package ui

import (
	"regexp"
	"strings"
)

// Standard modal widths
const (
	ModalWidthSmall  = 40 // Simple confirmations
	ModalWidthMedium = 50 // Standard modals with inputs
	ModalWidthLarge  = 60 // Modals with longer content
)

// ansiPattern matches ANSI escape codes
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// CalculateModalWidth returns an appropriate width based on content.
// Ensures width is within min/max bounds.
func CalculateModalWidth(content string, minWidth, maxWidth int) int {
	// Find longest line in content
	maxLineLen := 0
	for _, line := range strings.Split(content, "\n") {
		// Strip ANSI for accurate measurement
		stripped := ansiPattern.ReplaceAllString(line, "")
		if len(stripped) > maxLineLen {
			maxLineLen = len(stripped)
		}
	}

	// Add padding for modal borders/padding (typically 6 chars)
	width := maxLineLen + 6

	// Clamp to bounds
	if width < minWidth {
		return minWidth
	}
	if width > maxWidth {
		return maxWidth
	}
	return width
}

// ClampModalWidth clamps a width value between min and max screen bounds.
func ClampModalWidth(width, screenWidth int) int {
	maxAllowed := screenWidth - 10
	if width > maxAllowed {
		return maxAllowed
	}
	return width
}
