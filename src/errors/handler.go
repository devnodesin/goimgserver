package errors

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// HandleError handles an error and responds with appropriate HTTP status and JSON
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	
	// Extract request ID from context if available
	requestID, _ := c.Get("request_id")
	reqIDStr, _ := requestID.(string)
	
	// Try to cast to AppError
	var appErr *AppError
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		// Wrap standard error as internal error
		appErr = NewAppError(err.Error(), ErrorTypeInternal, err)
	}
	
	// Add request ID if not already present
	if reqIDStr != "" && appErr.requestID == "" {
		appErr = appErr.WithRequestID(reqIDStr)
	}
	
	// Convert to response
	response := appErr.ToResponse()
	
	// Send JSON response
	c.AbortWithStatusJSON(response.Status, response)
}

// ErrorHandlerMiddleware returns a middleware that handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID if available
				requestID, _ := c.Get("request_id")
				reqIDStr, _ := requestID.(string)
				
				// Create error from panic
				var appErr *AppError
				switch e := err.(type) {
				case error:
					appErr = NewAppError(e.Error(), ErrorTypeInternal, e)
				case string:
					appErr = NewAppError(e, ErrorTypeInternal, nil)
				default:
					appErr = NewAppError(fmt.Sprintf("panic: %v", err), ErrorTypeInternal, nil)
				}
				
				// Add request ID
				if reqIDStr != "" {
					appErr = appErr.WithRequestID(reqIDStr)
				}
				
				// Send error response
				response := appErr.ToResponse()
				c.AbortWithStatusJSON(response.Status, response)
			}
		}()
		
		c.Next()
		
		// Check for errors in context
		if len(c.Errors) > 0 {
			// Handle the last error
			lastErr := c.Errors.Last()
			if lastErr != nil && lastErr.Err != nil {
				HandleError(c, lastErr.Err)
			}
		}
	}
}

// AbortWithError aborts the request with an error
func AbortWithError(c *gin.Context, err error) {
	HandleError(c, err)
	c.Abort()
}

// AbortWithAppError aborts the request with an AppError
func AbortWithAppError(c *gin.Context, message string, errorType ErrorType) {
	err := NewAppError(message, errorType, nil)
	HandleError(c, err)
	c.Abort()
}
