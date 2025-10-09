package resolver

import (
	"sync"
)

// Cache provides thread-safe caching for file resolution results
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*ResolutionResult
}

// NewCache creates a new resolution cache
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*ResolutionResult),
	}
}

// Get retrieves a cached resolution result
func (c *Cache) Get(key string) (*ResolutionResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result, found := c.entries[key]
	return result, found
}

// Set stores a resolution result in the cache
func (c *Cache) Set(key string, result *ResolutionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.entries[key] = result
}

// Invalidate removes a specific entry from the cache
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.entries = make(map[string]*ResolutionResult)
}

// Size returns the number of cached entries
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.entries)
}
