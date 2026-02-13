package filebrowser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcus/sidecar/internal/modal"
	"github.com/marcus/sidecar/internal/styles"
)

// renderInfoModalContent renders the file info modal.
func (p *Plugin) renderInfoModalContent() string {
	p.ensureInfoModal()
	if p.infoModal == nil {
		return ""
	}
	return p.infoModal.Render(p.width, p.height, p.mouseHandler)
}

// ensureInfoModal builds/rebuilds the info modal.
func (p *Plugin) ensureInfoModal() {
	modalW := 60
	if modalW > p.width-4 {
		modalW = p.width - 4
	}
	if modalW < 30 {
		modalW = 30
	}

	if p.infoModal != nil && p.infoModalWidth == modalW {
		return
	}
	p.infoModalWidth = modalW

	title := "File Info"
	if path := p.infoTargetPath(); path != "" {
		title = filepath.Base(path)
	}

	p.infoModal = modal.New(title,
		modal.WithWidth(modalW),
		modal.WithHints(false),
	).
		AddSection(p.infoModalDetailsSection())
}

func (p *Plugin) clearInfoModal() {
	p.infoModal = nil
	p.infoModalWidth = 0
}

func (p *Plugin) infoTargetPath() string {
	if p.activePane == PanePreview && p.previewFile != "" {
		return p.previewFile
	}
	node := p.tree.GetNode(p.treeCursor)
	if node != nil {
		return node.Path
	}
	return ""
}

func (p *Plugin) infoModalDetailsSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		path := p.infoTargetPath()
		if path == "" {
			return modal.RenderedSection{Content: styles.Muted.Render("No file selected")}
		}

		fullPath := filepath.Join(p.ctx.WorkDir, path)
		info, err := os.Stat(fullPath)
		if err != nil {
			return modal.RenderedSection{Content: styles.StatusDeleted.Render("Error reading file: " + err.Error())}
		}

		isDir := info.IsDir()
		if p.activePane != PanePreview {
			if node := p.tree.GetNode(p.treeCursor); node != nil {
				isDir = node.IsDir
			}
		}

		name := info.Name()
		kind := "File"
		if isDir {
			kind = "Directory"
		} else {
			ext := filepath.Ext(name)
			if ext != "" && len(ext) > 1 {
				kind = strings.ToUpper(ext[1:]) + " File"
			}
		}

		size := formatSize(info.Size())
		if isDir {
			size = "--"
		}

		modTime := info.ModTime().Format("Jan 2, 2006 at 15:04")
		perms := info.Mode().String()

		labelStyle := styles.Muted.Width(12).Align(lipgloss.Right).MarginRight(2)
		valueStyle := lipgloss.NewStyle().Foreground(styles.TextPrimary)

		fields := []struct{ label, value string }{
			{"Kind:", kind},
			{"Size:", size},
			{"Where:", filepath.Dir(path)},
			{"Modified:", modTime},
			{"Permissions:", perms},
			{"Git Status:", p.gitStatus},
			{"Commit:", p.gitLastCommit},
		}

		var sb strings.Builder
		for i, f := range fields {
			line := lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(f.label),
				valueStyle.Render(f.value),
			)
			sb.WriteString(line)
			if i < len(fields)-1 {
				sb.WriteString("\n")
			}
		}

		return modal.RenderedSection{Content: sb.String()}
	}, nil)
}
