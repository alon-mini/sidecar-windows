package filebrowser

import (
	"testing"

	"github.com/marcus/sidecar/internal/markdown"
)

func TestIsMarkdownFile(t *testing.T) {
	tests := []struct {
		name        string
		previewFile string
		want        bool
	}{
		{"md extension", "README.md", true},
		{"markdown extension", "docs/guide.markdown", true},
		{"uppercase MD", "test.MD", true},
		{"mixed case", "Test.Md", true},
		{"go file", "main.go", false},
		{"txt file", "notes.txt", false},
		{"empty path", "", false},
		{"no extension", "README", false},
		{"md in path but not extension", "docs/md/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{
				previewFile: tt.previewFile,
			}
			if got := p.isMarkdownFile(); got != tt.want {
				t.Errorf("isMarkdownFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToggleMarkdownRender(t *testing.T) {
	t.Run("toggles mode for markdown file", func(t *testing.T) {
		renderer, _ := markdown.NewRenderer()
		p := &Plugin{
			previewFile:      "README.md",
			previewLines:     []string{"# Hello", "", "World"},
			markdownRenderer: renderer,
			previewWidth:     80,
		}

		// Initially off
		if p.markdownRenderMode {
			t.Error("markdownRenderMode should start false")
		}

		// Toggle on
		p.toggleMarkdownRender()
		if !p.markdownRenderMode {
			t.Error("markdownRenderMode should be true after toggle")
		}

		// Should have rendered content
		if len(p.markdownRendered) == 0 {
			t.Error("markdownRendered should have content after toggle on")
		}

		// Toggle off
		p.toggleMarkdownRender()
		if p.markdownRenderMode {
			t.Error("markdownRenderMode should be false after second toggle")
		}
	})

	t.Run("no-op for non-markdown file", func(t *testing.T) {
		p := &Plugin{
			previewFile: "main.go",
		}

		p.toggleMarkdownRender()
		if p.markdownRenderMode {
			t.Error("markdownRenderMode should remain false for non-markdown file")
		}
	})

	t.Run("no-op for empty preview file", func(t *testing.T) {
		p := &Plugin{
			previewFile: "",
		}

		p.toggleMarkdownRender()
		if p.markdownRenderMode {
			t.Error("markdownRenderMode should remain false for empty file")
		}
	})
}

func TestRenderMarkdownContent(t *testing.T) {
	t.Run("renders content with renderer", func(t *testing.T) {
		renderer, _ := markdown.NewRenderer()
		p := &Plugin{
			previewFile:      "test.md",
			previewLines:     []string{"# Header", "", "Some text here"},
			markdownRenderer: renderer,
			previewWidth:     80,
		}

		p.renderMarkdownContent()

		if len(p.markdownRendered) == 0 {
			t.Error("markdownRendered should have content")
		}
	})

	t.Run("safe with nil renderer", func(t *testing.T) {
		p := &Plugin{
			previewFile:      "test.md",
			previewLines:     []string{"# Header"},
			markdownRenderer: nil,
			previewWidth:     80,
		}

		// Should not panic
		p.renderMarkdownContent()

		if len(p.markdownRendered) != 0 {
			t.Error("markdownRendered should be empty with nil renderer")
		}
	})

	t.Run("safe with empty preview lines", func(t *testing.T) {
		renderer, _ := markdown.NewRenderer()
		p := &Plugin{
			previewFile:      "test.md",
			previewLines:     []string{},
			markdownRenderer: renderer,
			previewWidth:     80,
		}

		// Should not panic
		p.renderMarkdownContent()
	})

	t.Run("respects width for rendering", func(t *testing.T) {
		renderer, _ := markdown.NewRenderer()
		content := []string{"This is a very long line that should wrap when the width is narrow enough to cause wrapping behavior"}

		p40 := &Plugin{
			previewFile:      "test.md",
			previewLines:     content,
			markdownRenderer: renderer,
			previewWidth:     46, // 46-6=40 effective width
		}
		p40.renderMarkdownContent()

		p100 := &Plugin{
			previewFile:      "test.md",
			previewLines:     content,
			markdownRenderer: renderer,
			previewWidth:     106, // 106-6=100 effective width
		}
		p100.renderMarkdownContent()

		// Narrower width should produce more lines (or equal)
		if len(p40.markdownRendered) < len(p100.markdownRendered) {
			t.Errorf("narrow width produced fewer lines: %d vs %d",
				len(p40.markdownRendered), len(p100.markdownRendered))
		}
	})
}
