package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/marcus/sidecar/internal/styles"
)

// AnsiResetRe matches ANSI reset sequences (both \x1b[0m and \x1b[m).
var AnsiResetRe = regexp.MustCompile(`\x1b\[0?m`)

// GetSelectionBgANSI returns the ANSI 24-bit background code for selection highlight
// based on the current theme's BgTertiary color.
func GetSelectionBgANSI() string {
	theme := styles.GetCurrentTheme()
	hex := theme.Colors.BgTertiary
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		r, g, b = 55, 65, 81
	}
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

// InjectSelectionBackground adds a selection background while preserving ANSI resets.
// It prepends the background at the start and re-injects after any reset sequences.
func InjectSelectionBackground(s string) string {
	selectionBg := GetSelectionBgANSI()
	result := selectionBg + s
	result = AnsiResetRe.ReplaceAllString(result, "${0}"+selectionBg)
	return result + "\x1b[0m"
}

// ExpandTabs replaces tabs with spaces, preserving ANSI sequences and column widths.
func ExpandTabs(line string, tabWidth int) string {
	if tabWidth <= 0 || !strings.Contains(line, "\t") {
		return line
	}

	var sb strings.Builder
	sb.Grow(len(line))

	state := ansi.NormalState
	column := 0
	for len(line) > 0 {
		seq, width, n, newState := ansi.GraphemeWidth.DecodeSequenceInString(line, state, nil)
		if n <= 0 {
			sb.WriteString(line)
			break
		}
		if seq == "\t" && width == 0 {
			spaces := tabWidth - (column % tabWidth)
			if spaces == 0 {
				spaces = tabWidth
			}
			sb.WriteString(strings.Repeat(" ", spaces))
			column += spaces
		} else {
			sb.WriteString(seq)
			column += width
		}
		state = newState
		line = line[n:]
	}

	return sb.String()
}

// VisualSubstring extracts a substring by visual column range [startCol, endCol).
// endCol is EXCLUSIVE (one past last included column).
// Handles ANSI escape codes (skipped in column counting).
// If endCol is -1, extracts to end of string.
// Returns plain text (ANSI stripped) for clipboard use.
func VisualSubstring(s string, startCol, endCol int) string {
	if s == "" {
		return ""
	}

	var sb strings.Builder
	state := ansi.NormalState
	cumWidth := 0

	remaining := s
	for len(remaining) > 0 {
		seq, width, n, newState := ansi.GraphemeWidth.DecodeSequenceInString(remaining, state, nil)
		if n <= 0 {
			break
		}
		if width > 0 {
			charStart := cumWidth
			charEnd := cumWidth + width
			cumWidth = charEnd

			// Check if this character is within range
			inRange := false
			if endCol == -1 {
				inRange = charEnd > startCol
			} else {
				inRange = charStart < endCol && charEnd > startCol
			}
			if inRange {
				sb.WriteString(seq)
			}
			if endCol >= 0 && cumWidth >= endCol {
				break
			}
		}
		// Skip ANSI sequences (width == 0, not a visible character)
		state = newState
		remaining = remaining[n:]
	}

	return sb.String()
}

// InjectCharacterRangeBackground applies selection background to visual columns
// [startCol, endCol] (inclusive) within the line. startCol and endCol are in
// absolute visual space (post-tab-expansion). Handles ANSI codes correctly.
// If endCol is -1, highlights to end of line.
func InjectCharacterRangeBackground(line string, startCol, endCol int) string {
	if startCol == 0 && endCol == -1 {
		return InjectSelectionBackground(line)
	}

	selBg := GetSelectionBgANSI()
	var sb strings.Builder
	sb.Grow(len(line) + 64)

	state := ansi.NormalState
	cumWidth := 0
	inSelection := false

	remaining := line
	for len(remaining) > 0 {
		seq, width, n, newState := ansi.GraphemeWidth.DecodeSequenceInString(remaining, state, nil)
		if n <= 0 {
			sb.WriteString(remaining)
			break
		}

		if width > 0 {
			// Visible character
			charInRange := false
			if endCol == -1 {
				charInRange = cumWidth >= startCol
			} else {
				charInRange = cumWidth >= startCol && cumWidth <= endCol
			}

			if charInRange && !inSelection {
				sb.WriteString(selBg)
				inSelection = true
			} else if !charInRange && inSelection {
				sb.WriteString("\x1b[49m") // reset background only, preserve foreground
				inSelection = false
			}

			sb.WriteString(seq)
			cumWidth += width

			// Check if we've passed the end of selection
			if endCol >= 0 && cumWidth > endCol && inSelection {
				sb.WriteString("\x1b[49m") // reset background only, preserve foreground
				inSelection = false
			}
		} else {
			// ANSI sequence or control character
			sb.WriteString(seq)
			// If there's a reset within the selection, re-inject background
			if inSelection && AnsiResetRe.MatchString(seq) {
				sb.WriteString(selBg)
			}
		}

		state = newState
		remaining = remaining[n:]
	}

	if inSelection {
		sb.WriteString("\x1b[49m") // reset background only at end of line
	}

	return sb.String()
}

// VisualColAtRelativeX takes an already-expanded line and a relative X offset,
// walks graphemes using ansi.GraphemeWidth.DecodeSequenceInString, snaps to
// character boundaries, and clamps to last char if beyond end.
func VisualColAtRelativeX(expandedLine string, relX int) int {
	if relX < 0 {
		return 0
	}

	visualCol := relX

	// Walk expanded line grapheme-by-grapheme to find column
	state := ansi.NormalState
	cumWidth := 0
	lastCharCol := 0
	hasChars := false

	remaining := expandedLine
	for len(remaining) > 0 {
		seq, width, n, newState := ansi.GraphemeWidth.DecodeSequenceInString(remaining, state, nil)
		if n <= 0 {
			break
		}
		_ = seq
		if width > 0 {
			hasChars = true
			// If visualCol lands within this character's cells
			if visualCol >= cumWidth && visualCol < cumWidth+width {
				return cumWidth // snap to start of character
			}
			lastCharCol = cumWidth
			cumWidth += width
		}
		state = newState
		remaining = remaining[n:]
	}

	if !hasChars {
		return 0
	}

	// Beyond line end: clamp to last character column (inclusive)
	if visualCol >= cumWidth {
		return lastCharCol
	}
	return visualCol
}
