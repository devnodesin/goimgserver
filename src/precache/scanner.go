package precache

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

var supportedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// directoryScanner implements Scanner interface
type directoryScanner struct{}

// NewScanner creates a new directory scanner
func NewScanner() Scanner {
	return &directoryScanner{}
}

// Scan scans the given directory and returns a list of image paths
func (s *directoryScanner) Scan(ctx context.Context, imageDir string, defaultImagePath string) ([]string, error) {
	var images []string
	
	// Normalize default image path for comparison
	defaultImagePath = filepath.Clean(defaultImagePath)
	
	err := filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Check if file has supported extension
		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExtensions[ext] {
			return nil
		}
		
		// Skip the system default image if specified
		if defaultImagePath != "" && filepath.Clean(path) == defaultImagePath {
			return nil
		}
		
		images = append(images, path)
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return images, nil
}
