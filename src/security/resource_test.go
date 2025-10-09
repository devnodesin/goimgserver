package security

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestResourceProtection_MemoryLimits_ImageProcessing tests memory limits
func TestResourceProtection_MemoryLimits_ImageProcessing(t *testing.T) {
	limiter := NewMemoryLimiter(10 * 1024 * 1024) // 10MB limit

	tests := []struct {
		name          string
		requestedMem  int64
		shouldSucceed bool
	}{
		{
			name:          "within_limit",
			requestedMem:  5 * 1024 * 1024, // 5MB
			shouldSucceed: true,
		},
		{
			name:          "at_limit",
			requestedMem:  10 * 1024 * 1024, // 10MB
			shouldSucceed: true,
		},
		{
			name:          "exceeds_limit",
			requestedMem:  15 * 1024 * 1024, // 15MB
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := limiter.Reserve(tt.requestedMem)
			if tt.shouldSucceed {
				assert.NoError(t, err)
				limiter.Release(tt.requestedMem)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestResourceProtection_ProcessingTimeouts_Operations tests operation timeouts
func TestResourceProtection_ProcessingTimeouts_Operations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add timeout middleware (100ms timeout)
	router.Use(TimeoutMiddleware(100 * time.Millisecond))

	router.GET("/fast", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "done"})
	})

	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "done"})
	})

	// Fast request should succeed
	req1 := httptest.NewRequest("GET", "/fast", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code, "Fast request should succeed")

	// Slow request should timeout
	req2 := httptest.NewRequest("GET", "/slow", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusRequestTimeout, w2.Code, "Slow request should timeout")
}

// TestResourceProtection_ConcurrentLimits_Requests tests concurrent request limits
func TestResourceProtection_ConcurrentLimits_Requests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Allow max 3 concurrent requests
	router.Use(ConcurrencyLimiter(3))

	var processing atomic.Int32
	var maxConcurrent atomic.Int32

	router.GET("/test", func(c *gin.Context) {
		current := processing.Add(1)
		
		// Track max concurrent requests
		for {
			max := maxConcurrent.Load()
			if current <= max || maxConcurrent.CompareAndSwap(max, current) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond)
		processing.Add(-1)
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Launch 10 concurrent requests
	var wg sync.WaitGroup
	results := make([]int, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results[index] = w.Code
		}(i)
	}

	wg.Wait()

	// Some requests should succeed, some should be rate limited
	successCount := 0
	limitedCount := 0
	for _, code := range results {
		if code == http.StatusOK {
			successCount++
		} else if code == http.StatusTooManyRequests {
			limitedCount++
		}
	}

	// Should have limited some requests
	assert.Greater(t, limitedCount, 0, "Should have limited some requests")
	// Max concurrent should not exceed limit
	assert.LessOrEqual(t, maxConcurrent.Load(), int32(3), "Should not exceed concurrent limit")
}

// TestResourceProtection_DiskSpace_Monitoring tests disk space monitoring
func TestResourceProtection_DiskSpace_Monitoring(t *testing.T) {
	monitor := NewDiskSpaceMonitor("/tmp", 100*1024*1024) // 100MB threshold

	// Check current usage
	usage, err := monitor.GetUsage()
	assert.NoError(t, err)
	assert.NotNil(t, usage)

	// Usage should have valid values
	assert.GreaterOrEqual(t, usage.Total, int64(0))
	assert.GreaterOrEqual(t, usage.Used, int64(0))
	assert.GreaterOrEqual(t, usage.Available, int64(0))
	assert.GreaterOrEqual(t, usage.PercentUsed, float64(0))
	assert.LessOrEqual(t, usage.PercentUsed, float64(100))
}

// TestResourceProtection_DiskSpace_ThresholdAlert tests threshold alerts
func TestResourceProtection_DiskSpace_ThresholdAlert(t *testing.T) {
	// Create monitor with low threshold for testing
	monitor := NewDiskSpaceMonitor("/tmp", 1) // 1 byte threshold (will always trigger)

	usage, err := monitor.GetUsage()
	assert.NoError(t, err)

	// Check if threshold is exceeded
	exceeded := monitor.IsThresholdExceeded()
	
	// If we're using more than 1 byte (which we always are), threshold should be exceeded
	if usage.Used > 1 {
		assert.True(t, exceeded, "Should exceed threshold with 1 byte limit")
	}
}

// TestResourceProtection_MemoryLimits_Concurrent tests concurrent memory reservations
func TestResourceProtection_MemoryLimits_Concurrent(t *testing.T) {
	limiter := NewMemoryLimiter(100 * 1024 * 1024) // 100MB limit

	var wg sync.WaitGroup
	successes := atomic.Int32{}
	failures := atomic.Int32{}

	// Try to reserve 20x 10MB (total 200MB, but limit is 100MB)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := limiter.Reserve(10 * 1024 * 1024)
			if err == nil {
				successes.Add(1)
				time.Sleep(10 * time.Millisecond)
				limiter.Release(10 * 1024 * 1024)
			} else {
				failures.Add(1)
			}
		}()
	}

	wg.Wait()

	// Should have some failures due to limit
	assert.Greater(t, failures.Load(), int32(0), "Should have some failures")
	// Should have some successes
	assert.Greater(t, successes.Load(), int32(0), "Should have some successes")
	// Total should be 20
	assert.Equal(t, int32(20), successes.Load()+failures.Load())
}

// TestResourceProtection_RequestSizeLimit tests request size limiting
func TestResourceProtection_RequestSizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Limit requests to 1KB
	router.Use(RequestSizeLimiter(1024))

	router.POST("/upload", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	tests := []struct {
		name           string
		bodySize       int
		expectedStatus int
	}{
		{
			name:           "small_request",
			bodySize:       500,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "at_limit",
			bodySize:       1024,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "exceeds_limit",
			bodySize:       2048,
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/upload", nil)
			req.Header.Set("Content-Length", string(rune(tt.bodySize)))
			req.ContentLength = int64(tt.bodySize)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestResourceProtection_CircuitBreaker tests circuit breaker pattern
func TestResourceProtection_CircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond) // 3 failures, 100ms timeout

	// Simulate successful calls
	for i := 0; i < 5; i++ {
		err := cb.Call(func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, CircuitClosed, cb.State())
	}

	// Simulate failures
	for i := 0; i < 3; i++ {
		err := cb.Call(func() error {
			return assert.AnError
		})
		assert.Error(t, err)
	}

	// Circuit should be open now
	assert.Equal(t, CircuitOpen, cb.State())

	// Further calls should fail immediately
	err := cb.Call(func() error {
		t.Fatal("Should not execute when circuit is open")
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Circuit is still open, but will transition to half-open on next call
	assert.Equal(t, CircuitOpen, cb.State())

	// Next call will transition to half-open and execute
	err = cb.Call(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, CircuitClosed, cb.State())
}

// TestResourceProtection_ProcessingTimeout_Context tests context-based timeouts
func TestResourceProtection_ProcessingTimeout_Context(t *testing.T) {
	tests := []struct {
		name          string
		processingTime time.Duration
		timeout       time.Duration
		shouldTimeout bool
	}{
		{
			name:           "completes_in_time",
			processingTime: 50 * time.Millisecond,
			timeout:        100 * time.Millisecond,
			shouldTimeout:  false,
		},
		{
			name:           "exceeds_timeout",
			processingTime: 150 * time.Millisecond,
			timeout:        100 * time.Millisecond,
			shouldTimeout:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ProcessWithTimeout(func() error {
				time.Sleep(tt.processingTime)
				return nil
			}, tt.timeout)

			if tt.shouldTimeout {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
