# Pre-cache Package

The pre-cache package implements startup pre-caching of images for the goimgserver application.

## Overview

This package scans the images directory during application startup and creates default cached versions of all images. This improves initial response times by pre-warming the cache with commonly accessed image formats.

## Features

- **Recursive Directory Scanning**: Scans nested directories and grouped image structures
- **Default Image Exclusion**: Skips system default images to avoid redundant processing
- **Concurrent Processing**: Uses worker pools for parallel image processing
- **Progress Tracking**: Real-time progress reporting with structured logging
- **Error Handling**: Graceful error handling with detailed logging
- **Configurable**: Optional with CLI flags for enabling/disabling and worker count

## Default Pre-cache Settings

All images are pre-cached with the following default settings:
- **Dimensions**: 1000x1000 pixels
- **Format**: WebP
- **Quality**: 95

## Usage

### Basic Usage

```go
import "goimgserver/precache"

// Create configuration
config := &precache.PreCacheConfig{
    ImageDir:         "/path/to/images",
    CacheDir:         "/path/to/cache",
    DefaultImagePath: "/path/to/default.jpg",
    Enabled:          true,
    Workers:          4, // 0 = auto (uses CPU count)
}

// Create pre-cache instance
preCache, err := precache.New(
    config,
    fileResolver,
    cacheManager,
    processorAdapter,
)

// Run synchronously
stats, err := preCache.Run(context.Background())

// Or run asynchronously (non-blocking)
preCache.RunAsync(context.Background())
```

### CLI Flags

```bash
# Enable pre-cache (enabled by default)
./goimgserver -precache=true

# Disable pre-cache
./goimgserver -precache=false

# Specify number of workers
./goimgserver -precache-workers=8

# Auto workers (uses CPU count)
./goimgserver -precache-workers=0
```

## Architecture

### Components

1. **Scanner** (`scanner.go`): Recursively scans directories for supported image files
   - Supports JPEG, PNG, WebP formats
   - Excludes system default image
   - Handles nested directories and grouped images

2. **Processor** (`processor.go`): Processes individual images and stores in cache
   - Uses file resolution system for path handling
   - Skips already cached images
   - Stores with default pre-cache settings

3. **Progress Reporter** (`progress.go`): Tracks and logs pre-cache progress
   - Real-time progress updates
   - Error tracking and reporting
   - Completion statistics

4. **Concurrent Executor** (`concurrent.go`): Manages worker pool for parallel processing
   - Configurable worker count
   - Context-aware cancellation
   - Thread-safe statistics

5. **PreCache Coordinator** (`precache.go`): Main API for pre-cache operations
   - Coordinates all components
   - Provides sync and async execution
   - Auto-configures workers based on CPU count

## Integration with File Resolution System

The pre-cache system integrates with the file resolution system (gh-0015) for consistent path handling:

- Uses `FileResolver` to resolve image paths
- Supports grouped images (subdirectory organization)
- Handles group default images (default.* in subdirectories)
- Excludes system default image from pre-caching

## Integration with Default Image System

The pre-cache system respects the default image system (gh-0002):

- **Excludes system default image**: The system default image (e.g., `/images/default.jpg`) is not pre-cached
- **Includes group defaults**: Group default images (e.g., `/images/cats/default.jpg`) are pre-cached normally
- **Cache structure**: Pre-cached images use the same cache structure as regular requests

## Test Coverage

The package has comprehensive test coverage (92.1%):

- **Scanner Tests**: Empty directories, single/multiple images, nested directories, exclusions
- **Processor Tests**: Default settings, already cached, error handling, corrupted images
- **Progress Tests**: Tracking, logging, error reporting
- **Concurrent Tests**: Worker pool, thread safety, error handling, cancellation
- **Integration Tests**: Full workflow, async execution, configuration

Run tests:
```bash
cd src/precache
go test -v ./...
go test -cover ./...
```

## Performance Considerations

- **Asynchronous by Default**: Pre-cache runs asynchronously to not block server startup
- **Worker Pools**: Concurrent processing with configurable worker count
- **Skip Cached**: Already cached images are skipped to avoid redundant work
- **Context Cancellation**: Supports graceful cancellation via context

## Error Handling

The pre-cache system handles errors gracefully:

- **Per-image errors**: Individual image processing errors don't stop the overall process
- **Error tracking**: All errors are logged and tracked in statistics
- **Corrupted images**: Skipped with error logging
- **Missing files**: Skipped with error logging

## Statistics

After completion, pre-cache provides detailed statistics:

```go
type Stats struct {
    TotalImages   int           // Total images found
    ProcessedOK   int           // Successfully processed
    Skipped       int           // Skipped (already cached)
    Errors        int           // Processing errors
    Duration      time.Duration // Total duration
    StartTime     time.Time     // Start timestamp
    EndTime       time.Time     // End timestamp
}
```

## Example Output

```
2025/10/09 03:42:10 Starting pre-cache: scanning /path/to/images
2025/10/09 03:42:10 Found 5 images to pre-cache
2025/10/09 03:42:10 Starting pre-cache: 5 images to process
2025/10/09 03:42:10 Pre-cache progress: 1/5 (20.0%) - image0.jpg
2025/10/09 03:42:10 Pre-cache progress: 2/5 (40.0%) - image2.jpg
2025/10/09 03:42:10 Pre-cache progress: 3/5 (60.0%) - image3.jpg
2025/10/09 03:42:10 Pre-cache progress: 4/5 (80.0%) - image4.jpg
2025/10/09 03:42:10 Pre-cache progress: 5/5 (100.0%) - image1.jpg
2025/10/09 03:42:10 Pre-cache complete: 5 processed, 0 skipped, 0 errors in 877.057µs
```

## Dependencies

- `goimgserver/cache`: Cache management system (gh-0004)
- `goimgserver/resolver`: File resolution system (gh-0015)
- `goimgserver/config`: Core configuration (gh-0002)
- Image processor interface (adapted from gh-0003)

## Future Enhancements

- [ ] Pre-cache multiple sizes (not just 1000x1000)
- [ ] Pre-cache multiple formats (not just WebP)
- [ ] Progress persistence (resume after restart)
- [ ] Rate limiting to control resource usage
- [ ] Metrics collection for monitoring
- [ ] Web UI for pre-cache status

## License

Part of the goimgserver project.
