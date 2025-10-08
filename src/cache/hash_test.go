package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_GenerateHash_Consistency tests that hash generation is deterministic
func Test_GenerateHash_Consistency(t *testing.T) {
	// Arrange
	resolvedPath := "photo.jpg"
	params := ProcessingParams{
		Width:   800,
		Height:  600,
		Format:  "webp",
		Quality: 90,
	}

	// Act
	hash1 := generateHash(resolvedPath, params)
	hash2 := generateHash(resolvedPath, params)

	// Assert
	assert.Equal(t, hash1, hash2, "Same inputs should produce same hash")
	assert.NotEmpty(t, hash1)
	assert.Len(t, hash1, 64) // SHA256 produces 64 hex characters
}

// Test_GenerateHash_DifferentInputs tests that different inputs produce different hashes
func Test_GenerateHash_DifferentInputs(t *testing.T) {
	// Arrange
	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Act
	hash1 := generateHash("photo1.jpg", params)
	hash2 := generateHash("photo2.jpg", params)

	// Assert
	assert.NotEqual(t, hash1, hash2, "Different paths should produce different hashes")
}

// Test_GenerateHash_DifferentParams tests that different parameters produce different hashes
func Test_GenerateHash_DifferentParams(t *testing.T) {
	// Arrange
	path := "photo.jpg"
	params1 := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	params2 := ProcessingParams{Width: 400, Height: 300, Format: "png", Quality: 85}

	// Act
	hash1 := generateHash(path, params1)
	hash2 := generateHash(path, params2)

	// Assert
	assert.NotEqual(t, hash1, hash2, "Different parameters should produce different hashes")
}

// Test_GenerateHash_ParameterVariations tests specific parameter changes
func Test_GenerateHash_ParameterVariations(t *testing.T) {
	tests := []struct {
		name    string
		params1 ProcessingParams
		params2 ProcessingParams
		want    string // "equal" or "different"
	}{
		{
			name:    "Same parameters",
			params1: ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90},
			params2: ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90},
			want:    "equal",
		},
		{
			name:    "Different width",
			params1: ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90},
			params2: ProcessingParams{Width: 400, Height: 600, Format: "webp", Quality: 90},
			want:    "different",
		},
		{
			name:    "Different height",
			params1: ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90},
			params2: ProcessingParams{Width: 800, Height: 300, Format: "webp", Quality: 90},
			want:    "different",
		},
		{
			name:    "Different format",
			params1: ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90},
			params2: ProcessingParams{Width: 800, Height: 600, Format: "png", Quality: 90},
			want:    "different",
		},
		{
			name:    "Different quality",
			params1: ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90},
			params2: ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 75},
			want:    "different",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			hash1 := generateHash("photo.jpg", tt.params1)
			hash2 := generateHash("photo.jpg", tt.params2)

			// Assert
			if tt.want == "equal" {
				assert.Equal(t, hash1, hash2, "Hashes should be equal")
			} else {
				assert.NotEqual(t, hash1, hash2, "Hashes should be different")
			}
		})
	}
}

// Test_GenerateHash_EmptyPath tests hash generation with empty path
func Test_GenerateHash_EmptyPath(t *testing.T) {
	// Arrange
	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}

	// Act
	hash := generateHash("", params)

	// Assert
	assert.NotEmpty(t, hash, "Should generate hash even with empty path")
	assert.Len(t, hash, 64)
}

// Test_GenerateHash_SpecialCharacters tests hash generation with special characters in path
func Test_GenerateHash_SpecialCharacters(t *testing.T) {
	// Arrange
	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	paths := []string{
		"photo with spaces.jpg",
		"photo-with-dashes.jpg",
		"photo_with_underscores.jpg",
		"nested/path/photo.jpg",
		"Ññ-unicode.jpg",
	}

	// Act & Assert
	for _, path := range paths {
		hash := generateHash(path, params)
		assert.NotEmpty(t, hash, "Should generate hash for path: %s", path)
		assert.Len(t, hash, 64, "Hash length should be 64 for path: %s", path)
	}
}

// Benchmark_GenerateHash benchmarks hash generation performance
func Benchmark_GenerateHash(b *testing.B) {
	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	path := "photo.jpg"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateHash(path, params)
	}
}

// Benchmark_GenerateHash_LongPath benchmarks hash generation with long path
func Benchmark_GenerateHash_LongPath(b *testing.B) {
	params := ProcessingParams{Width: 800, Height: 600, Format: "webp", Quality: 90}
	path := "very/long/nested/path/to/some/deeply/nested/image/file/photo.jpg"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateHash(path, params)
	}
}
