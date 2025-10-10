# Testing Guide

## Overview

This guide covers the essential testing information for goimgserver. The project follows Test-Driven Development (TDD) principles with comprehensive test coverage.

## Test Structure

```
src/
├── testutils/              # Test utilities and helpers
├── integration/            # Integration test suite  
├── performance/            # Performance and load tests
└── security/              # Security tests
```

## Running Tests

### Quick Tests
```bash
# Run all tests in short mode
go test ./... -short

# Run with coverage
go test ./... -short -cover
```

### Full Test Suite
```bash
# Run all tests including long-running tests
go test ./...

# Run with verbose output and race detection
go test ./... -v -race
```

### Performance Tests
```bash
# Run benchmarks
go test ./performance -bench=.

# Run load tests 
go test ./performance -v -run=TestPerformance
```

## Test Categories

### Unit Tests
- Individual package functionality
- Located in package test files (`*_test.go`)
- High coverage requirements (>90%)

### Integration Tests  
- End-to-end workflows
- Component interaction testing
- Located in `src/integration/`

### Performance Tests
- Benchmarks and load testing
- Located in `src/performance/`
- Baseline performance metrics

### Security Tests
- Attack prevention testing
- Input validation testing
- Located in `src/security/`

## Coverage Goals

- **Overall Target**: >95% coverage
- **Core Packages**: >90% coverage
- **Test Utilities**: >75% coverage

## Best Practices

1. **Follow TDD**: Write tests before implementation
2. **Test Error Cases**: Don't just test happy paths
3. **Use Descriptive Names**: Clear test scenarios
4. **Keep Tests Independent**: No cross-test dependencies
5. **Clean Up Resources**: Use `t.TempDir()` and `t.Cleanup()`

## Continuous Integration

The test suite supports CI/CD with:
- Fast test mode for quick feedback
- Full test suite for comprehensive validation
- Coverage reporting and metrics
- Performance regression detection

For detailed information, see the comprehensive testing documentation in the design documents.