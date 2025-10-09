package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileResolver_PathTraversal_Prevention tests security path validation
func TestFileResolver_PathTraversal_Prevention(t *testing.T) {
	tmpDir := setupTestDir(t)
	
	// Create a file outside the image directory
	outsideDir := filepath.Dir(tmpDir)
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	require.NoError(t, os.WriteFile(outsideFile, []byte("secret data"), 0644))
	t.Cleanup(func() { os.Remove(outsideFile) })
	
	resolver := NewResolver(tmpDir)
	
	tests := []struct {
		name        string
		requestPath string
	}{
		{
			name:        "double dot traversal",
			requestPath: "../secret.txt",
		},
		{
			name:        "multiple traversals",
			requestPath: "../../etc/passwd",
		},
		{
			name:        "mixed path with traversal",
			requestPath: "cats/../../secret.txt",
		},
		{
			name:        "encoded traversal",
			requestPath: "..%2Fsecret.txt",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.requestPath)
			
			// Should either return error or fallback to default (not the traversed path)
			if err == nil {
				require.NotNil(t, result)
				// Ensure resolved path is within image directory
				absImageDir, _ := filepath.Abs(tmpDir)
				absResolved, _ := filepath.Abs(result.ResolvedPath)
				
				assert.True(t, 
					filepath.HasPrefix(absResolved, absImageDir),
					"Resolved path should stay within image directory",
				)
				
				// Should not resolve to the secret file
				assert.NotContains(t, result.ResolvedPath, "secret.txt")
			}
		})
	}
}

// TestFileResolver_SymlinkHandling_Safe tests safe symlink following
func TestFileResolver_SymlinkHandling_Safe(t *testing.T) {
	tmpDir := setupTestDir(t)
	
	// Create a symlink pointing to a file within the image directory
	targetFile := filepath.Join(tmpDir, "cat.jpg")
	symlinkFile := filepath.Join(tmpDir, "cat_link.jpg")
	
	err := os.Symlink(targetFile, symlinkFile)
	if err != nil {
		t.Skipf("Symlink creation failed (may not be supported): %v", err)
	}
	
	resolver := NewResolver(tmpDir)
	result, err := resolver.Resolve("cat_link.jpg")
	
	require.NoError(t, err, "Should resolve symlink within image directory")
	require.NotNil(t, result)
	
	// Verify the resolved path points to the actual file
	realPath, _ := filepath.EvalSymlinks(result.ResolvedPath)
	expectedPath := targetFile
	
	assert.Equal(t, expectedPath, realPath, "Should resolve symlink to target")
}

// TestFileResolver_SymlinkHandling_Dangerous tests dangerous symlink rejection
func TestFileResolver_SymlinkHandling_Dangerous(t *testing.T) {
	tmpDir := setupTestDir(t)
	
	// Create a file outside the image directory
	outsideDir := filepath.Dir(tmpDir)
	outsideFile := filepath.Join(outsideDir, "outside_secret.txt")
	require.NoError(t, os.WriteFile(outsideFile, []byte("secret data"), 0644))
	t.Cleanup(func() { os.Remove(outsideFile) })
	
	// Create a symlink pointing outside the image directory
	symlinkFile := filepath.Join(tmpDir, "dangerous_link.txt")
	err := os.Symlink(outsideFile, symlinkFile)
	if err != nil {
		t.Skipf("Symlink creation failed (may not be supported): %v", err)
	}
	
	resolver := NewResolver(tmpDir)
	result, err := resolver.Resolve("dangerous_link.txt")
	
	// Should either error or fallback, but not resolve to outside file
	if err == nil {
		require.NotNil(t, result)
		
		// Verify the resolved path doesn't escape the image directory
		absImageDir, _ := filepath.Abs(tmpDir)
		realPath, _ := filepath.EvalSymlinks(result.ResolvedPath)
		absResolved, _ := filepath.Abs(realPath)
		
		assert.True(t,
			filepath.HasPrefix(absResolved, absImageDir),
			"Resolved symlink should stay within image directory",
		)
		
		// Should not resolve to the outside file
		assert.NotEqual(t, outsideFile, realPath)
	}
}

// TestFileResolver_Security_DirectoryEscape tests directory escape prevention
func TestFileResolver_Security_DirectoryEscape(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	tests := []struct {
		name        string
		requestPath string
	}{
		{
			name:        "absolute path attempt",
			requestPath: "/etc/passwd",
		},
		{
			name:        "windows absolute path",
			requestPath: "C:\\Windows\\System32\\config",
		},
		{
			name:        "url encoded double dot",
			requestPath: "cats%2F..%2F..%2Fsecret.txt",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.requestPath)
			
			if err == nil && result != nil {
				absImageDir, _ := filepath.Abs(tmpDir)
				absResolved, _ := filepath.Abs(result.ResolvedPath)
				
				assert.True(t,
					filepath.HasPrefix(absResolved, absImageDir),
					"Should not escape image directory",
				)
			}
		})
	}
}

// TestFileResolver_Security_NullByteInjection tests null byte injection prevention
func TestFileResolver_Security_NullByteInjection(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	tests := []struct {
		name        string
		requestPath string
	}{
		{
			name:        "null byte in path",
			requestPath: "cat\x00.jpg",
		},
		{
			name:        "null byte with extension",
			requestPath: "secret\x00.txt.jpg",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.requestPath)
			
			// Should handle gracefully - either error or fallback
			if err == nil && result != nil {
				// Verify no null bytes in resolved path
				assert.NotContains(t, result.ResolvedPath, "\x00")
			}
		})
	}
}

// TestFileResolver_EdgeCases tests various edge cases
func TestFileResolver_EdgeCases(t *testing.T) {
	tmpDir := setupTestDir(t)
	
	tests := []struct {
		name          string
		setupFunc     func(string)
		requestPath   string
		expectFallback bool
	}{
		{
			name: "empty request path",
			requestPath: "",
			expectFallback: true,
		},
		{
			name: "directory without default",
			setupFunc: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "emptygroup"), 0755)
			},
			requestPath: "emptygroup",
			expectFallback: true,
		},
		{
			name: "file in subdirectory",
			setupFunc: func(dir string) {
				createTestFile(t, dir, "deep/nested/image.jpg")
			},
			requestPath: "deep/nested/image.jpg",
			expectFallback: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			require.NoError(t, os.MkdirAll(testDir, 0755))
			createTestFile(t, testDir, "default.jpg")
			
			if tt.setupFunc != nil {
				tt.setupFunc(testDir)
			}
			
			resolver := NewResolver(testDir)
			result, err := resolver.Resolve(tt.requestPath)
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			if tt.expectFallback {
				assert.True(t, result.IsFallback, "Should be a fallback")
			} else {
				assert.False(t, result.IsFallback, "Should not be a fallback")
			}
		})
	}
}

// TestFileResolver_MissingSystemDefault tests behavior when system default is missing
func TestFileResolver_MissingSystemDefault(t *testing.T) {
	tmpDir := t.TempDir()
	// No default image created
	
	resolver := NewResolver(tmpDir)
	result, err := resolver.Resolve("nonexistent.jpg")
	
	// Should return error when no default image exists
	assert.Error(t, err)
	assert.Equal(t, ErrFileNotFound, err)
	assert.Nil(t, result)
}
