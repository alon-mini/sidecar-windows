---
name: ui-features
description: Implementing UI/UX features in sidecar including modals (internal/modal library), keyboard shortcuts, mouse support, scrolling, pill/tab rendering, and pane resizing. Use when implementing UI features, handling user input, adding keyboard shortcuts, building modals, or working on UX improvements.
---

# UI Feature Implementation

Single entry point for sidecar UI work. All new modals must use `internal/modal`. For complete keyboard shortcut listings, see `references/keyboard-shortcuts-reference.md`.

## Quick Checklist

- Modals: use `internal/modal`, render with `ui.OverlayModal`, avoid manual hit region math
- Pills/chips/tabs: use `styles.RenderPillWithStyle`; auto-fallback when `nerdFontsEnabled` is false
- Keyboard: Commands + FocusContext + bindings must match; names short; priorities set
- Mouse: rebuild hit regions on each render; add general regions first, specific last
- Rendering: keep output within View width/height to avoid header/footer overlap. Use `contentHeight := height - headerLines - footerLines`
- Testing: verify keyboard, mouse, hover, scrolling, and footer hints
- Plugins must NOT render their own footer -- the app renders a unified footer from `Commands()`

## Modals (internal/modal)

All new modals must use `internal/modal`. See `docs/guides/declarative-modal-guide.md` for the full API.

### Create a modal

```go
m := modal.New("Delete Worktree?",
    modal.WithWidth(58),
    modal.WithVariant(modal.VariantDanger),
    modal.WithPrimaryAction("delete"),
).
    AddSection(modal.Text("Name: " + wt.Name)).
    AddSection(modal.Spacer()).
    AddSection(modal.Buttons(
        modal.Btn(" Delete ", "delete", modal.BtnDanger()),
        modal.Btn(" Cancel ", "cancel"),
    ))
```

### Render in View

```go
func (p *Plugin) renderDeleteView(width, height int) string {
    background := p.renderListView(width, height)
    rendered := p.deleteModal.Render(width, height, p.mouseHandler)
    return ui.OverlayModal(background, rendered, width, height)
}
```

### Handle input in Update

```go
case tea.KeyMsg:
    action, cmd := p.deleteModal.HandleKey(msg)
    if action != "" {
        return p.handleModalAction(action)
    }
    return p, cmd

case tea.MouseMsg:
    action := p.deleteModal.HandleMouse(msg, p.mouseHandler)
    if action != "" {
        return p.handleModalAction(action)
    }
    return p, nil
```

### Modal initialization and caching (critical)

Always call `ensureModal()` in BOTH View and Update handlers. Create an ensure function that:
1. Returns early if required state is missing
2. Caches based on width to avoid rebuilding every frame
3. Creates the modal only when needed

```go
func (p *Plugin) ensureMyModal() {
    if p.targetItem == nil {
        return
    }
    modalW := 50
    if modalW > p.width-4 { modalW = p.width - 4 }
    if modalW < 20 { modalW = 20 }
    if p.myModal != nil && p.myModalWidthCache == modalW {
        return
    }
    p.myModalWidthCache = modalW
    p.myModal = modal.New("Title", modal.WithWidth(modalW), ...).
        AddSection(...)
}
```

**The key handler MUST call ensure before checking nil:**

```go
func (p *Plugin) handleMyModalKeys(msg tea.KeyMsg) tea.Cmd {
    p.ensureMyModal()  // CRITICAL: Initialize before nil check
    if p.myModal == nil { return nil }
    action, cmd := p.myModal.HandleKey(msg)
    return cmd
}
```

### Async content invalidation

When modal content depends on async data, invalidate the cache when data arrives:

```go
case MyDataLoadedMsg:
    p.myData = msg.Data
    p.clearMyModal()  // Force rebuild with new content
    return p, nil
```

### Modal keyboard shortcuts and footer hints

Modals need their own focus context and commands for footer hints:

1. Return a dedicated context from `FocusContext()`
2. Add commands for the modal context in `Commands()`
3. Add bindings in `internal/keymap/bindings.go`
4. Intercept custom keys before `modal.HandleKey` (Tab/Enter/Esc are handled internally)

```go
func (p *Plugin) FocusContext() string {
    switch p.viewMode {
    case ViewModeError:  return "git-error"
    case ViewModePushMenu: return "git-push-menu"
    default: return "git-status"
    }
}
```

### Modal notes

- `HandleKey`/`HandleMouse` handle Tab, Shift+Tab, Enter, Esc internally
- Backdrop clicks return "cancel"; use `WithCloseOnBackdropClick(false)` to disable
- Use built-in sections (Text, Input, Textarea, Buttons, Checkbox, List, When) before custom layouts
- For bespoke layouts, use `modal.Custom` and return explicit focusable offsets
- `SetFocus(id)` auto-scrolls viewport to focused element
- Prefer `ui.OverlayModal(background, modal, width, height)` for dimmed overlays; do not pre-center with `lipgloss.Place`

### Background colors (critical)

Lipgloss `Background()` does not cascade into child content. ANSI resets clear the parent background. Solution: replace ANSI resets within viewport lines with reset + background re-apply, then pad short lines. See `fillBackground` in `internal/modal/layout.go`.

## Pill-Shaped Elements (internal/styles)

Controlled by `nerdFontsEnabled` in `~/.config/sidecar/config.json` (`ui.nerdFontsEnabled`).

```go
// With explicit colors
label := styles.RenderPill("Output", styles.TextPrimary, styles.Primary, "")

// With a lipgloss.Style (preferred for tabs/chips)
active := styles.RenderPillWithStyle("Output", styles.BarChipActive, "")
inactive := styles.RenderPillWithStyle("Diff", styles.BarChip, "")
```

Available styles: `styles.BarChip` (inactive), `styles.BarChipActive` (active), or custom `lipgloss.Style`.

Test with both `nerdFontsEnabled: true` and `false` to verify fallback.

## Keyboard Shortcuts

For complete per-plugin shortcut listings, see `references/keyboard-shortcuts-reference.md`.

### Three things must match

1. **Command ID** in `Commands()` (e.g., `"stage-file"`)
2. **Binding command** in `internal/keymap/bindings.go` (e.g., `"stage-file"`)
3. **Context string** in both places (e.g., `"git-status"`)

```go
// 1) Commands()
{ID: "stage-file", Name: "Stage", Context: "git-status", Priority: 1}

// 2) FocusContext()
func (p *Plugin) FocusContext() string { return "git-status" }

// 3) bindings.go
{Key: "s", Command: "stage-file", Context: "git-status"}
```

### Multiple contexts (view modes)

Return different context strings from `FocusContext()` for different modes. Each context gets its own footer hints and key bindings.

### Priority guidelines

- **1**: Primary actions (Stage, Commit, Open)
- **2**: Secondary actions (Diff, Search, Push)
- **3**: Tertiary actions (History, Refresh)
- **4+**: Palette only

### Root contexts (q behavior)

In root contexts, `q` shows quit confirmation. In non-root, `q` navigates back. Root contexts: `global`, `conversations`, `conversations-sidebar`, `git-status`, `git-status-commits`, `git-status-diff`, `file-browser-tree`, `workspace-list`, `td-monitor`.

Update `isRootContext()` in `internal/app/update.go` when adding new contexts.

### Text input contexts

When a view has text input, implement `plugin.TextInputConsumer` and return `true` while active. This prevents app-level shortcuts from intercepting typed characters.

```go
func (p *Plugin) ConsumesTextInput() bool {
    return p.showMyModal
}
```

### Footer rendering flow

```
footerHints()
    +-- pluginFooterHints() -> Commands() filtered by FocusContext(), sorted by Priority
    +-- globalFooterHints() -> App-level hints
renderHintLineTruncated(hints, availableWidth)
    -> Renders left-to-right until width exceeded
```

### Keyboard checklist

- Command in `Commands()` with ID, Name, Context, Priority
- `FocusContext()` returns matching context
- Binding in `internal/keymap/bindings.go`
- Key handled in `Update()` if app does not intercept
- No conflicting keys in same context
- Short footer hint names, primary actions Priority 1-2
- Verify `q` behavior with `isRootContext()`

### Core files

| File | Purpose |
|------|---------|
| `internal/plugin/plugin.go` | Command struct, Commands(), FocusContext(), TextInputConsumer |
| `internal/keymap/bindings.go` | Default key-to-command mappings |
| `internal/keymap/registry.go` | Runtime binding lookup |
| `internal/app/update.go` | Key routing, isRootContext() |
| `internal/app/view.go` | Footer rendering |

## Scrollbar (internal/ui)

```go
ui.RenderScrollbar(ui.ScrollbarParams{
    TotalItems:   len(items),
    ScrollOffset: p.scrollOffset,
    VisibleItems: visibleCount,
    TrackHeight:  height,
})
```

Pattern: reduce content width by 1, render content, render scrollbar, join horizontally with `lipgloss.JoinHorizontal(lipgloss.Top, content, scrollbar)`.

For multi-line items, set `TrackHeight` to actual terminal rows: `visibleCount * linesPerItem`.

## Mouse Support

### Setup

```go
type Plugin struct {
    mouseHandler *mouse.Handler
}
func New() *Plugin {
    return &Plugin{mouseHandler: mouse.NewHandler()}
}
```

### Register hit regions during render

```go
func (p *Plugin) View(width, height int) string {
    p.mouseHandler.Clear()
    p.mouseHandler.HitMap.AddRect("pane", 0, 0, width, height, nil)
    p.mouseHandler.HitMap.AddRect("item", 2, 5, width-4, 1, 0)
    return content
}
```

### Region ordering (critical)

Regions tested in reverse order. Add general regions first, specific regions last.

### Coordinate system

App offsets Y by 2 (header height) before forwarding to plugins. Plugins operate in local coords where Y=0 is plugin content top.

### Common patterns

- Click to select/focus, scroll wheel to move, double-click to open
- Drag regions for pane resizing
- Hover for visual feedback (focus takes precedence)

### Mouse troubleshooting

| Symptom | Fix |
|---------|-----|
| Clicks don't register | Check region order (pane first) |
| Y offsets wrong | Account for borders, padding, headers |
| Scroll over items broken | Include item regions in scroll routing |
| Double-click fails | Ensure consistent region ID/bounds |
| Drag broken | Call StartDrag on click, check DragRegion during drag |
