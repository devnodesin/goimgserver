package precache

import (
	"context"
	"fmt"
	"goimgserver/cache"
	"goimgserver/resolver"
	"os"
	"path/filepath"
)

// ProcessorInterface defines the minimal interface for image processing
type ProcessorInterface interface {
	Process(data []byte, opts interface{}) ([]byte, error)
}

// preCacheProcessor implements Processor interface
type preCacheProcessor struct {
	imageDir  string
	resolver  resolver.FileResolver
	cache     cache.CacheManager
	processor ProcessorInterface
}

// NewProcessor creates a new pre-cache processor
func NewProcessor(imageDir string, fileResolver resolver.FileResolver, cacheManager cache.CacheManager, processor ProcessorInterface) Processor {
	return &preCacheProcessor{
		imageDir:  imageDir,
		resolver:  fileResolver,
		cache:     cacheManager,
		processor: processor,
	}
}

// Process processes a single image and stores it in cache
func (p *preCacheProcessor) Process(ctx context.Context, imagePath string) error {
	// Get relative path from image directory
	relPath, err := filepath.Rel(p.imageDir, imagePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}
	
	// Resolve the file path
	result, err := p.resolver.Resolve(relPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	
	// Define default pre-cache parameters (1000x1000, WebP, q95)
	params := cache.ProcessingParams{
		Width:   1000,
		Height:  1000,
		Format:  "webp",
		Quality: 95,
	}
	
	// Check if already cached
	if p.cache.Exists(result.ResolvedPath, params) {
		// Skip already cached images
		return nil
	}
	
	// Read the image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}
	
	// Process the image with default settings
	processedData, err := p.processor.Process(imageData, params)
	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}
	
	// Store in cache
	err = p.cache.Store(result.ResolvedPath, params, processedData)
	if err != nil {
		return fmt.Errorf("failed to store in cache: %w", err)
	}
	
	return nil
}
