// ABOUTME: Example demonstrating provider metadata and dynamic registry
// ABOUTME: Shows capability discovery, model selection, and dynamic provider creation

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
	fmt.Println("=== Provider Metadata Example ===")

	// Create and configure registry
	registry := provider.NewDynamicRegistry()

	// Register default factories
	if err := provider.RegisterDefaultFactories(registry); err != nil {
		log.Fatal("Failed to register factories:", err)
	}

	// Example 1: Explore provider templates
	fmt.Println("\n1. Available Provider Templates:")
	exploreTemplates(registry)

	// Example 2: Create providers dynamically
	fmt.Println("\n2. Dynamic Provider Creation:")
	createProvidersDynamically(registry)

	// Example 3: Capability-based discovery
	fmt.Println("\n3. Capability-Based Discovery:")
	discoverByCapability(registry)

	// Example 4: Model comparison
	fmt.Println("\n4. Model Comparison:")
	compareModels(registry)

	// Example 5: Export/Import configuration
	fmt.Println("\n5. Configuration Management:")
	manageConfiguration(registry)

	// Example 6: Best model selection
	fmt.Println("\n6. Best Model Selection:")
	selectBestModel(registry)
}

func exploreTemplates(registry *provider.DynamicRegistry) {
	templates := registry.ListTemplates()

	for _, template := range templates {
		fmt.Printf("\n  Template: %s\n", template.Name)
		fmt.Printf("  Type: %s\n", template.Type)
		fmt.Printf("  Description: %s\n", template.Description)

		// Show required fields
		fmt.Println("  Required fields:")
		for name, field := range template.Schema.Fields {
			if field.Required {
				fmt.Printf("    - %s (%s): %s\n", name, field.Type, field.Description)
				if field.EnvVar != "" {
					fmt.Printf("      Can use env var: %s\n", field.EnvVar)
				}
			}
		}

		// Show examples
		if len(template.Examples) > 0 {
			fmt.Println("  Examples:")
			for _, example := range template.Examples {
				fmt.Printf("    - %s: %s\n", example.Name, example.Description)
			}
		}
	}
}

func createProvidersDynamically(registry *provider.DynamicRegistry) {
	// Create a mock provider for testing
	mockConfig := map[string]interface{}{
		"default_response": "I'm a dynamically created mock provider!",
		"responses": map[string]interface{}{
			"hello": "Hello from dynamic mock!",
			"test":  "This is a test response",
		},
	}

	err := registry.CreateProviderFromTemplate("mock", "dynamic-mock", mockConfig)
	if err != nil {
		log.Printf("Failed to create mock provider: %v", err)
		return
	}

	// Test the provider
	p, err := registry.GetProvider("dynamic-mock")
	if err != nil {
		log.Printf("Failed to get provider: %v", err)
		return
	}

	response, err := p.Generate(context.TODO(), "hello")
	if err != nil {
		log.Printf("Failed to generate: %v", err)
		return
	}

	fmt.Printf("  Created provider 'dynamic-mock'\n")
	fmt.Printf("  Test response: %s\n", response)

	// Create provider with metadata
	mockProvider := provider.NewMockProvider()
	metadata := &provider.BaseProviderMetadata{
		ProviderName:        "Custom Mock",
		ProviderDescription: "A custom mock provider with metadata",
		Capabilities:        []provider.Capability{provider.CapabilityStreaming},
		// Note: Models are now loaded dynamically via GetModels()
	}

	err = registry.RegisterProvider("custom-mock", mockProvider, metadata)
	if err != nil {
		log.Printf("Failed to register custom provider: %v", err)
		return
	}

	fmt.Printf("\n  Registered provider 'custom-mock' with metadata\n")
}

func discoverByCapability(registry *provider.DynamicRegistry) {
	// Create providers with different capabilities
	createTestProviders(registry)

	// Find providers with streaming
	streamingProviders := registry.ListProvidersByCapability(provider.CapabilityStreaming)
	fmt.Printf("  Providers with streaming: %v\n", streamingProviders)

	// Find providers with vision
	visionProviders := registry.ListProvidersByCapability(provider.CapabilityVision)
	fmt.Printf("  Providers with vision: %v\n", visionProviders)

	// Find providers with function calling
	functionProviders := registry.ListProvidersByCapability(provider.CapabilityFunctionCalling)
	fmt.Printf("  Providers with function calling: %v\n", functionProviders)
}

func compareModels(registry *provider.DynamicRegistry) {
	// Create test providers with real metadata
	openaiMeta := provider.NewOpenAIMetadata()
	anthropicMeta := provider.NewAnthropicMetadata()

	_ = registry.RegisterProvider("openai-test", provider.NewMockProvider(), openaiMeta)
	_ = registry.RegisterProvider("anthropic-test", provider.NewMockProvider(), anthropicMeta)

	ctx := context.Background()
	comparisons := provider.CompareProviders(registry, ctx)

	fmt.Println("  Provider Comparison:")
	fmt.Println("  ┌──────────────┬─────────┬──────────┬───────────┬───────────┐")
	fmt.Println("  │ Provider     │ Models  │ Streaming│ Vision    │ Functions │")
	fmt.Println("  ├──────────────┼─────────┼──────────┼───────────┼───────────┤")

	for _, comp := range comparisons {
		fmt.Printf("  │ %-12s │ %-7d │ %-8v │ %-9v │ %-9v │\n",
			comp.Provider,
			comp.ModelCount,
			comp.HasStreaming,
			comp.HasVision,
			comp.HasFunctions,
		)
	}
	fmt.Println("  └──────────────┴─────────┴──────────┴───────────┴───────────┘")

	// Show detailed model info for one provider
	fmt.Println("\n  OpenAI Models:")
	models, err := openaiMeta.GetModels(ctx)
	if err != nil {
		log.Printf("Failed to get models: %v", err)
		return
	}
	for i, model := range models {
		if i >= 3 { // Show only first 3
			fmt.Printf("  ... and %d more models\n", len(models)-3)
			break
		}
		fmt.Printf("    - %s: %s\n", model.ID, model.Name)
		fmt.Printf("      Context: %d tokens, Max output: %d tokens\n",
			model.ContextWindow, model.MaxTokens)
		if model.InputPricing != nil {
			fmt.Printf("      Pricing: $%.4f per %d tokens (input)\n",
				model.InputPricing.Price, model.InputPricing.PerTokens)
		}
	}
}

func manageConfiguration(registry *provider.DynamicRegistry) {
	// Export current configuration
	config, err := registry.ExportConfig()
	if err != nil {
		log.Printf("Failed to export config: %v", err)
		return
	}

	// Pretty print config
	configJSON, _ := json.MarshalIndent(config, "  ", "  ")
	fmt.Printf("  Exported configuration:\n  %s\n", configJSON)

	// Create a new registry and import
	newRegistry := provider.NewDynamicRegistry()
	_ = provider.RegisterDefaultFactories(newRegistry)

	err = newRegistry.ImportConfig(config)
	if err != nil {
		log.Printf("Failed to import config: %v", err)
		return
	}

	fmt.Printf("\n  Successfully imported configuration into new registry\n")
	fmt.Printf("  Providers in new registry: %v\n", newRegistry.ListProviders())
}

func selectBestModel(registry *provider.DynamicRegistry) {
	// Register providers with metadata
	openaiMeta := provider.NewOpenAIMetadata()
	anthropicMeta := provider.NewAnthropicMetadata()

	_ = registry.RegisterProvider("openai", provider.NewMockProvider(), openaiMeta)
	_ = registry.RegisterProvider("anthropic", provider.NewMockProvider(), anthropicMeta)

	// Find best model for specific requirements
	minContext := 100000 // Need at least 100K context
	maxPrice := 5.0      // Max $5 per million tokens input

	ctx := context.Background()
	p, model, err := provider.CreateProviderWithBestModel(
		registry,
		ctx,
		minContext,
		maxPrice,
		provider.CapabilityStreaming,
		provider.CapabilityVision,
	)

	if err != nil {
		log.Printf("Failed to find suitable model: %v", err)
		return
	}

	fmt.Printf("  Best model for requirements:\n")
	fmt.Printf("    Min context: %d tokens\n", minContext)
	fmt.Printf("    Max price: $%.2f per million tokens\n", maxPrice)
	fmt.Printf("    Required: Streaming, Vision\n")
	fmt.Printf("\n  Selected: %s (%s)\n", model.Name, model.ID)
	fmt.Printf("    Context: %d tokens\n", model.ContextWindow)
	if model.InputPricing != nil {
		fmt.Printf("    Price: $%.2f per %d tokens\n",
			model.InputPricing.Price, model.InputPricing.PerTokens)
	}
	fmt.Printf("    Provider configured with this model: %v\n", p != nil)
}

// Helper to create test providers
func createTestProviders(registry *provider.DynamicRegistry) {
	// Provider with streaming only
	_ = registry.RegisterProvider("stream-only", provider.NewMockProvider(), &provider.BaseProviderMetadata{
		ProviderName: "Stream Only",
		Capabilities: []provider.Capability{provider.CapabilityStreaming},
	})

	// Provider with vision
	_ = registry.RegisterProvider("vision-capable", provider.NewMockProvider(), &provider.BaseProviderMetadata{
		ProviderName: "Vision Capable",
		Capabilities: []provider.Capability{provider.CapabilityStreaming, provider.CapabilityVision},
	})

	// Provider with all capabilities
	_ = registry.RegisterProvider("full-featured", provider.NewMockProvider(), &provider.BaseProviderMetadata{
		ProviderName: "Full Featured",
		Capabilities: []provider.Capability{
			provider.CapabilityStreaming,
			provider.CapabilityVision,
			provider.CapabilityFunctionCalling,
			provider.CapabilityStructuredOutput,
		},
	})
}
