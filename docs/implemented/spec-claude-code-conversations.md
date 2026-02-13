# Conversations Plugin Enhancement Specification

## Executive Summary
Fix bug preventing sessions from loading, enhance with token stats & search, maintain adapter abstraction for reusability with other AI coding tools (opencode, codex, gemini-cli).

**Bug**: Sessions never load - watcher event loop incomplete, silent failures in Init()
**Priority Features**: Session slug display, token usage stats, search/filter

## Critical Files
- `internal/plugins/conversations/plugin.go` - Core bug fix + search state
- `internal/plugins/conversations/view.go` - Enhanced rendering + token display
- `internal/adapter/claudecode/types.go` - Add Slug field
- `internal/adapter/claudecode/adapter.go` - Extract slug from JSONL
- `internal/keymap/bindings.go` - Add keyboard shortcuts

## Phase 1: Fix Core Bugs (Critical)

### 1.1 Fix Watcher Event Loop
**Problem**: `startWatcher()` sets `p.watchChan` but never listens to it

**Location**: `internal/plugins/conversations/plugin.go`

**Changes**:
```go
// Add new method after startWatcher():
func (p *Plugin) listenForWatchEvents() tea.Cmd {
  if p.watchChan == nil {
    return nil
  }
  return func() tea.Msg {
    <-p.watchChan  // Block until event
    return WatchEventMsg{}
  }
}

// In Update() handler, add after line 123:
case WatchStartedMsg:
  p.watchChan = msg.Chan  // Store channel
  return p, p.listenForWatchEvents()  // Start listening

// Modify WatchEventMsg handler (line 132):
case WatchEventMsg:
  return p, tea.Batch(
    p.loadSessions(),
    p.listenForWatchEvents(),  // Continue listening
  )
```

**Add to message types**:
```go
type WatchStartedMsg struct {
  Chan <-chan adapter.Event
}
```

**Update startWatcher()** (line 313):
```go
func (p *Plugin) startWatcher() tea.Cmd {
  return func() tea.Msg {
    if p.adapter == nil {
      return nil
    }
    ch, err := p.adapter.Watch(p.ctx.WorkDir)
    if err != nil {
      return ErrorMsg{Err: err}
    }
    return WatchStartedMsg{Chan: ch}  // Return channel in msg
  }
}
```

### 1.2 Add Debug Logging
**Problem**: Silent failures in Init() - can't diagnose why sessions don't load

**Location**: `internal/plugins/conversations/plugin.go` lines 77-94

**Changes**:
```go
func (p *Plugin) Init(ctx *plugin.Context) error {
  p.ctx = ctx

  // Get Claude Code adapter
  if a, ok := ctx.Adapters["claude-code"]; ok {
    p.adapter = a
    ctx.Logger.Debug("conversations: adapter found")
  } else {
    ctx.Logger.Warn("conversations: no claude-code adapter")
    return nil
  }

  // Check if adapter can detect this project
  found, err := p.adapter.Detect(ctx.WorkDir)
  ctx.Logger.Debug("conversations: detect result", "found", found, "err", err, "workdir", ctx.WorkDir)
  if err != nil {
    ctx.Logger.Warn("conversations: detect error", "err", err)
    return nil
  }
  if !found {
    ctx.Logger.Info("conversations: no sessions for project", "workdir", ctx.WorkDir)
    return nil
  }

  return nil
}
```

### 1.3 Enhanced Diagnostics
**Location**: `internal/plugins/conversations/plugin.go` lines 261-276

**Changes**:
```go
func (p *Plugin) Diagnostics() []plugin.Diagnostic {
  if p.adapter == nil {
    return []plugin.Diagnostic{
      {ID: "conversations", Status: "disabled", Detail: "no adapter"},
    }
  }

  found, err := p.adapter.Detect(p.ctx.WorkDir)
  if err != nil {
    return []plugin.Diagnostic{
      {ID: "conversations", Status: "error", Detail: fmt.Sprintf("detect err: %v", err)},
    }
  }
  if !found {
    return []plugin.Diagnostic{
      {ID: "conversations", Status: "empty", Detail: "no sessions found"},
    }
  }

  detail := formatSessionCount(len(p.sessions))
  if len(p.sessions) == 0 {
    detail = "loaded but empty"
  }

  return []plugin.Diagnostic{
    {ID: "conversations", Status: "ok", Detail: detail},
  }
}
```

## Phase 2: Session Slug Display

### 2.1 Add Slug to Data Types
**Location**: `internal/adapter/claudecode/types.go`

**Changes**:
```go
// Line 18, add after GitBranch:
type RawMessage struct {
  // ... existing fields ...
  GitBranch string `json:"gitBranch,omitempty"`
  Slug      string `json:"slug,omitempty"`  // ADD THIS
}

// Line 72, add to SessionMetadata:
type SessionMetadata struct {
  // ... existing fields ...
  MsgCount  int
  Slug      string  // ADD THIS
}
```

### 2.2 Extract Slug in Parser
**Location**: `internal/adapter/claudecode/adapter.go`

**In parseSessionMetadata()** around line 249:
```go
if meta.FirstMsg.IsZero() {
  meta.FirstMsg = raw.Timestamp
  meta.CWD = raw.CWD
  meta.Version = raw.Version
  meta.GitBranch = raw.GitBranch
  meta.Slug = raw.Slug  // ADD THIS
}
```

**In Sessions() method** around line 93:
```go
sessions = append(sessions, adapter.Session{
  ID:        meta.SessionID,
  Name:      meta.Slug,  // CHANGE from meta.SessionID[:8]
  CreatedAt: meta.FirstMsg,
  UpdatedAt: meta.LastMsg,
  IsActive:  time.Since(meta.LastMsg) < 5*time.Minute,
})
```

**Add fallback** after the append:
```go
// If no slug, use short ID
if sessions[len(sessions)-1].Name == "" {
  sessions[len(sessions)-1].Name = meta.SessionID[:8]
}
```

## Phase 3: Token Usage Display

### 3.1 Enhanced Per-Message Token Display
**Location**: `internal/plugins/conversations/view.go` lines 143-190

**In renderMessage()** around line 156:
```go
tokens := ""
if msg.OutputTokens > 0 || msg.InputTokens > 0 {
  in := formatTokens(msg.InputTokens)
  out := formatTokens(msg.OutputTokens)
  cache := ""
  if msg.CacheRead > 0 {
    cache = fmt.Sprintf(" %s cache", formatTokens(msg.CacheRead))
  }
  tokens = fmt.Sprintf(" %s→%s%s", in, out, cache)
}
```

**Add helper function** at end of view.go:
```go
// formatTokens formats token counts with k/M suffix
func formatTokens(n int) string {
  if n == 0 {
    return "0"
  }
  if n < 1000 {
    return fmt.Sprintf("%d", n)
  }
  if n < 1000000 {
    return fmt.Sprintf("%dk", n/1000)
  }
  return fmt.Sprintf("%.1fM", float64(n)/1000000)
}
```

### 3.2 Session-Level Usage Stats
**Location**: `internal/plugins/conversations/plugin.go`

**Add to Plugin struct** around line 50:
```go
// Message view state
selectedSession string
messages        []adapter.Message
sessionUsage    *adapter.UsageStats  // ADD THIS
```

**Add loadUsage command**:
```go
func (p *Plugin) loadUsage(sessionID string) tea.Cmd {
  return func() tea.Msg {
    if p.adapter == nil {
      return UsageLoadedMsg{}
    }
    usage, err := p.adapter.Usage(sessionID)
    if err != nil {
      return ErrorMsg{Err: err}
    }
    return UsageLoadedMsg{Usage: usage}
  }
}
```

**Add message type**:
```go
type UsageLoadedMsg struct {
  Usage *adapter.UsageStats
}
```

**Update handler** in Update():
```go
case UsageLoadedMsg:
  p.sessionUsage = msg.Usage
  return p, nil
```

**Call in updateSessions()** around line 173:
```go
case "enter":
  if len(p.sessions) > 0 && p.cursor < len(p.sessions) {
    p.selectedSession = p.sessions[p.cursor].ID
    p.view = ViewMessages
    p.msgCursor = 0
    p.msgScrollOff = 0
    return p, tea.Batch(
      p.loadMessages(p.selectedSession),
      p.loadUsage(p.selectedSession),  // ADD THIS
    )
  }
```

**Display in header** - `internal/plugins/conversations/view.go` line 108:
```go
// Header
usageStr := ""
if p.sessionUsage != nil {
  usageStr = fmt.Sprintf("  %s→%s",
    formatTokens(p.sessionUsage.TotalInputTokens),
    formatTokens(p.sessionUsage.TotalOutputTokens))
}
header := fmt.Sprintf(" Session: %s%s                    %d msgs",
  sessionName, usageStr, len(p.messages))
```

## Phase 4: Search & Filter

### 4.1 Add Search State
**Location**: `internal/plugins/conversations/plugin.go`

**Add to Plugin struct** around line 37:
```go
// Current view
view View

// Search state
searchMode   bool    // ADD THIS
searchQuery  string  // ADD THIS
filteredSessions []int  // ADD THIS - indices of matching sessions
```

### 4.2 Search Mode Toggle
**In updateSessions()** around line 178:
```go
case "r":
  return p, p.loadSessions()

case "/":  // ADD THIS
  p.searchMode = true
  p.searchQuery = ""
  return p, nil

case "esc":  // ADD THIS
  if p.searchMode {
    p.searchMode = false
    p.searchQuery = ""
    p.filteredSessions = nil
    return p, nil
  }
```

**Handle search input** in Update():
```go
case tea.KeyMsg:
  if p.searchMode {
    return p.updateSearch(msg)
  }
  if p.view == ViewMessages {
    return p.updateMessages(msg)
  }
  return p.updateSessions(msg)
```

**Add updateSearch method**:
```go
func (p *Plugin) updateSearch(msg tea.KeyMsg) (plugin.Plugin, tea.Cmd) {
  switch msg.String() {
  case "esc":
    p.searchMode = false
    p.searchQuery = ""
    p.filteredSessions = nil
    return p, nil

  case "enter":
    p.searchMode = false
    p.filterSessions()
    return p, nil

  case "backspace":
    if len(p.searchQuery) > 0 {
      p.searchQuery = p.searchQuery[:len(p.searchQuery)-1]
      p.filterSessions()
    }
    return p, nil

  default:
    if len(msg.String()) == 1 {
      p.searchQuery += msg.String()
      p.filterSessions()
    }
  }
  return p, nil
}

func (p *Plugin) filterSessions() {
  if p.searchQuery == "" {
    p.filteredSessions = nil
    return
  }

  p.filteredSessions = nil
  query := strings.ToLower(p.searchQuery)
  for i, s := range p.sessions {
    if strings.Contains(strings.ToLower(s.Name), query) ||
       strings.Contains(strings.ToLower(s.ID), query) {
      p.filteredSessions = append(p.filteredSessions, i)
    }
  }
}
```

### 4.3 Render Filtered Sessions
**Location**: `internal/plugins/conversations/view.go`

**Update renderSessions()** around line 29:
```go
// Content
if len(p.sessions) == 0 {
  sb.WriteString(styles.Muted.Render(" No sessions found for this project"))
} else {
  displaySessions := p.sessions
  indices := make([]int, len(p.sessions))
  for i := range indices {
    indices[i] = i
  }

  if p.filteredSessions != nil {
    indices = p.filteredSessions
  }

  contentHeight := p.height - 2
  if contentHeight < 1 {
    contentHeight = 1
  }

  end := p.scrollOff + contentHeight
  if end > len(indices) {
    end = len(indices)
  }

  for i := p.scrollOff; i < end; i++ {
    idx := indices[i]
    session := displaySessions[idx]
    selected := i == p.cursor
    sb.WriteString(p.renderSessionRow(session, selected))
    sb.WriteString("\n")
  }
}

// Show search bar if in search mode
if p.searchMode {
  sb.WriteString("\n")
  sb.WriteString(styles.PanelHeader.Render(fmt.Sprintf(" Search: %s_", p.searchQuery)))
}
```

## Phase 5: Keyboard Shortcuts & Commands

### 5.1 Update Commands
**Location**: `internal/plugins/conversations/plugin.go` lines 244-250

```go
func (p *Plugin) Commands() []plugin.Command {
  if p.searchMode {
    return []plugin.Command{
      {ID: "cancel-search", Name: "Cancel", Context: "conversations-search"},
    }
  }

  if p.view == ViewMessages {
    return []plugin.Command{
      {ID: "back", Name: "Back", Context: "conversation-detail"},
      {ID: "scroll", Name: "Scroll", Context: "conversation-detail"},
    }
  }

  return []plugin.Command{
    {ID: "view", Name: "View", Context: "conversations"},
    {ID: "search", Name: "Search", Context: "conversations"},
    {ID: "refresh", Name: "Refresh", Context: "conversations"},
  }
}
```

### 5.2 Update FocusContext
```go
func (p *Plugin) FocusContext() string {
  if p.searchMode {
    return "conversations-search"
  }
  if p.view == ViewMessages {
    return "conversation-detail"
  }
  return "conversations"
}
```

### 5.3 Add Bindings
**Location**: `internal/keymap/bindings.go`

```go
// In defaultKeymap() function, add:
"conversations": {
  "j":     "move-down",
  "k":     "move-up",
  "down":  "move-down",
  "up":    "move-up",
  "g":     "go-top",
  "G":     "go-bottom",
  "enter": "view",
  "/":     "search",
  "r":     "refresh",
},
"conversation-detail": {
  "j":      "move-down",
  "k":      "move-up",
  "down":   "move-down",
  "up":     "move-up",
  "g":      "go-top",
  "G":      "go-bottom",
  "esc":    "back",
  "q":      "back",
  "ctrl+d": "page-down",
  "ctrl+u": "page-up",
},
"conversations-search": {
  "esc":   "cancel-search",
  "enter": "confirm-search",
},
```

## Testing Checklist

- [ ] Sessions load correctly from `~/.claude/projects/`
- [ ] Session slugs display instead of UUIDs
- [ ] Watcher updates on new sessions/messages
- [ ] Token counts show per message
- [ ] Session-level usage stats display in header
- [ ] Search filters sessions by name/slug
- [ ] ESC exits search mode
- [ ] All keyboard shortcuts work
- [ ] Diagnostics show useful error messages
- [ ] Plugin degrades gracefully without adapter

## Implementation Order

1. **Phase 1** (Critical): Bug fixes - watcher loop + logging
2. **Phase 2**: Slug extraction and display
3. **Phase 3.1**: Per-message token display
4. **Phase 3.2**: Session-level usage stats
5. **Phase 4**: Search implementation
6. **Phase 5**: Keyboard shortcuts

## Adapter Interface Reusability

**No changes to adapter interface needed** - maintains compatibility for other tools:
- `adapter.Adapter` interface unchanged
- `adapter.Session`, `adapter.Message`, `adapter.TokenUsage` unchanged
- Other tools (opencode, codex, gemini-cli) implement same interface
- Tool-specific fields (like `slug`) handled in adapter implementation, not interface

**For other tools to implement**:
1. Create adapter package (e.g., `internal/adapter/opencode/`)
2. Implement `adapter.Adapter` interface methods
3. Register in `cmd/sidecar/main.go`: `pluginCtx.Adapters["tool-name"] = adapter.New()`
4. Plugin will automatically work with new adapter
