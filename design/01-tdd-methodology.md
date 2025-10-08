# Test-Driven Development (TDD) Methodology

## Overview

goimgserver follows strict Test-Driven Development practices to ensure high code quality, maintainable architecture, and reliable functionality. This document outlines the TDD methodology requirements for all development work.

## TDD Principles

### Red-Green-Refactor Cycle

**MANDATORY**: All code must be developed following the Red-Green-Refactor cycle:

1. **Red**: Write a failing test that defines the desired behavior
   - Test must fail for the right reason (functionality doesn't exist yet)
   - Test should be specific and focused on one behavior
   - Write the minimal test that captures the requirement

2. **Green**: Write the minimal code to make the test pass
   - Implement only what's needed to make the test pass
   - Don't over-engineer or add unnecessary features
   - Focus on making tests pass quickly

3. **Refactor**: Improve the code while keeping tests green
   - Clean up code structure and design
   - Extract common functionality
   - Ensure all tests continue to pass
   - Refactor both production and test code

### Core TDD Rules

- **Never write production code without a failing test**
- **Write the smallest possible test that fails**
- **Write the smallest amount of production code to make the test pass**
- **Run tests frequently** (after every small change)
- **Keep test cycles short** (minutes, not hours)
- **One failing test at a time** - fix before moving to the next

## Test Organization

### Test Structure

```go
// Test function naming: Test_<FunctionName>_<Scenario>
func Test_ImageProcessor_Resize_ValidDimensions(t *testing.T) {
    // Arrange
    processor := NewImageProcessor()
    input := []byte("mock image data")
    
    // Act
    result, err := processor.Resize(input, 100, 200)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Test Categories

#### Unit Tests
- Test individual functions and methods in isolation
- Use mocks and stubs for dependencies
- Focus on single responsibility
- Name: `*_test.go` in same package

#### Integration Tests
- Test component interactions
- Use real dependencies where appropriate
- Test complete workflows
- Name: `*_integration_test.go`

#### End-to-End Tests
- Test complete user scenarios
- Use actual HTTP requests
- Test with real files and cache
- Name: `*_e2e_test.go`

### Test Helpers

```go
func setupTestEnvironment(t *testing.T) (string, string) {
    t.Helper() // Mark as helper function
    
    tempDir := t.TempDir() // Automatic cleanup
    imageDir := filepath.Join(tempDir, "images")
    cacheDir := filepath.Join(tempDir, "cache")
    
    os.MkdirAll(imageDir, 0755)
    os.MkdirAll(cacheDir, 0755)
    
    return imageDir, cacheDir
}
```

## Testing Requirements

### Coverage Requirements

- **Minimum 90% test coverage** across all packages
- **95% coverage for critical components**:
  - Image processing
  - Cache management
  - URL parsing
  - File resolution
- Use `go test -coverprofile=coverage.out` to measure coverage

### Test Types Required

#### For Each Component:
- [ ] Unit tests for all public functions
- [ ] Error case testing (invalid inputs, system failures)
- [ ] Edge case testing (boundary values, empty inputs)
- [ ] Concurrent access testing (where applicable)
- [ ] Performance benchmarks for critical paths

#### Component-Specific Requirements:

**Image Processing:**
- [ ] Format conversion tests
- [ ] Dimension validation tests
- [ ] Quality parameter tests
- [ ] Corruption handling tests

**Cache Management:**
- [ ] Cache hit/miss scenarios
- [ ] Concurrent access safety
- [ ] Cache invalidation tests
- [ ] Storage failure handling

**URL Parsing:**
- [ ] Valid parameter combinations
- [ ] Invalid parameter handling
- [ ] Graceful parsing scenarios
- [ ] Parameter precedence rules

**File Resolution:**
- [ ] Extension auto-detection
- [ ] Grouped image resolution
- [ ] Extension priority order
- [ ] Missing file scenarios

## Test Implementation Standards

### Arrange-Act-Assert Pattern

```go
func Test_CacheManager_Store_ValidData(t *testing.T) {
    // Arrange - Set up test data and dependencies
    manager := NewCacheManager(t.TempDir())
    testData := []byte("test image data")
    key := "test-key"
    
    // Act - Execute the function under test
    err := manager.Store(key, testData)
    
    // Assert - Verify the results
    assert.NoError(t, err)
    stored, exists := manager.Retrieve(key)
    assert.True(t, exists)
    assert.Equal(t, testData, stored)
}
```

### Table-Driven Tests

Use for multiple scenarios with similar logic:

```go
func Test_ParseDimensions_Various(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantW    int
        wantH    int
        wantErr  bool
    }{
        {"valid dimensions", "800x600", 800, 600, false},
        {"width only", "400", 400, 400, false},
        {"invalid format", "invalid", 0, 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w, h, err := ParseDimensions(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.wantW, w)
            assert.Equal(t, tt.wantH, h)
        })
    }
}
```

### Mock Usage

Use dependency injection and interfaces for testability:

```go
type ImageProcessor interface {
    Process(data []byte, params ProcessParams) ([]byte, error)
}

type MockImageProcessor struct {
    ProcessFunc func([]byte, ProcessParams) ([]byte, error)
}

func (m *MockImageProcessor) Process(data []byte, params ProcessParams) ([]byte, error) {
    if m.ProcessFunc != nil {
        return m.ProcessFunc(data, params)
    }
    return data, nil // Default behavior
}
```

## Performance Testing

### Benchmarks

Write benchmarks for performance-critical code:

```go
func Benchmark_ImageProcessor_Resize_SmallImage(b *testing.B) {
    processor := NewImageProcessor()
    testData := loadTestImage("small.jpg")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := processor.Resize(testData, 200, 200)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Load Testing

Test concurrent scenarios:

```go
func Test_CacheManager_Concurrent_Access(t *testing.T) {
    manager := NewCacheManager(t.TempDir())
    const goroutines = 100
    
    var wg sync.WaitGroup
    wg.Add(goroutines)
    
    for i := 0; i < goroutines; i++ {
        go func(id int) {
            defer wg.Done()
            key := fmt.Sprintf("key-%d", id)
            data := []byte(fmt.Sprintf("data-%d", id))
            
            err := manager.Store(key, data)
            assert.NoError(t, err)
            
            retrieved, exists := manager.Retrieve(key)
            assert.True(t, exists)
            assert.Equal(t, data, retrieved)
        }(i)
    }
    
    wg.Wait()
}
```

## Test Development Workflow

### Development Process

1. **Write failing test first**
   ```bash
   go test -v ./package -run TestNewFeature
   # Should fail with clear error message
   ```

2. **Implement minimal code**
   ```bash
   # Add just enough code to make test pass
   go test -v ./package -run TestNewFeature
   # Should pass
   ```

3. **Refactor with safety**
   ```bash
   # Clean up code while keeping tests green
   go test -v ./package
   # All tests should still pass
   ```

4. **Verify coverage**
   ```bash
   go test -coverprofile=coverage.out ./package
   go tool cover -html=coverage.out
   ```

### Continuous Testing

Set up test automation:

```bash
# Pre-commit hook
#!/bin/bash
go test -v ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
```

## Test Environment

### Test Data Management

```go
func setupTestData(t *testing.T) string {
    t.Helper()
    
    testDataDir := filepath.Join("testdata", t.Name())
    if !fileExists(testDataDir) {
        t.Skipf("Test data directory not found: %s", testDataDir)
    }
    
    return testDataDir
}
```

### Cleanup

```go
func Test_Feature(t *testing.T) {
    // Use t.TempDir() for automatic cleanup
    tempDir := t.TempDir()
    
    // Or manual cleanup
    cleanup := setupComplexEnvironment()
    defer cleanup()
}
```

## Quality Gates

### Before Code Review

- [ ] All tests pass (`go test -v ./...`)
- [ ] Race condition testing passes (`go test -race ./...`)
- [ ] Coverage requirements met
- [ ] Benchmarks show acceptable performance
- [ ] No test flakiness (run tests multiple times)

### Integration Requirements

- [ ] Tests run in CI/CD pipeline
- [ ] Coverage reports generated
- [ ] Performance regressions detected
- [ ] Test results visible in pull requests

## Best Practices

### Do's

- ✅ Write tests before implementation
- ✅ Use descriptive test names
- ✅ Test one thing at a time
- ✅ Use table-driven tests for multiple scenarios
- ✅ Mock external dependencies
- ✅ Test error cases thoroughly
- ✅ Use `t.Helper()` for test utilities
- ✅ Clean up resources properly

### Don'ts

- ❌ Skip writing tests for "simple" code
- ❌ Write tests after implementation
- ❌ Test implementation details
- ❌ Use global state in tests
- ❌ Ignore test flakiness
- ❌ Mock everything (test real behavior when reasonable)
- ❌ Write overly complex tests
- ❌ Duplicate test logic

## Tools and Libraries

### Testing Framework
- **Standard library**: `testing` package
- **Assertions**: `github.com/stretchr/testify/assert`
- **Mocking**: `github.com/stretchr/testify/mock`
- **HTTP testing**: `net/http/httptest`

### Coverage and Quality
- **Coverage**: `go test -cover`
- **Race detection**: `go test -race`
- **Linting**: `golangci-lint`
- **Benchmarking**: `go test -bench`

This TDD methodology ensures that goimgserver maintains high code quality, reliability, and maintainability throughout development.