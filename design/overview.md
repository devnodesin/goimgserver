## goimgserver

**_goimgserver_** is a backend service designed to store images and serve optimized, resized, and converted images dynamically based on URL parameters. It is built for high performance, scalability, and ease of use, with a focus on caching and efficient image processing.

## Documentation Placeholders

For consistency across all documentation, the following placeholders are used:

- **{image_dir}** - Image directory path (default: `./images`)
- **{cache_dir}** - Cache directory path (default: `./cache`)
- **{default_image}** - Default placeholder image file
- **{filename}** - Requested image filename
- **{width}** - Image width in pixels
- **{height}** - Image height in pixels
- **{format}** - Image format (webp, png, jpeg, jpg)
- **{quality}** - Image quality (1-100)
- **{hash}** - Cache key hash
- **{pwd}** - Present working directory

## Features

- **Backend only**: Dynamic image resizing service.
- **Resize**: Supports dimensions like `{width}x{height}` (e.g., 400x200). Default: 1000x1000 px. Maximum: 4000 px, Minimum: 10 px.
- **Format**: Supports WebP (default), PNG, JPEG.
- **Quality**: Adjustable quality via `q{quality}` (e.g., `q95`). Default: `q75`.
- **Caching**: Resized images are cached to accelerate delivery and improve performance. Cached files are served directly when available.
- **Default Image Fallback**: When requested image is not found, serves {default_image} with same processing parameters.
- **Automatic Placeholder Generation**: If {default_image} is missing, generates a programmatic placeholder (1000x1000px, white background, "goimgserver" text).
- **Performance**: Fast and secure for dynamic URL-based image processing.
- **CORS-based access**: Enables cross-origin resource sharing.
- **No UI**: Only returns the image when the URL is accessed.

## Request Flow

The following steps outline the process from receiving a request to serving the image:

- **Request Received**: The server receives an HTTP request with parameters specifying the image filename, dimensions, format, and quality.
- **Cache Lookup**: The server checks {cache_dir}/{filename}/{hash} for a pre-processed image matching the request parameters.
  - If a cached image is found, it is served immediately.
  - If no cached image is found, the request proceeds to the processing stage.
- **Image Processing**:
  - The original image is retrieved from {image_dir}.
  - **If image not found**: Falls back to {default_image} from {image_dir}/default.{jpg/jpeg/png/webp}
  - **If {default_image} not found**: Uses programmatically generated placeholder
  - The image (original, default, or generated) is resized, converted, and optimized based on the request parameters using libvips via bimg.
- **Cache Storage**:
  - The processed image is stored in {cache_dir}/{filename}/{hash} for future requests.
- **Response**:
  - The processed image is returned to the client with appropriate HTTP headers for caching and content type.

This flow ensures high performance by prioritizing cache hits and providing graceful fallback to {default_image} when requested images are missing.

## Application Startup

- **Load Configuration**: The application loads configuration from command-line arguments. If no arguments are provided, default settings are used.
- **Directory Checks**: The application ensures that all required directories (e.g., {cache_dir}, {image_dir}) exist. If the directories are missing, they are created automatically during startup.
- **Default Image Setup**:
  - Check for {default_image} in {image_dir} (looks for default.jpg, default.jpeg, default.png, default.webp)
  - If {default_image} not found, generate programmatic placeholder:
    - Dimensions: 1000x1000px
    - Background: White
    - Text: "goimgserver" (centered, black font, sufficient size)
    - Format: JPEG
    - Save as {image_dir}/default.jpg
- **Pre-cache Initialization**:
  - The application scans {image_dir} for original images.
  - For each image, a cached version is created using the default settings:
    - Dimensions: `1000x1000`
    - Format: `webp`
    - Quality: `95`
  - Cached images are stored in {cache_dir}/{filename}/{hash} directory structure.
- **Server Initialization**:
  - The HTTP server is started and begins listening for incoming requests.
- **Error Handling**:
  - Any critical errors during startup (e.g., missing permissions, invalid configuration) are logged, and the application exits gracefully.

This startup process ensures that the application is fully prepared to handle requests efficiently and always has a {default_image} available for fallback scenarios.

### Future Enhancements

- Add support for additional image formats (e.g., AVIF, TIFF).
- Implement authentication for secure access.
- Introduce rate limiting to prevent abuse.
- Enhance logging and monitoring capabilities.
- Add support for image watermarking.
- Provide detailed metrics for image processing performance.
- Add support for image transformations (rotation, cropping, filters).
- Implement CDN integration and advanced caching strategies.
- Add support for progressive JPEG and optimized delivery.

## Stack

- **Go**: Core programming language for the service.
- **bimg**: High-performance library for image processing.
- **Gin framework**: Lightweight and fast web framework for API and routing.
- **Git**: Used for syncing and managing original image files.

## Development Methodology

**goimgserver follows Test-Driven Development (TDD) principles:**

- **Test-First Development**: All functionality must be developed using the Red-Green-Refactor TDD cycle
- **Comprehensive Testing**: Unit tests, integration tests, and performance benchmarks are mandatory
- **Test Coverage**: Maintain >90% test coverage across all components
- **Continuous Testing**: Tests run on every change and before every commit

See [`design/tdd-methodology.md`](tdd-methodology.md) for detailed TDD implementation guidelines.

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
--imagesdir   (default: ./images)     Directory for storing original images ({image_dir}).
--cachedir    (default: ./cache)      Directory for storing cached images ({cache_dir}).
--dump        (optional)              dumps the current settings in {pwd} named settings.conf for debugging purposes.
```

{pwd} refers to the current working directory.

## References

- Usage: See [`README.md`](../README.md)
- Endpoints: See [`design/endpoints.md`](endpoints.md)
- URL Parsing: See [`design/url-parsing.md`](url-parsing.md)
- Default Image System: See [`design/default-image.md`](default-image.md)
- API Specification: See [`design/api-specification.md`](api-specification.md)
- Security: See [`design/security.md`](security.md)
- TDD Methodology: See [`design/tdd-methodology.md`](tdd-methodology.md)
- Deployment: See [`design/deployment.md`](deployment.md)
