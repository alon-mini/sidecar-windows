//go:build windows

package mux

// Default returns the platform-default multiplexer.
// On Windows, this is psmux (PowerShell multiplexer).
func Default() Multiplexer {
	return &PsmuxMux{}
}
