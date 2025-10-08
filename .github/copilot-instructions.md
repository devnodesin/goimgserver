# Copilot Instructions for goimgserver

Welcome to the `goimgserver` project! This document provides essential guidelines for AI coding agents to be productive in this codebase. Follow these instructions to understand the architecture, workflows, and conventions of the project.

## Project Overview

`goimgserver` is a backend service for dynamic image resizing and optimization. It processes images based on URL parameters and serves them efficiently. Key features include:

- **Dynamic Resizing**: Supports dimensions like `{width}x{height}`.
- **Format Conversion**: Converts images to WebP, PNG, or JPEG.
- **Caching**: Resized images are cached for performance.
- **No UI**: Operates entirely through API endpoints.

### Stack
- **Go**: Core programming language.
- **bimg**: High-performance image processing library.
- **Gin Framework**: Lightweight web framework for routing.

## Key Files and Directories

- `src/main.go`: Entry point for the application.

## Developer Workflows

### Running the Application

1. Clone the repository and navigate to the `src` directory.
2. Run the application:
   ```bash
   go run main.go --port 9000 --imagesdir /path/to/images --cachedir /path/to/cache
   ```
3. Access endpoints, e.g.,
   ```bash
   curl -X GET "http://localhost:9000/img/sample.jpg/600x400"
   ```

### Testing

- Use `curl` or similar tools to test endpoints.
- Verify caching by inspecting the `/cache` directory.

### Deployment

- **Local**: Run with `go run main.go`.
- **Production**: Build with `go build` and manage using `systemd` or `supervisord`.

## API Endpoints

### Command Endpoints
- `POST /cmd/{name}`: Executes a specific command.
- `POST /cmd/clear`: Clears the cache directory.

### Image Endpoints
- `GET /img/{filename}/{width}x{height}`: Resizes the image.
- `GET /img/{filename}/{width}x{height}/{format}`: Resizes and converts the image format.
- `GET /img/{filename}/clear`: Clears the cached file.

## Project-Specific Conventions

- **Caching**: Always check the cache before processing an image.
- **Error Handling**: Return appropriate HTTP status codes for invalid requests (e.g., 400 for bad parameters).
- **Logging**: Use the configured log level for debugging and monitoring.

## Future Enhancements

- Support for additional formats like AVIF and TIFF.
- Authentication and rate limiting.
- Enhanced logging and monitoring.

## Notes for AI Agents

- Follow the `README.md` for setup and usage instructions.
- Ensure compliance with the command-line arguments when modifying image processing logic.
- Maintain the lightweight and efficient nature of the service.

---

Feel free to iterate on this document as the project evolves.