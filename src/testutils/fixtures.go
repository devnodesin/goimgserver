package testutils

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

// ImageFormat represents supported image formats for testing
type ImageFormat string

const (
	FormatJPEG ImageFormat = "jpeg"
	FormatPNG  ImageFormat = "png"
	FormatWebP ImageFormat = "webp"
)

// ValidateImageFormat validates if the image format is supported
func ValidateImageFormat(format ImageFormat) error {
	switch format {
	case FormatJPEG, FormatPNG, FormatWebP:
		return nil
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
}

// CreateTestImage creates a test image with the specified dimensions and format
func CreateTestImage(filename string, width, height int, format ImageFormat) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}

	if err := ValidateImageFormat(format); err != nil {
		return err
	}

	// Create a simple colored image
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Fill with a gradient pattern for visual verification
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return saveImage(filename, img, format)
}

// CreateColoredImage creates a solid color image
func CreateColoredImage(filename string, width, height int, col color.Color, format ImageFormat) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}

	if err := ValidateImageFormat(format); err != nil {
		return err
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}

	return saveImage(filename, img, format)
}

// CreateTestImageWithText creates an image with text overlay
// Note: Basic implementation without font rendering for simplicity
func CreateTestImageWithText(filename string, width, height int, text string, format ImageFormat) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}

	if err := ValidateImageFormat(format); err != nil {
		return err
	}

	// Create white background
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	
	// Fill background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, white)
		}
	}
	
	// Draw a simple pattern in center (simulating text)
	centerX, centerY := width/2, height/2
	size := 20
	for y := centerY - size; y < centerY + size; y++ {
		for x := centerX - size; x < centerX + size; x++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				img.Set(x, y, black)
			}
		}
	}

	return saveImage(filename, img, format)
}

// saveImage saves an image to disk in the specified format
func saveImage(filename string, img image.Image, format ImageFormat) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	switch format {
	case FormatJPEG:
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 95})
	case FormatPNG:
		return png.Encode(f, img)
	case FormatWebP:
		// WebP encoding is more complex, for now we'll save as PNG
		// In production, use a proper WebP encoder
		return png.Encode(f, img)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// FixtureManager manages test image fixtures
type FixtureManager struct {
	baseDir string
}

// NewFixtureManager creates a new fixture manager
func NewFixtureManager(baseDir string) *FixtureManager {
	return &FixtureManager{
		baseDir: baseDir,
	}
}

// CreateFixtureSet creates a standard set of test fixtures
func (fm *FixtureManager) CreateFixtureSet() error {
	fixtures := []struct {
		name   string
		width  int
		height int
		format ImageFormat
	}{
		{"small_test.jpg", 100, 100, FormatJPEG},
		{"medium_test.jpg", 500, 500, FormatJPEG},
		{"large_test.jpg", 2000, 1500, FormatJPEG},
		{"test.png", 300, 300, FormatPNG},
		{"test.webp", 300, 300, FormatWebP},
	}

	for _, fixture := range fixtures {
		path := filepath.Join(fm.baseDir, fixture.name)
		if err := CreateTestImage(path, fixture.width, fixture.height, fixture.format); err != nil {
			return fmt.Errorf("failed to create fixture %s: %w", fixture.name, err)
		}
	}

	return nil
}

// GetFixturePath returns the full path to a fixture
func (fm *FixtureManager) GetFixturePath(name string) string {
	return filepath.Join(fm.baseDir, name)
}

// Cleanup removes all fixtures
func (fm *FixtureManager) Cleanup() error {
	return os.RemoveAll(fm.baseDir)
}
