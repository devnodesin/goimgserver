package handlers

import (
	"context"
	"fmt"
	"goimgserver/cache"
	"goimgserver/config"
	"goimgserver/git"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GitOperations defines the interface for git operations
type GitOperations interface {
	IsGitRepo(dir string) bool
	ExecuteGitPull(ctx context.Context, dir string) (*git.GitPullResult, error)
	ValidatePath(path, allowedBase string) bool
}

// CommandHandler handles administrative command endpoints
type CommandHandler struct {
	config       *config.Config
	cacheManager cache.CacheManager
	gitOps       GitOperations
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(cfg *config.Config, cacheManager cache.CacheManager, gitOps GitOperations) *CommandHandler {
	return &CommandHandler{
		config:       cfg,
		cacheManager: cacheManager,
		gitOps:       gitOps,
	}
}

// HandleClear handles the /cmd/clear endpoint
func (h *CommandHandler) HandleClear(c *gin.Context) {
	// Count files before clearing
	stats, err := h.cacheManager.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get cache stats",
		})
		return
	}

	clearedFiles := stats.TotalFiles
	freedSpace := stats.TotalSize

	// Clear the cache
	if err := h.cacheManager.ClearAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to clear cache",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Cache cleared successfully",
		"cleared_files": clearedFiles,
		"freed_space":   formatBytes(freedSpace),
	})
}

// HandleGitUpdate handles the /cmd/gitupdate endpoint
func (h *CommandHandler) HandleGitUpdate(c *gin.Context) {
	imagesDir := h.config.ImagesDir

	// Validate the path doesn't contain malicious characters
	if strings.Contains(imagesDir, ";") || strings.Contains(imagesDir, "&") || 
	   strings.Contains(imagesDir, "|") || strings.Contains(imagesDir, "`") {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid directory path",
			"code":    "INVALID_PATH",
		})
		return
	}

	// Check if it's a git repository
	if !h.gitOps.IsGitRepo(imagesDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Images directory is not a git repository",
			"code":    "GIT_NOT_FOUND",
		})
		return
	}

	// Execute git pull with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.gitOps.ExecuteGitPull(ctx, imagesDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "git update failed: " + err.Error(),
			"code":    "GIT_UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Git update completed",
		"changes":     result.Changes,
		"branch":      result.Branch,
		"last_commit": result.LastCommit,
	})
}

// HandleCommand handles the generic /cmd/{name} endpoint
func (h *CommandHandler) HandleCommand(c *gin.Context) {
	commandName := c.Param("name")

	// Validate command name
	if !h.validateCommand(commandName) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid command: " + commandName,
			"code":    "INVALID_COMMAND",
		})
		return
	}

	// Route to specific command handler
	switch commandName {
	case "clear":
		h.HandleClear(c)
	case "gitupdate":
		h.HandleGitUpdate(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid command: " + commandName,
			"code":    "INVALID_COMMAND",
		})
	}
}

// validateCommand checks if a command name is allowed
func (h *CommandHandler) validateCommand(command string) bool {
	allowedCommands := map[string]bool{
		"clear":     true,
		"gitupdate": true,
	}
	return allowedCommands[command]
}

// formatBytes formats bytes to human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// countFiles counts the number of files in a directory recursively
func countFiles(dir string) (int64, error) {
	var count int64
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}
