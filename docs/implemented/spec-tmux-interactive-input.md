# Tmux Interactive Input (Write Support) for Sidecar

## Overview

Add "write support" to Sidecar's tmux pane preview, enabling users to type into the tmux pane while keeping the Sidecar UI visible. The approach: intercept Bubble Tea keypresses and forward them to tmux via `send-keys`, while continuing to capture output for display.

**Core Insight**: This is a proxy/forwarding approach, not a terminal emulator. Tmux remains the PTY backend; Sidecar becomes an input/output relay.

---

## A) Focus & Escape UX

### Entering Interactive Mode

**Trigger**: `i` key when preview pane is focused (vim-style "insert")

- Doesn't conflict with existing `Enter` (which attaches/suspends UI)
- Clear mental model: `i` = inline input, `Enter` = full attach

```go
// In handleListKeys when activePane == PanePreview
case "i":
    if wt := p.selectedWorktree(); wt != nil && wt.Agent != nil {
        return p.startInteractiveMode(wt)
    }
    if shell := p.getSelectedShell(); shell != nil {
        return p.startInteractiveModeShell(shell)
    }
```

### Exiting Interactive Mode

**Problem**: Must work inside vim/fzf/htop which consume Escape.

**Design Decision: `Ctrl+\` as Primary Exit**

The original double-Escape approach has a flaw: if we forward single Escape immediately, we can't detect double-Escape without swallowing input. The alternatives:

| Approach                                   | Tradeoff                                                |
| ------------------------------------------ | ------------------------------------------------------- |
| Forward Esc immediately, detect double-Esc | Can't detect—second Esc arrives after first is sent     |
| Delay Esc ~200ms to detect double          | Adds latency to ALL Esc usage in vim/fzf (unacceptable) |
| `Ctrl+\` as primary exit                   | Unambiguous, no delay, rarely used in apps              |

**Chosen approach**: `Ctrl+\` is the **primary** exit mechanism. Double-Escape is a **secondary** convenience that uses a short delay.

```go
const escapeDetectDelay = 150 * time.Millisecond

type interactiveState struct {
    targetPaneID  string    // tmux pane ID (e.g., "%12")
    isShell       bool
    pendingEscape bool      // True if we're waiting to see if another Esc arrives
    escapeTimer   *time.Timer
}

func (p *Plugin) handleInteractiveKeys(msg tea.KeyMsg) tea.Cmd {
    // Primary exit: Ctrl+\ — immediate, unambiguous
    if msg.String() == "ctrl+\\" {
        return p.exitInteractiveMode()
    }

    // Secondary exit: Double-Escape with delay
    if msg.Type == tea.KeyEscape {
        if p.interactive.pendingEscape {
            // Second Escape within window: exit
            if p.interactive.escapeTimer != nil {
                p.interactive.escapeTimer.Stop()
            }
            p.interactive.pendingEscape = false
            return p.exitInteractiveMode()
        }
        // First Escape: start delay window
        p.interactive.pendingEscape = true
        p.interactive.escapeTimer = time.AfterFunc(escapeDetectDelay, func() {
            // Timer fired: forward the single Escape
            p.interactive.pendingEscape = false
            // Send delayed Escape to tmux (via channel/message)
        })
        return nil // Don't forward yet
    }

    // Any other key: cancel pending Escape, forward it, then forward this key
    if p.interactive.pendingEscape {
        p.interactive.pendingEscape = false
        if p.interactive.escapeTimer != nil {
            p.interactive.escapeTimer.Stop()
        }
        // Forward the pending Escape + current key
        return tea.Batch(
            p.sendKeyToTmux("Escape"),
            p.forwardKey(msg),
        )
    }

    return p.forwardKey(msg)
}
```

**Tradeoff acknowledged**: The 150ms delay on single Escape adds slight latency for vim mode switches. This is acceptable because:

1. `Ctrl+\` provides instant exit when needed
2. 150ms is at the edge of human perception
3. Most vim users use `Ctrl+[` for Escape anyway

---

## B) Key Event Mapping

### Bubble Tea → tmux send-keys Translation

| Bubble Tea                  | tmux send-keys             | Notes                           |
| --------------------------- | -------------------------- | ------------------------------- |
| `tea.KeyRunes`              | `-l "chars"`               | Literal mode for printable text |
| `tea.KeyEnter`              | `Enter`                    |                                 |
| `tea.KeySpace`              | `Space`                    |                                 |
| `tea.KeyTab`                | `Tab`                      |                                 |
| `tea.KeyBackspace`          | `BSpace`                   |                                 |
| `tea.KeyDelete`             | `DC`                       |                                 |
| `tea.KeyEsc`                | `Escape`                   |                                 |
| `tea.KeyUp/Down/Left/Right` | `Up`/`Down`/`Left`/`Right` |                                 |
| `tea.KeyHome/End`           | `Home`/`End`               |                                 |
| `tea.KeyPgUp/PgDown`        | `PPage`/`NPage`            |                                 |
| `tea.KeyCtrlA..Z`           | `C-a`..`C-z`               |                                 |
| Alt+x                       | `M-x`                      | msg.Alt flag                    |
| F1-F12                      | `F1`-`F12`                 |                                 |

### Key Mapping Realism

**Important limitation**: Modified special keys (Shift+Arrow, Ctrl+Arrow, etc.) are **unreliable** via tmux `send-keys` named keys. Terminal emulators encode these differently, and tmux's key name translation doesn't always match.

**MVP approach**:

- Prioritize common unmodified keys (arrows, function keys, Ctrl+letter)
- Rely on tmux's internal translation for what it supports
- Defer advanced modifier combinations (Shift+Arrow, Ctrl+Arrow) to post-MVP
- Document known limitations

```go
// keymap_tmux.go
func MapKeyToTmux(msg tea.KeyMsg) (keys []string, literal bool, supported bool) {
    switch msg.Type {
    case tea.KeyUp:
        if msg.Shift || msg.Ctrl {
            // Modified arrows are unreliable - skip for MVP
            return nil, false, false
        }
        return []string{"Up"}, false, true
    // ... etc
    }
}
```

### Application Cursor Mode

Terminal apps like vim use "application cursor keys" mode. We can detect this from captured ANSI:

- `ESC[?1h` = enable application cursor
- `ESC[?1l` = disable

**For MVP**: Skip detection. Modern tmux/terminals handle mode translation automatically. Only add detection if we observe issues.

---

## C) Sending Input to tmux

### Pane ID Targeting (Not Session Names)

**Critical change**: Use tmux **pane IDs** (`%12`) as the canonical target, not session names or window indices.

**Why pane IDs**:

- Globally unique within tmux server
- Stable across window/session renames
- Unambiguous targeting for multi-pane sessions

**Discovering pane IDs**:

```go
// When creating a session, capture the pane ID
func createTmuxSession(name, workDir string) (paneID string, err error) {
    // Create session and get pane ID in one command
    cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", workDir,
        "-P", "-F", "#{pane_id}")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil // Returns "%12"
}

// For existing sessions, query pane ID
func getPaneID(sessionName string) (string, error) {
    cmd := exec.Command("tmux", "list-panes", "-t", sessionName,
        "-F", "#{pane_id}")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    // Returns first pane (for single-pane sessions)
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    if len(lines) == 0 {
        return "", errors.New("no panes found")
    }
    return lines[0], nil
}
```

**Store pane ID in Agent/Shell structs**:

```go
type Agent struct {
    TmuxSession string // Session name (for display/attach)
    TmuxPaneID  string // Pane ID for send-keys/capture (e.g., "%12")
    // ...
}
```

### Basic Send Pattern

```go
func (p *Plugin) sendKeyToTmux(key string) tea.Cmd {
    return func() tea.Msg {
        paneID := p.interactive.targetPaneID
        cmd := exec.Command("tmux", "send-keys", "-t", paneID, key)
        if err := cmd.Run(); err != nil {
            return InteractiveErrorMsg{Err: err}
        }
        return InteractiveKeySentMsg{}
    }
}

func (p *Plugin) sendLiteralToTmux(text string) tea.Cmd {
    return func() tea.Msg {
        paneID := p.interactive.targetPaneID
        cmd := exec.Command("tmux", "send-keys", "-l", "-t", paneID, text)
        if err := cmd.Run(); err != nil {
            return InteractiveErrorMsg{Err: err}
        }
        return InteractiveKeySentMsg{}
    }
}
```

### Paste Strategy: tmux Native Buffers

**For MVP**: Do NOT use bracketed paste. It requires tracking paste mode state in the target application, which we can't reliably detect.

**Instead, use tmux's native buffer mechanism**:

```go
func (p *Plugin) sendPasteToTmux(text string) tea.Cmd {
    return func() tea.Msg {
        paneID := p.interactive.targetPaneID

        // Load text into tmux buffer
        loadCmd := exec.Command("tmux", "load-buffer", "-")
        loadCmd.Stdin = strings.NewReader(text)
        if err := loadCmd.Run(); err != nil {
            return InteractiveErrorMsg{Err: fmt.Errorf("load-buffer: %w", err)}
        }

        // Paste buffer into target pane
        pasteCmd := exec.Command("tmux", "paste-buffer", "-t", paneID)
        if err := pasteCmd.Run(); err != nil {
            return InteractiveErrorMsg{Err: fmt.Errorf("paste-buffer: %w", err)}
        }

        return InteractiveKeySentMsg{}
    }
}
```

**Why this is better**:

- Works regardless of application's paste mode
- tmux handles newline translation
- No escape sequence injection risks
- Atomic operation

**Later enhancement (post-MVP)**: Bracketed paste support requires:

1. Detecting if target app enabled bracketed paste mode (via captured ANSI)
2. Wrapping content in `ESC[200~` ... `ESC[201~`
3. Tracking mode changes across captures

---

## D) Rendering/Capture Improvements

### Capture Flag Requirements

**Current state check needed**: Verify if `capture-pane -e` is already in use. The `-e` flag is **required** to preserve ANSI escape sequences (colors, cursor movement) for proper rendering.

```bash
# Required capture command
tmux capture-pane -p -e -t %12
```

If currently using `-p` without `-e`, output will be plain text and full-screen apps will render incorrectly.

### Cursor Overlay (MVP Requirement)

**Critical for usability**: Most full-screen apps (vim, fzf, htop) do NOT render the cursor as a visible glyph in their output. They rely on the terminal's cursor. Since we're capturing text, not a live terminal, we must render our own cursor.

**MVP must include cursor position query and overlay**:

```go
type cursorState struct {
    x, y    int
    visible bool
}

func (p *Plugin) getCursorPosition() (cursorState, error) {
    paneID := p.interactive.targetPaneID

    // Query cursor position and visibility
    cmd := exec.Command("tmux", "display-message", "-t", paneID,
        "-p", "#{cursor_x},#{cursor_y},#{cursor_flag}")
    output, err := cmd.Output()
    if err != nil {
        return cursorState{}, err
    }

    parts := strings.Split(strings.TrimSpace(string(output)), ",")
    if len(parts) < 2 {
        return cursorState{}, errors.New("invalid cursor format")
    }

    x, _ := strconv.Atoi(parts[0])
    y, _ := strconv.Atoi(parts[1])
    visible := len(parts) < 3 || parts[2] != "0"

    return cursorState{x: x, y: y, visible: visible}, nil
}
```

**Rendering cursor overlay**:

```go
func (p *Plugin) renderWithCursor(content string, cursor cursorState, width int) string {
    if !cursor.visible {
        return content
    }

    lines := strings.Split(content, "\n")
    if cursor.y >= len(lines) {
        return content
    }

    line := lines[cursor.y]
    runes := []rune(line)

    if cursor.x >= len(runes) {
        // Cursor past end of line: append cursor block
        lines[cursor.y] = line + cursorStyle.Render(" ")
    } else {
        // Overlay cursor on existing character
        before := string(runes[:cursor.x])
        char := string(runes[cursor.x])
        after := string(runes[cursor.x+1:])
        lines[cursor.y] = before + cursorStyle.Render(char) + after
    }

    return strings.Join(lines, "\n")
}

var cursorStyle = lipgloss.NewStyle().Reverse(true)
```

### Polling Strategy with Decay

**Problem**: Constant fast polling (50ms) wastes CPU when user isn't actively typing.

**Solution**: Fast polling after input, decay to slower rate during inactivity.

```go
const (
    pollIntervalFast     = 50 * time.Millisecond   // During active input
    pollIntervalMedium   = 200 * time.Millisecond  // Recent activity
    pollIntervalSlow     = 500 * time.Millisecond  // Idle
    activityDecayTime    = 500 * time.Millisecond  // Time before slowing down
)

type interactiveState struct {
    // ...
    lastInputTime time.Time
}

func (p *Plugin) getInteractivePollInterval() time.Duration {
    if p.interactive == nil {
        return pollIntervalSlow
    }

    elapsed := time.Since(p.interactive.lastInputTime)

    switch {
    case elapsed < activityDecayTime:
        return pollIntervalFast   // Active typing: 50ms
    case elapsed < 2*activityDecayTime:
        return pollIntervalMedium // Recent activity: 200ms
    default:
        return pollIntervalSlow   // Idle: 500ms
    }
}

func (p *Plugin) handleInteractiveKeys(msg tea.KeyMsg) tea.Cmd {
    p.interactive.lastInputTime = time.Now() // Reset activity timer
    // ... key handling
}
```

**CPU impact considerations**:

- 50ms polling = 20 calls/sec during active typing (acceptable burst)
- 500ms polling = 2 calls/sec when idle (minimal overhead)
- Each `capture-pane` call: ~5-10ms subprocess overhead
- Combined with cursor query: ~10-15ms per poll cycle

---

## E) Bubble Tea Integration

### New State

```go
// types.go
const ViewModeInteractive ViewMode = 10 // After existing modes

type interactiveState struct {
    targetPaneID  string      // tmux pane ID (e.g., "%12")
    targetSession string      // Session name (for display)
    isShell       bool
    lastInputTime time.Time   // For polling decay
    pendingEscape bool        // For double-escape detection
    escapeTimer   *time.Timer
}

// plugin.go - add to Plugin struct
interactive *interactiveState
```

### FocusContext

```go
// commands.go
func (p *Plugin) FocusContext() string {
    switch p.viewMode {
    case ViewModeInteractive:
        return "worktree-interactive"
    // ... existing
    }
}
```

### Entry/Exit Functions

```go
// agent.go
func (p *Plugin) startInteractiveMode(wt *Worktree) tea.Cmd {
    if wt.Agent == nil || wt.Agent.TmuxPaneID == "" {
        return nil
    }

    p.interactive = &interactiveState{
        targetPaneID:  wt.Agent.TmuxPaneID,
        targetSession: wt.Agent.TmuxSession,
        isShell:       false,
        lastInputTime: time.Now(),
    }
    p.viewMode = ViewModeInteractive

    // Start with fast polling
    return p.scheduleInteractivePoll()
}

func (p *Plugin) exitInteractiveMode() tea.Cmd {
    if p.interactive == nil {
        return nil
    }

    if p.interactive.escapeTimer != nil {
        p.interactive.escapeTimer.Stop()
    }

    session := p.interactive.targetSession
    p.interactive = nil
    p.viewMode = ViewModeList
    p.activePane = PanePreview

    return p.scheduleAgentPoll(session, pollIntervalActive)
}

func (p *Plugin) scheduleInteractivePoll() tea.Cmd {
    interval := p.getInteractivePollInterval()
    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return interactivePollMsg{}
    })
}
```

### Visual Indicator

```go
// view_preview.go
func (p *Plugin) renderOutputContent(width, height int) string {
    var hint string
    if p.viewMode == ViewModeInteractive {
        hint = styles.InteractiveHint.Render("INTERACTIVE • Ctrl+\\ to exit")
    } else {
        hint = dimText("i to type • enter to attach")
    }
    // ... rest of render with cursor overlay
}
```

---

## F) Milestone Path

### MVP (M1): Basic Interactive Typing

**Goal**: Type printable chars, Enter, Backspace. See feedback with cursor in preview.

**Files to modify**:

- `types.go`: Add ViewModeInteractive, interactiveState
- `keys.go`: Add handleInteractiveKeys(), entry via `i` key
- `keymap_tmux.go`: NEW file with MapKeyToTmux()
- `agent.go`: Add sendKeyToTmux(), pane ID discovery, cursor query
- `commands.go`: Add FocusContext case
- `view_preview.go`: Add cursor overlay rendering, interactive indicator

**Tasks**:

1. Add ViewModeInteractive constant and interactiveState struct
2. Modify session creation to capture and store pane ID
3. Verify/add `-e` flag to capture-pane calls
4. Implement MapKeyToTmux() for: Runes, Enter, Backspace, Escape, arrows
5. Implement sendKeyToTmux() and sendLiteralToTmux() using pane ID
6. Add `i` key handler in handleListKeys (preview pane)
7. Implement handleInteractiveKeys() with `Ctrl+\` exit
8. Implement getCursorPosition() and renderWithCursor()
9. Add FocusContext case for "worktree-interactive"
10. Implement polling decay (fast → slow based on activity)
11. Update hint text in renderOutputContent

**Verification**:

- Start agent, focus preview, press `i`
- Verify cursor is visible at correct position
- Type "hello", press Enter
- Verify text appears and cursor moves
- Press `Ctrl+\` to exit

### Better (M2): Full Key Support & Escape Handling

**Goal**: All common keys work, double-escape as secondary exit.

**Tasks**:

1. Complete MapKeyToTmux() for Ctrl+_, Alt+_, F1-F12
2. Implement double-escape detection with 150ms delay
3. Implement paste via load-buffer + paste-buffer
4. Update Commands() to show interactive mode in footer
5. Document known key mapping limitations

**Verification**:

- Open vim in tmux, press `i`, type and navigate with arrows
- Use Ctrl+C to cancel
- Double-Escape to exit (with noted latency)
- Paste multi-line text

### Robust (M3): Production Quality

**Goal**: Handle edge cases, optimize, polish.

**Tasks**:

1. Graceful session/pane disconnect handling
2. Detect alternate screen mode, adjust hints
3. Add mouse support (forward mouse events to tmux)
4. Bracketed paste support (track paste mode state)
5. Handle modified special keys (Shift+Arrow, Ctrl+Arrow)
6. Visual feedback on key send errors
7. Configurable keybindings for exit

**Verification**:

- Run htop, fzf, less, vim - all work correctly
- Kill tmux session externally - graceful error handling
- Stress test with rapid typing
- Mouse clicks work in fzf

---

## G) Anti-Goals & Pitfalls

### Do NOT:

1. **Build a terminal emulator** - We proxy to tmux, not replace it
2. **Buffer output client-side** - Tmux handles scrollback
3. **Use bracketed paste in MVP** - Requires paste mode tracking
4. **Target by session name** - Use pane IDs for reliability
5. **Implement Shift/Ctrl+Arrow in MVP** - Unreliable via send-keys
6. **Poll at constant fast rate** - Decay to save CPU when idle

### Watch Out For:

1. **Tmux prefix conflict**: User presses Ctrl+b (default prefix). Solution: Forward ALL keys. Tmux handles its own prefix.

2. **Escape latency**: Double-escape requires delay. Document that `Ctrl+\` is instant.

3. **Missing cursor**: Apps don't render cursor as glyph. Solution: Query position, render overlay.

4. **CPU burn**: Fast polling is expensive. Solution: Decay polling rate after inactivity.

5. **Pane ID not stored**: Session name isn't enough. Solution: Capture pane ID on session creation.

6. **No `-e` flag**: Capture without ANSI codes breaks rendering. Verify flag is present.

---

## Critical Files Summary

| File                                        | Changes                                          |
| ------------------------------------------- | ------------------------------------------------ |
| `internal/plugins/worktree/types.go`        | Add ViewModeInteractive, interactiveState        |
| `internal/plugins/worktree/keys.go`         | handleInteractiveKeys(), entry (`i`) handling    |
| `internal/plugins/worktree/keymap_tmux.go`  | NEW: MapKeyToTmux() function                     |
| `internal/plugins/worktree/agent.go`        | sendKeyToTmux(), pane ID discovery, cursor query |
| `internal/plugins/worktree/commands.go`     | FocusContext case for "worktree-interactive"     |
| `internal/plugins/worktree/view_preview.go` | Cursor overlay, interactive mode indicator       |
| `internal/plugins/worktree/update.go`       | Handle InteractiveKeySentMsg, polling decay      |

---

## tmux Command Reference

```bash
# Create session and get pane ID
tmux new-session -d -s "mysession" -c /path -P -F "#{pane_id}"
# Returns: %12

# Get pane ID from existing session
tmux list-panes -t "mysession" -F "#{pane_id}"

# Send key to pane (use pane ID, not session name)
tmux send-keys -t %12 Enter
tmux send-keys -t %12 C-c
tmux send-keys -t %12 Escape

# Send literal text
tmux send-keys -l -t %12 "hello world"

# Paste via buffer (for multi-line)
echo "multi\nline\ntext" | tmux load-buffer -
tmux paste-buffer -t %12

# Get cursor position
tmux display-message -t %12 -p "#{cursor_x},#{cursor_y}"

# Capture with ANSI escape sequences (required)
tmux capture-pane -p -e -t %12

# Check if pane exists
tmux has-session -t %12 2>/dev/null && echo "exists"
```

---

## Verification Strategy

### Unit Tests

- MapKeyToTmux() coverage for supported key types
- Polling decay timing logic
- Cursor overlay positioning

### Integration Tests

- Pane ID discovery on session creation
- Send key, verify tmux receives it
- Cursor position query accuracy
- Escape detection timing

### Manual E2E Tests

- Basic typing in shell with visible cursor
- vim: insert mode, navigation, :wq
- fzf: fuzzy search, selection (cursor visible)
- htop: navigation, quit
- less: scrolling, search
- Rapid typing (no dropped chars, polling decay works)
- Session kill during interactive mode
- `Ctrl+\` exits instantly
- Double-Escape exits with ~150ms delay
