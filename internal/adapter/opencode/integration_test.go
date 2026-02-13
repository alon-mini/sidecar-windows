//go:build integration

package opencode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationWithRealData(t *testing.T) {
	a := New()

	home, _ := os.UserHomeDir()
	projects := []string{
		filepath.Join(home, "code", "sidecar"),
		filepath.Join(home, "code", "td"),
	}

	for _, p := range projects {
		t.Run(filepath.Base(p), func(t *testing.T) {
			found, err := a.Detect(p)
			if err != nil {
				t.Fatalf("Detect error: %v", err)
			}

			if !found {
				t.Skipf("no OpenCode sessions for %s", p)
			}

			sessions, err := a.Sessions(p)
			if err != nil {
				t.Fatalf("Sessions error: %v", err)
			}

			t.Logf("Found %d sessions", len(sessions))
			for i, s := range sessions {
				if i >= 5 {
					t.Logf("... and %d more", len(sessions)-5)
					break
				}
				t.Logf("  %s: %s (tokens=%d, subagent=%v)", s.ID, s.Name, s.TotalTokens, s.IsSubAgent)
			}

			if len(sessions) > 0 {
				msgs, err := a.Messages(sessions[0].ID)
				if err != nil {
					t.Fatalf("Messages error: %v", err)
				}
				t.Logf("\nFirst session has %d messages", len(msgs))
				for i, m := range msgs {
					if i >= 3 {
						break
					}
					content := m.Content
					if len(content) > 50 {
						content = content[:50] + "..."
					}
					t.Logf("  [%s] %s: %s", m.Timestamp.Format("15:04"), m.Role, content)
					if len(m.ToolUses) > 0 {
						t.Logf("    tools: %d", len(m.ToolUses))
					}
				}
			}
		})
	}
}
