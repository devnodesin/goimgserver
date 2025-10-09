package main

import (
	"goimgserver/server"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Example of using the enhanced server package
func main() {
	// Configure the server
	config := &server.Config{
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		EnableCORS:      true,
		EnableRateLimit: true,
		RateLimit:       100,           // 100 requests
		RatePer:         time.Minute,   // per minute
		Production:      false,
	}
	
	// Create server
	srv := server.New(config)
	
	// Add health checks
	srv.AddHealthCheck("database", func() bool {
		// Example: check database connectivity
		return true
	})
	
	srv.AddHealthCheck("cache", func() bool {
		// Example: check cache connectivity
		return true
	})
	
	// Register routes
	srv.Router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to goimgserver!",
		})
	})
	
	srv.Router.GET("/api/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "running",
			"features": []string{
				"CORS enabled",
				"Rate limiting enabled",
				"Security headers",
				"Request ID tracking",
				"Structured logging",
				"Error handling",
				"Health checks",
				"Graceful shutdown",
			},
		})
	})
	
	log.Println("Starting server with enhanced middleware...")
	log.Println("Visit http://localhost:8080/ for home")
	log.Println("Visit http://localhost:8080/health for health check")
	log.Println("Visit http://localhost:8080/live for liveness probe")
	log.Println("Visit http://localhost:8080/ready for readiness probe")
	log.Println("Visit http://localhost:8080/api/status for status")
	log.Println("Press Ctrl+C for graceful shutdown")
	
	// Run server with graceful shutdown
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
