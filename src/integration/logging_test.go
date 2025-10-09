package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"goimgserver/errors"
	"goimgserver/logging"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestLogging_Integration_HTTPHandlers tests logging in HTTP context
func TestLogging_Integration_HTTPHandlers(t *testing.T) {
	// Create logger
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(buf, slog.LevelInfo)
	
	// Create router with logging middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		
		logger.InfoWithFields("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
		)
	})
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify logging
	output := buf.String()
	assert.Contains(t, output, "request")
	assert.Contains(t, output, "/test")
	assert.Contains(t, output, "duration_ms")
}

// TestLogging_Integration_ErrorPropagation tests error flow through system
func TestLogging_Integration_ErrorPropagation(t *testing.T) {
	// Create logger
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(buf, slog.LevelInfo)
	
	// Create router with error handling
	router := gin.New()
	router.Use(errors.ErrorHandlerMiddleware())
	router.GET("/error", func(c *gin.Context) {
		err := errors.NewImageNotFoundError("missing.jpg")
		logger.ErrorWithFields("handler error",
			"error", err.Error(),
			"type", err.Type(),
		)
		errors.HandleError(c, err)
	})
	
	// Make request
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	// Verify error response
	var response errors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", response.Code)
	
	// Verify logging
	output := buf.String()
	assert.Contains(t, output, "handler error")
	assert.Contains(t, output, "missing.jpg")
}

// TestLogging_Integration_PerformanceTracking tests performance in HTTP context
func TestLogging_Integration_PerformanceTracking(t *testing.T) {
	// Create logger and performance logger
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(buf, slog.LevelInfo)
	perfLogger := logging.NewPerformanceLogger(logger)
	
	// Create router
	router := gin.New()
	router.GET("/process", func(c *gin.Context) {
		ctx := c.Request.Context()
		
		// Track operation
		ctx = perfLogger.StartOperation(ctx, "process_image")
		time.Sleep(10 * time.Millisecond)
		perfLogger.EndOperation(ctx, "process_image", map[string]interface{}{
			"image": "test.jpg",
			"size":  "800x600",
		})
		
		c.JSON(http.StatusOK, gin.H{"status": "processed"})
	})
	
	// Make request
	req := httptest.NewRequest("GET", "/process", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify performance logging
	output := buf.String()
	assert.Contains(t, output, "process_image")
	assert.Contains(t, output, "duration_ms")
	assert.Contains(t, output, "test.jpg")
}

// TestIntegration_FullStack tests complete logging and error handling stack
func TestIntegration_FullStack(t *testing.T) {
	// Create logger with config
	buf := &bytes.Buffer{}
	config := logging.ProductionConfig()
	logger := logging.NewLoggerFromConfig(buf, config)
	perfLogger := logging.NewPerformanceLogger(logger)
	
	// Create router with full middleware stack
	router := gin.New()
	
	// Request ID middleware
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-req-123")
		c.Next()
	})
	
	// Error handling middleware
	router.Use(errors.ErrorHandlerMiddleware())
	
	// Logging middleware
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
		
		duration := time.Since(start)
		logger.InfoWithFields("request completed",
			"request_id", c.GetString("request_id"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
		)
	})
	
	// Test endpoint
	router.GET("/api/image/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		
		if filename == "missing.jpg" {
			errors.HandleError(c, errors.NewImageNotFoundError(filename))
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"filename": filename, "status": "ok"})
	})
	
	// Test successful request
	t.Run("Successful request", func(t *testing.T) {
		buf.Reset()
		req := httptest.NewRequest("GET", "/api/image/test.jpg", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		output := buf.String()
		assert.Contains(t, output, "request completed")
		assert.Contains(t, output, "test-req-123")
	})
	
	// Test error request
	t.Run("Error request", func(t *testing.T) {
		buf.Reset()
		req := httptest.NewRequest("GET", "/api/image/missing.jpg", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response errors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "NOT_FOUND", response.Code)
		assert.Equal(t, "test-req-123", response.RequestID)
	})
}

// TestIntegration_ContextPropagation tests context values through the stack
func TestIntegration_ContextPropagation(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(buf, slog.LevelInfo)
	
	router := gin.New()
	
	// Context setup middleware
	router.Use(func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "user_id", "user-123")
		ctx = context.WithValue(ctx, "request_id", "req-456")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	
	router.GET("/test", func(c *gin.Context) {
		ctx := c.Request.Context()
		
		logger.InfoContext(ctx, "processing with context",
			"user_id", ctx.Value("user_id"),
			"request_id", ctx.Value("request_id"),
		)
		
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	output := buf.String()
	assert.Contains(t, output, "user-123")
	assert.Contains(t, output, "req-456")
}
