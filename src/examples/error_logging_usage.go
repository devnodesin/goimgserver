package main

import (
	"fmt"
	"goimgserver/errors"
	"goimgserver/logging"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Example demonstrating the new error handling and logging system
func main() {
	// Setup logging
	config := logging.DevelopmentConfig()
	logger := logging.NewLoggerFromConfig(os.Stdout, config)
	perfLogger := logging.NewPerformanceLogger(logger)

	// Create Gin router
	router := gin.New()

	// Middleware stack
	router.Use(func(c *gin.Context) {
		// Generate request ID
		requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
		c.Set("request_id", requestID)
		c.Next()
	})

	router.Use(errors.ErrorHandlerMiddleware())

	router.Use(func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()
		ctx = perfLogger.StartOperation(ctx, "handle_request")
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		perfLogger.EndOperation(ctx, "handle_request", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": c.Writer.Status(),
		})

		logger.InfoWithFields("request completed",
			"request_id", c.GetString("request_id"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})

	// Example handlers demonstrating error handling

	// Success case
	router.GET("/api/image/:filename", func(c *gin.Context) {
		filename := c.Param("filename")

		logger.InfoWithFields("processing image request",
			"filename", filename,
			"request_id", c.GetString("request_id"),
		)

		c.JSON(200, gin.H{
			"filename": filename,
			"status":   "processed",
		})
	})

	// Not found error
	router.GET("/api/error/notfound", func(c *gin.Context) {
		err := errors.NewImageNotFoundError("missing.jpg")
		logger.ErrorWithFields("image not found",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
		)
		errors.HandleError(c, err)
	})

	// Validation error
	router.GET("/api/error/validation", func(c *gin.Context) {
		valErr := errors.NewValidationError("Invalid request parameters")
		valErr = valErr.AddFieldError("width", "must be positive")
		valErr = valErr.AddFieldError("height", "must be positive")

		logger.WarnWithFields("validation failed",
			"request_id", c.GetString("request_id"),
		)
		errors.HandleError(c, valErr)
	})

	// Performance tracking example
	router.GET("/api/performance", func(c *gin.Context) {
		ctx := c.Request.Context()

		// Track multiple operations
		ctx = perfLogger.StartOperation(ctx, "load_image")
		time.Sleep(5 * time.Millisecond) // Simulate work
		perfLogger.EndOperation(ctx, "load_image", map[string]interface{}{
			"filename": "test.jpg",
		})

		ctx = perfLogger.StartOperation(ctx, "resize_image")
		time.Sleep(10 * time.Millisecond) // Simulate work
		perfLogger.EndOperation(ctx, "resize_image", map[string]interface{}{
			"width":  800,
			"height": 600,
		})

		// Log metrics
		perfLogger.LogMetrics()

		c.JSON(200, gin.H{"status": "completed"})
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		logger.Debug("health check requested")
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Start server
	logger.InfoWithFields("starting server",
		"port", 8080,
		"environment", "development",
	)

	if err := router.Run(":8080"); err != nil {
		logger.ErrorWithFields("server failed to start",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
