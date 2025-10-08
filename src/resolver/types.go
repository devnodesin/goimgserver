package resolver

import "errors"

// ResolutionResult represents the result of file resolution
type ResolutionResult struct {
	ResolvedPath string
	IsGrouped    bool
	IsFallback   bool
	FallbackType string
}

// FileResolver provides file resolution services
type FileResolver interface {
	Resolve(requestPath string) (*ResolutionResult, error)
	ResolveWithDefault(requestPath string, defaultPath string) (*ResolutionResult, error)
}

// Common errors
var (
	ErrInvalidPath     = errors.New("invalid path")
	ErrPathTraversal   = errors.New("path traversal attempt detected")
	ErrOutsideImageDir = errors.New("resolved path is outside image directory")
	ErrFileNotFound    = errors.New("file not found")
)
