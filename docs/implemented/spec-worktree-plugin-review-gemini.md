# Review of Worktree Manager Plugin Specification

**Date:** January 13, 2026
**Reviewer:** Gemini CLI
**Target:** `@docs/spec-worktree-plugin.md`

## Executive Summary

The specification is robust and well-aligned with the problem statement. The architecture generally fits the Sidecar plugin model, but there are **critical concurrency violations** in the proposed implementation details regarding Bubble Tea state management.

Additionally, the proposed `td` integration strategy (modifying `td` source code) might be unnecessary if environment variables can be leveraged, offering a cleaner integration path.

---

## 1. Concurrency & State Management (Critical)

**Issue:** The spec proposes a polling loop (`pollAgentOutput`) that directly mutates the `Worktree` and `Agent` structs inside a goroutine (`wt.Agent.OutputBuf.Update(output)`).

**Violation:** In Bubble Tea, **only** the `Update` function can mutate the model. Background goroutines must return `tea.Msg`s. Direct mutation causes race conditions and UI rendering glitches (panics) because the View reads these structs while the goroutine writes to them.

**Correction Required:**
Refactor the polling mechanism to use the standard Bubble Tea command loop pattern.

*   **Change in `Update`:**
    ```go
    // Bad (Current Spec):
    // go m.pollAgentOutput(wt)

    // Good (Bubble Tea Pattern):
    func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
        case AgentOutputMsg:
            // 1. Update state here
            m.agents[msg.WorktreeID].OutputBuf.Write(msg.Data)
            // 2. Schedule next poll
            return m, waitForNextOutput(msg.WorktreeID)
        }
    }

    func waitForNextOutput(id string) tea.Cmd {
        return func() tea.Msg {
            // Sleep and Fetch logic here
            return AgentOutputMsg{WorktreeID: id, Data: ...}
        }
    }
    ```

## 2. TD Integration Strategy

**Issue:** The spec proposes modifying `td` source code to read a `.td-root` file. While valid, this creates a "hard" dependency version requirement between `sidecar` and `td`.

**Suggestion:**
Check if `td` supports an environment variable to override the database path (e.g., `TD_DB_PATH` or `TD_CONFIG_DIR`).

1.  **If `td` supports Env Vars:**
    Instead of writing a `.td-root` file, the Worktree Plugin should simply inject this environment variable when spawning the agent or running `td` commands in that worktree.
    *   *Implementation:* In `getAgentCommand`, prepend `export TD_DB_PATH=/path/to/main/.todos; ...`

2.  **If `td` does NOT support Env Vars:**
    The `.td-root` file approach is acceptable, but consider adding the environment variable support to `td` instead of a custom file logic. It is a more standard CLI pattern.

## 3. Architecture & Code Reuse

**Omission:** The spec defines a new `TDClient`.
**Context:** `sidecar` already has `internal/plugins/tdmonitor`.

**Recommendation:**
*   Check `internal/plugins/tdmonitor` for an existing `td` client wrapper.
*   Refactor the `td` client logic into a shared package (e.g., `internal/pkg/tdclient`) so both the `tdmonitor` plugin and the `worktree` plugin reuse the same code for parsing `td` JSON output and handling IPC.

## 4. Tmux Integration & UX

**Issue:** `AttachToSession` uses `tea.ExecProcess`.
**Refinement:**
*   **Polling Pause:** When the user attaches to a session, the `Update` loop should **pause** polling for that specific agent. Polling while the user is interactively using the tmux session adds unnecessary CPU load and potential I/O contention.
*   **Resume:** When `tea.ExecProcess` returns (user detaches), immediately trigger a refresh/poll command to capture what happened while the user was away.

**Safety:**
The spec mentions `CleanupOrphanedSessions`. This should be extremely conservative.
*   *Risk:* If a user names their own session `sidecar-wt-manual-test`, the plugin might kill it.
*   *Fix:* Only kill sessions that explicitly match IDs known to the current Sidecar instance's state, or use a specific tmux variable/tag to strictly identify ownership.

## 5. Keymap & Plugin Guide Compliance

The spec follows the keymap guide well, but a few specific details need alignment with `internal/keymap`:

1.  **Context Naming:**
    The spec uses "List View" and "Kanban View".
    *   Ensure `FocusContext()` returns stable strings like `worktree-list` and `worktree-kanban` to allow `internal/keymap/bindings.go` to provide specific bindings for each view.

2.  **Footer Hints:**
    The spec shows a footer: `n:new y:approve...`.
    *   *Correction:* In Sidecar, plugins do **not** render their own footer. The plugin must implement `Commands() []plugin.Command`. The main app (`internal/app/model.go`) renders the footer based on these commands. The mockup in the spec should be updated to reflect that the footer is system-managed.

## 6. Directory Structure & Paths

**Observation:** The spec suggests sibling directories (`../sidecar-worktrees`).
**Constraint:** This assumes the user has write permissions to the parent directory.
**Improvement:**
*   Default to sibling directories.
*   Add a validation step in `Init`: If the parent directory is not writable, fallback to a subdirectory within the main repo (e.g., `.sidecar/worktrees/`) and add that path to `.gitignore` automatically.

## 7. Implementation Roadmap Adjustment

Based on the complexity of Bubble Tea concurrency, I suggest a modified Phase 1:

**Phase 1.5: The TUI Loop**
Before integrating Agents/Tmux, build the TUI list view using **mock data** to ensure the `Kanban` <-> `List` toggle and `Focus` handling works smoothly within the `View(w, h)` constraints. Sidecar panes can be narrow; the Kanban view might need a "min-width" check to auto-collapse to List view if the terminal is too small.

## Summary of Necessary Changes to Spec

1.  **Rewrite Section 6.3 (Capturing Output)** to use `tea.Cmd` and `Msg` instead of `go routine` + `mutex`.
2.  **Update Section 8.1 (TD Root)** to prefer Environment Variables over file markers if possible.
3.  **Update Section 4.1 (UI)** to remove the hardcoded footer and reference the `Commands()` interface.
4.  **Add Section 3.4 (Shared Code)** to explicitly mention reusing `td` parsing logic from the `tdmonitor` plugin if available.
5.  **Update Section 5.1 (Location)** to include a fallback if the parent directory is not writable.
