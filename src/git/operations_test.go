package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGitOperations_IsGitRepo_ValidRepo tests git repo detection with a valid repo
func TestGitOperations_IsGitRepo_ValidRepo(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())

	ops := NewOperations()

	// Act
	isRepo := ops.IsGitRepo(tempDir)

	// Assert
	assert.True(t, isRepo, "Should detect valid git repository")
}

// TestGitOperations_IsGitRepo_NotGitRepo tests non-git directory detection
func TestGitOperations_IsGitRepo_NotGitRepo(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	ops := NewOperations()

	// Act
	isRepo := ops.IsGitRepo(tempDir)

	// Assert
	assert.False(t, isRepo, "Should not detect non-git directory as repo")
}

// TestGitOperations_IsGitRepo_NonExistentDir tests non-existent directory
func TestGitOperations_IsGitRepo_NonExistentDir(t *testing.T) {
	// Arrange
	ops := NewOperations()
	nonExistentDir := "/path/that/does/not/exist"

	// Act
	isRepo := ops.IsGitRepo(nonExistentDir)

	// Assert
	assert.False(t, isRepo, "Should return false for non-existent directory")
}

// TestGitOperations_ExecuteGitPull_Success tests successful git pull
func TestGitOperations_ExecuteGitPull_Success(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	// Arrange - Create a bare repo and a clone
	tempDir := t.TempDir()
	bareRepo := filepath.Join(tempDir, "bare.git")
	cloneRepo := filepath.Join(tempDir, "clone")

	// Create bare repo
	cmd := exec.Command("git", "init", "--bare", bareRepo)
	require.NoError(t, cmd.Run())

	// Clone the repo
	cmd = exec.Command("git", "clone", bareRepo, cloneRepo)
	require.NoError(t, cmd.Run())

	// Configure user for commits
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = cloneRepo
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = cloneRepo
	require.NoError(t, cmd.Run())

	// Create a file and commit
	testFile := filepath.Join(cloneRepo, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = cloneRepo
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = cloneRepo
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = cloneRepo
	require.NoError(t, cmd.Run())

	ops := NewOperations()
	ctx := context.Background()

	// Act
	result, err := ops.ExecuteGitPull(ctx, cloneRepo)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

// TestGitOperations_ExecuteGitPull_NotGitRepo tests git pull on non-git directory
func TestGitOperations_ExecuteGitPull_NotGitRepo(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	ops := NewOperations()
	ctx := context.Background()

	// Act
	result, err := ops.ExecuteGitPull(ctx, tempDir)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not a git repository")
}

// TestGitOperations_ExecuteGitPull_Timeout tests command timeout
func TestGitOperations_ExecuteGitPull_Timeout(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	// Arrange
	tempDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())

	ops := NewOperations()
	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(10 * time.Millisecond)

	// Act
	result, err := ops.ExecuteGitPull(ctx, tempDir)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestGitOperations_ExecuteGitPull_InvalidPath tests invalid directory path
func TestGitOperations_ExecuteGitPull_InvalidPath(t *testing.T) {
	// Arrange
	ops := NewOperations()
	ctx := context.Background()
	invalidPath := "/path/that/does/not/exist"

	// Act
	result, err := ops.ExecuteGitPull(ctx, invalidPath)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestGitOperations_ValidatePath_WithinAllowedPath tests path validation
func TestGitOperations_ValidatePath_WithinAllowedPath(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	ops := NewOperations()

	// Act
	valid := ops.ValidatePath(subDir, tempDir)

	// Assert
	assert.True(t, valid, "Subdirectory should be valid")
}

// TestGitOperations_ValidatePath_OutsideAllowedPath tests path validation rejection
func TestGitOperations_ValidatePath_OutsideAllowedPath(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	outsideDir := t.TempDir()

	ops := NewOperations()

	// Act
	valid := ops.ValidatePath(outsideDir, tempDir)

	// Assert
	assert.False(t, valid, "Directory outside allowed path should be invalid")
}

// TestGitOperations_ValidatePath_TraversalAttempt tests path traversal prevention
func TestGitOperations_ValidatePath_TraversalAttempt(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	ops := NewOperations()
	traversalPath := filepath.Join(tempDir, "..", "..", "etc", "passwd")

	// Act
	valid := ops.ValidatePath(traversalPath, tempDir)

	// Assert
	assert.False(t, valid, "Path traversal attempt should be invalid")
}
