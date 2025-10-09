package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitPullResult represents the result of a git pull operation
type GitPullResult struct {
	Success    bool
	Branch     string
	Changes    int
	LastCommit string
	Output     string
}

// Operations defines the interface for git operations
type Operations interface {
	IsGitRepo(dir string) bool
	ExecuteGitPull(ctx context.Context, dir string) (*GitPullResult, error)
	ValidatePath(path, allowedBase string) bool
}

// operations implements the Operations interface
type operations struct{}

// NewOperations creates a new git operations instance
func NewOperations() Operations {
	return &operations{}
}

// IsGitRepo checks if the given directory is a git repository
func (o *operations) IsGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ExecuteGitPull executes git pull in the specified directory
func (o *operations) ExecuteGitPull(ctx context.Context, dir string) (*GitPullResult, error) {
	// Validate directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	// Check if it's a git repository
	if !o.IsGitRepo(dir) {
		return nil, errors.New("not a git repository")
	}

	// Execute git pull command
	cmd := exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = dir
	
	// Clean environment for security
	cmd.Env = []string{
		"GIT_TERMINAL_PROMPT=0", // Disable prompts
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check for context cancellation (timeout)
		if ctx.Err() != nil {
			return nil, fmt.Errorf("git pull timeout: %w", ctx.Err())
		}
		return nil, fmt.Errorf("git pull failed: %w", err)
	}

	// Parse output to extract information
	result := &GitPullResult{
		Success: true,
		Output:  string(output),
	}

	// Try to get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = dir
	if branchOutput, err := branchCmd.Output(); err == nil {
		result.Branch = strings.TrimSpace(string(branchOutput))
	}

	// Try to get last commit
	commitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	commitCmd.Dir = dir
	if commitOutput, err := commitCmd.Output(); err == nil {
		result.LastCommit = strings.TrimSpace(string(commitOutput))
	}

	// Count changes (simplified - just check if output mentions files changed)
	if strings.Contains(result.Output, "file") {
		result.Changes = 1 // Simplified - just indicate changes occurred
	}

	return result, nil
}

// ValidatePath validates that a path is within the allowed base directory
func (o *operations) ValidatePath(path, allowedBase string) bool {
	// Clean paths to resolve .. and .
	cleanPath := filepath.Clean(path)
	cleanBase := filepath.Clean(allowedBase)

	// Check if path starts with allowed base
	relPath, err := filepath.Rel(cleanBase, cleanPath)
	if err != nil {
		return false
	}

	// Ensure relative path doesn't start with .. (going outside base)
	return !strings.HasPrefix(relPath, "..")
}
