package resolver

import (
	"os"
	"path/filepath"
	"strings"
)

// Resolver implements FileResolver interface
type Resolver struct {
	imageDir string
	cache    *Cache
}

// NewResolver creates a new file resolver
func NewResolver(imageDir string) *Resolver {
	return &Resolver{
		imageDir: imageDir,
	}
}

// NewResolverWithCache creates a new file resolver with caching enabled
func NewResolverWithCache(imageDir string) *Resolver {
	return &Resolver{
		imageDir: imageDir,
		cache:    NewCache(),
	}
}

// Resolve resolves a request path to an actual file path
func (r *Resolver) Resolve(requestPath string) (*ResolutionResult, error) {
	// Check cache if available
	if r.cache != nil {
		if result, found := r.cache.Get(requestPath); found {
			return result, nil
		}
	}
	
	// Extension priority order
	extensions := []string{".jpg", ".jpeg", ".png", ".webp"}
	
	// Sanitize and validate the request path
	cleanPath, err := sanitizePath(requestPath, r.imageDir)
	if err != nil {
		// On security error, fall back to system default
		result, sysErr := r.resolveSystemDefault()
		if sysErr == nil && r.cache != nil {
			r.cache.Set(requestPath, result)
		}
		return result, sysErr
	}
	
	// Check if path has extension
	ext := filepath.Ext(cleanPath)
	hasExtension := ext != ""
	
	// Determine if this is a grouped path
	segments := strings.Split(cleanPath, string(filepath.Separator))
	isGrouped := len(segments) > 1
	
	if hasExtension {
		// Direct path with extension
		fullPath := filepath.Join(r.imageDir, cleanPath)
		if fileExists(fullPath) {
			// Validate the resolved path (for symlinks)
			if err := validateResolvedPath(fullPath, r.imageDir); err != nil {
				result, fbErr := r.resolveFallback(cleanPath, isGrouped)
				if fbErr == nil && r.cache != nil {
					r.cache.Set(requestPath, result)
				}
				return result, fbErr
			}
			result := &ResolutionResult{
				ResolvedPath: fullPath,
				IsGrouped:    isGrouped,
				IsFallback:   false,
			}
			if r.cache != nil {
				r.cache.Set(requestPath, result)
			}
			return result, nil
		}
		
		// If file doesn't exist, try fallback
		result, err := r.resolveFallback(cleanPath, isGrouped)
		if err == nil && r.cache != nil {
			r.cache.Set(requestPath, result)
		}
		return result, err
	}
	
	// No extension - try auto-detection
	basePath := cleanPath
	
	// Check if this might be a group (directory exists)
	groupPath := filepath.Join(r.imageDir, basePath)
	if dirExists(groupPath) {
		// Try to resolve group default
		for _, ext := range extensions {
			defaultPath := filepath.Join(groupPath, "default"+ext)
			if fileExists(defaultPath) {
				result := &ResolutionResult{
					ResolvedPath: defaultPath,
					IsGrouped:    true,
					IsFallback:   false,
				}
				if r.cache != nil {
					r.cache.Set(requestPath, result)
				}
				return result, nil
			}
		}
		// Group exists but no default found - fallback to system default
		result, err := r.resolveSystemDefault()
		if err == nil && r.cache != nil {
			r.cache.Set(requestPath, result)
		}
		return result, err
	}
	
	// Try to find file with extension priority
	for _, ext := range extensions {
		testPath := filepath.Join(r.imageDir, basePath+ext)
		if fileExists(testPath) {
			// Validate the resolved path (for symlinks)
			if err := validateResolvedPath(testPath, r.imageDir); err != nil {
				continue
			}
			result := &ResolutionResult{
				ResolvedPath: testPath,
				IsGrouped:    isGrouped,
				IsFallback:   false,
			}
			if r.cache != nil {
				r.cache.Set(requestPath, result)
			}
			return result, nil
		}
	}
	
	// Not found - try fallback
	result, err := r.resolveFallback(basePath, isGrouped)
	if err == nil && r.cache != nil {
		r.cache.Set(requestPath, result)
	}
	return result, err
}

// resolveFallback handles fallback resolution
func (r *Resolver) resolveFallback(requestPath string, isGrouped bool) (*ResolutionResult, error) {
	extensions := []string{".jpg", ".jpeg", ".png", ".webp"}
	
	// If grouped, try group default first
	if isGrouped {
		segments := strings.Split(requestPath, string(filepath.Separator))
		if len(segments) > 1 {
			groupName := segments[0]
			groupPath := filepath.Join(r.imageDir, groupName)
			
			// Try group default
			for _, ext := range extensions {
				defaultPath := filepath.Join(groupPath, "default"+ext)
				if fileExists(defaultPath) {
					return &ResolutionResult{
						ResolvedPath: defaultPath,
						IsGrouped:    true,
						IsFallback:   true,
						FallbackType: "group_default",
					}, nil
				}
			}
		}
	}
	
	// Fall back to system default
	return r.resolveSystemDefault()
}

// resolveSystemDefault resolves to the system default image
func (r *Resolver) resolveSystemDefault() (*ResolutionResult, error) {
	extensions := []string{".jpg", ".jpeg", ".png", ".webp"}
	
	for _, ext := range extensions {
		defaultPath := filepath.Join(r.imageDir, "default"+ext)
		if fileExists(defaultPath) {
			return &ResolutionResult{
				ResolvedPath: defaultPath,
				IsGrouped:    false,
				IsFallback:   true,
				FallbackType: "system_default",
			}, nil
		}
	}
	
	return nil, ErrFileNotFound
}

// ResolveWithDefault resolves with a specific default path
func (r *Resolver) ResolveWithDefault(requestPath string, defaultPath string) (*ResolutionResult, error) {
	result, err := r.Resolve(requestPath)
	if err == nil && !result.IsFallback {
		return result, nil
	}
	
	// Use provided default
	if fileExists(defaultPath) {
		return &ResolutionResult{
			ResolvedPath: defaultPath,
			IsGrouped:    false,
			IsFallback:   true,
			FallbackType: "provided_default",
		}, nil
	}
	
	return nil, ErrFileNotFound
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
