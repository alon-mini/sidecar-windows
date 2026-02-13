package fdmonitor

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	// DefaultWarningThreshold is the FD count that triggers a warning.
	DefaultWarningThreshold = 200
	// DefaultCriticalThreshold is the FD count that triggers a critical warning.
	DefaultCriticalThreshold = 500
	// MinCheckInterval prevents checking too frequently.
	MinCheckInterval = 10 * time.Second
)

var (
	lastCheck         time.Time
	lastCount         int
	lastCheckMu       sync.Mutex
	warningThreshold  = DefaultWarningThreshold
	criticalThreshold = DefaultCriticalThreshold
)

// SetThresholds configures the warning and critical thresholds.
func SetThresholds(warning, critical int) {
	warningThreshold = warning
	criticalThreshold = critical
}

// Thresholds returns the current warning and critical thresholds.
func Thresholds() (warning, critical int) {
	return warningThreshold, criticalThreshold
}

// Count returns the current number of open file descriptors for this process.
// On non-Linux/macOS platforms, returns -1.
func Count() int {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return -1
	}

	pid := os.Getpid()
	var fdDir string
	switch runtime.GOOS {
	case "darwin":
		// On macOS, /dev/fd shows the current process's FDs
		fdDir = "/dev/fd"
	case "linux":
		fdDir = fmt.Sprintf("/proc/%d/fd", pid)
	}

	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return -1
	}
	return len(entries)
}

// Check checks the current FD count and logs a warning if it exceeds thresholds.
// To avoid log spam, checks are rate-limited to MinCheckInterval.
// Returns the current FD count and whether a warning was logged.
func Check(logger *slog.Logger) (count int, warned bool) {
	lastCheckMu.Lock()
	defer lastCheckMu.Unlock()

	if time.Since(lastCheck) < MinCheckInterval {
		return lastCount, false
	}

	count = Count()
	if count < 0 {
		return count, false
	}

	lastCheck = time.Now()
	lastCount = count

	if count >= criticalThreshold {
		if logger != nil {
			logger.Warn("critical FD count", "count", count, "threshold", criticalThreshold)
		}
		return count, true
	}
	if count >= warningThreshold {
		if logger != nil {
			logger.Warn("high FD count", "count", count, "threshold", warningThreshold)
		}
		return count, true
	}

	return count, false
}

// CheckWithContext checks FD count and includes context about what triggered the check.
func CheckWithContext(logger *slog.Logger, context string) (count int, warned bool) {
	count, warned = Check(logger)
	if warned && logger != nil {
		logger.Debug("FD check context", "context", context, "count", count)
	}
	return count, warned
}

// DebugInfo returns detailed FD information for debugging.
// On non-Linux/macOS platforms, returns an empty map.
func DebugInfo() map[string]int {
	info := make(map[string]int)

	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return info
	}

	pid := os.Getpid()
	var fdDir string
	switch runtime.GOOS {
	case "darwin":
		fdDir = "/dev/fd"
	case "linux":
		fdDir = fmt.Sprintf("/proc/%d/fd", pid)
	}

	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return info
	}

	for _, e := range entries {
		fdPath := filepath.Join(fdDir, e.Name())
		target, err := os.Readlink(fdPath)
		if err != nil {
			continue
		}

		// Categorize by file type/pattern
		var category string
		switch {
		case target == "pipe" || target == "anon_inode:[pipe]":
			category = "pipe"
		case target == "socket" || len(target) > 0 && target[0] == '[':
			category = "socket"
		case filepath.Ext(target) == ".jsonl":
			category = "jsonl"
		case filepath.Ext(target) == ".json":
			category = "json"
		case filepath.Ext(target) == ".sqlite" || filepath.Ext(target) == ".db":
			category = "database"
		case isDirectory(target):
			category = "directory"
		default:
			category = "file"
		}
		info[category]++
	}

	return info
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
