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
1. **Split URL path** into segments after `/img/{filename}/`
2. **Iterate through segments** in order
3. **Pattern match** each segment against known parameter types
4. **First match wins** - subsequent matches of same type are ignored
5. **Invalid matches ignored** - continue processing remaining segments
6. **Apply defaults** for any missing parameters

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

## Implementation Examples

### Example 1: Complex Valid URL
```
URL: /img/photo.jpg/800x600/webp/q90

Parsing steps:
1. {filename}: "photo.jpg"
2. segments: ["800x600", "webp", "q90"]
3. Parse "800x600" → dimensions: width=800, height=600 ✓
4. Parse "webp" → format: webp ✓  
5. Parse "q90" → quality: 90 ✓

Result: 800x600px, WebP format, quality 90
Fallback: If photo.jpg not found, serve {default_image} with same processing
```

### Example 2: Invalid Parameters Ignored
```
URL: /img/photo.jpg/800x600/webp/q90/wow/extra

Parsing steps:
1. {filename}: "photo.jpg"
2. segments: ["800x600", "webp", "q90", "wow", "extra"]
3. Parse "800x600" → dimensions: width=800, height=600 ✓
4. Parse "webp" → format: webp ✓
5. Parse "q90" → quality: 90 ✓
6. Parse "wow" → no pattern match, ignored ✗
7. Parse "extra" → no pattern match, ignored ✗

Result: 800x600px, WebP format, quality 90 (ignores invalid parameters)
Fallback: If photo.jpg not found, serve {default_image} with same processing
```

### Example 3: Duplicate Parameters
```
URL: /img/logo.png/300/400/jpeg/png/q85/q95

Parsing steps:
1. filename: "logo.png"
2. segments: ["300", "400", "jpeg", "png", "q85", "q95"]
3. Parse "300" → width: 300 ✓ (first width)
4. Parse "400" → width: 400, but width already set, ignored ✗
5. Parse "jpeg" → format: jpeg ✓ (first format)
6. Parse "png" → format: png, but format already set, ignored ✗
7. Parse "q85" → quality: 85 ✓ (first quality)
8. Parse "q95" → quality: 95, but quality already set, ignored ✗

Result: 300px width, JPEG format, quality 85
```

### Example 4: Invalid Values with Fallback
```
URL: /img/banner.jpg/99999x1/invalidformat/q95005

Parsing steps:
1. filename: "banner.jpg"
2. segments: ["99999x1", "invalidformat", "q95005"]
3. Parse "99999x1" → width=99999 (>4000, invalid), height=1 (<10, invalid) ✗
4. Parse "invalidformat" → no format match ✗
5. Parse "q95005" → quality=95005 (>100, invalid) ✗
6. Apply defaults: 1000x1000, webp, q75

Result: 1000x1000px (default), WebP (default), quality 75 (default)
```

## Error Handling Strategy

### No HTTP Errors for Parameter Issues
- **Invalid parameters**: Ignored, continue processing
- **Malformed values**: Use defaults, continue processing  
- **Unknown parameters**: Ignored, continue processing
- **Duplicate parameters**: Use first valid, ignore rest

### HTTP Errors Only For:
- **File not found**: 404 - Image file doesn't exist
- **Processing failure**: 500 - Image processing failed
- **System errors**: 500 - File system, memory, or other system issues

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

### Monitoring and Metrics
- **Track ignored parameters** - Monitor for common user mistakes
- **Parameter usage stats** - Understand most common parameter combinations
- **Performance metrics** - Ensure parsing doesn't impact response time
- **Cache effectiveness** - Monitor cache hit rates with graceful parsing