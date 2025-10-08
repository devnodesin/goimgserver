# File Resolution System

## Overview

The file resolution system handles automatic file extension detection, grouped image organization, and maintains backward compatibility while adding flexible image access patterns.

## Core Features

### Extension Auto-Detection
Automatically finds image files when no extension is specified in the URL.

### Grouped Image Support
Organizes images in folders with default fallback mechanisms.

### Extension Priority Order
Consistent priority when multiple extensions exist for the same base filename.

## File Organization Patterns

### Single Images (Traditional)
```
{image_dir}/
├── cat.jpg                    # Accessible as /img/cat.jpg or /img/cat
├── dog.png                    # Accessible as /img/dog.png or /img/dog  
├── logo.webp                  # Accessible as /img/logo.webp or /img/logo
├── profile.jpg                # Accessible as /img/profile.jpg or /img/profile
├── profile.png                # Lower priority than profile.jpg
└── default.jpg                # System default image
```

### Grouped Images
```
{image_dir}/
├── cats/
│   ├── default.jpg            # Accessible as /img/cats
│   ├── cat_white.jpg          # Accessible as /img/cats/cat_white.jpg or /img/cats/cat_white
│   ├── cat_white.png          # Lower priority than cat_white.jpg
│   ├── funny_white.png        # Accessible as /img/cats/funny_white.png or /img/cats/funny_white
│   └── cute_cat.webp          # Accessible as /img/cats/cute_cat.webp or /img/cats/cute_cat
├── dogs/
│   ├── default.png            # Accessible as /img/dogs
│   ├── puppy.jpg              # Accessible as /img/dogs/puppy.jpg or /img/dogs/puppy
│   └── adult.webp             # Accessible as /img/dogs/adult.webp or /img/dogs/adult
└── default.jpg                # System default image
```

## Extension Priority Order

When multiple files exist with the same base name but different extensions, the system uses this priority order:

### Priority Sequence
1. **jpg/jpeg** (highest priority)
2. **png** (medium priority)
3. **webp** (lowest priority)

### Priority Examples
```
# Scenario 1: All extensions exist
Files: profile.jpg, profile.png, profile.webp
Request: GET /img/profile
Result: Serves profile.jpg

# Scenario 2: jpg/jpeg missing
Files: profile.png, profile.webp
Request: GET /img/profile  
Result: Serves profile.png

# Scenario 3: Only webp exists
Files: profile.webp
Request: GET /img/profile
Result: Serves profile.webp

# Scenario 4: jpeg vs jpg
Files: profile.jpeg, profile.jpg
Request: GET /img/profile
Result: Serves profile.jpeg (jpeg has same priority as jpg, first found wins)
```

## Resolution Algorithm

### Single Image Resolution
```
Input: /img/{filename_with_or_without_ext}

1. If filename contains extension:
   a. Check if {image_dir}/{filename} exists
   b. If exists: return path
   c. If not exists: proceed to default image fallback

2. If filename has no extension:
   a. Try {image_dir}/{filename}.jpg
   b. Try {image_dir}/{filename}.jpeg  
   c. Try {image_dir}/{filename}.png
   d. Try {image_dir}/{filename}.webp
   e. If any found: return first match following priority
   f. If none found: proceed to default image fallback

3. Default image fallback:
   a. Try {image_dir}/default.jpg
   b. Try {image_dir}/default.jpeg
   c. Try {image_dir}/default.png
   d. Try {image_dir}/default.webp
   e. If found: return path (will be processed as {default_image})
   f. If not found: generate programmatic placeholder
```

### Grouped Image Resolution
```
Input: /img/{foldername} or /img/{foldername}/{filename}

1. Check if {image_dir}/{foldername} is a directory
2. If not a directory: treat as single image resolution

3. For group default (/img/{foldername}):
   a. Try {image_dir}/{foldername}/default.jpg
   b. Try {image_dir}/{foldername}/default.jpeg
   c. Try {image_dir}/{foldername}/default.png
   d. Try {image_dir}/{foldername}/default.webp
   e. If found: return path
   f. If not found: fallback to system default image

4. For specific grouped image (/img/{foldername}/{filename}):
   a. If filename has extension:
      - Check {image_dir}/{foldername}/{filename}
      - If exists: return path
      - If not exists: fallback to group default, then system default
   
   b. If filename has no extension:
      - Try {image_dir}/{foldername}/{filename}.jpg
      - Try {image_dir}/{foldername}/{filename}.jpeg
      - Try {image_dir}/{foldername}/{filename}.png
      - Try {image_dir}/{foldername}/{filename}.webp
      - If found: return first match following priority
      - If not found: fallback to group default, then system default
```

## Implementation Examples

### Example 1: Single Image with Extension
```
Request: GET /img/cat.jpg
Resolution:
1. Check {image_dir}/cat.jpg → exists
2. Return: {image_dir}/cat.jpg
```

### Example 2: Single Image without Extension
```
Request: GET /img/profile
Files exist: profile.png, profile.webp
Resolution:
1. Try {image_dir}/profile.jpg → not found
2. Try {image_dir}/profile.jpeg → not found  
3. Try {image_dir}/profile.png → found!
4. Return: {image_dir}/profile.png
```

### Example 3: Grouped Image Default
```
Request: GET /img/cats
Directory exists: {image_dir}/cats/
Files in cats/: default.jpg, cat_white.png
Resolution:
1. Check {image_dir}/cats is directory → yes
2. Try {image_dir}/cats/default.jpg → found!
3. Return: {image_dir}/cats/default.jpg
```

### Example 4: Specific Grouped Image
```
Request: GET /img/cats/cat_white
Directory exists: {image_dir}/cats/
Files in cats/: cat_white.jpg, cat_white.png
Resolution:
1. Check {image_dir}/cats is directory → yes
2. Try {image_dir}/cats/cat_white.jpg → found!
3. Return: {image_dir}/cats/cat_white.jpg
```

### Example 5: Missing Grouped Image Fallback
```
Request: GET /img/cats/missing_cat
Directory exists: {image_dir}/cats/
Files in cats/: default.jpg
Resolution:
1. Check {image_dir}/cats is directory → yes
2. Try {image_dir}/cats/missing_cat.jpg → not found
3. Try {image_dir}/cats/missing_cat.jpeg → not found
4. Try {image_dir}/cats/missing_cat.png → not found  
5. Try {image_dir}/cats/missing_cat.webp → not found
6. Fallback to group default: {image_dir}/cats/default.jpg → found!
7. Return: {image_dir}/cats/default.jpg (marked as {default_image})
```

### Example 6: Complex Priority Resolution
```
Request: GET /img/dogs/puppy
Directory exists: {image_dir}/dogs/
Files in dogs/: puppy.webp, puppy.png, puppy.jpg
Resolution:
1. Check {image_dir}/dogs is directory → yes
2. Try {image_dir}/dogs/puppy.jpg → found! (highest priority)
3. Return: {image_dir}/dogs/puppy.jpg
```

## Cache Path Generation

The resolved file path determines the cache structure:

### Single Images
```
Resolved: {image_dir}/photo.jpg
Cache: {cache_dir}/photo.jpg/{hash}

Resolved: {image_dir}/profile.png (from /img/profile request)
Cache: {cache_dir}/profile.png/{hash}
```

### Grouped Images  
```
Resolved: {image_dir}/cats/default.jpg (from /img/cats request)
Cache: {cache_dir}/cats/default.jpg/{hash}

Resolved: {image_dir}/cats/cat_white.jpg (from /img/cats/cat_white request)  
Cache: {cache_dir}/cats/cat_white.jpg/{hash}
```

### Default Image Fallback
```
Request: /img/missing.jpg
Resolved: {image_dir}/default.jpg (fallback)
Cache: {cache_dir}/missing.jpg/{hash}  # Note: cached under original request path

Request: /img/cats/missing_cat
Resolved: {image_dir}/cats/default.jpg (fallback)
Cache: {cache_dir}/cats/missing_cat/{hash}  # Note: cached under original request path
```

## Error Handling

### File System Errors
- **Permission denied**: Log error, fallback to {default_image}
- **Directory inaccessible**: Log error, fallback to {default_image}
- **File corruption**: Log error, fallback to {default_image}

### Resolution Failures
- **No file found**: Always fallback to {default_image} (never 404)
- **Invalid path**: Sanitize path, attempt resolution
- **Symlink issues**: Follow symlinks safely within {image_dir}

## Performance Optimizations

### File System Caching
- Cache directory structure in memory
- Cache file existence checks for frequently accessed paths
- Invalidate cache when files change (use file system watchers)

### Efficient Resolution
- Use `os.Stat()` for existence checks (faster than opening files)
- Batch directory scans during startup
- Skip expensive checks for known missing files

### Priority Order Optimization
- Check highest priority extensions first
- Stop at first match (don't check all extensions)
- Cache extension mapping for resolved files

## Security Considerations

### Path Traversal Prevention
```go
func sanitizePath(path string) string {
    // Remove dangerous characters and path traversal attempts
    cleaned := filepath.Clean(path)
    
    // Ensure path stays within image directory
    if strings.Contains(cleaned, "..") {
        return ""
    }
    
    return cleaned
}
```

### Symlink Handling
- Follow symlinks only within {image_dir}
- Prevent symlinks pointing outside {image_dir}
- Log suspicious symlink access attempts

## Testing Requirements

### Unit Tests
- [ ] `TestResolveFile_SingleImage_WithExtension`
- [ ] `TestResolveFile_SingleImage_WithoutExtension`
- [ ] `TestResolveFile_SingleImage_ExtensionPriority`
- [ ] `TestResolveFile_GroupedImage_Default`
- [ ] `TestResolveFile_GroupedImage_Specific`
- [ ] `TestResolveFile_GroupedImage_MissingFallback`
- [ ] `TestResolveFile_PathTraversalPrevention`
- [ ] `TestResolveFile_InvalidPaths`
- [ ] `TestResolveFile_SymlinkHandling`

### Integration Tests
- [ ] `TestFileResolution_CompleteFlow`
- [ ] `TestFileResolution_CacheIntegration`
- [ ] `TestFileResolution_DefaultImageFallback`

### Performance Tests
- [ ] `BenchmarkResolveFile_SingleImage`
- [ ] `BenchmarkResolveFile_GroupedImage`
- [ ] `BenchmarkResolveFile_PriorityOrder`

## File Watching Integration

### Directory Monitoring
Monitor {image_dir} for changes to invalidate resolution cache:

```go
// Watch for file system changes
watcher, err := fsnotify.NewWatcher()
watcher.Add(imageDir)

// Invalidate cache on changes
go func() {
    for {
        select {
        case event := <-watcher.Events:
            invalidateResolutionCache(event.Name)
        }
    }
}()
```

### Cache Invalidation
- **File added**: Update resolution cache
- **File removed**: Invalidate cached resolutions
- **File renamed**: Update resolution mappings
- **Directory changes**: Rescan directory structure

This file resolution system provides robust, flexible image access while maintaining performance and security.