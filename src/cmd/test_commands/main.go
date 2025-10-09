package main

import (
	"context"
	"encoding/json"
	"fmt"
	"goimgserver/cache"
	"goimgserver/config"
	"goimgserver/git"
	"goimgserver/handlers"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.TestMode)
	
	// Setup test environment
	tempDir, err := os.MkdirTemp("", "cmd-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	
	imagesDir := filepath.Join(tempDir, "images")
	cacheDir := filepath.Join(tempDir, "cache")
	os.MkdirAll(imagesDir, 0755)
	os.MkdirAll(cacheDir, 0755)
	
	// Create test cache files
	for i := 0; i < 5; i++ {
		cachePath := filepath.Join(cacheDir, "test", fmt.Sprintf("hash%d.webp", i))
		os.MkdirAll(filepath.Dir(cachePath), 0755)
		os.WriteFile(cachePath, []byte("test data"), 0644)
	}
	
	cfg := &config.Config{
		ImagesDir: imagesDir,
		CacheDir:  cacheDir,
		Port:      9000,
	}
	
	cacheManager, err := cache.NewManager(cacheDir)
	if err != nil {
		panic(err)
	}
	
	gitOps := git.NewOperations()
	handler := handlers.NewCommandHandler(cfg, cacheManager, gitOps)
	
	router := gin.New()
	router.POST("/cmd/clear", handler.HandleClear)
	router.POST("/cmd/gitupdate", handler.HandleGitUpdate)
	router.POST("/cmd/:name", handler.HandleCommand)
	
	// Test 1: Clear cache
	fmt.Println("Test 1: Clear cache")
	req := httptest.NewRequest("POST", "/cmd/clear", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		fmt.Printf("FAIL: Expected 200, got %d\n", w.Code)
		fmt.Printf("Response: %s\n", w.Body.String())
		os.Exit(1)
	}
	
	var clearResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &clearResp)
	if !clearResp["success"].(bool) {
		fmt.Println("FAIL: Clear cache returned success=false")
		os.Exit(1)
	}
	fmt.Printf("PASS: Cache cleared, files: %.0f\n", clearResp["cleared_files"].(float64))
	
	// Test 2: Git update on non-git directory
	fmt.Println("\nTest 2: Git update on non-git directory")
	req = httptest.NewRequest("POST", "/cmd/gitupdate", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusBadRequest {
		fmt.Printf("FAIL: Expected 400, got %d\n", w.Code)
		os.Exit(1)
	}
	
	var gitResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &gitResp)
	if gitResp["success"].(bool) {
		fmt.Println("FAIL: Git update should fail on non-git directory")
		os.Exit(1)
	}
	fmt.Printf("PASS: Git update correctly rejected non-git directory\n")
	
	// Test 3: Generic command with invalid name
	fmt.Println("\nTest 3: Generic command with invalid name")
	req = httptest.NewRequest("POST", "/cmd/invalid", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusBadRequest {
		fmt.Printf("FAIL: Expected 400, got %d\n", w.Code)
		os.Exit(1)
	}
	fmt.Println("PASS: Invalid command correctly rejected")
	
	// Test 4: Git update with actual git repo
	fmt.Println("\nTest 4: Git update with actual git repo")
	gitDir := filepath.Join(tempDir, "git-repo")
	os.MkdirAll(gitDir, 0755)
	
	// Initialize git repo
	ctx := context.Background()
	_ = ctx
	cmd := exec.Command("git", "init")
	cmd.Dir = gitDir
	if err := cmd.Run(); err == nil {
		cfg.ImagesDir = gitDir
		handler = handlers.NewCommandHandler(cfg, cacheManager, gitOps)
		
		router = gin.New()
		router.POST("/cmd/gitupdate", handler.HandleGitUpdate)
		
		req = httptest.NewRequest("POST", "/cmd/gitupdate", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Git pull might fail if there's no remote, but it should at least try
		if w.Code == http.StatusOK || w.Code == http.StatusInternalServerError {
			fmt.Println("PASS: Git update attempted on valid repo")
		} else {
			fmt.Printf("FAIL: Unexpected status code %d\n", w.Code)
			os.Exit(1)
		}
	} else {
		fmt.Println("SKIP: Git not available for test")
	}
	
	fmt.Println("\n✓ All command handler tests passed!")
}
