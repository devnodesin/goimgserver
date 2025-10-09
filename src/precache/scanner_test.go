package precache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ScanDirectory_EmptyDirectory(t *testing.T) {
	// Create empty test directory
	tmpDir := t.TempDir()
	
	scanner := &directoryScanner{}
	
	images, err := scanner.Scan(context.Background(), tmpDir, "")
	
	require.NoError(t, err)
	assert.Empty(t, images, "Should return empty list for empty directory")
}

func Test_ScanDirectory_SingleImage(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a single test image
	testFile := filepath.Join(tmpDir, "test.jpg")
	err := os.WriteFile(testFile, []byte("fake image"), 0644)
	require.NoError(t, err)
	
	scanner := &directoryScanner{}
	
	images, err := scanner.Scan(context.Background(), tmpDir, "")
	
	require.NoError(t, err)
	assert.Len(t, images, 1, "Should find one image")
	assert.Contains(t, images[0], "test.jpg")
}

func Test_ScanDirectory_MultipleImages(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create multiple test images
	testFiles := []string{"test1.jpg", "test2.png", "test3.webp"}
	for _, file := range testFiles {
		err := os.WriteFile(filepath.Join(tmpDir, file), []byte("fake image"), 0644)
		require.NoError(t, err)
	}
	
	scanner := &directoryScanner{}
	
	images, err := scanner.Scan(context.Background(), tmpDir, "")
	
	require.NoError(t, err)
	assert.Len(t, images, 3, "Should find three images")
}

func Test_ScanDirectory_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create nested directory structure
	subdir := filepath.Join(tmpDir, "subdir")
	err := os.MkdirAll(subdir, 0755)
	require.NoError(t, err)
	
	// Create images in both directories
	err = os.WriteFile(filepath.Join(tmpDir, "root.jpg"), []byte("fake image"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subdir, "nested.jpg"), []byte("fake image"), 0644)
	require.NoError(t, err)
	
	scanner := &directoryScanner{}
	
	images, err := scanner.Scan(context.Background(), tmpDir, "")
	
	require.NoError(t, err)
	assert.Len(t, images, 2, "Should find images in nested directories")
}

func Test_ScanDirectory_ExcludeDefaultImage(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create regular image and default image
	err := os.WriteFile(filepath.Join(tmpDir, "regular.jpg"), []byte("fake image"), 0644)
	require.NoError(t, err)
	
	defaultImagePath := filepath.Join(tmpDir, "default.jpg")
	err = os.WriteFile(defaultImagePath, []byte("fake image"), 0644)
	require.NoError(t, err)
	
	scanner := &directoryScanner{}
	
	images, err := scanner.Scan(context.Background(), tmpDir, defaultImagePath)
	
	require.NoError(t, err)
	assert.Len(t, images, 1, "Should only find regular image, not default image")
	assert.Contains(t, images[0], "regular.jpg")
	assert.NotContains(t, images[0], "default.jpg")
}

func Test_ScanDirectory_UnsupportedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create supported and unsupported files
	err := os.WriteFile(filepath.Join(tmpDir, "image.jpg"), []byte("fake image"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "document.txt"), []byte("text file"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "script.sh"), []byte("#!/bin/bash"), 0644)
	require.NoError(t, err)
	
	scanner := &directoryScanner{}
	
	images, err := scanner.Scan(context.Background(), tmpDir, "")
	
	require.NoError(t, err)
	assert.Len(t, images, 1, "Should only find supported image files")
	assert.Contains(t, images[0], "image.jpg")
}

func Test_ScanDirectory_GroupedImages(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create grouped directory structure (images in subdirectories)
	groups := []string{"cats", "dogs", "birds"}
	for _, group := range groups {
		groupDir := filepath.Join(tmpDir, group)
		err := os.MkdirAll(groupDir, 0755)
		require.NoError(t, err)
		
		// Create images in each group
		err = os.WriteFile(filepath.Join(groupDir, "image1.jpg"), []byte("fake image"), 0644)
		require.NoError(t, err)
	}
	
	scanner := &directoryScanner{}
	
	images, err := scanner.Scan(context.Background(), tmpDir, "")
	
	require.NoError(t, err)
	assert.Len(t, images, 3, "Should find images in grouped directories")
}

func Test_ScanDirectory_GroupedDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create group with default image
	groupDir := filepath.Join(tmpDir, "cats")
	err := os.MkdirAll(groupDir, 0755)
	require.NoError(t, err)
	
	err = os.WriteFile(filepath.Join(groupDir, "cat1.jpg"), []byte("fake image"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(groupDir, "default.jpg"), []byte("fake image"), 0644)
	require.NoError(t, err)
	
	scanner := &directoryScanner{}
	
	// Group defaults should be included (they're different from system default)
	images, err := scanner.Scan(context.Background(), tmpDir, "")
	
	require.NoError(t, err)
	assert.Len(t, images, 2, "Should include group default images")
}

func Test_ScanDirectory_ExcludeSystemDefaultOnly(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create system default at root
	systemDefaultPath := filepath.Join(tmpDir, "default.jpg")
	err := os.WriteFile(systemDefaultPath, []byte("fake image"), 0644)
	require.NoError(t, err)
	
	// Create group default in subdirectory
	groupDir := filepath.Join(tmpDir, "cats")
	err = os.MkdirAll(groupDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(groupDir, "default.jpg"), []byte("fake image"), 0644)
	require.NoError(t, err)
	
	scanner := &directoryScanner{}
	
	images, err := scanner.Scan(context.Background(), tmpDir, systemDefaultPath)
	
	require.NoError(t, err)
	assert.Len(t, images, 1, "Should exclude system default but include group default")
	assert.Contains(t, images[0], filepath.Join("cats", "default.jpg"))
}

func Test_ScanDirectory_WithContext(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test images
	for i := 0; i < 5; i++ {
		imagePath := filepath.Join(tmpDir, fmt.Sprintf("image%d.jpg", i))
		err := os.WriteFile(imagePath, []byte("fake image"), 0644)
		require.NoError(t, err)
	}
	
	scanner := &directoryScanner{}
	
	// Test with context
	ctx := context.Background()
	images, err := scanner.Scan(ctx, tmpDir, "")
	
	require.NoError(t, err)
	assert.Len(t, images, 5)
}

func Test_ScanDirectory_NonExistentDirectory(t *testing.T) {
	scanner := &directoryScanner{}
	
	// Test with non-existent directory
	images, err := scanner.Scan(context.Background(), "/nonexistent/path", "")
	
	assert.Error(t, err)
	assert.Nil(t, images)
}
