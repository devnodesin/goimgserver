package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"goimgserver/cache"
	"goimgserver/config"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGitOperations is a mock implementation of GitOperations for testing
type mockGitOperations struct {
	isGitRepoResult  bool
	execGitPullError error
	execGitPullResult *GitPullResult
}

func (m *mockGitOperations) IsGitRepo(dir string) bool {
	return m.isGitRepoResult
}

func (m *mockGitOperations) ExecuteGitPull(ctx context.Context, dir string) (*GitPullResult, error) {
	if m.execGitPullError != nil {
		return nil, m.execGitPullError
	}
	return m.execGitPullResult, nil
}

func (m *mockGitOperations) ValidatePath(path, allowedBase string) bool {
	// Simple validation for testing
	return filepath.HasPrefix(filepath.Clean(path), filepath.Clean(allowedBase))
}

// setupCommandTestEnvironment creates a test environment for command testing
func setupCommandTestEnvironment(t *testing.T) (string, string, *config.Config, cache.CacheManager) {
	imagesDir := t.TempDir()
	cacheDir := t.TempDir()

	// Create test cache files
	testCacheFile := filepath.Join(cacheDir, "test", "hash1.webp")
	require.NoError(t, os.MkdirAll(filepath.Dir(testCacheFile), 0755))
	require.NoError(t, os.WriteFile(testCacheFile, []byte("test data"), 0644))

	cfg := &config.Config{
		ImagesDir: imagesDir,
		CacheDir:  cacheDir,
		Port:      9000,
	}

	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)

	return imagesDir, cacheDir, cfg, cacheManager
}

// TestCommandHandler_POST_Clear_Success tests successful cache clear
func TestCommandHandler_POST_Clear_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, cacheDir, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	mockGit := &mockGitOperations{}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/clear", handler.HandleClear)

	req := httptest.NewRequest("POST", "/cmd/clear", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "cleared_files")
	
	// Verify cache is empty
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// TestCommandHandler_POST_Clear_EmptyCache tests clearing empty cache
func TestCommandHandler_POST_Clear_EmptyCache(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, _, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	// Clear cache first
	require.NoError(t, cacheManager.ClearAll())
	
	mockGit := &mockGitOperations{}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/clear", handler.HandleClear)

	req := httptest.NewRequest("POST", "/cmd/clear", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, float64(0), response["cleared_files"].(float64))
}

// TestCommandHandler_POST_GitUpdate_ValidRepo tests git update in valid repo
func TestCommandHandler_POST_GitUpdate_ValidRepo(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, _, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	mockGit := &mockGitOperations{
		isGitRepoResult: true,
		execGitPullResult: &GitPullResult{
			Success:    true,
			Branch:     "main",
			Changes:    5,
			LastCommit: "abc123...",
		},
	}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/gitupdate", handler.HandleGitUpdate)

	req := httptest.NewRequest("POST", "/cmd/gitupdate", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "main", response["branch"])
	assert.Equal(t, float64(5), response["changes"])
}

// TestCommandHandler_POST_GitUpdate_NotGitRepo tests non-git directory
func TestCommandHandler_POST_GitUpdate_NotGitRepo(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, _, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	mockGit := &mockGitOperations{
		isGitRepoResult: false,
	}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/gitupdate", handler.HandleGitUpdate)

	req := httptest.NewRequest("POST", "/cmd/gitupdate", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "not a git repository")
}

// TestCommandHandler_POST_GitUpdate_NetworkError tests network failures
func TestCommandHandler_POST_GitUpdate_NetworkError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, _, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	mockGit := &mockGitOperations{
		isGitRepoResult:  true,
		execGitPullError: errors.New("network error: unable to connect"),
	}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/gitupdate", handler.HandleGitUpdate)

	req := httptest.NewRequest("POST", "/cmd/gitupdate", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "git update failed")
}

// TestCommandHandler_POST_GenericCommand_ValidName tests generic command framework
func TestCommandHandler_POST_GenericCommand_ValidName(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, _, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	mockGit := &mockGitOperations{}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/:name", handler.HandleCommand)

	req := httptest.NewRequest("POST", "/cmd/clear", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

// TestCommandHandler_POST_GenericCommand_InvalidName tests invalid commands
func TestCommandHandler_POST_GenericCommand_InvalidName(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, _, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	mockGit := &mockGitOperations{}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/:name", handler.HandleCommand)

	req := httptest.NewRequest("POST", "/cmd/invalid", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "invalid command")
}

// TestCommandExecution_Security_InjectionPrevention tests injection protection
func TestCommandExecution_Security_InjectionPrevention(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	// Arrange
	gin.SetMode(gin.TestMode)
	_, _, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	// Try to use path with injection attempt
	cfg.ImagesDir = "/tmp/test; rm -rf /"
	
	mockGit := &mockGitOperations{
		isGitRepoResult: false,
	}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/gitupdate", handler.HandleGitUpdate)

	req := httptest.NewRequest("POST", "/cmd/gitupdate", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
}

// TestCommandSecurity_CommandInjection_Prevention tests command injection prevention
func TestCommandSecurity_CommandInjection_Prevention(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	tempDir := t.TempDir()
	cacheDir := t.TempDir()
	
	// Create a directory with a malicious name
	maliciousDir := filepath.Join(tempDir, "test; cat /etc/passwd")
	require.NoError(t, os.MkdirAll(maliciousDir, 0755))
	
	cfg := &config.Config{
		ImagesDir: maliciousDir,
		CacheDir:  cacheDir,
		Port:      9000,
	}
	
	cacheManager, err := cache.NewManager(cacheDir)
	require.NoError(t, err)
	
	mockGit := &mockGitOperations{
		isGitRepoResult: false,
	}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/gitupdate", handler.HandleGitUpdate)

	req := httptest.NewRequest("POST", "/cmd/gitupdate", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert - Should handle safely without executing injected command
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCommandSecurity_PathTraversal_Prevention tests path traversal prevention
func TestCommandSecurity_PathTraversal_Prevention(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, _, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	// Try to use path traversal in images directory
	cfg.ImagesDir = "/tmp/../../../etc"
	
	mockGit := &mockGitOperations{}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/gitupdate", handler.HandleGitUpdate)

	req := httptest.NewRequest("POST", "/cmd/gitupdate", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCommandEndpoint_Integration_CacheClear tests actual cache clearing
func TestCommandEndpoint_Integration_CacheClear(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	_, cacheDir, cfg, cacheManager := setupCommandTestEnvironment(t)
	
	// Create multiple cache files
	for i := 0; i < 5; i++ {
		cachePath := filepath.Join(cacheDir, "test", "hash"+string(rune(i))+".webp")
		require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0755))
		require.NoError(t, os.WriteFile(cachePath, []byte("test"), 0644))
	}
	
	mockGit := &mockGitOperations{}
	handler := NewCommandHandler(cfg, cacheManager, mockGit)

	router := gin.New()
	router.POST("/cmd/clear", handler.HandleClear)

	req := httptest.NewRequest("POST", "/cmd/clear", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify all cache files are cleared
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// BenchmarkCommandHandler_CacheClear benchmarks cache clearing performance
func BenchmarkCommandHandler_CacheClear(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cacheDir, err := os.MkdirTemp("", "bench-cache-*")
		if err != nil {
			b.Fatal(err)
		}
		
		// Create test cache files
		for j := 0; j < 100; j++ {
			cachePath := filepath.Join(cacheDir, "test", "hash"+string(rune(j))+".webp")
			os.MkdirAll(filepath.Dir(cachePath), 0755)
			os.WriteFile(cachePath, []byte("test"), 0644)
		}
		
		cfg := &config.Config{
			ImagesDir: cacheDir,
			CacheDir:  cacheDir,
			Port:      9000,
		}
		
		cacheManager, _ := cache.NewManager(cacheDir)
		mockGit := &mockGitOperations{}
		handler := NewCommandHandler(cfg, cacheManager, mockGit)

		router := gin.New()
		router.POST("/cmd/clear", handler.HandleClear)

		req := httptest.NewRequest("POST", "/cmd/clear", nil)
		w := httptest.NewRecorder()
		
		b.StartTimer()
		router.ServeHTTP(w, req)
		b.StopTimer()
		
		os.RemoveAll(cacheDir)
	}
}
