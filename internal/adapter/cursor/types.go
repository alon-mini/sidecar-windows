package cursor

import (
	"encoding/json"
	"time"
)

// SessionMeta holds metadata stored in the meta table at key "0".
// The value is hex-encoded JSON.
type SessionMeta struct {
	AgentID          string `json:"agentId"`
	LatestRootBlobID string `json:"latestRootBlobId"`
	Name             string `json:"name"`
	Mode             string `json:"mode"`
	CreatedAt        int64  `json:"createdAt"` // Unix timestamp in milliseconds
	LastUsedModel    string `json:"lastUsedModel"`
}

// CreatedTime returns the createdAt timestamp as time.Time.
func (m SessionMeta) CreatedTime() time.Time {
	return time.UnixMilli(m.CreatedAt)
}

// MessageBlob represents a raw message blob from the SQLite database.
// Messages are stored as JSON with role and content fields.
type MessageBlob struct {
	ID      string          `json:"id,omitempty"`
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// ContentBlock represents a single block in the content array.
// Content can be a plain string or an array of these blocks.
type ContentBlock struct {
	Type            string          `json:"type"`
	Text            string          `json:"text,omitempty"`
	ToolCallID      string          `json:"toolCallId,omitempty"`
	ToolName        string          `json:"toolName,omitempty"`
	Args            json.RawMessage `json:"args,omitempty"`
	Result          json.RawMessage `json:"result,omitempty"`  // For tool-result blocks
	IsError         bool            `json:"isError,omitempty"` // For tool-result error status
	ProviderOptions *ProviderOpts   `json:"providerOptions,omitempty"`
	Signature       string          `json:"signature,omitempty"`
}

// ProviderOpts holds provider-specific options embedded in content blocks.
type ProviderOpts struct {
	Cursor *CursorOpts `json:"cursor,omitempty"`
}

// CursorOpts holds Cursor-specific message metadata.
type CursorOpts struct {
	ModelName       string `json:"modelName,omitempty"`
	RawToolCallArgs string `json:"rawToolCallArgs,omitempty"`
}

// SessionInfo holds parsed session information for display.
type SessionInfo struct {
	Path             string    // Path to store.db
	SessionID        string    // UUID of the session
	WorkspaceHash    string    // Hash of the workspace path
	Name             string    // Session name from metadata
	Mode             string    // Agent mode (e.g., "auto-run")
	Model            string    // Last used model
	CreatedAt        time.Time // When the session was created
	UpdatedAt        time.Time // Estimated from file mtime or blob analysis
	MessageCount     int       // Number of user/assistant messages
	TotalTokens      int       // Estimated token count (if available)
	FirstUserMessage string    // Content of the first user message (for title)
}
