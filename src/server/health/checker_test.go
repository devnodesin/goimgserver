package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck_BasicEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	checker := NewChecker()
	router.GET("/health", checker.HealthHandler)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestHealthCheck_DetailedStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	checker := NewChecker()
	router.GET("/health", checker.DetailedHealthHandler)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "uptime")
}

func TestHealthCheck_DependencyChecks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	checker := NewChecker()
	
	// Add dependency checks
	checker.AddCheck("cache", func() bool {
		return true
	})
	checker.AddCheck("filesystem", func() bool {
		return true
	})
	
	router.GET("/health", checker.DetailedHealthHandler)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	
	// Check dependencies
	if checks, ok := response["checks"].(map[string]interface{}); ok {
		assert.Equal(t, "ok", checks["cache"])
		assert.Equal(t, "ok", checks["filesystem"])
	}
}

func TestHealthCheck_FailedDependency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	checker := NewChecker()
	
	// Add failing dependency check
	checker.AddCheck("database", func() bool {
		return false
	})
	
	router.GET("/health", checker.DetailedHealthHandler)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 503 when a dependency fails
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "degraded", response["status"])
}

func TestHealthCheck_MultipleChecks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	checker := NewChecker()
	
	checker.AddCheck("check1", func() bool { return true })
	checker.AddCheck("check2", func() bool { return true })
	checker.AddCheck("check3", func() bool { return false })
	
	router.GET("/health", checker.DetailedHealthHandler)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// Check individual results
	if checks, ok := response["checks"].(map[string]interface{}); ok {
		assert.Equal(t, "ok", checks["check1"])
		assert.Equal(t, "ok", checks["check2"])
		assert.Equal(t, "failed", checks["check3"])
	}
}

func TestHealthCheck_ReadinessEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	checker := NewChecker()
	router.GET("/ready", checker.ReadinessHandler)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
}

func TestHealthCheck_LivenessEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	checker := NewChecker()
	router.GET("/live", checker.LivenessHandler)

	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
}
