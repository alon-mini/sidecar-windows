package conversations

import (
	"testing"
	"time"

	"github.com/marcus/sidecar/internal/adapter"
)

func TestGroupMessagesIntoTurns(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		messages []adapter.Message
		want     []Turn
	}{
		{
			name:     "Empty messages",
			messages: []adapter.Message{},
			want:     nil,
		},
		{
			name: "Single user message",
			messages: []adapter.Message{
				{Role: "user", Content: "Hello", Timestamp: now},
			},
			want: []Turn{
				{
					Role:       "user",
					Messages:   []adapter.Message{{Role: "user", Content: "Hello", Timestamp: now}},
					StartIndex: 0,
				},
			},
		},
		{
			name: "Alternating roles",
			messages: []adapter.Message{
				{Role: "user", Content: "Hi", Timestamp: now},
				{Role: "assistant", Content: "Hello", Timestamp: now},
			},
			want: []Turn{
				{
					Role:       "user",
					Messages:   []adapter.Message{{Role: "user", Content: "Hi", Timestamp: now}},
					StartIndex: 0,
				},
				{
					Role:       "assistant",
					Messages:   []adapter.Message{{Role: "assistant", Content: "Hello", Timestamp: now}},
					StartIndex: 1,
				},
			},
		},
		{
			name: "Consecutive same role messages",
			messages: []adapter.Message{
				{Role: "user", Content: "Command output 1", Timestamp: now},
				{Role: "user", Content: "Command output 2", Timestamp: now},
				{Role: "user", Content: "Actual question", Timestamp: now},
			},
			want: []Turn{
				{
					Role: "user",
					Messages: []adapter.Message{
						{Role: "user", Content: "Command output 1", Timestamp: now},
						{Role: "user", Content: "Command output 2", Timestamp: now},
						{Role: "user", Content: "Actual question", Timestamp: now},
					},
					StartIndex: 0,
				},
			},
		},
		{
			name: "Mixed complex case",
			messages: []adapter.Message{
				{Role: "user", Content: "Q1", Timestamp: now, TokenUsage: adapter.TokenUsage{InputTokens: 10}},
				{Role: "assistant", Content: "A1 part 1", Timestamp: now, TokenUsage: adapter.TokenUsage{OutputTokens: 20}},
				{Role: "assistant", Content: "A1 part 2", Timestamp: now, TokenUsage: adapter.TokenUsage{OutputTokens: 15}, ToolUses: []adapter.ToolUse{{}}},
				{Role: "user", Content: "Q2", Timestamp: now, TokenUsage: adapter.TokenUsage{InputTokens: 5}},
			},
			want: []Turn{
				{
					Role:           "user",
					Messages:       []adapter.Message{{Role: "user", Content: "Q1", Timestamp: now, TokenUsage: adapter.TokenUsage{InputTokens: 10}}},
					StartIndex:     0,
					TotalTokensIn:  10,
					TotalTokensOut: 0,
				},
				{
					Role: "assistant",
					Messages: []adapter.Message{
						{Role: "assistant", Content: "A1 part 1", Timestamp: now, TokenUsage: adapter.TokenUsage{OutputTokens: 20}},
						{Role: "assistant", Content: "A1 part 2", Timestamp: now, TokenUsage: adapter.TokenUsage{OutputTokens: 15}, ToolUses: []adapter.ToolUse{{}}},
					},
					StartIndex:     1,
					TotalTokensIn:  0,
					TotalTokensOut: 35,
					ToolCount:      1,
				},
				{
					Role:           "user",
					Messages:       []adapter.Message{{Role: "user", Content: "Q2", Timestamp: now, TokenUsage: adapter.TokenUsage{InputTokens: 5}}},
					StartIndex:     3,
					TotalTokensIn:  5,
					TotalTokensOut: 0,
				},
			},
		},
		{
			name: "Token counting",
			messages: []adapter.Message{
				{
					Role: "assistant",
					ThinkingBlocks: []adapter.ThinkingBlock{
						{TokenCount: 100},
						{TokenCount: 50},
					},
					TokenUsage: adapter.TokenUsage{
						InputTokens:  10,
						OutputTokens: 20,
					},
				},
			},
			want: []Turn{
				{
					Role: "assistant",
					Messages: []adapter.Message{{
						Role: "assistant",
						ThinkingBlocks: []adapter.ThinkingBlock{
							{TokenCount: 100},
							{TokenCount: 50},
						},
						TokenUsage: adapter.TokenUsage{
							InputTokens:  10,
							OutputTokens: 20,
						},
					}},
					StartIndex:     0,
					TotalTokensIn:  10,
					TotalTokensOut: 20,
					ThinkingTokens: 150,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupMessagesIntoTurns(tt.messages)

			if len(got) != len(tt.want) {
				t.Fatalf("got %d turns, want %d", len(got), len(tt.want))
			}

			for i := range got {
				g := got[i]
				w := tt.want[i]

				if g.Role != w.Role {
					t.Errorf("turn[%d].Role = %q, want %q", i, g.Role, w.Role)
				}
				if g.StartIndex != w.StartIndex {
					t.Errorf("turn[%d].StartIndex = %d, want %d", i, g.StartIndex, w.StartIndex)
				}
				if len(g.Messages) != len(w.Messages) {
					t.Errorf("turn[%d].Messages length = %d, want %d", i, len(g.Messages), len(w.Messages))
				}
				if g.TotalTokensIn != w.TotalTokensIn {
					t.Errorf("turn[%d].TotalTokensIn = %d, want %d", i, g.TotalTokensIn, w.TotalTokensIn)
				}
				if g.TotalTokensOut != w.TotalTokensOut {
					t.Errorf("turn[%d].TotalTokensOut = %d, want %d", i, g.TotalTokensOut, w.TotalTokensOut)
				}
				if g.ThinkingTokens != w.ThinkingTokens {
					t.Errorf("turn[%d].ThinkingTokens = %d, want %d", i, g.ThinkingTokens, w.ThinkingTokens)
				}
				if g.ToolCount != w.ToolCount {
					t.Errorf("turn[%d].ToolCount = %d, want %d", i, g.ToolCount, w.ToolCount)
				}
			}
		})
	}
}

func TestStripXMLTags(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "Hello world",
			want:  "Hello world",
		},
		{
			input: "<tag>content</tag>",
			want:  "content",
		},
		{
			input: "Multiple <t1>tags</t1> here <t2>and</t2> there",
			want:  "Multiple tags here and there",
		},
		{
			input: "<user_query>Specific query</user_query>",
			want:  "Specific query",
		},
		{
			input: "Prefix <user_query>Query</user_query> Suffix",
			want:  "Query",
		},
		{
			input: "Mixed <other>tags</other> with <user_query>Target</user_query>",
			want:  "Target",
		},
		{
			input: "",
			want:  "",
		},
		{
			input: "   ",
			want:  "",
		},
		{
			input: "<outer><inner>nested</inner></outer>",
			want:  "nested",
		},
		{
			input: "<br/> some text",
			want:  "some text",
		},
		{
			input: `<div class="foo">bar</div>`,
			want:  "bar",
		},
		{
			input: "<user_query>first</user_query> <user_query>second</user_query>",
			want:  "first",
		},
		{
			input: "<user_query>  spaced  </user_query>",
			want:  "spaced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripXMLTags(tt.input)
			if got != tt.want {
				t.Errorf("stripXMLTags(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAppendMessagesToTurns(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		turns       []Turn
		newMessages []adapter.Message
		startIndex  int
		wantTurns   int
		wantLast    Turn // expected last turn
	}{
		{
			name:        "Empty new messages",
			turns:       []Turn{{Role: "user", StartIndex: 0}},
			newMessages: []adapter.Message{},
			startIndex:  1,
			wantTurns:   1,
			wantLast:    Turn{Role: "user", StartIndex: 0}, // Unchanged
		},
		{
			name:  "Empty existing turns",
			turns: nil,
			newMessages: []adapter.Message{
				{Role: "user", Content: "Hello", Timestamp: now},
			},
			startIndex: 0,
			wantTurns:  1,
			wantLast:   Turn{Role: "user", StartIndex: 0},
		},
		{
			name: "Extend last turn same role",
			turns: []Turn{
				{
					Role:       "user",
					StartIndex: 0,
					Messages:   []adapter.Message{{Role: "user", Content: "Q1", Timestamp: now}},
				},
				{
					Role:       "assistant",
					StartIndex: 1,
					Messages:   []adapter.Message{{Role: "assistant", Content: "A1", Timestamp: now}},
				},
			},
			newMessages: []adapter.Message{
				{Role: "assistant", Content: "A1 continued", Timestamp: now, TokenUsage: adapter.TokenUsage{OutputTokens: 100}},
			},
			startIndex: 2,
			wantTurns:  2, // Should NOT create new turn
			wantLast:   Turn{Role: "assistant", StartIndex: 1, TotalTokensOut: 100},
		},
		{
			name: "New turn different role",
			turns: []Turn{
				{
					Role:       "assistant",
					StartIndex: 0,
					Messages:   []adapter.Message{{Role: "assistant", Content: "A1", Timestamp: now}},
				},
			},
			newMessages: []adapter.Message{
				{Role: "user", Content: "Q2", Timestamp: now, TokenUsage: adapter.TokenUsage{InputTokens: 50}},
			},
			startIndex: 1,
			wantTurns:  2, // Should create new turn
			wantLast:   Turn{Role: "user", StartIndex: 1, TotalTokensIn: 50},
		},
		{
			name: "Multiple new messages with role change",
			turns: []Turn{
				{Role: "user", StartIndex: 0, Messages: []adapter.Message{{Role: "user"}}},
			},
			newMessages: []adapter.Message{
				{Role: "assistant", Content: "Response 1", Timestamp: now},
				{Role: "assistant", Content: "Response 2", Timestamp: now, ToolUses: []adapter.ToolUse{{}}},
				{Role: "user", Content: "Follow up", Timestamp: now},
			},
			startIndex: 1,
			wantTurns:  3, // user -> assistant -> user
			wantLast:   Turn{Role: "user", StartIndex: 3},
		},
		{
			name: "Verify correct StartIndex with offset",
			turns: []Turn{
				{Role: "user", StartIndex: 0, Messages: []adapter.Message{{Role: "user"}}},
				{Role: "assistant", StartIndex: 1, Messages: []adapter.Message{{Role: "assistant"}}},
			},
			newMessages: []adapter.Message{
				{Role: "user", Content: "Q2", Timestamp: now},
				{Role: "assistant", Content: "A2", Timestamp: now},
			},
			startIndex: 2,
			wantTurns:  4,
			wantLast:   Turn{Role: "assistant", StartIndex: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendMessagesToTurns(tt.turns, tt.newMessages, tt.startIndex)

			if len(got) != tt.wantTurns {
				t.Fatalf("got %d turns, want %d", len(got), tt.wantTurns)
			}

			if tt.wantTurns > 0 && len(got) > 0 {
				last := got[len(got)-1]
				if last.Role != tt.wantLast.Role {
					t.Errorf("last turn Role = %q, want %q", last.Role, tt.wantLast.Role)
				}
				if last.StartIndex != tt.wantLast.StartIndex {
					t.Errorf("last turn StartIndex = %d, want %d", last.StartIndex, tt.wantLast.StartIndex)
				}
				if tt.wantLast.TotalTokensIn > 0 && last.TotalTokensIn != tt.wantLast.TotalTokensIn {
					t.Errorf("last turn TotalTokensIn = %d, want %d", last.TotalTokensIn, tt.wantLast.TotalTokensIn)
				}
				if tt.wantLast.TotalTokensOut > 0 && last.TotalTokensOut != tt.wantLast.TotalTokensOut {
					t.Errorf("last turn TotalTokensOut = %d, want %d", last.TotalTokensOut, tt.wantLast.TotalTokensOut)
				}
			}
		})
	}
}

func TestAppendMessagesToTurns_Equivalent(t *testing.T) {
	// Verify that incremental append produces same result as full grouping
	now := time.Now()
	allMessages := []adapter.Message{
		{Role: "user", Content: "Q1", Timestamp: now, TokenUsage: adapter.TokenUsage{InputTokens: 10}},
		{Role: "assistant", Content: "A1", Timestamp: now, TokenUsage: adapter.TokenUsage{OutputTokens: 20}},
		{Role: "assistant", Content: "A1 tool", Timestamp: now, ToolUses: []adapter.ToolUse{{}}},
		{Role: "user", Content: "Q2", Timestamp: now, TokenUsage: adapter.TokenUsage{InputTokens: 15}},
		{Role: "assistant", Content: "A2", Timestamp: now, TokenUsage: adapter.TokenUsage{OutputTokens: 30}},
	}

	// Full grouping
	fullTurns := GroupMessagesIntoTurns(allMessages)

	// Incremental: group first 2, then append rest
	initialTurns := GroupMessagesIntoTurns(allMessages[:2])
	incrementalTurns := AppendMessagesToTurns(initialTurns, allMessages[2:], 2)

	if len(fullTurns) != len(incrementalTurns) {
		t.Fatalf("turn count mismatch: full=%d, incremental=%d", len(fullTurns), len(incrementalTurns))
	}

	for i := range fullTurns {
		f, inc := fullTurns[i], incrementalTurns[i]
		if f.Role != inc.Role {
			t.Errorf("turn[%d] Role: full=%q, incremental=%q", i, f.Role, inc.Role)
		}
		if f.StartIndex != inc.StartIndex {
			t.Errorf("turn[%d] StartIndex: full=%d, incremental=%d", i, f.StartIndex, inc.StartIndex)
		}
		if len(f.Messages) != len(inc.Messages) {
			t.Errorf("turn[%d] Messages: full=%d, incremental=%d", i, len(f.Messages), len(inc.Messages))
		}
		if f.TotalTokensIn != inc.TotalTokensIn {
			t.Errorf("turn[%d] TotalTokensIn: full=%d, incremental=%d", i, f.TotalTokensIn, inc.TotalTokensIn)
		}
		if f.TotalTokensOut != inc.TotalTokensOut {
			t.Errorf("turn[%d] TotalTokensOut: full=%d, incremental=%d", i, f.TotalTokensOut, inc.TotalTokensOut)
		}
		if f.ToolCount != inc.ToolCount {
			t.Errorf("turn[%d] ToolCount: full=%d, incremental=%d", i, f.ToolCount, inc.ToolCount)
		}
	}
}

func TestTurnPreview(t *testing.T) {
	tests := []struct {
		name     string
		messages []adapter.Message
		maxLen   int
		want     string
	}{
		{
			name:     "Empty turn",
			messages: nil,
			maxLen:   10,
			want:     "",
		},
		{
			name: "Simple content",
			messages: []adapter.Message{
				{Content: "Hello world"},
			},
			maxLen: 20,
			want:   "Hello world",
		},
		{
			name: "Truncated content",
			messages: []adapter.Message{
				{Content: "Hello world"},
			},
			maxLen: 8,
			want:   "Hello...",
		},
		{
			name: "Skip tool result marker",
			messages: []adapter.Message{
				{Content: "[1 tool result(s)]"},
				{Content: "Actual content"},
			},
			maxLen: 20,
			want:   "Actual content",
		},
		{
			name: "XML tag stripping",
			messages: []adapter.Message{
				{Content: "<ant_thinking>Thinking...</ant_thinking>Response"},
			},
			maxLen: 20,
			want:   "Thinking...Response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			turn := Turn{Messages: tt.messages}
			got := turn.Preview(tt.maxLen)
			if got != tt.want {
				t.Errorf("Turn.Preview() = %q, want %q", got, tt.want)
			}
		})
	}
}
