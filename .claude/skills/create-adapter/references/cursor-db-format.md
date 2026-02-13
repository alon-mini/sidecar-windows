# Cursor Agent Database Format Reference

## Database Location

Per-session SQLite databases:
```
~/.cursor/chats/<workspace-hash>/<session-id>/store.db
```
- **workspace-hash**: MD5 hash of absolute project path (32 hex chars)
- **session-id**: UUID (directory name IS the session UUID)

Uses WAL mode. For consistent reads, include: `store.db`, `store.db-shm`, `store.db-wal`

## Schema

### meta Table

| Column | Type | Description |
|--------|------|-------------|
| `key` | TEXT | Always "0" for session metadata |
| `value` | TEXT | Hex-encoded JSON |

Decode: `hex.DecodeString(value)` -> JSON

Metadata JSON fields:
- `agentId` - Session UUID (same as folder name)
- `latestRootBlobId` - Hash of root blob in message tree
- `name` - User-visible session name
- `mode` - Agent mode: "auto-run", "manual", etc.
- `createdAt` - Unix timestamp in milliseconds
- `lastUsedModel` - Model identifier string

### blobs Table

Merkle tree message storage.

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT | SHA-256 hash (64 hex chars) |
| `data` | BLOB | Binary data (JSON or linking blob) |

## Blob Types

### Message Blobs (JSON)
Start with `{` byte. Contain conversation messages.

**User message**: `role: "user"`, content may contain `<user_query>` XML tags. Extract actual query with:
```go
userQueryRegex := regexp.MustCompile(`<user_query>\s*([\s\S]*?)\s*</user_query>`)
```

**Assistant message**: `role: "assistant"`, content is array of blocks. All assistant messages have `"id": "1"` (Cursor quirk - use blob hash as unique ID instead).

**Tool result**: `role: "tool"`, content array with `tool-result` blocks. Result can be string or array of content blocks.

**System message**: `role: "system"`, filtered during parsing if no user content.

### Linking Blobs (Binary)
Start with `0x0A 0x20` bytes. Contain references to child blobs.

Format: `[0x0A][0x20][32-byte SHA-256 hash]` repeated, optional embedded JSON at end.

## Content Block Types

| Type | Description | Key Fields |
|------|-------------|------------|
| `text` | Text response | `text` |
| `reasoning` | Thinking/reasoning | `text`, `signature` |
| `tool-call` | Tool invocation | `toolCallId`, `toolName`, `args` |
| `tool-result` | Tool output | `toolCallId`, `result`, `isError` |

## Important Quirks

- **All assistant messages have id="1"**: Use blob hash as unique message ID
- **Tool results as array**: Result field can be string or array of content blocks
- **System-context messages filtered**: Messages with only system context (no user content) are skipped
- **Timestamps not stored per-message**: Interpolate from `meta.createdAt` and file mtime

## Tree Traversal

1. Get `latestRootBlobId` from meta table
2. Load all blobs into `map[hash]data`
3. Recursively traverse from root:
   - If `data[0] == '{'`: JSON message blob, parse it
   - If starts with `0x0A 0x20`: linking blob, follow child references (34-byte chunks)
   - Check for embedded JSON after all references in linking blobs

## Model Identifiers

| Cursor Model ID | Display Name |
|-----------------|--------------|
| `claude-4.5-opus-high-thinking` | Claude Opus 4.5 (Thinking) |
| `claude-4-sonnet` | Claude Sonnet 4 |
| `gpt-4o` | GPT-4o |
| `gpt-4.1` | GPT-4.1 |
| `gemini-2.5-pro` | Gemini 2.5 Pro |

## File Watching

Watch `store.db-wal` for changes. Debounce rapid writes (100ms recommended). Re-read metadata to get updated `latestRootBlobId`, then traverse tree for new messages.
