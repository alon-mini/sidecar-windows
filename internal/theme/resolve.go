package theme

import (
	"github.com/marcus/sidecar/internal/community"
	"github.com/marcus/sidecar/internal/config"
	"github.com/marcus/sidecar/internal/styles"
)

// ResolvedTheme represents a fully-determined theme configuration.
type ResolvedTheme struct {
	BaseName      string
	CommunityName string
	Overrides     map[string]interface{}
}

// ResolveTheme determines the effective theme for a project path.
// Priority: project.Theme > global UI.Theme > "default".
func ResolveTheme(cfg *config.Config, projectPath string) ResolvedTheme {
	resolved := ResolvedTheme{
		BaseName:      cfg.UI.Theme.Name,
		CommunityName: cfg.UI.Theme.Community,
		Overrides:     cfg.UI.Theme.Overrides,
	}

	for _, proj := range cfg.Projects.List {
		if proj.Path == projectPath && proj.Theme != nil {
			resolved.BaseName = proj.Theme.Name
			resolved.CommunityName = proj.Theme.Community
			resolved.Overrides = proj.Theme.Overrides
			break
		}
	}

	if resolved.BaseName == "" {
		resolved.BaseName = "default"
	}

	return resolved
}

// ApplyResolved applies a resolved theme to the styles system.
func ApplyResolved(r ResolvedTheme) {
	if r.CommunityName != "" {
		scheme := community.GetScheme(r.CommunityName)
		if scheme != nil {
			palette := community.Convert(scheme)
			communityOverrides := community.PaletteToOverrides(palette)
			// Layer user overrides on top of community-derived colors
			for k, v := range r.Overrides {
				communityOverrides[k] = v
			}
			styles.ApplyThemeWithGenericOverrides(r.BaseName, communityOverrides)
			return
		}
	}

	if len(r.Overrides) > 0 {
		styles.ApplyThemeWithGenericOverrides(r.BaseName, r.Overrides)
	} else {
		styles.ApplyTheme(r.BaseName)
	}
}
