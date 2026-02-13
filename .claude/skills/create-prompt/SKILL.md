---
name: create-prompt
description: Create prompts for sidecar workspaces. Covers prompt structure (name, ticketMode, body), template variables (ticket with fallbacks), config file locations (global vs project), and scope overrides. Use when creating or modifying prompts in sidecar config files.
---

# Creating Prompts

Prompts are reusable templates that configure the initial context for agents when creating workspaces. They help standardize common workflows like code reviews, bug fixes, or feature implementation.

## Config File Locations

| Scope   | Path                                  | Override Priority |
|---------|---------------------------------------|-------------------|
| Global  | `~/.config/sidecar/config.json`       | Lower             |
| Project | `.sidecar/config.json` (in project)   | Higher            |

Project prompts override global prompts with the same name (case-sensitive match).

## Config Format

```json
{
  "prompts": [
    {
      "name": "Code Review",
      "ticketMode": "optional",
      "body": "Do a detailed code review of {{ticket || 'open reviews'}}.\nFocus on correctness, edge cases, and test coverage."
    }
  ]
}
```

## Prompt Fields

### name (required)
Display name shown in the prompt picker. Keep it concise.

### ticketMode (optional)
Controls how the ticket/task field behaves:

| Mode       | Behavior                                      |
|------------|-----------------------------------------------|
| `optional` | Task field shown, can be empty (default)      |
| `required` | Task must be selected before creating         |
| `none`     | Task field is hidden, prompt stands alone     |

### body (required)
The prompt text sent to the agent. Supports template variables. Use `\n` for newlines.

## Template Variables

### `{{ticket}}`
Expands to the selected task ID. Returns empty string if no task selected.

```json
"body": "Fix issue {{ticket}}."
// With task td-abc123: "Fix issue td-abc123."
// Without task: "Fix issue ."
```

### `{{ticket || 'fallback'}}`
Expands to task ID, or the fallback text if no task selected.

```json
"body": "Review {{ticket || 'all open items'}}."
// With task td-abc123: "Review td-abc123."
// Without task: "Review all open items."
```

**Only single quotes work for fallback text:**
- `{{ticket || 'default'}}` -- works
- `{{ticket || "default"}}` -- does NOT work

## Examples

### Task-Required Workflow
```json
{
  "name": "Implement Task",
  "ticketMode": "required",
  "body": "Begin work on {{ticket}}. Use td to track progress.\nRead the task description carefully before starting."
}
```

### Standalone Workflow
```json
{
  "name": "Run Tests",
  "ticketMode": "none",
  "body": "Run the full test suite. Fix any failures.\nReport a summary when complete."
}
```

### Flexible Workflow
```json
{
  "name": "Code Review Session",
  "ticketMode": "optional",
  "body": "Start a review session for {{ticket || 'open reviews'}}.\nCreate td tasks for any issues found."
}
```

### Backlog Refinement
```json
{
  "name": "Backlog Refinement",
  "ticketMode": "none",
  "body": "Start a backlog refinement session. Use td to find all tasks in the backlog starting with the oldest. For each task:\n\n1. Determine relevance - has it been implemented, is it still needed? If not relevant, comment on the task and close it. If questionable, leave open with notes.\n\n2. Update tasks with out-of-date code references to reflect the current state of the codebase.\n\nUse sub-agents to analyze multiple tasks in parallel. Present a report of findings and changes when complete."
}
```

## Default Prompts

If no config file exists, the app creates one with 5 built-in templates:
- Begin Work on Ticket (required)
- Code Review Ticket (required)
- Plan to Epic (No Impl) (none)
- Plan to Epic + Implement (none)
- TD Review Session (none)

Press `d` in the prompt picker to install defaults if none are configured.

## Scope Indicators

In the prompt picker, prompts show their source:
- `[G]` - Global prompt (from `~/.config/sidecar/`)
- `[P]` - Project prompt (from `.sidecar/`)

## Important Notes

- Prompt names must match exactly (case-sensitive) for project overrides
- Only single quotes work in fallback syntax
- `ticketMode` defaults to `optional` if omitted
- Config files are loaded when creating a workspace; restart the app to reload
- Keep names short (they appear in a table with limited width)
- Use `\n` for newlines in prompt bodies
