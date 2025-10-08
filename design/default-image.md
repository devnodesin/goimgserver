# Default Image Fallback System

## Overview

goimgserver implements a robust {default_image} fallback system that ensures **zero 404 errors** for image requests. When a requested {filename} is not found in {image_dir}, the system transparently serves a {default_image} with the same processing parameters.

## Default Image Hierarchy

### 1. User-Provided Default Image
**Location**: {image_dir}/default.{extension}  
**Supported Extensions**: `.jpg`, `.jpeg`, `.png`, `.webp`  
**Detection Order**: jpg → jpeg → png → webp

The system scans for {default_image} during startup in this order:
1. {image_dir}/default.jpg
2. {image_dir}/default.jpeg  
3. {image_dir}/default.png
4. {image_dir}/default.webp

**First found file becomes the {default_image}.**

### 2. Programmatically Generated Placeholder
**Fallback Behavior**: If no user-provided {default_image} is found during startup, the system automatically generates one.

**Generated Image Specifications**:
- **Dimensions**: 1000x1000 pixels
- **Background**: White (#FFFFFF)
- **Text**: "goimgserver" 
- **Font Color**: Black (#000000)
- **Text Position**: Centered (both horizontally and vertically)
- **Font Size**: Automatically calculated for optimal visibility (approximately 120-150px)
- **Format**: JPEG
- **Quality**: 95
- **Filename**: {image_dir}/default.jpg

## Fallback Processing Workflow

### Request Processing with Fallback
```
1. Receive request: GET /img/{filename}/800x600/webp/q90

2. Parse URL parameters gracefully:
   - dimensions: 800x600
   - format: webp  
   - quality: 90

3. Check cache: {cache_dir}/{filename}/{hash}
   - If found: serve cached image
   - If not found: continue to step 4

4. Locate source image:
   - Try: {image_dir}/{filename}
   - If found: use original image
   - If not found: use {default_image}

5. Process image (original or {default_image}):
   - Resize to 800x600
   - Convert to WebP
   - Apply quality 90

6. Cache processed image:
   - Store in: {cache_dir}/{filename}/{hash}
   - Note: Uses original {filename} in cache path

7. Serve processed image to client
```

### Cache Key Behavior
**Important**: When {default_image} is served due to missing {filename}, the cache still uses the **original {filename}** in the cache path. This ensures:

- **Consistent URLs**: Same URL always returns same result
- **Cache Efficiency**: Multiple requests for same missing file hit same cache
- **User Experience**: No difference between original and fallback images

## Implementation Details

### Startup Sequence
1. **Directory Initialization**: Ensure {image_dir} exists
2. **Default Image Detection**: Scan for existing {default_image}
3. **Placeholder Generation**: If no {default_image} found, generate programmatic placeholder
4. **Validation**: Verify {default_image} is readable and processable

### Generated Placeholder Creation
```go
// Pseudocode for generated placeholder
func GenerateDefaultImage() error {
    image := CreateImage(1000, 1000, White)
    font := LoadFont("Arial", 130) // Approximately 13% of image width
    text := "goimgserver"
    
    // Calculate center position
    textWidth := MeasureText(text, font)
    textHeight := GetFontHeight(font)
    x := (1000 - textWidth) / 2
    y := (1000 + textHeight) / 2
    
    DrawText(image, text, x, y, Black, font)
    SaveAsJPEG(image, "{image_dir}/default.jpg", quality=95)
}
```

### Runtime Fallback Logic
```go
// Pseudocode for runtime fallback
func ProcessImageRequest(filename string, params ProcessingParams) ([]byte, error) {
    // Try original image first
    imagePath := filepath.Join(imageDir, filename)
    imageData, err := ReadFile(imagePath)
    
    if err != nil {
        // File not found - use default image
        defaultPath := GetDefaultImagePath()
        imageData, err = ReadFile(defaultPath)
        if err != nil {
            return nil, fmt.Errorf("default image not available: %w", err)
        }
    }
    
    // Process image (same logic regardless of source)
    processedImage, err := ProcessImage(imageData, params)
    return processedImage, err
}
```

## Error Handling and Edge Cases

### Scenarios Handled Gracefully
1. **Missing Requested Image**: Serve {default_image} with same processing
2. **Missing Default Image at Startup**: Generate programmatic placeholder
3. **Corrupted Default Image**: Log error and generate new placeholder
4. **Processing Failure**: Attempt with different parameters or serve unprocessed {default_image}

### Failure Scenarios
- **No Write Permission**: Cannot generate placeholder - application fails to start
- **Insufficient Disk Space**: Cannot save generated placeholder - application fails to start
- **Image Processing Library Failure**: Cannot generate placeholder - application fails to start

### Monitoring and Logging
- **Default Image Usage**: Log when {default_image} is served instead of requested image
- **Placeholder Generation**: Log when programmatic placeholder is created
- **Fallback Statistics**: Track ratio of original vs. fallback image serves
- **Performance Impact**: Monitor processing time difference between original and default images

## Configuration Options

### Current Implementation
- **Automatic Detection**: System automatically finds and uses {default_image}
- **Automatic Generation**: Creates placeholder if no {default_image} exists
- **No Configuration Required**: Zero configuration for basic functionality

### Future Enhancements
- **Custom Default Image Path**: Allow specifying {default_image} location
- **Custom Placeholder Text**: Configure text in generated placeholder
- **Custom Placeholder Styling**: Configure colors, fonts, size
- **Multiple Default Images**: Different defaults per directory or image type
- **Default Image Caching**: Separate cache strategy for default images

## Benefits

### User Experience
- **Zero 404 Errors**: All image URLs return valid images
- **Consistent Behavior**: Same URL always returns same result
- **Fast Response**: Cached fallbacks serve quickly
- **Visual Feedback**: Placeholder indicates missing image rather than error

### Developer Experience
- **Simplified Error Handling**: No need to handle 404 cases in client code
- **Testing Friendly**: Missing images don't break tests or demonstrations
- **Deployment Resilient**: Missing files don't cause service failures
- **Monitoring Simplified**: Focus on processing errors rather than missing files

### System Reliability
- **High Availability**: Service remains functional with missing images
- **Graceful Degradation**: Missing content doesn't break user experience
- **Cache Efficiency**: Fallback images benefit from same caching system
- **Resource Optimization**: Single {default_image} serves multiple missing files

## Implementation Priority

### Phase 1: Basic Fallback (Required)
- Default image detection at startup
- Programmatic placeholder generation
- Runtime fallback logic
- Basic logging

### Phase 2: Enhanced Features (Optional)
- Custom placeholder styling
- Fallback statistics and monitoring
- Configuration options
- Performance optimization

### Phase 3: Advanced Features (Future)
- Multiple default images
- Dynamic placeholder generation
- Advanced cache strategies
- Integration with external placeholder services