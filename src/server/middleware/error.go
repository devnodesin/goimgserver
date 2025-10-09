package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a middleware that handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with request ID if available
				requestID := c.GetString("request_id")
				if requestID != "" {
					log.Printf("[%s] PANIC: %v", requestID, err)
				} else {
					log.Printf("PANIC: %v", err)
				}
				
				// Build error response
				response := gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_ERROR",
				}
				
				// Include request ID in response for tracing
				if requestID != "" {
					response["request_id"] = requestID
				}
				
				c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			}
		}()
		
		c.Next()
	}
}
