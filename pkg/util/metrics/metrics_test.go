package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestNewCounter(t *testing.T) {
	counter := NewCounter("test_counter")
	if counter == nil {
		t.Fatal("Expected non-nil counter")
	}
	
	if counter.name != "test_counter" {
		t.Errorf("Expected name to be 'test_counter', got '%s'", counter.name)
	}
	
	if counter.value != 0 {
		t.Errorf("Expected initial value to be 0, got %d", counter.value)
	}
}

func TestCounterIncrement(t *testing.T) {
	counter := NewCounter("test_counter")
	
	// Test single increment
	counter.Increment()
	if counter.value != 1 {
		t.Errorf("Expected value to be 1 after increment, got %d", counter.value)
	}
	
	// Test increment with value
	counter.IncrementBy(5)
	if counter.value != 6 {
		t.Errorf("Expected value to be 6 after incrementing by 5, got %d", counter.value)
	}
}

func TestCounterGetValue(t *testing.T) {
	counter := NewCounter("test_counter")
	counter.IncrementBy(10)
	
	if counter.GetValue() != 10 {
		t.Errorf("Expected GetValue() to return 10, got %d", counter.GetValue())
	}
}

func TestCounterConcurrency(t *testing.T) {
	counter := NewCounter("test_counter")
	
	var wg sync.WaitGroup
	workers := 10
	incrementsPerWorker := 1000
	
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerWorker; j++ {
				counter.Increment()
			}
		}()
	}
	
	wg.Wait()
	
	expectedValue := workers * incrementsPerWorker
	if counter.GetValue() != int64(expectedValue) {
		t.Errorf("Expected value to be %d after concurrent increments, got %d", expectedValue, counter.GetValue())
	}
}

func TestNewGauge(t *testing.T) {
	gauge := NewGauge("test_gauge")
	if gauge == nil {
		t.Fatal("Expected non-nil gauge")
	}
	
	if gauge.name != "test_gauge" {
		t.Errorf("Expected name to be 'test_gauge', got '%s'", gauge.name)
	}
	
	if gauge.value != 0 {
		t.Errorf("Expected initial value to be 0, got %f", gauge.value)
	}
}

func TestGaugeSetAndGet(t *testing.T) {
	gauge := NewGauge("test_gauge")
	
	// Test set value
	gauge.Set(42.5)
	if gauge.GetValue() != 42.5 {
		t.Errorf("Expected value to be 42.5 after Set, got %f", gauge.GetValue())
	}
	
	// Test increment
	gauge.Increment()
	if gauge.GetValue() != 43.5 {
		t.Errorf("Expected value to be 43.5 after increment, got %f", gauge.GetValue())
	}
	
	// Test decrement
	gauge.Decrement()
	if gauge.GetValue() != 42.5 {
		t.Errorf("Expected value to be 42.5 after decrement, got %f", gauge.GetValue())
	}
	
	// Test add
	gauge.Add(7.5)
	if gauge.GetValue() != 50.0 {
		t.Errorf("Expected value to be 50.0 after adding 7.5, got %f", gauge.GetValue())
	}
	
	// Test subtract
	gauge.Subtract(5.0)
	if gauge.GetValue() != 45.0 {
		t.Errorf("Expected value to be 45.0 after subtracting 5.0, got %f", gauge.GetValue())
	}
}

func TestGaugeConcurrency(t *testing.T) {
	gauge := NewGauge("test_gauge")
	
	var wg sync.WaitGroup
	workers := 10
	incrementsPerWorker := 1000
	
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerWorker; j++ {
				gauge.Add(0.1)
			}
		}()
	}
	
	wg.Wait()
	
	expectedValue := float64(workers * incrementsPerWorker) * 0.1
	actualValue := gauge.GetValue()
	// Check with a small epsilon for floating point comparison
	if actualValue < expectedValue-0.001 || actualValue > expectedValue+0.001 {
		t.Errorf("Expected value to be %f after concurrent adds, got %f", expectedValue, actualValue)
	}
}

func TestNewRatioCounter(t *testing.T) {
	ratio := NewRatioCounter("test_ratio")
	if ratio == nil {
		t.Fatal("Expected non-nil ratio counter")
	}
	
	if ratio.name != "test_ratio" {
		t.Errorf("Expected name to be 'test_ratio', got '%s'", ratio.name)
	}
	
	if ratio.numerator != 0 || ratio.denominator != 0 {
		t.Errorf("Expected initial values to be 0/0, got %d/%d", ratio.numerator, ratio.denominator)
	}
}

func TestRatioCounterIncrementAndGet(t *testing.T) {
	ratio := NewRatioCounter("test_ratio")
	
	// When denominator is 0, ratio should be 0
	ratio.IncrementNumerator()
	if ratio.GetRatio() != 0.0 {
		t.Errorf("Expected ratio to be 0.0 when denominator is 0, got %f", ratio.GetRatio())
	}
	
	// Test with 1/1
	ratio.IncrementDenominator()
	if ratio.GetRatio() != 1.0 {
		t.Errorf("Expected ratio to be 1.0, got %f", ratio.GetRatio())
	}
	
	// Test with 1/2
	ratio.IncrementDenominator()
	if ratio.GetRatio() != 0.5 {
		t.Errorf("Expected ratio to be 0.5, got %f", ratio.GetRatio())
	}
	
	// Test with 3/5
	ratio.IncrementNumerator()
	ratio.IncrementNumerator()
	ratio.IncrementDenominator()
	ratio.IncrementDenominator()
	ratio.IncrementDenominator()
	if ratio.GetRatio() != 0.6 {
		t.Errorf("Expected ratio to be 0.6, got %f", ratio.GetRatio())
	}
}

func TestRatioCounterIncByValueAndGet(t *testing.T) {
	ratio := NewRatioCounter("test_ratio")
	
	// Test with 5/10
	ratio.IncrementNumeratorBy(5)
	ratio.IncrementDenominatorBy(10)
	if ratio.GetRatio() != 0.5 {
		t.Errorf("Expected ratio to be 0.5, got %f", ratio.GetRatio())
	}
	
	// Get raw values
	if n, d := ratio.GetValues(); n != 5 || d != 10 {
		t.Errorf("Expected values to be 5/10, got %d/%d", n, d)
	}
}

func TestRatioCounterConcurrency(t *testing.T) {
	ratio := NewRatioCounter("test_ratio")
	
	var wg sync.WaitGroup
	workers := 10
	incrementsPerWorker := 1000
	
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerWorker; j++ {
				ratio.IncrementNumerator()
				ratio.IncrementDenominator()
				ratio.IncrementDenominator() // Make denominator twice the numerator
			}
		}()
	}
	
	wg.Wait()
	
	expectedRatio := 0.5 // Each worker adds N to numerator and 2N to denominator
	actualRatio := ratio.GetRatio()
	if actualRatio < expectedRatio-0.001 || actualRatio > expectedRatio+0.001 {
		t.Errorf("Expected ratio to be %f after concurrent increments, got %f", expectedRatio, actualRatio)
	}
}

func TestNewTimer(t *testing.T) {
	timer := NewTimer("test_timer")
	if timer == nil {
		t.Fatal("Expected non-nil timer")
	}
	
	if timer.name != "test_timer" {
		t.Errorf("Expected name to be 'test_timer', got '%s'", timer.name)
	}
}

func TestTimerStartStopAndRecord(t *testing.T) {
	timer := NewTimer("test_timer")
	
	// Test simple timing of a sleep
	timer.Start()
	time.Sleep(100 * time.Millisecond)
	timer.Stop()
	
	// Duration should be approximately 100ms (but we'll allow for some variation)
	if timer.GetLastDuration() < 90*time.Millisecond || timer.GetLastDuration() > 150*time.Millisecond {
		t.Errorf("Expected last duration around 100ms, got %v", timer.GetLastDuration())
	}
	
	// Test count and total time
	if timer.GetCount() != 1 {
		t.Errorf("Expected count to be 1, got %d", timer.GetCount())
	}
	
	if timer.GetTotalDuration() < 90*time.Millisecond || timer.GetTotalDuration() > 150*time.Millisecond {
		t.Errorf("Expected total duration around 100ms, got %v", timer.GetTotalDuration())
	}
	
	// Test another timing
	timer.Start()
	time.Sleep(50 * time.Millisecond)
	timer.Stop()
	
	if timer.GetCount() != 2 {
		t.Errorf("Expected count to be 2, got %d", timer.GetCount())
	}
	
	if timer.GetTotalDuration() < 140*time.Millisecond || timer.GetTotalDuration() > 200*time.Millisecond {
		t.Errorf("Expected total duration around 150ms, got %v", timer.GetTotalDuration())
	}
	
	// Test average
	avgExpected := timer.GetTotalDuration().Nanoseconds() / int64(timer.GetCount())
	avgActual := timer.GetAverageDuration().Nanoseconds()
	if avgActual != avgExpected {
		t.Errorf("Expected average duration to be %d ns, got %d ns", avgExpected, avgActual)
	}
}

func TestTimerTimeFunction(t *testing.T) {
	timer := NewTimer("test_timer")
	
	// Time a function execution
	result := timer.TimeFunction(func() interface{} {
		time.Sleep(100 * time.Millisecond)
		return 42
	})
	
	// Check result
	if result != 42 {
		t.Errorf("Expected result to be 42, got %v", result)
	}
	
	// Check timing
	if timer.GetCount() != 1 {
		t.Errorf("Expected count to be 1, got %d", timer.GetCount())
	}
	
	if timer.GetLastDuration() < 90*time.Millisecond || timer.GetLastDuration() > 150*time.Millisecond {
		t.Errorf("Expected last duration around 100ms, got %v", timer.GetLastDuration())
	}
}

func TestRegistry(t *testing.T) {
	// Reset the global registry before the test
	globalRegistry = &Registry{
		counters:      make(map[string]*Counter),
		gauges:        make(map[string]*Gauge),
		ratioCounters: make(map[string]*RatioCounter),
		timers:        make(map[string]*Timer),
	}

	reg := GetRegistry()
	if reg == nil {
		t.Fatal("Expected non-nil registry")
	}
	
	// Get again, should be the same instance
	reg2 := GetRegistry()
	if reg != reg2 {
		t.Error("Expected GetRegistry() to return the same instance")
	}
	
	// Test registering metrics
	counter := reg.GetOrCreateCounter("test_counter")
	counter.Increment()
	
	gauge := reg.GetOrCreateGauge("test_gauge")
	gauge.Set(42.0)
	
	ratio := reg.GetOrCreateRatioCounter("test_ratio")
	ratio.IncrementNumerator()
	ratio.IncrementDenominator()
	
	timer := reg.GetOrCreateTimer("test_timer")
	timer.RecordDuration(100 * time.Millisecond)
	
	// Test looking up metrics
	if reg.GetCounter("test_counter") != counter {
		t.Error("Expected to get the same counter that was registered")
	}
	
	if reg.GetGauge("test_gauge") != gauge {
		t.Error("Expected to get the same gauge that was registered")
	}
	
	if reg.GetRatioCounter("test_ratio") != ratio {
		t.Error("Expected to get the same ratio counter that was registered")
	}
	
	if reg.GetTimer("test_timer") != timer {
		t.Error("Expected to get the same timer that was registered")
	}
	
	// Test getting all metrics
	allCounters := reg.GetAllCounters()
	if len(allCounters) != 1 || allCounters["test_counter"] != counter {
		t.Error("GetAllCounters didn't return the expected counters")
	}
	
	allGauges := reg.GetAllGauges()
	if len(allGauges) != 1 || allGauges["test_gauge"] != gauge {
		t.Error("GetAllGauges didn't return the expected gauges")
	}
	
	allRatios := reg.GetAllRatioCounters()
	if len(allRatios) != 1 || allRatios["test_ratio"] != ratio {
		t.Error("GetAllRatioCounters didn't return the expected ratio counters")
	}
	
	allTimers := reg.GetAllTimers()
	if len(allTimers) != 1 || allTimers["test_timer"] != timer {
		t.Error("GetAllTimers didn't return the expected timers")
	}
}

func TestMultipleMetricTypes(t *testing.T) {
	// Reset the global registry before the test
	globalRegistry = &Registry{
		counters:      make(map[string]*Counter),
		gauges:        make(map[string]*Gauge),
		ratioCounters: make(map[string]*RatioCounter),
		timers:        make(map[string]*Timer),
	}
	reg := GetRegistry()
	
	// Test that we can have metrics with the same name but different types
	counter := reg.GetOrCreateCounter("test_metric")
	gauge := reg.GetOrCreateGauge("test_metric")
	ratio := reg.GetOrCreateRatioCounter("test_metric")
	timer := reg.GetOrCreateTimer("test_metric")
	
	// Update all metrics
	counter.Increment()
	gauge.Set(42.0)
	ratio.IncrementNumerator()
	ratio.IncrementDenominator()
	timer.RecordDuration(100 * time.Millisecond)
	
	// They should all be different instances
	if counter.GetValue() != 1 {
		t.Error("Counter value incorrect")
	}
	
	if gauge.GetValue() != 42.0 {
		t.Error("Gauge value incorrect")
	}
	
	if ratio.GetRatio() != 1.0 {
		t.Error("Ratio value incorrect")
	}
	
	if timer.GetCount() != 1 {
		t.Error("Timer count incorrect")
	}
}

func TestClear(t *testing.T) {
	// Reset the global registry before the test
	globalRegistry = &Registry{
		counters:      make(map[string]*Counter),
		gauges:        make(map[string]*Gauge),
		ratioCounters: make(map[string]*RatioCounter),
		timers:        make(map[string]*Timer),
	}
	reg := GetRegistry()
	
	// Create some metrics
	counter := reg.GetOrCreateCounter("test_counter")
	counter.Increment()
	
	gauge := reg.GetOrCreateGauge("test_gauge")
	gauge.Set(42.0)
	
	// Clear all metrics
	reg.Clear()
	
	// Metrics should be gone
	if reg.GetCounter("test_counter") != nil {
		t.Error("Expected counter to be nil after Clear()")
	}
	
	if reg.GetGauge("test_gauge") != nil {
		t.Error("Expected gauge to be nil after Clear()")
	}
	
	// Creating new metrics should work
	newCounter := reg.GetOrCreateCounter("test_counter")
	if newCounter.GetValue() != 0 {
		t.Error("Expected new counter to start at 0")
	}
}