//go:build !windows

package mux

// Default returns the platform-default multiplexer.
// On Unix systems (Linux, macOS), this is tmux.
func Default() Multiplexer {
	return &TmuxMux{}
}
