package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging returns a middleware that logs HTTP requests
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		
		// Process request
		c.Next()
		
		// Calculate request duration
		duration := time.Since(start)
		
		// Get request ID if available
		requestID := c.GetString("request_id")
		
		// Build log message
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		
		// Log format: [RequestID] ClientIP Method Path?Query Status Duration
		logMsg := ""
		if requestID != "" {
			logMsg += "[" + requestID + "] "
		}
		logMsg += clientIP + " " + method + " " + path
		if query != "" {
			logMsg += "?" + query
		}
		logMsg += " " + statusCodeString(statusCode) + " " + duration.String()
		
		log.Println(logMsg)
	}
}

func statusCodeString(code int) string {
	if code >= 500 {
		return "ERROR:" + intToString(code)
	} else if code >= 400 {
		return "WARN:" + intToString(code)
	}
	return intToString(code)
}

func intToString(i int) string {
	switch i {
	case 200:
		return "200"
	case 201:
		return "201"
	case 204:
		return "204"
	case 400:
		return "400"
	case 401:
		return "401"
	case 403:
		return "403"
	case 404:
		return "404"
	case 500:
		return "500"
	case 502:
		return "502"
	case 503:
		return "503"
	default:
		// Fallback for other codes
		return string(rune(i/100 + '0')) + string(rune((i/10)%10 + '0')) + string(rune(i%10 + '0'))
	}
}
