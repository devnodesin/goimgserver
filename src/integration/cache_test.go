package integration

import (
	"path/filepath"
	"sync"
	"testing"

	"goimgserver/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Cache_BasicOperations tests basic cache operations
func TestIntegration_Cache_BasicOperations(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	
	t.Run("Retrieve non-existent file", func(t *testing.T) {
		params := cache.ProcessingParams{Width: 800, Height: 600}
		data, found, err := cm.Retrieve("nonexistent.jpg", params)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, data)
	})
	
	t.Run("Store and Retrieve file", func(t *testing.T) {
		resolvedPath := "/images/test.jpg"
		params := cache.ProcessingParams{Width: 800, Height: 600}
		data := []byte("test image data")
		
		err := cm.Store(resolvedPath, params, data)
		require.NoError(t, err)
		
		retrieved, found, err := cm.Retrieve(resolvedPath, params)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, data, retrieved)
	})
	
	t.Run("Overwrite existing file", func(t *testing.T) {
		resolvedPath := "/images/overwrite.jpg"
		params := cache.ProcessingParams{Width: 800, Height: 600}
		data1 := []byte("original data")
		data2 := []byte("updated data")
		
		err := cm.Store(resolvedPath, params, data1)
		require.NoError(t, err)
		
		err = cm.Store(resolvedPath, params, data2)
		require.NoError(t, err)
		
		retrieved, found, err := cm.Retrieve(resolvedPath, params)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, data2, retrieved)
	})
}

// TestIntegration_Cache_ClearOperations tests cache clearing
func TestIntegration_Cache_ClearOperations(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	
	// Add multiple files with same resolved path but different params
	resolvedPath := "/images/test.jpg"
	params1 := cache.ProcessingParams{Width: 800, Height: 600}
	params2 := cache.ProcessingParams{Width: 1024, Height: 768}
	
	err = cm.Store(resolvedPath, params1, []byte("data1"))
	require.NoError(t, err)
	
	err = cm.Store(resolvedPath, params2, []byte("data2"))
	require.NoError(t, err)
	
	// Clear cache for this resolved path
	err = cm.Clear(resolvedPath)
	require.NoError(t, err)
	
	// Verify files are gone
	_, found1, _ := cm.Retrieve(resolvedPath, params1)
	assert.False(t, found1)
	
	_, found2, _ := cm.Retrieve(resolvedPath, params2)
	assert.False(t, found2)
}

// TestIntegration_Cache_ConcurrentAccess tests concurrent cache operations
func TestIntegration_Cache_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 50
	
	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
				resolvedPath := "/images/concurrent/file.jpg"
				params := cache.ProcessingParams{Width: 800, Height: 600}
				data := []byte("data from goroutine")
				err := cm.Store(resolvedPath, params, data)
				assert.NoError(t, err)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify final state
	resolvedPath := "/images/concurrent/file.jpg"
	params := cache.ProcessingParams{Width: 800, Height: 600}
	data, found, err := cm.Retrieve(resolvedPath, params)
	require.NoError(t, err)
	assert.True(t, found)
	assert.NotNil(t, data)
}

// TestIntegration_Cache_DifferentParams tests different processing parameters
func TestIntegration_Cache_DifferentParams(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	
	resolvedPath := "/images/test.jpg"
	
	// Store with different parameters
	params1 := cache.ProcessingParams{Width: 800, Height: 600}
	params2 := cache.ProcessingParams{Width: 1024, Height: 768}
	
	data1 := []byte("800x600 data")
	data2 := []byte("1024x768 data")
	
	err = cm.Store(resolvedPath, params1, data1)
	require.NoError(t, err)
	
	err = cm.Store(resolvedPath, params2, data2)
	require.NoError(t, err)
	
	// Retrieve should return correct data for each parameter set
	retrieved1, found1, err := cm.Retrieve(resolvedPath, params1)
	require.NoError(t, err)
	assert.True(t, found1)
	assert.Equal(t, data1, retrieved1)
	
	retrieved2, found2, err := cm.Retrieve(resolvedPath, params2)
	require.NoError(t, err)
	assert.True(t, found2)
	assert.Equal(t, data2, retrieved2)
}

// TestIntegration_Cache_KeyGeneration tests cache key generation
func TestIntegration_Cache_KeyGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	
	resolvedPath := "/images/test.jpg"
	params := cache.ProcessingParams{
		Width:   800,
		Height:  600,
		Format:  "webp",
		Quality: 85,
	}
	
	key := cm.GenerateKey(resolvedPath, params)
	assert.NotEmpty(t, key)
	
	// Same parameters should generate same key
	key2 := cm.GenerateKey(resolvedPath, params)
	assert.Equal(t, key, key2)
	
	// Different parameters should generate different key
	params2 := cache.ProcessingParams{
		Width:   1024,
		Height:  768,
		Format:  "webp",
		Quality: 85,
	}
	key3 := cm.GenerateKey(resolvedPath, params2)
	assert.NotEqual(t, key, key3)
}

// TestIntegration_Cache_PathHandling tests path handling
func TestIntegration_Cache_PathHandling(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	
	params := cache.ProcessingParams{Width: 800, Height: 600}
	
	tests := []struct {
		name string
		path string
	}{
		{"Simple path", "/images/test.jpg"},
		{"Nested path", "/images/category/subcategory/test.jpg"},
		{"Path with dashes", "/images/test-image-01.jpg"},
		{"Path with underscores", "/images/test_image_01.jpg"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("test data")
			
			err := cm.Store(tt.path, params, data)
			require.NoError(t, err)
			
			retrieved, found, err := cm.Retrieve(tt.path, params)
			require.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, data, retrieved)
		})
	}
}


