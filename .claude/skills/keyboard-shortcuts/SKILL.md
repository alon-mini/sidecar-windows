---
name: keyboard-shortcuts
description: >
  Reference for keyboard shortcut implementation, keybinding registration,
  shortcut parity with vim and other TUI tools, and the complete shortcut
  assignment table across all sidecar plugins. Use when adding or modifying
  keyboard shortcuts, checking shortcut assignments, resolving key conflicts,
  or assessing alignment with vim conventions.
---

# Keyboard Shortcuts

Complete shortcut listings and context reference for all sidecar plugins. For implementation patterns, see `docs/guides/ui-feature-guide.md`. For a detailed assessment of inconsistencies, vim alignment, mnemonic quality, and improvement proposals, see `references/assessment.md` in this skill directory.

## Architecture

- **Centralized binding registry**: `internal/keymap/bindings.go` is the single source of truth for key bindings.
- **Context-based dispatch**: Each plugin defines contexts; bindings are scoped to contexts.
- **Command palette** (`?`): Auto-discovers bindings for discoverability.
- **User overrides**: Supported via `~/.config/sidecar/config.json`.
- **Key sequences**: Compound commands like `g g` are supported with 500ms timeout.

### Adding a New Shortcut

1. Add the binding in `internal/keymap/bindings.go` under the appropriate context.
2. Add command handling in the plugin's `Update()` method (usually in a `handlers.go` file).
3. Add the command to the plugin's `Commands()` method for footer hint and command palette.
4. Keep command names short (1 word preferred) to prevent footer wrapping.

### TD Monitor Shortcuts

TD shortcuts are dynamically exported from TD itself via `ExportBindings()` and `ExportCommands()` in `pkg/monitor/keymap/`. TD is the single source of truth. To add TD shortcuts:

1. Add binding to TD's `pkg/monitor/keymap/bindings.go`
2. Add command constant to TD's `pkg/monitor/keymap/registry.go`
3. Add metadata to TD's `pkg/monitor/keymap/export.go`
4. Handle in TD's `pkg/monitor/model.go`

## Global Shortcuts

| Key | Command | Description |
|-----|---------|-------------|
| `j` / `down` | cursor-down | Move cursor down |
| `k` / `up` | cursor-up | Move cursor up |
| `G` | cursor-bottom | Jump to bottom |
| `g g` | cursor-top | Jump to top |
| `ctrl+d` | page-down | Page down |
| `ctrl+u` | page-up | Page up |
| `enter` | select | Select item |
| `esc` | back | Go back / close |
| `` ` `` | next-plugin | Next plugin |
| `~` | prev-plugin | Previous plugin |
| `1-5` | focus-plugin-N | Focus plugin by number |
| `?` | toggle-palette | Command palette |
| `!` | toggle-diagnostics | Diagnostics overlay |
| `@` | switch-project | Project switcher |
| `r` | refresh | Refresh |
| `q` | quit | Quit (root contexts only) |
| `ctrl+c` | quit | Force quit |

## Sidebar Controls (All Two-Pane Plugins)

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch focus between panes |
| `\` | Toggle sidebar visibility |
| `h` / `left` | Focus left pane |
| `l` / `right` | Focus right pane |

## Git Status Plugin

### Contexts

| Context | View |
|---------|------|
| `git-status` | File list (root) |
| `git-status-commits` | Recent commits sidebar (root) |
| `git-status-diff` | Inline diff pane (root) |
| `git-commit-preview` | Commit detail in right pane |
| `git-diff` | Full-screen diff |
| `git-commit` | Commit editor |
| `git-push-menu` | Push strategy selection |
| `git-pull-menu` | Pull strategy selection |
| `git-pull-conflict` | Conflict resolution |
| `git-history` | Commit history |
| `git-commit-detail` | Single commit view |

### File List Shortcuts

| Key | Command | Description |
|-----|---------|-------------|
| `s` | stage-file | Stage selected file |
| `u` | unstage-file | Unstage selected file |
| `S` | stage-all | Stage all modified |
| `U` | unstage-all | Unstage all |
| `c` | commit | Open commit editor |
| `A` | amend | Amend last commit |
| `d` / `enter` | show-diff | View file changes |
| `D` | discard-changes | Discard unstaged changes |
| `h` | show-history | Open commit history |
| `P` | push | Open push menu |
| `L` | pull | Open pull menu |
| `f` | fetch | Fetch from remote |
| `b` | branch | Branch operations |
| `z` | stash | Stash changes |
| `Z` | stash-pop | Pop stash |
| `ctrl+z` | stash-apply | Stash apply |
| `o` | open-in-github | Open in GitHub |
| `O` | open-in-file-browser | Open in file browser |
| `y` | yank-file | Copy file info |
| `Y` | yank-path | Copy file path |

### Commit List Shortcuts

| Key | Command | Description |
|-----|---------|-------------|
| `enter` / `d` | view-commit | Open commit details |
| `h` | show-history | Open history view |
| `y` | yank-commit | Copy commit as markdown |
| `Y` | yank-id | Copy commit hash |
| `/` | search-history | Search commit messages |
| `f` | filter-author | Filter by author |
| `p` | filter-path | Filter by path |
| `F` | clear-filter | Clear filters |
| `n` | next-match | Next search match |
| `N` | prev-match | Previous match |
| `o` | open-in-github | Open commit in GitHub |
| `v` | toggle-graph | Toggle commit graph |

### Pull Menu

| Key | Command |
|-----|---------|
| `p` | pull-merge |
| `r` | pull-rebase |
| `f` | pull-ff-only |
| `a` | pull-autostash |

## File Browser Plugin

### Contexts

| Context | View |
|---------|------|
| `file-browser-tree` | Tree view (root) |
| `file-browser-preview` | Preview pane |
| `file-browser-search` | Filename search |
| `file-browser-content-search` | Content search |
| `file-browser-quick-open` | Fuzzy file finder |
| `file-browser-project-search` | Ripgrep search modal |
| `file-browser-file-op` | File operation input |
| `file-browser-inline-edit` | Inline vim editor (all keys forwarded, global shortcuts bypassed) |

### Tree Shortcuts

| Key | Command | Description |
|-----|---------|-------------|
| `/` | search | Filter files by name |
| `ctrl+p` | quick-open | Fuzzy file finder |
| `ctrl+s` | project-search | Project-wide search |
| `a` | create-file | Create new file |
| `A` | create-dir | Create new directory |
| `d` | delete | Delete (with confirmation) |
| `t` | new-tab | Open in new tab |
| `[` | prev-tab | Previous tab |
| `]` | next-tab | Next tab |
| `x` | close-tab | Close active tab |
| `y` | yank | Copy to clipboard |
| `p` | paste | Paste from clipboard |
| `s` | sort | Cycle sort mode |
| `m` | move | Move file/directory |
| `R` | rename | Rename |
| `ctrl+r` | reveal | Reveal in file manager |

## Conversations Plugin

### Contexts

| Context | View |
|---------|------|
| `conversations` | Session list single-pane (root) |
| `conversations-sidebar` | Session list two-pane (root) |
| `conversations-main` | Messages pane |
| `conversations-search` | Search mode |
| `conversations-filter` | Adapter filter |
| `conversation-detail` | Turn list |
| `message-detail` | Single turn content |
| `analytics` | Usage stats |

## Workspaces Plugin

### Contexts

| Context | View |
|---------|------|
| `workspace-list` | Workspace list (root) |
| `workspace-preview` | Preview pane |
| `workspace-create` | Create worktree input |
| `workspace-task-link` | Task selection modal |
| `workspace-merge` | Merge workflow modal |
| `workspace-interactive` | Embedded terminal |

### List Shortcuts

| Key | Command | Description |
|-----|---------|-------------|
| `n` | new-workspace | Create new workspace |
| `v` | toggle-view | Toggle list/kanban |
| `D` | delete-workspace | Delete workspace |
| `p` | push | Push branch |
| `m` | merge-workflow | Start merge workflow |
| `T` | link-task | Link/unlink task |
| `s` | start-agent | Start agent |
| `enter` | interactive | Enter interactive mode |
| `t` | attach | Attach to tmux |
| `S` | stop-agent | Stop agent |
| `y` | approve | Approve agent prompt |
| `N` | reject | Reject agent prompt |
| `[` | prev-tab | Previous preview tab |
| `]` | next-tab | Next preview tab |

### Interactive Mode

| Key | Command |
|-----|---------|
| `ctrl+\` | exit |
| `ctrl+]` | attach |
| `alt+c` | copy |
| `alt+v` | paste |

## TD Monitor Plugin

Contexts: `td-monitor` (root), `td-modal`, `td-stats`, `td-search`, `td-confirm`, `td-epic-tasks`, `td-parent-epic`, `td-handoffs`.

Shortcuts are defined in TD's `pkg/monitor/keymap/` and auto-exported.

## Project Switcher

| Key | Command |
|-----|---------|
| `@` | toggle |
| `down` / `ctrl+n` | cursor-down |
| `up` / `ctrl+p` | cursor-up |
| `Enter` | select |
| `Esc` | close |

## Command Palette

Press `?` to open. Press `tab` to toggle between current-context and all-contexts view.

| Key | Action |
|-----|--------|
| `j` / `k` / `up` / `down` | Navigate |
| `ctrl+d` / `ctrl+u` | Page down/up |
| `enter` | Execute |
| `esc` | Close |
| `tab` | Toggle context filter |

## Known Conflicts and Design Decisions

Key conflicts exist across plugins (e.g., `d` = delete in file-browser, diff in git, delete-session in conversations). See `references/assessment.md` for the full inconsistency analysis, vim alignment audit, mnemonic analysis, and proposed improvement plan.

### Shift Modifier Convention (Current)

- `s`/`S`: stage / stage-all (git)
- `u`/`U`: unstage / unstage-all (git)
- `d`/`D`: diff / discard (git), delete/- (file-browser)
- `y`/`Y`: yank item / yank path
- `n`/`N`: next-match / prev-match (search contexts)
