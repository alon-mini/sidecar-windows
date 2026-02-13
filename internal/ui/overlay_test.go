package ui

import (
	"strings"
	"testing"
)

func TestMaxLineWidth(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  int
	}{
		{"empty", []string{}, 0},
		{"single", []string{"hello"}, 5},
		{"multiple", []string{"hi", "hello", "hey"}, 5},
		{"with ansi", []string{"\x1b[31mred\x1b[0m"}, 3}, // visual width is 3
		{"empty lines", []string{"", "", ""}, 0},
		{"mixed", []string{"short", "longer line", "mid"}, 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxLineWidth(tt.lines)
			if got != tt.want {
				t.Errorf("maxLineWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompositeRow(t *testing.T) {
	tests := []struct {
		name        string
		bgLine      string
		modalLine   string
		modalStartX int
		modalWidth  int
		totalWidth  int
		wantModal   bool // should contain modal content
	}{
		{
			name:        "basic centered",
			bgLine:      "background text here",
			modalLine:   "[MODAL]",
			modalStartX: 5,
			modalWidth:  7,
			totalWidth:  20,
			wantModal:   true,
		},
		{
			name:        "modal at left edge",
			bgLine:      "background",
			modalLine:   "[M]",
			modalStartX: 0,
			modalWidth:  3,
			totalWidth:  10,
			wantModal:   true,
		},
		{
			name:        "background shorter than modal position",
			bgLine:      "hi",
			modalLine:   "[MODAL]",
			modalStartX: 10,
			modalWidth:  7,
			totalWidth:  20,
			wantModal:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compositeRow(tt.bgLine, tt.modalLine, tt.modalStartX, tt.modalWidth, tt.totalWidth)

			if tt.wantModal && !strings.Contains(got, tt.modalLine) {
				t.Errorf("compositeRow() missing modal content %q", tt.modalLine)
			}
		})
	}
}

func TestOverlayModal(t *testing.T) {
	tests := []struct {
		name       string
		background string
		modal      string
		width      int
		height     int
		checkFn    func(t *testing.T, result string)
	}{
		{
			name:       "basic overlay",
			background: "line1\nline2\nline3\nline4\nline5",
			modal:      "[M]",
			width:      10,
			height:     5,
			checkFn: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				if len(lines) != 5 {
					t.Errorf("expected 5 lines, got %d", len(lines))
				}
				// Modal should be in middle line (line 2, 0-indexed)
				if !strings.Contains(lines[2], "[M]") {
					t.Errorf("modal not found in expected line")
				}
			},
		},
		{
			name:       "strips ansi from background",
			background: "\x1b[31mred\x1b[0m\n\x1b[32mgreen\x1b[0m",
			modal:      "X",
			width:      10,
			height:     3,
			checkFn: func(t *testing.T, result string) {
				// Original ANSI codes should be stripped
				if strings.Contains(result, "\x1b[31m") {
					t.Errorf("original red ANSI code should be stripped")
				}
				// Modal should still be present
				if !strings.Contains(result, "X") {
					t.Errorf("modal should be present")
				}
			},
		},
		{
			name:       "modal larger than background",
			background: "a\nb",
			modal:      "MODAL",
			width:      10,
			height:     5,
			checkFn: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				if len(lines) != 5 {
					t.Errorf("expected 5 lines, got %d", len(lines))
				}
				// Modal should still be centered
				found := false
				for _, line := range lines {
					if strings.Contains(line, "MODAL") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("modal not found in result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OverlayModal(tt.background, tt.modal, tt.width, tt.height)
			tt.checkFn(t, result)
		})
	}
}

func TestDimLine(t *testing.T) {
	// dimLine should strip ANSI codes
	input := "\x1b[31mred text\x1b[0m"
	result := dimLine(input)

	// Should not contain original red ANSI code
	if strings.Contains(result, "\x1b[31m") {
		t.Errorf("dimLine should strip original ANSI codes")
	}

	// Should contain the plain text
	if !strings.Contains(result, "red text") {
		t.Errorf("dimLine should preserve text content")
	}
}
