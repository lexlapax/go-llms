package llmutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ProviderPool is a pool of LLM providers for load balancing and fallback
type ProviderPool struct {
	providers   []domain.Provider
	strategy    PoolStrategy
	metrics     map[int]*ProviderMetrics
	mu          sync.RWMutex
	activeIndex int
}

// PoolStrategy defines how the provider pool selects a provider
type PoolStrategy int

const (
	// StrategyRoundRobin cycles through providers
	StrategyRoundRobin PoolStrategy = iota
	
	// StrategyFailover uses the first provider until it fails, then moves to the next
	StrategyFailover
	
	// StrategyFastest uses the provider with the lowest latency
	StrategyFastest
)

// ProviderMetrics tracks performance metrics for a provider
type ProviderMetrics struct {
	Requests         int
	Failures         int
	AvgLatencyMs     int64
	TotalLatencyMs   int64
	LastUsed         time.Time
	ConsecutiveErrors int
}

// NewProviderPool creates a new provider pool
func NewProviderPool(providers []domain.Provider, strategy PoolStrategy) *ProviderPool {
	metrics := make(map[int]*ProviderMetrics)
	for i := range providers {
		metrics[i] = &ProviderMetrics{
			LastUsed: time.Now(),
		}
	}
	
	return &ProviderPool{
		providers:   providers,
		strategy:    strategy,
		metrics:     metrics,
		activeIndex: 0,
	}
}

// Generate implements the Provider interface for the pool
func (p *ProviderPool) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	idx, provider, err := p.getProvider()
	if err != nil {
		return "", err
	}
	
	startTime := time.Now()
	result, err := provider.Generate(ctx, prompt, options...)
	duration := time.Since(startTime)
	
	p.updateMetrics(idx, err, duration)
	
	if err != nil {
		// If the selected provider fails, try to find another one
		if p.strategy == StrategyFailover {
			fallbackIdx, fallbackProvider, fallbackErr := p.getFallbackProvider(idx)
			if fallbackErr != nil {
				return "", err // Return original error if no fallback
			}
			
			fallbackStartTime := time.Now()
			fallbackResult, fallbackErr := fallbackProvider.Generate(ctx, prompt, options...)
			fallbackDuration := time.Since(fallbackStartTime)
			
			p.updateMetrics(fallbackIdx, fallbackErr, fallbackDuration)
			
			if fallbackErr != nil {
				return "", err // Return original error if fallback also fails
			}
			
			return fallbackResult, nil
		}
		
		return "", err
	}
	
	return result, nil
}

// GenerateMessage implements the Provider interface for the pool
func (p *ProviderPool) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	idx, provider, err := p.getProvider()
	if err != nil {
		return domain.Response{}, err
	}
	
	startTime := time.Now()
	result, err := provider.GenerateMessage(ctx, messages, options...)
	duration := time.Since(startTime)
	
	p.updateMetrics(idx, err, duration)
	
	if err != nil {
		// If the selected provider fails, try to find another one
		if p.strategy == StrategyFailover {
			fallbackIdx, fallbackProvider, fallbackErr := p.getFallbackProvider(idx)
			if fallbackErr != nil {
				return domain.Response{}, err // Return original error if no fallback
			}
			
			fallbackStartTime := time.Now()
			fallbackResult, fallbackErr := fallbackProvider.GenerateMessage(ctx, messages, options...)
			fallbackDuration := time.Since(fallbackStartTime)
			
			p.updateMetrics(fallbackIdx, fallbackErr, fallbackDuration)
			
			if fallbackErr != nil {
				return domain.Response{}, err // Return original error if fallback also fails
			}
			
			return fallbackResult, nil
		}
		
		return domain.Response{}, err
	}
	
	return result, nil
}

// GenerateWithSchema implements the Provider interface for the pool
func (p *ProviderPool) GenerateWithSchema(ctx context.Context, prompt string, schema interface{}, options ...domain.Option) (interface{}, error) {
	idx, provider, err := p.getProvider()
	if err != nil {
		return nil, err
	}
	
	startTime := time.Now()
	
	// Check if the schema is already of the correct type
	schemaObj, ok := schema.(*schemaDomain.Schema)
	if !ok {
		// For non-Schema types, use a default schema
		schemaObj = &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{},
		}
	}
	
	result, err := provider.GenerateWithSchema(ctx, prompt, schemaObj, options...)
	duration := time.Since(startTime)
	
	p.updateMetrics(idx, err, duration)
	
	if err != nil {
		// If the selected provider fails, try to find another one
		if p.strategy == StrategyFailover {
			fallbackIdx, fallbackProvider, fallbackErr := p.getFallbackProvider(idx)
			if fallbackErr != nil {
				return nil, err // Return original error if no fallback
			}
			
			fallbackStartTime := time.Now()
			fallbackResult, fallbackErr := fallbackProvider.GenerateWithSchema(ctx, prompt, schemaObj, options...)
			fallbackDuration := time.Since(fallbackStartTime)
			
			p.updateMetrics(fallbackIdx, fallbackErr, fallbackDuration)
			
			if fallbackErr != nil {
				return nil, err // Return original error if fallback also fails
			}
			
			return fallbackResult, nil
		}
		
		return nil, err
	}
	
	return result, nil
}

// Stream implements the Provider interface for the pool
func (p *ProviderPool) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	idx, provider, err := p.getProvider()
	if err != nil {
		return nil, err
	}
	
	startTime := time.Now()
	result, err := provider.Stream(ctx, prompt, options...)
	duration := time.Since(startTime)
	
	p.updateMetrics(idx, err, duration)
	
	if err != nil {
		// If the selected provider fails, try to find another one
		if p.strategy == StrategyFailover {
			fallbackIdx, fallbackProvider, fallbackErr := p.getFallbackProvider(idx)
			if fallbackErr != nil {
				return nil, err // Return original error if no fallback
			}
			
			fallbackStartTime := time.Now()
			fallbackResult, fallbackErr := fallbackProvider.Stream(ctx, prompt, options...)
			fallbackDuration := time.Since(fallbackStartTime)
			
			p.updateMetrics(fallbackIdx, fallbackErr, fallbackDuration)
			
			if fallbackErr != nil {
				return nil, err // Return original error if fallback also fails
			}
			
			return fallbackResult, nil
		}
		
		return nil, err
	}
	
	return result, nil
}

// StreamMessage implements the Provider interface for the pool
func (p *ProviderPool) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	idx, provider, err := p.getProvider()
	if err != nil {
		return nil, err
	}
	
	startTime := time.Now()
	result, err := provider.StreamMessage(ctx, messages, options...)
	duration := time.Since(startTime)
	
	p.updateMetrics(idx, err, duration)
	
	if err != nil {
		// If the selected provider fails, try to find another one
		if p.strategy == StrategyFailover {
			fallbackIdx, fallbackProvider, fallbackErr := p.getFallbackProvider(idx)
			if fallbackErr != nil {
				return nil, err // Return original error if no fallback
			}
			
			fallbackStartTime := time.Now()
			fallbackResult, fallbackErr := fallbackProvider.StreamMessage(ctx, messages, options...)
			fallbackDuration := time.Since(fallbackStartTime)
			
			p.updateMetrics(fallbackIdx, fallbackErr, fallbackDuration)
			
			if fallbackErr != nil {
				return nil, err // Return original error if fallback also fails
			}
			
			return fallbackResult, nil
		}
		
		return nil, err
	}
	
	return result, nil
}

// getProvider selects a provider based on the strategy
func (p *ProviderPool) getProvider() (int, domain.Provider, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if len(p.providers) == 0 {
		return -1, nil, fmt.Errorf("no providers available")
	}
	
	switch p.strategy {
	case StrategyRoundRobin:
		idx := p.activeIndex
		p.activeIndex = (p.activeIndex + 1) % len(p.providers)
		return idx, p.providers[idx], nil
		
	case StrategyFailover:
		return p.activeIndex, p.providers[p.activeIndex], nil
		
	case StrategyFastest:
		var fastestIdx int
		var fastestLatency int64 = -1
		
		for i, metrics := range p.metrics {
			// Skip providers with consecutive errors
			if metrics.ConsecutiveErrors > 3 {
				continue
			}
			
			if fastestLatency == -1 || (metrics.AvgLatencyMs > 0 && metrics.AvgLatencyMs < fastestLatency) {
				fastestLatency = metrics.AvgLatencyMs
				fastestIdx = i
			}
		}
		
		// If all providers have consecutive errors, use the first one
		if fastestLatency == -1 {
			return 0, p.providers[0], nil
		}
		
		return fastestIdx, p.providers[fastestIdx], nil
		
	default:
		return 0, p.providers[0], nil
	}
}

// getFallbackProvider finds a fallback provider when the current one fails
func (p *ProviderPool) getFallbackProvider(currentIdx int) (int, domain.Provider, error) {
	if len(p.providers) <= 1 {
		return -1, nil, fmt.Errorf("no fallback providers available")
	}
	
	// In failover strategy, move to the next provider
	if p.strategy == StrategyFailover {
		p.mu.Lock()
		p.activeIndex = (currentIdx + 1) % len(p.providers)
		idx := p.activeIndex
		p.mu.Unlock()
		
		return idx, p.providers[idx], nil
	}
	
	// Otherwise just pick a different provider
	nextIdx := (currentIdx + 1) % len(p.providers)
	return nextIdx, p.providers[nextIdx], nil
}

// updateMetrics updates the metrics for a provider
func (p *ProviderPool) updateMetrics(idx int, err error, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	metrics := p.metrics[idx]
	metrics.Requests++
	metrics.LastUsed = time.Now()
	
	if err != nil {
		metrics.Failures++
		metrics.ConsecutiveErrors++
	} else {
		metrics.ConsecutiveErrors = 0
		metrics.TotalLatencyMs += duration.Milliseconds()
		metrics.AvgLatencyMs = metrics.TotalLatencyMs / int64(metrics.Requests-metrics.Failures)
	}
}

// GetMetrics returns metrics for all providers
func (p *ProviderPool) GetMetrics() map[int]*ProviderMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	// Make a copy to avoid race conditions
	metricsCopy := make(map[int]*ProviderMetrics)
	for i, m := range p.metrics {
		metricsCopy[i] = &ProviderMetrics{
			Requests:         m.Requests,
			Failures:         m.Failures,
			AvgLatencyMs:     m.AvgLatencyMs,
			TotalLatencyMs:   m.TotalLatencyMs,
			LastUsed:         m.LastUsed,
			ConsecutiveErrors: m.ConsecutiveErrors,
		}
	}
	
	return metricsCopy
}