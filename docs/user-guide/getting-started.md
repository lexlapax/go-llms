# Getting Started with Go-LLMs

> **[Documentation Home](/REFERENCE.md) / [User Guide](/docs/user-guide/) / Getting Started**

This guide will help you get started with the Go-LLMs library, a Go implementation for creating LLM-powered applications with structured outputs and type safety.

*Related: [Multi-Provider Guide](multi-provider.md) | [Advanced Validation](advanced-validation.md) | [Error Handling](error-handling.md) | [API Reference](/docs/api/README.md)*

## Table of Contents

1. [Installation](#installation)
2. [Basic Usage](#basic-usage)
3. [Schema Validation](#schema-validation)
4. [LLM Providers](#llm-providers)
5. [Structured Output](#structured-output)
6. [Multi-Provider](#multi-provider)
7. [Agent Tools](#agent-tools)
8. [Next Steps](#next-steps)

## Installation

Install the Go-LLMs library using Go modules:

```bash
go get github.com/lexlapax/go-llms
```

## Basic Usage

### Simple Text Generation

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Create an OpenAI provider with the modern gpt-4o model
    llmProvider := provider.NewOpenAIProvider(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-4o",
    )
    
    // Generate text with a simple prompt
    response, err := llmProvider.Generate(context.Background(), "Explain quantum computing")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Response: %s\n", response)
}
```

## Schema Validation

Go-LLMs provides comprehensive schema validation capabilities:

```go
// Define a schema for validation
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "name":  {Type: "string", Description: "Person's name"},
        "age":   {Type: "integer", Minimum: float64Ptr(0), Maximum: float64Ptr(120)},
        "email": {Type: "string", Format: "email", Description: "Email address"},
    },
    Required: []string{"name", "email"},
}

// Create a validator
validator := validation.NewValidator()

// Validate JSON data against the schema
result, err := validator.Validate(schema, `{"name": "John Doe", "age": 30, "email": "john@example.com"}`)
if err != nil {
    log.Fatalf("Validation error: %v", err)
}

fmt.Printf("Validation result: %v\n", result.Valid)
```

## LLM Providers

Go-LLMs supports multiple LLM providers:

### OpenAI

```go
// Create an OpenAI provider
provider := provider.NewOpenAIProvider(
    os.Getenv("OPENAI_API_KEY"),
    "gpt-4o",
)
```

### Anthropic

```go
// Create an Anthropic provider
provider := provider.NewAnthropicProvider(
    os.Getenv("ANTHROPIC_API_KEY"),
    "claude-3-5-sonnet-latest",
)
```

### Mock Provider (for testing)

```go
// Create a mock provider for testing
provider := provider.NewMockProvider().
    WithResponse("This is a mock response").
    WithDelay(100 * time.Millisecond)
```

## Structured Output

Generate structured, schema-conforming outputs from LLMs:

```go
// Define your schema
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "name": {Type: "string"},
        "age":  {Type: "integer", Minimum: float64Ptr(0)},
        "email": {Type: "string", Format: "email"},
    },
    Required: []string{"name", "email"},
}

// Generate structured output
result, err := provider.GenerateWithSchema(
    context.Background(),
    "Generate information about a person",
    schema,
)
if err != nil {
    log.Fatalf("Structured generation error: %v", err)
}

// Access the structured result
person := result.(map[string]interface{})
fmt.Printf("Generated person: %s (%d)\n", person["name"], person["age"])
```

### Processing Raw Outputs

You can also process raw LLM outputs containing JSON:

```go
// Create a JSON processor
processor := processor.NewJsonProcessor()

// Process raw LLM output containing JSON
rawOutput := `I'll create a person profile for you:
{
  "name": "Jane Smith",
  "age": 35,
  "email": "jane.smith@example.com"
}
Hope this helps!`

// Extract and validate the JSON
data, err := processor.Process(schema, rawOutput)
if err != nil {
    log.Fatalf("Processing error: %v", err)
}

// Or map directly to a struct
type Person struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

var person Person
err = processor.ProcessTyped(schema, rawOutput, &person)
if err != nil {
    log.Fatalf("Processing error: %v", err)
}

fmt.Printf("Person: %s (%d)\n", person.Name, person.Age)
```

## Multi-Provider

The Multi-Provider feature allows you to work with multiple LLM providers simultaneously:

```go
// Create multiple providers
openaiProvider := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
anthropicProvider := provider.NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-5-sonnet-latest")

// Create provider weights
providers := []provider.ProviderWeight{
    {Provider: openaiProvider, Weight: 1.0, Name: "openai"},
    {Provider: anthropicProvider, Weight: 1.0, Name: "anthropic"},
}

// Create a multi-provider with the fastest strategy
fastestProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

// Or with the primary strategy (with fallback)
primaryProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).
    WithPrimaryProvider(0) // Use first provider as primary

// Or with consensus strategy
consensusProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
    WithConsensusStrategy(provider.ConsensusSimilarity).
    WithSimilarityThreshold(0.7)

// Use like any other provider
response, err := consensusProvider.Generate(context.Background(), "What are the three laws of robotics?")
```

For more details on the Multi-Provider feature, see the [Multi-Provider Guide](multi-provider.md).

## Agent Tools

The Agent feature allows LLMs to interact with tools and perform complex tasks:

```go
// Create a provider
llmProvider := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4o")

// Create an agent with string output type
agent := workflow.NewAgent[struct{}, string](llmProvider).
    SetSystemPrompt("You are a helpful assistant with access to tools.")

// Add a calculator tool
agent.AddTool(tools.NewTool(
    "calculator",
    "Perform mathematical calculations",
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        expression, _ := params["expression"].(string)
        // Implement calculation...
        return result, nil
    },
))

// Add a logging hook
agent.AddHook(workflow.NewLoggingHook(slog.Default(), workflow.LogLevelDetailed))

// Run the agent
result, err := agent.Run(context.Background(), "What is the square root of 144?", struct{}{})
if err != nil {
    log.Fatalf("Agent error: %v", err)
}

fmt.Printf("Agent result: %v\n", result)
```

## Prompt Enhancement

Enhance prompts with schema information:

```go
// Create a prompt enhancer
enhancer := processor.NewPromptEnhancer()

// Enhance a prompt with schema information
prompt := "Generate information about a person"
enhancedPrompt, err := enhancer.Enhance(prompt, schema)
if err != nil {
    log.Fatalf("Enhancement error: %v", err)
}

// Use the enhanced prompt with your LLM provider
response, err := provider.Generate(context.Background(), enhancedPrompt)
```

## Convenience Utilities

The library includes various convenience utilities:

```go
// Create a provider from config
config := llmutil.ModelConfig{
    Provider: "openai",
    Model:    "gpt-4o",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
}
provider, err := llmutil.CreateProvider(config)

// Generate responses in parallel
prompts := []string{
    "What is the capital of France?",
    "Give me a recipe for pancakes",
    "How many planets are in our solar system?",
}
results, errors := llmutil.BatchGenerate(context.Background(), provider, prompts)

// Generate with retry for transient errors
result, err := llmutil.GenerateWithRetry(
    context.Background(), 
    provider, 
    "Write a haiku about programming",
    3, // max retries
)

// Create a provider pool for load balancing
providerPool := llmutil.NewProviderPool(
    []domain.Provider{provider1, provider2, provider3},
    llmutil.StrategyRoundRobin,
)

// Create an agent with common configuration
agentConfig := llmutil.AgentConfig{
    Provider:      provider,
    SystemPrompt:  "You are a helpful assistant with access to tools.",
    EnableCaching: true,
    Tools:         []agentDomain.Tool{calculatorTool},
    Hooks:         []agentDomain.Hook{workflow.NewMetricsHook()},
}
agent := llmutil.CreateAgent(agentConfig)

// Run an agent with timeout
result, err := llmutil.RunWithTimeout(
    agent,
    "What is 7 * 6?",
    10*time.Second, // timeout
)
```

## Next Steps

Now that you're familiar with the basics, you can:

1. Explore the [API Reference](/docs/api/) for detailed information about each component
2. Review the [examples](/cmd/examples/) for more comprehensive examples
3. Learn about [advanced validation features](advanced-validation.md)
4. Understand [error handling patterns](error-handling.md)
5. Explore [performance optimization strategies](/docs/technical/performance.md)

Helper function for the examples above:

```go
func float64Ptr(v float64) *float64 {
    return &v
}
```