# Go-LLMs API Reference

This API reference provides a comprehensive guide to the Go-LLMs library, which offers a robust framework for working with Large Language Models (LLMs) in Go applications with structured outputs and type safety.

## Table of Contents

- [Core Features](#core-features)
- [Package Structure](#package-structure)
- [Getting Started](#getting-started)
- [Package Documentation](#package-documentation)
- [Examples](#examples)

## Core Features

- **Schema Validation**: Define, validate, and coerce structured outputs against JSON Schema
- **LLM Provider Integration**: Unified API for OpenAI, Anthropic, and custom providers
- **Multi-Provider Strategies**: Combine multiple LLM providers with different strategies (fastest, primary, consensus)
- **Structured Output Processing**: Generate structured, schema-conforming responses from LLMs
- **Agent System**: Build LLM agents with tools and workflow capabilities
- **Monitoring Hooks**: Track and measure LLM operations with logging and metrics hooks
- **Type-Safe API**: Generics-based design for type safety and IDE support

## Package Structure

Go-LLMs follows a vertical slicing approach where code is organized by feature:

```
go-llms/
├── pkg/                       # Public packages
│   ├── schema/                # Schema definition and validation feature
│   │   ├── domain/            # Core domain models
│   │   ├── validation/        # Validation logic
│   │   └── adapter/           # External adapters (reflection, etc.)
│   ├── llm/                   # LLM integration feature
│   │   ├── domain/            # Core LLM domain models
│   │   └── provider/          # LLM provider implementations
│   ├── structured/            # Structured output feature
│   │   ├── domain/            # Structured output domain
│   │   └── processor/         # Output processors
│   └── agent/                 # Agent feature (tools, workflows)
│       ├── domain/            # Agent domain models
│       ├── tools/             # Tool implementations
│       └── workflow/          # Agent workflows
```

## Getting Started

### Installation

```bash
go get github.com/lexlapax/go-llms
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

func main() {
    // Create an LLM provider
    llmProvider := provider.NewOpenAIProvider(
        "your-api-key",
        "gpt-4o",
    )
    
    // Define a schema for structured output
    schema := &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "name": {Type: "string"},
            "age": {Type: "integer", Minimum: float64Ptr(0)},
            "email": {Type: "string", Format: "email"},
        },
        Required: []string{"name", "email"},
    }
    
    // Create a prompt
    prompt := "Generate a profile for a fictional person."
    
    // Enhance the prompt with schema information
    enhancedPrompt, err := processor.EnhancePromptWithSchema(prompt, schema)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Generate a response
    response, err := llmProvider.Generate(context.Background(), enhancedPrompt)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Process the response
    proc := processor.NewJsonProcessor()
    result, err := proc.Process(schema, response)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Use the result
    person := result.(map[string]interface{})
    fmt.Printf("Name: %s\n", person["name"])
    fmt.Printf("Age: %v\n", person["age"])
    fmt.Printf("Email: %s\n", person["email"])
}

func float64Ptr(v float64) *float64 {
    return &v
}
```

## Package Documentation

Detailed documentation for each package is available in the API docs:

- [Schema Package](docs/api/schema.md) - Schema definition and validation
- [LLM Package](docs/api/llm.md) - LLM provider integration
- [Structured Package](docs/api/structured.md) - Structured output processing
- [Agent Package](docs/api/agent.md) - Agent and tool functionality

## Examples

The following examples demonstrate common usage patterns:

### Multi-Provider with Consensus Strategy

```go
// Create provider weights
providers := []provider.ProviderWeight{
    {Provider: openAIProvider, Weight: 1.0, Name: "openai"},
    {Provider: anthropicProvider, Weight: 1.0, Name: "anthropic"},
    {Provider: mockProvider, Weight: 0.5, Name: "mock"},
}

// Create a multi-provider with the consensus strategy
consensusProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
    WithConsensusStrategy(provider.ConsensusSimilarity).
    WithSimilarityThreshold(0.7)

// Generate a response
response, err := consensusProvider.Generate(
    context.Background(),
    "What are the three principles of object-oriented programming?",
)
```

### Agent with Tools

```go
// Create an agent
agent := workflow.NewAgent[struct{}, string](llmProvider).
    SetSystemPrompt("You are a helpful assistant with access to tools.")

// Add a calculator tool
agent.AddTool(tools.NewTool(
    "calculator",
    "Perform mathematical calculations",
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        expression, _ := params["expression"].(string)
        // Implement calculation logic...
        return result, nil
    },
))

// Run the agent
result, err := agent.Run(
    context.Background(),
    "What is the square root of 144 plus 36?",
    struct{}{},
)
```

### Typed Structured Output

```go
// Define a structured output type
type WeatherInfo struct {
    Location    string  `json:"location"`
    Temperature float64 `json:"temperature"`
    Condition   string  `json:"condition"`
    Humidity    int     `json:"humidity"`
}

// Create a schema
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "location": {Type: "string", Description: "The location name"},
        "temperature": {Type: "number", Description: "The temperature in Celsius"},
        "condition": {Type: "string", Description: "The weather condition"},
        "humidity": {Type: "integer", Description: "The humidity percentage"},
    },
    Required: []string{"location", "temperature", "condition"},
}

// Process output directly into a typed struct
var weather WeatherInfo
err := processor.ProcessTyped(schema, response, &weather)
```

For more examples, see the [examples directory](cmd/examples/).

## Additional Resources

- [Implementation Plan](implementation-plan.md) - Detailed project implementation plan
- [Coding Practices](coding-practices.md) - Coding standards and best practices
- [Python to Go Blueprint](pydantic-ai-to-go.md) - Blueprint for porting pydantic-ai to Go