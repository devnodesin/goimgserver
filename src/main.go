package main

import (
	"fmt"
	"goimgserver/config"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

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
