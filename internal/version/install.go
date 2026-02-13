package version

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// InstallMethod represents how sidecar was installed.
type InstallMethod string

const (
	InstallMethodHomebrew InstallMethod = "homebrew"
	InstallMethodScoop    InstallMethod = "scoop"
	InstallMethodGo       InstallMethod = "go"
	InstallMethodBinary   InstallMethod = "binary"
)

var (
	detectedMethod     InstallMethod
	detectedMethodOnce sync.Once
)

// DetectInstallMethod determines how sidecar was installed.
// Checks Homebrew first, then Go bin directories, falls back to binary.
// Result is cached for the lifetime of the process.
func DetectInstallMethod() InstallMethod {
	detectedMethodOnce.Do(func() {
		detectedMethod = detectInstallMethod()
	})
	return detectedMethod
}

func detectInstallMethod() InstallMethod {
	// Check Homebrew (macOS/Linux)
	if isHomebrewInstall() {
		return InstallMethodHomebrew
	}

	// Check Scoop (Windows)
	if isScoopInstall() {
		return InstallMethodScoop
	}

	// Check if binary is in a Go bin directory
	if isGoInstall() {
		return InstallMethodGo
	}

	return InstallMethodBinary
}

// isHomebrewInstall checks if sidecar was installed via Homebrew.
func isHomebrewInstall() bool {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return false
	}
	_, err := exec.LookPath("brew")
	if err != nil {
		return false
	}
	out, err := exec.Command("brew", "list", "--formula", "marcus/tap/sidecar").CombinedOutput()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// isGoInstall checks if the current binary is in a Go bin directory.
func isGoInstall() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return false
	}

	dir := filepath.Dir(exe)

	// Check GOBIN
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		if dir == gobin {
			return true
		}
	}

	// Check GOPATH/bin
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		if dir == filepath.Join(gopath, "bin") {
			return true
		}
	}

	// Check default ~/go/bin
	if home, err := os.UserHomeDir(); err == nil {
		if dir == filepath.Join(home, "go", "bin") {
			return true
		}
	}

	// Heuristic: path contains /go/bin/
	if strings.Contains(exe, string(filepath.Separator)+"go"+string(filepath.Separator)+"bin"+string(filepath.Separator)) {
		return true
	}

	return false
}

// isScoopInstall checks if sidecar was installed via Scoop (Windows package manager).
func isScoopInstall() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return false
	}
	// Scoop installs to ~/scoop/apps/<app>/current/ or ~/scoop/shims/
	normalized := strings.ToLower(filepath.ToSlash(exe))
	return strings.Contains(normalized, "/scoop/shims/") ||
		strings.Contains(normalized, "/scoop/apps/")
}
