# Advanced Usage Examples

## Advanced Image Processing

This guide covers advanced usage patterns and optimization techniques.

## Dynamic Dimensions with Query Parameters

Query parameters provide flexibility when path-based dimensions aren't sufficient:

```bash
# Override width only (maintain aspect ratio)
curl -X GET "http://localhost:9000/img/photo.jpg/800x600?width=1200" \
  -o wide.jpg

# Override height only
curl -X GET "http://localhost:9000/img/photo.jpg/800x600?height=900" \
  -o tall.jpg

# Override both dimensions
curl -X GET "http://localhost:9000/img/photo.jpg/800x600?width=1920&height=1080" \
  -o fullhd.jpg

# Query parameters take precedence over path parameters
curl -X GET "http://localhost:9000/img/photo.jpg/100x100?width=500&height=500" \
  -o large_override.jpg
```

## Quality Optimization

### Adaptive Quality

Adjust quality based on use case:

```bash
# Maximum quality for archival (95-100)
curl -X GET "http://localhost:9000/img/photo.jpg/2048x1536/webp?quality=100" \
  -o archive.webp

# High quality for hero images (85-95)
curl -X GET "http://localhost:9000/img/photo.jpg/1920x1080/webp?quality=90" \
  -o hero.webp

# Medium quality for galleries (75-85)
curl -X GET "http://localhost:9000/img/photo.jpg/800x600/webp?quality=80" \
  -o gallery.webp

# Lower quality for thumbnails (60-75)
curl -X GET "http://localhost:9000/img/photo.jpg/150x150/webp?quality=70" \
  -o thumb.webp
```

## Batch Processing

### Shell Script for Batch Conversion

```bash
#!/bin/bash

# Convert multiple images to WebP
IMAGES=(
    "photo1.jpg"
    "photo2.jpg"
    "photo3.jpg"
)

for img in "${IMAGES[@]}"; do
    echo "Processing: $img"
    curl -X GET "http://localhost:9000/img/${img}/800x600/webp" \
        -o "processed_${img%.jpg}.webp"
done

echo "Batch processing complete!"
```

### Parallel Processing

```bash
#!/bin/bash

# Process images in parallel using xargs
cat images.txt | xargs -P 4 -I {} curl -X GET \
    "http://localhost:9000/img/{}/800x600/webp" \
    -o "output/{}.webp"
```

## Responsive Image Sets

### Generate Multiple Sizes

Script to generate responsive image sets:

```bash
#!/bin/bash

IMAGE="photo.jpg"
SIZES=(
    "400x300"
    "800x600"
    "1200x900"
    "1920x1080"
)

for size in "${SIZES[@]}"; do
    curl -X GET "http://localhost:9000/img/${IMAGE}/${size}/webp" \
        -o "${IMAGE%.jpg}_${size}.webp"
    echo "Generated: ${size}"
done
```

### Automated Responsive HTML

Generate HTML with responsive images:

```bash
#!/bin/bash

IMAGE="photo.jpg"
OUTPUT="index.html"

cat > "$OUTPUT" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Responsive Gallery</title>
    <style>
        img { max-width: 100%; height: auto; }
    </style>
</head>
<body>
EOF

for img in images/*.jpg; do
    basename=$(basename "$img")
    cat >> "$OUTPUT" << EOF
    <img 
        src="http://localhost:9000/img/${basename}/800x600/webp"
        srcset="
            http://localhost:9000/img/${basename}/400x300/webp 400w,
            http://localhost:9000/img/${basename}/800x600/webp 800w,
            http://localhost:9000/img/${basename}/1200x900/webp 1200w
        "
        sizes="(max-width: 600px) 400px, (max-width: 1200px) 800px, 1200px"
        alt="${basename}">
EOF
done

cat >> "$OUTPUT" << 'EOF'
</body>
</html>
EOF

echo "Generated: $OUTPUT"
```

## Advanced JavaScript Integration

### Image Loader Class

```javascript
class ImageLoader {
    constructor(baseUrl = 'http://localhost:9000') {
        this.baseUrl = baseUrl;
        this.cache = new Map();
    }
    
    async loadImage(filename, options = {}) {
        const { width = 800, height = 600, format = 'webp', quality = 95 } = options;
        const cacheKey = `${filename}_${width}x${height}_${format}_${quality}`;
        
        // Check cache
        if (this.cache.has(cacheKey)) {
            return this.cache.get(cacheKey);
        }
        
        // Build URL
        let url = `${this.baseUrl}/img/${filename}/${width}x${height}`;
        if (format) {
            url += `/${format}`;
        }
        if (quality !== 95) {
            url += `?quality=${quality}`;
        }
        
        try {
            const response = await fetch(url);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const blob = await response.blob();
            const imageUrl = URL.createObjectURL(blob);
            
            // Cache the result
            this.cache.set(cacheKey, imageUrl);
            
            return imageUrl;
        } catch (error) {
            console.error(`Failed to load image ${filename}:`, error);
            throw error;
        }
    }
    
    async loadMultiple(images) {
        return Promise.all(images.map(img => this.loadImage(img.filename, img.options)));
    }
    
    clearCache() {
        this.cache.forEach(url => URL.revokeObjectURL(url));
        this.cache.clear();
    }
}

// Usage
const loader = new ImageLoader();

// Load single image
const imageUrl = await loader.loadImage('photo.jpg', { 
    width: 1920, 
    height: 1080, 
    quality: 90 
});

// Load multiple images
const images = [
    { filename: 'photo1.jpg', options: { width: 800, height: 600 } },
    { filename: 'photo2.jpg', options: { width: 1920, height: 1080 } }
];
const urls = await loader.loadMultiple(images);
```

### Lazy Loading with Intersection Observer

```javascript
class LazyImageLoader {
    constructor(baseUrl = 'http://localhost:9000') {
        this.baseUrl = baseUrl;
        this.observer = new IntersectionObserver(
            this.handleIntersection.bind(this),
            { rootMargin: '50px' }
        );
    }
    
    observe(img) {
        this.observer.observe(img);
    }
    
    handleIntersection(entries) {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const img = entry.target;
                const filename = img.dataset.filename;
                const width = img.dataset.width || 800;
                const height = img.dataset.height || 600;
                const format = img.dataset.format || 'webp';
                
                const url = `${this.baseUrl}/img/${filename}/${width}x${height}/${format}`;
                
                img.src = url;
                img.onload = () => img.classList.add('loaded');
                
                this.observer.unobserve(img);
            }
        });
    }
}

// HTML usage
// <img data-filename="photo.jpg" data-width="800" data-height="600" data-format="webp" 
//      class="lazy" alt="Photo">

// JavaScript initialization
const lazyLoader = new LazyImageLoader();
document.querySelectorAll('img.lazy').forEach(img => lazyLoader.observe(img));
```

## Performance Monitoring

### Request Performance Tracking

```javascript
async function measureImageLoad(filename, options) {
    const startTime = performance.now();
    
    try {
        const url = createImageUrl(filename, options);
        const response = await fetch(url);
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const blob = await response.blob();
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        const rateLimit = {
            limit: response.headers.get('X-RateLimit-Limit'),
            remaining: response.headers.get('X-RateLimit-Remaining'),
            reset: response.headers.get('X-RateLimit-Reset')
        };
        
        console.log({
            filename,
            duration: `${duration.toFixed(2)}ms`,
            size: `${(blob.size / 1024).toFixed(2)}KB`,
            rateLimit
        });
        
        return blob;
    } catch (error) {
        console.error('Image load failed:', error);
        throw error;
    }
}

// Usage
await measureImageLoad('photo.jpg', { width: 1920, height: 1080, format: 'webp' });
```

## Cache Management Strategies

### Automated Cache Clearing

```bash
#!/bin/bash

# Clear cache daily at midnight via cron
# 0 0 * * * /path/to/clear_cache.sh

CACHE_SIZE_LIMIT=10737418240  # 10GB in bytes

# Get current cache size
CACHE_SIZE=$(du -sb /path/to/cache | cut -f1)

if [ "$CACHE_SIZE" -gt "$CACHE_SIZE_LIMIT" ]; then
    echo "Cache size exceeded limit. Clearing cache..."
    curl -X POST "http://localhost:9000/cmd/clear"
    echo "Cache cleared at $(date)"
fi
```

### Selective Cache Clearing

Clear cache for specific images (manual implementation):

```bash
#!/bin/bash

# Remove cached versions of a specific image
IMAGE_NAME="photo.jpg"
CACHE_DIR="/path/to/cache"

find "$CACHE_DIR" -name "*${IMAGE_NAME}*" -delete
echo "Cleared cache for: $IMAGE_NAME"
```

## Git Integration for Image Updates

### Automated Image Updates

```bash
#!/bin/bash

# Update images from Git and notify
WEBHOOK_URL="https://example.com/webhook"

RESPONSE=$(curl -X POST "http://localhost:9000/cmd/gitupdate")

if echo "$RESPONSE" | grep -q '"success":true'; then
    CHANGES=$(echo "$RESPONSE" | jq -r '.changes')
    
    # Notify webhook
    curl -X POST "$WEBHOOK_URL" \
        -H "Content-Type: application/json" \
        -d "{\"event\":\"images_updated\",\"changes\":$CHANGES}"
    
    echo "Images updated: $CHANGES changes"
else
    echo "Update failed: $RESPONSE"
fi
```

## Rate Limit Handling

### Respect Rate Limits

```javascript
class RateLimitedImageLoader {
    constructor(baseUrl = 'http://localhost:9000') {
        this.baseUrl = baseUrl;
        this.queue = [];
        this.processing = false;
    }
    
    async loadImage(filename, options) {
        return new Promise((resolve, reject) => {
            this.queue.push({ filename, options, resolve, reject });
            this.processQueue();
        });
    }
    
    async processQueue() {
        if (this.processing || this.queue.length === 0) return;
        
        this.processing = true;
        const { filename, options, resolve, reject } = this.queue.shift();
        
        try {
            const url = this.buildUrl(filename, options);
            const response = await fetch(url);
            
            const remaining = parseInt(response.headers.get('X-RateLimit-Remaining'));
            const reset = parseInt(response.headers.get('X-RateLimit-Reset'));
            
            if (response.status === 429) {
                const retryAfter = parseInt(response.headers.get('Retry-After') || '60');
                console.log(`Rate limited. Retrying after ${retryAfter}s`);
                
                setTimeout(() => {
                    this.queue.unshift({ filename, options, resolve, reject });
                    this.processing = false;
                    this.processQueue();
                }, retryAfter * 1000);
                
                return;
            }
            
            const blob = await response.blob();
            resolve(URL.createObjectURL(blob));
            
            // Add delay if approaching rate limit
            if (remaining < 10) {
                const now = Date.now() / 1000;
                const delay = Math.max(0, reset - now) * 1000;
                await new Promise(r => setTimeout(r, delay / remaining));
            }
            
        } catch (error) {
            reject(error);
        } finally {
            this.processing = false;
            this.processQueue();
        }
    }
    
    buildUrl(filename, options) {
        const { width = 800, height = 600, format = 'webp', quality = 95 } = options;
        let url = `${this.baseUrl}/img/${filename}/${width}x${height}`;
        if (format) url += `/${format}`;
        if (quality !== 95) url += `?quality=${quality}`;
        return url;
    }
}
```

## Next Steps

- Review [API Reference](../README.md)
- Check [Performance Guide](../../performance/README.md)
- Learn about [Security Best Practices](../../security/README.md)
