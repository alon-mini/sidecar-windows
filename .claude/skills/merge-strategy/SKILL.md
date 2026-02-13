---
name: merge-strategy
description: >
  Git merge strategies, conflict resolution approaches, merge vs rebase
  recommendations, and branch integration patterns in sidecar. Covers pull
  strategy menu, direct merge workflow, squash merge, commit message templates,
  configurable defaults, and protected branches. Use when working on git merge
  features or making decisions about merge strategies.
---

# Merge Strategy Configuration

This skill covers sidecar's git merge and pull operations, their current implementation, known gaps, and recommended configuration patterns.

## Current Implementation

### Git Status Plugin Pull Strategies

Location: `internal/plugins/gitstatus/pull_menu.go`, `internal/plugins/gitstatus/remote.go`

The git-status plugin offers four pull strategies via a modal menu:

| Strategy | Command | Use Case |
|----------|---------|----------|
| Pull (merge) | `git pull` | Creates merge commit, preserves branch topology |
| Pull (rebase) | `git pull --rebase` | Replays local commits on top of upstream |
| Pull (fast-forward only) | `git pull --ff-only` | Only pulls if fast-forward possible (safest) |
| Pull (rebase + autostash) | `git pull --rebase --autostash` | Rebase with automatic stash/unstash |

### Workspace Plugin Merge Workflows

Location: `internal/plugins/workspace/merge.go`

**PR Workflow:**
1. Push branch to remote
2. Create PR via `gh pr create`
3. Wait for PR merge (polling)
4. Cleanup: delete worktree, branches, pull base

**Direct Merge (No PR)** at `merge.go:430-498`:
```
1. git fetch origin <baseBranch>
2. git checkout <baseBranch>
3. git pull origin <baseBranch>
4. git merge <branch> --no-ff -m "Merge branch '<branch>'"  // Hardcoded --no-ff
5. git push origin <baseBranch>
```

**Post-Merge Pull** at `merge.go:500-551`:
- When on base branch: `git pull --ff-only origin <branch>` (hardcoded)
- Otherwise: `git fetch` + `git update-ref`

**Divergence Resolution** at `merge.go:644-705`:
- Rebase option: `git pull --rebase origin <branch>`
- Merge option: `git pull origin <branch>`

## Known Issues and Gaps

1. **No configurable defaults** -- users must select strategy every time
2. **Limited merge strategies** -- direct merge hardcoded to `--no-ff`, missing `--ff`, `--ff-only`, `--squash`, rebase-based integration
3. **No squash merge support** -- no clean-history option for main branch
4. **Hardcoded commit messages** -- `fmt.Sprintf("Merge branch '%s'", branch)`, no templates
5. **Missing safety options** -- no GPG signing, no `--verify-signatures`, no `--force-with-lease`
6. **Autostash inconsistency** -- available for pull-rebase but not post-merge pull or divergence resolution
7. **No upstream tracking config** -- remote always `origin`, no configurable tracking
8. **Limited conflict recovery** -- only abort or dismiss, no continue/skip rebase
9. **No interactive rebase** -- no squash/reorder/edit before merge

## Recommended Configuration Schema

Add to `internal/config/config.go`:

```go
type GitConfig struct {
    DefaultPullStrategy  string   `json:"defaultPullStrategy,omitempty"`  // "merge", "rebase", "ff-only", "autostash"
    DefaultMergeStrategy string   `json:"defaultMergeStrategy,omitempty"` // "no-ff", "ff", "ff-only", "squash", "rebase"
    MergeCommitTemplate  string   `json:"mergeCommitTemplate,omitempty"`  // Go template: .Branch, .BaseBranch, .PRTitle, .PRNumber
    SignCommits          bool     `json:"signCommits,omitempty"`
    SignMerges           bool     `json:"signMerges,omitempty"`
    AutostashOnPull      bool     `json:"autostashOnPull,omitempty"`
    DefaultRemote        string   `json:"defaultRemote,omitempty"`        // Default: "origin"
    SetUpstreamOnPush    bool     `json:"setUpstreamOnPush,omitempty"`
    ProtectedBranches    []string `json:"protectedBranches,omitempty"`    // e.g. ["main", "master", "release/*"]
    PreMergeChecks       []string `json:"preMergeChecks,omitempty"`       // e.g. ["go test ./..."]
}
```

Add to `PluginsConfig`:
```go
type PluginsConfig struct {
    Git       GitConfig             `json:"git"`
    GitStatus GitStatusPluginConfig `json:"git-status"`
    // ... existing fields
}
```

Example user config (`~/.config/sidecar/config.json`):
```json
{
  "plugins": {
    "git": {
      "defaultPullStrategy": "rebase",
      "defaultMergeStrategy": "squash",
      "mergeCommitTemplate": "Merge {{.Branch}} into {{.BaseBranch}}",
      "signCommits": true,
      "autostashOnPull": true,
      "protectedBranches": ["main", "production"]
    }
  }
}
```

## Implementation Priorities

### Priority 1: Default Pull Strategy

**Files:** `internal/config/config.go`, `internal/plugins/gitstatus/pull_menu.go`, `internal/plugins/gitstatus/update_handlers.go`

- If `defaultPullStrategy` is set, execute directly (skip menu)
- If unset, show menu as today
- Consider adding "Always use this" option to menu

### Priority 2: Default Merge Strategy

**File:** `internal/plugins/workspace/merge.go` -- `performDirectMerge()`

```go
func (p *Plugin) performDirectMerge(wt *Worktree) tea.Cmd {
    strategy := p.ctx.Config.Plugins.Git.DefaultMergeStrategy
    if strategy == "" {
        strategy = "no-ff" // backward compatible default
    }
    var mergeArgs []string
    switch strategy {
    case "ff":
        mergeArgs = []string{"merge", branch, "-m", mergeMsg}
    case "ff-only":
        mergeArgs = []string{"merge", branch, "--ff-only"}
    case "squash":
        mergeArgs = []string{"merge", branch, "--squash"}
    case "rebase":
        // 1. git rebase baseBranch branch
        // 2. git checkout baseBranch
        // 3. git merge branch --ff-only
    default: // "no-ff"
        mergeArgs = []string{"merge", branch, "--no-ff", "-m", mergeMsg}
    }
}
```

### Priority 3: Squash Merge Support

After `git merge --squash`, a separate commit is needed:
```go
case "squash":
    mergeCmd := exec.Command("git", "merge", branch, "--squash")
    // execute...
    commitMsg := fmt.Sprintf("Squash merge branch '%s'", branch)
    commitCmd := exec.Command("git", "commit", "-m", commitMsg)
```

### Priority 4: Merge Commit Templates

```go
type MergeTemplateData struct {
    Branch, BaseBranch, PRTitle, PRNumber string
}

func renderMergeMessage(tmpl string, data MergeTemplateData) (string, error) {
    if tmpl == "" {
        tmpl = "Merge branch '{{.Branch}}'"
    }
    t, err := template.New("merge").Parse(tmpl)
    // ... execute template
}
```

### Priority 5: Protected Branch Warnings

```go
func isProtectedBranch(branch string, patterns []string) bool {
    for _, pattern := range patterns {
        if matched, _ := filepath.Match(pattern, branch); matched {
            return true
        }
    }
    return false
}
```

Show confirmation modal before direct merge to protected branches.

## Strategy Comparison

| Strategy | History | Commit Count | Best For |
|----------|---------|--------------|----------|
| No Fast-Forward (`--no-ff`) | Preserves branch topology | +1 merge commit | Feature branches, audit trails |
| Fast-Forward (`--ff`) | Linear when possible | No extra commits | Small changes, single commits |
| Fast-Forward Only (`--ff-only`) | Always linear | Fails if not possible | Strict linear history |
| Squash | Linear | Single commit | Clean history, large features |
| Rebase | Linear, preserves commits | No merge commit | Clean history with detail |

### Team Recommendations

- **Open source:** default `squash`, PR workflow recommended
- **Internal teams:** default `no-ff`, direct merge acceptable
- **Solo developers:** default `rebase` or `ff`, direct merge for speed

## Related Files

| File | Purpose |
|------|---------|
| `internal/config/config.go` | Config struct definitions |
| `internal/config/loader.go` | Config loading and defaults |
| `internal/plugins/gitstatus/remote.go` | Pull/fetch operations |
| `internal/plugins/gitstatus/pull_menu.go` | Pull strategy UI |
| `internal/plugins/workspace/merge.go` | Merge workflow |
| `internal/plugins/workspace/diff.go` | Base branch detection |
| `internal/plugins/workspace/worktree.go` | Branch operations |

## Migration Notes

- **Empty config = current behavior** -- all new fields default to empty/false
- **Config validation** -- warn on invalid strategy names, do not fail
- **Backward compatibility** -- if changing defaults, provide migration warnings
