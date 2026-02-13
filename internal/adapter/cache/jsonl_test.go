package cache

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScannerPool(t *testing.T) {
	// Get buffer
	buf := GetScannerBuffer()
	if len(buf) != DefaultScannerBufSize {
		t.Errorf("expected buffer size %d, got %d", DefaultScannerBufSize, len(buf))
	}

	// Put it back
	PutScannerBuffer(buf)

	// Get again (should reuse)
	buf2 := GetScannerBuffer()
	if len(buf2) != DefaultScannerBufSize {
		t.Errorf("expected buffer size %d, got %d", DefaultScannerBufSize, len(buf2))
	}
	PutScannerBuffer(buf2)
}

func TestNewScanner(t *testing.T) {
	content := "line1\nline2\nline3\n"
	reader := strings.NewReader(content)

	scanner, buf := NewScanner(reader)
	defer PutScannerBuffer(buf)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestIncrementalReader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Read from start
	r, err := NewIncrementalReader(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	}()

	line, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if string(line) != "line1" {
		t.Errorf("expected line1, got %s", line)
	}

	// Check offset tracking
	if r.Offset() != 6 { // "line1" + newline
		t.Errorf("expected offset 6, got %d", r.Offset())
	}

	line, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if string(line) != "line2" {
		t.Errorf("expected line2, got %s", line)
	}
}

func TestIncrementalReader_FromOffset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Read from offset 6 (after "line1\n")
	r, err := NewIncrementalReader(path, 6)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	}()

	line, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if string(line) != "line2" {
		t.Errorf("expected line2, got %s", line)
	}
}

func TestIncrementalReader_EOF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	if err := os.WriteFile(path, []byte("line1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := NewIncrementalReader(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	}()

	_, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.Next()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestTailReader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Read last 12 bytes (covers "line4\nline5\n")
	r, err := NewTailReader(path, 12)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	}()

	// First line should skip the partial line at seek point
	line, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	// Should get "line5" (skipped partial "line4")
	if string(line) != "line5" {
		t.Errorf("expected line5, got %s", line)
	}

	// EOF
	_, err = r.Next()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestTailReader_SmallFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	content := "line1\nline2\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Request more than file size
	r, err := NewTailReader(path, 1000)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	}()

	// Should read from start, no skip
	line, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if string(line) != "line1" {
		t.Errorf("expected line1, got %s", line)
	}

	line, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if string(line) != "line2" {
		t.Errorf("expected line2, got %s", line)
	}
}

func TestHeadReader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Read only first 3 lines
	r, err := NewHeadReader(path, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	}()

	var lines []string
	for {
		line, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, string(line))
	}

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Errorf("unexpected lines: %v", lines)
	}

	if r.LinesRead() != 3 {
		t.Errorf("expected LinesRead() = 3, got %d", r.LinesRead())
	}
}

func TestHeadReader_Offset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	content := "line1\nline2\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := NewHeadReader(path, 100)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	}()

	_, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if r.Offset() != 6 { // "line1" + newline
		t.Errorf("expected offset 6, got %d", r.Offset())
	}

	_, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if r.Offset() != 12 { // both lines
		t.Errorf("expected offset 12, got %d", r.Offset())
	}
}

func TestHeadReader_FewerLinesThanMax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	content := "line1\nline2\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Request more lines than exist
	r, err := NewHeadReader(path, 100)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	}()

	var count int
	for {
		_, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		count++
	}

	if count != 2 {
		t.Errorf("expected 2 lines, got %d", count)
	}
}
