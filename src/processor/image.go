package processor

import (
	"fmt"
	"github.com/h2non/bimg"
)

// bimgProcessor implements ImageProcessor using bimg
type bimgProcessor struct{}

// New creates a new ImageProcessor instance
func New() ImageProcessor {
	return &bimgProcessor{}
}

// Resize resizes an image to the specified dimensions
// If width or height is 0, aspect ratio is maintained
func (p *bimgProcessor) Resize(data []byte, width, height int) ([]byte, error) {
	if err := validateDimensions(width, height); err != nil {
		return nil, err
	}
	
	img := bimg.NewImage(data)
	
	options := bimg.Options{
		Width:  width,
		Height: height,
	}
	
	// If only one dimension is provided, bimg will maintain aspect ratio
	if width == 0 || height == 0 {
		options.Force = false
	}
	
	result, err := img.Process(options)
	if err != nil {
		return nil, fmt.Errorf("resize failed: %w", err)
	}
	
	return result, nil
}

// ConvertFormat converts an image to the specified format
func (p *bimgProcessor) ConvertFormat(data []byte, format ImageFormat) ([]byte, error) {
	bimgType, err := formatToBimgType(format)
	if err != nil {
		return nil, err
	}
	
	img := bimg.NewImage(data)
	
	options := bimg.Options{
		Type: bimgType,
	}
	
	result, err := img.Process(options)
	if err != nil {
		return nil, fmt.Errorf("format conversion failed: %w", err)
	}
	
	return result, nil
}

// AdjustQuality adjusts the quality of an image
func (p *bimgProcessor) AdjustQuality(data []byte, quality int) ([]byte, error) {
	if err := validateQuality(quality); err != nil {
		return nil, err
	}
	
	img := bimg.NewImage(data)
	
	options := bimg.Options{
		Quality: quality,
	}
	
	result, err := img.Process(options)
	if err != nil {
		return nil, fmt.Errorf("quality adjustment failed: %w", err)
	}
	
	return result, nil
}

// Process performs combined operations (resize + format + quality)
func (p *bimgProcessor) Process(data []byte, opts ProcessOptions) ([]byte, error) {
	// Validate input
	if err := p.ValidateImage(data); err != nil {
		return nil, err
	}
	
	if err := validateDimensions(opts.Width, opts.Height); err != nil {
		return nil, err
	}
	
	if err := validateQuality(opts.Quality); err != nil {
		return nil, err
	}
	
	bimgType, err := formatToBimgType(opts.Format)
	if err != nil {
		return nil, err
	}
	
	img := bimg.NewImage(data)
	
	bimgOpts := bimg.Options{
		Width:   opts.Width,
		Height:  opts.Height,
		Type:    bimgType,
		Quality: opts.Quality,
	}
	
	result, err := img.Process(bimgOpts)
	if err != nil {
		return nil, ErrInvalidImage
	}
	
	return result, nil
}

// ValidateImage checks if the data is a valid image
func (p *bimgProcessor) ValidateImage(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidImage
	}
	
	img := bimg.NewImage(data)
	imgType := img.Type()
	
	// bimg returns "unknown" for invalid images
	if imgType == "unknown" {
		return ErrInvalidImage
	}
	
	// Try to get size to validate the image is readable
	size, err := img.Size()
	if err != nil {
		return ErrInvalidImage
	}
	
	// Check for reasonable dimensions
	if size.Width == 0 || size.Height == 0 {
		return ErrInvalidImage
	}
	
	return nil
}

// validateDimensions checks if dimensions are within valid range
func validateDimensions(width, height int) error {
	// If both are 0, it's valid (no resize)
	if width == 0 && height == 0 {
		return nil
	}
	
	// If only one is 0, check the other
	if width == 0 {
		if height < MinDimension || height > MaxDimension {
			return ErrInvalidDimensions
		}
		return nil
	}
	
	if height == 0 {
		if width < MinDimension || width > MaxDimension {
			return ErrInvalidDimensions
		}
		return nil
	}
	
	// Both are non-zero, check both
	if width < MinDimension || width > MaxDimension {
		return ErrInvalidDimensions
	}
	
	if height < MinDimension || height > MaxDimension {
		return ErrInvalidDimensions
	}
	
	return nil
}

// validateQuality checks if quality is within valid range (1-100)
func validateQuality(quality int) error {
	if quality < MinQuality || quality > MaxQuality {
		return ErrInvalidQuality
	}
	return nil
}

// formatToBimgType converts ImageFormat to bimg.ImageType
func formatToBimgType(format ImageFormat) (bimg.ImageType, error) {
	switch format {
	case FormatWebP:
		return bimg.WEBP, nil
	case FormatPNG:
		return bimg.PNG, nil
	case FormatJPEG, FormatJPG:
		return bimg.JPEG, nil
	default:
		return bimg.UNKNOWN, ErrUnsupportedFormat
	}
}

// GetMetadata extracts metadata from image data using bimg
func GetMetadata(data []byte) (*ImageMetadata, error) {
	img := bimg.NewImage(data)
	
	size, err := img.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get image size: %w", err)
	}
	
	imgType := img.Type()
	
	return &ImageMetadata{
		Width:  size.Width,
		Height: size.Height,
		Type:   imgType,
	}, nil
}
