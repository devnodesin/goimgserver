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

- `GET /img/{filename}`: Retrieves the image with default settings (1000x1000px, q75, WebP). If {filename} not found, serves {default_image} with same processing.

### Default Image Fallback Behavior

When a requested {filename} is not found in {image_dir}:
1. **Fallback to {default_image}**: Looks for {image_dir}/default.{jpg/jpeg/png/webp}
2. **Apply same processing**: {default_image} is processed with the same parameters as requested
3. **Cache with original filename**: Processed {default_image} is cached using the original {filename} for future requests
4. **Graceful degradation**: If {default_image} is also missing, serves programmatically generated placeholder

### Examples of Default Image Behavior
```
GET /img/missing.jpg/800x600/webp/q90
→ Serves {default_image} resized to 800x600, WebP format, quality 90
→ Caches result as if it were "missing.jpg" for future requests

GET /img/nonexistent.png/400x300/png/q85  
→ Serves {default_image} resized to 400x300, PNG format, quality 85
→ Maintains consistent behavior regardless of missing files
```

### Custom Resized Images

- `GET /img/{filename}/600x400`: Retrieves an image resized to 600x400 pixels.
- `GET /img/{filename}/400`: Retrieves an image resized to 400 pixels in width, maintaining aspect ratio.

### Quality Settings

- `GET /img/{filename}/q50`: Retrieves the image with quality set to 50 (other settings default).
- `GET /img/{filename}/150/q50`: Retrieves the image resized to 150px width and quality set to 50.

### Cache Management

- `GET /img/{filename}/clear`: Clears all cached files for the specified {filename} (including {default_image} cached under this filename).

### Image Formats

- `GET /img/{filename}/600x400/png`: Retrieves an image resized to 600x400 pixels and converted to PNG format. Falls back to {default_image} if {filename} not found.
- `GET /img/{filename}/200x300/webp`: Retrieves an image resized to 200x300 pixels in WebP format. Falls back to {default_image} if {filename} not found.
- `GET /img/{filename}/100x100/jpeg`: Retrieves an image resized to 100x100 pixels in JPEG format. Falls back to {default_image} if {filename} not found.

### Advanced Parameter Combinations

All combinations follow graceful parsing rules:

- `GET /img/{filename}/800x600/webp/q90/invalid`: Processes 800x600, WebP, q90 (ignores 'invalid')
- `GET /img/{filename}/300/400/png/q85`: Uses width 300, PNG format, q85 (ignores '400')
- `GET /img/{filename}/large/png/q95005`: Uses defaults for dimensions and quality, PNG format
