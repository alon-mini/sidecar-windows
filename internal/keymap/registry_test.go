package keymap

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRegistry_SingleKey(t *testing.T) {
	r := NewRegistry()

	called := false
	r.RegisterCommand(Command{
		ID:   "test-cmd",
		Name: "Test",
		Handler: func() tea.Cmd {
			called = true
			return nil
		},
	})
	r.RegisterBinding(Binding{Key: "t", Command: "test-cmd", Context: "global"})

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	r.Handle(key, "global")

	if !called {
		t.Error("command handler not called")
	}
}

func TestRegistry_KeySequence(t *testing.T) {
	r := NewRegistry()

	called := false
	r.RegisterCommand(Command{
		ID:   "go-top",
		Name: "Go to top",
		Handler: func() tea.Cmd {
			called = true
			return nil
		},
	})
	r.RegisterBinding(Binding{Key: "g g", Command: "go-top", Context: "global"})

	// First 'g' should start sequence
	key1 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	r.Handle(key1, "global")

	if called {
		t.Error("should not call handler after first key")
	}
	if !r.HasPending() {
		t.Error("should have pending key")
	}

	// Second 'g' should complete sequence
	key2 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	r.Handle(key2, "global")

	if !called {
		t.Error("command handler not called for sequence")
	}
}

func TestRegistry_KeySequenceTimeout(t *testing.T) {
	r := NewRegistry()

	called := false
	r.RegisterCommand(Command{
		ID:   "go-top",
		Name: "Go to top",
		Handler: func() tea.Cmd {
			called = true
			return nil
		},
	})
	r.RegisterBinding(Binding{Key: "g g", Command: "go-top", Context: "global"})

	// First 'g'
	key1 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	r.Handle(key1, "global")

	// Wait for timeout
	time.Sleep(sequenceTimeout + 10*time.Millisecond)

	// Second 'g' should not complete sequence due to timeout
	key2 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	r.Handle(key2, "global")

	if called {
		t.Error("sequence should have timed out")
	}
}

func TestRegistry_ContextPrecedence(t *testing.T) {
	r := NewRegistry()

	globalCalled := false
	contextCalled := false

	r.RegisterCommand(Command{
		ID:   "global-action",
		Name: "Global Action",
		Handler: func() tea.Cmd {
			globalCalled = true
			return nil
		},
	})
	r.RegisterCommand(Command{
		ID:   "context-action",
		Name: "Context Action",
		Handler: func() tea.Cmd {
			contextCalled = true
			return nil
		},
	})

	r.RegisterBinding(Binding{Key: "s", Command: "global-action", Context: "global"})
	r.RegisterBinding(Binding{Key: "s", Command: "context-action", Context: "git-status"})

	// With git-status context, should use context binding
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	r.Handle(key, "git-status")

	if globalCalled {
		t.Error("global handler should not be called")
	}
	if !contextCalled {
		t.Error("context handler should be called")
	}
}

func TestRegistry_UserOverride(t *testing.T) {
	r := NewRegistry()

	defaultCalled := false
	overrideCalled := false

	r.RegisterCommand(Command{
		ID:   "default-action",
		Name: "Default",
		Handler: func() tea.Cmd {
			defaultCalled = true
			return nil
		},
	})
	r.RegisterCommand(Command{
		ID:   "override-action",
		Name: "Override",
		Handler: func() tea.Cmd {
			overrideCalled = true
			return nil
		},
	})

	r.RegisterBinding(Binding{Key: "x", Command: "default-action", Context: "global"})
	r.SetUserOverride("x", "override-action")

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	r.Handle(key, "global")

	if defaultCalled {
		t.Error("default handler should not be called")
	}
	if !overrideCalled {
		t.Error("override handler should be called")
	}
}

func TestRegistry_SpecialKeys(t *testing.T) {
	cases := []struct {
		keyType tea.KeyType
		expect  string
	}{
		{tea.KeyTab, "tab"},
		{tea.KeyEnter, "enter"},
		{tea.KeyEsc, "esc"},
		{tea.KeyUp, "up"},
		{tea.KeyDown, "down"},
		{tea.KeyCtrlC, "ctrl+c"},
		{tea.KeyShiftTab, "shift+tab"},
	}

	for _, tc := range cases {
		key := tea.KeyMsg{Type: tc.keyType}
		got := keyToString(key)
		if got != tc.expect {
			t.Errorf("keyToString(%v) = %q, want %q", tc.keyType, got, tc.expect)
		}
	}
}

func TestRegistry_GetCommand(t *testing.T) {
	r := NewRegistry()

	// Register a command
	r.RegisterCommand(Command{
		ID:      "test-cmd",
		Name:    "Test Command",
		Handler: func() tea.Cmd { return nil },
		Context: "global",
	})

	// Test finding existing command
	cmd, ok := r.GetCommand("test-cmd")
	if !ok {
		t.Error("GetCommand should find registered command")
	}
	if cmd.ID != "test-cmd" {
		t.Errorf("GetCommand returned wrong ID: got %q, want %q", cmd.ID, "test-cmd")
	}
	if cmd.Name != "Test Command" {
		t.Errorf("GetCommand returned wrong Name: got %q, want %q", cmd.Name, "Test Command")
	}

	// Test missing command
	_, ok = r.GetCommand("nonexistent")
	if ok {
		t.Error("GetCommand should return false for missing command")
	}
}
