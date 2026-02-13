// Package conversations provides integration tests for content search feature (td-6ac70a).
package conversations

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcus/sidecar/internal/adapter"
	"github.com/marcus/sidecar/internal/plugin"
)

// MockSearchAdapter implements adapter.MessageSearcher for testing.
type MockSearchAdapter struct {
	adapter.Adapter
	searchResults map[string][]adapter.MessageMatch
}

func (m *MockSearchAdapter) Detect(workDir string) (bool, error) {
	return true, nil
}

func (m *MockSearchAdapter) SearchMessages(sessionID, query string, opts adapter.SearchOptions) ([]adapter.MessageMatch, error) {
	if results, ok := m.searchResults[sessionID]; ok {
		return results, nil
	}
	return nil, nil
}

func TestOpenContentSearch(t *testing.T) {
	p := New()

	// Initially not in content search mode
	if p.contentSearchMode {
		t.Error("Plugin should not start in content search mode")
	}
	if p.contentSearchState != nil {
		t.Error("Plugin should not have content search state initially")
	}

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	if !p.contentSearchMode {
		t.Error("Plugin should be in content search mode after openContentSearch")
	}
	if p.contentSearchState == nil {
		t.Error("Plugin should have content search state after openContentSearch")
	}
	if p.contentSearchState.Query != "" {
		t.Error("Content search query should be empty initially")
	}
}

func TestContentSearchKeyHandling(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Test typing characters
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	p = result.(*Plugin)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	p = result.(*Plugin)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	p = result.(*Plugin)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	p = result.(*Plugin)

	if p.contentSearchState.Query != "test" {
		t.Errorf("Query should be 'test', got %q", p.contentSearchState.Query)
	}

	// Test backspace
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyBackspace})
	p = result.(*Plugin)
	if p.contentSearchState.Query != "tes" {
		t.Errorf("Query after backspace should be 'tes', got %q", p.contentSearchState.Query)
	}

	// Test escape to close
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyEscape})
	p = result.(*Plugin)
	if p.contentSearchMode {
		t.Error("Content search mode should be closed after escape")
	}
	if p.contentSearchState != nil {
		t.Error("Content search state should be nil after escape")
	}
}

func TestContentSearchCursorNavigation(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search and set up test results
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Add test results
	p.contentSearchState.Results = []SessionSearchResult{
		{
			Session: adapter.Session{ID: "session1", Name: "Session 1"},
			Messages: []adapter.MessageMatch{
				{
					MessageID: "msg1",
					Matches: []adapter.ContentMatch{
						{LineNo: 1, LineText: "test"},
						{LineNo: 5, LineText: "test again"},
					},
				},
			},
		},
	}

	// Initial cursor position
	if p.contentSearchState.Cursor != 0 {
		t.Errorf("Initial cursor should be 0, got %d", p.contentSearchState.Cursor)
	}

	// Move down
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyDown})
	p = result.(*Plugin)
	if p.contentSearchState.Cursor != 1 {
		t.Errorf("Cursor after down should be 1, got %d", p.contentSearchState.Cursor)
	}

	// Move up
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyUp})
	p = result.(*Plugin)
	if p.contentSearchState.Cursor != 0 {
		t.Errorf("Cursor after up should be 0, got %d", p.contentSearchState.Cursor)
	}
}

func TestContentSearchToggleOptions(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Initially regex and case sensitivity are off
	if p.contentSearchState.UseRegex {
		t.Error("Regex should be off initially")
	}
	if p.contentSearchState.CaseSensitive {
		t.Error("Case sensitivity should be off initially")
	}

	// Toggle regex with ctrl+r
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyCtrlR})
	p = result.(*Plugin)
	if !p.contentSearchState.UseRegex {
		t.Error("Regex should be on after ctrl+r")
	}

	// Toggle case sensitivity with alt+c
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}, Alt: true})
	p = result.(*Plugin)
	if !p.contentSearchState.CaseSensitive {
		t.Error("Case sensitivity should be on after alt+c")
	}
}

func TestContentSearchToggleCollapse(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search and set up test results
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Add test results
	p.contentSearchState.Results = []SessionSearchResult{
		{
			Session:   adapter.Session{ID: "session1", Name: "Session 1"},
			Collapsed: false,
			Messages: []adapter.MessageMatch{
				{MessageID: "msg1", Matches: []adapter.ContentMatch{{LineNo: 1}}},
			},
		},
	}

	// Cursor is on session, toggle collapse with tab (td-2467e8: changed from space)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyTab})
	p = result.(*Plugin)
	if !p.contentSearchState.Results[0].Collapsed {
		t.Error("Session should be collapsed after tab on session row")
	}

	// Toggle again to expand
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyTab})
	p = result.(*Plugin)
	if p.contentSearchState.Results[0].Collapsed {
		t.Error("Session should be expanded after second tab")
	}
}

func TestContentSearchFocusContext(t *testing.T) {
	p := New()

	// Not in content search mode
	if p.FocusContext() == "conversations-content-search" {
		t.Error("FocusContext should not be content-search when not in that mode")
	}

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	if p.FocusContext() != "conversations-content-search" {
		t.Errorf("FocusContext should be 'conversations-content-search', got %q", p.FocusContext())
	}
}

func TestContentSearchCommands(t *testing.T) {
	p := New()

	// Not in content search mode - should not have content search commands
	cmds := p.Commands()
	found := false
	for _, cmd := range cmds {
		if cmd.Context == "conversations-content-search" {
			found = true
			break
		}
	}
	if found {
		t.Error("Should not have content search commands when not in that mode")
	}

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	cmds = p.Commands()
	found = false
	for _, cmd := range cmds {
		if cmd.Context == "conversations-content-search" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have content search commands when in content search mode")
	}
}

func TestContentSearchUpdateDebounceMsg(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Set up state for debounce
	p.contentSearchState.Query = "test"
	p.contentSearchState.DebounceVersion = 5

	// Valid debounce message
	result2, cmd := p.Update(ContentSearchDebounceMsg{Version: 5, Query: "test"})
	p = result2.(*Plugin)
	if cmd == nil {
		t.Error("Should return command for valid debounce version")
	}

	// Invalid debounce message (wrong version)
	result2, cmd = p.Update(ContentSearchDebounceMsg{Version: 3, Query: "test"})
	_ = result2.(*Plugin)
	if cmd != nil {
		t.Error("Should not return command for invalid debounce version")
	}
}

func TestContentSearchUpdateResultsMsg(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)
	p.contentSearchState.IsSearching = true

	// Send results
	results := []SessionSearchResult{
		{Session: adapter.Session{ID: "s1", Name: "Session 1"}},
	}
	result2, _ := p.Update(ContentSearchResultsMsg{Results: results, Error: nil})
	p = result2.(*Plugin)

	if p.contentSearchState.IsSearching {
		t.Error("IsSearching should be false after receiving results")
	}
	if len(p.contentSearchState.Results) != 1 {
		t.Errorf("Should have 1 result, got %d", len(p.contentSearchState.Results))
	}
	if p.contentSearchState.Error != "" {
		t.Errorf("Error should be empty, got %q", p.contentSearchState.Error)
	}
}

func TestContentSearchJumpToResult(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Set up sessions
	p.sessions = []adapter.Session{
		{ID: "session1", Name: "Session 1"},
		{ID: "session2", Name: "Session 2"},
	}
	p.cursor = 0
	p.selectedSession = ""

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Add results pointing to session2
	p.contentSearchState.Results = []SessionSearchResult{
		{
			Session: adapter.Session{ID: "session2", Name: "Session 2"},
			Messages: []adapter.MessageMatch{
				{MessageID: "msg1", MessageIdx: 5, Matches: []adapter.ContentMatch{{LineNo: 1}}},
			},
		},
	}

	// Position cursor on a match
	p.contentSearchState.Cursor = 2 // match row

	// Jump to result
	cmd := p.jumpToSearchResult()

	// Should close search mode
	if p.contentSearchMode {
		t.Error("Content search mode should be closed after jump")
	}

	// Should select correct session
	if p.selectedSession != "session2" {
		t.Errorf("Selected session should be 'session2', got %q", p.selectedSession)
	}

	// Should switch to messages pane
	if p.activePane != PaneMessages {
		t.Error("Should switch to messages pane after jump")
	}

	// Command should not be nil (has load messages)
	if cmd == nil {
		t.Error("Jump should return a command to load messages")
	}
}

func TestContentSearchView(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Render view
	output := p.View(100, 40)
	if output == "" {
		t.Error("View should produce non-empty output in content search mode")
	}

	// Add some results and render again
	p.contentSearchState.Query = "test"
	p.contentSearchState.Results = []SessionSearchResult{
		{
			Session:   adapter.Session{ID: "s1", Name: "Test Session", UpdatedAt: time.Now()},
			Collapsed: false,
			Messages: []adapter.MessageMatch{
				{MessageID: "msg1", Role: "user", Timestamp: time.Now(), Matches: []adapter.ContentMatch{{LineNo: 1, LineText: "test match"}}},
			},
		},
	}

	output = p.View(100, 40)
	if output == "" {
		t.Error("View should produce non-empty output with results")
	}
}

func TestTriggerContentSearch(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Set a query with at least 2 characters (td-5dcadc: minimum query length)
	p.contentSearchState.Query = "te"

	initialVersion := p.contentSearchState.DebounceVersion

	// Trigger search
	cmd := p.triggerContentSearch()

	// Version should be incremented
	if p.contentSearchState.DebounceVersion != initialVersion+1 {
		t.Errorf("DebounceVersion should be %d, got %d", initialVersion+1, p.contentSearchState.DebounceVersion)
	}

	// IsSearching should be true
	if !p.contentSearchState.IsSearching {
		t.Error("IsSearching should be true after trigger")
	}

	// Should return a command
	if cmd == nil {
		t.Error("triggerContentSearch should return a command")
	}
}

func TestTriggerContentSearch_MinQueryLength(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Query with only 1 character should not trigger search (td-5dcadc)
	p.contentSearchState.Query = "t"
	p.contentSearchState.IsSearching = true // Set to true to verify it gets cleared

	cmd := p.triggerContentSearch()

	if cmd != nil {
		t.Error("triggerContentSearch should return nil for query < 2 chars")
	}
	if p.contentSearchState.IsSearching {
		t.Error("IsSearching should be false when query is too short")
	}

	// Empty query should also not trigger
	p.contentSearchState.Query = ""
	p.contentSearchState.IsSearching = true
	cmd = p.triggerContentSearch()

	if cmd != nil {
		t.Error("triggerContentSearch should return nil for empty query")
	}
}

func TestContentSearchNextPrevMatch(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Add results with multiple matches
	p.contentSearchState.Results = []SessionSearchResult{
		{
			Session:   adapter.Session{ID: "s1"},
			Collapsed: false,
			Messages: []adapter.MessageMatch{
				{MessageID: "msg1", Matches: []adapter.ContentMatch{{LineNo: 1}, {LineNo: 2}}},
			},
		},
	}

	// Start at session row (index 0)
	p.contentSearchState.Cursor = 0

	// Jump to next match (ctrl+n - td-2467e8)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	p = result.(*Plugin)

	// Should be at first match (index 2)
	if p.contentSearchState.Cursor != 2 {
		t.Errorf("Cursor should be at first match (2), got %d", p.contentSearchState.Cursor)
	}

	// Jump to next match again
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	p = result.(*Plugin)

	// Should be at second match (index 3)
	if p.contentSearchState.Cursor != 3 {
		t.Errorf("Cursor should be at second match (3), got %d", p.contentSearchState.Cursor)
	}

	// Jump to previous match (ctrl+p - td-2467e8)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyCtrlP})
	p = result.(*Plugin)

	// Should be back at first match
	if p.contentSearchState.Cursor != 2 {
		t.Errorf("Cursor should be at first match (2), got %d", p.contentSearchState.Cursor)
	}
}

func TestFKeyOpensContentSearch(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Initialize plugin with sessions
	p.sessions = []adapter.Session{
		{ID: "session1", Name: "Session 1"},
	}
	p.activePane = PaneSidebar

	// Test from sidebar
	result, _ := p.updateSessions(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})
	p = result.(*Plugin)

	if !p.contentSearchMode {
		t.Error("Shift+F should open content search from sidebar")
	}

	// Close and test from messages pane
	p.contentSearchMode = false
	p.contentSearchState = nil
	p.activePane = PaneMessages
	p.selectedSession = "session1"

	result, _ = p.updateMessages(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})
	p = result.(*Plugin)

	if !p.contentSearchMode {
		t.Error("Shift+F should open content search from messages pane")
	}
}

func TestContentSearchExpandCollapseAll(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Add multiple sessions
	p.contentSearchState.Results = []SessionSearchResult{
		{Session: adapter.Session{ID: "s1"}, Collapsed: false},
		{Session: adapter.Session{ID: "s2"}, Collapsed: false},
	}

	// Collapse all (C key)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}})
	p = result.(*Plugin)

	for i, sr := range p.contentSearchState.Results {
		if !sr.Collapsed {
			t.Errorf("Session %d should be collapsed after C", i)
		}
	}

	// Expand all (E key)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'E'}})
	p = result.(*Plugin)

	for i, sr := range p.contentSearchState.Results {
		if sr.Collapsed {
			t.Errorf("Session %d should be expanded after E", i)
		}
	}
}

func TestContentSearchGotoTopBottom(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40

	// Open content search
	result, _ := p.openContentSearch()
	p = result.(*Plugin)

	// Add results
	p.contentSearchState.Results = []SessionSearchResult{
		{
			Session:   adapter.Session{ID: "s1"},
			Collapsed: false,
			Messages: []adapter.MessageMatch{
				{MessageID: "msg1", Matches: []adapter.ContentMatch{{LineNo: 1}, {LineNo: 2}}},
			},
		},
	}

	// Move to middle
	p.contentSearchState.Cursor = 2

	// Go to top (home key - td-2467e8)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyHome})
	p = result.(*Plugin)

	if p.contentSearchState.Cursor != 0 {
		t.Errorf("Cursor should be at top (0), got %d", p.contentSearchState.Cursor)
	}

	// Go to bottom (end key - td-2467e8)
	result, _ = p.handleContentSearchKey(tea.KeyMsg{Type: tea.KeyEnd})
	p = result.(*Plugin)

	flatLen := p.contentSearchState.FlatLen()
	if p.contentSearchState.Cursor != flatLen-1 {
		t.Errorf("Cursor should be at bottom (%d), got %d", flatLen-1, p.contentSearchState.Cursor)
	}
}

// Ensure pending scroll is processed on message load (td-b74d9f)
func TestPendingScrollOnMessageLoad(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40
	p.selectedSession = "test-session"

	// Set up pending scroll state (as if from content search jump)
	// Uses message ID for pagination-safe navigation
	p.pendingScrollMsgID = "msg2"
	p.pendingScrollActive = true

	// Simulate MessagesLoadedMsg
	result, _ := p.Update(MessagesLoadedMsg{
		SessionID: "test-session",
		Messages: []adapter.Message{
			{ID: "msg1", Role: "user", Content: "First message"},
			{ID: "msg2", Role: "assistant", Content: "Second message"},
			{ID: "msg3", Role: "user", Content: "Third message"},
		},
		TotalCount: 3,
		Offset:     0,
	})
	p = result.(*Plugin)

	// Pending scroll should be cleared
	if p.pendingScrollActive {
		t.Error("Pending scroll should be cleared after processing")
	}
	if p.pendingScrollMsgID != "" {
		t.Errorf("Pending scroll message ID should be empty, got %s", p.pendingScrollMsgID)
	}

	// messageCursor should be updated to find the target
	// Note: exact behavior depends on visibleMessageIndices
	if p.messageCursor < 0 {
		t.Error("Message cursor should be non-negative")
	}
}

// TestPendingScrollWithPagination tests that scroll works correctly when messages
// are paginated (offset > 0). Uses message ID for reliable navigation (td-b74d9f).
func TestPendingScrollWithPagination(t *testing.T) {
	p := New()
	p.width = 100
	p.height = 40
	p.selectedSession = "test-session"

	// Simulate paginated scenario: messages 50-52 loaded from a larger set
	// The target message "msg51" is at index 1 in our loaded slice
	p.pendingScrollMsgID = "msg51"
	p.pendingScrollActive = true

	// Simulate MessagesLoadedMsg with pagination
	result, _ := p.Update(MessagesLoadedMsg{
		SessionID: "test-session",
		Messages: []adapter.Message{
			{ID: "msg50", Role: "user", Content: "Message at offset 50"},
			{ID: "msg51", Role: "assistant", Content: "Target message at offset 51"},
			{ID: "msg52", Role: "user", Content: "Message at offset 52"},
		},
		TotalCount: 100, // Total messages in session
		Offset:     50,  // Starting from message 50
	})
	p = result.(*Plugin)

	// Pending scroll should be cleared
	if p.pendingScrollActive {
		t.Error("Pending scroll should be cleared after processing")
	}
	if p.pendingScrollMsgID != "" {
		t.Error("Pending scroll message ID should be cleared")
	}

	// Cursor should be positioned correctly in the loaded message array
	// The target "msg51" is at index 1 in the loaded messages
	if p.messageCursor < 0 {
		t.Error("Message cursor should be non-negative")
	}
}

// Helper to ensure the interface is satisfied
var _ plugin.Plugin = (*Plugin)(nil)
