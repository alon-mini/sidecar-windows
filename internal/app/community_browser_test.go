package app

import (
	"testing"

	"github.com/marcus/sidecar/internal/community"
	"github.com/marcus/sidecar/internal/styles"
)

func TestBuildUnifiedThemeList(t *testing.T) {
	entries := buildUnifiedThemeList()
	builtInCount := len(styles.ListThemes())
	communityCount := len(community.ListSchemes())

	// +1 for separator between built-in and community
	expectedTotal := builtInCount + 1 + communityCount
	if len(entries) != expectedTotal {
		t.Errorf("expected %d entries (inc separator), got %d", expectedTotal, len(entries))
	}

	// Built-in themes should come first
	for i := 0; i < builtInCount; i++ {
		if !entries[i].IsBuiltIn {
			t.Errorf("entry %d should be built-in", i)
		}
	}

	// Separator at boundary
	if !entries[builtInCount].IsSeparator {
		t.Error("expected separator between built-in and community themes")
	}

	// Community themes after separator
	for i := builtInCount + 1; i < len(entries); i++ {
		if entries[i].IsBuiltIn || entries[i].IsSeparator {
			t.Errorf("entry %d should be community, got built-in or separator", i)
		}
	}
}

func TestFilterThemeEntries(t *testing.T) {
	entries := buildUnifiedThemeList()

	// Filter for a built-in theme
	filtered := filterThemeEntries(entries, "dracula")
	found := false
	for _, e := range filtered {
		if e.IsBuiltIn && e.ThemeKey == "dracula" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find dracula in filtered results")
	}

	// Empty query returns all (including separator)
	all := filterThemeEntries(entries, "")
	if len(all) != len(entries) {
		t.Errorf("empty filter: expected %d, got %d", len(entries), len(all))
	}

	// Filtering excludes separators
	filtered2 := filterThemeEntries(entries, "a")
	for _, e := range filtered2 {
		if e.IsSeparator {
			t.Error("filtered results should not contain separators")
		}
	}

	// No matches
	none := filterThemeEntries(entries, "zzz-nonexistent-theme-xyz")
	if len(none) != 0 {
		t.Errorf("expected 0 matches, got %d", len(none))
	}
}

func TestUnifiedThemeCursorNavigation(t *testing.T) {
	var m Model
	m.width = 80
	m.height = 40
	m.initThemeSwitcher()

	if len(m.themeSwitcherFiltered) == 0 {
		t.Fatal("expected themes to be available")
	}

	// Verify list has both built-in and community entries
	hasBuiltIn := false
	hasCommunity := false
	for _, e := range m.themeSwitcherFiltered {
		if e.IsBuiltIn {
			hasBuiltIn = true
		} else {
			hasCommunity = true
		}
	}
	if !hasBuiltIn {
		t.Error("expected built-in themes in unified list")
	}
	if !hasCommunity {
		t.Error("expected community themes in unified list")
	}
}
