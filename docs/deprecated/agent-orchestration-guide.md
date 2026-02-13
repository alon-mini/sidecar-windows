# Agent Orchestration Guide

Sidecar's agent orchestration automates multi-phase task execution: an AI agent plans, implements, and validates code changes with configurable iteration loops. You select a task, pick a provider, and the orchestrator handles the rest -- spawning agents in isolated workspaces, running validators in parallel, and iterating on rejections until the work passes review.

## Quick start

### Prerequisites

At least one supported agent CLI must be installed and on your `$PATH`:

| Provider | CLI command | Install |
|----------|-----------|---------|
| Claude | `claude` | `npm install -g @anthropic-ai/claude-code` |
| Codex | `codex` | `npm install -g @openai/codex` |
| Gemini | `gemini` | `npm install -g @anthropic-ai/gemini-cli` |

You also need `td` (task engine) installed for task tracking.

### Launch your first orchestration

1. Open sidecar and press the tab key assigned to the orchestrator plugin (check your tab bar).
2. You'll see a list of open tasks from `td`. Use `j`/`k` to navigate, then press `Enter` on a task.
3. The **launch modal** appears. Review the defaults and press `Enter` to run.
4. The **plan review** screen shows the planner's decisions. Press `a` to accept the plan.
5. Watch the **progress view** as the implementer works and validators review.
6. On completion, press `m` to merge the worktree or `Esc` to return.

### Start from an idea (no existing task)

Press `n` from the task list to enter **idea mode**. Type a description of what you want to build (up to 256 characters), then press `Enter`. The orchestrator creates a `td` task automatically and runs the full pipeline.

## Orchestration flow

```
Select task ──> Launch modal ──> Plan ──> Accept? ──> Implement ──> Validate
                                           │                          │
                                           No ── Cancel               │
                                                                 All pass? ── Yes ──> Complete
                                                                      │
                                                                 No + iters left ──> Implement (retry)
                                                                      │
                                                                 No + exhausted ──> Failed
```

Each phase spawns a fresh agent CLI process in the workspace directory. The agent reads task context via `td`, makes changes, and exits. The orchestrator interprets exit codes and td logs to decide what happens next.

See `docs/agent-orchestration-sequence.md` for the full Mermaid sequence diagram.

## The launch modal

The launch modal configures the orchestration run. Open it by pressing `Enter` on a task or `n` for idea mode.

### Provider

A scrollable list of available agent CLIs. Unavailable providers are shown but grayed out. Your last selection is persisted to `~/.config/sidecar/orchestrator-state.json`.

### Template

Press `t` to cycle through orchestration templates:

| Template | Phases | Use case |
|----------|--------|----------|
| **default** | Plan -> Implement -> Validate | Full pipeline with planning and review |
| **simple** | Implement only | Quick changes that don't need planning or validation |
| **review-only** | Validate only | Review existing changes without new implementation |

The modal hides irrelevant options based on the template. For example, "simple" hides iteration and validator controls since those phases are skipped.

### Iterations

Press `Left`/`Right` (or `h`/`l`) to adjust the max iteration count (1-10). Each iteration re-runs the implementer with validator feedback, then re-validates. Higher values give the agent more chances to fix issues.

### Validators

Press `+`/`-` to adjust the validator count (0-5). Validators run in parallel after implementation. Setting to 0 skips validation entirely.

### Validator profiles

Press `p` to cycle all validators through preset focus areas:

| Profile | Focus |
|---------|-------|
| **general** | Broad code review (default) |
| **security** | Input validation, auth, injection vulnerabilities, secrets |
| **tests** | Test quality, coverage, edge cases, isolation, assertions |
| **performance** | Complexity, memory, concurrency, leaks, I/O, scalability |

Each profile appends focused instructions to the validator's prompt.

### Workspace mode

Press `w` to toggle between workspace isolation modes:

| Mode | Behavior |
|------|----------|
| **worktree** | Creates a git worktree on an `agent/{task}` branch. Changes are isolated from your working tree. Merge when ready. |
| **direct** | Works in your current directory. No isolation. Use for quick one-offs. |

### Progressive defaults

When you select a task, the modal loads its metadata and sets smart defaults:

- **Chores or small tasks** (<=3 points): 0 validators, 1 iteration
- **Tasks with acceptance criteria**: 2 validators, 3 iterations
- **Bug fixes**: 1 validator, 2 iterations
- **Everything else**: 1 validator, 2 iterations

You can always override these before launching.

### Capacity

The orchestrator supports up to 3 concurrent runs. If you're at capacity, a warning appears and the Run button is disabled. Finish or cancel a run first.

## Plan review

After the planner agent finishes, the plan review screen shows its decisions. The planner reads the task context via `td` and logs its reasoning.

| Key | Action |
|-----|--------|
| `a` or `Enter` | Accept the plan and start implementation |
| `e` | Open the task in your external editor via `td edit` |
| `r` | Regenerate the plan (re-runs the planner) |
| `j`/`k` | Scroll the decision log |
| `Esc` | Reject the plan and cancel the run |

If the planner produces no updates (empty session), you'll see a warning with the option to retry.

## Progress view

During implementation and validation, the progress view shows:

- **Phase indicator**: Current phase with spinner (PLANNING, IMPLEMENTING, VALIDATING, ITERATING)
- **Iteration counter**: Which attempt you're on (e.g., "iteration 2/5")
- **Modified files**: Files the agent has changed (up to 15 shown)
- **Validator results**: Per-validator approve/reject with finding counts
- **Last activity**: Time since last agent output

| Key | Action |
|-----|--------|
| `c` | Cancel the orchestration |
| `d` | Jump to git-status plugin to view the diff |
| `f` | Jump to file-browser plugin to browse changes |
| `l` | Switch to the run list (all active runs) |

## Completion view

When the run finishes (success or failure):

| Key | Action |
|-----|--------|
| `m` | Merge the worktree branch into your base branch (press twice to confirm) |
| `d` | View the final diff |
| `r` | Retry from the planning phase |
| `Esc` / `q` | Return to task selection |
| `l` | Switch to the run list |

## Managing multiple runs

### Run list

Press `l` from any orchestrator view to see all tracked runs. Each row shows:

- Status icon (spinner = running, checkmark = complete, X = failed)
- Task ID and status label
- Elapsed time

Navigate with `j`/`k` and press `Enter` to switch to a run. Press `Esc` to return.

### Run detail modal

The run detail modal shows a timeline of events for a specific run:

| Key | Action |
|-----|--------|
| `j`/`k` | Scroll timeline |
| `c` | Cancel a running orchestration |
| `d` | View diff |
| `Esc` | Close modal |

### Workspace integration

Orchestration worktrees appear in the workspace plugin with status badges:

- **Planning** -- planner agent is active
- **Implementing** -- implementer is working
- **Complete** -- run finished successfully
- **Failed** -- run ended with errors
- **Interrupted** -- run was cancelled or crashed

Pressing `Enter` on an orchestration worktree in the workspace plugin jumps to the orchestrator's run detail view for that run.

## Configuration

Add to `~/.config/sidecar/config.json` under `plugins.orchestrator`:

```json
{
  "plugins": {
    "orchestrator": {
      "enabled": true,
      "provider": "claude",
      "maxIterations": 3,
      "validatorCount": 2,
      "workspace": "worktree",
      "autoMerge": false,
      "providerBinary": ""
    }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable the orchestrator plugin |
| `provider` | string | `"claude"` | Default agent provider |
| `maxIterations` | int | `3` | Default max rejection-loop iterations (1-10) |
| `validatorCount` | int | `2` | Default number of validators (0-5) |
| `workspace` | string | `"worktree"` | Default workspace mode |
| `autoMerge` | bool | `false` | Auto-merge worktree on successful completion |
| `providerBinary` | string | `""` | Path to agent binary (auto-detected if empty) |

All settings are overridable per-run in the launch modal.

## Keyboard reference

### Task selection (`orchestrator-select`)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate task list |
| `Enter` | Open launch modal |
| `n` | New idea (no task) |
| `/` | Filter tasks |
| `r` | Refresh task list |

### Launch modal (`orchestrator-launch`)

| Key | Action |
|-----|--------|
| `Enter` | Start run |
| `Esc` | Cancel |
| `t` | Cycle template |
| `Left`/`Right` | Adjust iterations |
| `+`/`-` | Adjust validators |
| `p` | Cycle validator profiles |
| `w` | Toggle workspace mode |

### Plan review (`orchestrator-plan`)

| Key | Action |
|-----|--------|
| `a` / `Enter` | Accept plan |
| `e` | Edit task externally |
| `r` | Regenerate plan |
| `j` / `k` | Scroll decisions |
| `Esc` | Reject and cancel |

### Running (`orchestrator-running`)

| Key | Action |
|-----|--------|
| `c` | Cancel run |
| `d` | View diff |
| `f` | Browse files |
| `l` | Run list |

### Complete (`orchestrator-complete`)

| Key | Action |
|-----|--------|
| `m` | Merge worktree |
| `d` | View diff |
| `r` | Retry |
| `Esc` / `q` | Back to tasks |
| `l` | Run list |

### Run list (`orchestrator-runlist`)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate |
| `Enter` | Select run |
| `Esc` | Back |

### Run detail (`orchestrator-detail`)

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll timeline |
| `c` | Cancel |
| `d` | Diff |
| `Esc` | Close |
