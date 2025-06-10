// ABOUTME: Tests for OpenAPI specification caching system
// ABOUTME: Verifies cache behavior, TTL, and thread safety

package web

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestOpenAPICache_Basic(t *testing.T) {
	cache := NewOpenAPICache(5 * time.Second)

	// Test empty cache
	spec, discovery, found := cache.Get("https://example.com/spec.json")
	if found {
		t.Error("Expected no result from empty cache")
	}
	if spec != nil || discovery != nil {
		t.Error("Expected nil results from empty cache")
	}

	// Add to cache
	testSpec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: InfoObject{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}
	testDiscovery := NewOperationDiscovery(testSpec)

	cache.Set("https://example.com/spec.json", testSpec, testDiscovery)

	// Retrieve from cache
	spec, discovery, found = cache.Get("https://example.com/spec.json")
	if !found {
		t.Error("Expected to find cached item")
	}
	if spec == nil || discovery == nil {
		t.Error("Expected non-nil results from cache")
		return
	}
	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", spec.Info.Title)
	}
}

func TestOpenAPICache_TTL(t *testing.T) {
	cache := NewOpenAPICache(100 * time.Millisecond)

	// Add to cache
	testSpec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: InfoObject{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}
	testDiscovery := NewOperationDiscovery(testSpec)

	cache.Set("https://example.com/spec.json", testSpec, testDiscovery)

	// Should be in cache immediately
	_, _, found := cache.Get("https://example.com/spec.json")
	if !found {
		t.Error("Expected to find cached item immediately after insertion")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, _, found = cache.Get("https://example.com/spec.json")
	if found {
		t.Error("Expected cached item to be expired")
	}
}

func TestOpenAPICache_Concurrent(t *testing.T) {
	cache := NewOpenAPICache(5 * time.Second)

	// Test concurrent access
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			url := "https://example.com/spec.json"

			// Half the goroutines read
			if id%2 == 0 {
				_, _, _ = cache.Get(url)
			} else {
				// Half write
				spec := &OpenAPISpec{
					OpenAPI: "3.0.0",
					Info: InfoObject{
						Title:   "Test API",
						Version: "1.0.0",
					},
				}
				discovery := NewOperationDiscovery(spec)
				cache.Set(url, spec, discovery)
			}
		}(i)
	}

	wg.Wait()
}

func TestOpenAPICache_Delete(t *testing.T) {
	cache := NewOpenAPICache(5 * time.Second)

	// Add to cache
	testSpec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: InfoObject{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}
	testDiscovery := NewOperationDiscovery(testSpec)

	url := "https://example.com/spec.json"
	cache.Set(url, testSpec, testDiscovery)

	// Verify it's cached
	_, _, found := cache.Get(url)
	if !found {
		t.Error("Expected to find cached item")
	}

	// Delete from cache
	cache.Delete(url)

	// Verify it's gone
	_, _, found = cache.Get(url)
	if found {
		t.Error("Expected item to be deleted from cache")
	}
}

func TestOpenAPICache_Clear(t *testing.T) {
	cache := NewOpenAPICache(5 * time.Second)

	// Add multiple items
	for i := 0; i < 5; i++ {
		spec := &OpenAPISpec{
			OpenAPI: "3.0.0",
			Info: InfoObject{
				Title:   "Test API",
				Version: "1.0.0",
			},
		}
		discovery := NewOperationDiscovery(spec)
		cache.Set(fmt.Sprintf("https://example.com/spec%d.json", i), spec, discovery)
	}

	// Verify stats
	total, active := cache.Stats()
	if total != 5 || active != 5 {
		t.Errorf("Expected 5 total and 5 active entries, got %d total, %d active", total, active)
	}

	// Clear cache
	cache.Clear()

	// Verify all cleared
	total, active = cache.Stats()
	if total != 0 || active != 0 {
		t.Errorf("Expected 0 entries after clear, got %d total, %d active", total, active)
	}
}

func TestOpenAPICache_Stats(t *testing.T) {
	cache := NewOpenAPICache(200 * time.Millisecond)

	// Add items with different expiration times
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: InfoObject{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}
	discovery := NewOperationDiscovery(spec)

	// Add first item
	cache.Set("https://example.com/spec1.json", spec, discovery)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Add second item
	cache.Set("https://example.com/spec2.json", spec, discovery)

	// Check stats immediately
	total, active := cache.Stats()
	if total != 2 || active != 2 {
		t.Errorf("Expected 2 total and 2 active, got %d total, %d active", total, active)
	}

	// Wait for first item to expire
	time.Sleep(150 * time.Millisecond)

	// Force a cleanup to ensure expired items are visible in stats
	cache.Cleanup()

	// Check stats - first should be expired and removed
	total, active = cache.Stats()
	if total != 1 || active != 1 {
		t.Errorf("Expected 1 total and 1 active after cleanup, got %d total, %d active", total, active)
	}
}

func TestGetOpenAPICache_Singleton(t *testing.T) {
	cache1 := GetOpenAPICache()
	cache2 := GetOpenAPICache()

	if cache1 != cache2 {
		t.Error("Expected GetOpenAPICache to return the same instance")
	}
}
