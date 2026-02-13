# Worktree Plugin Spec Review

A detailed review of `docs/spec-worktree-plugin.md` against actual sidecar and td codebase patterns.

---

## Executive Summary

The spec is well-structured and demonstrates deep understanding of the domain. However, several areas need revision to align with sidecar's established patterns. The main gaps are:

1. **Plugin interface alignment** - spec doesn't match sidecar's actual Plugin interface
2. **Missing keymap integration** - spec doesn't use sidecar's Registry-based keymap system
3. **View rendering issues** - footer rendering violates sidecar patterns (app handles footer)
4. **TD integration oversimplified** - td's database path is actually in `db.go`, not `root.go`
5. **Missing message types** - needs tea.Msg types for async operations
6. **Context system gaps** - FocusContext() not fully specified

---

## Critical Changes Required

### 1. Plugin Interface Mismatch (Section 3.3)

**Spec has:**
```go
type WorktreeManager struct {
    worktrees []*Worktree
    agents    map[string]*Agent
    tdClient  *TDClient
}
```

**Sidecar requires:**
```go
type Plugin struct {
    // Required by plugin.Plugin interface
    ctx       *plugin.Context
    focused   bool
    width     int
    height    int

    // Worktree-specific
    worktrees []*Worktree
    agents    map[string]*Agent

    // View state
    viewMode       ViewMode
    activePane     FocusPane
    selectedIdx    int
    previewOffset  int

    // Async state
    refreshing     bool
    lastRefresh    time.Time
}

// Must implement full interface
func (p *Plugin) ID() string           { return "worktree-manager" }
func (p *Plugin) Name() string         { return "Worktrees" }
func (p *Plugin) Icon() string         { return "W" }
func (p *Plugin) IsFocused() bool      { return p.focused }
func (p *Plugin) SetFocused(f bool)    { p.focused = f }
func (p *Plugin) Commands() []plugin.Command { ... }
func (p *Plugin) FocusContext() string { ... }
```

### 2. Missing Message Types (Add new section)

Sidecar plugins use typed messages for all async operations:

```go
// Add to types.go
type RefreshMsg struct{}
type RefreshDoneMsg struct {
    Worktrees []*Worktree
    Err       error
}
type WatchEventMsg struct {
    Path string
}
type AgentOutputMsg struct {
    WorktreeName string
    Output       string
    Status       WorktreeStatus
}
type AgentStoppedMsg struct {
    WorktreeName string
    Err          error
}
type TmuxAttachFinishedMsg struct {
    Err error
}
type DiffLoadedMsg struct {
    WorktreeName string
    Content      string
    Raw          string
}
type TaskSearchResultsMsg struct {
    Tasks []*Task
    Err   error
}
```

### 3. View Rendering Violates Patterns (Section 4)

**Problem:** Spec shows plugin rendering its own footer:
```
│ n:new  y:approve  ↵:attach  d:diff  p:push  m:merge  D:delete  ?:help       │
```

**Fix:** Remove footer from View(). Define commands instead:

```go
func (p *Plugin) Commands() []plugin.Command {
    return []plugin.Command{
        // List view commands
        {ID: "worktree-new", Name: "New", Context: "worktree-list", Priority: 1},
        {ID: "worktree-approve", Name: "Approve", Context: "worktree-list", Priority: 1},
        {ID: "worktree-attach", Name: "Attach", Context: "worktree-list", Priority: 2},
        {ID: "worktree-diff", Name: "Diff", Context: "worktree-list", Priority: 2},
        {ID: "worktree-push", Name: "Push", Context: "worktree-list", Priority: 3},
        {ID: "worktree-merge", Name: "Merge", Context: "worktree-list", Priority: 3},
        {ID: "worktree-delete", Name: "Delete", Context: "worktree-list", Priority: 4},

        // Output pane commands
        {ID: "worktree-scroll-up", Name: "Up", Context: "worktree-output", Priority: 1},
        {ID: "worktree-scroll-down", Name: "Down", Context: "worktree-output", Priority: 1},

        // Diff view commands
        {ID: "worktree-close-diff", Name: "Close", Context: "worktree-diff", Priority: 1},

        // Kanban view commands
        {ID: "worktree-list-view", Name: "List", Context: "worktree-kanban", Priority: 1},
    }
}
```

### 4. Missing FocusContext Implementation (Add to Section 4)

```go
func (p *Plugin) FocusContext() string {
    switch p.viewMode {
    case ViewModeList:
        if p.activePane == PaneOutput {
            return "worktree-output"
        }
        if p.activePane == PaneDiff {
            return "worktree-diff"
        }
        if p.activePane == PaneTask {
            return "worktree-task"
        }
        return "worktree-list"
    case ViewModeKanban:
        return "worktree-kanban"
    case ViewModeNewModal:
        return "worktree-new-modal"
    case ViewModeConfirmDelete:
        return "worktree-confirm-delete"
    default:
        return "worktree-list"
    }
}
```

### 5. Keymap Bindings Missing (Add new section)

Add to `internal/keymap/bindings.go`:

```go
// Worktree plugin bindings
{Key: "n", Command: "worktree-new", Context: "worktree-list"},
{Key: "N", Command: "worktree-new-with-prompt", Context: "worktree-list"},
{Key: "y", Command: "worktree-approve", Context: "worktree-list"},
{Key: "Y", Command: "worktree-approve-all", Context: "worktree-list"},
{Key: "enter", Command: "worktree-attach", Context: "worktree-list"},
{Key: "d", Command: "worktree-diff", Context: "worktree-list"},
{Key: "D", Command: "worktree-delete", Context: "worktree-list"},
{Key: "p", Command: "worktree-push", Context: "worktree-list"},
{Key: "P", Command: "worktree-create-pr", Context: "worktree-list"},
{Key: "m", Command: "worktree-merge", Context: "worktree-list"},
{Key: "t", Command: "worktree-link-task", Context: "worktree-list"},
{Key: "r", Command: "worktree-resume", Context: "worktree-list"},
{Key: "R", Command: "worktree-refresh", Context: "worktree-list"},
{Key: "/", Command: "worktree-filter", Context: "worktree-list"},
{Key: "v", Command: "worktree-toggle-view", Context: "worktree-list"},
{Key: "h", Command: "worktree-pane-left", Context: "worktree-list"},
{Key: "l", Command: "worktree-pane-right", Context: "worktree-list"},
{Key: "tab", Command: "worktree-cycle-preview", Context: "worktree-list"},
{Key: "escape", Command: "worktree-close-modal", Context: "worktree-new-modal"},
{Key: "enter", Command: "worktree-confirm-create", Context: "worktree-new-modal"},
```

### 6. Height Constraint Pattern Missing (Section 4)

**Critical:** Add to View() implementation:

```go
func (p *Plugin) View(width, height int) string {
    p.width, p.height = width, height

    var content string
    switch p.viewMode {
    case ViewModeList:
        content = p.renderListView()
    case ViewModeKanban:
        content = p.renderKanbanView()
    }

    // Handle modals as overlays
    if p.viewMode == ViewModeNewModal {
        modal := p.renderNewWorktreeModal()
        content = ui.OverlayModal(content, modal, width, height)
    }

    // CRITICAL: Constrain to allocated height (prevents header scroll-off)
    return lipgloss.NewStyle().
        Width(width).
        Height(height).
        MaxHeight(height).
        Render(content)
}
```

---

## TD Integration Corrections

### 1. Database Path Location Wrong (Section 8.1)

**Spec says (line 1169-1179):**
```go
// In cmd/root.go or db/db.go - add to baseDir resolution
```

**Actual location:** The database path is defined in `internal/db/db.go`:
```go
const dbFile = ".todos/issues.db"

func Open(baseDir string) (*DB, error) {
    dbPath := filepath.Join(baseDir, dbFile)
    // ...
}
```

**BaseDir resolution** is in `cmd/root.go` via `initBaseDir()` which uses `os.Getwd()`.

**Correction:** The `.td-root` check should be added to `internal/db/db.go`'s `Open()` function:

```go
func Open(baseDir string) (*DB, error) {
    // Check for worktree redirection
    tdRootPath := filepath.Join(baseDir, ".td-root")
    if content, err := os.ReadFile(tdRootPath); err == nil {
        baseDir = strings.TrimSpace(string(content))
    }

    dbPath := filepath.Join(baseDir, dbFile)
    // ...
}
```

### 2. `.td-root` File Format Unspecified

**Add:** Document the exact format:
```
# .td-root file format
# Single line containing absolute path to main worktree
# Example content:
/Users/dev/code/sidecar
```

### 3. Session Scoping Already Exists

**Good news:** TD already has branch-scoped sessions at `.todos/sessions/<branch>/<agent-pid>.json`. The spec should note this provides automatic isolation per worktree.

### 4. Task Search Commands Need Correction

**Spec shows (line 1286):**
```go
cmd := exec.Command("td", "query", fmt.Sprintf("title ~ '%s'", query))
```

**Better approach using JSON output:**
```go
cmd := exec.Command("td", "list", "--json", "--status", "open,in_progress")
```

Then filter in Go for faster fuzzy matching with the UI.

---

## Missing Sections to Add

### 1. DiagnosticProvider Implementation

Sidecar plugins can implement optional diagnostics:

```go
func (p *Plugin) Diagnostics() []plugin.Diagnostic {
    var diags []plugin.Diagnostic

    // Check tmux availability
    if _, err := exec.LookPath("tmux"); err != nil {
        diags = append(diags, plugin.Diagnostic{
            Level:   plugin.DiagnosticError,
            Message: "tmux not found in PATH",
            Hint:    "Install tmux: brew install tmux",
        })
    }

    // Check for orphaned sessions
    orphaned := p.findOrphanedSessions()
    if len(orphaned) > 0 {
        diags = append(diags, plugin.Diagnostic{
            Level:   plugin.DiagnosticWarning,
            Message: fmt.Sprintf("%d orphaned tmux sessions", len(orphaned)),
            Hint:    "Press 'c' to cleanup",
        })
    }

    return diags
}
```

### 2. Event Bus Integration

The spec mentions events but doesn't show implementation:

```go
func (p *Plugin) Init(ctx *plugin.Context) error {
    p.ctx = ctx

    // Subscribe to git events for auto-refresh
    gitEvents := ctx.EventBus.Subscribe("git:status-changed")
    go func() {
        for range gitEvents {
            // Trigger refresh when git status changes
            // This catches external commits, pushes, etc.
        }
    }()

    return nil
}

// Publish worktree events for other plugins
func (p *Plugin) publishAgentStatus(wt *Worktree) {
    p.ctx.EventBus.Publish("worktree:agent-status", event.NewEvent(
        event.TypeRefreshNeeded,
        "worktree:agent-status",
        map[string]interface{}{
            "worktree": wt.Name,
            "status":   wt.Status,
        },
    ))
}
```

### 3. Inter-Plugin Communication

Add section showing how worktree plugin communicates with other plugins:

```go
// Navigate to file in file browser
func (p *Plugin) openFileInBrowser(worktreePath, filePath string) tea.Cmd {
    return tea.Batch(
        app.FocusPlugin("file-browser"),
        func() tea.Msg {
            // Use relative path from worktree
            return filebrowser.NavigateToFileMsg{Path: filePath}
        },
    )
}

// Navigate to diff in git plugin
func (p *Plugin) showInGitPlugin(path string) tea.Cmd {
    return tea.Batch(
        app.FocusPlugin("git-status"),
        func() tea.Msg {
            return gitstatus.ShowDiffMsg{Path: path}
        },
    )
}
```

### 4. Start() Method Pattern

The spec shows business logic but not the proper Start() pattern:

```go
func (p *Plugin) Start() tea.Cmd {
    return tea.Batch(
        p.refresh(),           // Load worktrees
        p.startWatcher(),      // Watch for git changes
        p.reconnectAgents(),   // Find existing tmux sessions
    )
}

func (p *Plugin) refresh() tea.Cmd {
    return func() tea.Msg {
        worktrees, err := p.listWorktrees()
        return RefreshDoneMsg{Worktrees: worktrees, Err: err}
    }
}

func (p *Plugin) startWatcher() tea.Cmd {
    return func() tea.Msg {
        // Watch .git/worktrees for changes
        watcher, err := fsnotify.NewWatcher()
        if err != nil {
            return WatcherErrorMsg{Err: err}
        }
        p.watcher = watcher

        go func() {
            for event := range watcher.Events {
                // Debounce and send refresh message
            }
        }()

        watcher.Add(filepath.Join(p.ctx.WorkDir, ".git", "worktrees"))
        return WatcherStartedMsg{}
    }
}
```

### 5. Update() Handler Structure

```go
func (p *Plugin) Update(msg tea.Msg) (plugin.Plugin, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return p.handleKey(msg)
    case tea.MouseMsg:
        return p.handleMouse(msg)

    // Refresh cycle
    case RefreshMsg:
        p.refreshing = true
        return p, p.refresh()
    case RefreshDoneMsg:
        p.refreshing = false
        p.lastRefresh = time.Now()
        if msg.Err != nil {
            return p, app.Toast(msg.Err.Error(), 3*time.Second)
        }
        p.worktrees = msg.Worktrees
        return p, nil

    // Agent output polling
    case AgentOutputMsg:
        if wt := p.findWorktree(msg.WorktreeName); wt != nil {
            wt.Agent.OutputBuf.Update(msg.Output)
            wt.Status = msg.Status
        }
        return p, nil

    // Tmux attach
    case TmuxAttachFinishedMsg:
        // Returned from ExecProcess
        return p, p.refresh()

    // App-level messages
    case app.PluginFocusedMsg:
        return p, p.refresh()
    case app.RefreshMsg:
        return p, p.refresh()
    }

    return p, nil
}
```

### 6. Stop() Cleanup

```go
func (p *Plugin) Stop() {
    // Stop file watcher
    if p.watcher != nil {
        p.watcher.Close()
    }

    // Stop agent polling goroutines
    close(p.stopChan)

    // Note: Don't kill tmux sessions - they should persist
    // for agent resume functionality
}
```

---

## Architecture Recommendations

### 1. Split View Implementation

The spec shows a split-pane layout but doesn't explain implementation. Use sidecar's existing pattern from git plugin:

```go
func (p *Plugin) renderListView() string {
    // Calculate pane widths
    totalWidth := p.width
    listWidth := min(60, totalWidth/3)
    previewWidth := totalWidth - listWidth - 1 // -1 for border

    // Render each pane
    listPane := p.renderWorktreeList(listWidth, p.height)
    previewPane := p.renderPreviewPane(previewWidth, p.height)

    // Join with border
    border := strings.Repeat("│\n", p.height)
    return lipgloss.JoinHorizontal(
        lipgloss.Top,
        listPane,
        border,
        previewPane,
    )
}
```

### 2. Modal Pattern

Use `ui.OverlayModal` instead of custom modal rendering:

```go
func (p *Plugin) renderNewWorktreeModal() string {
    width := ui.ModalWidthMedium

    var b strings.Builder
    b.WriteString(ui.ModalTitle("New Worktree", width))
    b.WriteString(p.renderBranchInput())
    b.WriteString(p.renderTaskSearch())
    b.WriteString(p.renderAgentSelector())
    b.WriteString(p.renderOptions())
    b.WriteString(ui.ModalFooter("Esc:Cancel  Enter:Create", width))

    return b.String()
}
```

### 3. Agent Type Configuration

Move agent configuration to use sidecar's config system:

```go
// In config/config.go
type WorktreeConfig struct {
    Enabled         bool              `json:"enabled"`
    WorktreeDir     string            `json:"worktreeDir"`
    DefaultAgent    string            `json:"defaultAgent"`
    Agents          map[string]Agent  `json:"agents"`
    RefreshInterval string            `json:"refreshInterval"`
}

type Agent struct {
    Command    string `json:"command"`
    PromptFlag string `json:"promptFlag,omitempty"`
}
```

---

## Minor Issues & Typos

1. **Line 53-58:** Table formatting is good but could link to actual repos
2. **Line 1095-1136:** Claude Code hooks section references `code.claude.com` which may not exist
3. **Line 1654-1720:** Phase timeline estimates should be removed per CLAUDE.md guidance
4. **Line 1258:** `--prompt` flag for claude may not exist - verify
5. **Section 4.2 Kanban:** Good concept but implementation details sparse

---

## Recommended Spec Updates

1. Add "Plugin Interface Implementation" section showing full interface compliance
2. Add "Message Types" section with all tea.Msg types
3. Add "Keymap Bindings" section with binding definitions
4. Remove footer from View mockups (app handles this)
5. Add "FocusContext Implementation" showing context switching logic
6. Correct TD database path location
7. Add "Inter-Plugin Communication" section
8. Add "DiagnosticProvider" implementation
9. Add "Event Bus Usage" section
10. Remove time estimates from implementation phases

---

## Files to Create/Modify

```
internal/plugins/worktree/
├── plugin.go       # Plugin interface + lifecycle
├── model.go        # State types + helpers
├── view.go         # List view rendering
├── view_kanban.go  # Kanban view
├── messages.go     # tea.Msg type definitions
├── keymap.go       # NOT needed - use registry
├── worktree.go     # Git worktree operations
├── agent.go        # Agent management
├── tmux.go         # tmux integration
├── td.go           # TD integration
├── config.go       # Config struct
└── types.go        # Shared types

internal/keymap/bindings.go  # Add worktree bindings

~/code/td/internal/db/db.go  # Add .td-root support
```

---

## Summary

The spec is 80% complete and well-thought-out. Main work needed:

1. **High priority:** Align with Plugin interface, add message types, fix view rendering
2. **Medium priority:** Add keymap bindings, FocusContext, diagnostics
3. **Low priority:** Correct TD paths, add inter-plugin communication docs

The worktree concept and tmux integration design are solid. The main gaps are around sidecar-specific patterns that aren't documented elsewhere.
