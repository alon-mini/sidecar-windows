package ui

import "testing"

func TestCalculateModalWidth(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		min      int
		max      int
		expected int
	}{
		{
			name:     "short content uses min",
			content:  "Hi",
			min:      40,
			max:      80,
			expected: 40,
		},
		{
			name:     "long content capped at max",
			content:  "This is a very long line that exceeds the maximum allowed width by far",
			min:      40,
			max:      60,
			expected: 60,
		},
		{
			name:     "multiline uses longest",
			content:  "Short\nThis is the longest line here\nMedium",
			min:      20,
			max:      100,
			expected: 35, // 29 + 6 padding
		},
		{
			name:     "strips ANSI codes",
			content:  "\x1b[31mRed text\x1b[0m",
			min:      20,
			max:      100,
			expected: 20, // "Red text" = 8 + 6 = 14, min is 20
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateModalWidth(tt.content, tt.min, tt.max)
			if got != tt.expected {
				t.Errorf("got %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestClampModalWidth(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		screenWidth int
		expected    int
	}{
		{
			name:        "within bounds",
			width:       50,
			screenWidth: 100,
			expected:    50,
		},
		{
			name:        "exceeds screen",
			width:       100,
			screenWidth: 80,
			expected:    70, // 80 - 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampModalWidth(tt.width, tt.screenWidth)
			if got != tt.expected {
				t.Errorf("got %d, want %d", got, tt.expected)
			}
		})
	}
}
