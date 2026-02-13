# Project Switching Implementation Details

## Modal Builder Pattern

```go
m.projectSwitcherModal = modal.New("Switch Project",
    modal.WithWidth(modalW),
    modal.WithHints(false),
).
    AddSection(m.projectSwitcherInputSection()).
    AddSection(m.projectSwitcherCountSection()).
    AddSection(m.projectSwitcherListSection()).
    AddSection(m.projectSwitcherHintsSection())
```

## Custom Section Pattern

Custom sections receive `contentWidth`, `focusID`, and `hoverID`:

```go
modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
    return modal.RenderedSection{
        Content:    renderedString,
        Focusables: []modal.FocusableInfo{...},
    }
}, updateFn)
```

## initProjectSwitcher() (model.go:433-454)

```go
func (m *Model) initProjectSwitcher() {
    m.clearProjectSwitcherModal()
    ti := textinput.New()
    ti.Placeholder = "Filter projects..."
    ti.Focus()
    ti.CharLimit = 50
    ti.Width = 40
    m.projectSwitcherInput = ti
    m.projectSwitcherFiltered = m.cfg.Projects.List
    m.projectSwitcherCursor = 0
    m.projectSwitcherScroll = 0
    for i, proj := range m.projectSwitcherFiltered {
        if proj.Path == m.ui.WorkDir {
            m.projectSwitcherCursor = i
            break
        }
    }
    m.previewProjectTheme()
}
```

## resetProjectSwitcher() (model.go:413-423)

```go
func (m *Model) resetProjectSwitcher() {
    m.showProjectSwitcher = false
    m.projectSwitcherCursor = 0
    m.projectSwitcherScroll = 0
    m.projectSwitcherFiltered = nil
    m.clearProjectSwitcherModal()
    m.resetProjectAdd()
    resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
    theme.ApplyResolved(resolved)
}
```

## Mouse Click IDs (view.go:29-31)

```go
const projectSwitcherItemPrefix = "project-switcher-item-"

func projectSwitcherItemID(idx int) string {
    return fmt.Sprintf("%s%d", projectSwitcherItemPrefix, idx)
}
```

## Esc Behavior (update.go)

```go
case tea.KeyEsc:
    if m.projectSwitcherInput.Value() != "" {
        m.projectSwitcherInput.SetValue("")
        m.projectSwitcherFiltered = m.cfg.Projects.List
        m.projectSwitcherCursor = 0
        m.projectSwitcherScroll = 0
        return m, nil
    }
    m.resetProjectSwitcher()
    m.updateContext()
    return m, nil
```

## Navigation with Scroll (update.go)

```go
case tea.KeyDown:
    m.projectSwitcherCursor++
    if m.projectSwitcherCursor >= len(projects) {
        m.projectSwitcherCursor = len(projects) - 1
    }
    if m.projectSwitcherCursor < 0 {
        m.projectSwitcherCursor = 0
    }
    m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(
        m.projectSwitcherCursor, m.projectSwitcherScroll, 8)
    m.previewProjectTheme()
    return m, nil
```

## Filter Input Handling (update.go)

```go
var cmd tea.Cmd
m.projectSwitcherInput, cmd = m.projectSwitcherInput.Update(msg)
m.projectSwitcherFiltered = filterProjects(allProjects, m.projectSwitcherInput.Value())
m.clearProjectSwitcherModal()
if m.projectSwitcherCursor >= len(m.projectSwitcherFiltered) {
    m.projectSwitcherCursor = len(m.projectSwitcherFiltered) - 1
}
if m.projectSwitcherCursor < 0 {
    m.projectSwitcherCursor = 0
}
m.projectSwitcherScroll = 0
m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(...)
m.previewProjectTheme()
```

## filterProjects() (model.go:457-470)

```go
func filterProjects(all []config.ProjectConfig, query string) []config.ProjectConfig {
    if query == "" {
        return all
    }
    q := strings.ToLower(query)
    var matches []config.ProjectConfig
    for _, p := range all {
        if strings.Contains(strings.ToLower(p.Name), q) ||
           strings.Contains(strings.ToLower(p.Path), q) {
            matches = append(matches, p)
        }
    }
    return matches
}
```

## Scroll Helper (model.go:474-482)

```go
func projectSwitcherEnsureCursorVisible(cursor, scroll, maxVisible int) int {
    if cursor < scroll {
        return cursor
    }
    if cursor >= scroll+maxVisible {
        return cursor - maxVisible + 1
    }
    return scroll
}
```

## Mouse Handling (update.go:1073-1115)

```go
func (m Model) handleProjectSwitcherMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
    m.ensureProjectSwitcherModal()
    if m.projectSwitcherModal == nil {
        return m, nil
    }
    if m.projectSwitcherMouseHandler == nil {
        m.projectSwitcherMouseHandler = mouse.NewHandler()
    }
    action := m.projectSwitcherModal.HandleMouse(msg, m.projectSwitcherMouseHandler)
    if strings.HasPrefix(action, projectSwitcherItemPrefix) {
        var idx int
        if _, err := fmt.Sscanf(action, projectSwitcherItemPrefix+"%d", &idx); err == nil {
            projects := m.projectSwitcherFiltered
            if idx >= 0 && idx < len(projects) {
                selectedProject := projects[idx]
                m.resetProjectSwitcher()
                m.updateContext()
                return m, m.switchProject(selectedProject.Path)
            }
        }
    }
}
```

## List Section with Focusables (view.go:160-257)

```go
func (m *Model) projectSwitcherListSection() modal.Section {
    return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
        var b strings.Builder
        focusables := make([]modal.FocusableInfo, 0, visibleCount)
        for i := scrollOffset; i < scrollOffset+visibleCount && i < len(projects); i++ {
            itemID := projectSwitcherItemID(i)
            isHovered := itemID == hoverID
            // Render project with hover/cursor styling...
            focusables = append(focusables, modal.FocusableInfo{
                ID:      itemID,
                OffsetX: 0,
                OffsetY: lineOffset + (i-scrollOffset)*2,
                Width:   contentWidth,
                Height:  2,  // name + path = 2 lines
            })
        }
        return modal.RenderedSection{Content: b.String(), Focusables: focusables}
    }, m.projectSwitcherListUpdate)
}
```

## Project Item Styling

| State | Name Style |
|-------|------------|
| Normal | Secondary (blue) |
| Cursor or Hover | Primary + Bold |
| Current | Success (green) + Bold |

Current project shows "(current)" label.

## Empty States (view.go:169-178)

1. **No projects configured**: Muted "No projects configured" text
2. **No filter matches**: Shows "No matches"

## switchProject() (model.go:485-577)

```go
func (m *Model) switchProject(projectPath string) tea.Cmd {
    if projectPath == m.ui.WorkDir {
        return func() tea.Msg {
            return ToastMsg{Message: "Already on this project", Duration: 2 * time.Second}
        }
    }
    oldWorkDir := m.ui.WorkDir
    if activePlugin := m.ActivePlugin(); activePlugin != nil {
        state.SetActivePlugin(oldWorkDir, activePlugin.ID())
    }
    // Worktree restoration logic...
    m.ui.WorkDir = targetPath
    m.intro.RepoName = GetRepoName(targetPath)
    resolved := theme.ResolveTheme(m.cfg, targetPath)
    theme.ApplyResolved(resolved)
    startCmds := m.registry.Reinit(targetPath)
    // WindowSizeMsg broadcasting...
    newActivePluginID := state.GetActivePlugin(targetPath)
    if newActivePluginID != "" {
        m.FocusPluginByID(newActivePluginID)
    }
    return tea.Batch(
        tea.Batch(startCmds...),
        func() tea.Msg {
            return ToastMsg{
                Message:  fmt.Sprintf("Switched to %s", GetRepoName(targetPath)),
                Duration: 3 * time.Second,
            }
        },
    )
}
```

## Adding a Keyboard Shortcut

Add case in `update.go` string switch (after KeyType switch), return early:

```go
case "ctrl+d":
    // Custom action
    return m, nil
```

## Adding Project Metadata

1. Extend `config.ProjectConfig` in `internal/config/types.go`
2. Update `filterProjects()` to search new fields
3. Update `projectSwitcherListSection()` to display new fields

## Changing Filter Algorithm

Replace `filterProjects()` body. Options: fuzzy matching, regex, field-specific search (`name:foo`).

## Adding a New Modal Section

1. Create section function returning `modal.Section`:

```go
func (m *Model) projectSwitcherNewSection() modal.Section {
    return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
        return modal.RenderedSection{Content: "..."}
    }, nil)
}
```

2. Add to modal builder in `ensureProjectSwitcherModal()`.

## Worktree-Specific Behavior

- **Deleted worktrees**: Graceful fallback to main branch
- **Last-active restoration**: When switching to project's main repo, restores last-active worktree
- **Non-git repos**: Shows "No worktrees found"

## Testing

Recommended coverage:
1. `filterProjects()` -- various query inputs
2. `projectSwitcherEnsureCursorVisible()` -- scroll boundary cases
3. Keyboard navigation -- cursor bounds, scroll sync
4. Mouse click detection -- via modal library integration

## Example Configs

### Minimal

```json
{
  "projects": {
    "list": [
      {"name": "work", "path": "~/work/main-project"}
    ]
  }
}
```

### Multiple Projects with Themes

```json
{
  "projects": {
    "list": [
      {"name": "sidecar", "path": "~/code/sidecar"},
      {"name": "work", "path": "~/work/main", "theme": "dark"},
      {"name": "personal", "path": "~/code/personal", "theme": "light"}
    ]
  }
}
```

## Troubleshooting

- **"No projects configured"**: Add projects to config file
- **Path doesn't exist**: Verify paths with `ls`
- **Current project not highlighted**: Path in config must match workdir exactly (after `~` expansion)
- **Switch seems to hang**: Complex projects take longer (stopping watchers, scanning directory, loading git status)
