package workspace

import (
	"reflect"
	"sort"
	"testing"
)

func TestIntersection(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "no overlap",
			a:        []string{"a", "b", "c"},
			b:        []string{"d", "e", "f"},
			expected: nil,
		},
		{
			name:     "partial overlap",
			a:        []string{"a", "b", "c"},
			b:        []string{"b", "c", "d"},
			expected: []string{"b", "c"},
		},
		{
			name:     "full overlap",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty a",
			a:        []string{},
			b:        []string{"a", "b"},
			expected: nil,
		},
		{
			name:     "empty b",
			a:        []string{"a", "b"},
			b:        []string{},
			expected: nil,
		},
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: nil,
		},
		{
			name:     "single element overlap",
			a:        []string{"a", "b", "c"},
			b:        []string{"b"},
			expected: []string{"b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := intersection(tt.a, tt.b)

			// Sort both for comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("intersection(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestHasConflict(t *testing.T) {
	p := &Plugin{}

	conflicts := []Conflict{
		{
			Worktrees: []string{"wt1", "wt2"},
			Files:     []string{"file.go"},
		},
		{
			Worktrees: []string{"wt2", "wt3"},
			Files:     []string{"other.go"},
		},
	}

	tests := []struct {
		name         string
		worktreeName string
		expected     bool
	}{
		{"wt1 has conflict", "wt1", true},
		{"wt2 has conflict", "wt2", true},
		{"wt3 has conflict", "wt3", true},
		{"wt4 has no conflict", "wt4", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.hasConflict(tt.worktreeName, conflicts)
			if result != tt.expected {
				t.Errorf("hasConflict(%q) = %v, want %v", tt.worktreeName, result, tt.expected)
			}
		})
	}
}

func TestGetConflictingFiles(t *testing.T) {
	p := &Plugin{}

	conflicts := []Conflict{
		{
			Worktrees: []string{"wt1", "wt2"},
			Files:     []string{"file1.go", "file2.go"},
		},
		{
			Worktrees: []string{"wt1", "wt3"},
			Files:     []string{"file3.go"},
		},
	}

	tests := []struct {
		name         string
		worktreeName string
		expected     []string
	}{
		{
			name:         "wt1 multiple conflicts",
			worktreeName: "wt1",
			expected:     []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:         "wt2 single conflict",
			worktreeName: "wt2",
			expected:     []string{"file1.go", "file2.go"},
		},
		{
			name:         "wt4 no conflict",
			worktreeName: "wt4",
			expected:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.getConflictingFiles(tt.worktreeName, conflicts)

			// Sort for comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getConflictingFiles(%q) = %v, want %v", tt.worktreeName, result, tt.expected)
			}
		})
	}
}

func TestGetConflictingWorktrees(t *testing.T) {
	p := &Plugin{}

	conflicts := []Conflict{
		{
			Worktrees: []string{"wt1", "wt2"},
			Files:     []string{"file1.go"},
		},
		{
			Worktrees: []string{"wt1", "wt3"},
			Files:     []string{"file2.go"},
		},
	}

	tests := []struct {
		name         string
		worktreeName string
		expected     []string
	}{
		{
			name:         "wt1 conflicts with wt2 and wt3",
			worktreeName: "wt1",
			expected:     []string{"wt2", "wt3"},
		},
		{
			name:         "wt2 conflicts with wt1",
			worktreeName: "wt2",
			expected:     []string{"wt1"},
		},
		{
			name:         "wt4 no conflicts",
			worktreeName: "wt4",
			expected:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.getConflictingWorktrees(tt.worktreeName, conflicts)

			// Sort for comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getConflictingWorktrees(%q) = %v, want %v", tt.worktreeName, result, tt.expected)
			}
		})
	}
}
