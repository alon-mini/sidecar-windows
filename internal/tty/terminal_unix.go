//go:build !windows

package tty

// IsWindowsTerminal always returns false on non-Windows platforms.
func IsWindowsTerminal() bool { return false }

// IsConhost always returns false on non-Windows platforms.
func IsConhost() bool { return false }

// SupportsNerdFonts always returns true on Unix platforms.
// Both macOS Terminal/iTerm2 and modern Linux terminals support Nerd Fonts.
func SupportsNerdFonts() bool { return true }

// Supports24BitColor always returns true on Unix platforms.
// Most modern terminals support 24-bit color.
func Supports24BitColor() bool { return true }
