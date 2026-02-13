package ui

import (
	"strings"
	"testing"
)

func TestNewSkeleton(t *testing.T) {
	s := NewSkeleton(5, nil)

	if s.Rows != 5 {
		t.Errorf("expected 5 rows, got %d", s.Rows)
	}
	if len(s.RowWidths) == 0 {
		t.Error("expected default row widths")
	}
	if !s.IsActive() {
		t.Error("expected skeleton to be active by default")
	}
}

func TestNewSkeletonCustomWidths(t *testing.T) {
	widths := []int{100, 50, 75}
	s := NewSkeleton(3, widths)

	if len(s.RowWidths) != 3 {
		t.Errorf("expected 3 widths, got %d", len(s.RowWidths))
	}
	if s.RowWidths[0] != 100 {
		t.Errorf("expected first width 100, got %d", s.RowWidths[0])
	}
}

func TestSkeletonView(t *testing.T) {
	s := NewSkeleton(3, []int{100, 50, 75})
	view := s.View(20)

	lines := strings.Split(view, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Check that view contains skeleton characters
	if !strings.ContainsAny(view, "░▒") {
		t.Error("expected skeleton characters in view")
	}
}

func TestSkeletonViewMinWidth(t *testing.T) {
	s := NewSkeleton(2, []int{100})
	view := s.View(5) // Less than minimum

	// Should still render something
	if len(view) == 0 {
		t.Error("expected non-empty view")
	}
}

func TestSkeletonStartStop(t *testing.T) {
	s := NewSkeleton(3, nil)

	if !s.IsActive() {
		t.Error("expected active after creation")
	}

	s.Stop()
	if s.IsActive() {
		t.Error("expected inactive after Stop()")
	}

	s.Start()
	if !s.IsActive() {
		t.Error("expected active after Start()")
	}
}

func TestSkeletonUpdate(t *testing.T) {
	s := NewSkeleton(3, nil)
	initialFrame := s.frame

	// Update with tick message should advance frame
	s.Update(SkeletonTickMsg{})

	if s.frame != initialFrame+1 {
		t.Errorf("expected frame %d, got %d", initialFrame+1, s.frame)
	}
}

func TestSkeletonUpdateWhenStopped(t *testing.T) {
	s := NewSkeleton(3, nil)
	s.Stop()

	initialFrame := s.frame
	cmd := s.Update(SkeletonTickMsg{})

	// Frame should not advance when stopped
	if s.frame != initialFrame {
		t.Error("frame should not advance when stopped")
	}
	// Should not return a new tick command
	if cmd != nil {
		t.Error("should not return command when stopped")
	}
}

func TestSkeletonShimmerAnimation(t *testing.T) {
	s := NewSkeleton(2, []int{100})

	view1 := s.View(30)
	s.Update(SkeletonTickMsg{})
	view2 := s.View(30)

	// Views should be different due to shimmer movement
	if view1 == view2 {
		t.Error("expected different views after tick (shimmer should move)")
	}
}
