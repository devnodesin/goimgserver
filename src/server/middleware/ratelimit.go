package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate int
	lastRefill time.Time
	mu         sync.Mutex
}

func newRateLimiter(rate int, burst int) *rateLimiter {
	return &rateLimiter{
		tokens:     burst,
		maxTokens:  burst,
		refillRate: rate,
		lastRefill: time.Now(),
	}
}

func (rl *rateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	
	// Calculate tokens to add based on elapsed time and refill rate
	tokensToAdd := int(elapsed.Seconds() * float64(rl.refillRate))
	
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}
	
	// Check if request can be allowed
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	
	return false
}

// RateLimit returns a middleware that limits requests globally
func RateLimit(rate int, per time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(int(float64(rate)/per.Seconds()), rate)
	
	return func(c *gin.Context) {
		if !limiter.allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": per.String(),
			})
			return
		}
		c.Next()
	}
}

// RateLimitPerIP returns a middleware that limits requests per IP address
func RateLimitPerIP(rate int, per time.Duration) gin.HandlerFunc {
	limiters := make(map[string]*rateLimiter)
	var mu sync.Mutex
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		mu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			limiter = newRateLimiter(int(float64(rate)/per.Seconds()), rate)
			limiters[ip] = limiter
		}
		mu.Unlock()
		
		if !limiter.allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": per.String(),
			})
			return
		}
		c.Next()
	}
}

// RateLimitWithBurst returns a middleware that limits requests with burst capacity
func RateLimitWithBurst(rate int, per time.Duration, burst int) gin.HandlerFunc {
	limiter := newRateLimiter(int(float64(rate)/per.Seconds()), burst)
	
	return func(c *gin.Context) {
		if !limiter.allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": per.String(),
			})
			return
		}
		c.Next()
	}
}
