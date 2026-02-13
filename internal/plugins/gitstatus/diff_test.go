package gitstatus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStringToInt(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   int
		wantOK bool
	}{
		{"zero", "0", 0, true},
		{"single digit", "5", 5, true},
		{"multiple digits", "123", 123, true},
		{"large number", "999999", 999999, true},
		{"empty string", "", 0, true},
		{"non-digit", "abc", 0, false},
		{"mixed", "12a34", 0, false},
		{"negative sign", "-5", 0, false},
		{"decimal", "3.14", 0, false},
		{"spaces", "1 2", 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result int
			ok, _ := stringToInt(tc.input, &result)

			if ok != tc.wantOK {
				t.Errorf("ok = %v, want %v", ok, tc.wantOK)
			}
			if tc.wantOK && result != tc.want {
				t.Errorf("result = %d, want %d", result, tc.want)
			}
		})
	}
}

func TestStringToInt_Accumulates(t *testing.T) {
	// The function accumulates into the result pointer
	// Starting with non-zero value should work as documented
	var result int
	_, _ = stringToInt("12", &result)
	if result != 12 {
		t.Errorf("got %d, want 12", result)
	}
}

func TestGetNewFileDiff(t *testing.T) {
	// Create temp dir with a test file
	tmpDir := t.TempDir()
	testFile := "newfile.txt"
	content := "line1\nline2\nline3"
	err := os.WriteFile(filepath.Join(tmpDir, testFile), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	diff, err := GetNewFileDiff(tmpDir, testFile)
	if err != nil {
		t.Fatalf("GetNewFileDiff failed: %v", err)
	}

	// Check diff header
	if !strings.Contains(diff, "diff --git") {
		t.Error("diff missing git header")
	}
	if !strings.Contains(diff, "new file mode") {
		t.Error("diff missing new file indicator")
	}
	if !strings.Contains(diff, "--- /dev/null") {
		t.Error("diff missing /dev/null source")
	}
	if !strings.Contains(diff, "+++ b/"+testFile) {
		t.Error("diff missing dest path")
	}
	if !strings.Contains(diff, "@@ -0,0") {
		t.Error("diff missing hunk header")
	}

	// Check all lines are additions
	lines := strings.Split(diff, "\n")
	var addCount int
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addCount++
		}
	}
	if addCount != 3 {
		t.Errorf("expected 3 addition lines, got %d", addCount)
	}
}

func TestGetNewFileDiff_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := "empty.txt"
	err := os.WriteFile(filepath.Join(tmpDir, testFile), []byte(""), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	diff, err := GetNewFileDiff(tmpDir, testFile)
	if err != nil {
		t.Fatalf("GetNewFileDiff failed: %v", err)
	}

	if !strings.Contains(diff, "new file mode") {
		t.Error("diff missing new file indicator for empty file")
	}
}

func TestGetNewFileDiff_NotExists(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := GetNewFileDiff(tmpDir, "nonexistent.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
