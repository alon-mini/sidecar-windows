package ui

import (
	"strings"
	"testing"

	"github.com/marcus/sidecar/internal/styles"
)

func TestResolveButtonStyle_FocusMatch(t *testing.T) {
	style := ResolveButtonStyle(1, -1, 1)
	if style.GetBold() != styles.ButtonFocused.GetBold() {
		t.Error("expected ButtonFocused when focusIdx matches btnIdx")
	}
}

func TestResolveButtonStyle_HoverMatch(t *testing.T) {
	style := ResolveButtonStyle(-1, 2, 2)
	// ButtonHover should be returned when no focus but hover matches
	if style.GetBackground() != styles.ButtonHover.GetBackground() {
		t.Error("expected ButtonHover when hoverIdx matches btnIdx (no focus)")
	}
}

func TestResolveButtonStyle_NoMatch(t *testing.T) {
	style := ResolveButtonStyle(-1, -1, 1)
	if style.GetBackground() != styles.Button.GetBackground() {
		t.Error("expected Button when neither matches")
	}
}

func TestResolveButtonStyle_FocusPrecedence(t *testing.T) {
	// Focus takes precedence over hover when both match
	style := ResolveButtonStyle(1, 1, 1)
	if style.GetBold() != styles.ButtonFocused.GetBold() {
		t.Error("focus should take precedence over hover")
	}
}

func TestRenderButtonPair_Output(t *testing.T) {
	result := RenderButtonPair("Confirm", "Cancel", 0, 0)

	// Should contain both labels
	if !strings.Contains(result, "Confirm") {
		t.Error("output should contain confirm label")
	}
	if !strings.Contains(result, "Cancel") {
		t.Error("output should contain cancel label")
	}
}

func TestRenderButtonPair_FocusConfirm(t *testing.T) {
	// When focus is on confirm (1), it should use focused style
	result := RenderButtonPair("OK", "No", 1, 0)
	if !strings.Contains(result, "OK") || !strings.Contains(result, "No") {
		t.Error("output should contain both labels")
	}
}

func TestRenderButtonPair_FocusCancel(t *testing.T) {
	// When focus is on cancel (2), it should use focused style
	result := RenderButtonPair("Yes", "No", 2, 0)
	if !strings.Contains(result, "Yes") || !strings.Contains(result, "No") {
		t.Error("output should contain both labels")
	}
}
