# Adding a New Provider to Go-LLMs

> **[Documentation Home](/REFERENCE.md) / [Technical Documentation](/docs/technical/README.md) / Adding a New Provider**

This guide provides a step-by-step approach to adding a new LLM provider to the Go-LLMs library. By following these instructions, you'll learn how to implement, test, and integrate a new provider into the existing architecture.

## Overview

Adding a new provider to Go-LLMs involves several steps:

1. Implementing the provider interface
2. Writing unit tests
3. Creating integration tests
4. Updating utility functions
5. Integrating with multi-provider functionality
6. Creating example applications
7. Updating documentation

## Step 1: Implement the Provider

### Create Provider Files

First, create a new file in the `pkg/llm/provider` directory:

```bash
touch pkg/llm/provider/new_provider.go
touch pkg/llm/provider/new_provider_test.go
```

### Implement the Provider Interface

Your provider must implement the `domain.Provider` interface defined in `pkg/llm/domain/interfaces.go`. Here's a template for implementing a new provider:

```go
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// NewProviderOption is a functional option for configuring the provider
type NewProviderOption func(*NewProvider)

// NewProvider implements the Provider interface for the new service
type NewProvider struct {
	apiKey   string
	model    string
	baseURL  string
	client   *http.Client
	timeout  time.Duration
	// Add other necessary fields
}

// Create configuration options

// WithNewProviderBaseURL sets a custom base URL for the provider
func WithNewProviderBaseURL(baseURL string) NewProviderOption {
	return func(p *NewProvider) {
		p.baseURL = baseURL
	}
}

// WithNewProviderTimeout sets a custom timeout for the provider
func WithNewProviderTimeout(timeout time.Duration) NewProviderOption {
	return func(p *NewProvider) {
		p.timeout = timeout
	}
}

// NewNewProvider creates a new provider with the given API key and model
func NewNewProvider(apiKey, model string, options ...NewProviderOption) *NewProvider {
	provider := &NewProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.newprovider.com/v1", // Default base URL
		client:  &http.Client{},
		timeout: 30 * time.Second,
	}

	// Apply options
	for _, option := range options {
		option(provider)
	}

	return provider
}

// Generate generates text based on a prompt
func (p *NewProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	// Parse options
	opts := domain.NewOptions()
	for _, opt := range options {
		opt(&opts)
	}

	// Create request
	// Make API call
	// Handle response
	// Return result and any errors
}

// GenerateMessage generates a response based on a conversation
func (p *NewProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	// Implement message-based generation
}

// GenerateWithSchema generates structured data according to a schema
func (p *NewProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	// Implement schema-based generation
}

// Stream streams a response token by token
func (p *NewProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	// Implement streaming
}

// StreamMessage streams a response for a conversation
func (p *NewProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	// Implement message-based streaming
}

// Helper functions specific to this provider
// ...
```

### Implement Error Handling

Create provider-specific error handling that maps the provider's API errors to the standard error types defined in `pkg/llm/domain/errors.go`. This ensures consistent error handling across all providers:

```go
// Handle specific provider API errors
func (p *NewProvider) handleAPIError(statusCode int, errorResponse map[string]interface{}) error {
	// Map provider-specific errors to domain error types
	if statusCode == 401 || statusCode == 403 {
		return domain.ErrAPIKeyInvalid
	} else if statusCode == 429 {
		return domain.ErrRateLimitExceeded
	}
	// Add other error mappings
	
	return fmt.Errorf("provider API error: %v", errorResponse)
}
```

## Step 2: Write Unit Tests

Create comprehensive unit tests for your provider in the `new_provider_test.go` file. Make sure to test all the methods in the interface, including edge cases and error scenarios:

```go
package provider

import (
	"context"
	"testing"
	
	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestNewProvider(t *testing.T) {
	// Test provider creation with default and custom options
	t.Run("NewNewProvider", func(t *testing.T) {
		provider := NewNewProvider("test-key", "test-model")
		if provider.apiKey != "test-key" {
			t.Errorf("Expected API key to be 'test-key', got '%s'", provider.apiKey)
		}
		if provider.model != "test-model" {
			t.Errorf("Expected model to be 'test-model', got '%s'", provider.model)
		}
		
		// Test with custom options
		provider = NewNewProvider("test-key", "test-model", 
			WithNewProviderBaseURL("https://custom.example.com"),
			WithNewProviderTimeout(60*time.Second),
		)
		if provider.baseURL != "https://custom.example.com" {
			t.Errorf("Expected baseURL to be 'https://custom.example.com', got '%s'", provider.baseURL)
		}
	})
	
	// Test each method with mock responses
	// ...
}
```

Use mocks for the HTTP client to simulate API responses without making actual network calls:

```go
// mockHTTPClient simulates HTTP responses for testing
type mockHTTPClient struct {
	doFunc func(*http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// Helper function to create a mock response
func mockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
```

## Step 3: Create Integration Tests

Integration tests verify that your provider works correctly with actual API calls. Create a new file in the `tests/integration` directory:

```bash
touch tests/integration/new_provider_e2e_test.go
```

Implement the integration test:

```go
package integration

import (
	"context"
	"os"
	"testing"
	
	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestNewProviderIntegration(t *testing.T) {
	// Skip if API key not available
	apiKey := os.Getenv("NEW_PROVIDER_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: NEW_PROVIDER_API_KEY environment variable not set")
	}
	
	// Create provider
	newProvider := provider.NewNewProvider(apiKey, "default-model")
	
	// Basic generation test
	t.Run("Generate", func(t *testing.T) {
		response, err := newProvider.Generate(context.Background(), "Hello, world!")
		if err != nil {
			t.Fatalf("Generation error: %v", err)
		}
		if response == "" {
			t.Error("Expected non-empty response")
		}
	})
	
	// Add tests for other methods
	// ...
}
```

Also add an agent integration test to verify that the provider works with the agent system:

```bash
touch tests/integration/new_provider_agent_e2e_test.go
```

```go
package integration

import (
	"context"
	"os"
	"testing"
	
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestNewProviderWithAgent(t *testing.T) {
	// Skip if API key not available
	apiKey := os.Getenv("NEW_PROVIDER_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: NEW_PROVIDER_API_KEY environment variable not set")
	}
	
	// Create provider and agent
	newProvider := provider.NewNewProvider(apiKey, "default-model")
	agent := workflow.NewAgent(newProvider)
	
	// Add some tools for testing
	agent.AddTool(tools.NewTool("date", "Get current date", 
		func() string { return time.Now().Format("2006-01-02") },
		nil,
	))
	
	// Test agent with the provider
	result, err := agent.Run(context.Background(), "What is today's date?")
	if err != nil {
		t.Fatalf("Agent error: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
}
```

## Step 4: Update Utility Functions

Update the utility functions in `pkg/util/llmutil` to support the new provider:

### Update Model Configuration

In `pkg/util/llmutil/llmutil.go`, add the new provider to the `WithProviderOptions` function:

```go
func WithProviderOptions(config ModelConfig) ([]interface{}, error) {
	var options []interface{}
	
	if config.BaseURL != "" {
		switch config.Provider {
		case "openai":
			options = append(options, provider.WithBaseURL(config.BaseURL))
		case "anthropic":
			options = append(options, provider.WithAnthropicBaseURL(config.BaseURL))
		case "gemini":
			options = append(options, provider.WithGeminiBaseURL(config.BaseURL))
		case "newprovider":
			options = append(options, provider.WithNewProviderBaseURL(config.BaseURL))
		}
	}
	
	return options, nil
}
```

### Update CreateProvider Function

Add the new provider to the `CreateProvider` function:

```go
func CreateProvider(config ModelConfig) (domain.Provider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	var llmProvider domain.Provider
	
	options, err := WithProviderOptions(config)
	if err != nil {
		return nil, err
	}
	
	switch config.Provider {
	case "openai":
		openAIOptions := make([]provider.OpenAIOption, 0, len(options))
		for _, opt := range options {
			if o, ok := opt.(provider.OpenAIOption); ok {
				openAIOptions = append(openAIOptions, o)
			}
		}
		llmProvider = provider.NewOpenAIProvider(config.APIKey, config.Model, openAIOptions...)
	
	case "anthropic":
		anthropicOptions := make([]provider.AnthropicOption, 0, len(options))
		for _, opt := range options {
			if o, ok := opt.(provider.AnthropicOption); ok {
				anthropicOptions = append(anthropicOptions, o)
			}
		}
		llmProvider = provider.NewAnthropicProvider(config.APIKey, config.Model, anthropicOptions...)
	
	case "gemini":
		geminiOptions := make([]provider.GeminiOption, 0, len(options))
		for _, opt := range options {
			if o, ok := opt.(provider.GeminiOption); ok {
				geminiOptions = append(geminiOptions, o)
			}
		}
		llmProvider = provider.NewGeminiProvider(config.APIKey, config.Model, geminiOptions...)
		
	case "newprovider":
		newProviderOptions := make([]provider.NewProviderOption, 0, len(options))
		for _, opt := range options {
			if o, ok := opt.(provider.NewProviderOption); ok {
				newProviderOptions = append(newProviderOptions, o)
			}
		}
		llmProvider = provider.NewNewProvider(config.APIKey, config.Model, newProviderOptions...)
	
	case "mock":
		llmProvider = provider.NewMockProvider()
	
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	return llmProvider, nil
}
```

### Update ProviderFromEnv Function

Add the new provider to the `ProviderFromEnv` function:

```go
func ProviderFromEnv() (domain.Provider, string, string, error) {
	// Check for API keys in environment variables
	openAIKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY") 
	geminiKey := os.Getenv("GEMINI_API_KEY")
	newProviderKey := os.Getenv("NEW_PROVIDER_API_KEY")
	
	// Default models for each provider
	openAIModel := os.Getenv("OPENAI_MODEL")
	if openAIModel == "" {
		openAIModel = "gpt-4o"
	}
	
	anthropicModel := os.Getenv("ANTHROPIC_MODEL")
	if anthropicModel == "" {
		anthropicModel = "claude-3-5-sonnet-latest"
	}
	
	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-2.0-flash-lite"
	}
	
	newProviderModel := os.Getenv("NEW_PROVIDER_MODEL")
	if newProviderModel == "" {
		newProviderModel = "default-model"
	}
	
	// Try to create a provider in order of preference
	if openAIKey != "" {
		provider := provider.NewOpenAIProvider(openAIKey, openAIModel)
		return provider, "openai", openAIModel, nil
	}
	
	if anthropicKey != "" {
		provider := provider.NewAnthropicProvider(anthropicKey, anthropicModel)
		return provider, "anthropic", anthropicModel, nil
	}
	
	if geminiKey != "" {
		provider := provider.NewGeminiProvider(geminiKey, geminiModel)
		return provider, "gemini", geminiModel, nil
	}
	
	if newProviderKey != "" {
		provider := provider.NewNewProvider(newProviderKey, newProviderModel)
		return provider, "newprovider", newProviderModel, nil
	}
	
	// If no API keys are found, create a mock provider
	mockProvider := provider.NewMockProvider()
	return mockProvider, "mock", "default", nil
}
```

## Step 5: Integrate with Multi-Provider Functionality

If you want to use the new provider with the multi-provider system, you need to ensure it works correctly with consensus strategies:

### Test Multi-Provider Integration

Create tests to verify that the new provider works correctly with multi-provider strategies:

```go
func TestMultiProviderWithNewProvider(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()
	
	// Create the new provider (or mock it)
	newProvider := provider.NewNewProvider("test-key", "test-model")
	
	// Create provider weights
	providers := []provider.ProviderWeight{
		{Provider: mockProvider, Weight: 1.0, Name: "mock"},
		{Provider: newProvider, Weight: 1.0, Name: "new"},
	}
	
	// Create multi-provider with fastest strategy
	multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)
	
	// Test generation
	response, err := multiProvider.Generate(context.Background(), "Test prompt")
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
	
	// Test other strategies
	// ...
}
```

## Step 6: Create Example Applications

Create an example application to showcase the new provider:

```bash
mkdir -p cmd/examples/newprovider
touch cmd/examples/newprovider/main.go
touch cmd/examples/newprovider/main_test.go
touch cmd/examples/newprovider/README.md
```

### Implement the Example Application

Implement the example application in `cmd/examples/newprovider/main.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	structuredProcessor "github.com/lexlapax/go-llms/pkg/structured/processor"
)

func main() {
	// Check if API key is provided
	apiKey := os.Getenv("NEW_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("No NEW_PROVIDER_API_KEY environment variable found. Using mock provider instead.")
		runWithMockProvider()
		return
	}

	// Create the new provider
	newProvider := provider.NewNewProvider(
		apiKey,
		"default-model",
	)

	// Create structured processor components
	validator := validation.NewValidator()
	processor := structuredProcessor.NewStructuredProcessor(validator)
	promptEnhancer := structuredProcessor.NewPromptEnhancer()

	// Create a schema for examples
	exampleSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:        "string",
				Description: "Example name",
			},
			"description": {
				Type:        "string",
				Description: "Example description",
			},
			"examples": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				Description: "List of examples",
			},
		},
		Required: []string{"name", "description", "examples"},
	}

	fmt.Println("Go-LLMs New Provider Example")
	fmt.Println("============================")

	// Example 1: Simple generation
	fmt.Println("\nExample 1: Simple generation")
	response, err := newProvider.Generate(
		context.Background(),
		"Explain what Go channels are in a paragraph",
	)
	if err != nil {
		log.Fatalf("Generation error: %v", err)
	}
	fmt.Printf("Response: %s\n", response)

	// Example 2: Using message-based conversation
	fmt.Println("\nExample 2: Message-based conversation")
	messages := []domain.Message{
		{Role: domain.RoleSystem, Content: "You are a helpful coding assistant specializing in Go."},
		{Role: domain.RoleUser, Content: "What's the difference between a slice and an array in Go?"},
	}
	messageResponse, err := newProvider.GenerateMessage(context.Background(), messages)
	if err != nil {
		log.Fatalf("Message generation error: %v", err)
	}
	fmt.Printf("Response: %s\n", messageResponse.Content)

	// Example 3: Structured output with schema
	fmt.Println("\nExample 3: Structured output generation with schema")

	// Enhance the prompt with schema information
	prompt := "Generate three examples of Go best practices"
	enhancedPrompt, err := promptEnhancer.Enhance(prompt, exampleSchema)
	if err != nil {
		log.Fatalf("Prompt enhancement error: %v", err)
	}

	// Generate the structured output
	structuredResponse, err := newProvider.Generate(context.Background(), enhancedPrompt)
	if err != nil {
		log.Fatalf("Structured generation error: %v", err)
	}

	// Process the raw response
	exampleData, err := processor.Process(exampleSchema, structuredResponse)
	if err != nil {
		log.Fatalf("Processing error: %v", err)
	}

	// Pretty print the result
	resultJSON, _ := json.MarshalIndent(exampleData, "", "  ")
	fmt.Printf("Structured Response:\n%s\n", string(resultJSON))

	// Example 4: Stream the response
	fmt.Println("\nExample 4: Streaming response")
	stream, err := newProvider.Stream(
		context.Background(),
		"List 3 benefits of Go's garbage collector in short bullet points",
	)
	if err != nil {
		log.Fatalf("Stream error: %v", err)
	}

	fmt.Println("Streamed Response:")
	for token := range stream {
		fmt.Print(token.Text)
		if token.Finished {
			fmt.Println()
		}
	}
}

// runWithMockProvider runs the example with a mock provider
func runWithMockProvider() {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()

	// Set a custom response for generation
	mockProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		return "This is a mock response for prompt: " + prompt, nil
	})

	fmt.Println("Go-LLMs Mock Example (simulating New Provider)")
	fmt.Println("=============================================")

	fmt.Println("\nExample 1: Simple generation")
	response, _ := mockProvider.Generate(
		context.Background(),
		"Explain what Go channels are in a paragraph",
	)
	fmt.Printf("Response: %s\n", response)

	// See the rest of the examples in the main function...
}

// Helper function for creating float pointers
func float64Ptr(v float64) *float64 {
	return &v
}
```

### Create a Test File

Create a test file for the example application in `cmd/examples/newprovider/main_test.go`:

```go
package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestNewProviderExample(t *testing.T) {
	// Skip test if no API key provided
	apiKey := os.Getenv("NEW_PROVIDER_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: NEW_PROVIDER_API_KEY environment variable not set")
	}

	// Create provider
	newProvider := provider.NewNewProvider(apiKey, "default-model")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test simple text generation
	response, err := newProvider.Generate(ctx, "Say hello!")
	if err != nil {
		t.Fatalf("Error generating text: %v", err)
	}

	if response == "" {
		t.Error("Empty response received from API")
	} else {
		t.Logf("Received response: %s", response)
	}
}
```

### Create a README for the Example

Create a README for the example application in `cmd/examples/newprovider/README.md`:

```markdown
# New Provider Example

This example demonstrates how to use the New Provider with Go-LLMs to interact with the service.

## Features

- Text generation with prompt
- Message-based conversation
- Structured data generation with schema validation
- Streaming responses

## Prerequisites

- Go 1.21 or later
- A New Provider API key

## Running the Example

1. Set your New Provider API key as an environment variable:

```bash
export NEW_PROVIDER_API_KEY=your_api_key_here
```

2. Build and run the example:

```bash
make example EXAMPLE=newprovider
./bin/newprovider
```

## What This Example Demonstrates

1. **Simple Text Generation**: Generate text with a single prompt
2. **Conversation**: Create a conversation with multiple messages 
3. **Structured Output**: Generate structured data with schema validation
4. **Streaming**: Stream tokens as they're generated

## Output

The example produces outputs from the New Provider for different types of prompts and shows how to use various features of the Go-LLMs library.

## Notes

- By default, this example uses the "default-model" model
- You can customize the model by setting the NEW_PROVIDER_MODEL environment variable
- If no API key is provided, the example falls back to a mock provider
```

## Step 7: Update Command Line Interface

Update the command line interface in `cmd/main.go` to support the new provider:

### Update Provider Flag Description

Update the provider flag description to include the new provider:

```go
rootCmd.PersistentFlags().String("provider", "openai", "LLM provider to use (openai, anthropic, gemini, newprovider, mock)")
```

### Add Default Model in Config

Add the default model for the new provider:

```go
// Set default provider models
viper.SetDefault("providers.openai.default_model", "gpt-4o")
viper.SetDefault("providers.anthropic.default_model", "claude-3-5-sonnet-latest")
viper.SetDefault("providers.gemini.default_model", "gemini-2.0-flash-lite")
viper.SetDefault("providers.newprovider.default_model", "default-model")
```

### Update Provider Creation in Commands

Update the provider creation in each command to include the new provider:

```go
switch providerType {
case "openai":
    llmProvider = provider.NewOpenAIProvider(apiKey, modelName)
case "anthropic":
    llmProvider = provider.NewAnthropicProvider(apiKey, modelName)
case "gemini":
    llmProvider = provider.NewGeminiProvider(apiKey, modelName)
case "newprovider":
    llmProvider = provider.NewNewProvider(apiKey, modelName)
case "mock":
    llmProvider = provider.NewMockProvider()
default:
    fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", providerType)
    os.Exit(1)
}
```

## Step 8: Update Documentation

### Update Main README

Update the main README.md to include the new provider:

1. Update the "Features" section:
   ```markdown
   - **Multiple providers**: Support for OpenAI, Anthropic, Google Gemini, New Provider, and extensible for other providers
   ```

2. Update the "Project Goals" section:
   ```markdown
   3. Support modern LLM providers (OpenAI, Anthropic, Google Gemini, New Provider, etc.)
   ```

3. Update the directory structure:
   ```markdown
   │   │   ├── provider/          # Provider implementations (OpenAI, Anthropic, Gemini, New Provider)
   ```

4. Update the code examples to include the new provider:
   ```go
   // Create multiple providers
   openaiProvider := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
   anthropicProvider := provider.NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-5-sonnet-latest")
   geminiProvider := provider.NewGeminiProvider(os.Getenv("GEMINI_API_KEY"), "gemini-2.0-flash-lite")
   newProvider := provider.NewNewProvider(os.Getenv("NEW_PROVIDER_API_KEY"), "default-model")
   ```

5. Update the "Example Applications" section:
   ```markdown
   - [New Provider Example](cmd/examples/newprovider/README.md) - Integration with New Provider
   ```

### Update API Documentation

Update the LLM API documentation in `docs/api/llm.md` to include the new provider.

### Update Technical Documentation

Update the [Architecture](docs/technical/architecture.md) document to mention the new provider in the appropriate sections.

## Common Challenges and Solutions

### Rate Limiting

Most LLM providers have rate limits. Implement appropriate backoff and retry mechanisms:

```go
func (p *NewProvider) doWithRetry(req *http.Request) (*http.Response, error) {
	// Implement exponential backoff for rate limits
	// ...
}
```

### Error Handling

Ensure all provider-specific errors are correctly mapped to the standard error types defined in `domain/errors.go`:

```go
// Domain error mapping
if strings.Contains(errMsg, "rate limit exceeded") {
    return domain.ErrRateLimitExceeded
}
```

### Stream Handling

Streaming responses can be complex. Implement proper connection and stream management:

```go
func (p *NewProvider) handleStreamResponse(ctx context.Context, resp *http.Response) (domain.ResponseStream, error) {
	// Handle SSE or other streaming formats
	// ...
}
```

## Conclusion

Adding a new provider to Go-LLMs involves implementing the provider interface, writing tests, updating utility functions, and updating documentation. By following this guide, you should be able to successfully integrate any LLM provider into the Go-LLMs library.

Remember to thoroughly test your implementation, both with unit tests and integration tests, to ensure compatibility with the rest of the library.

## References

- [Provider Interface Definition](/pkg/llm/domain/interfaces.go)
- [Domain Error Types](/pkg/llm/domain/errors.go)
- [OpenAI Provider Implementation](/pkg/llm/provider/openai.go) (reference implementation)
- [Anthropic Provider Implementation](/pkg/llm/provider/anthropic.go) (reference implementation)
- [Gemini Provider Implementation](/pkg/llm/provider/gemini.go) (reference implementation)