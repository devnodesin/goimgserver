# API Specification

## Overview

This document defines the complete HTTP API specification for goimgserver, including endpoints, request/response formats, headers, and error handling.

## Base Configuration

### Server Information
- **Protocol**: HTTP/1.1
- **Default Port**: 9000
- **Content-Type**: Determined by requested format (image/webp, image/png, image/jpeg)
- **CORS**: Enabled for cross-origin requests

### Request Timeouts
- **Connection Timeout**: 30 seconds
- **Processing Timeout**: 60 seconds (for complex image operations)
- **Keep-Alive**: Enabled

## Image Endpoints

### Single Image Access

#### GET /img/{filename}
Retrieve image with default settings.

**Path Parameters:**
- `filename` (string): Image filename with or without extension

**Query Parameters:** None (parameters are part of URL path)

**Response:**
- **Content-Type**: `image/webp` (default)
- **Status Code**: 200
- **Body**: Processed image binary data

**Examples:**
```http
GET /img/photo.jpg
GET /img/logo.png  
GET /img/profile    # Auto-detects extension
```

#### GET /img/{filename}/{parameters...}
Retrieve image with custom processing parameters.

**Path Parameters:**
- `filename` (string): Image filename with or without extension
- `parameters` (string[]): Processing parameters in any order

**Parameter Types:**
- **Dimensions**: `{width}x{height}` or `{width}` (e.g., `800x600`, `400`)
- **Quality**: `q{1-100}` (e.g., `q90`, `q75`)
- **Format**: `webp` | `png` | `jpeg` | `jpg`

**Response:**
- **Content-Type**: Varies based on format parameter
- **Status Code**: 200
- **Body**: Processed image binary data

**Examples:**
```http
GET /img/photo.jpg/800x600
GET /img/photo.jpg/800x600/webp/q90
GET /img/logo/400/png
GET /img/profile/600x400/jpeg/q85/invalid  # Invalid params ignored
```

### Grouped Image Access

#### GET /img/{foldername}
Retrieve group default image with default settings.

**Path Parameters:**
- `foldername` (string): Image group folder name

**Response:**
- **Content-Type**: `image/webp` (default)
- **Status Code**: 200
- **Body**: Processed group default image

**Examples:**
```http
GET /img/cats      # Serves cats/default.*
GET /img/products  # Serves products/default.*
```

#### GET /img/{foldername}/{parameters...}
Retrieve group default image with custom parameters.

**Path Parameters:**
- `foldername` (string): Image group folder name
- `parameters` (string[]): Processing parameters

**Examples:**
```http
GET /img/cats/400x300/png
GET /img/products/600x600/webp/q90
```

#### GET /img/{foldername}/{filename}
Retrieve specific image from group with default settings.

**Path Parameters:**
- `foldername` (string): Image group folder name
- `filename` (string): Specific image filename with or without extension

**Examples:**
```http
GET /img/cats/cat_white.jpg
GET /img/cats/fluffy        # Auto-detects extension
GET /img/products/item123
```

#### GET /img/{foldername}/{filename}/{parameters...}
Retrieve specific grouped image with custom parameters.

**Path Parameters:**
- `foldername` (string): Image group folder name
- `filename` (string): Specific image filename
- `parameters` (string[]): Processing parameters

**Examples:**
```http
GET /img/cats/cat_white/300x300/webp
GET /img/products/item123/800x600/png/q95
```

### Cache Management

#### GET /img/{path}/clear
Clear cached versions of specific image.

**Path Parameters:**
- `path` (string): Image path (can be single image or grouped image path)

**Response:**
- **Content-Type**: `application/json`
- **Status Code**: 200
- **Body**: Cache clear confirmation

**Examples:**
```http
GET /img/photo.jpg/clear           # Clear single image cache
GET /img/cats/clear                # Clear group default cache  
GET /img/cats/cat_white/clear      # Clear specific grouped image cache
```

**Response Format:**
```json
{
  "success": true,
  "message": "Cache cleared for: photo.jpg",
  "cleared_files": 5
}
```

## Command Endpoints

### POST /cmd/clear
Clear entire cache directory.

**Request:**
- **Method**: POST
- **Content-Type**: Not required
- **Body**: Empty

**Response:**
```json
{
  "success": true,
  "message": "Cache cleared successfully",
  "cleared_files": 1234,
  "freed_space": "2.5GB"
}
```

### POST /cmd/gitupdate
Update image repository via Git.

**Request:**
- **Method**: POST
- **Content-Type**: Not required
- **Body**: Empty

**Response:**
```json
{
  "success": true,
  "message": "Git update completed",
  "changes": 5,
  "branch": "main",
  "last_commit": "abc123..."
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Not a git repository",
  "code": "GIT_NOT_FOUND"
}
```

### POST /cmd/{name}
Execute generic command.

**Path Parameters:**
- `name` (string): Command name

**Response:**
```json
{
  "success": true,
  "command": "command_name",
  "message": "Command executed successfully"
}
```

## Response Headers

### Image Responses

#### Standard Headers
```http
Content-Type: image/webp | image/png | image/jpeg
Content-Length: {file_size}
Cache-Control: public, max-age=31536000, immutable
ETag: "{hash}"
Last-Modified: {date}
```

#### CORS Headers
```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Accept
Access-Control-Max-Age: 86400
```

#### Performance Headers
```http
X-Cache-Status: HIT | MISS
X-Processing-Time: 45ms
X-Image-Source: original | default | generated
X-Resolution: 800x600
X-Format-Conversion: jpg->webp
```

### Command Responses

#### Standard Headers
```http
Content-Type: application/json
Cache-Control: no-cache, no-store, must-revalidate
```

## Error Handling

### HTTP Status Codes

#### Success Codes
- **200 OK**: Image served successfully
- **200 OK**: Command executed successfully

#### Error Codes
- **422 Unprocessable Entity**: Image file corrupted or invalid format
- **500 Internal Server Error**: Processing failure or system error

#### Notable: No 404 Errors
The system never returns 404 errors for image requests due to default image fallback system.

### Error Response Format

#### Image Processing Errors
```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json

{
  "error": "Image processing failed",
  "code": "PROCESSING_ERROR",
  "message": "Corrupted image file",
  "requested_path": "/img/photo.jpg/800x600",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### System Errors
```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{
  "error": "Internal server error",
  "code": "SYSTEM_ERROR", 
  "message": "File system unavailable",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Command Errors
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "Command failed",
  "code": "COMMAND_ERROR",
  "message": "Invalid command name",
  "command": "invalid_command"
}
```

## Parameter Validation

### Graceful Parsing Behavior
- **Invalid parameters**: Silently ignored
- **Duplicate parameters**: First valid occurrence used
- **Malformed values**: Default values applied
- **Unknown parameters**: Ignored, processing continues

### Parameter Limits

#### Dimensions
- **Minimum**: 10x10 pixels
- **Maximum**: 4000x4000 pixels
- **Default**: 1000x1000 pixels

#### Quality
- **Range**: 1-100
- **Default**: 75

#### Format
- **Supported**: webp, png, jpeg, jpg
- **Default**: webp

## Content Negotiation

### Format Priority
When multiple formats are acceptable:
1. Explicitly requested format parameter
2. Accept header preference  
3. Default format (WebP)

### Accept Headers
```http
Accept: image/webp,image/png,image/*;q=0.8
Accept: image/jpeg,image/png;q=0.9,*/*;q=0.1
```

## Caching Strategy

### Client-Side Caching
```http
Cache-Control: public, max-age=31536000, immutable
ETag: "sha256-{hash}"
```

### CDN Integration
- **ETags**: Support for conditional requests
- **Immutable**: Processed images never change
- **Long TTL**: 1 year cache lifetime

### Cache Invalidation
- Manual cache clearing via `/clear` endpoints
- Automatic invalidation when source images change

## Rate Limiting

### Current Status
Not implemented (planned for future enhancement)

### Planned Implementation
- **Per-IP limits**: 1000 requests/hour
- **Burst allowance**: 100 requests/minute
- **Command limits**: 10 commands/hour per IP

## Authentication

### Current Status
Not implemented (public access)

### Planned Implementation
- **API Key**: For command endpoints
- **JWT Tokens**: For administrative access
- **Public Images**: No authentication required

## Monitoring and Metrics

### Health Check
```http
GET /health

Response:
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "72h15m",
  "cache_size": "1.2GB",
  "images_processed": 15420
}
```

### Performance Metrics
- **Response Time**: Average, 95th percentile
- **Cache Hit Rate**: Percentage of cache hits
- **Processing Time**: Image processing duration
- **Default Image Usage**: Fallback frequency

## OpenAPI Specification

### Swagger Documentation
Available at: `/swagger` (planned implementation)

### API Schema
```yaml
openapi: 3.0.0
info:
  title: goimgserver API
  version: 1.0.0
  description: Dynamic Image Processing Service

paths:
  /img/{filename}:
    get:
      summary: Get image with default settings
      parameters:
        - name: filename
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Image data
          content:
            image/*:
              schema:
                type: string
                format: binary
```

## Client Libraries

### Future Support
Planned client libraries for:
- **JavaScript/TypeScript**: npm package
- **Python**: pip package  
- **Go**: Go module
- **cURL**: Command-line examples

This API specification ensures consistent, predictable behavior across all image processing operations while maintaining flexibility and fault tolerance.