package resolver

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCache_BasicOperations tests basic cache operations
func TestCache_BasicOperations(t *testing.T) {
	cache := NewCache()
	
	// Test empty cache
	_, found := cache.Get("key1")
	assert.False(t, found, "Cache should be empty initially")
	assert.Equal(t, 0, cache.Size())
	
	// Test set and get
	result := &ResolutionResult{
		ResolvedPath: "/path/to/image.jpg",
		IsGrouped:    false,
		IsFallback:   false,
	}
	
	cache.Set("key1", result)
	assert.Equal(t, 1, cache.Size())
	
	retrieved, found := cache.Get("key1")
	assert.True(t, found, "Should find cached entry")
	assert.Equal(t, result.ResolvedPath, retrieved.ResolvedPath)
	
	// Test invalidate
	cache.Invalidate("key1")
	assert.Equal(t, 0, cache.Size())
	
	_, found = cache.Get("key1")
	assert.False(t, found, "Entry should be invalidated")
}

// TestCache_Clear tests cache clearing
func TestCache_Clear(t *testing.T) {
	cache := NewCache()
	
	// Add multiple entries
	cache.Set("key1", &ResolutionResult{ResolvedPath: "/path1"})
	cache.Set("key2", &ResolutionResult{ResolvedPath: "/path2"})
	cache.Set("key3", &ResolutionResult{ResolvedPath: "/path3"})
	
	assert.Equal(t, 3, cache.Size())
	
	// Clear cache
	cache.Clear()
	assert.Equal(t, 0, cache.Size())
	
	_, found := cache.Get("key1")
	assert.False(t, found)
}

// TestCache_Concurrent tests thread-safe operations
func TestCache_Concurrent(t *testing.T) {
	cache := NewCache()
	
	var wg sync.WaitGroup
	iterations := 100
	
	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('a' + (idx % 26)))
			cache.Set(key, &ResolutionResult{
				ResolvedPath: "/path/" + key,
			})
		}(i)
	}
	
	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('a' + (idx % 26)))
			cache.Get(key)
		}(i)
	}
	
	wg.Wait()
	
	// Cache should have 26 entries (a-z)
	assert.LessOrEqual(t, cache.Size(), 26)
	assert.Greater(t, cache.Size(), 0)
}

// TestFileResolver_CacheResolution_Performance tests resolution caching
func TestFileResolver_CacheResolution_Performance(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolverWithCache(tmpDir)
	
	// First resolution - should hit filesystem
	result1, err := resolver.Resolve("cat.jpg")
	require.NoError(t, err)
	require.NotNil(t, result1)
	
	// Second resolution - should hit cache
	result2, err := resolver.Resolve("cat.jpg")
	require.NoError(t, err)
	require.NotNil(t, result2)
	
	// Results should be identical
	assert.Equal(t, result1.ResolvedPath, result2.ResolvedPath)
}

// TestFileResolver_Concurrent_ThreadSafety tests thread-safe operations
func TestFileResolver_Concurrent_ThreadSafety(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolverWithCache(tmpDir)
	
	var wg sync.WaitGroup
	iterations := 50
	
	paths := []string{
		"cat.jpg",
		"dog.png",
		"profile",
		"cats",
		"cats/cat_white",
	}
	
	// Concurrent resolutions
	for i := 0; i < iterations; i++ {
		for _, path := range paths {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				result, err := resolver.Resolve(p)
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}(path)
		}
	}
	
	wg.Wait()
}
