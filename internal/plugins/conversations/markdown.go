package conversations

import "github.com/marcus/sidecar/internal/markdown"

// GlamourRenderer is an alias to the shared markdown renderer.
type GlamourRenderer = markdown.Renderer

// NewGlamourRenderer creates a new renderer instance.
func NewGlamourRenderer() (*GlamourRenderer, error) {
	return markdown.NewRenderer()
}

// wrapText wraps text to fit within maxWidth.
func wrapText(text string, maxWidth int) []string {
	return markdown.WrapText(text, maxWidth)
}
