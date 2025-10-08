package cache

import (
	"time"
)

// CacheManager defines the interface for cache operations
type CacheManager interface {
	// GenerateKey creates a cache key from resolved file path and processing parameters
	GenerateKey(resolvedPath string, params ProcessingParams) string

	// Store saves processed image data to cache with atomic operations
	Store(resolvedPath string, params ProcessingParams, data []byte) error

	// Retrieve fetches cached image data if it exists
	Retrieve(resolvedPath string, params ProcessingParams) ([]byte, bool, error)

	// Exists checks if a cached file exists
	Exists(resolvedPath string, params ProcessingParams) bool

	// Clear removes cached files for a specific resolved path
	Clear(resolvedPath string) error

	// ClearAll removes all cached files
	ClearAll() error

	// GetPath returns the cache path for given parameters
	GetPath(resolvedPath string, params ProcessingParams) string

	// GetStats returns cache statistics
	GetStats() (*Stats, error)
}

// ProcessingParams represents normalized image processing parameters
type ProcessingParams struct {
	Width   int
	Height  int
	Format  string
	Quality int
}

// Stats contains cache statistics
type Stats struct {
	TotalFiles     int64
	TotalSize      int64
	HitCount       int64
	MissCount      int64
	LastClearTime  time.Time
	OldestFileTime time.Time
	NewestFileTime time.Time
}

// Metadata contains cache file metadata
type Metadata struct {
	CreatedAt time.Time
	Size      int64
	Params    ProcessingParams
	Hash      string
}
