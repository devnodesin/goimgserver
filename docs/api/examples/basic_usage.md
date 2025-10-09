# Basic Usage Examples

## Getting Started

This guide covers basic usage patterns for the goimgserver API.

## Simple Image Requests

### Get Original Image

Request the original image without any transformations:

```bash
curl -X GET "http://localhost:9000/img/photo.jpg" \
  -o output.jpg
```

### Resize Image

Resize an image to specific dimensions:

```bash
# Resize to 800x600
curl -X GET "http://localhost:9000/img/photo.jpg/800x600" \
  -o resized.jpg
```

### Common Sizes

Pre-defined common sizes for web and mobile:

```bash
# Thumbnail (150x150)
curl -X GET "http://localhost:9000/img/photo.jpg/150x150" \
  -o thumb.jpg

# Small (400x300)
curl -X GET "http://localhost:9000/img/photo.jpg/400x300" \
  -o small.jpg

# Medium (800x600)
curl -X GET "http://localhost:9000/img/photo.jpg/800x600" \
  -o medium.jpg

# Large (1920x1080)
curl -X GET "http://localhost:9000/img/photo.jpg/1920x1080" \
  -o large.jpg
```

## Format Conversion

### Convert to WebP

WebP provides better compression and quality:

```bash
curl -X GET "http://localhost:9000/img/photo.jpg/800x600/webp" \
  -o photo.webp
```

### Convert to PNG

PNG for lossless compression:

```bash
curl -X GET "http://localhost:9000/img/photo.jpg/800x600/png" \
  -o photo.png
```

### Custom Quality

Specify quality for lossy formats (1-100):

```bash
# High quality WebP
curl -X GET "http://localhost:9000/img/photo.jpg/800x600/webp?quality=95" \
  -o high_quality.webp

# Lower quality for smaller file size
curl -X GET "http://localhost:9000/img/photo.jpg/800x600/webp?quality=75" \
  -o low_quality.webp
```

## HTML Integration

### Using in HTML

```html
<!DOCTYPE html>
<html>
<head>
    <title>Image Gallery</title>
</head>
<body>
    <!-- Original size -->
    <img src="http://localhost:9000/img/photo.jpg" alt="Photo">
    
    <!-- Thumbnail -->
    <img src="http://localhost:9000/img/photo.jpg/150x150" alt="Thumbnail">
    
    <!-- Responsive images with srcset -->
    <img 
        src="http://localhost:9000/img/photo.jpg/800x600/webp"
        srcset="
            http://localhost:9000/img/photo.jpg/400x300/webp 400w,
            http://localhost:9000/img/photo.jpg/800x600/webp 800w,
            http://localhost:9000/img/photo.jpg/1920x1080/webp 1920w
        "
        sizes="(max-width: 600px) 400px, (max-width: 1200px) 800px, 1920px"
        alt="Responsive Photo">
</body>
</html>
```

### Using Picture Element

```html
<picture>
    <!-- WebP for modern browsers -->
    <source 
        srcset="http://localhost:9000/img/photo.jpg/800x600/webp" 
        type="image/webp">
    
    <!-- JPEG fallback -->
    <img 
        src="http://localhost:9000/img/photo.jpg/800x600" 
        alt="Photo">
</picture>
```

## JavaScript Integration

### Fetch API

```javascript
// Fetch and display image
async function loadImage(filename, width, height) {
    const url = `http://localhost:9000/img/${filename}/${width}x${height}/webp`;
    
    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const blob = await response.blob();
        const imageUrl = URL.createObjectURL(blob);
        
        const img = document.createElement('img');
        img.src = imageUrl;
        document.body.appendChild(img);
        
    } catch (error) {
        console.error('Error loading image:', error);
    }
}

// Usage
loadImage('photo.jpg', 800, 600);
```

### Dynamic Image Loading

```javascript
function createImageUrl(filename, options = {}) {
    const { 
        width = 800, 
        height = 600, 
        format = 'webp', 
        quality = 95 
    } = options;
    
    let url = `http://localhost:9000/img/${filename}/${width}x${height}`;
    
    if (format) {
        url += `/${format}`;
    }
    
    const params = new URLSearchParams();
    if (quality !== 95) {
        params.append('quality', quality);
    }
    
    const queryString = params.toString();
    if (queryString) {
        url += `?${queryString}`;
    }
    
    return url;
}

// Usage examples
const thumbUrl = createImageUrl('photo.jpg', { width: 150, height: 150 });
const largeUrl = createImageUrl('photo.jpg', { width: 1920, height: 1080, quality: 90 });
```

## Cache Management

### Clear Cache

Clear all cached images:

```bash
curl -X POST "http://localhost:9000/cmd/clear"
```

Response:
```json
{
  "success": true,
  "message": "Cache cleared successfully",
  "cleared_files": 1234,
  "freed_space": "2.5GB"
}
```

### Update Images from Git

If your images directory is a Git repository:

```bash
curl -X POST "http://localhost:9000/cmd/gitupdate"
```

Response:
```json
{
  "success": true,
  "message": "Git update completed",
  "changes": 5,
  "branch": "main"
}
```

## Query Parameter Overrides

Use query parameters to override path-based dimensions:

```bash
# Path says 800x600, but query params override to 1000x750
curl -X GET "http://localhost:9000/img/photo.jpg/800x600?width=1000&height=750" \
  -o custom.jpg
```

This is useful for dynamic sizing requirements.

## Health Checks

### Check Server Health

```bash
curl -X GET "http://localhost:9000/health"
```

Response:
```json
{
  "status": "healthy",
  "timestamp": 1699999999,
  "uptime": "2h30m15s"
}
```

### Check Readiness

```bash
curl -X GET "http://localhost:9000/health/ready"
```

Response:
```json
{
  "ready": true
}
```

## Error Handling

### Handle Missing Images

```bash
curl -X GET "http://localhost:9000/img/nonexistent.jpg" \
  -w "\nHTTP Status: %{http_code}\n"
```

Response:
```
HTTP Status: 404
{
  "error": "Image not found",
  "code": "FILE_NOT_FOUND",
  "status": 404
}
```

### Handle Invalid Parameters

```bash
curl -X GET "http://localhost:9000/img/photo.jpg/invalid" \
  -w "\nHTTP Status: %{http_code}\n"
```

Response:
```
HTTP Status: 400
{
  "error": "Invalid dimensions",
  "code": "VALIDATION_ERROR",
  "status": 400
}
```

## Tips and Best Practices

1. **Always use WebP for web** - Better compression and quality
2. **Cache on client side** - Images are cacheable, leverage browser caching
3. **Use appropriate dimensions** - Don't request larger images than needed
4. **Set reasonable quality** - 85-95 is usually sufficient
5. **Monitor rate limits** - Check X-RateLimit-* headers
6. **Handle errors gracefully** - Always check HTTP status codes

## Next Steps

- Learn about [Advanced Usage](advanced_usage.md)
- Review the [API Reference](../README.md)
- Check [OpenAPI Specification](../openapi.yaml)
