package styles

// TabThemePreset defines a named tab color scheme
type TabThemePreset struct {
	Name        string   // Internal name (e.g., "sunset")
	DisplayName string   // Display name (e.g., "Sunset")
	Style       string   // "gradient", "per-tab", "solid", "minimal"
	Colors      []string // Hex colors for gradient stops or per-tab colors
}

// TabThemePresets contains all built-in tab theme presets
var TabThemePresets = map[string]TabThemePreset{
	// Gradient themes - colors flow continuously across all tabs
	"rainbow": {
		Name:        "rainbow",
		DisplayName: "Rainbow",
		Style:       "gradient",
		Colors:      []string{"#DC3C3C", "#3CDC3C", "#3C3CDC", "#9C3CDC"}, // red → green → blue → purple
	},
	"sunset": {
		Name:        "sunset",
		DisplayName: "Sunset",
		Style:       "gradient",
		Colors:      []string{"#FF6B35", "#F7C59F", "#FF006E"}, // orange → peach → pink
	},
	"ocean": {
		Name:        "ocean",
		DisplayName: "Ocean",
		Style:       "gradient",
		Colors:      []string{"#0077B6", "#00B4D8", "#90E0EF"}, // deep → mid → light blue
	},
	"aurora": {
		Name:        "aurora",
		DisplayName: "Aurora",
		Style:       "gradient",
		Colors:      []string{"#9D4EDD", "#5A189A", "#48BFE3"}, // purple → dark purple → teal
	},
	"neon": {
		Name:        "neon",
		DisplayName: "Neon",
		Style:       "gradient",
		Colors:      []string{"#FF00FF", "#00FFFF", "#00FF00"}, // magenta → cyan → green
	},
	"fire": {
		Name:        "fire",
		DisplayName: "Fire",
		Style:       "gradient",
		Colors:      []string{"#FF4500", "#FF8C00", "#FFD700"}, // red-orange → orange → gold
	},
	"forest": {
		Name:        "forest",
		DisplayName: "Forest",
		Style:       "gradient",
		Colors:      []string{"#2D5016", "#4C8B2F", "#A8E063"}, // dark → mid → light green
	},
	"candy": {
		Name:        "candy",
		DisplayName: "Candy",
		Style:       "gradient",
		Colors:      []string{"#FF69B4", "#9370DB", "#40E0D0"}, // pink → purple → turquoise
	},

	// Per-tab themes - each tab gets a distinct color (cycles through array)
	"pastel": {
		Name:        "pastel",
		DisplayName: "Pastel",
		Style:       "per-tab",
		Colors:      []string{"#FFB3BA", "#BAFFC9", "#BAE1FF", "#FFFFBA"}, // pink, green, blue, yellow
	},
	"jewel": {
		Name:        "jewel",
		DisplayName: "Jewel Tones",
		Style:       "per-tab",
		Colors:      []string{"#9B2335", "#0F4C81", "#5B5EA6", "#9C6644"}, // ruby, sapphire, amethyst, topaz
	},
	"terminal": {
		Name:        "terminal",
		DisplayName: "Terminal",
		Style:       "per-tab",
		Colors:      []string{"#FF5555", "#50FA7B", "#8BE9FD", "#F1FA8C"}, // red, green, cyan, yellow
	},

	// Solid themes - use theme primary/accent colors
	"mono": {
		Name:        "mono",
		DisplayName: "Monochrome",
		Style:       "solid",
		Colors:      []string{}, // Uses theme primary color
	},
	"accent": {
		Name:        "accent",
		DisplayName: "Accent",
		Style:       "solid",
		Colors:      []string{}, // Uses theme accent color
	},

	// Minimal themes - subtle styling without background colors
	"underline": {
		Name:        "underline",
		DisplayName: "Underline",
		Style:       "minimal",
		Colors:      []string{}, // No background, underline active tab
	},
	"dim": {
		Name:        "dim",
		DisplayName: "Dim",
		Style:       "minimal",
		Colors:      []string{}, // No background, dim inactive tabs
	},
}

// GetTabPreset returns a tab theme preset by name, or nil if not found
func GetTabPreset(name string) *TabThemePreset {
	if preset, ok := TabThemePresets[name]; ok {
		return &preset
	}
	return nil
}

// ListTabPresets returns the names of all available tab presets
func ListTabPresets() []string {
	names := make([]string, 0, len(TabThemePresets))
	for name := range TabThemePresets {
		names = append(names, name)
	}
	return names
}
