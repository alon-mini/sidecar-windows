package conversations

import (
	"strings"
	"testing"
	"time"

	"github.com/marcus/sidecar/internal/adapter"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal", "normal-name", "normal-name"},
		{"slashes", "file/with\\slashes", "file-with-slashes"},
		{"special chars", "a:b*c?d\"e<f>g|h", "a-bcdefgh"},
		{"empty", "", "session"},
		{"only dashes", "---", "session"},
		{"only spaces", "   ", "session"},
		{"newlines", "name\nwith\nnewlines", "name with newlines"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename_LongName(t *testing.T) {
	long := strings.Repeat("a", 100)
	got := sanitizeFilename(long)
	if len([]rune(got)) > 50 {
		t.Errorf("sanitizeFilename should truncate to 50 runes, got %d", len([]rune(got)))
	}
}

func TestFormatExportDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"seconds", 30 * time.Second, "30s"},
		{"minutes", 5 * time.Minute, "5m"},
		{"hours", 2 * time.Hour, "2h"},
		{"hours and minutes", 2*time.Hour + 30*time.Minute, "2h 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatExportDuration(tt.d); got != tt.want {
				t.Errorf("formatExportDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestExportSessionAsMarkdown_NilSession(t *testing.T) {
	messages := []adapter.Message{
		{Role: "user", Content: "hello", Timestamp: time.Now()},
	}
	result := ExportSessionAsMarkdown(nil, messages)
	if !strings.Contains(result, "Unknown Session") {
		t.Error("nil session should use 'Unknown Session'")
	}
	if !strings.Contains(result, "hello") {
		t.Error("message content should be present")
	}
}

func TestExportSessionAsMarkdown_BasicSession(t *testing.T) {
	session := &adapter.Session{
		ID:          "ses_test123",
		Name:        "Test Session",
		CreatedAt:   time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		Duration:    45 * time.Minute,
		TotalTokens: 5000,
		EstCost:     0.15,
	}

	messages := []adapter.Message{
		{
			Role:      "user",
			Content:   "How does auth work?",
			Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			Role:      "assistant",
			Content:   "Auth uses JWT tokens.",
			Timestamp: time.Date(2025, 1, 15, 10, 31, 0, 0, time.UTC),
			Model:     "claude-sonnet-4-20250514",
			TokenUsage: adapter.TokenUsage{
				InputTokens:  1000,
				OutputTokens: 500,
			},
			ToolUses: []adapter.ToolUse{
				{Name: "Read", Input: `file_path="/src/auth.go"`},
			},
		},
	}

	result := ExportSessionAsMarkdown(session, messages)

	checks := []string{
		"Test Session",
		"2025-01-15",
		"45m",
		"5000",
		"$0.15",
		"User",
		"Assistant",
		"Auth uses JWT tokens.",
		"Read",
		"---",
	}
	for _, c := range checks {
		if !strings.Contains(result, c) {
			t.Errorf("expected output to contain %q", c)
		}
	}
}

func TestExportSessionAsMarkdown_EmptyMessages(t *testing.T) {
	session := &adapter.Session{
		Name:      "Empty",
		CreatedAt: time.Now(),
	}
	result := ExportSessionAsMarkdown(session, nil)
	if !strings.Contains(result, "Empty") {
		t.Error("session name should be in header")
	}
	if strings.Contains(result, "## User") || strings.Contains(result, "## Assistant") {
		t.Error("should not have message headers with no messages")
	}
}

func TestExportSessionAsMarkdown_WithThinkingBlocks(t *testing.T) {
	session := &adapter.Session{
		Name:      "Thinking",
		CreatedAt: time.Now(),
	}
	messages := []adapter.Message{
		{
			Role:      "assistant",
			Content:   "The answer is 42.",
			Timestamp: time.Now(),
			ThinkingBlocks: []adapter.ThinkingBlock{
				{Content: "Let me think about this...", TokenCount: 100},
			},
		},
	}

	result := ExportSessionAsMarkdown(session, messages)
	if !strings.Contains(result, "<details>") {
		t.Error("thinking blocks should use <details> tag")
	}
	if !strings.Contains(result, "<summary>") {
		t.Error("thinking blocks should use <summary> tag")
	}
	if !strings.Contains(result, "100 tokens") {
		t.Error("should show token count in thinking summary")
	}
}
