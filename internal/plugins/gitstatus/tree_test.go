package gitstatus

import (
	"os"
	"testing"
)

func TestNewFileTree(t *testing.T) {
	tree := NewFileTree("/tmp/test")
	if tree == nil {
		t.Fatal("expected non-nil tree")
	}
	if tree.workDir != "/tmp/test" {
		t.Errorf("expected workDir /tmp/test, got %s", tree.workDir)
	}
}

func TestFileTreeRefresh(t *testing.T) {
	// Skip if not in a git repo
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("couldn't get working directory")
	}

	tree := NewFileTree(cwd)
	err = tree.Refresh()
	if err != nil {
		t.Skipf("not in git repo or git not available: %v", err)
	}

	// Should have loaded something (or be clean)
	t.Logf("summary: %s", tree.Summary())
	t.Logf("total: %d files", tree.TotalCount())
}

func TestFileTreeSummary(t *testing.T) {
	tree := &FileTree{}

	// Empty tree
	if s := tree.Summary(); s != "clean" {
		t.Errorf("expected 'clean', got %q", s)
	}

	// With staged
	tree.Staged = []*FileEntry{{Path: "a.txt"}}
	if s := tree.Summary(); s != "1 staged" {
		t.Errorf("expected '1 staged', got %q", s)
	}

	// With modified
	tree.Modified = []*FileEntry{{Path: "b.txt"}, {Path: "c.txt"}}
	if s := tree.Summary(); s != "1 staged, 2 modified" {
		t.Errorf("expected '1 staged, 2 modified', got %q", s)
	}

	// With untracked
	tree.Untracked = []*FileEntry{{Path: "d.txt"}}
	if s := tree.Summary(); s != "1 staged, 2 modified, 1 untracked" {
		t.Errorf("expected '1 staged, 2 modified, 1 untracked', got %q", s)
	}
}

func TestFileTreeAllEntries(t *testing.T) {
	tree := &FileTree{
		Staged:    []*FileEntry{{Path: "a.txt"}},
		Modified:  []*FileEntry{{Path: "b.txt"}},
		Untracked: []*FileEntry{{Path: "c.txt"}},
	}

	entries := tree.AllEntries()
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Check order: staged, modified, untracked
	if entries[0].Path != "a.txt" {
		t.Errorf("expected first entry to be a.txt, got %s", entries[0].Path)
	}
	if entries[1].Path != "b.txt" {
		t.Errorf("expected second entry to be b.txt, got %s", entries[1].Path)
	}
	if entries[2].Path != "c.txt" {
		t.Errorf("expected third entry to be c.txt, got %s", entries[2].Path)
	}
}

func TestParseOrdinaryEntry(t *testing.T) {
	tree := &FileTree{}

	// Modified in both index and worktree
	line := "1 MM N... 100644 100644 100644 abc123 def456 src/main.go"
	entry := tree.parseOrdinaryEntry(line)
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Path != "src/main.go" {
		t.Errorf("expected path src/main.go, got %s", entry.Path)
	}
	if !entry.Staged {
		t.Error("expected Staged to be true")
	}
	if !entry.Unstaged {
		t.Error("expected Unstaged to be true")
	}

	// Only staged
	line = "1 M. N... 100644 100644 100644 abc123 def456 staged.go"
	entry = tree.parseOrdinaryEntry(line)
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if !entry.Staged {
		t.Error("expected Staged to be true")
	}
	if entry.Unstaged {
		t.Error("expected Unstaged to be false")
	}

	// Only unstaged
	line = "1 .M N... 100644 100644 100644 abc123 def456 unstaged.go"
	entry = tree.parseOrdinaryEntry(line)
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Staged {
		t.Error("expected Staged to be false")
	}
	if !entry.Unstaged {
		t.Error("expected Unstaged to be true")
	}
}

func TestDiffStats(t *testing.T) {
	ds := DiffStats{Additions: 10, Deletions: 5}
	if ds.Additions != 10 {
		t.Errorf("expected 10 additions, got %d", ds.Additions)
	}
	if ds.Deletions != 5 {
		t.Errorf("expected 5 deletions, got %d", ds.Deletions)
	}
}

func TestParseOrdinaryEntry_AllStatuses(t *testing.T) {
	tree := &FileTree{}

	tests := []struct {
		name       string
		line       string
		wantPath   string
		wantStatus FileStatus
		wantStaged bool
		wantUnstg  bool
	}{
		{
			name:       "added staged",
			line:       "1 A. N... 000000 100644 100644 0000000 abc123 new.go",
			wantPath:   "new.go",
			wantStatus: StatusAdded,
			wantStaged: true,
			wantUnstg:  false,
		},
		{
			name:       "deleted staged",
			line:       "1 D. N... 100644 000000 000000 abc123 0000000 deleted.go",
			wantPath:   "deleted.go",
			wantStatus: StatusDeleted,
			wantStaged: true,
			wantUnstg:  false,
		},
		{
			name:       "modified unstaged only",
			line:       "1 .M N... 100644 100644 100644 abc123 abc123 changed.go",
			wantPath:   "changed.go",
			wantStatus: StatusModified,
			wantStaged: false,
			wantUnstg:  true,
		},
		{
			name:       "deleted unstaged",
			line:       "1 .D N... 100644 100644 000000 abc123 abc123 gone.go",
			wantPath:   "gone.go",
			wantStatus: StatusDeleted,
			wantStaged: false,
			wantUnstg:  true,
		},
		{
			name:       "deep path",
			line:       "1 M. N... 100644 100644 100644 abc123 def456 internal/plugins/gitstatus/tree.go",
			wantPath:   "internal/plugins/gitstatus/tree.go",
			wantStatus: StatusModified,
			wantStaged: true,
			wantUnstg:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry := tree.parseOrdinaryEntry(tc.line)
			if entry == nil {
				t.Fatal("expected non-nil entry")
			}
			if entry.Path != tc.wantPath {
				t.Errorf("path = %q, want %q", entry.Path, tc.wantPath)
			}
			if entry.Status != tc.wantStatus {
				t.Errorf("status = %q, want %q", entry.Status, tc.wantStatus)
			}
			if entry.Staged != tc.wantStaged {
				t.Errorf("staged = %v, want %v", entry.Staged, tc.wantStaged)
			}
			if entry.Unstaged != tc.wantUnstg {
				t.Errorf("unstaged = %v, want %v", entry.Unstaged, tc.wantUnstg)
			}
		})
	}
}

func TestParseOrdinaryEntry_InvalidInput(t *testing.T) {
	tree := &FileTree{}

	tests := []struct {
		name string
		line string
	}{
		{"empty", ""},
		{"too few fields", "1 MM N... 100644"},
		{"wrong prefix", "? some/file.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry := tree.parseOrdinaryEntry(tc.line)
			if entry != nil {
				t.Errorf("expected nil for invalid input, got %+v", entry)
			}
		})
	}
}

func TestParseRenamedEntry(t *testing.T) {
	tree := &FileTree{}

	tests := []struct {
		name       string
		line       string
		wantPath   string
		wantStaged bool
		wantUnstg  bool
	}{
		{
			name:       "renamed only",
			line:       "2 R. N... 100644 100644 100644 abc123 def456 R100 new_name.go",
			wantPath:   "new_name.go",
			wantStaged: true,
			wantUnstg:  false,
		},
		{
			name:       "renamed with modifications",
			line:       "2 RM N... 100644 100644 100644 abc123 def456 R090 renamed_modified.go",
			wantPath:   "renamed_modified.go",
			wantStaged: true,
			wantUnstg:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry := tree.parseRenamedEntry(tc.line)
			if entry == nil {
				t.Fatal("expected non-nil entry")
			}
			if entry.Path != tc.wantPath {
				t.Errorf("path = %q, want %q", entry.Path, tc.wantPath)
			}
			if entry.Status != StatusRenamed {
				t.Errorf("status = %q, want %q", entry.Status, StatusRenamed)
			}
			if entry.Staged != tc.wantStaged {
				t.Errorf("staged = %v, want %v", entry.Staged, tc.wantStaged)
			}
			if entry.Unstaged != tc.wantUnstg {
				t.Errorf("unstaged = %v, want %v", entry.Unstaged, tc.wantUnstg)
			}
		})
	}
}

func TestParseRenamedEntry_InvalidInput(t *testing.T) {
	tree := &FileTree{}

	entry := tree.parseRenamedEntry("2 R. N... 100644")
	if entry != nil {
		t.Errorf("expected nil for invalid input, got %+v", entry)
	}
}

func TestParseUnmergedEntry(t *testing.T) {
	tree := &FileTree{}

	line := "u UU N... 100644 100644 100644 100644 abc123 def456 ghi789 conflict.go"
	entry := tree.parseUnmergedEntry(line)

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Path != "conflict.go" {
		t.Errorf("path = %q, want %q", entry.Path, "conflict.go")
	}
	if entry.Status != StatusUnmerged {
		t.Errorf("status = %q, want %q", entry.Status, StatusUnmerged)
	}
	if !entry.Unstaged {
		t.Error("unstaged should be true for unmerged")
	}
}

func TestParseUnmergedEntry_InvalidInput(t *testing.T) {
	tree := &FileTree{}

	entry := tree.parseUnmergedEntry("u UU N... 100644")
	if entry != nil {
		t.Errorf("expected nil for invalid input, got %+v", entry)
	}
}

func TestParseStatus(t *testing.T) {
	tree := &FileTree{}

	// Simulate git status --porcelain=v2 -z output (null-separated)
	output := []byte("1 M. N... 100644 100644 100644 abc def staged.go\x00" +
		"1 .M N... 100644 100644 100644 abc abc modified.go\x00" +
		"? untracked.go\x00")

	err := tree.parseStatus(output)
	if err != nil {
		t.Fatalf("parseStatus error: %v", err)
	}

	if len(tree.Staged) != 1 {
		t.Errorf("expected 1 staged, got %d", len(tree.Staged))
	}
	if len(tree.Modified) != 1 {
		t.Errorf("expected 1 modified, got %d", len(tree.Modified))
	}
	if len(tree.Untracked) != 1 {
		t.Errorf("expected 1 untracked, got %d", len(tree.Untracked))
	}

	if tree.Staged[0].Path != "staged.go" {
		t.Errorf("staged path = %q, want staged.go", tree.Staged[0].Path)
	}
	if tree.Modified[0].Path != "modified.go" {
		t.Errorf("modified path = %q, want modified.go", tree.Modified[0].Path)
	}
	if tree.Untracked[0].Path != "untracked.go" {
		t.Errorf("untracked path = %q, want untracked.go", tree.Untracked[0].Path)
	}
}

func TestParseStatus_BothStagedAndModified(t *testing.T) {
	tree := &FileTree{}

	// File with both staged and unstaged changes
	output := []byte("1 MM N... 100644 100644 100644 abc def both.go\x00")

	err := tree.parseStatus(output)
	if err != nil {
		t.Fatalf("parseStatus error: %v", err)
	}

	// Should appear in both staged and modified
	if len(tree.Staged) != 1 {
		t.Errorf("expected 1 staged, got %d", len(tree.Staged))
	}
	if len(tree.Modified) != 1 {
		t.Errorf("expected 1 modified, got %d", len(tree.Modified))
	}
	if tree.Staged[0].Path != "both.go" {
		t.Errorf("staged path = %q, want both.go", tree.Staged[0].Path)
	}
	if tree.Modified[0].Path != "both.go" {
		t.Errorf("modified path = %q, want both.go", tree.Modified[0].Path)
	}
}

func TestParseStatus_Empty(t *testing.T) {
	tree := &FileTree{}

	err := tree.parseStatus([]byte{})
	if err != nil {
		t.Fatalf("parseStatus error: %v", err)
	}

	if tree.TotalCount() != 0 {
		t.Errorf("expected 0 total, got %d", tree.TotalCount())
	}
	if tree.Summary() != "clean" {
		t.Errorf("expected 'clean', got %q", tree.Summary())
	}
}

func TestParseStatus_RenamedFile(t *testing.T) {
	tree := &FileTree{}

	// Renamed file: new path on line, old path follows after null
	output := []byte("2 R. N... 100644 100644 100644 abc def R100 newname.go\x00oldname.go\x00")

	err := tree.parseStatus(output)
	if err != nil {
		t.Fatalf("parseStatus error: %v", err)
	}

	if len(tree.Staged) != 1 {
		t.Fatalf("expected 1 staged, got %d", len(tree.Staged))
	}
	if tree.Staged[0].Path != "newname.go" {
		t.Errorf("path = %q, want newname.go", tree.Staged[0].Path)
	}
	if tree.Staged[0].OldPath != "oldname.go" {
		t.Errorf("oldPath = %q, want oldname.go", tree.Staged[0].OldPath)
	}
	if tree.Staged[0].Status != StatusRenamed {
		t.Errorf("status = %q, want R", tree.Staged[0].Status)
	}
}

func TestAddEntry(t *testing.T) {
	tree := &FileTree{}

	// Staged only
	tree.addEntry(&FileEntry{Path: "staged.go", Staged: true})
	if len(tree.Staged) != 1 {
		t.Errorf("expected 1 staged, got %d", len(tree.Staged))
	}

	// Unstaged only
	tree.addEntry(&FileEntry{Path: "unstaged.go", Unstaged: true})
	if len(tree.Modified) != 1 {
		t.Errorf("expected 1 modified, got %d", len(tree.Modified))
	}

	// Both staged and unstaged
	tree.addEntry(&FileEntry{Path: "both.go", Staged: true, Unstaged: true})
	if len(tree.Staged) != 2 {
		t.Errorf("expected 2 staged, got %d", len(tree.Staged))
	}
	if len(tree.Modified) != 2 {
		t.Errorf("expected 2 modified, got %d", len(tree.Modified))
	}
}

func TestTotalCount(t *testing.T) {
	tree := &FileTree{
		Staged:    []*FileEntry{{}, {}},
		Modified:  []*FileEntry{{}},
		Untracked: []*FileEntry{{}, {}, {}},
	}

	if tree.TotalCount() != 6 {
		t.Errorf("expected 6, got %d", tree.TotalCount())
	}
}
