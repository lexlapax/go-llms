package workflow

import (
	"context"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// MetricsHook implements Hook for collecting metrics
type MetricsHook struct {
	mu            sync.Mutex
	requests      int
	toolCalls     int
	errorCount    int
	totalTokens   int
	generateTimes []time.Duration
	toolTimes     map[string][]time.Duration
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{
		generateTimes: make([]time.Duration, 0),
		toolTimes:     make(map[string][]time.Duration),
	}
}

// BeforeGenerate is called before generating a response
func (h *MetricsHook) BeforeGenerate(ctx context.Context, messages []domain.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.requests++

	// Estimate token count (very rough estimation)
	for _, msg := range messages {
		h.totalTokens += len(msg.Content) / 4 // rough approximation of tokens
	}

	// Store start time in context
	setMetricContextValue(ctx, "generateStartTime", time.Now())
}

// AfterGenerate is called after generating a response
func (h *MetricsHook) AfterGenerate(ctx context.Context, response domain.Response, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err != nil {
		h.errorCount++
		return
	}

	// Add response tokens
	h.totalTokens += len(response.Content) / 4 // rough approximation

	// Calculate time
	startTime, ok := getMetricContextValue(ctx, "generateStartTime").(time.Time)
	if ok {
		duration := time.Since(startTime)
		h.generateTimes = append(h.generateTimes, duration)
	}
}

// BeforeToolCall is called before executing a tool
func (h *MetricsHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	setMetricContextValue(ctx, "toolStartTime", time.Now())
}

// AfterToolCall is called after executing a tool
func (h *MetricsHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.toolCalls++

	if err != nil {
		h.errorCount++
		return
	}

	startTime, ok := getMetricContextValue(ctx, "toolStartTime").(time.Time)
	if ok {
		duration := time.Since(startTime)

		if _, exists := h.toolTimes[tool]; !exists {
			h.toolTimes[tool] = make([]time.Duration, 0)
		}
		h.toolTimes[tool] = append(h.toolTimes[tool], duration)
	}
}

// Metrics returns the collected metrics
type Metrics struct {
	Requests         int
	ToolCalls        int
	ErrorCount       int
	TotalTokens      int
	AverageGenTimeMs float64
	ToolStats        map[string]ToolStats
}

// ToolStats holds statistics for a specific tool
type ToolStats struct {
	Calls         int
	AverageTimeMs float64
	FastestCallMs float64
	SlowestCallMs float64
}

// GetMetrics returns the collected metrics
func (h *MetricsHook) GetMetrics() Metrics {
	h.mu.Lock()
	defer h.mu.Unlock()

	metrics := Metrics{
		Requests:    h.requests,
		ToolCalls:   h.toolCalls,
		ErrorCount:  h.errorCount,
		TotalTokens: h.totalTokens,
		ToolStats:   make(map[string]ToolStats),
	}

	// Calculate average generation time
	if len(h.generateTimes) > 0 {
		var total time.Duration
		for _, t := range h.generateTimes {
			total += t
		}
		metrics.AverageGenTimeMs = float64(total.Milliseconds()) / float64(len(h.generateTimes))
	}

	// Calculate tool statistics
	for tool, times := range h.toolTimes {
		if len(times) == 0 {
			continue
		}

		stats := ToolStats{
			Calls:         len(times),
			FastestCallMs: float64(times[0].Milliseconds()),
			SlowestCallMs: float64(times[0].Milliseconds()),
		}

		var total time.Duration
		for _, t := range times {
			total += t

			if float64(t.Milliseconds()) < stats.FastestCallMs {
				stats.FastestCallMs = float64(t.Milliseconds())
			}
			if float64(t.Milliseconds()) > stats.SlowestCallMs {
				stats.SlowestCallMs = float64(t.Milliseconds())
			}
		}

		stats.AverageTimeMs = float64(total.Milliseconds()) / float64(len(times))
		metrics.ToolStats[tool] = stats
	}

	return metrics
}

// Reset resets all metrics
func (h *MetricsHook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requests = 0
	h.toolCalls = 0
	h.errorCount = 0
	h.totalTokens = 0
	h.generateTimes = make([]time.Duration, 0)
	h.toolTimes = make(map[string][]time.Duration)
}

// Helper functions for storing metrics context values
type metricsContextKey string

// setMetricContextValue stores a value in the context
func setMetricContextValue(ctx context.Context, key string, value interface{}) {
	if ctx.Value(metricsContextKey("metricsValues")) == nil {
		// Context doesn't have our values map yet, so we can't store anything
		return
	}

	valuesMap, ok := ctx.Value(metricsContextKey("metricsValues")).(map[string]interface{})
	if !ok {
		return
	}

	valuesMap[key] = value
}

// getMetricContextValue retrieves a value from the context
func getMetricContextValue(ctx context.Context, key string) interface{} {
	if ctx.Value(metricsContextKey("metricsValues")) == nil {
		return nil
	}

	valuesMap, ok := ctx.Value(metricsContextKey("metricsValues")).(map[string]interface{})
	if !ok {
		return nil
	}

	return valuesMap[key]
}

// WithMetrics wraps a context to enable metrics collection
func WithMetrics(ctx context.Context) context.Context {
	return context.WithValue(ctx, metricsContextKey("metricsValues"), make(map[string]interface{}))
}
