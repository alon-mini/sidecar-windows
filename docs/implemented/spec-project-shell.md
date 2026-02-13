# Plan: Project Shell in Worktrees Plugin

## Summary

Add a "Project Shell" entry as the first item in the worktrees list. This special entry provides a tmux session in the project root directory, reusing all existing tmux infrastructure (capture, polling, output rendering). No new plugin needed.

## Design

```
┌─ Worktrees ─────────────────┬─ Output ────────────────────────┐
│                             │                                  │
│ → ● Project Shell           │ $ npm run dev                    │
│     (sidecar)               │ > vite                           │
│                             │   VITE v5.2.0 ready in 500ms     │
│   ⧗ feature/auth            │   → Local: http://localhost:5173 │
│     claude · 3m · td-123    │                                  │
│                             │                                  │
│   ✓ fix/login-bug           │                                  │
│     claude · 1h             │                                  │
│                             │                                  │
└─────────────────────────────┴──────────────────────────────────┘
```

**Key behaviors:**

- Shell entry always appears first, visually distinct (different icon, no branch)
- Session name: `sidecar-sh-{project-name}` (distinct from `sidecar-wt-*`)
- Same Output/Diff/Task tabs work (Output shows shell, Diff/Task show empty state)
- Shell-specific commands: Create session, Attach, Kill session
- Worktree commands (Delete, Push, Merge, Task link) disabled for shell

## Implementation

### 1. Add shell session tracking to Plugin struct

**File: `internal/plugins/worktree/plugin.go`**

```go
type Plugin struct {
    // ... existing fields ...

    // Project shell session (not a git worktree)
    shellSession     *Agent      // Reuse Agent struct for tmux session
    shellSessionName string      // "sidecar-sh-{project}"
    shellSelected    bool        // True when shell entry is selected (vs a worktree)
}
```

### 2. Add shell session management

**File: `internal/plugins/worktree/shell.go`** (new file)

```go
// Shell session lifecycle - reuses agent.go patterns

func (p *Plugin) initShellSession() {
    projectName := filepath.Base(p.ctx.WorkDir)
    p.shellSessionName = "sidecar-sh-" + sanitizeName(projectName)

    // Check if session exists from previous run
    if sessionExists(p.shellSessionName) {
        p.shellSession = &Agent{
            Type:        AgentTypeShell, // New agent type constant
            TmuxSession: p.shellSessionName,
            OutputBuf:   NewOutputBuffer(outputBufferCap),
        }
    }
}

func (p *Plugin) createShellSession() tea.Cmd {
    // tmux new-session -d -s {name} -c {workdir}
    // Same pattern as StartAgent() but without agent command
}

func (p *Plugin) attachToShell() tea.Cmd {
    return tea.ExecProcess(
        exec.Command("tmux", "attach", "-t", p.shellSessionName),
        func(err error) tea.Msg { return ShellDetachedMsg{} },
    )
}

func (p *Plugin) killShellSession() tea.Cmd {
    // tmux kill-session -t {name}
}
```

### 3. Modify sidebar rendering

**File: `internal/plugins/worktree/view_list.go`**

```go
func (p *Plugin) renderSidebarContent(width, height int) string {
    var lines []string

    // === Shell entry (always first) ===
    shellLine := p.renderShellEntry(width, p.shellSelected)
    lines = append(lines, shellLine)
    lines = append(lines, "") // Separator

    // === Worktrees ===
    for i, wt := range p.worktrees {
        selected := !p.shellSelected && i == p.selectedIdx
        lines = append(lines, p.renderWorktreeItem(wt, width, selected))
    }

    return strings.Join(lines, "\n")
}

func (p *Plugin) renderShellEntry(width int, selected bool) string {
    // Icon: "●" if session exists, "○" if no session
    // Name: "Project Shell" or project name from config
    // Status: shows session state (running/no session)
    icon := "○"
    status := "no session"
    if p.shellSession != nil {
        icon = p.shellSession.Status.Icon()
        status = "running"
    }

    // Different visual style from worktrees (no branch, no agent type)
    // Example: "● Project Shell (sidecar)"
}
```

### 4. Modify selection logic

**File: `internal/plugins/worktree/update.go`**

```go
func (p *Plugin) handleKeyDown() {
    if p.shellSelected {
        // Moving down from shell -> first worktree
        if len(p.worktrees) > 0 {
            p.shellSelected = false
            p.selectedIdx = 0
        }
    } else if p.selectedIdx < len(p.worktrees)-1 {
        p.selectedIdx++
    }
}

func (p *Plugin) handleKeyUp() {
    if !p.shellSelected && p.selectedIdx == 0 {
        // Moving up from first worktree -> shell
        p.shellSelected = true
    } else if !p.shellSelected {
        p.selectedIdx--
    }
}

func (p *Plugin) selectedWorktree() *Worktree {
    if p.shellSelected {
        return nil // Shell is not a worktree
    }
    // ... existing logic ...
}
```

### 5. Modify commands based on selection

**File: `internal/plugins/worktree/commands.go`**

```go
func (p *Plugin) Commands() []plugin.Command {
    // ... existing modal handling ...

    // Base commands (always available)
    cmds := []plugin.Command{
        {ID: "new-worktree", Name: "New", ...},
        {ID: "toggle-view", Name: viewToggleName, ...},
        // ...
    }

    if p.shellSelected {
        // Shell-specific commands
        if p.shellSession == nil {
            cmds = append(cmds,
                plugin.Command{ID: "create-shell", Name: "Create", Description: "Create shell session", Priority: 10},
            )
        } else {
            cmds = append(cmds,
                plugin.Command{ID: "attach", Name: "Attach", Description: "Attach to shell", Priority: 10},
                plugin.Command{ID: "kill-shell", Name: "Kill", Description: "Kill shell session", Priority: 11},
            )
        }
        return cmds
    }

    // Worktree-specific commands (existing logic)
    wt := p.selectedWorktree()
    if wt != nil {
        // ... existing worktree commands ...
    }
    return cmds
}
```

### 6. Modify output rendering

**File: `internal/plugins/worktree/view_preview.go`**

```go
func (p *Plugin) renderOutputContent(width, height int) string {
    if p.shellSelected {
        // Render shell output
        if p.shellSession == nil {
            return p.renderNoShellMessage(width, height)
        }
        return p.renderAgentOutput(p.shellSession, width, height)
    }

    // Existing worktree output rendering
    wt := p.selectedWorktree()
    // ...
}

func (p *Plugin) renderNoShellMessage(width, height int) string {
    return `
  No shell session

  Press 'n' to create a tmux session
  in the project directory.

  For running builds, dev servers, or
  quick fixes without worktrees.
`
}
```

### 7. Handle project switching

**File: `internal/plugins/worktree/plugin.go`**

```go
func (p *Plugin) Init(ctx *plugin.Context) error {
    p.ctx = ctx
    // ... existing init ...

    // Initialize shell session tracking for new project
    p.initShellSession()

    return nil
}

func (p *Plugin) Stop() {
    // Stop shell polling but keep session alive (survives project switch)
    // ... existing cleanup ...
}
```

### 8. Add keybindings

**File: `internal/keymap/bindings.go`**

```go
// Shell-specific bindings (context: worktree-list when shell selected)
{Key: "n", Command: "create-shell", Context: "worktree-shell"},
{Key: "enter", Command: "attach", Context: "worktree-shell"},
{Key: "K", Command: "kill-shell", Context: "worktree-shell"},
```

## Files to modify

| File                                        | Change                                |
| ------------------------------------------- | ------------------------------------- |
| `internal/plugins/worktree/plugin.go`       | Add shell fields, init logic          |
| `internal/plugins/worktree/types.go`        | Add `AgentTypeShell` constant         |
| `internal/plugins/worktree/commands.go`     | Shell-specific commands               |
| `internal/plugins/worktree/update.go`       | Selection navigation, key handling    |
| `internal/plugins/worktree/view_list.go`    | Render shell entry in sidebar         |
| `internal/plugins/worktree/view_preview.go` | Render shell output, no-session state |
| `internal/keymap/bindings.go`               | Shell context bindings                |

## New files

| File                                 | Purpose                                      |
| ------------------------------------ | -------------------------------------------- |
| `internal/plugins/worktree/shell.go` | Shell session lifecycle (create/attach/kill) |

## Verification

1. **Navigation**: Press `k` on first worktree to select shell, `j` to go back
2. **Create session**: Select shell, press `n` (or key for create-shell), session starts
3. **Attach**: Press `Enter` on shell, attached to tmux, `Ctrl-b d` detaches
4. **Output streaming**: Run a command in shell, output appears in Output tab
5. **Project switch**: `@` to switch projects, shell session persists, reconnects when returning
6. **Worktree commands**: Select a worktree, Delete/Push/Merge commands available; select shell, those commands hidden
7. **tmux list-sessions**: Shows both `sidecar-sh-*` and `sidecar-wt-*` sessions

## Benefits of this approach

- **No code duplication**: Reuses all tmux capture, polling, rendering code
- **Consistent UX**: Same keybindings, same output preview, same attach/detach flow
- **Clear separation**: Shell is visually distinct, commands differ based on selection
- **Project-aware**: Shell session tied to current project, survives switches
