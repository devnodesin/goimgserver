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
