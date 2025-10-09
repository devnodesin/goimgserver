package resolver

import (
	"testing"
)

// BenchmarkFileResolver_SingleImage_WithExtension benchmarks direct resolution
func BenchmarkFileResolver_SingleImage_WithExtension(b *testing.B) {
	tmpDir := setupTestDir(&testing.T{})
	resolver := NewResolver(tmpDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve("cat.jpg")
	}
}

// BenchmarkFileResolver_SingleImage_AutoDetection benchmarks auto-detection
func BenchmarkFileResolver_SingleImage_AutoDetection(b *testing.B) {
	tmpDir := setupTestDir(&testing.T{})
	resolver := NewResolver(tmpDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve("cat")
	}
}

// BenchmarkFileResolver_GroupedImage_Default benchmarks group default resolution
func BenchmarkFileResolver_GroupedImage_Default(b *testing.B) {
	tmpDir := setupTestDir(&testing.T{})
	resolver := NewResolver(tmpDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve("cats")
	}
}

// BenchmarkFileResolver_GroupedImage_Specific benchmarks specific grouped image
func BenchmarkFileResolver_GroupedImage_Specific(b *testing.B) {
	tmpDir := setupTestDir(&testing.T{})
	resolver := NewResolver(tmpDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve("cats/cat_white.jpg")
	}
}

// BenchmarkFileResolver_CacheHit benchmarks cached resolution
func BenchmarkFileResolver_CacheHit(b *testing.B) {
	tmpDir := setupTestDir(&testing.T{})
	resolver := NewResolverWithCache(tmpDir)
	
	// Warm up cache
	_, _ = resolver.Resolve("cat.jpg")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve("cat.jpg")
	}
}

// BenchmarkFileResolver_CacheMiss benchmarks uncached resolution
func BenchmarkFileResolver_CacheMiss(b *testing.B) {
	tmpDir := setupTestDir(&testing.T{})
	resolver := NewResolverWithCache(tmpDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache each iteration to simulate cache miss
		if resolver.cache != nil {
			resolver.cache.Clear()
		}
		_, _ = resolver.Resolve("cat.jpg")
	}
}

// BenchmarkFileResolver_ExtensionPriority benchmarks priority resolution
func BenchmarkFileResolver_ExtensionPriority(b *testing.B) {
	tmpDir := setupTestDir(&testing.T{})
	resolver := NewResolver(tmpDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve("profile")
	}
}

// BenchmarkFileResolver_PathTraversal benchmarks security validation
func BenchmarkFileResolver_PathTraversal(b *testing.B) {
	tmpDir := setupTestDir(&testing.T{})
	resolver := NewResolver(tmpDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve("../secret.txt")
	}
}
