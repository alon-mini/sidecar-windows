package theme

import (
	"testing"

	"github.com/marcus/sidecar/internal/config"
	"github.com/marcus/sidecar/internal/styles"
)

func TestResolveTheme(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		projectPath string
		want        ResolvedTheme
	}{
		{
			name: "global theme only, no projects",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{Name: "dracula"},
				},
			},
			projectPath: "/some/path",
			want:        ResolvedTheme{BaseName: "dracula"},
		},
		{
			name: "project without theme field falls back to global",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{Name: "monokai"},
				},
				Projects: config.ProjectsConfig{
					List: []config.ProjectConfig{
						{Name: "proj", Path: "/code/proj"},
					},
				},
			},
			projectPath: "/code/proj",
			want:        ResolvedTheme{BaseName: "monokai"},
		},
		{
			name: "project with theme overrides global",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{Name: "monokai"},
				},
				Projects: config.ProjectsConfig{
					List: []config.ProjectConfig{
						{Name: "proj", Path: "/code/proj", Theme: &config.ThemeConfig{Name: "dracula"}},
					},
				},
			},
			projectPath: "/code/proj",
			want:        ResolvedTheme{BaseName: "dracula"},
		},
		{
			name: "community name propagated from global",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{Name: "default", Community: "Dracula"},
				},
			},
			projectPath: "/code/proj",
			want:        ResolvedTheme{BaseName: "default", CommunityName: "Dracula"},
		},
		{
			name: "community name from project theme",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{Name: "default"},
				},
				Projects: config.ProjectsConfig{
					List: []config.ProjectConfig{
						{Name: "proj", Path: "/code/proj", Theme: &config.ThemeConfig{
							Name: "default", Community: "Solarized Dark",
						}},
					},
				},
			},
			projectPath: "/code/proj",
			want:        ResolvedTheme{BaseName: "default", CommunityName: "Solarized Dark"},
		},
		{
			name: "empty base name defaults to default",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{Name: ""},
				},
			},
			projectPath: "/code/proj",
			want:        ResolvedTheme{BaseName: "default"},
		},
		{
			name: "overrides propagated from global",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{
						Name:      "default",
						Overrides: map[string]interface{}{"primary": "#ff0000"},
					},
				},
			},
			projectPath: "/code/proj",
			want: ResolvedTheme{
				BaseName:  "default",
				Overrides: map[string]interface{}{"primary": "#ff0000"},
			},
		},
		{
			name: "project overrides replace global overrides",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{
						Name:      "default",
						Overrides: map[string]interface{}{"primary": "#ff0000"},
					},
				},
				Projects: config.ProjectsConfig{
					List: []config.ProjectConfig{
						{Name: "proj", Path: "/code/proj", Theme: &config.ThemeConfig{
							Name:      "default",
							Overrides: map[string]interface{}{"primary": "#00ff00"},
						}},
					},
				},
			},
			projectPath: "/code/proj",
			want: ResolvedTheme{
				BaseName:  "default",
				Overrides: map[string]interface{}{"primary": "#00ff00"},
			},
		},
		{
			name: "unmatched project path uses global",
			cfg: &config.Config{
				UI: config.UIConfig{
					Theme: config.ThemeConfig{Name: "dracula"},
				},
				Projects: config.ProjectsConfig{
					List: []config.ProjectConfig{
						{Name: "other", Path: "/code/other", Theme: &config.ThemeConfig{Name: "monokai"}},
					},
				},
			},
			projectPath: "/code/proj",
			want:        ResolvedTheme{BaseName: "dracula"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveTheme(tt.cfg, tt.projectPath)
			if got.BaseName != tt.want.BaseName {
				t.Errorf("BaseName = %q, want %q", got.BaseName, tt.want.BaseName)
			}
			if got.CommunityName != tt.want.CommunityName {
				t.Errorf("CommunityName = %q, want %q", got.CommunityName, tt.want.CommunityName)
			}
			if len(got.Overrides) != len(tt.want.Overrides) {
				t.Errorf("Overrides len = %d, want %d", len(got.Overrides), len(tt.want.Overrides))
			}
			for k, wantV := range tt.want.Overrides {
				if gotV, ok := got.Overrides[k]; !ok || gotV != wantV {
					t.Errorf("Overrides[%q] = %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}

func TestApplyResolved(t *testing.T) {
	t.Run("no community no overrides applies base theme", func(t *testing.T) {
		ApplyResolved(ResolvedTheme{BaseName: "dracula"})
		if got := styles.GetCurrentThemeName(); got != "dracula" {
			t.Errorf("theme = %q, want %q", got, "dracula")
		}
	})

	t.Run("with overrides applies base plus overrides", func(t *testing.T) {
		ApplyResolved(ResolvedTheme{
			BaseName:  "default",
			Overrides: map[string]interface{}{"primary": "#ff0000"},
		})
		if got := styles.GetCurrentThemeName(); got != "default" {
			t.Errorf("theme = %q, want %q", got, "default")
		}
		th := styles.GetCurrentTheme()
		if th.Colors.Primary != "#ff0000" {
			t.Errorf("primary = %q, want %q", th.Colors.Primary, "#ff0000")
		}
	})

	t.Run("community theme applies converted palette", func(t *testing.T) {
		ApplyResolved(ResolvedTheme{
			BaseName:      "default",
			CommunityName: "Dracula",
		})
		if got := styles.GetCurrentThemeName(); got != "default" {
			t.Errorf("theme = %q, want %q", got, "default")
		}
		th := styles.GetCurrentTheme()
		// Dracula community theme should set primary to a non-default value
		if th.Colors.Primary == "" {
			t.Error("primary should be set by community theme")
		}
	})

	t.Run("community with user overrides layers correctly", func(t *testing.T) {
		ApplyResolved(ResolvedTheme{
			BaseName:      "default",
			CommunityName: "Dracula",
			Overrides:     map[string]interface{}{"primary": "#123456"},
		})
		th := styles.GetCurrentTheme()
		if th.Colors.Primary != "#123456" {
			t.Errorf("primary = %q, want user override %q", th.Colors.Primary, "#123456")
		}
	})

	t.Run("unknown community name falls back to overrides", func(t *testing.T) {
		ApplyResolved(ResolvedTheme{
			BaseName:      "default",
			CommunityName: "NonExistentScheme",
			Overrides:     map[string]interface{}{"primary": "#abcdef"},
		})
		th := styles.GetCurrentTheme()
		if th.Colors.Primary != "#abcdef" {
			t.Errorf("primary = %q, want %q", th.Colors.Primary, "#abcdef")
		}
	})

	t.Run("unknown community name no overrides applies base", func(t *testing.T) {
		ApplyResolved(ResolvedTheme{
			BaseName:      "monokai",
			CommunityName: "NonExistentScheme",
		})
		if got := styles.GetCurrentThemeName(); got != "monokai" {
			t.Errorf("theme = %q, want %q", got, "monokai")
		}
	})
}
