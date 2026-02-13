---
name: inline-editor
description: >
  Inline text editing implementation within the file browser preview pane using
  tmux PTY backend, cursor movement, text manipulation, and editor state
  management. Covers entry/exit lifecycle, dimension calculations, confirmation
  dialogs, click-away detection, mouse forwarding, and app-level key routing.
  Use when working on inline editing features, text input components, or
  debugging editor rendering/input issues in the file browser plugin.
---

# Inline Editor Implementation

## Overview

The inline editor (`tmux_inline_edit`) lets users edit files directly within the file browser preview pane using their preferred terminal editor (vim, nvim, nano, etc.) without leaving the TUI. The file tree remains visible during editing.

**Core Principle**: This is NOT a terminal emulator. Tmux manages the PTY backend; Sidecar acts as an input/output relay, similar to the workspace plugin's interactive mode.

## Architecture

### Components

1. **Entry Layer** (`internal/plugins/filebrowser/inline_edit.go`): Creates tmux sessions, manages editor lifecycle
2. **Rendering Layer** (`internal/plugins/filebrowser/view.go`, `inline_edit.go`): Renders editor content within preview pane
3. **Input Layer** (`internal/plugins/filebrowser/plugin.go`, `mouse.go`): Routes keys/clicks to editor or confirmation dialog
4. **TTY Model** (`internal/tty/tty.go`): Handles tmux capture, cursor overlay, and input forwarding

### Data Flow

```
User presses 'e' on file
  -> enterInlineEditMode()
  -> tmux new-session -d -s {sessionName} {editor} {path}
  -> InlineEditStartedMsg
  -> handleInlineEditStarted()
  -> tty.Model.Enter()
  -> Start polling tmux capture-pane
  -> renderInlineEditorContent() in preview pane
  -> User types -> tty.Model forwards to tmux
  -> User exits -> SessionDeadMsg or exit keys
  -> exitInlineEditMode()
  -> Refresh preview
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/plugins/filebrowser/inline_edit.go` | Editor lifecycle, confirmation dialog, dimension calculations |
| `internal/plugins/filebrowser/view.go` | Preview pane rendering, gradient border |
| `internal/plugins/filebrowser/mouse.go` | Click-away detection |
| `internal/plugins/filebrowser/plugin.go` | State management, Update routing |
| `internal/tty/tty.go` | TTY model for tmux interaction (shared with workspace) |
| `internal/app/update.go` | App-level key routing for inline edit context |

## Critical Implementation Details

### 1. Preview Pane Rendering (Not Full-Screen)

The editor renders within `renderPreviewPane()`, NOT as a full-screen takeover. The file tree stays visible.

```go
// view.go - renderPreviewPane()
func (p *Plugin) renderPreviewPane(visibleHeight int) string {
    if p.inlineEditMode && p.inlineEditor != nil && p.inlineEditor.IsActive() {
        return p.renderInlineEditorContent(visibleHeight)
    }
    // ... normal preview rendering
}
```

### 2. Dimension Calculations

The tty.Model needs exact dimensions matching the preview pane content area:

```go
func (p *Plugin) calculateInlineEditorWidth() int {
    if !p.treeVisible {
        return p.width - 4 // borders + padding
    }
    p.calculatePaneWidths()
    return p.previewWidth - 4
}

func (p *Plugin) calculateInlineEditorHeight() int {
    paneHeight := p.height
    innerHeight := paneHeight - 2 // pane borders
    contentHeight := innerHeight - 2 // header lines
    if len(p.tabs) > 1 {
        contentHeight-- // tab line
    }
    return contentHeight
}
```

**These MUST stay in sync with `renderInlineEditorContent()` layout calculations.**

### 3. Confirmation Behavior

**Rule: session alive = show confirmation, session dead = exit immediately.**

Always show confirmation when the session is alive, regardless of file modification status. Vim's modification status cannot be reliably detected externally.

```go
func (p *Plugin) isInlineEditSessionAlive() bool {
    if p.inlineEditSession == "" {
        return false
    }
    err := exec.Command("tmux", "has-session", "-t", p.inlineEditSession).Run()
    return err == nil
}
```

Check session alive status:
1. At the start of `Update()` when in inline edit mode - if dead, exit immediately
2. In click-away handling - if dead, skip confirmation and clean up

### 4. Exit Confirmation Dialog

State fields:

```go
showExitConfirmation bool        // Dialog visible
pendingClickRegion   string      // Where user clicked
pendingClickData     interface{} // Click data (tree index, tab index)
exitConfirmSelection int         // 0=Save&Exit, 1=Exit without saving, 2=Cancel
```

Options:
- **Save & Exit**: Sends editor-appropriate save-and-quit sequence, waits for session death
- **Exit without saving**: Kills tmux session immediately
- **Cancel**: Returns to editing

### 5. Click-Away Detection

Mouse regions are registered during render. Clicks between items may miss regions, so always include position-based fallback:

```go
if p.inlineEditMode && p.inlineEditor != nil && p.inlineEditor.IsActive() {
    action := p.mouseHandler.HandleMouse(msg)

    handleClickAway := func(regionID string, regionData interface{}) (*Plugin, tea.Cmd) {
        if !p.isInlineEditSessionAlive() {
            p.exitInlineEditMode()
            p.pendingClickRegion = regionID
            p.pendingClickData = regionData
            return p.processPendingClickAction()
        }
        p.pendingClickRegion = regionID
        p.pendingClickData = regionData
        p.showExitConfirmation = true
        p.exitConfirmSelection = 0
        return p, nil
    }

    if action.Type == mouse.ActionClick {
        if action.Region != nil {
            switch action.Region.ID {
            case regionTreePane, regionTreeItem, regionPreviewTab:
                return handleClickAway(action.Region.ID, action.Region.Data)
            }
        }
        // Fallback: position-based detection
        if p.treeVisible && action.X < p.treeWidth {
            return handleClickAway(regionTreePane, nil)
        }
    }
    // Forward to tty model
    return p, p.inlineEditor.Update(msg)
}
```

### 6. Gradient Border Feedback

Visual indicator that edit mode is active:

```go
if p.inlineEditMode && p.inlineEditor != nil && p.inlineEditor.IsActive() {
    rightPane = styles.RenderPanelWithGradient(previewContent, p.previewWidth,
        paneHeight, styles.GetInteractiveGradient())
}
```

### 7. Mouse Support (SGR Protocol)

Full mouse interaction including text selection via SGR (1006) protocol. Mouse events are forwarded to the tty model which translates them into SGR escape sequences: `\x1b[<button;x;y;M/m` where `M` = press/drag, `m` = release.

### 8. Multi-Editor Support

`sendEditorSaveAndQuit()` detects which editor is running and sends the appropriate sequence:

| Editor | Save & Quit Command |
|--------|---------------------|
| vim, nvim, vi | `Escape :wq Enter` |
| nano | `Ctrl+O Enter Ctrl+X` |
| emacs | `Ctrl+X Ctrl+S Ctrl+X Ctrl+C` |
| helix | `Escape :wq Enter` |
| micro | `Ctrl+S Ctrl+Q` |
| kakoune | `Escape :write-quit Enter` |
| joe | `Ctrl+K X` |
| ne | `Escape :SaveQuit Enter` |
| amp | `Ctrl+S Ctrl+Q` |

## Exit Paths

| Method | Confirmation | Description |
|--------|--------------|-------------|
| `Ctrl+\` | No | Immediate exit (tty.Config.ExitKey) |
| Double-ESC | No | Exit with 150ms delay (vim ESC compatibility) |
| `:q`, `:wq` in vim | No | Normal editor exit, session death detected |
| Click tree/tab | Yes (if alive) | Shows confirmation when session alive; exits immediately if dead |

## State Management

### Plugin State (plugin.go)

```go
inlineEditor         *tty.Model // Embeddable tty model
inlineEditMode       bool       // Currently editing
inlineEditSession    string     // Tmux session name
inlineEditFile       string     // File being edited
showExitConfirmation bool
pendingClickRegion   string
pendingClickData     interface{}
exitConfirmSelection int
```

### Update Priority (plugin.go)

```go
func (p *Plugin) Update(msg tea.Msg) (plugin.Plugin, tea.Cmd) {
    // 1. Handle exit confirmation dialog FIRST
    if p.showExitConfirmation { /* j/k navigation, Enter confirm, Esc cancel */ }
    // 2. Handle inline edit mode
    if p.inlineEditMode && p.inlineEditor.IsActive() { /* Delegate to tty.Model */ }
    // 3. Normal plugin handling
}
```

### App-Level Key Routing

The app intercepts global shortcuts (`q`, `1-5`, `` ` ``, `~`, `?`, `!`, `@`) before plugins. For inline edit to receive ALL keys, `internal/app/update.go` must recognize the context:

```go
if m.activeContext == "workspace-interactive" || m.activeContext == "file-browser-inline-edit" {
    // Forward ALL keys to plugin
}
```

The plugin returns `"file-browser-inline-edit"` from `FocusContext()`.

## Common Pitfalls

1. **Do NOT render full-screen** - render within preview pane only, keeping file tree visible
2. **Always include position-based click fallback** - mouse regions may not cover every pixel
3. **Always confirm on click-away** when session is alive - prevents accidental data loss
4. **Check `IsActive()` on every message** - tty model can become inactive asynchronously; without this check, users get a blank screen after vim exits
5. **Never query tmux synchronously in View()** - use cached content from `tty.Model.View()`
6. **Handle tab row clicks by Y position BEFORE region-based detection** - prevents incorrect handling as preview pane clicks
7. **Ensure app-level key routing** recognizes `"file-browser-inline-edit"` context - otherwise typing `q` in vim triggers quit instead of inserting character

## Feature Flag

Gated behind `tmux_inline_edit`. Enable in `~/.config/sidecar/config.json`:

```json
{ "features": { "tmux_inline_edit": true } }
```

Fallback: `features.IsEnabled(features.TmuxInlineEdit.Name)` returns false -> opens external editor.

## Keyboard Shortcuts

| Key | Command | Description |
|-----|---------|-------------|
| `e` | `edit` | Edit file inline (within preview pane) |
| `E` | `edit-external` | Edit in full terminal (suspends TUI) |

Registered in:
- `internal/plugins/filebrowser/plugin.go` - `Commands()` method
- `internal/plugins/filebrowser/handlers.go` - Key handling
- `internal/keymap/bindings.go` - Key bindings for `file-browser-tree` and `file-browser-preview` contexts

## References

- Interactive shell guide: `docs/guides/interactive-shell-implementation.md`
- TTY model: `internal/tty/tty.go`
- Feature flags: `internal/features/features.go`
