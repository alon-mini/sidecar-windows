---
name: feature-flags
description: Creating and using feature flags in sidecar for gating experimental functionality. Covers flag registration, checking flags in code, config file and CLI overrides, and priority resolution. Use when adding feature flags, toggling features, or gating new functionality behind flags.
user-invocable: false
---

# Feature Flags

Feature flags gate experimental functionality behind user-configurable settings, enabling safe rollout (default off), user opt-in, and easy rollback.

## Checking Feature State

```go
import "github.com/marcus/sidecar/internal/features"

if features.IsEnabled("tmux_interactive_input") {
    // Feature-gated code
}
```

## Adding a New Feature Flag

1. Define the feature in `internal/features/features.go`:

```go
var MyNewFeature = Feature{
    Name:        "my_new_feature",
    Default:     false,
    Description: "Description of what this enables",
}
```

2. Add to the `allFeatures` slice:

```go
var allFeatures = []Feature{
    TmuxInteractiveInput,
    MyNewFeature, // Add here
}
```

3. Use the feature check in your code:

```go
if features.IsEnabled("my_new_feature") {
    // New functionality
}
```

## User Configuration

### Config file (`~/.config/sidecar/config.json`)

```json
{
  "features": {
    "flags": {
      "tmux_interactive_input": true
    }
  }
}
```

### CLI override (takes precedence over config)

```bash
sidecar --enable-feature=tmux_interactive_input
sidecar --disable-feature=tmux_interactive_input
sidecar --enable-feature=feature1,feature2   # Multiple features
```

Unknown feature names in CLI flags produce a warning but do not prevent startup.

## Priority Order

Feature state resolves in this order (first match wins):
1. CLI override (`--enable-feature`, `--disable-feature`)
2. Config file (`features.flags` in config.json)
3. Default value (defined in code)

## Available Features

| Feature | Default | Description |
|---------|---------|-------------|
| `tmux_interactive_input` | true | Write support for tmux panes |
| `tmux_inline_edit` | true | Inline file editing via tmux in files plugin |
| `notes_plugin` | false | Notes plugin for capturing quick notes |

## API Reference

```go
features.IsEnabled(name string) bool           // Check if enabled
features.List() map[string]bool                 // All features with current state
features.ListAll() []Feature                    // All features with metadata
features.SetEnabled(name string, enabled bool) error  // Persist to config
features.SetOverride(name string, enabled bool) // Runtime override (not persisted)
features.IsKnownFeature(name string) bool       // Check if registered
```

## Best Practices

- Use `snake_case` for feature names (e.g., `my_new_feature`)
- New experimental features should default to `false`
- Provide clear descriptions for each feature
- Document features in `docs/guides/feature-flags.md` when adding them
- Remove feature flags once features are stable
