# Sidebar List Implementation Reference

Conventions and pitfalls for building scrollable sidebar lists in sidecar plugins.

## Layout Accounting (Most Common Source of Bugs)

Sidebar list bugs almost always come from mismatched height math. The renderer and scroll logic must use the same line budget.

- Compute pane height once at the view layer, then pass `visibleHeight` into list rendering and sizing logic.
- When using bordered panels with padding, account for borders and header lines exactly once.
- Treat `visibleHeight` as post-header content height. Do not subtract header lines again later.
- Track every line the sidebar renders (headers, blank lines, separators, status lines, rows).
- Truncate headers to sidebar width to prevent line wrapping from stealing rows.

Typical layout flow:
```
paneHeight = height - 2         // outer borders
innerHeight = paneHeight - 2    // panel border + header
filesSection = 3                // minimum for status or staged files
separator = 1
commitsHeader = 1
commitCapacity = innerHeight - filesSection - separator - commitsHeader
```

Centralize capacity calculation in a helper used by both render and scroll logic. See `gitstatus.commitSectionCapacity()` for reference.

## Scroll + Cursor Stability

- Keep scroll offsets clamped after any data change.
- Only adjust scroll if selected row is outside visible range.
- Never mutate scroll offsets in render functions. Rendering must be pure.
- Use absolute indices for hit regions to avoid drift with filters/pagination.

Explicit clamping pattern:
```go
if p.commitScrollOff > len(commits) - visibleCommits {
    p.commitScrollOff = len(commits) - visibleCommits
}
if p.commitScrollOff < 0 {
    p.commitScrollOff = 0
}
```

Paging keys: `j`/`down`, `k`/`up`, `ctrl+d`, `ctrl+u`, `g`, `G`

## Async Data + Refresh Safety

- When loading more history, append to existing data.
- When refreshing, merge by hash/ID to preserve older entries.
- Preserve selection by stable ID (hash), not index.
- Update metadata in place instead of replacing the list when possible.
- After any merge/replace, clamp scroll offsets and cursor ranges immediately.
- Auto-load additional pages if viewport is taller than initial page size.

## Empty Rows at Bottom

Caused by renderer/scroll disagreement on visible row count. Fix by:
- Using one shared capacity calculation for both rendering and scroll clamping.
- Accounting for optional rows (status lines, "empty" placeholders).
- Not double-counting panel headers or borders.
- Clamping `scrollOff` to `len(items) - visibleCount` (minimum 0).
- Loading more data when list is shorter than visible window.

## Mouse Hit Regions

- Register hit regions after the line is rendered and after updating `currentY`.
- Use absolute indices so clicks map correctly even when scrolled.
- Register in order of increasing priority (last wins):
```go
p.HitMap.AddRect(regionSidebar, x, y, w, h, data)        // lowest
p.HitMap.AddRect(regionCommitLine, x, y, w, h, data)     // medium
p.HitMap.AddRect(regionCommitDivider, x, y, w, h, data)  // highest
```

## Height Enforcement

Always constrain output height:
```go
lipgloss.NewStyle().Width(w).Height(h).MaxHeight(h).Render(content)
```

Never rely on the app to truncate overflow. Never mutate plugin state during rendering.

## Common Pitfalls Checklist

- Double-counted header lines or borders
- Scroll offsets adjusted inside render methods
- Refresh replaced list, losing older entries
- No clamp after list size changes
- Hit regions using visible index instead of absolute index
- Placeholder lines not counted in height math
