package processor

import (
	"testing"
	"time"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestSchemaCacheExpiration(t *testing.T) {
	// Create cache with short expiration time
	cache := NewSchemaCacheWithOptions(5, 10*time.Millisecond)

	// Create some test schemas
	schema1 := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"field1": {Type: "string"},
		},
	}

	schema2 := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"field2": {Type: "integer"},
		},
	}

	// Generate keys
	key1 := GenerateSchemaKey(schema1)
	key2 := GenerateSchemaKey(schema2)

	// Store data in cache
	cache.Set(key1, []byte(`{"type":"object","properties":{"field1":{"type":"string"}}}`))
	cache.Set(key2, []byte(`{"type":"object","properties":{"field2":{"type":"integer"}}}`))

	// Verify items were stored
	_, found1 := cache.Get(key1)
	_, found2 := cache.Get(key2)

	if !found1 || !found2 {
		t.Fatalf("Expected both items to be found in cache")
	}

	// Wait for items to expire
	time.Sleep(15 * time.Millisecond)

	// Force expiration check
	cache.lastCleanup = time.Now().Add(-2 * time.Minute)

	// Try to get an item - this should trigger cleanup
	_, _ = cache.Get(key1)

	// Check that the entries were expired
	_, found1 = cache.Get(key1)
	_, found2 = cache.Get(key2)

	if found1 || found2 {
		t.Errorf("Expected items to be expired but they were still found")
	}
}

func TestSchemaCacheLRUEviction(t *testing.T) {
	// Create cache with max size of 3
	cache := NewSchemaCacheWithOptions(3, 0)

	// Create test schemas
	var schemas []*schemaDomain.Schema
	var keys []uint64

	for i := 0; i < 5; i++ {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"field": {
					Type:        "string",
					Description: "Test " + string(rune('A'+i)),
				},
			},
		}
		// Make sure the append result is assigned back to the variable
		schemas = append(schemas, schema)
		_ = schemas // Explicitly use schemas to avoid unused result warning
		keys = append(keys, GenerateSchemaKey(schema))
	}

	// Set cache entries in sequence
	for i := 0; i < 5; i++ {
		cache.Set(keys[i], []byte(`{"test":"data"}`+string(rune('A'+i))))
	}

	// We should have only 3 entries due to LRU eviction
	if cache.GetSize() != 3 {
		t.Errorf("Expected cache size to be 3 due to LRU policy, got %d", cache.GetSize())
	}

	// The oldest 2 entries should be evicted
	_, found0 := cache.Get(keys[0])
	_, found1 := cache.Get(keys[1])

	if found0 || found1 {
		t.Errorf("Expected oldest entries to be evicted")
	}

	// The newest 3 entries should still be in the cache
	_, found2 := cache.Get(keys[2])
	_, found3 := cache.Get(keys[3])
	_, found4 := cache.Get(keys[4])

	if !found2 || !found3 || !found4 {
		t.Errorf("Expected newest entries to be in cache, found: %v, %v, %v", found2, found3, found4)
	}
}

func TestSchemaCacheAccessOrder(t *testing.T) {
	// Create cache with max size of 2
	cache := NewSchemaCacheWithOptions(2, 0)

	// Create 3 test schemas
	schema1 := &schemaDomain.Schema{Type: "object", Properties: map[string]schemaDomain.Property{"name": {Type: "string"}}}
	schema2 := &schemaDomain.Schema{Type: "object", Properties: map[string]schemaDomain.Property{"age": {Type: "integer"}}}
	schema3 := &schemaDomain.Schema{Type: "object", Properties: map[string]schemaDomain.Property{"email": {Type: "string"}}}

	key1 := GenerateSchemaKey(schema1)
	key2 := GenerateSchemaKey(schema2)
	key3 := GenerateSchemaKey(schema3)

	// Add first two items
	cache.Set(key1, []byte(`{"test":"data1"}`))
	cache.Set(key2, []byte(`{"test":"data2"}`))

	// Access the first item to make it most recently used
	cache.Get(key1)

	// Add third item (should evict the second item since it's least recently used)
	cache.Set(key3, []byte(`{"test":"data3"}`))

	// Check that most recently used item is still there
	_, found1 := cache.Get(key1)
	if !found1 {
		t.Errorf("Expected most recently used item to remain in cache")
	}

	// Check that least recently used item was evicted
	_, found2 := cache.Get(key2)
	if found2 {
		t.Errorf("Expected least recently used item to be evicted")
	}

	// Check that newest item is in the cache
	_, found3 := cache.Get(key3)
	if !found3 {
		t.Errorf("Expected newest item to be in cache")
	}
}
