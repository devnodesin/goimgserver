# URL Parsing Strategy

## Philosophy

goimgserver implements a **graceful parameter parsing** strategy that prioritizes user experience, fault tolerance, and service availability over strict parameter validation.

## Core Principles

### 1. Fault Tolerance First
- **Never fail on invalid parameters** - Unknown or malformed parameters are silently ignored
- **Always serve content** - Service continues with best-effort parameter interpretation
- **Graceful degradation** - Fall back to sensible defaults when parameters are invalid

### 2. First Valid Parameter Wins
When duplicate parameter types are encountered in a URL, the **first valid occurrence** takes precedence:

```
/img/logo.png/300/400/jpeg/q85
         ↑     ↑
    first width wins (300)
    second width ignored (400)

/img/logo.png/300/png/jpeg/q85
              ↑    ↑
         first format wins (png)
         second format ignored (jpeg)
```

### 3. Invalid Values Use Defaults
Malformed parameter values fall back to application defaults:

```
/img/banner.jpg/1200x400/png/q95005
                             ↑
                    invalid quality (q95005)
                    falls back to default (q75)
```

## Parameter Recognition

### Parsing Algorithm
1. **Extract path segments** after `/img/{path}/` where {path} can be:
   - `{filename}` - single image
   - `{foldername}` - group default image
   - `{foldername}/{filename}` - specific grouped image
2. **Identify base path** and separate processing parameters
3. **Iterate through parameter segments** in order
4. **Pattern match** each segment against known parameter types
5. **First match wins** - subsequent matches of same type are ignored
6. **Invalid matches ignored** - continue processing remaining segments
7. **Apply defaults** for any missing parameters

### Parameter Type Patterns

#### Dimensions
- **Pattern**: `{width}x{height}` or `{width}`
- **Validation**: Both width and height must be 10-4000 pixels
- **Examples**: 
  - Valid: `800x600`, `300`, `1920x1080`
  - Invalid: `0x0`, `99999x1`, `800x`, `x600`
- **Default**: `1000x1000`

#### Quality
- **Pattern**: `q{1-100}`
- **Validation**: Integer between 1 and 100 (inclusive)
- **Examples**:
  - Valid: `q90`, `q75`, `q1`, `q100`
  - Invalid: `q0`, `q101`, `q95005`, `quality90`
- **Default**: `q75`

#### Format
- **Pattern**: Exact string match (case-sensitive)
- **Valid Values**: `webp`, `png`, `jpeg`, `jpg`
- **Examples**:
  - Valid: `webp`, `png`, `jpeg`, `jpg`
  - Invalid: `WEBP`, `PNG`, `gif`, `tiff`, `bmp`
- **Default**: `webp`

#### Special Commands
- **Pattern**: Exact string match
- **Valid Values**: `clear`
- **Purpose**: Cache management operations

## URL Path Structure

### Single Images
```
/img/{filename}/{parameters...}
/img/{filename_no_ext}/{parameters...}

Examples:
/img/photo.jpg/800x600/webp/q90
/img/logo/400x200/png
```

### Grouped Images
```
/img/{foldername}/{parameters...}           # Group default
/img/{foldername}/{filename}/{parameters...} # Specific grouped image

Examples:
/img/cats/600x400/webp                      # cats/default.* → 600x400 WebP
/img/cats/cat_white/300x300/png             # cats/cat_white.* → 300x300 PNG  
/img/cats/cute_cat.jpg/200x200              # cats/cute_cat.jpg → 200x200
```

## Implementation Examples

### Example 1: Single Image - Complex Valid URL
```
URL: /img/photo.jpg/800x600/webp/q90

Parsing steps:
1. Base path: "photo.jpg"
2. Parameters: ["800x600", "webp", "q90"]
3. Parse "800x600" → dimensions: width=800, height=600 ✓
4. Parse "webp" → format: webp ✓  
5. Parse "q90" → quality: 90 ✓

Result: photo.jpg resized to 800x600px, WebP format, quality 90
Fallback: If photo.jpg not found, serve {default_image} with same processing
```

### Example 2: Grouped Image - Default Access
```
URL: /img/cats/400x300/png/q85

Parsing steps:
1. Check if "cats" is a folder in {image_dir}
2. Base path: "cats/default.*" (following extension priority)
3. Parameters: ["400x300", "png", "q85"]
4. Parse "400x300" → dimensions: width=400, height=300 ✓
5. Parse "png" → format: png ✓
6. Parse "q85" → quality: 85 ✓

Result: cats/default.* resized to 400x300px, PNG format, quality 85
Fallback: If cats/default.* not found, serve {default_image} with same processing
```

### Example 3: Grouped Image - Specific File
```
URL: /img/cats/cat_white/300x300/webp

Parsing steps:
1. Check if "cats" is a folder in {image_dir}
2. Base path: "cats/cat_white.*" (following extension priority)
3. Parameters: ["300x300", "webp"]
4. Parse "300x300" → dimensions: width=300, height=300 ✓
5. Parse "webp" → format: webp ✓

Result: cats/cat_white.* resized to 300x300px, WebP format, default quality
Fallback: If cats/cat_white.* not found, serve {default_image} with same processing
```

### Example 4: Auto-Extension Detection
```
URL: /img/profile/600x400/jpeg

Parsing steps:
1. Base path: "profile" (no extension)
2. File resolution: Search for profile.jpg → profile.png → profile.webp
3. Parameters: ["600x400", "jpeg"]
4. Parse "600x400" → dimensions: width=600, height=400 ✓
5. Parse "jpeg" → format: jpeg ✓

Result: Found file (e.g., profile.jpg) converted to 600x400px, JPEG format
Fallback: If no profile.* found, serve {default_image} with same processing
```

### Example 5: Invalid Parameters Ignored
```
URL: /img/cats/cat_white/800x600/webp/q90/wow/extra

Parsing steps:
1. Base path: "cats/cat_white.*"
2. Parameters: ["800x600", "webp", "q90", "wow", "extra"]
3. Parse "800x600" → dimensions: width=800, height=600 ✓
4. Parse "webp" → format: webp ✓
5. Parse "q90" → quality: 90 ✓
6. Parse "wow" → no pattern match, ignored ✗
7. Parse "extra" → no pattern match, ignored ✗

Result: cats/cat_white.* resized to 800x600px, WebP format, quality 90 (ignores invalid parameters)
```

### Example 6: Duplicate Parameters
```
URL: /img/logo/300/400/jpeg/png/q85/q95

Parsing steps:
1. Base path: "logo.*" (auto-detect extension)
2. Parameters: ["300", "400", "jpeg", "png", "q85", "q95"]
3. Parse "300" → width: 300 ✓ (first width)
4. Parse "400" → width: 400, but width already set, ignored ✗
5. Parse "jpeg" → format: jpeg ✓ (first format)
6. Parse "png" → format: png, but format already set, ignored ✗
7. Parse "q85" → quality: 85 ✓ (first quality)
8. Parse "q95" → quality: 95, but quality already set, ignored ✗

Result: logo.* resized to 300px width, JPEG format, quality 85
```

### Example 7: Invalid Values with Fallback
```
URL: /img/cats/banner/99999x1/invalidformat/q95005

Parsing steps:
1. Base path: "cats/banner.*"
2. Parameters: ["99999x1", "invalidformat", "q95005"]
3. Parse "99999x1" → width=99999 (>4000, invalid), height=1 (<10, invalid) ✗
4. Parse "invalidformat" → no format match ✗
5. Parse "q95005" → quality=95005 (>100, invalid) ✗
6. Apply defaults: 1000x1000, webp, q75

Result: cats/banner.* using default dimensions (1000x1000px), WebP format, quality 75 (default)
```

## File Resolution Integration

The URL parsing system works in conjunction with the file resolution system:

### Resolution Process
1. **Parse URL** to extract base path and parameters
2. **Resolve file path** using extension auto-detection and grouping rules
3. **Apply graceful parameter parsing** to determine processing options
4. **Process image** with resolved file and parsed parameters

### Path Resolution Examples
```
# Auto-extension resolution
/img/cat → resolves to {image_dir}/cat.jpg (if exists, following priority)

# Grouped image resolution  
/img/cats → resolves to {image_dir}/cats/default.jpg (if exists, following priority)
/img/cats/white → resolves to {image_dir}/cats/white.jpg (if exists, following priority)

# Explicit extension (no resolution needed)
/img/cat.png → resolves to {image_dir}/cat.png
/img/cats/white.webp → resolves to {image_dir}/cats/white.webp
```

## Error Handling Strategy

### No HTTP Errors for Parameter Issues
- **Invalid parameters**: Ignored, continue processing
- **Malformed values**: Use defaults, continue processing  
- **Unknown parameters**: Ignored, continue processing
- **Duplicate parameters**: Use first valid, ignore rest

### HTTP Errors Only For:
- **Processing failure**: 500 - Image processing failed
- **System errors**: 500 - File system, memory, or other system issues
- **Note**: File not found no longer generates 404 due to {default_image} fallback system

## Benefits of Graceful Parsing

### User Experience
- **URLs always work** - No frustrating 400 errors for typos
- **Typo tolerance** - Minor mistakes don't break functionality
- **Exploration friendly** - Users can experiment with parameters

### System Reliability  
- **High availability** - Service continues even with malformed requests
- **Reduced error rates** - Fewer HTTP errors in logs and monitoring
- **Cache efficiency** - Invalid parameters don't create cache pollution

### Development Benefits
- **API evolution** - New parameters can be added without breaking old URLs
- **Testing tolerance** - Tests don't break on minor URL variations
- **Debugging ease** - Invalid parameters logged but don't stop processing

## Implementation Guidelines

### Parser Design
- **Stateless parsing** - Each segment parsed independently
- **Early wins** - First valid parameter of each type wins
- **Continue on errors** - Don't stop processing on invalid segments
- **Comprehensive logging** - Log ignored parameters for debugging

### Testing Strategy
- **Test all valid combinations** - Ensure all parameter types work
- **Test invalid parameters** - Verify graceful ignoring
- **Test duplicate parameters** - Verify first-wins behavior  
- **Test edge cases** - Boundary values, empty segments, special characters
- **Test performance** - Parsing should be fast even with many parameters
- **Test file resolution integration** - Verify URL parsing works with auto-detection

### Monitoring and Metrics
- **Track ignored parameters** - Monitor for common user mistakes
- **Parameter usage stats** - Understand most common parameter combinations
- **Performance metrics** - Ensure parsing doesn't impact response time
- **Cache effectiveness** - Monitor cache hit rates with graceful parsing
- **File resolution stats** - Track auto-detection success rates