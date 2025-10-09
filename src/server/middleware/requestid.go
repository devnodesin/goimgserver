package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestID returns a middleware that generates or propagates request IDs
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get existing request ID from header
		requestID := c.GetHeader("X-Request-ID")
		
		// Generate new request ID if not provided
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		// Store in context for handlers to use
		c.Set("request_id", requestID)
		
		// Set in response header
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to simple ID if random fails
		return "req-fallback"
	}
	return hex.EncodeToString(b)
}
