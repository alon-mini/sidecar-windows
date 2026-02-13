package claudecode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcus/sidecar/internal/adapter/cache"
	"github.com/marcus/sidecar/internal/adapter/testutil"
)

// Benchmark targets (td-336ee0):
// - Full parse (1MB): <50ms
// - Full parse (10MB): <500ms
// - Incremental (1MB + 1KB): <10ms
// - Cache hit: <1ms

func BenchmarkMessages_FullParse_1MB(b *testing.B) {
	tmpDir := b.TempDir()
	sessionFile := filepath.Join(tmpDir, "session.jsonl")

	// ~1MB file with 500 message pairs, ~1KB each
	messageCount := testutil.ApproximateMessageCount(1*1024*1024, 1024)
	if err := testutil.GenerateClaudeCodeSessionFile(sessionFile, messageCount, 1024); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	info, _ := os.Stat(sessionFile)
	b.Logf("Generated file size: %d bytes (%d messages)", info.Size(), messageCount*2)

	a := New()
	a.sessionIndex["bench-session-001"] = sessionFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear cache to force full parse
		a.msgCache = cache.New[messageCacheEntry](msgCacheMaxEntries)
		_, err := a.Messages("bench-session-001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_FullParse_10MB(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping 10MB benchmark in short mode")
	}

	tmpDir := b.TempDir()
	sessionFile := filepath.Join(tmpDir, "session.jsonl")

	// ~10MB file
	messageCount := testutil.ApproximateMessageCount(10*1024*1024, 1024)
	if err := testutil.GenerateClaudeCodeSessionFile(sessionFile, messageCount, 1024); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	info, _ := os.Stat(sessionFile)
	b.Logf("Generated file size: %d bytes (%d messages)", info.Size(), messageCount*2)

	a := New()
	a.sessionIndex["bench-session-001"] = sessionFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[messageCacheEntry](msgCacheMaxEntries)
		_, err := a.Messages("bench-session-001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_CacheHit(b *testing.B) {
	tmpDir := b.TempDir()
	sessionFile := filepath.Join(tmpDir, "session.jsonl")

	// ~1MB file
	messageCount := testutil.ApproximateMessageCount(1*1024*1024, 1024)
	if err := testutil.GenerateClaudeCodeSessionFile(sessionFile, messageCount, 1024); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.sessionIndex["bench-session-001"] = sessionFile

	// Warm the cache
	_, err := a.Messages("bench-session-001")
	if err != nil {
		b.Fatalf("initial Messages failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := a.Messages("bench-session-001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_IncrementalParse(b *testing.B) {
	tmpDir := b.TempDir()
	sessionFile := filepath.Join(tmpDir, "session.jsonl")

	// Start with ~1MB file
	messageCount := testutil.ApproximateMessageCount(1*1024*1024, 1024)
	if err := testutil.GenerateClaudeCodeSessionFile(sessionFile, messageCount, 1024); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.sessionIndex["bench-session-001"] = sessionFile

	// Warm the cache
	_, err := a.Messages("bench-session-001")
	if err != nil {
		b.Fatalf("initial Messages failed: %v", err)
	}

	// Pre-generate content to append
	appendContent := []byte(`{"type":"user","uuid":"msg-append","sessionId":"bench-session-001","timestamp":"2024-01-15T12:00:00Z","message":{"role":"user","content":[{"type":"text","text":"appended message"}]}}` + "\n")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Append a small message to trigger incremental parse
		f, _ := os.OpenFile(sessionFile, os.O_APPEND|os.O_WRONLY, 0644)
		_, _ = f.Write(appendContent)
		_ = f.Close()
		b.StartTimer()

		_, err := a.Messages("bench-session-001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkSessions_50Files(b *testing.B) {
	tmpDir := b.TempDir()

	// Claude Code stores sessions at: ~/.claude/projects/<hash>/<session>.jsonl
	// The hash is based on the project path: /foo/bar -> -foo-bar
	// Claude Code replaces "/", ".", and "_" with "-"
	projectRoot := "/bench/project"
	projectHash := strings.ReplaceAll(projectRoot, "/", "-")
	projectHash = strings.ReplaceAll(projectHash, ".", "-")
	projectHash = strings.ReplaceAll(projectHash, "_", "-")
	sessionsDir := filepath.Join(tmpDir, projectHash)
	_ = os.MkdirAll(sessionsDir, 0755)

	// Create 50 session files of ~100KB each
	for i := 0; i < 50; i++ {
		sessionFile := filepath.Join(sessionsDir, formatSessionFile(i))

		messageCount := testutil.ApproximateMessageCount(100*1024, 512)
		if err := testutil.GenerateClaudeCodeSessionFile(sessionFile, messageCount, 512); err != nil {
			b.Fatalf("failed to generate test file %d: %v", i, err)
		}
	}

	a := New()
	a.projectsDir = tmpDir

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear caches
		a.sessionIndex = make(map[string]string)
		a.metaCache = make(map[string]sessionMetaCacheEntry)
		_, err := a.Sessions(projectRoot)
		if err != nil {
			b.Fatalf("Sessions failed: %v", err)
		}
	}
}

func formatSessionFile(index int) string {
	return "session_" + formatSessionID(index) + ".jsonl"
}

func formatSessionID(index int) string {
	// Format like "01ABCDEF..."
	return "bench" + string(rune('A'+index%26)) + string(rune('0'+index%10))
}

// BenchmarkMessages_Allocs specifically tracks allocations for optimization.
func BenchmarkMessages_Allocs(b *testing.B) {
	tmpDir := b.TempDir()
	sessionFile := filepath.Join(tmpDir, "session.jsonl")

	// Small file to focus on allocation patterns
	messageCount := 100
	if err := testutil.GenerateClaudeCodeSessionFile(sessionFile, messageCount, 256); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.sessionIndex["bench-session-001"] = sessionFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[messageCacheEntry](msgCacheMaxEntries)
		_, _ = a.Messages("bench-session-001")
	}
}
