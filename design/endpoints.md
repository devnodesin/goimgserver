# Endpoints

## Command Endpoints

- `POST /cmd/{name}`: Executes a specific command based on the provided name.
- `POST /cmd/clear`: Clears the entire cache directory.
- `POST /cmd/gitupdate`: run `git update` inside image_dir, (only image_dir is a git dir)

## Image Endpoints

- `GET /img/{filename}`: Retrieves the image with default settings (1000x1000px, q75, WebP).

### Custom Resized Images

- `GET /img/{filename}/600x400`: Retrieves an image resized to 600x400 pixels.
- `GET /img/{filename}/400`: Retrieves an image resized to 400 pixels in width, maintaining aspect ratio.

### Quality Settings

- `GET /img/{filename}/q50`: Retrieves the image with quality set to 50 (other settings default).
- `GET /img/{filename}/150/q50`: Retrieves the image resized to 150px width and quality set to 50.

### Cache Management

- `GET /img/{filename}/clear`: Clears the all cached file for the specified filename.

### Image Formats

- `GET /img/{filename}/600x400/png`: Retrieves an image resized to 600x400 pixels and converted to PNG format.
- `GET /img/{filename}/200x300/webp`: Retrieves an image resized to 200x300 pixels in WebP format.
- `GET /img/{filename}/100x100/jpeg`: Retrieves an image resized to 100x100 pixels in JPEG format.
