package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func BenchmarkMiddleware_CORS_Overhead(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_Logging_Overhead(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_RequestID_Overhead(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_SecurityHeaders_Overhead(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_ErrorHandler_Overhead(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_FullStack(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Full middleware stack
	router.Use(RequestID())
	router.Use(SecurityHeaders())
	router.Use(CORS())
	router.Use(ErrorHandler())
	router.Use(Logging())
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkServer_RequestThroughput(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(RequestID())
	router.Use(SecurityHeaders())
	router.Use(CORS())
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkRateLimit_HighLoad(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// High rate limit for benchmarking
	router.Use(RateLimit(100000, time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
