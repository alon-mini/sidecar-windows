package features

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/marcus/sidecar/internal/config"
)

// setupTestConfig sets up a temp config path for tests that write to config.
func setupTestConfig(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	config.SetTestConfigPath(filepath.Join(tmpDir, "config.json"))
	t.Cleanup(config.ResetTestConfigPath)
}

func TestIsEnabled_DefaultValue(t *testing.T) {
	// Reset global manager
	globalManager = nil

	// Without initialization, should return defaults
	if IsEnabled(TmuxInteractiveInput.Name) != TmuxInteractiveInput.Default {
		t.Errorf("expected default value %v for %s", TmuxInteractiveInput.Default, TmuxInteractiveInput.Name)
	}
}

func TestIsEnabled_UnknownFeature(t *testing.T) {
	globalManager = nil
	if IsEnabled("unknown_feature") != false {
		t.Error("unknown features should default to false")
	}
}

func TestIsEnabled_ConfigOverride(t *testing.T) {
	cfg := config.Default()
	cfg.Features.Flags["tmux_interactive_input"] = true

	Init(cfg)
	defer func() { globalManager = nil }()

	if !IsEnabled("tmux_interactive_input") {
		t.Error("config override should enable feature")
	}
}

func TestIsEnabled_CLIOverrideTakesPrecedence(t *testing.T) {
	cfg := config.Default()
	cfg.Features.Flags["tmux_interactive_input"] = false

	Init(cfg)
	defer func() { globalManager = nil }()

	// CLI override should take precedence
	SetOverride("tmux_interactive_input", true)

	if !IsEnabled("tmux_interactive_input") {
		t.Error("CLI override should take precedence over config")
	}
}

func TestList(t *testing.T) {
	cfg := config.Default()
	Init(cfg)
	defer func() { globalManager = nil }()

	list := List()
	if len(list) == 0 {
		t.Error("List should return at least one feature")
	}

	// Verify known feature is in list
	if _, ok := list[TmuxInteractiveInput.Name]; !ok {
		t.Errorf("expected %s in list", TmuxInteractiveInput.Name)
	}
}

func TestListAll(t *testing.T) {
	all := ListAll()
	if len(all) == 0 {
		t.Error("ListAll should return at least one feature")
	}

	// Verify metadata is present
	found := false
	for _, f := range all {
		if f.Name == TmuxInteractiveInput.Name {
			found = true
			if f.Description == "" {
				t.Error("feature should have description")
			}
		}
	}
	if !found {
		t.Errorf("expected %s in ListAll", TmuxInteractiveInput.Name)
	}
}

func TestSetOverride_NilManager(t *testing.T) {
	globalManager = nil
	// Should not panic
	SetOverride("test", true)
}

func TestSetEnabled_NilManager(t *testing.T) {
	globalManager = nil
	err := SetEnabled("test", true)
	if err != ErrNotInitialized {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

func TestSetEnabled_UpdatesConfig(t *testing.T) {
	setupTestConfig(t)

	cfg := config.Default()
	Init(cfg)
	defer func() { globalManager = nil }()

	// SetEnabled modifies config and saves to temp file
	_ = SetEnabled("tmux_interactive_input", true)

	// Verify the config was updated
	if !cfg.Features.Flags["tmux_interactive_input"] {
		t.Error("SetEnabled should update config")
	}
}

func TestSetEnabled_InitializesNilFlagsMap(t *testing.T) {
	setupTestConfig(t)

	cfg := config.Default()
	cfg.Features.Flags = nil // Force nil map
	Init(cfg)
	defer func() { globalManager = nil }()

	_ = SetEnabled("tmux_interactive_input", true)

	if cfg.Features.Flags == nil {
		t.Error("SetEnabled should initialize nil Flags map")
	}
}

func TestIsKnownFeature(t *testing.T) {
	if !IsKnownFeature("tmux_interactive_input") {
		t.Error("tmux_interactive_input should be a known feature")
	}
	if IsKnownFeature("unknown_feature") {
		t.Error("unknown_feature should not be a known feature")
	}
}

func TestListAllReturnsCopy(t *testing.T) {
	original := ListAll()
	originalLen := len(original)

	// Modify the returned slice
	if len(original) > 0 {
		original[0].Name = "modified"
	}

	// Get fresh copy
	fresh := ListAll()
	if len(fresh) != originalLen {
		t.Error("ListAll should return consistent length")
	}
	if fresh[0].Name == "modified" {
		t.Error("ListAll should return a copy, not the original slice")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cfg := config.Default()
	Init(cfg)
	defer func() { globalManager = nil }()

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent reads and writes
	for i := 0; i < goroutines; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = IsEnabled("tmux_interactive_input")
		}()
		go func() {
			defer wg.Done()
			SetOverride("tmux_interactive_input", true)
		}()
		go func() {
			defer wg.Done()
			_ = List()
		}()
	}
	wg.Wait()
}

func TestConcurrentSetEnabled(t *testing.T) {
	setupTestConfig(t)

	cfg := config.Default()
	Init(cfg)
	defer func() { globalManager = nil }()

	var wg sync.WaitGroup
	const goroutines = 20

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(val bool) {
			defer wg.Done()
			// SetEnabled may fail on Save, but shouldn't race
			_ = SetEnabled("tmux_interactive_input", val)
		}(i%2 == 0)
	}
	wg.Wait()
}
