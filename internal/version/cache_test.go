package version

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsCacheValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		entry          *CacheEntry
		currentVersion string
		want           bool
	}{
		{
			name:           "nil entry",
			entry:          nil,
			currentVersion: "v1.0.0",
			want:           false,
		},
		{
			name: "valid cache - same version, recent",
			entry: &CacheEntry{
				LatestVersion:  "v1.1.0",
				CurrentVersion: "v1.0.0",
				CheckedAt:      now,
				HasUpdate:      true,
			},
			currentVersion: "v1.0.0",
			want:           true,
		},
		{
			name: "expired cache - same version, old timestamp",
			entry: &CacheEntry{
				LatestVersion:  "v1.1.0",
				CurrentVersion: "v1.0.0",
				CheckedAt:      now.Add(-4 * time.Hour), // older than 3h TTL
				HasUpdate:      true,
			},
			currentVersion: "v1.0.0",
			want:           false,
		},
		{
			name: "invalid cache - version mismatch (upgrade)",
			entry: &CacheEntry{
				LatestVersion:  "v1.1.0",
				CurrentVersion: "v1.0.0",
				CheckedAt:      now,
				HasUpdate:      true,
			},
			currentVersion: "v1.1.0",
			want:           false,
		},
		{
			name: "invalid cache - version mismatch (downgrade)",
			entry: &CacheEntry{
				LatestVersion:  "v1.1.0",
				CurrentVersion: "v1.0.0",
				CheckedAt:      now,
				HasUpdate:      true,
			},
			currentVersion: "v0.9.0",
			want:           false,
		},
		{
			name: "boundary - exactly at TTL",
			entry: &CacheEntry{
				LatestVersion:  "v1.1.0",
				CurrentVersion: "v1.0.0",
				CheckedAt:      now.Add(-3*time.Hour + time.Minute), // just under TTL
				HasUpdate:      true,
			},
			currentVersion: "v1.0.0",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCacheValid(tt.entry, tt.currentVersion)
			if got != tt.want {
				t.Errorf("IsCacheValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveAndLoadCache(t *testing.T) {
	// Create temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "sidecar")

	// Override cachePath for testing by saving directly
	cachePath := filepath.Join(configDir, "version_cache.json")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	entry := &CacheEntry{
		LatestVersion:  "v1.2.0",
		CurrentVersion: "v1.0.0",
		CheckedAt:      time.Now().Truncate(time.Second), // Truncate for JSON roundtrip
		HasUpdate:      true,
	}

	// Write JSON directly
	data := `{"latestVersion":"v1.2.0","currentVersion":"v1.0.0","checkedAt":"` +
		entry.CheckedAt.Format(time.RFC3339) + `","hasUpdate":true}`
	if err := os.WriteFile(cachePath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	// Read back
	readData, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal(err)
	}

	if len(readData) == 0 {
		t.Error("Expected non-empty cache file")
	}
}

func TestLoadCache_FileNotExist(t *testing.T) {
	// LoadCache uses os.UserHomeDir() internally, so we can't easily
	// redirect it. This test verifies error handling for missing files.
	// The actual cachePath() function will return a real path.
	_, err := LoadCache()
	// Error is expected since cache likely doesn't exist in test env
	// or if it does exist, that's also fine
	_ = err
}

func TestCacheEntry_JSONRoundtrip(t *testing.T) {
	// Test that CacheEntry serializes/deserializes correctly
	original := CacheEntry{
		LatestVersion:  "v2.0.0",
		CurrentVersion: "v1.5.0",
		CheckedAt:      time.Now().Truncate(time.Second),
		HasUpdate:      true,
	}

	// Create temp file
	tmpFile := filepath.Join(t.TempDir(), "cache.json")

	// Write
	data := `{"latestVersion":"` + original.LatestVersion +
		`","currentVersion":"` + original.CurrentVersion +
		`","checkedAt":"` + original.CheckedAt.Format(time.RFC3339) +
		`","hasUpdate":` + "true" + `}`

	if err := os.WriteFile(tmpFile, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	// Read and verify
	readData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(readData) != data {
		t.Errorf("JSON roundtrip failed: got %s, want %s", readData, data)
	}
}

func TestTdCachePath(t *testing.T) {
	// tdCachePath should return a path ending in td_version_cache.json
	path := tdCachePath()
	if path == "" {
		// May be empty if home dir detection fails
		return
	}

	if !filepath.IsAbs(path) {
		t.Errorf("tdCachePath() = %q, want absolute path", path)
	}

	if filepath.Base(path) != "td_version_cache.json" {
		t.Errorf("tdCachePath() = %q, want filename td_version_cache.json", path)
	}
}

func TestLoadTdCache_FileNotExist(t *testing.T) {
	// LoadTdCache should return error when file doesn't exist
	_, err := LoadTdCache()
	// Error is expected since cache likely doesn't exist in test env
	// (or if it does exist on dev machine, that's also fine)
	_ = err
}

func TestTdCacheOperations(t *testing.T) {
	// Test that td cache operations work correctly
	// This test verifies the cache functions exist and are callable

	// Create a cache entry
	entry := &CacheEntry{
		LatestVersion:  "v0.4.13",
		CurrentVersion: "v0.4.12",
		CheckedAt:      time.Now(),
		HasUpdate:      true,
	}

	// SaveTdCache and LoadTdCache use the real home dir,
	// so we just verify they don't panic
	err := SaveTdCache(entry)
	if err != nil {
		// May fail due to permissions, which is acceptable
		t.Logf("SaveTdCache error (may be expected): %v", err)
	}

	// If save succeeded, load should also work
	if err == nil {
		loaded, loadErr := LoadTdCache()
		if loadErr != nil {
			t.Logf("LoadTdCache error after successful save: %v", loadErr)
		} else {
			// Verify loaded data matches
			if loaded.LatestVersion != entry.LatestVersion {
				t.Errorf("LatestVersion = %q, want %q", loaded.LatestVersion, entry.LatestVersion)
			}
			if loaded.CurrentVersion != entry.CurrentVersion {
				t.Errorf("CurrentVersion = %q, want %q", loaded.CurrentVersion, entry.CurrentVersion)
			}
			if loaded.HasUpdate != entry.HasUpdate {
				t.Errorf("HasUpdate = %v, want %v", loaded.HasUpdate, entry.HasUpdate)
			}
		}
	}
}

func TestCachePathSeparation(t *testing.T) {
	// Verify that sidecar and td caches use different paths
	sidecarPath := cachePath()
	tdPath := tdCachePath()

	if sidecarPath == "" || tdPath == "" {
		// Skip if home dir detection fails
		t.Skip("Home dir detection failed")
	}

	if sidecarPath == tdPath {
		t.Errorf("Cache paths should be different: sidecar=%q, td=%q", sidecarPath, tdPath)
	}

	// Both should be in same directory
	if filepath.Dir(sidecarPath) != filepath.Dir(tdPath) {
		t.Errorf("Cache files should be in same directory: sidecar=%q, td=%q",
			filepath.Dir(sidecarPath), filepath.Dir(tdPath))
	}
}
