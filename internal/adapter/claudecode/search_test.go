package claudecode

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcus/sidecar/internal/adapter"
)

func TestSearchMessages_InterfaceCompliance(t *testing.T) {
	a := New()
	// Verify interface compliance at compile time
	var _ adapter.MessageSearcher = a
}

func TestSearchMessages_EmptySession(t *testing.T) {
	// Create temp dir with empty session
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create empty session file
	sessionFile := filepath.Join(projectDir, "empty-session.jsonl")
	if err := os.WriteFile(sessionFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	a := &Adapter{
		projectsDir:  filepath.Join(tmpDir),
		sessionIndex: map[string]string{"empty-session": sessionFile},
		metaCache:    make(map[string]sessionMetaCacheEntry),
	}

	results, err := a.SearchMessages("empty-session", "test", adapter.DefaultSearchOptions())
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty session, got %v", results)
	}
}

func TestSearchMessages_WithMessages(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create session file with messages
	sessionContent := `{"type":"user","uuid":"u1","timestamp":"2025-01-01T12:00:00Z","message":{"role":"user","content":"hello world"}}
{"type":"assistant","uuid":"a1","timestamp":"2025-01-01T12:00:01Z","message":{"role":"assistant","content":"hi there, hello"}}`

	sessionFile := filepath.Join(projectDir, "test-session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatal(err)
	}

	a := &Adapter{
		projectsDir:  tmpDir,
		sessionIndex: map[string]string{"test-session": sessionFile},
		metaCache:    make(map[string]sessionMetaCacheEntry),
	}

	results, err := a.SearchMessages("test-session", "hello", adapter.DefaultSearchOptions())
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 message matches, got %d", len(results))
	}

	// First match should be in user message
	if results[0].Role != "user" {
		t.Errorf("expected first match role 'user', got %q", results[0].Role)
	}
	// Second match should be in assistant message
	if results[1].Role != "assistant" {
		t.Errorf("expected second match role 'assistant', got %q", results[1].Role)
	}
}

func TestSearchMessages_RegexSearch(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	sessionContent := `{"type":"user","uuid":"u1","timestamp":"2025-01-01T12:00:00Z","message":{"role":"user","content":"test123 test456"}}`
	sessionFile := filepath.Join(projectDir, "regex-session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatal(err)
	}

	a := &Adapter{
		projectsDir:  tmpDir,
		sessionIndex: map[string]string{"regex-session": sessionFile},
		metaCache:    make(map[string]sessionMetaCacheEntry),
	}

	opts := adapter.SearchOptions{UseRegex: true, MaxResults: 50}
	results, err := a.SearchMessages("regex-session", "test\\d+", opts)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 message match, got %d", len(results))
	}
	if len(results[0].Matches) != 2 {
		t.Errorf("expected 2 content matches, got %d", len(results[0].Matches))
	}
}

func TestSearchMessages_ToolUseSearch(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Message with tool_use block
	sessionContent := `{"type":"assistant","uuid":"a1","timestamp":"2025-01-01T12:00:00Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"tu1","name":"Bash","input":{"command":"grep -r pattern"}}]}}`
	sessionFile := filepath.Join(projectDir, "tool-session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatal(err)
	}

	a := &Adapter{
		projectsDir:  tmpDir,
		sessionIndex: map[string]string{"tool-session": sessionFile},
		metaCache:    make(map[string]sessionMetaCacheEntry),
	}

	results, err := a.SearchMessages("tool-session", "Bash", adapter.DefaultSearchOptions())
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected match for tool name")
	}
}

func TestSearchMessages_NonExistentSession(t *testing.T) {
	a := New()
	results, err := a.SearchMessages("nonexistent-session-xyz", "test", adapter.DefaultSearchOptions())
	if err != nil {
		t.Fatalf("expected no error for nonexistent session, got %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results, got %v", results)
	}
}

func TestSearchMessages_CaseSensitive(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	sessionContent := `{"type":"user","uuid":"u1","timestamp":"2025-01-01T12:00:00Z","message":{"role":"user","content":"Hello HELLO hello"}}`
	sessionFile := filepath.Join(projectDir, "case-session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatal(err)
	}

	a := &Adapter{
		projectsDir:  tmpDir,
		sessionIndex: map[string]string{"case-session": sessionFile},
		metaCache:    make(map[string]sessionMetaCacheEntry),
	}

	// Case insensitive should find all 3
	opts := adapter.SearchOptions{CaseSensitive: false, MaxResults: 50}
	results, err := a.SearchMessages("case-session", "hello", opts)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results[0].Matches) != 3 {
		t.Errorf("expected 3 case-insensitive matches, got %d", len(results[0].Matches))
	}

	// Case sensitive should find only 1 (lowercase)
	opts = adapter.SearchOptions{CaseSensitive: true, MaxResults: 50}
	results, err = a.SearchMessages("case-session", "hello", opts)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results[0].Matches) != 1 {
		t.Errorf("expected 1 case-sensitive match, got %d", len(results[0].Matches))
	}
}

// Ensure timestamp is included in matches
func TestSearchMessages_TimestampPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	sessionContent := `{"type":"user","uuid":"u1","timestamp":"2025-06-15T10:30:00Z","message":{"role":"user","content":"searchable text"}}`
	sessionFile := filepath.Join(projectDir, "ts-session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatal(err)
	}

	a := &Adapter{
		projectsDir:  tmpDir,
		sessionIndex: map[string]string{"ts-session": sessionFile},
		metaCache:    make(map[string]sessionMetaCacheEntry),
	}

	results, err := a.SearchMessages("ts-session", "searchable", adapter.DefaultSearchOptions())
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected match")
	}

	expected := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	if !results[0].Timestamp.Equal(expected) {
		t.Errorf("expected timestamp %v, got %v", expected, results[0].Timestamp)
	}
}
