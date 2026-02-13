package mux

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// TmuxMux implements Multiplexer for tmux (Unix).
type TmuxMux struct{}

func (t *TmuxMux) Name() string { return "tmux" }

func (t *TmuxMux) IsInstalled() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func (t *TmuxMux) NewSession(name, startDir string, env map[string]string) error {
	args := []string{"new-session", "-d", "-s", name}
	if startDir != "" {
		args = append(args, "-c", startDir)
	}
	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}
	return exec.Command("tmux", args...).Run()
}

func (t *TmuxMux) KillSession(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

func (t *TmuxMux) HasSession(name string) bool {
	err := exec.Command("tmux", "has-session", "-t", name).Run()
	return err == nil
}

func (t *TmuxMux) ListSessions(prefix string) ([]string, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		return nil, err
	}
	var sessions []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && (prefix == "" || strings.HasPrefix(line, prefix)) {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}

func (t *TmuxMux) SendKeys(target string, keys ...string) error {
	args := []string{"send-keys", "-t", target}
	args = append(args, keys...)
	return exec.Command("tmux", args...).Run()
}

func (t *TmuxMux) SendLiteral(target, text string) error {
	args := []string{"send-keys", "-t", target, "-l", text}
	return exec.Command("tmux", args...).Run()
}

func (t *TmuxMux) SendSGRMouse(target string, button, col, row int, release bool) error {
	releaseChar := "M"
	if release {
		releaseChar = "m"
	}
	// Build SGR mouse escape sequence: ESC [ < button ; col ; row M/m
	seq := fmt.Sprintf("\x1b[<%d;%d;%d%s", button, col, row, releaseChar)
	return exec.Command("tmux", "send-keys", "-t", target, "-l", seq).Run()
}

func (t *TmuxMux) ResizePane(target string, width, height int) error {
	return exec.Command("tmux", "resize-pane", "-t", target,
		"-x", fmt.Sprint(width), "-y", fmt.Sprint(height)).Run()
}

func (t *TmuxMux) SetWindowSizeManual(session string) error {
	return exec.Command("tmux", "set-option", "-t", session,
		"window-size", "manual").Run()
}

func (t *TmuxMux) QueryPaneSize(target string) (width, height int, ok bool) {
	out, err := exec.Command("tmux", "display-message", "-t", target,
		"-p", "#{pane_width} #{pane_height}").Output()
	if err != nil {
		return 0, 0, false
	}
	_, scanErr := fmt.Sscanf(strings.TrimSpace(string(out)), "%d %d", &width, &height)
	return width, height, scanErr == nil
}

func (t *TmuxMux) CapturePaneOutput(target string, scrollback int) (string, error) {
	args := []string{"capture-pane", "-t", target, "-p", "-e"}
	if scrollback > 0 {
		args = append(args, "-S", fmt.Sprintf("-%d", scrollback))
	}
	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (t *TmuxMux) LoadBuffer(text string) error {
	cmd := exec.Command("tmux", "load-buffer", "-")
	cmd.Stdin = bytes.NewReader([]byte(text))
	return cmd.Run()
}

func (t *TmuxMux) PasteBuffer(target string) error {
	return exec.Command("tmux", "paste-buffer", "-t", target, "-d", "-p").Run()
}
