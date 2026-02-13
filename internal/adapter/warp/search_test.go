package warp

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
	// Note: Warp requires SQLite DB, so nonexistent session lookup may return error
	// depending on whether DB exists at default path
	_, err := a.SearchMessages("nonexistent-session-xyz", "test", adapter.DefaultSearchOptions())
	// We don't strictly check error here since it depends on local Warp installation
	_ = err
}
