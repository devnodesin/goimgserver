package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID_Generation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	var capturedID string
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		capturedID = c.GetString("request_id")
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedID, "Request ID should be generated")
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"), "Request ID should be in response header")
	assert.Equal(t, capturedID, w.Header().Get("X-Request-ID"), "Request ID should match")
}

func TestRequestID_Propagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	var capturedID string
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		capturedID = c.GetString("request_id")
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	existingID := "test-request-id-123"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, existingID, capturedID, "Should use existing request ID")
	assert.Equal(t, existingID, w.Header().Get("X-Request-ID"), "Should propagate existing request ID")
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	var requestIDs []string
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestIDs = append(requestIDs, c.GetString("request_id"))
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Make multiple requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Check all IDs are unique
	assert.Len(t, requestIDs, 5)
	uniqueIDs := make(map[string]bool)
	for _, id := range requestIDs {
		assert.False(t, uniqueIDs[id], "Request ID should be unique: %s", id)
		uniqueIDs[id] = true
	}
}

func TestRequestID_ContextStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		assert.NotEmpty(t, requestID, "Request ID should be stored in context")
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
