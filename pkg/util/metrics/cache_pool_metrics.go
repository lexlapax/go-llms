package metrics

import (
	"sync/atomic"
	"time"
)

// CacheMetrics tracks cache hit/miss statistics
type CacheMetrics struct {
	name       string
	hits       int64
	misses     int64
	hitRate    *RatioCounter
	accessTime *Timer
}

// NewCacheMetrics creates a new cache metrics collector with the given name
func NewCacheMetrics(name string) *CacheMetrics {
	registry := GetRegistry()

	return &CacheMetrics{
		name:       name,
		hits:       0,
		misses:     0,
		hitRate:    registry.GetOrCreateRatioCounter(name + ".hit_rate"),
		accessTime: registry.GetOrCreateTimer(name + ".access_time"),
	}
}

// RecordHit records a cache hit
func (c *CacheMetrics) RecordHit() {
	atomic.AddInt64(&c.hits, 1)
	c.hitRate.IncrementNumerator()
	c.hitRate.IncrementDenominator()
}

// RecordMiss records a cache miss
func (c *CacheMetrics) RecordMiss() {
	atomic.AddInt64(&c.misses, 1)
	c.hitRate.IncrementDenominator()
}

// RecordAccessTime records the time taken to access the cache
func (c *CacheMetrics) RecordAccessTime(duration time.Duration) {
	c.accessTime.RecordDuration(duration)
}

// TimeAccess times a cache access operation and records the result
func (c *CacheMetrics) TimeAccess(fn func() (interface{}, bool)) (interface{}, bool) {
	c.accessTime.Start()
	result, hit := fn()
	c.accessTime.Stop()

	if hit {
		c.RecordHit()
	} else {
		c.RecordMiss()
	}

	return result, hit
}

// GetHitRate returns the current cache hit rate
func (c *CacheMetrics) GetHitRate() float64 {
	return c.hitRate.GetRatio()
}

// GetHitsMisses returns the current cache hits, misses, and total accesses
func (c *CacheMetrics) GetHitsMisses() (int64, int64, int64) {
	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)
	return hits, misses, hits + misses
}

// GetAverageAccessTime returns the average time taken to access the cache
func (c *CacheMetrics) GetAverageAccessTime() time.Duration {
	return c.accessTime.GetAverageDuration()
}

// Reset resets all cache metrics to zero
func (c *CacheMetrics) Reset() {
	atomic.StoreInt64(&c.hits, 0)
	atomic.StoreInt64(&c.misses, 0)

	// Create new metrics in the registry with the same names
	registry := GetRegistry()
	c.hitRate = registry.GetOrCreateRatioCounter(c.name + ".hit_rate")
	c.accessTime = registry.GetOrCreateTimer(c.name + ".access_time")
}

// PoolMetrics tracks memory pool statistics
type PoolMetrics struct {
	name           string
	poolSize       int64
	allocations    int64
	returns        int64
	waitTime       *Timer
	allocationTime *Timer
}

// NewPoolMetrics creates a new pool metrics collector with the given name
func NewPoolMetrics(name string) *PoolMetrics {
	registry := GetRegistry()

	return &PoolMetrics{
		name:           name,
		poolSize:       0,
		allocations:    0,
		returns:        0,
		waitTime:       registry.GetOrCreateTimer(name + ".wait_time"),
		allocationTime: registry.GetOrCreateTimer(name + ".allocation_time"),
	}
}

// RecordAllocation records an object allocation from the pool
func (p *PoolMetrics) RecordAllocation() {
	atomic.AddInt64(&p.allocations, 1)
}

// RecordReturn records an object being returned to the pool
func (p *PoolMetrics) RecordReturn() {
	atomic.AddInt64(&p.returns, 1)
}

// IncrementPoolSize adds to the current pool size
func (p *PoolMetrics) IncrementPoolSize(delta int64) {
	atomic.AddInt64(&p.poolSize, delta)
}

// SetPoolSize sets the current pool size
func (p *PoolMetrics) SetPoolSize(size int64) {
	atomic.StoreInt64(&p.poolSize, size)
}

// RecordWaitTime records the time waiting for an object from the pool
func (p *PoolMetrics) RecordWaitTime(duration time.Duration) {
	p.waitTime.RecordDuration(duration)
}

// RecordAllocationTime records the time taken to allocate a new object
func (p *PoolMetrics) RecordAllocationTime(duration time.Duration) {
	p.allocationTime.RecordDuration(duration)
}

// TimeAllocation times the allocation of a new object
func (p *PoolMetrics) TimeAllocation(fn func() interface{}) interface{} {
	p.allocationTime.Start()
	result := fn()
	p.allocationTime.Stop()

	p.RecordAllocation()
	return result
}

// GetPoolSize returns the current pool size
func (p *PoolMetrics) GetPoolSize() int64 {
	return atomic.LoadInt64(&p.poolSize)
}

// GetAllocationCount returns the number of allocations and returns
func (p *PoolMetrics) GetAllocationCount() (int64, int64) {
	allocations := atomic.LoadInt64(&p.allocations)
	returns := atomic.LoadInt64(&p.returns)
	return allocations, returns
}

// GetAverageWaitTime returns the average time waiting for an object
func (p *PoolMetrics) GetAverageWaitTime() time.Duration {
	return p.waitTime.GetAverageDuration()
}

// GetAverageAllocationTime returns the average time to allocate a new object
func (p *PoolMetrics) GetAverageAllocationTime() time.Duration {
	return p.allocationTime.GetAverageDuration()
}

// Reset resets all pool metrics to zero
func (p *PoolMetrics) Reset() {
	atomic.StoreInt64(&p.poolSize, 0)
	atomic.StoreInt64(&p.allocations, 0)
	atomic.StoreInt64(&p.returns, 0)

	// Create new metrics in the registry with the same names
	registry := GetRegistry()
	p.waitTime = registry.GetOrCreateTimer(p.name + ".wait_time")
	p.allocationTime = registry.GetOrCreateTimer(p.name + ".allocation_time")
}
