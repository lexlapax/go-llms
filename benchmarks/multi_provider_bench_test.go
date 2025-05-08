package benchmarks

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Counter for benchmarking provider calls
var (
	generateCounter        int64
	generateMessageCounter int64
	generateSchemaCounter  int64
	streamCounter          int64
	streamMessageCounter   int64
)

// resetCounters resets all counters for clean test runs
func resetCounters() {
	atomic.StoreInt64(&generateCounter, 0)
	atomic.StoreInt64(&generateMessageCounter, 0)
	atomic.StoreInt64(&generateSchemaCounter, 0)
	atomic.StoreInt64(&streamCounter, 0)
	atomic.StoreInt64(&streamMessageCounter, 0)
}

// BenchmarkProviderTypes compares different provider configurations
func BenchmarkProviderTypes(b *testing.B) {
	// Fast mock provider that succeeds immediately
	fastProvider := newCountingMockProvider(0, false)

	// Slow mock provider that takes 50ms to respond
	slowProvider := newCountingMockProvider(50*time.Millisecond, false)

	// Failing mock provider that always returns an error
	failingProvider := newCountingMockProvider(0, true)

	// Variable delay provider (100-300ms)
	variableProvider := newVariableDelayMockProvider(100*time.Millisecond, 300*time.Millisecond)

	// Test input data
	prompt := "Test prompt for benchmarking"
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: "Test message for benchmarking"},
	}
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"result": {Type: "string"},
		},
	}

	ctx := context.Background()

	// Single provider (baseline)
	b.Run("SingleProvider", func(b *testing.B) {
		resetCounters()
		b.Run("Generate", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = fastProvider.Generate(ctx, prompt)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateCounter)), "provider_calls")
		})

		b.Run("GenerateMessage", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = fastProvider.GenerateMessage(ctx, messages)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateMessageCounter)), "provider_calls")
		})

		b.Run("GenerateWithSchema", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = fastProvider.GenerateWithSchema(ctx, prompt, schema)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateSchemaCounter)), "provider_calls")
		})
	})

	// Multi-provider with fastest strategy (optimal case)
	b.Run("MultiProvider_Fastest_Optimal", func(b *testing.B) {
		providers := []provider.ProviderWeight{
			{Provider: fastProvider, Weight: 1.0, Name: "fast"},
			{Provider: slowProvider, Weight: 1.0, Name: "slow"},
		}
		multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

		b.Run("Generate", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = multiProvider.Generate(ctx, prompt)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateCounter)), "provider_calls")
		})

		b.Run("GenerateMessage", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = multiProvider.GenerateMessage(ctx, messages)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateMessageCounter)), "provider_calls")
		})

		b.Run("GenerateWithSchema", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = multiProvider.GenerateWithSchema(ctx, prompt, schema)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateSchemaCounter)), "provider_calls")
		})
	})

	// Multi-provider with primary strategy (primary succeeds)
	b.Run("MultiProvider_Primary_Success", func(b *testing.B) {
		providers := []provider.ProviderWeight{
			{Provider: fastProvider, Weight: 1.0, Name: "fast"},
			{Provider: slowProvider, Weight: 1.0, Name: "slow"},
		}
		multiProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).WithPrimaryProvider(0)

		b.Run("Generate", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = multiProvider.Generate(ctx, prompt)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateCounter)), "provider_calls")
		})

		b.Run("GenerateMessage", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = multiProvider.GenerateMessage(ctx, messages)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateMessageCounter)), "provider_calls")
		})
	})

	// Multi-provider with primary strategy (primary fails)
	b.Run("MultiProvider_Primary_Fallback", func(b *testing.B) {
		providers := []provider.ProviderWeight{
			{Provider: failingProvider, Weight: 1.0, Name: "failing"},
			{Provider: fastProvider, Weight: 1.0, Name: "fast"},
		}
		multiProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).WithPrimaryProvider(0)

		b.Run("Generate", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = multiProvider.Generate(ctx, prompt)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateCounter)), "provider_calls")
		})

		b.Run("GenerateMessage", func(b *testing.B) {
			resetCounters()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = multiProvider.GenerateMessage(ctx, messages)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateMessageCounter)), "provider_calls")
		})
	})

	// Multi-provider with variable delay providers (realistic simulation)
	b.Run("MultiProvider_VariableDelay", func(b *testing.B) {
		providers := []provider.ProviderWeight{
			{Provider: variableProvider, Weight: 1.0, Name: "variable1"},
			{Provider: newVariableDelayMockProvider(200*time.Millisecond, 400*time.Millisecond), Weight: 1.0, Name: "variable2"},
			{Provider: newVariableDelayMockProvider(50*time.Millisecond, 500*time.Millisecond), Weight: 1.0, Name: "variable3"},
		}
		multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

		// Only run a small number of iterations since this is slow
		b.Run("Generate", func(b *testing.B) {
			resetCounters()
			count := b.N
			if count > 20 {
				count = 20 // Limit to 20 iterations for slow tests
			}
			b.ResetTimer()
			for i := 0; i < count; i++ {
				_, _ = multiProvider.Generate(ctx, prompt)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateCounter)), "provider_calls")
		})

		b.Run("GenerateMessage", func(b *testing.B) {
			resetCounters()
			count := b.N
			if count > 20 {
				count = 20 // Limit to 20 iterations for slow tests
			}
			b.ResetTimer()
			for i := 0; i < count; i++ {
				_, _ = multiProvider.GenerateMessage(ctx, messages)
			}
			b.StopTimer()
			b.ReportMetric(float64(atomic.LoadInt64(&generateMessageCounter)), "provider_calls")
		})
	})
}

// BenchmarkProviderTimeout tests timeout handling with mock providers
func BenchmarkProviderTimeout(b *testing.B) {
	// Create providers with different delays
	verySlowProvider := newCountingMockProvider(2*time.Second, false)
	slowerProvider := newCountingMockProvider(500*time.Millisecond, false)
	fasterProvider := newCountingMockProvider(100*time.Millisecond, false)

	// Test input data
	prompt := "Test prompt for timeout benchmarking"
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: "Test message for timeout benchmarking"},
	}

	// Timeout duration will be used for each test context

	// Create multi-provider with fastest strategy
	providers := []provider.ProviderWeight{
		{Provider: verySlowProvider, Weight: 1.0, Name: "very_slow"},
		{Provider: slowerProvider, Weight: 1.0, Name: "slower"},
		{Provider: fasterProvider, Weight: 1.0, Name: "faster"},
	}
	multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

	// Only run a small number of iterations since this is slow
	b.Run("Generate_WithTimeout", func(b *testing.B) {
		resetCounters()
		count := b.N
		if count > 10 {
			count = 10 // Limit to 10 iterations for slow tests
		}
		b.ResetTimer()
		for i := 0; i < count; i++ {
			// Use a fresh context for each iteration
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			_, _ = multiProvider.Generate(timeoutCtx, prompt)
			timeoutCancel()
		}
		b.StopTimer()
		b.ReportMetric(float64(atomic.LoadInt64(&generateCounter)), "provider_calls")
	})

	b.Run("GenerateMessage_WithTimeout", func(b *testing.B) {
		resetCounters()
		count := b.N
		if count > 10 {
			count = 10 // Limit to 10 iterations for slow tests
		}
		b.ResetTimer()
		for i := 0; i < count; i++ {
			// Use a fresh context for each iteration
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			_, _ = multiProvider.GenerateMessage(timeoutCtx, messages)
			timeoutCancel()
		}
		b.StopTimer()
		b.ReportMetric(float64(atomic.LoadInt64(&generateMessageCounter)), "provider_calls")
	})
}

// Mock providers for benchmarking

// countingMockProvider counts calls and simulates delay and failures
type countingMockProvider struct {
	delay      time.Duration
	alwaysFail bool
}

func newCountingMockProvider(delay time.Duration, alwaysFail bool) *countingMockProvider {
	return &countingMockProvider{
		delay:      delay,
		alwaysFail: alwaysFail,
	}
}

func (p *countingMockProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	atomic.AddInt64(&generateCounter, 1)

	if p.delay > 0 {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(p.delay):
			// Continue after delay
		}
	}

	if p.alwaysFail {
		return "", errors.New("simulated provider failure")
	}

	return "Mock response for: " + prompt, nil
}

func (p *countingMockProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	atomic.AddInt64(&generateMessageCounter, 1)

	if p.delay > 0 {
		select {
		case <-ctx.Done():
			return domain.Response{}, ctx.Err()
		case <-time.After(p.delay):
			// Continue after delay
		}
	}

	if p.alwaysFail {
		return domain.Response{}, errors.New("simulated provider failure")
	}

	// Create response with message content
	var content string
	if len(messages) > 0 {
		content = "Mock response for message: " + messages[len(messages)-1].Content
	} else {
		content = "Mock response for empty messages"
	}

	return domain.GetResponsePool().NewResponse(content), nil
}

func (p *countingMockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	atomic.AddInt64(&generateSchemaCounter, 1)

	if p.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(p.delay):
			// Continue after delay
		}
	}

	if p.alwaysFail {
		return nil, errors.New("simulated provider failure")
	}

	// Create a simple mock response
	mockResponse := map[string]interface{}{
		"result": "Mock structured response for: " + prompt,
	}

	return mockResponse, nil
}

func (p *countingMockProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	atomic.AddInt64(&streamCounter, 1)

	if p.alwaysFail {
		return nil, errors.New("simulated provider failure")
	}

	tokenCh := make(chan domain.Token, 5)

	go func() {
		defer close(tokenCh)

		// Simulate token streaming with delay
		tokens := []string{"Mock ", "streaming ", "response ", "for: ", prompt}

		for i, token := range tokens {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return
			case <-time.After(p.delay / 5): // Divide the delay among tokens
				// Send the token
				isLast := i == len(tokens)-1

				select {
				case <-ctx.Done():
					return
				case tokenCh <- domain.GetTokenPool().NewToken(token, isLast):
					// Token sent successfully
				}
			}
		}
	}()

	return tokenCh, nil
}

func (p *countingMockProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	atomic.AddInt64(&streamMessageCounter, 1)

	if p.alwaysFail {
		return nil, errors.New("simulated provider failure")
	}

	tokenCh := make(chan domain.Token, 5)

	go func() {
		defer close(tokenCh)

		// Get the last message content for the response
		var lastContent string
		if len(messages) > 0 {
			lastContent = messages[len(messages)-1].Content
		} else {
			lastContent = "empty messages"
		}

		// Simulate token streaming with delay
		tokens := []string{"Mock ", "streaming ", "response ", "for: ", lastContent}

		for i, token := range tokens {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return
			case <-time.After(p.delay / 5): // Divide the delay among tokens
				// Send the token
				isLast := i == len(tokens)-1

				select {
				case <-ctx.Done():
					return
				case tokenCh <- domain.GetTokenPool().NewToken(token, isLast):
					// Token sent successfully
				}
			}
		}
	}()

	return tokenCh, nil
}

// variableDelayMockProvider introduces random but bounded delay between min and max
type variableDelayMockProvider struct {
	minDelay time.Duration
	maxDelay time.Duration
}

func newVariableDelayMockProvider(minDelay, maxDelay time.Duration) *variableDelayMockProvider {
	return &variableDelayMockProvider{
		minDelay: minDelay,
		maxDelay: maxDelay,
	}
}

// getRandomDelay returns a duration between min and max delay
func (p *variableDelayMockProvider) getRandomDelay() time.Duration {
	// Simple non-random pattern for testing - cycle between min, mid, and max
	// This gives more predictable results than random while still testing the concept
	now := time.Now().UnixNano()
	switch now % 3 {
	case 0:
		return p.minDelay
	case 1:
		return (p.minDelay + p.maxDelay) / 2
	default:
		return p.maxDelay
	}
}

// Implement all provider methods with variable delay

func (p *variableDelayMockProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	atomic.AddInt64(&generateCounter, 1)
	delay := p.getRandomDelay()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(delay):
		return fmt.Sprintf("Response after %v delay: %s", delay, prompt), nil
	}
}

func (p *variableDelayMockProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	atomic.AddInt64(&generateMessageCounter, 1)
	delay := p.getRandomDelay()

	select {
	case <-ctx.Done():
		return domain.Response{}, ctx.Err()
	case <-time.After(delay):
		var content string
		if len(messages) > 0 {
			content = fmt.Sprintf("Response after %v delay for: %s", delay, messages[len(messages)-1].Content)
		} else {
			content = fmt.Sprintf("Response after %v delay for empty messages", delay)
		}
		return domain.GetResponsePool().NewResponse(content), nil
	}
}

func (p *variableDelayMockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	atomic.AddInt64(&generateSchemaCounter, 1)
	delay := p.getRandomDelay()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
		return map[string]interface{}{
			"result": fmt.Sprintf("Structured response after %v delay for: %s", delay, prompt),
		}, nil
	}
}

func (p *variableDelayMockProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	atomic.AddInt64(&streamCounter, 1)
	tokenCh := make(chan domain.Token, 5)

	go func() {
		defer close(tokenCh)

		tokens := []string{"Variable ", "delay ", "streaming ", "response ", "for: ", prompt}

		for i, token := range tokens {
			delay := p.getRandomDelay() / 6 // Divide total delay among tokens

			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
				isLast := i == len(tokens)-1

				select {
				case <-ctx.Done():
					return
				case tokenCh <- domain.GetTokenPool().NewToken(token, isLast):
					// Token sent successfully
				}
			}
		}
	}()

	return tokenCh, nil
}

func (p *variableDelayMockProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	atomic.AddInt64(&streamMessageCounter, 1)
	tokenCh := make(chan domain.Token, 5)

	go func() {
		defer close(tokenCh)

		var lastContent string
		if len(messages) > 0 {
			lastContent = messages[len(messages)-1].Content
		} else {
			lastContent = "empty messages"
		}

		tokens := []string{"Variable ", "delay ", "streaming ", "response ", "for: ", lastContent}

		for i, token := range tokens {
			delay := p.getRandomDelay() / 6 // Divide total delay among tokens

			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
				isLast := i == len(tokens)-1

				select {
				case <-ctx.Done():
					return
				case tokenCh <- domain.GetTokenPool().NewToken(token, isLast):
					// Token sent successfully
				}
			}
		}
	}()

	return tokenCh, nil
}
