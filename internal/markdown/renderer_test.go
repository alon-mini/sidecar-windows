package markdown

import (
	"strings"
	"sync"
	"testing"
)

func TestRenderer_Basic(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error: %v", err)
	}

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "header",
			content: "# Header\n\nSome text",
		},
		{
			name:    "list",
			content: "- Item 1\n- Item 2\n- Item 3",
		},
		{
			name:    "code block",
			content: "```go\nfunc main() {}\n```",
		},
		{
			name:    "mixed",
			content: "# Title\n\nParagraph with **bold** and *italic*.\n\n- List item\n\n```\ncode\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := r.RenderContent(tt.content, 80)
			if len(lines) == 0 {
				t.Error("RenderContent returned empty output")
			}
			// Verify content is present (stripped of ANSI)
			joined := strings.Join(lines, "\n")
			if len(joined) == 0 {
				t.Error("Joined output is empty")
			}
		})
	}
}

func TestRenderer_WidthChange(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error: %v", err)
	}

	content := "This is a test paragraph that should wrap differently at different widths."

	lines80 := r.RenderContent(content, 80)
	lines40 := r.RenderContent(content, 40)

	// At minimum, both should produce output
	if len(lines80) == 0 || len(lines40) == 0 {
		t.Error("Width change produced empty output")
	}
}

func TestRenderer_NarrowFallback(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error: %v", err)
	}

	content := "# Header\n\nSome content"

	// Width < 30 should use plain WrapText fallback
	lines := r.RenderContent(content, 20)

	// Should have output
	if len(lines) == 0 {
		t.Error("Narrow fallback returned empty output")
	}
}

func TestRenderer_Cache(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error: %v", err)
	}

	content := "# Test\n\nContent here"

	// First call
	lines1 := r.RenderContent(content, 80)

	// Second call with same content and width
	lines2 := r.RenderContent(content, 80)

	// Results should be identical
	if len(lines1) != len(lines2) {
		t.Errorf("Cache miss: different line counts %d vs %d", len(lines1), len(lines2))
	}

	for i := range lines1 {
		if i < len(lines2) && lines1[i] != lines2[i] {
			t.Errorf("Cache miss: line %d differs", i)
			break
		}
	}
}

func TestRenderer_CacheEviction(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error: %v", err)
	}

	// Render 101 unique strings to trigger eviction
	for i := 0; i < 101; i++ {
		content := strings.Repeat("x", i+50) // Unique content
		lines := r.RenderContent(content, 80)
		if lines == nil {
			t.Fatalf("RenderContent returned nil at iteration %d", i)
		}
	}

	// Should still work after eviction
	lines := r.RenderContent("# Final test", 80)
	if len(lines) == 0 {
		t.Error("RenderContent failed after cache eviction")
	}
}

func TestRenderer_EmptyContent(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error: %v", err)
	}

	lines := r.RenderContent("", 80)

	// Should return empty slice, not panic
	if lines == nil {
		t.Error("RenderContent returned nil for empty content")
	}
	if len(lines) != 0 {
		t.Errorf("RenderContent returned %d lines for empty content, expected 0", len(lines))
	}
}

func TestRenderer_Concurrent(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error: %v", err)
	}

	var wg sync.WaitGroup
	contents := []string{
		"# Header 1\n\nContent",
		"# Header 2\n\n- List",
		"```go\ncode\n```",
		"Plain text",
	}

	// Run multiple goroutines concurrently
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			content := contents[idx%len(contents)]
			width := 60 + (idx%3)*20 // 60, 80, or 100
			lines := r.RenderContent(content, width)
			if len(lines) == 0 {
				t.Errorf("Concurrent render %d returned empty", idx)
			}
		}(i)
	}

	wg.Wait()
}

func TestRenderer_WidthZero(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error: %v", err)
	}

	// Width 0 should fall back to WrapText behavior
	lines := r.RenderContent("Some content", 0)

	// Should not panic and should return something
	if lines == nil {
		t.Error("Width 0 returned nil")
	}
}

func TestWrapText_Basic(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		wantMin  int // minimum expected lines
	}{
		{"short", "Hello", 80, 1},
		{"wraps", "Hello world this is a test", 10, 2},
		{"zero width", "Test", 0, 1},
		{"empty", "", 80, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := WrapText(tt.text, tt.maxWidth)
			if len(lines) < tt.wantMin {
				t.Errorf("WrapText() got %d lines, want >= %d", len(lines), tt.wantMin)
			}
		})
	}
}
