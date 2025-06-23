# Documentation Style Guide for Contributors

This guide outlines the documentation standards and conventions for the go-llms project. Following these guidelines ensures consistency and quality across all code documentation.

## Table of Contents

- [Overview](#overview)
- [ABOUTME Comments](#aboutme-comments)
- [Package Documentation](#package-documentation)
- [Function and Method Documentation](#function-and-method-documentation)
- [Type and Interface Documentation](#type-and-interface-documentation)
- [Examples and Usage](#examples-and-usage)
- [Common Patterns](#common-patterns)
- [Tools and Validation](#tools-and-validation)

## Overview

All Go code in this project must include comprehensive documentation that follows Go's documentation conventions. Documentation serves multiple purposes:

- Provides clear API documentation via `go doc`
- Enables IDE tooltips and auto-completion
- Helps new contributors understand code purpose and usage
- Ensures maintainability and knowledge transfer

## ABOUTME Comments

Every `.go` file must include ABOUTME comments that provide a quick summary of the file's purpose.

### Format

```go
// ABOUTME: Brief description of what this file does (one line)
// ABOUTME: Additional context or key functionality (second line)
```

### Rules

1. **Exactly 2 lines** starting with `// ABOUTME: `
2. **First line** describes the main purpose/functionality
3. **Second line** provides additional context, key features, or important details
4. **Keep concise** - each line should be under 80 characters
5. **Placed after package declaration** and before imports

### Examples

```go
// ABOUTME: Core JSON schema validator with type validation and constraints
// ABOUTME: Features object pooling, regex caching, and optional type coercion

package validation
```

```go
// ABOUTME: OpenAI provider implementation with streaming and function calling
// ABOUTME: Supports all GPT models, embeddings, and chat completions API
```

## Package Documentation

Each package must have comprehensive package-level documentation in at least one file (typically the main file or `doc.go`).

### Format

```go
// Package name provides brief description of the package purpose.
// 
// Detailed description explaining the package's role, key features,
// and primary use cases. Include architectural notes if relevant.
//
// # Key Features
//
// - Feature 1: Description
// - Feature 2: Description
// - Feature 3: Description
//
// # Usage Example
//
//   provider := provider.NewOpenAIProvider(apiKey, model)
//   result, err := provider.Generate(ctx, prompt)
//   if err != nil {
//       // handle error
//   }
//
// # Architecture Notes
//
// Additional context about design decisions, patterns used,
// or integration points with other packages.
package name
```

### Rules

1. **Start with package name** followed by purpose
2. **Include blank line** after first sentence
3. **Use markdown headers** (# ## ###) for sections
4. **Provide usage examples** where appropriate
5. **Document key types** and their relationships
6. **Explain architectural decisions** when relevant

### Examples

```go
// Package provider implements LLM provider interfaces and implementations.
//
// This package provides a unified interface for interacting with various
// LLM providers including OpenAI, Anthropic, Google, and others. It handles
// authentication, request formatting, response parsing, and error handling
// for each provider while maintaining a consistent API.
//
// # Supported Providers
//
// - OpenAI: GPT models, embeddings, function calling
// - Anthropic: Claude models with streaming support  
// - Google: Gemini models via Vertex AI and direct API
// - Ollama: Local model inference
// - OpenRouter: Access to multiple models via single API
//
// # Usage Example
//
//   provider := provider.NewOpenAIProvider(apiKey, "gpt-4")
//   result, err := provider.Generate(ctx, prompt)
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println(result)
package provider
```

## Function and Method Documentation

All exported functions and methods must have comprehensive documentation.

### Format

```go
// FunctionName performs specific action with given parameters.
// Detailed description of what the function does, including any
// important behavior, side effects, or assumptions.
//
// Parameters explain what each parameter is for and any constraints.
// Return values describe what is returned and under what conditions.
//
// Returns the result and any error that occurred during processing.
func FunctionName(param1 Type1, param2 Type2) (ResultType, error) {
    // implementation
}
```

### Rules

1. **Start with function name** and brief description
2. **Use present tense** ("performs", "creates", "validates")
3. **Document all parameters** with their purpose and constraints
4. **Document return values** and error conditions
5. **Include usage examples** for complex functions
6. **Mention side effects** or important behavior
7. **Keep first line under 80 characters**

### Examples

```go
// Generate creates a completion response from the given prompt using the configured model.
// This method handles request formatting, authentication, and response parsing
// for the OpenAI API. It supports both streaming and non-streaming responses
// based on the provided options.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - prompt: The input text to generate a completion for
//   - options: Optional settings like temperature, max tokens, etc.
//
// Returns the generated text response or an error if the request failed.
// Common errors include authentication failures, rate limiting, and network issues.
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
```

```go
// ValidateSchema checks if the provided data conforms to the JSON schema.
// Validation includes type checking, constraint validation, and format
// verification. The method uses caching for improved performance on
// repeated validations of the same schema.
//
// Parameters:
//   - schema: The JSON schema to validate against
//   - data: JSON string representation of the data to validate
//
// Returns validation results with detailed error messages if validation fails.
func (v *Validator) ValidateSchema(schema *Schema, data string) (*ValidationResult, error) {
```

## Type and Interface Documentation

All exported types and interfaces require clear documentation.

### Structs

```go
// ProviderConfig represents configuration settings for LLM providers.
// It contains authentication details, model selection, and optional
// provider-specific settings used during provider initialization.
type ProviderConfig struct {
    // Provider identifies the LLM provider (e.g., "openai", "anthropic")
    Provider string
    
    // Model specifies the model name to use (e.g., "gpt-4", "claude-3")
    Model string
    
    // APIKey contains the authentication key for the provider
    APIKey string
    
    // BaseURL optionally overrides the default provider endpoint
    BaseURL string
}
```

### Interfaces

```go
// Provider defines the interface for LLM provider implementations.
// All providers must implement these methods to ensure consistent
// behavior across different LLM services. The interface supports
// both simple text generation and advanced features like streaming.
type Provider interface {
    // Generate creates a text completion from the given prompt.
    // Returns the generated text or an error if generation failed.
    Generate(ctx context.Context, prompt string, options ...Option) (string, error)
    
    // Stream generates a streaming response for real-time text generation.
    // Returns a channel of partial responses and an error channel.
    Stream(ctx context.Context, prompt string, options ...Option) (<-chan string, <-chan error)
    
    // Name returns the provider's identifier (e.g., "openai", "anthropic").
    Name() string
}
```

### Rules

1. **Document the purpose** and role of the type
2. **Explain key fields** and their constraints
3. **Describe relationships** with other types
4. **Include usage context** when helpful
5. **Document interface contracts** clearly
6. **Mention implementation requirements** for interfaces

## Examples and Usage

Include practical examples in documentation when they add value.

### When to Include Examples

- Complex APIs with multiple parameters
- Non-obvious usage patterns
- Common integration scenarios
- Error handling patterns

### Example Format

```go
// ProcessWithRetry attempts processing with automatic retries on failure.
// It implements exponential backoff between retry attempts and respects
// context cancellation for graceful shutdowns.
//
// Example usage:
//
//   result, err := ProcessWithRetry(ctx, func() (interface{}, error) {
//       return client.MakeRequest(data)
//   }, 3, time.Second)
//   if err != nil {
//       log.Printf("Processing failed after retries: %v", err)
//       return
//   }
//   fmt.Printf("Result: %v\n", result)
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - fn: Function to execute with retry logic
//   - maxRetries: Maximum number of retry attempts
//   - baseDelay: Initial delay between retries (doubles each attempt)
//
// Returns the function result or the last error encountered.
func ProcessWithRetry(ctx context.Context, fn func() (interface{}, error), maxRetries int, baseDelay time.Duration) (interface{}, error) {
```

## Common Patterns

### Error Documentation

Always document error conditions clearly:

```go
// Connect establishes a connection to the LLM provider.
// Returns an error if authentication fails, the provider is unavailable,
// or network connectivity issues prevent connection establishment.
func (p *Provider) Connect() error {
```

### Context Parameters

Document context usage consistently:

```go
// Generate creates a completion using the provided context for cancellation.
// The context can be used to set timeouts, cancel in-flight requests,
// or pass request-scoped values like trace IDs.
func (p *Provider) Generate(ctx context.Context, prompt string) (string, error) {
```

### Option Parameters

Document variadic options clearly:

```go
// CreateProvider initializes a new provider with the given configuration.
// Additional options can be provided to customize behavior such as
// timeout settings, retry policies, or debug logging.
//
// Example:
//   provider := CreateProvider(config, 
//       WithTimeout(30*time.Second),
//       WithRetryPolicy(3, time.Second))
func CreateProvider(config Config, options ...Option) *Provider {
```

## Tools and Validation

### Checking Documentation

Use these commands to validate documentation:

```bash
# Check that all exported items have documentation
go doc -all ./pkg/provider

# Generate documentation for review
godoc -http=:6060

# Lint documentation quality
golangci-lint run --enable=godox,godot
```

### Common Issues to Avoid

1. **Missing documentation** for exported items
2. **Inconsistent ABOUTME format** (wrong number of lines, missing prefix)
3. **Unclear parameter descriptions** 
4. **Missing error documentation**
5. **Outdated examples** that don't match current API
6. **Too verbose** documentation that obscures key information
7. **Too brief** documentation that lacks essential details

### Documentation Checklist

Before submitting code, verify:

- [ ] All `.go` files have properly formatted ABOUTME comments
- [ ] Package has comprehensive documentation in at least one file
- [ ] All exported functions/methods are documented
- [ ] All exported types and interfaces are documented
- [ ] Parameter and return value documentation is clear
- [ ] Error conditions are documented
- [ ] Examples are included for complex APIs
- [ ] Documentation follows Go conventions (present tense, etc.)
- [ ] No spelling or grammar errors

## Conclusion

Good documentation is essential for maintainable, usable code. These guidelines ensure consistency across the project and help contributors write clear, helpful documentation. When in doubt, err on the side of more detail rather than less, and always consider the perspective of someone encountering the code for the first time.

For questions about documentation standards or clarification on these guidelines, please refer to existing code in the project or ask in pull request reviews.