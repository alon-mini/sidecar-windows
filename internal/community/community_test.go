package community

import "testing"

func TestSchemeCount(t *testing.T) {
	if got := SchemeCount(); got != 453 {
		t.Errorf("SchemeCount() = %d, want 453", got)
	}
}

func TestListSchemes(t *testing.T) {
	schemes := ListSchemes()
	if len(schemes) == 0 {
		t.Fatal("ListSchemes() returned empty")
	}
	// Verify sorted
	for i := 1; i < len(schemes); i++ {
		if schemes[i] < schemes[i-1] {
			t.Errorf("ListSchemes() not sorted: %s < %s at index %d", schemes[i], schemes[i-1], i)
			break
		}
	}
}

func TestGetScheme(t *testing.T) {
	s := GetScheme("Catppuccin Mocha")
	if s == nil {
		t.Fatal("GetScheme('Catppuccin Mocha') returned nil")
	}
	if s.Background != "#1e1e2e" {
		t.Errorf("Catppuccin Mocha background = %s, want #1e1e2e", s.Background)
	}
	if s.Foreground != "#cdd6f4" {
		t.Errorf("Catppuccin Mocha foreground = %s, want #cdd6f4", s.Foreground)
	}
}

func TestGetSchemeNotFound(t *testing.T) {
	if s := GetScheme("nonexistent-theme-xyz"); s != nil {
		t.Error("GetScheme for nonexistent theme should return nil")
	}
}

func TestSchemeFieldsPopulated(t *testing.T) {
	s := GetScheme("Dracula")
	if s == nil {
		t.Fatal("GetScheme('Dracula') returned nil")
	}
	fields := []struct {
		name, val string
	}{
		{"Black", s.Black},
		{"Red", s.Red},
		{"Green", s.Green},
		{"Blue", s.Blue},
		{"Background", s.Background},
		{"Foreground", s.Foreground},
	}
	for _, f := range fields {
		if f.val == "" {
			t.Errorf("Dracula.%s is empty", f.name)
		}
		if len(f.val) != 7 || f.val[0] != '#' {
			t.Errorf("Dracula.%s = %q, not valid hex", f.name, f.val)
		}
	}
}
