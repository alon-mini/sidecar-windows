package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FillBackground ensures each line has a uniform background color.
// Inner styled elements emit ANSI resets (\x1b[0m) that clear all attributes
// including the parent container's background, leaving terminal-default black
// for the remainder of the line. We fix this by re-applying the background
// ANSI sequence after every reset, then padding short lines with
// background-colored spaces.
func FillBackground(content string, width int, bgColor lipgloss.Color) string {
	if width <= 0 {
		return content
	}
	bgSeq := BgANSISeqFor(bgColor)
	if bgSeq == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Re-apply background after every ANSI reset within the line
		line = strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+bgSeq)

		// Pad short lines to target width with background-colored spaces
		w := lipgloss.Width(line)
		if w < width {
			line += strings.Repeat(" ", width-w)
		}

		// Ensure clean reset at end of line
		if !strings.HasSuffix(line, "\x1b[0m") {
			line += "\x1b[0m"
		}

		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// BgANSISeqFor extracts the raw ANSI escape sequence for the given background
// color by rendering a marker character and taking everything before it.
func BgANSISeqFor(bgColor lipgloss.Color) string {
	const marker = "\x01"
	s := lipgloss.NewStyle().Background(bgColor).Render(marker)
	idx := strings.Index(s, marker)
	if idx > 0 {
		return s[:idx]
	}
	return ""
}
