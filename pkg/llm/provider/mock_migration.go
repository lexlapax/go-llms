// ABOUTME: Migration utilities for transitioning to new mock infrastructure
// ABOUTME: Provides compatibility layer and helpers for upgrading existing mock usage

package provider

import (
	"context"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/fixtures"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// MockProviderAdapter wraps the new MockProvider to maintain backward compatibility
// Deprecated: Use fixtures.ChatGPTMockProvider() or fixtures.ClaudeMockProvider() for new code
type MockProviderAdapter struct {
	underlying *mocks.MockProvider
}

// NewMockProviderCompat creates a backward-compatible mock provider
// Deprecated: Use fixtures.ChatGPTMockProvider() or fixtures.ClaudeMockProvider() for new code
func NewMockProviderCompat(options ...domain.ProviderOption) *MockProviderAdapter {
	// Create a provider with basic responses
	provider := fixtures.ChatGPTMockProvider()

	return &MockProviderAdapter{
		underlying: provider,
	}
}

// Generate produces text from a prompt
func (p *MockProviderAdapter) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	return p.underlying.Generate(ctx, prompt, options...)
}

// GenerateMessage produces text from a list of messages
func (p *MockProviderAdapter) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	return p.underlying.GenerateMessage(ctx, messages, options...)
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *MockProviderAdapter) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	// For backward compatibility, generate mock data based on schema
	return generateMockDataFromSchema(schema), nil
}

// Stream streams responses token by token
func (p *MockProviderAdapter) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	return p.underlying.Stream(ctx, prompt, options...)
}

// StreamMessage streams responses from a list of messages
func (p *MockProviderAdapter) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	return p.underlying.StreamMessage(ctx, messages, options...)
}

// WithGenerateFunc sets a custom generate function
// Deprecated: Use fixtures or mocks.MockProvider.OnGenerate for new code
func (p *MockProviderAdapter) WithGenerateFunc(f func(ctx context.Context, prompt string, options ...domain.Option) (string, error)) *MockProviderAdapter {
	// Adapt the function to work with the new interface
	p.underlying.OnGenerate = f
	return p
}

// WithGenerateMessageFunc sets a custom generate message function
// Deprecated: Use mocks.MockProvider.OnGenerateMessage for new code
func (p *MockProviderAdapter) WithGenerateMessageFunc(f func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error)) *MockProviderAdapter {
	p.underlying.OnGenerateMessage = f
	return p
}

// WithPredefinedResponses sets predefined responses for specific prompts
// Deprecated: Use mocks.MockProvider.WithPatternResponse for new code
func (p *MockProviderAdapter) WithPredefinedResponses(responses map[string]string) *MockProviderAdapter {
	// Convert to pattern responses
	for prompt, response := range responses {
		p.underlying.WithPatternResponse(prompt, mocks.Response{
			Content: response,
			Metadata: map[string]interface{}{
				"mock_type": "predefined",
				"prompt":    prompt,
			},
		})
	}
	return p
}

// Helper functions for migration

// generateMockDataFromSchema creates mock data based on schema (simplified version)
func generateMockDataFromSchema(schema *schemaDomain.Schema) interface{} {
	if schema == nil {
		return map[string]interface{}{"result": "mock response"}
	}

	switch schema.Type {
	case "object":
		result := make(map[string]interface{})
		for propName, prop := range schema.Properties {
			switch prop.Type {
			case "string":
				result[propName] = "mock_" + propName
			case "integer":
				result[propName] = 42
			case "number":
				result[propName] = 42.5
			case "boolean":
				result[propName] = true
			case "array":
				result[propName] = []string{"item1", "item2"}
			default:
				result[propName] = "mock_" + propName
			}
		}
		return result
	case "array":
		return []string{"item1", "item2"}
	case "string":
		return "mock_string"
	case "integer":
		return 42
	case "number":
		return 42.5
	case "boolean":
		return true
	default:
		return "mock_default"
	}
}

// Migration helpers for common patterns

// CreateChatGPTLikeMock creates a mock provider that behaves like ChatGPT
func CreateChatGPTLikeMock() *MockProviderAdapter {
	provider := fixtures.ChatGPTMockProvider()
	return &MockProviderAdapter{underlying: provider}
}

// CreateClaudeLikeMock creates a mock provider that behaves like Claude
func CreateClaudeLikeMock() *MockProviderAdapter {
	provider := fixtures.ClaudeMockProvider()
	return &MockProviderAdapter{underlying: provider}
}

// CreateSlowMock creates a mock provider with configurable delays
func CreateSlowMock(delay time.Duration) *MockProviderAdapter {
	provider := fixtures.SlowMockProvider(delay)
	return &MockProviderAdapter{underlying: provider}
}

// CreateErrorMock creates a mock provider that returns errors
func CreateErrorMock(errorType string) *MockProviderAdapter {
	provider := fixtures.ErrorMockProvider(errorType)
	return &MockProviderAdapter{underlying: provider}
}

// CreateStreamingMock creates a mock provider with streaming responses
func CreateStreamingMock() *MockProviderAdapter {
	provider := fixtures.StreamingMockProvider()
	return &MockProviderAdapter{underlying: provider}
}
