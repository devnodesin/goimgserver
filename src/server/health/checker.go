package health

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck is a function that returns true if a component is healthy
type HealthCheck func() bool

// Checker manages health checks for the application
type Checker struct {
	checks    map[string]HealthCheck
	startTime time.Time
	mu        sync.RWMutex
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		checks:    make(map[string]HealthCheck),
		startTime: time.Now(),
	}
}

// AddCheck registers a health check function
func (c *Checker) AddCheck(name string, check HealthCheck) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// HealthHandler returns a simple health check endpoint
func (c *Checker) HealthHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// DetailedHealthHandler returns a detailed health check with dependency status
func (c *Checker) DetailedHealthHandler(ctx *gin.Context) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	checkResults := make(map[string]string)
	allHealthy := true
	
	// Run all health checks
	for name, check := range c.checks {
		if check() {
			checkResults[name] = "ok"
		} else {
			checkResults[name] = "failed"
			allHealthy = false
		}
	}
	
	// Calculate uptime
	uptime := time.Since(c.startTime)
	
	status := "ok"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}
	
	response := gin.H{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    uptime.String(),
	}
	
	if len(checkResults) > 0 {
		response["checks"] = checkResults
	}
	
	ctx.JSON(statusCode, response)
}

// ReadinessHandler returns readiness status (can handle traffic)
func (c *Checker) ReadinessHandler(ctx *gin.Context) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Check if all dependencies are healthy
	for _, check := range c.checks {
		if !check() {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
			})
			return
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// LivenessHandler returns liveness status (application is running)
func (c *Checker) LivenessHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
