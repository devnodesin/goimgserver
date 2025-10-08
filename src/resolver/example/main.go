package main

import (
	"fmt"
	"goimgserver/resolver"
	"log"
	"os"
	"path/filepath"
)

// This example demonstrates the file resolution system
func main() {
	// Setup a temporary directory structure for demonstration
	tmpDir, err := os.MkdirTemp("", "resolver-example-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Example directory: %s\n\n", tmpDir)

	// Create example file structure
	setupExampleFiles(tmpDir)

	// Create a resolver with caching
	res := resolver.NewResolverWithCache(tmpDir)

	// Example 1: Direct file resolution with extension
	fmt.Println("=== Example 1: Direct file resolution ===")
	result, err := res.Resolve("cat.jpg")
	if err == nil {
		fmt.Printf("Request: cat.jpg\n")
		fmt.Printf("Resolved: %s\n", result.ResolvedPath)
		fmt.Printf("Is Fallback: %v\n\n", result.IsFallback)
	}

	// Example 2: Auto-detection without extension
	fmt.Println("=== Example 2: Extension auto-detection ===")
	result, err = res.Resolve("dog")
	if err == nil {
		fmt.Printf("Request: dog\n")
		fmt.Printf("Resolved: %s\n", result.ResolvedPath)
		fmt.Printf("Extension auto-detected\n\n")
	}

	// Example 3: Extension priority (jpg > png > webp)
	fmt.Println("=== Example 3: Extension priority ===")
	result, err = res.Resolve("profile")
	if err == nil {
		fmt.Printf("Request: profile\n")
		fmt.Printf("Resolved: %s\n", result.ResolvedPath)
		fmt.Printf("Priority: jpg over png/webp\n\n")
	}

	// Example 4: Grouped image default
	fmt.Println("=== Example 4: Grouped image default ===")
	result, err = res.Resolve("cats")
	if err == nil {
		fmt.Printf("Request: cats\n")
		fmt.Printf("Resolved: %s\n", result.ResolvedPath)
		fmt.Printf("Is Grouped: %v\n\n", result.IsGrouped)
	}

	// Example 5: Specific grouped image
	fmt.Println("=== Example 5: Specific grouped image ===")
	result, err = res.Resolve("cats/cat_white")
	if err == nil {
		fmt.Printf("Request: cats/cat_white\n")
		fmt.Printf("Resolved: %s\n", result.ResolvedPath)
		fmt.Printf("Auto-detected extension with priority\n\n")
	}

	// Example 6: Missing file fallback
	fmt.Println("=== Example 6: Missing file with fallback ===")
	result, err = res.Resolve("cats/missing_cat")
	if err == nil {
		fmt.Printf("Request: cats/missing_cat\n")
		fmt.Printf("Resolved: %s\n", result.ResolvedPath)
		fmt.Printf("Is Fallback: %v\n", result.IsFallback)
		fmt.Printf("Fallback Type: %s\n\n", result.FallbackType)
	}

	// Example 7: Security - path traversal prevention
	fmt.Println("=== Example 7: Security - path traversal ===")
	result, err = res.Resolve("../../../etc/passwd")
	if err == nil {
		fmt.Printf("Request: ../../../etc/passwd\n")
		fmt.Printf("Resolved: %s\n", result.ResolvedPath)
		fmt.Printf("Safe: Falls back to system default\n\n")
	}

	// Example 8: Caching performance
	fmt.Println("=== Example 8: Caching performance ===")
	fmt.Println("First resolution (filesystem):")
	result1, _ := res.Resolve("cat.jpg")
	fmt.Printf("Resolved: %s\n", result1.ResolvedPath)

	fmt.Println("Second resolution (cached):")
	result2, _ := res.Resolve("cat.jpg")
	fmt.Printf("Resolved: %s\n", result2.ResolvedPath)
	fmt.Printf("Same result returned from cache\n")
}

func setupExampleFiles(dir string) {
	files := []string{
		"cat.jpg",
		"dog.png",
		"logo.webp",
		"profile.jpg",
		"profile.png",
		"profile.webp",
		"cats/default.jpg",
		"cats/cat_white.jpg",
		"cats/cat_white.png",
		"dogs/default.png",
		"dogs/puppy.jpg",
		"default.jpg",
	}

	for _, file := range files {
		fullPath := filepath.Join(dir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("example content"), 0644); err != nil {
			log.Fatal(err)
		}
	}
}
