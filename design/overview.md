## goimgserver

goimgserver is a backend service that stores images and serves optimized, resized, and converted images on the fly based on URL parameters.

## Features

- **Backend only**: Dynamic image resizing service.
- **Resize**: Supports dimensions like `{width}x{height}` (e.g., 400x200). Default: 1000x1000 px. Maximum: 4000 px, Minimum: 10 px.
- **Format**: Supports WebP (default), PNG, JPEG.
- **Quality**: Adjustable quality via `qXX` (e.g., `q95`). Default: `q75`.
- **Caching**: Resized images are cached to accelerate delivery and improve performance. Cached files are served directly when available.
- **Performance**: Fast and secure for dynamic URL-based image processing.
- **CORS-based access**: Enables cross-origin resource sharing.
- **No UI**: Only returns the image when the URL is accessed.

## Request Flow

The following steps outline the process from receiving a request to serving the image:

- **Request Received**: The server receives an HTTP request with parameters specifying the image filename, dimensions, format, and quality.
- **Cache Lookup**: The server checks the cache directory (`/cache/{filename}/{hash}`) for a pre-processed image matching the request parameters.
  - If a cached image is found, it is served immediately.
  - If no cached image is found, the request proceeds to the processing stage.
- **Image Processing**:
  - The original image is retrieved from the images directory (managed via Git).
  - The image is resized, converted, and optimized based on the request parameters using libvips via bimg.
- **Cache Storage**:
  - The processed image is stored in the cache directory under `/cache/{filename}/{hash}` for future requests.
- **Response**:
  - The processed image is returned to the client with appropriate HTTP headers for caching and content type.

This flow ensures high performance by prioritizing cache hits and minimizing redundant processing.

## Application Startup

- **Load Configuration**: The application loads configuration from command-line arguments. If no arguments are provided, default settings are used.
- **Directory Checks**: The application ensures that all required directories (e.g., `{cache_dir}`, `{image_dir}`) exist. If the directoires are missing, it is created automatically during startup.
- **Pre-cache Initialization**:
  - The application scans the images directory for original images.
  - For each image, a cached version is created using the default settings:
    - Dimensions: `1000x1000`
    - Format: `webp`
    - Quality: `95`
  - Cached images are stored in the `cache/{filename}/{hash}` directory structure.
- **Server Initialization**:
  - The HTTP server is started and begins listening for incoming requests.
- **Error Handling**:
  - Any critical errors during startup (e.g., missing permissions, invalid configuration) are logged, and the application exits gracefully.

This startup process ensures that the application is fully prepared to handle requests efficiently from the moment it begins running.

### Future Enhancements

- Add support for additional image formats (e.g., AVIF, TIFF).
- Implement authentication for secure access.
- Introduce rate limiting to prevent abuse.
- Enhance logging and monitoring capabilities.
- Add support for image watermarking.
- Provide detailed metrics for image processing performance.

## Stack

- **Go**: Core programming language for the service.
- **bimg**: High-performance library for image processing.
- **Gin framework**: Lightweight and fast web framework for API and routing.
- **Git**: Used for syncing and managing original image files.

## Folder Structure

```plaintext
/design          ← Software design documents and specifications
/src
  main.go        ← Entry point for the application
```

## Endpoints

See [`design/endpoints.md`](design/endpoints.md) for full details.

- **Image endpoints**: Resize, convert, and serve images via URL parameters.
- **Command endpoints**: Perform cache management and other administrative actions.

pwd = current working directory

## Command-Line Arguments

The application uses command-line arguments for configuration:

```plaintext
--port        (default: 9000)         Specifies the port on which the server runs.
--imagesdir   (default: ./images)     Directory for storing original images.
--cachedir    (default: ./cache)      Directory for storing cached images.
--dump        (optional)              dumps the current settings in pwd named settings.conf for debugging purposes.
```

`pwd` refers to the current working directory.

## References

- Usage: See [`README.md`](../README.md)
- Endpoints: See [`design/endpoints.md`](endpoints.md)
- Deployment: See [`design/deployment.md`](deployment.md)
