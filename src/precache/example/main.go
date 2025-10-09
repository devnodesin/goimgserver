package main

import (
	"context"
	"fmt"
	"goimgserver/cache"
	"goimgserver/precache"
	"goimgserver/resolver"
	"log"
	"os"
	"path/filepath"
	"time"
)

// mockProcessor is a simple mock for demonstration
type mockProcessor struct{}

func (m *mockProcessor) Process(data []byte, opts interface{}) ([]byte, error) {
	// In a real application, this would process the image
	// For demo, we just simulate some work
	time.Sleep(10 * time.Millisecond)
	return []byte("processed image data"), nil
}

func main() {
	fmt.Println("Pre-cache Example")
	fmt.Println("=================")
	fmt.Println()
	
	// Setup temporary directories
	tmpDir, err := os.MkdirTemp("", "precache-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	imageDir := filepath.Join(tmpDir, "images")
	cacheDir := filepath.Join(tmpDir, "cache")
	
	// Create image directory structure
	fmt.Println("Setting up test environment...")
	err = os.MkdirAll(imageDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	
	// Create some test images
	testImages := []string{
		"photo1.jpg",
		"photo2.jpg",
		"photo3.jpg",
		"cats/cat1.jpg",
		"cats/cat2.jpg",
		"cats/default.jpg",
		"dogs/dog1.jpg",
	}
	
	for _, imagePath := range testImages {
		fullPath := filepath.Join(imageDir, imagePath)
		dir := filepath.Dir(fullPath)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		
		// Create minimal JPEG file
		err = os.WriteFile(fullPath, getTestJPEGData(), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	
	// Create system default image
	defaultImagePath := filepath.Join(imageDir, "default.jpg")
	err = os.WriteFile(defaultImagePath, getTestJPEGData(), 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Created test images in: %s\n", imageDir)
	fmt.Printf("Cache directory: %s\n", cacheDir)
	fmt.Println()
	
	// Initialize components
	fmt.Println("Initializing components...")
	fileResolver := resolver.NewResolverWithCache(imageDir)
	cacheManager, err := cache.NewManager(cacheDir)
	if err != nil {
		log.Fatal(err)
	}
	mockProc := &mockProcessor{}
	
	// Configure pre-cache
	config := &precache.PreCacheConfig{
		ImageDir:         imageDir,
		CacheDir:         cacheDir,
		DefaultImagePath: defaultImagePath,
		Enabled:          true,
		Workers:          2,
	}
	
	fmt.Printf("Configuration:\n")
	fmt.Printf("  - Image Directory: %s\n", config.ImageDir)
	fmt.Printf("  - Cache Directory: %s\n", config.CacheDir)
	fmt.Printf("  - Default Image: %s\n", config.DefaultImagePath)
	fmt.Printf("  - Workers: %d\n", config.Workers)
	fmt.Println()
	
	// Create pre-cache instance
	preCache, err := precache.New(config, fileResolver, cacheManager, mockProc)
	if err != nil {
		log.Fatal(err)
	}
	
	// Run pre-cache synchronously
	fmt.Println("Starting pre-cache (synchronous)...")
	fmt.Println("-----------------------------------")
	
	startTime := time.Now()
	stats, err := preCache.Run(context.Background())
	duration := time.Since(startTime)
	
	if err != nil {
		log.Fatal(err)
	}
	
	// Display results
	fmt.Println("\nPre-cache Complete!")
	fmt.Println("-----------------------------------")
	fmt.Printf("Total Images Found: %d\n", stats.TotalImages)
	fmt.Printf("Successfully Processed: %d\n", stats.ProcessedOK)
	fmt.Printf("Skipped (already cached): %d\n", stats.Skipped)
	fmt.Printf("Errors: %d\n", stats.Errors)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Println()
	
	// Demonstrate async execution
	fmt.Println("Demonstrating async execution...")
	fmt.Println("-----------------------------------")
	
	// Clear cache for fresh run
	err = cacheManager.ClearAll()
	if err != nil {
		log.Fatal(err)
	}
	
	// Create new pre-cache instance
	preCache2, err := precache.New(config, fileResolver, cacheManager, mockProc)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Starting pre-cache (asynchronous)...")
	preCache2.RunAsync(context.Background())
	fmt.Println("Pre-cache running in background (non-blocking)")
	fmt.Println("Application can continue with other work...")
	
	// Simulate other work
	for i := 1; i <= 3; i++ {
		fmt.Printf("  Doing other work... %d/3\n", i)
		time.Sleep(100 * time.Millisecond)
	}
	
	// Wait a bit for async to complete
	time.Sleep(2 * time.Second)
	
	fmt.Println("\nExample complete!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("  ✓ Directory scanning (recursive)")
	fmt.Println("  ✓ Grouped image support (cats/, dogs/)")
	fmt.Println("  ✓ Default image exclusion (default.jpg)")
	fmt.Println("  ✓ Concurrent processing (worker pools)")
	fmt.Println("  ✓ Progress tracking")
	fmt.Println("  ✓ Synchronous execution")
	fmt.Println("  ✓ Asynchronous execution")
}

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
