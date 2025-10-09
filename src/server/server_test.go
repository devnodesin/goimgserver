package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestServer_Integration_FullStack(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &Config{
		Port:            9000,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		ShutdownTimeout: 5 * time.Second,
		EnableCORS:      true,
		EnableRateLimit: false, // Disable for test
	}
	
	srv := New(config)
	
	// Add a test route
	srv.Router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["message"])
	
	// Verify middleware headers are set
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestServer_Integration_ShutdownFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &Config{
		Port:            9001,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 2 * time.Second,
		EnableCORS:      true,
		EnableRateLimit: false,
	}
	
	srv := New(config)
	
	// Add a slow handler to test graceful shutdown
	srv.Router.GET("/slow", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "slow"})
	})
	
	// Start server in background
	go func() {
		srv.Start()
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Shutdown should complete within timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	err := srv.Shutdown(ctx)
	assert.NoError(t, err, "Shutdown should complete without error")
}

func TestServer_Integration_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &Config{
		Port:            9002,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 2 * time.Second,
		EnableCORS:      true,
		EnableRateLimit: false,
	}
	
	srv := New(config)
	
	// Add a route that panics
	srv.Router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	srv.Router.ServeHTTP(w, req)

	// Error should be caught and return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	
	// Request ID should still be set
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestServer_HealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &Config{
		Port:            9003,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 2 * time.Second,
		EnableCORS:      false,
		EnableRateLimit: false,
	}
	
	srv := New(config)

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"Health", "/health", http.StatusOK},
		{"Liveness", "/live", http.StatusOK},
		{"Readiness", "/ready", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			srv.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "status")
		})
	}
}

func TestServer_RateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &Config{
		Port:            9004,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 2 * time.Second,
		EnableCORS:      false,
		EnableRateLimit: true,
		RateLimit:       5,
		RatePer:         time.Second,
	}
	
	srv := New(config)
	
	srv.Router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	successCount := 0
	limitedCount := 0

	// Make requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		srv.Router.ServeHTTP(w, req)
		
		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			limitedCount++
		}
	}

	assert.Equal(t, 5, successCount, "Should allow 5 requests")
	assert.Equal(t, 5, limitedCount, "Should rate limit 5 requests")
}
