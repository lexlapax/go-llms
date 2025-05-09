package testutils_test

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// TestDeterministicMultiProvider tests the primary provider strategy in a deterministic way
// This test demonstrates the correct approach to testing the primary provider strategy
func TestDeterministicMultiProvider(t *testing.T) {
	// Create providers with tracking capability
	provider1Calls := 0
	provider1 := &testutils.TestMockProvider{
		GenerateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			provider1Calls++
			return "Provider1 Response", nil
		},
	}

	provider2Calls := 0
	provider2 := &testutils.TestMockProvider{
		GenerateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			provider2Calls++
			return "Provider2 Response", nil
		},
	}

	// Test with primary provider strategy
	// Instead of looking at the response content, we'll verify which provider was called
	t.Run("ExplicitPrimaryProvider", func(t *testing.T) {
		// Reset call counters
		provider1Calls = 0
		provider2Calls = 0

		// Create providers list (order doesn't matter)
		providers := []provider.ProviderWeight{
			{Provider: provider1, Weight: 1.0, Name: "provider1"},
			{Provider: provider2, Weight: 1.0, Name: "provider2"},
		}

		// Create MultiProvider and explicitly set provider at index 1 (provider2) as primary
		multiProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).
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
		failingProvider := &testutils.TestMockProvider{
			GenerateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
				return "", context.DeadlineExceeded
			},
		}

		// Create providers list with failing provider first
		providers := []provider.ProviderWeight{
			{Provider: failingProvider, Weight: 1.0, Name: "failing"},
			{Provider: provider1, Weight: 1.0, Name: "provider1"},
		}

		// Create MultiProvider with default primary (index 0 - the failing provider)
		multiProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary)

		// Generate - should use primary first, then fall back
		_, err := multiProvider.Generate(context.Background(), "test prompt")
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		// Verify provider1 was called exactly once for fallback
		if provider1Calls != 1 {
			t.Errorf("Expected provider1 to be called once as fallback, got %d calls", provider1Calls)
		}
	})
}
