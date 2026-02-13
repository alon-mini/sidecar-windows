# Git Status Plugin: Staging & Commit Message Feature

## Summary

Add commit message writing and "stage all" functionality to the existing git status plugin.

## User Preferences

- Keep single-file s/u staging, add "stage all" with `S` key
- Commit execution: `Alt+Enter`
- Show summary stats (+/- totals, file count) in commit view header

## Files to Modify

| File                                        | Changes                                           |
| ------------------------------------------- | ------------------------------------------------- |
| `go.mod`                                    | Add `github.com/charmbracelet/bubbles` dependency |
| `internal/plugins/gitstatus/plugin.go`      | Add ViewModeCommit, commit state fields, handlers |
| `internal/plugins/gitstatus/tree.go`        | Add `Commit()` and `StageAll()` functions         |
| `internal/plugins/gitstatus/commit_view.go` | **NEW** - Commit view rendering                   |

## Implementation Steps

### 1. Add Dependency

```bash
go get github.com/charmbracelet/bubbles
```

### 2. Add ViewModeCommit to plugin.go (line ~27)

```go
ViewModeCommit  // Commit message editor
```

### 3. Add Commit State Fields to Plugin struct (after line 67)

```go
// Commit state
commitMessage    textarea.Model
commitError      string
commitInProgress bool
```

### 4. Add Git Operations to tree.go

**StageAll** - Stage all modified/untracked files:

```go
func (t *FileTree) StageAll() error {
    cmd := exec.Command("git", "add", "-A")
    cmd.Dir = t.workDir
    return cmd.Run()
}
```

**Commit** - Execute git commit:

```go
func Commit(workDir, message string) (string, error) {
    cmd := exec.Command("git", "commit", "-m", message)
    cmd.Dir = workDir
    output, err := cmd.CombinedOutput()
    // Parse output for commit hash
    // Return error with git message if failed
}
```

### 5. Add Key Handlers in plugin.go

**In updateStatus()** - Add these cases:

- `"S"` - Stage all files, refresh tree
- `"c"` - Enter commit mode (only if staged files exist)

**New updateCommit() method:**

- `Esc` - Cancel, return to status
- `Alt+Enter` - Execute commit (validate non-empty message)
- Other keys - Pass to textarea

### 6. Add Message Types

```go
type CommitSuccessMsg struct { Hash, Subject string }
type CommitErrorMsg struct { Err error }
```

### 7. Create commit_view.go

Renders:

```
 Commit                    [3 files: +45 -12]
 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Staged Files (3)
   M cmd/sidecar/main.go              +12 -3
   A internal/plugins/new.go          +45 -0
   D old/deprecated.go                +0 -20
 ──────────────────────────────────────────
 Commit Message
 ┌────────────────────────────────────────┐
 │                                        │
 │ (textarea)                             │
 │                                        │
 └────────────────────────────────────────┘
 ──────────────────────────────────────────
  Esc  Cancel   Alt+Enter  Commit
```

### 8. Wire Up in Update() and View()

**Update()** - Add cases for:

- `ViewModeCommit` in KeyMsg switch -> call updateCommit()
- `CommitSuccessMsg` -> reset to status, refresh
- `CommitErrorMsg` -> show error, keep message

**View()** - Add case for ViewModeCommit -> renderCommit()

### 9. Update Commands() and FocusContext()

- Add `"git-commit"` focus context
- Add commit-related commands to list

## Keybindings Summary

| Key          | Context    | Action                          |
| ------------ | ---------- | ------------------------------- |
| `s`          | git-status | Stage file under cursor         |
| `u`          | git-status | Unstage file under cursor       |
| `S`          | git-status | Stage all modified/untracked    |
| `c`          | git-status | Enter commit mode (if staged)   |
| `Esc`        | git-commit | Cancel commit, return to status |
| `Alt+Enter`  | git-commit | Execute commit                  |

## Error Handling

- Empty message: Show inline error, don't execute
- Git commit fails: Show error, preserve message for retry
- No staged files: `c` key is no-op
