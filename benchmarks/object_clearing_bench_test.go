package benchmarks

import (
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// generateLargeString creates a test string of the specified size
func generateLargeString(size int) string {
	var sb strings.Builder
	// Pre-allocate to avoid reallocations
	sb.Grow(size)
	
	// Generate a test pattern that repeats
	pattern := "This is a test string for benchmarking object clearing operations. "
	
	// Repeat until we get close to the target size
	for sb.Len() < size {
		sb.WriteString(pattern)
	}
	
	return sb.String()[:size]
}

// BenchmarkObjectClearing provides comprehensive benchmarks for object clearing operations
func BenchmarkObjectClearing(b *testing.B) {
	// Test with different content sizes to see the impact of our optimizations
	smallSize := 64       // 64 bytes
	mediumSize := 1024    // 1KB
	largeSize := 4096     // 4KB
	veryLargeSize := 32768 // 32KB
	
	// Generate test data
	smallContent := generateLargeString(smallSize)
	mediumContent := generateLargeString(mediumSize)
	largeContent := generateLargeString(largeSize)
	veryLargeContent := generateLargeString(veryLargeSize)
	
	// 1. Benchmark standard clearing (setting to empty string)
	b.Run("StandardClearing", func(b *testing.B) {
		b.Run("SmallContent", func(b *testing.B) {
			var s string = smallContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = smallContent
				s = ""
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("MediumContent", func(b *testing.B) {
			var s string = mediumContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = mediumContent
				s = ""
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("LargeContent", func(b *testing.B) {
			var s string = largeContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = largeContent
				s = ""
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("VeryLargeContent", func(b *testing.B) {
			var s string = veryLargeContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = veryLargeContent
				s = ""
				_ = s // Use s to prevent compiler optimizations
			}
		})
	})
	
	// 2. Benchmark optimized clearing using zeroString
	b.Run("OptimizedClearing", func(b *testing.B) {
		b.Run("SmallContent", func(b *testing.B) {
			var s string = smallContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = smallContent
				domain.ZeroString(&s)
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("MediumContent", func(b *testing.B) {
			var s string = mediumContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = mediumContent
				domain.ZeroString(&s)
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("LargeContent", func(b *testing.B) {
			var s string = largeContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = largeContent
				domain.ZeroString(&s)
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("VeryLargeContent", func(b *testing.B) {
			var s string = veryLargeContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = veryLargeContent
				domain.ZeroString(&s)
				_ = s // Use s to prevent compiler optimizations
			}
		})
	})
	
	// 3. Benchmark hybrid approach (our implementation)
	b.Run("HybridClearing", func(b *testing.B) {
		b.Run("SmallContent", func(b *testing.B) {
			var s string = smallContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = smallContent
				if len(s) > 1024 {
					domain.ZeroString(&s)
				} else {
					s = ""
				}
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("MediumContent", func(b *testing.B) {
			var s string = mediumContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = mediumContent
				if len(s) > 1024 {
					domain.ZeroString(&s)
				} else {
					s = ""
				}
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("LargeContent", func(b *testing.B) {
			var s string = largeContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = largeContent
				if len(s) > 1024 {
					domain.ZeroString(&s)
				} else {
					s = ""
				}
				_ = s // Use s to prevent compiler optimizations
			}
		})

		b.Run("VeryLargeContent", func(b *testing.B) {
			var s string = veryLargeContent
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = veryLargeContent
				if len(s) > 1024 {
					domain.ZeroString(&s)
				} else {
					s = ""
				}
				_ = s // Use s to prevent compiler optimizations
			}
		})
	})
	
	// 4. Benchmark standard Response pool clearing
	b.Run("ResponsePool", func(b *testing.B) {
		b.Run("StandardPool_SmallContent", func(b *testing.B) {
			pool := domain.NewResponsePool()
			resp := pool.Get()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp.Content = smallContent
				pool.Put(resp)
				resp = pool.Get()
			}
		})
		
		b.Run("StandardPool_LargeContent", func(b *testing.B) {
			pool := domain.NewResponsePool()
			resp := pool.Get()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp.Content = largeContent
				pool.Put(resp)
				resp = pool.Get()
			}
		})
		
		b.Run("StandardPool_VeryLargeContent", func(b *testing.B) {
			pool := domain.NewResponsePool()
			resp := pool.Get()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp.Content = veryLargeContent
				pool.Put(resp)
				resp = pool.Get()
			}
		})
	})
	
	// We've removed the separate optimized response pool implementation
	// and integrated the optimizations directly into the standard pool
	
	// 6. Compare full end-to-end operations
	b.Run("EndToEnd", func(b *testing.B) {
		b.Run("Standard_SmallContent", func(b *testing.B) {
			pool := domain.NewResponsePool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				response := pool.NewResponse(smallContent)
				_ = response.Content // Use the response
			}
		})
		
		b.Run("Standard_LargeContent", func(b *testing.B) {
			pool := domain.NewResponsePool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				response := pool.NewResponse(largeContent)
				_ = response.Content // Use the response
			}
		})
		
		// We've removed the separate optimized response pool implementation
		// and integrated the optimizations directly into the standard pool
	})
}