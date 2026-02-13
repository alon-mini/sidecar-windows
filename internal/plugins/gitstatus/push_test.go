package gitstatus

import (
	"testing"
)

func TestIsCommitPushed_FullHashMatch(t *testing.T) {
	ps := &PushStatus{
		HasUpstream:    true,
		UnpushedHashes: []string{"abc1234567890abcdef0123456789abcdef01234"},
	}

	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "unpushed commit full hash",
			hash:     "abc1234567890abcdef0123456789abcdef01234",
			expected: false,
		},
		{
			name:     "pushed commit full hash",
			hash:     "def4567890abcdef0123456789abcdef012345a",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ps.IsCommitPushed(tt.hash)
			if got != tt.expected {
				t.Errorf("IsCommitPushed(%q) = %v, want %v", tt.hash, got, tt.expected)
			}
		})
	}
}

func TestIsCommitPushed_ShortHashMatch(t *testing.T) {
	ps := &PushStatus{
		HasUpstream:    true,
		UnpushedHashes: []string{"abc1234567890abcdef0123456789abcdef01234"},
	}

	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "unpushed commit short hash",
			hash:     "abc1234",
			expected: false,
		},
		{
			name:     "pushed commit short hash",
			hash:     "def4567",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ps.IsCommitPushed(tt.hash)
			if got != tt.expected {
				t.Errorf("IsCommitPushed(%q) = %v, want %v", tt.hash, got, tt.expected)
			}
		})
	}
}

func TestIsCommitPushed_NoUpstream(t *testing.T) {
	ps := &PushStatus{
		HasUpstream:    false,
		UnpushedHashes: []string{},
	}

	got := ps.IsCommitPushed("abc1234")
	if got != false {
		t.Errorf("IsCommitPushed with no upstream should return false, got %v", got)
	}
}

func TestIsCommitPushed_MultipleUnpushed(t *testing.T) {
	ps := &PushStatus{
		HasUpstream: true,
		UnpushedHashes: []string{
			"abc1234567890abcdef0123456789abcdef01234",
			"def4567890abcdef0123456789abcdef012345a1",
			"ghi7890abcdef0123456789abcdef012345a12bc",
		},
	}

	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "first unpushed commit",
			hash:     "abc1234",
			expected: false,
		},
		{
			name:     "second unpushed commit",
			hash:     "def4567",
			expected: false,
		},
		{
			name:     "third unpushed commit",
			hash:     "ghi7890",
			expected: false,
		},
		{
			name:     "pushed commit",
			hash:     "jkl0123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ps.IsCommitPushed(tt.hash)
			if got != tt.expected {
				t.Errorf("IsCommitPushed(%q) = %v, want %v", tt.hash, got, tt.expected)
			}
		})
	}
}

func TestFormatAheadBehind_Synced(t *testing.T) {
	ps := &PushStatus{
		HasUpstream: true,
		Ahead:       0,
		Behind:      0,
	}

	got := ps.FormatAheadBehind()
	if got != "" {
		t.Errorf("FormatAheadBehind() for synced branch should be empty, got %q", got)
	}
}

func TestFormatAheadBehind_Ahead(t *testing.T) {
	ps := &PushStatus{
		HasUpstream: true,
		Ahead:       3,
		Behind:      0,
	}

	got := ps.FormatAheadBehind()
	expected := "↑3"
	if got != expected {
		t.Errorf("FormatAheadBehind() = %q, want %q", got, expected)
	}
}

func TestFormatAheadBehind_Behind(t *testing.T) {
	ps := &PushStatus{
		HasUpstream: true,
		Ahead:       0,
		Behind:      2,
	}

	got := ps.FormatAheadBehind()
	expected := "↓2"
	if got != expected {
		t.Errorf("FormatAheadBehind() = %q, want %q", got, expected)
	}
}

func TestFormatAheadBehind_Diverged(t *testing.T) {
	ps := &PushStatus{
		HasUpstream: true,
		Ahead:       3,
		Behind:      2,
	}

	got := ps.FormatAheadBehind()
	expected := "↑3 ↓2"
	if got != expected {
		t.Errorf("FormatAheadBehind() = %q, want %q", got, expected)
	}
}

func TestFormatAheadBehind_NoUpstream(t *testing.T) {
	ps := &PushStatus{
		HasUpstream: false,
		Ahead:       0,
		Behind:      0,
	}

	got := ps.FormatAheadBehind()
	expected := "no upstream"
	if got != expected {
		t.Errorf("FormatAheadBehind() = %q, want %q", got, expected)
	}
}

func TestFormatAheadBehind_DetachedHead(t *testing.T) {
	ps := &PushStatus{
		DetachedHead: true,
		HasUpstream:  false,
	}

	got := ps.FormatAheadBehind()
	expected := "detached"
	if got != expected {
		t.Errorf("FormatAheadBehind() = %q, want %q", got, expected)
	}
}

func TestNeedsForce(t *testing.T) {
	tests := []struct {
		name     string
		ps       *PushStatus
		expected bool
	}{
		{
			name:     "ahead only",
			ps:       &PushStatus{Ahead: 3, Behind: 0},
			expected: false,
		},
		{
			name:     "behind only",
			ps:       &PushStatus{Ahead: 0, Behind: 2},
			expected: false,
		},
		{
			name:     "both ahead and behind",
			ps:       &PushStatus{Ahead: 3, Behind: 2},
			expected: true,
		},
		{
			name:     "neither ahead nor behind",
			ps:       &PushStatus{Ahead: 0, Behind: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ps.NeedsForce()
			if got != tt.expected {
				t.Errorf("NeedsForce() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCanPush(t *testing.T) {
	tests := []struct {
		name     string
		ps       *PushStatus
		expected bool
	}{
		{
			name: "ahead of upstream",
			ps: &PushStatus{
				HasUpstream:  true,
				Ahead:        3,
				DetachedHead: false,
			},
			expected: true,
		},
		{
			name: "no upstream, not detached",
			ps: &PushStatus{
				HasUpstream:  false,
				Ahead:        0,
				DetachedHead: false,
			},
			expected: true,
		},
		{
			name: "detached head",
			ps: &PushStatus{
				HasUpstream:  false,
				Ahead:        0,
				DetachedHead: true,
			},
			expected: false,
		},
		{
			name: "synced with upstream",
			ps: &PushStatus{
				HasUpstream:  true,
				Ahead:        0,
				DetachedHead: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ps.CanPush()
			if got != tt.expected {
				t.Errorf("CanPush() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParsePushOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "empty output",
			output:   "",
			expected: "Push completed",
		},
		{
			name:     "up to date",
			output:   "Everything up-to-date",
			expected: "Already up-to-date",
		},
		{
			name:     "successful push",
			output:   "abc1234..def5678  main -> main",
			expected: "Pushed successfully",
		},
		{
			name:     "new branch",
			output:   " * [new branch]      feature -> feature",
			expected: "Created remote branch",
		},
		{
			name:     "generic completion",
			output:   "Some git output",
			expected: "Push completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePushOutput(tt.output)
			if got != tt.expected {
				t.Errorf("ParsePushOutput(%q) = %q, want %q", tt.output, got, tt.expected)
			}
		})
	}
}

func TestPopulatePushStatus(t *testing.T) {
	commits := []*Commit{
		{Hash: "abc1234567890abcdef0123456789abcdef01234", Subject: "commit 1"},
		{Hash: "def4567890abcdef0123456789abcdef012345a1", Subject: "commit 2"},
		{Hash: "ghi7890abcdef0123456789abcdef012345a12bc", Subject: "commit 3"},
	}

	ps := &PushStatus{
		HasUpstream:    true,
		UnpushedHashes: []string{"def4567890abcdef0123456789abcdef012345a1"},
	}

	PopulatePushStatus(commits, ps)

	expectedPushed := []bool{true, false, true}
	for i, commit := range commits {
		if commit.Pushed != expectedPushed[i] {
			t.Errorf("commits[%d].Pushed = %v, want %v", i, commit.Pushed, expectedPushed[i])
		}
	}
}

func TestPopulatePushStatus_NilStatus(t *testing.T) {
	commits := []*Commit{
		{Hash: "abc1234567890abcdef0123456789abcdef01234", Subject: "commit 1"},
	}

	// Should handle nil gracefully
	PopulatePushStatus(commits, nil)

	if commits[0].Pushed != false {
		t.Errorf("Pushed should remain false when status is nil")
	}
}
