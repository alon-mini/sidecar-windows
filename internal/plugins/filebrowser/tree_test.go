package filebrowser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSystemFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"macOS DS_Store", ".DS_Store", true},
		{"macOS Spotlight", ".Spotlight-V100", true},
		{"macOS Trashes", ".Trashes", true},
		{"macOS fseventsd", ".fseventsd", true},
		{"macOS TemporaryItems", ".TemporaryItems", true},
		{"macOS DocumentRevisions", ".DocumentRevisions-V100", true},
		{"Windows Thumbs.db", "Thumbs.db", true},
		{"Windows desktop.ini", "desktop.ini", true},
		{"Windows recycle bin", "$RECYCLE.BIN", true},
		{"macOS resource fork", "._something", true},
		{"macOS resource fork minimal", "._", true},
		{"regular file", "main.go", false},
		{"dotfile", ".gitignore", false},
		{"empty string", "", false},
		{"underscore only", "_test.go", false},
		{"dot only", ".", false},
		{"double dot", "..", false},
		{"similar but not system", ".DS_Store2", false},
		{"case sensitive DS_Store", ".ds_store", false},
		{"case sensitive Thumbs", "thumbs.db", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSystemFile(tt.input); got != tt.expected {
				t.Errorf("isSystemFile(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFileTree_SystemFilesFiltered(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".DS_Store"), []byte(""), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "Thumbs.db"), []byte(""), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "._resource"), []byte(""), 0644)

	tree := NewFileTree(tmpDir)
	if err := tree.Build(); err != nil {
		t.Fatal(err)
	}

	for _, node := range tree.FlatList {
		if isSystemFile(node.Name) {
			t.Errorf("system file %q should have been filtered from tree", node.Name)
		}
	}

	var foundMain bool
	for _, node := range tree.FlatList {
		if node.Name == "main.go" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Error("Expected main.go to be in the tree")
	}
}

func TestFileTree_Build(t *testing.T) {
	// Use current directory for testing
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tree := NewFileTree(cwd)
	if err := tree.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if tree.Len() == 0 {
		t.Error("Expected non-empty tree")
	}

	// Verify root is set
	if tree.Root == nil {
		t.Error("Expected root to be set")
	}
}

func TestFileTree_ExpandCollapse(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, "subdir", "nested"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.txt"), []byte("test"), 0644)

	tree := NewFileTree(tmpDir)
	if err := tree.Build(); err != nil {
		t.Fatal(err)
	}

	// Find the subdir node
	var subdirNode *FileNode
	for _, node := range tree.FlatList {
		if node.Name == "subdir" && node.IsDir {
			subdirNode = node
			break
		}
	}

	if subdirNode == nil {
		t.Fatal("Expected to find subdir node")
	}

	initialLen := tree.Len()

	// Expand
	if err := tree.Expand(subdirNode); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if tree.Len() <= initialLen {
		t.Error("Expected tree length to increase after expand")
	}

	expandedLen := tree.Len()

	// Collapse
	tree.Collapse(subdirNode)

	if tree.Len() >= expandedLen {
		t.Error("Expected tree length to decrease after collapse")
	}
}

func TestFileTree_GetNode(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	tree := NewFileTree(tmpDir)
	if err := tree.Build(); err != nil {
		t.Fatal(err)
	}

	// Valid index
	node := tree.GetNode(0)
	if node == nil {
		t.Error("Expected node at index 0")
	}

	// Invalid indices
	if tree.GetNode(-1) != nil {
		t.Error("Expected nil for negative index")
	}
	if tree.GetNode(1000) != nil {
		t.Error("Expected nil for out of bounds index")
	}
}

func TestSortChildren(t *testing.T) {
	children := []*FileNode{
		{Name: "zebra.txt", IsDir: false},
		{Name: "alpha", IsDir: true},
		{Name: "beta.txt", IsDir: false},
		{Name: "delta", IsDir: true},
	}

	sortChildren(children, SortByName)

	// Directories should come first
	if !children[0].IsDir || !children[1].IsDir {
		t.Error("Directories should be sorted first")
	}

	// Then alphabetical
	if children[0].Name != "alpha" || children[1].Name != "delta" {
		t.Error("Directories should be alphabetically sorted")
	}
	if children[2].Name != "beta.txt" || children[3].Name != "zebra.txt" {
		t.Error("Files should be alphabetically sorted")
	}
}

func TestFileTree_RefreshPreservesExpandedState(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, "dir1", "nested"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "dir2"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "dir1", "file.txt"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "dir1", "nested", "deep.txt"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "dir2", "other.txt"), []byte("test"), 0644)

	tree := NewFileTree(tmpDir)
	if err := tree.Build(); err != nil {
		t.Fatal(err)
	}

	// Find and expand dir1
	var dir1Node *FileNode
	for _, node := range tree.FlatList {
		if node.Name == "dir1" && node.IsDir {
			dir1Node = node
			break
		}
	}
	if dir1Node == nil {
		t.Fatal("Expected to find dir1 node")
	}

	if err := tree.Expand(dir1Node); err != nil {
		t.Fatal(err)
	}

	// Find and expand nested
	var nestedNode *FileNode
	for _, node := range tree.FlatList {
		if node.Name == "nested" && node.IsDir {
			nestedNode = node
			break
		}
	}
	if nestedNode == nil {
		t.Fatal("Expected to find nested node")
	}

	if err := tree.Expand(nestedNode); err != nil {
		t.Fatal(err)
	}

	// Remember the length after expanding
	expandedLen := tree.Len()

	// Verify expanded paths are tracked
	expandedPaths := tree.GetExpandedPaths()
	if !expandedPaths["dir1"] {
		t.Error("Expected dir1 to be in expanded paths")
	}
	if !expandedPaths["dir1/nested"] {
		t.Error("Expected dir1/nested to be in expanded paths")
	}

	// Add a new file (simulating external change)
	_ = os.WriteFile(filepath.Join(tmpDir, "newfile.txt"), []byte("new"), 0644)

	// Refresh the tree
	if err := tree.Refresh(); err != nil {
		t.Fatal(err)
	}

	// Verify expanded state is preserved
	// Tree should have similar length (plus the new file)
	if tree.Len() < expandedLen {
		t.Errorf("Tree length decreased after refresh: got %d, want >= %d", tree.Len(), expandedLen)
	}

	// Find dir1 again and verify it's still expanded
	var dir1AfterRefresh *FileNode
	for _, node := range tree.FlatList {
		if node.Name == "dir1" && node.IsDir {
			dir1AfterRefresh = node
			break
		}
	}
	if dir1AfterRefresh == nil {
		t.Fatal("Expected to find dir1 after refresh")
	}
	if !dir1AfterRefresh.IsExpanded {
		t.Error("dir1 should still be expanded after refresh")
	}

	// Find nested and verify it's still expanded
	var nestedAfterRefresh *FileNode
	for _, node := range tree.FlatList {
		if node.Name == "nested" && node.IsDir {
			nestedAfterRefresh = node
			break
		}
	}
	if nestedAfterRefresh == nil {
		t.Fatal("Expected to find nested after refresh")
	}
	if !nestedAfterRefresh.IsExpanded {
		t.Error("nested should still be expanded after refresh")
	}

	// Verify new file is visible
	var foundNewFile bool
	for _, node := range tree.FlatList {
		if node.Name == "newfile.txt" {
			foundNewFile = true
			break
		}
	}
	if !foundNewFile {
		t.Error("Expected to find newfile.txt after refresh")
	}
}

func TestFileTree_GetExpandedPaths_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755)

	tree := NewFileTree(tmpDir)
	if err := tree.Build(); err != nil {
		t.Fatal(err)
	}

	// No directories expanded
	paths := tree.GetExpandedPaths()
	if len(paths) != 0 {
		t.Errorf("Expected no expanded paths, got %d", len(paths))
	}
}

func TestFileTree_RestoreExpandedPaths_DeletedDir(t *testing.T) {
	// Test that restoring works gracefully when a previously expanded dir is deleted
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, "willdelete"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "willdelete", "file.txt"), []byte("test"), 0644)

	tree := NewFileTree(tmpDir)
	if err := tree.Build(); err != nil {
		t.Fatal(err)
	}

	// Expand the directory
	var dirNode *FileNode
	for _, node := range tree.FlatList {
		if node.Name == "willdelete" && node.IsDir {
			dirNode = node
			break
		}
	}
	if dirNode == nil {
		t.Fatal("Expected to find willdelete node")
	}
	_ = tree.Expand(dirNode)

	// Delete the directory
	_ = os.RemoveAll(filepath.Join(tmpDir, "willdelete"))

	// Refresh should not error even though expanded dir is gone
	if err := tree.Refresh(); err != nil {
		t.Fatalf("Refresh failed after deleting expanded dir: %v", err)
	}

	// Tree should not contain the deleted directory
	for _, node := range tree.FlatList {
		if node.Name == "willdelete" {
			t.Error("Deleted directory should not be in tree")
		}
	}
}
