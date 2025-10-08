# API Specification

## Overview

The goimgserver API provides endpoints for dynamic image processing and administrative operations. All endpoints return appropriate HTTP status codes and headers.

## URL Parameter Parsing Philosophy

goimgserver implements **graceful parameter parsing** for maximum user experience and fault tolerance:

### Core Parsing Principles

1. **Fault Tolerance**: Invalid or unrecognized parameters are silently ignored
2. **First Valid Wins**: When multiple parameters of the same type exist, the first valid one is used
3. **Default Fallback**: Invalid parameter values fall back to application defaults
4. **Best Effort Processing**: Service continues with partial parameter recognition

### Parameter Recognition Logic

#### Dimension Parameters
- **Pattern**: `{width}x{height}` or `{width}` (single dimension)
- **Valid Range**: 10-4000 pixels
- **Examples**: `800x600`, `300`, `1920x1080`
- **Invalid Examples**: `99999x99999` (falls back to default 1000x1000)

#### Quality Parameters  
- **Pattern**: `q{1-100}`
- **Valid Range**: 1-100
- **Examples**: `q90`, `q75`, `q50`
- **Invalid Examples**: `q95005`, `q0`, `q101` (falls back to default q75)

#### Format Parameters
- **Valid Values**: `webp`, `png`, `jpeg`, `jpg`
- **Case Sensitive**: Lowercase only
- **Invalid Examples**: `WEBP`, `gif`, `tiff` (falls back to default WebP)

#### Special Commands
- **Valid Values**: `clear` (cache management)

### Parsing Examples

```
URL: /img/photo.jpg/800x600/webp/q90/wow
Parsed: dimensions=800x600, format=webp, quality=q90
Ignored: 'wow' (unrecognized parameter)

URL: /img/logo.png/300/400/jpeg/q85  
Parsed: width=300, format=jpeg, quality=q85
Ignored: '400' (duplicate dimension, first wins)

URL: /img/logo.png/300/png/jpeg/q85
Parsed: width=300, format=png, quality=q85  
Ignored: 'jpeg' (duplicate format, first wins)

URL: /img/banner.jpg/1200x400/png/q95005
Parsed: dimensions=1200x400, format=png, quality=q75 (default)
Ignored: 'q95005' (invalid quality value)

URL: /img/test.jpg/large/png/q150/small
Parsed: format=png, quality=q75 (default), dimensions=1000x1000 (default)
Ignored: 'large', 'q150', 'small' (all invalid)
```

## Image Processing Endpoints

### Base Image Endpoint

**GET** `/img/{filename}`

Returns the image with default settings. If {filename} is not found in {image_dir}, automatically serves {default_image} with the same processing parameters.

- **Default Settings**: 1000x1000px, q75, WebP format
- **Response**: Image binary data (original image or {default_image})
- **Content-Type**: Determined by output format
- **Cache-Control**: Set for optimal caching
- **Fallback Behavior**: Transparent fallback to {default_image} when requested image missing

#### Parameters
- `{filename}`: The name of the image file (path parameter)

#### Responses
- `200 OK`: Image successfully processed and returned (original or {default_image})
- `500 Internal Server Error`: Processing failed (rare, as {default_image} should always be available)

**Note**: 404 responses are eliminated through {default_image} fallback mechanism.

### Resized Image Endpoints

**GET** `/img/{filename}/{width}x{height}`

Returns image resized to specific dimensions.

#### Parameters
- `filename`: Image file name
- `width`: Target width in pixels (10-4000)
- `height`: Target height in pixels (10-4000)

**GET** `/img/{filename}/{width}`

Returns image resized to specific width, maintaining aspect ratio.

#### Parameters
- `filename`: Image file name
- `width`: Target width in pixels (10-4000)

### Quality Control

**GET** `/img/{filename}/q{quality}`

Returns image with specified quality setting.

#### Parameters
- `filename`: Image file name
- `quality`: Quality level (1-100)

### Format Conversion

**GET** `/img/{filename}/{dimensions}/{format}`

Returns image in specified format.

#### Supported Formats
- `webp`: WebP format (default)
- `png`: PNG format
- `jpeg` or `jpg`: JPEG format

### Combined Parameters

**GET** `/img/{filename}/{param1}/{param2}/.../{paramN}`

Combines multiple processing options using graceful parsing.

#### Parameter Order Flexibility
Parameters can appear in any order. The parser identifies each parameter by pattern:

```
/img/photo.jpg/q90/800x600/webp    # Quality first
/img/photo.jpg/webp/q90/800x600    # Format first  
/img/photo.jpg/800x600/webp/q90    # Dimensions first
```

All above URLs produce identical results: 800x600 pixels, WebP format, quality 90.

#### Graceful Error Handling
- **Invalid parameters ignored**: `/img/photo.jpg/800x600/webp/q90/invalid` 
- **Duplicate parameters**: First valid occurrence used
- **Malformed values**: Fall back to defaults
- **No HTTP errors**: Service always attempts processing with best available parameters

#### Example URLs with Graceful Parsing
```
# Complex valid URL
/img/sample.jpg/1920x1080/webp/q95
→ 1920x1080px, WebP format, quality 95

# URL with invalid parameters (gracefully handled)  
/img/sample.jpg/800x600/webp/q90/wow/extra/stuff
→ 800x600px, WebP format, quality 90 (ignores 'wow', 'extra', 'stuff')

# URL with duplicate parameters
/img/sample.jpg/300/400/png/jpeg/q85/q95
→ 300px width, PNG format, quality 85 (first valid of each type)

# URL with invalid values  
/img/sample.jpg/99999x99999/invalidformat/q999
→ 1000x1000px (default), WebP (default), quality 75 (default)
```

### Cache Management

**GET** `/img/{filename}/clear`

Clears cached versions of the specified image (including any {default_image} cached under this {filename}).

#### Responses
- `200 OK`: Cache cleared successfully
- `500 Internal Server Error`: Cache clear operation failed

**Note**: No 404 error as cache clear operations are always valid.

## Administrative Endpoints

### Cache Operations

**POST** `/cmd/clear`

Clears the entire cache directory.

#### Request
- **Method**: POST
- **Content-Type**: application/json (optional)
- **Body**: Empty or JSON with options

#### Response
```json
{
  "status": "success",
  "message": "Cache cleared successfully",
  "files_removed": 1247,
  "bytes_freed": 1048576000
}
```

### Git Operations

**POST** `/cmd/gitupdate`

Updates the images directory from Git repository.

#### Requirements
- Images directory must be a Git repository
- Git must be available in system PATH
- Appropriate permissions for Git operations

#### Response
```json
{
  "status": "success",
  "message": "Git update completed",
  "commits_pulled": 3,
  "files_updated": ["image1.jpg", "image2.png"]
}
```

### Generic Command Execution

**POST** `/cmd/{name}`

Generic command execution framework for future extensions.

#### Security Note
This endpoint should be heavily restricted and potentially disabled in production.

## Response Headers

### Image Responses
- `Content-Type`: Appropriate MIME type for the image format
- `Cache-Control`: Optimized for client and proxy caching
- `ETag`: Hash-based entity tag for cache validation
- `Last-Modified`: Based on cache file timestamp
- `X-Processing-Time`: Time taken to process the image (optional)

### API Responses
- `Content-Type`: application/json
- `X-Request-Id`: Unique request identifier for tracing

## Error Responses

The graceful parsing approach and {default_image} fallback minimize errors significantly:

```json
{
  "error": {
    "code": "PROCESSING_FAILED",
    "message": "Image processing operation failed",
    "details": {
      "filename": "problematic.jpg",
      "fallback_used": "default_image",
      "parsed_parameters": {
        "dimensions": "800x600",
        "format": "webp", 
        "quality": 90,
        "ignored": ["invalid", "extra"]
      }
    }
  },
  "request_id": "12345-67890-abcde"
}
```

### Error Codes
- `PROCESSING_FAILED`: Image processing operation failed (500) - rare due to {default_image} fallback
- `CACHE_ERROR`: Cache operation failed (500)
- `PERMISSION_DENIED`: Insufficient permissions for operation (403)
- `RATE_LIMITED`: Request rate limit exceeded (429)

**Notes**: 
- **FILE_NOT_FOUND errors eliminated** through {default_image} fallback mechanism
- Parameter parsing errors do not generate HTTP errors (graceful parsing)
- {default_image} ensures high availability and minimal error responses

## Rate Limiting

- **Default Limit**: 100 requests per minute per IP
- **Header**: `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- **Status**: `429 Too Many Requests` when limit exceeded

## Authentication

Administrative endpoints may require authentication:

- **Header**: `Authorization: Bearer <token>` (if enabled)
- **Status**: `401 Unauthorized` for missing/invalid credentials
- **Status**: `403 Forbidden` for insufficient permissions

## CORS Support

Cross-Origin Resource Sharing is enabled for image endpoints:

- `Access-Control-Allow-Origin`: Configurable (default: *)
- `Access-Control-Allow-Methods`: GET, POST, OPTIONS
- `Access-Control-Allow-Headers`: Content-Type, Authorization
- `Access-Control-Max-Age`: 86400 (24 hours)

## Health and Monitoring

### Health Check

**GET** `/health`

Returns service health status.

#### Response
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "72h30m15s",
  "checks": {
    "image_directory": "ok",
    "cache_directory": "ok",
    "disk_space": "ok",
    "memory": "ok"
  }
}
```

### Metrics

**GET** `/metrics`

Returns Prometheus-formatted metrics (if enabled).

## Backward Compatibility

The API is designed to be backward compatible:

- New parameters are additive
- Existing endpoint behavior is preserved
- Deprecation notices are provided before removing features