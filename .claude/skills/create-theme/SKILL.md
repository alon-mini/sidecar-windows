---
name: create-theme
description: >
  Create custom color themes for Sidecar, including base theme selection,
  color overrides, gradient borders, tab styles, per-project themes,
  community themes, and programmatic theme registration. Use when creating
  or modifying themes, adjusting UI appearance, or debugging color/style
  issues. See references/palette-reference.md for the full color palette
  with all keys and per-theme values.
---

# Create Theme

## Configuration Location

Themes are configured in `~/.config/sidecar/config.json`:

```json
{
  "ui": {
    "showFooter": true,
    "showClock": true,
    "nerdFontsEnabled": false,
    "theme": {
      "name": "default",
      "overrides": {
        "primary": "#FF5500",
        "success": "#00FF00"
      }
    }
  }
}
```

## Available Base Themes

- **default** - Dark theme with purple/blue accents
- **dracula** - Dracula-inspired dark theme with vibrant colors
- **molokai** - Vibrant, high-contrast dark theme
- **nord** - Arctic, north-bluish color palette
- **solarized-dark** - Precision color scheme for reduced blue light
- **tokyo-night** - Clean dark theme celebrating Downtown Tokyo

## Creating a Custom Theme

### Method 1: Override Specific Colors

Start from a base theme and override specific colors:
```json
{
  "ui": {
    "theme": {
      "name": "default",
      "overrides": {
        "primary": "#E91E63",
        "success": "#4CAF50",
        "error": "#F44336",
        "syntaxTheme": "github"
      }
    }
  }
}
```

### Method 2: Full Theme Override

Override all colors for complete control. See `references/palette-reference.md` for every available color key and their default values across themes.

### Method 3: Custom Gradient Borders

Panel borders support angled gradients (default 30 degrees) flowing diagonally:
```json
{
  "ui": {
    "theme": {
      "overrides": {
        "gradientBorderActive": ["#FF0000", "#FF7F00", "#FFFF00", "#00FF00", "#0000FF", "#8B00FF"],
        "gradientBorderAngle": 45
      }
    }
  }
}
```

Gradients support 2+ color stops. If not specified, solid `borderActive`/`borderNormal` colors are fallback.

## Tab Styles

Configure with `tabStyle` and `tabColors` in overrides:

**Tab Styles:**
- `gradient` - Colors flow continuously across all tabs (per-character interpolation)
- `per-tab` - Each tab gets a distinct solid color from array (cycles)
- `solid` - Uses theme primary/tertiary colors
- `minimal` - No background, active tab uses underline

**Built-in Presets** (use as `tabStyle` value):
- `rainbow` - Red -> Green -> Blue -> Purple (gradient)
- `sunset` - Orange -> Peach -> Pink (gradient)
- `ocean` - Deep Blue -> Cyan -> Light Blue (gradient)
- `aurora` - Purple -> Dark Purple -> Teal (gradient)
- `neon` - Magenta -> Cyan -> Green (gradient)
- `fire` - Red-Orange -> Orange -> Gold (gradient)
- `forest` - Dark Green -> Mid Green -> Light Green (gradient)
- `candy` - Pink -> Purple -> Turquoise (gradient)
- `pastel` - Pink, Green, Blue, Yellow (per-tab)
- `jewel` - Ruby, Sapphire, Amethyst, Topaz (per-tab)
- `terminal` - Red, Green, Cyan, Yellow (per-tab)
- `mono` - Theme primary color (solid)
- `accent` - Theme accent color (solid)
- `underline` - No background, underlined active (minimal)
- `dim` - No background, dim inactive (minimal)

Examples:
```json
// Use a preset
{ "overrides": { "tabStyle": "sunset" } }

// Custom gradient
{ "overrides": { "tabStyle": "gradient", "tabColors": ["#FF6B35", "#F7C59F", "#FF006E"] } }

// Per-tab distinct colors
{ "overrides": { "tabStyle": "per-tab", "tabColors": ["#FF5555", "#50FA7B", "#8BE9FD", "#F1FA8C"] } }
```

## Color Key Categories

All colors use hex format (`#RRGGBB`). Key categories:

- **Brand**: `primary`, `secondary`, `accent`
- **Status**: `success`, `warning`, `error`, `info`
- **Text**: `textPrimary`, `textSecondary`, `textMuted`, `textSubtle`, `textHighlight`, `textSelection`, `textInverse`
- **Background**: `bgPrimary`, `bgSecondary`, `bgTertiary`, `bgOverlay`
- **Border**: `borderNormal`, `borderActive`, `borderMuted`
- **Gradient border**: `gradientBorderActive`, `gradientBorderNormal` (arrays), `gradientBorderAngle` (number)
- **Tab**: `tabStyle`, `tabColors` (array)
- **Diff**: `diffAddFg`, `diffAddBg`, `diffRemoveFg`, `diffRemoveBg`
- **UI elements**: `buttonHover`, `tabTextInactive`, `link`, `toastSuccessText`, `toastErrorText`
- **Danger**: `dangerLight`, `dangerDark`, `dangerBright`, `dangerHover`
- **Blame age**: `blameAge1` through `blameAge5`
- **Third-party**: `syntaxTheme` (Chroma theme name), `markdownTheme` (`dark`/`light`)

Full color values for all themes: see `references/palette-reference.md`.

## Syntax Themes

The `syntaxTheme` value can be any Chroma theme:
- `monokai`, `dracula`, `github`, `github-dark`, `nord`, `onedark`, `solarized-dark`, `solarized-light`, `vs`, `vim`

See [Chroma Style Gallery](https://xyproto.github.io/splash/docs/) for all options.

## Color Validation

Colors must be valid hex in `#RRGGBB` format. Invalid colors are ignored.
- Valid: `"#FF5500"`, `"#ff5500"` (lowercase ok)
- Invalid: `"FF5500"` (missing #), `"#F50"` (shorthand), `"red"` (named colors)

## Nerd Fonts

When `nerdFontsEnabled` is true: pill-shaped tabs (Powerline chars), pill-shaped buttons. Requires a Nerd Font installed in your terminal.

## Community Themes

Press `#` to open theme switcher, then `Tab` to browse 453 community color schemes. Supports search, live preview, color swatches. Press `Enter` to save.

Community themes are converted from iTerm2 color schemes. Stored by scheme name:
```json
{
  "ui": {
    "theme": {
      "name": "default",
      "community": "Catppuccin Mocha",
      "overrides": { "primary": "#ff79c6" }
    }
  }
}
```

To regenerate community themes from upstream:
```bash
git clone https://github.com/mbadolato/iTerm2-Color-Schemes ~/code/iTerm2-Color-Schemes
./scripts/generate-schemes.sh [path-to-repo]
```

## Per-Project Themes

Each project can have its own theme. When switching with `@`, theme changes automatically.

```json
{
  "projects": {
    "list": [
      { "name": "api", "path": "~/code/api", "theme": { "name": "dracula" } },
      { "name": "web", "path": "~/code/web", "theme": { "name": "default", "community": "Catppuccin Mocha" } },
      { "name": "tools", "path": "~/code/tools" }
    ]
  }
}
```

Set per-project: press `#`, then `ctrl+s` to toggle scope to "Set for this project".

Resolution order: project theme > global `ui.theme` > `"default"`.

## Programmatic Theme Registration

```go
import "github.com/marcus/sidecar/internal/styles"

myTheme := styles.Theme{
    Name:        "my-theme",
    DisplayName: "My Custom Theme",
    Colors: styles.ColorPalette{
        Primary:   "#FF5500",
        Secondary: "#00FF55",
        // ... all other colors
    },
}

styles.RegisterTheme(myTheme)
styles.ApplyTheme("my-theme")
```

## API Reference

```go
styles.ListThemes()                    // []string of available theme names
styles.GetTheme("dracula")             // Theme struct
styles.IsValidTheme("my-theme")        // bool
styles.IsValidHexColor("#FF5500")       // bool
styles.GetCurrentTheme()               // Theme
styles.GetCurrentThemeName()           // string
styles.ApplyTheme("dracula")
styles.ApplyThemeWithOverrides("default", map[string]string{"primary": "#FF5500"})

// Resolve effective theme for a project path (project > global > default)
import "github.com/marcus/sidecar/internal/theme"
resolved := theme.ResolveTheme(cfg, "/path/to/project")
theme.ApplyResolved(resolved)
```

## Design Tips

1. **Contrast**: Ensure text colors have sufficient contrast against backgrounds
2. **Consistency**: Use related colors from the same palette (Tailwind, Material, etc.)
3. **Diff visibility**: Diff backgrounds should be subtle but visible
4. **Toast readability**: Toast text colors should contrast with success/error backgrounds
