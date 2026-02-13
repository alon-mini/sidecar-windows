package codex

import (
	"encoding/json"
	"time"
)

// RawRecord represents a single JSONL line in a Codex session file.
type RawRecord struct {
	Timestamp time.Time       `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

// SessionMetaPayload holds metadata about a Codex session.
type SessionMetaPayload struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	CWD       string    `json:"cwd"`
	Source    string    `json:"source"`
}

// ResponseItemBase holds the response item type.
type ResponseItemBase struct {
	Type string `json:"type"`
}

// ResponseMessagePayload represents a user or assistant message.
type ResponseMessagePayload struct {
	Type    string         `json:"type"`
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock is a content fragment inside a message.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ResponseToolCallPayload represents a tool call request.
type ResponseToolCallPayload struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	CallID    string          `json:"call_id"`
}

// ResponseToolOutputPayload represents a tool call response.
type ResponseToolOutputPayload struct {
	Type   string          `json:"type"`
	CallID string          `json:"call_id"`
	Output json.RawMessage `json:"output,omitempty"`
}

// ResponseReasoningPayload represents reasoning metadata.
type ResponseReasoningPayload struct {
	Type    string             `json:"type"`
	Summary []ReasoningSummary `json:"summary"`
}

// ReasoningSummary holds summarized reasoning text.
type ReasoningSummary struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// EventMsgPayload represents an event message.
type EventMsgPayload struct {
	Type string          `json:"type"`
	Text string          `json:"text,omitempty"`
	Info *TokenCountInfo `json:"info,omitempty"`
}

// TokenCountInfo contains token usage stats.
type TokenCountInfo struct {
	TotalTokenUsage *TokenUsage `json:"total_token_usage"`
	LastTokenUsage  *TokenUsage `json:"last_token_usage"`
}

// TokenUsage represents token usage metrics.
type TokenUsage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
	TotalTokens           int `json:"total_tokens"`
}

// TurnContextPayload carries model context per turn.
type TurnContextPayload struct {
	Model string `json:"model"`
}

// SessionMetadata aggregates metadata extracted from a session file.
type SessionMetadata struct {
	Path             string
	SessionID        string
	CWD              string
	FirstMsg         time.Time
	LastMsg          time.Time
	MsgCount         int
	TotalTokens      int
	FirstUserMessage string // Content of the first user message (for title)
}
