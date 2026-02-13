package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcus/sidecar/internal/community"
	"github.com/marcus/sidecar/internal/modal"
	"github.com/marcus/sidecar/internal/mouse"
	"github.com/marcus/sidecar/internal/styles"
	"github.com/marcus/sidecar/internal/ui"
)

const (
	themeSwitcherFilterID   = "theme-switcher-filter"
	themeSwitcherItemPrefix = "theme-switcher-item-"
)

// themeEntry represents a single theme in the unified theme list.
type themeEntry struct {
	Name          string // display name
	IsBuiltIn     bool
	ThemeKey      string // built-in: theme registry key; community: scheme name
	IsSeparator   bool   // non-selectable separator line
	SeparatorText string // e.g. "Community Themes (453)"
}

// buildUnifiedThemeList returns all themes: built-in first, then community.
func buildUnifiedThemeList() []themeEntry {
	builtIn := styles.ListThemes()
	communityNames := community.ListSchemes()
	entries := make([]themeEntry, 0, len(builtIn)+len(communityNames))
	for _, name := range builtIn {
		displayName := name
		if t := styles.GetTheme(name); t.DisplayName != "" {
			displayName = t.DisplayName
		}
		entries = append(entries, themeEntry{Name: displayName, IsBuiltIn: true, ThemeKey: name})
	}
	if len(communityNames) > 0 {
		entries = append(entries, themeEntry{
			IsSeparator:   true,
			SeparatorText: fmt.Sprintf("Community Themes (%d)", len(communityNames)),
		})
	}
	for _, name := range communityNames {
		entries = append(entries, themeEntry{Name: name, IsBuiltIn: false, ThemeKey: name})
	}
	return entries
}

// filterThemeEntries filters entries by case-insensitive substring match on Name.
// Separators are included only when unfiltered; they are excluded when a query is active.
func filterThemeEntries(entries []themeEntry, query string) []themeEntry {
	if query == "" {
		return entries
	}
	q := strings.ToLower(query)
	var matches []themeEntry
	for _, e := range entries {
		if e.IsSeparator {
			continue
		}
		if strings.Contains(strings.ToLower(e.Name), q) {
			matches = append(matches, e)
		}
	}
	return matches
}

// themeSwitcherItemID returns the ID for a theme item at the given index.
func themeSwitcherItemID(idx int) string {
	return fmt.Sprintf("%s%d", themeSwitcherItemPrefix, idx)
}

// ensureThemeSwitcherModal builds/rebuilds the theme switcher modal.
func (m *Model) ensureThemeSwitcherModal() {
	modalW := 72
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalW < 30 {
		modalW = 30
	}

	// Only rebuild if modal doesn't exist or width changed
	if m.themeSwitcherModal != nil && m.themeSwitcherModalWidth == modalW {
		return
	}
	m.themeSwitcherModalWidth = modalW

	m.themeSwitcherModal = modal.New("Switch Theme",
		modal.WithWidth(modalW),
		modal.WithHints(false),
	).
		AddSection(modal.When(m.themeSwitcherHasProject, m.themeSwitcherScopeSection())).
		AddSection(modal.Input(themeSwitcherFilterID, &m.themeSwitcherInput, modal.WithSubmitOnEnter(false))).
		AddSection(m.themeSwitcherCountSection()).
		AddSection(modal.Spacer()).
		AddSection(m.themeSwitcherListSection()).
		AddSection(modal.Spacer()).
		AddSection(m.themeSwitcherHintsSection())
}

// themeSwitcherHasProject returns true if the current project is in the project list.
func (m *Model) themeSwitcherHasProject() bool {
	return m.currentProjectConfig() != nil
}

// themeSwitcherCountSection renders the theme count info.
func (m *Model) themeSwitcherCountSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		all := buildUnifiedThemeList()
		allCount := 0
		for _, e := range all {
			if !e.IsSeparator {
				allCount++
			}
		}
		filteredCount := 0
		for _, e := range m.themeSwitcherFiltered {
			if !e.IsSeparator {
				filteredCount++
			}
		}

		var text string
		if m.themeSwitcherInput.Value() != "" {
			text = fmt.Sprintf("%d of %d themes", filteredCount, allCount)
		}

		if text == "" {
			return modal.RenderedSection{Content: ""}
		}
		return modal.RenderedSection{Content: styles.Muted.Render(text)}
	}, nil)
}

// themeSwitcherListSection renders the theme list with selection.
func (m *Model) themeSwitcherListSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		themes := m.themeSwitcherFiltered

		if len(themes) == 0 {
			return modal.RenderedSection{Content: styles.Muted.Render("No matches")}
		}

		cursorStyle := lipgloss.NewStyle().Foreground(styles.Primary)
		nameNormalStyle := lipgloss.NewStyle().Foreground(styles.Secondary)
		nameSelectedStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
		nameCurrentStyle := lipgloss.NewStyle().Foreground(styles.Success).Bold(true)

		maxVisible := 12
		visibleCount := min(maxVisible, len(themes))

		selectedIdx := m.themeSwitcherSelectedIdx
		if selectedIdx < 0 {
			selectedIdx = 0
		}
		if selectedIdx >= len(themes) {
			selectedIdx = len(themes) - 1
		}

		scrollOffset := 0
		if selectedIdx >= maxVisible {
			scrollOffset = selectedIdx - maxVisible + 1
		}
		if scrollOffset > len(themes)-visibleCount {
			scrollOffset = len(themes) - visibleCount
		}
		if scrollOffset < 0 {
			scrollOffset = 0
		}

		var sb strings.Builder
		focusables := make([]modal.FocusableInfo, 0, visibleCount)
		lineOffset := 0

		if scrollOffset > 0 {
			sb.WriteString(styles.Muted.Render(fmt.Sprintf("  ↑ %d more above", scrollOffset)))
			sb.WriteString("\n")
			lineOffset++
		}

		for i := scrollOffset; i < scrollOffset+visibleCount && i < len(themes); i++ {
			entry := themes[i]

			// Render separator lines (non-selectable)
			if entry.IsSeparator {
				sb.WriteString(styles.Muted.Render(fmt.Sprintf("  ── %s ──", entry.SeparatorText)))
				sb.WriteString("\n")
				continue
			}

			isSelected := i == selectedIdx
			itemID := themeSwitcherItemID(i)
			isHovered := itemID == hoverID
			isCurrent := entry.IsBuiltIn == m.themeSwitcherOriginal.IsBuiltIn && entry.ThemeKey == m.themeSwitcherOriginal.ThemeKey

			if isSelected {
				sb.WriteString(cursorStyle.Render("> "))
			} else {
				sb.WriteString("  ")
			}

			// Color swatch for all themes
			if entry.IsBuiltIn {
				t := styles.GetTheme(entry.ThemeKey)
				swatchColors := []string{t.Colors.Primary, t.Colors.Success, t.Colors.Secondary, t.Colors.Error}
				for _, sc := range swatchColors {
					sb.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(sc)).Render(" "))
				}
				sb.WriteString(" ")
			} else {
				scheme := community.GetScheme(entry.ThemeKey)
				if scheme != nil {
					swatchColors := []string{scheme.Red, scheme.Green, scheme.Blue, scheme.Purple}
					for _, sc := range swatchColors {
						sb.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(sc)).Render(" "))
					}
					sb.WriteString(" ")
				}
			}

			var nameStyle lipgloss.Style
			if isCurrent {
				nameStyle = nameCurrentStyle
			} else if isSelected || isHovered {
				nameStyle = nameSelectedStyle
			} else {
				nameStyle = nameNormalStyle
			}

			sb.WriteString(nameStyle.Render(entry.Name))

			if isCurrent {
				sb.WriteString(styles.Muted.Render(" (current)"))
			}
			sb.WriteString("\n")

			focusables = append(focusables, modal.FocusableInfo{
				ID:      itemID,
				OffsetX: 0,
				OffsetY: lineOffset + (i - scrollOffset),
				Width:   contentWidth,
				Height:  1,
			})
		}

		remaining := len(themes) - (scrollOffset + visibleCount)
		if remaining > 0 {
			sb.WriteString(styles.Muted.Render(fmt.Sprintf("  ↓ %d more below", remaining)))
		}

		return modal.RenderedSection{Content: strings.TrimRight(sb.String(), "\n"), Focusables: focusables}
	}, m.themeSwitcherListUpdate)
}

// themeSwitcherListUpdate handles key events for the theme list.
func (m *Model) themeSwitcherListUpdate(msg tea.Msg, focusID string) (string, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return "", nil
	}

	themes := m.themeSwitcherFiltered
	if len(themes) == 0 {
		return "", nil
	}

	switch keyMsg.String() {
	case "up", "k", "ctrl+p":
		if m.themeSwitcherSelectedIdx > 0 {
			m.themeSwitcherSelectedIdx--
			// Skip separators
			for m.themeSwitcherSelectedIdx > 0 && themes[m.themeSwitcherSelectedIdx].IsSeparator {
				m.themeSwitcherSelectedIdx--
			}
			m.themeSwitcherModalWidth = 0
			if m.themeSwitcherSelectedIdx < len(themes) && !themes[m.themeSwitcherSelectedIdx].IsSeparator {
				m.previewThemeEntry(themes[m.themeSwitcherSelectedIdx])
			}
		}
		return "", nil

	case "down", "j", "ctrl+n":
		if m.themeSwitcherSelectedIdx < len(themes)-1 {
			m.themeSwitcherSelectedIdx++
			// Skip separators
			for m.themeSwitcherSelectedIdx < len(themes)-1 && themes[m.themeSwitcherSelectedIdx].IsSeparator {
				m.themeSwitcherSelectedIdx++
			}
			m.themeSwitcherModalWidth = 0
			if m.themeSwitcherSelectedIdx < len(themes) && !themes[m.themeSwitcherSelectedIdx].IsSeparator {
				m.previewThemeEntry(themes[m.themeSwitcherSelectedIdx])
			}
		}
		return "", nil

	case "enter":
		if m.themeSwitcherSelectedIdx >= 0 && m.themeSwitcherSelectedIdx < len(themes) && !themes[m.themeSwitcherSelectedIdx].IsSeparator {
			return "select", nil
		}
		return "", nil
	}

	return "", nil
}

// themeSwitcherScopeSection renders the scope selector.
func (m *Model) themeSwitcherScopeSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		var sb strings.Builder

		activeStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)

		scopeGlobal := "Global"
		scopeProject := "This project"
		if m.themeSwitcherScope == "project" {
			sb.WriteString(styles.Muted.Render(scopeGlobal))
			sb.WriteString(styles.Muted.Render("  │  "))
			sb.WriteString(activeStyle.Render(scopeProject))
		} else {
			sb.WriteString(activeStyle.Render(scopeGlobal))
			sb.WriteString(styles.Muted.Render("  │  "))
			sb.WriteString(styles.Muted.Render(scopeProject))
		}

		return modal.RenderedSection{Content: sb.String()}
	}, nil)
}

// themeSwitcherHintsSection renders the help text.
func (m *Model) themeSwitcherHintsSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		var sb strings.Builder

		if len(m.themeSwitcherFiltered) == 0 {
			sb.WriteString(styles.KeyHint.Render("esc"))
			sb.WriteString(styles.Muted.Render(" clear filter  "))
			sb.WriteString(styles.KeyHint.Render("#"))
			sb.WriteString(styles.Muted.Render(" close"))
		} else {
			sb.WriteString(styles.KeyHint.Render("enter"))
			sb.WriteString(styles.Muted.Render(" select  "))
			sb.WriteString(styles.KeyHint.Render("↑/↓"))
			sb.WriteString(styles.Muted.Render(" navigate"))
			if m.currentProjectConfig() != nil {
				sb.WriteString(styles.Muted.Render("  "))
				sb.WriteString(styles.KeyHint.Render("←/→"))
				sb.WriteString(styles.Muted.Render(" scope"))
			}
			sb.WriteString(styles.Muted.Render("  "))
			sb.WriteString(styles.KeyHint.Render("esc"))
			sb.WriteString(styles.Muted.Render(" cancel"))
		}

		return modal.RenderedSection{Content: sb.String()}
	}, nil)
}

// renderThemeSwitcherModal renders the theme switcher modal using the modal library.
func (m *Model) renderThemeSwitcherModal(content string) string {
	m.ensureThemeSwitcherModal()
	if m.themeSwitcherModal == nil {
		return content
	}

	if m.themeSwitcherMouseHandler == nil {
		m.themeSwitcherMouseHandler = mouse.NewHandler()
	}
	modalContent := m.themeSwitcherModal.Render(m.width, m.height, m.themeSwitcherMouseHandler)
	return ui.OverlayModal(content, modalContent, m.width, m.height)
}
