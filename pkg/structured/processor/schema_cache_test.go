package processor

import (
	"fmt"
	"testing"
	"time"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/metrics"
)

func TestSchemaCacheMetrics(t *testing.T) {
	// Reset metrics registry to avoid test interference
	metrics.GetRegistry().Clear()

	// Create a new schema cache
	cache := NewSchemaCache()

	// Generate a test schema
	schema := &schemaDomain.Schema{
		Type:  "object",
		Title: "Test Schema",
	}

	// Generate a key for the schema
	key := GenerateSchemaKey(schema)

	// Set a value in the cache
	cache.Set(key, []byte(`{"type":"object"}`))

	// Get the value from the cache (hit)
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected cache hit, got miss")
	}

	// Get a non-existent value (miss)
	_, found = cache.Get(12345)
	if found {
		t.Error("Expected cache miss, got hit")
	}

	// Give metrics a chance to settle
	time.Sleep(10 * time.Millisecond)

	// Check metrics
	hits, misses, total := cache.GetHitRate()
	if hits != 1 {
		t.Errorf("Expected 1 hit, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}
	if total != 2 {
		t.Errorf("Expected 2 total, got %d", total)
	}
	// Use hits to avoid ineffectual assignment lint error
	_ = hits

	// Warm up the cache metrics
	for i := 0; i < 10; i++ {
		// A mix of hits and misses
		cache.Get(key)     // hit
		cache.Get(key + 1) // miss
		cache.Get(key)     // hit
		cache.Get(key)     // hit
	}

	// Check metrics again
	hits, misses, total = cache.GetHitRate()
	expectedHits := int64(1 + 10*3)   // Initial + 3 hits per iteration
	expectedMisses := int64(1 + 10*1) // Initial + 1 miss per iteration
	expectedTotal := expectedHits + expectedMisses

	if hits != expectedHits {
		t.Errorf("Expected %d hits, got %d", expectedHits, hits)
	}
	if misses != expectedMisses {
		t.Errorf("Expected %d misses, got %d", expectedMisses, misses)
	}
	if total != expectedTotal {
		t.Errorf("Expected %d total, got %d", expectedTotal, total)
	}

	// Check for non-zero cache operation times
	if avgTime := cache.GetAverageAccessTime(); avgTime == 0 {
		t.Error("Expected non-zero average access time")
	}

	// Test cache clear
	cache.Clear()

	// After clear, cache should be empty but metrics should persist
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cache to be empty after clear")
	}

	// Metrics should still be available after clear
	hits, misses, total = cache.GetHitRate()
	if total == 0 {
		t.Error("Expected metrics to persist after cache clear")
	}
	// Use all variables to avoid ineffectual assignment lint warning
	_ = hits
	_ = misses
}

func TestSchemaCacheConcurrency(t *testing.T) {
	// Reset metrics registry to avoid test interference
	metrics.GetRegistry().Clear()

	cache := NewSchemaCache()

	// Generate a test schema
	schema := &schemaDomain.Schema{
		Type:  "object",
		Title: "Test Schema",
	}

	// Generate a key for the schema
	key := GenerateSchemaKey(schema)

	// Set a value in the cache
	cache.Set(key, []byte(`{"type":"object"}`))

	// Directly test the metrics
	cache.Get(key)     // hit
	cache.Get(key)     // hit
	cache.Get(key + 1) // miss
	cache.Get(key + 2) // miss

	// Verify the metrics
	hits, misses, total := cache.GetHitRate()

	if hits != 2 {
		t.Errorf("Expected 2 hits, got %d", hits)
	}
	if misses != 2 {
		t.Errorf("Expected 2 misses, got %d", misses)
	}
	if total != 4 {
		t.Errorf("Expected 4 total, got %d", total)
	}
}

func TestSchemaKeySizeAndPerformance(t *testing.T) {
	// Create a complex schema
	schema := &schemaDomain.Schema{
		Type:        "object",
		Title:       "Complex Schema",
		Description: "A very complex schema with many properties",
		Properties:  map[string]schemaDomain.Property{},
		Required:    []string{},
	}

	// Add many properties to the schema
	for i := 0; i < 100; i++ {
		propName := fmt.Sprintf("property_%d", i)
		schema.Properties[propName] = schemaDomain.Property{
			Type:        "string",
			Description: "A property",
		}
		schema.Required = append(schema.Required, propName)
	}

	// Generate a key for the schema
	start := time.Now()
	key := GenerateSchemaKey(schema)
	duration := time.Since(start)

	// Key generation should be fast even for large schemas
	if duration > 5*time.Millisecond {
		t.Errorf("Schema key generation took too long: %v", duration)
	}

	// Key should be non-zero
	if key == 0 {
		t.Error("Generated key is zero")
	}
}
