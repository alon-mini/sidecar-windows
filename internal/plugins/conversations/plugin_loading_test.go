package conversations

import "testing"

func TestDeriveWorktreeNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		wtPath   string
		mainPath string
		want     string
	}{
		{
			name:     "standard prefixed path",
			wtPath:   "/Users/foo/code/myrepo-feature-auth",
			mainPath: "/Users/foo/code/myrepo",
			want:     "feature-auth",
		},
		{
			name:     "path without prefix",
			wtPath:   "/Users/foo/code/some-other-dir",
			mainPath: "/Users/foo/code/myrepo",
			want:     "some-other-dir",
		},
		{
			name:     "repo name with hyphen",
			wtPath:   "/Users/foo/code/my-repo-feature",
			mainPath: "/Users/foo/code/my-repo",
			want:     "feature",
		},
		{
			name:     "nested paths",
			wtPath:   "/a/b/c/repo-branch",
			mainPath: "/a/b/c/repo",
			want:     "branch",
		},
		{
			name:     "same directory",
			wtPath:   "/Users/foo/code/myrepo",
			mainPath: "/Users/foo/code/myrepo",
			want:     "myrepo",
		},
		{
			name:     "multi-part branch name",
			wtPath:   "/code/sidecar-fix-bug-123",
			mainPath: "/code/sidecar",
			want:     "fix-bug-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveWorktreeNameFromPath(tt.wtPath, tt.mainPath)
			if got != tt.want {
				t.Errorf("deriveWorktreeNameFromPath(%q, %q) = %q, want %q",
					tt.wtPath, tt.mainPath, got, tt.want)
			}
		})
	}
}
