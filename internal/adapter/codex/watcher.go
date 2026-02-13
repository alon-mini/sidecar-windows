package codex

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/marcus/sidecar/internal/adapter"
)

// NewWatcher creates a watcher for Codex session changes.
// Only watches root and month directories to reduce FD count (td-0f0e68).
// fsnotify on macOS propagates events from subdirectories.
func NewWatcher(root string) (<-chan adapter.Event, io.Closer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}

	// Watch root for new year directories
	if err := watcher.Add(root); err != nil {
		_ = watcher.Close()
		return nil, nil, err
	}

	// Watch only month directories (not recursive) to reduce FD count (td-0f0e68)
	for _, monthDir := range recentSessionDirs(root) {
		if info, err := os.Stat(monthDir); err == nil && info.IsDir() {
			_ = watcher.Add(monthDir)
		}
	}

	events := make(chan adapter.Event, 32)

	go func() {
		var debounceTimer *time.Timer
		var lastEvent fsnotify.Event
		debounceDelay := 200 * time.Millisecond // Increased from 100ms (td-11c31ccd)

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

				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						// Only watch year/month directories, not day dirs (td-0f0e68)
						rel, _ := filepath.Rel(root, event.Name)
						depth := len(strings.Split(rel, string(filepath.Separator)))
						if depth <= 2 { // year or year/month
							_ = watcher.Add(event.Name)
						}
						// Scan for sessions in new directory
						scanNewDirForSessions(event.Name, events)
						continue
					}
				}

				if !strings.HasSuffix(event.Name, ".jsonl") {
					continue
				}

				mu.Lock()
				lastEvent = event
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					mu.Lock()
					defer mu.Unlock()

					if closed {
						return
					}

					sessionID := strings.TrimSuffix(filepath.Base(lastEvent.Name), ".jsonl")
					var eventType adapter.EventType
					switch {
					case lastEvent.Op&fsnotify.Create != 0:
						eventType = adapter.EventSessionCreated
					case lastEvent.Op&fsnotify.Write != 0:
						eventType = adapter.EventMessageAdded
					case lastEvent.Op&fsnotify.Remove != 0:
						return
					default:
						eventType = adapter.EventSessionUpdated
					}

					select {
					case events <- adapter.Event{Type: eventType, SessionID: sessionID}:
					default:
					}
				})
				mu.Unlock()

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return events, watcher, nil
}

// recentSessionDirs returns directories for current and previous months (td-ae05cd6a).
// Codex organizes sessions by date: sessions/YYYY/MM/DD/session.jsonl
func recentSessionDirs(root string) []string {
	now := time.Now()
	dirs := make([]string, 0, 2)

	// Current month
	dirs = append(dirs, filepath.Join(root, now.Format("2006"), now.Format("01")))

	// Previous month (for sessions started last month)
	prev := now.AddDate(0, -1, 0)
	dirs = append(dirs, filepath.Join(root, prev.Format("2006"), prev.Format("01")))

	return dirs
}

// scanNewDirForSessions checks for JSONL files in a newly created directory
// and sends events for any found. This handles the race condition where a
// directory and its files are created before the watcher is added (td-ba9f8c12).
func scanNewDirForSessions(dir string, events chan<- adapter.Event) {
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".jsonl") {
			sessionID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
			select {
			case events <- adapter.Event{Type: adapter.EventSessionCreated, SessionID: sessionID}:
			default:
				// Channel full, skip
			}
		}
		return nil
	})
}
