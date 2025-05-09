// Package metrics provides utilities for collecting and reporting performance metrics
// in the Go-LLMs project. It supports various metric types including counters, gauges,
// ratio counters, and timers, all with thread-safe operations.
package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Counter is a monotonically increasing counter
type Counter struct {
	name  string
	value int64
}

// NewCounter creates a new counter with a given name
func NewCounter(name string) *Counter {
	return &Counter{
		name:  name,
		value: 0,
	}
}

// Increment increments the counter by 1
func (c *Counter) Increment() {
	atomic.AddInt64(&c.value, 1)
}

// IncrementBy increments the counter by the specified value
func (c *Counter) IncrementBy(value int64) {
	atomic.AddInt64(&c.value, value)
}

// GetValue returns the current value of the counter
func (c *Counter) GetValue() int64 {
	return atomic.LoadInt64(&c.value)
}

// Gauge is a metric that can go up and down
type Gauge struct {
	name  string
	value float64
	mu    sync.RWMutex
}

// NewGauge creates a new gauge with a given name
func NewGauge(name string) *Gauge {
	return &Gauge{
		name:  name,
		value: 0,
	}
}

// Set sets the gauge to a specific value
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Increment increments the gauge by 1
func (g *Gauge) Increment() {
	g.Add(1)
}

// Decrement decrements the gauge by 1
func (g *Gauge) Decrement() {
	g.Subtract(1)
}

// Add adds the given value to the gauge
func (g *Gauge) Add(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += value
}

// Subtract subtracts the given value from the gauge
func (g *Gauge) Subtract(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value -= value
}

// GetValue returns the current value of the gauge
func (g *Gauge) GetValue() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// RatioCounter tracks a ratio between two counters (e.g., cache hit rate)
type RatioCounter struct {
	name        string
	numerator   int64
	denominator int64
}

// NewRatioCounter creates a new ratio counter with a given name
func NewRatioCounter(name string) *RatioCounter {
	return &RatioCounter{
		name:        name,
		numerator:   0,
		denominator: 0,
	}
}

// IncrementNumerator increments the numerator by 1
func (r *RatioCounter) IncrementNumerator() {
	atomic.AddInt64(&r.numerator, 1)
}

// IncrementDenominator increments the denominator by 1
func (r *RatioCounter) IncrementDenominator() {
	atomic.AddInt64(&r.denominator, 1)
}

// IncrementNumeratorBy increments the numerator by the specified value
func (r *RatioCounter) IncrementNumeratorBy(value int64) {
	atomic.AddInt64(&r.numerator, value)
}

// IncrementDenominatorBy increments the denominator by the specified value
func (r *RatioCounter) IncrementDenominatorBy(value int64) {
	atomic.AddInt64(&r.denominator, value)
}

// GetRatio returns the current ratio (numerator/denominator)
// Returns 0 if denominator is 0 to avoid division by zero
func (r *RatioCounter) GetRatio() float64 {
	num := atomic.LoadInt64(&r.numerator)
	den := atomic.LoadInt64(&r.denominator)
	
	if den == 0 {
		return 0.0
	}
	
	return float64(num) / float64(den)
}

// GetValues returns the raw numerator and denominator values
func (r *RatioCounter) GetValues() (int64, int64) {
	return atomic.LoadInt64(&r.numerator), atomic.LoadInt64(&r.denominator)
}

// Timer tracks execution duration of operations
type Timer struct {
	name         string
	startTime    time.Time
	count        int64
	totalTime    int64 // nanoseconds
	lastDuration int64 // nanoseconds
	running      bool
	mu           sync.Mutex
}

// NewTimer creates a new timer with a given name
func NewTimer(name string) *Timer {
	return &Timer{
		name:      name,
		count:     0,
		totalTime: 0,
		running:   false,
	}
}

// Start starts the timer
func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if !t.running {
		t.startTime = time.Now()
		t.running = true
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if !t.running {
		return 0
	}
	
	duration := time.Since(t.startTime)
	t.recordDurationLocked(duration)
	t.running = false
	
	return duration
}

// RecordDuration manually records a duration without starting/stopping the timer
func (t *Timer) RecordDuration(duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.recordDurationLocked(duration)
}

// recordDurationLocked records a duration without locking (must be called with lock held)
func (t *Timer) recordDurationLocked(duration time.Duration) {
	t.lastDuration = duration.Nanoseconds()
	t.totalTime += t.lastDuration
	t.count++
}

// TimeFunction times the execution of a function and returns its result
func (t *Timer) TimeFunction(fn func() interface{}) interface{} {
	t.Start()
	result := fn()
	t.Stop()
	return result
}

// GetLastDuration returns the duration of the last timed operation
func (t *Timer) GetLastDuration() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	return time.Duration(t.lastDuration)
}

// GetTotalDuration returns the total duration of all timed operations
func (t *Timer) GetTotalDuration() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	return time.Duration(t.totalTime)
}

// GetCount returns the number of timed operations
func (t *Timer) GetCount() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	return t.count
}

// GetAverageDuration returns the average duration of all timed operations
func (t *Timer) GetAverageDuration() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.count == 0 {
		return 0
	}
	
	return time.Duration(t.totalTime / t.count)
}

// Registry is a central registry for all metrics
type Registry struct {
	counters      map[string]*Counter
	gauges        map[string]*Gauge
	ratioCounters map[string]*RatioCounter
	timers        map[string]*Timer
	mu            sync.RWMutex
}

// Global singleton registry
var (
	globalRegistry *Registry
	registryOnce   sync.Once
)

// GetRegistry returns the singleton global registry
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		globalRegistry = &Registry{
			counters:      make(map[string]*Counter),
			gauges:        make(map[string]*Gauge),
			ratioCounters: make(map[string]*RatioCounter),
			timers:        make(map[string]*Timer),
		}
	})
	return globalRegistry
}

// GetOrCreateCounter gets or creates a counter with the given name
func (r *Registry) GetOrCreateCounter(name string) *Counter {
	r.mu.RLock()
	counter, ok := r.counters[name]
	r.mu.RUnlock()
	
	if ok {
		return counter
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check again in case another goroutine created it
	counter, ok = r.counters[name]
	if ok {
		return counter
	}
	
	counter = NewCounter(name)
	r.counters[name] = counter
	return counter
}

// GetOrCreateGauge gets or creates a gauge with the given name
func (r *Registry) GetOrCreateGauge(name string) *Gauge {
	r.mu.RLock()
	gauge, ok := r.gauges[name]
	r.mu.RUnlock()
	
	if ok {
		return gauge
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check again in case another goroutine created it
	gauge, ok = r.gauges[name]
	if ok {
		return gauge
	}
	
	gauge = NewGauge(name)
	r.gauges[name] = gauge
	return gauge
}

// GetOrCreateRatioCounter gets or creates a ratio counter with the given name
func (r *Registry) GetOrCreateRatioCounter(name string) *RatioCounter {
	r.mu.RLock()
	ratio, ok := r.ratioCounters[name]
	r.mu.RUnlock()
	
	if ok {
		return ratio
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check again in case another goroutine created it
	ratio, ok = r.ratioCounters[name]
	if ok {
		return ratio
	}
	
	ratio = NewRatioCounter(name)
	r.ratioCounters[name] = ratio
	return ratio
}

// GetOrCreateTimer gets or creates a timer with the given name
func (r *Registry) GetOrCreateTimer(name string) *Timer {
	r.mu.RLock()
	timer, ok := r.timers[name]
	r.mu.RUnlock()
	
	if ok {
		return timer
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check again in case another goroutine created it
	timer, ok = r.timers[name]
	if ok {
		return timer
	}
	
	timer = NewTimer(name)
	r.timers[name] = timer
	return timer
}

// GetCounter gets a counter by name (or nil if not found)
func (r *Registry) GetCounter(name string) *Counter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.counters[name]
}

// GetGauge gets a gauge by name (or nil if not found)
func (r *Registry) GetGauge(name string) *Gauge {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.gauges[name]
}

// GetRatioCounter gets a ratio counter by name (or nil if not found)
func (r *Registry) GetRatioCounter(name string) *RatioCounter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ratioCounters[name]
}

// GetTimer gets a timer by name (or nil if not found)
func (r *Registry) GetTimer(name string) *Timer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.timers[name]
}

// GetAllCounters returns all registered counters
func (r *Registry) GetAllCounters() map[string]*Counter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Make a copy to avoid race conditions
	result := make(map[string]*Counter, len(r.counters))
	for k, v := range r.counters {
		result[k] = v
	}
	return result
}

// GetAllGauges returns all registered gauges
func (r *Registry) GetAllGauges() map[string]*Gauge {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Make a copy to avoid race conditions
	result := make(map[string]*Gauge, len(r.gauges))
	for k, v := range r.gauges {
		result[k] = v
	}
	return result
}

// GetAllRatioCounters returns all registered ratio counters
func (r *Registry) GetAllRatioCounters() map[string]*RatioCounter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Make a copy to avoid race conditions
	result := make(map[string]*RatioCounter, len(r.ratioCounters))
	for k, v := range r.ratioCounters {
		result[k] = v
	}
	return result
}

// GetAllTimers returns all registered timers
func (r *Registry) GetAllTimers() map[string]*Timer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Make a copy to avoid race conditions
	result := make(map[string]*Timer, len(r.timers))
	for k, v := range r.timers {
		result[k] = v
	}
	return result
}

// Clear removes all metrics from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.counters = make(map[string]*Counter)
	r.gauges = make(map[string]*Gauge)
	r.ratioCounters = make(map[string]*RatioCounter)
	r.timers = make(map[string]*Timer)
}