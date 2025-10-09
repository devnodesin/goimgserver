package performance

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"goimgserver/cache"
	"goimgserver/resolver"
	"goimgserver/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestPerformance_LoadTesting_ConcurrentUsers tests concurrent load
func TestPerformance_LoadTesting_ConcurrentUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	concurrencyLevels := []int{10, 50, 100}
	
	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("%d_concurrent_users", concurrency), func(t *testing.T) {
			tmpDir := t.TempDir()
			imagesDir := filepath.Join(tmpDir, "images")
			cacheDir := filepath.Join(tmpDir, "cache")
			
			// Setup
			fixtureManager := testutils.NewFixtureManager(imagesDir)
			if err := fixtureManager.CreateFixtureSet(); err != nil {
				t.Fatalf("Failed to create fixtures: %v", err)
			}
			
			cm, err := cache.NewManager(cacheDir)
			if err != nil {
				t.Fatalf("Failed to create cache manager: %v", err)
			}
			
			res := resolver.NewResolver(imagesDir)
			
			// Create router
			router := gin.New()
			router.GET("/img/:filename/:dimensions", func(c *gin.Context) {
				filename := c.Param("filename")
				
				result, err := res.Resolve(filename)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
					return
				}
				
				params := cache.ProcessingParams{Width: 800, Height: 600}
				_, found, _ := cm.Retrieve(result.ResolvedPath, params)
				
				if !found {
					data := []byte("processed data")
					_ = cm.Store(result.ResolvedPath, params, data)
				}
				
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
			
			// Load test
			var wg sync.WaitGroup
			var successCount int64
			var errorCount int64
			
			startTime := time.Now()
			
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					
					for j := 0; j < 100; j++ {
						req := testutils.NewRequestBuilder("GET", "/img/test.jpg/800x600").Build()
						rec := testutils.NewResponseRecorder()
						
						router.ServeHTTP(rec, req)
						
						if rec.Code == http.StatusOK {
							atomic.AddInt64(&successCount, 1)
						} else {
							atomic.AddInt64(&errorCount, 1)
						}
					}
				}(i)
			}
			
			wg.Wait()
			duration := time.Since(startTime)
			
			totalRequests := successCount + errorCount
			requestsPerSecond := float64(totalRequests) / duration.Seconds()
			
			t.Logf("Concurrency: %d", concurrency)
			t.Logf("Total requests: %d", totalRequests)
			t.Logf("Successful: %d", successCount)
			t.Logf("Errors: %d", errorCount)
			t.Logf("Duration: %v", duration)
			t.Logf("Requests/sec: %.2f", requestsPerSecond)
			
			// Assertions
			assert.Greater(t, successCount, int64(0))
			assert.LessOrEqual(t, errorCount, successCount/10) // Max 10% error rate
		})
	}
}

// TestPerformance_StressTesting_ResourceLimits tests resource limits
func TestPerformance_StressTesting_ResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}
	
	t.Run("Large number of files", func(t *testing.T) {
		numFiles := 1000
		params := cache.ProcessingParams{Width: 800, Height: 600}
		data := make([]byte, 10*1024) // 10KB per file
		
		startTime := time.Now()
		
		// Store many files
		for i := 0; i < numFiles; i++ {
			path := fmt.Sprintf("/images/stress%d.jpg", i)
			err := cm.Store(path, params, data)
			assert.NoError(t, err)
		}
		
		storeDuration := time.Since(startTime)
		
		// Retrieve all files
		startTime = time.Now()
		
		for i := 0; i < numFiles; i++ {
			path := fmt.Sprintf("/images/stress%d.jpg", i)
			_, found, err := cm.Retrieve(path, params)
			assert.NoError(t, err)
			assert.True(t, found)
		}
		
		retrieveDuration := time.Since(startTime)
		
		t.Logf("Stored %d files in %v", numFiles, storeDuration)
		t.Logf("Retrieved %d files in %v", numFiles, retrieveDuration)
		t.Logf("Avg store time: %v", storeDuration/time.Duration(numFiles))
		t.Logf("Avg retrieve time: %v", retrieveDuration/time.Duration(numFiles))
	})
	
	t.Run("Large file sizes", func(t *testing.T) {
		fileSizes := []int{
			1 * 1024 * 1024,  // 1MB
			5 * 1024 * 1024,  // 5MB
			10 * 1024 * 1024, // 10MB
		}
		
		params := cache.ProcessingParams{Width: 2000, Height: 1500}
		
		for _, size := range fileSizes {
			data := make([]byte, size)
			path := fmt.Sprintf("/images/large_%dmb.jpg", size/(1024*1024))
			
			startTime := time.Now()
			err := cm.Store(path, params, data)
			storeDuration := time.Since(startTime)
			
			assert.NoError(t, err)
			
			startTime = time.Now()
			retrieved, found, err := cm.Retrieve(path, params)
			retrieveDuration := time.Since(startTime)
			
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, len(data), len(retrieved))
			
			t.Logf("Size: %dMB, Store: %v, Retrieve: %v", 
				size/(1024*1024), storeDuration, retrieveDuration)
		}
	})
}

// TestPerformance_SustainedLoad tests sustained load over time
func TestPerformance_SustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping sustained load test in short mode")
	}
	
	tmpDir := t.TempDir()
	imagesDir := filepath.Join(tmpDir, "images")
	cacheDir := filepath.Join(tmpDir, "cache")
	
	fixtureManager := testutils.NewFixtureManager(imagesDir)
	if err := fixtureManager.CreateFixtureSet(); err != nil {
		t.Fatalf("Failed to create fixtures: %v", err)
	}
	
	cm, err := cache.NewManager(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}
	
	res := resolver.NewResolver(imagesDir)
	
	// Run for 10 seconds with constant load
	duration := 10 * time.Second
	concurrency := 20
	
	var requestCount int64
	var errorCount int64
	stopChan := make(chan struct{})
	
	var wg sync.WaitGroup
	
	startTime := time.Now()
	
	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			params := cache.ProcessingParams{Width: 800, Height: 600}
			
			for {
				select {
				case <-stopChan:
					return
				default:
					result, err := res.Resolve("test.jpg")
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
						continue
					}
					
					_, found, _ := cm.Retrieve(result.ResolvedPath, params)
					if !found {
						data := []byte("processed")
						_ = cm.Store(result.ResolvedPath, params, data)
					}
					
					atomic.AddInt64(&requestCount, 1)
				}
			}
		}()
	}
	
	// Let it run for duration
	time.Sleep(duration)
	close(stopChan)
	
	wg.Wait()
	
	actualDuration := time.Since(startTime)
	requestsPerSecond := float64(requestCount) / actualDuration.Seconds()
	
	t.Logf("Duration: %v", actualDuration)
	t.Logf("Total requests: %d", requestCount)
	t.Logf("Errors: %d", errorCount)
	t.Logf("Requests/sec: %.2f", requestsPerSecond)
	t.Logf("Concurrency: %d", concurrency)
	
	assert.Greater(t, requestCount, int64(0))
	assert.LessOrEqual(t, errorCount, requestCount/20) // Max 5% error rate
}

// TestPerformance_MemoryPressure tests behavior under memory pressure
func TestPerformance_MemoryPressure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory pressure test in short mode")
	}
	
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	
	cm, err := cache.NewManager(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}
	
	// Create large dataset
	numFiles := 100
	fileSize := 1 * 1024 * 1024 // 1MB each
	params := cache.ProcessingParams{Width: 2000, Height: 1500}
	
	for i := 0; i < numFiles; i++ {
		data := make([]byte, fileSize)
		path := fmt.Sprintf("/images/memory%d.jpg", i)
		err := cm.Store(path, params, data)
		assert.NoError(t, err)
	}
	
	// Concurrent access under memory pressure
	var wg sync.WaitGroup
	concurrency := 50
	
	startTime := time.Now()
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < 100; j++ {
				fileIdx := (id*100 + j) % numFiles
				path := fmt.Sprintf("/images/memory%d.jpg", fileIdx)
				_, found, err := cm.Retrieve(path, params)
				assert.NoError(t, err)
				assert.True(t, found)
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(startTime)
	
	totalOps := concurrency * 100
	opsPerSecond := float64(totalOps) / duration.Seconds()
	
	t.Logf("Operations: %d", totalOps)
	t.Logf("Duration: %v", duration)
	t.Logf("Ops/sec: %.2f", opsPerSecond)
}
