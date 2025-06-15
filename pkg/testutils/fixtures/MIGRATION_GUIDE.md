# Provider Fixtures Migration Guide

This guide shows how to migrate from inline provider configurations to using centralized fixtures.

## Before: Inline Provider Configuration

```go
func TestWithInlineProvider(t *testing.T) {
    // Inline configuration - repetitive and hard to maintain
    provider := mocks.NewMockProvider("test-openai")
    provider.WithPatternResponse("(?i).*summarize.*", mocks.Response{
        Content: "Here's a summary...",
        Metadata: map[string]interface{}{
            "model": "gpt-4-turbo",
            "usage": map[string]interface{}{
                "prompt_tokens":     25,
                "completion_tokens": 15,
                "total_tokens":      40,
            },
        },
    })
    provider.WithDefaultResponse(mocks.Response{
        Content: "Default response",
        Metadata: map[string]interface{}{
            "model": "gpt-4-turbo",
            "usage": map[string]interface{}{
                "prompt_tokens":     10,
                "completion_tokens": 8,
                "total_tokens":      18,
            },
        },
    })
    
    // Test logic...
}

func TestWithInlineStreamingProvider(t *testing.T) {
    // Complex inline streaming setup
    provider := mocks.NewMockProvider("streaming")
    provider.OnStreamMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
        ch := make(chan ldomain.Token)
        go func() {
            defer close(ch)
            // Complex streaming logic repeated in every test...
        }()
        return ch, nil
    }
    
    // Test logic...
}
```

## After: Using Centralized Fixtures

```go
import "github.com/lexlapax/go-llms/pkg/testutils/fixtures"

func TestWithFixtureProvider(t *testing.T) {
    // Clean, standardized provider configuration
    provider := fixtures.OpenAIMockProvider()
    
    // Test logic...
}

func TestWithBasicProvider(t *testing.T) {
    // Simple provider for basic test scenarios
    provider := fixtures.BasicMockProvider()
    
    // Test logic...
}

func TestWithCustomContent(t *testing.T) {
    // Basic provider with specific content
    provider := fixtures.BasicMockProviderWithContent("Custom response")
    
    // Test logic...
}

func TestWithConfiguredProvider(t *testing.T) {
    // Easily configure specific settings
    provider := fixtures.ConfiguredOpenAIProvider("gpt-4o", 0.7, 500)
    
    // Test logic...
}

func TestWithStreamingProvider(t *testing.T) {
    // Realistic streaming behavior out of the box
    provider := fixtures.RealisticStreamingProvider()
    
    // Test logic...
}

func TestErrorScenarios(t *testing.T) {
    t.Run("rate_limit", func(t *testing.T) {
        provider := fixtures.RateLimitErrorProvider()
        // Test error handling...
    })
    
    t.Run("auth_failure", func(t *testing.T) {
        provider := fixtures.AuthenticationErrorProvider()
        // Test error handling...
    })
    
    t.Run("intermittent_failures", func(t *testing.T) {
        provider := fixtures.IntermittentErrorProvider(0.7) // 70% success rate
        // Test retry logic...
    })
}
```

## Available Provider Fixtures

### Basic Provider Types
- `BasicMockProvider()` - Simple provider for basic testing scenarios
- `BasicMockProviderWithContent(content)` - Simple provider with custom content
- `ChatGPTMockProvider()` - OpenAI ChatGPT-style responses
- `ClaudeMockProvider()` - Anthropic Claude-style responses  
- `OpenAIMockProvider()` - Enhanced OpenAI with realistic metadata
- `AnthropicMockProvider()` - Enhanced Anthropic with realistic metadata
- `GeminiMockProvider()` - Google Gemini with safety ratings

### Streaming Providers
- `StreamingMockProvider()` - Basic streaming functionality
- `RealisticStreamingProvider()` - Variable delays and realistic patterns
- `FastStreamingProvider()` - Minimal latency streaming

### Error Simulation
- `ErrorMockProvider(errorType)` - Generic error scenarios
- `RateLimitErrorProvider()` - Rate limiting with retry-after
- `AuthenticationErrorProvider()` - Authentication failures
- `NetworkErrorProvider()` - Network timeout/connection issues
- `IntermittentErrorProvider(successRate)` - Occasional failures

### Configuration-Specific
- `ConfiguredOpenAIProvider(model, temperature, maxTokens)` - Custom OpenAI config
- `ConfiguredAnthropicProvider(model, maxTokens, temperature)` - Custom Claude config
- `SlowMockProvider(delay)` - Providers with configurable latency

## Migration Benefits

1. **Reduced Duplication**: Common provider setups defined once
2. **Realistic Testing**: Provider-specific metadata and response patterns
3. **Better Error Testing**: Comprehensive error scenario coverage
4. **Maintainability**: Updates to provider behavior in one place
5. **Consistency**: Standardized testing patterns across the codebase

## Migration Steps

1. **Identify Inline Configurations**: Search for `NewMockProvider` in test files
2. **Choose Appropriate Fixture**: Select the fixture that matches your test needs
3. **Replace Inline Setup**: Replace provider creation with fixture call
4. **Update Test Logic**: Adjust test assertions if needed
5. **Verify Behavior**: Ensure tests still pass with expected behavior

## Custom Fixtures

If you need provider behavior not covered by existing fixtures, consider adding new fixtures to `pkg/testutils/fixtures/providers.go` rather than creating inline configurations.