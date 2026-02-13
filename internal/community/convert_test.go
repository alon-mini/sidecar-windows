package community

import (
	"strings"
	"testing"

	"github.com/marcus/sidecar/internal/styles"
)

func TestConvertCatppuccinMocha(t *testing.T) {
	scheme := GetScheme("Catppuccin Mocha")
	if scheme == nil {
		t.Fatal("Catppuccin Mocha not found")
	}

	palette := Convert(scheme)

	// Verify key mappings
	if palette.Primary != scheme.Blue {
		t.Errorf("Primary = %s, want %s (blue)", palette.Primary, scheme.Blue)
	}
	if palette.BgPrimary != scheme.Background {
		t.Errorf("BgPrimary = %s, want %s", palette.BgPrimary, scheme.Background)
	}
	if palette.Error != scheme.Red {
		t.Errorf("Error = %s, want %s (red)", palette.Error, scheme.Red)
	}
	if palette.Success != scheme.Green {
		t.Errorf("Success = %s, want %s (green)", palette.Success, scheme.Green)
	}

	// Verify derived colors are valid hex
	derivedFields := []struct {
		name, val string
	}{
		{"TextSecondary", palette.TextSecondary},
		{"TextPrimary", palette.TextPrimary},
		{"TextSubtle", palette.TextSubtle},
		{"BgSecondary", palette.BgSecondary},
		{"BgTertiary", palette.BgTertiary},
		{"BorderMuted", palette.BorderMuted},
		{"DiffAddBg", palette.DiffAddBg},
		{"DiffRemoveBg", palette.DiffRemoveBg},
	}
	for _, f := range derivedFields {
		if !isValidHex(f.val) {
			t.Errorf("%s = %q, not valid hex", f.name, f.val)
		}
	}

	// Dark theme should get dark markdown theme
	if palette.MarkdownTheme != "dark" {
		t.Errorf("MarkdownTheme = %s, want dark", palette.MarkdownTheme)
	}

	// Syntax theme should be a known chroma theme
	if palette.SyntaxTheme == "" {
		t.Error("SyntaxTheme is empty")
	}
}

func TestConvertLightTheme(t *testing.T) {
	// Find a light theme
	scheme := GetScheme("Alabaster")
	if scheme == nil {
		// Try another known light theme
		scheme = GetScheme("Apple System Colors Light")
	}
	if scheme == nil {
		t.Skip("No known light theme found")
	}

	palette := Convert(scheme)

	if Luminance(scheme.Background) >= 0.5 {
		if palette.MarkdownTheme != "light" {
			t.Errorf("Light theme MarkdownTheme = %s, want light", palette.MarkdownTheme)
		}
	}
}

func TestConvertTabGradient(t *testing.T) {
	scheme := GetScheme("Catppuccin Mocha")
	if scheme == nil {
		t.Fatal("scheme not found")
	}

	palette := Convert(scheme)

	if palette.TabStyle != "gradient" {
		t.Errorf("TabStyle = %s, want gradient", palette.TabStyle)
	}
	if len(palette.TabColors) < 3 {
		t.Errorf("TabColors has %d colors, want >= 3", len(palette.TabColors))
	}
	for i, c := range palette.TabColors {
		if !isValidHex(c) {
			t.Errorf("TabColors[%d] = %q, not valid hex", i, c)
		}
	}
}

func TestConvertGradientBorders(t *testing.T) {
	scheme := GetScheme("Dracula")
	if scheme == nil {
		t.Fatal("Dracula not found")
	}

	palette := Convert(scheme)

	if len(palette.GradientBorderActive) != 2 {
		t.Errorf("GradientBorderActive has %d colors, want 2", len(palette.GradientBorderActive))
	}
	if len(palette.GradientBorderNormal) != 2 {
		t.Errorf("GradientBorderNormal has %d colors, want 2", len(palette.GradientBorderNormal))
	}
	if palette.GradientBorderAngle != 30 {
		t.Errorf("GradientBorderAngle = %f, want 30", palette.GradientBorderAngle)
	}
}

func TestPaletteToOverrides(t *testing.T) {
	scheme := GetScheme("Catppuccin Mocha")
	if scheme == nil {
		t.Fatal("scheme not found")
	}

	palette := Convert(scheme)
	overrides := PaletteToOverrides(palette)

	// Verify string fields
	if v, ok := overrides["primary"].(string); !ok || v != palette.Primary {
		t.Errorf("overrides[primary] = %v, want %s", overrides["primary"], palette.Primary)
	}
	if v, ok := overrides["bgPrimary"].(string); !ok || v != palette.BgPrimary {
		t.Errorf("overrides[bgPrimary] = %v, want %s", overrides["bgPrimary"], palette.BgPrimary)
	}

	// Verify gradient arrays
	if arr, ok := overrides["gradientBorderActive"].([]interface{}); !ok || len(arr) != 2 {
		t.Errorf("gradientBorderActive not a 2-element array: %v", overrides["gradientBorderActive"])
	}

	// Verify tab colors
	if arr, ok := overrides["tabColors"].([]interface{}); !ok || len(arr) < 3 {
		t.Errorf("tabColors should have >= 3 elements: %v", overrides["tabColors"])
	}

	// Verify angle
	if v, ok := overrides["gradientBorderAngle"].(float64); !ok || v != 30 {
		t.Errorf("gradientBorderAngle = %v, want 30", overrides["gradientBorderAngle"])
	}
}

func TestMatchSyntaxTheme(t *testing.T) {
	// Dark themes
	got := matchSyntaxTheme("#282a36")
	if got != "dracula" {
		t.Errorf("matchSyntaxTheme(#282a36) = %s, want dracula", got)
	}
	got = matchSyntaxTheme("#1e1e2e")
	if got != "catppuccin-mocha" {
		t.Errorf("matchSyntaxTheme(#1e1e2e) = %s, want catppuccin-mocha", got)
	}
	// Light themes - #ffffff matches both "github" and "vs"
	got = matchSyntaxTheme("#ffffff")
	if got != "github" && got != "vs" {
		t.Errorf("matchSyntaxTheme(#ffffff) = %s, want github or vs", got)
	}
}

func TestDeriveTabGradient(t *testing.T) {
	scheme := &CommunityScheme{
		Red: "#ff0000", Green: "#00ff00", Yellow: "#ffff00",
		Blue: "#0000ff", Purple: "#ff00ff", Cyan: "#00ffff",
		BrightRed: "#ff5555", BrightGreen: "#55ff55", BrightYellow: "#ffff55",
		BrightBlue: "#5555ff", BrightPurple: "#ff55ff", BrightCyan: "#55ffff",
	}
	result := deriveTabGradient(scheme)
	if len(result) < 3 || len(result) > 4 {
		t.Errorf("deriveTabGradient returned %d colors, want 3-4", len(result))
	}

	// Verify sorted by hue
	for i := 1; i < len(result); i++ {
		if HueDegrees(result[i]) < HueDegrees(result[i-1]) {
			t.Errorf("tab gradient not sorted by hue at index %d", i)
		}
	}
}

func TestDeriveTabGradientDesaturatedFallback(t *testing.T) {
	scheme := &CommunityScheme{
		Red: "#666666", Green: "#707070", Yellow: "#7a7a7a",
		Blue: "#848484", Purple: "#8e8e8e", Cyan: "#989898",
		BrightRed: "#a2a2a2", BrightGreen: "#acacac", BrightYellow: "#b6b6b6",
		BrightBlue: "#c0c0c0", BrightPurple: "#cacaca", BrightCyan: "#d4d4d4",
		Background: "#1a1a1a",
	}
	result := deriveTabGradient(scheme)
	if len(result) < 3 || len(result) > 4 {
		t.Errorf("deriveTabGradient returned %d colors, want 3-4", len(result))
	}

	unique := make(map[string]bool)
	for i, c := range result {
		if !isValidHex(c) {
			t.Errorf("TabColors[%d] = %q, not valid hex", i, c)
		}
		unique[strings.ToLower(c)] = true
	}
	if len(unique) < 3 {
		t.Errorf("deriveTabGradient returned %d unique colors, want >= 3", len(unique))
	}
}

func TestConvertSelectionBackgroundFallback(t *testing.T) {
	base := GetScheme("Catppuccin Mocha")
	if base == nil {
		t.Skip("Catppuccin Mocha not found")
	}
	scheme := *base
	scheme.SelectionBackground = base.Background

	palette := Convert(&scheme)
	if palette.BgTertiary == scheme.Background {
		t.Errorf("BgTertiary = %s, want fallback distinct from background", palette.BgTertiary)
	}
}

func TestConvertSelectionBackgroundEmpty(t *testing.T) {
	base := GetScheme("Catppuccin Mocha")
	if base == nil {
		t.Skip("Catppuccin Mocha not found")
	}
	scheme := *base
	scheme.SelectionBackground = ""

	palette := Convert(&scheme)
	if palette.BgTertiary == scheme.Background {
		t.Errorf("BgTertiary = %s, want fallback distinct from background", palette.BgTertiary)
	}
}

func TestConvertBgOverlayHandlesAlpha(t *testing.T) {
	base := GetScheme("Catppuccin Mocha")
	if base == nil {
		t.Skip("Catppuccin Mocha not found")
	}
	scheme := *base
	scheme.Background = "#112233aa"

	palette := Convert(&scheme)
	if palette.BgOverlay != "#11223380" {
		t.Errorf("BgOverlay = %s, want #11223380", palette.BgOverlay)
	}
}

func TestPaletteToOverridesZeroGradientAngle(t *testing.T) {
	palette := styles.ColorPalette{
		GradientBorderActive: []string{"#111111", "#222222"},
		GradientBorderAngle:  0,
	}
	overrides := PaletteToOverrides(palette)

	if v, ok := overrides["gradientBorderAngle"].(float64); !ok || v != 0 {
		t.Errorf("gradientBorderAngle = %v, want 0", overrides["gradientBorderAngle"])
	}
}

func TestAllSchemesMinimumContrast(t *testing.T) {
	schemes := ListSchemes()
	if len(schemes) == 0 {
		t.Fatal("no schemes loaded")
	}

	for _, name := range schemes {
		scheme := GetScheme(name)
		if scheme == nil {
			continue
		}
		palette := Convert(scheme)
		bg := palette.BgPrimary
		bgTertiary := palette.BgTertiary

		// TextMuted must have at least 3:1 contrast against background
		if ratio := ContrastRatio(palette.TextMuted, bg); ratio < 3.0 {
			t.Errorf("%s: TextMuted contrast %.2f < 3.0 (fg=%s, bg=%s)", name, ratio, palette.TextMuted, bg)
		}
		// TextSubtle must have at least 2.5:1
		if ratio := ContrastRatio(palette.TextSubtle, bg); ratio < 2.5 {
			t.Errorf("%s: TextSubtle contrast %.2f < 2.5 (fg=%s, bg=%s)", name, ratio, palette.TextSubtle, bg)
		}
		// TabTextInactive must have at least 3:1
		if ratio := ContrastRatio(palette.TabTextInactive, bg); ratio < 3.0 {
			t.Errorf("%s: TabTextInactive contrast %.2f < 3.0 (fg=%s, bg=%s)", name, ratio, palette.TabTextInactive, bg)
		}
		// TextPrimary must have at least 4.5:1 against primary background.
		if ratio := ContrastRatio(palette.TextPrimary, bg); ratio < 4.5 {
			t.Errorf("%s: TextPrimary/BgPrimary contrast %.2f < 4.5 (fg=%s, bg=%s)", name, ratio, palette.TextPrimary, bg)
		}
		// TextSecondary must have at least 3.5:1 against primary background.
		if ratio := ContrastRatio(palette.TextSecondary, bg); ratio < 3.5 {
			t.Errorf("%s: TextSecondary/BgPrimary contrast %.2f < 3.5 (fg=%s, bg=%s)", name, ratio, palette.TextSecondary, bg)
		}
		derivedTertiary := scheme.SelectionBackground == "" || ColorDistance(bg, scheme.SelectionBackground) < 20
		if derivedTertiary {
			// TextPrimary must have at least 4.5:1 against BgTertiary (selected rows).
			if ratio := ContrastRatio(palette.TextPrimary, bgTertiary); ratio < 4.5 {
				t.Errorf("%s: TextPrimary/BgTertiary contrast %.2f < 4.5 (fg=%s, bg=%s)", name, ratio, palette.TextPrimary, bgTertiary)
			}
			// TextSecondary must have at least 3.5:1 against BgTertiary (buttons).
			if ratio := ContrastRatio(palette.TextSecondary, bgTertiary); ratio < 3.5 {
				t.Errorf("%s: TextSecondary/BgTertiary contrast %.2f < 3.5 (fg=%s, bg=%s)", name, ratio, palette.TextSecondary, bgTertiary)
			}
		}
	}
}

func isValidHex(s string) bool {
	if len(s) < 7 || s[0] != '#' {
		return false
	}
	hex := s[1:7]
	for _, c := range hex {
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}
