// ABOUTME: Integration helpers to add metadata support to existing providers
// ABOUTME: Wraps providers with metadata capabilities for backward compatibility

package provider

import (
	"context"
	"fmt"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// MetadataProviderWrapper wraps a provider with metadata support
type MetadataProviderWrapper struct {
	domain.Provider
	metadata ProviderMetadata
}

// NewMetadataProviderWrapper creates a new wrapper
func NewMetadataProviderWrapper(provider domain.Provider, metadata ProviderMetadata) *MetadataProviderWrapper {
	return &MetadataProviderWrapper{
		Provider: provider,
		metadata: metadata,
	}
}

// GetMetadata implements MetadataProvider
func (w *MetadataProviderWrapper) GetMetadata() ProviderMetadata {
	return w.metadata
}

// WrapWithMetadata adds metadata support to a provider
func WrapWithMetadata(provider domain.Provider, metadata ProviderMetadata) domain.Provider {
	// If provider already supports metadata, return as-is
	if _, ok := provider.(MetadataProvider); ok {
		return provider
	}
	return NewMetadataProviderWrapper(provider, metadata)
}

// GetProviderMetadata retrieves metadata from a provider if supported
func GetProviderMetadata(provider domain.Provider) (ProviderMetadata, bool) {
	if metaProvider, ok := provider.(MetadataProvider); ok {
		return metaProvider.GetMetadata(), true
	}
	return nil, false
}

// SelectProviderByCapability selects a provider from registry based on capability
func SelectProviderByCapability(registry *DynamicRegistry, capability Capability) (domain.Provider, error) {
	providers := registry.ListProvidersByCapability(capability)
	if len(providers) == 0 {
		return nil, fmt.Errorf("no provider found with capability: %s", capability)
	}

	// Return the first matching provider
	return registry.GetProvider(providers[0])
}

// SelectModelByCapability selects a model that has specific capabilities
func SelectModelByCapability(registry *DynamicRegistry, ctx context.Context, capabilities ...Capability) (string, *ModelInfo, error) {
	providers := registry.ListProviders()

	for _, providerName := range providers {
		metadata, err := registry.GetMetadata(providerName)
		if err != nil {
			continue
		}

		// Check each model
		models, err := metadata.GetModels(ctx)
		if err != nil {
			continue // Skip providers that fail to load models
		}

		for _, model := range models {
			hasAll := true
			for _, required := range capabilities {
				found := false
				for _, cap := range model.Capabilities {
					if cap == required {
						found = true
						break
					}
				}
				if !found {
					hasAll = false
					break
				}
			}

			if hasAll && !model.Deprecated {
				return providerName, &model, nil
			}
		}
	}

	return "", nil, fmt.Errorf("no model found with all required capabilities")
}

// ProviderComparison represents a comparative analysis of a provider's capabilities and features.
// It is used to compare different LLM providers based on their supported features.
type ProviderComparison struct {
	Provider     string
	Metadata     ProviderMetadata
	Capabilities []Capability
	ModelCount   int
	HasVision    bool
	HasStreaming bool
	HasFunctions bool
}

// CompareProviders generates a comparison of all providers in the registry
func CompareProviders(registry *DynamicRegistry, ctx context.Context) []ProviderComparison {
	var comparisons []ProviderComparison

	for _, name := range registry.ListProviders() {
		metadata, err := registry.GetMetadata(name)
		if err != nil {
			continue
		}

		// Get models with context
		models, err := metadata.GetModels(ctx)
		modelCount := 0
		if err == nil {
			modelCount = len(models)
		}

		comp := ProviderComparison{
			Provider:     name,
			Metadata:     metadata,
			Capabilities: metadata.GetCapabilities(),
			ModelCount:   modelCount,
		}

		// Check specific capabilities
		for _, cap := range metadata.GetCapabilities() {
			switch cap {
			case CapabilityVision:
				comp.HasVision = true
			case CapabilityStreaming:
				comp.HasStreaming = true
			case CapabilityFunctionCalling:
				comp.HasFunctions = true
			}
		}

		comparisons = append(comparisons, comp)
	}

	return comparisons
}

// CreateProviderWithBestModel creates a provider with the best model for given constraints
func CreateProviderWithBestModel(registry *DynamicRegistry, ctx context.Context, minContext int, maxPrice float64, capabilities ...Capability) (domain.Provider, *ModelInfo, error) {
	var bestProvider string
	var bestModel *ModelInfo
	bestPrice := maxPrice + 1 // Start with higher than max

	for _, providerName := range registry.ListProviders() {
		metadata, err := registry.GetMetadata(providerName)
		if err != nil {
			continue
		}

		models, err := metadata.GetModels(ctx)
		if err != nil {
			continue // Skip providers that fail to load models
		}

		for _, model := range models {
			// Skip deprecated models
			if model.Deprecated {
				continue
			}

			// Check context window
			if model.ContextWindow < minContext {
				continue
			}

			// Check capabilities
			hasAll := true
			for _, required := range capabilities {
				found := false
				for _, cap := range model.Capabilities {
					if cap == required {
						found = true
						break
					}
				}
				if !found {
					hasAll = false
					break
				}
			}

			if !hasAll {
				continue
			}

			// Check price (use input pricing as reference)
			if model.InputPricing != nil && model.InputPricing.Price < bestPrice {
				bestPrice = model.InputPricing.Price
				bestProvider = providerName
				bestModel = &model
			}
		}
	}

	if bestProvider == "" {
		return nil, nil, fmt.Errorf("no suitable model found within constraints")
	}

	// Get the provider
	provider, err := registry.GetProvider(bestProvider)
	if err != nil {
		return nil, nil, err
	}

	// Configure provider with the selected model
	configuredProvider := &modelConfiguredProvider{
		Provider: provider,
		model:    bestModel.ID,
	}

	return configuredProvider, bestModel, nil
}

// modelConfiguredProvider wraps a provider with a specific model configuration
type modelConfiguredProvider struct {
	domain.Provider
	model string
}

func (p *modelConfiguredProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	// Prepend model option
	opts := append([]domain.Option{domain.WithModel(p.model)}, options...)
	return p.Provider.Generate(ctx, prompt, opts...)
}

func (p *modelConfiguredProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	opts := append([]domain.Option{domain.WithModel(p.model)}, options...)
	return p.Provider.GenerateMessage(ctx, messages, opts...)
}

func (p *modelConfiguredProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...domain.Option) (interface{}, error) {
	opts := append([]domain.Option{domain.WithModel(p.model)}, options...)
	return p.Provider.GenerateWithSchema(ctx, prompt, schema, opts...)
}

func (p *modelConfiguredProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	opts := append([]domain.Option{domain.WithModel(p.model)}, options...)
	return p.Provider.Stream(ctx, prompt, opts...)
}

func (p *modelConfiguredProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	opts := append([]domain.Option{domain.WithModel(p.model)}, options...)
	return p.Provider.StreamMessage(ctx, messages, opts...)
}
