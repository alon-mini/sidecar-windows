package cmd

import (
	"github.com/marcus/td/internal/features"
	"github.com/spf13/cobra"
)

// SyncFeatureHooks exposes optional sync hooks that can be registered by
// sync-specific command files. These hooks are gated by feature flags.
type SyncFeatureHooks struct {
	OnStartup         func(commandName string)
	OnAfterMutation   func()
	IsMutatingCommand func(commandName string) bool
}

var syncFeatureHooks SyncFeatureHooks

// RegisterSyncFeatureHooks registers sync-related hooks.
func RegisterSyncFeatureHooks(hooks SyncFeatureHooks) {
	syncFeatureHooks = hooks
}

// AddFeatureGatedCommand registers a command only when its feature is enabled
// for the current process (env overrides + defaults).
func AddFeatureGatedCommand(featureName string, command *cobra.Command) {
	if features.IsEnabledForProcess(featureName) {
		rootCmd.AddCommand(command)
	}
}

func runGatedSyncStartupHook(cmd *cobra.Command) {
	if syncFeatureHooks.OnStartup == nil {
		return
	}
	if !features.IsEnabled(getBaseDir(), features.SyncAutosync.Name) {
		return
	}
	syncFeatureHooks.OnStartup(resolveCommandName(cmd))
}

func runGatedSyncMutationHook(cmd *cobra.Command) {
	if syncFeatureHooks.OnAfterMutation == nil {
		return
	}
	if !features.IsEnabled(getBaseDir(), features.SyncAutosync.Name) {
		return
	}

	commandName := resolveCommandName(cmd)
	if syncFeatureHooks.IsMutatingCommand != nil && !syncFeatureHooks.IsMutatingCommand(commandName) {
		return
	}

	syncFeatureHooks.OnAfterMutation()
}

func resolveCommandName(cmd *cobra.Command) string {
	name := cmd.Name()
	if cmd.Parent() != nil && cmd.Parent().Name() != "td" {
		name = cmd.Parent().Name()
	}
	return name
}
