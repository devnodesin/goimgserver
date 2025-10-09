# API Documentation

## Overview

goimgserver provides a simple REST API for dynamic image processing and transformation. All image operations are performed on-the-fly with automatic caching.

## Base URL

```
http://localhost:9000
```

## Authentication

Currently, the API does not require authentication. For production use, consider implementing authentication via a reverse proxy (nginx, Apache).

## Endpoints

### Image Endpoints

#### GET /img/{filename}

Returns the original image or default cached version.

**Parameters:**
- `filename` (path parameter, required): The name of the image file

**Example Request:**
```bash
curl -X GET "http://localhost:9000/img/sample.jpg"
```

**Response:**
- **Status Code:** 200 OK
- **Content-Type:** image/jpeg (or appropriate image type)
- **Body:** Binary image data

---

#### GET /img/{filename}/{dimensions}

Returns the image resized to specified dimensions.

**Parameters:**
- `filename` (path parameter, required): The name of the image file
- `dimensions` (path parameter, required): Image dimensions in format `{width}x{height}` (e.g., `800x600`)

**Example Request:**
```bash
curl -X GET "http://localhost:9000/img/sample.jpg/800x600"
```

**Response:**
- **Status Code:** 200 OK
- **Content-Type:** image/jpeg
- **Body:** Resized image data

---

#### GET /img/{filename}/{dimensions}/{format}

Returns the image resized and converted to specified format.

**Parameters:**
- `filename` (path parameter, required): The name of the image file
- `dimensions` (path parameter, required): Image dimensions in format `{width}x{height}`
- `format` (path parameter, required): Output format (`webp`, `png`, `jpeg`, `jpg`)

**Query Parameters (Optional):**
- `quality` (integer, 1-100): Output quality for lossy formats (default: 95)
- `width` (integer): Override width from dimensions
- `height` (integer): Override height from dimensions

**Example Requests:**
```bash
# Convert to WebP format
curl -X GET "http://localhost:9000/img/sample.jpg/800x600/webp"

# Convert with custom quality
curl -X GET "http://localhost:9000/img/sample.jpg/800x600/webp?quality=85"

# Query parameters override path parameters
curl -X GET "http://localhost:9000/img/sample.jpg/800x600?width=1000&height=750"
```

**Response:**
- **Status Code:** 200 OK
- **Content-Type:** image/webp (or specified format)
- **Body:** Processed image data

**Error Responses:**
- **400 Bad Request:** Invalid dimensions or format
- **404 Not Found:** Image file not found
- **500 Internal Server Error:** Processing error

---

### Command Endpoints

#### POST /cmd/clear

Clears the entire cache directory.

**Example Request:**
```bash
curl -X POST "http://localhost:9000/cmd/clear"
```

**Response:**
```json
{
  "success": true,
  "message": "Cache cleared successfully",
  "cleared_files": 1234,
  "freed_space": "2.5GB"
}
```

---

#### POST /cmd/gitupdate

Updates the images directory via `git pull` if it's a git repository.

**Example Request:**
```bash
curl -X POST "http://localhost:9000/cmd/gitupdate"
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Git update completed",
  "changes": 5,
  "branch": "main",
  "last_commit": "abc123..."
}
```

**Response (Not a Git Repo):**
```json
{
  "success": false,
  "error": "Images directory is not a git repository",
  "code": "GIT_NOT_FOUND"
}
```

---

#### POST /cmd/:name

Generic command router that dispatches to specific command handlers.

**Parameters:**
- `name` (path parameter, required): Command name (`clear`, `gitupdate`)

**Valid Commands:**
- `clear` - Clear cache
- `gitupdate` - Update images from git

---

### Health Check Endpoints

#### GET /health

Returns server health status.

**Example Request:**
```bash
curl -X GET "http://localhost:9000/health"
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1699999999,
  "uptime": "2h30m15s"
}
```

---

#### GET /health/ready

Returns readiness status (always returns ready if server is running).

**Example Request:**
```bash
curl -X GET "http://localhost:9000/health/ready"
```

**Response:**
```json
{
  "ready": true
}
```

---

## Response Formats

### Success Response

Image endpoints return binary image data with appropriate Content-Type headers.

Command endpoints return JSON responses:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  // Additional fields depending on the command
}
```

### Error Response

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "status": 400
}
```

**Common Error Codes:**
- `VALIDATION_ERROR` - Invalid input parameters
- `FILE_NOT_FOUND` - Requested image not found
- `PROCESSING_ERROR` - Image processing failed
- `GIT_NOT_FOUND` - Git repository not found
- `INTERNAL_ERROR` - Internal server error

---

## Rate Limiting

The API implements rate limiting per IP address:

- **Default Limit:** 100 requests per minute
- **Burst:** 10 requests

When rate limit is exceeded:

**Response:**
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "status": 429,
  "retry_after": 60
}
```

**Headers:**
- `X-RateLimit-Limit`: Maximum requests per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the rate limit resets (Unix timestamp)

---

## CORS Headers

The server includes CORS headers for browser-based clients:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

---

## Request ID

Each request is assigned a unique request ID for tracing:

**Response Headers:**
```
X-Request-ID: abc123def456...
```

This ID can be used for debugging and log correlation.

---

## Caching

Processed images are automatically cached on the server. Cache keys are generated based on:

- Original filename
- Dimensions
- Format
- Quality

The cache is stored in the configured cache directory and persists across server restarts.

---

## Best Practices

1. **Use WebP format** for web delivery (smaller file sizes)
2. **Specify quality** when converting to lossy formats
3. **Pre-cache common sizes** using the pre-cache system
4. **Monitor rate limits** in production environments
5. **Use a CDN** for high-traffic deployments
6. **Clear cache periodically** to free disk space
7. **Use query parameters** for dynamic dimension requirements

---

## Examples

See the [examples](examples/) directory for detailed usage examples:

- [Basic Usage](examples/basic_usage.md)
- [Advanced Usage](examples/advanced_usage.md)

---

## OpenAPI Specification

A complete OpenAPI 3.0 specification is available in [openapi.yaml](openapi.yaml).
