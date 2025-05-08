# Go-LLMs API Documentation

> **[Documentation Home](/REFERENCE.md) / API Documentation**

This documentation provides detailed information about the Go-LLMs public API, organized by packages and features.

*Related: [User Guide](/docs/user-guide/) | [Technical Documentation](/docs/technical/)*

## Core Features

- **Schema Validation**: Define, validate, and coerce structured outputs against JSON Schema
- **LLM Provider Integration**: Unified API for OpenAI, Anthropic, and custom providers
- **Multi-Provider Strategies**: Combine multiple LLM providers with different strategies (fastest, primary, consensus)
- **Structured Output Processing**: Generate structured, schema-conforming responses from LLMs
- **Agent System**: Build LLM agents with tools and workflow capabilities
- **Monitoring Hooks**: Track and measure LLM operations with logging and metrics hooks
- **Type-Safe API**: Generics-based design for type safety and IDE support
- **Convenience Utilities**: Helper functions for common LLM operations and patterns

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
│   └── util/                  # Utility packages
│       ├── json/              # JSON utilities 
│       └── llmutil/           # LLM convenience functions
```

## Packages

- [schema](schema.md) - Schema definition and validation
- [llm](llm.md) - LLM provider integration
- [structured](structured.md) - Structured output processing
- [agent](agent.md) - Agent and tool functionality

## Getting Started

For most applications, you'll need to use multiple packages together. Here's a typical usage flow:

1. Define a schema for the structured output you want to receive
2. Create an LLM provider (e.g., OpenAI, Anthropic)
3. Use the structured output processor to generate and validate responses
4. Optionally, use the agent system to add tools and workflow capabilities

## Basic Usage

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

## Examples

Check each package's documentation for detailed examples and guides:

- [schema](schema.md) - Schema definition and validation
- [llm](llm.md) - LLM provider integration
- [structured](structured.md) - Structured output processing
- [agent](agent.md) - Agent and tool functionality

For more examples, see the [examples directory](/cmd/examples/) in the project repository.