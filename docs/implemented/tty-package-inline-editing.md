# Plan: Abstract Interactive Shell for Cross-Plugin Reuse

**Epic**: td-50cf27
**Status**: Planning complete, ready for implementation

## Goal
Extract the interactive shell from the workspace plugin into a reusable `internal/tty` package so plugins (files, git) can run CLI editors (vim, nano) in-pane.

## Architecture

Create `internal/tty/` package with embeddable `tty.Model`:
- Plugins embed `tty.Model` and delegate Update/View when active
- Package handles tmux session management, key mapping, polling, cursor rendering
- Plugins only manage: when to start, what command to run, what to do on exit

## Files to Create

| File | Purpose |
|------|---------|
| `internal/tty/tty.go` | Core `Model` struct, `Config`, `New()`, `Start()`, `Stop()`, `Active()`, `Update()`, `View()` |
| `internal/tty/keymap.go` | `MapKeyToTmux()` extracted from `workspace/interactive.go:129-287` |
| `internal/tty/cursor.go` | `renderWithCursor()` - cursor overlay rendering |
| `internal/tty/terminal_mode.go` | `detectBracketedPasteMode()`, `detectMouseReportingMode()` |
| `internal/tty/paste.go` | Paste handling with bracketed paste support |
| `internal/tty/session.go` | `CreateSession()`, `KillSession()`, `ResizePane()`, capture helpers |
| `internal/tty/polling.go` | Adaptive polling (50ms→200ms→500ms decay) |
| `internal/tty/messages.go` | `TTYOutputMsg`, `TTYSessionDeadMsg`, `TTYExitRequestedMsg` |
| `internal/tty/output_buffer.go` | `OutputBuffer` with hash-based change detection |

## Files to Modify

| File | Changes |
|------|---------|
| `internal/plugins/workspace/interactive.go` | Import from `tty` package, delegate to extracted functions |
| `internal/plugins/workspace/types.go` | Remove `OutputBuffer` (moved to tty) |
| `internal/plugins/filebrowser/plugin.go` | Add `ttyModel tty.Model`, `editingFile string` fields |
| `internal/plugins/filebrowser/handlers.go` | Add handler for `e` key to enter inline edit mode |
| `internal/plugins/filebrowser/view.go` | Render `ttyModel.View()` when editing |
| `internal/app/update.go` | Add `"file-browser-interactive"` to text input contexts |

## Public API (tty.Model)

```go
type Config struct {
    ExitKey    string        // Default: "ctrl+\\"
    CopyKey    string        // Default: "alt+c"
    PasteKey   string        // Default: "alt+v"
    PollFast   time.Duration // Default: 50ms
    PollMedium time.Duration // Default: 200ms
    PollSlow   time.Duration // Default: 500ms
}

func New(cfg Config) Model
func (m *Model) Start(sessionName, paneID string, width, height int) tea.Cmd
func (m *Model) Stop() tea.Cmd
func (m Model) Active() bool
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)
func (m Model) View(width, height int) string
```

## Implementation Phases

### Phase 1: Extract Core Utilities (Non-Breaking)
1. Create `internal/tty/keymap.go` - copy `MapKeyToTmux()` verbatim
2. Create `internal/tty/cursor.go` - copy cursor rendering functions
3. Create `internal/tty/terminal_mode.go` - copy mode detection
4. Create `internal/tty/paste.go` - copy paste handling
5. Update workspace imports to use `tty.*` functions
6. Run tests - no behavior change

### Phase 2: Create TTY Model
1. Create `internal/tty/tty.go` with `Model`, `Config`
2. Create `internal/tty/session.go` - tmux session lifecycle
3. Create `internal/tty/polling.go` - adaptive polling
4. Create `internal/tty/messages.go` - message types
5. Create `internal/tty/output_buffer.go` - move from workspace/types.go
6. Add tests

### Phase 3: Integrate with Files Plugin
1. Add `ttyModel` field to files plugin
2. Add `editInPreview()` function that:
   - Creates tmux session: `sidecar-edit-{timestamp}`
   - Runs: `$EDITOR <filepath>`
   - Calls `ttyModel.Start()`
3. Handle `TTYExitRequestedMsg` / `TTYSessionDeadMsg` - reload preview
4. Update View() to render `ttyModel.View()` when active
5. Add `"file-browser-interactive"` FocusContext
6. Update `Commands()` for edit hints

### Phase 4: Migrate Workspace Plugin (Optional)
1. Add `ttyModel` field alongside existing `interactiveState`
2. Gradually delegate to `ttyModel`
3. Remove duplicated code

## Files Plugin Integration Example

```go
// In handlers.go, 'e' key when preview focused:
func (p *Plugin) editInPreview() tea.Cmd {
    editor := os.Getenv("EDITOR")
    if editor == "" { editor = "vim" }

    sessionName := fmt.Sprintf("sidecar-edit-%d", time.Now().UnixNano())
    fullPath := filepath.Join(p.ctx.WorkDir, p.previewFile)

    return func() tea.Msg {
        sess, paneID, err := tty.CreateSession(sessionName, []string{editor, fullPath}, p.ctx.WorkDir)
        if err != nil { return EditErrorMsg{Err: err} }
        return EditStartedMsg{Session: sess, PaneID: paneID}
    }
}

// In Update():
case EditStartedMsg:
    p.editingFile = p.previewFile
    width, height := p.previewDimensions()
    return p, p.ttyModel.Start(msg.Session, msg.PaneID, width, height)

case tty.TTYExitRequestedMsg, tty.TTYSessionDeadMsg:
    p.ttyModel.Stop()
    p.editingFile = ""
    return p, p.refreshPreview()
```

## Verification

1. **Unit tests**: Key mapping, cursor rendering, polling decay
2. **Manual testing**:
   - Workspace: `i` to enter interactive mode, type commands, `Ctrl+\` to exit
   - Files: `e` on a file, vim opens in preview, `:wq` returns to preview
3. **Edge cases**:
   - Editor not found (fallback to vim)
   - Session death during edit
   - Window resize while editing
   - Double-escape exit

## Decisions

- **Key binding**: Replace `e` in files plugin (inline editing becomes default)
- **Scope**: Files plugin only (git plugin can be added later)
- **Feature flag**: New `tmux_inline_edit` flag (separate from workspace's `tmux_interactive_input`)

## Feature Flag

```json
{
  "features": {
    "tmux_inline_edit": true
  }
}
```

Add to `internal/features/features.go`:
```go
var TmuxInlineEdit = Feature{
    Name:        "tmux_inline_edit",
    Description: "Enable inline editing with CLI editors in file browser preview pane",
}
```

## Critical Files Reference

- `internal/plugins/workspace/interactive.go` - Source for extraction (~900 lines)
- `internal/plugins/workspace/types.go:231-292` - `InteractiveState` struct
- `internal/plugins/filebrowser/operations.go:22-40` - Current editor opening
- `internal/app/update.go:200-211` - `OpenFileMsg` handling pattern

## Tasks (td-50cf27)

| Task ID | Description |
|---------|-------------|
| td-d0fdfd | Create internal/tty/keymap.go with MapKeyToTmux function |
| td-c2e874 | Create internal/tty/cursor.go with cursor rendering functions |
| td-6c477d | Create internal/tty/terminal_mode.go with mode detection |
| td-22e317 | Create internal/tty/paste.go with clipboard paste handling |
| td-3c54c4 | Create internal/tty/output_buffer.go with hash-based change detection |
| td-90ba29 | Create internal/tty/session.go with tmux session lifecycle helpers |
| td-708d20 | Create internal/tty/polling.go with adaptive polling logic |
| td-576411 | Create internal/tty/messages.go with tea.Msg types |
| td-e6d11e | Create internal/tty/tty.go with core Model struct and API |
| td-f5e761 | Add tmux_inline_edit feature flag |
| td-8fbd23 | Update workspace plugin to use tty package functions |
| td-3636ff | Integrate tty.Model into files plugin for inline editing |
| td-ad3891 | Add unit tests for tty package |
