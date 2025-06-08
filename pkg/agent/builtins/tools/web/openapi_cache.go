// ABOUTME: In-memory cache for OpenAPI specifications to improve performance
// ABOUTME: Provides TTL-based caching and thread-safe access to parsed specs

package web

import (
	"sync"
	"time"
)

// OpenAPICache provides thread-safe caching for OpenAPI specifications
type OpenAPICache struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry
	ttl   time.Duration
}

// cacheEntry represents a cached OpenAPI spec with expiration
type cacheEntry struct {
	spec      *OpenAPISpec
	discovery *OperationDiscovery
	expiresAt time.Time
}

// Global cache instance
var (
	globalCache     *OpenAPICache
	globalCacheOnce sync.Once
)

// GetOpenAPICache returns the global cache instance
func GetOpenAPICache() *OpenAPICache {
	globalCacheOnce.Do(func() {
		globalCache = NewOpenAPICache(15 * time.Minute) // 15-minute TTL
	})
	return globalCache
}

// NewOpenAPICache creates a new cache with the specified TTL
func NewOpenAPICache(ttl time.Duration) *OpenAPICache {
	cache := &OpenAPICache{
		cache: make(map[string]*cacheEntry),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a cached spec and its discovery instance
func (c *OpenAPICache) Get(url string) (*OpenAPISpec, *OperationDiscovery, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[url]
	if !exists {
		return nil, nil, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return nil, nil, false
	}

	return entry.spec, entry.discovery, true
}

// Set stores a spec and its discovery instance in the cache
func (c *OpenAPICache) Set(url string, spec *OpenAPISpec, discovery *OperationDiscovery) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[url] = &cacheEntry{
		spec:      spec,
		discovery: discovery,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes a spec from the cache
func (c *OpenAPICache) Delete(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, url)
}

// Clear removes all entries from the cache
func (c *OpenAPICache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
}

// cleanupLoop periodically removes expired entries
func (c *OpenAPICache) cleanupLoop() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.Cleanup()
	}
}

// Cleanup removes expired entries (public for testing)
func (c *OpenAPICache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for url, entry := range c.cache {
		if now.After(entry.expiresAt) {
			delete(c.cache, url)
		}
	}
}

// Stats returns cache statistics
func (c *OpenAPICache) Stats() (totalEntries int, activeEntries int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalEntries = len(c.cache)
	now := time.Now()

	for _, entry := range c.cache {
		if now.Before(entry.expiresAt) {
			activeEntries++
		}
	}

	return totalEntries, activeEntries
}

// SetTTL updates the cache TTL for new entries
func (c *OpenAPICache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
}
