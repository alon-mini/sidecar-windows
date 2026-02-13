package styles

import "testing"

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"#FF0000", true},
		{"#ff0000", true},
		{"#aaBBcc", true},
		{"#FF0000AA", true},
		{"#ff0000aa", true},
		{"", false},
		{"FF0000", false},
		{"#FFF", false},
		{"#GGGGGG", false},
		{"#FF00001", false},
		{"#FF000", false},
		{"#FF0000AAB", false},
		{"hello", false},
		{"#", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsValidHexColor(tt.input); got != tt.want {
				t.Errorf("IsValidHexColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidTheme(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"default", true},
		{"dracula", true},
		{"nonexistent-theme", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidTheme(tt.name); got != tt.want {
				t.Errorf("IsValidTheme(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGetTheme(t *testing.T) {
	theme := GetTheme("default")
	if theme.Name == "" {
		t.Error("GetTheme(\"default\") returned theme with empty name")
	}

	fallback := GetTheme("nonexistent")
	if fallback.Name != DefaultTheme.Name {
		t.Errorf("GetTheme(\"nonexistent\") = %q, want default theme %q", fallback.Name, DefaultTheme.Name)
	}
}
