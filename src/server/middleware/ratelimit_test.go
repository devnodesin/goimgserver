package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiting_BasicLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Set low limit for testing (5 requests per second)
	router.Use(RateLimit(5, time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	successCount := 0
	limitedCount := 0

	// Make 10 requests quickly
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			limitedCount++
		}
	}

	assert.Equal(t, 5, successCount, "Should allow 5 requests")
	assert.Equal(t, 5, limitedCount, "Should rate limit 5 requests")
}

func TestRateLimiting_PerIPLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Set per-IP rate limit
	router.Use(RateLimitPerIP(3, time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Test from first IP
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if i < 3 {
			assert.Equal(t, http.StatusOK, w.Code, "First 3 requests from IP1 should succeed")
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code, "Requests 4+ from IP1 should be limited")
		}
	}

	// Test from second IP (should have separate limit)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code, "Requests from IP2 should succeed (separate limit)")
	}
}

func TestRateLimiting_BurstHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Set rate limit with burst capacity
	router.Use(RateLimitWithBurst(2, time.Second, 5))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// First burst of 5 should succeed (burst capacity)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code, "Burst requests should succeed")
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request exceeding burst should be limited")
}

func TestRateLimiting_RefillOverTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Set rate limit (2 requests per 100ms)
	router.Use(RateLimit(2, 100*time.Millisecond))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 3rd request should be limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should work again after refill
	req4 := httptest.NewRequest("GET", "/test", nil)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code, "Should work after rate limit window")
}

func TestRateLimiting_ErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(RateLimit(1, time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// First request succeeds
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request is rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Contains(t, w2.Body.String(), "error")
	// Should have retry-after or similar information
	assert.NotEmpty(t, w2.Body.String())
}
