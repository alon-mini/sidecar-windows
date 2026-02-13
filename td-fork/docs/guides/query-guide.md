# TDQ Query Language Guide

TDQ (Todo Query Language) is a powerful query language for searching issues in td. It supports filtering by any issue field, boolean logic, cross-entity search (logs, comments, handoffs), and relative dates.

## Quick Start

```bash
# Basic field queries
td query "status = open"
td query "type = bug"
td query "priority <= P1"

# Text search
td query "authentication"              # Searches title, description, ID
td query 'title ~ "auth"'              # Title contains "auth"

# Boolean logic
td query "type = bug AND status = open"
td query "priority = P0 OR priority = P1"
td query "NOT status = closed"

# Relative dates
td query "created >= -7d"              # Created in last 7 days
td query "updated >= today"            # Updated today
```

## Monitor Integration

In the monitor TUI (`td monitor`):
- Press `/` to start search mode
- Type any TDQ expression or plain text
- Press `?` while searching to show TDQ syntax help
- Press `Enter` to confirm, `Esc` to cancel

Plain text searches title/description/ID. TDQ syntax is auto-detected when you use operators like `=`, `~`, `AND`, `OR`, or functions like `is()`, `has()`.

## Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `=` | Exact match | `status = open` |
| `!=` | Not equal | `status != closed` |
| `~` | Contains (case-insensitive) | `title ~ "auth"` |
| `!~` | Does not contain | `title !~ "test"` |
| `<` | Less than | `priority < P2` |
| `>` | Greater than | `points > 3` |
| `<=` | Less than or equal | `priority <= P1` |
| `>=` | Greater than or equal | `created >= -7d` |

## Boolean Logic

| Syntax | Description |
|--------|-------------|
| `expr AND expr` | Both must match |
| `expr OR expr` | Either matches |
| `NOT expr` | Negation |
| `-expr` | Shorthand for NOT |
| `(expr)` | Grouping |
| `expr expr` | Implicit AND (space between expressions) |

Priority: NOT > AND > OR

## Fields

### Issue Fields

| Field | Type | Values |
|-------|------|--------|
| `id` | string | td-* format |
| `title` | string | any text |
| `description` | string | any text |
| `status` | enum | open, in_progress, blocked, in_review, closed |
| `type` | enum | bug, feature, task, epic, chore |
| `priority` | ordinal | P0, P1, P2, P3, P4 |
| `points` | number | 1, 2, 3, 5, 8, 13, 21 |
| `labels` | string | comma-separated |
| `parent` | string | direct parent issue ID |
| `epic` | string | ancestor epic ID (recursive) |
| `implementer` | string | session ID or @me |
| `reviewer` | string | session ID or @me |
| `minor` | bool | true, false |
| `branch` | string | git branch name |
| `created` | date | ISO or relative |
| `updated` | date | ISO or relative |
| `closed` | date | ISO or relative |

### Cross-Entity Fields

Search across related data:

| Prefix | Fields |
|--------|--------|
| `log.` | message, type, timestamp, session |
| `comment.` | text, created, session |
| `handoff.` | done, remaining, decisions, uncertain |
| `file.` | path, role |

Log types: progress, blocker, decision, hypothesis, tried, result
File roles: implementation, test, reference, config

### Note Fields

Notes are standalone entities (not linked to issues). Use the `note.` prefix to query notes via `ExecuteNotes`:

| Field | Type | Description |
|-------|------|-------------|
| `note.title` | string | Note title |
| `note.content` | string | Note body content |
| `note.created` | date | Creation timestamp |
| `note.updated` | date | Last update timestamp |
| `note.pinned` | bool | Pinned status (true/false) |
| `note.archived` | bool | Archived status (true/false) |

## Functions

| Function | Description | Example |
|----------|-------------|---------|
| `has(field)` | Field is not empty | `has(labels)` |
| `is(status)` | Shorthand for status check | `is(open)` |
| `any(field, v1, v2, ...)` | Field matches any value | `any(type, bug, feature)` |
| `all(field, v1, v2, ...)` | Field matches all values | `all(labels, urgent, backend)` |
| `none(field, v1, v2, ...)` | Field matches none | `none(labels, wontfix)` |
| `blocks(id)` | Issues that block given id | `blocks(td-abc)` |
| `blocked_by(id)` | Issues blocked by given id | `blocked_by(td-xyz)` |
| `child_of(id)` | Direct children of issue | `child_of(td-epic)` |
| `descendant_of(id)` | All descendants (recursive) | `descendant_of(td-epic)` |
| `linked_to(path)` | Issues linked to file path | `linked_to("cmd/query.go")` |

## Special Values

| Value | Description |
|-------|-------------|
| `@me` | Current session ID |
| `EMPTY` | Empty/null field |
| `NULL` | Null field |

## Relative Dates

| Syntax | Description |
|--------|-------------|
| `today` | Start of current day |
| `yesterday` | Start of previous day |
| `this_week` | Start of current week (Monday) |
| `last_week` | Start of previous week |
| `this_month` | Start of current month |
| `-Nd` | N days ago |
| `-Nw` | N weeks ago |
| `-Nm` | N months ago |
| `+Nd` | N days from now |

## Examples

### Basic Filtering

```bash
# All open issues
td query "status = open"

# High priority bugs
td query "type = bug AND priority <= P1"

# Large issues (8+ points)
td query "points >= 8"

# Issues with specific label
td query "labels ~ urgent"
```

### Date Queries

```bash
# Created in last 7 days
td query "created >= -7d"

# Updated today
td query "updated >= today"

# Stale open issues (not updated in 30 days)
td query "status = open AND updated < -30d"

# Closed this week
td query "status = closed AND closed >= this_week"
```

### Session Queries

```bash
# My in-progress work
td query "implementer = @me AND is(in_progress)"

# Issues I can review
td query "status = in_review AND implementer != @me"

# My recently created issues
td query "implementer = @me AND created >= -7d"
```

### Cross-Entity Queries

```bash
# Issues with blocker logs
td query "log.type = blocker"

# Issues mentioning "fixed" in logs
td query 'log.message ~ "fixed"'

# Issues with approved comments
td query 'comment.text ~ "approved"'

# Issues with remaining TODOs in handoff
td query 'handoff.remaining ~ "TODO"'

# Issues with test files linked
td query "file.role = test"
```

### Note Queries

Notes use the `note.` prefix and are queried separately from issues via `ExecuteNotes`:

```bash
# Find notes by title
note.title ~ "meeting"

# Search note content
note.content ~ "project timeline"

# Pinned notes only
note.pinned = true

# Non-archived notes
NOT note.archived = true

# Recently updated notes
note.updated >= -7d

# Combined: pinned notes mentioning "design"
note.pinned = true AND note.content ~ "design"

# Bare text search across note title, content, and ID
"important"
```

### Complex Queries

```bash
# High priority bugs created this week, not in review
td query "type = bug AND priority <= P1 AND created >= this_week AND status != in_review"

# Open features with no labels
td query "type = feature AND is(open) AND NOT has(labels)"

# All tasks in an epic that are blocked
td query "descendant_of(td-epic1) AND is(blocked)"

# Bugs or features with urgent label
td query "(type = bug OR type = feature) AND labels ~ urgent"
```

## CLI Options

```bash
td query "EXPRESSION" [flags]

Flags:
  -o, --output string   Output format: table, json, ids, count (default "table")
  -n, --limit int       Limit results (default 50)
      --sort string     Sort by field (prefix with - for descending)
      --explain         Show query parsing without executing
      --examples        Show query examples
      --fields          List all searchable fields
```

## Inline Sort

You can add sort clauses directly in the query string:

```bash
# Sort by priority (ascending)
td query "type = bug sort:priority"

# Sort by priority descending
td query "type = bug sort:-priority"

# Multiple sort fields
td query "status = open sort:-priority sort:created"
```

Prefix with `-` for descending order. Multiple `sort:` clauses are applied in order.

## Tips

1. **Enum values are case-insensitive**: `priority = p0` and `priority = P0` both work, as do `status = OPEN`, `type = Bug`, etc.
2. **Quote strings with spaces**: `title ~ "multi word search"`
3. **Use functions for common patterns**: `is(open)` instead of `status = open`
4. **Combine with AND for precision**: `type = bug AND priority = P0 AND created >= -7d`
5. **Implicit AND**: `type = bug priority = P0` is equivalent to `type = bug AND priority = P0`
6. **Use cross-entity search to find issues**: `log.type = blocker` finds issues with blockers
7. **Inline sort**: Add `sort:field` or `sort:-field` to any query for ordering
8. **In monitor, plain text still works**: Just type to do simple text search
