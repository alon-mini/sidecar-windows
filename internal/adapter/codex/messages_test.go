package codex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMessagesAndUsage(t *testing.T) {
	root := t.TempDir()
	sessionsDir := filepath.Join(root, "sessions")
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	path := filepath.Join(sessionsDir, "2025", "11", "20")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	lines := []string{
		`{"timestamp":"2025-11-21T04:13:55.791Z","type":"session_meta","payload":{"id":"id-1","timestamp":"2025-11-21T04:13:55.777Z","cwd":"` + projectDir + `"}}`,
		`{"timestamp":"2025-11-21T04:13:56.000Z","type":"turn_context","payload":{"model":"gpt-4.1"}}`,
		`{"timestamp":"2025-11-21T04:14:00.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"hello"}]}}`,
		`{"timestamp":"2025-11-21T04:14:01.000Z","type":"response_item","payload":{"type":"function_call","name":"shell_command","arguments":"{\"command\":\"ls\"}","call_id":"call-1"}}`,
		`{"timestamp":"2025-11-21T04:14:02.000Z","type":"response_item","payload":{"type":"function_call_output","call_id":"call-1","output":"OK"}}`,
		`{"timestamp":"2025-11-21T04:14:03.000Z","type":"response_item","payload":{"type":"reasoning","summary":[{"type":"summary_text","text":"Thinking"}]}}`,
		`{"timestamp":"2025-11-21T04:14:03.500Z","type":"event_msg","payload":{"type":"agent_reasoning","text":"extra reasoning"}}`,
		`{"timestamp":"2025-11-21T04:14:04.000Z","type":"event_msg","payload":{"type":"token_count","info":{"last_token_usage":{"input_tokens":10,"cached_input_tokens":2,"output_tokens":5,"reasoning_output_tokens":1}}}}`,
		`{"timestamp":"2025-11-21T04:14:05.000Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"done"}]}}`,
		`{"timestamp":"2025-11-21T04:14:06.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"apply_patch","input":"patch","call_id":"call-2"}}`,
		`{"timestamp":"2025-11-21T04:14:07.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call-2","output":"applied"}}`,
		`{"timestamp":"2025-11-21T04:14:08.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"next"}]}}`,
	}
	if err := writeSessionFile(filepath.Join(path, "rollout-1.jsonl"), lines); err != nil {
		t.Fatalf("write session file: %v", err)
	}

	a := New()
	a.sessionsDir = sessionsDir

	messages, err := a.Messages("id-1")
	if err != nil {
		t.Fatalf("Messages error: %v", err)
	}
	if len(messages) != 4 {
		t.Fatalf("Messages() = %d, want 4", len(messages))
	}
	if messages[0].Role != "user" || messages[0].Content != "hello" {
		t.Fatalf("unexpected first message: %+v", messages[0])
	}
	if messages[1].Role != "assistant" || messages[1].Content != "done" {
		t.Fatalf("unexpected assistant message: %+v", messages[1])
	}
	if messages[1].Model != "gpt-4.1" {
		t.Fatalf("Model = %q, want gpt-4.1", messages[1].Model)
	}
	if len(messages[1].ToolUses) != 1 || messages[1].ToolUses[0].Output != "OK" {
		t.Fatalf("tool uses not attached: %+v", messages[1].ToolUses)
	}
	if len(messages[1].ThinkingBlocks) != 2 {
		t.Fatalf("thinking blocks = %d, want 2", len(messages[1].ThinkingBlocks))
	}
	if messages[1].InputTokens != 10 || messages[1].OutputTokens != 6 || messages[1].CacheRead != 2 {
		t.Fatalf("token usage mismatch: %+v", messages[1].TokenUsage)
	}
	if messages[2].Role != "assistant" || messages[2].Content != "tool calls" {
		t.Fatalf("unexpected synthetic message: %+v", messages[2])
	}
	if len(messages[2].ToolUses) != 1 || messages[2].ToolUses[0].ID != "call-2" {
		t.Fatalf("synthetic tool uses mismatch: %+v", messages[2].ToolUses)
	}
	if messages[3].Role != "user" || messages[3].Content != "next" {
		t.Fatalf("unexpected final user message: %+v", messages[3])
	}

	usage, err := a.Usage("id-1")
	if err != nil {
		t.Fatalf("Usage error: %v", err)
	}
	if usage.TotalInputTokens != 10 || usage.TotalOutputTokens != 6 || usage.TotalCacheRead != 2 {
		t.Fatalf("usage mismatch: %+v", usage)
	}
	if usage.MessageCount != 4 {
		t.Fatalf("usage MessageCount = %d, want 4", usage.MessageCount)
	}
}
