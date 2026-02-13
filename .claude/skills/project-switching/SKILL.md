---
name: project-switching
description: >
  Project switching implementation in sidecar: project discovery, state management,
  UI flow, modal rendering, filtering, theme preview, and plugin reinitialization.
  Use when working on the project switcher feature, project management, worktree
  switching, or the project configuration system.
user-invocable: false
---

# Project Switching

Switch between git repositories without restarting sidecar. Press `@` to open the project switcher modal.

## Architecture Overview

The project switcher uses the app's declarative modal library (`internal/modal/`) for rendering and mouse handling.

### Key Files

| File | Contents |
|------|----------|
| `internal/app/model.go` | State, init, reset, filter, switch, theme preview |
| `internal/app/update.go` | Keyboard and mouse handlers |
| `internal/app/view.go` | Modal building and section rendering |
| `internal/modal/` | Modal library (builder, sections, layout) |
| `internal/config/types.go` | ProjectConfig struct |
| `internal/config/loader.go` | Config loading with path validation |

### Model State

```go
// internal/app/model.go
showProjectSwitcher         bool
projectSwitcherCursor       int
projectSwitcherScroll       int
projectSwitcherInput        textinput.Model
projectSwitcherFiltered     []config.ProjectConfig
projectSwitcherModal        *modal.Modal           // Cached modal instance
projectSwitcherModalWidth   int                    // Width for cache invalidation
projectSwitcherMouseHandler *mouse.Handler
```

## Project vs Worktree Switching

- **Project Switching** (`@`): Switch between configured projects from `config.json` (arbitrary repos)
- **Worktree Switching** (`W`): Switch between git worktrees within the current repository

## Configuration

Projects are configured in `~/.config/sidecar/config.json`:

```json
{
  "projects": {
    "list": [
      {"name": "sidecar", "path": "~/code/sidecar"},
      {"name": "td", "path": "~/code/td", "theme": "dark"}
    ]
  }
}
```

Paths support `~` expansion. Projects can have per-project themes.

## Core Flow

### Opening (`@` key)

```go
case "@":
    m.showProjectSwitcher = !m.showProjectSwitcher
    if m.showProjectSwitcher {
        m.activeContext = "project-switcher"
        m.initProjectSwitcher()
    } else {
        m.resetProjectSwitcher()
        m.updateContext()
    }
```

### Initialization (`model.go:433-454`)

`initProjectSwitcher()` clears cached modal, creates text input with "Filter projects..." placeholder, loads all projects, pre-selects current project, and previews its theme.

### Cleanup (`model.go:413-423`)

`resetProjectSwitcher()` resets all state, clears modal cache, and restores current project's theme (undoing any live preview).

## Modal Rendering

Uses `internal/modal/` with a builder pattern and lazy caching.

### Modal Structure

```
+-------------------------------------------+
| Switch Project                            |  <- Title
| [Filter projects...                    ]  |  <- Input section
| 3 of 10 projects                          |  <- Count section
|   ^ 2 more above                          |  <- Scroll indicator
| > sidecar                                 |  <- Selected item
|   ~/code/sidecar                          |
|   td (current)                            |  <- Current project (green)
|   v 5 more below                          |  <- Scroll indicator
| enter switch  up/down navigate  esc close |  <- Hints section
+-------------------------------------------+
```

### Caching (`view.go:114-137`)

`ensureProjectSwitcherModal()` builds modal only when it does not exist or width changed. Call `clearProjectSwitcherModal()` when content changes (filter input, cursor movement with scroll).

### Section Types

| Type | Factory | Purpose |
|------|---------|---------|
| Input | `modal.Input()` | Text input with focus |
| Custom | `modal.Custom()` | Complex content with focusables |
| Text | `modal.Text()` | Static text |
| Buttons | `modal.Buttons()` | Button row |

## Keyboard Handling (`update.go:600-690`)

Priority: KeyType switch (special keys) -> String switch (named keys) -> Fallthrough to textinput.

| Key | Action |
|-----|--------|
| `Esc` | Clear filter (if set) or close modal |
| `Enter` | Switch to selected project |
| `Up/Down` | Arrow navigation |
| `ctrl+n/ctrl+p` | Emacs-style navigation |
| Other keys | Forwarded to text input for filtering |

### Esc Behavior

First Esc clears filter if set. Second Esc (or first with empty filter) closes modal.

### Navigation

Cursor movement updates `projectSwitcherCursor` and calls `projectSwitcherEnsureCursorVisible()` to maintain scroll window (max 8 visible items).

## Filtering (`model.go:457-470`)

`filterProjects()` does case-insensitive substring match on both `Name` and `Path` fields. On filter change: clear modal cache, clamp cursor, reset scroll, preview theme.

## Mouse Handling (`update.go:1073-1115`)

Uses modal library's `HandleMouse()`. Each project item has a focusable ID (`project-switcher-item-N`). Click on item triggers switch. Hover state managed by modal library via `hoverID` string.

## Theme Preview (`model.go:580-586`)

`previewProjectTheme()` applies the selected project's theme live. Called on init, cursor movement, and filter changes. Theme is restored in `resetProjectSwitcher()`.

## Project Switching (`model.go:485-577`)

`switchProject()` performs:

1. Skip if same project (show toast)
2. Save active plugin state for old workdir
3. Check for saved worktree to restore
4. Update `m.ui.WorkDir` and repo name
5. Apply project-specific theme
6. Reinitialize all plugins via `m.registry.Reinit(targetPath)`
7. Send `WindowSizeMsg` for layout recalculation
8. Restore previously active plugin for new workdir
9. Return toast notification

### What Happens on Switch

1. All plugins stop (file watchers, git commands, etc.)
2. Plugin context updates to new working directory
3. All plugins reinitialize with new path
4. Previously active plugin for that project is restored
5. Toast notification confirms the switch

## State Persistence

Per-project state saved in `~/.config/sidecar/state.json`:
- Active plugin per project
- File browser cursor position and expanded directories
- Sidebar widths and view preferences

## Common Pitfalls

1. **Forgetting updateContext()** -- Call after closing modal to restore app context
2. **Stale modal cache** -- Call `clearProjectSwitcherModal()` when content changes
3. **Cursor out of bounds** -- Always clamp after filtering
4. **Printable keys vs navigation** -- Keep printable characters routed to textinput
5. **Theme preview cleanup** -- Always restore theme in `resetProjectSwitcher()`
6. **Focusable coordinates** -- Each project takes 2 lines (name + path)

## Adding Features

See [references/implementation-details.md](references/implementation-details.md) for detailed implementation patterns including adding keyboard shortcuts, project metadata, filter algorithms, and new modal sections.
