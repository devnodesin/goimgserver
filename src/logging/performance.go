package logging

import (
	"context"
	"sync"
	"time"
)

// PerformanceLogger wraps a logger with performance tracking capabilities
type PerformanceLogger struct {
	logger  *Logger
	metrics *Metrics
	mu      sync.RWMutex
}

// Metrics represents performance metrics
type Metrics struct {
	TotalOperations int64
	TotalDuration   time.Duration
	ErrorCount      int64
}

// NewPerformanceLogger creates a new performance logger
func NewPerformanceLogger(logger *Logger) *PerformanceLogger {
	return &PerformanceLogger{
		logger:  logger,
		metrics: &Metrics{},
	}
}

// operationKey is the context key for operation tracking
type operationKey struct{}

// operationContext holds operation timing information
type operationContext struct {
	name      string
	startTime time.Time
}

// StartOperation begins tracking an operation
func (p *PerformanceLogger) StartOperation(ctx context.Context, name string) context.Context {
	opCtx := &operationContext{
		name:      name,
		startTime: time.Now(),
	}
	return context.WithValue(ctx, operationKey{}, opCtx)
}

// EndOperation completes tracking an operation
func (p *PerformanceLogger) EndOperation(ctx context.Context, name string, details map[string]interface{}) {
	opCtx, ok := ctx.Value(operationKey{}).(*operationContext)
	if !ok || opCtx == nil {
		return
	}
	
	duration := time.Since(opCtx.startTime)
	
	// Update metrics
	p.mu.Lock()
	p.metrics.TotalOperations++
	p.metrics.TotalDuration += duration
	p.mu.Unlock()
	
	// Log the operation
	args := []interface{}{
		"operation", name,
		"duration_ms", duration.Milliseconds(),
	}
	
	// Add details if provided
	if details != nil {
		for k, v := range details {
			args = append(args, k, v)
		}
	}
	
	p.logger.InfoWithFields("operation completed", args...)
}

// EndOperationWithError completes tracking an operation that errored
func (p *PerformanceLogger) EndOperationWithError(ctx context.Context, name string, err error) {
	opCtx, ok := ctx.Value(operationKey{}).(*operationContext)
	if !ok || opCtx == nil {
		return
	}
	
	duration := time.Since(opCtx.startTime)
	
	// Update metrics
	p.mu.Lock()
	p.metrics.TotalOperations++
	p.metrics.TotalDuration += duration
	p.metrics.ErrorCount++
	p.mu.Unlock()
	
	// Log the error
	p.logger.ErrorWithFields("operation failed",
		"operation", name,
		"duration_ms", duration.Milliseconds(),
		"error", err.Error(),
	)
}

// GetMetrics returns current metrics
func (p *PerformanceLogger) GetMetrics() Metrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return Metrics{
		TotalOperations: p.metrics.TotalOperations,
		TotalDuration:   p.metrics.TotalDuration,
		ErrorCount:      p.metrics.ErrorCount,
	}
}

// ResetMetrics resets all metrics to zero
func (p *PerformanceLogger) ResetMetrics() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.metrics = &Metrics{}
}

// LogMetrics logs current metrics
func (p *PerformanceLogger) LogMetrics() {
	metrics := p.GetMetrics()
	
	avgDuration := time.Duration(0)
	if metrics.TotalOperations > 0 {
		avgDuration = metrics.TotalDuration / time.Duration(metrics.TotalOperations)
	}
	
	p.logger.InfoWithFields("performance metrics",
		"total_operations", metrics.TotalOperations,
		"total_duration_ms", metrics.TotalDuration.Milliseconds(),
		"avg_duration_ms", avgDuration.Milliseconds(),
		"error_count", metrics.ErrorCount,
	)
}
