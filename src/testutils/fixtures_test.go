package testutils

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelpers_ImageFixtures_Creation tests image fixture generation
func TestHelpers_ImageFixtures_Creation(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		format      ImageFormat
		shouldError bool
	}{
		{
			name:        "Small JPEG",
			width:       100,
			height:      100,
			format:      FormatJPEG,
			shouldError: false,
		},
		{
			name:        "Medium PNG",
			width:       500,
			height:      500,
			format:      FormatPNG,
			shouldError: false,
		},
		{
			name:        "Large JPEG",
			width:       2000,
			height:      1500,
			format:      FormatJPEG,
			shouldError: false,
		},
		{
			name:        "Invalid dimensions - zero width",
			width:       0,
			height:      100,
			format:      FormatJPEG,
			shouldError: true,
		},
		{
			name:        "Invalid dimensions - negative height",
			width:       100,
			height:      -100,
			format:      FormatJPEG,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "test_image.jpg")

			err := CreateTestImage(filename, tt.width, tt.height, tt.format)

			if tt.shouldError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			
			// Verify file exists
			info, err := os.Stat(filename)
			require.NoError(t, err)
			assert.Greater(t, info.Size(), int64(0))

			// Verify image can be decoded
			f, err := os.Open(filename)
			require.NoError(t, err)
			defer f.Close()

			img, _, err := image.Decode(f)
			require.NoError(t, err)
			assert.NotNil(t, img)
			
			// Verify dimensions
			bounds := img.Bounds()
			assert.Equal(t, tt.width, bounds.Dx())
			assert.Equal(t, tt.height, bounds.Dy())
		})
	}
}

// TestFixtureManager_CreateFixtureSet tests creating a set of test fixtures
func TestFixtureManager_CreateFixtureSet(t *testing.T) {
	tmpDir := t.TempDir()
	
	manager := NewFixtureManager(tmpDir)
	
	err := manager.CreateFixtureSet()
	require.NoError(t, err)
	
	// Verify standard fixtures were created
	standardFixtures := []string{
		"small_test.jpg",
		"medium_test.jpg",
		"large_test.jpg",
		"test.png",
		"test.webp",
	}
	
	for _, fixture := range standardFixtures {
		path := filepath.Join(tmpDir, fixture)
		info, err := os.Stat(path)
		assert.NoError(t, err, "Fixture %s should exist", fixture)
		if err == nil {
			assert.Greater(t, info.Size(), int64(0), "Fixture %s should not be empty", fixture)
		}
	}
}

// TestFixtureManager_GetFixturePath tests fixture path retrieval
func TestFixtureManager_GetFixturePath(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewFixtureManager(tmpDir)
	
	err := manager.CreateFixtureSet()
	require.NoError(t, err)
	
	tests := []struct {
		name          string
		fixtureName   string
		shouldExist   bool
	}{
		{
			name:        "Existing fixture",
			fixtureName: "small_test.jpg",
			shouldExist: true,
		},
		{
			name:        "Non-existing fixture",
			fixtureName: "nonexistent.jpg",
			shouldExist: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := manager.GetFixturePath(tt.fixtureName)
			
			if tt.shouldExist {
				assert.FileExists(t, path)
			} else {
				_, err := os.Stat(path)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

// TestFixtureManager_Cleanup tests fixture cleanup
func TestFixtureManager_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewFixtureManager(tmpDir)
	
	err := manager.CreateFixtureSet()
	require.NoError(t, err)
	
	// Verify fixtures exist
	path := manager.GetFixturePath("small_test.jpg")
	assert.FileExists(t, path)
	
	// Cleanup
	err = manager.Cleanup()
	require.NoError(t, err)
	
	// Verify fixtures are removed
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

// TestCreateColoredImage tests creating images with specific colors
func TestCreateColoredImage(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		color  color.Color
	}{
		{
			name:   "Red image",
			width:  100,
			height: 100,
			color:  color.RGBA{R: 255, A: 255},
		},
		{
			name:   "Blue image",
			width:  200,
			height: 150,
			color:  color.RGBA{B: 255, A: 255},
		},
		{
			name:   "Green image",
			width:  150,
			height: 150,
			color:  color.RGBA{G: 255, A: 255},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "colored.jpg")
			
			err := CreateColoredImage(filename, tt.width, tt.height, tt.color, FormatJPEG)
			require.NoError(t, err)
			
			// Verify image
			f, err := os.Open(filename)
			require.NoError(t, err)
			defer f.Close()
			
			img, _, err := image.Decode(f)
			require.NoError(t, err)
			
			bounds := img.Bounds()
			assert.Equal(t, tt.width, bounds.Dx())
			assert.Equal(t, tt.height, bounds.Dy())
		})
	}
}

// TestCreateTestImageWithText tests creating images with text
func TestCreateTestImageWithText(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "text_image.jpg")
	
	err := CreateTestImageWithText(filename, 300, 200, "Test Image", FormatJPEG)
	require.NoError(t, err)
	
	// Verify file exists and has content
	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
	
	// Verify image can be decoded
	f, err := os.Open(filename)
	require.NoError(t, err)
	defer f.Close()
	
	img, _, err := image.Decode(f)
	require.NoError(t, err)
	assert.NotNil(t, img)
}

// TestImageFormatValidation tests image format validation
func TestImageFormatValidation(t *testing.T) {
	tests := []struct {
		name        string
		format      ImageFormat
		shouldError bool
	}{
		{
			name:        "Valid JPEG format",
			format:      FormatJPEG,
			shouldError: false,
		},
		{
			name:        "Valid PNG format",
			format:      FormatPNG,
			shouldError: false,
		},
		{
			name:        "Valid WebP format",
			format:      FormatWebP,
			shouldError: false,
		},
		{
			name:        "Invalid format",
			format:      ImageFormat("invalid"),
			shouldError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageFormat(tt.format)
			
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
