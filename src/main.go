package main

import (
	"context"
	"fmt"
	"goimgserver/cache"
	"goimgserver/config"
	"goimgserver/git"
	"goimgserver/handlers"
	"goimgserver/precache"
	"goimgserver/processor"
	"goimgserver/resolver"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// imageProcessorAdapter adapts processor.ImageProcessor to precache.ProcessorInterface
type imageProcessorAdapter struct {
	processor processor.ImageProcessor
}

// Process adapts the ImageProcessor.Process to match precache.ProcessorInterface
func (a *imageProcessorAdapter) Process(data []byte, opts interface{}) ([]byte, error) {
	// Convert cache.ProcessingParams to processor.ProcessOptions
	params, ok := opts.(cache.ProcessingParams)
	if !ok {
		return nil, fmt.Errorf("invalid options type")
	}
	
	processOpts := processor.ProcessOptions{
		Width:   params.Width,
		Height:  params.Height,
		Format:  processor.ImageFormat(params.Format),
		Quality: params.Quality,
	}
	
	return a.processor.Process(data, processOpts)
}

func main() {
	// Parse command-line arguments
	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to parse arguments: %v", err)
	}

	// Validate configuration and create directories
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Setup default image
	if err := cfg.SetupDefaultImage(); err != nil {
		log.Fatalf("Failed to setup default image: %v", err)
	}

	// Dump settings if requested
	if cfg.Dump {
		pwd, _ := os.Getwd()
		settingsFile := filepath.Join(pwd, "settings.conf")
		if err := cfg.DumpSettings(settingsFile); err != nil {
			log.Fatalf("Failed to dump settings: %v", err)
		}
		fmt.Printf("Settings dumped to: %s\n", settingsFile)
	}

	// Log configuration
	log.Println("Starting goimgserver with configuration:")
	log.Print(cfg.String())

	// Initialize components
	log.Println("Initializing components...")
	
	// Create resolver
	fileResolver := resolver.NewResolverWithCache(cfg.ImagesDir)
	log.Println("File resolver initialized")
	
	// Create cache manager
	cacheManager, err := cache.NewManager(cfg.CacheDir)
	if err != nil {
		log.Fatalf("Failed to create cache manager: %v", err)
	}
	log.Println("Cache manager initialized")
	
	// Create image processor
	imageProcessor := processor.New()
	log.Println("Image processor initialized")
	
	// Create image handler
	imageHandler := handlers.NewImageHandler(cfg, fileResolver, cacheManager, imageProcessor)
	log.Println("Image handler initialized")
	
	// Create git operations
	gitOps := git.NewOperations()
	log.Println("Git operations initialized")
	
	// Create command handler
	commandHandler := handlers.NewCommandHandler(cfg, cacheManager, gitOps)
	log.Println("Command handler initialized")
	
	// Run pre-cache if enabled
	if cfg.PreCacheEnabled {
		log.Println("Starting pre-cache initialization...")
		preCacheConfig := &precache.PreCacheConfig{
			ImageDir:         cfg.ImagesDir,
			CacheDir:         cfg.CacheDir,
			DefaultImagePath: cfg.DefaultImagePath,
			Enabled:          cfg.PreCacheEnabled,
			Workers:          cfg.PreCacheWorkers,
		}
		
		// Create processor adapter for pre-cache (adapts processor.ImageProcessor to precache.ProcessorInterface)
		processorAdapter := &imageProcessorAdapter{processor: imageProcessor}
		
		// Create pre-cache instance
		preCache, err := precache.New(preCacheConfig, fileResolver, cacheManager, processorAdapter)
		if err != nil {
			log.Printf("Warning: Failed to create pre-cache: %v", err)
		} else {
			// Run pre-cache asynchronously to not block server startup
			preCache.RunAsync(context.Background())
		}
	} else {
		log.Println("Pre-cache disabled")
	}

	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	// Set trusted proxies to localhost only (removes Gin warning)
	err = r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		panic(err)
	}

	// Define a simple GET endpoint
	r.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	
	// Image endpoints
	r.GET("/img/*path", imageHandler.ServeImage)
	log.Println("Image endpoints registered")
	
	// Command endpoints
	r.POST("/cmd/clear", commandHandler.HandleClear)
	r.POST("/cmd/gitupdate", commandHandler.HandleGitUpdate)
	r.POST("/cmd/:name", commandHandler.HandleCommand)
	log.Println("Command endpoints registered")

	// Print server startup message
	fmt.Println("Server started and running.")
	fmt.Printf("Server will listen on 127.0.0.1:%d (localhost:%d on Windows)\n", cfg.Port, cfg.Port)
	fmt.Printf("GET http://127.0.0.1:%d/ping to test; you should see message pong.\n", cfg.Port)
	fmt.Printf("Images directory: %s\n", cfg.ImagesDir)
	fmt.Printf("Cache directory: %s\n", cfg.CacheDir)
	fmt.Printf("Default image: %s\n", cfg.DefaultImagePath)

	// Start server on configured port
	addr := fmt.Sprintf(":%d", cfg.Port)
	r.Run(addr)
}
