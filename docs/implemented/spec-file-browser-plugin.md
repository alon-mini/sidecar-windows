# File Browser Plugin Specification

## Overview

A 2-pane file browser plugin for sidecar that provides IDE-like file exploration without leaving the terminal. Tree view on the left (~30%), file preview with syntax highlighting on the right (~70%).

## Goals

- Eliminate need for VS Code/IDE to browse files when using CLI agents
- Vim-style navigation (j/k, h/l for expand/collapse)
- Syntax highlighting for 500+ languages via chroma
- Show gitignored files with muted styling
- Open files in user's $EDITOR

## UI Layout

```
+------------------------------------------+
| Agent Sidecar  [ F File Browser ]  14:32 |
+------------------------------------------+
| Tree (~30%)     │ Preview (~70%)         |
|-----------------|------------------------|
| > src/          | // main.go             |
|   main.go       | package main           |
|   go.mod        |                        |
|   .gitignore    | import (               |
| > internal/     |     "fmt"              |
|   + plugin/     |     "os"               |
|   + app/        | )                      |
| > docs/         |                        |
|   .env (muted)  | func main() {          |
|                 |     fmt.Println("hi")  |
+------------------------------------------+
| tab pane  j/k nav  e edit  / search      |
+------------------------------------------+
```

### Visual Indicators

| Element | Style |
|---------|-------|
| Directories | Bold blue, prefix `>` (expanded) or `+` (collapsed) |
| Files | Normal text |
| Gitignored | Muted/dimmed color |
| Selected | Background highlight |
| Active pane | Purple border |
| Inactive pane | Gray border |

## Keybindings

### Tree Pane (`file-browser-tree`)

| Key | Action |
|-----|--------|
| `j` / `down` | Move cursor down |
| `k` / `up` | Move cursor up |
| `l` / `right` / `enter` | Expand directory OR preview file |
| `h` / `left` | Collapse directory OR jump to parent |
| `e` / `o` | Open file in $EDITOR |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `/` | Enter search mode |
| `.` | Toggle hidden files |
| `r` | Refresh tree |
| `tab` | Switch to preview pane |

### Preview Pane (`file-browser-preview`)

| Key | Action |
|-----|--------|
| `j` / `down` | Scroll down |
| `k` / `up` | Scroll up |
| `g` | Scroll to top |
| `G` | Scroll to bottom |
| `ctrl+d` | Page down |
| `ctrl+u` | Page up |
| `tab` | Switch to tree pane |

### Search Mode (`file-browser-search`)

| Key | Action |
|-----|--------|
| `<typing>` | Filter matches |
| `up` / `ctrl+p` | Previous match |
| `down` / `ctrl+n` | Next match |
| `enter` | Jump to match, exit search |
| `esc` | Exit search mode |
| `n` | Next match (after exiting search) |
| `N` | Previous match |

## Data Structures

### FileNode

```go
type FileNode struct {
    Name       string      // File/directory name
    Path       string      // Relative path from root
    IsDir      bool
    IsExpanded bool        // For directories
    IsIgnored  bool        // Matches .gitignore
    Children   []*FileNode
    Parent     *FileNode
    Depth      int         // Indentation level
    Size       int64
    ModTime    time.Time
}
```

### FileTree

```go
type FileTree struct {
    Root      *FileNode
    RootDir   string
    FlatList  []*FileNode  // Flattened visible nodes for cursor navigation
    gitIgnore *GitIgnore
}
```

### Plugin State

```go
type FocusPane int
const (
    PaneTree FocusPane = iota
    PanePreview
)

type Plugin struct {
    ctx          *plugin.Context
    tree         *FileTree
    focused      bool
    activePane   FocusPane

    // Tree pane
    treeCursor    int
    treeScrollOff int

    // Preview pane
    previewContent    string
    previewLines      []string
    previewHighlighted []string  // Syntax highlighted
    previewScroll     int
    previewFile       string
    isBinary          bool

    // Search
    searchMode    bool
    searchQuery   string
    searchMatches []*FileNode

    // Dimensions
    width, height int
    treeWidth     int   // ~30%
    previewWidth  int   // ~70%

    watcher *Watcher
}
```

## File Structure

```
internal/plugins/filebrowser/
├── plugin.go      # Main plugin, Update/View, keybindings
├── tree.go        # FileTree, FileNode, expand/collapse
├── view.go        # 2-pane rendering, renderTreePane, renderPreviewPane
├── preview.go     # File loading, binary detection, highlighting
├── gitignore.go   # .gitignore parsing
└── watcher.go     # fsnotify-based file watching
```

## Implementation Phases

### Phase 1: Core Tree (~5 pts)
- FileNode and FileTree data structures
- Build(), Expand(), Collapse(), Flatten()
- Directories sorted before files, alphabetical

### Phase 2: Plugin Skeleton (~3 pts)
- Implement plugin.Plugin interface
- Single-pane tree rendering
- j/k/l/h navigation
- Register in main.go

### Phase 3: 2-Pane Layout (~5 pts)
- lipgloss.JoinHorizontal() for side-by-side
- 30%/70% width split
- Tab to switch focus
- Border styling based on active pane

### Phase 4: File Preview (~5 pts)
- Async file loading (tea.Cmd)
- Binary detection (null bytes in first 512)
- Large file truncation (500KB limit)
- Line numbers
- Scroll with j/k/g/G

### Phase 5: Syntax Highlighting (~5 pts)
- Add chroma dependency
- Detect language from extension
- terminal256 formatter, monokai theme
- Cache highlighted content

### Phase 6: Gitignore (~3 pts)
- Parse .gitignore patterns
- Glob to regex conversion
- Mark IsIgnored on nodes
- Muted styling

### Phase 7: Editor Integration (~2 pts)
- 'e' opens in $EDITOR
- tea.ExecProcess for terminal handoff
- Fallback to vim

### Phase 8: Search & Watcher (~5 pts)
- '/' enters search mode
- Fuzzy match across all files
- n/N to navigate matches
- fsnotify for auto-refresh
- 100ms debounce

## Dependencies

```go
require (
    github.com/alecthomas/chroma/v2 v2.12.0
)
```

Note: fsnotify already in go.mod (used by gitstatus)

## Styles to Add

```go
// internal/styles/styles.go
FileBrowserDir = lipgloss.NewStyle().
    Foreground(Secondary).
    Bold(true)

FileBrowserFile = lipgloss.NewStyle().
    Foreground(TextPrimary)

FileBrowserIgnored = lipgloss.NewStyle().
    Foreground(TextSubtle)

FileBrowserLineNumber = lipgloss.NewStyle().
    Foreground(TextMuted).
    Width(5).
    Align(lipgloss.Right)
```

## Edge Cases

### Binary Files
- Detect by checking for null bytes in first 512 bytes
- Display "Binary file" message in preview

### Large Files
- Read max 500KB
- Display max 10,000 lines
- Show truncation indicator: "... (file truncated, 15MB total)"

### Permission Errors
- Display error message gracefully
- Continue with other files

### Empty Directories
- Show as expandable but empty when expanded

### Deep Nesting
- Support arbitrary depth
- Consider limiting initial scan depth for performance

## Testing

### Unit Tests
- tree_test.go: Build, Expand, Collapse, Flatten
- gitignore_test.go: Pattern matching
- preview_test.go: Binary detection, truncation

### Integration Tests
- Full plugin rendering
- Navigation flows
- Search functionality

## TD Tasks

All implementation tasks tracked under epic `td-b520d0d7`:

```
td tree td-b520d0d7
```

Tasks include detailed implementation instructions, code snippets, and acceptance criteria.
