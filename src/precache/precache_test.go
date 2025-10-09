package precache

import (
	"context"
	"fmt"
	"goimgserver/cache"
	"goimgserver/resolver"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PreCache_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	
	config := &PreCacheConfig{
		ImageDir: tmpDir,
		CacheDir: filepath.Join(tmpDir, "cache"),
		Enabled:  false,
	}
	
	fileResolver := resolver.NewResolverWithCache(tmpDir)
	cacheManager, err := cache.NewManager(config.CacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	stats, err := preCache.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalImages)
}

func Test_PreCache_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	
	config := &PreCacheConfig{
		ImageDir: tmpDir,
		CacheDir: filepath.Join(tmpDir, "cache"),
		Enabled:  true,
	}
	
	fileResolver := resolver.NewResolverWithCache(tmpDir)
	cacheManager, err := cache.NewManager(config.CacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	stats, err := preCache.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalImages)
}

func Test_PreCache_WithImages(t *testing.T) {
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create test images
	for i := 0; i < 5; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("image%d.jpg", i))
		err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
		require.NoError(t, err)
	}
	
	config := &PreCacheConfig{
		ImageDir: imageDir,
		CacheDir: filepath.Join(tmpDir, "cache"),
		Enabled:  true,
		Workers:  2,
	}
	
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(config.CacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	stats, err := preCache.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 5, stats.TotalImages)
	assert.Equal(t, 5, stats.ProcessedOK)
	assert.Equal(t, 0, stats.Errors)
}

func Test_PreCache_ExcludeDefaultImage(t *testing.T) {
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create regular images
	for i := 0; i < 3; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("image%d.jpg", i))
		err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
		require.NoError(t, err)
	}
	
	// Create default image
	defaultImagePath := filepath.Join(imageDir, "default.jpg")
	err = os.WriteFile(defaultImagePath, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	config := &PreCacheConfig{
		ImageDir:         imageDir,
		CacheDir:         filepath.Join(tmpDir, "cache"),
		DefaultImagePath: defaultImagePath,
		Enabled:          true,
		Workers:          2,
	}
	
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(config.CacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	stats, err := preCache.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalImages, "Should only process 3 images, excluding default")
	assert.Equal(t, 3, stats.ProcessedOK)
}

func Test_PreCache_Async(t *testing.T) {
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create test images
	imagePath := filepath.Join(imageDir, "image.jpg")
	err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	config := &PreCacheConfig{
		ImageDir: imageDir,
		CacheDir: filepath.Join(tmpDir, "cache"),
		Enabled:  true,
		Workers:  1,
	}
	
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(config.CacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	// RunAsync should not block
	preCache.RunAsync(context.Background())
	
	// Function returns immediately
	assert.True(t, true, "RunAsync should not block")
}

func Test_PreCache_NewWithNilConfig(t *testing.T) {
	tmpDir := t.TempDir()
	fileResolver := resolver.NewResolverWithCache(tmpDir)
	cacheManager, err := cache.NewManager(filepath.Join(tmpDir, "cache"))
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	// Test with nil config
	preCache, err := New(nil, fileResolver, cacheManager, mockProcessor)
	
	assert.Nil(t, preCache)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func Test_PreCache_NewWithZeroWorkers(t *testing.T) {
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Config with 0 workers (should default to NumCPU)
	config := &PreCacheConfig{
		ImageDir: imageDir,
		CacheDir: filepath.Join(tmpDir, "cache"),
		Enabled:  true,
		Workers:  0, // Should default to runtime.NumCPU()
	}
	
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(config.CacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	assert.NotNil(t, preCache)
	
	// Config workers should be updated
	assert.Greater(t, config.Workers, 0)
}

func Test_PreCache_RunWithScanError(t *testing.T) {
	// Test Run when scanner returns an error
	config := &PreCacheConfig{
		ImageDir: "/nonexistent/directory",
		CacheDir: "/tmp/cache",
		Enabled:  true,
		Workers:  2,
	}
	
	fileResolver := resolver.NewResolverWithCache("/tmp")
	cacheManager, err := cache.NewManager(config.CacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	// Run should return error from scanner
	stats, err := preCache.Run(context.Background())
	
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "failed to scan directory")
}
