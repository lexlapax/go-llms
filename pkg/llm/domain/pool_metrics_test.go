package domain

import (
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/util/metrics"
)

func TestResponsePoolMetrics(t *testing.T) {
	// Reset metrics registry
	metrics.GetRegistry().Clear()
	
	// Get a new response pool
	pool := NewResponsePoolWithMetrics()
	
	// Check initial metrics
	size, allocated, returned := pool.GetMetrics()
	if size != 0 || allocated != 0 || returned != 0 {
		t.Errorf("Expected initial metrics to be 0/0/0, got %d/%d/%d", size, allocated, returned)
	}
	
	// Get a response from the pool (allocation)
	resp := pool.Get()
	
	// Check metrics after allocation
	size, allocated, returned = pool.GetMetrics()
	if allocated != 1 {
		t.Errorf("Expected 1 allocation, got %d", allocated)
	}
	
	// Return response to the pool
	pool.Put(resp)
	
	// Check metrics after return
	size, allocated, returned = pool.GetMetrics()
	if returned != 1 {
		t.Errorf("Expected 1 return, got %d", returned)
	}
	
	// Get allocation time metrics
	allocTime := pool.GetAverageAllocationTime()
	if allocTime == 0 {
		t.Error("Expected non-zero allocation time")
	}
	
	// Test the NewResponse method
	response := pool.NewResponse("test content")
	if response.Content != "test content" {
		t.Errorf("Expected content to be 'test content', got '%s'", response.Content)
	}
	
	// Metrics should show one more allocation and return
	size, allocated, returned = pool.GetMetrics()
	if allocated != 2 || returned != 2 {
		t.Errorf("Expected 2 allocations and 2 returns, got %d/%d", allocated, returned)
	}
}

func TestTokenPoolMetrics(t *testing.T) {
	// Reset metrics registry
	metrics.GetRegistry().Clear()
	
	// Get a new token pool
	pool := NewTokenPoolWithMetrics()
	
	// Check initial metrics
	size, allocated, returned := pool.GetMetrics()
	if size != 0 || allocated != 0 || returned != 0 {
		t.Errorf("Expected initial metrics to be 0/0/0, got %d/%d/%d", size, allocated, returned)
	}
	
	// Get a token from the pool (allocation)
	token := pool.Get()
	
	// Check metrics after allocation
	size, allocated, returned = pool.GetMetrics()
	if allocated != 1 {
		t.Errorf("Expected 1 allocation, got %d", allocated)
	}
	
	// Return token to the pool
	pool.Put(token)
	
	// Check metrics after return
	size, allocated, returned = pool.GetMetrics()
	if returned != 1 {
		t.Errorf("Expected 1 return, got %d", returned)
	}
	
	// Test the NewToken method
	token2 := pool.NewToken("test", true)
	if token2.Text != "test" || !token2.Finished {
		t.Errorf("Expected token text 'test' and finished true, got '%s' and %v", token2.Text, token2.Finished)
	}
	
	// Metrics should show one more allocation and return
	size, allocated, returned = pool.GetMetrics()
	if allocated != 2 || returned != 2 {
		t.Errorf("Expected 2 allocations and 2 returns, got %d/%d", allocated, returned)
	}
}

func TestChannelPoolMetrics(t *testing.T) {
	// Reset metrics registry
	metrics.GetRegistry().Clear()
	
	// Get a new channel pool
	pool := NewChannelPoolWithMetrics()
	
	// Check initial metrics
	size, allocated, returned := pool.GetMetrics()
	if size != 0 || allocated != 0 || returned != 0 {
		t.Errorf("Expected initial metrics to be 0/0/0, got %d/%d/%d", size, allocated, returned)
	}
	
	// Get a channel from the pool (allocation)
	ch := pool.Get()
	
	// Check metrics after allocation
	size, allocated, returned = pool.GetMetrics()
	if allocated != 1 {
		t.Errorf("Expected 1 allocation, got %d", allocated)
	}
	
	// Return channel to the pool
	pool.Put(ch)
	
	// Check metrics after return
	size, allocated, returned = pool.GetMetrics()
	if returned != 1 {
		t.Errorf("Expected 1 return, got %d", returned)
	}
	
	// Test the GetResponseStream method
	stream, ch2 := pool.GetResponseStream()
	if stream == nil || ch2 == nil {
		t.Error("Expected non-nil stream and channel")
	}
	
	// Metrics should show one more allocation
	size, allocated, returned = pool.GetMetrics()
	if allocated != 2 {
		t.Errorf("Expected 2 allocations, got %d", allocated)
	}
	
	// Return the channel
	pool.Put(ch2)
	
	// Metrics should show one more return
	size, allocated, returned = pool.GetMetrics()
	if returned != 2 {
		t.Errorf("Expected 2 returns, got %d", returned)
	}
}

func TestPoolMetricsConcurrent(t *testing.T) {
	// Reset metrics registry
	metrics.GetRegistry().Clear()
	
	// Create pool with metrics
	pool := NewResponsePoolWithMetrics()
	
	// Run concurrent allocations and returns
	const workers = 5
	const iterations = 10
	
	done := make(chan struct{})
	
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				resp := pool.Get()
				time.Sleep(time.Millisecond) // Simulate work
				pool.Put(resp)
			}
			done <- struct{}{}
		}()
	}
	
	// Wait for all workers to finish
	for i := 0; i < workers; i++ {
		<-done
	}
	
	// Check metrics
	_, allocated, returned := pool.GetMetrics()
	expectedAllocs := workers * iterations
	expectedReturns := workers * iterations
	
	if allocated != int64(expectedAllocs) {
		t.Errorf("Expected %d allocations, got %d", expectedAllocs, allocated)
	}
	if returned != int64(expectedReturns) {
		t.Errorf("Expected %d returns, got %d", expectedReturns, returned)
	}
}