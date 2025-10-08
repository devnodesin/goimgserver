# Endpoints

## URL Parameter Parsing Strategy

The goimgserver uses a **graceful parameter parsing** approach that prioritizes user experience and fault tolerance:

### Parsing Rules
1. **Invalid parameters are discarded** - Unknown parameters are ignored (e.g., `wow` in `/img/photo.jpg/800x600/webp/q90/wow`)
2. **First valid parameter wins** - When duplicate parameter types exist, use the first valid one:
   - `/img/logo.png/300/400/jpeg/q85` → uses `300` for width (ignores `400`)
   - `/img/logo.png/300/png/jpeg/q85` → uses `png` format (ignores `jpeg`)
3. **Invalid values fall back to defaults** - Malformed parameters use default values:
   - `/img/banner.jpg/1200x400/png/q95005` → `q95005` is invalid, uses default quality (`q75`)
4. **Graceful degradation** - Service continues with best-effort parameter interpretation

### Parameter Types Recognition
- **Dimensions**: `{width}x{height}` (e.g., `800x600`) or single `{width}` (e.g., `300`)
- **Quality**: `q{1-100}` (e.g., `q90`, `q75`)
- **Format**: `webp`, `png`, `jpeg`, `jpg`
- **Commands**: `clear` (cache clearing)

### Examples of Graceful Parsing
```
/img/photo.jpg/800x600/webp/q90/wow
→ Dimensions: 800x600, Format: webp, Quality: q90 (ignores 'wow')

/img/logo.png/300/400/jpeg/q85
→ Width: 300, Format: jpeg, Quality: q85 (ignores '400')

/img/logo.png/300/png/jpeg/q85
→ Width: 300, Format: png, Quality: q85 (ignores 'jpeg')

/img/banner.jpg/1200x400/png/q95005
→ Dimensions: 1200x400, Format: png, Quality: q75 (default, ignores 'q95005')
```

## Command Endpoints

- `POST /cmd/{name}`: Executes a specific command based on the provided name.
- `POST /cmd/clear`: Clears the entire {cache_dir}.
- `POST /cmd/gitupdate`: run `git update` inside {image_dir}, (only if {image_dir} is a git directory)

## Image Endpoints

### Basic Image Access

#### With Extension
- `GET /img/{filename}`: Retrieves the image with default settings (1000x1000px, q75, WebP). If {filename} not found, serves {default_image} with same processing.

#### Without Extension (Auto-detection)
- `GET /img/{filename_no_ext}`: Automatically detects file extension following priority order: jpg/jpeg, png, webp
  - Example: `GET /img/cat` → searches for `cat.jpg`, then `cat.png`, then `cat.webp`

### Grouped Images

#### Folder Access
- `GET /img/{foldername}`: Serves default image from `{image_dir}/{foldername}/default.{ext}` (following extension priority)
  - Example: `GET /img/cats` → serves `{image_dir}/cats/default.jpg` (or .png, .webp based on availability)

#### Individual Images in Groups
- `GET /img/{foldername}/{filename}`: Serves specific image from group folder
  - With extension: `GET /img/cats/cat_white.jpg` → serves `{image_dir}/cats/cat_white.jpg`
  - Without extension: `GET /img/cats/cute_cat` → searches `{image_dir}/cats/` for `cute_cat.jpg`, then `.png`, then `.webp`

### File Extension Priority Order

When no extension is specified or multiple extensions exist for the same base filename, the system follows this priority:
1. **jpg/jpeg** (highest priority)
2. **png** (medium priority)  
3. **webp** (lowest priority)

Examples:
```
# If files exist: cat.jpg, cat.png, cat.webp
GET /img/cat → serves cat.jpg

# If files exist: cat.png, cat.webp  
GET /img/cat → serves cat.png

# If only cat.webp exists
GET /img/cat → serves cat.webp
```

### Default Image Fallback Behavior

When a requested {filename} or {foldername}/{filename} is not found:
1. **Fallback to {default_image}**: Looks for {image_dir}/default.{jpg/jpeg/png/webp} (following extension priority)
2. **Apply same processing**: {default_image} is processed with the same parameters as requested
3. **Cache with original path**: Processed {default_image} is cached using the original request path for future requests
4. **Graceful degradation**: If {default_image} is also missing, serves programmatically generated placeholder

### Examples of Default Image Behavior
```
# Single image not found
GET /img/missing.jpg/800x600/webp/q90
→ Serves {default_image} resized to 800x600, WebP format, quality 90
→ Caches result as if it were "missing.jpg" for future requests

# Grouped image not found
GET /img/cats/nonexistent/400x300/png/q85  
→ Serves {default_image} resized to 400x300, PNG format, quality 85
→ Maintains consistent behavior regardless of missing files

# Group folder not found
GET /img/missing_folder/300x300
→ Serves {default_image} resized to 300x300, default format and quality
```

### Custom Resized Images

- `GET /img/{filename}/600x400`: Retrieves an image resized to 600x400 pixels.
- `GET /img/{filename}/400`: Retrieves an image resized to 400 pixels in width, maintaining aspect ratio.
- `GET /img/{foldername}/{filename}/600x400`: Retrieves grouped image resized to 600x400 pixels.
- `GET /img/{foldername}/600x400`: Retrieves group default image resized to 600x400 pixels.

### Quality Settings

- `GET /img/{filename}/q50`: Retrieves the image with quality set to 50 (other settings default).
- `GET /img/{filename}/150/q50`: Retrieves the image resized to 150px width and quality set to 50.
- `GET /img/{foldername}/{filename}/150/q50`: Retrieves grouped image with custom quality.
- `GET /img/{foldername}/150/q50`: Retrieves group default image with custom quality.

### Cache Management

- `GET /img/{filename}/clear`: Clears all cached files for the specified {filename} (including {default_image} cached under this filename).
- `GET /img/{foldername}/{filename}/clear`: Clears all cached files for the specified grouped image.
- `GET /img/{foldername}/clear`: Clears all cached files for the group default image.

### Image Formats

#### Single Images
- `GET /img/{filename}/600x400/png`: Retrieves an image resized to 600x400 pixels and converted to PNG format.
- `GET /img/{filename}/200x300/webp`: Retrieves an image resized to 200x300 pixels in WebP format.
- `GET /img/{filename}/100x100/jpeg`: Retrieves an image resized to 100x100 pixels in JPEG format.

#### Grouped Images
- `GET /img/{foldername}/600x400/png`: Retrieves group default image resized and converted to PNG.
- `GET /img/{foldername}/{filename}/200x300/webp`: Retrieves specific grouped image resized and converted to WebP.

### Advanced Parameter Combinations

All combinations follow graceful parsing rules and work with both single and grouped images:

#### Single Images
- `GET /img/{filename}/800x600/webp/q90/invalid`: Processes 800x600, WebP, q90 (ignores 'invalid')
- `GET /img/{filename}/300/400/png/q85`: Uses width 300, PNG format, q85 (ignores '400')
- `GET /img/{filename}/large/png/q95005`: Uses defaults for dimensions and quality, PNG format

#### Grouped Images  
- `GET /img/{foldername}/800x600/webp/q90/invalid`: Processes group default image 800x600, WebP, q90 (ignores 'invalid')
- `GET /img/{foldername}/{filename}/300/400/png/q85`: Uses width 300, PNG format, q85 (ignores '400')
- `GET /img/{foldername}/{filename}/large/png/q95005`: Uses defaults for dimensions and quality, PNG format

### Complete Endpoint Examples

#### Traditional Single Images
```
GET /img/logo.png                          # Default settings
GET /img/logo.png/400x200                  # Custom dimensions
GET /img/logo.png/400x200/webp             # Custom dimensions + format
GET /img/logo.png/400x200/webp/q90         # All parameters
GET /img/logo                              # Auto-detect extension
```

#### Grouped Images
```
GET /img/cats                              # Group default image
GET /img/cats/400x200                      # Group default resized
GET /img/cats/cat_white                    # Specific image in group
GET /img/cats/cat_white.jpg                # Specific image with extension
GET /img/cats/cat_white/300x300/webp       # Specific image with parameters
GET /img/cats/funny_white.png/300x300/webp # Mixed: PNG source → WebP output
```

#### Auto-Detection Examples
```
# Server searches: cute_cat.jpg → cute_cat.png → cute_cat.webp
GET /img/cats/cute_cat/150/q50

# Server searches: profile.jpg → profile.png → profile.webp  
GET /img/profile/600x600/jpeg
```

This comprehensive endpoint system provides flexible image access while maintaining backward compatibility and graceful error handling.