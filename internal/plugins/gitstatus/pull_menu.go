package gitstatus

import (
	"fmt"
	"strings"

	"github.com/marcus/sidecar/internal/modal"
	"github.com/marcus/sidecar/internal/styles"
	"github.com/marcus/sidecar/internal/ui"
)

const (
	pullMenuOptionMerge     = "pull-merge"       // List item ID for merge strategy
	pullMenuOptionRebase    = "pull-rebase"      // List item ID for rebase strategy
	pullMenuOptionFFOnly    = "pull-ff-only"     // List item ID for fast-forward only
	pullMenuOptionAutostash = "pull-autostash"   // List item ID for rebase + autostash
	pullMenuActionID        = "pull-menu-action" // Primary action (Enter key)

	pullMenuModalWidth = 50 // Default modal width
	pullMenuMinWidth   = 20 // Minimum modal width

	pullConflictAbortID   = "pull-conflict-abort"
	pullConflictDismissID = "pull-conflict-dismiss"
)

// ensurePullModal builds/rebuilds the pull menu modal.
func (p *Plugin) ensurePullModal() {
	modalW := pullMenuModalWidth
	if modalW > p.width-4 {
		modalW = p.width - 4
	}
	if modalW < pullMenuMinWidth {
		modalW = pullMenuMinWidth
	}

	// Only rebuild if modal doesn't exist or width changed
	if p.pullModal != nil && p.pullModalWidth == modalW {
		return
	}
	p.pullModalWidth = modalW

	items := []modal.ListItem{
		{ID: pullMenuOptionMerge, Label: "Pull (merge)"},
		{ID: pullMenuOptionRebase, Label: "Pull (rebase)"},
		{ID: pullMenuOptionFFOnly, Label: "Pull (fast-forward only)"},
		{ID: pullMenuOptionAutostash, Label: "Pull (rebase + autostash)"},
	}

	p.pullModal = modal.New("Pull",
		modal.WithWidth(modalW),
		modal.WithPrimaryAction(pullMenuActionID),
	).
		AddSection(modal.List("pull-options", items, &p.pullSelectedIdx, modal.WithMaxVisible(4)))
}

// renderPullMenu renders the pull options popup menu.
func (p *Plugin) renderPullMenu() string {
	background := p.renderThreePaneView()

	p.ensurePullModal()
	if p.pullModal == nil {
		return background
	}

	modalContent := p.pullModal.Render(p.width, p.height, p.mouseHandler)
	return ui.OverlayModal(background, modalContent, p.width, p.height)
}

// ensurePullConflictModal builds/rebuilds the pull conflict modal.
func (p *Plugin) ensurePullConflictModal() {
	modalW := ui.ModalWidthMedium
	if modalW > p.width-4 {
		modalW = p.width - 4
	}
	if modalW < pullMenuMinWidth {
		modalW = pullMenuMinWidth
	}

	if p.pullConflictModal != nil && p.pullConflictWidth == modalW {
		return
	}
	p.pullConflictWidth = modalW

	p.pullConflictModal = modal.New("Conflicts",
		modal.WithWidth(modalW),
		modal.WithVariant(modal.VariantDanger),
		modal.WithHints(false),
		modal.WithPrimaryAction(pullConflictAbortID),
	).
		AddSection(p.pullConflictSummarySection()).
		AddSection(modal.Spacer()).
		AddSection(p.pullConflictFilesSection()).
		AddSection(modal.Spacer()).
		AddSection(p.pullConflictResolutionSection()).
		AddSection(modal.Spacer()).
		AddSection(modal.Buttons(
			modal.Btn(" Abort ", pullConflictAbortID, modal.BtnDanger()),
			modal.Btn(" Dismiss ", pullConflictDismissID),
		))
}

// renderPullConflict renders the pull conflict resolution modal.
func (p *Plugin) renderPullConflict() string {
	background := p.renderThreePaneView()

	p.ensurePullConflictModal()
	if p.pullConflictModal == nil {
		return background
	}

	modalContent := p.pullConflictModal.Render(p.width, p.height, p.mouseHandler)
	return ui.OverlayModal(background, modalContent, p.width, p.height)
}

func (p *Plugin) pullConflictSummarySection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		conflictLabel := "Merge"
		if p.pullConflictType == "rebase" {
			conflictLabel = "Rebase"
		}
		content := styles.Muted.Render(fmt.Sprintf("%s produced conflicts in %d file(s):", conflictLabel, len(p.pullConflictFiles)))
		return modal.RenderedSection{Content: content}
	}, nil)
}

func (p *Plugin) pullConflictFilesSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		if len(p.pullConflictFiles) == 0 {
			return modal.RenderedSection{Content: styles.Muted.Render("No conflicted files detected.")}
		}

		var sb strings.Builder
		maxFiles := 8
		for i, f := range p.pullConflictFiles {
			if i >= maxFiles {
				sb.WriteString(styles.Muted.Render(fmt.Sprintf("  ... and %d more", len(p.pullConflictFiles)-maxFiles)))
				break
			}
			sb.WriteString(styles.StatusModified.Render("  U " + f))
			if i < len(p.pullConflictFiles)-1 {
				sb.WriteString("\n")
			}
		}
		return modal.RenderedSection{Content: sb.String()}
	}, nil)
}

func (p *Plugin) pullConflictResolutionSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		content := styles.Muted.Render("Resolve conflicts in your editor, then commit.")
		return modal.RenderedSection{Content: content}
	}, nil)
}
