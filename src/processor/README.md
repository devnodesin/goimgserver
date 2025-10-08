# Image Processor Package

This package provides image processing functionality using the bimg library (libvips wrapper).

## Features

- **Image Resizing**: Resize images with dimension constraints (10px-4000px)
  - Both dimensions: `Resize(data, 200, 150)`
  - Width only (maintain aspect ratio): `Resize(data, 200, 0)`
  - Height only (maintain aspect ratio): `Resize(data, 0, 150)`

- **Format Conversion**: Convert between image formats
  - Supported formats: WebP (default), PNG, JPEG/JPG
  - Example: `ConvertFormat(data, FormatWebP)`

- **Quality Adjustment**: Adjust image quality (1-100)
  - Default quality: 75
  - Example: `AdjustQuality(data, 85)`

- **Combined Operations**: Process with multiple operations
  - Resize + format conversion + quality adjustment
  - Example: `Process(data, ProcessOptions{Width: 300, Height: 200, Format: FormatWebP, Quality: 85})`

- **Image Validation**: Validate image headers and integrity
  - Uses magic numbers to detect file types
  - Example: `ValidateImage(data)`

## Usage

```go
import "goimgserver/processor"

// Create a new processor
processor := processor.New()

// Resize an image
resized, err := processor.Resize(imageData, 400, 300)

// Convert to WebP
webp, err := processor.ConvertFormat(imageData, processor.FormatWebP)

// Process with all operations
opts := processor.ProcessOptions{
    Width:   300,
    Height:  200,
    Format:  processor.FormatWebP,
    Quality: 75,
}
result, err := processor.Process(imageData, opts)
```

## Constants

- `MinDimension`: 10px (minimum allowed dimension)
- `MaxDimension`: 4000px (maximum allowed dimension)
- `DefaultWidth`: 1000px
- `DefaultHeight`: 1000px
- `DefaultQuality`: 75
- `MinQuality`: 1
- `MaxQuality`: 100

## Error Handling

The package defines specific errors for different failure scenarios:
- `ErrInvalidDimensions`: Dimensions outside allowed range
- `ErrInvalidQuality`: Quality outside 1-100 range
- `ErrUnsupportedFormat`: Unsupported image format
- `ErrInvalidImage`: Corrupted or invalid image data
- `ErrUnsupportedInputFormat`: Input format not supported

## Test Coverage

- **95.2%** test coverage
- **58** test cases covering:
  - Valid operations
  - Edge cases
  - Error scenarios
  - Boundary conditions
  - Performance benchmarks

## Performance Benchmarks

On AMD EPYC 7763 64-Core Processor:
- Small image resize (50x50): ~2.1ms per operation, 680 B/op
- Large image resize (3000x2000): ~11.4ms per operation, 3504 B/op
- WebP conversion: ~6.4ms per operation, 608 B/op

## Dependencies

- `github.com/h2non/bimg`: Go bindings for libvips
- Requires libvips (v8.15+) installed on the system

## Development

This package was developed using Test-Driven Development (TDD) methodology:
1. **Red**: All tests written first
2. **Green**: Minimal implementation to pass tests
3. **Refactor**: Code optimization while maintaining tests

For more details, see `image_test.go` for comprehensive test examples.
