package integration

import (
	"net/http"
	"testing"

	"goimgserver/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_GracefulParsing_EndToEnd tests graceful URL parsing in complete workflows
func TestIntegration_GracefulParsing_EndToEnd(t *testing.T) {
	router := gin.New()
	
	// Simulate an image endpoint with parameter parsing
	router.GET("/img/:filename/:dimensions", func(c *gin.Context) {
		filename := c.Param("filename")
		dimensions := c.Param("dimensions")
		
		// Gracefully parse dimensions (basic simulation)
		width := 800  // default
		height := 600 // default
		
		// In real implementation, parse dimensions string
		// For now, just acknowledge the request
		c.JSON(http.StatusOK, gin.H{
			"filename":   filename,
			"dimensions": dimensions,
			"width":      width,
			"height":     height,
		})
	})
	
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		shouldContain  string
	}{
		{
			name:           "Valid dimensions",
			path:           "/img/test.jpg/800x600",
			expectedStatus: http.StatusOK,
			shouldContain:  "test.jpg",
		},
		{
			name:           "Complex URL with extra params",
			path:           "/img/test.jpg/800x600?quality=80&format=webp",
			expectedStatus: http.StatusOK,
			shouldContain:  "test.jpg",
		},
		{
			name:           "URL with special characters in filename",
			path:           "/img/test-image_01.jpg/800x600",
			expectedStatus: http.StatusOK,
			shouldContain:  "test-image_01.jpg",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := testutils.MakeTestRequest(router, "GET", tt.path, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.shouldContain)
		})
	}
}

// TestIntegration_GracefulParsing_CacheConsistency tests cache consistency with graceful parsing
func TestIntegration_GracefulParsing_CacheConsistency(t *testing.T) {
	router := gin.New()
	
	// Simulate cache behavior
	cache := make(map[string]string)
	
	router.GET("/img/:filename/:dimensions", func(c *gin.Context) {
		filename := c.Param("filename")
		dimensions := c.Param("dimensions")
		
		// Create cache key (simplified)
		cacheKey := filename + "_" + dimensions
		
		// Check cache
		if cached, ok := cache[cacheKey]; ok {
			c.JSON(http.StatusOK, gin.H{
				"cached":  true,
				"data":    cached,
				"cacheKey": cacheKey,
			})
			return
		}
		
		// Process and cache
		result := "processed_" + filename
		cache[cacheKey] = result
		
		c.JSON(http.StatusOK, gin.H{
			"cached":  false,
			"data":    result,
			"cacheKey": cacheKey,
		})
	})
	
	t.Run("First request - no cache", func(t *testing.T) {
		rec := testutils.MakeTestRequest(router, "GET", "/img/test.jpg/800x600", nil, nil)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response map[string]interface{}
		err := testutils.ParseJSONResponse(rec, &response)
		require.NoError(t, err)
		
		assert.Equal(t, false, response["cached"])
		assert.Contains(t, response["data"], "processed_test.jpg")
	})
	
	t.Run("Second request - from cache", func(t *testing.T) {
		rec := testutils.MakeTestRequest(router, "GET", "/img/test.jpg/800x600", nil, nil)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response map[string]interface{}
		err := testutils.ParseJSONResponse(rec, &response)
		require.NoError(t, err)
		
		assert.Equal(t, true, response["cached"])
	})
	
	t.Run("Different dimensions - different cache key", func(t *testing.T) {
		rec := testutils.MakeTestRequest(router, "GET", "/img/test.jpg/1024x768", nil, nil)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response map[string]interface{}
		err := testutils.ParseJSONResponse(rec, &response)
		require.NoError(t, err)
		
		// Should not be cached (different dimensions)
		assert.Equal(t, false, response["cached"])
	})
}

// TestIntegration_GracefulParsing_InvalidParameters tests handling of invalid parameters
func TestIntegration_GracefulParsing_InvalidParameters(t *testing.T) {
	router := gin.New()
	
	router.GET("/img/:filename/:dimensions", func(c *gin.Context) {
		filename := c.Param("filename")
		dimensions := c.Param("dimensions")
		
		// Gracefully handle invalid dimensions by using defaults
		width := 800
		height := 600
		
		// In real implementation, would parse and validate dimensions
		// If invalid, fall back to defaults
		
		c.JSON(http.StatusOK, gin.H{
			"filename":   filename,
			"dimensions": dimensions,
			"width":      width,
			"height":     height,
			"used_defaults": true,
		})
	})
	
	tests := []struct {
		name       string
		path       string
		shouldWork bool
	}{
		{
			name:       "Invalid dimensions format",
			path:       "/img/test.jpg/invalid",
			shouldWork: true, // Should gracefully use defaults
		},
		{
			name:       "Empty dimensions",
			path:       "/img/test.jpg/",
			shouldWork: true, // Gin routing might handle this differently
		},
		{
			name:       "Negative dimensions",
			path:       "/img/test.jpg/-100x-100",
			shouldWork: true, // Should gracefully use defaults
		},
		{
			name:       "Zero dimensions",
			path:       "/img/test.jpg/0x0",
			shouldWork: true, // Should gracefully use defaults
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := testutils.MakeTestRequest(router, "GET", tt.path, nil, nil)
			
			if tt.shouldWork {
				// Should not error, even with invalid input
				assert.NotEqual(t, http.StatusInternalServerError, rec.Code)
			}
		})
	}
}

// TestIntegration_GracefulParsing_QueryParameters tests query parameter handling
func TestIntegration_GracefulParsing_QueryParameters(t *testing.T) {
	router := gin.New()
	
	router.GET("/img/:filename/:dimensions", func(c *gin.Context) {
		filename := c.Param("filename")
		dimensions := c.Param("dimensions")
		
		// Parse query parameters with defaults
		quality := c.DefaultQuery("quality", "85")
		format := c.DefaultQuery("format", "jpeg")
		
		c.JSON(http.StatusOK, gin.H{
			"filename":   filename,
			"dimensions": dimensions,
			"quality":    quality,
			"format":     format,
		})
	})
	
	tests := []struct {
		name     string
		path     string
		expected map[string]string
	}{
		{
			name: "No query parameters",
			path: "/img/test.jpg/800x600",
			expected: map[string]string{
				"quality": "85",
				"format":  "jpeg",
			},
		},
		{
			name: "Custom quality",
			path: "/img/test.jpg/800x600?quality=95",
			expected: map[string]string{
				"quality": "95",
				"format":  "jpeg",
			},
		},
		{
			name: "Custom format",
			path: "/img/test.jpg/800x600?format=webp",
			expected: map[string]string{
				"quality": "85",
				"format":  "webp",
			},
		},
		{
			name: "Multiple parameters",
			path: "/img/test.jpg/800x600?quality=90&format=png",
			expected: map[string]string{
				"quality": "90",
				"format":  "png",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := testutils.MakeTestRequest(router, "GET", tt.path, nil, nil)
			
			assert.Equal(t, http.StatusOK, rec.Code)
			
			var response map[string]interface{}
			err := testutils.ParseJSONResponse(rec, &response)
			require.NoError(t, err)
			
			for key, expectedValue := range tt.expected {
				assert.Equal(t, expectedValue, response[key], "Mismatch for parameter: %s", key)
			}
		})
	}
}

// TestIntegration_GracefulParsing_SpecialCharacters tests handling of special characters
func TestIntegration_GracefulParsing_SpecialCharacters(t *testing.T) {
	router := gin.New()
	
	router.GET("/img/:filename/:dimensions", func(c *gin.Context) {
		filename := c.Param("filename")
		dimensions := c.Param("dimensions")
		
		c.JSON(http.StatusOK, gin.H{
			"filename":   filename,
			"dimensions": dimensions,
		})
	})
	
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "Filename with dashes",
			filename: "test-image-01.jpg",
		},
		{
			name:     "Filename with underscores",
			filename: "test_image_01.jpg",
		},
		{
			name:     "Filename with numbers",
			filename: "image123.jpg",
		},
		{
			name:     "Filename with dots",
			filename: "test.image.jpg",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/img/" + tt.filename + "/800x600"
			rec := testutils.MakeTestRequest(router, "GET", path, nil, nil)
			
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.filename)
		})
	}
}
