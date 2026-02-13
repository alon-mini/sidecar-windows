package mux

// Multiplexer abstracts terminal multiplexer operations (tmux, psmux, etc.).
// This interface decouples the application from any specific multiplexer backend,
// enabling cross-platform support: tmux on Unix, psmux on Windows.
type Multiplexer interface {
	// Name returns the multiplexer executable name (e.g. "tmux", "psmux").
	Name() string

	// IsInstalled checks if the multiplexer binary is available on PATH.
	IsInstalled() bool

	// Session management
	NewSession(name, startDir string, env map[string]string) error
	KillSession(name string) error
	HasSession(name string) bool
	ListSessions(prefix string) ([]string, error)

	// Key/input
	SendKeys(target string, keys ...string) error
	SendLiteral(target, text string) error
	SendSGRMouse(target string, button, col, row int, release bool) error

	// Pane management
	ResizePane(target string, width, height int) error
	SetWindowSizeManual(session string) error
	QueryPaneSize(target string) (width, height int, ok bool)
	CapturePaneOutput(target string, scrollback int) (string, error)

	// Clipboard/paste
	LoadBuffer(text string) error
	PasteBuffer(target string) error
}
