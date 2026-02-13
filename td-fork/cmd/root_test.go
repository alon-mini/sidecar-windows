package cmd

import "testing"

func TestFirstNonFlagArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "skips leading flags",
			args: []string{"--flag", "unknown-cmd"},
			want: "unknown-cmd",
		},
		{
			name: "all flags",
			args: []string{"-h", "--help"},
			want: "",
		},
		{
			name: "finds command after help",
			args: []string{"--help", "list"},
			want: "list",
		},
		{
			name: "no args",
			args: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstNonFlagArg(tt.args); got != tt.want {
				t.Errorf("firstNonFlagArg(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}
