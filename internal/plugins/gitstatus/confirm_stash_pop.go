package gitstatus

import (
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcus/sidecar/internal/modal"
	"github.com/marcus/sidecar/internal/plugin"
	"github.com/marcus/sidecar/internal/styles"
	"github.com/marcus/sidecar/internal/ui"
)

// buildStashPopModal creates or updates the stash pop confirmation modal.
func (p *Plugin) buildStashPopModal() {
	if p.stashPopItem == nil {
		p.stashPopModal = nil
		return
	}

	stash := p.stashPopItem

	// Truncate long messages (rune-based for Unicode safety)
	msg := stash.Message
	if runes := []rune(msg); len(runes) > 50 {
		msg = string(runes[:47]) + "..."
	}

	modalWidth := 55
	if modalWidth > p.width-10 {
		modalWidth = p.width - 10
	}
	if modalWidth < 20 {
		modalWidth = 20
	}

	sections := []modal.Section{
		modal.Text(styles.Subtitle.Render(stash.Ref)),
	}
	if msg != "" {
		sections = append(sections, modal.Text(styles.Muted.Render(msg)))
	}
	sections = append(sections,
		modal.Spacer(),
		modal.Text(lipgloss.NewStyle().Foreground(styles.Warning).Bold(true).Render("Warning: ")+"This may cause merge conflicts."),
		modal.Text(styles.Muted.Render("The stash will be removed if successful.")),
		modal.Spacer(),
		modal.Buttons(
			modal.Btn(" Pop ", "pop", modal.BtnDanger()),
			modal.Btn(" Cancel ", "cancel"),
		),
	)

	m := modal.New("Pop Stash",
		modal.WithVariant(modal.VariantDanger),
		modal.WithWidth(modalWidth),
	)
	for _, s := range sections {
		m = m.AddSection(s)
	}
	p.stashPopModal = m
}

// renderConfirmStashPop renders the confirm stash pop modal overlay.
func (p *Plugin) renderConfirmStashPop() string {
	background := p.renderThreePaneView()

	if p.stashPopItem == nil {
		return background
	}

	if p.stashPopModal == nil {
		p.buildStashPopModal()
	}

	modalContent := p.stashPopModal.Render(p.width, p.height, p.mouseHandler)
	return ui.OverlayModal(background, modalContent, p.width, p.height)
}

// handleStashPopMouse handles mouse events for the stash pop confirmation modal.
func (p *Plugin) handleStashPopMouse(msg tea.MouseMsg) (plugin.Plugin, tea.Cmd) {
	if p.stashPopModal == nil {
		return p, nil
	}

	action := p.stashPopModal.HandleMouse(msg, p.mouseHandler)

	switch action {
	case "pop":
		return p.executeStashPop()
	case "cancel":
		return p.cancelStashPop()
	}

	return p, nil
}

// executeStashPop performs the stash pop and closes the modal.
func (p *Plugin) executeStashPop() (plugin.Plugin, tea.Cmd) {
	var cmd tea.Cmd
	if p.stashPopItem != nil {
		cmd = p.doStashPop()
	}
	p.viewMode = ViewModeStatus
	p.stashPopItem = nil
	p.stashPopModal = nil
	return p, cmd
}

// cancelStashPop closes the modal without popping.
func (p *Plugin) cancelStashPop() (plugin.Plugin, tea.Cmd) {
	p.viewMode = ViewModeStatus
	p.stashPopItem = nil
	p.stashPopModal = nil
	return p, nil
}

// updateConfirmStashPop handles key events in the confirm stash pop modal.
func (p *Plugin) updateConfirmStashPop(msg tea.KeyMsg) (plugin.Plugin, tea.Cmd) {
	if p.stashPopModal == nil {
		p.buildStashPopModal()
	}
	if p.stashPopModal == nil {
		return p, nil
	}

	// Quick confirm shortcut
	switch msg.String() {
	case "y", "Y":
		return p.executeStashPop()
	}

	action, cmd := p.stashPopModal.HandleKey(msg)

	switch action {
	case "pop":
		return p.executeStashPop()
	case "cancel":
		return p.cancelStashPop()
	}

	return p, cmd
}
