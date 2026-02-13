package styles

import "math"

func contrastRatio(fg, bg RGB) float64 {
	l1 := relativeLuminance(fg)
	l2 := relativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func minContrastRatio(fg RGB, bgs []RGB) float64 {
	if len(bgs) == 0 {
		return contrastRatio(fg, RGB{0, 0, 0})
	}
	minRatio := math.MaxFloat64
	for _, bg := range bgs {
		if ratio := contrastRatio(fg, bg); ratio < minRatio {
			minRatio = ratio
		}
	}
	return minRatio
}

func relativeLuminance(c RGB) float64 {
	r := linearize(c.R / 255.0)
	g := linearize(c.G / 255.0)
	b := linearize(c.B / 255.0)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func linearize(v float64) float64 {
	if v <= 0.03928 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}
