package precache

import (
	"context"
	"fmt"
	"goimgserver/cache"
	"goimgserver/resolver"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkPreCache_ScanDirectory_SmallSet(b *testing.B) {
	// Create test directory with 10 images
	tmpDir := b.TempDir()
	for i := 0; i < 10; i++ {
		imagePath := filepath.Join(tmpDir, fmt.Sprintf("image%d.jpg", i))
		err := os.WriteFile(imagePath, getTestJPEGData(), 0644)
		if err != nil {
			b.Fatal(err)
		}
	}
	
	scanner := NewScanner()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(context.Background(), tmpDir, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPreCache_ScanDirectory_LargeSet(b *testing.B) {
	// Create test directory with 100 images
	tmpDir := b.TempDir()
	for i := 0; i < 100; i++ {
		imagePath := filepath.Join(tmpDir, fmt.Sprintf("image%d.jpg", i))
		err := os.WriteFile(imagePath, getTestJPEGData(), 0644)
		if err != nil {
			b.Fatal(err)
		}
	}
	
	scanner := NewScanner()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(context.Background(), tmpDir, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPreCache_ProcessImage_Sequential(b *testing.B) {
	tmpDir := b.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	if err != nil {
		b.Fatal(err)
	}
	
	// Create test image
	testImage := filepath.Join(imageDir, "test.jpg")
	err = os.WriteFile(testImage, getTestJPEGData(), 0644)
	if err != nil {
		b.Fatal(err)
	}
	
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	if err != nil {
		b.Fatal(err)
	}
	mockProcessor := &mockImageProcessor{}
	
	processor := NewProcessor(imageDir, fileResolver, cacheManager, mockProcessor)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Clear cache between runs
		cacheManager.ClearAll()
		
		err := processor.Process(context.Background(), testImage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPreCache_ProcessImage_Concurrent(b *testing.B) {
	tmpDir := b.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	imageDir := filepath.Join(tmpDir, "images")
	err := os.MkdirAll(imageDir, 0755)
	if err != nil {
		b.Fatal(err)
	}
	
	// Create 10 test images
	numImages := 10
	imagePaths := make([]string, numImages)
	for i := 0; i < numImages; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("image%d.jpg", i))
		err = os.WriteFile(imagePath, getTestJPEGData(), 0644)
		if err != nil {
			b.Fatal(err)
		}
		imagePaths[i] = imagePath
	}
	
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	if err != nil {
		b.Fatal(err)
	}
	mockProcessor := &mockImageProcessor{}
	
	processor := NewProcessor(imageDir, fileResolver, cacheManager, mockProcessor)
	executor := NewConcurrentExecutor(processor, 4, NewProgress())
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Clear cache between runs
		cacheManager.ClearAll()
		
		_, err := executor.Execute(context.Background(), imagePaths)
		if err != nil {
			b.Fatal(err)
		}
	}
}
