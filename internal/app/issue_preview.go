package app

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// IssueSearchResult holds a single search result from td search.
type IssueSearchResult struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Type     string `json:"type"`
	Priority string `json:"priority"`
}

// tdSearchResultWrapper wraps td search JSON output: {"Issue": {...}, "Score": N}.
type tdSearchResultWrapper struct {
	Issue struct {
		IssueSearchResult
		UpdatedAt string `json:"updated_at"`
	} `json:"Issue"`
	Score int `json:"Score"`
}

// IssueSearchResultMsg carries search results back to the app.
type IssueSearchResultMsg struct {
	Query   string
	Results []IssueSearchResult
	Error   error
}

// issueSearchCmd runs `td search <query> --json -n 50` asynchronously.
// When includeClosed is false, filters to non-closed statuses.
// workDir sets the command's working directory so td uses the correct project database.
func issueSearchCmd(workDir, query string, includeClosed bool) tea.Cmd {
	return func() tea.Msg {
		args := []string{"search", query, "--json", "-n", "50"}
		if !includeClosed {
			args = append(args, "-s", "open", "-s", "in_progress", "-s", "blocked", "-s", "in_review")
		}
		cmd := exec.Command("td", args...)
		cmd.Dir = workDir
		out, err := cmd.Output()
		if err != nil {
			return IssueSearchResultMsg{Query: query, Error: err}
		}
		var wrappers []tdSearchResultWrapper
		if err := json.Unmarshal(out, &wrappers); err != nil {
			return IssueSearchResultMsg{Query: query, Error: err}
		}
		// Sort by updated_at descending (most recently updated first).
		sort.Slice(wrappers, func(i, j int) bool {
			ti, _ := time.Parse(time.RFC3339Nano, wrappers[i].Issue.UpdatedAt)
			tj, _ := time.Parse(time.RFC3339Nano, wrappers[j].Issue.UpdatedAt)
			return ti.After(tj)
		})
		results := make([]IssueSearchResult, len(wrappers))
		for i, w := range wrappers {
			results[i] = w.Issue.IssueSearchResult
		}
		return IssueSearchResultMsg{Query: query, Results: results}
	}
}

// IssuePreviewData holds lightweight issue data fetched via CLI.
type IssuePreviewData struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Points      int      `json:"points"`
	Description string   `json:"description"`
	ParentID    string   `json:"parent_id"`
	Labels      []string `json:"labels"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// IssuePreviewResultMsg carries fetched issue data back to the app.
type IssuePreviewResultMsg struct {
	Data  *IssuePreviewData
	Error error
}

// OpenFullIssueMsg is broadcast to plugins to open the full rich issue view.
// Currently handled by the TD monitor plugin via monitor.OpenIssueByIDMsg.
type OpenFullIssueMsg struct {
	IssueID string
}

// fetchIssuePreviewCmd runs `td show <id> -f json` and returns the result.
// workDir sets the command's working directory so td uses the correct project database.
func fetchIssuePreviewCmd(workDir, issueID string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("td", "show", issueID, "-f", "json")
		cmd.Dir = workDir
		out, err := cmd.Output()
		if err != nil {
			// stdout may contain "ERROR: <message>" from td CLI
			if msg := extractTdError(string(out)); msg != "" {
				return IssuePreviewResultMsg{Error: fmt.Errorf("%s", msg)}
			}
			// stderr may contain usage help + "Error: <message>" on last line
			if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
				if msg := extractTdError(string(exitErr.Stderr)); msg != "" {
					return IssuePreviewResultMsg{Error: fmt.Errorf("%s", msg)}
				}
			}
			return IssuePreviewResultMsg{Error: fmt.Errorf("issue %q not found", issueID)}
		}
		var data IssuePreviewData
		if err := json.Unmarshal(out, &data); err != nil {
			return IssuePreviewResultMsg{Error: err}
		}
		return IssuePreviewResultMsg{Data: &data}
	}
}

// extractTdError finds the last "ERROR: ..." or "Error: ..." line in td output.
func extractTdError(output string) string {
	for _, line := range reverseLines(strings.TrimSpace(output)) {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "error:") {
			return strings.TrimSpace(line[len("error:"):])
		}
	}
	return ""
}

func reverseLines(s string) []string {
	lines := strings.Split(s, "\n")
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}
