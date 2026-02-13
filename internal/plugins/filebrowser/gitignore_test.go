package filebrowser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		glob     string
		anchored bool
		input    string
		match    bool
	}{
		{"*.go", false, "main.go", true},
		{"*.go", false, "foo/bar.go", true},
		{"*.go", false, "main.txt", false},
		{"node_modules", false, "node_modules", true},
		{"node_modules", false, "foo/node_modules", true},
		{"build/", false, "build", true},
		{"src/*.js", true, "src/app.js", true},
		{"src/*.js", true, "lib/app.js", false},
		{"**/*.log", false, "foo/bar/baz.log", true},
		{"?.txt", false, "a.txt", true},
		{"?.txt", false, "ab.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.glob+"_"+tt.input, func(t *testing.T) {
			regex := globToRegex(tt.glob, tt.anchored)
			gi := NewGitIgnore()
			gi.addPattern(tt.glob)

			if len(gi.patterns) == 0 {
				t.Fatal("pattern not added")
			}

			result := gi.patterns[0].regex.MatchString(tt.input)
			if result != tt.match {
				t.Errorf("globToRegex(%q, %v) matching %q = %v, want %v (regex: %s)",
					tt.glob, tt.anchored, tt.input, result, tt.match, regex)
			}
		})
	}
}

func TestGitIgnore_IsIgnored(t *testing.T) {
	gi := NewGitIgnore()
	gi.addPattern("*.log")
	gi.addPattern("node_modules/")
	gi.addPattern("!important.log")
	gi.addPattern("build/")

	tests := []struct {
		path    string
		isDir   bool
		ignored bool
	}{
		{"debug.log", false, true},
		{"foo/error.log", false, true},
		{"important.log", false, false}, // negated
		{"main.go", false, false},
		{"node_modules", true, true},
		{"build", true, true},
		{"build", false, false}, // dir-only pattern
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := gi.IsIgnored(tt.path, tt.isDir)
			if result != tt.ignored {
				t.Errorf("IsIgnored(%q, %v) = %v, want %v", tt.path, tt.isDir, result, tt.ignored)
			}
		})
	}
}

func TestGitIgnore_LoadFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .gitignore
	gitignore := `# Comment
*.log
node_modules/

# Another comment
!keep.log
build/
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatal(err)
	}

	gi := NewGitIgnore()
	if err := gi.LoadFile(filepath.Join(tmpDir, ".gitignore")); err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	// Should have 4 patterns (comments excluded)
	if len(gi.patterns) != 4 {
		t.Errorf("patterns = %d, want 4", len(gi.patterns))
	}

	// Test loaded patterns work
	if !gi.IsIgnored("debug.log", false) {
		t.Error("expected debug.log to be ignored")
	}
	if gi.IsIgnored("keep.log", false) {
		t.Error("expected keep.log to NOT be ignored (negated)")
	}
	if !gi.IsIgnored("node_modules", true) {
		t.Error("expected node_modules dir to be ignored")
	}
}

func TestGitIgnore_LoadFile_NotFound(t *testing.T) {
	gi := NewGitIgnore()
	err := gi.LoadFile("/nonexistent/.gitignore")
	if err != nil {
		t.Errorf("LoadFile for non-existent file should return nil, got %v", err)
	}
}

func TestGitIgnore_Cache(t *testing.T) {
	gi := NewGitIgnore()
	gi.addPattern("*.log")

	// First call should compute
	_ = gi.IsIgnored("test.log", false)
	if len(gi.cache) != 1 {
		t.Error("expected cache to have 1 entry")
	}

	// Clear and verify
	gi.ClearCache()
	if len(gi.cache) != 0 {
		t.Error("expected cache to be empty after clear")
	}
}
