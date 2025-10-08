package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Test default image detection in image directory
func Test_DefaultImage_Detection(t *testing.T) {
	tests := []struct {
		name          string
		createFiles   []string
		expectedFound bool
		expectedExt   string
	}{
		{
			name:          "DetectDefaultJPG",
			createFiles:   []string{"default.jpg"},
			expectedFound: true,
			expectedExt:   ".jpg",
		},
		{
			name:          "DetectDefaultJPEG",
			createFiles:   []string{"default.jpeg"},
			expectedFound: true,
			expectedExt:   ".jpeg",
		},
		{
			name:          "DetectDefaultPNG",
			createFiles:   []string{"default.png"},
			expectedFound: true,
			expectedExt:   ".png",
		},
		{
			name:          "DetectDefaultWebP",
			createFiles:   []string{"default.webp"},
			expectedFound: true,
			expectedExt:   ".webp",
		},
		{
			name:          "PriorityJPGOverJPEG",
			createFiles:   []string{"default.jpg", "default.jpeg"},
			expectedFound: true,
			expectedExt:   ".jpg",
		},
		{
			name:          "PriorityJPEGOverPNG",
			createFiles:   []string{"default.jpeg", "default.png"},
			expectedFound: true,
			expectedExt:   ".jpeg",
		},
		{
			name:          "NoDefaultImage",
			createFiles:   []string{"other.jpg"},
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpDir := t.TempDir()
			
			// Create test files
			for _, filename := range tt.createFiles {
				filePath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Act
			defaultImgPath, found := DetectDefaultImage(tmpDir)

			// Assert
			if found != tt.expectedFound {
				t.Errorf("Expected found=%v, got %v", tt.expectedFound, found)
			}

			if tt.expectedFound {
				if defaultImgPath == "" {
					t.Error("Expected non-empty path when found=true")
				}
				ext := filepath.Ext(defaultImgPath)
				if ext != tt.expectedExt {
					t.Errorf("Expected extension %s, got %s", tt.expectedExt, ext)
				}
			}
		})
	}
}

// Test programmatic placeholder generation
func Test_DefaultImage_GeneratePlaceholder(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "default.jpg")

	// Act
	err := GenerateDefaultPlaceholder(outputPath)

	// Assert
	if err != nil {
		t.Fatalf("GenerateDefaultPlaceholder() returned error: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Generated placeholder file should exist: %v", err)
	}

	// Verify file has reasonable size (not empty, not too small)
	if info.Size() < 100 {
		t.Errorf("Generated placeholder is too small: %d bytes", info.Size())
	}
}

// Test default image readability validation
func Test_DefaultImage_ValidationReadable(t *testing.T) {
	t.Run("ValidImageFile", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		imgPath := filepath.Join(tmpDir, "default.jpg")
		
		// Generate a valid placeholder
		if err := GenerateDefaultPlaceholder(imgPath); err != nil {
			t.Fatalf("Failed to create test image: %v", err)
		}

		// Act
		err := ValidateDefaultImage(imgPath)

		// Assert
		if err != nil {
			t.Errorf("ValidateDefaultImage() should succeed for valid image: %v", err)
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		// Arrange
		imgPath := "/nonexistent/default.jpg"

		// Act
		err := ValidateDefaultImage(imgPath)

		// Assert
		if err == nil {
			t.Error("ValidateDefaultImage() should fail for non-existent file")
		}
	})

	t.Run("EmptyFile", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		imgPath := filepath.Join(tmpDir, "empty.jpg")
		if err := os.WriteFile(imgPath, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		// Act
		err := ValidateDefaultImage(imgPath)

		// Assert
		if err == nil {
			t.Error("ValidateDefaultImage() should fail for empty file")
		}
	})
}

// Test default image setup integration
func Test_DefaultImage_Setup(t *testing.T) {
	t.Run("ExistingDefaultImage", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		existingDefault := filepath.Join(tmpDir, "default.jpg")
		
		// Create a valid JPEG image using GenerateDefaultPlaceholder
		if err := GenerateDefaultPlaceholder(existingDefault); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cfg := &Config{
			Port:      9000,
			ImagesDir: tmpDir,
			CacheDir:  filepath.Join(tmpDir, "cache"),
		}

		// Act
		err := cfg.SetupDefaultImage()

		// Assert
		if err != nil {
			t.Fatalf("SetupDefaultImage() returned error: %v", err)
		}

		if cfg.DefaultImagePath == "" {
			t.Error("DefaultImagePath should be set")
		}

		if cfg.DefaultImagePath != existingDefault {
			t.Errorf("Expected path %s, got %s", existingDefault, cfg.DefaultImagePath)
		}
	})

	t.Run("GenerateWhenMissing", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		expectedPath := filepath.Join(tmpDir, "default.jpg")

		cfg := &Config{
			Port:      9000,
			ImagesDir: tmpDir,
			CacheDir:  filepath.Join(tmpDir, "cache"),
		}

		// Act
		err := cfg.SetupDefaultImage()

		// Assert
		if err != nil {
			t.Fatalf("SetupDefaultImage() returned error: %v", err)
		}

		if cfg.DefaultImagePath == "" {
			t.Error("DefaultImagePath should be set")
		}

		// Verify file was generated
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Error("Default image should be generated at expected path")
		}
	})
}

// Test SetupDefaultImage with unwritable directory
func Test_DefaultImage_Setup_UnwritableDir(t *testing.T) {
	t.Run("CannotGeneratePlaceholder", func(t *testing.T) {
		// Skip on systems where we can't reliably test this
		if os.Getuid() == 0 {
			t.Skip("Skipping test when running as root")
		}

		// Arrange
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
			t.Fatalf("Failed to create readonly dir: %v", err)
		}

		cfg := &Config{
			Port:      9000,
			ImagesDir: readOnlyDir,
			CacheDir:  filepath.Join(tmpDir, "cache"),
		}

		// Act
		err := cfg.SetupDefaultImage()

		// Assert
		if err == nil {
			t.Error("SetupDefaultImage() should fail when directory is not writable")
		}
	})
}

// Test GenerateDefaultPlaceholder with invalid output path
func Test_DefaultImage_GeneratePlaceholder_InvalidPath(t *testing.T) {
	// Arrange - use invalid path
	outputPath := "/dev/null/impossible/default.jpg"

	// Act
	err := GenerateDefaultPlaceholder(outputPath)

	// Assert
	if err == nil {
		t.Error("GenerateDefaultPlaceholder() should fail for invalid path")
	}
}

// Test ValidateDefaultImage with corrupted image
func Test_DefaultImage_ValidationReadable_CorruptedImage(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "corrupted.jpg")
	
	// Create a file with invalid JPEG data
	if err := os.WriteFile(imgPath, []byte("not a real image file"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Act
	err := ValidateDefaultImage(imgPath)

	// Assert
	if err == nil {
		t.Error("ValidateDefaultImage() should fail for corrupted image")
	}
}

// Test SetupDefaultImage when found image fails validation
func Test_DefaultImage_Setup_FoundButInvalid(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	corruptedDefault := filepath.Join(tmpDir, "default.jpg")
	
	// Create corrupted image file
	if err := os.WriteFile(corruptedDefault, []byte("corrupted"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &Config{
		Port:      9000,
		ImagesDir: tmpDir,
		CacheDir:  filepath.Join(tmpDir, "cache"),
	}

	// Act
	err := cfg.SetupDefaultImage()

	// Assert
	if err == nil {
		t.Error("SetupDefaultImage() should fail when found image is invalid")
	}
}
