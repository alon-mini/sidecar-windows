package gitstatus

import (
	"testing"
)

func TestParseUnifiedDiff_BasicDiff(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 package main
+
 func foo() {
 }
`

	parsed, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.OldFile != "file.go" {
		t.Errorf("OldFile = %q, want %q", parsed.OldFile, "file.go")
	}
	if parsed.NewFile != "file.go" {
		t.Errorf("NewFile = %q, want %q", parsed.NewFile, "file.go")
	}
	if len(parsed.Hunks) != 1 {
		t.Fatalf("len(Hunks) = %d, want 1", len(parsed.Hunks))
	}

	hunk := parsed.Hunks[0]
	if hunk.OldStart != 1 || hunk.OldCount != 3 {
		t.Errorf("old range = %d,%d, want 1,3", hunk.OldStart, hunk.OldCount)
	}
	if hunk.NewStart != 1 || hunk.NewCount != 4 {
		t.Errorf("new range = %d,%d, want 1,4", hunk.NewStart, hunk.NewCount)
	}
}

func TestParseUnifiedDiff_MultipleHunks(t *testing.T) {
	diff := `--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line1
-old
+new
 line3
@@ -10,3 +10,4 @@
 line10
+added
 line11
 line12
`

	parsed, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Hunks) != 2 {
		t.Fatalf("len(Hunks) = %d, want 2", len(parsed.Hunks))
	}

	if parsed.Hunks[0].OldStart != 1 {
		t.Errorf("first hunk OldStart = %d, want 1", parsed.Hunks[0].OldStart)
	}
	if parsed.Hunks[1].OldStart != 10 {
		t.Errorf("second hunk OldStart = %d, want 10", parsed.Hunks[1].OldStart)
	}
}

func TestParseUnifiedDiff_BinaryFile(t *testing.T) {
	diff := `Binary files a/image.png and b/image.png differ`

	parsed, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !parsed.Binary {
		t.Error("expected Binary = true")
	}
}

func TestParseUnifiedDiff_LineTypes(t *testing.T) {
	// Note: no trailing newline to avoid empty context line
	diff := `--- a/file.txt
+++ b/file.txt
@@ -1,4 +1,4 @@
 context
-removed
+added
 more context`

	parsed, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Hunks) != 1 {
		t.Fatalf("len(Hunks) = %d, want 1", len(parsed.Hunks))
	}

	lines := parsed.Hunks[0].Lines
	if len(lines) != 4 {
		t.Fatalf("len(Lines) = %d, want 4", len(lines))
	}

	// Check line types
	if lines[0].Type != LineContext {
		t.Errorf("line 0 type = %v, want LineContext", lines[0].Type)
	}
	if lines[1].Type != LineRemove {
		t.Errorf("line 1 type = %v, want LineRemove", lines[1].Type)
	}
	if lines[2].Type != LineAdd {
		t.Errorf("line 2 type = %v, want LineAdd", lines[2].Type)
	}
	if lines[3].Type != LineContext {
		t.Errorf("line 3 type = %v, want LineContext", lines[3].Type)
	}
}

func TestParseUnifiedDiff_LineNumbers(t *testing.T) {
	diff := `--- a/file.txt
+++ b/file.txt
@@ -5,4 +5,4 @@
 context
-removed
+added
 more context
`

	parsed, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := parsed.Hunks[0].Lines

	// Context line: both line numbers
	if lines[0].OldLineNo != 5 || lines[0].NewLineNo != 5 {
		t.Errorf("context line numbers = %d,%d, want 5,5", lines[0].OldLineNo, lines[0].NewLineNo)
	}

	// Removed line: only old line number
	if lines[1].OldLineNo != 6 || lines[1].NewLineNo != 0 {
		t.Errorf("removed line numbers = %d,%d, want 6,0", lines[1].OldLineNo, lines[1].NewLineNo)
	}

	// Added line: only new line number
	if lines[2].OldLineNo != 0 || lines[2].NewLineNo != 6 {
		t.Errorf("added line numbers = %d,%d, want 0,6", lines[2].OldLineNo, lines[2].NewLineNo)
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"hello world", []string{"hello", " ", "world"}},
		{"a  b", []string{"a", "  ", "b"}},
		{"\tfoo", []string{"\t", "foo"}},
		{"", nil},
		{"   ", []string{"   "}},
		{"word", []string{"word"}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := tokenize(tc.input)
			if len(got) != len(tc.want) {
				t.Errorf("len = %d, want %d", len(got), len(tc.want))
				t.Errorf("got: %v", got)
				return
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("token[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParsedDiff_TotalLines(t *testing.T) {
	// No trailing newline
	diff := `--- a/file.txt
+++ b/file.txt
@@ -1,2 +1,2 @@
 context
-old
+new`

	parsed, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 content lines + 1 hunk header
	if parsed.TotalLines() != 4 {
		t.Errorf("TotalLines() = %d, want 4", parsed.TotalLines())
	}
}

func TestParsedDiff_MaxLineNumber(t *testing.T) {
	// No trailing newline
	diff := `--- a/file.txt
+++ b/file.txt
@@ -100,2 +100,3 @@
 context
+added
 more`

	parsed, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Max line number should be 102 (100 + 2 context lines)
	max := parsed.MaxLineNumber()
	if max != 102 {
		t.Errorf("MaxLineNumber() = %d, want 102", max)
	}
}
