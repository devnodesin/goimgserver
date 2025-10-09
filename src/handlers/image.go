package handlers

import (
	"fmt"
	"goimgserver/cache"
	"goimgserver/config"
	"goimgserver/processor"
	"goimgserver/resolver"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// ImageHandler handles image serving requests
type ImageHandler struct {
	config       *config.Config
	resolver     resolver.FileResolver
	cache        cache.CacheManager
	processor    processor.ImageProcessor
}

// NewImageHandler creates a new image handler
func NewImageHandler(cfg *config.Config, res resolver.FileResolver, cacheManager cache.CacheManager, proc processor.ImageProcessor) *ImageHandler {
	return &ImageHandler{
		config:    cfg,
		resolver:  res,
		cache:     cacheManager,
		processor: proc,
	}
}

// ServeImage handles image requests with parameter parsing and processing
func (h *ImageHandler) ServeImage(c *gin.Context) {
	// Get the full path from the wildcard
	requestPath := c.Param("path")
	
	// Remove leading slash
	requestPath = strings.TrimPrefix(requestPath, "/")
	
	// Split path into segments
	segments := strings.Split(requestPath, "/")
	if len(segments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}
	
	// Check for clear command
	if hasClearCommand(segments) {
		h.handleCacheClear(c, segments)
		return
	}
	
	// Parse path and parameters
	basePath, paramSegments := h.parsePathAndParams(segments)
	params := parseParameters(paramSegments)
	
	// Resolve the file path
	result, err := h.resolver.Resolve(basePath)
	if err != nil {
		// If resolution fails, try to use default image
		if h.config.DefaultImagePath != "" {
			result = &resolver.ResolutionResult{
				ResolvedPath: h.config.DefaultImagePath,
				IsFallback:   true,
				FallbackType: "system_default",
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "file resolution failed"})
			return
		}
	}
	
	// If file not found in resolution result, use default image
	if result.IsFallback && h.config.DefaultImagePath != "" {
		result.ResolvedPath = h.config.DefaultImagePath
	}
	
	// Convert params to cache params
	cacheParams := cache.ProcessingParams{
		Width:   params.Width,
		Height:  params.Height,
		Format:  params.Format,
		Quality: params.Quality,
	}
	
	// Check cache first (cache under the original request path for fallback images)
	cacheKey := basePath
	if result.IsFallback {
		// For fallback images, cache under the original requested path
		cacheKey = basePath
	} else {
		cacheKey = result.ResolvedPath
	}
	
	cachedData, found, err := h.cache.Retrieve(cacheKey, cacheParams)
	if err == nil && found {
		// Serve from cache
		h.serveImageData(c, cachedData, params.Format)
		return
	}
	
	// Read the image file
	imageData, err := os.ReadFile(result.ResolvedPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image"})
		return
	}
	
	// Validate image
	if err := h.processor.ValidateImage(imageData); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "corrupted or invalid image"})
		return
	}
	
	// Process the image
	processedData, err := h.processImage(imageData, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("image processing failed: %v", err)})
		return
	}
	
	// Store in cache
	if err := h.cache.Store(cacheKey, cacheParams, processedData); err != nil {
		log.Printf("Warning: failed to cache image: %v", err)
	}
	
	// Serve the processed image
	h.serveImageData(c, processedData, params.Format)
}

// parsePathAndParams separates the base path from processing parameters
func (h *ImageHandler) parsePathAndParams(segments []string) (string, []string) {
	// Need to determine where the filename/path ends and parameters begin
	// This is tricky because we support grouped images like /cats/cat_white.jpg
	
	// Strategy: 
	// - If first segment looks like a file (has extension), it's the base path
	// - Otherwise, check if it's a directory
	// - If it's a directory, check if next segment is a file or parameter
	
	if len(segments) == 0 {
		return "", nil
	}
	
	// Check if first segment has extension
	firstExt := filepath.Ext(segments[0])
	if firstExt != "" {
		// Single file with extension
		return segments[0], segments[1:]
	}
	
	// Check if first segment is a directory
	firstPath := filepath.Join(h.config.ImagesDir, segments[0])
	info, err := os.Stat(firstPath)
	if err == nil && info.IsDir() {
		// It's a directory (grouped image)
		if len(segments) == 1 {
			// Just the folder name, accessing default
			return segments[0], nil
		}
		
		// Check if second segment is a file or parameter
		secondExt := filepath.Ext(segments[1])
		if secondExt != "" {
			// Second segment is a file
			return filepath.Join(segments[0], segments[1]), segments[2:]
		}
		
		// Check if second segment looks like a parameter
		if h.looksLikeParameter(segments[1]) {
			// It's a parameter, so first segment is folder (access default)
			return segments[0], segments[1:]
		}
		
		// Second segment is likely a filename without extension
		return filepath.Join(segments[0], segments[1]), segments[2:]
	}
	
	// First segment is likely a filename without extension
	return segments[0], segments[1:]
}

// looksLikeParameter checks if a segment looks like a processing parameter
func (h *ImageHandler) looksLikeParameter(segment string) bool {
	// Check common parameter patterns
	if strings.Contains(segment, "x") {
		// Dimensions like "800x600"
		return true
	}
	if strings.HasPrefix(segment, "q") {
		// Quality like "q90"
		return true
	}
	if validFormats[segment] {
		// Format like "webp", "png", "jpeg"
		return true
	}
	if segment == "clear" {
		return true
	}
	// Check if it's a pure number (width only)
	if len(segment) > 0 && segment[0] >= '0' && segment[0] <= '9' {
		return true
	}
	return false
}

// processImage processes the image with the given parameters
func (h *ImageHandler) processImage(data []byte, params cache.ProcessingParams) ([]byte, error) {
	opts := processor.ProcessOptions{
		Width:   params.Width,
		Height:  params.Height,
		Format:  processor.ImageFormat(params.Format),
		Quality: params.Quality,
	}
	
	return h.processor.Process(data, opts)
}

// serveImageData sends the image data to the client with appropriate headers
func (h *ImageHandler) serveImageData(c *gin.Context, data []byte, format string) {
	// Set CORS headers
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Accept, Content-Type")
	
	// Set cache headers
	c.Header("Cache-Control", "public, max-age=31536000") // 1 year
	
	// Set content type based on format
	contentType := h.getContentType(format)
	c.Header("Content-Type", contentType)
	
	// Send the data
	c.Data(http.StatusOK, contentType, data)
}

// getContentType returns the MIME type for the given format
func (h *ImageHandler) getContentType(format string) string {
	switch format {
	case "webp":
		return "image/webp"
	case "png":
		return "image/png"
	case "jpeg", "jpg":
		return "image/jpeg"
	default:
		return "image/webp"
	}
}

// handleCacheClear clears cache for the specified path
func (h *ImageHandler) handleCacheClear(c *gin.Context, segments []string) {
	// Remove "clear" from segments to get the path
	pathSegments := make([]string, 0)
	for _, seg := range segments {
		if seg != "clear" {
			pathSegments = append(pathSegments, seg)
		}
	}
	
	if len(pathSegments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no path specified for cache clear"})
		return
	}
	
	// Build the path
	basePath := strings.Join(pathSegments, "/")
	
	// Resolve the file path
	result, err := h.resolver.Resolve(basePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	
	// Clear cache for this path
	if err := h.cache.Clear(result.ResolvedPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to clear cache: %v", err)})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "cache cleared",
		"path":    basePath,
	})
}
