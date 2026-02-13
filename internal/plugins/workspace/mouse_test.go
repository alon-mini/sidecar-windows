package workspace

import (
	"testing"

	"github.com/marcus/sidecar/internal/mouse"
)

func TestIsModalViewMode(t *testing.T) {
	tests := []struct {
		mode    ViewMode
		isModal bool
	}{
		{ViewModeList, false},
		{ViewModeKanban, false},
		{ViewModeInteractive, false},
		{ViewModeCreate, true},
		{ViewModeTaskLink, true},
		{ViewModeMerge, true},
		{ViewModeAgentChoice, true},
		{ViewModeConfirmDelete, true},
		{ViewModeConfirmDeleteShell, true},
		{ViewModeCommitForMerge, true},
		{ViewModePromptPicker, true},
		{ViewModeTypeSelector, true},
		{ViewModeRenameShell, true},
		{ViewModeFilePicker, true},
	}

	for _, tt := range tests {
		p := &Plugin{viewMode: tt.mode}
		got := p.isModalViewMode()
		if got != tt.isModal {
			t.Errorf("isModalViewMode() for %d = %v, want %v", tt.mode, got, tt.isModal)
		}
	}
}

func TestIsBackgroundRegion(t *testing.T) {
	background := []string{
		regionSidebar, regionPreviewPane, regionPaneDivider,
		regionWorktreeItem, regionPreviewTab,
		regionCreateWorktreeButton, regionShellsPlusButton, regionWorkspacesPlusButton,
		regionKanbanCard, regionKanbanColumn, regionViewToggle,
	}
	for _, id := range background {
		if !isBackgroundRegion(id) {
			t.Errorf("isBackgroundRegion(%q) = false, want true", id)
		}
	}

	modal := []string{
		agentChoiceConfirmID, agentChoiceCancelID,
		deleteConfirmDeleteID, deleteConfirmCancelID,
		regionCreateBackdrop, regionCreateModalBody, regionCreateInput,
		mergeMethodListID, mergeCleanUpButtonID, mergeSkipButtonID, // Merge modal element IDs
		typeSelectorListID, typeSelectorConfirmID, typeSelectorCancelID, typeSelectorInputID, // Type selector modal element IDs
		regionPromptItem, regionPromptFilter,
	}
	for _, id := range modal {
		if isBackgroundRegion(id) {
			t.Errorf("isBackgroundRegion(%q) = true, want false", id)
		}
	}
}

func TestModalClickGuard(t *testing.T) {
	modalModes := []ViewMode{
		ViewModeCreate, ViewModeMerge, ViewModeAgentChoice,
		ViewModeConfirmDelete, ViewModeConfirmDeleteShell,
		ViewModeRenameShell, ViewModeTypeSelector,
		ViewModePromptPicker, ViewModeTaskLink,
		ViewModeCommitForMerge, ViewModeFilePicker,
	}
	backgroundRegions := []string{
		regionSidebar, regionPreviewPane, regionWorktreeItem,
		regionPaneDivider, regionPreviewTab,
	}

	for _, mode := range modalModes {
		for _, region := range backgroundRegions {
			p := &Plugin{
				viewMode:     mode,
				mouseHandler: mouse.NewHandler(),
			}
			action := mouse.MouseAction{
				Type:   mouse.ActionClick,
				Region: &mouse.Region{ID: region},
			}
			cmd := p.handleMouseClick(action)
			if cmd != nil {
				t.Errorf("handleMouseClick(mode=%d, region=%q) returned non-nil cmd", mode, region)
			}
		}
	}
}

func TestModalClickGuardAllowsModalRegions(t *testing.T) {
	// Merge modal now uses modal library which handles its own regions via handleMergeModalMouse
	// This test verifies that handleMouseClick returns nil for merge modal (since it's handled separately)
	p := &Plugin{
		viewMode:     ViewModeMerge,
		mouseHandler: mouse.NewHandler(),
	}
	action := mouse.MouseAction{
		Type:   mouse.ActionClick,
		Region: &mouse.Region{ID: mergeCleanUpButtonID, Data: nil},
	}
	// handleMouseClick should not process merge modal clicks (handled by handleMergeModalMouse)
	cmd := p.handleMouseClick(action)
	if cmd != nil {
		t.Errorf("merge modal click via handleMouseClick returned unexpected cmd")
	}
}

func TestModalDoubleClickGuard(t *testing.T) {
	p := &Plugin{
		viewMode:     ViewModeMerge,
		mouseHandler: mouse.NewHandler(),
	}
	action := mouse.MouseAction{
		Type:   mouse.ActionDoubleClick,
		Region: &mouse.Region{ID: regionPreviewPane},
	}
	cmd := p.handleMouseDoubleClick(action)
	if cmd != nil {
		t.Error("handleMouseDoubleClick should return nil when modal is open")
	}
}

func TestModalScrollGuard(t *testing.T) {
	p := &Plugin{
		viewMode:     ViewModeConfirmDelete,
		mouseHandler: mouse.NewHandler(),
	}
	action := mouse.MouseAction{
		Type:   mouse.ActionScrollDown,
		Region: &mouse.Region{ID: regionSidebar},
	}
	cmd := p.handleMouseScroll(action)
	if cmd != nil {
		t.Error("handleMouseScroll should return nil when modal is open with background region")
	}
}

func TestModalScrollGuardNilRegion(t *testing.T) {
	p := &Plugin{
		viewMode:     ViewModeConfirmDelete,
		mouseHandler: mouse.NewHandler(),
	}
	action := mouse.MouseAction{
		Type:   mouse.ActionScrollDown,
		Region: nil,
	}
	cmd := p.handleMouseScroll(action)
	if cmd != nil {
		t.Error("handleMouseScroll should return nil when modal is open with nil region")
	}
}

func TestModalDragGuard(t *testing.T) {
	p := &Plugin{
		viewMode:     ViewModeRenameShell,
		mouseHandler: mouse.NewHandler(),
		sidebarWidth: 40,
	}
	action := mouse.MouseAction{
		Type: mouse.ActionDrag,
		X:    50,
	}
	cmd := p.handleMouseDrag(action)
	if cmd != nil {
		t.Error("handleMouseDrag should return nil when modal is open")
	}
	if p.sidebarWidth != 40 {
		t.Errorf("sidebarWidth changed during modal drag: got %d, want 40", p.sidebarWidth)
	}
}

func TestModalHoverGuard(t *testing.T) {
	p := &Plugin{
		viewMode:     ViewModeTaskLink,
		mouseHandler: mouse.NewHandler(),
	}
	action := mouse.MouseAction{
		Type:   mouse.ActionHover,
		Region: &mouse.Region{ID: regionCreateWorktreeButton},
	}
	cmd := p.handleMouseHover(action)
	if cmd != nil {
		t.Error("handleMouseHover should return nil for background region when modal is open")
	}
	if p.hoverNewButton {
		t.Error("hoverNewButton should not be set when modal is open")
	}
}

func TestNonModalClickPassesThrough(t *testing.T) {
	p := &Plugin{
		viewMode:     ViewModeList,
		mouseHandler: mouse.NewHandler(),
		width:        100,
		sidebarWidth: 40,
	}
	action := mouse.MouseAction{
		Type:   mouse.ActionClick,
		Region: &mouse.Region{ID: regionSidebar},
	}
	_ = p.handleMouseClick(action)
	if p.activePane != PaneSidebar {
		t.Error("sidebar click in List mode should set activePane to PaneSidebar")
	}
}
