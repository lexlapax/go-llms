package processor

import (
	"hash/fnv"
	"sync"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// SchemaCache provides caching for schema JSON to avoid repeated marshaling
type SchemaCache struct {
	lock  sync.RWMutex
	cache map[uint64][]byte
}

// NewSchemaCache creates a new schema cache with a default capacity
func NewSchemaCache() *SchemaCache {
	return &SchemaCache{
		cache: make(map[uint64][]byte, 10), // Default capacity of 10 schemas
	}
}

// Get retrieves cached schema JSON for the given key
// Returns the cached JSON and a boolean indicating if it was found
func (c *SchemaCache) Get(key uint64) ([]byte, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value, ok := c.cache[key]
	return value, ok
}

// Set stores schema JSON in the cache
func (c *SchemaCache) Set(key uint64, value []byte) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.cache[key] = value
}

// Clear empties the cache
func (c *SchemaCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// If the cache is already large, allocate a new one with the same capacity
	if len(c.cache) > 10 {
		c.cache = make(map[uint64][]byte, len(c.cache))
	} else {
		// Otherwise just clear the map
		for k := range c.cache {
			delete(c.cache, k)
		}
	}
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
