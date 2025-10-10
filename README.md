# goimgserver

goimgserver - These services allow you to store images, then serve optimized, resized, and converted images on the fly based on URL parameters.

## Features

- **Dynamic Image Resizing**: Resize images on-the-fly with URL parameters
- **Format Conversion**: Convert images to WebP, PNG, or JPEG
- **Caching**: Automatic caching of processed images for performance
- **Pre-cache System**: Startup pre-caching of images for improved initial response times
- **Command Endpoints**: Administrative commands for cache management and Git operations

## Usage

1. **Setup**:

```bash
sudo apt-get update && sudo apt-get install -y libvips-dev
cd src
go build
```

```bash
cd src
go run main.go
```

All parameters are optional: if not specified, they will start with default values:

- `--port XXXX` defaults to `9000`
- `--imagesdir /path/to/images` defaults to `{pwd}/images`
- `--cachedir /path/to/cache` defaults to `{pwd}/cache`
- `--precache` defaults to `true` (enables pre-caching on startup)
- `--precache-workers N` defaults to `0` (auto, uses CPU count)

2. **Access the endpoints**:

### Image Endpoints

```bash
# Get image with default settings
curl -X GET "http://localhost:9000/img/sample.jpg"

# Resize image
curl -X GET "http://localhost:9000/img/sample.jpg/600x400"

# Resize and convert format
curl -X GET "http://localhost:9000/img/sample.jpg/600x400/webp"
```

### Command Endpoints

```bash
# Clear entire cache
curl -X POST "http://localhost:9000/cmd/clear"

# Update images from git (if images directory is a git repository)
curl -X POST "http://localhost:9000/cmd/gitupdate"
```

- Use tools like `curl` or a browser to test the endpoints.

## Command Endpoints

The server provides administrative command endpoints for maintenance operations:

### POST /cmd/clear
Clears the entire cache directory and returns statistics about freed space.

**Example Response:**
```json
{
  "success": true,
  "message": "Cache cleared successfully",
  "cleared_files": 1234,
  "freed_space": "2.5GB"
}
```

### POST /cmd/gitupdate
Updates the images directory via `git pull` if it's a git repository.

**Example Response (Success):**
```json
{
  "success": true,
  "message": "Git update completed",
  "changes": 5,
  "branch": "main",
  "last_commit": "abc123..."
}
```

**Example Response (Not a Git Repo):**
```json
{
  "success": false,
  "error": "Images directory is not a git repository",
  "code": "GIT_NOT_FOUND"
}
```

### POST /cmd/:name
Generic command router that dispatches to specific command handlers.

**Valid Commands:** `clear`, `gitupdate`

## Pre-cache System

The server includes an automatic pre-caching system that runs during startup to improve initial response times.

### How It Works

When the server starts, it:
1. Scans the images directory for all supported image files (JPEG, PNG, WebP)
2. Creates default cached versions of each image:
   - **Dimensions**: 1000x1000 pixels
   - **Format**: WebP
   - **Quality**: 95
3. Stores pre-cached images in the cache directory
4. Skips already cached images to avoid redundant work

### Configuration

```bash
# Enable pre-cache (default)
go run main.go --precache=true

# Disable pre-cache
go run main.go --precache=false

# Set number of workers (0 = auto, uses CPU count)
go run main.go --precache-workers=8
```

### Features

- **Async Execution**: Runs asynchronously to not block server startup
- **Concurrent Processing**: Uses worker pools for fast processing
- **Progress Logging**: Real-time progress updates in logs
- **Error Handling**: Gracefully handles corrupted or missing images
- **Smart Exclusion**: Skips system default images and already cached files

For more details, see [Pre-cache Package Documentation](src/precache/README.md).

## Testing

Run the comprehensive test suite:

```bash
# Run all tests with summary
./run_test.sh

# Run tests with coverage report
./run_test.sh --coverage

# Run only short tests (faster)
./run_test.sh --short

# Run with verbose output
./run_test.sh --verbose

# Run benchmarks
./run_test.sh --bench
```

Alternatively, run tests directly:

```bash
cd src
go test ./...
```

Test command endpoints (requires server to be running):

```bash
./test_commands.sh
```

For detailed testing information, see the [Testing Guide](docs/testing.md).

## Documentation

### Quick Start Guides

- [API Documentation](docs/api/README.md) - Complete API reference with examples
- [Basic Usage](docs/api/examples/basic_usage.md) - Getting started with image processing
- [Advanced Usage](docs/api/examples/advanced_usage.md) - Advanced patterns and optimization

### Deployment

- [Deployment Guide](docs/deployment/README.md) - Complete deployment instructions
  - [Systemd Deployment](docs/deployment/systemd/) - Linux service installation
  - [Docker Deployment](docs/deployment/docker/) - Container-based deployment
  - [Nginx Configuration](docs/deployment/nginx/) - Reverse proxy setup

### Operations

- [Performance Guide](docs/performance/README.md) - Optimization and tuning
- [Security Guide](docs/security/README.md) - Security hardening and best practices
- [Troubleshooting Guide](docs/troubleshooting/README.md) - Common issues and solutions

### Reference

- [OpenAPI Specification](docs/api/openapi.yaml) - Machine-readable API specification
- [Design Documents](design/) - Architecture and design decisions

## Quick Examples

### Basic Image Operations

```bash
# Get original image
curl http://localhost:9000/img/photo.jpg -o photo.jpg

# Thumbnail (150x150)
curl http://localhost:9000/img/photo.jpg/150x150 -o thumb.jpg

# Medium size with WebP format
curl http://localhost:9000/img/photo.jpg/800x600/webp -o medium.webp

# Custom quality
curl "http://localhost:9000/img/photo.jpg/1920x1080/webp?quality=90" -o hd.webp
```

### HTML Integration

```html
<!-- Simple responsive image -->
<img 
    src="http://localhost:9000/img/photo.jpg/800x600/webp"
    srcset="
        http://localhost:9000/img/photo.jpg/400x300/webp 400w,
        http://localhost:9000/img/photo.jpg/800x600/webp 800w,
        http://localhost:9000/img/photo.jpg/1920x1080/webp 1920w
    "
    sizes="(max-width: 600px) 400px, (max-width: 1200px) 800px, 1920px"
    alt="Photo">
```

### Maintenance Operations

```bash
# Check server health
curl http://localhost:9000/health

# Clear cache
curl -X POST http://localhost:9000/cmd/clear

# Update images from Git
curl -X POST http://localhost:9000/cmd/gitupdate
```

## Production Deployment

### Quick Start with Docker

```bash
# Clone repository
git clone https://github.com/devnodesin/goimgserver.git
cd goimgserver

# Prepare images
mkdir -p docs/deployment/docker/images
cp /path/to/your/images/* docs/deployment/docker/images/

# Start with Docker Compose
cd docs/deployment/docker
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f
```

### Quick Start with Systemd

```bash
# Build application
cd src
go build -o goimgserver main.go

# Install as service
cd ../docs/deployment/systemd
sudo cp goimgserver /opt/goimgserver/bin/
sudo ./install.sh

# Start service
sudo systemctl start goimgserver
sudo systemctl status goimgserver
```

## Performance

### Benchmarks

Typical performance on modern hardware (8 cores, 16GB RAM, NVMe SSD):

- **Cached Requests**: 800+ requests/sec, 2-5ms latency
- **Processing**: 150+ requests/sec, 50-200ms latency
- **Cache Hit Rate**: 80-95% (typical workload)

See the [Performance Guide](docs/performance/README.md) for optimization tips.

## Security

Key security features:

- ✓ Rate limiting (100 requests/min per IP)
- ✓ Input validation (path traversal prevention)
- ✓ Resource limits (max dimensions, quality)
- ✓ Security headers (X-Frame-Options, CSP)
- ✓ Access control for admin endpoints

See the [Security Guide](docs/security/README.md) for hardening instructions.

## Contributing

Contributions are welcome! Please see the [design documents](design/) for architecture details and follow TDD methodology outlined in [design/01-tdd-methodology.md](design/01-tdd-methodology.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

