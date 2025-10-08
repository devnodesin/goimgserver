package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper to create test directory structure
func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	
	// Create some test files for single images
	createTestFile(t, tmpDir, "cat.jpg")
	createTestFile(t, tmpDir, "dog.png")
	createTestFile(t, tmpDir, "logo.webp")
	
	// Create files with multiple extensions (for priority testing)
	createTestFile(t, tmpDir, "profile.jpg")
	createTestFile(t, tmpDir, "profile.png")
	createTestFile(t, tmpDir, "profile.webp")
	
	// Create grouped images
	createTestFile(t, tmpDir, "cats/default.jpg")
	createTestFile(t, tmpDir, "cats/cat_white.jpg")
	createTestFile(t, tmpDir, "cats/cat_white.png")
	createTestFile(t, tmpDir, "cats/funny_white.png")
	
	createTestFile(t, tmpDir, "dogs/default.png")
	createTestFile(t, tmpDir, "dogs/puppy.jpg")
	
	// Create system default
	createTestFile(t, tmpDir, "default.jpg")
	
	return tmpDir
}

func createTestFile(t *testing.T, baseDir, relPath string) {
	t.Helper()
	fullPath := filepath.Join(baseDir, relPath)
	dir := filepath.Dir(fullPath)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	
	if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", fullPath, err)
	}
}

// TestFileResolver_New_ValidInstance tests resolver creation
func TestFileResolver_New_ValidInstance(t *testing.T) {
	tmpDir := setupTestDir(t)
	
	resolver := NewResolver(tmpDir)
	
	assert.NotNil(t, resolver, "Resolver should not be nil")
}

// TestFileResolver_ResolveSingle_WithExtension tests direct file resolution
func TestFileResolver_ResolveSingle_WithExtension(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	tests := []struct {
		name         string
		requestPath  string
		expectedFile string
	}{
		{
			name:         "resolve cat.jpg",
			requestPath:  "cat.jpg",
			expectedFile: "cat.jpg",
		},
		{
			name:         "resolve dog.png",
			requestPath:  "dog.png",
			expectedFile: "dog.png",
		},
		{
			name:         "resolve logo.webp",
			requestPath:  "logo.webp",
			expectedFile: "logo.webp",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.requestPath)
			
			require.NoError(t, err, "Should resolve without error")
			require.NotNil(t, result, "Result should not be nil")
			
			expectedPath := filepath.Join(tmpDir, tt.expectedFile)
			assert.Equal(t, expectedPath, result.ResolvedPath, "Should resolve to correct path")
			assert.False(t, result.IsFallback, "Should not be a fallback")
		})
	}
}

// TestFileResolver_ResolveSingle_WithoutExtension tests extension auto-detection
func TestFileResolver_ResolveSingle_WithoutExtension(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	tests := []struct {
		name         string
		requestPath  string
		expectedFile string
	}{
		{
			name:         "auto-detect cat.jpg",
			requestPath:  "cat",
			expectedFile: "cat.jpg",
		},
		{
			name:         "auto-detect dog.png",
			requestPath:  "dog",
			expectedFile: "dog.png",
		},
		{
			name:         "auto-detect logo.webp",
			requestPath:  "logo",
			expectedFile: "logo.webp",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.requestPath)
			
			require.NoError(t, err, "Should resolve without error")
			require.NotNil(t, result, "Result should not be nil")
			
			expectedPath := filepath.Join(tmpDir, tt.expectedFile)
			assert.Equal(t, expectedPath, result.ResolvedPath, "Should resolve to correct path")
			assert.False(t, result.IsFallback, "Should not be a fallback")
		})
	}
}

// TestFileResolver_ResolveSingle_ExtensionPriority tests jpg > png > webp priority
func TestFileResolver_ResolveSingle_ExtensionPriority(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	// Test with profile which has all three extensions
	result, err := resolver.Resolve("profile")
	
	require.NoError(t, err, "Should resolve without error")
	require.NotNil(t, result, "Result should not be nil")
	
	expectedPath := filepath.Join(tmpDir, "profile.jpg")
	assert.Equal(t, expectedPath, result.ResolvedPath, "Should prioritize .jpg over .png and .webp")
	assert.False(t, result.IsFallback, "Should not be a fallback")
}

// TestFileResolver_ResolveSingle_MultipleExtensions tests priority with multiple files
func TestFileResolver_ResolveSingle_MultipleExtensions(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name          string
		createFiles   []string
		requestPath   string
		expectedFile  string
	}{
		{
			name:         "jpg wins over png and webp",
			createFiles:  []string{"test.jpg", "test.png", "test.webp"},
			requestPath:  "test",
			expectedFile: "test.jpg",
		},
		{
			name:         "png wins over webp when jpg missing",
			createFiles:  []string{"test.png", "test.webp"},
			requestPath:  "test",
			expectedFile: "test.png",
		},
		{
			name:         "webp when others missing",
			createFiles:  []string{"test.webp"},
			requestPath:  "test",
			expectedFile: "test.webp",
		},
		{
			name:         "jpeg same priority as jpg",
			createFiles:  []string{"test.jpeg", "test.png"},
			requestPath:  "test",
			expectedFile: "test.jpeg",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			require.NoError(t, os.MkdirAll(testDir, 0755))
			
			for _, file := range tt.createFiles {
				createTestFile(t, testDir, file)
			}
			
			resolver := NewResolver(testDir)
			result, err := resolver.Resolve(tt.requestPath)
			
			require.NoError(t, err, "Should resolve without error")
			require.NotNil(t, result, "Result should not be nil")
			
			expectedPath := filepath.Join(testDir, tt.expectedFile)
			assert.Equal(t, expectedPath, result.ResolvedPath, "Should resolve to correct priority file")
		})
	}
}

// TestFileResolver_ResolveGrouped_DefaultImage tests group default resolution
func TestFileResolver_ResolveGrouped_DefaultImage(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	tests := []struct {
		name         string
		requestPath  string
		expectedFile string
	}{
		{
			name:         "resolve cats group default",
			requestPath:  "cats",
			expectedFile: "cats/default.jpg",
		},
		{
			name:         "resolve dogs group default",
			requestPath:  "dogs",
			expectedFile: "dogs/default.png",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.requestPath)
			
			require.NoError(t, err, "Should resolve without error")
			require.NotNil(t, result, "Result should not be nil")
			
			expectedPath := filepath.Join(tmpDir, tt.expectedFile)
			assert.Equal(t, expectedPath, result.ResolvedPath, "Should resolve to group default")
			assert.True(t, result.IsGrouped, "Should be marked as grouped")
			assert.False(t, result.IsFallback, "Group default is not a fallback")
		})
	}
}

// TestFileResolver_ResolveGrouped_SpecificImage tests specific grouped image resolution
func TestFileResolver_ResolveGrouped_SpecificImage(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	tests := []struct {
		name         string
		requestPath  string
		expectedFile string
	}{
		{
			name:         "resolve specific cat image",
			requestPath:  "cats/cat_white.jpg",
			expectedFile: "cats/cat_white.jpg",
		},
		{
			name:         "resolve specific dog image",
			requestPath:  "dogs/puppy.jpg",
			expectedFile: "dogs/puppy.jpg",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.requestPath)
			
			require.NoError(t, err, "Should resolve without error")
			require.NotNil(t, result, "Result should not be nil")
			
			expectedPath := filepath.Join(tmpDir, tt.expectedFile)
			assert.Equal(t, expectedPath, result.ResolvedPath, "Should resolve to specific grouped image")
			assert.True(t, result.IsGrouped, "Should be marked as grouped")
			assert.False(t, result.IsFallback, "Direct file is not a fallback")
		})
	}
}

// TestFileResolver_ResolveGrouped_WithExtension tests grouped image with extension
func TestFileResolver_ResolveGrouped_WithExtension(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	result, err := resolver.Resolve("cats/funny_white.png")
	
	require.NoError(t, err, "Should resolve without error")
	require.NotNil(t, result, "Result should not be nil")
	
	expectedPath := filepath.Join(tmpDir, "cats/funny_white.png")
	assert.Equal(t, expectedPath, result.ResolvedPath, "Should resolve grouped image with extension")
	assert.True(t, result.IsGrouped, "Should be marked as grouped")
}

// TestFileResolver_ResolveGrouped_WithoutExtension tests grouped auto-detection
func TestFileResolver_ResolveGrouped_WithoutExtension(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	// cat_white has both .jpg and .png, should prefer .jpg
	result, err := resolver.Resolve("cats/cat_white")
	
	require.NoError(t, err, "Should resolve without error")
	require.NotNil(t, result, "Result should not be nil")
	
	expectedPath := filepath.Join(tmpDir, "cats/cat_white.jpg")
	assert.Equal(t, expectedPath, result.ResolvedPath, "Should auto-detect with priority in grouped images")
	assert.True(t, result.IsGrouped, "Should be marked as grouped")
}

// TestFileResolver_ResolveGrouped_MissingFallback tests fallback to group default
func TestFileResolver_ResolveGrouped_MissingFallback(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	// Request a non-existent file in cats group
	result, err := resolver.Resolve("cats/missing_cat")
	
	require.NoError(t, err, "Should resolve without error (fallback)")
	require.NotNil(t, result, "Result should not be nil")
	
	expectedPath := filepath.Join(tmpDir, "cats/default.jpg")
	assert.Equal(t, expectedPath, result.ResolvedPath, "Should fallback to group default")
	assert.True(t, result.IsGrouped, "Should be marked as grouped")
	assert.True(t, result.IsFallback, "Should be marked as fallback")
	assert.Equal(t, "group_default", result.FallbackType, "Should indicate group default fallback")
}

// TestFileResolver_ResolveWithDefault tests custom default resolution
func TestFileResolver_ResolveWithDefault(t *testing.T) {
	tmpDir := setupTestDir(t)
	resolver := NewResolver(tmpDir)
	
	customDefault := filepath.Join(tmpDir, "dog.png")
	
	tests := []struct {
		name        string
		requestPath string
		defaultPath string
		expectPath  string
		expectError bool
	}{
		{
			name:        "existing file ignores default",
			requestPath: "cat.jpg",
			defaultPath: customDefault,
			expectPath:  filepath.Join(tmpDir, "cat.jpg"),
			expectError: false,
		},
		{
			name:        "missing file uses default",
			requestPath: "missing.jpg",
			defaultPath: customDefault,
			expectPath:  customDefault,
			expectError: false,
		},
		{
			name:        "missing file with missing default",
			requestPath: "missing.jpg",
			defaultPath: filepath.Join(tmpDir, "nonexistent.jpg"),
			expectPath:  "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveWithDefault(tt.requestPath, tt.defaultPath)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectPath, result.ResolvedPath)
			}
		})
	}
}
