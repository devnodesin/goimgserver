package precache

import (
	"context"
	"errors"
	"time"
)

// Common errors
var (
	ErrInvalidOptions = errors.New("invalid options type")
)

// Scanner scans directories for images to pre-cache
type Scanner interface {
	// Scan scans the given directory and returns a list of image paths
	Scan(ctx context.Context, imageDir string, defaultImagePath string) ([]string, error)
}

// Processor processes images and stores them in cache
type Processor interface {
	// Process processes a single image and stores it in cache
	Process(ctx context.Context, imagePath string) error
}

// ProgressReporter reports progress during pre-caching
type ProgressReporter interface {
	// Start begins progress tracking
	Start(total int)
	
	// Update updates progress with processed count
	Update(processed int, current string)
	
	// Complete marks progress as complete
	Complete(processed int, skipped int, errors int, duration time.Duration)
	
	// Error reports an error
	Error(imagePath string, err error)
}

// PreCacheConfig contains configuration for pre-caching
type PreCacheConfig struct {
	ImageDir         string
	CacheDir         string
	DefaultImagePath string
	Enabled          bool
	Workers          int
}

// Stats contains pre-cache statistics
type Stats struct {
	TotalImages   int
	ProcessedOK   int
	Skipped       int
	Errors        int
	Duration      time.Duration
	StartTime     time.Time
	EndTime       time.Time
}
