# goimgserver

goimgserver - These services allow you to store images, then serve optimized, resized, and converted images on the fly based on URL parameters.

## Features

- **Backend only** - dynamic image resizing service
- **Resize**: Supports dimensions like `{width}x{height}` (e.g: 400x200), 1000x1000 px (default), Max: 4000 px, min: 10 px
- **Format**: Supports WebP (default), PNG, JPEG
- **Quality**: Adjustable quality qXX (e.g., q95), q75 (default)
- **Caching**: Resized images are cached to accelerate delivery and improve performance. Directly serves files from cache when available.
- **Performance**: Fast and secure for dynamic URL-based image processing
- **CORS-based access**: Ensures cross-origin resource sharing
- **No UI**: Only returns the image when the URL is accessed

### Future Enhancements

- Add support for additional image formats (e.g., AVIF, TIFF)
- Implement authentication for secure access
- Introduce rate limiting to prevent abuse
- Enhance logging and monitoring capabilities
- Add support for image watermarking
- Provide detailed metrics for image processing performance

## Stack

- **Go**: The core programming language used for building the service
- **bimg**: High-performance library for image processing
- **Gin framework**: Lightweight and fast web framework for API and routing
- **Git**: Used for syncing and managing original image files

## Folder Structure

```plaintext
/src
  main.go        ← Entry point for the application
```

## Endpoints

### Command Endpoints

- `POST /cmd/{name}`: Executes a specific command based on the provided name
- `POST /cmd/clear` clears the entire cache directory

### Image Endpoints

- `GET /img/{filename}/600x400`: Retrieves an image resized to 600x400 pixels
- `GET /img/{filename}/400`: Retrieves an image resized to 400 pixels in width, maintaining aspect ratio
- `GET /img/{filename}/600x400/png`: Retrieves an image resized to 600x400 pixels and converted to PNG format
- `GET /img/{filename}/clear`: Clears the cached file for the specified filename
- **More examples**:
  - `GET /img/{filename}/200x300/webp`: Retrieves an image resized to 200x300 pixels in WebP format
  - `GET /img/{filename}/100x100/jpeg`: Retrieves an image resized to 100x100 pixels in JPEG format

## Command-Line Arguments

The application now uses command-line arguments for configuration. The following options are available:

- `--port` (default: `9000`): Specifies the port on which the server runs.
- `--imagesdir` (default: `pwd/images`): Specifies the directory for storing original images.
- `--cachedir` (default: `pwd/cache`): Specifies the directory for storing cached images.
- `--dump` (optional): Outputs the current settings for debugging purposes.

## Usage

1. **Setup**:

   - Clone the repository and navigate to the `src` directory.

   ```bash
   go run main.go --port 9000 --imagesdir /path/to/images --cachedir /path/to/cache
   ```

2. **Access the endpoints**:

   ```bash
   curl -X GET "http://localhost:8080/img/sample.jpg/600x400"
   ```

3. **Testing**:
   - Use tools like curl to test the endpoints.
   - Verify caching by checking the `/cache` directory for processed images.

## Deployment

- **Local Deployment**:
  - Ensure Go is installed on the system.
  - Run the application using `go run main.go`.
- **Production Deployment**:
  - Build the application using `go build`.
  - Use a process manager like `systemd` or `supervisord` to manage the application.



