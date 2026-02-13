package cursor

import (
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/marcus/sidecar/internal/adapter"
)

// NewWatcher creates a watcher for Cursor CLI session changes.
// It watches only the workspace directory - fsnotify on macOS reports events
// for files in subdirectories when watching the parent (td-0f0e68).
func NewWatcher(workspaceDir string) (<-chan adapter.Event, io.Closer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}

	// Only watch workspace directory to reduce FD count (td-0f0e68)
	// fsnotify on macOS propagates events from subdirectories
	if err := watcher.Add(workspaceDir); err != nil {
		_ = watcher.Close()
		return nil, nil, err
	}

	events := make(chan adapter.Event, 32)

	go func() {
		// Debounce timer
		var debounceTimer *time.Timer
		debounceDelay := 100 * time.Millisecond

		// Protect against sending to closed channel from timer callback
		var closed bool
		var mu sync.Mutex

		defer func() {
			mu.Lock()
			closed = true
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			mu.Unlock()
			close(events)
		}()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Watch for store.db changes or new session directories
				if strings.HasSuffix(event.Name, "store.db") ||
					strings.HasSuffix(event.Name, "store.db-wal") {
					// Capture event for closure to avoid race condition
					capturedEvent := event

					mu.Lock()
					// Debounce rapid events
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(debounceDelay, func() {
						mu.Lock()
						defer mu.Unlock()

						if closed {
							return
						}

						// Extract session ID from path (use capturedEvent to avoid race)
						sessionID := filepath.Base(filepath.Dir(capturedEvent.Name))

						var eventType adapter.EventType
						switch {
						case capturedEvent.Op&fsnotify.Create != 0:
							eventType = adapter.EventSessionCreated
						case capturedEvent.Op&fsnotify.Write != 0:
							eventType = adapter.EventMessageAdded
						case capturedEvent.Op&fsnotify.Remove != 0:
							return
						default:
							eventType = adapter.EventSessionUpdated
						}

						select {
						case events <- adapter.Event{
							Type:      eventType,
							SessionID: sessionID,
						}:
						default:
							// Channel full, drop event
						}
					})
					mu.Unlock()
				}
				// Don't add per-session directory watches - parent watch suffices (td-0f0e68)

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
				// Log error but continue watching
			}
		}
	}()

	return events, watcher, nil
}
