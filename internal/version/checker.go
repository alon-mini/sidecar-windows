package version

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// UpdateAvailableMsg is sent when a new sidecar version is available.
type UpdateAvailableMsg struct {
	CurrentVersion string
	LatestVersion  string
	UpdateCommand  string
	ReleaseNotes   string
	ReleaseURL     string
	InstallMethod  InstallMethod
}

// TdVersionMsg is sent with td version info (installed or not).
type TdVersionMsg struct {
	Installed      bool
	CurrentVersion string
	LatestVersion  string
	HasUpdate      bool
}

// updateCommand generates the update command based on install method.
func updateCommand(version string, method InstallMethod) string {
	switch method {
	case InstallMethodHomebrew:
		return "brew upgrade sidecar"
	case InstallMethodScoop:
		return "scoop update sidecar"
	case InstallMethodBinary:
		return fmt.Sprintf("https://github.com/marcus/sidecar/releases/tag/%s", version)
	default:
		return fmt.Sprintf(
			"go install -ldflags \"-X main.Version=%s\" github.com/marcus/sidecar/cmd/sidecar@%s",
			version, version,
		)
	}
}

// CheckAsync returns a Bubble Tea command that checks for updates in background.
func CheckAsync(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		method := DetectInstallMethod()

		// Check cache first
		if cached, err := LoadCache(); err == nil && IsCacheValid(cached, currentVersion) {
			if cached.HasUpdate {
				return UpdateAvailableMsg{
					CurrentVersion: currentVersion,
					LatestVersion:  cached.LatestVersion,
					UpdateCommand:  updateCommand(cached.LatestVersion, method),
					InstallMethod:  method,
				}
			}
			return nil // up-to-date, cached
		}

		// Cache miss or invalid, fetch from GitHub
		result := Check(currentVersion)

		// Only cache successful checks (don't cache network errors)
		if result.Error == nil {
			_ = SaveCache(&CacheEntry{
				LatestVersion:  result.LatestVersion,
				CurrentVersion: currentVersion,
				CheckedAt:      time.Now(),
				HasUpdate:      result.HasUpdate,
			})
		}

		if result.HasUpdate {
			return UpdateAvailableMsg{
				CurrentVersion: currentVersion,
				LatestVersion:  result.LatestVersion,
				UpdateCommand:  updateCommand(result.LatestVersion, method),
				ReleaseNotes:   result.ReleaseNotes,
				ReleaseURL:     result.UpdateURL,
				InstallMethod:  method,
			}
		}

		return nil
	}
}

// ForceCheckAsync checks for updates, ignoring the cache.
func ForceCheckAsync(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		method := DetectInstallMethod()
		result := Check(currentVersion)
		if result.Error == nil {
			_ = SaveCache(&CacheEntry{
				LatestVersion:  result.LatestVersion,
				CurrentVersion: currentVersion,
				CheckedAt:      time.Now(),
				HasUpdate:      result.HasUpdate,
			})
		}
		if result.HasUpdate {
			return UpdateAvailableMsg{
				CurrentVersion: currentVersion,
				LatestVersion:  result.LatestVersion,
				UpdateCommand:  updateCommand(result.LatestVersion, method),
				ReleaseNotes:   result.ReleaseNotes,
				ReleaseURL:     result.UpdateURL,
				InstallMethod:  method,
			}
		}
		return nil
	}
}

// tdUpdateCommand generates the update command for td based on install method.
func tdUpdateCommand(version string, method InstallMethod) string {
	switch method {
	case InstallMethodHomebrew:
		return "brew upgrade td"
	case InstallMethodScoop:
		return "scoop update td"
	default:
		return fmt.Sprintf(
			"go install github.com/marcus/td@%s",
			version,
		)
	}
}

// GetTdVersion returns the installed td version by running `td version --short`.
// Returns empty string if td is not installed or command fails.
func GetTdVersion() string {
	out, err := exec.Command("td", "version", "--short").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// CheckTdAsync returns a Bubble Tea command that checks td version in background.
// Returns TdVersionMsg with installation status and version info.
func CheckTdAsync() tea.Cmd {
	return func() tea.Msg {
		tdVersion := GetTdVersion()

		// td not installed
		if tdVersion == "" {
			return TdVersionMsg{Installed: false}
		}

		// Check cache first
		if cached, err := LoadTdCache(); err == nil && IsCacheValid(cached, tdVersion) {
			return TdVersionMsg{
				Installed:      true,
				CurrentVersion: tdVersion,
				LatestVersion:  cached.LatestVersion,
				HasUpdate:      cached.HasUpdate,
			}
		}

		// Cache miss or invalid, fetch from GitHub
		result := CheckTd(tdVersion)

		// Only cache successful checks
		if result.Error == nil {
			_ = SaveTdCache(&CacheEntry{
				LatestVersion:  result.LatestVersion,
				CurrentVersion: tdVersion,
				CheckedAt:      time.Now(),
				HasUpdate:      result.HasUpdate,
			})
		}

		return TdVersionMsg{
			Installed:      true,
			CurrentVersion: tdVersion,
			LatestVersion:  result.LatestVersion,
			HasUpdate:      result.HasUpdate,
		}
	}
}

// ForceCheckTdAsync checks for td updates, ignoring the cache.
func ForceCheckTdAsync() tea.Cmd {
	return func() tea.Msg {
		tdVersion := GetTdVersion()
		if tdVersion == "" {
			return TdVersionMsg{Installed: false}
		}
		result := CheckTd(tdVersion)
		if result.Error == nil {
			_ = SaveTdCache(&CacheEntry{
				LatestVersion:  result.LatestVersion,
				CurrentVersion: tdVersion,
				CheckedAt:      time.Now(),
				HasUpdate:      result.HasUpdate,
			})
		}
		return TdVersionMsg{
			Installed:      true,
			CurrentVersion: tdVersion,
			LatestVersion:  result.LatestVersion,
			HasUpdate:      result.HasUpdate,
		}
	}
}
