// ABOUTME: Metrics collection hook for monitoring agent performance
// ABOUTME: Tracks execution times, token usage, and tool invocation statistics

package core

import (
	"context"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// LLMMetricsHook implements Hook for collecting LLM-specific metrics.
// It tracks request counts, token usage, execution times, and tool invocation
// statistics. Thread-safe for concurrent use by multiple agents.
type LLMMetricsHook struct {
	mu            sync.RWMutex
	requests      int
	toolCalls     int
	errorCount    int
	totalTokens   int
	generateTimes []time.Duration
	toolTimes     map[string][]time.Duration

	// Context storage for timing
	startTimes sync.Map
}

// NewLLMMetricsHook creates a new LLM metrics hook.
// The hook starts with empty metrics that accumulate as the agent operates.
// Use the GetMetrics method to retrieve current statistics.
func NewLLMMetricsHook() *LLMMetricsHook {
	return &LLMMetricsHook{
		generateTimes: make([]time.Duration, 0),
		toolTimes:     make(map[string][]time.Duration),
	}
}

// BeforeGenerate is called before generating a response.
// It increments request count, estimates token usage, and records the start time
// for measuring generation duration.
func (h *LLMMetricsHook) BeforeGenerate(ctx context.Context, messages []domain.Message) {
	h.mu.Lock()
	h.requests++

	// Estimate token count (very rough estimation)
	tokenCount := 0
	for _, msg := range messages {
		for _, part := range msg.Content {
			if part.Type == domain.ContentTypeText {
				tokenCount += len(part.Text) / 4 // rough approximation of tokens
			}
		}
	}
	h.totalTokens += tokenCount
	h.mu.Unlock()

	// Store start time
	h.startTimes.Store(ctx, time.Now())
}

// AfterGenerate is called after generating a response
func (h *LLMMetricsHook) AfterGenerate(ctx context.Context, response domain.Response, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err != nil {
		h.errorCount++
		return
	}

	// Add response tokens
	h.totalTokens += len(response.Content) / 4 // rough approximation

	// Calculate time
	if startTimeVal, ok := h.startTimes.LoadAndDelete(ctx); ok {
		if startTime, ok := startTimeVal.(time.Time); ok {
			duration := time.Since(startTime)
			h.generateTimes = append(h.generateTimes, duration)
		}
	}
}

// BeforeToolCall is called before executing a tool
func (h *LLMMetricsHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	// Store start time for this tool call
	key := struct {
		ctx  context.Context
		tool string
	}{ctx, tool}
	h.startTimes.Store(key, time.Now())
}

// AfterToolCall is called after executing a tool
func (h *LLMMetricsHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.toolCalls++

	if err != nil {
		h.errorCount++
		return
	}

	// Calculate time
	key := struct {
		ctx  context.Context
		tool string
	}{ctx, tool}

	if startTimeVal, ok := h.startTimes.LoadAndDelete(key); ok {
		if startTime, ok := startTimeVal.(time.Time); ok {
			duration := time.Since(startTime)

			if _, exists := h.toolTimes[tool]; !exists {
				h.toolTimes[tool] = make([]time.Duration, 0)
			}
			h.toolTimes[tool] = append(h.toolTimes[tool], duration)
		}
	}
}

// Metrics represents the collected metrics
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
func (h *LLMMetricsHook) GetMetrics() Metrics {
	h.mu.RLock()
	defer h.mu.RUnlock()

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
func (h *LLMMetricsHook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requests = 0
	h.toolCalls = 0
	h.errorCount = 0
	h.totalTokens = 0
	h.generateTimes = make([]time.Duration, 0)
	h.toolTimes = make(map[string][]time.Duration)

	// Clear any stored start times
	h.startTimes.Range(func(key, value interface{}) bool {
		h.startTimes.Delete(key)
		return true
	})
}
