# Task: Create OpenClaw Adapter for Sidecar Conversations Plugin

## Goal
Build a new adapter that reads OpenClaw session JSONL files so they appear in sidecar's conversations plugin alongside Claude Code and Codex sessions. This lets OpenClaw users see their AI agent conversations directly in sidecar.

## Background
OpenClaw is an AI agent platform (like Claude Code but with messaging, cron jobs, tools, etc). It stores sessions as JSONL files in `~/.openclaw/agents/main/sessions/`. The format is similar to Claude Code's but with key differences. There are already 211+ sessions totaling ~14MB in the test environment.

## Architecture

### Existing Pattern
Follow the Claude Code adapter (`internal/adapter/claudecode/`) as the primary template. The file structure should be:

```
internal/adapter/openclaw/
├── adapter.go          # Main adapter implementation
├── adapter_test.go     # Tests
├── doc.go              # Package doc
├── register.go         # init() registration
├── search.go           # Content search (optional, can add later)
├── search_test.go
├── stats.go            # Cost estimation
├── types.go            # OpenClaw-specific types
└── watcher.go          # fsnotify watcher
```

### Watch Scope
OpenClaw is a **global-scope** adapter (like Codex). All sessions live in one directory (`~/.openclaw/agents/main/sessions/`) regardless of project. The adapter should implement `WatchScopeProvider` and return `WatchScopeGlobal`.

### Session Filtering
Unlike Claude Code (which has per-project directories), OpenClaw stores ALL sessions in one flat directory. Each session's first JSONL line is a `type: "session"` header containing the `cwd`. The adapter must:
1. Read the session header to get `cwd`
2. Only return sessions whose `cwd` matches the `projectRoot` passed to `Sessions()`
3. Cache the cwd mapping to avoid re-reading headers on every call

## OpenClaw JSONL Format

### Line Types

#### Session Header (first line)
```json
{
  "type": "session",
  "version": 3,
  "id": "ec314c57-bdad-4045-ba03-c0c57add5291",
  "timestamp": "2026-02-01T22:02:17.040Z",
  "cwd": "/Users/marcusvorwaller/.openclaw/workspace"
}
```

#### Model Change
```json
{
  "type": "model_change",
  "id": "6ef314c4",
  "parentId": null,
  "timestamp": "2026-02-03T17:35:22.200Z",
  "provider": "anthropic",
  "modelId": "claude-opus-4-5"
}
```

#### Thinking Level Change
```json
{
  "type": "thinking_level_change",
  "id": "93a5a453",
  "parentId": "6ef314c4",
  "timestamp": "2026-02-03T17:35:22.200Z",
  "thinkingLevel": "low"
}
```

#### Custom Events (model snapshots, etc.)
```json
{
  "type": "custom",
  "customType": "model-snapshot",
  "data": {"timestamp": 1770140122200, "provider": "anthropic", "modelApi": "anthropic-messages", "modelId": "claude-opus-4-5"},
  "id": "e0158071",
  "parentId": "30de8b3f"
}
```

#### User Message
```json
{
  "type": "message",
  "id": "d0759492",
  "parentId": "30de8b3f",
  "timestamp": "2026-02-01T22:06:59.076Z",
  "message": {
    "role": "user",
    "content": [
      {"type": "text", "text": "the actual user message"},
      {"type": "image", "...": "..."}
    ]
  }
}
```

#### Assistant Message (with thinking + text + tool calls)
```json
{
  "type": "message",
  "id": "e33712cc",
  "parentId": "e5241729",
  "timestamp": "2026-02-01T22:06:59.076Z",
  "message": {
    "role": "assistant",
    "content": [
      {
        "type": "thinking",
        "thinking": "Let me think about this...",
        "thinkingSignature": "ErACCkYI..."
      },
      {
        "type": "text",
        "text": "Here's my response."
      },
      {
        "type": "toolCall",
        "id": "toolu_014idxUeCvvdBYWfhCHzWyJV",
        "name": "edit",
        "arguments": {"path": "/some/file.md", "oldText": "...", "newText": "..."}
      }
    ],
    "api": "anthropic-messages",
    "provider": "anthropic",
    "model": "claude-opus-4-5",
    "usage": {
      "input": 15000,
      "output": 2000,
      "cacheRead": 12000,
      "cacheWrite": 3000,
      "totalTokens": 17000,
      "cost": {
        "input": 0.045,
        "output": 0.15,
        "cacheRead": 0.0036,
        "cacheWrite": 0.01125,
        "total": 0.20985
      }
    },
    "stopReason": "end_turn",
    "timestamp": 1770140122202
  }
}
```

#### Tool Result
```json
{
  "type": "message",
  "id": "02316264",
  "parentId": "e33712cc",
  "timestamp": "2026-02-01T22:06:59.090Z",
  "message": {
    "role": "toolResult",
    "toolCallId": "toolu_014idxUeCvvdBYWfhCHzWyJV",
    "toolName": "edit",
    "content": [
      {"type": "text", "text": "Successfully replaced text in /some/file.md."}
    ],
    "details": {
      "diff": "  1 - old line\n  1 + new line"
    }
  }
}
```

## Key Differences from Claude Code Format

| Aspect | Claude Code | OpenClaw |
|--------|------------|----------|
| Message line type | `"type": "user"` or `"type": "assistant"` | Always `"type": "message"`, role in `message.role` |
| Role values | `user`, `assistant` | `user`, `assistant`, `toolResult` |
| Tool calls | `content[].type = "tool_use"` | `content[].type = "toolCall"` |
| Tool call input | `content[].input` (any) | `content[].arguments` (object) |
| Tool results | Embedded as `tool_result` blocks in next `user` message | Separate `message` with `role: "toolResult"` |
| Tool result linking | `content[].tool_use_id` | `message.toolCallId` |
| ID field | `uuid` | `id` |
| Parent ID | `parentUuid` | `parentId` |
| Session ID | `sessionId` (repeated in every line) | Only in `type: "session"` header line |
| Usage tokens | `usage.input_tokens`, `usage.output_tokens` | `usage.input`, `usage.output` |
| Cache tokens | `usage.cache_read_input_tokens`, `usage.cache_creation_input_tokens` | `usage.cacheRead`, `usage.cacheWrite` |
| Cost | Calculated from token counts | Pre-calculated in `usage.cost.total` |
| Session directory | `~/.claude/projects/<encoded-path>/` (per-project) | `~/.openclaw/agents/main/sessions/` (global) |
| Session metadata | Scattered across message lines (`cwd`, `version`, `gitBranch`) | Concentrated in `type: "session"` header + `type: "model_change"` lines |
| Sub-agents | Filename prefix `agent-` | Unknown — may need to detect from session content or ignore initially |
| Thinking blocks | `content[].type = "thinking"`, `content[].thinking` | Same: `content[].type = "thinking"`, `content[].thinking` (also has `thinkingSignature`) |

## Implementation Guide

### types.go

Define OpenClaw-specific types:

```go
package openclaw

import (
    "encoding/json"
    "time"
)

// RawLine represents any JSONL line from an OpenClaw session.
type RawLine struct {
    Type      string          `json:"type"`
    ID        string          `json:"id"`
    ParentID  string          `json:"parentId"`
    Timestamp time.Time       `json:"timestamp"`
    Message   *MessageContent `json:"message,omitempty"`

    // Session header fields (type="session")
    Version int    `json:"version,omitempty"`
    CWD     string `json:"cwd,omitempty"`

    // Model change fields (type="model_change")
    Provider string `json:"provider,omitempty"`
    ModelID  string `json:"modelId,omitempty"`

    // Custom event fields (type="custom")
    CustomType string          `json:"customType,omitempty"`
    Data       json.RawMessage `json:"data,omitempty"`
}

type MessageContent struct {
    Role       string          `json:"role"`    // "user", "assistant", "toolResult"
    Content    json.RawMessage `json:"content"` // array of content blocks
    Model      string          `json:"model,omitempty"`
    Provider   string          `json:"provider,omitempty"`
    API        string          `json:"api,omitempty"`
    Usage      *Usage          `json:"usage,omitempty"`
    StopReason string          `json:"stopReason,omitempty"`
    ToolCallID string          `json:"toolCallId,omitempty"` // for toolResult role
    ToolName   string          `json:"toolName,omitempty"`   // for toolResult role
    Details    *Details        `json:"details,omitempty"`    // for toolResult extra info
}

type Usage struct {
    Input       int      `json:"input"`
    Output      int      `json:"output"`
    CacheRead   int      `json:"cacheRead"`
    CacheWrite  int      `json:"cacheWrite"`
    TotalTokens int      `json:"totalTokens"`
    Cost        *Cost    `json:"cost,omitempty"`
}

type Cost struct {
    Input      float64 `json:"input"`
    Output     float64 `json:"output"`
    CacheRead  float64 `json:"cacheRead"`
    CacheWrite float64 `json:"cacheWrite"`
    Total      float64 `json:"total"`
}

type Details struct {
    Diff string `json:"diff,omitempty"`
}

// ContentBlock represents a block in the content array.
type ContentBlock struct {
    Type              string          `json:"type"`                         // "text", "thinking", "toolCall", "image"
    Text              string          `json:"text,omitempty"`               // for text blocks
    Thinking          string          `json:"thinking,omitempty"`           // for thinking blocks
    ThinkingSignature string          `json:"thinkingSignature,omitempty"`  // for thinking blocks
    ID                string          `json:"id,omitempty"`                 // for toolCall (the tool_use_id)
    Name              string          `json:"name,omitempty"`               // for toolCall (tool name)
    Arguments         json.RawMessage `json:"arguments,omitempty"`          // for toolCall (tool input)
}

type SessionMetadata struct {
    Path             string
    SessionID        string
    CWD              string
    Version          int
    FirstMsg         time.Time
    LastMsg          time.Time
    MsgCount         int
    TotalTokens      int
    EstCost          float64
    PrimaryModel     string
    FirstUserMessage string
}
```

### adapter.go Key Decisions

1. **Directory**: `~/.openclaw/agents/main/sessions/` — hardcode for now, could make configurable later
2. **Detect()**: Check if the directory exists and has `.jsonl` files. Also check if any session's `cwd` matches `projectRoot`
3. **Sessions()**: Read all `.jsonl` files, parse header to get `cwd`, filter by `projectRoot` match, return sorted by UpdatedAt
4. **Messages()**: Parse message lines, map `toolCall` → `tool_use` in adapter types, link `toolResult` messages to their corresponding `toolCall` blocks
5. **Usage()**: OpenClaw pre-calculates cost in `usage.cost.total` — use it directly instead of estimating from model rates

### Tool Result Linking

This is the trickiest part. In Claude Code, tool results are embedded in the next `user` message as `tool_result` content blocks with a `tool_use_id`. In OpenClaw, tool results are separate `message` lines with `role: "toolResult"` and a `toolCallId` field.

Strategy:
- When parsing, track pending tool calls by their `id` (the `toolCall` block's `id` field)
- When a `toolResult` message arrives, look up the `toolCallId` in the pending map
- Link the result content back to the tool call in the assistant message
- This is similar to what the Claude Code adapter does, just with different field names

### register.go

```go
package openclaw

import "github.com/marcus/sidecar/internal/adapter"

func init() {
    adapter.RegisterFactory(func() adapter.Adapter {
        return New()
    })
}
```

### watcher.go

Watch `~/.openclaw/agents/main/sessions/` for new/modified `.jsonl` files. Since it's a global directory, implement `WatchScopeProvider` returning `WatchScopeGlobal` to avoid duplicate watchers.

The watcher implementation can be nearly identical to the Claude Code watcher — fsnotify on the sessions directory, debounce rapid events, emit `EventSessionCreated`/`EventMessageAdded`.

### CWD Caching

Since all sessions are in one directory but we need to filter by project, maintain a `cwdCache map[string]string` mapping session file path → CWD. This avoids re-reading session headers on every `Sessions()` call. Invalidate entries when files are modified (check mtime).

### Performance Considerations

- 211 sessions currently, will grow. The CWD filter means we scan all session headers even for a specific project.
- Use a two-level cache: (1) CWD cache for filtering, (2) metadata cache for session details
- The CWD is always in the first line of the file — fast to read
- Incremental metadata parsing (like Claude Code adapter) works here too since JSONL is append-only

### Icon & Name

```go
func (a *Adapter) ID() string   { return "openclaw" }
func (a *Adapter) Name() string { return "OpenClaw" }
func (a *Adapter) Icon() string { return "🐾" }
```

## Test Data

Real OpenClaw session files are at `~/.openclaw/agents/main/sessions/`. For tests, create minimal JSONL fixtures:

```jsonl
{"type":"session","version":3,"id":"test-session-1","timestamp":"2026-02-01T00:00:00Z","cwd":"/test/project"}
{"type":"model_change","id":"mc1","parentId":null,"timestamp":"2026-02-01T00:00:01Z","provider":"anthropic","modelId":"claude-opus-4-5"}
{"type":"message","id":"m1","parentId":"mc1","timestamp":"2026-02-01T00:00:02Z","message":{"role":"user","content":[{"type":"text","text":"Hello, what files are in this directory?"}]}}
{"type":"message","id":"m2","parentId":"m1","timestamp":"2026-02-01T00:00:03Z","message":{"role":"assistant","content":[{"type":"thinking","thinking":"Let me check the files."},{"type":"text","text":"Let me look at the directory contents."},{"type":"toolCall","id":"tc1","name":"exec","arguments":{"command":"ls"}}],"model":"claude-opus-4-5","usage":{"input":500,"output":100,"cacheRead":400,"cacheWrite":100,"totalTokens":600,"cost":{"total":0.015}}}}
{"type":"message","id":"m3","parentId":"m2","timestamp":"2026-02-01T00:00:04Z","message":{"role":"toolResult","toolCallId":"tc1","toolName":"exec","content":[{"type":"text","text":"file1.go\nfile2.go\nREADME.md"}]}}
{"type":"message","id":"m4","parentId":"m3","timestamp":"2026-02-01T00:00:05Z","message":{"role":"assistant","content":[{"type":"text","text":"I can see three files: file1.go, file2.go, and README.md."}],"model":"claude-opus-4-5","usage":{"input":600,"output":50,"cacheRead":500,"cacheWrite":100,"totalTokens":650,"cost":{"total":0.012}}}}
```

## Files to Create

1. `internal/adapter/openclaw/types.go` — Type definitions
2. `internal/adapter/openclaw/adapter.go` — Main adapter (Detect, Sessions, Messages, Usage, Watch)
3. `internal/adapter/openclaw/adapter_test.go` — Unit tests with fixture data
4. `internal/adapter/openclaw/register.go` — Auto-registration via init()
5. `internal/adapter/openclaw/doc.go` — Package documentation
6. `internal/adapter/openclaw/watcher.go` — fsnotify file watcher
7. `internal/adapter/openclaw/stats.go` — Cost calculation (can use pre-calculated costs)
8. `internal/adapter/openclaw/testdata/` — Test fixture JSONL files

## Verification

After implementation:
1. `go test ./internal/adapter/openclaw/...`
2. `go build ./...` — ensure no import cycles
3. Open sidecar in `~/.openclaw/workspace` — OpenClaw sessions should appear in conversations plugin with 🐾 badge
4. Verify session list shows correct titles, timestamps, costs
5. Verify message view shows user messages, assistant responses, thinking blocks, and tool calls with results
6. Verify live updates work (watcher detects new messages in active sessions)

## Community Impact
This adapter would let any OpenClaw user see their agent conversations in sidecar. The OpenClaw community would benefit significantly from this integration — it's a natural fit since both tools are used by AI-assisted developers.
