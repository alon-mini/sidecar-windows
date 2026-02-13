package cursor

import (
	"os"
	"testing"
)

func TestID(t *testing.T) {
	a := New()
	if a.ID() != "cursor-cli" {
		t.Errorf("expected ID 'cursor-cli', got %q", a.ID())
	}
}

func TestName(t *testing.T) {
	a := New()
	if a.Name() != "Cursor CLI" {
		t.Errorf("expected Name 'Cursor CLI', got %q", a.Name())
	}
}

func TestCapabilities(t *testing.T) {
	a := New()
	caps := a.Capabilities()

	if !caps["sessions"] {
		t.Error("expected sessions capability")
	}
	if !caps["messages"] {
		t.Error("expected messages capability")
	}
	if caps["usage"] {
		t.Error("usage capability should be false for cursor")
	}
	if !caps["watch"] {
		t.Error("expected watch capability")
	}
}

func TestDetect(t *testing.T) {
	a := New()

	// Get the current working directory for testing
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Try to detect sessions for current project (may or may not exist)
	found, err := a.Detect(cwd)
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	t.Logf("Cursor CLI sessions for %s: %v", cwd, found)

	// Should not detect for non-existent project
	found, err = a.Detect("/nonexistent/path")
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	if found {
		t.Error("should not find sessions for nonexistent path")
	}
}

func TestDetect_RelativePath(t *testing.T) {
	a := New()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Both relative and absolute should return same result
	foundRel, errRel := a.Detect(".")
	foundAbs, errAbs := a.Detect(cwd)

	if errRel != nil {
		t.Fatalf("Detect(.) error: %v", errRel)
	}
	if errAbs != nil {
		t.Fatalf("Detect(cwd) error: %v", errAbs)
	}

	if foundRel != foundAbs {
		t.Errorf("Detect('.') = %v, Detect(cwd) = %v - should be equal", foundRel, foundAbs)
	}
}

func TestWorkspacePath(t *testing.T) {
	a := New()

	// Test that workspace path is computed consistently
	path1 := a.workspacePath("/Users/test/code/project")
	path2 := a.workspacePath("/Users/test/code/project")

	if path1 != path2 {
		t.Errorf("same input should produce same path: %q != %q", path1, path2)
	}

	// Different paths should produce different hashes
	path3 := a.workspacePath("/Users/test/code/other")
	if path1 == path3 {
		t.Error("different inputs should produce different paths")
	}
}

func TestSessions(t *testing.T) {
	a := New()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	sessions, err := a.Sessions(cwd)
	if err != nil {
		t.Fatalf("Sessions error: %v", err)
	}

	if len(sessions) == 0 {
		t.Skip("no cursor sessions found for testing")
	}

	t.Logf("found %d sessions", len(sessions))

	// Check first session has required fields
	s := sessions[0]
	if s.ID == "" {
		t.Error("session ID should not be empty")
	}
	if s.Name == "" {
		t.Error("session Name should not be empty")
	}
	if s.AdapterID != "cursor-cli" {
		t.Errorf("expected AdapterID 'cursor-cli', got %q", s.AdapterID)
	}
	if s.AdapterName != "Cursor CLI" {
		t.Errorf("expected AdapterName 'Cursor CLI', got %q", s.AdapterName)
	}
	if s.CreatedAt.IsZero() {
		t.Error("session CreatedAt should not be zero")
	}
	if s.UpdatedAt.IsZero() {
		t.Error("session UpdatedAt should not be zero")
	}

	t.Logf("newest session: %s (created %v, updated %v)", s.ID, s.CreatedAt, s.UpdatedAt)
}

func TestSessions_RelativePath(t *testing.T) {
	a := New()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Both should return same sessions
	sessionsRel, errRel := a.Sessions(".")
	sessionsAbs, errAbs := a.Sessions(cwd)

	if errRel != nil {
		t.Fatalf("Sessions(.) error: %v", errRel)
	}
	if errAbs != nil {
		t.Fatalf("Sessions(cwd) error: %v", errAbs)
	}

	if len(sessionsRel) != len(sessionsAbs) {
		t.Errorf("Sessions('.') = %d sessions, Sessions(cwd) = %d sessions - should be equal",
			len(sessionsRel), len(sessionsAbs))
	}
}

func TestMessages(t *testing.T) {
	a := New()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	sessions, err := a.Sessions(cwd)
	if err != nil {
		t.Fatalf("Sessions error: %v", err)
	}

	if len(sessions) == 0 {
		t.Skip("no cursor sessions found for testing")
	}

	// Get messages from the most recent session
	messages, err := a.Messages(sessions[0].ID)
	if err != nil {
		t.Fatalf("Messages error: %v", err)
	}

	if len(messages) == 0 {
		t.Skip("no messages in session")
	}

	t.Logf("found %d messages", len(messages))

	// Check message roles
	userCount, assistantCount := 0, 0
	for _, m := range messages {
		switch m.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		}
	}

	t.Logf("user messages: %d, assistant messages: %d", userCount, assistantCount)

	// Check for tool uses in assistant messages
	toolUseCount := 0
	for _, msg := range messages {
		if msg.Role == "assistant" && len(msg.ToolUses) > 0 {
			toolUseCount += len(msg.ToolUses)
		}
	}
	t.Logf("found %d tool uses across messages", toolUseCount)

	// Check for thinking blocks
	thinkingCount := 0
	for _, msg := range messages {
		if len(msg.ThinkingBlocks) > 0 {
			thinkingCount += len(msg.ThinkingBlocks)
		}
	}
	t.Logf("found %d thinking blocks across messages", thinkingCount)
}

func TestUsage(t *testing.T) {
	a := New()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	sessions, err := a.Sessions(cwd)
	if err != nil {
		t.Fatalf("Sessions error: %v", err)
	}

	if len(sessions) == 0 {
		t.Skip("no cursor sessions found for testing")
	}

	usage, err := a.Usage(sessions[0].ID)
	if err != nil {
		t.Fatalf("Usage error: %v", err)
	}

	t.Logf("usage: input=%d output=%d messages=%d",
		usage.TotalInputTokens, usage.TotalOutputTokens,
		usage.MessageCount)
}

func TestShortID(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"12345678", "12345678"},
		{"123456789abcdef", "12345678"},
		{"1234567", "1234567"},
		{"abc", "abc"},
		{"", ""},
	}

	for _, tt := range tests {
		result := shortID(tt.id)
		if result != tt.expected {
			t.Errorf("shortID(%q) = %q, expected %q", tt.id, result, tt.expected)
		}
	}
}

// TestSessionCacheInitialized verifies the session cache is initialized (td-107eea24)
func TestSessionCacheInitialized(t *testing.T) {
	a := New()
	if a.sessionCache == nil {
		t.Error("expected sessionCache to be initialized")
	}
}

// TestParseContentBlocks_TextBlock verifies text blocks produce ContentBlock with Type='text'
func TestParseContentBlocks_TextBlock(t *testing.T) {
	a := New()
	content := []byte(`[{"type":"text","text":"Hello world"}]`)

	_, _, _, contentBlocks, _ := a.parseContent(content)

	if len(contentBlocks) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(contentBlocks))
	}
	if contentBlocks[0].Type != "text" {
		t.Errorf("expected Type='text', got %q", contentBlocks[0].Type)
	}
	if contentBlocks[0].Text != "Hello world" {
		t.Errorf("expected Text='Hello world', got %q", contentBlocks[0].Text)
	}
}

// TestParseContentBlocks_ToolCall verifies tool-call blocks produce ContentBlock with Type='tool_use'
func TestParseContentBlocks_ToolCall(t *testing.T) {
	a := New()
	content := []byte(`[{"type":"tool-call","toolCallId":"tc_123","toolName":"Read","args":{"path":"foo.go"}}]`)

	_, toolUses, _, contentBlocks, _ := a.parseContent(content)

	// Verify ToolUses array
	if len(toolUses) != 1 {
		t.Fatalf("expected 1 tool use, got %d", len(toolUses))
	}
	if toolUses[0].ID != "tc_123" {
		t.Errorf("expected ToolUse.ID='tc_123', got %q", toolUses[0].ID)
	}

	// Verify ContentBlocks
	if len(contentBlocks) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(contentBlocks))
	}
	if contentBlocks[0].Type != "tool_use" {
		t.Errorf("expected Type='tool_use', got %q", contentBlocks[0].Type)
	}
	if contentBlocks[0].ToolUseID != "tc_123" {
		t.Errorf("expected ToolUseID='tc_123', got %q", contentBlocks[0].ToolUseID)
	}
	if contentBlocks[0].ToolName != "Read" {
		t.Errorf("expected ToolName='Read', got %q", contentBlocks[0].ToolName)
	}
}

// TestParseContentBlocks_ToolResult verifies tool-result blocks produce ContentBlock with Type='tool_result'
func TestParseContentBlocks_ToolResult(t *testing.T) {
	a := New()
	content := []byte(`[{"type":"tool-result","toolCallId":"tc_123","result":"file contents here"}]`)

	_, _, _, contentBlocks, _ := a.parseContent(content)

	if len(contentBlocks) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(contentBlocks))
	}
	if contentBlocks[0].Type != "tool_result" {
		t.Errorf("expected Type='tool_result', got %q", contentBlocks[0].Type)
	}
	if contentBlocks[0].ToolUseID != "tc_123" {
		t.Errorf("expected ToolUseID='tc_123', got %q", contentBlocks[0].ToolUseID)
	}
	if contentBlocks[0].ToolOutput != "file contents here" {
		t.Errorf("expected ToolOutput='file contents here', got %q", contentBlocks[0].ToolOutput)
	}
}

// TestParseContentBlocks_ToolResultArray verifies tool-result with array content is extracted
func TestParseContentBlocks_ToolResultArray(t *testing.T) {
	a := New()
	content := []byte(`[{"type":"tool-result","toolCallId":"tc_456","result":[{"type":"text","text":"line 1"},{"type":"text","text":"line 2"}]}]`)

	_, _, _, contentBlocks, _ := a.parseContent(content)

	if len(contentBlocks) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(contentBlocks))
	}
	if contentBlocks[0].Type != "tool_result" {
		t.Errorf("expected Type='tool_result', got %q", contentBlocks[0].Type)
	}
	if contentBlocks[0].ToolOutput != "line 1\nline 2" {
		t.Errorf("expected ToolOutput='line 1\\nline 2', got %q", contentBlocks[0].ToolOutput)
	}
}

// TestParseContentBlocks_Reasoning verifies reasoning blocks produce ContentBlock with Type='thinking'
func TestParseContentBlocks_Reasoning(t *testing.T) {
	a := New()
	content := []byte(`[{"type":"reasoning","text":"Let me think about this..."}]`)

	_, _, thinkingBlocks, contentBlocks, _ := a.parseContent(content)

	// Verify ThinkingBlocks array
	if len(thinkingBlocks) != 1 {
		t.Fatalf("expected 1 thinking block, got %d", len(thinkingBlocks))
	}
	if thinkingBlocks[0].Content != "Let me think about this..." {
		t.Errorf("expected ThinkingBlock.Content='Let me think about this...', got %q", thinkingBlocks[0].Content)
	}

	// Verify ContentBlocks
	if len(contentBlocks) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(contentBlocks))
	}
	if contentBlocks[0].Type != "thinking" {
		t.Errorf("expected Type='thinking', got %q", contentBlocks[0].Type)
	}
	if contentBlocks[0].Text != "Let me think about this..." {
		t.Errorf("expected Text='Let me think about this...', got %q", contentBlocks[0].Text)
	}
}

// TestParseContentBlocks_Mixed verifies mixed content produces correct ContentBlocks in order
func TestParseContentBlocks_Mixed(t *testing.T) {
	a := New()
	content := []byte(`[
		{"type":"text","text":"I will read the file"},
		{"type":"tool-call","toolCallId":"tc_001","toolName":"Read","args":{"path":"main.go"}}
	]`)

	text, toolUses, _, contentBlocks, _ := a.parseContent(content)

	// Verify content string
	if text != "I will read the file" {
		t.Errorf("expected content='I will read the file', got %q", text)
	}

	// Verify ToolUses
	if len(toolUses) != 1 {
		t.Fatalf("expected 1 tool use, got %d", len(toolUses))
	}

	// Verify ContentBlocks order
	if len(contentBlocks) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(contentBlocks))
	}
	if contentBlocks[0].Type != "text" {
		t.Errorf("blocks[0].Type=%q, want 'text'", contentBlocks[0].Type)
	}
	if contentBlocks[1].Type != "tool_use" {
		t.Errorf("blocks[1].Type=%q, want 'tool_use'", contentBlocks[1].Type)
	}
}

// TestParseContentBlocks_StringContent verifies string content produces single text ContentBlock
func TestParseContentBlocks_StringContent(t *testing.T) {
	a := New()
	content := []byte(`"Just a simple string message"`)

	text, _, _, contentBlocks, _ := a.parseContent(content)

	if text != "Just a simple string message" {
		t.Errorf("expected content='Just a simple string message', got %q", text)
	}
	if len(contentBlocks) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(contentBlocks))
	}
	if contentBlocks[0].Type != "text" {
		t.Errorf("expected Type='text', got %q", contentBlocks[0].Type)
	}
	if contentBlocks[0].Text != "Just a simple string message" {
		t.Errorf("expected Text='Just a simple string message', got %q", contentBlocks[0].Text)
	}
}

// TestParseContentBlocks_ToolResultError verifies IsError flag is captured
func TestParseContentBlocks_ToolResultError(t *testing.T) {
	a := New()
	content := []byte(`[{"type":"tool-result","toolCallId":"tc_err","result":"Error: file not found","isError":true}]`)

	_, _, _, contentBlocks, _ := a.parseContent(content)

	if len(contentBlocks) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(contentBlocks))
	}
	if !contentBlocks[0].IsError {
		t.Error("expected IsError=true, got false")
	}
}
