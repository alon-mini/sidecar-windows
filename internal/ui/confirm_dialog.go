package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/marcus/sidecar/internal/modal"
	"github.com/marcus/sidecar/internal/styles"
)

// ConfirmDialog is a reusable confirmation modal with interactive buttons.
type ConfirmDialog struct {
	Title        string
	Message      string
	ConfirmLabel string         // e.g., " Confirm ", " Delete ", " Yes "
	CancelLabel  string         // e.g., " Cancel ", " No "
	BorderColor  lipgloss.Color // Modal border color
	Width        int            // Modal width (default 50)
}

// NewConfirmDialog creates a dialog with sensible defaults.
func NewConfirmDialog(title, message string) *ConfirmDialog {
	return &ConfirmDialog{
		Title:        title,
		Message:      message,
		ConfirmLabel: " Confirm ",
		CancelLabel:  " Cancel ",
		BorderColor:  styles.Primary,
		Width:        ModalWidthMedium,
	}
}

// ToModal adapts the dialog configuration into a modal.Modal instance.
func (d *ConfirmDialog) ToModal() *modal.Modal {
	variant := modal.VariantDefault
	switch d.BorderColor {
	case styles.Error:
		variant = modal.VariantDanger
	case styles.Warning:
		variant = modal.VariantWarning
	case styles.Info:
		variant = modal.VariantInfo
	}

	return modal.New(d.Title,
		modal.WithWidth(d.Width),
		modal.WithVariant(variant),
		modal.WithPrimaryAction("confirm"),
		modal.WithHints(false),
	).
		AddSection(modal.Text(d.Message)).
		AddSection(modal.Spacer()).
		AddSection(modal.Buttons(
			modal.Btn(d.ConfirmLabel, "confirm"),
			modal.Btn(d.CancelLabel, "cancel"),
		))
}
