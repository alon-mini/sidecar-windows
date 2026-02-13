# Warp Terminal SQLite Schema Reference

## Database Location

```
~/Library/Group Containers/2BBY89MBSN.dev.warp/Library/Application Support/dev.warp.Warp-Stable/warp.sqlite
```

Uses WAL mode. For consistent reads, include: `warp.sqlite`, `warp.sqlite-shm`, `warp.sqlite-wal`

## Key Tables

### ai_queries

User queries sent to AI. One row per exchange (turn).

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | Primary key |
| `exchange_id` | TEXT | Unique ID for this exchange |
| `conversation_id` | TEXT | Groups exchanges into conversations |
| `start_ts` | DATETIME | Timestamp of query |
| `input` | TEXT | JSON array with Query object |
| `working_directory` | TEXT | Project path (for filtering) |
| `output_status` | TEXT | "Completed", etc. |
| `model_id` | TEXT | Model used |
| `planning_model_id` | TEXT | Planning model if different |
| `coding_model_id` | TEXT | Coding model if different |

**Input JSON**: Only the first exchange has the full Query (with `text`, `context` array containing Directory, Git, ProjectRules, CurrentTime, ExecutionEnvironment, Codebase). Subsequent exchanges have empty `[]` input.

### agent_conversations

Usage metadata per conversation.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | Primary key |
| `conversation_id` | TEXT | Links to ai_queries |
| `conversation_data` | TEXT | JSON with usage stats |
| `last_modified_at` | TIMESTAMP | Auto-updated |

`conversation_data` JSON contains: `server_conversation_token`, `conversation_usage_metadata` (with `credits_spent`, `token_usage` array, `tool_usage_metadata`).

### agent_tasks

Protobuf-encoded task lists. Currently not decoded by the Warp adapter.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | Primary key |
| `conversation_id` | TEXT | Links to conversation |
| `task_id` | TEXT | Unique task ID |
| `task` | BLOB | Protobuf-encoded task data |
| `last_modified_at` | TIMESTAMP | Auto-updated |

### blocks

Terminal command blocks. Links to AI via `ai_metadata`.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | Primary key |
| `pane_leaf_uuid` | BLOB | Terminal pane reference |
| `stylized_command` | BLOB | ANSI-encoded command text |
| `stylized_output` | BLOB | ANSI-encoded output text |
| `pwd` | TEXT | Working directory |
| `exit_code` | INTEGER | Command exit code |
| `start_ts` | DATETIME | Command start time |
| `completed_ts` | DATETIME | Completion time |
| `ai_metadata` | TEXT | JSON with `action_id`, `conversation_id`, etc. |

ANSI stripping regex: `\x1b\[[0-9;]*m`

## Critical Limitation

**Warp does NOT store AI assistant text responses locally.** Only available: user queries, tool executions (blocks with ai_metadata), usage statistics. Not stored: AI reasoning/response text, thinking content, full conversation transcript.

## Model IDs

| Warp Model ID | Display Name |
|---------------|--------------|
| `claude-4-5-opus` | Claude Opus 4.5 |
| `claude-4-5-opus-thinking` | Claude Opus 4.5 (Thinking) |
| `gpt-5-1-high-reasoning` | GPT-5 |

## Useful Queries

List conversations by project:
```sql
SELECT DISTINCT conversation_id, working_directory, model_id,
       MIN(start_ts) as first_msg, MAX(start_ts) as last_msg,
       COUNT(*) as exchange_count
FROM ai_queries
WHERE working_directory LIKE ? OR working_directory = ?
GROUP BY conversation_id
ORDER BY last_msg DESC;
```

Get blocks for a conversation:
```sql
SELECT id, stylized_command, stylized_output, exit_code, ai_metadata
FROM blocks
WHERE ai_metadata LIKE '%"conversation_id":"<uuid>"%'
ORDER BY start_ts;
```

## Other Tables

| Table | Purpose |
|-------|---------|
| `terminal_panes` | Terminal session state, `conversation_ids` field |
| `commands` | Command history (not AI-specific) |
| `projects` | Known project paths |
| `workspace_metadata` | Workspace navigation history |
| `mcp_server_installations` | MCP server configs |

## File Watching

Watch `warp.sqlite-wal` for changes. Debounce rapid writes (100ms recommended).
