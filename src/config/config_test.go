package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test default values when no arguments are provided
func Test_ParseArgs_DefaultValues(t *testing.T) {
	// Arrange
	args := []string{}

	// Act
	cfg, err := ParseArgs(args)

	// Assert
	if err != nil {
		t.Fatalf("ParseArgs() returned error: %v", err)
	}
	if cfg.Port != 9000 {
		t.Errorf("Expected default port 9000, got %d", cfg.Port)
	}
	if cfg.ImagesDir != "./images" {
		t.Errorf("Expected default imagesdir './images', got %s", cfg.ImagesDir)
	}
	if cfg.CacheDir != "./cache" {
		t.Errorf("Expected default cachedir './cache', got %s", cfg.CacheDir)
	}
	if cfg.Dump {
		t.Error("Expected dump to be false by default")
	}
}

// Test custom port argument parsing
func Test_ParseArgs_CustomPort(t *testing.T) {
	// Arrange
	args := []string{"--port", "8080"}

	// Act
	cfg, err := ParseArgs(args)

	// Assert
	if err != nil {
		t.Fatalf("ParseArgs() returned error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}
}

// Test custom directory path parsing
func Test_ParseArgs_CustomDirectories(t *testing.T) {
	// Arrange
	args := []string{"--imagesdir", "/custom/images", "--cachedir", "/custom/cache"}

	// Act
	cfg, err := ParseArgs(args)

	// Assert
	if err != nil {
		t.Fatalf("ParseArgs() returned error: %v", err)
	}
	if cfg.ImagesDir != "/custom/images" {
		t.Errorf("Expected imagesdir '/custom/images', got %s", cfg.ImagesDir)
	}
	if cfg.CacheDir != "/custom/cache" {
		t.Errorf("Expected cachedir '/custom/cache', got %s", cfg.CacheDir)
	}
}

// Test dump flag functionality
func Test_ParseArgs_DumpFlag(t *testing.T) {
	// Arrange
	args := []string{"--dump"}

	// Act
	cfg, err := ParseArgs(args)

	// Assert
	if err != nil {
		t.Fatalf("ParseArgs() returned error: %v", err)
	}
	if !cfg.Dump {
		t.Error("Expected dump to be true")
	}
}

// Test port validation with valid range
func Test_ValidatePort_ValidRange(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"MinValidPort", 1},
		{"MidRangePort", 8080},
		{"DefaultPort", 9000},
		{"MaxValidPort", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpDir := t.TempDir()
			cfg := &Config{
				Port:      tt.port,
				ImagesDir: filepath.Join(tmpDir, "images"),
				CacheDir:  filepath.Join(tmpDir, "cache"),
			}

			// Act
			err := cfg.Validate()

			// Assert
			if err != nil {
				t.Errorf("Port %d should be valid, got error: %v", tt.port, err)
			}
		})
	}
}

// Test invalid port rejection
func Test_ValidatePort_InvalidRange(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"ZeroPort", 0},
		{"NegativePort", -1},
		{"TooLargePort", 65536},
		{"VeryLargePort", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := &Config{Port: tt.port, ImagesDir: "/tmp", CacheDir: "/tmp"}

			// Act
			err := cfg.Validate()

			// Assert
			if err == nil {
				t.Errorf("Port %d should be invalid, but no error was returned", tt.port)
			}
		})
	}
}

// Test existing directory validation
func Test_ValidateDirectories_ExistingDirs(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	imgDir := filepath.Join(tmpDir, "images")
	cacheDir := filepath.Join(tmpDir, "cache")
	
	// Create directories
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	cfg := &Config{
		Port:      9000,
		ImagesDir: imgDir,
		CacheDir:  cacheDir,
	}

	// Act
	err := cfg.Validate()

	// Assert
	if err != nil {
		t.Errorf("Validation should succeed for existing directories, got error: %v", err)
	}

	// Verify directories still exist
	if _, err := os.Stat(imgDir); os.IsNotExist(err) {
		t.Error("ImagesDir should exist after validation")
	}
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("CacheDir should exist after validation")
	}
}

// Test directory creation for missing directories
func Test_ValidateDirectories_CreateMissing(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	imgDir := filepath.Join(tmpDir, "new_images")
	cacheDir := filepath.Join(tmpDir, "new_cache")

	cfg := &Config{
		Port:      9000,
		ImagesDir: imgDir,
		CacheDir:  cacheDir,
	}

	// Act
	err := cfg.Validate()

	// Assert
	if err != nil {
		t.Fatalf("Validation should create missing directories, got error: %v", err)
	}

	// Verify directories were created
	if _, err := os.Stat(imgDir); os.IsNotExist(err) {
		t.Error("ImagesDir should be created during validation")
	}
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("CacheDir should be created during validation")
	}
}

// Test settings dump functionality
func Test_DumpSettings_ValidOutput(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "settings.conf")

	cfg := &Config{
		Port:      9000,
		ImagesDir: "./images",
		CacheDir:  "./cache",
		Dump:      true,
	}

	// Act
	err := cfg.DumpSettings(outputFile)

	// Assert
	if err != nil {
		t.Fatalf("DumpSettings() returned error: %v", err)
	}

	// Verify file exists and contains expected content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("Settings file should not be empty")
	}
}

// Test configuration string representation
func Test_String_ProperFormat(t *testing.T) {
	// Arrange
	cfg := &Config{
		Port:      9000,
		ImagesDir: "./images",
		CacheDir:  "./cache",
		Dump:      false,
	}

	// Act
	result := cfg.String()

	// Assert
	if result == "" {
		t.Error("String() should return non-empty string")
	}
	// Should contain key configuration values
	if len(result) < 20 {
		t.Error("String() should contain meaningful configuration information")
	}
}

// Test String with DefaultImagePath set
func Test_String_WithDefaultImagePath(t *testing.T) {
	// Arrange
	cfg := &Config{
		Port:             9000,
		ImagesDir:        "./images",
		CacheDir:         "./cache",
		Dump:             false,
		DefaultImagePath: "/path/to/default.jpg",
	}

	// Act
	result := cfg.String()

	// Assert
	if result == "" {
		t.Error("String() should return non-empty string")
	}
	// Should contain default image path
	if !strings.Contains(result, "DefaultImagePath") {
		t.Error("String() should contain DefaultImagePath when set")
	}
}

// Test ParseArgs with invalid arguments
func Test_ParseArgs_InvalidArgs(t *testing.T) {
	// Arrange
	args := []string{"--port", "invalid"}

	// Act
	_, err := ParseArgs(args)

	// Assert
	if err == nil {
		t.Error("ParseArgs() should return error for invalid port value")
	}
}

// Test Validate with directory creation failure (impossible path on Unix)
func Test_ValidateDirectories_InvalidPath(t *testing.T) {
	// Arrange - use null device path which cannot be created as directory
	cfg := &Config{
		Port:      9000,
		ImagesDir: "/dev/null/impossible",
		CacheDir:  filepath.Join(t.TempDir(), "cache"),
	}

	// Act
	err := cfg.Validate()

	// Assert
	if err == nil {
		t.Error("Validate() should fail for impossible directory path")
	}
}
