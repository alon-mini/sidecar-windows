package gitstatus

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// doCommit executes the git commit asynchronously.
func (p *Plugin) doCommit(message string) tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		hash, err := ExecuteCommit(workDir, message)
		if err != nil {
			return CommitErrorMsg{Err: err}
		}
		// Extract first line as subject
		subject := strings.Split(message, "\n")[0]
		return CommitSuccessMsg{Hash: hash, Subject: subject}
	}
}

// doAmend executes git commit --amend asynchronously.
func (p *Plugin) doAmend(message string) tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		hash, err := ExecuteAmend(workDir, message)
		if err != nil {
			return CommitErrorMsg{Err: err}
		}
		subject := strings.Split(message, "\n")[0]
		return CommitSuccessMsg{Hash: hash, Subject: subject}
	}
}

// doPush executes a git push asynchronously.
func (p *Plugin) doPush(force bool) tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		output, err := ExecutePush(workDir, force)
		if err != nil {
			return PushErrorMsg{Err: err}
		}
		return PushSuccessMsg{Output: output}
	}
}

// doPushForce executes a force push with lease.
func (p *Plugin) doPushForce() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		output, err := ExecutePushForce(workDir)
		if err != nil {
			return PushErrorMsg{Err: err}
		}
		return PushSuccessMsg{Output: output}
	}
}

// doPushSetUpstream executes a push with upstream tracking.
func (p *Plugin) doPushSetUpstream() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		output, err := ExecutePushSetUpstream(workDir)
		if err != nil {
			return PushErrorMsg{Err: err}
		}
		return PushSuccessMsg{Output: output}
	}
}

// canPush returns true if there are commits that can be pushed.
func (p *Plugin) canPush() bool {
	return p.pushStatus != nil && p.pushStatus.CanPush()
}

// doStashPush stashes all current changes.
func (p *Plugin) doStashPush() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		err := StashPush(workDir)
		return StashResultMsg{Operation: "push", Err: err}
	}
}

// doStashPop pops the latest stash.
func (p *Plugin) doStashPop() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		err := StashPop(workDir)
		return StashResultMsg{Operation: "pop", Ref: "stash@{0}", Err: err}
	}
}

// doStashApply applies the latest stash without removing it.
func (p *Plugin) doStashApply() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		err := StashApply(workDir, "stash@{0}")
		return StashResultMsg{Operation: "apply", Ref: "stash@{0}", Err: err}
	}
}

// doFetch fetches from remote.
func (p *Plugin) doFetch() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		output, err := ExecuteFetch(workDir)
		if err != nil {
			return FetchErrorMsg{Err: err}
		}
		return FetchSuccessMsg{Output: output}
	}
}

// doPull pulls from remote (default merge strategy).
func (p *Plugin) doPull() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		output, err := ExecutePull(workDir)
		if err != nil {
			return PullErrorMsg{Err: err, Strategy: "merge"}
		}
		return PullSuccessMsg{Output: output}
	}
}

// doPullRebase pulls from remote with rebase.
func (p *Plugin) doPullRebase() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		output, err := ExecutePullRebase(workDir)
		if err != nil {
			return PullErrorMsg{Err: err, Strategy: "rebase"}
		}
		return PullSuccessMsg{Output: output}
	}
}

// doPullFFOnly pulls from remote with fast-forward only.
func (p *Plugin) doPullFFOnly() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		output, err := ExecutePullFFOnly(workDir)
		if err != nil {
			return PullErrorMsg{Err: err, Strategy: "ff-only"}
		}
		return PullSuccessMsg{Output: output}
	}
}

// doPullAutostash pulls from remote with rebase and autostash.
func (p *Plugin) doPullAutostash() tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		output, err := ExecutePullAutostash(workDir)
		if err != nil {
			return PullErrorMsg{Err: err, Strategy: "autostash"}
		}
		return PullSuccessMsg{Output: output}
	}
}

// doAbortPull aborts the current merge or rebase.
func (p *Plugin) doAbortPull() tea.Cmd {
	workDir := p.repoRoot
	conflictType := p.pullConflictType
	return func() tea.Msg {
		var err error
		if conflictType == "rebase" {
			err = AbortRebase(workDir)
		} else {
			err = AbortMerge(workDir)
		}
		if err != nil {
			return PullErrorMsg{Err: err}
		}
		return PullAbortedMsg{}
	}
}

// canPull returns true if pull is possible (has remote, not detached HEAD).
func (p *Plugin) canPull() bool {
	if p.pushStatus != nil && p.pushStatus.DetachedHead {
		return false
	}
	return HasRemote(p.repoRoot)
}

// doDiscard executes the git discard operation.
func (p *Plugin) doDiscard(entry *FileEntry) tea.Cmd {
	workDir := p.repoRoot
	return func() tea.Msg {
		var err error
		if entry.Status == StatusUntracked {
			// Remove untracked file
			err = DiscardUntracked(workDir, entry.Path)
		} else if entry.Staged {
			// Unstage and restore staged file
			err = DiscardStaged(workDir, entry.Path)
		} else {
			// Restore modified file
			err = DiscardModified(workDir, entry.Path)
		}
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return RefreshDoneMsg{}
	}
}
