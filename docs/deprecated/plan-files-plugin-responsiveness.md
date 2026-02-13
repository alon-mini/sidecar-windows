# Files Plugin Responsiveness Fix

## Problem

Clicking a file in the tree pane caused a 1.4s freeze before the preview loaded.

## Root Cause

`syncTreeSelection()` in `tabs.go` called `walkTree()` on every file click. `walkTree()` recursively visits ALL directories and loads unloaded children from disk via `os.ReadDir()` + `entry.Info()`. For this project, it visited **47,912 nodes** taking **1.4 seconds** — blocking the entire bubbletea event loop.

The call path: click → `handleMouseClick` → `loadPreviewForCursor` → `openTab` → `applyActiveTab` → `syncTreeSelection` → `walkTree` (loads entire tree from disk).

The irony: `handleMouseClick` already set `p.treeCursor = idx` correctly. Then `syncTreeSelection` walked the entire tree just to set it again.

## Fix (implemented)

Added fast paths to `syncTreeSelection` (`tabs.go`):

```go
func (p *Plugin) syncTreeSelection(path string) {
    // Fast path: cursor already points to this file (click case)
    if node := p.tree.GetNode(p.treeCursor); node != nil && node.Path == path {
        return
    }

    // Try FlatList lookup (no disk I/O, O(n) over visible nodes)
    for i, node := range p.tree.FlatList {
        if node.Path == path {
            p.treeCursor = i
            p.ensureTreeCursorVisible()
            return
        }
    }

    // Fallback: walk full tree (only for files in unexpanded directories)
    ...
}
```

## Additional Enhancement

Arrow keys (`j`/`k`/`up`/`down`) in the tree pane now load the preview for the file under the cursor, enabling fast file browsing by keyboard.

## Theories Investigated and Disproven

1. **Tree rebuild/state restoration blocks UI** — Tree ops are 7-18ms, not the bottleneck
2. **LoadPreview (file I/O + chroma) is slow** — Goroutine finishes in 5-12ms
3. **Scheduling delay** — Only 30-85us
4. **Mouse event flood** — Only 6-7 updates in the 1.5s gap
5. **App-level View() blocking** — View is 10-15ms

## Future Work (not yet implemented)

The original atomic-swap plan (build new tree in background, swap when ready) would further improve responsiveness by eliminating the data race in `refresh()` where `p.tree.Build()` is called from a goroutine while the main loop reads the same tree. This is a correctness issue more than a performance one — the current tree rebuild is fast (7-18ms) so the race window is small.
