package performance

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"testing"

	"goimgserver/cache"
	"goimgserver/resolver"
	"goimgserver/testutils"
)

// BenchmarkFullPipeline_SmallImage benchmarks full pipeline with small images
func BenchmarkFullPipeline_SmallImage(b *testing.B) {
	// Setup
	tmpDir := b.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	// Setup cache
	cm, err := cache.NewManager(cacheDir)
	if err != nil {
		b.Fatalf("Failed to create cache manager: %v", err)
	}
	
	params := cache.ProcessingParams{
		Width:   100,
		Height:  100,
		Format:  "jpeg",
		Quality: 85,
	}
	
	resolvedPath := "/images/small_test.jpg"
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Check cache
		_, found, _ := cm.Retrieve(resolvedPath, params)
		if !found {
			// Simulate processing
			data := []byte("processed image data")
			_ = cm.Store(resolvedPath, params, data)
		}
	}
}

// BenchmarkFullPipeline_LargeImage benchmarks full pipeline with large images
func BenchmarkFullPipeline_LargeImage(b *testing.B) {
	tmpDir := b.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	if err != nil {
		b.Fatalf("Failed to create cache manager: %v", err)
	}
	
	params := cache.ProcessingParams{
		Width:   2000,
		Height:  1500,
		Format:  "jpeg",
		Quality: 85,
	}
	
	resolvedPath := "/images/large_test.jpg"
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, found, _ := cm.Retrieve(resolvedPath, params)
		if !found {
			data := make([]byte, 500*1024) // 500KB simulated data
			_ = cm.Store(resolvedPath, params, data)
		}
	}
}

// BenchmarkCacheOperations benchmarks cache store and retrieve operations
func BenchmarkCacheOperations(b *testing.B) {
	tmpDir := b.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	if err != nil {
		b.Fatalf("Failed to create cache manager: %v", err)
	}
	
	params := cache.ProcessingParams{
		Width:   800,
		Height:  600,
		Format:  "jpeg",
		Quality: 85,
	}
	
	data := make([]byte, 100*1024) // 100KB
	resolvedPath := "/images/test.jpg"
	
	b.Run("Store", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path := fmt.Sprintf("/images/test%d.jpg", i)
			_ = cm.Store(path, params, data)
		}
	})
	
	b.Run("Retrieve", func(b *testing.B) {
		// Pre-populate cache
		_ = cm.Store(resolvedPath, params, data)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = cm.Retrieve(resolvedPath, params)
		}
	})
	
	b.Run("KeyGeneration", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cm.GenerateKey(resolvedPath, params)
		}
	})
}

// BenchmarkResolverOperations benchmarks file resolution
func BenchmarkResolverOperations(b *testing.B) {
	tmpDir := b.TempDir()
	imagesDir := filepath.Join(tmpDir, "images")
	
	fixtureManager := testutils.NewFixtureManager(imagesDir)
	if err := fixtureManager.CreateFixtureSet(); err != nil {
		b.Fatalf("Failed to create fixtures: %v", err)
	}
	
	res := resolver.NewResolver(imagesDir)
	
	b.Run("Resolve_WithExtension", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = res.Resolve("small_test.jpg")
		}
	})
	
	b.Run("Resolve_WithoutExtension", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = res.Resolve("small_test")
		}
	})
}

// BenchmarkImageDecoding benchmarks image decoding operations
func BenchmarkImageDecoding(b *testing.B) {
	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	jpegData := buf.Bytes()
	
	buf.Reset()
	png.Encode(&buf, img)
	pngData := buf.Bytes()
	
	b.Run("DecodeJPEG", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = image.Decode(bytes.NewReader(jpegData))
		}
	})
	
	b.Run("DecodePNG", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = image.Decode(bytes.NewReader(pngData))
		}
	})
}

// BenchmarkMemoryUsage_ImageProcessing benchmarks memory usage
func BenchmarkMemoryUsage_ImageProcessing(b *testing.B) {
	tmpDir := b.TempDir()
	imagesDir := filepath.Join(tmpDir, "images")
	
	fixtureManager := testutils.NewFixtureManager(imagesDir)
	if err := fixtureManager.CreateFixtureSet(); err != nil {
		b.Fatalf("Failed to create fixtures: %v", err)
	}
	
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"100x100", 100, 100},
		{"500x500", 500, 500},
		{"1000x1000", 1000, 1000},
		{"2000x2000", 2000, 2000},
	}
	
	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				// Create test image
				img := image.NewRGBA(image.Rect(0, 0, size.width, size.height))
				
				// Encode to JPEG
				var buf bytes.Buffer
				_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
				
				// Decode
				_, _, _ = image.Decode(bytes.NewReader(buf.Bytes()))
			}
		})
	}
}

// BenchmarkConcurrentCacheAccess benchmarks concurrent cache access
func BenchmarkConcurrentCacheAccess(b *testing.B) {
	tmpDir := b.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	if err != nil {
		b.Fatalf("Failed to create cache manager: %v", err)
	}
	
	params := cache.ProcessingParams{
		Width:   800,
		Height:  600,
		Format:  "jpeg",
		Quality: 85,
	}
	
	data := make([]byte, 100*1024) // 100KB
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := fmt.Sprintf("/images/concurrent%d.jpg", i%100)
			_ = cm.Store(path, params, data)
			_, _, _ = cm.Retrieve(path, params)
			i++
		}
	})
}

// BenchmarkGracefulParsing_URLComplexity benchmarks parsing performance with complex URLs
func BenchmarkGracefulParsing_URLComplexity(b *testing.B) {
	testURLs := []string{
		"/img/test.jpg/800x600",
		"/img/test.jpg/800x600?quality=85",
		"/img/test.jpg/800x600?quality=85&format=webp",
		"/img/test.jpg/800x600?quality=85&format=webp&cache=true&foo=bar",
	}
	
	for _, url := range testURLs {
		b.Run(url, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate URL parsing
				_ = parseURL(url)
			}
		})
	}
}

// parseURL simulates URL parsing logic
func parseURL(url string) map[string]string {
	result := make(map[string]string)
	// Simple parsing simulation
	result["url"] = url
	return result
}
