package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseImageRequest_GracefulParsing_ValidParams tests parsing of valid parameters
func TestParseImageRequest_GracefulParsing_ValidParams(t *testing.T) {
	tests := []struct {
		name           string
		segments       []string
		expectedWidth  int
		expectedHeight int
		expectedFormat string
		expectedQual   int
	}{
		{
			name:           "Dimensions and format",
			segments:       []string{"800x600", "webp"},
			expectedWidth:  800,
			expectedHeight: 600,
			expectedFormat: "webp",
			expectedQual:   75, // default
		},
		{
			name:           "Width only",
			segments:       []string{"400"},
			expectedWidth:  400,
			expectedHeight: 0, // maintain aspect ratio
			expectedFormat: "webp",
			expectedQual:   75,
		},
		{
			name:           "Quality only",
			segments:       []string{"q90"},
			expectedWidth:  1000, // default
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   90,
		},
		{
			name:           "All parameters",
			segments:       []string{"1920x1080", "jpeg", "q95"},
			expectedWidth:  1920,
			expectedHeight: 1080,
			expectedFormat: "jpeg",
			expectedQual:   95,
		},
		{
			name:           "PNG format",
			segments:       []string{"300x300", "png", "q85"},
			expectedWidth:  300,
			expectedHeight: 300,
			expectedFormat: "png",
			expectedQual:   85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			params := parseParameters(tt.segments)

			// Assert
			assert.Equal(t, tt.expectedWidth, params.Width)
			assert.Equal(t, tt.expectedHeight, params.Height)
			assert.Equal(t, tt.expectedFormat, params.Format)
			assert.Equal(t, tt.expectedQual, params.Quality)
		})
	}
}

// TestParseImageRequest_GracefulParsing_InvalidParamsIgnored tests that invalid parameters are ignored
func TestParseImageRequest_GracefulParsing_InvalidParamsIgnored(t *testing.T) {
	tests := []struct {
		name           string
		segments       []string
		expectedWidth  int
		expectedHeight int
		expectedFormat string
		expectedQual   int
	}{
		{
			name:           "Invalid parameter ignored",
			segments:       []string{"800x600", "webp", "q90", "wow"},
			expectedWidth:  800,
			expectedHeight: 600,
			expectedFormat: "webp",
			expectedQual:   90,
		},
		{
			name:           "Multiple invalid parameters",
			segments:       []string{"800x600", "invalid", "webp", "extra", "q90"},
			expectedWidth:  800,
			expectedHeight: 600,
			expectedFormat: "webp",
			expectedQual:   90,
		},
		{
			name:           "All invalid uses defaults",
			segments:       []string{"wow", "invalid", "extra"},
			expectedWidth:  1000, // default
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			params := parseParameters(tt.segments)

			// Assert
			assert.Equal(t, tt.expectedWidth, params.Width)
			assert.Equal(t, tt.expectedHeight, params.Height)
			assert.Equal(t, tt.expectedFormat, params.Format)
			assert.Equal(t, tt.expectedQual, params.Quality)
		})
	}
}

// TestParseImageRequest_GracefulParsing_DuplicateParamsFirstWins tests first valid parameter wins
func TestParseImageRequest_GracefulParsing_DuplicateParamsFirstWins(t *testing.T) {
	tests := []struct {
		name           string
		segments       []string
		expectedWidth  int
		expectedHeight int
		expectedFormat string
		expectedQual   int
	}{
		{
			name:           "Duplicate dimensions - first wins",
			segments:       []string{"300", "400"},
			expectedWidth:  300,
			expectedHeight: 0,
			expectedFormat: "webp",
			expectedQual:   75,
		},
		{
			name:           "Duplicate formats - first wins",
			segments:       []string{"800x600", "jpeg", "png"},
			expectedWidth:  800,
			expectedHeight: 600,
			expectedFormat: "jpeg",
			expectedQual:   75,
		},
		{
			name:           "Duplicate quality - first wins",
			segments:       []string{"q85", "q95"},
			expectedWidth:  1000,
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   85,
		},
		{
			name:           "Multiple duplicates",
			segments:       []string{"300", "400", "jpeg", "png", "q85", "q95"},
			expectedWidth:  300,
			expectedHeight: 0,
			expectedFormat: "jpeg",
			expectedQual:   85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			params := parseParameters(tt.segments)

			// Assert
			assert.Equal(t, tt.expectedWidth, params.Width)
			assert.Equal(t, tt.expectedHeight, params.Height)
			assert.Equal(t, tt.expectedFormat, params.Format)
			assert.Equal(t, tt.expectedQual, params.Quality)
		})
	}
}

// TestParseImageRequest_GracefulParsing_InvalidValuesUseDefaults tests fallback to defaults
func TestParseImageRequest_GracefulParsing_InvalidValuesUseDefaults(t *testing.T) {
	tests := []struct {
		name           string
		segments       []string
		expectedWidth  int
		expectedHeight int
		expectedFormat string
		expectedQual   int
	}{
		{
			name:           "Invalid dimensions - too large",
			segments:       []string{"99999x1"},
			expectedWidth:  1000, // default
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   75,
		},
		{
			name:           "Invalid dimensions - too small",
			segments:       []string{"5x5"},
			expectedWidth:  1000, // default
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   75,
		},
		{
			name:           "Invalid quality - too high",
			segments:       []string{"q95005"},
			expectedWidth:  1000,
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   75, // default
		},
		{
			name:           "Invalid quality - zero",
			segments:       []string{"q0"},
			expectedWidth:  1000,
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   75,
		},
		{
			name:           "Invalid format",
			segments:       []string{"gif"},
			expectedWidth:  1000,
			expectedHeight: 1000,
			expectedFormat: "webp", // default
			expectedQual:   75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			params := parseParameters(tt.segments)

			// Assert
			assert.Equal(t, tt.expectedWidth, params.Width)
			assert.Equal(t, tt.expectedHeight, params.Height)
			assert.Equal(t, tt.expectedFormat, params.Format)
			assert.Equal(t, tt.expectedQual, params.Quality)
		})
	}
}

// TestParseImageRequest_GracefulParsing_ComplexURL tests complex URL with mixed valid/invalid params
func TestParseImageRequest_GracefulParsing_ComplexURL(t *testing.T) {
	tests := []struct {
		name           string
		segments       []string
		expectedWidth  int
		expectedHeight int
		expectedFormat string
		expectedQual   int
	}{
		{
			name:           "Mixed valid and invalid",
			segments:       []string{"800x600", "webp", "q90", "wow", "extra", "999"},
			expectedWidth:  800,
			expectedHeight: 600,
			expectedFormat: "webp",
			expectedQual:   90,
		},
		{
			name:           "Valid params surrounded by invalid",
			segments:       []string{"invalid1", "300x300", "bad", "png", "wrong", "q85", "extra"},
			expectedWidth:  300,
			expectedHeight: 300,
			expectedFormat: "png",
			expectedQual:   85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			params := parseParameters(tt.segments)

			// Assert
			assert.Equal(t, tt.expectedWidth, params.Width)
			assert.Equal(t, tt.expectedHeight, params.Height)
			assert.Equal(t, tt.expectedFormat, params.Format)
			assert.Equal(t, tt.expectedQual, params.Quality)
		})
	}
}

// TestParseImageRequest_GracefulParsing_ParameterOrder tests parameter order independence
func TestParseImageRequest_GracefulParsing_ParameterOrder(t *testing.T) {
	tests := []struct {
		name     string
		segments []string
	}{
		{
			name:     "Order 1: dimensions, format, quality",
			segments: []string{"800x600", "webp", "q90"},
		},
		{
			name:     "Order 2: quality, dimensions, format",
			segments: []string{"q90", "800x600", "webp"},
		},
		{
			name:     "Order 3: format, quality, dimensions",
			segments: []string{"webp", "q90", "800x600"},
		},
	}

	expectedWidth := 800
	expectedHeight := 600
	expectedFormat := "webp"
	expectedQual := 90

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			params := parseParameters(tt.segments)

			// Assert - all orders should produce same result
			assert.Equal(t, expectedWidth, params.Width)
			assert.Equal(t, expectedHeight, params.Height)
			assert.Equal(t, expectedFormat, params.Format)
			assert.Equal(t, expectedQual, params.Quality)
		})
	}
}

// TestParseImageRequest_GracefulParsing_EdgeCases tests edge cases and boundary values
func TestParseImageRequest_GracefulParsing_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		segments       []string
		expectedWidth  int
		expectedHeight int
		expectedFormat string
		expectedQual   int
	}{
		{
			name:           "Empty segments",
			segments:       []string{},
			expectedWidth:  1000,
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   75,
		},
		{
			name:           "Minimum valid dimensions",
			segments:       []string{"10x10"},
			expectedWidth:  10,
			expectedHeight: 10,
			expectedFormat: "webp",
			expectedQual:   75,
		},
		{
			name:           "Maximum valid dimensions",
			segments:       []string{"4000x4000"},
			expectedWidth:  4000,
			expectedHeight: 4000,
			expectedFormat: "webp",
			expectedQual:   75,
		},
		{
			name:           "Minimum quality",
			segments:       []string{"q1"},
			expectedWidth:  1000,
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   1,
		},
		{
			name:           "Maximum quality",
			segments:       []string{"q100"},
			expectedWidth:  1000,
			expectedHeight: 1000,
			expectedFormat: "webp",
			expectedQual:   100,
		},
		{
			name:           "jpg format (alternative)",
			segments:       []string{"jpg"},
			expectedWidth:  1000,
			expectedHeight: 1000,
			expectedFormat: "jpg",
			expectedQual:   75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			params := parseParameters(tt.segments)

			// Assert
			assert.Equal(t, tt.expectedWidth, params.Width)
			assert.Equal(t, tt.expectedHeight, params.Height)
			assert.Equal(t, tt.expectedFormat, params.Format)
			assert.Equal(t, tt.expectedQual, params.Quality)
		})
	}
}

// TestParseImageRequest_GracefulParsing_Performance tests parsing performance with many parameters
func TestParseImageRequest_GracefulParsing_Performance(t *testing.T) {
	// Arrange
	segments := make([]string, 100)
	for i := 0; i < 100; i++ {
		if i%3 == 0 {
			segments[i] = "invalid"
		} else if i%3 == 1 {
			segments[i] = "extra"
		} else {
			segments[i] = "wrong"
		}
	}
	// Add some valid parameters in the middle
	segments[50] = "800x600"
	segments[51] = "webp"
	segments[52] = "q90"

	// Act
	params := parseParameters(segments)

	// Assert
	assert.Equal(t, 800, params.Width)
	assert.Equal(t, 600, params.Height)
	assert.Equal(t, "webp", params.Format)
	assert.Equal(t, 90, params.Quality)
}

// TestParseClearCommand tests parsing of clear command
func TestParseClearCommand(t *testing.T) {
	tests := []struct {
		name        string
		segments    []string
		expectClear bool
	}{
		{
			name:        "Clear command present",
			segments:    []string{"clear"},
			expectClear: true,
		},
		{
			name:        "Clear command with other params",
			segments:    []string{"800x600", "clear"},
			expectClear: true,
		},
		{
			name:        "No clear command",
			segments:    []string{"800x600", "webp"},
			expectClear: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isClear := hasClearCommand(tt.segments)

			// Assert
			assert.Equal(t, tt.expectClear, isClear)
		})
	}
}
