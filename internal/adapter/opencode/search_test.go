package opencode

import (
	"testing"

	"github.com/marcus/sidecar/internal/adapter"
)

func TestSearchMessages_InterfaceCompliance(t *testing.T) {
	a := New()
	// Verify interface compliance at compile time
	var _ adapter.MessageSearcher = a
}

func TestSearchMessages_NonExistentSession(t *testing.T) {
	a := New()
	results, err := a.SearchMessages("nonexistent-session-xyz", "test", adapter.DefaultSearchOptions())
	if err != nil {
		t.Fatalf("expected no error for nonexistent session, got %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results, got %v", results)
	}
}
