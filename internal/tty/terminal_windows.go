//go:build windows

package tty

import "os"

// IsWindowsTerminal returns true if running inside Windows Terminal.
// Windows Terminal sets the WT_SESSION environment variable.
func IsWindowsTerminal() bool {
	return os.Getenv("WT_SESSION") != ""
}

// IsConhost returns true if running in the legacy Windows console host.
// This is the fallback when not running in Windows Terminal.
func IsConhost() bool {
	return !IsWindowsTerminal()
}

// SupportsNerdFonts returns true if the terminal likely supports Nerd Fonts.
// Windows Terminal supports custom fonts including Nerd Fonts.
// Legacy conhost has limited font support.
func SupportsNerdFonts() bool {
	return IsWindowsTerminal()
}

// Supports24BitColor returns true if the terminal supports 24-bit (true) color.
// Windows Terminal supports 24-bit color; conhost may not.
func Supports24BitColor() bool {
	return IsWindowsTerminal()
}
