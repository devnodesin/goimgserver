package config

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// DetectDefaultImage scans the directory for a default image file
// Returns the path and whether it was found
// Priority: default.jpg -> default.jpeg -> default.png -> default.webp
func DetectDefaultImage(dir string) (string, bool) {
	extensions := []string{".jpg", ".jpeg", ".png", ".webp"}

	for _, ext := range extensions {
		path := filepath.Join(dir, "default"+ext)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}

	return "", false
}

// GenerateDefaultPlaceholder creates a 1000x1000px placeholder image
// with white background and "goimgserver" text in black
func GenerateDefaultPlaceholder(outputPath string) error {
	const (
		width  = 1000
		height = 1000
	)

	// Create image with white background
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, white)
		}
	}

	// Draw text
	text := "goimgserver"
	
	// Use basicfont for simplicity (it's part of golang.org/x/image/font)
	// For better appearance, we'll draw the text multiple times to make it larger
	point := fixed.Point26_6{
		X: fixed.Int26_6(width/2 - 200) * 64,
		Y: fixed.Int26_6(height/2 + 30) * 64,
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{0, 0, 0, 255}),
		Face: basicfont.Face7x13,
		Dot:  point,
	}

	// Draw text larger by drawing multiple times with offset
	scale := 10
	for i := 0; i < scale; i++ {
		for j := 0; j < scale; j++ {
			drawer.Dot = fixed.Point26_6{
				X: (fixed.Int26_6(width/2-200) + fixed.Int26_6(i)) * 64,
				Y: (fixed.Int26_6(height/2+30) + fixed.Int26_6(j)) * 64,
			}
			drawer.DrawString(text)
		}
	}

	// Save as JPEG
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	opts := &jpeg.Options{Quality: 95}
	if err := jpeg.Encode(file, img, opts); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

// ValidateDefaultImage checks if the image file is readable and valid
func ValidateDefaultImage(path string) error {
	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("default image not accessible: %w", err)
	}

	// Check if file is not empty
	if info.Size() == 0 {
		return fmt.Errorf("default image is empty")
	}

	// Try to open and read the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open default image: %w", err)
	}
	defer file.Close()

	// Try to decode as image (basic validation)
	_, _, err = image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("default image is not a valid image file: %w", err)
	}

	return nil
}

// SetupDefaultImage detects or generates the default image
func (c *Config) SetupDefaultImage() error {
	// Try to detect existing default image
	if path, found := DetectDefaultImage(c.ImagesDir); found {
		// Validate the found image
		if err := ValidateDefaultImage(path); err != nil {
			return fmt.Errorf("found default image is invalid: %w", err)
		}
		c.DefaultImagePath = path
		return nil
	}

	// Generate placeholder if not found
	defaultPath := filepath.Join(c.ImagesDir, "default.jpg")
	if err := GenerateDefaultPlaceholder(defaultPath); err != nil {
		return fmt.Errorf("failed to generate default placeholder: %w", err)
	}

	c.DefaultImagePath = defaultPath
	return nil
}
