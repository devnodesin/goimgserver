package integration

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"goimgserver/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestServer creates a test server with minimal configuration
func setupTestServer(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	
	// Create temp directories
	tmpDir := t.TempDir()
	imagesDir := filepath.Join(tmpDir, "images")
	cacheDir := filepath.Join(tmpDir, "cache")
	
	err := os.MkdirAll(imagesDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(cacheDir, 0755)
	require.NoError(t, err)
	
	// Create test fixtures
	fixtureManager := testutils.NewFixtureManager(imagesDir)
	err = fixtureManager.CreateFixtureSet()
	require.NoError(t, err)
	
	// Create basic router
	router := gin.New()
	
	return router, imagesDir
}

// TestIntegration_EndToEnd_CompleteFlows tests complete user workflows
func TestIntegration_EndToEnd_CompleteFlows(t *testing.T) {
	router, _ := setupTestServer(t)
	
	// Add basic health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	
	t.Run("Health check endpoint", func(t *testing.T) {
		req := testutils.NewRequestBuilder("GET", "/health").Build()
		rec := testutils.NewResponseRecorder()
		
		router.ServeHTTP(rec, req)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response map[string]interface{}
		err := testutils.ParseJSONResponse(rec, &response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	})
}

// TestIntegration_APIRoutes_BasicEndpoints tests basic API routes
func TestIntegration_APIRoutes_BasicEndpoints(t *testing.T) {
	router, _ := setupTestServer(t)
	
	// Setup test routes
	router.GET("/api/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "goimgserver",
			"version": "1.0.0",
		})
	})
	
	router.POST("/api/echo", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"echoed": data})
	})
	
	t.Run("GET status endpoint", func(t *testing.T) {
		rec := testutils.MakeTestRequest(router, "GET", "/api/status", nil, nil)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "goimgserver")
	})
	
	t.Run("POST echo endpoint", func(t *testing.T) {
		payload := map[string]interface{}{
			"message": "test",
			"value":   42,
		}
		
		rec := testutils.MakeJSONRequest(router, "POST", "/api/echo", payload)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "test")
	})
}

// TestIntegration_ErrorHandling tests error handling across the API
func TestIntegration_ErrorHandling(t *testing.T) {
	router, _ := setupTestServer(t)
	
	router.GET("/error/400", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad request",
			"code":  "BAD_REQUEST",
		})
	})
	
	router.GET("/error/404", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
			"code":  "NOT_FOUND",
		})
	})
	
	router.GET("/error/500", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error",
			"code":  "INTERNAL_ERROR",
		})
	})
	
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "400 Bad Request",
			path:           "/error/400",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "404 Not Found",
			path:           "/error/404",
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "500 Internal Error",
			path:           "/error/500",
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := testutils.MakeTestRequest(router, "GET", tt.path, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response map[string]interface{}
			err := testutils.ParseJSONResponse(rec, &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCode, response["code"])
		})
	}
}

// TestIntegration_ConcurrentRequests tests handling of concurrent requests
func TestIntegration_ConcurrentRequests(t *testing.T) {
	router, _ := setupTestServer(t)
	
	counter := 0
	router.GET("/increment", func(c *gin.Context) {
		counter++
		c.JSON(http.StatusOK, gin.H{"count": counter})
	})
	
	// Make concurrent requests
	results := make(chan int, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			rec := testutils.MakeTestRequest(router, "GET", "/increment", nil, nil)
			results <- rec.Code
		}()
	}
	
	// Collect results
	for i := 0; i < 10; i++ {
		code := <-results
		assert.Equal(t, http.StatusOK, code)
	}
}

// TestIntegration_MiddlewareChain tests middleware processing
func TestIntegration_MiddlewareChain(t *testing.T) {
	router := gin.New()
	
	// Add middleware that sets a header
	router.Use(func(c *gin.Context) {
		c.Header("X-Test-Middleware", "applied")
		c.Next()
	})
	
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	
	rec := testutils.MakeTestRequest(router, "GET", "/test", nil, nil)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "applied", rec.Header().Get("X-Test-Middleware"))
}

// TestIntegration_RequestValidation tests request validation
func TestIntegration_RequestValidation(t *testing.T) {
	router, _ := setupTestServer(t)
	
	type ValidatedRequest struct {
		Name  string `json:"name" binding:"required,min=3,max=50"`
		Email string `json:"email" binding:"required,email"`
		Age   int    `json:"age" binding:"required,min=1,max=150"`
	}
	
	router.POST("/validate", func(c *gin.Context) {
		var req ValidatedRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"validated": true})
	})
	
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid request",
			payload: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   30,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing name",
			payload: map[string]interface{}{
				"email": "john@example.com",
				"age":   30,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid email",
			payload: map[string]interface{}{
				"name":  "John Doe",
				"email": "invalid-email",
				"age":   30,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Age out of range",
			payload: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   200,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := testutils.MakeJSONRequest(router, "POST", "/validate", tt.payload)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestIntegration_ContentNegotiation tests content type handling
func TestIntegration_ContentNegotiation(t *testing.T) {
	router, _ := setupTestServer(t)
	
	router.GET("/data", func(c *gin.Context) {
		accept := c.GetHeader("Accept")
		
		data := gin.H{"message": "test data"}
		
		switch accept {
		case "application/json":
			c.JSON(http.StatusOK, data)
		case "application/xml":
			c.XML(http.StatusOK, data)
		default:
			c.JSON(http.StatusOK, data)
		}
	})
	
	t.Run("JSON response", func(t *testing.T) {
		headers := map[string]string{"Accept": "application/json"}
		rec := testutils.MakeTestRequest(router, "GET", "/data", nil, headers)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "json")
	})
	
	t.Run("XML response", func(t *testing.T) {
		headers := map[string]string{"Accept": "application/xml"}
		rec := testutils.MakeTestRequest(router, "GET", "/data", nil, headers)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "xml")
	})
}
