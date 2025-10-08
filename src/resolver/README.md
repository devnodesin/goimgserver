# File Resolver Package

A comprehensive file resolution system that supports extension auto-detection, grouped image organization, and extension priority ordering with security validation.

## Features

- **Extension Auto-Detection**: Automatically finds files when no extension is specified
- **Extension Priority**: jpg/jpeg (highest) → png (medium) → webp (lowest)
- **Grouped Images**: Organize images in folders with default fallback
- **Security**: Path traversal prevention and safe symlink handling
- **Caching**: Optional thread-safe caching for performance
- **Fallback Chain**: Structured fallback system for missing files

## Installation

```go
import "goimgserver/resolver"
```

## Usage

### Basic Resolution

```go
// Create a resolver
res := resolver.NewResolver("/path/to/images")

// Resolve a file with extension
result, err := res.Resolve("cat.jpg")
if err == nil {
    fmt.Println(result.ResolvedPath)  // /path/to/images/cat.jpg
}

// Auto-detect extension
result, err = res.Resolve("cat")
// Searches for cat.jpg, cat.jpeg, cat.png, cat.webp in priority order
```

### Extension Priority

When multiple extensions exist for the same file:

```go
// Files: profile.jpg, profile.png, profile.webp
result, err := res.Resolve("profile")
// Returns: profile.jpg (highest priority)
```

### Grouped Images

```go
// Request group default
result, err := res.Resolve("cats")
// Returns: /path/to/images/cats/default.jpg

// Request specific grouped image
result, err := res.Resolve("cats/cat_white")
// Auto-detects: /path/to/images/cats/cat_white.jpg
```

### Caching

```go
// Create resolver with caching enabled
res := resolver.NewResolverWithCache("/path/to/images")

// First call hits filesystem
result1, _ := res.Resolve("cat.jpg")

// Second call hits cache (much faster)
result2, _ := res.Resolve("cat.jpg")
```

### Custom Default

```go
result, err := res.ResolveWithDefault("missing.jpg", "/path/to/custom/default.jpg")
// Falls back to custom default if missing.jpg not found
```

## Resolution Algorithm

### Single Images

```
Request: /img/cat
1. Try cat.jpg → found? return
2. Try cat.jpeg → found? return
3. Try cat.png → found? return
4. Try cat.webp → found? return
5. Fallback to system default
```

### Grouped Images

```
Request: /img/cats/cat_white
1. Try cats/cat_white.jpg → found? return
2. Try cats/cat_white.jpeg → found? return
3. Try cats/cat_white.png → found? return
4. Try cats/cat_white.webp → found? return
5. Fallback to cats/default.*
6. Fallback to system default
```

## Fallback Chain

1. **Requested File**: Try to resolve the exact file requested
2. **Group Default** (grouped images only): Fall back to group's default image
3. **System Default**: Fall back to system-wide default image
4. **Error**: Return `ErrFileNotFound` if no default exists

## Security

The resolver includes multiple security features:

- **Path Traversal Prevention**: Blocks `../` attempts to escape image directory
- **Absolute Path Rejection**: Rejects absolute paths like `/etc/passwd`
- **Null Byte Protection**: Sanitizes null bytes in paths
- **Symlink Validation**: Only follows symlinks within the image directory

## API Reference

### Types

```go
type ResolutionResult struct {
    ResolvedPath string   // Absolute path to resolved file
    IsGrouped    bool     // Whether this is a grouped image
    IsFallback   bool     // Whether this is a fallback result
    FallbackType string   // Type of fallback if applicable
}
```

### Functions

```go
// NewResolver creates a basic resolver
func NewResolver(imageDir string) *Resolver

// NewResolverWithCache creates a resolver with caching enabled
func NewResolverWithCache(imageDir string) *Resolver

// Resolve resolves a request path to an actual file
func (r *Resolver) Resolve(requestPath string) (*ResolutionResult, error)

// ResolveWithDefault resolves with a custom default fallback
func (r *Resolver) ResolveWithDefault(requestPath string, defaultPath string) (*ResolutionResult, error)
```

## Performance

Benchmark results (on modern hardware):

```
BenchmarkFileResolver_CacheHit-4              15.22 ns/op      0 B/op    0 allocs/op
BenchmarkFileResolver_SingleImage-4           9649 ns/op    1400 B/op   21 allocs/op
BenchmarkFileResolver_AutoDetection-4        11256 ns/op    1712 B/op   26 allocs/op
BenchmarkFileResolver_GroupedImage-4          4560 ns/op     672 B/op   10 allocs/op
```

Cache hits are **~600x faster** than filesystem resolution.

## Testing

Run tests:

```bash
go test ./resolver -v
```

Run benchmarks:

```bash
go test ./resolver -bench=. -benchmem
```

Check coverage:

```bash
go test ./resolver -cover
```

## Example

See [example/main.go](example/main.go) for a complete working example demonstrating all features.

## File Organization

```
resolver/
├── types.go           # Core types and interfaces
├── resolver.go        # Main resolution logic
├── security.go        # Security validation
├── cache.go          # Thread-safe caching
├── resolver_test.go   # Core unit tests
├── security_test.go   # Security tests
├── cache_test.go      # Cache tests
├── benchmark_test.go  # Performance benchmarks
└── example/          # Usage examples
    └── main.go
```

## Test Coverage

Current coverage: **91.2%** of statements

All critical paths are tested including:
- Extension priority resolution
- Grouped image handling
- Security validation
- Path traversal prevention
- Symlink handling
- Caching operations
- Concurrent access

## License

Part of the goimgserver project.
