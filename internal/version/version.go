package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	repoOwner   = "marcus"
	repoName    = "sidecar"
	tdRepoOwner = "marcus"
	tdRepoName  = "td"
	apiURL      = "https://api.github.com/repos/%s/%s/releases/latest"
)

// Release represents a GitHub release response.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// CheckResult holds the result of a version check.
type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	UpdateURL      string
	ReleaseNotes   string
	HasUpdate      bool
	Error          error
}

// Check fetches the latest release from GitHub and compares versions.
func Check(currentVersion string) CheckResult {
	return CheckRepo(repoOwner, repoName, currentVersion)
}

// CheckTd fetches the latest td release from GitHub and compares versions.
func CheckTd(currentVersion string) CheckResult {
	return CheckRepo(tdRepoOwner, tdRepoName, currentVersion)
}

// CheckRepo fetches the latest release for a repo and compares versions.
func CheckRepo(owner, repo, currentVersion string) CheckResult {
	result := CheckResult{CurrentVersion: currentVersion}

	if isDevelopmentVersion(currentVersion) {
		return result
	}

	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf(apiURL, owner, repo)

	resp, err := client.Get(url)
	if err != nil {
		result.Error = err
		return result
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("github api: %s", resp.Status)
		return result
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		result.Error = err
		return result
	}

	result.LatestVersion = release.TagName
	result.UpdateURL = release.HTMLURL
	result.ReleaseNotes = release.Body
	result.HasUpdate = isNewer(release.TagName, currentVersion)

	return result
}

// isDevelopmentVersion returns true for non-release versions.
func isDevelopmentVersion(v string) bool {
	if v == "" || v == "unknown" || v == "devel" {
		return true
	}
	if strings.HasPrefix(v, "devel+") {
		return true
	}
	return false
}
