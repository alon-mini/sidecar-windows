package gitstatus

import (
	"errors"
	"testing"
)

func TestIsConflictError_MergeConflict(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "merge conflict output",
			err:      &RemoteError{Output: "CONFLICT (content): Merge conflict in file.go\nAutomatic merge failed; fix conflicts and then commit.", Err: errors.New("exit 1")},
			expected: true,
		},
		{
			name:     "rebase conflict output",
			err:      &RemoteError{Output: "error: could not apply abc1234... commit msg\nCONFLICT (content): Merge conflict in main.go", Err: errors.New("exit 1")},
			expected: true,
		},
		{
			name:     "non-conflict error",
			err:      &RemoteError{Output: "fatal: Not a git repository", Err: errors.New("exit 128")},
			expected: false,
		},
		{
			name:     "non-RemoteError",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "automatic merge failed",
			err:      &RemoteError{Output: "Automatic merge failed; fix conflicts and then commit the result.", Err: errors.New("exit 1")},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConflictError(tt.err)
			if got != tt.expected {
				t.Errorf("IsConflictError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPullErrorMsg_Strategy(t *testing.T) {
	msg := PullErrorMsg{
		Err:      errors.New("failed"),
		Strategy: "rebase",
	}

	if msg.Strategy != "rebase" {
		t.Errorf("PullErrorMsg.Strategy = %q, want %q", msg.Strategy, "rebase")
	}
}
