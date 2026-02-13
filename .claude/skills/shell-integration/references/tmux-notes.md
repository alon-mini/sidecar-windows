# Tmux Integration Notes

Detailed reference for Go-based tmux UI integration techniques. This project uses direct tmux CLI commands via `exec.Command`, not a wrapper library.

## Forwarding Input and Capturing Output

Core approach: tmux as headless terminal driver. Use `capture-pane` to poll pane text, `send-keys` to inject input.

Always use `capture-pane` with `-e` flag to preserve ANSI escape codes. Without `-e`, color and formatting is stripped.

Alternative: tmux control mode (`tmux -C`) provides structured `%output` events with pane output and accepts commands. Avoids writing a full VT100 parser since tmux sends already-rendered text.

## Cursor Position Tracking

Tmux exposes `#{cursor_x}` and `#{cursor_y}` (zero-based coordinates). Also `#{cursor_flag}` for visibility.

### The "+1" Offset

`capture-pane -p` may not include trailing empty line where cursor sits. Solutions:
- Use `-J` flag to include blank trailing lines
- Adjust index manually

The `-J` flag behavior depends on `tmux_interactive_input` feature flag. When enabled, `-J` preserves trailing whitespace for accurate cursor positioning.

Additional cursor variables: `cursor_shape` (block/line/underline), `cursor_blinking`.

## Synchronizing Terminal Resizes

Use `tmux resize-pane -x <cols> -y <rows>` when UI resizes. For detached sessions, use `tmux resize-window`.

Failure to sync causes:
- Horizontal padding or truncation (widths differ by 1)
- Height mismatch with scroll region issues
- Background filled with spaces if tmux thinks pane is larger

Always propagate resize events from `WindowSizeMsg`.

## Bracketed Paste Mode

Wrap pasted text: `\x1b[200~` + text + `\x1b[201~`. Tmux flag: `#{bracketed_paste_flag}`.

Key rules:
- Only the outermost terminal (your UI) should add escape wrappers
- Don't double-wrap (if both UI and tmux add wrappers, app sees extra markers)
- Modern shells enable bracketed paste by default, tmux 2.6+ honors it

## Forwarding Mouse Events

For apps inside tmux to receive mouse events, keep tmux's mouse mode OFF and forward encoded sequences.

SGR extended mode (`\x1b[?1000h` + `\x1b[?1006h`): `ESC [ < b;x;y M` (press) or `m` (release).

Coordinates are 1-based in escape sequences. Button codes: 0=left, 1=middle, 2=right, 64=scroll-up, 65=scroll-down. Add 32 for drag.

**Important:** Strip SGR mouse escape sequences from captured output to prevent visual artifacts.

## Modified Keys

Modified keys use xterm CSI sequences:
- Shift+Up: `ESC [ 1;2A`
- Ctrl+Up: `ESC [ 1;5A`
- Alt+key: ESC then the character

Tmux `send-keys` accepts `C-` and `M-` prefixes. For reliable forwarding, send literal escape sequences via `send-keys -l` rather than relying on key name parsing.

Platform quirk: macOS Terminal.app may not send distinct codes for Shift+Arrow without configuration.

Tmux `extended-keys` option (3.2+) enables CSI u encoding for additional modifier support.

## Adaptive Polling

Strategy: poll fast during output bursts, slow down when idle. Exponential backoff from ~16ms (60 FPS) up to ~500ms.

Hash-based skip: compare captured content hash to avoid processing when unchanged. Tmux `#{history_bytes}` can serve as a quick change indicator.

For control mode: `pause-after` can buffer bursts (~100ms) and send in one chunk.

## Batch Capture Optimization

When polling multiple panes, capture all active panes in one coordinated operation and cache results briefly (~300ms). Reduces subprocess spawns.

## Runaway Session Detection

If >20 polls detect output within 3 seconds, throttle to 20s intervals. Reset when session becomes idle.

## Capture Size Limits

Configurable via `plugins.workspace.tmuxCaptureMaxBytes` in config. Default cap: 2MB.

## Debugging Desynchronization

### Logging

- Bubble Tea: instrument Update to dump every message to a file
- tmux: `tmux -vv new` produces logs in `/tmp/tmux-{client,server,out}-*.log`
- Server log shows all input/output to panes and state changes

### Techniques

1. Manual `tmux capture-pane -p` to compare with UI output
2. Log cursor_x,y each frame alongside UI's believed position
3. Use `tmux pipe-pane -O` to log all output bytes to a pane
4. Run a real tmux client alongside for visual comparison
5. Snapshot trigger: capture pane to file + dump UI buffer, then diff

### Common Issues

- Missing output: check tmux-out log for delivered bytes vs UI display
- Stale cursor: rapid cursor movements between polls
- Timing: poll interval missed output chunk (solution: extra poll on input events)
