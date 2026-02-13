//go:build !windows

package workspace

import "fmt"

// setEnvCmd returns a shell command to set an environment variable.
// On Unix, this uses the export syntax.
func setEnvCmd(key, value string) string {
	return "export " + key + "=" + shellQuote(value)
}

// unsetEnvCmd returns a shell command to unset an environment variable.
// On Unix, this uses the unset command.
func unsetEnvCmd(key string) string {
	return "unset " + key
}

// envInlinePrefix returns the inline env-var-setting prefix for a command.
// On Unix: export KEY='val' && cmd
func envInlinePrefix(key, value string) string {
	return fmt.Sprintf("export %s=%s && ", key, shellQuote(value))
}
