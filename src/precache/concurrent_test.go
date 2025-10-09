package precache

import (
	"context"
	"fmt"
	"goimgserver/cache"
	"goimgserver/resolver"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Concurrent_Processing(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create multiple test images
	numImages := 10
	imagePaths := make([]string, numImages)
	for i := 0; i < numImages; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("image%d.jpg", i))
		err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
		require.NoError(t, err)
		imagePaths[i] = imagePath
	}
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	processor := NewProcessor(imageDir, fileResolver, cacheManager, mockProcessor)
	
	// Create concurrent executor
	executor := NewConcurrentExecutor(processor, 4, NewProgress())
	
	// Execute concurrent processing
	stats, err := executor.Execute(context.Background(), imagePaths)
	
	require.NoError(t, err)
	assert.Equal(t, numImages, stats.ProcessedOK)
	assert.Equal(t, 0, stats.Errors)
}

func Test_Concurrent_ThreadSafety(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create test images
	numImages := 20
	imagePaths := make([]string, numImages)
	for i := 0; i < numImages; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("image%d.jpg", i))
		err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
		require.NoError(t, err)
		imagePaths[i] = imagePath
	}
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	processor := NewProcessor(imageDir, fileResolver, cacheManager, mockProcessor)
	
	// Run multiple concurrent executors
	var wg sync.WaitGroup
	numExecutors := 3
	
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			executor := NewConcurrentExecutor(processor, 4, NewProgress())
			_, err := executor.Execute(context.Background(), imagePaths)
			assert.NoError(t, err)
		}()
	}
	
	wg.Wait()
}

func Test_Concurrent_ErrorHandling(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create mix of valid and invalid images
	validImage := filepath.Join(imageDir, "valid.jpg")
	err = os.WriteFile(validImage, getTestJPEGData(), 0644)
	require.NoError(t, err)
	
	invalidImage := filepath.Join(imageDir, "invalid.jpg")
	// Don't create this file - it will cause an error
	
	imagePaths := []string{validImage, invalidImage}
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	processor := NewProcessor(imageDir, fileResolver, cacheManager, mockProcessor)
	
	// Create concurrent executor
	executor := NewConcurrentExecutor(processor, 2, NewProgress())
	
	// Execute - should handle errors gracefully
	stats, err := executor.Execute(context.Background(), imagePaths)
	
	require.NoError(t, err)
	assert.Equal(t, 1, stats.ProcessedOK)
	assert.Equal(t, 1, stats.Errors)
}

func Test_Concurrent_Cancellation(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create many test images
	numImages := 100
	imagePaths := make([]string, numImages)
	for i := 0; i < numImages; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("image%d.jpg", i))
		err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
		require.NoError(t, err)
		imagePaths[i] = imagePath
	}
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	processor := NewProcessor(imageDir, fileResolver, cacheManager, mockProcessor)
	
	// Create concurrent executor
	executor := NewConcurrentExecutor(processor, 2, NewProgress())
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	
	// Execute - should respect cancellation
	stats, err := executor.Execute(ctx, imagePaths)
	
	// Should have processed some but not all due to cancellation
	assert.True(t, stats.ProcessedOK+stats.Errors < numImages, "Should not process all images due to cancellation")
}

func Test_Concurrent_WorkerPool(t *testing.T) {
	// Create test directories
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	require.NoError(t, err)
	
	// Create test images
	numImages := 15
	imagePaths := make([]string, numImages)
	for i := 0; i < numImages; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("image%d.jpg", i))
		err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
		require.NoError(t, err)
		imagePaths[i] = imagePath
	}
	
	// Create dependencies
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	mockProcessor := &mockImageProcessor{}
	
	processor := NewProcessor(imageDir, fileResolver, cacheManager, mockProcessor)
	
	// Test different worker pool sizes
	for _, workers := range []int{1, 2, 4, 8} {
		t.Run(fmt.Sprintf("Workers=%d", workers), func(t *testing.T) {
			executor := NewConcurrentExecutor(processor, workers, NewProgress())
			stats, err := executor.Execute(context.Background(), imagePaths)
			
			require.NoError(t, err)
			assert.Equal(t, numImages, stats.ProcessedOK, "All images should be processed")
		})
	}
}
