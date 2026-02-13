package workspace

import (
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcus/sidecar/internal/app"
)

// openInBrowser opens the URL in the default browser.
func openInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		default:
			return nil
		}
		_ = cmd.Start()
		return nil
	}
}

// openInGitTab opens the selected worktree in the git status tab.
// It switches to the worktree directory and focuses the git-status plugin.
func (p *Plugin) openInGitTab(wt *Worktree) tea.Cmd {
	if wt == nil {
		return nil
	}
	// Sequence: switch to worktree first (triggers plugin reinit), then focus git-status plugin.
	// Must use Sequence not Batch to avoid deadlock during concurrent plugin reinit + fork/exec.
	return tea.Sequence(
		app.SwitchWorktree(wt.Path),
		app.FocusPlugin("git-status"),
	)
}
