package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_Order_Correct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	var order []string
	
	// Add middleware that tracks execution order
	router.Use(func(c *gin.Context) {
		order = append(order, "first-before")
		c.Next()
		order = append(order, "first-after")
	})
	
	router.Use(func(c *gin.Context) {
		order = append(order, "second-before")
		c.Next()
		order = append(order, "second-after")
	})
	
	router.GET("/test", func(c *gin.Context) {
		order = append(order, "handler")
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check execution order
	expected := []string{
		"first-before",
		"second-before",
		"handler",
		"second-after",
		"first-after",
	}
	assert.Equal(t, expected, order)
}

func TestMiddleware_Chain_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Set up complete middleware chain
	router.Use(RequestID())
	router.Use(SecurityHeaders())
	router.Use(CORS())
	router.Use(ErrorHandler())
	router.Use(Logging())
	
	router.GET("/test", func(c *gin.Context) {
		// Verify request ID was set
		requestID := c.GetString("request_id")
		assert.NotEmpty(t, requestID)
		
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify headers from different middleware
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"), "Request ID header should be set")
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"), "Security headers should be set")
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"), "CORS headers should be set")
}

func TestMiddleware_Performance_Overhead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Router without middleware
	routerBasic := gin.New()
	routerBasic.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	
	// Router with full middleware stack
	routerFull := gin.New()
	routerFull.Use(RequestID())
	routerFull.Use(SecurityHeaders())
	routerFull.Use(CORS())
	routerFull.Use(ErrorHandler())
	routerFull.Use(Logging())
	routerFull.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	
	// Measure basic request
	startBasic := time.Now()
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		routerBasic.ServeHTTP(w, req)
	}
	basicDuration := time.Since(startBasic)
	
	// Measure with middleware
	startFull := time.Now()
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		routerFull.ServeHTTP(w, req)
	}
	fullDuration := time.Since(startFull)
	
	// Middleware overhead should be reasonable (less than 10x slower)
	overhead := float64(fullDuration) / float64(basicDuration)
	assert.Less(t, overhead, 10.0, "Middleware overhead should be reasonable")
	
	t.Logf("Basic: %v, With middleware: %v, Overhead: %.2fx", basicDuration, fullDuration, overhead)
}

func TestMiddleware_Concurrent_Safety(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add middleware
	router.Use(RequestID())
	router.Use(SecurityHeaders())
	router.Use(CORS())
	router.Use(ErrorHandler())
	router.Use(RateLimit(1000, time.Second)) // High limit for concurrent test
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Run concurrent requests
	done := make(chan bool)
	errors := make(chan error, 50)
	
	for i := 0; i < 50; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				errors <- assert.AnError
			}
			done <- true
		}()
	}
	
	// Wait for all requests to complete
	for i := 0; i < 50; i++ {
		<-done
	}
	close(errors)
	
	// Check no errors occurred
	errorCount := 0
	for range errors {
		errorCount++
	}
	assert.Equal(t, 0, errorCount, "No errors should occur in concurrent requests")
}

func TestMiddleware_ErrorPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(RequestID())
	router.Use(ErrorHandler())
	router.Use(Logging())
	
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Error middleware should catch panic
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	// Request ID should still be in response
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestMiddleware_AbortPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	handlerExecuted := false
	
	router.Use(func(c *gin.Context) {
		// Abort early
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	})
	
	router.GET("/test", func(c *gin.Context) {
		handlerExecuted = true
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, handlerExecuted, "Handler should not execute after abort")
}
