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
// Create an OpenAI provider
provider := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))

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
// Create a structured processor
processor := processor.NewStructuredProcessor(validator)

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

### Using Agents with Tools

```go
// Create a provider
llmProvider := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))

// Create an agent
agent := workflow.NewAgent(llmProvider)

// Add tools
agent.AddTool(tools.NewCalculatorTool())
agent.AddTool(tools.NewWebSearchTool())

// Set system prompt
agent.SetSystemPrompt("You are a helpful assistant that can search the web and perform calculations.")

// Add a logging hook
agent.WithHook(hooks.NewLoggingHook())

// Run the agent
result, err := agent.Run(context.Background(), "What is the square root of 144, and can you find information about the Fibonacci sequence?")
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

## Development Status

This project is currently in active development. The API is unstable and subject to change.

## License

This project is licensed under the MIT License - see the LICENSE file for details.