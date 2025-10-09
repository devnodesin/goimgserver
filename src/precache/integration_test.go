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

// Test_Integration_RealImagesCaching tests pre-caching with actual image files
func Test_Integration_RealImagesCaching(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create test images with realistic structure
	testImages := []string{
		"photo1.jpg",
		"photo2.png",
		"subdir/photo3.webp",
		"products/item1.jpg",
		"products/item2.jpg",
	}
	
	for _, img := range testImages {
		imgPath := filepath.Join(imageDir, img)
		err = os.MkdirAll(filepath.Dir(imgPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(imgPath, getTestJPEGData(), 0644)
		require.NoError(t, err)
	}
	
	// Create default image
	defaultImagePath := filepath.Join(imageDir, "default.jpg")
	err = os.WriteFile(defaultImagePath, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	// Setup pre-cache configuration
	config := &PreCacheConfig{
		ImageDir:         imageDir,
		CacheDir:         cacheDir,
		DefaultImagePath: defaultImagePath,
		Enabled:          true,
		Workers:          2,
	}
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	// Create and run pre-cache
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	stats, err := preCache.Run(context.Background())
	require.NoError(t, err)
	
	// Verify results
	assert.Equal(t, len(testImages), stats.TotalImages, "Should process all test images except default")
	assert.Equal(t, len(testImages), stats.ProcessedOK, "All images should be processed successfully")
	assert.Equal(t, 0, stats.Errors, "Should have no errors")
	
	// Verify cache was created for each image
	params := cache.ProcessingParams{
		Width:   1000,
		Height:  1000,
		Format:  "webp",
		Quality: 95,
	}
	
	for _, img := range testImages {
		result, err := fileResolver.Resolve(img)
		require.NoError(t, err)
		exists := cacheManager.Exists(result.ResolvedPath, params)
		assert.True(t, exists, "Cache should exist for %s", img)
	}
}

// Test_Integration_CacheCreation tests actual cache directory creation
func Test_Integration_CacheCreation(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create test image
	testImage := filepath.Join(imageDir, "test.jpg")
	err = os.WriteFile(testImage, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	// Setup configuration
	config := &PreCacheConfig{
		ImageDir: imageDir,
		CacheDir: cacheDir,
		Enabled:  true,
		Workers:  1,
	}
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	// Run pre-cache
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	stats, err := preCache.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, stats.ProcessedOK)
	
	// Verify cache directory structure was created
	assert.DirExists(t, cacheDir, "Cache directory should be created")
	
	// Verify cache file exists
	params := cache.ProcessingParams{
		Width:   1000,
		Height:  1000,
		Format:  "webp",
		Quality: 95,
	}
	
	result, err := fileResolver.Resolve("test.jpg")
	require.NoError(t, err)
	exists := cacheManager.Exists(result.ResolvedPath, params)
	assert.True(t, exists, "Cache file should exist")
}

// Test_Integration_StartupFlow tests the complete startup flow
func Test_Integration_StartupFlow(t *testing.T) {
	// This test simulates the actual application startup flow
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create test images
	numImages := 10
	for i := 0; i < numImages; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("image%d.jpg", i))
		err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
		require.NoError(t, err)
	}
	
	// Create default image
	defaultImagePath := filepath.Join(imageDir, "default.jpg")
	err = os.WriteFile(defaultImagePath, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	// Step 1: Initialize components (as in main.go)
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	// Step 2: Configure pre-cache
	config := &PreCacheConfig{
		ImageDir:         imageDir,
		CacheDir:         cacheDir,
		DefaultImagePath: defaultImagePath,
		Enabled:          true,
		Workers:          4,
	}
	
	// Step 3: Create pre-cache instance
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	// Step 4: Run pre-cache (simulating async startup)
	ctx := context.Background()
	stats, err := preCache.Run(ctx)
	require.NoError(t, err)
	
	// Step 5: Verify pre-cache completed successfully
	assert.Equal(t, numImages, stats.TotalImages, "Should find all images except default")
	assert.Equal(t, numImages, stats.ProcessedOK, "All images should be processed")
	assert.Equal(t, 0, stats.Errors, "Should have no errors")
	assert.Greater(t, stats.Duration.Milliseconds(), int64(0), "Duration should be recorded")
	
	// Step 6: Verify cache is ready for serving requests
	params := cache.ProcessingParams{
		Width:   1000,
		Height:  1000,
		Format:  "webp",
		Quality: 95,
	}
	
	for i := 0; i < numImages; i++ {
		filename := fmt.Sprintf("image%d.jpg", i)
		result, err := fileResolver.Resolve(filename)
		require.NoError(t, err)
		exists := cacheManager.Exists(result.ResolvedPath, params)
		assert.True(t, exists, "Cache should be ready for serving %s", filename)
	}
}

// Test_Integration_FileResolution tests integration with file resolution system
func Test_Integration_FileResolution(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	
	// Create nested directory structure
	subdirs := []string{"products", "users", "categories"}
	for _, subdir := range subdirs {
		dir := filepath.Join(imageDir, subdir)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
		
		// Create images in each subdir
		for i := 0; i < 2; i++ {
			imagePath := filepath.Join(dir, fmt.Sprintf("image%d.jpg", i))
			err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
			require.NoError(t, err)
		}
	}
	
	// Setup and run pre-cache
	config := &PreCacheConfig{
		ImageDir: imageDir,
		CacheDir: cacheDir,
		Enabled:  true,
		Workers:  2,
	}
	
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	preCache, err := New(config, fileResolver, cacheManager, mockProcessor)
	require.NoError(t, err)
	
	stats, err := preCache.Run(context.Background())
	require.NoError(t, err)
	
	// Verify all images were resolved and cached correctly
	expectedImages := len(subdirs) * 2
	assert.Equal(t, expectedImages, stats.TotalImages)
	assert.Equal(t, expectedImages, stats.ProcessedOK)
	
	// Verify file resolution worked correctly for nested paths
	params := cache.ProcessingParams{
		Width:   1000,
		Height:  1000,
		Format:  "webp",
		Quality: 95,
	}
	
	for _, subdir := range subdirs {
		for i := 0; i < 2; i++ {
			relativePath := filepath.Join(subdir, fmt.Sprintf("image%d.jpg", i))
			result, err := fileResolver.Resolve(relativePath)
			require.NoError(t, err)
			exists := cacheManager.Exists(result.ResolvedPath, params)
			assert.True(t, exists, "Cache should exist for %s", relativePath)
		}
	}
}

// Test_Integration_ConfigurationOptional tests that pre-cache can be disabled
func Test_Integration_ConfigurationOptional(t *testing.T) {
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create test image
	testImage := filepath.Join(imageDir, "test.jpg")
	err = os.WriteFile(testImage, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	// Test with pre-cache disabled
	config := &PreCacheConfig{
		ImageDir: imageDir,
		CacheDir: filepath.Join(tmpDir, "cache"),
		Enabled:  false, // Disabled
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
	
	// When disabled, should return empty stats without processing
	assert.Equal(t, 0, stats.TotalImages, "Should not process any images when disabled")
	assert.Equal(t, 0, stats.ProcessedOK)
}
