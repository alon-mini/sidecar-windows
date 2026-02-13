package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewConfirmDialog(t *testing.T) {
	d := NewConfirmDialog("Test Title", "Test message")

	if d.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", d.Title)
	}
	if d.Message != "Test message" {
		t.Errorf("expected message 'Test message', got %q", d.Message)
	}
	if d.ConfirmLabel != " Confirm " {
		t.Errorf("expected default confirm label ' Confirm ', got %q", d.ConfirmLabel)
	}
	if d.CancelLabel != " Cancel " {
		t.Errorf("expected default cancel label ' Cancel ', got %q", d.CancelLabel)
	}
	if d.Width != ModalWidthMedium {
		t.Errorf("expected width %d, got %d", ModalWidthMedium, d.Width)
	}
}

func TestConfirmDialog_ToModal(t *testing.T) {
	d := NewConfirmDialog("Delete File?", "Are you sure?")
	d.ConfirmLabel = " Delete "
	d.CancelLabel = " Cancel "

	modal := d.ToModal()
	output := modal.Render(80, 24, nil)

	if !strings.Contains(output, "Delete File?") {
		t.Error("render should contain title")
	}
	if !strings.Contains(output, "Are you sure?") {
		t.Error("render should contain message")
	}
	if !strings.Contains(output, "Delete") {
		t.Error("render should contain confirm label")
	}
	if !strings.Contains(output, "Cancel") {
		t.Error("render should contain cancel label")
	}
	if strings.Contains(output, "Tab to switch") {
		t.Error("render should not include modal hint line")
	}
}

func TestConfirmDialog_ToModalActions(t *testing.T) {
	d := NewConfirmDialog("Test", "Message")
	m := d.ToModal()
	m.Render(80, 24, nil)

	action, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != "confirm" {
		t.Errorf("expected confirm action, got %q", action)
	}

	m.SetFocus("cancel")
	action, _ = m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != "cancel" {
		t.Errorf("expected cancel action, got %q", action)
	}
}
