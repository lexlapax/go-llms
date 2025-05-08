package provider

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestDeterministicMultiProvider demonstrates a deterministic way to test
// the primary provider strategy in MultiProvider
func TestDeterministicMultiProvider(t *testing.T) {
	// Create providers with tracking capability
	provider1Calls := 0
	provider1 := &mockProviderInTest{
		response:    "Provider1 Response",
		callCounter: &provider1Calls,
	}

	provider2Calls := 0
	provider2 := &mockProviderInTest{
		response:    "Provider2 Response",
		callCounter: &provider2Calls,
	}

	// Test with primary provider strategy
	// Instead of looking at the response content, we'll verify which provider was called
	t.Run("ExplicitPrimaryProvider", func(t *testing.T) {
		// Reset call counters
		provider1Calls = 0
		provider2Calls = 0

		// Create providers list (order doesn't matter)
		providers := []ProviderWeight{
			{Provider: provider1, Weight: 1.0, Name: "provider1"},
			{Provider: provider2, Weight: 1.0, Name: "provider2"},
		}

		// Create MultiProvider and explicitly set provider at index 1 (provider2) as primary
		multiProvider := NewMultiProvider(providers, StrategyPrimary).
			WithPrimaryProvider(1)

		// Generate - should use primary provider first
		_, err := multiProvider.Generate(context.Background(), "test prompt")
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		// Verify provider2 was called exactly once
		if provider2Calls != 1 {
			t.Errorf("Expected provider2 to be called once as the primary provider, got %d calls", provider2Calls)
		}

		// Verify provider1 was not called (since provider2 succeeded)
		if provider1Calls != 0 {
			t.Errorf("Expected provider1 not to be called when primary succeeds, got %d calls", provider1Calls)
		}
	})

	// Test primary provider fallback when primary fails
	t.Run("PrimaryFallback", func(t *testing.T) {
		// Reset call counters
		provider1Calls = 0
		provider2Calls = 0

		// Create failing primary provider
		failingProviderCalls := 0
		failingProvider := &mockProviderInTest{
			err:         context.DeadlineExceeded,
			callCounter: &failingProviderCalls,
		}

		// Create providers list with failing provider first
		providers := []ProviderWeight{
			{Provider: failingProvider, Weight: 1.0, Name: "failing"},
			{Provider: provider1, Weight: 1.0, Name: "provider1"},
		}

		// Create MultiProvider with default primary (index 0 - the failing provider)
		multiProvider := NewMultiProvider(providers, StrategyPrimary)

		// Generate - should use primary first, then fall back
		_, err := multiProvider.Generate(context.Background(), "test prompt")
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		// Verify failing provider was called
		if failingProviderCalls != 1 {
			t.Errorf("Expected failing provider to be called once, got %d calls", failingProviderCalls)
		}

		// Verify provider1 was called exactly once for fallback
		if provider1Calls != 1 {
			t.Errorf("Expected provider1 to be called once as fallback, got %d calls", provider1Calls)
		}
	})
}

// mockProviderInTest is a deterministic mock provider with call tracking for this test file
type mockProviderInTest struct {
	response    string
	err         error
	callCounter *int
}

func (m *mockProviderInTest) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	// Increment call counter if provided
	if m.callCounter != nil {
		*m.callCounter++
	}

	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *mockProviderInTest) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	// Increment call counter if provided
	if m.callCounter != nil {
		*m.callCounter++
	}

	if m.err != nil {
		return domain.Response{}, m.err
	}
	return domain.Response{Content: m.response}, nil
}

func (m *mockProviderInTest) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	// Increment call counter if provided
	if m.callCounter != nil {
		*m.callCounter++
	}

	if m.err != nil {
		return nil, m.err
	}

	ch := make(chan domain.Token)
	go func() {
		defer close(ch)
		ch <- domain.Token{Text: m.response, Finished: true}
	}()
	return ch, nil
}

func (m *mockProviderInTest) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	// Increment call counter if provided
	if m.callCounter != nil {
		*m.callCounter++
	}

	if m.err != nil {
		return nil, m.err
	}

	ch := make(chan domain.Token)
	go func() {
		defer close(ch)
		ch <- domain.Token{Text: m.response, Finished: true}
	}()
	return ch, nil
}

func (m *mockProviderInTest) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	// Increment call counter if provided
	if m.callCounter != nil {
		*m.callCounter++
	}

	if m.err != nil {
		return nil, m.err
	}
	return map[string]interface{}{"result": m.response}, nil
}
