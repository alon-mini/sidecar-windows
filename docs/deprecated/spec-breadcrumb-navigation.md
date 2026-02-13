# Navigation Breadcrumb Implementation Plan

## Summary

Add a breadcrumb trail at the top of plugins showing navigation depth (e.g., `History › abc1234 › file.go`). Helps users understand how many times to press escape to return to top level.

## Design Decisions

- **Separator**: `›` (arrow) with space padding: `›`
- **No escape hint**: Just show the path, no "(2x esc)" suffix
- **Status view**: No breadcrumb when focusing diff pane (use border highlighting only)
- **Breadcrumb replaces existing headers**: No additional height impact

## Files to Modify

### 1. New Shared Utility

**Create**: `internal/ui/breadcrumb.go`

```go
package ui

// Render builds a breadcrumb string from path segments
// - Uses styles.Muted for parent items
// - Uses styles.Title (bold) for current location (last item)
// - Separator: " › " with styles.Subtle
func Render(items []string, width int) string
```

### 2. Git Plugin

**File**: `internal/plugins/gitstatus/plugin.go`

- Add `breadcrumb() []string` method that returns path based on `viewMode`:
  - `ViewModeStatus`: `["Git Status"]` (top level, no breadcrumb shown)
  - `ViewModeHistory`: `["Git Status", "History"]`
  - `ViewModeCommitDetail`: `["Git Status", "History", shortHash]`
  - `ViewModeDiff` (from commit): `["Git Status", "History", shortHash, filename]`
  - `ViewModeDiff` (from status): `["Git Status", filename]`
  - `ViewModeCommit`: `["Git Status", "Commit"]`

**File**: `internal/plugins/gitstatus/history_view.go`

- `renderHistory()` (line 11): Replace static header with breadcrumb
- `renderCommitDetail()` (line 85): Replace static header with breadcrumb

**File**: `internal/plugins/gitstatus/view.go`

- `renderDiffModal()`: Add breadcrumb header

**File**: `internal/plugins/gitstatus/commit_view.go`

- `renderCommit()`: Add breadcrumb header

### 3. Conversations Plugin

**File**: `internal/plugins/conversations/plugin.go`

- Add `breadcrumb() []string` method:
  - `ViewSessions`: `["Sessions"]` (top level, no breadcrumb)
  - `ViewMessages`: `["Sessions", sessionName]`
  - `ViewAnalytics`: `["Sessions", "Analytics"]`

**File**: `internal/plugins/conversations/view.go`

- `renderMessages()` (line 202): Add breadcrumb before session header
- `renderAnalytics()`: Add breadcrumb header (need to locate/create this function)

## Implementation Steps

1. **Create breadcrumb utility** (`internal/ui/breadcrumb.go`)
   - `Render(items []string, width int) string` - renders breadcrumb line
   - Uses existing styles: `styles.Muted`, `styles.Title`, `styles.Subtle`

2. **Update Git Plugin**
   - Add `breadcrumb()` method to Plugin struct
   - Modify `renderHistory()` to use breadcrumb instead of static " Commit History"
   - Modify `renderCommitDetail()` to show commit hash in breadcrumb
   - Modify `renderDiffModal()` to show full path (with context awareness)
   - Modify `renderCommit()` for commit message editor

3. **Update Conversations Plugin**
   - Add `breadcrumb()` method
   - Modify `renderMessages()` to include breadcrumb
   - Modify analytics view to include breadcrumb

4. **Test height constraints**
   - Verify header stays in place after navigating through views
   - Test with small terminal sizes

## Height Calculation Notes

- Breadcrumb **replaces** existing headers (e.g., " Commit History" becomes "Git Status › History")
- No net change to `contentHeight` calculations
- Existing pattern: `contentHeight := p.height - 3` (header + separator + padding)
- Keep this pattern unchanged

## Visual Example

```
Git Status › History › abc1234
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Author: John Doe <john@example.com>
  Date:   Mon Jan 2 15:04:05 2024

  feat: add user authentication
```

## Styling Reference

- Parent items: `styles.Muted` (gray, #6B7280)
- Current item: `styles.Title` (bold, white #F9FAFB)
- Separator `›`: `styles.Subtle` (darker gray, #4B5563)
- Separator line: `styles.Muted.Render(strings.Repeat("━", width-2))`
