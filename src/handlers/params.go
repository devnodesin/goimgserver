package handlers

import (
	"goimgserver/cache"
	"regexp"
	"strconv"
)

// Parameter parsing constants
const (
	DefaultWidth   = 1000
	DefaultHeight  = 1000
	DefaultFormat  = "webp"
	DefaultQuality = 75
	MinDimension   = 10
	MaxDimension   = 4000
	MinQuality     = 1
	MaxQuality     = 100
)

// Valid image formats
var validFormats = map[string]bool{
	"webp": true,
	"png":  true,
	"jpeg": true,
	"jpg":  true,
}

// Regular expressions for parameter parsing
var (
	dimensionsRegex = regexp.MustCompile(`^(\d+)x(\d+)$`)
	widthOnlyRegex  = regexp.MustCompile(`^(\d+)$`)
	qualityRegex    = regexp.MustCompile(`^q(\d+)$`)
)

// parseParameters parses URL segments into ProcessingParams with graceful handling
// Invalid parameters are ignored, first valid parameter of each type wins
func parseParameters(segments []string) cache.ProcessingParams {
	params := cache.ProcessingParams{
		Width:   DefaultWidth,
		Height:  DefaultHeight,
		Format:  DefaultFormat,
		Quality: DefaultQuality,
	}

	// Track which parameters have been set (first wins)
	hasDimensions := false
	hasFormat := false
	hasQuality := false

	for _, segment := range segments {
		// Skip empty segments
		if segment == "" {
			continue
		}

		// Try to parse dimensions (WxH)
		if !hasDimensions {
			if matches := dimensionsRegex.FindStringSubmatch(segment); matches != nil {
				width, _ := strconv.Atoi(matches[1])
				height, _ := strconv.Atoi(matches[2])
				if isValidDimension(width) && isValidDimension(height) {
					params.Width = width
					params.Height = height
					hasDimensions = true
					continue
				}
			}
		}

		// Try to parse width only
		if !hasDimensions {
			if matches := widthOnlyRegex.FindStringSubmatch(segment); matches != nil {
				width, _ := strconv.Atoi(matches[1])
				if isValidDimension(width) {
					params.Width = width
					params.Height = 0 // maintain aspect ratio
					hasDimensions = true
					continue
				}
			}
		}

		// Try to parse quality
		if !hasQuality {
			if matches := qualityRegex.FindStringSubmatch(segment); matches != nil {
				quality, _ := strconv.Atoi(matches[1])
				if isValidQuality(quality) {
					params.Quality = quality
					hasQuality = true
					continue
				}
			}
		}

		// Try to parse format
		if !hasFormat {
			if validFormats[segment] {
				params.Format = segment
				hasFormat = true
				continue
			}
		}

		// If we reach here, the segment is invalid - ignore it
	}

	return params
}

// hasClearCommand checks if clear command is present in segments
func hasClearCommand(segments []string) bool {
	for _, segment := range segments {
		if segment == "clear" {
			return true
		}
	}
	return false
}

// isValidDimension checks if a dimension value is within valid range
func isValidDimension(value int) bool {
	return value >= MinDimension && value <= MaxDimension
}

// isValidQuality checks if a quality value is within valid range
func isValidQuality(value int) bool {
	return value >= MinQuality && value <= MaxQuality
}
