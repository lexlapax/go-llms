package processor

import (
	"sync"
	"time"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/metrics"
)

// CacheEntry represents a single entry in the schema cache with expiration
type CacheEntry struct {
	Value      []byte    // The JSON data
	LastAccess time.Time // Last time this entry was accessed
}

// SchemaCache provides caching for schema JSON to avoid repeated marshaling
type SchemaCache struct {
	lock           sync.RWMutex
	cache          map[uint64]CacheEntry
	metrics        *metrics.CacheMetrics
	size           int64
	operations     int64
	maxSize        int           // Maximum number of entries (0 means unlimited)
	expirationTime time.Duration // How long entries live (0 means never expire)
	lastCleanup    time.Time     // Last time cache was cleaned up
}

// NewSchemaCache creates a new schema cache with a default capacity
func NewSchemaCache() *SchemaCache {
	return &SchemaCache{
		cache:          make(map[uint64]CacheEntry, 100), // Default capacity of 100 schemas
		metrics:        metrics.NewCacheMetrics("schema_cache"),
		size:           0,
		operations:     0,
		maxSize:        1000,             // Default max size of 1000 entries
		expirationTime: 30 * time.Minute, // Default expiration of 30 minutes
		lastCleanup:    time.Now(),
	}
}

// NewSchemaCacheWithOptions creates a new schema cache with custom settings
func NewSchemaCacheWithOptions(maxSize int, expirationTime time.Duration) *SchemaCache {
	initialCapacity := 100
	if maxSize > 0 && maxSize < initialCapacity {
		initialCapacity = maxSize
	}

	return &SchemaCache{
		cache:          make(map[uint64]CacheEntry, initialCapacity),
		metrics:        metrics.NewCacheMetrics("schema_cache"),
		size:           0,
		operations:     0,
		maxSize:        maxSize,
		expirationTime: expirationTime,
		lastCleanup:    time.Now(),
	}
}

// Get retrieves cached schema JSON for the given key
// Returns the cached JSON and a boolean indicating if it was found
func (c *SchemaCache) Get(key uint64) ([]byte, bool) {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordAccessTime(time.Since(startTime))
	}()

	// Try to cleanup expired entries periodically
	c.maybeCleanupExpired()

	// Use a wrapper function to time the actual cache access
	result, found := func() ([]byte, bool) {
		c.lock.RLock()
		defer c.lock.RUnlock()

		entry, ok := c.cache[key]
		if !ok {
			return nil, false
		}

		// Check if the entry is expired
		if c.expirationTime > 0 && time.Since(entry.LastAccess) > c.expirationTime {
			// Don't remove it here to avoid write lock contention
			// It will be cleaned up during next cleanup cycle
			return nil, false
		}

		// Update last access time (done separately with a write lock)
		return entry.Value, true
	}()

	// Update last access time if found (needs write lock)
	if found {
		c.updateAccessTime(key)
		c.metrics.RecordHit()
	} else {
		c.metrics.RecordMiss()
	}

	return result, found
}

// updateAccessTime updates the last access time for a cache entry
func (c *SchemaCache) updateAccessTime(key uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if entry, ok := c.cache[key]; ok {
		entry.LastAccess = time.Now()
		c.cache[key] = entry
	}
}

// Set stores schema JSON in the cache
func (c *SchemaCache) Set(key uint64, value []byte) {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordAccessTime(time.Since(startTime))
	}()

	c.lock.Lock()
	defer c.lock.Unlock()

	// If we're at capacity and this is a new key, evict something
	if c.maxSize > 0 && c.size >= int64(c.maxSize) && c.cache[key].Value == nil {
		c.evictLeastRecentlyUsed()
	}

	// Update cache size tracking if this is a new key
	if c.cache[key].Value == nil {
		c.size++
	}

	// Store the entry with current time
	c.cache[key] = CacheEntry{
		Value:      value,
		LastAccess: time.Now(),
	}
}

// Clear empties the cache
func (c *SchemaCache) Clear() {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordAccessTime(time.Since(startTime))
	}()

	c.lock.Lock()
	defer c.lock.Unlock()

	// If the cache is already large, allocate a new one with the same capacity
	capacity := len(c.cache)
	if capacity > 100 {
		c.cache = make(map[uint64]CacheEntry, capacity)
	} else {
		// Otherwise just clear the map
		for k := range c.cache {
			delete(c.cache, k)
		}
	}

	// Reset size counter, but keep metrics
	c.size = 0
}

// evictLeastRecentlyUsed removes the least recently used entry from the cache
// NOTE: This function assumes the write lock is already held
func (c *SchemaCache) evictLeastRecentlyUsed() {
	if len(c.cache) == 0 {
		return
	}

	var (
		oldestKey       uint64
		oldestTime      time.Time
		oldestTimeFound bool
	)

	// Find the oldest entry
	for key, entry := range c.cache {
		if !oldestTimeFound || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
			oldestTimeFound = true
		}
	}

	// Remove the oldest entry if found
	if oldestTimeFound {
		delete(c.cache, oldestKey)
		c.size--
	}
}

// maybeCleanupExpired periodically cleans up expired entries
func (c *SchemaCache) maybeCleanupExpired() {
	// Only check every minute to avoid lock contention
	if time.Since(c.lastCleanup) < time.Minute {
		return
	}

	// Try to get the lock, but don't block if we can't
	// This avoids slowing down cache hits for cleanup
	if c.lock.TryLock() {
		defer c.lock.Unlock()

		// Update last cleanup time
		c.lastCleanup = time.Now()

		// If no expiration time is set, don't bother checking
		if c.expirationTime <= 0 {
			return
		}

		// Find and remove expired entries
		expiredKeys := make([]uint64, 0)
		for key, entry := range c.cache {
			if time.Since(entry.LastAccess) > c.expirationTime {
				expiredKeys = append(expiredKeys, key)
			}
		}

		// Remove the expired entries
		for _, key := range expiredKeys {
			delete(c.cache, key)
			c.size--
		}
	}
}

// GenerateSchemaKey creates a hash key for a schema
// This is used for cache lookups to avoid repeated JSON marshaling
func GenerateSchemaKey(schema *schemaDomain.Schema) uint64 {
	// Use the improved implementation for better key generation
	return ImprovedGenerateSchemaKey(schema)
}

// GetHitRate returns the current cache hit/miss statistics
func (c *SchemaCache) GetHitRate() (hits int64, misses int64, total int64) {
	return c.metrics.GetHitsMisses()
}

// GetHitRateValue returns the current cache hit rate as a float64
func (c *SchemaCache) GetHitRateValue() float64 {
	return c.metrics.GetHitRate()
}

// GetAverageAccessTime returns the average time taken to access the cache
func (c *SchemaCache) GetAverageAccessTime() time.Duration {
	return c.metrics.GetAverageAccessTime()
}

// GetSize returns the number of entries in the cache
func (c *SchemaCache) GetSize() int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.size
}

// ResetMetrics resets only the metrics while keeping the cache contents
func (c *SchemaCache) ResetMetrics() {
	c.metrics.Reset()
}
