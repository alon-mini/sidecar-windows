package gitstatus

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcus/sidecar/internal/plugin"
	"github.com/marcus/sidecar/internal/styles"
)

// HistorySearchState holds state for commit history search.
type HistorySearchState struct {
	Query     string
	Matches   []*Commit // Commits matching the query
	Cursor    int       // Index in matches for n/N navigation
	Committed bool      // True after Enter (enables n/N)

	// Search options
	UseRegex      bool
	CaseSensitive bool
}

// NewHistorySearchState creates a new search state.
func NewHistorySearchState() *HistorySearchState {
	return &HistorySearchState{
		Matches: make([]*Commit, 0),
	}
}

// Reset clears the search state.
func (s *HistorySearchState) Reset() {
	s.Query = ""
	s.Matches = nil
	s.Cursor = 0
	s.Committed = false
}

// searchCommits filters commits by query (message or author).
func (p *Plugin) searchCommits(query string, useRegex, caseSensitive bool) []*Commit {
	if query == "" {
		return nil
	}

	var matches []*Commit
	var re *regexp.Regexp

	if useRegex {
		flags := ""
		if !caseSensitive {
			flags = "(?i)"
		}
		var err error
		re, err = regexp.Compile(flags + query)
		if err != nil {
			return nil // Invalid regex
		}
	}

	// Search within active commits (filtered or all)
	commits := p.activeCommits()
	for _, c := range commits {
		var match bool
		if useRegex && re != nil {
			match = re.MatchString(c.Subject) || re.MatchString(c.Author)
		} else {
			subject := c.Subject
			author := c.Author
			q := query
			if !caseSensitive {
				subject = strings.ToLower(subject)
				author = strings.ToLower(author)
				q = strings.ToLower(query)
			}
			match = strings.Contains(subject, q) || strings.Contains(author, q)
		}
		if match {
			matches = append(matches, c)
		}
	}

	return matches
}

// findCommitIndex returns the index of a commit in active commits by hash.
func (p *Plugin) findCommitIndex(hash string) int {
	commits := p.activeCommits()
	for i, c := range commits {
		if c.Hash == hash {
			return i
		}
	}
	return -1
}

// renderHistorySearchModal renders the search modal overlay.
func (p *Plugin) renderHistorySearchModal(width int) string {
	state := p.historySearchState
	if state == nil {
		state = NewHistorySearchState()
	}

	// Modal dimensions
	modalWidth := width - 4
	if modalWidth > 70 {
		modalWidth = 70
	}
	if modalWidth < 40 {
		modalWidth = 40
	}

	var sb strings.Builder

	// Header: Search input with cursor
	cursor := "█"
	prefix := "/ "
	available := modalWidth - len(prefix) - 1

	query := state.Query
	if len(query) > available {
		query = "..." + query[len(query)-available+3:]
	}

	header := prefix + query + cursor
	sb.WriteString(styles.ModalTitle.Render(header))
	sb.WriteString("\n")

	// Options bar
	var opts []string
	if state.UseRegex {
		opts = append(opts, styles.BarChipActive.Render(".*"))
	} else {
		opts = append(opts, styles.BarChip.Render(".*"))
	}
	if state.CaseSensitive {
		opts = append(opts, styles.BarChipActive.Render("Aa"))
	} else {
		opts = append(opts, styles.BarChip.Render("Aa"))
	}
	sb.WriteString(strings.Join(opts, " "))
	sb.WriteString("\n\n")

	// Status line
	if state.Query == "" {
		sb.WriteString(styles.Muted.Render("Type to search commits..."))
		sb.WriteString("\n")
	} else if len(state.Matches) == 0 {
		sb.WriteString(styles.Muted.Render("No matches found"))
		sb.WriteString("\n")
	} else {
		// Match count header
		var matchInfo string
		if len(state.Matches) == 1 {
			matchInfo = "1 match"
		} else {
			matchInfo = formatInt(len(state.Matches)) + " matches"
		}
		sb.WriteString(styles.Muted.Render(matchInfo))
		sb.WriteString("\n\n")

		// Display matches (up to 8)
		maxVisible := 8
		if len(state.Matches) < maxVisible {
			maxVisible = len(state.Matches)
		}

		// Calculate scroll offset to keep cursor visible
		scrollOff := 0
		if state.Cursor >= maxVisible {
			scrollOff = state.Cursor - maxVisible + 1
		}

		for i := scrollOff; i < scrollOff+maxVisible && i < len(state.Matches); i++ {
			c := state.Matches[i]

			// Cursor indicator
			if i == state.Cursor {
				sb.WriteString(styles.ListCursor.Render("▸ "))
			} else {
				sb.WriteString("  ")
			}

			// Short hash (muted)
			sb.WriteString(styles.Subtle.Render(c.ShortHash))
			sb.WriteString(" ")

			// Subject (truncate to fit)
			subjectWidth := modalWidth - 12 // hash + cursor + padding
			subject := c.Subject
			if len(subject) > subjectWidth {
				subject = subject[:subjectWidth-3] + "..."
			}
			if i == state.Cursor {
				sb.WriteString(styles.ListItemSelected.Render(subject))
			} else {
				sb.WriteString(subject)
			}
			sb.WriteString("\n")
		}

		// Show scroll indicator if more matches
		if len(state.Matches) > maxVisible {
			remaining := len(state.Matches) - scrollOff - maxVisible
			if remaining > 0 {
				sb.WriteString(styles.Muted.Render("  ↓ " + formatInt(remaining) + " more"))
				sb.WriteString("\n")
			}
		}
	}

	sb.WriteString("\n")
	// Hint
	sb.WriteString(styles.Muted.Render("j/k nav · enter select · alt+r regex · esc cancel"))

	content := sb.String()
	return styles.ModalBox.Width(modalWidth).Render(content)
}

// formatInt converts int to string without importing strconv in view logic.
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + formatInt(-n)
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// updateHistorySearch handles key events when in search mode.
func (p *Plugin) updateHistorySearch(msg tea.KeyMsg) (plugin.Plugin, tea.Cmd) {
	state := p.historySearchState
	if state == nil {
		state = NewHistorySearchState()
		p.historySearchState = state
	}

	key := msg.String()

	switch key {
	case "esc":
		// Cancel search, close modal, and clear search state
		p.historySearchMode = false
		p.clearSearchState()
		return p, nil

	case "enter":
		// Select current match and jump to it
		if len(state.Matches) > 0 {
			state.Committed = true
			p.historySearchMode = false
			return p, p.jumpToSearchMatch()
		}
		// No matches, just close and clear
		p.historySearchMode = false
		p.clearSearchState()
		return p, nil

	case "j", "down", "ctrl+n":
		// Navigate down in matches
		if len(state.Matches) > 0 {
			state.Cursor++
			if state.Cursor >= len(state.Matches) {
				state.Cursor = 0 // Wrap around
			}
		}
		return p, nil

	case "k", "up", "ctrl+p":
		// Navigate up in matches
		if len(state.Matches) > 0 {
			state.Cursor--
			if state.Cursor < 0 {
				state.Cursor = len(state.Matches) - 1 // Wrap around
			}
		}
		return p, nil

	case "backspace":
		if len(state.Query) > 0 {
			state.Query = state.Query[:len(state.Query)-1]
			state.Matches = p.searchCommits(state.Query, state.UseRegex, state.CaseSensitive)
			state.Cursor = 0 // Reset cursor when query changes
		}
		return p, nil

	case "alt+r":
		// Toggle regex
		state.UseRegex = !state.UseRegex
		state.Matches = p.searchCommits(state.Query, state.UseRegex, state.CaseSensitive)
		state.Cursor = 0
		return p, nil

	case "alt+c":
		// Toggle case sensitivity
		state.CaseSensitive = !state.CaseSensitive
		state.Matches = p.searchCommits(state.Query, state.UseRegex, state.CaseSensitive)
		state.Cursor = 0
		return p, nil

	default:
		// Append printable characters to query
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			state.Query += key
			state.Matches = p.searchCommits(state.Query, state.UseRegex, state.CaseSensitive)
			state.Cursor = 0 // Reset cursor when query changes
		}
		return p, nil
	}
}

// clearSearchState clears the history search state.
func (p *Plugin) clearSearchState() {
	if p.historySearchState != nil {
		p.historySearchState.Reset()
	}
}

// updatePathFilter handles key events when in path filter mode.
func (p *Plugin) updatePathFilter(msg tea.KeyMsg) (plugin.Plugin, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		// Cancel path filter, close modal
		p.pathFilterMode = false
		p.pathFilterInput = ""
		return p, nil

	case "enter":
		// Apply path filter
		if p.pathFilterInput != "" {
			p.historyFilterPath = p.pathFilterInput
			p.historyFilterActive = true
			p.pathFilterMode = false
			return p, p.loadFilteredCommits()
		}
		// Empty input, just close
		p.pathFilterMode = false
		return p, nil

	case "backspace":
		if len(p.pathFilterInput) > 0 {
			p.pathFilterInput = p.pathFilterInput[:len(p.pathFilterInput)-1]
		}
		return p, nil

	default:
		// Append printable characters to path
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			p.pathFilterInput += key
		}
		return p, nil
	}
}

// renderPathFilterModal renders the path filter input modal.
func (p *Plugin) renderPathFilterModal(width int) string {
	// Modal dimensions
	modalWidth := width - 4
	if modalWidth > 60 {
		modalWidth = 60
	}
	if modalWidth < 30 {
		modalWidth = 30
	}

	var sb strings.Builder

	// Title
	sb.WriteString(styles.ModalTitle.Render("Filter by Path"))
	sb.WriteString("\n\n")

	// Input with cursor
	cursor := "█"
	prefix := "Path: "
	available := modalWidth - len(prefix) - 1

	input := p.pathFilterInput
	if len(input) > available {
		input = "..." + input[len(input)-available+3:]
	}

	sb.WriteString(prefix + input + cursor)
	sb.WriteString("\n\n")

	// Hint
	sb.WriteString(styles.Muted.Render("Examples: *.go, internal/, README.md"))
	sb.WriteString("\n")
	sb.WriteString(styles.Muted.Render("enter apply · esc cancel"))

	content := sb.String()
	return styles.ModalBox.Width(modalWidth).Render(content)
}

// jumpToSearchMatch moves cursor to the current search match.
func (p *Plugin) jumpToSearchMatch() tea.Cmd {
	state := p.historySearchState
	if state == nil || len(state.Matches) == 0 {
		return nil
	}

	match := state.Matches[state.Cursor]
	idx := p.findCommitIndex(match.Hash)
	if idx < 0 {
		return nil
	}

	// Move cursor to match (entries count + commit index)
	entries := p.tree.AllEntries()
	p.cursor = len(entries) + idx
	p.ensureCommitVisible(idx)

	return p.autoLoadCommitPreview()
}
