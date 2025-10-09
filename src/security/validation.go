package security

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Validation errors
var (
	ErrInvalidPath      = errors.New("invalid path")
	ErrPathTraversal    = errors.New("path traversal attempt detected")
	ErrInvalidDimension = errors.New("invalid dimension")
	ErrInvalidQuality   = errors.New("invalid quality value")
	ErrInvalidFormat    = errors.New("invalid image format")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrInvalidExtension = errors.New("invalid file extension")
	ErrFileTooLarge     = errors.New("file size exceeds limit")
	ErrFileEmpty        = errors.New("file is empty")
)

// Validation constants
const (
	MinDimension   = 10
	MaxDimension   = 4000
	DefaultWidth   = 1000
	DefaultHeight  = 1000
	MinQuality     = 1
	MaxQuality     = 100
	DefaultQuality = 75
	DefaultFormat  = "webp"
	MaxFileSize    = 50 * 1024 * 1024 // 50MB default max file size
)

// Valid formats map
var validFormats = map[string]bool{
	"webp": true,
	"png":  true,
	"jpeg": true,
	"jpg":  true,
}

// Regular expressions for parameter parsing
var (
	dimensionsRegex = regexp.MustCompile(`^(\d+)x(\d+)$`)
	widthOnlyRegex  = regexp.MustCompile(`^(\d+)$`)
	qualityRegex    = regexp.MustCompile(`^q(\d+)$`)
)

// ValidatePath validates a file path to prevent security issues
func ValidatePath(path string) error {
	// Empty path is invalid
	if path == "" {
		return ErrInvalidPath
	}

	// Remove null bytes (null byte injection attack)
	if strings.Contains(path, "\x00") {
		return ErrInvalidPath
	}

	// Clean the path to normalize it
	cleanPath := filepath.Clean(path)

	// Reject absolute paths
	if filepath.IsAbs(cleanPath) {
		return ErrInvalidPath
	}

	// Additional check: split by path separator and look for exactly ".." in segments
	segments := strings.Split(cleanPath, string(filepath.Separator))
	for _, segment := range segments {
		if segment == ".." {
			return ErrPathTraversal
		}
	}

	// Check for backslash (Windows path separator) in Unix-style paths
	// This catches attempts to bypass path traversal checks
	if strings.Contains(path, "\\") && strings.Contains(path, "..") {
		return ErrPathTraversal
	}

	return nil
}

// ValidateDimensions validates and sanitizes width and height values
// Returns safe default values if inputs are invalid
func ValidateDimensions(width, height int) (int, int, error) {
	validatedWidth := width
	validatedHeight := height

	// Validate width
	if width < MinDimension || width > MaxDimension {
		validatedWidth = DefaultWidth
	}

	// Validate height - special case: 0 means maintain aspect ratio (valid)
	if height == 0 {
		validatedHeight = 0
	} else if height < MinDimension || height > MaxDimension {
		validatedHeight = DefaultHeight
	}

	return validatedWidth, validatedHeight, nil
}

// ValidateQuality validates and sanitizes quality value
// Returns safe default if input is invalid
func ValidateQuality(quality int) int {
	if quality < MinQuality || quality > MaxQuality {
		return DefaultQuality
	}
	return quality
}

// ValidateFormat validates and sanitizes image format
// Returns safe default if input is invalid
func ValidateFormat(format string) string {
	// Normalize to lowercase
	normalized := strings.ToLower(strings.TrimSpace(format))

	// Check if format is valid
	if !validFormats[normalized] {
		return DefaultFormat
	}

	return normalized
}

// ProcessingParams represents validated image processing parameters
type ProcessingParams struct {
	Width   int
	Height  int
	Quality int
	Format  string
}

// ParseAndValidateParameters parses and validates URL segments into safe ProcessingParams
// This function implements graceful parsing with security validation
func ParseAndValidateParameters(segments []string) ProcessingParams {
	params := ProcessingParams{
		Width:   DefaultWidth,
		Height:  DefaultHeight,
		Format:  DefaultFormat,
		Quality: DefaultQuality,
	}

	// Track which parameters have been set (first wins)
	hasDimensions := false
	hasFormat := false
	hasQuality := false

	for _, segment := range segments {
		// Skip empty segments
		if segment == "" {
			continue
		}

		// Security check: reject segments with null bytes
		if strings.Contains(segment, "\x00") {
			continue
		}

		// Try to parse dimensions (WxH)
		if !hasDimensions {
			if matches := dimensionsRegex.FindStringSubmatch(segment); matches != nil {
				width, _ := strconv.Atoi(matches[1])
				height, _ := strconv.Atoi(matches[2])

				// Validate and apply
				validatedWidth, validatedHeight, _ := ValidateDimensions(width, height)
				params.Width = validatedWidth
				params.Height = validatedHeight
				hasDimensions = true
				continue
			}
		}

		// Try to parse width only
		if !hasDimensions {
			if matches := widthOnlyRegex.FindStringSubmatch(segment); matches != nil {
				width, _ := strconv.Atoi(matches[1])

				// Validate width
				if width >= MinDimension && width <= MaxDimension {
					params.Width = width
					params.Height = 0 // maintain aspect ratio
					hasDimensions = true
					continue
				}
			}
		}

		// Try to parse quality
		if !hasQuality {
			if matches := qualityRegex.FindStringSubmatch(segment); matches != nil {
				quality, _ := strconv.Atoi(matches[1])

				// Validate and apply
				params.Quality = ValidateQuality(quality)
				hasQuality = true
				continue
			}
		}

		// Try to parse format
		if !hasFormat {
			// Validate format (this will return default if invalid)
			validatedFormat := ValidateFormat(segment)
			if validatedFormat != DefaultFormat || validFormats[strings.ToLower(segment)] {
				params.Format = validatedFormat
				hasFormat = true
				continue
			}
		}

		// If we reach here, the segment is unrecognized - safely ignore it
		// This prevents injection attacks via ignored parameters
	}

	return params
}

// ValidateFileType validates file type by checking magic numbers (file headers)
func ValidateFileType(data []byte) (string, error) {
	if len(data) < 8 {
		return "", ErrInvalidFileType
	}

	// Check PNG first (starts with 89 50 4E 47 0D 0A 1A 0A)
	if data[0] == 0x89 &&
		data[1] == 0x50 &&
		data[2] == 0x4E &&
		data[3] == 0x47 &&
		data[4] == 0x0D &&
		data[5] == 0x0A &&
		data[6] == 0x1A &&
		data[7] == 0x0A {
		return "png", nil
	}

	// Check JPEG (starts with FF D8 and third byte is also FF)
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpeg", nil
	}

	// Check WebP (RIFF....WEBP)
	if len(data) >= 12 &&
		data[0] == 0x52 && // R
		data[1] == 0x49 && // I
		data[2] == 0x46 && // F
		data[3] == 0x46 && // F
		data[8] == 0x57 && // W
		data[9] == 0x45 && // E
		data[10] == 0x42 && // B
		data[11] == 0x50 { // P
		return "webp", nil
	}

	return "", fmt.Errorf("%w: unrecognized file signature", ErrInvalidFileType)
}

// ValidateFileExtension validates file extension
func ValidateFileExtension(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	// Remove the leading dot
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}

	// Check if extension is valid
	if !validFormats[ext] {
		return fmt.Errorf("%w: %s", ErrInvalidExtension, ext)
	}

	// Additional check: ensure the file doesn't have multiple extensions
	// e.g., image.jpg.exe should be rejected
	// But allow hidden files (starting with .)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// Remove leading dots (hidden files on Unix)
	baseName := strings.TrimLeft(filepath.Base(nameWithoutExt), ".")
	
	if strings.Contains(baseName, ".") {
		// Check if there's another extension
		secondExt := filepath.Ext(nameWithoutExt)
		if len(secondExt) > 1 {
			// Has a second extension - this is suspicious
			return fmt.Errorf("%w: multiple extensions detected", ErrInvalidExtension)
		}
	}

	return nil
}

// ValidateFileSize validates file size against a maximum limit
func ValidateFileSize(size, maxSize int64) error {
	if size <= 0 {
		return ErrFileEmpty
	}

	if size > maxSize {
		return fmt.Errorf("%w: size %d exceeds limit %d", ErrFileTooLarge, size, maxSize)
	}

	return nil
}
