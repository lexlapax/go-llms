package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestCacheMetrics(t *testing.T) {
	// Reset registry
	globalRegistry = &Registry{
		counters:      make(map[string]*Counter),
		gauges:        make(map[string]*Gauge),
		ratioCounters: make(map[string]*RatioCounter),
		timers:        make(map[string]*Timer),
	}

	// Create cache metrics
	metrics := NewCacheMetrics("test_cache")

	// Test initial values
	if metrics.GetHitRate() != 0.0 {
		t.Errorf("Expected initial hit rate to be 0.0, got %f", metrics.GetHitRate())
	}

	if hits, misses, total := metrics.GetHitsMisses(); hits != 0 || misses != 0 || total != 0 {
		t.Errorf("Expected initial hits/misses/total to be 0/0/0, got %d/%d/%d", hits, misses, total)
	}

	// Test recording hits and misses
	metrics.RecordHit()
	metrics.RecordMiss()
	metrics.RecordHit()

	// Expected: 2 hits, 1 miss, 3 total, 0.6667 hit rate
	expectedHitRate := 2.0 / 3.0
	if hitRate := metrics.GetHitRate(); hitRate < expectedHitRate-0.01 || hitRate > expectedHitRate+0.01 {
		t.Errorf("Expected hit rate to be %.4f, got %.4f", expectedHitRate, hitRate)
	}

	hits, misses, total := metrics.GetHitsMisses()
	if hits != 2 || misses != 1 || total != 3 {
		t.Errorf("Expected hits/misses/total to be 2/1/3, got %d/%d/%d", hits, misses, total)
	}

	// Reset metrics
	metrics.Reset()
	if hits, misses, total := metrics.GetHitsMisses(); hits != 0 || misses != 0 || total != 0 {
		t.Errorf("Expected hits/misses/total to be 0/0/0 after reset, got %d/%d/%d", hits, misses, total)
	}
}

func TestCacheMetricsConcurrency(t *testing.T) {
	// Reset registry
	globalRegistry = &Registry{
		counters:      make(map[string]*Counter),
		gauges:        make(map[string]*Gauge),
		ratioCounters: make(map[string]*RatioCounter),
		timers:        make(map[string]*Timer),
	}

	metrics := NewCacheMetrics("test_cache_concurrent")

	var wg sync.WaitGroup
	workers := 10
	hitsPerWorker := 7
	missesPerWorker := 3

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < hitsPerWorker; j++ {
				metrics.RecordHit()
			}
			for j := 0; j < missesPerWorker; j++ {
				metrics.RecordMiss()
			}
		}()
	}

	wg.Wait()

	hits, misses, total := metrics.GetHitsMisses()
	expectedHits := workers * hitsPerWorker
	expectedMisses := workers * missesPerWorker
	expectedTotal := expectedHits + expectedMisses

	if hits != int64(expectedHits) || misses != int64(expectedMisses) || total != int64(expectedTotal) {
		t.Errorf("Expected hits/misses/total to be %d/%d/%d, got %d/%d/%d",
			expectedHits, expectedMisses, expectedTotal, hits, misses, total)
	}

	expectedHitRate := float64(expectedHits) / float64(expectedTotal)
	if hitRate := metrics.GetHitRate(); hitRate < expectedHitRate-0.01 || hitRate > expectedHitRate+0.01 {
		t.Errorf("Expected hit rate to be %.4f, got %.4f", expectedHitRate, hitRate)
	}
}

func TestPoolMetrics(t *testing.T) {
	// Reset registry
	globalRegistry = &Registry{
		counters:      make(map[string]*Counter),
		gauges:        make(map[string]*Gauge),
		ratioCounters: make(map[string]*RatioCounter),
		timers:        make(map[string]*Timer),
	}

	metrics := NewPoolMetrics("test_pool")

	// Test initial values
	if size := metrics.GetPoolSize(); size != 0 {
		t.Errorf("Expected initial pool size to be 0, got %d", size)
	}

	if allocated, returned := metrics.GetAllocationCount(); allocated != 0 || returned != 0 {
		t.Errorf("Expected initial allocations/returns to be 0/0, got %d/%d", allocated, returned)
	}

	if avgWaitTime := metrics.GetAverageWaitTime(); avgWaitTime != 0 {
		t.Errorf("Expected initial average wait time to be 0, got %v", avgWaitTime)
	}

	// Record pool operations
	metrics.RecordAllocation()
	metrics.RecordAllocation()
	metrics.RecordReturn()
	metrics.IncrementPoolSize(5)
	metrics.RecordWaitTime(100 * time.Millisecond)
	metrics.RecordWaitTime(300 * time.Millisecond)

	// Check values
	if size := metrics.GetPoolSize(); size != 5 {
		t.Errorf("Expected pool size to be 5, got %d", size)
	}

	if allocated, returned := metrics.GetAllocationCount(); allocated != 2 || returned != 1 {
		t.Errorf("Expected allocations/returns to be 2/1, got %d/%d", allocated, returned)
	}

	if avgWaitTime := metrics.GetAverageWaitTime(); avgWaitTime != 200*time.Millisecond {
		t.Errorf("Expected average wait time to be 200ms, got %v", avgWaitTime)
	}

	// Reset metrics
	metrics.Reset()
	if size := metrics.GetPoolSize(); size != 0 {
		t.Errorf("Expected pool size to be 0 after reset, got %d", size)
	}
}

func TestPoolMetricsConcurrency(t *testing.T) {
	// Reset registry
	globalRegistry = &Registry{
		counters:      make(map[string]*Counter),
		gauges:        make(map[string]*Gauge),
		ratioCounters: make(map[string]*RatioCounter),
		timers:        make(map[string]*Timer),
	}

	metrics := NewPoolMetrics("test_pool_concurrent")

	var wg sync.WaitGroup
	workers := 10
	allocsPerWorker := 10
	returnsPerWorker := 7

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < allocsPerWorker; j++ {
				metrics.RecordAllocation()
			}
			for j := 0; j < returnsPerWorker; j++ {
				metrics.RecordReturn()
			}
			metrics.IncrementPoolSize(1)
			metrics.RecordWaitTime(100 * time.Millisecond)
		}()
	}

	wg.Wait()

	allocated, returned := metrics.GetAllocationCount()
	expectedAllocs := workers * allocsPerWorker
	expectedReturns := workers * returnsPerWorker

	if allocated != int64(expectedAllocs) || returned != int64(expectedReturns) {
		t.Errorf("Expected allocations/returns to be %d/%d, got %d/%d",
			expectedAllocs, expectedReturns, allocated, returned)
	}

	expectedPoolSize := workers // Each worker incremented by 1
	if size := metrics.GetPoolSize(); size != int64(expectedPoolSize) {
		t.Errorf("Expected pool size to be %d, got %d", expectedPoolSize, size)
	}

	// Each worker added one wait time measurement of 100ms
	expectedAvg := 100 * time.Millisecond
	avgWaitTime := metrics.GetAverageWaitTime()
	if avgWaitTime < expectedAvg-time.Millisecond || avgWaitTime > expectedAvg+time.Millisecond {
		t.Errorf("Expected average wait time to be %v, got %v", expectedAvg, avgWaitTime)
	}
}
