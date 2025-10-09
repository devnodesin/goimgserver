package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"goimgserver/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGitRepo creates a temporary git repository for testing
func setupGitRepo(t *testing.T) string {
	t.Helper()
	
	tmpDir := t.TempDir()
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Git init output: %s", output)
		t.Skip("Git not available or failed to initialize")
	}
	
	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	
	// Create and commit a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	
	return tmpDir
}

// TestIntegration_Git_IsRepository tests repository detection
func TestIntegration_Git_IsRepository(t *testing.T) {
	gitOps := git.NewOperations()
	
	t.Run("Valid git repository", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		
		isRepo := gitOps.IsGitRepo(repoDir)
		
		assert.True(t, isRepo, "Should detect valid git repository")
	})
	
	t.Run("Non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		isRepo := gitOps.IsGitRepo(tmpDir)
		
		assert.False(t, isRepo, "Should not detect non-git directory as repository")
	})
}

// TestIntegration_Git_ExecutePull tests git pull operation
func TestIntegration_Git_ExecutePull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git pull test in short mode")
	}
	
	repoDir := setupGitRepo(t)
	gitOps := git.NewOperations()
	
	// Pull should work even if there's nothing to pull
	ctx := context.Background()
	result, err := gitOps.ExecuteGitPull(ctx, repoDir)
	
	// The actual result depends on git configuration
	// We just verify it doesn't crash
	_ = result
	_ = err
}

// TestIntegration_Git_NonRepoOperations tests operations on non-repository
func TestIntegration_Git_NonRepoOperations(t *testing.T) {
	tmpDir := t.TempDir()
	gitOps := git.NewOperations()
	
	t.Run("Pull on non-repo", func(t *testing.T) {
		ctx := context.Background()
		result, err := gitOps.ExecuteGitPull(ctx, tmpDir)
		// Should error or handle gracefully
		_ = result
		_ = err
	})
}

// TestIntegration_Git_PathValidation tests path validation
func TestIntegration_Git_PathValidation(t *testing.T) {
	gitOps := git.NewOperations()
	
	tests := []struct {
		name        string
		path        string
		allowedBase string
		shouldPass  bool
	}{
		{
			name:        "Valid path within base",
			path:        "/data/images/test.jpg",
			allowedBase: "/data",
			shouldPass:  true,
		},
		{
			name:        "Path outside base",
			path:        "/etc/passwd",
			allowedBase: "/data",
			shouldPass:  false,
		},
		{
			name:        "Path traversal attempt",
			path:        "/data/../etc/passwd",
			allowedBase: "/data",
			shouldPass:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gitOps.ValidatePath(tt.path, tt.allowedBase)
			assert.Equal(t, tt.shouldPass, result)
		})
	}
}

// TestIntegration_Git_ConcurrentAccess tests concurrent git operations
func TestIntegration_Git_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}
	
	repoDir := setupGitRepo(t)
	gitOps := git.NewOperations()
	
	// Multiple concurrent checks should all work
	for i := 0; i < 5; i++ {
		go func() {
			isRepo := gitOps.IsGitRepo(repoDir)
			assert.True(t, isRepo)
		}()
	}
}

// TestIntegration_Git_ErrorHandling tests error scenarios
func TestIntegration_Git_ErrorHandling(t *testing.T) {
	gitOps := git.NewOperations()
	
	t.Run("Invalid path", func(t *testing.T) {
		isRepo := gitOps.IsGitRepo("/nonexistent/path/to/repo")
		assert.False(t, isRepo)
	})
	
	t.Run("Empty path", func(t *testing.T) {
		isRepo := gitOps.IsGitRepo("")
		assert.False(t, isRepo)
	})
}

