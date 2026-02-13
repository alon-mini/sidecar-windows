---
name: drag-pane
description: >
  Drag-and-drop pane resizing implementation for two-pane plugin layouts.
  Covers mouse event handling via the internal/mouse package, hit region
  registration, drag delta calculation, width clamping, state persistence,
  and pane layout management. Use when working on pane resizing, drag
  interactions, layout management, or adding drag-to-resize to a new plugin.
---

# Drag-to-Resize Pane Implementation

## Overview

Add drag-to-resize support for two-pane plugin layouts (sidebar + main content). Users click and drag the divider between panes to resize them.

## Prerequisites

- Plugin already has a two-pane layout (sidebar + main content)
- State persistence functions exist in `internal/state/state.go` (each plugin has its own getter/setter)
- Familiarity with `internal/mouse` package

## Existing Implementations

| Plugin | State Functions | Mouse File |
|--------|----------------|------------|
| FileBrowser | `GetFileBrowserTreeWidth()` / `SetFileBrowserTreeWidth()` | `internal/plugins/filebrowser/mouse.go` |
| GitStatus | `GetGitStatusSidebarWidth()` / `SetGitStatusSidebarWidth()` | `internal/plugins/gitstatus/mouse.go` |
| Conversations | `GetConversationsSideWidth()` / `SetConversationsSideWidth()` | `internal/plugins/conversations/mouse.go` |
| Workspace | `GetWorkspaceSidebarWidth()` / `SetWorkspaceSidebarWidth()` | `internal/plugins/workspace/view_list.go` |

## Implementation Steps

### Step 1: Add Mouse Handler to Plugin Struct

```go
import "github.com/marcus/sidecar/internal/mouse"

type Plugin struct {
    // ... other fields
    mouseHandler *mouse.Handler
    sidebarWidth int  // Current sidebar width (persisted)
}

func New() *Plugin {
    return &Plugin{
        mouseHandler: mouse.NewHandler(),
    }
}
```

### Step 2: Define Hit Region Constants

```go
const (
    regionSidebar     = "sidebar"
    regionMainPane    = "main-pane"
    regionPaneDivider = "pane-divider"
    dividerWidth      = 1  // Visual divider width
)
```

### Step 3: Initialize Width on First Render (NOT in Init)

**Important:** Do NOT load width in `Init()` - plugin dimensions (`p.width`) are not available yet. Initialize lazily on first render:

```go
func (p *Plugin) renderTwoPane() string {
    p.mouseHandler.HitMap.Clear() // CRITICAL: clear every render

    if p.sidebarWidth == 0 {
        p.sidebarWidth = state.GetYourPluginSidebarWidth()
        if p.sidebarWidth == 0 {
            available := p.width - dividerWidth
            p.sidebarWidth = available * 30 / 100 // Default 30%
        }
    }
    // ... rest of render
}
```

### Step 4: Handle MouseMsg in Update

```go
func (p *Plugin) Update(msg tea.Msg) (plugin.Plugin, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        return p.handleMouse(msg)
    }
}
```

### Step 5: Create mouse.go with Handlers

```go
func (p *Plugin) handleMouse(msg tea.MouseMsg) (*Plugin, tea.Cmd) {
    action := p.mouseHandler.HandleMouse(msg)
    switch action.Type {
    case mouse.ActionClick:
        return p.handleMouseClick(action)
    case mouse.ActionDrag:
        return p.handleMouseDrag(action)
    case mouse.ActionDragEnd:
        return p.handleMouseDragEnd()
    }
    return p, nil
}

func (p *Plugin) handleMouseClick(action mouse.MouseAction) (*Plugin, tea.Cmd) {
    if action.Region == nil {
        return p, nil
    }
    switch action.Region.ID {
    case regionSidebar:
        p.activePane = PaneSidebar
    case regionMainPane:
        p.activePane = PaneMain
    case regionPaneDivider:
        p.mouseHandler.StartDrag(action.X, action.Y, regionPaneDivider, p.sidebarWidth)
    }
    return p, nil
}

func (p *Plugin) handleMouseDrag(action mouse.MouseAction) (*Plugin, tea.Cmd) {
    if p.mouseHandler.DragRegion() != regionPaneDivider {
        return p, nil
    }
    startValue := p.mouseHandler.DragStartValue()
    newWidth := startValue + action.DragDX

    // Clamp to bounds
    // NOTE: Offset varies by plugin (border styling differences):
    // GitStatus: -5, FileBrowser: -6, Conversations: -5, Workspace: just dividerWidth
    available := p.width - 5 - dividerWidth
    minWidth := 25
    maxWidth := available - 40
    if newWidth < minWidth {
        newWidth = minWidth
    } else if newWidth > maxWidth {
        newWidth = maxWidth
    }
    p.sidebarWidth = newWidth
    return p, nil
}

func (p *Plugin) handleMouseDragEnd() (*Plugin, tea.Cmd) {
    _ = state.SetYourPluginSidebarWidth(p.sidebarWidth)
    return p, nil
}
```

### Step 6: Register Hit Regions in Render

**This is where most bugs occur.** Follow this pattern exactly:

```go
func (p *Plugin) renderTwoPane() string {
    p.mouseHandler.HitMap.Clear() // CRITICAL: clear every render

    available := p.width - 5 - dividerWidth
    sidebarWidth := p.sidebarWidth
    if sidebarWidth == 0 {
        sidebarWidth = available * 30 / 100
    }
    if sidebarWidth < 25 {
        sidebarWidth = 25
    }
    if sidebarWidth > available-40 {
        sidebarWidth = available - 40
    }
    mainWidth := available - sidebarWidth
    p.sidebarWidth = sidebarWidth

    // ... render panes and divider ...

    // CRITICAL: Register in priority order (last = highest priority)
    p.mouseHandler.HitMap.AddRect(regionSidebar, 0, 0, sidebarWidth, p.height, nil)
    mainX := sidebarWidth + dividerWidth
    p.mouseHandler.HitMap.AddRect(regionMainPane, mainX, 0, mainWidth, p.height, nil)
    // Divider LAST = highest priority
    dividerX := sidebarWidth
    dividerHitWidth := 3 // Wider than visual for easier clicking
    p.mouseHandler.HitMap.AddRect(regionPaneDivider, dividerX, 0, dividerHitWidth, p.height, nil)

    return content
}
```

### Step 7: Render Visible Divider

```go
func (p *Plugin) renderDivider(height int) string {
    dividerStyle := lipgloss.NewStyle().
        Foreground(styles.BorderNormal).
        MarginTop(1) // Aligns with pane content (below top border)

    var sb strings.Builder
    for i := 0; i < height; i++ {
        sb.WriteString("|")
        if i < height-1 {
            sb.WriteString("\n")
        }
    }
    return dividerStyle.Render(sb.String())
}
```

### Step 8: Add State Persistence

Add plugin-specific functions to `internal/state/state.go`:

```go
// In State struct
YourPluginSidebarWidth int `json:"yourPluginSidebarWidth,omitempty"`

// Getter
func GetYourPluginSidebarWidth() int {
    mu.RLock()
    defer mu.RUnlock()
    if current == nil { return 0 }
    return current.YourPluginSidebarWidth
}

// Setter
func SetYourPluginSidebarWidth(width int) error {
    mu.Lock()
    if current == nil { current = &State{} }
    current.YourPluginSidebarWidth = width
    mu.Unlock()
    return Save()
}
```

## Critical Rules

### Rule 1: Never Reset Width in View()

**WRONG:**
```go
func (p *Plugin) View(width, height int) string {
    p.sidebarWidth = width * 30 / 100 // BUG: Overwrites drag changes every render!
}
```

**CORRECT:** Width is only set when `sidebarWidth == 0`. All other code paths must not unconditionally overwrite it.

### Rule 2: Hit Region X Coordinates

Divider X position = `sidebarWidth`, NOT `sidebarWidth + borderWidth`.

When lipgloss renders `Width(sidebarWidth)`, the pane occupies columns 0 to sidebarWidth-1. The divider starts at column sidebarWidth.

### Rule 3: Hit Region Priority (Registration Order)

`HitMap.Test()` checks regions in **reverse order** - last added = checked first.

The divider region MUST be registered LAST so it takes priority over overlapping pane regions.

```go
// CORRECT ORDER:
p.mouseHandler.HitMap.AddRect(regionSidebar, ...)      // Lowest priority
p.mouseHandler.HitMap.AddRect(regionMainPane, ...)     // Medium priority
p.mouseHandler.HitMap.AddRect(regionPaneDivider, ...)  // HIGHEST priority (last)
```

### Rule 4: Divider Hit Width

Use `dividerHitWidth := 3` (wider than the visual 1-character divider) to make clicking easier.

### Rule 5: Height for Hit Regions

Use `p.height` for hit region height, not `paneHeight` or `paneHeight + 2`.

## Performance Optimization: Hit Region Caching

For plugins with many hit regions, use a dirty flag to avoid rebuilding every render:

```go
type Plugin struct {
    hitRegionsDirty bool
    prevWidth       int
    prevHeight      int
    prevScrollOff   int
}

func (p *Plugin) renderTwoPane() string {
    if p.width != p.prevWidth || p.height != p.prevHeight {
        p.hitRegionsDirty = true
        p.prevWidth = p.width
        p.prevHeight = p.height
    }
    if p.scrollOffset != p.prevScrollOff {
        p.hitRegionsDirty = true
        p.prevScrollOff = p.scrollOffset
    }

    // ... render content ...

    if p.hitRegionsDirty {
        p.mouseHandler.HitMap.Clear()
        // Register all hit regions...
        p.hitRegionsDirty = false
    }
    return content
}
```

Also mark `hitRegionsDirty = true` when:
- View mode changes (toggling list/detail)
- Content changes (items loaded, expanded/collapsed)
- Sidebar visibility toggles

See `internal/plugins/conversations/view_layout.go` and `plugin_input.go` for a complete implementation.

## Debugging

If drag is not working, add temporary logging:

```go
func (p *Plugin) handleMouseClick(action mouse.MouseAction) (*Plugin, tea.Cmd) {
    log.Printf("CLICK x=%d y=%d region=%v", action.X, action.Y, action.Region)
}
```

Common issues:
- **Region is nil or wrong pane:** Check X coordinate calculation and registration order
- **Drag starts but width does not change:** Check that `handleMouseDrag` is being called
- **Width resets after drag:** Search for code that sets `sidebarWidth` unconditionally
