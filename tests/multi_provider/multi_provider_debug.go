package multi_provider

import (
	"context"
	"testing"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// TestDebug provides a detailed debug of the multi-provider behavior
func TestDebug(t *testing.T) {
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
}
