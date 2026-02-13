package geminicli

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/marcus/sidecar/internal/adapter"
)

// sessionIDPattern extracts sessionId field from partial JSON
var sessionIDPattern = regexp.MustCompile(`"sessionId"\s*:\s*"([^"]+)"`)

// NewWatcher creates a watcher for Gemini CLI session changes.
func NewWatcher(chatsDir string) (<-chan adapter.Event, io.Closer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}

	if err := watcher.Add(chatsDir); err != nil {
		_ = watcher.Close()
		return nil, nil, err
	}

	events := make(chan adapter.Event, 32)

	go func() {
		// Debounce timer
		var debounceTimer *time.Timer
		var lastEvent fsnotify.Event
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

				// Only watch session-*.json files
				name := filepath.Base(event.Name)
				if !strings.HasPrefix(name, "session-") || !strings.HasSuffix(name, ".json") {
					continue
				}

				mu.Lock()
				lastEvent = event

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

					sessionID := extractSessionID(lastEvent.Name)
					if sessionID == "" {
						return
					}

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
					case events <- adapter.Event{
						Type:      eventType,
						SessionID: sessionID,
					}:
					default:
						// Channel full, drop event
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

// extractSessionID reads the session file and extracts the sessionId field.
// First attempts with a 2048-byte buffer, falling back to full file read
// if the buffer was full but no match was found (td-8d9c18c2).
func extractSessionID(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	// Try reading first 2048 bytes - sessionId is usually near the start
	const bufSize = 2048
	buf := make([]byte, bufSize)
	n, err := file.Read(buf)
	if err != nil || n == 0 {
		return ""
	}

	// Extract sessionId using regex
	if match := sessionIDPattern.FindSubmatch(buf[:n]); match != nil {
		return validateSessionID(string(match[1]))
	}

	// Fallback: if buffer was full, sessionId might be beyond buffer boundary
	if n == bufSize {
		// Read entire file as fallback
		_, _ = file.Seek(0, 0)
		data, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		if match := sessionIDPattern.FindSubmatch(data); match != nil {
			return validateSessionID(string(match[1]))
		}
	}

	return ""
}

// validateSessionID checks that sessionId looks valid (non-empty, reasonable length).
func validateSessionID(id string) string {
	// SessionIDs should be non-empty and reasonably sized (typically UUIDs or similar)
	if len(id) == 0 || len(id) > 128 {
		return ""
	}
	return id
}
