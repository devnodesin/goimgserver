package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// generateHash creates a SHA256 hash from resolved file path and processing parameters
func generateHash(resolvedPath string, params ProcessingParams) string {
	h := sha256.New()

	// Write resolved path
	h.Write([]byte(resolvedPath))

	// Write normalized parameters
	h.Write([]byte(fmt.Sprintf("%dx%d", params.Width, params.Height)))
	h.Write([]byte(params.Format))
	h.Write([]byte(fmt.Sprintf("q%d", params.Quality)))

	return hex.EncodeToString(h.Sum(nil))
}
