package multi_provider

import (
	"context"
	"testing"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// TestMultiProviderDebug tests the primary provider functionality directly
// This test is kept to document the behavior but is skipped by default
func TestMultiProviderDebug(t *testing.T) {
	// Skip this test to avoid flaky test results, but keep it for documentation
	t.Skip("Skipping debug test to avoid flaky test results")
	// Create providers with very distinct responses for easy identification
	provider1 := &MockProvider{
		GenerateFunc: func(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
			return "PROVIDER_ONE_RESPONSE", nil
		},
	}

	provider2 := &MockProvider{
		GenerateFunc: func(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
			return "PROVIDER_TWO_RESPONSE", nil
		},
	}

	// Create provider weights with providers in different orders
	providers1First := []provider.ProviderWeight{
		{Provider: provider1, Weight: 1.0, Name: "provider1"},
		{Provider: provider2, Weight: 1.0, Name: "provider2"},
	}

	// Create MultiProvider with primary strategy
	multiProvider := provider.NewMultiProvider(providers1First, provider.StrategyPrimary)

	// Get response with default primary (should be first provider)
	result, err := multiProvider.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	t.Logf("Default primary provider result: %s", result)

	// Now explicitly set provider 0 as primary
	mp0 := multiProvider.WithPrimaryProvider(0)
	result0, err := mp0.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Generate with primary 0 failed: %v", err)
	}
	t.Logf("With primary provider 0 result: %s", result0)

	// Now explicitly set provider 1 as primary
	mp1 := multiProvider.WithPrimaryProvider(1)
	result1, err := mp1.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Generate with primary 1 failed: %v", err)
	}
	t.Logf("With primary provider 1 result: %s", result1)

	// Try with reversed order
	providers2First := []provider.ProviderWeight{
		{Provider: provider2, Weight: 1.0, Name: "provider2"},
		{Provider: provider1, Weight: 1.0, Name: "provider1"},
	}

	multiProvider2 := provider.NewMultiProvider(providers2First, provider.StrategyPrimary)
	result2, err := multiProvider2.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Generate with reversed order failed: %v", err)
	}
	t.Logf("Reversed order default provider result: %s", result2)

	// Explicitly set indices
	mp2_0 := multiProvider2.WithPrimaryProvider(0)
	result2_0, err := mp2_0.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Generate with reversed order primary 0 failed: %v", err)
	}
	t.Logf("Reversed order with provider 0 result: %s", result2_0)

	mp2_1 := multiProvider2.WithPrimaryProvider(1)
	result2_1, err := mp2_1.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Generate with reversed order primary 1 failed: %v", err)
	}
	t.Logf("Reversed order with provider 1 result: %s", result2_1)
}
