package community

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/marcus/sidecar/internal/styles"
)

// Convert maps a CommunityScheme to a full Sidecar ColorPalette.
func Convert(scheme *CommunityScheme) styles.ColorPalette {
	bg := scheme.Background
	fg := scheme.Foreground
	isDark := Luminance(bg) < 0.5

	bgSecondary := adjustBg(bg, 0.08, isDark)
	bgTertiary := scheme.SelectionBackground
	if bgTertiary == "" || ColorDistance(bg, bgTertiary) < 20 {
		bgTertiary = adjustBg(bg, 0.16, isDark)
	}

	// Compute muted text colors with contrast enforcement
	textMuted := EnsureContrast(scheme.BrightBlack, bg, 3.0)
	textSubtle := EnsureContrast(Blend(scheme.BrightBlack, bg, 0.30), bg, 2.5)
	tabTextInactive := EnsureContrast(scheme.BrightBlack, bg, 3.0)

	// Ensure primary/secondary text has sufficient contrast against the main background.
	textPrimary := EnsureContrast(fg, bg, 4.5)
	textSecondary := EnsureContrast(Blend(fg, bg, 0.25), bg, 3.5)

	// Compute selection text color with guaranteed contrast against BgTertiary.
	// This is independent of main background - selection text must always be readable.
	textSelection := EnsureContrast(fg, bgTertiary, 4.5)

	// Improve contrast on tertiary backgrounds if we can keep main background contrast.
	if ContrastRatio(textPrimary, bgTertiary) < 4.5 {
		adjusted := EnsureContrast(textPrimary, bgTertiary, 4.5)
		if ContrastRatio(adjusted, bg) >= 4.5 {
			textPrimary = adjusted
		}
	}
	if ContrastRatio(textSecondary, bgTertiary) < 3.5 {
		adjusted := EnsureContrast(textSecondary, bgTertiary, 3.5)
		if ContrastRatio(adjusted, bg) >= 3.5 {
			textSecondary = adjusted
		}
	}

	return styles.ColorPalette{
		Primary:   scheme.Blue,
		Secondary: scheme.Cyan,
		Accent:    scheme.Yellow,

		Success: scheme.Green,
		Warning: scheme.Yellow,
		Error:   scheme.Red,
		Info:    scheme.Cyan,

		TextPrimary:   textPrimary,
		TextSecondary: textSecondary,
		TextMuted:     textMuted,
		TextSubtle:    textSubtle,
		TextSelection: textSelection,
		TextHighlight: scheme.BrightWhite,

		BgPrimary:   bg,
		BgSecondary: bgSecondary,
		BgTertiary:  bgTertiary,
		BgOverlay:   WithAlpha(bg, "80"),

		BorderNormal: scheme.BrightBlack,
		BorderActive: scheme.Blue,
		BorderMuted:  adjustBg(bg, 0.06, isDark),

		GradientBorderActive: []string{scheme.Blue, scheme.Purple},
		GradientBorderNormal: []string{scheme.BrightBlack, bgTertiary},
		GradientBorderAngle:  30,

		TabStyle:  "gradient",
		TabColors: deriveTabGradient(scheme),

		DiffAddFg:    scheme.Green,
		DiffAddBg:    Blend(bg, scheme.Green, 0.15),
		DiffRemoveFg: scheme.Red,
		DiffRemoveBg: Blend(bg, scheme.Red, 0.15),

		ButtonHover:      scheme.Purple,
		TabTextInactive:  tabTextInactive,
		Link:             scheme.BrightBlue,
		ToastSuccessText: contrastText(scheme.Green),
		ToastErrorText:   contrastText(scheme.Red),

		SyntaxTheme:   matchSyntaxTheme(bg),
		MarkdownTheme: markdownTheme(isDark),
	}
}

// PaletteToOverrides serializes a ColorPalette to the override map format for config.json.
func PaletteToOverrides(p styles.ColorPalette) map[string]interface{} {
	m := map[string]interface{}{
		"primary":          p.Primary,
		"secondary":        p.Secondary,
		"accent":           p.Accent,
		"success":          p.Success,
		"warning":          p.Warning,
		"error":            p.Error,
		"info":             p.Info,
		"textPrimary":      p.TextPrimary,
		"textSecondary":    p.TextSecondary,
		"textMuted":        p.TextMuted,
		"textSubtle":       p.TextSubtle,
		"textSelection":    p.TextSelection,
		"textHighlight":    p.TextHighlight,
		"bgPrimary":        p.BgPrimary,
		"bgSecondary":      p.BgSecondary,
		"bgTertiary":       p.BgTertiary,
		"bgOverlay":        p.BgOverlay,
		"borderNormal":     p.BorderNormal,
		"borderActive":     p.BorderActive,
		"borderMuted":      p.BorderMuted,
		"diffAddFg":        p.DiffAddFg,
		"diffAddBg":        p.DiffAddBg,
		"diffRemoveFg":     p.DiffRemoveFg,
		"diffRemoveBg":     p.DiffRemoveBg,
		"buttonHover":      p.ButtonHover,
		"tabTextInactive":  p.TabTextInactive,
		"link":             p.Link,
		"toastSuccessText": p.ToastSuccessText,
		"toastErrorText":   p.ToastErrorText,
		"syntaxTheme":      p.SyntaxTheme,
		"markdownTheme":    p.MarkdownTheme,
		"tabStyle":         p.TabStyle,
	}

	if len(p.GradientBorderActive) > 0 {
		arr := make([]interface{}, len(p.GradientBorderActive))
		for i, c := range p.GradientBorderActive {
			arr[i] = c
		}
		m["gradientBorderActive"] = arr
	}
	if len(p.GradientBorderNormal) > 0 {
		arr := make([]interface{}, len(p.GradientBorderNormal))
		for i, c := range p.GradientBorderNormal {
			arr[i] = c
		}
		m["gradientBorderNormal"] = arr
	}
	m["gradientBorderAngle"] = p.GradientBorderAngle
	if len(p.TabColors) > 0 {
		arr := make([]interface{}, len(p.TabColors))
		for i, c := range p.TabColors {
			arr[i] = c
		}
		m["tabColors"] = arr
	}

	return m
}

// adjustBg lightens for dark themes, darkens for light themes.
func adjustBg(bg string, amount float64, isDark bool) string {
	if isDark {
		return Lighten(bg, amount)
	}
	return Darken(bg, amount)
}

// contrastText returns black or white for maximum contrast.
func contrastText(bg string) string {
	if Luminance(bg) > 0.5 {
		return "#000000"
	}
	return "#ffffff"
}

func markdownTheme(isDark bool) string {
	if isDark {
		return "dark"
	}
	return "light"
}

type colorInfo struct {
	hex string
	sat float64
	hue float64
}

// deriveTabGradient picks 3-4 saturated distinct colors sorted by hue.
func deriveTabGradient(scheme *CommunityScheme) []string {
	candidates := []string{
		scheme.Red, scheme.Green, scheme.Yellow, scheme.Blue,
		scheme.Purple, scheme.Cyan,
		scheme.BrightRed, scheme.BrightGreen, scheme.BrightYellow,
		scheme.BrightBlue, scheme.BrightPurple, scheme.BrightCyan,
	}

	// Filter to saturated colors
	var saturated []colorInfo
	for _, c := range filterCandidateColors(candidates) {
		s := Saturation(c)
		if s > 0.3 {
			saturated = append(saturated, colorInfo{c, s, HueDegrees(c)})
		}
	}

	if len(saturated) < 3 {
		fallback := greedyDistancePick(filterCandidateColors(candidates), 4)
		fallback = ensureFallbackTabColors(fallback, scheme.Background)
		sort.Slice(fallback, func(i, j int) bool {
			return HueDegrees(fallback[i]) < HueDegrees(fallback[j])
		})
		return fallback
	}

	// Sort by saturation descending, take top 6
	sort.Slice(saturated, func(i, j int) bool {
		return saturated[i].sat > saturated[j].sat
	})
	if len(saturated) > 6 {
		saturated = saturated[:6]
	}

	// Greedily pick 4 colors with maximum hue distance
	picked := greedyHuePick(saturated, 4)

	// Sort by hue for smooth gradient
	sort.Slice(picked, func(i, j int) bool {
		return HueDegrees(picked[i]) < HueDegrees(picked[j])
	})

	return picked
}

func filterCandidateColors(candidates []string) []string {
	filtered := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c == "" || !styles.IsValidHexColor(c) {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

func greedyDistancePick(candidates []string, n int) []string {
	if len(candidates) <= n {
		result := make([]string, len(candidates))
		copy(result, candidates)
		return result
	}

	picked := []string{candidates[0]}
	used := map[int]bool{0: true}

	for len(picked) < n {
		bestIdx := -1
		bestMinDist := -1.0

		for i, c := range candidates {
			if used[i] {
				continue
			}
			minDist := math.MaxFloat64
			for _, p := range picked {
				dist := ColorDistance(c, p)
				if dist < minDist {
					minDist = dist
				}
			}
			if minDist > bestMinDist {
				bestMinDist = minDist
				bestIdx = i
			}
		}

		if bestIdx < 0 {
			break
		}
		picked = append(picked, candidates[bestIdx])
		used[bestIdx] = true
	}

	return picked
}

func ensureFallbackTabColors(picked []string, background string) []string {
	seen := make(map[string]bool, len(picked))
	for _, c := range picked {
		seen[strings.ToLower(c)] = true
	}

	addUnique := func(color string) {
		if color == "" {
			return
		}
		key := strings.ToLower(color)
		if seen[key] {
			return
		}
		seen[key] = true
		picked = append(picked, color)
	}

	if len(picked) >= 3 {
		return picked
	}

	isDark := Luminance(background) < 0.5
	addUnique(adjustBg(background, 0.12, isDark))
	addUnique(adjustBg(background, 0.20, isDark))
	addUnique(adjustBg(background, 0.28, isDark))
	addUnique(background)

	if len(picked) > 4 {
		return picked[:4]
	}
	return picked
}

// greedyHuePick selects n colors with maximum mutual hue separation.
func greedyHuePick(candidates []colorInfo, n int) []string {
	if len(candidates) <= n {
		result := make([]string, len(candidates))
		for i, c := range candidates {
			result[i] = c.hex
		}
		return result
	}

	picked := []colorInfo{candidates[0]}
	used := map[int]bool{0: true}

	for len(picked) < n {
		bestIdx := -1
		bestMinDist := -1.0

		for i, c := range candidates {
			if used[i] {
				continue
			}
			minDist := math.MaxFloat64
			for _, p := range picked {
				dist := hueDist(c.hue, p.hue)
				if dist < minDist {
					minDist = dist
				}
			}
			if minDist > bestMinDist {
				bestMinDist = minDist
				bestIdx = i
			}
		}

		if bestIdx < 0 {
			break
		}
		picked = append(picked, candidates[bestIdx])
		used[bestIdx] = true
	}

	result := make([]string, len(picked))
	for i, c := range picked {
		result[i] = c.hex
	}
	return result
}

func hueDist(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}

// Known Chroma theme backgrounds for matching.
var chromaThemes = map[string]string{
	"monokai":          "#272822",
	"dracula":          "#282a36",
	"nord":             "#2e3440",
	"solarized-dark":   "#002b36",
	"github":           "#ffffff",
	"github-dark":      "#24292e",
	"onedark":          "#282c34",
	"gruvbox":          "#282828",
	"catppuccin-mocha": "#1e1e2e",
	"vs":               "#ffffff",
	"solarized-light":  "#fdf6e3",
}

// matchSyntaxTheme finds the closest Chroma theme by background color.
func matchSyntaxTheme(bg string) string {
	isDark := Luminance(bg) < 0.5
	best := "monokai"
	if !isDark {
		best = "github"
	}
	bestDist := math.MaxFloat64

	for name, themeBg := range chromaThemes {
		// Only match dark themes to dark, light to light
		themeIsDark := Luminance(themeBg) < 0.5
		if themeIsDark != isDark {
			continue
		}
		dist := ColorDistance(bg, themeBg)
		if dist < bestDist {
			bestDist = dist
			best = name
		}
	}

	// If no close match, use safe defaults
	if bestDist > 100 {
		if isDark {
			return "monokai"
		}
		return "github"
	}
	return best
}

// FormatSchemeInfo returns a brief description for display.
func FormatSchemeInfo(scheme *CommunityScheme) string {
	lum := Luminance(scheme.Background)
	mode := "dark"
	if lum >= 0.5 {
		mode = "light"
	}
	return fmt.Sprintf("%s (%s)", scheme.Name, mode)
}
