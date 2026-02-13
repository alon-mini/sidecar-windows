# Form Text Wrapping Investigation

**Issue**: [sidecar#95](https://github.com/marcus/sidecar/issues/95) — Text in the description field doesn't wrap at the modal edge when creating/editing issues. Text wraps at roughly double the visible width (~130 chars instead of ~84).

**Status**: Unresolved. Root cause not identified.

## What Was Done

### Code Changes (currently in working tree)

1. **`pkg/monitor/form.go`** — Added `Width int` field to `FormState`, applied via `Form.WithWidth()` at end of `buildForm()` so width survives form rebuilds (e.g. toggle extended fields).

2. **`pkg/monitor/form_operations.go`** — Both `openNewIssueForm()` and `openEditIssueForm()` now compute `formWidth = modalWidth - 4` (accounting for `Padding(1,2)`) and call `Form.WithWidth(formWidth)` before returning `Form.Init()`.

3. **`pkg/monitor/commands.go`** — `handleFormMessage()` recalculates and reapplies form width on `tea.WindowSizeMsg`.

4. **`pkg/monitor/view.go`** — `renderFormModal()` sets form width at render time and truncates the footer to prevent `JoinVertical` from widening the content beyond `formWidth`.

### Known Unfixed Issue

`CmdFormToggleExtend` handler (commands.go:1216-1220) calls `ToggleExtended()` but returns `nil` cmd instead of `Form.Init()`. The rebuilt form doesn't get re-initialized. Not related to the main wrapping bug but should be fixed.

## What Was Investigated

### Width Chain Trace

Traced the full chain: `Form.WithWidth(n)` → `Group.WithWidth(n)` → `field.WithWidth(n)` → `textarea.SetWidth(n - theme.Base.GetHorizontalFrameSize())`. All values checked and correct. The Dracula theme's `Focused.Base` and `Blurred.Base` both have `GetHorizontalFrameSize() == 2`.

### huh Library Internals (v0.8.0)

- `Form.Update` on `WindowSizeMsg`: only auto-sizes if `f.width == 0`. Since we set width explicitly, this branch is skipped. However, the `WindowSizeMsg` is still forwarded to `Group.Update` → all fields.
- `Group.Update`: forwards all non-key messages to every field, then calls `buildView()` to rebuild the viewport cache.
- `Group.View`: renders from cached viewport content, NOT fresh field views.
- `Text.Update`: falls through to `textarea.Update` for unhandled messages.
- `textarea.Update` (bubbles): does NOT handle `WindowSizeMsg` — won't override width.
- `textarea.SetWidth`: calculates `m.width = inputWidth - reservedOuter - reservedInner`. Wrapping uses `m.width` via `memoizedWrap`.

### Theme / Focus Investigation

- Focus/Blur only changes the style pointer, doesn't reset width.
- Both focused and blurred base styles have identical horizontal frame size (2).

### Isolated Reproduction Attempts

Built standalone test programs using the exact same library versions (pinned via `go.mod` replace directives):

1. **Static form width test**: Created a form with `WithWidth(86)`, rendered it → output was correctly 86 chars wide.
2. **Full lifecycle test**: Created form, called `Init()`, sent `WindowSizeMsg{Width:160}`, typed a long string character by character → output was correctly 86 chars wide.
3. **Full modal rendering test**: Wrapped the form in the same lipgloss modal styling used by `renderFormModal()` → output was correctly 92 chars wide (86 + 4 padding + 2 border).

All three tests showed correct wrapping behavior. The bug could not be reproduced outside the live TUI.

## Hypotheses Not Yet Tested

1. **Message ordering / race condition**: `Form.Init()` sends `tea.WindowSize()` asynchronously. If the resulting `WindowSizeMsg` arrives and is processed by the form before our `WithWidth()` call takes effect in the bubbletea event loop, the form could auto-size to terminal width. However, our code calls `WithWidth()` before returning `Init()` as a cmd, so the width should already be set when the WindowSizeMsg arrives.

2. **Something in the main Update loop resetting form state**: There may be a code path in the main `Update` function that re-creates or replaces the form/group objects without preserving width. A `grep` for `NewForm`, `buildForm`, `WithWidth` could reveal unexpected callers.

3. **The WindowSizeMsg forwarded through Group.Update to fields**: Even though `textarea.Update` doesn't handle `WindowSizeMsg`, other field types or the `updateFieldMsg` sent by Group.Update after every field update might trigger a width recalculation. The `updateFieldMsg` handler in `Text.Update` could be worth examining.

4. **Viewport width in Group**: `Group.WithWidth` sets `g.viewport.Width` and `g.viewport.Style.MaxWidth`. If `buildView()` (called every `Group.Update`) recalculates viewport dimensions from somewhere else, it could override the field widths.

5. **Runtime debug logging**: Writing width values to a file during actual TUI execution (in `handleFormMessage` and `renderFormModal`) would confirm whether the width is correct at runtime or being overridden by something.

## Files Modified

- `pkg/monitor/form.go` — `Width` field + applied in `buildForm()`
- `pkg/monitor/form_operations.go` — Width set on form creation
- `pkg/monitor/commands.go` — Width set on window resize
- `pkg/monitor/view.go` — Width set at render time, footer truncated
