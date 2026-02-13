package geminicli

import (
	"time"
)

// Session represents a Gemini CLI session JSON file.
type Session struct {
	SessionID   string    `json:"sessionId"`
	ProjectHash string    `json:"projectHash"`
	StartTime   time.Time `json:"startTime"`
	LastUpdated time.Time `json:"lastUpdated"`
	Messages    []Message `json:"messages"`
}

// Message represents a message in a Gemini CLI session.
type Message struct {
	ID        string     `json:"id"`
	Timestamp time.Time  `json:"timestamp"`
	Type      string     `json:"type"` // "user", "gemini", "info"
	Content   string     `json:"content"`
	Model     string     `json:"model,omitempty"`
	Tokens    *Tokens    `json:"tokens,omitempty"`
	ToolCalls []ToolCall `json:"toolCalls,omitempty"`
	Thoughts  []Thought  `json:"thoughts,omitempty"`
}

// Tokens holds token usage for a message.
type Tokens struct {
	Input    int `json:"input"`
	Output   int `json:"output"`
	Cached   int `json:"cached"`
	Thoughts int `json:"thoughts"`
	Tool     int `json:"tool"`
	Total    int `json:"total"`
}

// ToolCall represents a tool invocation.
type ToolCall struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Args        any    `json:"args,omitempty"`
	Result      any    `json:"result,omitempty"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

// Thought represents a thinking step from Gemini.
type Thought struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// SessionMetadata holds parsed metadata about a session file.
type SessionMetadata struct {
	Path             string
	SessionID        string
	ProjectHash      string
	StartTime        time.Time
	LastUpdated      time.Time
	MsgCount         int
	TotalTokens      int
	EstCost          float64
	PrimaryModel     string
	FirstUserMessage string // Content of the first user message (for title)
}
