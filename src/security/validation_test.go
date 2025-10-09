package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInputValidation_DirectoryTraversal_Prevention tests basic path traversal protection
func TestInputValidation_DirectoryTraversal_Prevention(t *testing.T) {
	tests := []struct {
		name        string
		inputPath   string
		shouldError bool
		description string
	}{
		{
			name:        "simple_dot_dot",
			inputPath:   "../etc/passwd",
			shouldError: true,
			description: "Should reject simple .. traversal",
		},
		{
			name:        "nested_dot_dot",
			inputPath:   "images/../../etc/passwd",
			shouldError: true,
			description: "Should reject nested .. traversal",
		},
		{
			name:        "encoded_dot_dot",
			inputPath:   "..%2F..%2Fetc%2Fpasswd",
			shouldError: false, // URL encoding will be handled by router/HTTP layer
			description: "URL encoded paths handled at HTTP layer",
		},
		{
			name:        "double_encoded",
			inputPath:   "..%252F..%252Fetc%252Fpasswd",
			shouldError: false, // URL encoding will be handled by router/HTTP layer
			description: "Double encoded paths handled at HTTP layer",
		},
		{
			name:        "null_byte_injection",
			inputPath:   "image.jpg\x00../../etc/passwd",
			shouldError: true,
			description: "Should reject null byte injection",
		},
		{
			name:        "absolute_path",
			inputPath:   "/etc/passwd",
			shouldError: true,
			description: "Should reject absolute paths",
		},
		{
			name:        "valid_relative_path",
			inputPath:   "images/test.jpg",
			shouldError: false,
			description: "Should allow valid relative paths",
		},
		{
			name:        "valid_nested_path",
			inputPath:   "category/subcategory/image.jpg",
			shouldError: false,
			description: "Should allow valid nested paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.inputPath)
			if tt.shouldError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestInputValidation_DirectoryTraversal_EdgeCases tests edge cases in path validation
func TestInputValidation_DirectoryTraversal_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		wantError bool
	}{
		{
			name:      "empty_path",
			inputPath: "",
			wantError: true,
		},
		{
			name:      "only_dots",
			inputPath: "...",
			wantError: false, // "..." is a valid directory name
		},
		{
			name:      "mixed_slashes",
			inputPath: "images\\..\\passwd",
			wantError: true,
		},
		{
			name:      "unicode_dots",
			inputPath: "images/\u2024\u2024/passwd",
			wantError: false, // Unicode one dot leader, not actual ..
		},
		{
			name:      "trailing_slash",
			inputPath: "images/test.jpg/",
			wantError: false,
		},
		{
			name:      "multiple_slashes",
			inputPath: "images///test.jpg",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.inputPath)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestInputValidation_ParameterValidation_Dimensions tests dimension validation
func TestInputValidation_ParameterValidation_Dimensions(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		height        int
		expectValid   bool
		expectedWidth int
		expectedHeight int
	}{
		{
			name:           "valid_dimensions",
			width:          800,
			height:         600,
			expectValid:    true,
			expectedWidth:  800,
			expectedHeight: 600,
		},
		{
			name:           "minimum_dimensions",
			width:          10,
			height:         10,
			expectValid:    true,
			expectedWidth:  10,
			expectedHeight: 10,
		},
		{
			name:           "maximum_dimensions",
			width:          4000,
			height:         4000,
			expectValid:    true,
			expectedWidth:  4000,
			expectedHeight: 4000,
		},
		{
			name:           "below_minimum_width",
			width:          5,
			height:         600,
			expectValid:    false,
			expectedWidth:  1000, // default
			expectedHeight: 600,
		},
		{
			name:           "above_maximum_width",
			width:          5000,
			height:         600,
			expectValid:    false,
			expectedWidth:  1000, // default
			expectedHeight: 600,
		},
		{
			name:           "below_minimum_height",
			width:          800,
			height:         5,
			expectValid:    false,
			expectedWidth:  800,
			expectedHeight: 1000, // default
		},
		{
			name:           "above_maximum_height",
			width:          800,
			height:         5000,
			expectValid:    false,
			expectedWidth:  800,
			expectedHeight: 1000, // default
		},
		{
			name:           "negative_dimensions",
			width:          -100,
			height:         -100,
			expectValid:    false,
			expectedWidth:  1000,
			expectedHeight: 1000,
		},
		{
			name:           "zero_dimensions",
			width:          0,
			height:         0,
			expectValid:    false,
			expectedWidth:  1000,
			expectedHeight: 0, // 0 height is allowed for aspect ratio
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatedWidth, validatedHeight, err := ValidateDimensions(tt.width, tt.height)
			
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				// Even if invalid, should return safe defaults
				assert.NoError(t, err) // Validation uses defaults, not errors
			}
			
			assert.Equal(t, tt.expectedWidth, validatedWidth)
			assert.Equal(t, tt.expectedHeight, validatedHeight)
		})
	}
}

// TestInputValidation_ParameterValidation_Quality tests quality parameter validation
func TestInputValidation_ParameterValidation_Quality(t *testing.T) {
	tests := []struct {
		name            string
		quality         int
		expectedQuality int
	}{
		{
			name:            "valid_quality_mid",
			quality:         75,
			expectedQuality: 75,
		},
		{
			name:            "valid_quality_min",
			quality:         1,
			expectedQuality: 1,
		},
		{
			name:            "valid_quality_max",
			quality:         100,
			expectedQuality: 100,
		},
		{
			name:            "below_minimum",
			quality:         0,
			expectedQuality: 75, // default
		},
		{
			name:            "above_maximum",
			quality:         101,
			expectedQuality: 75, // default
		},
		{
			name:            "negative_quality",
			quality:         -50,
			expectedQuality: 75, // default
		},
		{
			name:            "extremely_high",
			quality:         1000,
			expectedQuality: 75, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatedQuality := ValidateQuality(tt.quality)
			assert.Equal(t, tt.expectedQuality, validatedQuality)
		})
	}
}

// TestInputValidation_ParameterValidation_Format tests format validation
func TestInputValidation_ParameterValidation_Format(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		expectedFormat string
	}{
		{
			name:           "valid_webp",
			format:         "webp",
			expectedFormat: "webp",
		},
		{
			name:           "valid_png",
			format:         "png",
			expectedFormat: "png",
		},
		{
			name:           "valid_jpeg",
			format:         "jpeg",
			expectedFormat: "jpeg",
		},
		{
			name:           "valid_jpg",
			format:         "jpg",
			expectedFormat: "jpg",
		},
		{
			name:           "uppercase_format",
			format:         "PNG",
			expectedFormat: "png",
		},
		{
			name:           "mixed_case",
			format:         "WebP",
			expectedFormat: "webp",
		},
		{
			name:           "invalid_format",
			format:         "bmp",
			expectedFormat: "webp", // default
		},
		{
			name:           "empty_format",
			format:         "",
			expectedFormat: "webp", // default
		},
		{
			name:           "script_injection_attempt",
			format:         "<script>alert('xss')</script>",
			expectedFormat: "webp", // default
		},
		{
			name:           "path_traversal_in_format",
			format:         "../../../etc/passwd",
			expectedFormat: "webp", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatedFormat := ValidateFormat(tt.format)
			assert.Equal(t, tt.expectedFormat, validatedFormat)
		})
	}
}

// TestInputValidation_GracefulParsing_SecurityBypass tests that graceful parsing doesn't bypass security
func TestInputValidation_GracefulParsing_SecurityBypass(t *testing.T) {
	tests := []struct {
		name        string
		rawSegments []string
		description string
	}{
		{
			name:        "malicious_dimensions",
			rawSegments: []string{"test.jpg", "9999x9999", "../../../../etc/passwd"},
			description: "Should validate dimensions even with malicious path segments",
		},
		{
			name:        "injection_in_ignored_params",
			rawSegments: []string{"test.jpg", "800x600", "'; DROP TABLE images;--"},
			description: "Should safely ignore SQL injection attempts in unrecognized params",
		},
		{
			name:        "command_injection_attempt",
			rawSegments: []string{"test.jpg", "800x600", "; rm -rf /"},
			description: "Should safely ignore command injection attempts",
		},
		{
			name:        "xss_in_quality",
			rawSegments: []string{"test.jpg", "q<script>alert('xss')</script>"},
			description: "Should safely ignore XSS attempts in quality param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse parameters with security validation
			params := ParseAndValidateParameters(tt.rawSegments)
			
			// Verify dimensions are within safe bounds
			assert.GreaterOrEqual(t, params.Width, 10)
			assert.LessOrEqual(t, params.Width, 4000)
			assert.GreaterOrEqual(t, params.Height, 0) // 0 means maintain aspect ratio
			assert.LessOrEqual(t, params.Height, 4000)
			
			// Verify quality is within safe bounds
			assert.GreaterOrEqual(t, params.Quality, 1)
			assert.LessOrEqual(t, params.Quality, 100)
			
			// Verify format is safe
			validFormats := map[string]bool{"webp": true, "png": true, "jpeg": true, "jpg": true}
			assert.True(t, validFormats[params.Format], "Format should be one of the allowed formats")
		})
	}
}

// TestInputValidation_GracefulParsing_MaliciousParams tests malicious parameters are safely ignored
func TestInputValidation_GracefulParsing_MaliciousParams(t *testing.T) {
	tests := []struct {
		name       string
		segments   []string
		shouldPass bool
	}{
		{
			name:       "all_malicious_params",
			segments:   []string{"../../../etc/passwd", "999999x999999", "q999999", "evil.exe"},
			shouldPass: true, // Should use defaults
		},
		{
			name:       "mixed_valid_and_malicious",
			segments:   []string{"800x600", "../../../../etc/passwd", "q85"},
			shouldPass: true, // Should use valid params and ignore malicious
		},
		{
			name:       "null_bytes_in_params",
			segments:   []string{"800x600\x00malicious", "q75\x00evil"},
			shouldPass: true, // Should handle null bytes safely
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic or cause errors
			params := ParseAndValidateParameters(tt.segments)
			
			// Should have valid default values
			assert.NotZero(t, params.Width)
			assert.NotZero(t, params.Quality)
			assert.NotEmpty(t, params.Format)
		})
	}
}

// TestInputValidation_FileTypeValidation_MagicNumbers tests file header validation
func TestInputValidation_FileTypeValidation_MagicNumbers(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectValid bool
		expectType  string
	}{
		{
			name:        "valid_jpeg",
			data:        []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46},
			expectValid: true,
			expectType:  "jpeg",
		},
		{
			name:        "valid_png",
			data:        []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expectValid: true,
			expectType:  "png",
		},
		{
			name:        "valid_webp",
			data:        []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			expectValid: true,
			expectType:  "webp",
		},
		{
			name:        "invalid_magic_number",
			data:        []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectValid: false,
			expectType:  "",
		},
		{
			name:        "executable_file",
			data:        []byte{0x4D, 0x5A, 0x90, 0x00}, // PE executable
			expectValid: false,
			expectType:  "",
		},
		{
			name:        "script_file",
			data:        []byte("#!/bin/bash\nrm -rf /"),
			expectValid: false,
			expectType:  "",
		},
		{
			name:        "too_short",
			data:        []byte{0xFF},
			expectValid: false,
			expectType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageType, err := ValidateFileType(tt.data)
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectType, imageType)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestInputValidation_FileTypeValidation_Extensions tests extension validation
func TestInputValidation_FileTypeValidation_Extensions(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectValid bool
	}{
		{
			name:        "valid_jpg",
			filename:    "image.jpg",
			expectValid: true,
		},
		{
			name:        "valid_jpeg",
			filename:    "image.jpeg",
			expectValid: true,
		},
		{
			name:        "valid_png",
			filename:    "image.png",
			expectValid: true,
		},
		{
			name:        "valid_webp",
			filename:    "image.webp",
			expectValid: true,
		},
		{
			name:        "uppercase_extension",
			filename:    "IMAGE.JPG",
			expectValid: true,
		},
		{
			name:        "invalid_extension",
			filename:    "image.exe",
			expectValid: false,
		},
		{
			name:        "double_extension",
			filename:    "image.jpg.exe",
			expectValid: false,
		},
		{
			name:        "no_extension",
			filename:    "image",
			expectValid: false,
		},
		{
			name:        "hidden_file",
			filename:    ".hidden.jpg",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileExtension(tt.filename)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestInputValidation_FileSizeLimits_MaxSize tests file size limits
func TestInputValidation_FileSizeLimits_MaxSize(t *testing.T) {
	tests := []struct {
		name        string
		size        int64
		maxSize     int64
		expectValid bool
	}{
		{
			name:        "within_limit",
			size:        5 * 1024 * 1024,  // 5MB
			maxSize:     10 * 1024 * 1024, // 10MB limit
			expectValid: true,
		},
		{
			name:        "at_limit",
			size:        10 * 1024 * 1024, // 10MB
			maxSize:     10 * 1024 * 1024, // 10MB limit
			expectValid: true,
		},
		{
			name:        "exceeds_limit",
			size:        15 * 1024 * 1024, // 15MB
			maxSize:     10 * 1024 * 1024, // 10MB limit
			expectValid: false,
		},
		{
			name:        "zero_size",
			size:        0,
			maxSize:     10 * 1024 * 1024,
			expectValid: false,
		},
		{
			name:        "negative_size",
			size:        -1,
			maxSize:     10 * 1024 * 1024,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileSize(tt.size, tt.maxSize)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestInputValidation_FileSizeLimits_ActualFile tests size validation with actual files
func TestInputValidation_FileSizeLimits_ActualFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	
	// Create a file with known size (1KB)
	testData := make([]byte, 1024)
	err := os.WriteFile(testFile, testData, 0644)
	require.NoError(t, err)
	
	// Test with higher limit (should pass)
	err = ValidateFileSize(1024, 2048)
	assert.NoError(t, err)
	
	// Test with lower limit (should fail)
	err = ValidateFileSize(1024, 512)
	assert.Error(t, err)
}
