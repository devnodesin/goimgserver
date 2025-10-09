package resolver

import (
	"path/filepath"
	"strings"
)

// sanitizePath cleans and validates a path to prevent security issues
func sanitizePath(requestPath string, imageDir string) (string, error) {
	// Remove null bytes
	if strings.Contains(requestPath, "\x00") {
		return "", ErrInvalidPath
	}
	
	// Clean the path (removes .. and . components)
	cleanPath := filepath.Clean(requestPath)
	
	// Prevent absolute paths
	if filepath.IsAbs(cleanPath) {
		return "", ErrInvalidPath
	}
	
	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return "", ErrPathTraversal
	}
	
	// Build the full path
	fullPath := filepath.Join(imageDir, cleanPath)
	
	// Verify the resolved path is within the image directory
	absImageDir, err := filepath.Abs(imageDir)
	if err != nil {
		return "", err
	}
	
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}
	
	// Ensure the path starts with the image directory
	if !filepath.HasPrefix(absFullPath, absImageDir) {
		return "", ErrOutsideImageDir
	}
	
	return cleanPath, nil
}

// validateResolvedPath ensures a resolved path is within the image directory
// This is especially important for symlinks
func validateResolvedPath(resolvedPath string, imageDir string) error {
	absImageDir, err := filepath.Abs(imageDir)
	if err != nil {
		return err
	}
	
	// Resolve any symlinks
	realPath, err := filepath.EvalSymlinks(resolvedPath)
	if err != nil {
		// If symlink evaluation fails, check the path itself
		realPath = resolvedPath
	}
	
	absResolved, err := filepath.Abs(realPath)
	if err != nil {
		return err
	}
	
	// Ensure the resolved path is within image directory
	if !filepath.HasPrefix(absResolved, absImageDir) {
		return ErrOutsideImageDir
	}
	
	return nil
}
