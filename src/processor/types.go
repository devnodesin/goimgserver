package processor

import "errors"

// ImageFormat represents supported image formats
type ImageFormat string

const (
	FormatWebP ImageFormat = "webp"
	FormatPNG  ImageFormat = "png"
	FormatJPEG ImageFormat = "jpeg"
	FormatJPG  ImageFormat = "jpg"
)

// Dimension constraints
const (
	MinDimension     = 10
	MaxDimension     = 4000
	DefaultWidth     = 1000
	DefaultHeight    = 1000
	DefaultQuality   = 75
	MinQuality       = 1
	MaxQuality       = 100
)

// Common errors
var (
	ErrInvalidDimensions      = errors.New("invalid dimensions: must be between 10 and 4000 pixels")
	ErrInvalidQuality         = errors.New("invalid quality: must be between 1 and 100")
	ErrUnsupportedFormat      = errors.New("unsupported image format")
	ErrInvalidImage           = errors.New("invalid or corrupted image data")
	ErrUnsupportedInputFormat = errors.New("unsupported input image format")
)

// ProcessOptions contains options for image processing
type ProcessOptions struct {
	Width   int
	Height  int
	Format  ImageFormat
	Quality int
}

// ImageMetadata contains basic image information
type ImageMetadata struct {
	Width  int
	Height int
	Type   string
}

// ImageProcessor defines the interface for image processing operations
type ImageProcessor interface {
	// Resize resizes an image to the specified dimensions
	// If width or height is 0, aspect ratio is maintained
	Resize(data []byte, width, height int) ([]byte, error)
	
	// ConvertFormat converts an image to the specified format
	ConvertFormat(data []byte, format ImageFormat) ([]byte, error)
	
	// AdjustQuality adjusts the quality of an image
	AdjustQuality(data []byte, quality int) ([]byte, error)
	
	// Process performs combined operations (resize + format + quality)
	Process(data []byte, opts ProcessOptions) ([]byte, error)
	
	// ValidateImage checks if the data is a valid image
	ValidateImage(data []byte) error
}
