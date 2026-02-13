package styles

import (
	"github.com/marcus/td/pkg/monitor"
)

// CreateTDPanelRenderer creates a PanelRenderer that uses sidecar's gradient borders.
// Maps td monitor PanelState values to appropriate gradients from the current theme.
func CreateTDPanelRenderer() monitor.PanelRenderer {
	return func(content string, width, height int, state monitor.PanelState) string {
		gradient := getTDPanelGradient(state)
		return RenderGradientBorder(content, width, height, gradient, 1)
	}
}

// CreateTDModalRenderer creates a ModalRenderer that uses sidecar's gradient borders.
// Maps td monitor ModalType and depth values to appropriate gradients from the current theme.
func CreateTDModalRenderer() monitor.ModalRenderer {
	return func(content string, width, height int, modalType monitor.ModalType, depth int) string {
		gradient := getTDModalGradient(modalType, depth)
		return RenderGradientBorder(content, width, height, gradient, 1)
	}
}

// getTDPanelGradient returns the appropriate gradient for a panel state.
func getTDPanelGradient(state monitor.PanelState) Gradient {
	theme := GetCurrentTheme()
	angle := theme.Colors.GradientBorderAngle
	if angle == 0 {
		angle = DefaultGradientAngle
	}

	switch state {
	case monitor.PanelStateActive:
		// Active panel: use theme's active gradient (purple→blue)
		colors := theme.Colors.GradientBorderActive
		if len(colors) < 2 {
			colors = []string{theme.Colors.BorderActive, theme.Colors.BorderActive}
		}
		return NewGradient(colors, angle)

	case monitor.PanelStateHover:
		// Hover: lightened version of normal gradient
		colors := theme.Colors.GradientBorderNormal
		if len(colors) < 2 {
			colors = []string{theme.Colors.BorderNormal, theme.Colors.BorderNormal}
		}
		// Lighten colors by blending with white
		lightened := make([]string, len(colors))
		for i, c := range colors {
			rgb := HexToRGB(c)
			// Blend 30% toward white
			lighter := LerpRGB(rgb, RGB{255, 255, 255}, 0.3)
			lightened[i] = RGBToHex(lighter)
		}
		return NewGradient(lightened, angle)

	case monitor.PanelStateDividerHover:
		// Divider hover: cyan gradient
		return NewGradient([]string{"#00BCD4", "#26C6DA"}, angle)

	case monitor.PanelStateDividerActive:
		// Divider active (dragging): orange gradient
		return NewGradient([]string{"#FF9800", "#FFB74D"}, angle)

	default:
		// Normal panel: use theme's normal gradient (dark gray)
		colors := theme.Colors.GradientBorderNormal
		if len(colors) < 2 {
			colors = []string{theme.Colors.BorderNormal, theme.Colors.BorderNormal}
		}
		return NewGradient(colors, angle)
	}
}

// getTDModalGradient returns the appropriate gradient for a modal type and depth.
func getTDModalGradient(modalType monitor.ModalType, depth int) Gradient {
	theme := GetCurrentTheme()
	angle := theme.Colors.GradientBorderAngle
	if angle == 0 {
		angle = DefaultGradientAngle
	}

	// Check for special modal types first
	switch modalType {
	case monitor.ModalTypeHandoffs:
		// Handoffs: green gradient
		return NewGradient([]string{"#10B981", "#34D399"}, angle)

	case monitor.ModalTypeConfirmation:
		// Confirmation: red gradient
		return NewGradient([]string{"#EF4444", "#F87171"}, angle)
	}

	// For other types, use depth-based coloring
	switch depth {
	case 1:
		// Depth 1: active gradient (purple→blue)
		colors := theme.Colors.GradientBorderActive
		if len(colors) < 2 {
			colors = []string{theme.Colors.BorderActive, theme.Colors.BorderActive}
		}
		return NewGradient(colors, angle)

	case 2:
		// Depth 2: cyan gradient
		return NewGradient([]string{"#00BCD4", "#26C6DA"}, angle)

	default:
		// Depth 3+: orange gradient
		return NewGradient([]string{"#FF9800", "#FFB74D"}, angle)
	}
}
