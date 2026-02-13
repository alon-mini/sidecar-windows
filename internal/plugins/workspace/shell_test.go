package workspace

import (
	"strings"
	"testing"
)

func TestTmuxNotationToHuman(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"C-b notation", "C-b", "Ctrl-b"},
		{"C-a notation", "C-a", "Ctrl-a"},
		{"M-x notation", "M-x", "Alt-x"},
		{"M-a notation", "M-a", "Alt-a"},
		{"Short input single char", "C", "C"},
		{"Short input empty", "", ""},
		{"Unhandled notation", "F1", "F1"},
		{"Unknown prefix", "X-y", "X-y"},
		{"Lowercase c prefix", "c-b", "c-b"},
		{"Multiple dashes", "C-b-c", "Ctrl-b-c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tmuxNotationToHuman(tt.input)
			if got != tt.expected {
				t.Errorf("tmuxNotationToHuman(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetTmuxPrefix(t *testing.T) {
	// getTmuxPrefix() uses sync.Once for caching, so we can only test it once per process.
	// This test verifies the function returns a valid prefix format.
	prefix := getTmuxPrefix()

	// Should return a non-empty string
	if prefix == "" {
		t.Error("getTmuxPrefix() returned empty string")
	}

	// Should return a human-readable format (contains "Ctrl-" or "Alt-")
	// or the default "Ctrl-b"
	if !strings.Contains(prefix, "Ctrl-") && !strings.Contains(prefix, "Alt-") {
		t.Errorf("getTmuxPrefix() = %q, expected format with 'Ctrl-' or 'Alt-'", prefix)
	}

	// Verify caching: calling again should return the same value
	prefix2 := getTmuxPrefix()
	if prefix != prefix2 {
		t.Errorf("getTmuxPrefix() caching failed: first call = %q, second call = %q", prefix, prefix2)
	}
}

func TestGetTmuxPrefixConcurrency(t *testing.T) {
	// Test concurrent access to cached value
	done := make(chan string, 10)
	for i := 0; i < 10; i++ {
		go func() {
			done <- getTmuxPrefix()
		}()
	}

	// Collect all results
	var results []string
	for i := 0; i < 10; i++ {
		results = append(results, <-done)
	}

	// All results should be identical (cached value)
	first := results[0]
	for i, result := range results {
		if result != first {
			t.Errorf("Concurrent call %d returned %q, expected %q", i, result, first)
		}
	}
}
