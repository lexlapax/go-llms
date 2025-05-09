package processor

import (
	"hash/fnv"
	"sync"
	"time"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/metrics"
)

// SchemaCache provides caching for schema JSON to avoid repeated marshaling
type SchemaCache struct {
	lock       sync.RWMutex
	cache      map[uint64][]byte
	metrics    *metrics.CacheMetrics
	size       int64
	operations int64
}

// NewSchemaCache creates a new schema cache with a default capacity
func NewSchemaCache() *SchemaCache {
	return &SchemaCache{
		cache:   make(map[uint64][]byte, 10), // Default capacity of 10 schemas
		metrics: metrics.NewCacheMetrics("schema_cache"),
		size:    0,
		operations: 0,
	}
}

// Get retrieves cached schema JSON for the given key
// Returns the cached JSON and a boolean indicating if it was found
func (c *SchemaCache) Get(key uint64) ([]byte, bool) {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordAccessTime(time.Since(startTime))
	}()

	// Use a wrapper function to time the actual cache access
	result, found := func() ([]byte, bool) {
		c.lock.RLock()
		defer c.lock.RUnlock()

		value, ok := c.cache[key]
		return value, ok
	}()

	// Record cache hit/miss in metrics
	if found {
		c.metrics.RecordHit()
	} else {
		c.metrics.RecordMiss()
	}

	return result, found
}

// Set stores schema JSON in the cache
func (c *SchemaCache) Set(key uint64, value []byte) {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordAccessTime(time.Since(startTime))
	}()

	c.lock.Lock()
	defer c.lock.Unlock()

	// Update cache size tracking if this is a new key
	if _, exists := c.cache[key]; !exists {
		c.size++
	}

	c.cache[key] = value
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
	if capacity > 10 {
		c.cache = make(map[uint64][]byte, capacity)
	} else {
		// Otherwise just clear the map
		for k := range c.cache {
			delete(c.cache, k)
		}
	}

	// Reset size counter, but keep metrics
	c.size = 0
}

// GenerateSchemaKey creates a hash key for a schema
// This is used for cache lookups to avoid repeated JSON marshaling
func GenerateSchemaKey(schema *schemaDomain.Schema) uint64 {
	hasher := fnv.New64()

	// Add schema type to hash
	hasher.Write([]byte(schema.Type))

	// Add required fields to hash
	for _, req := range schema.Required {
		hasher.Write([]byte(req))
	}

	// Add properties to hash (keys and types are the most important for uniqueness)
	for k, prop := range schema.Properties {
		hasher.Write([]byte(k))
		hasher.Write([]byte(prop.Type))
		// Add description to hash
		hasher.Write([]byte(prop.Description))
	}

	// Add title and description to hash
	hasher.Write([]byte(schema.Title))
	hasher.Write([]byte(schema.Description))

	return hasher.Sum64()
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
