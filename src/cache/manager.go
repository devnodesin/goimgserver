package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// manager implements the CacheManager interface
type manager struct {
	cacheDir string
	mu       sync.RWMutex
}

// NewManager creates a new cache manager instance
func NewManager(cacheDir string) (CacheManager, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &manager{
		cacheDir: cacheDir,
	}, nil
}

// GenerateKey creates a cache key from resolved file path and processing parameters
func (m *manager) GenerateKey(resolvedPath string, params ProcessingParams) string {
	return generateHash(resolvedPath, params)
}

// Store saves processed image data to cache with atomic operations
func (m *manager) Store(resolvedPath string, params ProcessingParams, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cachePath := m.GetPath(resolvedPath, params)

	// Create directory structure
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write atomically using temporary file
	tempFile := cachePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, cachePath); err != nil {
		os.Remove(tempFile) // Cleanup on failure
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// Retrieve fetches cached image data if it exists
func (m *manager) Retrieve(resolvedPath string, params ProcessingParams) ([]byte, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cachePath := m.GetPath(resolvedPath, params)

	// Check if file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, false, nil
	}

	// Read the file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read cache file: %w", err)
	}

	return data, true, nil
}

// Exists checks if a cached file exists
func (m *manager) Exists(resolvedPath string, params ProcessingParams) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cachePath := m.GetPath(resolvedPath, params)
	_, err := os.Stat(cachePath)
	return err == nil
}

// Clear removes cached files for a specific resolved path
func (m *manager) Clear(resolvedPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the directory for this resolved path
	// Cache structure: {cache_dir}/{filename}/{hash}
	// So we need to remove the entire {cache_dir}/{filename} directory
	pathDir := filepath.Join(m.cacheDir, resolvedPath)

	// Check if directory exists
	if _, err := os.Stat(pathDir); os.IsNotExist(err) {
		// Not an error if directory doesn't exist
		return nil
	}

	// Remove the directory and all its contents
	if err := os.RemoveAll(pathDir); err != nil {
		return fmt.Errorf("failed to clear cache for %s: %w", resolvedPath, err)
	}

	return nil
}

// ClearAll removes all cached files
func (m *manager) ClearAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove all contents of cache directory
	entries, err := os.ReadDir(m.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(m.cacheDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// GetPath returns the cache path for given parameters
func (m *manager) GetPath(resolvedPath string, params ProcessingParams) string {
	hash := m.GenerateKey(resolvedPath, params)

	// Cache structure: {cache_dir}/{filename}/{hash}
	// Clean the resolved path to remove any leading slashes
	cleanPath := strings.TrimPrefix(resolvedPath, "/")

	return filepath.Join(m.cacheDir, cleanPath, hash)
}

// GetStats returns cache statistics
func (m *manager) GetStats() (*Stats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &Stats{
		LastClearTime: time.Time{},
	}

	// Walk the cache directory to gather statistics
	err := filepath.WalkDir(m.cacheDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return nil // Skip files we can't stat
		}

		stats.TotalFiles++
		stats.TotalSize += info.Size()

		// Track oldest and newest files
		modTime := info.ModTime()
		if stats.OldestFileTime.IsZero() || modTime.Before(stats.OldestFileTime) {
			stats.OldestFileTime = modTime
		}
		if stats.NewestFileTime.IsZero() || modTime.After(stats.NewestFileTime) {
			stats.NewestFileTime = modTime
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to gather cache stats: %w", err)
	}

	return stats, nil
}
