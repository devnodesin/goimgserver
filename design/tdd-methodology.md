# Test-Driven Development (TDD) Methodology for goimgserver

## Overview

The goimgserver project **must** be developed using Test-Driven Development (TDD) methodology. This document outlines the TDD principles, practices, and specific guidelines for implementing the image processing service.

## TDD Fundamentals

### The Red-Green-Refactor Cycle

1. **🔴 RED**: Write a failing test
   - Write the smallest possible test that defines desired behavior
   - Ensure the test fails for the right reason (functionality doesn't exist)
   - Tests should be specific and focused on one behavior

2. **🟢 GREEN**: Make the test pass
   - Write the minimal code needed to make the test pass
   - Don't over-engineer or add unnecessary features
   - Focus on making the test pass quickly

3. **🔵 REFACTOR**: Improve the code
   - Clean up code structure while keeping tests green
   - Extract common functionality and improve design
   - Ensure all tests continue to pass

### TDD Rules (Uncle Bob's Three Laws)

1. **You are not allowed to write any production code** unless it is to make a failing unit test pass
2. **You are not allowed to write any more of a unit test** than is sufficient to fail; and compilation failures are failures
3. **You are not allowed to write any more production code** than is sufficient to pass the one failing unit test

## TDD Implementation Strategy for goimgserver

### 1. Start with Interface Design

Before implementing any functionality, design interfaces through tests:

```go
// Example: Start by testing the desired interface
func TestImageProcessor_Resize_ShouldResizeImage(t *testing.T) {
    processor := NewImageProcessor()
    imageData := loadTestImage("sample.jpg")
    
    result, err := processor.Resize(imageData, ResizeOptions{
        Width:  400,
        Height: 300,
        Format: "webp",
        Quality: 75,
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 400, result.Width)
    assert.Equal(t, 300, result.Height)
    assert.Equal(t, "webp", result.Format)
}
```

### 2. Test Categories for Each Component

#### Unit Tests (Primary Focus)
- **Configuration parsing**: Test CLI argument parsing and validation
- **Image processing**: Test resize, format conversion, quality adjustment
- **Cache operations**: Test cache storage, retrieval, and invalidation
- **URL parsing**: Test endpoint parameter extraction and validation
- **Error handling**: Test error scenarios and edge cases

#### Integration Tests
- **HTTP endpoints**: Test complete request/response cycles
- **File system operations**: Test image file access and cache directory operations
- **Git operations**: Test git update functionality

#### End-to-End Tests
- **Complete workflows**: Test full image processing pipelines
- **Performance tests**: Test under load and stress conditions

### 3. TDD Implementation Order

#### Phase 1: Foundation (TDD First)
1. **Configuration System**
   ```go
   func TestConfig_ParseArgs_DefaultValues(t *testing.T)
   func TestConfig_ParseArgs_CustomValues(t *testing.T)
   func TestConfig_ValidateDirectories_CreatesIfMissing(t *testing.T)
   ```

2. **Image Processing Core**
   ```go
   func TestImageProcessor_Resize_ValidDimensions(t *testing.T)
   func TestImageProcessor_ConvertFormat_WebP(t *testing.T)
   func TestImageProcessor_AdjustQuality_ValidRange(t *testing.T)
   ```

3. **Cache Management**
   ```go
   func TestCache_Store_ValidKey(t *testing.T)
   func TestCache_Retrieve_ExistingKey(t *testing.T)
   func TestCache_GenerateKey_ConsistentHash(t *testing.T)
   ```

#### Phase 2: HTTP Layer (TDD Driven)
1. **URL Parameter Parsing**
   ```go
   func TestParseImageRequest_ValidDimensions(t *testing.T)
   func TestParseImageRequest_InvalidDimensions_ReturnsError(t *testing.T)
   ```

2. **HTTP Handlers**
   ```go
   func TestImageHandler_GET_ExistingImage(t *testing.T)
   func TestImageHandler_GET_MissingImage_Returns404(t *testing.T)
   ```

### 4. Test Organization Structure

```
src/
├── config/
│   ├── config.go
│   └── config_test.go
├── processor/
│   ├── image.go
│   ├── image_test.go
│   └── testdata/
│       ├── sample.jpg
│       └── sample.png
├── cache/
│   ├── manager.go
│   └── manager_test.go
├── handlers/
│   ├── image.go
│   ├── image_test.go
│   ├── command.go
│   └── command_test.go
└── main.go
```

### 5. Test Data Management

#### Test Fixtures
- Create small, representative test images in multiple formats
- Use embedded files or testdata directory for test assets
- Generate test images programmatically for specific test cases

#### Mock and Stub Strategy
- Mock external dependencies (file system, HTTP clients)
- Use interfaces to enable easy mocking
- Stub complex operations for unit tests

### 6. TDD Best Practices for goimgserver

#### Test Naming Conventions
```go
// Pattern: Test_<MethodName>_<Scenario>_<ExpectedResult>
func TestImageProcessor_Resize_ValidDimensions_ReturnsResizedImage(t *testing.T)
func TestImageProcessor_Resize_InvalidDimensions_ReturnsError(t *testing.T)
func TestCache_Store_ExistingKey_OverwritesValue(t *testing.T)
```

#### Table-Driven Tests for Multiple Scenarios
```go
func TestImageProcessor_Resize_VariousDimensions(t *testing.T) {
    tests := []struct {
        name           string
        width, height  int
        expectedError  string
    }{
        {"valid dimensions", 400, 300, ""},
        {"too small width", 5, 300, "width too small"},
        {"too large height", 400, 5000, "height too large"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### URL Parameter Parsing (Graceful Parsing)
```go
func TestParseImageRequest_GracefulParsing_InvalidParamsIgnored(t *testing.T) {
    tests := []struct {
        name           string
        url            string
        expectedWidth  int
        expectedHeight int
        expectedFormat string
        expectedQuality int
        ignored        []string
    }{
        {
            "invalid params ignored",
            "/img/photo.jpg/800x600/webp/q90/wow",
            800, 600, "webp", 90,
            []string{"wow"},
        },
        {
            "duplicate params first wins",
            "/img/logo.png/300/400/png/jpeg/q85/q95",
            300, 1000, "png", 85,  // height defaults, first format and quality win
            []string{"400", "jpeg", "q95"},
        },
        {
            "invalid values use defaults",
            "/img/banner.jpg/99999x1/invalidformat/q95005",
            1000, 1000, "webp", 75,  // all use defaults
            []string{"99999x1", "invalidformat", "q95005"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, ignored := ParseImageRequest(tt.url)
            
            assert.Equal(t, tt.expectedWidth, result.Width)
            assert.Equal(t, tt.expectedHeight, result.Height)
            assert.Equal(t, tt.expectedFormat, result.Format)
            assert.Equal(t, tt.expectedQuality, result.Quality)
            assert.Equal(t, tt.ignored, ignored)
        })
    }
}
```
```

### 7. Continuous Testing Workflow

#### Local Development
1. Run tests after every small change: `go test ./...`
2. Use test coverage to identify untested code: `go test -cover ./...`
3. Run tests before committing: `go test ./... && go vet ./...`

#### CI/CD Integration
1. Run full test suite on every commit
2. Enforce minimum test coverage thresholds
3. Run performance benchmarks on performance-critical changes
4. Include race condition detection: `go test -race ./...`

### 8. Performance Testing Strategy

#### Benchmark Tests
```go
func BenchmarkImageProcessor_Resize_LargeImage(b *testing.B) {
    processor := NewImageProcessor()
    largeImage := loadLargeTestImage()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        processor.Resize(largeImage, 800, 600)
    }
}
```

#### Memory and Performance Monitoring
- Test memory usage with large images
- Benchmark concurrent processing scenarios
- Test cache performance under load

## Benefits of TDD for goimgserver

1. **Design Quality**: Tests drive better API design and interfaces
2. **Documentation**: Tests serve as living documentation of expected behavior
3. **Confidence**: Refactoring and changes are safer with comprehensive test coverage
4. **Bug Prevention**: Issues are caught early in the development cycle
5. **Maintainability**: Well-tested code is easier to maintain and extend

## Common TDD Pitfalls to Avoid

1. **Writing tests after implementation** - defeats the purpose of TDD
2. **Testing implementation details** instead of behavior
3. **Writing too many tests at once** - stay focused on one failing test
4. **Ignoring the refactor step** - code quality degrades without refactoring
5. **Not running tests frequently** - feedback loop becomes too long

## Success Metrics

- **Test Coverage**: Maintain >90% test coverage
- **Test Speed**: Unit tests should complete in <5 seconds
- **Integration Test Speed**: Should complete in <30 seconds
- **Bug Escape Rate**: Minimize bugs found in production
- **Refactoring Confidence**: Easy to make changes without breaking functionality

## Graceful URL Parsing TDD Examples

### URL Parameter Parsing with Fault Tolerance
```go
func TestParseImageRequest_GracefulParsing_InvalidParamsIgnored(t *testing.T) {
    tests := []struct {
        name           string
        url            string
        expectedWidth  int
        expectedHeight int
        expectedFormat string
        expectedQuality int
        ignored        []string
    }{
        {
            "invalid params ignored",
            "/img/photo.jpg/800x600/webp/q90/wow",
            800, 600, "webp", 90,
            []string{"wow"},
        },
        {
            "duplicate params first wins",
            "/img/logo.png/300/400/png/jpeg/q85/q95",
            300, 1000, "png", 85,  // height defaults, first format and quality win
            []string{"400", "jpeg", "q95"},
        },
        {
            "invalid values use defaults",
            "/img/banner.jpg/99999x1/invalidformat/q95005",
            1000, 1000, "webp", 75,  // all use defaults
            []string{"99999x1", "invalidformat", "q95005"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, ignored := ParseImageRequest(tt.url)
            
            assert.Equal(t, tt.expectedWidth, result.Width)
            assert.Equal(t, tt.expectedHeight, result.Height)
            assert.Equal(t, tt.expectedFormat, result.Format)
            assert.Equal(t, tt.expectedQuality, result.Quality)
            assert.Equal(t, tt.ignored, ignored)
        })
    }
}
```

### Cache Key Consistency with Graceful Parsing
```go
func TestCacheManager_GenerateKey_GracefulParsing_Consistency(t *testing.T) {
    manager := NewCacheManager()
    
    // These URLs should generate the same cache key because they resolve to same parameters
    tests := []struct {
        name string
        urls []string
        expectedSameKey bool
    }{
        {
            "same valid params with different invalid params",
            []string{
                "/img/photo.jpg/800x600/webp/q90",
                "/img/photo.jpg/800x600/webp/q90/invalid",
                "/img/photo.jpg/800x600/webp/q90/wow/extra",
            },
            true,
        },
        {
            "duplicate params resolve to same",
            []string{
                "/img/logo.png/300/png/q85",
                "/img/logo.png/300/400/png/jpeg/q85/q95",
            },
            true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var keys []string
            for _, url := range tt.urls {
                params := ParseImageRequest(url)
                key := manager.GenerateKey("test.jpg", params)
                keys = append(keys, key)
            }
            
            if tt.expectedSameKey {
                // All keys should be identical
                for i := 1; i < len(keys); i++ {
                    assert.Equal(t, keys[0], keys[i], 
                        "URLs should generate same cache key: %v", tt.urls)
                }
            }
        })
    }
}
```