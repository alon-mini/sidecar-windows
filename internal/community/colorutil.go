package community

import (
	"fmt"
	"math"
	"strings"

	"github.com/marcus/sidecar/internal/styles"
)

const achromaticEpsilon = 1e-6

// HexToHSL converts a hex color (#RRGGBB) to HSL (h: 0-360, s: 0-1, l: 0-1).
func HexToHSL(hex string) (h, s, l float64) {
	rgb := styles.HexToRGB(hex)
	r := rgb.R / 255.0
	g := rgb.G / 255.0
	b := rgb.B / 255.0

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2.0

	if math.Abs(max-min) < achromaticEpsilon {
		return 0, 0, l
	}

	d := max - min
	if l > 0.5 {
		s = d / (2.0 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h *= 60

	return h, s, l
}

// HSLToHex converts HSL (h: 0-360, s: 0-1, l: 0-1) to a hex color string.
func HSLToHex(h, s, l float64) string {
	if s == 0 {
		v := clamp(l*255, 0, 255)
		return styles.RGBToHex(styles.RGB{R: v, G: v, B: v})
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	hNorm := h / 360.0
	r := hueToRGB(p, q, hNorm+1.0/3.0)
	g := hueToRGB(p, q, hNorm)
	b := hueToRGB(p, q, hNorm-1.0/3.0)

	return styles.RGBToHex(styles.RGB{
		R: clamp(r*255, 0, 255),
		G: clamp(g*255, 0, 255),
		B: clamp(b*255, 0, 255),
	})
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	default:
		return p
	}
}

// Luminance returns relative luminance (0-1) using sRGB formula.
func Luminance(hex string) float64 {
	rgb := styles.HexToRGB(hex)
	r := linearize(rgb.R / 255.0)
	g := linearize(rgb.G / 255.0)
	b := linearize(rgb.B / 255.0)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func linearize(v float64) float64 {
	if v <= 0.03928 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// Blend mixes two hex colors: result = (1-t)*c1 + t*c2. t is clamped to [0,1].
func Blend(c1, c2 string, t float64) string {
	t = math.Max(0, math.Min(1, t))
	rgb1 := styles.HexToRGB(c1)
	rgb2 := styles.HexToRGB(c2)
	return styles.RGBToHex(styles.RGB{
		R: rgb1.R*(1-t) + rgb2.R*t,
		G: rgb1.G*(1-t) + rgb2.G*t,
		B: rgb1.B*(1-t) + rgb2.B*t,
	})
}

// Lighten increases HSL lightness by pct (0-1).
func Lighten(hex string, pct float64) string {
	h, s, l := HexToHSL(hex)
	l = math.Min(1.0, l+pct)
	return HSLToHex(h, s, l)
}

// Darken decreases HSL lightness by pct (0-1).
func Darken(hex string, pct float64) string {
	h, s, l := HexToHSL(hex)
	l = math.Max(0.0, l-pct)
	return HSLToHex(h, s, l)
}

// Saturation returns the HSL saturation (0-1) of a hex color.
func Saturation(hex string) float64 {
	_, s, _ := HexToHSL(hex)
	return s
}

// HueDegrees returns hue in degrees (0-360).
func HueDegrees(hex string) float64 {
	h, _, _ := HexToHSL(hex)
	return h
}

// ColorDistance returns euclidean distance in RGB space (0-441.67).
func ColorDistance(a, b string) float64 {
	c1 := styles.HexToRGB(a)
	c2 := styles.HexToRGB(b)
	dr := c1.R - c2.R
	dg := c1.G - c2.G
	db := c1.B - c2.B
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// FormatHex ensures a color string is in #rrggbb lowercase format.
func FormatHex(hex string) string {
	rgb := styles.HexToRGB(hex)
	return fmt.Sprintf("#%02x%02x%02x", clampByte(rgb.R), clampByte(rgb.G), clampByte(rgb.B))
}

// WithAlpha returns a lowercase #rrggbbaa color, normalizing the base hex if needed.
func WithAlpha(hex, alpha string) string {
	trimmed := strings.TrimPrefix(hex, "#")
	if len(trimmed) >= 6 {
		return "#" + strings.ToLower(trimmed[:6]) + strings.ToLower(strings.TrimPrefix(alpha, "#"))
	}
	base := FormatHex(hex)
	return base + strings.ToLower(strings.TrimPrefix(alpha, "#"))
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(math.Round(v))
}

// ContrastRatio returns the WCAG 2.0 contrast ratio between two colors (1 to 21).
func ContrastRatio(fg, bg string) float64 {
	l1 := Luminance(fg)
	l2 := Luminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// EnsureContrast adjusts fg until the contrast ratio against bg meets minRatio.
// Blends toward whichever pole (white or black) achieves the target with the smallest shift.
// Returns the original fg if already sufficient.
func EnsureContrast(fg, bg string, minRatio float64) string {
	return ensureContrastForBackgrounds(fg, []string{bg}, minRatio)
}

func ensureContrastForBackgrounds(fg string, bgs []string, minRatio float64) string {
	if len(bgs) == 0 {
		return fg
	}

	if minContrastRatio(fg, bgs) >= minRatio {
		return fg
	}

	targets := []string{"#ffffff", "#000000"}
	bestTarget := ""
	bestBlend := 0.0
	bestBlendFound := false
	bestTargetContrast := 0.0
	bestTargetColor := ""

	for _, target := range targets {
		targetMin := minContrastRatio(target, bgs)
		if targetMin > bestTargetContrast {
			bestTargetContrast = targetMin
			bestTargetColor = target
		}
		if targetMin < minRatio {
			continue
		}
		lo, hi := 0.0, 1.0
		for i := 0; i < 16; i++ {
			mid := (lo + hi) / 2
			if minContrastRatio(Blend(fg, target, mid), bgs) >= minRatio {
				hi = mid
			} else {
				lo = mid
			}
		}
		if !bestBlendFound || hi < bestBlend {
			bestBlendFound = true
			bestBlend = hi
			bestTarget = target
		}
	}

	if bestBlendFound {
		return Blend(fg, bestTarget, bestBlend)
	}
	if bestTargetColor != "" && bestTargetContrast > minContrastRatio(fg, bgs) {
		return bestTargetColor
	}
	return fg
}

func minContrastRatio(fg string, bgs []string) float64 {
	if len(bgs) == 0 {
		return ContrastRatio(fg, "#000000")
	}
	minRatio := math.MaxFloat64
	for _, bg := range bgs {
		if ratio := ContrastRatio(fg, bg); ratio < minRatio {
			minRatio = ratio
		}
	}
	return minRatio
}
