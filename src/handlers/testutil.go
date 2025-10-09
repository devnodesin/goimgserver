package handlers

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
)

// createTestImage creates a simple test image with specified dimensions and text
func createTestImage(path string, width, height int) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	// Save as JPEG
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
}
