#!/bin/bash

# run_test.sh - Comprehensive test runner for goimgserver
# Execute all Go tests with detailed reporting and summaries

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
VERBOSE=false
SHORT=false
COVERAGE=false
BENCHMARKS=false
RACE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -v|--verbose)
      VERBOSE=true
      shift
      ;;
    -s|--short)
      SHORT=true
      shift
      ;;
    -c|--coverage)
      COVERAGE=true
      shift
      ;;
    -b|--bench)
      BENCHMARKS=true
      shift
      ;;
    -r|--race)
      RACE=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  -v, --verbose    Verbose test output"
      echo "  -s, --short      Run only short tests (skip long-running tests)"
      echo "  -c, --coverage   Generate coverage report"
      echo "  -b, --bench      Run benchmarks"
      echo "  -r, --race       Enable race detection"
      echo "  -h, --help       Show this help message"
      echo ""
      echo "Examples:"
      echo "  $0                    # Run all tests"
      echo "  $0 -s -v             # Run short tests with verbose output"
      echo "  $0 -c                # Run tests with coverage report"
      echo "  $0 -b                # Run benchmarks"
      echo "  $0 -r -v             # Run with race detection and verbose output"
      exit 0
      ;;
    *)
      echo "Unknown option $1"
      exit 1
      ;;
  esac
done

echo -e "${BLUE}=== goimgserver Test Suite ===${NC}"
echo "Starting comprehensive test execution..."
echo ""

# Check if we're in the right directory
if [ ! -f "src/go.mod" ]; then
  echo -e "${RED}Error: src/go.mod not found. Please run from the project root directory.${NC}"
  exit 1
fi

# Change to src directory where go.mod is located
cd src

# Build test flags
TEST_FLAGS=""
if [ "$VERBOSE" = true ]; then
  TEST_FLAGS="$TEST_FLAGS -v"
fi
if [ "$SHORT" = true ]; then
  TEST_FLAGS="$TEST_FLAGS -short"
fi
if [ "$RACE" = true ]; then
  TEST_FLAGS="$TEST_FLAGS -race"
fi

# Start timing
START_TIME=$(date +%s)

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

echo -e "${BLUE}--- Running Go Tests ---${NC}"

# Run tests with or without coverage
if [ "$COVERAGE" = true ]; then
  echo "Running tests with coverage analysis..."
  if go test $TEST_FLAGS -coverprofile=coverage.out ./...; then
    echo -e "${GREEN}✓ Tests completed successfully${NC}"
    
    # Generate coverage report
    echo ""
    echo -e "${BLUE}--- Coverage Report ---${NC}"
    go tool cover -func=coverage.out | tail -n 1
    
    # Generate HTML coverage report
    echo "Generating HTML coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report saved to: coverage.html"
    
  else
    echo -e "${RED}✗ Some tests failed${NC}"
    FAILED_TESTS=1
  fi
else
  echo "Running tests..."
  if go test $TEST_FLAGS ./...; then
    echo -e "${GREEN}✓ Tests completed successfully${NC}"
  else
    echo -e "${RED}✗ Some tests failed${NC}"
    FAILED_TESTS=1
  fi
fi

# Run benchmarks if requested
if [ "$BENCHMARKS" = true ]; then
  echo ""
  echo -e "${BLUE}--- Running Benchmarks ---${NC}"
  if go test -bench=. -benchmem ./performance ./precache ./resolver ./server/middleware; then
    echo -e "${GREEN}✓ Benchmarks completed successfully${NC}"
  else
    echo -e "${YELLOW}⚠ Some benchmarks failed or were skipped${NC}"
  fi
fi

# Calculate execution time
END_TIME=$(date +%s)
EXECUTION_TIME=$((END_TIME - START_TIME))

# Test summary
echo ""
echo -e "${BLUE}=== Test Summary ===${NC}"

# Extract test results (this is a simplified approach)
# In a real implementation, you might want to parse the actual test output
if [ $FAILED_TESTS -eq 0 ]; then
  echo -e "${GREEN}Status: PASSED${NC}"
else
  echo -e "${RED}Status: FAILED${NC}"
fi

echo "Execution time: ${EXECUTION_TIME}s"

# Show available test categories
echo ""
echo -e "${BLUE}--- Test Categories ---${NC}"
echo "• Unit Tests: Individual package functionality"
echo "• Integration Tests: End-to-end workflows"  
echo "• Performance Tests: Benchmarks and load testing"
echo "• Security Tests: Attack prevention and validation"

# Additional information
echo ""
echo -e "${BLUE}--- Additional Commands ---${NC}"
echo "View coverage in browser:"
echo "  go tool cover -html=coverage.out"
echo ""
echo "Run specific test packages:"
echo "  go test ./cache -v"
echo "  go test ./integration -v"
echo "  go test ./performance -bench=."
echo ""
echo "Run with different options:"
echo "  go test ./... -short     # Skip long-running tests"
echo "  go test ./... -race      # Enable race detection"
echo "  go test ./... -count=1   # Disable test caching"

# Cleanup
if [ -f "coverage.out" ] && [ "$COVERAGE" = false ]; then
  rm -f coverage.out
fi

# Exit with appropriate code
if [ $FAILED_TESTS -eq 0 ]; then
  echo ""
  echo -e "${GREEN}All tests passed! 🎉${NC}"
  exit 0
else
  echo ""
  echo -e "${RED}Some tests failed! ❌${NC}"
  echo "Please review the test output above for details."
  exit 1
fi