# Proposal: Consistent Agent Session Identity

## Problem Statement

Sessions must be:
1. **Stable within a context window** - same agent session = same td session
2. **Unique across concurrent agents** - two agents on same branch = different sessions
3. **Cross-tool compatible** - works with Claude Code, Cursor, Gemini, Codex, etc.
4. **Seamless** - no manual intervention required

Current branch-scoped sessions solve (1) but fail (2) - two Claude Code sessions on `main` share one session.

## Key Discovery: Parent Agent PID is Stable

Within any AI agent session, there's a **parent orchestrator process** that spawns all Bash commands. This PID is:
- Stable across ALL commands in that session
- Unique per agent instance
- Different even for two agents on the same branch

**Example from live Claude Code session:**
```
Terminal (Warp)
  └─ zsh (4200)
       └─ claude (15261)  ← STABLE IDENTIFIER
            └─ zsh (each command)
                 └─ actual command
```

Every `td` command in this session has PID 15261 as an ancestor. Another Claude Code session has a different PID.

---

## Creative Proposals

### Proposal 1: Agent Ancestry Detection (Recommended)

Walk up the process tree looking for known agent parent processes.

```go
func getAgentFingerprint() (agentType string, pid int, found bool) {
    agentPatterns := map[string]string{
        "claude":   "claude-code",
        "cursor":   "cursor",
        "codex":    "codex",
        "windsurf": "windsurf",
        "zed":      "zed",
        "aider":    "aider",
        "copilot":  "copilot",
    }

    pid := os.Getppid()
    for depth := 0; depth < 15; depth++ {
        name, ppid := getProcessInfo(pid)

        for pattern, agentType := range agentPatterns {
            if strings.Contains(strings.ToLower(name), pattern) {
                return agentType, pid, true
            }
        }

        if ppid <= 1 { break }
        pid = ppid
    }
    return "", 0, false
}
```

**Session path structure:**
```
.todos/sessions/
  main/
    _terminal.json             # human terminal (no agent detected)
    claude-code_15261.json     # Claude Code session A
    claude-code_83769.json     # Claude Code session B (parallel)
    cursor_45678.json          # Cursor session
```

**Pros:**
- Zero configuration required
- Works across all tools automatically
- Stable within session, unique across sessions
- Clear session attribution in filenames

**Cons:**
- Requires maintaining list of agent process names
- Process names could change (mitigated by extensible config)

---

### Proposal 2: SSE Port Detection (Claude Code specific)

Claude Code runs a persistent SSE server. Each session has a unique port.

```go
func getClaudeCodePort() (int, bool) {
    // Walk up process tree to find node process
    // Check its listening ports via lsof
    // Return the SSE port (5175, 5177, etc.)
}
```

**Observed:**
- Session A: node (PID 27150) → port 5175
- Session B: node (PID 82600) → port 5177

**Pros:**
- Very reliable for Claude Code
- Port is visible and debuggable

**Cons:**
- Claude Code specific
- Requires lsof/netstat access

---

### Proposal 3: Environment Variable Cascade

Prioritized detection with fallbacks:

```go
func getSessionIdentifier() string {
    // 1. Explicit override (most reliable)
    if id := os.Getenv("TD_SESSION_ID"); id != "" {
        return "explicit:" + id
    }

    // 2. Agent-provided session IDs (when available)
    for _, env := range []string{
        "CLAUDE_SESSION_ID",     // Feature request pending
        "CURSOR_AGENT",          // Cursor provides this
        "COPILOT_SESSION_ID",    // GitHub Copilot
    } {
        if val := os.Getenv(env); val != "" {
            return "env:" + env + "=" + val
        }
    }

    // 3. Agent ancestry detection (our own)
    if agent, pid, found := getAgentFingerprint(); found {
        return fmt.Sprintf("agent:%s:%d", agent, pid)
    }

    // 4. Terminal session (for humans)
    for _, env := range []string{
        "TERM_SESSION_ID", "TMUX_PANE", "STY",
    } {
        if val := os.Getenv(env); val != "" {
            return "term:" + val
        }
    }

    // 5. Fallback: branch only
    return "branch:" + getCurrentBranch()
}
```

**Pros:**
- Comprehensive fallback chain
- Adapts as tools add session IDs
- Explicit override for edge cases

**Cons:**
- Complexity
- Many tools don't expose session IDs yet

---

### Proposal 4: Session Claim Protocol

Instead of trying to detect identity, use a claim-based system:

```go
type Session struct {
    ID            string
    Branch        string
    ClaimedBy     string    // fingerprint of claiming process
    ClaimedAt     time.Time
    LastActivity  time.Time
}

func GetOrCreate(baseDir string) (*Session, error) {
    sess := loadSession()
    fingerprint := getFingerprint()

    // If unclaimed or stale (>5min since activity), claim it
    if sess.ClaimedBy == "" || time.Since(sess.LastActivity) > 5*time.Minute {
        sess.ClaimedBy = fingerprint
        sess.ClaimedAt = time.Now()
    }

    // If claimed by someone else AND recent → collision!
    if sess.ClaimedBy != fingerprint && time.Since(sess.LastActivity) < 5*time.Minute {
        // Fork: create new session for this agent
        return forkSession(sess)
    }

    // Same claimer → continue
    sess.LastActivity = time.Now()
    return sess, nil
}
```

**Pros:**
- Self-healing: automatically forks on collision
- Works even with imperfect fingerprinting
- Graceful degradation

**Cons:**
- 5-minute window could cause issues
- Slightly more complex logic

---

### Proposal 5: File Lock Session Affinity

Use OS-level file locking for definitive session ownership:

```go
func claimSession(path string) (*os.File, error) {
    f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0644)
    if err != nil { return nil, err }

    // Try non-blocking exclusive lock
    err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
    if err == syscall.EWOULDBLOCK {
        // Someone else has the lock → need different session
        f.Close()
        return nil, ErrSessionLocked
    }

    return f, nil // We own this session
}
```

**Pros:**
- Definitive ownership
- Works across all tools
- OS-level guarantees

**Cons:**
- Lock files can be orphaned (process crash)
- Requires cleanup logic
- Platform differences (Windows vs Unix)

---

### Proposal 6: Hybrid Approach (Recommended Implementation)

Combine the best elements:

```go
// Session identity: agent ancestry + branch
func getSessionPath(baseDir string) string {
    branch := getCurrentBranch()
    branchDir := filepath.Join(baseDir, ".todos/sessions", sanitize(branch))

    // Try agent detection first
    if agent, pid, found := getAgentFingerprint(); found {
        return filepath.Join(branchDir, fmt.Sprintf("%s_%d.json", agent, pid))
    }

    // Fallback to terminal session
    if termID := getTerminalSessionID(); termID != "" {
        return filepath.Join(branchDir, fmt.Sprintf("term_%s.json", hash(termID)[:8]))
    }

    // Ultimate fallback
    return filepath.Join(branchDir, "_default.json")
}

// With claim protocol for safety
func GetOrCreate(baseDir string) (*Session, error) {
    path := getSessionPath(baseDir)

    // Load or create
    sess, err := loadSession(path)
    if err != nil {
        return createSession(path)
    }

    // Verify claim (safety net)
    fingerprint := getFingerprint()
    if sess.ClaimedBy != "" && sess.ClaimedBy != fingerprint {
        if time.Since(sess.LastActivity) < 5*time.Minute {
            // Collision! Log warning but continue (path should have prevented this)
            log.Printf("WARN: session collision detected")
        }
    }

    sess.ClaimedBy = fingerprint
    sess.LastActivity = time.Now()
    return sess, nil
}
```

---

## Comparison Matrix

| Approach | Multi-Agent Same Branch | Cross-Tool | Zero Config | Reliability |
|----------|------------------------|------------|-------------|-------------|
| Current (branch only) | ❌ | ✅ | ✅ | Medium |
| Agent Ancestry | ✅ | ✅ | ✅ | High |
| SSE Port | ✅ | ❌ (Claude only) | ✅ | High |
| Env Cascade | ✅ | ✅ | ✅ | Medium |
| Claim Protocol | ✅ | ✅ | ✅ | Medium |
| File Lock | ✅ | ✅ | ✅ | High |
| Hybrid | ✅ | ✅ | ✅ | Very High |

---

## Recommendation

Implement **Proposal 1 (Agent Ancestry Detection)** with **Proposal 4 (Claim Protocol)** as a safety net.

### Why:

1. **Agent ancestry** provides stable, unique identifiers across all tools
2. **Claim protocol** handles edge cases and collision detection
3. **Zero configuration** - just works
4. **Clear attribution** - session files show which agent owns them
5. **Extensible** - easy to add new agent patterns

### Implementation Steps:

1. Add `getAgentFingerprint()` function to walk process tree
2. Modify `sessionPathForBranch()` to include agent fingerprint
3. Add `ClaimedBy` field to Session struct
4. Add collision detection in `GetOrCreate()`
5. Add `td session list` to show all sessions by branch+agent

### Configuration Extension:

Allow users to add custom agent patterns:

```json
// .todos/config.json
{
  "agent_patterns": {
    "my-custom-agent": "custom-agent"
  }
}
```

---

## Edge Cases Addressed

| Scenario | How Handled |
|----------|-------------|
| Two Claude Code sessions on main | Different PIDs → different sessions |
| Agent switches branches | Same PID → same session follows |
| Agent crashes, new one starts | Different PID → new session |
| Human in terminal, agent in IDE | Different ancestry → different sessions |
| Same branch, Cursor + Claude Code | Different agent types → different sessions |
| Unknown agent tool | Falls back to terminal or default session |
| Parallel agents in CI/CD | Different PIDs → isolated sessions |

---

## Related Issues

- [GitHub #13733](https://github.com/anthropics/claude-code/issues/13733): Request for CLAUDE_SESSION_ID env var
- [GitHub #13735](https://github.com/anthropics/claude-code/issues/13735): Persistent env vars across Bash calls

Once Claude Code exposes `CLAUDE_SESSION_ID`, we can use it directly as the fingerprint, making detection even more reliable.
