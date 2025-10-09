package precache

import (
	"context"
	"fmt"
	"goimgserver/cache"
	"goimgserver/resolver"
	"log"
	"runtime"
)

// PreCache is the main pre-cache coordinator
type PreCache struct {
	config   *PreCacheConfig
	scanner  Scanner
	executor *ConcurrentExecutor
}

// New creates a new PreCache instance
func New(config *PreCacheConfig, fileResolver resolver.FileResolver, cacheManager cache.CacheManager, processor ProcessorInterface) (*PreCache, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	// Set default workers if not specified
	if config.Workers <= 0 {
		config.Workers = runtime.NumCPU()
	}
	
	scanner := NewScanner()
	preCacheProcessor := NewProcessor(config.ImageDir, fileResolver, cacheManager, processor)
	progress := NewProgress()
	executor := NewConcurrentExecutor(preCacheProcessor, config.Workers, progress)
	
	return &PreCache{
		config:   config,
		scanner:  scanner,
		executor: executor,
	}, nil
}

// Run executes the pre-cache process
func (p *PreCache) Run(ctx context.Context) (*Stats, error) {
	if !p.config.Enabled {
		log.Println("Pre-cache disabled, skipping")
		return &Stats{}, nil
	}
	
	log.Printf("Starting pre-cache: scanning %s", p.config.ImageDir)
	
	// Scan for images
	imagePaths, err := p.scanner.Scan(ctx, p.config.ImageDir, p.config.DefaultImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}
	
	log.Printf("Found %d images to pre-cache", len(imagePaths))
	
	if len(imagePaths) == 0 {
		log.Println("No images found, pre-cache complete")
		return &Stats{}, nil
	}
	
	// Execute pre-caching with concurrent processing
	stats, err := p.executor.Execute(ctx, imagePaths)
	if err != nil {
		return stats, fmt.Errorf("pre-cache execution failed: %w", err)
	}
	
	return stats, nil
}

// RunAsync executes the pre-cache process asynchronously
func (p *PreCache) RunAsync(ctx context.Context) {
	go func() {
		_, err := p.Run(ctx)
		if err != nil {
			log.Printf("Pre-cache async error: %v", err)
		}
	}()
}
