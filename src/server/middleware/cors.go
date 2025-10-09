package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that handles CORS with wildcard origin
func CORS() gin.HandlerFunc {
	return CORSWithOrigins([]string{"*"})
}

// CORSWithOrigins returns a middleware that handles CORS with specific allowed origins
func CORSWithOrigins(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Determine if origin is allowed
		var allowOrigin string
		if len(allowedOrigins) > 0 && allowedOrigins[0] == "*" {
			allowOrigin = "*"
		} else if origin != "" {
			for _, allowed := range allowedOrigins {
				if allowed == origin {
					allowOrigin = origin
					break
				}
			}
		}

		// Set CORS headers if origin is allowed
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Access-Control-Allow-Credentials", "false")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Request-ID")
			c.Header("Access-Control-Expose-Headers", "Content-Length, X-Processing-Time, X-Request-ID")
			c.Header("Access-Control-Max-Age", "43200") // 12 hours
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
