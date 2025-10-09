package precache

import (
	"context"
	"goimgserver/cache"
	"goimgserver/resolver"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockImageProcessor is a mock for testing without vips dependency
type mockImageProcessor struct {
	shouldFail bool
}

func (m *mockImageProcessor) Process(data []byte, opts interface{}) ([]byte, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return []byte("processed image data"), nil
}

func Test_ProcessImage_DefaultSettings(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create a test image file
	testImage := filepath.Join(imageDir, "test.jpg")
	err = os.WriteFile(testImage, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	// Create processor
	proc := &preCacheProcessor{
		imageDir:  imageDir,
		resolver:  fileResolver,
		cache:     cacheManager,
		processor: mockProcessor,
	}
	
	// Process the image
	err = proc.Process(context.Background(), testImage)
	
	// Should process successfully
	require.NoError(t, err)
	
	// Verify cache was created with default settings (1000x1000, WebP, q95)
	params := cache.ProcessingParams{
		Width:   1000,
		Height:  1000,
		Format:  "webp",
		Quality: 95,
	}
	
	// Get resolved path
	relPath, err := filepath.Rel(imageDir, testImage)
	require.NoError(t, err)
	result, err := fileResolver.Resolve(relPath)
	require.NoError(t, err)
	
	exists := cacheManager.Exists(result.ResolvedPath, params)
	assert.True(t, exists, "Cache should exist with default settings")
}

func Test_ProcessImage_AlreadyCached(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create a test image file
	testImage := filepath.Join(imageDir, "test.jpg")
	err = os.WriteFile(testImage, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	// Pre-populate cache
	relPath, err := filepath.Rel(imageDir, testImage)
	require.NoError(t, err)
	result, err := fileResolver.Resolve(relPath)
	require.NoError(t, err)
	
	params := cache.ProcessingParams{
		Width:   1000,
		Height:  1000,
		Format:  "webp",
		Quality: 95,
	}
	
	err = cacheManager.Store(result.ResolvedPath, params, []byte("cached data"))
	require.NoError(t, err)
	
	// Create processor
	proc := &preCacheProcessor{
		imageDir:  imageDir,
		resolver:  fileResolver,
		cache:     cacheManager,
		processor: mockProcessor,
	}
	
	// Process should skip already cached image
	err = proc.Process(context.Background(), testImage)
	
	// Should succeed (skip without error)
	require.NoError(t, err)
}

func Test_ProcessImage_ErrorHandling(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	// Create processor
	proc := &preCacheProcessor{
		imageDir:  imageDir,
		resolver:  fileResolver,
		cache:     cacheManager,
		processor: mockProcessor,
	}
	
	// Process non-existent image
	err = proc.Process(context.Background(), filepath.Join(imageDir, "nonexistent.jpg"))
	
	// Should handle error gracefully
	assert.Error(t, err, "Should return error for non-existent file")
}

func Test_ProcessImage_CorruptedImage(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create a corrupted image file
	testImage := filepath.Join(imageDir, "corrupted.jpg")
	err = os.WriteFile(testImage, []byte("This is not an image"), 0644)
	require.NoError(t, err)
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{shouldFail: true}
	
	// Create processor
	proc := &preCacheProcessor{
		imageDir:  imageDir,
		resolver:  fileResolver,
		cache:     cacheManager,
		processor: mockProcessor,
	}
	
	// Process corrupted image
	err = proc.Process(context.Background(), testImage)
	
	// Should handle error gracefully
	assert.Error(t, err, "Should return error for corrupted image")
}

// getTestJPEGData returns a minimal valid JPEG image (1x1 red pixel)
func getTestJPEGData() []byte {
	return []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43,
		0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
		0x09, 0x08, 0x0a, 0x0c, 0x14, 0x0d, 0x0c, 0x0b, 0x0b, 0x0c, 0x19, 0x12,
		0x13, 0x0f, 0x14, 0x1d, 0x1a, 0x1f, 0x1e, 0x1d, 0x1a, 0x1c, 0x1c, 0x20,
		0x24, 0x2e, 0x27, 0x20, 0x22, 0x2c, 0x23, 0x1c, 0x1c, 0x28, 0x37, 0x29,
		0x2c, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1f, 0x27, 0x39, 0x3d, 0x38, 0x32,
		0x3c, 0x2e, 0x33, 0x34, 0x32, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4, 0x00, 0x14, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xff, 0xc4, 0x00, 0x14, 0x10, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xff, 0xda, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3f, 0x00,
		0x7f, 0xff, 0xd9,
	}
}
