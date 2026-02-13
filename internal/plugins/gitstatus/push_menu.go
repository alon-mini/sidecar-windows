package gitstatus

import (
	"github.com/marcus/sidecar/internal/modal"
	"github.com/marcus/sidecar/internal/styles"
	"github.com/marcus/sidecar/internal/ui"
)

const (
	pushMenuOptionPush     = "push-menu-push"
	pushMenuOptionForce    = "push-menu-force"
	pushMenuOptionUpstream = "push-menu-upstream"
	pushMenuActionID       = "push-menu-action"

	pushMenuMinWidth = 20
)

func (p *Plugin) ensurePushMenuModal() {
	modalW := ui.ModalWidthMedium
	if modalW > p.width-4 {
		modalW = p.width - 4
	}
	if modalW < pushMenuMinWidth {
		modalW = pushMenuMinWidth
	}

	if p.pushMenuModal != nil && p.pushMenuModalWidth == modalW {
		return
	}
	p.pushMenuModalWidth = modalW

	items := []modal.ListItem{
		{ID: pushMenuOptionPush, Label: "Push to origin"},
		{ID: pushMenuOptionForce, Label: "Force push (--force-with-lease)"},
		{ID: pushMenuOptionUpstream, Label: "Push & set upstream (-u)"},
	}

	p.pushMenuModal = modal.New("Push",
		modal.WithWidth(modalW),
		modal.WithPrimaryAction(pushMenuActionID),
		modal.WithHints(false),
	).
		AddSection(modal.List("push-options", items, &p.pushMenuFocus, modal.WithMaxVisible(len(items)))).
		AddSection(modal.Spacer()).
		AddSection(p.pushMenuHintsSection())
}

func (p *Plugin) clearPushMenuModal() {
	p.pushMenuModal = nil
	p.pushMenuModalWidth = 0
}

func (p *Plugin) pushMenuHintsSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		return modal.RenderedSection{
			Content: styles.Muted.Render("p/f/u shortcut · Enter to select · Esc to cancel"),
		}
	}, nil)
}

// renderPushMenu renders the push options popup menu.
func (p *Plugin) renderPushMenu() string {
	// Render the background (current view dimmed)
	background := p.renderThreePaneView()

	p.ensurePushMenuModal()
	if p.pushMenuModal == nil {
		return background
	}

	modalContent := p.pushMenuModal.Render(p.width, p.height, p.mouseHandler)
	return ui.OverlayModal(background, modalContent, p.width, p.height)
}
