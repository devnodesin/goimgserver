package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPerformanceLogging_RequestDuration tests request timing
func TestPerformanceLogging_RequestDuration(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	// Simulate request processing
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	duration := time.Since(start)
	
	logger.InfoWithFields("request completed",
		"duration_ms", duration.Milliseconds(),
		"method", "GET",
		"path", "/img/test.jpg",
	)
	
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Contains(t, logEntry, "duration_ms")
	assert.GreaterOrEqual(t, logEntry["duration_ms"].(float64), 10.0)
}

// TestPerformanceLogging_MemoryUsage tests memory usage logging
func TestPerformanceLogging_MemoryUsage(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	logger.InfoWithFields("memory stats",
		"alloc_mb", float64(m.Alloc)/1024/1024,
		"total_alloc_mb", float64(m.TotalAlloc)/1024/1024,
		"sys_mb", float64(m.Sys)/1024/1024,
		"num_gc", m.NumGC,
	)
	
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Contains(t, logEntry, "alloc_mb")
	assert.Contains(t, logEntry, "num_gc")
}

// TestPerformanceLogging_CacheHitMiss tests cache metrics
func TestPerformanceLogging_CacheHitMiss(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	// Simulate cache metrics
	metrics := map[string]interface{}{
		"cache_hits":   100,
		"cache_misses": 20,
		"hit_rate":     0.833,
	}
	
	logger.InfoWithFields("cache metrics",
		"cache_hits", metrics["cache_hits"],
		"cache_misses", metrics["cache_misses"],
		"hit_rate", metrics["hit_rate"],
	)
	
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, float64(100), logEntry["cache_hits"])
	assert.Equal(t, float64(20), logEntry["cache_misses"])
}

// TestPerformanceLogging_ConcurrentRequests tests concurrent metrics
func TestPerformanceLogging_ConcurrentRequests(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	// Simulate concurrent request tracking
	activeRequests := 5
	totalRequests := 1000
	avgDuration := 45.3
	
	logger.InfoWithFields("concurrent stats",
		"active_requests", activeRequests,
		"total_requests", totalRequests,
		"avg_duration_ms", avgDuration,
	)
	
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, float64(5), logEntry["active_requests"])
	assert.Equal(t, float64(1000), logEntry["total_requests"])
}

// TestPerformanceLogging_ImageProcessing_Duration tests performance logging
func TestPerformanceLogging_ImageProcessing_Duration(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	// Create performance logger wrapper
	perfLogger := NewPerformanceLogger(logger)
	
	// Simulate image processing
	ctx := context.Background()
	ctx = perfLogger.StartOperation(ctx, "resize_image")
	
	time.Sleep(15 * time.Millisecond)
	
	perfLogger.EndOperation(ctx, "resize_image", map[string]interface{}{
		"width":  800,
		"height": 600,
		"format": "jpeg",
	})
	
	output := buf.String()
	assert.Contains(t, output, "resize_image")
	assert.Contains(t, output, "duration_ms")
}

// TestPerformanceLogging_CacheOperations_Metrics tests cache performance logs
func TestPerformanceLogging_CacheOperations_Metrics(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	perfLogger := NewPerformanceLogger(logger)
	
	// Simulate cache operation
	ctx := context.Background()
	ctx = perfLogger.StartOperation(ctx, "cache_read")
	
	time.Sleep(5 * time.Millisecond)
	
	perfLogger.EndOperation(ctx, "cache_read", map[string]interface{}{
		"cache_key": "test.jpg_800x600",
		"hit":       true,
	})
	
	output := buf.String()
	assert.Contains(t, output, "cache_read")
	assert.Contains(t, output, "hit")
}

// TestPerformanceLogger_Nested tests nested operation tracking
func TestPerformanceLogger_Nested(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	perfLogger := NewPerformanceLogger(logger)
	
	ctx := context.Background()
	
	// Outer operation
	ctx = perfLogger.StartOperation(ctx, "process_request")
	time.Sleep(5 * time.Millisecond)
	
	// Inner operation 1
	ctx1 := perfLogger.StartOperation(ctx, "load_image")
	time.Sleep(3 * time.Millisecond)
	perfLogger.EndOperation(ctx1, "load_image", nil)
	
	// Inner operation 2
	ctx2 := perfLogger.StartOperation(ctx, "resize_image")
	time.Sleep(3 * time.Millisecond)
	perfLogger.EndOperation(ctx2, "resize_image", nil)
	
	perfLogger.EndOperation(ctx, "process_request", nil)
	
	// Should have logged all operations
	output := buf.String()
	assert.Contains(t, output, "load_image")
	assert.Contains(t, output, "resize_image")
	assert.Contains(t, output, "process_request")
}

// TestPerformanceLogger_ErrorTracking tests tracking operations that error
func TestPerformanceLogger_ErrorTracking(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	perfLogger := NewPerformanceLogger(logger)
	
	ctx := context.Background()
	ctx = perfLogger.StartOperation(ctx, "failing_operation")
	
	time.Sleep(5 * time.Millisecond)
	
	perfLogger.EndOperationWithError(ctx, "failing_operation", assert.AnError)
	
	output := buf.String()
	assert.Contains(t, output, "failing_operation")
	assert.Contains(t, output, "error")
}

// TestPerformanceLogger_Metrics tests metrics collection
func TestPerformanceLogger_Metrics(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	perfLogger := NewPerformanceLogger(logger)
	
	// Collect metrics
	metrics := perfLogger.GetMetrics()
	assert.NotNil(t, metrics)
	
	// Initially should be empty or have zero values
	assert.GreaterOrEqual(t, metrics.TotalOperations, int64(0))
}

// TestPerformanceLogger_Reset tests metrics reset
func TestPerformanceLogger_Reset(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	perfLogger := NewPerformanceLogger(logger)
	
	// Perform some operations
	ctx := perfLogger.StartOperation(context.Background(), "test_op")
	perfLogger.EndOperation(ctx, "test_op", nil)
	
	// Reset metrics
	perfLogger.ResetMetrics()
	
	metrics := perfLogger.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalOperations)
}
