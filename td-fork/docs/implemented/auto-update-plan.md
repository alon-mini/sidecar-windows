# Auto-Update Plan for `td`

## Summary
Non-blocking, silent-failure auto-updates via GitHub Releases API. Check once per day, prompt after command completes, update via `go install`.

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/update/update.go` | Create - version check & update logic |
| `cmd/root.go` | Modify - add async check hook |
| `cmd/system.go` | Modify - add `update` command |

## Implementation Steps

### Step 1: Create `internal/update/update.go`

```go
package update

import (
    "context"
    "encoding/json"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

const (
    repoAPI      = "https://api.github.com/repos/marcus/td/releases/latest"
    modulePath   = "github.com/marcus/td@latest"
    checkTimeout = 3 * time.Second
    cacheFile    = "update-check"
    cacheTTL     = 24 * time.Hour
)

type UpdateInfo struct {
    Available  bool
    CurrentVer string
    LatestVer  string
}

// CheckForUpdate runs async, returns channel with result (nil on any error)
func CheckForUpdate(currentVersion string) <-chan *UpdateInfo

// ShouldCheck returns true if cache is stale or missing (>24h since last check)
func ShouldCheck(configDir string) bool

// SaveCheckTime writes current timestamp to cache file
func SaveCheckTime(configDir string)

// PerformUpdate runs `go install github.com/marcus/td@latest`
// Returns error for caller to handle (not printed)
func PerformUpdate() error
```

**Key behaviors:**
- `CheckForUpdate`: Starts goroutine, 3s timeout, returns nil on any error (network, parse, timeout)
- Skip check if version contains "dev" or "devel"
- Compare versions with semver (strip "v" prefix, compare major.minor.patch)
- Cache stored in `~/.config/td/update-check` (just a timestamp file)

### Step 2: Modify `cmd/root.go`

Add at package level:
```go
var updateChan <-chan *update.UpdateInfo
```

Add to `init()` after version is set:
```go
// Start async update check (non-blocking)
configDir := filepath.Join(os.UserHomeDir(), ".config", "td")
if update.ShouldCheck(configDir) {
    updateChan = update.CheckForUpdate(version)
}
```

Add `PersistentPostRun` to rootCmd:
```go
PersistentPostRun: func(cmd *cobra.Command, args []string) {
    if updateChan == nil {
        return
    }
    select {
    case info := <-updateChan:
        if info != nil && info.Available {
            promptForUpdate(info)
        }
    default:
        // Check not complete, don't block
    }
}
```

Add prompt function:
```go
func promptForUpdate(info *UpdateInfo) {
    fmt.Printf("\nUpdate available: %s → %s\n", info.CurrentVer, info.LatestVer)
    fmt.Print("Update now? [y/N] ")

    var response string
    fmt.Scanln(&response)

    if strings.ToLower(response) == "y" {
        fmt.Print("Updating...")
        if err := update.PerformUpdate(); err == nil {
            fmt.Printf(" done! Restart td to use %s\n", info.LatestVer)
        } else {
            fmt.Println(" failed. Try `td update` manually.")
        }
    }

    // Save check time regardless of user response
    update.SaveCheckTime(configDir)
}
```

### Step 3: Add `td update` Command in `cmd/system.go`

```go
var updateCmd = &cobra.Command{
    Use:     "update",
    Short:   "Update td to latest version",
    GroupID: "system",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Checking for updates...")

        // Force check (ignore cache)
        info := <-update.CheckForUpdate(version)
        if info == nil {
            fmt.Println("Could not check for updates. Are you online?")
            return
        }

        if !info.Available {
            fmt.Printf("Already on latest version (%s)\n", version)
            return
        }

        fmt.Printf("Updating %s → %s...\n", info.CurrentVer, info.LatestVer)
        if err := update.PerformUpdate(); err != nil {
            fmt.Printf("Update failed: %v\n", err)
            return
        }

        fmt.Printf("Updated to %s. Restart td to use new version.\n", info.LatestVer)
        update.SaveCheckTime(configDir)
    },
}
```

Register in init(): `rootCmd.AddCommand(updateCmd)`

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Offline/timeout | Return nil, no prompt |
| GitHub API error | Return nil, no prompt |
| Rate limited (403) | Return nil, no prompt |
| Dev version | Skip check entirely |
| go install fails | Brief error only if user-initiated |

## Testing Notes

- Test with `TD_UPDATE_CHECK_DISABLE=1` env var to skip checks during tests
- Mock HTTP responses for unit tests
- Integration test: verify `go install` command is correct (don't actually run)
