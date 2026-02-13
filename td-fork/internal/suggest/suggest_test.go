package suggest

import (
	"testing"
)

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		// Empty strings
		{"both empty", "", "", 0},
		{"first empty", "", "abc", 3},
		{"second empty", "abc", "", 3},

		// Identical strings
		{"identical single char", "a", "a", 0},
		{"identical short", "abc", "abc", 0},
		{"identical longer", "hello world", "hello world", 0},

		// Single character difference
		{"single substitution", "cat", "bat", 1},
		{"single insertion", "cat", "cats", 1},
		{"single deletion", "cats", "cat", 1},
		{"single insertion at start", "cat", "acat", 1},
		{"single deletion at start", "acat", "cat", 1},

		// Multiple differences
		{"two substitutions", "cat", "dog", 3},
		{"substitution and insertion", "abc", "abcd", 1},
		{"multiple edits", "kitten", "sitting", 3},
		{"prefix mismatch", "abcdef", "xyzdef", 3},
		{"suffix mismatch", "abcdef", "abcxyz", 3},
		{"middle mismatch", "abcdef", "abxyef", 2},

		// Case sensitivity
		{"case difference", "ABC", "abc", 3},
		{"partial case", "Abc", "abc", 1},

		// Special characters
		{"with dashes", "--help", "-help", 1},
		{"with underscores", "foo_bar", "foobar", 1},

		// Long strings
		{"long similar", "description", "desciption", 1},
		{"long different", "description", "explanation", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := levenshtein(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestLevenshteinSymmetric(t *testing.T) {
	// Levenshtein distance should be symmetric
	pairs := []struct{ a, b string }{
		{"cat", "bat"},
		{"hello", "world"},
		{"abc", "def"},
		{"", "test"},
		{"kitten", "sitting"},
	}

	for _, p := range pairs {
		ab := levenshtein(p.a, p.b)
		ba := levenshtein(p.b, p.a)
		if ab != ba {
			t.Errorf("levenshtein not symmetric: (%q,%q)=%d but (%q,%q)=%d",
				p.a, p.b, ab, p.b, p.a, ba)
		}
	}
}

func TestFlag(t *testing.T) {
	validFlags := []string{"--help", "--status", "--description", "--reason", "--issue", "--all", "--verbose"}

	tests := []struct {
		name       string
		unknown    string
		validFlags []string
		wantFirst  string // first result should match (empty string means no results expected)
	}{
		// Exact match (distance 0)
		{"exact match with dashes", "--help", validFlags, "--help"},
		{"exact match without dashes", "help", validFlags, "--help"},
		{"exact match status", "status", validFlags, "--status"},

		// Close matches (small edit distance)
		{"typo in help", "hlep", validFlags, "--help"},
		{"typo in status", "staus", validFlags, "--status"},
		{"typo in description", "desciption", validFlags, "--description"},
		{"missing letter", "verbos", validFlags, "--verbose"},
		{"extra letter", "issuee", validFlags, "--issue"},

		// Dash handling
		{"single dash", "-help", validFlags, "--help"},
		{"double dash", "--help", validFlags, "--help"},
		{"triple dash normalized", "---help", validFlags, "--help"},

		// No match (too different)
		{"very long unknown", "thisisaverylongflagthatdoesnotmatch", validFlags, ""},

		// Edge cases
		{"empty valid flags", "help", []string{}, ""},
		{"single valid flag", "hlep", []string{"--help"}, "--help"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Flag(tt.unknown, tt.validFlags)

			if tt.wantFirst == "" {
				if len(result) != 0 {
					t.Errorf("Flag(%q) expected no results, got %v", tt.unknown, result)
				}
				return
			}

			if len(result) == 0 {
				t.Errorf("Flag(%q) returned no results, want first=%q", tt.unknown, tt.wantFirst)
				return
			}

			if result[0] != tt.wantFirst {
				t.Errorf("Flag(%q) first result = %q, want %q",
					tt.unknown, result[0], tt.wantFirst)
			}
		})
	}
}

func TestFlagEmptyInput(t *testing.T) {
	validFlags := []string{"--help", "--status", "--all"}
	result := Flag("", validFlags)

	// Empty input has distance equal to normalized flag length
	// "--all" -> "all" (3 chars) -> distance 3 from "" which equals maxDist=max(3,0)=3
	// Should match flags within distance threshold
	if len(result) == 0 {
		t.Error("Flag(\"\") with valid flags should return some matches")
	}

	// --all (normalized "all", len 3) should be first since it's shortest
	if result[0] != "--all" {
		t.Errorf("Flag(\"\") first result = %q, want --all", result[0])
	}
}

func TestFlagShortInput(t *testing.T) {
	validFlags := []string{"--help", "--status", "--all", "--verbose"}

	// Short input "xyz" has maxDist = max(3, 3/2) = max(3,1) = 3
	// "all" has distance 3 from "xyz", so it matches
	result := Flag("xyz", validFlags)
	if len(result) == 0 {
		t.Error("Flag(\"xyz\") should match --all (distance 3)")
	}
}

func TestFlagSortsByScore(t *testing.T) {
	// Flags ordered by increasing distance from "sta"
	validFlags := []string{"--status", "--state", "--start", "--stat", "--stop", "--help"}

	result := Flag("sta", validFlags)

	// Should be sorted by distance
	if len(result) < 2 {
		t.Fatalf("Expected at least 2 results, got %d", len(result))
	}

	// "stat" (distance 1) should come before "status" (distance 3) or "state" (distance 2)
	// The exact order depends on levenshtein("sta", normalized)
	// "sta" -> "stat" = 1, "sta" -> "state" = 2, "sta" -> "status" = 3
	if result[0] != "--stat" {
		t.Errorf("Expected --stat first (distance 1), got %q", result[0])
	}
}

func TestFlagLimitsToThree(t *testing.T) {
	// Many similar flags
	validFlags := []string{
		"--aa", "--ab", "--ac", "--ad", "--ae", "--af", "--ag",
	}

	result := Flag("a", validFlags)

	if len(result) > 3 {
		t.Errorf("Flag should return at most 3 results, got %d: %v", len(result), result)
	}
}

func TestFlagDistanceThreshold(t *testing.T) {
	validFlags := []string{"--description"}

	tests := []struct {
		name    string
		unknown string
		wantHit bool
	}{
		// "description" has 11 chars, so maxDist = max(3, 11/2) = max(3, 5) = 5
		{"within threshold", "dscription", true},  // distance 1
		{"at threshold", "xxxription", true},      // distance 4
		{"beyond threshold", "xxxxxxxxxx", false}, // distance > 5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Flag(tt.unknown, validFlags)
			gotHit := len(result) > 0
			if gotHit != tt.wantHit {
				t.Errorf("Flag(%q) match = %v, want %v", tt.unknown, gotHit, tt.wantHit)
			}
		})
	}
}

func TestCommonFlagAliases(t *testing.T) {
	// Verify expected aliases exist
	expectedAliases := map[string]string{
		"note":    "--description, -d",
		"notes":   "--description, -d",
		"msg":     "--description, -d",
		"comment": "--reason, -m",
		"message": "--reason, -m",
		"id":      "--issue",
		"task":    "--issue",
		"force":   "(not supported - use confirmation prompt)",
		"version": "use: td version",
		"state":   "--status, -s",
		"review":  "use: td review <issue-id> (after td handoff)",
	}

	for alias, expected := range expectedAliases {
		if got, ok := CommonFlagAliases[alias]; !ok {
			t.Errorf("CommonFlagAliases missing key %q", alias)
		} else if got != expected {
			t.Errorf("CommonFlagAliases[%q] = %q, want %q", alias, got, expected)
		}
	}
}

func TestGetFlagHint(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		expected string
	}{
		// Known aliases
		{"note alias", "note", "--description, -d"},
		{"message alias", "message", "--reason, -m"},
		{"task alias", "task", "--issue"},
		{"version alias", "version", "use: td version"},
		{"state alias", "state", "--status, -s"},

		// With leading dashes (should be normalized)
		{"single dash", "-note", "--description, -d"},
		{"double dash", "--note", "--description, -d"},
		{"triple dash", "---note", "--description, -d"},

		// Case normalization
		{"uppercase", "NOTE", "--description, -d"},
		{"mixed case", "NoTe", "--description, -d"},

		// Unknown flags
		{"unknown flag", "unknown", ""},
		{"random text", "xyz", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFlagHint(tt.flag)
			if result != tt.expected {
				t.Errorf("GetFlagHint(%q) = %q, want %q", tt.flag, result, tt.expected)
			}
		})
	}
}

func TestFlagWithRealWorldTypos(t *testing.T) {
	validFlags := []string{
		"--help", "--status", "--description", "--reason",
		"--issue", "--all", "--verbose", "--quiet",
	}

	tests := []struct {
		typo     string
		expected string
	}{
		// Common typos
		{"--hepl", "--help"},
		{"--hlep", "--help"},
		{"--statsu", "--status"},
		{"--stauts", "--status"},
		{"--descrption", "--description"},
		{"--descriptin", "--description"},
		{"--reson", "--reason"},
		{"--resaon", "--reason"},
		{"--isue", "--issue"},
		{"--isseu", "--issue"},
		{"--verbsoe", "--verbose"},
		{"--queit", "--quiet"},
	}

	for _, tt := range tests {
		t.Run(tt.typo, func(t *testing.T) {
			result := Flag(tt.typo, validFlags)
			if len(result) == 0 {
				t.Errorf("Flag(%q) returned no results, expected %q", tt.typo, tt.expected)
				return
			}
			if result[0] != tt.expected {
				t.Errorf("Flag(%q) = %q, want %q", tt.typo, result[0], tt.expected)
			}
		})
	}
}

func TestFlagPreservesFlagFormat(t *testing.T) {
	// Ensure returned flags preserve original formatting
	tests := []struct {
		name       string
		validFlags []string
		unknown    string
		wantFormat string
	}{
		{"preserves double dash", []string{"--help"}, "help", "--help"},
		{"preserves single dash", []string{"-h"}, "h", "-h"},
		{"preserves no dash", []string{"help"}, "help", "help"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Flag(tt.unknown, tt.validFlags)
			if len(result) == 0 {
				t.Fatalf("No results for %q", tt.unknown)
			}
			if result[0] != tt.wantFormat {
				t.Errorf("Flag format not preserved: got %q, want %q", result[0], tt.wantFormat)
			}
		})
	}
}

func TestFlagMultipleResults(t *testing.T) {
	// Test that multiple similar flags are returned in order
	validFlags := []string{"--stat", "--state", "--status", "--start"}

	result := Flag("stat", validFlags)

	// Should get exactly --stat first (distance 0)
	if len(result) == 0 {
		t.Fatal("Expected at least one result")
	}
	if result[0] != "--stat" {
		t.Errorf("First result should be exact match --stat, got %q", result[0])
	}

	// Should have multiple results
	if len(result) < 2 {
		t.Errorf("Expected multiple suggestions, got %d", len(result))
	}
}

func BenchmarkLevenshtein(b *testing.B) {
	for i := 0; i < b.N; i++ {
		levenshtein("description", "desciption")
	}
}

func BenchmarkFlag(b *testing.B) {
	validFlags := []string{
		"--help", "--status", "--description", "--reason",
		"--issue", "--all", "--verbose", "--quiet",
	}
	for i := 0; i < b.N; i++ {
		Flag("--statsu", validFlags)
	}
}
