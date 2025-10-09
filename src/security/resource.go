package security

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrMemoryLimitExceeded     = errors.New("memory limit exceeded")
	ErrTimeoutExceeded         = errors.New("timeout exceeded")
	ErrConcurrencyLimitReached = errors.New("concurrency limit reached")
	ErrCircuitBreakerOpen      = errors.New("circuit breaker is open")
)

// MemoryLimiter limits memory usage
type MemoryLimiter struct {
	maxMemory int64
	current   atomic.Int64
	mu        sync.Mutex
}

// NewMemoryLimiter creates a new memory limiter
func NewMemoryLimiter(maxMemory int64) *MemoryLimiter {
	return &MemoryLimiter{
		maxMemory: maxMemory,
	}
}

// Reserve attempts to reserve memory
func (m *MemoryLimiter) Reserve(amount int64) error {
	for {
		current := m.current.Load()
		newValue := current + amount
		if newValue > m.maxMemory {
			return ErrMemoryLimitExceeded
		}
		if m.current.CompareAndSwap(current, newValue) {
			return nil
		}
	}
}

// Release releases reserved memory
func (m *MemoryLimiter) Release(amount int64) {
	m.current.Add(-amount)
}

// TimeoutMiddleware creates middleware that enforces request timeouts
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			c.Next()
			close(finished)
		}()

		select {
		case <-finished:
			// Request completed in time
		case <-ctx.Done():
			// Timeout occurred
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timeout",
				"code":  "TIMEOUT",
			})
		}
	}
}

// ConcurrencyLimiter creates middleware that limits concurrent requests
func ConcurrencyLimiter(maxConcurrent int) gin.HandlerFunc {
	sem := make(chan struct{}, maxConcurrent)

	return func(c *gin.Context) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many concurrent requests",
				"code":  "CONCURRENCY_LIMIT_EXCEEDED",
			})
		}
	}
}

// DiskSpaceMonitor monitors disk space usage
type DiskSpaceMonitor struct {
	path      string
	threshold int64
	mu        sync.RWMutex
}

// NewDiskSpaceMonitor creates a new disk space monitor
func NewDiskSpaceMonitor(path string, threshold int64) *DiskSpaceMonitor {
	return &DiskSpaceMonitor{
		path:      path,
		threshold: threshold,
	}
}

// DiskUsage represents disk usage statistics
type DiskUsage struct {
	Total       int64
	Used        int64
	Available   int64
	PercentUsed float64
}

// GetUsage returns current disk usage
func (m *DiskSpaceMonitor) GetUsage() (*DiskUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(m.path, &stat)
	if err != nil {
		return nil, err
	}

	// Calculate disk usage
	total := int64(stat.Blocks) * int64(stat.Bsize)
	available := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - available
	percentUsed := float64(used) / float64(total) * 100

	return &DiskUsage{
		Total:       total,
		Used:        used,
		Available:   available,
		PercentUsed: percentUsed,
	}, nil
}

// IsThresholdExceeded checks if disk usage exceeds threshold
func (m *DiskSpaceMonitor) IsThresholdExceeded() bool {
	usage, err := m.GetUsage()
	if err != nil {
		return false
	}
	return usage.Used > m.threshold
}

// RequestSizeLimiter creates middleware that limits request body size
func RequestSizeLimiter(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
				"code":  "REQUEST_TOO_LARGE",
				"limit": maxSize,
			})
			return
		}
		c.Next()
	}
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int
	timeout      time.Duration
	failures     atomic.Int32
	state        atomic.Int32
	lastFailTime atomic.Int64
	mu           sync.Mutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

// State returns the current circuit state
func (cb *CircuitBreaker) State() CircuitState {
	return CircuitState(cb.state.Load())
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	// Check state
	state := cb.State()

	if state == CircuitOpen {
		// Check if timeout has passed
		lastFail := time.Unix(0, cb.lastFailTime.Load())
		if time.Since(lastFail) > cb.timeout {
			// Transition to half-open
			cb.state.Store(int32(CircuitHalfOpen))
		} else {
			return fmt.Errorf("%w", ErrCircuitBreakerOpen)
		}
	}

	// Execute function
	err := fn()

	if err != nil {
		// Record failure
		failures := cb.failures.Add(1)
		cb.lastFailTime.Store(time.Now().UnixNano())

		if failures >= int32(cb.maxFailures) {
			// Open circuit
			cb.state.Store(int32(CircuitOpen))
		}
		return err
	}

	// Success - reset failures and close circuit
	cb.failures.Store(0)
	cb.state.Store(int32(CircuitClosed))
	return nil
}

// ProcessWithTimeout executes a function with a timeout
func ProcessWithTimeout(fn func() error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- fn()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrTimeoutExceeded, ctx.Err())
	}
}
