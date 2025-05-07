# Step-by-Step Implementation Plan for go-llms: A Pydantic-ai Port to Go
This implementation plan outlines the specific steps needed to create a Go port of pydantic-ai (go-llms), following Go's idioms and best practices. The plan adopts a test-driven development approach with vertical feature slices, ensuring that each component is properly tested before integration.
## Phase 1: Project Setup and Architecture (1-2 weeks)

### Step 1: Initialize project structure

1. Create the initial directory structure:
```
go-llms/
├── cmd/                       # Application entry points
│   └── examples/              # Example applications
├── internal/                  # Internal packages
├── pkg/                       # Public packages
│   ├── schema/                # Schema definition and validation
│   ├── llm/                   # LLM integration
│   ├── structured/            # Structured output
│   └── agent/                 # Agent feature
├── examples/                  # Usage examples
└── tests/                     # Integration tests
```

2. Initialize Go module:
```bash
go mod init github.com/lexlapax/go-llms
```

3. Create initial README.md with project goals and architecture overview

### Step 2: Define core interfaces

1. Create domain interfaces for the schema validation system:
```go
// pkg/schema/domain/interfaces.go
package domain

// Schema represents a validation schema for structured data
type Schema struct {
    Type        string              `json:"type"`
    Properties  map[string]Property `json:"properties,omitempty"`
    Required    []string            `json:"required,omitempty"`
    // Other schema properties...
}

// Validator defines the contract for schema validation
type Validator interface {
    // Validate checks if data conforms to the schema
    Validate(schema *Schema, data string) (*ValidationResult, error)
    
    // ValidateStruct validates a Go struct against a schema
    ValidateStruct(schema *Schema, obj interface{}) (*ValidationResult, error)
}
```

2. Create domain interfaces for the LLM providers:
```go
// pkg/llm/domain/interfaces.go
package domain

import "context"

// Provider defines the contract for LLM providers
type Provider interface {
    // Generate produces text from a prompt
    Generate(ctx context.Context, prompt string, options ...Option) (string, error)
    
    // GenerateWithSchema produces structured output conforming to a schema
    GenerateWithSchema(ctx context.Context, prompt string, schema interface{}, options ...Option) (interface{}, error)
    
    // Stream streams responses token by token
    Stream(ctx context.Context, prompt string, options ...Option) (<-chan Token, error)
}
```

3. Create domain interfaces for agents and tools:
```go
// pkg/agent/domain/interfaces.go
package domain

import "context"

// Tool represents a capability the LLM can invoke
type Tool interface {
    // Name returns the tool's name
    Name() string
    
    // Description provides information about the tool
    Description() string
    
    // Execute runs the tool with parameters
    Execute(ctx context.Context, params interface{}) (interface{}, error)
    
    // ParameterSchema returns the schema for the tool parameters
    ParameterSchema() interface{}
}

// Agent coordinates interactions with LLMs
type Agent interface {
    // Run executes the agent with given inputs
    Run(ctx context.Context, input interface{}) (interface{}, error)
    
    // AddTool registers a tool with the agent
    AddTool(tool Tool) Agent
    
    // SetSystemPrompt configures the agent's system prompt
    SetSystemPrompt(prompt string) Agent
}
```

## Phase 2: Schema Validation Implementation (2-3 weeks)

### Step 3: Implement basic schema models

1. Create property types for the schema validation:
```go
// pkg/schema/domain/schema.go
package domain

// Property represents a property in a schema
type Property struct {
    Type        string      `json:"type"`
    Format      string      `json:"format,omitempty"`
    Description string      `json:"description,omitempty"`
    Minimum     *float64    `json:"minimum,omitempty"`
    Maximum     *float64    `json:"maximum,omitempty"`
    MinLength   *int        `json:"minLength,omitempty"`
    MaxLength   *int        `json:"maxLength,omitempty"`
    Pattern     string      `json:"pattern,omitempty"`
    Enum        []string    `json:"enum,omitempty"`
    Items       *Property   `json:"items,omitempty"`
    Required    bool        `json:"required,omitempty"`
}

// ValidationResult represents the outcome of a validation
type ValidationResult struct {
    Valid  bool     `json:"valid"`
    Errors []string `json:"errors,omitempty"`
}
```


### Step 4: Implement the validator using TDD

1. Write tests for different validation scenarios:
```go
// pkg/schema/validation/validator_test.go
package validation

import "testing"

func TestStringValidation(t *testing.T) {
    // Test cases for string validation
}

func TestNumberValidation(t *testing.T) {
    // Test cases for number validation
}

func TestObjectValidation(t *testing.T) {
    // Test cases for object validation
}
```

2. Implement the validator to pass tests:
```go
// pkg/schema/validation/validator.go
package validation

import (
    "encoding/json"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DefaultValidator implements schema validation
type DefaultValidator struct{}

// NewValidator creates a new validator
func NewValidator() *DefaultValidator {
    return &DefaultValidator{}
}

// Validate validates a JSON string against a schema
func (v *DefaultValidator) Validate(schema *domain.Schema, jsonStr string) (*domain.ValidationResult, error) {
    // Implementation following TDD approach
}
```


### Step 5: Implement schema generation from Go structs

1. Write tests for schema generation:
```go
// pkg/schema/adapter/reflection/schema_generator_test.go
package reflection

import "testing"

func TestGenerateSchemaFromStruct(t *testing.T) {
    // Test cases for generating schemas from structs
}
```

2. Implement schema generation:
```go
// pkg/schema/adapter/reflection/schema_generator.go
package reflection

import (
    "reflect"
    "strings"
    
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// GenerateSchema creates a schema from a Go struct
func GenerateSchema(obj interface{}) (*domain.Schema, error) {
    // Implementation using reflection
}
```

### Step 6: Implement type coercion

1. Create tests for type coercion:
```go
// pkg/schema/validation/coercion_test.go
package validation

import "testing"

func TestCoerceToString(t *testing.T) {
    // Test cases for string coercion
}

func TestCoerceToNumber(t *testing.T) {
    // Test cases for number coercion
}
```

2. Implement type coercion:
```go
// pkg/schema/validation/coercion.go
package validation

// Coerce attempts to convert a value to the expected type
func (v *DefaultValidator) Coerce(targetType string, value interface{}) (interface{}, bool) {
    // Implementation for different types
}
```


## Phase 3: LLM Provider Integration (2-3 weeks)
### Step 7: Implement base LLM provider interface

1. Create message and token models:
```go
// pkg/llm/domain/message.go
package domain

// Role represents the role of a message sender
type Role string

const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleTool      Role = "tool"
)

// Message represents a message in a conversation
type Message struct {
    Role    Role   `json:"role"`
    Content string `json:"content"`
}

// Token represents a token in a streamed response
type Token struct {
    Text     string `json:"text"`
    Finished bool   `json:"finished"`
}
```

2. Create provider options:
```go
// pkg/llm/domain/options.go
package domain

// Option configures LLM provider behavior
type Option func(*ProviderOptions)

// ProviderOptions stores configuration for LLM providers
type ProviderOptions struct {
    Temperature      float64
    MaxTokens        int
    StopSequences    []string
    TopP             float64
    FrequencyPenalty float64
    PresencePenalty  float64
}

// WithTemperature sets the temperature for generation
func WithTemperature(temp float64) Option {
    return func(o *ProviderOptions) {
        o.Temperature = temp
    }
}

// Additional option functions...
```


### Step 8: Implement OpenAI provider

1. Write tests for the OpenAI provider:
```go
// pkg/llm/provider/openai_test.go
package provider

import "testing"

func TestOpenAIGenerate(t *testing.T) {
    // Test cases for OpenAI generation
}
```

2. Implement the OpenAI provider:
```go
// pkg/llm/provider/openai.go
package provider

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
    apiKey     string
    model      string
    validator  domain.Validator
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
    return &OpenAIProvider{
        apiKey: apiKey,
        model:  model,
    }
}

// Generate sends a prompt to OpenAI and returns the response
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
    // Implementation
}

// GenerateWithSchema generates LLM output conforming to a schema
func (p *OpenAIProvider) GenerateWithSchema(ctx context.Context, prompt string, schema interface{}, options ...domain.Option) (interface{}, error) {
    // Implementation
}
```

### Step 9: Implement mock provider for testing

1. Create a mock provider for testing:
```go
// pkg/llm/provider/mock.go
package provider

import (
    "context"
    
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
    generateFunc func(ctx context.Context, prompt string, options ...domain.Option) (string, error)
    streamFunc   func(ctx context.Context, prompt string, options ...domain.Option) (<-chan domain.Token, error)
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
    return &MockProvider{
        generateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
            return `{"result": "This is a mock response"}`, nil
        },
        streamFunc: func(ctx context.Context, prompt string, options ...domain.Option) (<-chan domain.Token, error) {
            ch := make(chan domain.Token)
            go func() {
                defer close(ch)
                ch <- domain.Token{Text: "This is a mock response", Finished: true}
            }()
            return ch, nil
        },
    }
}

// Implementation of Provider interface...
```


### Step 10: Implement Anthropic provider

1. Write tests for the Anthropic provider:
```go
// pkg/llm/provider/anthropic_test.go
package provider

import "testing"

func TestAnthropicGenerate(t *testing.T) {
    // Test cases for Anthropic generation
}
```

2. Implement the Anthropic provider:
```go
// pkg/llm/provider/anthropic.go
package provider

import (
    "context"
    
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// AnthropicProvider implements the Provider interface for Anthropic
type AnthropicProvider struct {
    apiKey string
    model  string
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
    return &AnthropicProvider{
        apiKey: apiKey,
        model:  model,
    }
}

// Implementation of Provider interface...
```

## Phase 4: Structured Output Implementation (1-2 weeks)
### Step 11: Implement structured output processor

1. Write tests for structured output processing:
```go
// pkg/structured/processor/processor_test.go
package processor

import "testing"

func TestProcessStructuredOutput(t *testing.T) {
    // Test cases for structured output processing
}
```

2. Implement structured output processor:
```go
// pkg/structured/processor/processor.go
package processor

import (
    "encoding/json"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/schema/validation"
)

// StructuredProcessor handles processing of structured LLM outputs
type StructuredProcessor struct {
    validator domain.Validator
}

// NewStructuredProcessor creates a new structured processor
func NewStructuredProcessor() *StructuredProcessor {
    return &StructuredProcessor{
        validator: validation.NewValidator(),
    }
}

// Process processes a structured output against a schema
func (p *StructuredProcessor) Process(schema *domain.Schema, output string) (interface{}, error) {
    // Implementation
}
```

### Step 12: Implement prompt enhancement for structured outputs

1. Create tests for prompt enhancement:
```go
// pkg/structured/processor/prompt_test.go
package processor

import "testing"

func TestEnhancePromptWithSchema(t *testing.T) {
    // Test cases for prompt enhancement
}
```

1. Implement prompt enhancement:
```go
// pkg/structured/processor/prompt.go
package processor

import (
    "encoding/json"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// EnhancePromptWithSchema adds schema information to a prompt
func EnhancePromptWithSchema(prompt string, schema *domain.Schema) (string, error) {
    // Implementation
}
```

## Phase 5: Agent and Tool Implementation (2-3 weeks)
### Step 13: Implement tool system

1. Create tests for tool execution:
```go
// pkg/agent/tools/tool_test.go
package tools

import "testing"

func TestExecuteTool(t *testing.T) {
    // Test cases for tool execution
}
```

2. Implement basic tool system:
```go
// pkg/agent/tools/base_tool.go
package tools

import (
    "context"
    "fmt"
    "reflect"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/schema/adapter/reflection"
)

// BaseTool provides a foundation for tool implementations
type BaseTool struct {
    name        string
    description string
    fn          interface{}
    paramSchema interface{}
}

// NewTool creates a new tool from a function
func NewTool(name, description string, fn interface{}) domain.Tool {
    paramSchema := reflection.ExtractParameterSchema(fn)
    return &BaseTool{
        name:        name,
        description: description,
        fn:          fn,
        paramSchema: paramSchema,
    }
}

// Implementation of Tool interface...
```

### Step 14: Implement context for dependency injection

1. Create a run context for dependency injection:
```go
// pkg/agent/domain/context.go
package domain

import "context"

// RunContext carries dependencies for a run
type RunContext[D any] struct {
    ctx  context.Context
    deps D
}

// NewRunContext creates a new run context
func NewRunContext[D any](ctx context.Context, deps D) *RunContext[D] {
    return &RunContext[D]{
        ctx:  ctx,
        deps: deps,
    }
}

// Deps returns the dependencies
func (r *RunContext[D]) Deps() D {
    return r.deps
}

// Context returns the context
func (r *RunContext[D]) Context() context.Context {
    return r.ctx
}
```

### Step 15: Implement agent system

1. Write tests for the agent system:
```go
// pkg/agent/workflow/agent_test.go
package workflow

import "testing"

func TestAgentRun(t *testing.T) {
    // Test cases for agent execution
}
```

2. Implement agent system:
```go
// pkg/agent/workflow/agent.go
package workflow

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// Agent[D, O] is a generic agent that handles dependencies and outputs
type Agent[D any, O any] struct {
    llmProvider domain.Provider
    tools       map[string]domain.Tool
    systemPrompt string
    outputType   reflect.Type
}

// NewAgent creates a new agent
func NewAgent[D any, O any](provider domain.Provider) *Agent[D, O] {
    return &Agent[D, O]{
        llmProvider: provider,
        tools:       make(map[string]domain.Tool),
    }
}

// Implementation of Agent interface...
```

### Step 16: Implement hooks for monitoring

1. Create hook interfaces:
```go
// pkg/agent/domain/hooks.go
package domain

import "context"

// Hook provides callbacks for monitoring agent operations
type Hook interface {
    // BeforeGenerate is called before generating a response
    BeforeGenerate(ctx context.Context, messages []Message)
    
    // AfterGenerate is called after generating a response
    AfterGenerate(ctx context.Context, response Response, err error)
    
    // BeforeToolCall is called before executing a tool
    BeforeToolCall(ctx context.Context, tool string, params map[string]interface{})
    
    // AfterToolCall is called after executing a tool
    AfterToolCall(ctx context.Context, tool string, result interface{}, err error)
}
```

2. Implement a logging hook:
```go
// pkg/agent/workflow/hooks.go
package workflow

import (
    "context"
    "log/slog"
    
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// LoggingHook implements Hook for logging
type LoggingHook struct {
    logger *slog.Logger
}

// NewLoggingHook creates a new logging hook
func NewLoggingHook(logger *slog.Logger) *LoggingHook {
    return &LoggingHook{
        logger: logger,
    }
}

// Implementation of Hook interface...
```

## Phase 6: Integration and Examples (1-2 weeks)
### Step 17: Create example applications

1. Create a simple example:
```go
// examples/simple/main.go
package main

import (
    "context"
    "fmt"
    "log/slog"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func main() {
    // Create a schema
    schema := &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "name": {Type: "string", Description: "Person's name"},
            "age": {Type: "integer", Minimum: float64Ptr(0)},
            "email": {Type: "string", Format: "email"},
        },
        Required: []string{"name", "email"},
    }
    
    // Create an LLM provider
    llmProvider := provider.NewOpenAIProvider("your-api-key", "gpt-4o")
    
    // Generate structured output
    prompt := "Generate information about a person including their name, age, and email."
    
    result, err := llmProvider.GenerateWithSchema(context.Background(), prompt, schema)
    if err != nil {
        slog.Error("Failed to generate with schema", "error", err)
        return
    }
    
    // Use the validated result
    person := result.(map[string]interface{})
    fmt.Printf("Name: %s\n", person["name"])
    fmt.Printf("Age: %v\n", person["age"])
    fmt.Printf("Email: %s\n", person["email"])
}
```

2. Create an agent example:
```go
// examples/agent/main.go
package main

import (
    "context"
    "fmt"
    "log/slog"
    
    "github.com/lexlapax/go-llms/pkg/agent/tools"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Implementation of a more complex agent example...
```

### Step 18: Write integration tests

1. Create integration tests for key flows:
```go
// tests/integration/validation_test.go
package integration

import "testing"

func TestEndToEndValidation(t *testing.T) {
    // End-to-end tests for validation
}
```

2. Create integration tests for agents:
```go
// tests/integration/agent_test.go
package integration

import "testing"

func TestEndToEndAgent(t *testing.T) {
    // End-to-end tests for agents
}
```

### Step 19: Create comprehensive documentation

Create detailed README.md
Create usage documentation
Create API documentation using Go's documentation tools
Create architecture documentation

## Phase 7: Performance Optimization and Refinement (1-2 weeks)

### Step 20: Optimize performance

Identify and resolve performance bottlenecks
Implement benchmarks:
```go
// pkg/schema/validation/validator_benchmark_test.go
package validation

import "testing"

func BenchmarkValidation(b *testing.B) {
    // Benchmarks for validation
}
```

Optimize memory usage and reduce allocations

### Step 21: Refine the API

Get feedback from users
Ensure consistent error handling
Implement more provider-specific options
Add convenience functions for common operations

### Step 22: Final testing and release

Run full test suite on multiple platforms
Create release notes
Tag and release the first version
Publish package documentation

Timeline Summary

Phase 1: Project Setup and Architecture (1-2 weeks)
Phase 2: Schema Validation Implementation (2-3 weeks)
Phase 3: LLM Provider Integration (2-3 weeks)
Phase 4: Structured Output Implementation (1-2 weeks)
Phase 5: Agent and Tool Implementation (2-3 weeks)
Phase 6: Integration and Examples (1-2 weeks)
Phase 7: Performance Optimization and Refinement (1-2 weeks)

Total estimated time: 10-17 weeks
This implementation plan provides a structured, step-by-step approach to creating go-llms, a Go port of pydantic-ai, using test-driven development with a clean architecture organized by vertical feature slices. Each phase builds on the previous one, ensuring a solid foundation for the library while maintaining proper separation of concerns throughout the development process.