package gitstatus

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcus/sidecar/internal/mouse"
	"github.com/marcus/sidecar/internal/plugin"
)

func TestDiscardModalKeyboardFlow(t *testing.T) {
	// Create a plugin with a mock discard file
	p := &Plugin{
		viewMode:     ViewModeStatus,
		width:        80,
		height:       24,
		mouseHandler: mouse.NewHandler(),
	}

	// Create a mock file entry
	entry := &FileEntry{
		Path:   "test/file.go",
		Status: StatusModified,
		Staged: false,
	}

	// Simulate pressing D to open the discard modal
	p.discardFile = entry
	p.discardReturnMode = p.viewMode
	p.viewMode = ViewModeConfirmDiscard
	p.buildDiscardModal()

	if p.discardModal == nil {
		t.Fatal("expected discardModal to be created")
	}

	// Simulate rendering (which populates focusIDs)
	p.discardModal.Render(p.width, p.height, p.mouseHandler)

	// Check that focus is on "discard" button
	if p.discardModal.FocusedID() != "discard" {
		t.Errorf("expected focus on 'discard', got %q", p.discardModal.FocusedID())
	}

	// Simulate pressing Enter
	action, _ := p.discardModal.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	t.Logf("HandleKey returned action: %q", action)

	if action != "discard" {
		t.Errorf("expected action 'discard', got %q", action)
	}

	// Test Tab to switch to cancel, then Enter
	p.discardModal.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
	if p.discardModal.FocusedID() != "cancel" {
		t.Errorf("expected focus on 'cancel' after Tab, got %q", p.discardModal.FocusedID())
	}

	action, _ = p.discardModal.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != "cancel" {
		t.Errorf("expected action 'cancel', got %q", action)
	}
}

func TestDiscardModalMouseFlow(t *testing.T) {
	// Create a plugin with a mock discard file
	p := &Plugin{
		viewMode:     ViewModeStatus,
		width:        80,
		height:       24,
		mouseHandler: mouse.NewHandler(),
	}

	// Create a mock file entry
	entry := &FileEntry{
		Path:   "test/file.go",
		Status: StatusModified,
		Staged: false,
	}

	// Simulate opening the discard modal
	p.discardFile = entry
	p.discardReturnMode = p.viewMode
	p.viewMode = ViewModeConfirmDiscard
	p.buildDiscardModal()

	if p.discardModal == nil {
		t.Fatal("expected discardModal to be created")
	}

	// Render to populate hit regions
	p.discardModal.Render(p.width, p.height, p.mouseHandler)

	// Find the discard button region
	regions := p.mouseHandler.HitMap.Regions()
	var discardRegion *mouse.Region
	for i := range regions {
		if regions[i].ID == "discard" {
			discardRegion = &regions[i]
			break
		}
	}

	if discardRegion == nil {
		t.Fatal("expected 'discard' button region to be registered")
	}

	t.Logf("Discard region: x=%d, y=%d, w=%d, h=%d", discardRegion.Rect.X, discardRegion.Rect.Y, discardRegion.Rect.W, discardRegion.Rect.H)

	// Simulate clicking on the discard button
	clickX := discardRegion.Rect.X + discardRegion.Rect.W/2
	clickY := discardRegion.Rect.Y
	action := p.discardModal.HandleMouse(tea.MouseMsg{
		X:      clickX,
		Y:      clickY,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}, p.mouseHandler)

	if action != "discard" {
		t.Errorf("expected action 'discard' on click, got %q", action)
	}
}

// TestPluginUpdateConfirmDiscard tests the full flow through Plugin.Update
func TestPluginUpdateConfirmDiscard(t *testing.T) {
	// Create a plugin with a mock discard file
	p := &Plugin{
		viewMode:     ViewModeStatus,
		width:        80,
		height:       24,
		mouseHandler: mouse.NewHandler(),
	}

	// Create a mock file entry
	entry := &FileEntry{
		Path:   "test/file.go",
		Status: StatusModified,
		Staged: false,
	}

	// Simulate pressing D to open the discard modal
	p.discardFile = entry
	p.discardReturnMode = p.viewMode
	p.viewMode = ViewModeConfirmDiscard
	p.buildDiscardModal()

	// Simulate rendering (which populates focusIDs)
	p.discardModal.Render(p.width, p.height, p.mouseHandler)

	t.Logf("Before Update: viewMode=%d discardFile=%v focusedID=%q",
		p.viewMode, p.discardFile != nil, p.discardModal.FocusedID())

	// Now call Update with Enter key
	result, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedPlugin := result.(*Plugin)

	t.Logf("After Update: viewMode=%d discardFile=%v cmd=%v",
		updatedPlugin.viewMode, updatedPlugin.discardFile != nil, cmd != nil)

	// The viewMode should have returned to status
	if updatedPlugin.viewMode != ViewModeStatus {
		t.Errorf("expected viewMode to return to ViewModeStatus, got %d", updatedPlugin.viewMode)
	}

	// The discardFile should be nil (modal closed)
	if updatedPlugin.discardFile != nil {
		t.Errorf("expected discardFile to be nil after confirm")
	}

	// A command should have been returned (doDiscard)
	if cmd == nil {
		t.Errorf("expected a command to be returned (doDiscard)")
	}
}

// TestPluginUpdateConfirmDiscardY tests the Y shortcut
func TestPluginUpdateConfirmDiscardY(t *testing.T) {
	// Create a plugin with a mock discard file
	p := &Plugin{
		viewMode:     ViewModeStatus,
		width:        80,
		height:       24,
		mouseHandler: mouse.NewHandler(),
	}

	// Create a mock file entry
	entry := &FileEntry{
		Path:   "test/file.go",
		Status: StatusModified,
		Staged: false,
	}

	// Simulate pressing D to open the discard modal
	p.discardFile = entry
	p.discardReturnMode = p.viewMode
	p.viewMode = ViewModeConfirmDiscard
	p.buildDiscardModal()

	// Simulate rendering
	p.discardModal.Render(p.width, p.height, p.mouseHandler)

	// Now call Update with Y key
	result, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	updatedPlugin := result.(*Plugin)

	// The viewMode should have returned to status
	if updatedPlugin.viewMode != ViewModeStatus {
		t.Errorf("expected viewMode to return to ViewModeStatus, got %d", updatedPlugin.viewMode)
	}

	// A command should have been returned (doDiscard)
	if cmd == nil {
		t.Errorf("expected a command to be returned (doDiscard) for Y shortcut")
	}
}

// Ensure plugin implements the interface
var _ plugin.Plugin = (*Plugin)(nil)
