# goimgserver

goimgserver - These services allow you to store images, then serve optimized, resized, and converted images on the fly based on URL parameters.

## Features

- **Dynamic Image Resizing**: Resize images on-the-fly with URL parameters
- **Format Conversion**: Convert images to WebP, PNG, or JPEG
- **Caching**: Automatic caching of processed images for performance
- **Command Endpoints**: Administrative commands for cache management and Git operations

## Usage

1. **Setup**:

```bash
cd src
go run main.go
```

All parameters are optional: if not specified, they will start with default values:

- `--port XXXX` defaults to `9000`
- `--imagesdir /path/to/images` defaults to `{pwd}/images`
- `--cachedir /path/to/cache` defaults to `{pwd}/cache`

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

## Testing

Run the test suite:

```bash
cd src
go test ./...
```

Test command endpoints (requires server to be running):

```bash
./test_commands.sh
```

## Documentation

- [Complete Design](design/complete_design.md)
- [API Specification](design/06-api-specification.md)
- [Security](design/07-security.md)
- [Command Implementation](COMMAND_IMPLEMENTATION.md)

