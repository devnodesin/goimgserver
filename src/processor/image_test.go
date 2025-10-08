package processor

import (
	"os"
	"testing"
)

// Test processor creation
func TestImageProcessor_New_ValidInstance(t *testing.T) {
	processor := New()
	
	if processor == nil {
		t.Fatal("New() returned nil processor")
	}
}

// Test basic resizing with both dimensions
func TestImageProcessor_Resize_ValidDimensions(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	result, err := processor.Resize(data, 200, 150)
	
	if err != nil {
		t.Fatalf("Resize() failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("Resize() returned empty data")
	}
	
	// Verify the image is valid and has correct dimensions
	metadata, err := getImageMetadata(result)
	if err != nil {
		t.Fatalf("Failed to get metadata from resized image: %v", err)
	}
	if metadata.Width != 200 || metadata.Height != 150 {
		t.Errorf("Expected dimensions 200x150, got %dx%d", metadata.Width, metadata.Height)
	}
}

// Test single dimension resize (maintaining aspect ratio)
func TestImageProcessor_Resize_MaintainAspectRatio(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg") // 400x300
	
	tests := []struct {
		name           string
		width, height  int
		expectWidth    int
		expectHeight   int
	}{
		{"Width only", 200, 0, 200, 150},
		{"Height only", 0, 100, 133, 100},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.Resize(data, tt.width, tt.height)
			
			if err != nil {
				t.Fatalf("Resize() failed: %v", err)
			}
			
			metadata, err := getImageMetadata(result)
			if err != nil {
				t.Fatalf("Failed to get metadata: %v", err)
			}
			
			// Allow small rounding differences
			if abs(metadata.Width-tt.expectWidth) > 1 || abs(metadata.Height-tt.expectHeight) > 1 {
				t.Errorf("Expected dimensions ~%dx%d, got %dx%d", 
					tt.expectWidth, tt.expectHeight, metadata.Width, metadata.Height)
			}
		})
	}
}

// Test minimum dimension constraints (10px)
func TestImageProcessor_Resize_MinimumConstraints(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	tests := []struct {
		name          string
		width, height int
		shouldFail    bool
	}{
		{"Valid minimum", 10, 10, false},
		{"Below minimum width", 5, 100, true},
		{"Below minimum height", 100, 5, true},
		{"Both below minimum", 5, 5, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.Resize(data, tt.width, tt.height)
			
			if tt.shouldFail && err == nil {
				t.Error("Expected error for dimensions below minimum, got nil")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if tt.shouldFail && err != ErrInvalidDimensions {
				t.Errorf("Expected ErrInvalidDimensions, got: %v", err)
			}
		})
	}
}

// Test maximum dimension constraints (4000px)
func TestImageProcessor_Resize_MaximumConstraints(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	tests := []struct {
		name          string
		width, height int
		shouldFail    bool
	}{
		{"Valid maximum", 4000, 4000, false},
		{"Above maximum width", 4001, 1000, true},
		{"Above maximum height", 1000, 4001, true},
		{"Both above maximum", 5000, 5000, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.Resize(data, tt.width, tt.height)
			
			if tt.shouldFail && err == nil {
				t.Error("Expected error for dimensions above maximum, got nil")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if tt.shouldFail && err != ErrInvalidDimensions {
				t.Errorf("Expected ErrInvalidDimensions, got: %v", err)
			}
		})
	}
}

// Test WebP format conversion
func TestImageProcessor_ConvertFormat_WebP(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	result, err := processor.ConvertFormat(data, FormatWebP)
	
	if err != nil {
		t.Fatalf("ConvertFormat() failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("ConvertFormat() returned empty data")
	}
	
	// Verify the output is WebP
	metadata, err := getImageMetadata(result)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	if metadata.Type != "webp" {
		t.Errorf("Expected format webp, got %s", metadata.Type)
	}
}

// Test PNG format conversion
func TestImageProcessor_ConvertFormat_PNG(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	result, err := processor.ConvertFormat(data, FormatPNG)
	
	if err != nil {
		t.Fatalf("ConvertFormat() failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("ConvertFormat() returned empty data")
	}
	
	// Verify the output is PNG
	metadata, err := getImageMetadata(result)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	if metadata.Type != "png" {
		t.Errorf("Expected format png, got %s", metadata.Type)
	}
}

// Test JPEG format conversion
func TestImageProcessor_ConvertFormat_JPEG(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.png")
	
	result, err := processor.ConvertFormat(data, FormatJPEG)
	
	if err != nil {
		t.Fatalf("ConvertFormat() failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("ConvertFormat() returned empty data")
	}
	
	// Verify the output is JPEG
	metadata, err := getImageMetadata(result)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	if metadata.Type != "jpeg" {
		t.Errorf("Expected format jpeg, got %s", metadata.Type)
	}
}

// Test unsupported format error handling
func TestImageProcessor_ConvertFormat_UnsupportedFormat(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	_, err := processor.ConvertFormat(data, ImageFormat("bmp"))
	
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
	if err != ErrUnsupportedFormat {
		t.Errorf("Expected ErrUnsupportedFormat, got: %v", err)
	}
}

// Test valid quality range (1-100)
func TestImageProcessor_AdjustQuality_ValidRange(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	tests := []struct {
		name    string
		quality int
	}{
		{"Min quality", 1},
		{"Low quality", 25},
		{"Default quality", 75},
		{"High quality", 95},
		{"Max quality", 100},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.AdjustQuality(data, tt.quality)
			
			if err != nil {
				t.Errorf("AdjustQuality(%d) failed: %v", tt.quality, err)
			}
			if len(result) == 0 {
				t.Error("AdjustQuality() returned empty data")
			}
		})
	}
}

// Test invalid quality range
func TestImageProcessor_AdjustQuality_InvalidRange(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	tests := []struct {
		name    string
		quality int
	}{
		{"Zero quality", 0},
		{"Negative quality", -10},
		{"Above maximum", 101},
		{"Very high", 200},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.AdjustQuality(data, tt.quality)
			
			if err == nil {
				t.Error("Expected error for invalid quality, got nil")
			}
			if err != ErrInvalidQuality {
				t.Errorf("Expected ErrInvalidQuality, got: %v", err)
			}
		})
	}
}

// Test combined operations (resize + format + quality)
func TestImageProcessor_Process_CombinedOperations(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	opts := ProcessOptions{
		Width:   300,
		Height:  200,
		Format:  FormatWebP,
		Quality: 85,
	}
	
	result, err := processor.Process(data, opts)
	
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("Process() returned empty data")
	}
	
	// Verify all operations were applied
	metadata, err := getImageMetadata(result)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	if metadata.Width != 300 || metadata.Height != 200 {
		t.Errorf("Expected dimensions 300x200, got %dx%d", metadata.Width, metadata.Height)
	}
	if metadata.Type != "webp" {
		t.Errorf("Expected format webp, got %s", metadata.Type)
	}
}

// Test error handling for corrupted images
func TestImageProcessor_Process_CorruptedImage(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "corrupted.jpg")
	
	opts := ProcessOptions{
		Width:   200,
		Height:  150,
		Format:  FormatWebP,
		Quality: 75,
	}
	
	_, err := processor.Process(data, opts)
	
	if err == nil {
		t.Error("Expected error for corrupted image, got nil")
	}
	if err != ErrInvalidImage {
		t.Errorf("Expected ErrInvalidImage, got: %v", err)
	}
}

// Test unsupported input format validation
func TestImageProcessor_Process_UnsupportedInputFormat(t *testing.T) {
	processor := New()
	
	// Create fake data that looks like an unsupported format
	invalidData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header but corrupted
	invalidData = append(invalidData, make([]byte, 100)...)
	
	opts := ProcessOptions{
		Width:   200,
		Height:  150,
		Format:  FormatWebP,
		Quality: 75,
	}
	
	_, err := processor.Process(invalidData, opts)
	
	if err == nil {
		t.Error("Expected error for invalid input, got nil")
	}
}

// Test image header validation
func TestImageProcessor_Validate_ImageHeaders(t *testing.T) {
	processor := New()
	
	tests := []struct {
		name      string
		filename  string
		shouldErr bool
	}{
		{"Valid JPEG", "sample.jpg", false},
		{"Valid PNG", "sample.png", false},
		{"Valid WebP", "sample.webp", false},
		{"Corrupted", "corrupted.jpg", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := loadTestImage(t, tt.filename)
			err := processor.ValidateImage(data)
			
			if tt.shouldErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// Test empty data validation
func TestImageProcessor_Validate_EmptyData(t *testing.T) {
	processor := New()
	
	err := processor.ValidateImage([]byte{})
	
	if err != ErrInvalidImage {
		t.Errorf("Expected ErrInvalidImage for empty data, got: %v", err)
	}
}

// Test resize with both dimensions zero
func TestImageProcessor_Resize_BothZero(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	result, err := processor.Resize(data, 0, 0)
	
	if err != nil {
		t.Fatalf("Resize with 0,0 should not fail: %v", err)
	}
	if len(result) == 0 {
		t.Error("Resize returned empty data")
	}
}

// Test Process with default values
func TestImageProcessor_Process_DefaultValues(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	opts := ProcessOptions{
		Width:   DefaultWidth,
		Height:  DefaultHeight,
		Format:  FormatWebP,
		Quality: DefaultQuality,
	}
	
	result, err := processor.Process(data, opts)
	
	if err != nil {
		t.Fatalf("Process with default values failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("Process returned empty data")
	}
}

// Test GetMetadata with various formats
func TestGetMetadata_VariousFormats(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType string
	}{
		{"JPEG", "sample.jpg", "jpeg"},
		{"PNG", "sample.png", "png"},
		{"WebP", "sample.webp", "webp"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := loadTestImage(t, tt.filename)
			
			metadata, err := GetMetadata(data)
			
			if err != nil {
				t.Fatalf("GetMetadata failed: %v", err)
			}
			if metadata.Type != tt.wantType {
				t.Errorf("Expected type %s, got %s", tt.wantType, metadata.Type)
			}
			if metadata.Width == 0 || metadata.Height == 0 {
				t.Error("Expected non-zero dimensions")
			}
		})
	}
}

// Test GetMetadata with corrupted data
func TestGetMetadata_CorruptedData(t *testing.T) {
	data := loadTestImage(t, "corrupted.jpg")
	
	_, err := GetMetadata(data)
	
	if err == nil {
		t.Error("Expected error for corrupted data")
	}
}

// Test Process with invalid options
func TestImageProcessor_Process_InvalidOptions(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	tests := []struct {
		name string
		opts ProcessOptions
	}{
		{
			"Invalid dimensions",
			ProcessOptions{Width: 5, Height: 100, Format: FormatWebP, Quality: 75},
		},
		{
			"Invalid quality",
			ProcessOptions{Width: 100, Height: 100, Format: FormatWebP, Quality: 150},
		},
		{
			"Invalid format",
			ProcessOptions{Width: 100, Height: 100, Format: ImageFormat("bmp"), Quality: 75},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.Process(data, tt.opts)
			
			if err == nil {
				t.Error("Expected error for invalid options")
			}
		})
	}
}

// Test edge case: Resize to same dimensions
func TestImageProcessor_Resize_SameDimensions(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "small.png") // 50x50
	
	result, err := processor.Resize(data, 50, 50)
	
	if err != nil {
		t.Fatalf("Resize to same dimensions failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("Resize returned empty data")
	}
}

// Test format conversion: JPG alias
func TestImageProcessor_ConvertFormat_JPG(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.png")
	
	result, err := processor.ConvertFormat(data, FormatJPG)
	
	if err != nil {
		t.Fatalf("ConvertFormat to JPG failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("ConvertFormat returned empty data")
	}
	
	metadata, err := GetMetadata(result)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	if metadata.Type != "jpeg" {
		t.Errorf("Expected format jpeg, got %s", metadata.Type)
	}
}

// Test Resize with only width specified
func TestImageProcessor_Resize_WidthOnly(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	result, err := processor.Resize(data, 300, 0)
	
	if err != nil {
		t.Fatalf("Resize with width only failed: %v", err)
	}
	
	metadata, err := GetMetadata(result)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	
	if metadata.Width != 300 {
		t.Errorf("Expected width 300, got %d", metadata.Width)
	}
}

// Test Resize with only height specified
func TestImageProcessor_Resize_HeightOnly(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "sample.jpg")
	
	result, err := processor.Resize(data, 0, 200)
	
	if err != nil {
		t.Fatalf("Resize with height only failed: %v", err)
	}
	
	metadata, err := GetMetadata(result)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	
	if metadata.Height != 200 {
		t.Errorf("Expected height 200, got %d", metadata.Height)
	}
}

// Test ConvertFormat with invalid image data
func TestImageProcessor_ConvertFormat_InvalidData(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "corrupted.jpg")
	
	_, err := processor.ConvertFormat(data, FormatWebP)
	
	if err == nil {
		t.Error("Expected error for corrupted data")
	}
}

// Test AdjustQuality with invalid image data
func TestImageProcessor_AdjustQuality_InvalidData(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "corrupted.jpg")
	
	_, err := processor.AdjustQuality(data, 75)
	
	if err == nil {
		t.Error("Expected error for corrupted data")
	}
}

// Test Resize with invalid image data
func TestImageProcessor_Resize_InvalidData(t *testing.T) {
	processor := New()
	data := loadTestImage(t, "corrupted.jpg")
	
	_, err := processor.Resize(data, 100, 100)
	
	if err == nil {
		t.Error("Expected error for corrupted data")
	}
}

// Test ValidateImage with zero-dimension image (edge case)
func TestImageProcessor_ValidateImage_ZeroDimension(t *testing.T) {
	processor := New()
	
	// Create minimal invalid image data
	invalidData := []byte{0x00, 0x00, 0x00, 0x00}
	
	err := processor.ValidateImage(invalidData)
	
	if err != ErrInvalidImage {
		t.Errorf("Expected ErrInvalidImage for zero dimension, got: %v", err)
	}
}

// Helper function to load test images
func loadTestImage(t *testing.T, filename string) []byte {
	t.Helper()
	
	path := "testdata/" + filename
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load test image %s: %v", filename, err)
	}
	return data
}

// Helper function to get image metadata
func getImageMetadata(data []byte) (*ImageMetadata, error) {
	return GetMetadata(data)
}

// Helper function for absolute value
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Benchmark tests

// BenchmarkImageProcessor_Resize_SmallImage benchmarks small image resizing
func BenchmarkImageProcessor_Resize_SmallImage(b *testing.B) {
	processor := New()
	data, err := os.ReadFile("testdata/small.png")
	if err != nil {
		b.Fatalf("Failed to load test image: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.Resize(data, 100, 100)
		if err != nil {
			b.Fatalf("Resize failed: %v", err)
		}
	}
}

// BenchmarkImageProcessor_Resize_LargeImage benchmarks large image resizing
func BenchmarkImageProcessor_Resize_LargeImage(b *testing.B) {
	processor := New()
	data, err := os.ReadFile("testdata/large.jpg")
	if err != nil {
		b.Fatalf("Failed to load test image: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.Resize(data, 800, 600)
		if err != nil {
			b.Fatalf("Resize failed: %v", err)
		}
	}
}

// BenchmarkImageProcessor_ConvertFormat_WebP benchmarks WebP conversion
func BenchmarkImageProcessor_ConvertFormat_WebP(b *testing.B) {
	processor := New()
	data, err := os.ReadFile("testdata/sample.jpg")
	if err != nil {
		b.Fatalf("Failed to load test image: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ConvertFormat(data, FormatWebP)
		if err != nil {
			b.Fatalf("ConvertFormat failed: %v", err)
		}
	}
}
