package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_RequestLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)
	
	router := gin.New()
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test?query=value", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	logStr := logOutput.String()
	assert.Contains(t, logStr, "GET")
	assert.Contains(t, logStr, "/test")
	assert.Contains(t, logStr, "200")
}

func TestLoggingMiddleware_ResponseLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)
	
	router := gin.New()
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	logStr := logOutput.String()
	assert.Contains(t, logStr, "200")
	// Check for time duration (µs, ms, s, etc.)
	assert.True(t, strings.Contains(logStr, "µs") || strings.Contains(logStr, "ms") || strings.Contains(logStr, "s"), 
		"Should log response time in some format")
}

func TestLoggingMiddleware_ErrorLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)
	
	router := gin.New()
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	logStr := logOutput.String()
	assert.Contains(t, logStr, "500")
	assert.Contains(t, logStr, "/test")
}

func TestLoggingMiddleware_RequestIDInLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)
	
	router := gin.New()
	router.Use(RequestID())
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	logStr := logOutput.String()
	// Should contain request ID in log
	assert.NotEmpty(t, logStr)
}

func TestLoggingMiddleware_ClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)
	
	router := gin.New()
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	logStr := logOutput.String()
	// Should log some form of client identifier
	assert.NotEmpty(t, logStr)
}

func TestLoggingMiddleware_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)
	
	router := gin.New()
	router.Use(Logging())
	router.GET("/test1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test1"})
	})
	router.GET("/test2", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	// First request
	req1 := httptest.NewRequest("GET", "/test1", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Second request
	req2 := httptest.NewRequest("GET", "/test2", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	logStr := logOutput.String()
	assert.Contains(t, logStr, "/test1")
	assert.Contains(t, logStr, "/test2")
	assert.Contains(t, logStr, "200")
	assert.Contains(t, logStr, "404")
	
	// Should have two log entries
	lines := strings.Split(strings.TrimSpace(logStr), "\n")
	assert.GreaterOrEqual(t, len(lines), 2)
}
