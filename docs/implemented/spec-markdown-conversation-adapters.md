# Add Glamour Markdown Rendering to Conversations Plugin

## Overview

Add markdown rendering to LLM responses in the conversations plugin using Glamour. The content from adapters is already markdown but currently rendered as plain text with simple word wrapping that destroys formatting.

**User preferences:**

- Markdown enabled by default
- No toggle command needed
- Thinking blocks: plain text only
- Turn list previews: plain text only
- Markdown rendering only in detail pane and full-screen view

## Architecture

Create an abstraction layer for content rendering that:

- Wraps Glamour markdown rendering
- Falls back to plain text on errors
- Works independently of adapter implementations
- Provides caching for performance

## Implementation Steps

### 1. Create Markdown Renderer Module

**File:** `internal/plugins/conversations/markdown.go` (NEW)

**Core components:**

```go
type ContentRenderer interface {
    RenderContent(content string, width int) []string
}

type GlamourRenderer struct {
    renderer        *glamour.TermRenderer
    lastWidth       int
    cache           map[string][]string
    maxCacheEntries int
}
```

**Key functionality:**

- `NewGlamourRenderer()` - Initialize with "dark" theme matching `internal/styles/styles.go` colors
- `RenderContent()` - Render markdown to []string lines, cache results, handle width changes
- Fallback to `wrapText()` on any Glamour errors
- Cache with simple key: `content_prefix:length:width`
- Recreate renderer when width changes (clear cache)
- Thread-safe with mutex

**Theme config:**

- Use `glamour.WithStylePath("dark")` to match existing dark color scheme
- Use `glamour.WithWordWrap(width)` for width constraints

### 2. Integrate Renderer into Plugin

**File:** `internal/plugins/conversations/plugin.go` (MODIFY)

Add to Plugin struct:

```go
contentRenderer *GlamourRenderer
```

Initialize in `New()`:

```go
renderer, err := NewGlamourRenderer()
if err != nil {
    // Log warning, set to nil (will use plain text fallback)
}
p.contentRenderer = renderer
```

Add helper method:

```go
func (p *Plugin) renderContent(content string, width int) []string {
    if p.contentRenderer != nil {
        return p.contentRenderer.RenderContent(content, width)
    }
    return wrapText(content, width)
}
```

### 3. Update View Rendering

**File:** `internal/plugins/conversations/view.go` (MODIFY)

**Location 1: Detail pane rendering (lines 1579-1585)**

Replace:

```go
msgLines := wrapText(msg.Content, contentWidth-2)
for _, line := range msgLines {
    contentLines = append(contentLines, styles.Body.Render(line))
}
```

With:

```go
msgLines := p.renderContent(msg.Content, contentWidth-2)
for _, line := range msgLines {
    contentLines = append(contentLines, line) // Glamour already styles
}
```

**Location 2: Full-screen detail view (lines 1972-1980)**

Replace:

```go
msgLines := wrapText(msg.Content, p.width-4)
for _, line := range msgLines {
    contentLines = append(contentLines, " "+styles.Body.Render(line))
}
```

With:

```go
msgLines := p.renderContent(msg.Content, p.width-4)
for _, line := range msgLines {
    contentLines = append(contentLines, " "+line)
}
```

**Keep unchanged:**

- Thinking blocks (lines 1571-1574) - continue using `wrapText()` with `styles.Muted`
- Turn previews in `renderCompactTurn()` - continue using plain text
- Tool use rendering - continue using `styles.Code`

### 4. Add Tests

**File:** `internal/plugins/conversations/markdown_test.go` (NEW)

Test coverage:

- `TestGlamourRenderer_Basic` - Basic markdown rendering (headers, lists, code blocks)
- `TestGlamourRenderer_WidthChange` - Renderer recreates with new width, cache clears
- `TestGlamourRenderer_Fallback` - Falls back to wrapText on errors
- `TestGlamourRenderer_Cache` - Cache hit/miss behavior
- `TestGlamourRenderer_EmptyContent` - Edge case handling
- `TestRenderContent_Integration` - Plugin.renderContent() with nil renderer fallback

### 5. Implementation Details

**Glamour configuration:**

```go
renderer, err := glamour.NewTermRenderer(
    glamour.WithStylePath("dark"),
    glamour.WithWordWrap(width),
)
```

**Width change handling:**

- Detect width delta (recreate renderer if changed)
- Clear cache on width change
- Thread-safe with mutex

**Caching strategy:**

```go
// Simple cache key
func cacheKey(content string, width int) string {
    prefix := content
    if len(prefix) > 50 {
        prefix = prefix[:50]
    }
    return fmt.Sprintf("%s:%d:%d", prefix, len(content), width)
}
```

**Performance considerations:**

- Max 100 cache entries (clear if exceeded)
- Cache only for rendered content, not previews
- No caching for thinking blocks (always plain text)

**Error handling:**

- Log Glamour initialization errors
- Fall back to `wrapText()` if renderer is nil
- Fall back to `wrapText()` if Glamour.Render() errors

## Files to Modify/Create

**CREATE:**

- `internal/plugins/conversations/markdown.go` - Core markdown renderer with Glamour
- `internal/plugins/conversations/markdown_test.go` - Comprehensive test suite

**MODIFY:**

- `internal/plugins/conversations/plugin.go` - Add renderer field, initialize, add helper method
- `internal/plugins/conversations/view.go` - Update 2 rendering locations (lines ~1579-1585, ~1972-1980)

**REFERENCE (no changes):**

- `internal/styles/styles.go` - Color palette for Glamour theme matching
- `internal/adapter/adapter.go` - Message interface (content already markdown)

## Success Criteria

- ✅ Markdown code blocks render with proper formatting
- ✅ Lists, headers, emphasis render correctly
- ✅ Width constraints respected (no header scroll-off)
- ✅ Fallback to plain text works if Glamour fails
- ✅ Thinking blocks remain plain text with muted styling
- ✅ Turn previews remain plain text
- ✅ Tests pass with good coverage
- ✅ No performance regression (caching effective)

## Edge Cases Handled

- Invalid markdown → Glamour treats as plain text
- Very large content → Cache size limits prevent memory issues
- Width changes → Renderer recreates, cache clears
- Glamour initialization failure → Graceful fallback to plain text
- Thread safety → Mutex protects renderer and cache

## Notes

- Glamour v0.10.0 already in go.mod dependencies
- No breaking changes to adapter interface
- Solution is adapter-agnostic (works with any `adapter.Message.Content`)
- wrapText() function kept for backward compatibility and non-content uses
- Clean separation: rendering logic separate from view composition
