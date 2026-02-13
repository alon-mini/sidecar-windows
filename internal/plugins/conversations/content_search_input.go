// Package conversations provides content search keyboard handling for
// cross-conversation search (td-6ac70a).
package conversations

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcus/sidecar/internal/app"
	"github.com/marcus/sidecar/internal/plugin"
)

// handleContentSearchKey handles keyboard input when in content search mode.
func (p *Plugin) handleContentSearchKey(msg tea.KeyMsg) (plugin.Plugin, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		p.contentSearchMode = false
		p.contentSearchState = nil
		p.hitRegionsDirty = true
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		return p, p.jumpToSearchResult()

	case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
		// Use only arrow keys for navigation to allow typing letters (td-2467e8)
		if p.contentSearchState != nil {
			p.contentSearchState.Cursor++
			p.contentSearchState.ClampCursor()
			p.contentSearchState.EnsureCursorVisible(p.contentSearchViewportHeight())
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("up"))):
		if p.contentSearchState != nil {
			p.contentSearchState.Cursor--
			p.contentSearchState.ClampCursor()
			p.contentSearchState.EnsureCursorVisible(p.contentSearchViewportHeight())
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+n"))):
		// Jump to next match (skips session and message rows) - ctrl+n instead of n (td-2467e8)
		if p.contentSearchState != nil {
			nextIdx := p.contentSearchState.NextMatchIndex(p.contentSearchState.Cursor)
			if nextIdx >= 0 {
				p.contentSearchState.Cursor = nextIdx
				p.contentSearchState.EnsureCursorVisible(p.contentSearchViewportHeight())
			}
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+p"))):
		// Jump to previous match (skips session and message rows) - ctrl+p instead of N (td-2467e8)
		if p.contentSearchState != nil {
			prevIdx := p.contentSearchState.PrevMatchIndex(p.contentSearchState.Cursor)
			if prevIdx >= 0 {
				p.contentSearchState.Cursor = prevIdx
				p.contentSearchState.EnsureCursorVisible(p.contentSearchViewportHeight())
			}
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("left"))):
		// Move to session (navigate up in hierarchy) - only arrow key, not 'h' (td-2467e8)
		if p.contentSearchState != nil {
			p.contentSearchState.MoveToSession()
			p.contentSearchState.EnsureCursorVisible(p.contentSearchViewportHeight())
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("right", "tab"))):
		// Toggle collapse/expand for session - use tab instead of space (td-2467e8)
		if p.contentSearchState != nil {
			p.contentSearchState.ToggleCollapse()
			p.contentSearchState.ClampCursor()
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+r", "alt+r"))):
		// Toggle regex mode
		if p.contentSearchState != nil {
			p.contentSearchState.UseRegex = !p.contentSearchState.UseRegex
			return p, p.triggerContentSearch()
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("alt+c"))):
		// Toggle case sensitivity
		if p.contentSearchState != nil {
			p.contentSearchState.CaseSensitive = !p.contentSearchState.CaseSensitive
			return p, p.triggerContentSearch()
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+d"))):
		// Page down
		if p.contentSearchState != nil {
			p.contentSearchState.Cursor += 10
			p.contentSearchState.ClampCursor()
			p.contentSearchState.EnsureCursorVisible(p.contentSearchViewportHeight())
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+u"))):
		// Page up
		if p.contentSearchState != nil {
			p.contentSearchState.Cursor -= 10
			p.contentSearchState.ClampCursor()
			p.contentSearchState.EnsureCursorVisible(p.contentSearchViewportHeight())
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("home", "ctrl+a"))):
		// Go to top - ctrl+a instead of g (td-2467e8)
		if p.contentSearchState != nil {
			p.contentSearchState.Cursor = 0
			p.contentSearchState.ScrollOffset = 0
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("end", "ctrl+e"))):
		// Go to bottom - ctrl+e instead of G (td-2467e8)
		if p.contentSearchState != nil {
			flatLen := p.contentSearchState.FlatLen()
			if flatLen > 0 {
				p.contentSearchState.Cursor = flatLen - 1
				p.contentSearchState.EnsureCursorVisible(p.contentSearchViewportHeight())
			}
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("alt+e"))):
		// Expand all sessions - alt+e instead of E (td-2467e8)
		if p.contentSearchState != nil {
			p.contentSearchState.ExpandAll()
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("alt+x"))):
		// Collapse all sessions - alt+x instead of C (td-2467e8)
		if p.contentSearchState != nil {
			p.contentSearchState.CollapseAll()
			p.contentSearchState.ClampCursor()
		}
		return p, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("backspace"))):
		// Delete character from query
		if p.contentSearchState != nil && len(p.contentSearchState.Query) > 0 {
			runes := []rune(p.contentSearchState.Query)
			p.contentSearchState.Query = string(runes[:len(runes)-1])
			return p, p.triggerContentSearch()
		}
		return p, nil

	case msg.Type == tea.KeyRunes:
		// Add character to query
		if p.contentSearchState != nil {
			p.contentSearchState.Query += string(msg.Runes)
			return p, p.triggerContentSearch()
		}
		return p, nil
	}

	return p, nil
}

// openContentSearch opens the content search modal.
func (p *Plugin) openContentSearch() (plugin.Plugin, tea.Cmd) {
	p.contentSearchMode = true
	p.contentSearchState = NewContentSearchState()
	p.hitRegionsDirty = true
	return p, nil
}

// minQueryLength is the minimum characters required before search triggers (td-5dcadc)
const minQueryLength = 2

// triggerContentSearch initiates a debounced search.
func (p *Plugin) triggerContentSearch() tea.Cmd {
	if p.contentSearchState == nil {
		return nil
	}

	// Require minimum query length (td-5dcadc)
	queryRunes := []rune(p.contentSearchState.Query)
	if len(queryRunes) < minQueryLength {
		// Clear results when query is too short
		p.contentSearchState.Results = nil
		p.contentSearchState.IsSearching = false
		p.contentSearchState.Cursor = 0
		p.contentSearchState.ScrollOffset = 0
		p.contentSearchState.Skeleton.Stop() // Stop skeleton animation (td-e740e4)
		return nil
	}

	p.contentSearchState.DebounceVersion++
	p.contentSearchState.IsSearching = true
	p.contentSearchState.Error = ""

	// Start skeleton animation for search in progress (td-e740e4)
	return tea.Batch(
		scheduleContentSearch(p.contentSearchState.Query, p.contentSearchState.DebounceVersion),
		p.contentSearchState.Skeleton.Start(),
	)
}

// contentSearchViewportHeight returns the viewport height for the content search results.
func (p *Plugin) contentSearchViewportHeight() int {
	// Modal height minus header, options, stats sections
	return p.height - 14
}

// jumpToSearchResult selects the session and message from the current search result.
func (p *Plugin) jumpToSearchResult() tea.Cmd {
	if p.contentSearchState == nil {
		return nil
	}

	session, msgMatch, _ := p.contentSearchState.GetSelectedResult()
	if session == nil {
		return nil
	}

	// Close search modal
	p.contentSearchMode = false
	p.hitRegionsDirty = true

	// Find the session in our list and select it
	found := false
	for i := range p.sessions {
		if p.sessions[i].ID == session.ID {
			p.cursor = i
			p.ensureCursorVisible()
			found = true
			break
		}
	}

	if !found {
		// Session not in current list (shouldn't happen, but handle gracefully)
		p.contentSearchState = nil
		return func() tea.Msg {
			return app.ToastMsg{Message: "Session not found", Duration: 2 * time.Second, IsError: true}
		}
	}

	// Set selected session and switch to messages pane
	p.setSelectedSession(session.ID)
	p.activePane = PaneMessages
	p.contentSearchState = nil

	// Set pending scroll target if we have a specific message to jump to (td-b74d9f)
	// This will be processed after messages are loaded in MessagesLoadedMsg handler
	// Use message ID (not index) to handle pagination correctly
	if msgMatch != nil && msgMatch.MessageID != "" {
		p.pendingScrollMsgID = msgMatch.MessageID
		p.pendingScrollActive = true
	} else {
		p.pendingScrollMsgID = ""
		p.pendingScrollActive = false
	}

	// Build commands to load messages
	return tea.Batch(
		p.loadMessages(session.ID),
		p.loadUsage(session.ID),
	)
}
