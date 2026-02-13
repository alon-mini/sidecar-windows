---
name: shell-integration
description: >
  Interactive shell/TTY integration with tmux session management, shell command
  execution, and output capture in sidecar. Covers the tty package, key mapping,
  adaptive polling, cursor rendering, scrolling, paste handling, and inline editing.
  Use when working on shell integration, tmux features, command execution, or
  interactive mode.
user-invocable: false
---

# Shell Integration

Sidecar's interactive shell allows users to type directly into tmux sessions from within the TUI. It is NOT a terminal emulator -- tmux is the PTY backend, sidecar acts as an input/output relay.

## Package Structure

```
internal/tty/                    # Shared tmux terminal abstraction
  tty.go                         # Core Model and State types
  keymap.go                      # Bubble Tea -> tmux key translation
  messages.go                    # Message types (CaptureResultMsg, PollTickMsg, etc.)
  session.go                     # tmux operations (send-keys, capture-pane, resize)
  polling.go                     # Polling interval constants and calculation
  cursor.go                      # Cursor rendering and position query
  paste.go                       # Paste handling (clipboard, bracketed paste)
  terminal_mode.go               # Terminal mode detection (mouse, bracketed paste)
  output_buffer.go               # Thread-safe buffer with hash-based change detection

internal/plugins/workspace/
  interactive.go                 # Workspace-specific interactive mode logic
  interactive_selection.go       # Text selection in interactive mode
  view_preview.go                # Rendering with cursor overlay and scroll offset
  mouse.go                       # Scroll handling
  types.go                       # InteractiveState type

internal/plugins/filebrowser/
  inline_edit.go                 # Inline editor mode using tty.Model
  handlers.go                    # Message handling for inline edit
```

## Data Flow

```
User Keypress -> handleInteractiveKeys()
              -> tty.MapKeyToTmux()
              -> tmux send-keys
              -> schedulePoll(20ms debounce)
              -> capture-pane + cursor query
              -> CaptureResultMsg
              -> OutputBuffer.Update()
              -> pollInteractivePane() (adaptive 50-250ms)
              -> renderWithCursor()
```

## Core Abstractions

### tty.Model

Embeddable component for interactive tmux functionality:

```go
type Model struct {
    Config   Config        // Exit key, copy/paste keys, scrollback lines
    State    *State        // Current interactive state
    Width    int
    Height   int
    OnExit   func() tea.Cmd
    OnAttach func() tea.Cmd
}

// Usage:
p.inlineEditor = tty.New(&tty.Config{
    ExitKey: "ctrl+\\",
    ScrollbackLines: 600,
})
cmd := p.inlineEditor.Enter(sessionName, paneID)
```

### tty.State

```go
type State struct {
    Active        bool
    TargetPane    string      // tmux pane ID (e.g., "%12")
    TargetSession string
    LastKeyTime   time.Time   // For polling decay
    CursorRow, CursorCol int
    CursorVisible        bool
    PaneHeight, PaneWidth int
    BracketedPasteEnabled bool
    MouseReportingEnabled bool
    OutputBuf      *OutputBuffer
    PollGeneration int          // For invalidating stale polls
}
```

### tty.OutputBuffer

Thread-safe bounded buffer with hash-based change detection:

```go
func (b *OutputBuffer) Update(content string) bool {
    rawHash := maphash.String(seed, content)
    if rawHash == b.lastRawHash { return false }  // Skip ALL processing
    content = mouseEscapeRegex.ReplaceAllString(content, "")
    b.lines = strings.Split(content, "\n")
    return true
}
func (b *OutputBuffer) LinesRange(start, end int) []string
```

## Key Mapping (`keymap.go`)

```go
func MapKeyToTmux(msg tea.KeyMsg) (key string, useLiteral bool) {
    switch msg.Type {
    case tea.KeyEnter:     return "Enter", false
    case tea.KeyBackspace: return "BSpace", false
    case tea.KeyTab:       return "Tab", false
    case tea.KeyUp:        return "Up", false
    case tea.KeyCtrlC:     return "C-c", false
    case tea.KeyRunes:     return string(msg.Runes), true  // Literal mode
    }
}
```

Modified keys use CSI sequences:
```go
case "shift+up":   return "\x1b[1;2A", true
case "ctrl+up":    return "\x1b[1;5A", true
case "alt+up":     return "\x1b[1;3A", true
case "shift+tab":  return "\x1b[Z", true
```

For printable characters, `tmux send-keys -l` prevents interpretation.

## Adaptive Polling (`polling.go`)

```go
const (
    PollingDecayFast   = 50ms    // During active typing
    PollingDecayMedium = 200ms   // After 2s inactivity
    PollingDecaySlow   = 250ms   // After 10s inactivity
    KeystrokeDebounce  = 20ms    // Delay after keystroke
)
```

### Three-State Visibility Polling (Workspace)

| State | Active | Idle |
|-------|--------|------|
| Visible + focused | 200ms | 2s |
| Visible + unfocused | 500ms | 500ms |
| Not visible | 10-20s | 10-20s |

### Poll Generation

Stale polls invalidated using generation counter:

```go
func (m *Model) schedulePoll(delay time.Duration) tea.Cmd {
    m.State.PollGeneration++
    gen := m.State.PollGeneration
    return tea.Tick(delay, func(t time.Time) tea.Msg {
        return PollTickMsg{Generation: gen}
    })
}
```

### Performance Per Keystroke

1. `tmux send-keys` (~10ms)
2. 20ms debounce
3. `capture-pane` (~5ms) + cursor query (~5ms)
4. Hash check (~1ms), regex if changed (~5ms), buffer split (~1ms)
5. Cursor overlay (<1ms)

Total: ~42ms worst case, ~36ms typical.

## Cursor Positioning (`cursor.go`)

### Query

```go
func QueryCursorPositionSync(target string) (row, col, paneHeight, paneWidth int, visible, ok bool) {
    cmd := exec.Command("tmux", "display-message", "-t", target,
        "-p", "#{cursor_x},#{cursor_y},#{cursor_flag},#{pane_height},#{pane_width}")
}
```

### Rendering

Cursor is rendered as a block character overlaid on captured output. Handles cursor past end of line (pad with spaces) and cursor within line (ANSI-aware slicing with `ansi.Cut`).

### Height Mismatch Adjustment

When display height differs from tmux pane height:
```go
if paneHeight > displayHeight {
    relativeRow = cursorRow - (paneHeight - displayHeight)
} else if paneHeight > 0 && paneHeight < displayHeight {
    relativeRow = cursorRow + (displayHeight - paneHeight)
}
```

## Scrolling

Scrolling operates on the captured buffer. No tmux copy-mode involved.

```go
type Plugin struct {
    previewOffset    int   // Lines from bottom (0 = at bottom/live)
    autoScrollOutput bool  // Auto-follow new output?
}
```

- Scroll UP: pause auto-scroll, increment `previewOffset`
- Scroll DOWN: decrement `previewOffset`, re-enable auto-scroll at 0
- Bounded by capture window (default 600 lines)
- Instant response (pure state manipulation, no subprocess calls)

## Copy/Paste

- Copy: `alt+c` (configurable via `interactiveCopyKey`)
- Paste: `alt+v` (configurable via `interactivePasteKey`)

Paste wraps text with bracketed paste sequences (`\x1b[200~`...`\x1b[201~`) when the application has enabled bracketed paste mode.

## Terminal Mode Detection (`terminal_mode.go`)

Detects bracketed paste and mouse reporting modes by scanning output for enable/disable escape sequences. Latest enable > latest disable = mode active.

## Width Synchronization

Tmux panes are resized in background at all times (not just interactive mode):

```go
func ResizeTmuxPane(paneID string, width, height int) {
    // resize-window, fallback to resize-pane for older tmux
}
```

Resize triggers: window resize, sidebar toggle/drag, selection change, agent/shell creation, interactive mode entry.

## Inline Edit Mode (Filebrowser)

Uses `tty.Model` for vim/nano/emacs editing in the file preview pane:

```go
func (p *Plugin) enterInlineEditMode(path string) tea.Cmd {
    editor := os.Getenv("EDITOR")
    sessionName := fmt.Sprintf("sidecar-edit-%d", time.Now().UnixNano())
    exec.Command("tmux", "new-session", "-d", "-s", sessionName, editor, path).Run()
    return InlineEditStartedMsg{SessionName: sessionName}
}
```

## Entry and Exit

**Workspace Plugin:**
- Enter: `i` when preview pane focused with output tab
- Exit: `Ctrl+\` (instant) or double-Escape (150ms delay)
- Attach: `Ctrl+]` (full tmux attach)

**Filebrowser Plugin:**
- Enter: `e` or `Enter` on a file (if inline edit enabled)
- Exit: `Ctrl+\` or double-Escape
- Attach: `Ctrl+]`

## Feature Flags

```json
{
  "features": {
    "tmux_interactive_input": true,
    "tmux_inline_edit": true
  }
}
```

## Configuration

```json
{
  "plugins": {
    "workspace": {
      "interactiveExitKey": "ctrl+\\",
      "interactiveAttachKey": "ctrl+]",
      "interactiveCopyKey": "alt+c",
      "interactivePasteKey": "alt+v",
      "tmuxCaptureMaxBytes": 600
    }
  }
}
```

## Critical Rules

1. **Never clear OutputBuffer** -- breaks hash-based change detection and rendering
2. **Always increment poll generation** on entering interactive mode to avoid duplicate poll chains (causes 200% CPU)
3. **Never call subprocesses from View()** -- cursor queries and tmux ops must run in poll handlers
4. **Don't mix shell/workspace polling** -- shells use `scheduleShellPollByName()` + `shellPollGeneration`, workspaces use `scheduleAgentPoll()` + `pollGeneration`
5. **Hash before regex** -- massive CPU savings when content unchanged
6. **Debouncing works** -- 20ms delay reduces subprocess spam ~60%
7. **Atomic cursor capture** -- query cursor with output to avoid race conditions
8. **Width sync matters** -- resize panes in background at all times

## References

- [Tmux integration notes](references/tmux-notes.md) -- detailed tmux CLI techniques, cursor tracking, resize sync, bracketed paste, mouse forwarding, modified keys, adaptive polling, debugging
- Original spec: `docs/spec-tmux-interactive-input.md`
