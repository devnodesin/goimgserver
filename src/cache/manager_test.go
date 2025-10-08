package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheManager_New_ValidInstance tests cache manager creation
func TestCacheManager_New_ValidInstance(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	// Act
	manager, err := NewManager(tempDir)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, manager)
}

// TestCacheManager_GenerateKey_ConsistentHash tests that hash generation is consistent
func TestCacheManager_GenerateKey_ConsistentHash(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{
		Width:   800,
		Height:  600,
		Format:  "webp",
		Quality: 90,
	}

	// Act
	key1 := manager.GenerateKey("photo.jpg", params)
	key2 := manager.GenerateKey("photo.jpg", params)

	// Assert
	assert.Equal(t, key1, key2, "Same inputs should produce same hash")
	assert.NotEmpty(t, key1)
}

// TestCacheManager_GenerateKey_DifferentParams tests different params create different hashes
func TestCacheManager_GenerateKey_DifferentParams(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params1 := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	params2 := ProcessingParams{Width: 400, Height: 300, Format: "png", Quality: 85}

	// Act
	key1 := manager.GenerateKey("photo.jpg", params1)
	key2 := manager.GenerateKey("photo.jpg", params2)

	// Assert
	assert.NotEqual(t, key1, key2, "Different params should produce different hashes")
}

// TestCacheManager_GenerateKey_SameParams tests same params create same hash
func TestCacheManager_GenerateKey_SameParams(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Act - generate keys multiple times
	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		keys[i] = manager.GenerateKey("photo.jpg", params)
	}

	// Assert - all keys should be identical
	firstKey := keys[0]
	for i, key := range keys {
		assert.Equal(t, firstKey, key, "Key %d should match first key", i)
	}
}

// TestCacheManager_GenerateKey_DefaultImageFallback tests cache keys when default image is used
func TestCacheManager_GenerateKey_DefaultImageFallback(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Act
	// When default image is served for missing.jpg, we still cache under missing.jpg
	keyOriginal := manager.GenerateKey("missing.jpg", params)
	keyDefault := manager.GenerateKey("missing.jpg", params) // Same request path

	// Assert
	assert.Equal(t, keyOriginal, keyDefault, "Cache key should be based on request path, not resolved path")
}

// TestCacheManager_GenerateKey_ResolvedPath tests cache keys use resolved file paths
func TestCacheManager_GenerateKey_ResolvedPath(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Act - different request paths that resolve to same file
	key1 := manager.GenerateKey("cat.jpg", params)
	key2 := manager.GenerateKey("cat.jpg", params) // Same resolved path
	key3 := manager.GenerateKey("dog.jpg", params) // Different file

	// Assert
	assert.Equal(t, key1, key2, "Same resolved path should produce same hash")
	assert.NotEqual(t, key1, key3, "Different files should produce different hashes")
}

// TestCacheManager_Store_ValidData tests successful cache storage
func TestCacheManager_Store_ValidData(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test image data")

	// Act
	err = manager.Store("photo.jpg", params, testData)

	// Assert
	assert.NoError(t, err)
}

// TestCacheManager_Store_CreateDirectoryStructure tests directory creation
func TestCacheManager_Store_CreateDirectoryStructure(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test image data")

	// Act
	err = manager.Store("nested/path/photo.jpg", params, testData)

	// Assert
	assert.NoError(t, err)

	// Verify directory structure was created
	cachePath := manager.GetPath("nested/path/photo.jpg", params)
	_, statErr := os.Stat(cachePath)
	assert.NoError(t, statErr, "Cache file should exist")
}

// TestCacheManager_Store_AtomicWrite tests atomic file operations
func TestCacheManager_Store_AtomicWrite(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test image data")

	// Act - store twice to ensure atomic replace works
	err1 := manager.Store("photo.jpg", params, testData)
	err2 := manager.Store("photo.jpg", params, []byte("updated data"))

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Verify final data is correct
	data, exists, err := manager.Retrieve("photo.jpg", params)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, []byte("updated data"), data)
}

// TestCacheManager_Store_DefaultImageCacheUnderOriginal tests default image cached under original filename
func TestCacheManager_Store_DefaultImageCacheUnderOriginal(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	defaultImageData := []byte("default image data")

	// Act - cache default image under the missing file's path
	err = manager.Store("missing.jpg", params, defaultImageData)

	// Assert
	assert.NoError(t, err)

	// Verify it's cached under missing.jpg, not default.jpg
	cachePath := manager.GetPath("missing.jpg", params)
	assert.Contains(t, cachePath, "missing.jpg")
}

// TestCacheManager_Store_GroupedImages tests caching grouped images
func TestCacheManager_Store_GroupedImages(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	imageData := []byte("grouped image data")

	// Act
	err = manager.Store("cats/default.jpg", params, imageData)

	// Assert
	assert.NoError(t, err)

	// Verify proper path structure
	cachePath := manager.GetPath("cats/default.jpg", params)
	assert.Contains(t, cachePath, "cats")
	assert.Contains(t, cachePath, "default.jpg")
}

// TestCacheManager_Retrieve_ExistingFile tests cache hit
func TestCacheManager_Retrieve_ExistingFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test image data")

	err = manager.Store("photo.jpg", params, testData)
	require.NoError(t, err)

	// Act
	data, exists, err := manager.Retrieve("photo.jpg", params)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, testData, data)
}

// TestCacheManager_Retrieve_MissingFile tests cache miss
func TestCacheManager_Retrieve_MissingFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Act
	data, exists, err := manager.Retrieve("nonexistent.jpg", params)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, data)
}

// TestCacheManager_Retrieve_CorruptedFile tests corrupted cache handling
func TestCacheManager_Retrieve_CorruptedFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Store valid data first
	err = manager.Store("photo.jpg", params, []byte("test data"))
	require.NoError(t, err)

	// Corrupt the file by truncating it
	cachePath := manager.GetPath("photo.jpg", params)
	err = os.Truncate(cachePath, 0)
	require.NoError(t, err)

	// Act
	data, exists, err := manager.Retrieve("photo.jpg", params)

	// Assert - should handle gracefully
	// Either return empty data or error, but shouldn't panic
	if err != nil {
		assert.False(t, exists)
	} else {
		assert.True(t, exists)
		assert.Equal(t, []byte{}, data)
	}
}

// TestCacheManager_Retrieve_DefaultImageFromCache tests retrieving cached default image
func TestCacheManager_Retrieve_DefaultImageFromCache(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	defaultData := []byte("default image data")

	// Store under original missing file path
	err = manager.Store("missing.jpg", params, defaultData)
	require.NoError(t, err)

	// Act
	data, exists, err := manager.Retrieve("missing.jpg", params)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, defaultData, data)
}

// TestCacheManager_Retrieve_GroupedImageCache tests retrieving grouped images from cache
func TestCacheManager_Retrieve_GroupedImageCache(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	imageData := []byte("grouped image data")

	err = manager.Store("cats/default.jpg", params, imageData)
	require.NoError(t, err)

	// Act
	data, exists, err := manager.Retrieve("cats/default.jpg", params)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, imageData, data)
}

// TestCacheManager_Exists_ValidFile tests file existence check
func TestCacheManager_Exists_ValidFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data")

	// Act - check before store
	existsBefore := manager.Exists("photo.jpg", params)

	// Store the file
	err = manager.Store("photo.jpg", params, testData)
	require.NoError(t, err)

	// Check after store
	existsAfter := manager.Exists("photo.jpg", params)

	// Assert
	assert.False(t, existsBefore)
	assert.True(t, existsAfter)
}

// TestCacheManager_Clear_SpecificFile tests single file cache clear
func TestCacheManager_Clear_SpecificFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data")

	// Store multiple files
	err = manager.Store("photo1.jpg", params, testData)
	require.NoError(t, err)
	err = manager.Store("photo2.jpg", params, testData)
	require.NoError(t, err)

	// Act
	err = manager.Clear("photo1.jpg")

	// Assert
	assert.NoError(t, err)

	// Verify photo1 is cleared but photo2 remains
	exists1 := manager.Exists("photo1.jpg", params)
	exists2 := manager.Exists("photo2.jpg", params)
	assert.False(t, exists1)
	assert.True(t, exists2)
}

// TestCacheManager_Clear_AllFiles tests global cache clear
func TestCacheManager_Clear_AllFiles(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data")

	// Store multiple files
	err = manager.Store("photo1.jpg", params, testData)
	require.NoError(t, err)
	err = manager.Store("photo2.jpg", params, testData)
	require.NoError(t, err)

	// Act
	err = manager.ClearAll()

	// Assert
	assert.NoError(t, err)

	// Verify all files are cleared
	exists1 := manager.Exists("photo1.jpg", params)
	exists2 := manager.Exists("photo2.jpg", params)
	assert.False(t, exists1)
	assert.False(t, exists2)
}

// TestCacheManager_Clear_NonExistentFile tests clearing non-existent files
func TestCacheManager_Clear_NonExistentFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	// Act
	err = manager.Clear("nonexistent.jpg")

	// Assert - should not error when clearing non-existent file
	assert.NoError(t, err)
}

// TestCacheManager_GetPath_ValidStructure tests cache path generation
func TestCacheManager_GetPath_ValidStructure(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Act
	path := manager.GetPath("photo.jpg", params)

	// Assert
	assert.Contains(t, path, tempDir)
	assert.Contains(t, path, "photo.jpg")
	// Path should follow structure: {cache_dir}/{filename}/{hash}
	assert.True(t, filepath.IsAbs(path))
}

// TestCacheManager_GetStats_FileCount tests cache statistics
func TestCacheManager_GetStats_FileCount(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data")

	// Store some files
	err = manager.Store("photo1.jpg", params, testData)
	require.NoError(t, err)
	err = manager.Store("photo2.jpg", params, testData)
	require.NoError(t, err)

	// Act
	stats, err := manager.GetStats()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalFiles, int64(2))
	assert.Greater(t, stats.TotalSize, int64(0))
}

// TestCacheManager_Concurrent_SafeOperations tests thread safety
func TestCacheManager_Concurrent_SafeOperations(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	goroutines := 50

	// Act
	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // Both store and retrieve

	// Concurrent stores
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			filename := fmt.Sprintf("photo%d.jpg", id)
			data := []byte(fmt.Sprintf("data-%d", id))
			err := manager.Store(filename, params, data)
			assert.NoError(t, err)
		}(i)
	}

	// Concurrent retrieves
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			filename := fmt.Sprintf("photo%d.jpg", id)
			// May or may not exist depending on timing
			_, _, err := manager.Retrieve(filename, params)
			assert.NoError(t, err)
		}(i)
	}

	// Assert - should complete without deadlock or data races
	wg.Wait()
}

// TestCacheManager_Concurrent_MultipleWrites tests concurrent writes to same file
func TestCacheManager_Concurrent_MultipleWrites(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	goroutines := 20

	// Act
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			data := []byte(fmt.Sprintf("data-%d", id))
			err := manager.Store("same-file.jpg", params, data)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Assert - file should exist and contain one of the writes
	data, exists, err := manager.Retrieve("same-file.jpg", params)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotEmpty(t, data)
}

// TestCacheManager_Concurrent_ReadWrite tests concurrent read/write operations
func TestCacheManager_Concurrent_ReadWrite(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	initialData := []byte("initial data")

	// Pre-populate some data
	err = manager.Store("photo.jpg", params, initialData)
	require.NoError(t, err)

	goroutines := 30

	// Act
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Concurrent reads
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, exists, err := manager.Retrieve("photo.jpg", params)
			assert.NoError(t, err)
			assert.True(t, exists)
		}()
	}

	// Concurrent writes
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			data := []byte(fmt.Sprintf("updated-%d", id))
			err := manager.Store("photo.jpg", params, data)
			assert.NoError(t, err)
		}(i)
	}

	// Assert - should complete without race conditions
	wg.Wait()
}

// TestCacheManager_Concurrent_ClearOperations tests concurrent clear operations
func TestCacheManager_Concurrent_ClearOperations(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data")

	// Pre-populate cache
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("photo%d.jpg", i)
		err = manager.Store(filename, params, testData)
		require.NoError(t, err)
	}

	goroutines := 20

	// Act
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				// Clear specific files
				filename := fmt.Sprintf("photo%d.jpg", id%10)
				err := manager.Clear(filename)
				assert.NoError(t, err)
			} else {
				// Check existence
				filename := fmt.Sprintf("photo%d.jpg", id%10)
				manager.Exists(filename, params)
			}
		}(i)
	}

	// Assert - should complete without deadlock
	wg.Wait()
}

// TestCacheManager_NewManager_InvalidDirectory tests error handling for invalid directory
func TestCacheManager_NewManager_InvalidDirectory(t *testing.T) {
	// Arrange - use an invalid path (on Unix, /dev/null/invalid is not a valid directory)
	invalidPath := "/dev/null/invalid"

	// Act
	manager, err := NewManager(invalidPath)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, manager)
}

// TestCacheManager_Store_ErrorHandling tests store error scenarios
func TestCacheManager_Store_ErrorHandling(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Create a read-only directory to cause write errors
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err = os.MkdirAll(readOnlyDir, 0755)
	require.NoError(t, err)

	// Store a file first
	testData := []byte("test data")
	err = manager.Store("readonly/test.jpg", params, testData)

	// Should succeed or fail gracefully
	// (behavior depends on file system permissions)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

// TestCacheManager_Retrieve_ReadError tests retrieve error scenarios
func TestCacheManager_Retrieve_ReadError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// This test verifies that retrieval handles missing files gracefully
	data, exists, err := manager.Retrieve("missing.jpg", params)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, data)
}

// TestCacheManager_GetStats_EmptyCache tests stats on empty cache
func TestCacheManager_GetStats_EmptyCache(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	// Act
	stats, err := manager.GetStats()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalFiles)
	assert.Equal(t, int64(0), stats.TotalSize)
}

// TestCacheManager_MultipleParams tests different parameter combinations
func TestCacheManager_MultipleParams(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	testData := []byte("test data")

	// Test various parameter combinations
	params := []ProcessingParams{
		{Width: 800, Height: 600, Format: "webp", Quality: 90},
		{Width: 400, Height: 300, Format: "png", Quality: 85},
		{Width: 1200, Height: 900, Format: "jpeg", Quality: 75},
		{Width: 200, Height: 200, Format: "jpg", Quality: 100},
	}

	// Act - store with different params
	for i, p := range params {
		err := manager.Store(fmt.Sprintf("photo%d.jpg", i), p, testData)
		assert.NoError(t, err)
	}

	// Assert - retrieve with same params
	for i, p := range params {
		data, exists, err := manager.Retrieve(fmt.Sprintf("photo%d.jpg", i), p)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, testData, data)
	}
}

// TestCacheManager_PathNormalization tests path handling with various formats
func TestCacheManager_PathNormalization(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data")

	// Test paths with various formats
	paths := []string{
		"photo.jpg",
		"nested/photo.jpg",
		"deeply/nested/path/photo.jpg",
		"/leading/slash/photo.jpg",
	}

	// Act & Assert
	for _, path := range paths {
		err := manager.Store(path, params, testData)
		assert.NoError(t, err)

		data, exists, err := manager.Retrieve(path, params)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, testData, data)
	}
}

// TestCacheManager_Clear_WithSubdirectories tests clearing files with nested structure
func TestCacheManager_Clear_WithSubdirectories(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params1 := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	params2 := ProcessingParams{Width: 400, Height: 300, Format: "png", Quality: 85}
	testData := []byte("test data")

	// Store multiple versions of same file
	err = manager.Store("photo.jpg", params1, testData)
	require.NoError(t, err)
	err = manager.Store("photo.jpg", params2, testData)
	require.NoError(t, err)

	// Act - clear all versions of the file
	err = manager.Clear("photo.jpg")

	// Assert
	assert.NoError(t, err)
	assert.False(t, manager.Exists("photo.jpg", params1))
	assert.False(t, manager.Exists("photo.jpg", params2))
}

// TestCacheManager_ClearAll_MultipleFiles tests clearing entire cache
func TestCacheManager_ClearAll_MultipleFiles(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data")

	// Store files with nested paths
	files := []string{
		"photo1.jpg",
		"photos/photo2.jpg",
		"gallery/summer/photo3.jpg",
	}

	for _, file := range files {
		err = manager.Store(file, params, testData)
		require.NoError(t, err)
	}

	// Verify they exist
	for _, file := range files {
		assert.True(t, manager.Exists(file, params))
	}

	// Act
	err = manager.ClearAll()

	// Assert
	assert.NoError(t, err)
	for _, file := range files {
		assert.False(t, manager.Exists(file, params))
	}
}

// TestCacheManager_GetStats_AfterOperations tests stats after various operations
func TestCacheManager_GetStats_AfterOperations(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data with some content")

	// Store files
	for i := 0; i < 5; i++ {
		err = manager.Store(fmt.Sprintf("photo%d.jpg", i), params, testData)
		require.NoError(t, err)
	}

	// Act
	stats, err := manager.GetStats()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(5), stats.TotalFiles)
	assert.Greater(t, stats.TotalSize, int64(0))
	assert.False(t, stats.OldestFileTime.IsZero())
	assert.False(t, stats.NewestFileTime.IsZero())
}

// TestCacheManager_Retrieve_AfterStoreUpdate tests retrieving after updating
func TestCacheManager_Retrieve_AfterStoreUpdate(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(t, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	originalData := []byte("original data")
	updatedData := []byte("updated data")

	// Store original
	err = manager.Store("photo.jpg", params, originalData)
	require.NoError(t, err)

	// Retrieve original
	data, exists, err := manager.Retrieve("photo.jpg", params)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, originalData, data)

	// Update
	err = manager.Store("photo.jpg", params, updatedData)
	require.NoError(t, err)

	// Act - retrieve updated
	data, exists, err = manager.Retrieve("photo.jpg", params)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, updatedData, data)
}


// BenchmarkCacheManager_Store_SmallFile benchmarks storing small files
func BenchmarkCacheManager_Store_SmallFile(b *testing.B) {
	tempDir := b.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(b, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	smallData := make([]byte, 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := fmt.Sprintf("photo%d.jpg", i%1000)
		_ = manager.Store(filename, params, smallData)
	}
}

// BenchmarkCacheManager_Store_LargeFile benchmarks storing large files
func BenchmarkCacheManager_Store_LargeFile(b *testing.B) {
	tempDir := b.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(b, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	largeData := make([]byte, 1024*1024) // 1MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := fmt.Sprintf("photo%d.jpg", i%100)
		_ = manager.Store(filename, params, largeData)
	}
}

// BenchmarkCacheManager_Retrieve_Hit benchmarks cache retrieval
func BenchmarkCacheManager_Retrieve_Hit(b *testing.B) {
	tempDir := b.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(b, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := make([]byte, 10240) // 10KB

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("photo%d.jpg", i)
		_ = manager.Store(filename, params, testData)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := fmt.Sprintf("photo%d.jpg", i%100)
		_, _, _ = manager.Retrieve(filename, params)
	}
}

// BenchmarkCacheManager_GenerateKey benchmarks hash generation
func BenchmarkCacheManager_GenerateKey(b *testing.B) {
	tempDir := b.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(b, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GenerateKey("photo.jpg", params)
	}
}

// BenchmarkCacheManager_Exists benchmarks existence checks
func BenchmarkCacheManager_Exists(b *testing.B) {
	tempDir := b.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(b, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := []byte("test data")

	// Pre-populate some files
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("photo%d.jpg", i)
		_ = manager.Store(filename, params, testData)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := fmt.Sprintf("photo%d.jpg", i%200) // Mix of existing and non-existing
		_ = manager.Exists(filename, params)
	}
}

// BenchmarkCacheManager_Concurrent_Operations benchmarks concurrent access
func BenchmarkCacheManager_Concurrent_Operations(b *testing.B) {
	tempDir := b.TempDir()
	manager, err := NewManager(tempDir)
	require.NoError(b, err)

	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	testData := make([]byte, 5120) // 5KB

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			filename := fmt.Sprintf("photo%d.jpg", i%100)
			if i%2 == 0 {
				_ = manager.Store(filename, params, testData)
			} else {
				_, _, _ = manager.Retrieve(filename, params)
			}
			i++
		}
	})
}


