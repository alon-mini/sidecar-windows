# td-sync Admin API Plan

Work required in this repo to support the td-watch admin module. td-watch is a SvelteKit web app (BFF + frontend) that proxies all data access through td-sync's HTTP API. td-watch never touches SQLite directly.

**td-watch repo:** `~/code/td-watch/`
**Full platform spec:** `~/code/td-watch/docs/sync-admin-web-spec.md`

## Architecture Context

```
Browser → td-watch (SvelteKit BFF) → td-sync admin API → server.db / events.db
```

td-watch's server routes authenticate to td-sync using a dedicated admin API key. The key is stored as an env var in td-watch, never exposed to the browser. td-sync is the sole database accessor — td-watch is a session-aware proxy.

## Current State (as of 2025-02-07)

What exists today in td-sync:

- **Database:** `server.db` (users, api_keys, projects, memberships, sync_cursors, auth_requests) + per-project `events.db`
- **Auth:** Device-flow auth, API keys with comma-separated scopes (only "sync" used), per-project roles (owner/writer/reader)
- **Endpoints:** healthz, metricz, auth flow, project CRUD, membership CRUD, sync push/pull/status/snapshot
- **Metrics:** In-memory atomic counters, reset on restart
- **Rate limiting:** In-memory fixed-window buckets, violations not persisted
- **Pagination:** None on project/member list endpoints; offset-based on sync pull only
- **CORS:** None
- **Admin concept:** Does not exist
- **Schema version:** 2

What does NOT exist:

- `is_admin` flag or any server-wide admin role
- Admin CLI commands
- Admin scopes on API keys
- Admin middleware
- Any `/v1/admin/*` endpoints
- `auth_events` table (auth_requests are temporary, cleaned up after 1 hour)
- `rate_limit_events` table
- Cursor-based pagination
- CORS headers
- Persistent metrics
- Entity type list endpoint

## Work Items

### 1. Schema Migration (v2 → v3)

Add to `internal/serverdb/schema.go`:

**`is_admin` column on `users`:**
```sql
ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT 0;
```

**`auth_events` table:**
```sql
CREATE TABLE auth_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    auth_request_id TEXT NOT NULL,
    email TEXT NOT NULL,
    event_type TEXT NOT NULL,  -- started, code_verified, key_issued, expired, failed
    metadata TEXT DEFAULT '{}', -- JSON: IP, user_agent, failure_reason, etc.
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_auth_events_type ON auth_events(event_type);
CREATE INDEX idx_auth_events_email ON auth_events(email);
CREATE INDEX idx_auth_events_created ON auth_events(created_at);
```

**`rate_limit_events` table:**
```sql
CREATE TABLE rate_limit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id TEXT,         -- NULL for IP-based limits (auth endpoints)
    ip TEXT,
    endpoint_class TEXT NOT NULL,  -- auth, push, pull, other
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_rle_created ON rate_limit_events(created_at);
CREATE INDEX idx_rle_key ON rate_limit_events(key_id);
```

Migration function pattern: follow existing `migrateV1ToV2` in schema.go. Add `migrateV2ToV3`.

First-user-is-admin: when creating the first user (check `SELECT COUNT(*) FROM users`), set `is_admin = 1`.

### 2. Admin CLI Commands

New subcommand group: `td-sync admin`. Add to `cmd/td-sync/`:

- **`td-sync admin grant --email <email>`** — sets `is_admin = 1` on the user. Errors if user doesn't exist.
- **`td-sync admin revoke --email <email>`** — sets `is_admin = 0`. Errors if user doesn't exist or is the last admin.

These operate directly on `server.db` (same process or via direct DB open if server isn't running). Simple `UPDATE users SET is_admin = ? WHERE email = ?`.

### 3. Admin Scopes and API Key Changes

Current scopes: comma-separated string, only "sync" used.

New admin scopes:
- `admin:read:server` — server overview, config, metrics
- `admin:read:projects` — list/inspect any project, members, sync status
- `admin:read:events` — view event streams for any project
- `admin:read:snapshots` — query derived state for any project
- `admin:export` — download/export event data

No changes to the `api_keys` table schema needed — scopes are already a text field. The `td-sync admin grant` command should also support creating an admin API key:

- **`td-sync admin create-key --email <email> --scopes "admin:read:server,admin:read:projects,..." --name "td-watch"`** — creates an API key with admin scopes for a user who has `is_admin = 1`. Returns the key (shown once). This is what operators use to configure `SYNC_ADMIN_API_KEY` in td-watch's environment.

### 4. Admin Middleware

New middleware in `internal/api/`:

**`requireAdmin(scope string)`** — checks:
1. Request has valid API key (existing `requireAuth` does this)
2. The key's user has `is_admin = 1`
3. The key's scopes include the required scope

Returns 403 with `{"error": {"code": "insufficient_admin_scope", "message": "..."}}` if any check fails.

Implementation: add a `IsAdmin` field to the `AuthUser` struct populated during auth middleware. `requireAdmin` wraps `requireAuth` and adds the admin+scope check.

### 5. Auth Event Logging

Instrument the device-flow auth lifecycle to write to `auth_events`:

| Auth step | event_type | Where in code |
|-----------|-----------|---------------|
| `/v1/auth/login/start` | `started` | `handleAuthLoginStart` |
| `/auth/verify` POST (success) | `code_verified` | `handleAuthVerify` |
| `/v1/auth/login/poll` (key issued) | `key_issued` | `handleAuthLoginPoll` |
| Cleanup goroutine (expired) | `expired` | `CleanupExpiredAuthRequests` |
| `/auth/verify` POST (wrong code) | `failed` | `handleAuthVerify` |

Metadata JSON: include IP address, user_agent where available, failure_reason for failed events.

**Retention:** Background goroutine deletes rows older than 90 days (configurable via `SYNC_AUTH_EVENT_RETENTION`). Can run alongside the existing auth_requests cleanup.

**Performance:** These are simple INSERT statements on an append-only table with no unique constraints. Acceptable to do synchronously since auth endpoints are low-throughput. If needed later, batch via a channel.

### 6. Rate Limit Event Logging

In `internal/api/ratelimit.go`, when a 429 is returned, INSERT into `rate_limit_events`:

```go
// In the rate limit check, after determining the request should be rejected:
db.InsertRateLimitEvent(keyID, ip, endpointClass)
```

**Retention:** Background goroutine deletes rows older than 30 days (configurable via `SYNC_RATE_LIMIT_EVENT_RETENTION`).

**Performance:** 429s are relatively rare. Synchronous INSERT is fine. If a burst of violations causes concern, buffer writes via a channel with a small batch window.

### 7. CORS Configuration

New env var: `SYNC_CORS_ALLOWED_ORIGINS` (comma-separated, default empty = CORS disabled).

Add CORS middleware in `internal/api/server.go` (applied to `/v1/admin/*` routes):
- Parse allowed origins on startup
- On request: check `Origin` header against allowlist
- Set `Access-Control-Allow-Origin`, `Access-Control-Allow-Headers` (Authorization, Content-Type), `Access-Control-Allow-Methods` (GET, POST, OPTIONS)
- Handle `OPTIONS` preflight requests with 204

Note: in the standard deployment, browser never talks directly to td-sync (td-watch BFF proxies everything server-to-server). CORS is a safety net for future direct-access scenarios.

### 8. Cursor-Based Pagination

New pagination contract for all list endpoints:

**Request:** `?cursor=<opaque>&limit=<int>`
**Response:**
```json
{
  "data": [...],
  "next_cursor": "string|null",
  "has_more": true|false
}
```

Default limit: 50. Max limit: 200.

**Cursor encoding:** base64-encoded JSON with the sort key value (e.g., `{"id":"..."}` or `{"created_at":"...","id":"..."}`). Opaque to clients.

**Retrofit existing endpoints:**
- `GET /v1/projects` — currently returns all user's projects as array
- `GET /v1/projects/{id}/members` — currently returns all members as array

**Backwards compatibility:** omitting `cursor` and `limit` returns all results (existing behavior). Existing clients won't break.

**Helper:** create a generic pagination helper in `internal/serverdb/` that all list queries use:
```go
func PaginatedQuery[T any](db *sql.DB, baseQuery string, args []any, limit int, cursor string, scanRow func(*sql.Rows) (T, error)) (*PaginatedResult[T], error)
```

### 9. Admin Endpoints

All require `requireAdmin(scope)` middleware. All follow the standard error contract. All list endpoints use cursor-based pagination.

#### 9.1 Server-Level (scope: `admin:read:server`)

**`GET /v1/admin/server/overview`**
Returns: uptime, health status, current counters (from existing metricz data), total project/user/member counts (COUNT queries on server.db).

**`GET /v1/admin/server/config`**
Returns non-secret config: listen addr, rate limits (per endpoint class), signup toggle, log level, log format, CORS origins. Read from the in-memory config struct, not from env vars directly (avoids leaking secrets that happen to be in env).

**`GET /v1/admin/server/rate-limit-violations?from=&to=&key_id=&ip=`**
Queries `rate_limit_events` table with filters. Returns paginated results.

**`GET /v1/admin/server/errors?from=&to=&code=&route=`**
Requires persistent error logging (see section 11). If deferred, this endpoint can return data from the existing in-memory `client_errors`/`server_errors` counters with a note that historical data requires persistent metrics.

**`GET /v1/admin/server/metrics?from=&to=&step=`**
Requires persistent time-series metrics (see section 11). Can be deferred if persistent metrics are out of initial scope — td-watch can poll `/v1/admin/server/overview` and build client-side charts from snapshots.

#### 9.2 Users/Auth (scope: `admin:read:server`)

**`GET /v1/admin/users?cursor=&limit=&q=`**
Paginated user list from `users` table. `q` searches by email (LIKE). Include: id, email, is_admin, created_at, project_count (subquery on memberships), last_activity (MAX of api_keys.last_used_at).

**`GET /v1/admin/users/{id}`**
Single user detail. Same fields as list plus full project list with roles.

**`GET /v1/admin/users/{id}/keys`**
API keys for user. Return: key_prefix, name, scopes, created_at, last_used_at, expires_at. Never return the key hash.

**`GET /v1/admin/auth/events?status=&from=&to=&email=`**
Paginated query on `auth_events` table. `status` filters by event_type.

#### 9.3 Projects (scope: `admin:read:projects`)

**`GET /v1/admin/projects?cursor=&limit=&q=&include_deleted=`**
All projects (not scoped to a user like `GET /v1/projects`). `q` searches by name. `include_deleted=true` includes soft-deleted projects. Include: id, name, created_at, updated_at, deleted_at, member_count, event_count (from events.db or cached), last_activity.

Note: event_count requires opening each project's events.db. For the list endpoint, consider caching event counts in server.db or returning them lazily. A simple approach: add `event_count` and `last_event_at` columns to `projects` table, updated on each push.

**`GET /v1/admin/projects/{id}`**
Project detail including deleted_at. Same data as list view plus description.

**`GET /v1/admin/projects/{id}/members`**
All members of project with user email, role, invited_by, created_at.

**`GET /v1/admin/projects/{id}/sync/status`**
Event count, head server_seq, last event timestamp. Read from the project's events.db.

**`GET /v1/admin/projects/{id}/sync/cursors`**
All sync cursors for this project from `sync_cursors` table. Each entry: client_id, last_event_id, last_sync_at, distance_from_head (computed: head_seq - last_event_id).

#### 9.4 Event Introspection (scope: `admin:read:events`)

**`GET /v1/admin/projects/{id}/events?after_seq=&limit=&entity_type=&action_type=&from=&to=&device_id=&session_id=&entity_id=`**
Paginated, filterable query on project's events.db. This is the most complex endpoint — it needs to build a dynamic WHERE clause from the filters.

Filters:
- `after_seq` — cursor (events with server_seq > value)
- `entity_type` — exact match (validated against allowlist)
- `action_type` — exact match (create/update/delete/soft_delete)
- `from`, `to` — server_timestamp range
- `device_id`, `session_id` — exact match
- `entity_id` — exact match

**`GET /v1/admin/projects/{id}/events/{server_seq}`**
Single event with full payload.

**`GET /v1/admin/entity-types`**
Returns the `allowedEntityTypes` map as a JSON array. Currently hardcoded in `internal/api/sync.go`. Move to a shared constant or config. No auth required beyond admin.

#### 9.5 Derived State / Snapshots (scope: `admin:read:snapshots`)

These endpoints query the materialized snapshot database (built from full event replay). The snapshot mechanism already exists for `GET /v1/projects/{id}/sync/snapshot`.

**`GET /v1/admin/projects/{id}/snapshot/meta`**
Returns: snapshot server_seq, current head seq, staleness (head - snapshot seq), build timestamp, entity counts per type.

**`GET /v1/admin/projects/{id}/snapshot/query?q=&cursor=&limit=`**
Single query endpoint powered by TDQ (see `docs/guides/query-guide.md`). Instead of per-entity-type endpoints with bespoke filters, reuse the existing TDQ query engine server-side.

TDQ already supports:
- Filtering by any issue field (status, type, priority, dates, etc.)
- Boolean logic (AND, OR, NOT, grouping)
- Cross-entity search (log., comment., handoff., file. prefixes)
- Functions (has(), is(), any(), blocks(), descendant_of(), etc.)
- Relative dates (-7d, today, this_week)
- Inline sort (sort:priority, sort:-created)

The `q` parameter accepts a TDQ expression. Examples:
- `?q=status = open AND type = bug` — open bugs
- `?q=log.type = blocker` — issues with blocker logs
- `?q=priority <= P1 sort:-updated` — high priority, recently updated first
- `?q=created >= -7d` — created in last 7 days

**What this requires:** The TDQ engine (parser + evaluator in `internal/query/` or `pkg/query/`) currently assumes it's running against the local project database. It needs to be refactored so the evaluator can accept an arbitrary `*sql.DB` (the snapshot database) rather than assuming the local DB path. The parser and SQL generation are already DB-agnostic — the coupling is likely just in how the DB connection is obtained.

This replaces the per-entity-type endpoints from the original spec:
- ~~`/snapshot/issues?status=&q=`~~
- ~~`/snapshot/logs?issue_id=`~~
- ~~`/snapshot/comments?entity_id=`~~
- ~~`/snapshot/handoffs`~~
- ~~`/snapshot/boards`~~

One TDQ-powered endpoint is more powerful than five bespoke ones, and avoids building a second, weaker query surface for the same data. td-watch's UI can provide entity-type dropdown + filter chips that compile to TDQ expressions client-side.

Snapshot is rebuilt if stale (more than 1000 events behind head, configurable).

Implementation: build or reuse snapshot, open as read-only, pass to TDQ evaluator, paginate results.

#### 9.6 Export (scope: `admin:export`)

**`GET /v1/admin/projects/{id}/events/export?format=json|csv&entity_type=&action_type=&from=&to=&limit=`**

Synchronous streaming export. Hard cap at 100k events. Returns `Content-Disposition: attachment; filename="events-{project_id}-{timestamp}.{format}"`.

- JSON format: newline-delimited JSON (one event per line)
- CSV format: headers + rows

Returns error code `export_too_large` if the filtered result set exceeds the cap (check with COUNT first or use LIMIT+1 trick).

### 10. Error Response Contract Updates

Current error format is already correct:
```json
{"error": {"code": "string", "message": "string"}}
```

New error codes to add:
- `insufficient_admin_scope` — valid admin user but missing required scope (used by admin middleware)
- `project_deleted` — project exists but is soft-deleted
- `snapshot_unavailable` — snapshot cannot be built or is in progress
- `export_too_large` — export exceeds synchronous cap

Audit all existing endpoints for consistent use of the error format.

### 11. Persistent Metrics (Can Be Deferred)

The spec calls for persistent time-series metrics for historical charts. This is the heaviest infrastructure change and can be phased:

**Phase 0 (minimum for td-watch v1):**
- Extend the existing in-memory counters with more granularity: status codes by route class (2xx/4xx/5xx/429), auth funnel counters, per-project push/pull counts
- td-watch polls `/v1/admin/server/overview` on an interval and builds client-side charts from snapshots (no server-side time-series storage needed)

**Phase 1 (proper time-series):**
- Add Prometheus exposition format to `/metricz` alongside existing JSON
- Or: persist counter snapshots to a `metrics_snapshots` table in server.db at regular intervals (e.g., every 60s)
- `GET /v1/admin/server/metrics?from=&to=&step=` queries historical data

**Phase 2 (full observability):**
- OpenTelemetry integration
- External Prometheus/Grafana stack

Recommendation: start with Phase 0. td-watch can build useful dashboards from polling alone. Add persistent metrics when the polling approach proves insufficient.

### 12. Project Event Count Caching

The admin project list needs event counts per project. Querying every project's events.db on each list request is expensive.

Options:
1. **Cache in server.db:** Add `event_count` and `last_event_at` columns to `projects` table. Update atomically during push (already have a transaction context). Cheap, accurate.
2. **Lazy loading:** Return projects without event counts in the list, let the UI fetch per-project stats on demand. Simpler but worse UX.

Recommendation: option 1. It's a few lines in the push handler.

### 13. Retention Cleanup Jobs

Add background goroutines (alongside existing auth_requests cleanup):

- **auth_events cleanup:** Delete rows older than `SYNC_AUTH_EVENT_RETENTION` (default 90 days). Run every hour.
- **rate_limit_events cleanup:** Delete rows older than `SYNC_RATE_LIMIT_EVENT_RETENTION` (default 30 days). Run every hour.

## New Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SYNC_CORS_ALLOWED_ORIGINS` | (empty) | Comma-separated allowed origins for CORS |
| `SYNC_AUTH_EVENT_RETENTION` | `90d` | How long to keep auth_events rows |
| `SYNC_RATE_LIMIT_EVENT_RETENTION` | `30d` | How long to keep rate_limit_events rows |

## File-Level Change Map

| File/Package | Changes |
|-------------|---------|
| `internal/serverdb/schema.go` | v3 migration: is_admin, auth_events, rate_limit_events tables |
| `internal/serverdb/users.go` (new or extend) | IsAdmin check, grant/revoke, first-user-is-admin |
| `internal/serverdb/auth_events.go` (new) | InsertAuthEvent, QueryAuthEvents, CleanupAuthEvents |
| `internal/serverdb/rate_limit_events.go` (new) | InsertRateLimitEvent, QueryRateLimitEvents, CleanupRateLimitEvents |
| `internal/serverdb/pagination.go` (new) | Generic cursor-based pagination helper |
| `internal/api/admin_middleware.go` (new) | requireAdmin middleware |
| `internal/api/admin_server.go` (new) | Server overview, config, rate-limit-violations, errors endpoints |
| `internal/api/admin_users.go` (new) | Users list/detail/keys, auth events endpoints |
| `internal/api/admin_projects.go` (new) | Projects list/detail/members, sync status/cursors endpoints |
| `internal/api/admin_events.go` (new) | Event introspection, entity-types, export endpoints |
| `internal/api/admin_snapshots.go` (new) | Snapshot meta + TDQ-powered query endpoint |
| `internal/query/` or `pkg/query/` | Refactor TDQ evaluator to accept arbitrary *sql.DB (decouple from local project DB) |
| `internal/api/server.go` | Register admin routes, CORS middleware, add retention cleanup goroutines |
| `internal/api/ratelimit.go` | Add rate_limit_events INSERT on 429 |
| `internal/api/auth.go` | Add auth_events INSERT at each auth lifecycle step |
| `internal/api/projects.go` | Add cursor pagination to existing list endpoints |
| `internal/api/members.go` | Add cursor pagination to existing list endpoint |
| `internal/api/errors.go` | Add new error codes |
| `cmd/td-sync/admin.go` (new) | CLI: `td-sync admin grant`, `revoke`, `create-key` |

## Implementation Order

Suggested ordering to enable incremental progress:

1. **Schema migration + is_admin + CLI** — foundation, can be tested standalone
2. **Admin middleware** — needed by all admin endpoints
3. **Cursor-based pagination helper** — needed by all list endpoints
4. **Server overview + config endpoints** — simplest admin endpoints, good first integration test with td-watch
5. **User/auth endpoints** — straightforward queries on server.db
6. **Auth event logging + endpoint** — instrument auth flow, add query endpoint
7. **Rate limit event logging + endpoint** — instrument rate limiter, add query endpoint
8. **Project list + detail endpoints** — includes event count caching
9. **Sync status + cursors endpoints** — reads from events.db and sync_cursors
10. **Event introspection endpoints** — most complex (dynamic WHERE, cross-db)
11. **Entity-types endpoint** — simple, expose existing allowlist
12. **Export endpoint** — streaming response, builds on event introspection
13. **Snapshot query endpoint** — refactor TDQ to accept arbitrary DB, wire up to snapshot
14. **CORS** — low priority since td-watch proxies everything
15. **Pagination retrofit on existing endpoints** — backwards-compatible, can land anytime
16. **Retention cleanup jobs** — can run without the rest

## Testing Strategy

- Unit tests for each new serverdb function (auth_events, rate_limit_events, pagination, admin queries)
- Integration tests for admin endpoints (same pattern as existing API tests in `internal/api/`)
- Test admin middleware with valid admin key, non-admin key, missing scope, no key
- Test pagination with empty results, single page, multi-page, max limit
- Test event introspection filters individually and in combination
- Test TDQ evaluator against snapshot DB (reuse existing TDQ test cases with server-side DB)
- Test export with JSON and CSV formats, cap enforcement
- Test retention cleanup actually deletes old rows
- Test first-user-is-admin logic
- Test grant/revoke CLI commands

## Endpoint Summary

| Method | Path | Scope | Description |
|--------|------|-------|-------------|
| GET | `/v1/admin/server/overview` | admin:read:server | Uptime, health, counters, entity counts |
| GET | `/v1/admin/server/config` | admin:read:server | Non-secret config values |
| GET | `/v1/admin/server/metrics` | admin:read:server | Time-series metrics (deferred) |
| GET | `/v1/admin/server/rate-limit-violations` | admin:read:server | Rate limit violation log |
| GET | `/v1/admin/server/errors` | admin:read:server | Error log (deferred) |
| GET | `/v1/admin/users` | admin:read:server | Paginated user list |
| GET | `/v1/admin/users/{id}` | admin:read:server | User detail |
| GET | `/v1/admin/users/{id}/keys` | admin:read:server | User's API keys |
| GET | `/v1/admin/auth/events` | admin:read:server | Auth event log |
| GET | `/v1/admin/projects` | admin:read:projects | All projects (inc. deleted) |
| GET | `/v1/admin/projects/{id}` | admin:read:projects | Project detail |
| GET | `/v1/admin/projects/{id}/members` | admin:read:projects | Project members |
| GET | `/v1/admin/projects/{id}/sync/status` | admin:read:projects | Sync status |
| GET | `/v1/admin/projects/{id}/sync/cursors` | admin:read:projects | Client sync cursors |
| GET | `/v1/admin/projects/{id}/events` | admin:read:events | Filtered event stream |
| GET | `/v1/admin/projects/{id}/events/{seq}` | admin:read:events | Single event |
| GET | `/v1/admin/entity-types` | admin:read:events | Valid entity type list |
| GET | `/v1/admin/projects/{id}/snapshot/meta` | admin:read:snapshots | Snapshot metadata |
| GET | `/v1/admin/projects/{id}/snapshot/query` | admin:read:snapshots | TDQ-powered query over snapshot |
| GET | `/v1/admin/projects/{id}/events/export` | admin:export | Streaming event export |
