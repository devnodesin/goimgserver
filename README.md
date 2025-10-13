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
make build
```

```bash
cd src
make run
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

## Testing

Run the comprehensive test suite:

```bash
# Run all tests with summary
make test

# Run tests with coverage report
cd src
./run_test.sh --coverage

# Run only short tests (faster)
cd src
./run_test.sh --short

# Run with verbose output
cd src
./run_test.sh --verbose

# Run benchmarks
cd src
./run_test.sh --bench
```

Alternatively, run tests directly:

```bash
cd src
make test
```

Test command endpoints (requires server to be running):

```bash
cd test
./test_commands.sh
./test_server_middleware.sh
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
make build

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

