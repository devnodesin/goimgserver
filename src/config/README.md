# Configuration Package

This package provides command-line argument parsing and configuration management for goimgserver.

## Features

- Command-line argument parsing with sensible defaults
- Directory validation and automatic creation
- Port validation (1-65535)
- Default image detection and generation
- Configuration dump functionality

## Usage

### Command-Line Arguments

```bash
goimgserver [options]

Options:
  --port int          Server port (default: 9000)
  --imagesdir string  Images directory (default: ./images)
  --cachedir string   Cache directory (default: ./cache)
  --dump             Dump settings to settings.conf
```

### Examples

**Run with default settings:**
```bash
goimgserver
```

**Run with custom port:**
```bash
goimgserver --port 8080
```

**Run with custom directories:**
```bash
goimgserver --imagesdir /var/images --cachedir /var/cache
```

**Dump configuration to file:**
```bash
goimgserver --dump
# Creates settings.conf in current directory
```

## Default Image Management

The configuration system automatically manages a default image:

1. **Detection**: Scans for existing default image in priority order:
   - default.jpg
   - default.jpeg
   - default.png
   - default.webp

2. **Generation**: If no default image exists, generates a placeholder:
   - Dimensions: 1000x1000 pixels
   - Background: White (#FFFFFF)
   - Text: "goimgserver" (centered, black)
   - Format: JPEG (quality: 95)
   - Saved as: {imagesdir}/default.jpg

3. **Validation**: Ensures the default image is readable and processable

## Testing

The package follows Test-Driven Development (TDD) with comprehensive test coverage:

```bash
cd src/config
go test -v -cover
```

**Coverage: 96.2%** (exceeds the >95% requirement)

### Test Categories

- **Argument Parsing**: Default values, custom values, flags
- **Port Validation**: Valid range (1-65535), invalid ports
- **Directory Management**: Existing dirs, automatic creation
- **Default Image**: Detection, generation, validation
- **Settings Dump**: File creation, content verification
- **Error Handling**: Invalid paths, corrupted images, unwritable dirs

## Architecture

### Core Types

```go
type Config struct {
    Port             int
    ImagesDir        string
    CacheDir         string
    Dump             bool
    DefaultImagePath string
}
```

### Key Functions

- `ParseArgs(args []string) (*Config, error)`: Parse command-line arguments
- `Validate() error`: Validate configuration and create directories
- `SetupDefaultImage() error`: Detect or generate default image
- `DumpSettings(filename string) error`: Write configuration to file
- `String() string`: String representation of configuration

## Implementation Details

Built following TDD principles:
1. ✅ Write failing tests (RED)
2. ✅ Implement minimal code to pass (GREEN)
3. ✅ Refactor while keeping tests green (REFACTOR)

All 22 tests pass with 96.2% code coverage.
