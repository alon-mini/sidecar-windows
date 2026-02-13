package monitor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcus/td/internal/db"
)

func TestSharedDB_SingleConnection(t *testing.T) {
	// Create a temp directory with a test database
	tmpDir, err := os.MkdirTemp("", "td-dbpool-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize the database
	todosDir := filepath.Join(tmpDir, ".todos")
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create initial database
	database, err := db.Initialize(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	database.Close()

	// Clear the pool before testing
	clearDBPool()

	// Get shared DB twice
	db1, err := getSharedDB(tmpDir)
	if err != nil {
		t.Fatalf("first getSharedDB failed: %v", err)
	}

	db2, err := getSharedDB(tmpDir)
	if err != nil {
		t.Fatalf("second getSharedDB failed: %v", err)
	}

	// Both should return the same pointer
	if db1 != db2 {
		t.Error("getSharedDB returned different pointers for same path")
	}

	// Check reference count
	resolvedDir := db.ResolveBaseDir(tmpDir)
	dbPool.mu.RLock()
	entry := dbPool.conns[resolvedDir]
	refs := entry.refs
	dbPool.mu.RUnlock()

	if refs != 2 {
		t.Errorf("expected refs=2, got refs=%d", refs)
	}

	// Release once
	if err := releaseSharedDB(tmpDir); err != nil {
		t.Fatalf("first releaseSharedDB failed: %v", err)
	}

	dbPool.mu.RLock()
	entry = dbPool.conns[resolvedDir]
	refs = entry.refs
	dbPool.mu.RUnlock()

	if refs != 1 {
		t.Errorf("expected refs=1 after first release, got refs=%d", refs)
	}

	// Release again - should close and remove
	if err := releaseSharedDB(tmpDir); err != nil {
		t.Fatalf("second releaseSharedDB failed: %v", err)
	}

	dbPool.mu.RLock()
	_, exists := dbPool.conns[resolvedDir]
	dbPool.mu.RUnlock()

	if exists {
		t.Error("expected connection to be removed from pool after all releases")
	}
}

func TestSharedDB_DifferentPaths(t *testing.T) {
	// Create two temp directories with test databases
	tmpDir1, err := os.MkdirTemp("", "td-dbpool-test1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "td-dbpool-test2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir2)

	// Initialize both databases
	for _, dir := range []string{tmpDir1, tmpDir2} {
		todosDir := filepath.Join(dir, ".todos")
		if err := os.MkdirAll(todosDir, 0755); err != nil {
			t.Fatal(err)
		}
		database, err := db.Initialize(dir)
		if err != nil {
			t.Fatal(err)
		}
		database.Close()
	}

	// Clear the pool before testing
	clearDBPool()

	// Get shared DBs for different paths
	db1, err := getSharedDB(tmpDir1)
	if err != nil {
		t.Fatalf("getSharedDB for dir1 failed: %v", err)
	}

	db2, err := getSharedDB(tmpDir2)
	if err != nil {
		t.Fatalf("getSharedDB for dir2 failed: %v", err)
	}

	// Should be different pointers
	if db1 == db2 {
		t.Error("getSharedDB returned same pointer for different paths")
	}

	// Clean up
	releaseSharedDB(tmpDir1)
	releaseSharedDB(tmpDir2)
}
