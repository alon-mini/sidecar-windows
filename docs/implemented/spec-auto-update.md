# Plan: In-App Update Feature for Sidecar

## Summary

Add an interactive "Update" button to the diagnostics modal (`!`) that updates both sidecar and td when either has an available update. Uses `go install`, shows progress, and prompts user to restart after success.

## User Requirements

- Single button updates both sidecar and td (when updates available)
- Go is required (use `go install`)
- Prompt user to restart after success (no auto-restart)
- Button should support mouse clicks (not just keyboard)

## Implementation Steps

### Step 1: Add Message Types

**File**: `internal/app/commands.go`

Add after existing message types:

```go
// UpdateSuccessMsg signals update completed
type UpdateSuccessMsg struct {
    SidecarUpdated    bool
    TdUpdated         bool
    NewSidecarVersion string
    NewTdVersion      string
}

// UpdateErrorMsg signals update failed
type UpdateErrorMsg struct {
    Step string // "sidecar", "td", or "check"
    Err  error
}
```

### Step 2: Add Model State Fields

**File**: `internal/app/model.go`

Add import for mouse package and add to Model struct (after `tdVersionInfo`):

```go
import "github.com/marcus/sidecar/internal/mouse"

// Update feature state
updateButtonFocus  bool
updateInProgress   bool
updateError        string
needsRestart       bool
updateButtonBounds mouse.Rect // button position for mouse clicks
```

Add helper method:

```go
func (m *Model) hasUpdatesAvailable() bool {
    if m.updateAvailable != nil {
        return true
    }
    if m.tdVersionInfo != nil && m.tdVersionInfo.HasUpdate && m.tdVersionInfo.Installed {
        return true
    }
    return false
}
```

### Step 3: Add Update Execution Function

**File**: `internal/app/model.go` (or new `internal/app/updater.go`)

```go
func (m *Model) doUpdate() tea.Cmd {
    sidecarUpdate := m.updateAvailable
    tdUpdate := m.tdVersionInfo

    return func() tea.Msg {
        // Check Go is available
        if _, err := exec.LookPath("go"); err != nil {
            return UpdateErrorMsg{Step: "check", Err: fmt.Errorf("go not found in PATH")}
        }

        var sidecarUpdated, tdUpdated bool
        var newSidecarVersion, newTdVersion string

        // Update sidecar
        if sidecarUpdate != nil {
            args := []string{"install", "-ldflags",
                fmt.Sprintf("-X main.Version=%s", sidecarUpdate.LatestVersion),
                fmt.Sprintf("github.com/marcus/sidecar/cmd/sidecar@%s", sidecarUpdate.LatestVersion)}
            cmd := exec.Command("go", args...)
            if output, err := cmd.CombinedOutput(); err != nil {
                return UpdateErrorMsg{Step: "sidecar", Err: fmt.Errorf("%v: %s", err, output)}
            }
            sidecarUpdated = true
            newSidecarVersion = sidecarUpdate.LatestVersion
        }

        // Update td
        if tdUpdate != nil && tdUpdate.HasUpdate && tdUpdate.Installed {
            cmd := exec.Command("go", "install",
                fmt.Sprintf("github.com/marcus/td@%s", tdUpdate.LatestVersion))
            if output, err := cmd.CombinedOutput(); err != nil {
                return UpdateErrorMsg{Step: "td", Err: fmt.Errorf("%v: %s", err, output)}
            }
            tdUpdated = true
            newTdVersion = tdUpdate.LatestVersion
        }

        return UpdateSuccessMsg{sidecarUpdated, tdUpdated, newSidecarVersion, newTdVersion}
    }
}
```

### Step 4: Handle Keys in Diagnostics Modal

**File**: `internal/app/update.go`

In key handling for diagnostics modal (around line 335), add:

```go
if m.showDiagnostics && !m.updateInProgress {
    switch msg.String() {
    case "tab":
        if m.hasUpdatesAvailable() {
            m.updateButtonFocus = !m.updateButtonFocus
        }
        return m, nil
    case "enter":
        if m.updateButtonFocus {
            m.updateInProgress = true
            m.updateError = ""
            return m, m.doUpdate()
        }
    case "u":
        if m.hasUpdatesAvailable() {
            m.updateInProgress = true
            m.updateError = ""
            return m, m.doUpdate()
        }
        return m, nil
    }
}
```

### Step 4b: Handle Mouse Clicks on Update Button

**File**: `internal/app/update.go`

Add field to Model for button bounds:

```go
// In model.go
updateButtonBounds mouse.Rect // bounds of update button in screen coords
```

Modify mouse handler (around line 58) - instead of ignoring all mouse events for diagnostics, handle update button clicks:

```go
case tea.MouseMsg:
    // ... existing palette handling ...

    // Handle diagnostics modal mouse events
    if m.showDiagnostics {
        if m.hasUpdatesAvailable() && !m.updateInProgress && !m.needsRestart {
            // Check if click is on update button
            if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
                if m.updateButtonBounds.Contains(msg.X, msg.Y) {
                    m.updateInProgress = true
                    m.updateError = ""
                    return m, m.doUpdate()
                }
            }
        }
        return m, nil // Ignore other mouse events in modal
    }
    // ... rest of existing mouse handling ...
```

### Step 5: Handle Update Messages

**File**: `internal/app/update.go`

In main Update() switch, add cases:

```go
case UpdateSuccessMsg:
    m.updateInProgress = false
    m.needsRestart = true
    if msg.SidecarUpdated {
        m.updateAvailable = nil
    }
    if msg.TdUpdated && m.tdVersionInfo != nil {
        m.tdVersionInfo.HasUpdate = false
    }
    m.ShowToast("Update complete! Restart sidecar to use new version", 10*time.Second)
    return m, nil

case UpdateErrorMsg:
    m.updateInProgress = false
    m.updateError = fmt.Sprintf("Failed to update %s: %s", msg.Step, msg.Err)
    m.ShowToast("Update failed: "+msg.Err.Error(), 5*time.Second)
    m.statusIsError = true
    return m, nil
```

### Step 6: Modify Diagnostics View + Button Bounds

**File**: `internal/app/view.go`

Replace the static "Update:" section (lines 601-605) with interactive button.
**Important**: Calculate button bounds for mouse clicks. The button position depends on modal centering.

In `renderDiagnosticsOverlay()`, after calling `buildDiagnosticsContent()`:

```go
// Calculate modal position (centered)
modalWidth := lipgloss.Width(content)
modalHeight := lipgloss.Height(content)
modalX := (m.width - modalWidth) / 2
modalY := (m.height - modalHeight) / 2

// Calculate button position within modal content
// Button is at line N in content, column 2 (for "  " indent)
if m.hasUpdatesAvailable() && !m.updateInProgress && !m.needsRestart {
    buttonLineInModal := countLinesBeforeButton(content) // helper to find button line
    buttonX := modalX + 2 + 1 // modal padding + indent + border
    buttonY := modalY + buttonLineInModal + 1 // +1 for border
    buttonWidth := 8 // " Update "
    m.updateButtonBounds = mouse.Rect{X: buttonX, Y: buttonY, W: buttonWidth, H: 1}
}
```

In `buildDiagnosticsContent()`, replace curl command section:

```go
// Show update controls if any updates available
if m.updateAvailable != nil || (m.tdVersionInfo != nil && m.tdVersionInfo.HasUpdate) {
    b.WriteString("\n")

    if m.updateInProgress {
        b.WriteString("  ")
        b.WriteString(styles.StatusInProgress.Render("● "))
        b.WriteString("Installing update...")
        b.WriteString("\n")
    } else if m.needsRestart {
        b.WriteString("  ")
        b.WriteString(styles.StatusCompleted.Render("✓ "))
        b.WriteString("Update complete. ")
        b.WriteString(styles.StatusModified.Render("Restart sidecar to use new version"))
        b.WriteString("\n")
    } else {
        // Show Update button (click or press u)
        buttonStyle := styles.Button
        if m.updateButtonFocus {
            buttonStyle = styles.ButtonFocused
        }
        label := m.buildUpdateLabel()
        b.WriteString("  ")
        b.WriteString(buttonStyle.Render(" Update "))
        b.WriteString("  ")
        b.WriteString(styles.Muted.Render(label))
        b.WriteString("  ")
        b.WriteString(styles.KeyHint.Render("u"))
        b.WriteString("\n")
    }

    if m.updateError != "" {
        b.WriteString("  ")
        b.WriteString(styles.StatusBlocked.Render("✗ " + m.updateError))
        b.WriteString("\n")
    }
}
```

Add helper:

```go
func (m Model) buildUpdateLabel() string {
    var parts []string
    if m.updateAvailable != nil {
        parts = append(parts, "sidecar "+m.updateAvailable.LatestVersion)
    }
    if m.tdVersionInfo != nil && m.tdVersionInfo.HasUpdate && m.tdVersionInfo.Installed {
        parts = append(parts, "td "+m.tdVersionInfo.LatestVersion)
    }
    return strings.Join(parts, " + ")
}
```

### Step 7: Reset State on Modal Close

**File**: `internal/app/update.go`

When toggling diagnostics off:

```go
m.updateButtonFocus = false
// Keep updateError visible if user reopens
```

## Files to Modify

1. `internal/app/commands.go` - Add UpdateSuccessMsg, UpdateErrorMsg
2. `internal/app/model.go` - Add state fields (incl. updateButtonBounds), hasUpdatesAvailable(), doUpdate()
3. `internal/app/update.go` - Add key handling, mouse click handling, message handlers
4. `internal/app/view.go` - Replace static curl command with interactive button, calculate button bounds

## Testing

1. Build with old version: `go build -ldflags "-X main.Version=v0.0.1" ./cmd/sidecar`
2. Run and press `!` to see update button
3. Test keyboard: Press `u` or Tab+Enter to trigger update
4. Test mouse: Click on the Update button in modal
5. Verify success message and restart prompt
6. Test error case by disconnecting network mid-update
