# TestHarness API Reference

## Table of Contents

- [Constructor](#constructor)
- [HTTP Methods](#http-methods)
- [Seed Data Methods](#seed-data-methods)
- [State Builder](#state-builder)
- [TestState Accessors](#teststate-accessors)
- [Assertion Helpers](#assertion-helpers)
- [Common Test Patterns](#common-test-patterns)
- [Admin Endpoints](#admin-endpoints)

## Constructor

```go
func newTestHarness(t *testing.T, opts ...func(*Config)) *TestHarness
```

Creates an isolated test server with:
- Fresh SQLite database in `t.TempDir()`
- Project data directory
- `httptest.NewServer` wrapping `srv.routes()` (full middleware chain)
- Rate limits set to 100,000 (effectively disabled)
- Auto-cleanup via `t.Cleanup`

Config overrides:
```go
h := newTestHarness(t, func(cfg *Config) {
    cfg.CORSAllowedOrigins = []string{"https://admin.example.com"}
    cfg.RateLimitPush = 5  // test rate limiting
})
```

Exposed fields: `h.Server`, `h.Store`, `h.BaseURL`

## HTTP Methods

### Do

```go
func (h *TestHarness) Do(method, path, token string, body any) *http.Response
```

Sends a real HTTP request. Caller must close `resp.Body` unless using assertion helpers.

- `path`: relative, e.g. `/v1/admin/users` (prepended with `h.BaseURL`)
- `token`: bearer token string, or `""` for unauthenticated
- `body`: any JSON-serializable value, or `nil`
- Fatals on network error

### DoJSON

```go
func (h *TestHarness) DoJSON(method, path, token string, body any, out any) *http.Response
```

Calls `Do`, then decodes JSON response into `out`. Fatals if status >= 400 (prints response body) or if JSON decode fails.

```go
var overview serverOverviewResponse
h.DoJSON("GET", "/v1/admin/server/overview", adminToken, nil, &overview)
```

## Seed Data Methods

### CreateUser

```go
func (h *TestHarness) CreateUser(email string) (userID, token string)
```

Creates user + sync-scoped API key via the store (not HTTP). Note: first user created is automatically admin.

### CreateAdminUser

```go
func (h *TestHarness) CreateAdminUser(email, scopes string) (userID, token string)
```

Creates user, sets `is_admin=1`, generates API key with given scopes.

Scopes are comma-separated: `"admin:read:server,admin:read:projects,sync"`

### CreateProject

```go
func (h *TestHarness) CreateProject(ownerToken, name string) string
```

Creates project via `POST /v1/projects` (real HTTP). Returns project ID. Owner is auto-added as member.

### PushEvents

```go
func (h *TestHarness) PushEvents(token, projectID string, events []EventInput)
```

Pushes events via `POST /v1/projects/{id}/sync/push`. Uses fixed device/session IDs.

### BuildSnapshot

```go
func (h *TestHarness) BuildSnapshot(token, projectID string)
```

Triggers snapshot build via `GET /v1/projects/{id}/sync/snapshot`.

## State Builder

Fluent API for complex test setup. Steps are deferred and executed in order by `Done()`.

```go
state := h.Build().
    WithUser("alice@test.com").
    WithUser("bob@test.com").
    WithAdmin("admin@test.com", "admin:read:server,admin:read:projects,admin:read:events,admin:read:snapshots,sync").
    WithProject("proj1", "alice@test.com").
    WithMember("proj1", "bob@test.com", "writer").
    WithEvents("proj1", "alice@test.com", 10).
    WithSnapshot("proj1").
    WithAuthEvents(5).
    WithRateLimitEvents(3).
    Done()
```

### Builder Methods

| Method | Description |
|--------|-------------|
| `WithUser(email)` | Creates non-admin user with sync key |
| `WithAdmin(email, scopes)` | Creates admin user with scoped key |
| `WithProject(name, ownerEmail)` | Creates project (owner must exist) |
| `WithMember(projectName, email, role)` | Adds member using owner's token |
| `WithEvents(projectName, userEmail, count)` | Pushes events cycling issues/logs/comments |
| `WithSnapshot(projectName)` | Builds snapshot using owner's token |
| `WithAuthEvents(count)` | Inserts auth events directly to DB |
| `WithRateLimitEvents(count)` | Inserts rate-limit events directly to DB |
| `Done()` | Executes all steps, returns `*TestState` |

### Ordering Constraints

- `WithUser`/`WithAdmin` before `WithProject` (owner must exist)
- `WithProject` before `WithMember`, `WithEvents`, `WithSnapshot`
- `WithUser` before `WithMember` (member must exist)
- `WithEvents` before `WithSnapshot` (need events to snapshot)

## TestState Accessors

```go
state.UserToken(email string) string    // fatals if not found
state.UserID(email string) string       // fatals if not found
state.AdminToken(email string) string   // fatals if not found or not admin
state.ProjectID(name string) string     // fatals if not found
state.Harness() *TestHarness            // underlying harness
```

## Assertion Helpers

### AssertStatus

```go
func AssertStatus(t *testing.T, resp *http.Response, expected int)
```

Checks status code. Reads and prints body on failure. Does NOT close body on success.

### AssertErrorResponse

```go
func AssertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedCode string)
```

Reads body, checks status and error code. Closes body. Error format: `{"error":{"code":"...","message":"..."}}`.

### ReadJSON

```go
func ReadJSON[T any](t *testing.T, resp *http.Response) T
```

Generic JSON decoder. Closes body via defer.

### AssertPaginated

```go
func AssertPaginated[T any](t *testing.T, resp *http.Response, expectedCount int, expectHasMore bool) PaginatedResponse[T]
```

Checks status 200, decodes paginated response, asserts item count and `has_more`. Closes body.

Returns `PaginatedResponse[T]` with `Data`, `HasMore`, `NextCursor` for cursor follow-up.

### CORS Assertions

```go
func AssertCORSHeaders(t *testing.T, resp *http.Response, expectedOrigin string)
func AssertNoCORSHeaders(t *testing.T, resp *http.Response)
```

### AssertRequiresAdminScope

```go
func (h *TestHarness) AssertRequiresAdminScope(t *testing.T, method, path, wrongScopeToken string)
```

Sends request with wrong-scope token, asserts 403 with `"insufficient_admin_scope"`.

## Common Test Patterns

### Pagination Follow

```go
resp := h.Do("GET", "/v1/admin/projects?limit=2", token, nil)
page1 := AssertPaginated[serverdb.AdminProject](t, resp, 2, true)

resp = h.Do("GET", fmt.Sprintf("/v1/admin/projects?limit=2&cursor=%s", page1.NextCursor), token, nil)
page2 := AssertPaginated[serverdb.AdminProject](t, resp, 1, false)
```

### Scope Enforcement

```go
// Admin with wrong scope
h.AssertRequiresAdminScope(t, "GET", "/v1/admin/projects", serverScopeToken)

// Non-admin user
resp := h.Do("GET", "/v1/admin/server/overview", regularToken, nil)
AssertErrorResponse(t, resp, http.StatusForbidden, "insufficient_admin_scope")
```

### Non-Admin Denial

```go
h.CreateUser("first@test.com")  // consume auto-admin slot
_, regularToken := h.CreateUser("regular@test.com")
resp := h.Do("GET", "/v1/admin/server/overview", regularToken, nil)
AssertErrorResponse(t, resp, http.StatusForbidden, "insufficient_admin_scope")
```

### CORS

```go
h := newTestHarness(t, func(cfg *Config) {
    cfg.CORSAllowedOrigins = []string{"https://admin.example.com"}
})
// Must use http.NewRequest for Origin header (h.Do doesn't support custom headers)
req, _ := http.NewRequest("GET", h.BaseURL+"/v1/admin/server/overview", nil)
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Origin", "https://admin.example.com")
resp, _ := (&http.Client{}).Do(req)
AssertCORSHeaders(t, resp, "https://admin.example.com")
```

### Event Filtering

```go
state := h.Build().
    WithUser("u@test.com").
    WithAdmin("a@test.com", "admin:read:events,sync").
    WithProject("p1", "u@test.com").
    WithEvents("p1", "u@test.com", 9). // 3 each of issues/logs/comments
    Done()

var events adminEventsResponse
h.DoJSON("GET", fmt.Sprintf("/v1/admin/projects/%s/events?entity_type=issues", pid), token, nil, &events)
// events.Data has 3 items, all EntityType == "issues"
```

## Admin Endpoints

All prefixed with `/v1/admin/`.

### Server (scope: `admin:read:server`)

| Method | Path | Response Type |
|--------|------|---------------|
| GET | `/server/overview` | `serverOverviewResponse` |
| GET | `/server/config` | `serverConfigResponse` |
| GET | `/server/rate-limit-violations?key_id=&ip=&from=&to=&cursor=&limit=` | paginated |
| GET | `/users?cursor=&limit=&q=` | paginated `serverdb.AdminUser` |
| GET | `/users/{id}` | `serverdb.AdminUser` |
| GET | `/users/{id}/keys` | paginated key list |
| GET | `/auth/events?status=&from=&to=&email=&cursor=&limit=` | paginated |

### Projects (scope: `admin:read:projects`)

| Method | Path | Response Type |
|--------|------|---------------|
| GET | `/projects?cursor=&limit=&q=&include_deleted=` | paginated `serverdb.AdminProject` |
| GET | `/projects/{id}` | `serverdb.AdminProject` |
| GET | `/projects/{id}/members` | paginated `serverdb.AdminProjectMember` |
| GET | `/projects/{id}/sync/status` | `adminSyncStatusResponse` |
| GET | `/projects/{id}/sync/cursors` | `{Data []adminCursorEntry}` |

### Events (scope: `admin:read:events`)

| Method | Path | Response Type |
|--------|------|---------------|
| GET | `/projects/{id}/events?after_seq=&limit=&entity_type=&action_type=&device_id=&session_id=&entity_id=` | `adminEventsResponse` |
| GET | `/projects/{id}/events/{server_seq}` | `adminEvent` |
| GET | `/entity-types` | `{EntityTypes []string}` |

### Snapshots (scope: `admin:read:snapshots`)

| Method | Path | Response Type |
|--------|------|---------------|
| GET | `/projects/{id}/snapshot/meta` | snapshot metadata |
| GET | `/projects/{id}/snapshot/query?q=&cursor=&limit=` | TDQ-powered query |
