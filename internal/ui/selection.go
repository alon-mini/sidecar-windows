package ui

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/marcus/sidecar/internal/mouse"
)

// SelectionPoint represents a position as (line, col) in visual space.
// Col is post-tab-expansion, accounting for multi-width chars.
// EndCol convention: INCLUSIVE — the character under the cursor IS selected.
type SelectionPoint struct {
	Line int // buffer line index, -1 = unset
	Col  int // visual column, -1 = unset
}

// Before returns true if p is before other in document order.
func (p SelectionPoint) Before(other SelectionPoint) bool {
	return p.Line < other.Line || (p.Line == other.Line && p.Col < other.Col)
}

// Valid returns true if the point has been set.
func (p SelectionPoint) Valid() bool {
	return p.Line >= 0 && p.Col >= 0
}

// SelectionState holds all selection state for a single selectable region.
type SelectionState struct {
	Active   bool
	Start    SelectionPoint
	End      SelectionPoint
	Anchor   SelectionPoint
	ViewRect mouse.Rect
}

// Clear resets all fields, Start/End/Anchor to {-1,-1}.
func (s *SelectionState) Clear() {
	s.Active = false
	s.Start = SelectionPoint{-1, -1}
	s.End = SelectionPoint{-1, -1}
	s.Anchor = SelectionPoint{-1, -1}
	s.ViewRect = mouse.Rect{}
}

// HasSelection returns true if both Start and End are valid.
func (s *SelectionState) HasSelection() bool {
	return s.Start.Valid() && s.End.Valid()
}

// IsLineSelected checks if line is in [Start.Line, End.Line].
func (s *SelectionState) IsLineSelected(lineIdx int) bool {
	if !s.HasSelection() {
		return false
	}
	start := s.Start.Line
	end := s.End.Line
	if start > end {
		start, end = end, start
	}
	return lineIdx >= start && lineIdx <= end
}

// GetLineSelectionCols returns the visual column range [start, end] inclusive
// for the given line. Returns (-1,-1) if not selected.
// endCol==-1 means "to end of line".
// Logic: single-line returns both cols; first line of multi returns (startCol, -1);
// last line returns (0, endCol); middle returns (0, -1).
func (s *SelectionState) GetLineSelectionCols(lineIdx int) (startCol, endCol int) {
	if !s.HasSelection() {
		return -1, -1
	}

	selStart := s.Start
	selEnd := s.End

	if lineIdx < selStart.Line || lineIdx > selEnd.Line {
		return -1, -1
	}

	if selStart.Line == selEnd.Line {
		// Single-line selection
		return selStart.Col, selEnd.Col
	} else if lineIdx == selStart.Line {
		// First line: from startCol to end of line
		return selStart.Col, -1
	} else if lineIdx == selEnd.Line {
		// Last line: from start to endCol (inclusive)
		return 0, selEnd.Col
	}
	// Middle line: entire line
	return 0, -1
}

// PrepareDrag stores the click position and starts drag tracking without
// initializing selection. Selection only activates on actual drag motion.
func (s *SelectionState) PrepareDrag(lineIdx, col int, viewRect mouse.Rect) {
	s.Active = false
	s.Start = SelectionPoint{-1, -1}
	s.End = SelectionPoint{-1, -1}
	s.Anchor = SelectionPoint{lineIdx, col}
	s.ViewRect = viewRect
}

// HandleDrag updates selection state during a drag operation.
// If start not valid, initialize start=end=anchor. Set Active=true.
// Order start/end by document position relative to anchor.
func (s *SelectionState) HandleDrag(lineIdx, col int) {
	current := SelectionPoint{lineIdx, col}
	anchor := s.Anchor

	// First drag motion: activate selection
	if !s.Start.Valid() {
		s.Start = anchor
		s.End = anchor
	}

	s.Active = true

	// Order start/end by document position
	if current.Before(anchor) {
		s.Start = current
		s.End = anchor
	} else {
		s.Start = anchor
		s.End = current
	}
}

// FinishDrag completes a drag operation. If start not valid (no drag motion),
// calls Clear(). Otherwise sets Active=false.
func (s *SelectionState) FinishDrag() {
	if !s.Start.Valid() {
		// No drag occurred — click without motion, clear state
		s.Clear()
		return
	}
	s.Active = false
}

// SelectedText takes raw lines in the selected range starting at startLine index,
// expands tabs, extracts with column precision using VisualSubstring.
// For single-line: extract [startCol, endCol+1).
// For multi-line: first line from startCol to end, middle lines ANSI-stripped,
// last line from 0 to endCol+1.
func (s *SelectionState) SelectedText(lines []string, startLine int, tabWidth int) []string {
	if !s.HasSelection() || len(lines) == 0 {
		return nil
	}

	// Expand tabs to match visual columns
	expanded := make([]string, len(lines))
	for i := range lines {
		expanded[i] = ExpandTabs(lines[i], tabWidth)
	}

	startCol := s.Start.Col
	endCol := s.End.Col

	if len(expanded) == 1 {
		// Single line: extract [startCol, endCol] inclusive
		expanded[0] = VisualSubstring(expanded[0], startCol, endCol+1)
	} else {
		// First line: from startCol to end
		expanded[0] = VisualSubstring(expanded[0], startCol, -1)
		// Last line: from start to endCol (inclusive)
		lastIdx := len(expanded) - 1
		expanded[lastIdx] = VisualSubstring(expanded[lastIdx], 0, endCol+1)
		// Middle lines: full text, ANSI stripped
		for i := 1; i < lastIdx; i++ {
			expanded[i] = ansi.Strip(expanded[i])
		}
	}

	return expanded
}
