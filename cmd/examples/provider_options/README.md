# Provider Options Example

This example demonstrates how to use the interface-based provider option system to configure LLM providers with both common and provider-specific options.

## Features Demonstrated

- Using common options that work across all providers
  - BaseURL - Setting a custom API endpoint
  - HTTPClient - Using a custom HTTP client with timeout
  - Headers - Adding custom headers to requests
  
- Using provider-specific options
  - OpenAI: Organization ID and logit bias
  - Anthropic: System prompt and metadata
  - Gemini: Generation config and safety settings

- Creating and composing multiple options
- Using options with environment variables

## Running the Example

```bash
# Set your API keys as environment variables
export OPENAI_API_KEY=your_openai_key
export ANTHROPIC_API_KEY=your_anthropic_key
export GEMINI_API_KEY=your_gemini_key

# Build and run
make example EXAMPLE=provider_options
./bin/provider_options
```

## Code Structure

- `main.go` - Main example showcasing different provider options
- `main_test.go` - Tests for the example

## Key Concepts

### The Option Pattern

The interface-based provider option system uses Go's interface capabilities to create a flexible, type-safe way to configure providers. Options implement provider-specific interfaces that apply configuration to the appropriate provider types.

### Common Options

Common options implement all provider-specific interfaces and can be used with any provider:

```go
// Create common options
baseURLOption := domain.NewBaseURLOption("https://custom-api.example.com")
httpClientOption := domain.NewHTTPClientOption(&http.Client{Timeout: 30 * time.Second})

// These options can be used with any provider
openaiProvider := provider.NewOpenAIProvider(apiKey, "gpt-4o", baseURLOption, httpClientOption)
anthropicProvider := provider.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-latest", baseURLOption, httpClientOption)
geminiProvider := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash-lite", baseURLOption, httpClientOption)
```

### Provider-Specific Options

Provider-specific options only implement a single provider interface and are used with only one provider type:

```go
// OpenAI-specific options
orgOption := domain.NewOpenAIOrganizationOption("org-123")
logitBiasOption := domain.NewOpenAILogitBiasOption(map[string]float64{"50256": -100})

// Anthropic-specific options
systemPromptOption := domain.NewAnthropicSystemPromptOption("You are a helpful assistant.")
metadataOption := domain.NewAnthropicMetadataOption(map[string]string{"user_id": "user-123"})

// Gemini-specific options
generationConfigOption := domain.NewGeminiGenerationConfigOption().WithTopK(20)
safetySettingsOption := domain.NewGeminiSafetySettingsOption([]map[string]interface{}{
    {"category": "HARM_CATEGORY_HARASSMENT", "threshold": "BLOCK_MEDIUM_AND_ABOVE"},
})
```

### Creating Providers with Options

Options are passed to provider constructors, which apply the appropriate options based on their interfaces:

```go
// Only options implementing OpenAIOption will be applied
openaiProvider := provider.NewOpenAIProvider(apiKey, "gpt-4o", baseURLOption, orgOption)

// Only options implementing AnthropicOption will be applied
anthropicProvider := provider.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-latest", baseURLOption, systemPromptOption)

// Only options implementing GeminiOption will be applied
geminiProvider := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash-lite", baseURLOption, generationConfigOption)
```