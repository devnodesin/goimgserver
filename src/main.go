package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	// Set trusted proxies to localhost only (removes Gin warning)
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
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
	println("Server started and running.")
	println("Server will listen on 127.0.0.1:9000 (localhost:9000 on Windows)")
	println("GET http://127.0.0.1:9000/ping to test; you should see message pong.")

	// Start server on port 9000
	r.Run(":9000")
}
