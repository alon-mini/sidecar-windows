package amp

import (
	"encoding/json"
	"strings"
	"time"
)

// Thread represents a top-level Amp thread from threads/T-{uuid}.json.
type Thread struct {
	V        int       `json:"v"`
	ID       string    `json:"id"`
	Created  int64     `json:"created"` // Unix millis
	Messages []Message `json:"messages"`
	Env      *Env      `json:"env,omitempty"`
}

// CreatedTime returns the thread creation time.
func (t *Thread) CreatedTime() time.Time {
	if t.Created == 0 {
		return time.Time{}
	}
	return time.UnixMilli(t.Created).Local()
}

// Message represents a message in an Amp thread.
type Message struct {
	Role      string         `json:"role"` // "user" or "assistant"
	MessageID int            `json:"messageId"`
	Content   []ContentBlock `json:"content"`
	State     *MessageState  `json:"state,omitempty"`
	Usage     *Usage         `json:"usage,omitempty"`
	Meta      *MessageMeta   `json:"meta,omitempty"`
}

// ContentBlock represents a content block in an Amp message.
type ContentBlock struct {
	Type string `json:"type"` // "text", "thinking", "tool_use", "tool_result"

	// Text block fields
	Text string `json:"text,omitempty"`

	// Thinking block fields
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
	Provider  string `json:"provider,omitempty"`

	// Tool use fields
	Complete bool            `json:"complete,omitempty"`
	BlockID  string          `json:"id,omitempty"`
	Name     string          `json:"name,omitempty"`
	Input    json.RawMessage `json:"input,omitempty"`

	// Tool result fields
	ToolUseID string     `json:"toolUseID,omitempty"`
	Run       *ToolRun   `json:"run,omitempty"`
	Content_  []SubBlock `json:"content,omitempty"` // Nested content for tool results
}

// SubBlock represents nested content within a tool result.
type SubBlock struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// ToolRun holds the result of a tool execution.
type ToolRun struct {
	Status string          `json:"status,omitempty"`
	Result json.RawMessage `json:"result,omitempty"` // Can be {output,exitCode}, []string, or string
}

// ToolResult holds the output of a tool execution (dict form).
type ToolResult struct {
	Output   string `json:"output,omitempty"`
	ExitCode int    `json:"exitCode,omitempty"`
}

// ParseResult tries to extract output text and exit code from the polymorphic result field.
func (r *ToolRun) ParseResult() (output string, exitCode int) {
	if r == nil || len(r.Result) == 0 {
		return "", 0
	}

	// Try dict: {"output": "...", "exitCode": 0}
	var dict ToolResult
	if err := json.Unmarshal(r.Result, &dict); err == nil && dict.Output != "" {
		return dict.Output, dict.ExitCode
	}

	// Try string
	var str string
	if err := json.Unmarshal(r.Result, &str); err == nil {
		return str, 0
	}

	// Try array of strings
	var arr []string
	if err := json.Unmarshal(r.Result, &arr); err == nil {
		return strings.Join(arr, "\n"), 0
	}

	// Fallback: raw JSON
	return string(r.Result), 0
}

// MessageState indicates the completion state of an assistant message.
type MessageState struct {
	Type       string `json:"type,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
}

// Usage holds per-message token usage information.
type Usage struct {
	Model                    string `json:"model,omitempty"`
	InputTokens              int    `json:"inputTokens,omitempty"`
	OutputTokens             int    `json:"outputTokens,omitempty"`
	CacheCreationInputTokens int    `json:"cacheCreationInputTokens,omitempty"`
	CacheReadInputTokens     int    `json:"cacheReadInputTokens,omitempty"`
	TotalInputTokens         int    `json:"totalInputTokens,omitempty"`
	Timestamp                string `json:"timestamp,omitempty"`
}

// MessageMeta holds metadata for user messages.
type MessageMeta struct {
	SentAt int64 `json:"sentAt,omitempty"` // Unix millis
}

// SentAtTime returns the sentAt time.
func (m *MessageMeta) SentAtTime() time.Time {
	if m == nil || m.SentAt == 0 {
		return time.Time{}
	}
	return time.UnixMilli(m.SentAt).Local()
}

// Env holds environment information for the thread.
type Env struct {
	Initial *EnvInitial `json:"initial,omitempty"`
}

// EnvInitial holds initial environment info including workspace trees.
type EnvInitial struct {
	Trees    []Tree    `json:"trees,omitempty"`
	Platform *Platform `json:"platform,omitempty"`
}

// Tree represents a workspace tree (project) in the Amp environment.
type Tree struct {
	DisplayName string      `json:"displayName,omitempty"`
	URI         string      `json:"uri,omitempty"` // file:// URI
	Repository  *Repository `json:"repository,omitempty"`
}

// Repository holds git repository info for a tree.
type Repository struct {
	URL string `json:"url,omitempty"`
	Ref string `json:"ref,omitempty"`
	SHA string `json:"sha,omitempty"`
}

// Platform holds platform information.
type Platform struct {
	OS            string `json:"os,omitempty"`
	Client        string `json:"client,omitempty"`
	ClientVersion string `json:"clientVersion,omitempty"`
}

// threadMeta holds cached metadata for fast session listing.
type threadMeta struct {
	ThreadID         string
	Path             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	MsgCount         int
	TotalTokens      int
	FirstUserMessage string
	Model            string
}
