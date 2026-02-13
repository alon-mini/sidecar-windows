# Fixed Footer Layout Reference

## App Shell Structure

Sidecar renders a fixed app shell:
- **Header**: plugin tabs, title, clock
- **Content**: plugin view area (scrollable within the plugin)
- **Footer**: global + plugin-specific key hints and status

The footer is rendered by the app shell (`internal/app/view.go`), not by plugins.

## Plugin Responsibilities

Plugins only render their main content area. Do not render a footer or assume extra terminal rows beyond your content area.

When implementing a plugin view:
- Treat the `height` argument as the full usable content height
- Keep your own header (optional) inside that height
- Calculate visible rows based on your own header/spacing
- Do not add footer rows or key hints

## Key Hints in the Footer

The footer builds hints from:
- Global bindings (tab switch, help, quit)
- Active plugin commands (`Plugin.Commands()` + active context bindings)

To expose hints for a plugin:
1. Add command entries in `Plugin.Commands()` with the correct `Context`.
2. Ensure key bindings exist in `internal/keymap/bindings.go` (or user overrides).
3. Return the active context from `Plugin.FocusContext()`.

### Example
```go
func (p *Plugin) Commands() []plugin.Command {
    return []plugin.Command{
        {ID: "open-item", Name: "Open", Context: "my-plugin", Priority: 1},
        {ID: "back", Name: "Back", Context: "my-plugin-detail", Priority: 1},
    }
}

func (p *Plugin) FocusContext() string {
    if p.showDetail {
        return "my-plugin-detail"
    }
    return "my-plugin"
}
```

With bindings:
```go
{Key: "enter", Command: "open-item", Context: "my-plugin"},
{Key: "esc", Command: "back", Context: "my-plugin-detail"},
```

### Priority

Use Priority 1 for essential commands, 2-4 for secondary, 5+ for rare. Default (0) is treated as 99 (lowest). Commands shown in priority order (lower = more important).

### Footer Behavior

The footer auto-truncates hints exceeding available width. Keep Command Names to 1-2 words. Plugin hints appear before global hints. Don't redefine global commands.

## Layout Math

Inside a plugin view, compute visible rows by subtracting only the rows you render yourself (headers, section titles, blank lines). Do not subtract footer height or assume extra space beyond the content area provided by the app shell.
