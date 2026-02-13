package notes

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
)

func TestNotesLoadedSyncsEditorAfterOutOfBandSave(t *testing.T) {
	p := New()
	p.height = 24
	p.editorTextarea = textarea.New()
	p.editorNote = &Note{ID: "nt-1", Content: "before"}
	p.editorTextarea.SetValue("before")
	p.previewLines = []string{"before"}
	p.pendingEditorSyncID = "nt-1"

	_, _ = p.Update(NotesLoadedMsg{
		Notes: []Note{
			{ID: "nt-1", Content: "after"},
		},
	})

	if got := p.editorTextarea.Value(); got != "after" {
		t.Fatalf("expected editor content to sync, got %q", got)
	}
	if len(p.previewLines) != 1 || p.previewLines[0] != "after" {
		t.Fatalf("expected preview lines to sync, got %#v", p.previewLines)
	}
	if p.pendingEditorSyncID != "" {
		t.Fatalf("expected pending sync marker to be cleared, got %q", p.pendingEditorSyncID)
	}
}

func TestNotesLoadedDoesNotSyncEditorWithoutPendingMarker(t *testing.T) {
	p := New()
	p.height = 24
	p.editorTextarea = textarea.New()
	p.editorNote = &Note{ID: "nt-1", Content: "before"}
	p.editorTextarea.SetValue("local edit buffer")
	p.previewLines = []string{"local edit buffer"}

	_, _ = p.Update(NotesLoadedMsg{
		Notes: []Note{
			{ID: "nt-1", Content: "after"},
		},
	})

	if got := p.editorTextarea.Value(); got != "local edit buffer" {
		t.Fatalf("expected editor content to remain unchanged, got %q", got)
	}
	if p.pendingEditorSyncID != "" {
		t.Fatalf("expected no pending sync marker, got %q", p.pendingEditorSyncID)
	}
}
