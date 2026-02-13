package warp

import (
	"testing"
	"time"
)

func TestParseWarpTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2025-12-27 10:12:11", time.Date(2025, 12, 27, 10, 12, 11, 0, time.UTC)},
		{"2025-12-27T10:12:11Z", time.Date(2025, 12, 27, 10, 12, 11, 0, time.UTC)},
		{"", time.Time{}},
		{"invalid", time.Time{}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseWarpTimestamp(tc.input)
			if !result.Equal(tc.expected) {
				t.Errorf("parseWarpTimestamp(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"\x1b[1mBold\x1b[0m", "Bold"},
		{"\x1b[31mRed\x1b[0m text", "Red text"},
		{"\x1b[1;32mGreen Bold\x1b[0m", "Green Bold"},
		{"", ""},
		{"no escape codes", "no escape codes"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := stripANSI(tc.input)
			if result != tc.expected {
				t.Errorf("stripANSI(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestExtractQueryText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "empty array",
			input:    "[]",
			expected: "",
		},
		{
			name:     "valid query",
			input:    `[{"Query":{"text":"What is the meaning of life?","context":[]}}]`,
			expected: "What is the meaning of life?",
		},
		{
			name:     "invalid json",
			input:    "not json",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractQueryText(tc.input)
			if result != tc.expected {
				t.Errorf("extractQueryText(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long text", 10, "this is..."},
		{"hello\nworld", 50, "hello world"},
		{"", 10, ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := truncateText(tc.input, tc.maxLen)
			if result != tc.expected {
				t.Errorf("truncateText(%q, %d) = %q, want %q", tc.input, tc.maxLen, result, tc.expected)
			}
		})
	}
}

func TestShortConversationID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"e05f25f4-4e57-411a-b126-f4c9ebca6781", "e05f25f4"},
		{"short", "short"},
		{"12345678", "12345678"},
		{"1234567", "1234567"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := shortConversationID(tc.input)
			if result != tc.expected {
				t.Errorf("shortConversationID(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestCwdMatchesProject(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		cwd         string
		expected    bool
	}{
		{"empty project", "", "/some/path", false},
		{"empty cwd", "/some/path", "", false},
		{"exact match", "/Users/test/project", "/Users/test/project", true},
		{"subdirectory", "/Users/test/project", "/Users/test/project/src", true},
		{"different path", "/Users/test/project", "/Users/other/project", false},
		{"parent directory", "/Users/test/project/src", "/Users/test/project", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cwdMatchesProject(tc.projectRoot, tc.cwd)
			if result != tc.expected {
				t.Errorf("cwdMatchesProject(%q, %q) = %v, want %v",
					tc.projectRoot, tc.cwd, result, tc.expected)
			}
		})
	}
}

func TestAdapterBasics(t *testing.T) {
	a := New()

	if a.ID() != "warp" {
		t.Errorf("ID() = %q, want %q", a.ID(), "warp")
	}

	if a.Name() != "Warp" {
		t.Errorf("Name() = %q, want %q", a.Name(), "Warp")
	}

	caps := a.Capabilities()
	if !caps["sessions"] {
		t.Error("Expected sessions capability")
	}
	if !caps["messages"] {
		t.Error("Expected messages capability")
	}
	if !caps["usage"] {
		t.Error("Expected usage capability")
	}
	if !caps["watch"] {
		t.Error("Expected watch capability")
	}
}
