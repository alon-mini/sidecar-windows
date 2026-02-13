# Plan: Stacking Modal System for Epic Task Navigation

## Overview

Add a stacking modal system to TD monitor that allows viewing tasks within an epic, with navigation between parent/child modals and visual stack indicators.

## User Decisions

- **Navigation**: j/k selects task in list, Enter opens (keep flexible for future changes)
- **Visual indicators**: Breadcrumb trail + Border color change by depth
- **Deliverable**: Create td epic with sub-task issues

---

## Data Structure Design

### ModalEntry (new type in model.go)

```go
type ModalEntry struct {
    // Core
    IssueID     string
    SourcePanel Panel  // Only for base entry (depth 1)

    // Display
    Scroll      int

    // Async data
    Loading     bool
    Error       error
    Issue       *models.Issue
    Handoff     *models.Handoff
    Logs        []models.Log
    BlockedBy   []models.Issue
    Blocks      []models.Issue
    DescRender  string
    AcceptRender string

    // Epic-specific (when Issue.Type == "epic")
    EpicTasks          []models.Issue
    EpicTasksCursor    int
    TaskSectionFocused bool
}
```

### Model Changes

Replace flat modal fields with:

```go
ModalStack []ModalEntry  // Stack of modals (empty = no modal)
```

Helper methods: `ModalOpen()`, `ModalDepth()`, `CurrentModal()`, `ModalBreadcrumb()`

---

## Keymap Changes

### New Context

```go
ContextEpicTasks Context = "epic-tasks"  // When task list is focused
```

### New Commands

```go
CmdFocusTaskSection Command = "focus-task-section"
CmdOpenEpicTask     Command = "open-epic-task"
```

### New Bindings (bindings.go)

```go
// Epic tasks context (when task section focused)
{Key: "j", Command: CmdCursorDown, Context: ContextEpicTasks}
{Key: "k", Command: CmdCursorUp, Context: ContextEpicTasks}
{Key: "enter", Command: CmdOpenEpicTask, Context: ContextEpicTasks}
{Key: "tab", Command: CmdFocusTaskSection, Context: ContextEpicTasks}
{Key: "esc", Command: CmdClose, Context: ContextEpicTasks}

// Modal context addition
{Key: "tab", Command: CmdFocusTaskSection, Context: ContextModal}
```

---

## Visual Design

### Breadcrumb (footer when depth > 1)

```
epic: td-abc123 > task: td-xyz789
```

### Border Colors by Depth

- Depth 1: Purple/Magenta (primaryColor, 212)
- Depth 2: Cyan (45)
- Depth 3+: Orange (214)

### Epic Tasks Section in Modal

```
TASKS IN EPIC (5)
  td-abc001 [open]     First subtask title
  td-abc002 [in_prog]  Second subtask title
> td-abc003 [open]     Selected task (highlighted)
  td-abc004 [closed]   Fourth subtask title
```

---

## Files to Modify

| File                             | Changes                                                                                                                                      |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/monitor/model.go`           | ModalEntry type, ModalStack, helper methods, openModal→pushModal, closeModal (pop), executeCommand, currentContext, IssueDetailsMsg handling |
| `pkg/monitor/view.go`            | renderModal (epic tasks section), wrapModalWithDepth (breadcrumb + border colors), footer help updates                                       |
| `pkg/monitor/keymap/registry.go` | ContextEpicTasks constant, CmdFocusTaskSection, CmdOpenEpicTask                                                                              |
| `pkg/monitor/keymap/bindings.go` | Epic tasks context bindings                                                                                                                  |
| `pkg/monitor/keymap/help.go`     | Update help text for epic navigation                                                                                                         |

---

## Implementation Steps

### Task 1: Modal Stack Foundation

1. Add `ModalEntry` struct to model.go
2. Replace flat modal fields with `ModalStack []ModalEntry`
3. Add helper methods: `ModalOpen()`, `ModalDepth()`, `CurrentModal()`, `ModalBreadcrumb()`
4. Update `openModal()` to push onto stack
5. Update `closeModal()` to pop (return to prev if depth > 1)
6. Update `IssueDetailsMsg` handling to target top of stack
7. Update `MarkdownRenderedMsg` handling similarly

### Task 2: Keymap Extensions

1. Add `ContextEpicTasks` to registry.go
2. Add `CmdFocusTaskSection` and `CmdOpenEpicTask` commands
3. Add epic task context bindings to bindings.go
4. Update `currentContext()` to return `ContextEpicTasks` when `TaskSectionFocused`

### Task 3: Epic Tasks Data Fetching

1. Add `EpicTasks []models.Issue` to `IssueDetailsMsg`
2. Update `fetchIssueDetails()` to fetch child tasks when `Issue.Type == "epic"`
3. Use `db.ListIssues(ListIssuesOptions{ParentID: issueID})`

### Task 4: Epic Tasks UI Rendering

1. Add "TASKS IN EPIC" section to renderModal when issue is epic
2. Show task list with cursor highlighting when TaskSectionFocused
3. Format: ID, status badge, truncated title

### Task 5: Epic Task Navigation

1. Add `CmdFocusTaskSection` handling in executeCommand (toggle focus)
2. Update cursor movement to handle epic task cursor when TaskSectionFocused
3. Add `CmdOpenEpicTask` handling to push selected task as new modal
4. Add `pushModal(issueID)` method

### Task 6: Visual Stack Indicators

1. Create `wrapModalWithDepth()` replacing `wrapModal()`
2. Implement border color selection by depth
3. Add breadcrumb to footer when depth > 1
4. Update footer help text based on context/depth

### Task 7: Testing & Polish

1. Test stack operations (push, pop, depth)
2. Test escape at various depths
3. Test epic with 0 children (no task section)
4. Test non-epic issues (unchanged behavior)
5. Test h/l navigation only at depth 1

---

## Edge Cases

- **Deep nesting**: Epics can contain epics (natural via parent_id). Consider max depth limit (5?)
- **Empty epics**: No TASKS IN EPIC section if epic has 0 children
- **Refresh**: `r` key refreshes current modal only, not entire stack
- **h/l navigation**: Only available at depth 1 (base modal)

---

## After Plan Approval

Create td epic "Stacking modal system for epic task navigation" with sub-tasks for each implementation step.
