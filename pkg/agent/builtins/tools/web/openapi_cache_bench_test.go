// ABOUTME: Benchmarks for OpenAPI caching system to measure performance improvements
// ABOUTME: Compares cached vs uncached spec fetching and operation discovery

package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// Mock OpenAPI spec for benchmarking
var benchmarkSpec = []byte(`{
	"openapi": "3.0.0",
	"info": {"title": "Benchmark API", "version": "1.0.0"},
	"paths": {
		"/users": {
			"get": {
				"operationId": "listUsers",
				"summary": "List users",
				"parameters": [
					{"name": "limit", "in": "query", "schema": {"type": "integer"}}
				]
			},
			"post": {
				"operationId": "createUser",
				"summary": "Create user",
				"requestBody": {
					"required": true,
					"content": {
						"application/json": {
							"schema": {
								"type": "object",
								"properties": {
									"name": {"type": "string"},
									"email": {"type": "string"}
								},
								"required": ["name", "email"]
							}
						}
					}
				}
			}
		},
		"/users/{id}": {
			"get": {
				"operationId": "getUser",
				"summary": "Get user by ID",
				"parameters": [
					{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}
				]
			}
		}
	}
}`)

func BenchmarkOpenAPIParser_FetchSpec_NoCache(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(benchmarkSpec)
	}))
	defer server.Close()

	// Create parser without cache
	parser := &OpenAPIParser{
		client:  &http.Client{Timeout: 30 * time.Second},
		timeout: 30 * time.Second,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Direct fetch without cache
			resp, err := parser.client.Get(server.URL)
			if err != nil {
				b.Fatal(err)
			}
			body := make([]byte, len(benchmarkSpec))
			_, _ = resp.Body.Read(body)
			_ = resp.Body.Close()

			_, err = parser.ParseSpec(body, server.URL)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkOpenAPIParser_FetchSpec_WithCache(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(benchmarkSpec)
	}))
	defer server.Close()

	// Create parser with cache
	parser := NewOpenAPIParser()

	// Prime the cache
	_, _ = parser.FetchSpec(server.URL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := parser.FetchSpec(server.URL)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkOperationDiscovery_EnumerateOperations_NoCache(b *testing.B) {
	parser := NewOpenAPIParser()
	spec, _ := parser.ParseSpec(benchmarkSpec, "test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		discovery := &OperationDiscovery{
			spec:      spec,
			validator: validation.NewValidator(),
		}
		_ = discovery.EnumerateOperations()
	}
}

func BenchmarkOperationDiscovery_EnumerateOperations_WithCache(b *testing.B) {
	parser := NewOpenAPIParser()
	spec, _ := parser.ParseSpec(benchmarkSpec, "test")
	discovery := NewOperationDiscovery(spec)

	// Prime the cache
	discovery.EnumerateOperations()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = discovery.EnumerateOperations()
	}
}

func BenchmarkOperationIndex_FindOperation(b *testing.B) {
	parser := NewOpenAPIParser()
	spec, _ := parser.ParseSpec(benchmarkSpec, "test")
	discovery := NewOperationDiscovery(spec)
	ops := discovery.EnumerateOperations()
	index := NewOperationIndex(ops)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Alternate between different operations
			if i := b.N % 3; i == 0 {
				index.FindOperation("GET", "/users")
			} else if i == 1 {
				index.FindOperation("POST", "/users")
			} else {
				index.FindOperation("GET", "/users/{id}")
			}
		}
	})
}

func BenchmarkOperationDiscovery_LinearSearch(b *testing.B) {
	parser := NewOpenAPIParser()
	spec, _ := parser.ParseSpec(benchmarkSpec, "test")
	discovery := NewOperationDiscovery(spec)
	ops := discovery.EnumerateOperations()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Linear search through operations
			target := "/users/{id}"
			method := "GET"
			for _, op := range ops {
				if op.Path == target && op.Method == method {
					break
				}
			}
		}
	})
}

func BenchmarkCache_ConcurrentAccess(b *testing.B) {
	cache := NewOpenAPICache(5 * time.Minute)

	// Create multiple specs
	specs := make([]*OpenAPISpec, 10)
	discoveries := make([]*OperationDiscovery, 10)

	for i := 0; i < 10; i++ {
		spec := &OpenAPISpec{
			OpenAPI: "3.0.0",
			Info: InfoObject{
				Title:   fmt.Sprintf("API %d", i),
				Version: "1.0.0",
			},
		}
		specs[i] = spec
		discoveries[i] = NewOperationDiscovery(spec)
		cache.Set(fmt.Sprintf("https://api%d.com/spec.json", i), spec, discoveries[i])
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Random read/write operations
			i := b.N % 10
			url := fmt.Sprintf("https://api%d.com/spec.json", i)

			if i%3 == 0 {
				// Write
				cache.Set(url, specs[i], discoveries[i])
			} else {
				// Read
				cache.Get(url)
			}
		}
	})
}
