//go:build windows

package workspace

import "fmt"

// setEnvCmd returns a PowerShell command to set an environment variable.
// Uses $env:VAR = "value" syntax native to PowerShell.
func setEnvCmd(key, value string) string {
	return fmt.Sprintf("$env:%s = \"%s\"", key, value)
}

// unsetEnvCmd returns a PowerShell command to remove an environment variable.
// Uses Remove-Item Env:\VAR syntax.
func unsetEnvCmd(key string) string {
	return fmt.Sprintf("Remove-Item Env:\\%s -ErrorAction SilentlyContinue", key)
}

// envInlinePrefix returns the inline env-var-setting prefix for a command.
// On Windows/PowerShell: $env:KEY = "val"; cmd
func envInlinePrefix(key, value string) string {
	return fmt.Sprintf("$env:%s = \"%s\"; ", key, value)
}
