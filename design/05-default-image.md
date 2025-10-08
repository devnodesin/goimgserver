# Default Image System

## Overview

The default image system provides seamless fallback functionality when requested images are not found, ensuring zero 404 errors and consistent user experience through automatic placeholder serving.

## System Design

### Default Image Sources

#### System Default Image
- **Location**: `{image_dir}/default.{ext}` (following extension priority: jpg/jpeg, png, webp)
- **Purpose**: Ultimate fallback when specific images or group defaults are missing
- **Auto-generation**: Created programmatically if missing during startup

#### Group Default Images
- **Location**: `{image_dir}/{foldername}/default.{ext}`
- **Purpose**: Fallback for missing images within specific image groups
- **Hierarchy**: Falls back to system default if group default is missing

### Fallback Hierarchy

```
1. Requested Image
   ↓ (if not found)
2. Group Default Image (for grouped images only)
   ↓ (if not found)
3. System Default Image
   ↓ (if not found)
4. Programmatically Generated Placeholder
```

## Default Image Behavior

### Transparent Fallback Processing
When a requested image is not found:

1. **Automatic Substitution**: System serves {default_image} instead
2. **Same Processing Applied**: {default_image} is processed with identical parameters as requested
3. **Original Path Caching**: Processed {default_image} is cached under the original request path
4. **Transparent to User**: No indication that fallback was used (same URL, same response)
5. **Zero 404 Errors**: All valid image requests return valid images

### Processing Consistency
```
Original Request: GET /img/missing.jpg/800x600/webp/q90
Fallback Behavior:
- Serve: {image_dir}/default.jpg
- Process: 800x600 pixels, WebP format, quality 90
- Cache as: {cache_dir}/missing.jpg/{hash}
- Response: Processed default image (transparent fallback)
```

## Default Image Types

### System Default Image

#### Automatic Discovery
System searches for default image in this order:
1. `{image_dir}/default.jpg`
2. `{image_dir}/default.jpeg`
3. `{image_dir}/default.png`
4. `{image_dir}/default.webp`

#### Programmatic Generation
If no default image is found, system generates:
- **Dimensions**: 1000x1000 pixels
- **Background**: White (#FFFFFF)
- **Text**: "goimgserver" (centered, black font, appropriate size)
- **Format**: JPEG
- **Filename**: `{image_dir}/default.jpg`
- **Quality**: High quality for future processing

```go
// Example programmatic generation
func generateDefaultImage() []byte {
    // Create 1000x1000 white canvas
    // Add centered "goimgserver" text
    // Save as JPEG with high quality
    // Return image data
}
```

### Group Default Images

#### Purpose
Provide contextual defaults for image groups:
- **Cats group**: `{image_dir}/cats/default.jpg` might show a generic cat
- **Products group**: `{image_dir}/products/default.jpg` might show a product placeholder
- **Profiles group**: `{image_dir}/profiles/default.jpg` might show an avatar placeholder

#### Fallback Chain
```
Request: GET /img/cats/missing_cat/400x300
Resolution chain:
1. Try: {image_dir}/cats/missing_cat.{ext} → Not found
2. Try: {image_dir}/cats/default.{ext} → Found? Use it
3. Try: {image_dir}/default.{ext} → System fallback
4. Generate: Programmatic placeholder
```

## Cache Integration

### Cache Path Strategy
Processed default images are cached using the **original request path**, not the default image path:

#### Single Image Fallback
```
Request: GET /img/logo.png/600x400/webp
Actual file: {image_dir}/default.jpg (fallback)
Cache path: {cache_dir}/logo.png/{hash}
Cache key: Based on "logo.png" + processing parameters
```

#### Grouped Image Fallback
```
Request: GET /img/cats/missing/300x300/png
Actual file: {image_dir}/cats/default.jpg (group fallback)
Cache path: {cache_dir}/cats/missing/{hash}
Cache key: Based on "cats/missing" + processing parameters
```

#### System Default Fallback
```
Request: GET /img/cats/missing/300x300/png
Group default: Not found
Actual file: {image_dir}/default.jpg (system fallback)
Cache path: {cache_dir}/cats/missing/{hash}
Cache key: Based on "cats/missing" + processing parameters
```

### Cache Benefits
- **Consistent URLs**: Same request URL always returns same cached result
- **Performance**: Subsequent requests for missing files serve from cache
- **Transparency**: User never knows fallback was used
- **Efficiency**: No duplicate processing for same missing file requests

## Implementation Examples

### Example 1: Simple Single Image Fallback
```
Request: GET /img/nonexistent.jpg
Files exist: {image_dir}/default.jpg
Process:
1. Look for {image_dir}/nonexistent.jpg → Not found
2. Fallback to {image_dir}/default.jpg → Found
3. Process default.jpg with default settings (1000x1000, webp, q75)
4. Cache as {cache_dir}/nonexistent.jpg/{hash}
5. Return processed default image
```

### Example 2: Grouped Image with Parameters
```
Request: GET /img/cats/fluffy/800x600/png/q90
Files exist: {image_dir}/cats/default.jpg, {image_dir}/default.jpg
Process:
1. Look for {image_dir}/cats/fluffy.{ext} → Not found
2. Fallback to {image_dir}/cats/default.jpg → Found (group default)
3. Process cats/default.jpg: 800x600, PNG format, quality 90
4. Cache as {cache_dir}/cats/fluffy/{hash}
5. Return processed group default image
```

### Example 3: Complete Fallback Chain
```
Request: GET /img/products/missing/400x400/webp
Files exist: {image_dir}/default.jpg only
Process:
1. Look for {image_dir}/products/missing.{ext} → Not found
2. Look for {image_dir}/products/default.{ext} → Not found
3. Fallback to {image_dir}/default.jpg → Found (system default)
4. Process default.jpg: 400x400, WebP format, default quality
5. Cache as {cache_dir}/products/missing/{hash}
6. Return processed system default image
```

### Example 4: Programmatic Generation
```
Request: GET /img/test.jpg/200x200
Files exist: None (no default.* files)
Process:
1. Look for {image_dir}/test.jpg → Not found
2. Look for {image_dir}/default.{ext} → Not found
3. Generate programmatic placeholder (1000x1000, white, "goimgserver" text)
4. Process generated image: 200x200, default format and quality
5. Cache as {cache_dir}/test.jpg/{hash}
6. Return processed generated placeholder
```

## Startup Integration

### Default Image Setup Process
During application startup:

1. **Scan for System Default**:
   ```go
   defaultExts := []string{"jpg", "jpeg", "png", "webp"}
   for _, ext := range defaultExts {
       path := filepath.Join(imageDir, "default."+ext)
       if fileExists(path) {
           systemDefault = path
           break
       }
   }
   ```

2. **Generate if Missing**:
   ```go
   if systemDefault == "" {
       generatedImage := generateDefaultImage()
       defaultPath := filepath.Join(imageDir, "default.jpg")
       writeFile(defaultPath, generatedImage)
       systemDefault = defaultPath
   }
   ```

3. **Scan for Group Defaults**:
   ```go
   // During directory scan, identify group default images
   groupDefaults := make(map[string]string)
   // Cache group default locations for quick access
   ```

### Pre-cache Exclusion
Default images should not be pre-cached during startup:
- Avoid redundant cache entries
- Default images are processed on-demand when used as fallbacks
- Cache entries use original request paths, not default image paths

## Error Handling

### Default Image Processing Failures
If default image processing fails:
1. **Log Error**: Record the processing failure
2. **Try Alternative**: Attempt other available default images
3. **Generate Fallback**: Use programmatic generation as ultimate fallback
4. **Return 500**: Only if all fallback mechanisms fail (extremely rare)

### Generation Failures
If programmatic generation fails:
1. **Log Critical Error**: This indicates serious system issues
2. **Return 500**: Service cannot provide any image
3. **Alert Monitoring**: Trigger system alerts for investigation

### File System Issues
- **Permission Errors**: Log and attempt alternative defaults
- **Disk Full**: Log critical error, attempt alternative storage
- **Corruption**: Detect and regenerate default images

## Testing Requirements

### Unit Tests
- [ ] `TestDefaultImage_SystemDefault_Discovery`
- [ ] `TestDefaultImage_SystemDefault_Generation`
- [ ] `TestDefaultImage_GroupDefault_Fallback`
- [ ] `TestDefaultImage_ProcessingSameParameters`
- [ ] `TestDefaultImage_CacheWithOriginalPath`
- [ ] `TestDefaultImage_FallbackChain`
- [ ] `TestDefaultImage_GenerationFailure`
- [ ] `TestDefaultImage_ProcessingFailure`

### Integration Tests
- [ ] `TestDefaultImage_CompleteFlow`
- [ ] `TestDefaultImage_CacheIntegration`
- [ ] `TestDefaultImage_StartupIntegration`
- [ ] `TestDefaultImage_FileResolutionIntegration`

### End-to-End Tests
- [ ] `TestDefaultImage_HTTPRequests`
- [ ] `TestDefaultImage_MissingFileScenarios`
- [ ] `TestDefaultImage_GroupedImageScenarios`

## Performance Considerations

### Default Image Caching
- Keep processed default images in memory for frequently accessed fallbacks
- Monitor cache hit rates for default image usage
- Optimize default image sizes for common use cases

### Processing Optimization
- Pre-process common default image sizes during startup
- Use efficient image formats for default images
- Consider SVG for scalable default images (future enhancement)

### Monitoring Metrics
- Track default image usage frequency
- Monitor processing time for default images
- Alert on high default image usage (may indicate missing files)

## Security Considerations

### Default Image Content
- Ensure default images contain no sensitive information
- Use generic, safe content for system-generated defaults
- Validate user-provided default images during setup

### Access Control
- Default images follow same access control as regular images
- No special permissions required for default image access
- Secure default image generation process

This default image system ensures robust fallback behavior while maintaining performance and user experience transparency.