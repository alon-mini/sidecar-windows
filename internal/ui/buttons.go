package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcus/sidecar/internal/styles"
)

// ResolveButtonStyle returns the appropriate style based on focus and hover state.
// focusIdx: which button index is focused (0-based), -1 for none
// hoverIdx: which button index is hovered (0-based), -1 for none
// btnIdx: the index of the button being styled
func ResolveButtonStyle(focusIdx, hoverIdx, btnIdx int) lipgloss.Style {
	if focusIdx == btnIdx {
		return styles.ButtonFocused
	}
	if hoverIdx == btnIdx {
		return styles.ButtonHover
	}
	return styles.Button
}

// RenderButtonPair renders a confirm/cancel button pair with proper spacing.
// focusIdx: 1 for confirm focused, 2 for cancel focused, 0 for neither
// hoverIdx: 1 for confirm hovered, 2 for cancel hovered, 0 for neither
func RenderButtonPair(confirmLabel, cancelLabel string, focusIdx, hoverIdx int) string {
	confirmStyle := ResolveButtonStyle(focusIdx, hoverIdx, 1)
	cancelStyle := ResolveButtonStyle(focusIdx, hoverIdx, 2)

	var sb strings.Builder
	sb.WriteString(confirmStyle.Render(confirmLabel))
	sb.WriteString("  ")
	sb.WriteString(cancelStyle.Render(cancelLabel))
	return sb.String()
}
