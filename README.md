# Go-LLMs: A Go Implementation of Pydantic-AI

Go-LLMs is a Go library for creating LLM-powered applications with structured outputs and type safety. It aims to port the core functionality of pydantic-ai to Go while embracing Go's idioms and strengths.

## Features

- **Structured responses**: Validates LLM outputs against predefined schemas
- **Model-agnostic interface**: Provides a unified API across different LLM providers
- **Type safety**: Leverages Go's type system for better developer experience
- **Dependency injection**: Enables passing data and services into agents
- **Tool integration**: Allows LLMs to interact with external systems through function calls
- **Multiple providers**: Support for OpenAI, Anthropic, and extensible for other providers
- **Schema validation**: Comprehensive JSON schema validation with type coercion
- **Monitoring hooks**: Hooks for logging, metrics, and debugging
- **Multi-provider strategies**: Combine providers using fastest, primary, or consensus approaches

## Project Goals

1. Create an idiomatic Go implementation of pydantic-ai
2. Minimize external dependencies by leveraging Go's standard library
3. Support modern LLM providers (OpenAI, Anthropic, etc.)
4. Provide comprehensive validation for LLM outputs
5. Follow clean architecture principles with vertical feature slices

## Project Structure

The project follows a vertical slicing approach where code is organized by feature:

```
go-llms/
├── cmd/                       # Application entry points
│   └── examples/              # Example applications
├── internal/                  # Internal packages
├── pkg/                       # Public packages
│   ├── schema/                # Schema definition and validation
│   │   ├── domain/            # Core domain models and interfaces
│   │   ├── validation/        # Validation implementation
│   │   └── adapter/           # Schema generation from Go structs
│   ├── llm/                   # LLM integration
│   │   ├── domain/            # Core domain models and interfaces
│   │   ├── provider/          # Provider implementations (OpenAI, Anthropic)
│   │   └── prompt/            # Prompt templates and formatting
│   ├── structured/            # Structured output processing
│   │   ├── domain/            # Core domain models and interfaces
│   │   ├── processor/         # JSON extraction and validation
│   │   └── adapter/           # External format adapters
│   └── agent/                 # Agent orchestration
│       ├── domain/            # Core domain models and interfaces
│       ├── tools/             # Tool implementations
│       └── workflow/          # Agent execution flow
└── examples/                  # Usage examples
```

## Installation

```bash
go get github.com/lexlapax/go-llms
```

## Basic Usage

### Schema Validation

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

### Using LLM Providers

```go
// Create an OpenAI provider with the modern gpt-4o model
provider := provider.NewOpenAIProvider(
    os.Getenv("OPENAI_API_KEY"),
    "gpt-4o",
)

// Generate text with a simple prompt
response, err := provider.Generate(context.Background(), "Explain quantum computing")
if err != nil {
    log.Fatalf("Generation error: %v", err)
}

fmt.Printf("Response: %s\n", response)

// Generate structured output with schema validation
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

### Processing Structured Outputs

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

### Multi-Provider Strategies

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

### Using Agents with Tools

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

### Prompt Enhancement

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

See the `cmd/examples/` directory for more comprehensive examples.

## Documentation

For detailed API documentation, see the [API Reference](docs/API_REFERENCE.md).

## Development Status

The core functionality is complete and working. Current focus is on:

1. Enhanced API documentation
2. Additional examples
3. Performance optimization
4. Error handling standardization

## License

This project is licensed under the MIT License - see the LICENSE file for details.