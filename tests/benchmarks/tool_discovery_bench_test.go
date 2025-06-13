// ABOUTME: Benchmark tests for tool discovery API to ensure no performance regression
// ABOUTME: Measures discovery operations performance for optimization tracking

package benchmarks

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
)

func BenchmarkDiscovery_ListTools(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toolList := discovery.ListTools()
		if len(toolList) == 0 {
			b.Fatal("Expected tools to be available")
		}
	}
}

func BenchmarkDiscovery_SearchTools(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := discovery.SearchTools("json")
		_ = results // Use results to prevent optimization
	}
}

func BenchmarkDiscovery_ListByCategory(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := discovery.ListByCategory("math")
		_ = results // Use results to prevent optimization
	}
}

func BenchmarkDiscovery_GetToolSchema(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema, err := discovery.GetToolSchema("calculator")
		if err != nil {
			b.Fatal(err)
		}
		_ = schema // Use schema to prevent optimization
	}
}

func BenchmarkDiscovery_GetToolExamples(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		examples, err := discovery.GetToolExamples("calculator")
		if err != nil {
			b.Fatal(err)
		}
		_ = examples // Use examples to prevent optimization
	}
}

func BenchmarkDiscovery_GetToolHelp(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		help, err := discovery.GetToolHelp("calculator")
		if err != nil {
			b.Fatal(err)
		}
		_ = help // Use help to prevent optimization
	}
}

func BenchmarkGetToolMetadata(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metadata := tools.GetToolMetadata()
		if len(metadata) == 0 {
			b.Fatal("Expected metadata to be available")
		}
	}
}

// Comparative benchmarks against legacy registry
func BenchmarkLegacyRegistry_List(b *testing.B) {
	// This benchmarks the legacy tools.Tools.List() for comparison
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This would require importing the legacy tools package
		// Placeholder for performance comparison
		// allTools := tools.Tools.List()
		// _ = allTools
	}
	b.Skip("Legacy comparison requires tool imports")
}

// Memory allocation benchmarks
func BenchmarkDiscovery_ListTools_Allocs(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toolList := discovery.ListTools()
		_ = toolList
	}
}

func BenchmarkDiscovery_SearchTools_Allocs(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := discovery.SearchTools("json")
		_ = results
	}
}

// Concurrent access benchmarks
func BenchmarkDiscovery_ConcurrentAccess(b *testing.B) {
	discovery := tools.NewDiscovery()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix different operations to test concurrent safety
			switch b.N % 4 {
			case 0:
				toolList := discovery.ListTools()
				_ = toolList
			case 1:
				results := discovery.SearchTools("data")
				_ = results
			case 2:
				category := discovery.ListByCategory("web")
				_ = category
			case 3:
				schema, _ := discovery.GetToolSchema("calculator")
				_ = schema
			}
		}
	})
}
