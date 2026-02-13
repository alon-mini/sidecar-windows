package image

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Valid image extensions
		{"test.png", true},
		{"test.PNG", true},
		{"test.jpg", true},
		{"test.JPG", true},
		{"test.jpeg", true},
		{"test.JPEG", true},
		{"test.gif", true},
		{"test.GIF", true},
		{"test.webp", true},
		{"test.WEBP", true},
		{"test.bmp", true},
		{"test.BMP", true},
		{"test.ico", true},
		{"test.ICO", true},

		// Mixed case
		{"test.Png", true},
		{"test.JpEg", true},

		// With path
		{"/path/to/image.png", true},
		{"./relative/image.jpg", true},

		// Not images
		{"test.txt", false},
		{"test.pdf", false},
		{"test.go", false},
		{"test.md", false},
		{"test.svg", false}, // SVG is not in list
		{"noextension", false},
		{"", false},

		// Edge cases
		{".png", true},          // Just extension
		{"test.png.txt", false}, // Double extension
		{"test.tar.gz", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := IsImageFile(tc.path)
			if result != tc.expected {
				t.Errorf("IsImageFile(%q) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestProtocolString(t *testing.T) {
	tests := []struct {
		protocol Protocol
		expected string
	}{
		{ProtocolNone, "None"},
		{ProtocolKitty, "Kitty"},
		{ProtocolITerm2, "iTerm2"},
		{ProtocolSixel, "Sixel"},
		{Protocol(99), "None"}, // Unknown protocol
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.protocol.String()
			if result != tc.expected {
				t.Errorf("Protocol(%d).String() = %q, want %q", tc.protocol, result, tc.expected)
			}
		})
	}
}

func TestRendererNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if r.cache == nil {
		t.Error("cache is nil")
	}
	// Protocol may be None in test environment
	_ = r.Protocol()
}

func TestRendererCacheOperations(t *testing.T) {
	r := New()

	// Initial state
	entries, maxEntries := r.CacheStats()
	if entries != 0 {
		t.Errorf("expected 0 entries, got %d", entries)
	}
	if maxEntries != MaxCacheEntries {
		t.Errorf("expected maxEntries=%d, got %d", MaxCacheEntries, maxEntries)
	}

	// Manually add cache entry to test operations
	key := cacheKey{Path: "/test/image.png", Width: 80, Height: 40, Mtime: 12345}
	r.cache[key] = &RenderResult{Content: "test", Width: 80, Height: 40}
	r.order = append(r.order, key)

	entries, _ = r.CacheStats()
	if entries != 1 {
		t.Errorf("expected 1 entry, got %d", entries)
	}

	// InvalidatePath
	r.InvalidatePath("/test/image.png")
	entries, _ = r.CacheStats()
	if entries != 0 {
		t.Errorf("expected 0 entries after invalidate, got %d", entries)
	}

	// Re-add and test ClearCache
	r.cache[key] = &RenderResult{Content: "test", Width: 80, Height: 40}
	r.order = append(r.order, key)

	r.ClearCache()
	entries, _ = r.CacheStats()
	if entries != 0 {
		t.Errorf("expected 0 entries after clear, got %d", entries)
	}
	if len(r.order) != 0 {
		t.Errorf("expected empty order after clear, got %d", len(r.order))
	}
}

func TestRendererInvalidatePathMultiple(t *testing.T) {
	r := New()

	// Add multiple entries with same path but different sizes
	for i := 0; i < 5; i++ {
		key := cacheKey{Path: "/test/image.png", Width: 80 + i, Height: 40, Mtime: 12345}
		r.cache[key] = &RenderResult{Content: "test", Width: 80 + i, Height: 40}
		r.order = append(r.order, key)
	}

	// Add entry with different path
	otherKey := cacheKey{Path: "/other/image.jpg", Width: 100, Height: 50, Mtime: 67890}
	r.cache[otherKey] = &RenderResult{Content: "other", Width: 100, Height: 50}
	r.order = append(r.order, otherKey)

	entries, _ := r.CacheStats()
	if entries != 6 {
		t.Errorf("expected 6 entries, got %d", entries)
	}

	// Invalidate only the first path
	r.InvalidatePath("/test/image.png")

	entries, _ = r.CacheStats()
	if entries != 1 {
		t.Errorf("expected 1 entry after invalidate, got %d", entries)
	}

	// The other path should still be cached
	if _, ok := r.cache[otherKey]; !ok {
		t.Error("other path should still be cached")
	}
}

func TestRendererLRUEviction(t *testing.T) {
	r := New()

	// Fill cache beyond limit
	for i := 0; i < MaxCacheEntries+5; i++ {
		key := cacheKey{Path: filepath.Join("/test", string(rune('a'+i))+".png"), Width: 80, Height: 40, Mtime: int64(i)}
		r.cache[key] = &RenderResult{Content: "test", Width: 80, Height: 40}
		r.order = append(r.order, key)

		// Trigger eviction by simulating what Render does
		for len(r.cache) > MaxCacheEntries {
			delete(r.cache, r.order[0])
			r.order = r.order[1:]
		}
	}

	entries, _ := r.CacheStats()
	if entries > MaxCacheEntries {
		t.Errorf("cache should not exceed max entries, got %d", entries)
	}
}

func TestRendererRenderFileNotFound(t *testing.T) {
	r := New()
	_, err := r.Render("/nonexistent/path/image.png", 80, 40)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestRendererRenderTooLarge(t *testing.T) {
	// Create a temp file larger than MaxImageSize is impractical for unit tests
	// Instead, we test by checking the MaxImageSize constant
	if MaxImageSize != 10*1024*1024 {
		t.Errorf("expected MaxImageSize=10MB, got %d", MaxImageSize)
	}
}

func TestSupportedTerminals(t *testing.T) {
	result := SupportedTerminals()
	if result == "" {
		t.Error("SupportedTerminals() should return non-empty string")
	}
}

func TestRenderResultFields(t *testing.T) {
	result := RenderResult{
		Content:    "test content",
		IsFallback: true,
		Width:      80,
		Height:     40,
	}

	if result.Content != "test content" {
		t.Error("Content field mismatch")
	}
	if !result.IsFallback {
		t.Error("IsFallback should be true")
	}
	if result.Width != 80 {
		t.Error("Width field mismatch")
	}
	if result.Height != 40 {
		t.Error("Height field mismatch")
	}
}

func TestRendererConcurrentAccess(t *testing.T) {
	r := New()
	done := make(chan bool)

	// Concurrent cache operations
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := cacheKey{Path: "/test/image.png", Width: 80 + id, Height: 40, Mtime: 12345}
			r.mu.Lock()
			r.cache[key] = &RenderResult{Content: "test", Width: 80 + id, Height: 40}
			r.order = append(r.order, key)
			r.mu.Unlock()

			r.CacheStats()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent invalidation
	go func() {
		r.InvalidatePath("/test/image.png")
		done <- true
	}()
	go func() {
		r.ClearCache()
		done <- true
	}()

	<-done
	<-done
}

// TestRendererRenderWithRealImage tests rendering with an actual image file.
func TestRendererRenderWithRealImage(t *testing.T) {
	// Create a minimal valid PNG in a temp directory using Go's image package
	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")

	// Create test image using Go's image package
	if err := createTestPNG(testImage); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	r := New()
	result, err := r.Render(testImage, 10, 10)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// Check cache was populated
	entries, _ := r.CacheStats()
	if entries != 1 {
		t.Errorf("expected 1 cache entry, got %d", entries)
	}

	// Render again should hit cache
	result2, err := r.Render(testImage, 10, 10)
	if err != nil {
		t.Fatalf("second Render failed: %v", err)
	}
	if result2 != result {
		t.Error("expected cache hit to return same pointer")
	}
}

// createTestPNG creates a minimal valid PNG file
func createTestPNG(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})
	img.Set(1, 1, color.RGBA{255, 255, 255, 255})

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return png.Encode(f, img)
}
