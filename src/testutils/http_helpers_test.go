package testutils

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestHelpers_HTTPTestUtils_RequestBuilder tests HTTP request building
func TestHelpers_HTTPTestUtils_RequestBuilder(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		headers        map[string]string
		expectedMethod string
		expectedPath   string
	}{
		{
			name:           "GET request without body",
			method:         "GET",
			path:           "/api/test",
			body:           nil,
			headers:        nil,
			expectedMethod: "GET",
			expectedPath:   "/api/test",
		},
		{
			name:           "POST request with body",
			method:         "POST",
			path:           "/api/data",
			body:           []byte(`{"key":"value"}`),
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedMethod: "POST",
			expectedPath:   "/api/data",
		},
		{
			name:   "PUT request with custom headers",
			method: "PUT",
			path:   "/api/update",
			body:   []byte(`{"updated":true}`),
			headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
			expectedMethod: "PUT",
			expectedPath:   "/api/update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRequestBuilder(tt.method, tt.path)
			
			if tt.body != nil {
				builder.WithBody(tt.body)
			}
			
			for key, value := range tt.headers {
				builder.WithHeader(key, value)
			}
			
			req := builder.Build()
			
			assert.Equal(t, tt.expectedMethod, req.Method)
			assert.Equal(t, tt.expectedPath, req.URL.Path)
			
			// Verify headers
			for key, expectedValue := range tt.headers {
				assert.Equal(t, expectedValue, req.Header.Get(key))
			}
			
			// Verify body
			if tt.body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.body, bodyBytes)
			}
		})
	}
}

// TestRequestBuilder_WithQueryParams tests query parameter handling
func TestRequestBuilder_WithQueryParams(t *testing.T) {
	builder := NewRequestBuilder("GET", "/api/search")
	builder.WithQueryParam("q", "test")
	builder.WithQueryParam("page", "1")
	builder.WithQueryParam("limit", "10")
	
	req := builder.Build()
	
	assert.Equal(t, "test", req.URL.Query().Get("q"))
	assert.Equal(t, "1", req.URL.Query().Get("page"))
	assert.Equal(t, "10", req.URL.Query().Get("limit"))
}

// TestRequestBuilder_Chaining tests method chaining
func TestRequestBuilder_Chaining(t *testing.T) {
	req := NewRequestBuilder("POST", "/api/test").
		WithHeader("Content-Type", "application/json").
		WithHeader("Authorization", "Bearer token").
		WithBody([]byte(`{"test":true}`)).
		WithQueryParam("debug", "true").
		Build()
	
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/api/test", req.URL.Path)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", req.Header.Get("Authorization"))
	assert.Equal(t, "true", req.URL.Query().Get("debug"))
}

// TestResponseRecorder_HelperMethods tests response recording helpers
func TestResponseRecorder_HelperMethods(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := NewRequestBuilder("GET", "/test").Build()
	rec := NewResponseRecorder()
	
	router.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "success")
	assert.Contains(t, rec.Header().Get("Content-Type"), "json")
}

// TestMakeTestRequest tests simple request execution
func TestMakeTestRequest(t *testing.T) {
	router := gin.New()
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	
	rec := MakeTestRequest(router, "GET", "/ping", nil, nil)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())
}

// TestMakeTestRequestWithHeaders tests request with custom headers
func TestMakeTestRequestWithHeaders(t *testing.T) {
	router := gin.New()
	router.GET("/auth", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "Bearer valid-token" {
			c.JSON(http.StatusOK, gin.H{"authenticated": true})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		}
	})
	
	headers := map[string]string{
		"Authorization": "Bearer valid-token",
	}
	
	rec := MakeTestRequest(router, "GET", "/auth", nil, headers)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "true")
}

// TestMakeJSONRequest tests JSON request handling
func TestMakeJSONRequest(t *testing.T) {
	router := gin.New()
	router.POST("/data", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"received": data})
	})
	
	payload := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}
	
	rec := MakeJSONRequest(router, "POST", "/data", payload)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test")
	assert.Contains(t, rec.Body.String(), "123")
}

// TestAssertJSONResponse tests JSON response assertions
func TestAssertJSONResponse(t *testing.T) {
	router := gin.New()
	router.GET("/user", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":   1,
			"name": "John Doe",
			"age":  30,
		})
	})
	
	rec := MakeTestRequest(router, "GET", "/user", nil, nil)
	
	var response map[string]interface{}
	err := AssertJSONResponse(t, rec, &response)
	require.NoError(t, err)
	
	assert.Equal(t, float64(1), response["id"])
	assert.Equal(t, "John Doe", response["name"])
	assert.Equal(t, float64(30), response["age"])
}

// TestAssertErrorResponse tests error response format
func TestAssertErrorResponse(t *testing.T) {
	router := gin.New()
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
			"code":  "INVALID_INPUT",
		})
	})
	
	rec := MakeTestRequest(router, "GET", "/error", nil, nil)
	
	var response map[string]interface{}
	err := AssertJSONResponse(t, rec, &response)
	require.NoError(t, err)
	
	assert.Equal(t, "invalid input", response["error"])
	assert.Equal(t, "INVALID_INPUT", response["code"])
}

// TestCreateTestRouter tests test router creation
func TestCreateTestRouter(t *testing.T) {
	router := CreateTestRouter()
	
	assert.NotNil(t, router)
	
	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})
	
	rec := MakeTestRequest(router, "GET", "/test", nil, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test response", rec.Body.String())
}

// TestCreateTestRouterWithMiddleware tests router with middleware
func TestCreateTestRouterWithMiddleware(t *testing.T) {
	middleware := func(c *gin.Context) {
		c.Set("middleware_called", true)
		c.Next()
	}
	
	router := CreateTestRouterWithMiddleware(middleware)
	router.GET("/test", func(c *gin.Context) {
		called, exists := c.Get("middleware_called")
		if exists && called.(bool) {
			c.String(http.StatusOK, "middleware worked")
		} else {
			c.String(http.StatusInternalServerError, "middleware failed")
		}
	})
	
	rec := MakeTestRequest(router, "GET", "/test", nil, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "middleware worked", rec.Body.String())
}

// TestParseJSONResponse tests JSON response parsing utility
func TestParseJSONResponse(t *testing.T) {
	jsonData := `{"name":"test","value":42,"active":true}`
	rec := httptest.NewRecorder()
	rec.Body = bytes.NewBufferString(jsonData)
	
	var result map[string]interface{}
	err := ParseJSONResponse(rec, &result)
	require.NoError(t, err)
	
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, float64(42), result["value"])
	assert.Equal(t, true, result["active"])
}
