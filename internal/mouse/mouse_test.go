package mouse

import (
	"testing"
)

func TestRect_Contains(t *testing.T) {
	r := Rect{X: 10, Y: 20, W: 5, H: 3}

	tests := []struct {
		name string
		x, y int
		want bool
	}{
		{"inside", 12, 21, true},
		{"top-left corner", 10, 20, true},
		{"bottom-right edge excluded", 15, 23, false},
		{"just inside right", 14, 22, true},
		{"left of rect", 9, 21, false},
		{"above rect", 12, 19, false},
		{"below rect", 12, 23, false},
		{"right of rect", 15, 21, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Contains(tt.x, tt.y); got != tt.want {
				t.Errorf("Rect(%+v).Contains(%d, %d) = %v, want %v", r, tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestRect_Contains_ZeroSize(t *testing.T) {
	r := Rect{X: 5, Y: 5, W: 0, H: 0}
	if r.Contains(5, 5) {
		t.Error("zero-size rect should not contain any point")
	}
}

func TestHitMap_AddAndTest(t *testing.T) {
	hm := NewHitMap()
	hm.Add("btn1", Rect{X: 0, Y: 0, W: 10, H: 5}, "button1")

	region := hm.Test(5, 3)
	if region == nil {
		t.Fatal("expected to hit btn1")
	}
	if region.ID != "btn1" {
		t.Errorf("region.ID = %q, want %q", region.ID, "btn1")
	}
	if region.Data != "button1" {
		t.Errorf("region.Data = %v, want %q", region.Data, "button1")
	}
}

func TestHitMap_OverlappingRegions(t *testing.T) {
	hm := NewHitMap()
	hm.Add("bottom", Rect{X: 0, Y: 0, W: 20, H: 20}, nil)
	hm.Add("top", Rect{X: 5, Y: 5, W: 10, H: 10}, nil)

	region := hm.Test(7, 7)
	if region == nil {
		t.Fatal("expected to hit a region")
	}
	if region.ID != "top" {
		t.Errorf("overlapping region: got %q, want %q (later region should win)", region.ID, "top")
	}
}

func TestHitMap_Clear(t *testing.T) {
	hm := NewHitMap()
	hm.Add("a", Rect{X: 0, Y: 0, W: 10, H: 10}, nil)
	hm.Clear()

	if region := hm.Test(5, 5); region != nil {
		t.Error("expected nil after Clear()")
	}
}

func TestHitMap_AddRect(t *testing.T) {
	hm := NewHitMap()
	hm.AddRect("item", 10, 20, 5, 3, 42)

	region := hm.Test(12, 21)
	if region == nil {
		t.Fatal("expected to hit item")
	}
	if region.ID != "item" {
		t.Errorf("ID = %q, want %q", region.ID, "item")
	}
	if region.Data != 42 {
		t.Errorf("Data = %v, want 42", region.Data)
	}
}

func TestHitMap_Regions(t *testing.T) {
	hm := NewHitMap()
	hm.Add("a", Rect{X: 0, Y: 0, W: 5, H: 5}, nil)
	hm.Add("b", Rect{X: 10, Y: 10, W: 5, H: 5}, nil)

	regions := hm.Regions()
	if len(regions) != 2 {
		t.Fatalf("len(Regions()) = %d, want 2", len(regions))
	}

	// Verify it's a copy
	regions[0].ID = "modified"
	original := hm.Regions()
	if original[0].ID == "modified" {
		t.Error("Regions() should return a copy")
	}
}

func TestHitMap_TestMiss(t *testing.T) {
	hm := NewHitMap()
	hm.Add("a", Rect{X: 100, Y: 100, W: 5, H: 5}, nil)

	if region := hm.Test(0, 0); region != nil {
		t.Error("expected nil for miss")
	}
}

func TestHandler_HandleClick(t *testing.T) {
	h := NewHandler()
	h.HitMap.Add("btn", Rect{X: 0, Y: 0, W: 10, H: 10}, nil)

	result := h.HandleClick(5, 5)
	if result.Region == nil {
		t.Fatal("expected to hit btn")
	}
	if result.Region.ID != "btn" {
		t.Errorf("region ID = %q, want %q", result.Region.ID, "btn")
	}
	if result.IsDoubleClick {
		t.Error("first click should not be double click")
	}
}

func TestHandler_HandleClick_Miss(t *testing.T) {
	h := NewHandler()
	h.HitMap.Add("btn", Rect{X: 100, Y: 100, W: 5, H: 5}, nil)

	result := h.HandleClick(0, 0)
	if result.Region != nil {
		t.Error("expected nil region for miss")
	}
}

func TestHandler_DoubleClick(t *testing.T) {
	h := NewHandler()
	h.HitMap.Add("btn", Rect{X: 0, Y: 0, W: 10, H: 10}, nil)

	// First click
	r1 := h.HandleClick(5, 5)
	if r1.IsDoubleClick {
		t.Error("first click should not be double click")
	}

	// Second click immediately — within 400ms
	r2 := h.HandleClick(5, 5)
	if !r2.IsDoubleClick {
		t.Error("second immediate click should be double click")
	}

	// Third click — should NOT be double click (reset after double)
	r3 := h.HandleClick(5, 5)
	if r3.IsDoubleClick {
		t.Error("third click should not be double click (reset after double)")
	}
}

func TestHandler_DragLifecycle(t *testing.T) {
	h := NewHandler()

	if h.IsDragging() {
		t.Error("should not be dragging initially")
	}

	h.StartDrag(10, 20, "sidebar", 200)

	if !h.IsDragging() {
		t.Error("should be dragging after StartDrag")
	}
	if h.DragRegion() != "sidebar" {
		t.Errorf("DragRegion = %q, want %q", h.DragRegion(), "sidebar")
	}
	if h.DragStartValue() != 200 {
		t.Errorf("DragStartValue = %d, want 200", h.DragStartValue())
	}

	dx, dy := h.DragDelta(15, 25)
	if dx != 5 || dy != 5 {
		t.Errorf("DragDelta = (%d, %d), want (5, 5)", dx, dy)
	}

	h.EndDrag()

	if h.IsDragging() {
		t.Error("should not be dragging after EndDrag")
	}
	if h.DragRegion() != "" {
		t.Errorf("DragRegion after end = %q, want empty", h.DragRegion())
	}
}

func TestHandler_Clear(t *testing.T) {
	h := NewHandler()
	h.HitMap.Add("a", Rect{X: 0, Y: 0, W: 10, H: 10}, nil)
	h.Clear()

	if region := h.HitMap.Test(5, 5); region != nil {
		t.Error("expected nil after Clear()")
	}
}
