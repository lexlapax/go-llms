// ABOUTME: GraphQL schema caching with TTL support and memory optimization
// ABOUTME: Provides fast schema lookups and operation discovery caching

package web

import (
	"sync"
	"time"

	"github.com/vektah/gqlparser/v2/ast"
)

// GraphQLCache provides caching for GraphQL schemas and discovery results
type GraphQLCache struct {
	mu              sync.RWMutex
	schemas         map[string]*cachedSchema
	discoveries     map[string]*cachedDiscovery
	defaultTTL      time.Duration
	maxCacheSize    int
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// cachedSchema represents a cached GraphQL schema
type cachedSchema struct {
	schema    *ast.Schema
	expiresAt time.Time
	endpoint  string
}

// cachedDiscovery represents cached discovery results
type cachedDiscovery struct {
	result    *GraphQLDiscoveryResult
	expiresAt time.Time
}

// NewGraphQLCache creates a new GraphQL cache
func NewGraphQLCache(defaultTTL time.Duration) *GraphQLCache {
	cache := &GraphQLCache{
		schemas:         make(map[string]*cachedSchema),
		discoveries:     make(map[string]*cachedDiscovery),
		defaultTTL:      defaultTTL,
		maxCacheSize:    100, // Maximum number of cached schemas
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// GetSchema retrieves a cached schema
func (c *GraphQLCache) GetSchema(endpoint string) (*ast.Schema, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.schemas[endpoint]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(cached.expiresAt) {
		return nil, false
	}

	return cached.schema, true
}

// SetSchema caches a GraphQL schema
func (c *GraphQLCache) SetSchema(endpoint string, schema *ast.Schema, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// Check cache size limit
	if len(c.schemas) >= c.maxCacheSize {
		// Remove oldest entry
		c.evictOldest()
	}

	c.schemas[endpoint] = &cachedSchema{
		schema:    schema,
		expiresAt: time.Now().Add(ttl),
		endpoint:  endpoint,
	}
}

// GetDiscovery retrieves cached discovery results
func (c *GraphQLCache) GetDiscovery(endpoint string) (*GraphQLDiscoveryResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.discoveries[endpoint]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(cached.expiresAt) {
		return nil, false
	}

	return cached.result, true
}

// SetDiscovery caches discovery results
func (c *GraphQLCache) SetDiscovery(endpoint string, result *GraphQLDiscoveryResult, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	c.discoveries[endpoint] = &cachedDiscovery{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}
}

// Clear removes all cached entries
func (c *GraphQLCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.schemas = make(map[string]*cachedSchema)
	c.discoveries = make(map[string]*cachedDiscovery)
}

// Close stops the cleanup goroutine
func (c *GraphQLCache) Close() {
	close(c.stopCleanup)
}

// evictOldest removes the oldest cache entry
func (c *GraphQLCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	// Find oldest schema
	for key, cached := range c.schemas {
		if oldestKey == "" || cached.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.expiresAt
		}
	}

	if oldestKey != "" {
		delete(c.schemas, oldestKey)
	}
}

// cleanupExpired periodically removes expired entries
func (c *GraphQLCache) cleanupExpired() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// removeExpired removes all expired entries
func (c *GraphQLCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Remove expired schemas
	for key, cached := range c.schemas {
		if now.After(cached.expiresAt) {
			delete(c.schemas, key)
		}
	}

	// Remove expired discoveries
	for key, cached := range c.discoveries {
		if now.After(cached.expiresAt) {
			delete(c.discoveries, key)
		}
	}
}

// Stats returns cache statistics
func (c *GraphQLCache) Stats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]int{
		"schemas":     len(c.schemas),
		"discoveries": len(c.discoveries),
		"total":       len(c.schemas) + len(c.discoveries),
	}
}

// Global GraphQL cache instance
var globalGraphQLCache = NewGraphQLCache(15 * time.Minute)

// GetGlobalGraphQLCache returns the global GraphQL cache instance
func GetGlobalGraphQLCache() *GraphQLCache {
	return globalGraphQLCache
}
