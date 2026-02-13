package filebrowser

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SortMode represents how files are sorted in the tree.
type SortMode int

const (
	SortByName SortMode = iota
	SortBySize
	SortByTime
	SortByType
)

// SortModeLabel returns a short label for display.
func (s SortMode) Label() string {
	switch s {
	case SortByName:
		return "name"
	case SortBySize:
		return "size"
	case SortByTime:
		return "time"
	case SortByType:
		return "type"
	default:
		return "name"
	}
}

// NextSortMode cycles to the next sort mode.
func (s SortMode) Next() SortMode {
	return (s + 1) % 4
}

// FileNode represents a file or directory in the tree.
type FileNode struct {
	Name       string
	Path       string // Relative path from root
	IsDir      bool
	IsExpanded bool
	IsIgnored  bool // Set by gitignore
	Children   []*FileNode
	Parent     *FileNode
	Depth      int
	Size       int64
	ModTime    time.Time
}

// FileTree manages the hierarchical file structure.
type FileTree struct {
	Root        *FileNode
	RootDir     string
	FlatList    []*FileNode // Flattened visible nodes for cursor navigation
	gitIgnore   *GitIgnore
	SortMode    SortMode // Current sort mode
	ShowIgnored bool     // Whether to include ignored files in FlatList
}

// NewFileTree creates a new file tree rooted at the given directory.
func NewFileTree(rootDir string) *FileTree {
	return &FileTree{
		RootDir:     rootDir,
		FlatList:    make([]*FileNode, 0),
		gitIgnore:   NewGitIgnore(),
		ShowIgnored: true, // Show ignored files by default
	}
}

// Build initializes the tree by loading the root directory's children.
func (t *FileTree) Build() error {
	// Load .gitignore from root
	t.gitIgnore = NewGitIgnore()
	_ = t.gitIgnore.LoadFile(filepath.Join(t.RootDir, ".gitignore"))

	t.Root = &FileNode{
		Name:       filepath.Base(t.RootDir),
		Path:       "",
		IsDir:      true,
		IsExpanded: true,
		Depth:      -1, // Root is hidden, children start at depth 0
	}

	if err := t.loadChildren(t.Root); err != nil {
		return err
	}

	t.Flatten()
	return nil
}

// isSystemFile returns true for OS-generated files that clutter file browsers.
func isSystemFile(name string) bool {
	// Exact matches
	switch name {
	case ".DS_Store", ".Spotlight-V100", ".Trashes", ".fseventsd",
		".TemporaryItems", ".DocumentRevisions-V100",
		"Thumbs.db", "desktop.ini", "$RECYCLE.BIN":
		return true
	}
	// macOS resource fork files (._*)
	if strings.HasPrefix(name, "._") {
		return true
	}
	return false
}

// loadChildren populates a node's children from the filesystem.
func (t *FileTree) loadChildren(node *FileNode) error {
	fullPath := filepath.Join(t.RootDir, node.Path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return err
	}

	node.Children = make([]*FileNode, 0, len(entries))

	for _, entry := range entries {
		if isSystemFile(entry.Name()) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		childPath := filepath.Join(node.Path, entry.Name())
		child := &FileNode{
			Name:      entry.Name(),
			Path:      childPath,
			IsDir:     entry.IsDir(),
			IsIgnored: t.gitIgnore.IsIgnored(childPath, entry.IsDir()),
			Parent:    node,
			Depth:     node.Depth + 1,
			Size:      info.Size(),
			ModTime:   info.ModTime(),
		}

		node.Children = append(node.Children, child)
	}

	sortChildren(node.Children, t.SortMode)
	return nil
}

// sortChildren sorts nodes according to the given mode.
func sortChildren(children []*FileNode, mode SortMode) {
	sort.Slice(children, func(i, j int) bool {
		// Directories always come before files
		if children[i].IsDir != children[j].IsDir {
			return children[i].IsDir
		}

		switch mode {
		case SortBySize:
			// Larger files first
			if children[i].Size != children[j].Size {
				return children[i].Size > children[j].Size
			}
			// Fall back to name
			return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)

		case SortByTime:
			// Newer files first
			if !children[i].ModTime.Equal(children[j].ModTime) {
				return children[i].ModTime.After(children[j].ModTime)
			}
			// Fall back to name
			return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)

		case SortByType:
			// Sort by extension, then by name
			exti := strings.ToLower(filepath.Ext(children[i].Name))
			extj := strings.ToLower(filepath.Ext(children[j].Name))
			if exti != extj {
				return exti < extj
			}
			return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)

		default: // SortByName
			return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
		}
	})
}

// Expand opens a directory node, loading children if needed.
func (t *FileTree) Expand(node *FileNode) error {
	if !node.IsDir {
		return nil
	}

	if len(node.Children) == 0 {
		if err := t.loadChildren(node); err != nil {
			return err
		}
	}

	node.IsExpanded = true
	t.Flatten()
	return nil
}

// Collapse closes a directory node.
func (t *FileTree) Collapse(node *FileNode) {
	node.IsExpanded = false
	t.Flatten()
}

// Toggle expands or collapses a directory node.
func (t *FileTree) Toggle(node *FileNode) error {
	if !node.IsDir {
		return nil
	}

	if node.IsExpanded {
		t.Collapse(node)
		return nil
	}
	return t.Expand(node)
}

// Flatten rebuilds the FlatList from visible nodes.
func (t *FileTree) Flatten() []*FileNode {
	t.FlatList = t.FlatList[:0] // Reuse slice
	if t.Root != nil {
		t.flattenNode(t.Root)
	}
	return t.FlatList
}

func (t *FileTree) flattenNode(node *FileNode) {
	for _, child := range node.Children {
		// Skip ignored files/folders when ShowIgnored is false
		if !t.ShowIgnored && child.IsIgnored {
			continue
		}
		t.FlatList = append(t.FlatList, child)
		if child.IsDir && child.IsExpanded {
			t.flattenNode(child)
		}
	}
}

// GetNode returns the node at the given index, or nil if out of bounds.
func (t *FileTree) GetNode(index int) *FileNode {
	if index < 0 || index >= len(t.FlatList) {
		return nil
	}
	return t.FlatList[index]
}

// Len returns the number of visible nodes.
func (t *FileTree) Len() int {
	return len(t.FlatList)
}

// FindParentDir returns the parent directory node, or nil if at root.
func (t *FileTree) FindParentDir(node *FileNode) *FileNode {
	if node == nil || node.Parent == nil || node.Parent == t.Root {
		return nil
	}
	return node.Parent
}

// IndexOf returns the index of a node in the flat list, or -1 if not found.
func (t *FileTree) IndexOf(node *FileNode) int {
	for i, n := range t.FlatList {
		if n == node {
			return i
		}
	}
	return -1
}

// FindByPath returns the node with the given relative path, or nil if not found.
func (t *FileTree) FindByPath(path string) *FileNode {
	for _, n := range t.FlatList {
		if n.Path == path {
			return n
		}
	}
	return nil
}

// GetExpandedPaths returns the paths of all expanded directories.
func (t *FileTree) GetExpandedPaths() map[string]bool {
	expanded := make(map[string]bool)
	if t.Root != nil {
		t.collectExpanded(t.Root, expanded)
	}
	return expanded
}

func (t *FileTree) collectExpanded(node *FileNode, expanded map[string]bool) {
	for _, child := range node.Children {
		if child.IsDir && child.IsExpanded {
			expanded[child.Path] = true
			t.collectExpanded(child, expanded)
		}
	}
}

// RestoreExpandedPaths expands directories that were previously expanded.
func (t *FileTree) RestoreExpandedPaths(paths map[string]bool) {
	if t.Root == nil || len(paths) == 0 {
		return
	}
	t.restoreExpanded(t.Root, paths)
	t.Flatten()
}

func (t *FileTree) restoreExpanded(node *FileNode, paths map[string]bool) {
	for _, child := range node.Children {
		if child.IsDir && paths[child.Path] {
			// Load children if needed and expand
			if len(child.Children) == 0 {
				_ = t.loadChildren(child)
			}
			child.IsExpanded = true
			t.restoreExpanded(child, paths)
		}
	}
}

// Refresh reloads the tree from disk, preserving expanded state.
func (t *FileTree) Refresh() error {
	// Save expanded state before rebuild
	expandedPaths := t.GetExpandedPaths()

	// Rebuild tree
	if err := t.Build(); err != nil {
		return err
	}

	// Restore expanded state
	t.RestoreExpandedPaths(expandedPaths)
	return nil
}

// SetSortMode changes the sort mode and re-sorts the tree.
func (t *FileTree) SetSortMode(mode SortMode) {
	t.SortMode = mode
	if t.Root != nil {
		t.resortNode(t.Root)
		t.Flatten()
	}
}

// resortNode recursively re-sorts a node and its children.
func (t *FileTree) resortNode(node *FileNode) {
	if len(node.Children) > 0 {
		sortChildren(node.Children, t.SortMode)
		for _, child := range node.Children {
			if child.IsDir {
				t.resortNode(child)
			}
		}
	}
}
