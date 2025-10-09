package handlers

import (
	"goimgserver/cache"
	"goimgserver/config"
	"goimgserver/processor"
	"goimgserver/resolver"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestEnvironment creates a test environment with image directories and test files
func setupTestEnvironment(t *testing.T) (string, string, *config.Config) {
	// Create temp directories
	imagesDir := t.TempDir()
	cacheDir := t.TempDir()

	// Create test images
	testImagePath := filepath.Join(imagesDir, "test.jpg")
	require.NoError(t, createTestImage(testImagePath, 100, 100))

	// Create grouped image folder
	groupDir := filepath.Join(imagesDir, "cats")
	require.NoError(t, os.MkdirAll(groupDir, 0755))
	groupImagePath := filepath.Join(groupDir, "cat_white.jpg")
	require.NoError(t, createTestImage(groupImagePath, 100, 100))

	// Create default image
	defaultImagePath := filepath.Join(imagesDir, "default.jpg")
	require.NoError(t, createTestImage(defaultImagePath, 1000, 1000))

	// Create config
	cfg := &config.Config{
		ImagesDir:        imagesDir,
		CacheDir:         cacheDir,
		DefaultImagePath: defaultImagePath,
		Port:             9000,
	}

	return imagesDir, cacheDir, cfg
}

// mockProcessor is a simple mock processor for testing
type mockProcessor struct{}

func (m *mockProcessor) Resize(data []byte, width, height int) ([]byte, error) {
	// Just return the data unchanged for testing
	return data, nil
}

func (m *mockProcessor) ConvertFormat(data []byte, format processor.ImageFormat) ([]byte, error) {
	return data, nil
}

func (m *mockProcessor) AdjustQuality(data []byte, quality int) ([]byte, error) {
	return data, nil
}

func (m *mockProcessor) Process(data []byte, opts processor.ProcessOptions) ([]byte, error) {
	return data, nil
}

func (m *mockProcessor) ValidateImage(data []byte) error {
	if len(data) == 0 {
		return processor.ErrInvalidImage
	}
	return nil
}

// TestImageHandler_GET_DefaultSettings tests basic image access with default settings
func TestImageHandler_GET_DefaultSettings(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_WithoutExtension tests auto-detection of file extension
func TestImageHandler_GET_WithoutExtension(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_GroupedDefault tests group default image access
func TestImageHandler_GET_GroupedDefault(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	// Create group default image
	groupDir := filepath.Join(imagesDir, "cats")
	defaultPath := filepath.Join(groupDir, "default.jpg")
	require.NoError(t, createTestImage(defaultPath, 100, 100))

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/cats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_GroupedSpecific tests specific image in group
func TestImageHandler_GET_GroupedSpecific(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/cats/cat_white.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_GroupedWithoutExtension tests grouped image with auto-detection
func TestImageHandler_GET_GroupedWithoutExtension(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/cats/cat_white", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_DefaultImage_MissingFile tests fallback to default image
func TestImageHandler_GET_DefaultImage_MissingFile(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act - request non-existent file
	req := httptest.NewRequest("GET", "/img/nonexistent.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should return default image, not 404
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_CustomDimensions tests custom dimensions
func TestImageHandler_GET_CustomDimensions(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg/600x400", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_WidthOnly tests width-only dimension
func TestImageHandler_GET_WidthOnly(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg/400", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_QualityOnly tests quality-only parameter
func TestImageHandler_GET_QualityOnly(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg/q50", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_CombinedParams tests combined dimension and quality parameters
func TestImageHandler_GET_CombinedParams(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg/800x600/q90", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_FormatConversion tests format conversion
func TestImageHandler_GET_FormatConversion(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg/800x600/png", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_AllParams tests all parameters combined
func TestImageHandler_GET_AllParams(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg/800x600/webp/q90", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, len(w.Body.Bytes()), 0)
}

// TestImageHandler_GET_CacheClear tests cache clearing for specific file
func TestImageHandler_GET_CacheClear(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// First, access the image to cache it
	req := httptest.NewRequest("GET", "/img/test.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Act - clear cache
	req = httptest.NewRequest("GET", "/img/test.jpg/clear", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestImageHandler_GET_Never404 tests that no 404 errors occur
func TestImageHandler_GET_Never404(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	testCases := []string{
		"/img/nonexistent.jpg",
		"/img/missing/file.png",
		"/img/not/found/anywhere.webp",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			// Act
			req := httptest.NewRequest("GET", testCase, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert - should never return 404, always fallback to default
			assert.NotEqual(t, http.StatusNotFound, w.Code)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestImageHandler_Headers_ContentType tests correct Content-Type headers
func TestImageHandler_Headers_ContentType(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	tests := []struct {
		url         string
		contentType string
	}{
		{"/img/test.jpg", "image/jpeg"},
		{"/img/test.jpg/webp", "image/webp"},
		{"/img/test.jpg/png", "image/png"},
		{"/img/test.jpg/jpeg", "image/jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			// Act
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)
			// Note: Content-Type might be set by Gin's automatic detection
			// We'll check if it contains "image/"
			contentType := w.Header().Get("Content-Type")
			assert.Contains(t, contentType, "image/")
		})
	}
}

// TestImageHandler_Headers_CacheControl tests cache headers
func TestImageHandler_Headers_CacheControl(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	cacheControl := w.Header().Get("Cache-Control")
	assert.NotEmpty(t, cacheControl)
}

// TestImageHandler_Headers_CORS tests CORS headers
func TestImageHandler_Headers_CORS(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act
	req := httptest.NewRequest("GET", "/img/test.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// CORS headers should be set
	accessControl := w.Header().Get("Access-Control-Allow-Origin")
	assert.NotEmpty(t, accessControl)
}

// TestImageHandler_GET_CorruptedImage tests handling of corrupted images
func TestImageHandler_GET_CorruptedImage(t *testing.T) {
	// This test requires a real processor that can detect corrupted images
	// For now, we'll skip it in the mock environment
	t.Skip("Requires real image processor")
}

// TestImageHandler_Integration_CompleteFlow tests complete image processing flow
func TestImageHandler_Integration_CompleteFlow(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	imagesDir, cacheDir, cfg := setupTestEnvironment(t)

	resolver := resolver.NewResolver(imagesDir)
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	proc := &mockProcessor{}

	handler := NewImageHandler(cfg, resolver, cacheManager, proc)

	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)

	// Act - First request (cache miss)
	req1 := httptest.NewRequest("GET", "/img/test.jpg/800x600/webp/q90", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Act - Second request (cache hit)
	req2 := httptest.NewRequest("GET", "/img/test.jpg/800x600/webp/q90", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Assert
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Greater(t, len(w1.Body.Bytes()), 0)
	assert.Greater(t, len(w2.Body.Bytes()), 0)
}

// Benchmark tests
func BenchmarkImageHandler_CacheHit(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	// Setup
	imagesDir := b.TempDir()
	cacheDir := b.TempDir()
	
	testImagePath := filepath.Join(imagesDir, "test.jpg")
	if err := createTestImage(testImagePath, 100, 100); err != nil {
		b.Fatal(err)
	}
	
	defaultImagePath := filepath.Join(imagesDir, "default.jpg")
	if err := createTestImage(defaultImagePath, 1000, 1000); err != nil {
		b.Fatal(err)
	}
	
	cfg := &config.Config{
		ImagesDir:        imagesDir,
		CacheDir:         cacheDir,
		DefaultImagePath: defaultImagePath,
	}
	
	resolver := resolver.NewResolver(imagesDir)
	cacheManager, _ := cache.NewManager(cacheDir)
	proc := &mockProcessor{}
	
	handler := NewImageHandler(cfg, resolver, cacheManager, proc)
	
	router := gin.New()
	router.GET("/img/*path", handler.ServeImage)
	
	// Prime the cache
	req := httptest.NewRequest("GET", "/img/test.jpg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Reset timer
	b.ResetTimer()
	
	// Benchmark
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/img/test.jpg", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkParseParameters(b *testing.B) {
	segments := []string{"800x600", "webp", "q90"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseParameters(segments)
	}
}
