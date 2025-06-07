// ABOUTME: Performance benchmarks for the API Client Tool
// ABOUTME: Measures execution speed, memory allocation, and concurrent performance

package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// BenchmarkAPIClientTool_SimpleGET benchmarks simple GET requests
func BenchmarkAPIClientTool_SimpleGET(b *testing.B) {
	// Create test server that returns a simple JSON response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      123,
			"message": "Hello, World!",
			"timestamp": time.Now().Unix(),
		})
	}))
	defer server.Close()

	tool := NewAPIClientTool()
	state := domain.NewState()
	ctx := &domain.ToolContext{
		Context:   context.Background(),
		State:     domain.NewStateReader(state),
		RunID:     "bench-run",
		StartTime: time.Now(),
		Agent: domain.AgentInfo{
			ID:   "bench-agent",
			Name: "Benchmark Agent",
			Type: domain.AgentTypeCustom,
		},
	}

	params := map[string]interface{}{
		"base_url": server.URL,
		"endpoint": "/test",
		"method":   "GET",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}

// BenchmarkAPIClientTool_POST benchmarks POST requests with JSON body
func BenchmarkAPIClientTool_POST(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      456,
			"created": true,
			"data":    body,
		})
	}))
	defer server.Close()

	tool := NewAPIClientTool()
	state := domain.NewState()
	ctx := &domain.ToolContext{
		Context:   context.Background(),
		State:     domain.NewStateReader(state),
		RunID:     "bench-run",
		StartTime: time.Now(),
		Agent: domain.AgentInfo{
			ID:   "bench-agent",
			Name: "Benchmark Agent",
			Type: domain.AgentTypeCustom,
		},
	}

	params := map[string]interface{}{
		"base_url": server.URL,
		"endpoint": "/items",
		"method":   "POST",
		"body": map[string]interface{}{
			"name":        "Test Item",
			"description": "Benchmark test",
			"value":       123.45,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}

// BenchmarkAPIClientTool_PathParams benchmarks path parameter substitution
func BenchmarkAPIClientTool_PathParams(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"path": r.URL.Path,
		})
	}))
	defer server.Close()

	tool := NewAPIClientTool()
	state := domain.NewState()
	ctx := &domain.ToolContext{
		Context:   context.Background(),
		State:     domain.NewStateReader(state),
		RunID:     "bench-run",
		StartTime: time.Now(),
		Agent: domain.AgentInfo{
			ID:   "bench-agent",
			Name: "Benchmark Agent",
			Type: domain.AgentTypeCustom,
		},
	}

	params := map[string]interface{}{
		"base_url": server.URL,
		"endpoint": "/users/{user_id}/posts/{post_id}/comments/{comment_id}",
		"method":   "GET",
		"path_params": map[string]string{
			"user_id":    "alice",
			"post_id":    "42",
			"comment_id": "123",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}

// BenchmarkAPIClientTool_LargeResponse benchmarks handling of large JSON responses
func BenchmarkAPIClientTool_LargeResponse(b *testing.B) {
	// Create a large response with many items
	largeData := make([]map[string]interface{}, 1000)
	for i := range largeData {
		largeData[i] = map[string]interface{}{
			"id":          i,
			"name":        "Item " + string(rune(i)),
			"description": "This is a longer description for the item to make the response larger",
			"value":       float64(i) * 1.23,
			"active":      i%2 == 0,
			"tags":        []string{"tag1", "tag2", "tag3"},
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": len(largeData),
			"items": largeData,
		})
	}))
	defer server.Close()

	tool := NewAPIClientTool()
	state := domain.NewState()
	ctx := &domain.ToolContext{
		Context:   context.Background(),
		State:     domain.NewStateReader(state),
		RunID:     "bench-run",
		StartTime: time.Now(),
		Agent: domain.AgentInfo{
			ID:   "bench-agent",
			Name: "Benchmark Agent",
			Type: domain.AgentTypeCustom,
		},
	}

	params := map[string]interface{}{
		"base_url": server.URL,
		"endpoint": "/large-data",
		"method":   "GET",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}

// BenchmarkAPIClientTool_Concurrent benchmarks concurrent API calls
func BenchmarkAPIClientTool_Concurrent(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      r.URL.Query().Get("id"),
			"message": "Concurrent response",
		})
	}))
	defer server.Close()

	tool := NewAPIClientTool()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			state := domain.NewState()
			ctx := &domain.ToolContext{
				Context:   context.Background(),
				State:     domain.NewStateReader(state),
				RunID:     "bench-run",
				StartTime: time.Now(),
				Agent: domain.AgentInfo{
					ID:   "bench-agent",
					Name: "Benchmark Agent",
					Type: domain.AgentTypeCustom,
				},
			}

			params := map[string]interface{}{
				"base_url": server.URL,
				"endpoint": "/concurrent",
				"method":   "GET",
				"query_params": map[string]string{
					"id": string(rune(i)),
				},
			}

			_, err := tool.Execute(ctx, params)
			if err != nil {
				b.Fatalf("Execute failed: %v", err)
			}
			i++
		}
	})
}