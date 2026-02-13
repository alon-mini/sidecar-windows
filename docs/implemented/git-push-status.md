# Git Push Status & Push Integration

Implementation notes for git push status indicators and push functionality in the git plugin.

## Overview

Added push/unpush status indicators to commits and implemented push functionality. Users can now see which commits have been pushed vs unpushed, and push commits directly from the UI.

## Files Modified

- `internal/plugins/gitstatus/push.go` (new) - Push status detection and execution
- `internal/plugins/gitstatus/history.go` - Added `Pushed` field to Commit struct
- `internal/plugins/gitstatus/plugin.go` - State management, keyboard handlers, message types
- `internal/plugins/gitstatus/history_view.go` - Push indicators in history view
- `internal/plugins/gitstatus/sidebar_view.go` - Push indicators in recent commits

## Key Features

### Visual Indicators

- `↑` (amber) - Unpushed commits
- No indicator - Pushed commits
- Header shows ahead/behind status: `↑3 ↓2`

### Keyboard Shortcuts

- `P` - Push to remote (in status view and history view)

### Edge Case Handling

| Scenario | Behavior |
|----------|----------|
| Detached HEAD | Shows "detached" in status, push disabled |
| No upstream | Shows "no upstream", push creates tracking branch |
| No remote | Git error displayed to user |
| Force push needed | Uses `--force-with-lease` for safety |

## Git Commands Used

```bash
# Get current branch
git branch --show-current

# Get upstream branch name
git rev-parse --abbrev-ref @{upstream}

# Get ahead/behind counts
git rev-list --count --left-right @{upstream}...HEAD

# Get unpushed commit hashes
git log @{upstream}..HEAD --format=%H

# Push with upstream tracking
git push -u origin HEAD

# Force push (when needed)
git push --force-with-lease -u origin HEAD
```

## Architecture Decisions

### Async Loading

Push status is loaded asynchronously alongside commit history to avoid blocking the UI. The `GetCommitHistoryWithPushStatus` function combines both operations.

### Message Flow

```
loadRecentCommits() -> RecentCommitsLoadedMsg{Commits, PushStatus}
loadHistory() -> HistoryLoadedMsg{Commits, PushStatus}
doPush() -> PushSuccessMsg/PushErrorMsg
```

### State Management

Push status is stored in `Plugin.pushStatus` and refreshed on:
- Plugin focus
- Manual refresh (r key)
- File system changes
- Successful push

## UI/UX Notes

- Push indicator appears before commit hash for clear visual grouping
- Ahead/behind counts shown in section headers
- Push in progress shows "Pushing..." status
- Push errors displayed with `✗` prefix (truncated if needed)
- Error cleared on refresh or successful push

## Testing Notes

- Build: `go build ./...`
- Tests: `go test ./internal/plugins/gitstatus/...`
- Manual testing recommended for:
  - New repos without any pushes
  - Repos with no remote configured
  - Detached HEAD state
  - Branches that need force push
