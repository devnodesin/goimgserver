# Makefile for goimgserver

# Variables
BINARY_NAME=goimgserver
SRC_DIR=.
TEST_SCRIPT=run_test.sh

# Default target
all: build

# Build the binary
build:
	@echo "Building the binary..."
	go build -o $(BINARY_NAME) $(SRC_DIR)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	./$(TEST_SCRIPT)

# Clean up
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Help
.PHONY: help
help:
	@echo "Makefile for goimgserver"
	@echo "Targets:"
	@echo "  all       - Build the binary (default)"
	@echo "  build     - Build the binary"
	@echo "  test      - Run tests using run_test.sh"
	@echo "  clean     - Clean up generated files"
	@echo "  help      - Show this help message"