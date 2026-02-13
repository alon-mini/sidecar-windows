package conversations

import (
	"testing"

	"github.com/marcus/sidecar/internal/adapter"
)

func TestAdapterAbbrev(t *testing.T) {
	tests := []struct {
		name    string
		session adapter.Session
		want    string
	}{
		{
			name:    "claude-code",
			session: adapter.Session{AdapterID: "claude-code"},
			want:    "CC",
		},
		{
			name:    "codex",
			session: adapter.Session{AdapterID: "codex"},
			want:    "CX",
		},
		{
			name:    "opencode",
			session: adapter.Session{AdapterID: "opencode"},
			want:    "OC",
		},
		{
			name:    "gemini-cli",
			session: adapter.Session{AdapterID: "gemini-cli"},
			want:    "GC",
		},
		{
			name:    "custom adapter with name",
			session: adapter.Session{AdapterID: "mytool", AdapterName: "My Tool"},
			want:    "MY",
		},
		{
			name:    "custom adapter with only ID",
			session: adapter.Session{AdapterID: "warp"},
			want:    "WA",
		},
		{
			name:    "empty ID and name",
			session: adapter.Session{},
			want:    "",
		},
		{
			name:    "short name",
			session: adapter.Session{AdapterID: "x", AdapterName: "X"},
			want:    "X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapterAbbrev(tt.session)
			if got != tt.want {
				t.Errorf("adapterAbbrev() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCategoryBadgeText(t *testing.T) {
	tests := []struct {
		name    string
		session adapter.Session
		want    string
	}{
		{
			name:    "interactive returns empty",
			session: adapter.Session{SessionCategory: adapter.SessionCategoryInteractive},
			want:    "",
		},
		{
			name:    "empty category returns empty",
			session: adapter.Session{},
			want:    "",
		},
		{
			name:    "cron returns cron",
			session: adapter.Session{SessionCategory: adapter.SessionCategoryCron},
			want:    "cron",
		},
		{
			name:    "system returns sys",
			session: adapter.Session{SessionCategory: adapter.SessionCategorySystem},
			want:    "sys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categoryBadgeText(tt.session)
			if got != tt.want {
				t.Errorf("categoryBadgeText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderCategoryBadge(t *testing.T) {
	// Interactive/empty should produce no badge
	if badge := renderCategoryBadge(adapter.Session{}); badge != "" {
		t.Errorf("renderCategoryBadge(empty) = %q, want empty", badge)
	}
	if badge := renderCategoryBadge(adapter.Session{SessionCategory: adapter.SessionCategoryInteractive}); badge != "" {
		t.Errorf("renderCategoryBadge(interactive) = %q, want empty", badge)
	}

	// Cron should produce non-empty styled string containing "cron"
	cronBadge := renderCategoryBadge(adapter.Session{SessionCategory: adapter.SessionCategoryCron})
	if cronBadge == "" {
		t.Error("renderCategoryBadge(cron) is empty, want styled badge")
	}

	// System should produce non-empty styled string containing "sys"
	sysBadge := renderCategoryBadge(adapter.Session{SessionCategory: adapter.SessionCategorySystem})
	if sysBadge == "" {
		t.Error("renderCategoryBadge(system) is empty, want styled badge")
	}
}

func TestAdapterBadgeText(t *testing.T) {
	tests := []struct {
		name    string
		session adapter.Session
		want    string
	}{
		{
			name:    "with adapter icon",
			session: adapter.Session{AdapterIcon: "🤖"},
			want:    "🤖",
		},
		{
			name:    "no icon with known adapter",
			session: adapter.Session{AdapterID: "claude-code"},
			want:    "●CC",
		},
		{
			name:    "no icon empty ID and name",
			session: adapter.Session{},
			want:    "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapterBadgeText(tt.session)
			if got != tt.want {
				t.Errorf("adapterBadgeText() = %q, want %q", got, tt.want)
			}
		})
	}
}
