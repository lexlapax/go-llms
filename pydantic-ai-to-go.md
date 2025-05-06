# Porting pydantic-ai to Go: a blueprint for idiomatic implementation

Pydantic-ai is a sophisticated framework for building LLM-powered applications with structured outputs and type safety. This plan outlines how to create a Go implementation that captures its core functionality while embracing Go's idioms and strengths.

## Core functionality and architecture of pydantic-ai

Pydantic-ai solves several key problems in LLM application development:

1. **Structured responses**: It validates LLM outputs against predefined schemas to ensure consistency.
2. **Model-agnostic interface**: It provides a unified API across different LLM providers.
3. **Type safety**: It leverages Python's type system for better developer experience.
4. **Dependency injection**: It enables passing data and services into agents via a context object.
5. **Tool integration**: It allows LLMs to interact with external systems through function calls.

The system follows an agent-based architecture with clear separation between:

- **Agent**: Central interface for interacting with LLMs
- **Models**: Abstract vendor-specific API details
- **Providers**: Handle authentication and connections
- **RunContext**: Carries dependencies to tools and system prompts
- **Messages**: Represent communication with the LLM

## Go port architecture

### High-level design principles

The Go implementation will follow these principles:

- **Interface-based design**: Clearly defined interfaces following Go's composition pattern
- **Struct-tag validation**: Schema validation using struct tags instead of Python's type hints
- **Error-return pattern**: Return errors rather than throwing exceptions
- **Concurrency support**: Leverage goroutines and channels for parallel operations
- **Minimal dependencies**: Use standard library where possible, with few external libraries

### Core components

```go
// Simplified component diagram
┌────────────┐     ┌───────────┐     ┌────────────┐
│   Agent    │────▶│   Model   │────▶│  Provider  │
└────────────┘     └───────────┘     └────────────┘
       │                 │                  │
       ▼                 ▼                  ▼
┌────────────┐     ┌───────────┐     ┌────────────┐
│   Tools    │◀───▶│ Messages  │◀───▶│ Validation │
└────────────┘     └───────────┘     └────────────┘
```

#### Agent

```go
// Agent is the main interface for interacting with LLMs
type Agent[D any, O any] struct {
    model       Model
    tools       map[string]Tool[D]
    systemPrompt string
    outputType   reflect.Type
}

// NewAgent creates a new agent with functional options
func NewAgent[D any, O any](model Model, opts ...AgentOption[D, O]) *Agent[D, O] {
    agent := &Agent[D, O]{
        model: model,
        tools: make(map[string]Tool[D]),
    }
    
    for _, opt := range opts {
        opt(agent)
    }
    
    return agent
}

// Run executes the agent with the given input and dependencies
func (a *Agent[D, O]) Run(ctx context.Context, input string, deps D) (O, error) {
    // Implement agent execution logic
}

// RegisterTool adds a tool to the agent
func (a *Agent[D, O]) RegisterTool(name string, tool Tool[D]) {
    a.tools[name] = tool
}
```

#### Model interface

```go
// Model represents a language model provider
type Model interface {
    // Generate creates a completion from messages
    Generate(ctx context.Context, messages []Message) (Response, error)
    
    // GenerateStream streams a completion from messages
    GenerateStream(ctx context.Context, messages []Message) (ResponseStream, error)
}

// Concrete implementations
type OpenAIModel struct {
    client    *openai.Client
    modelName string
    // Configuration options
}

type AnthropicModel struct {
    client    *anthropic.Client
    modelName string
    // Configuration options
}
```

#### Tool system

```go
// Tool represents a function that can be called by the LLM
type Tool[D any] interface {
    // Name returns the name of the tool
    Name() string
    
    // Description returns a description of the tool
    Description() string
    
    // Schema returns the JSON schema for the tool's parameters
    Schema() map[string]interface{}
    
    // Execute runs the tool with the given parameters and dependencies
    Execute(ctx context.Context, params map[string]interface{}, deps D) (interface{}, error)
}

// ToolFunc creates a Tool from a function using reflection
func ToolFunc[D any](name, description string, fn interface{}) Tool[D] {
    // Implementation using reflection to extract parameter information
    // and generate JSON schema
}
```

#### Validation system

```go
// Validator validates and converts data
type Validator interface {
    // Validate validates the data against the schema
    Validate(data interface{}) error
    
    // Convert converts the data to the target type
    Convert(data interface{}, target interface{}) error
}

// SchemaGenerator generates JSON schema from Go types
type SchemaGenerator interface {
    // GenerateSchema generates a JSON schema from a Go type
    GenerateSchema(t reflect.Type) map[string]interface{}
}
```

#### RunContext

```go
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

## Schema validation approach

Go doesn't have Python's rich type hints, so we'll use a combination of struct tags and reflection:

```go
// Example of struct tag validation
type UserProfile struct {
    Username string `json:"username" validate:"required,min=3,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18,lte=99"`
}

// ValidateStruct validates a struct using reflection and tags
func ValidateStruct(s interface{}) error {
    v := reflect.ValueOf(s)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }
    
    t := v.Type()
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        value := v.Field(i)
        
        tag := field.Tag.Get("validate")
        if tag == "" {
            continue
        }
        
        rules := strings.Split(tag, ",")
        for _, rule := range rules {
            if err := validateRule(rule, value); err != nil {
                return fmt.Errorf("field %s: %w", field.Name, err)
            }
        }
    }
    
    return nil
}
```

## Hooks for debugging and monitoring

Using Go's interface system, we can create hooks for debugging and monitoring:

```go
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

// Agent with hooks
type Agent[D any, O any] struct {
    // ...existing fields
    hooks []Hook
}

// WithHook adds a hook to the agent
func WithHook[D any, O any](hook Hook) AgentOption[D, O] {
    return func(a *Agent[D, O]) {
        a.hooks = append(a.hooks, hook)
    }
}

// LoggingHook implements Hook for logging purposes
type LoggingHook struct {
    logger Logger
}

func (h *LoggingHook) BeforeGenerate(ctx context.Context, messages []Message) {
    h.logger.Info("Generating response", "messages", messages)
}

// More hook methods...
```

## Recommended Go libraries

To minimize external dependencies while providing the necessary functionality:

1. **go-playground/validator**: For struct validation with tags
2. **encoding/json**: Standard library for JSON handling
3. **reflect**: For runtime type inspection
4. **context**: For cancellation and timeouts
5. **sync**: For synchronization primitives
6. **openai-go**: Official OpenAI client for Go
7. **pkg/errors**: For advanced error handling (optional)

## Implementation challenges and solutions

### Challenge 1: Dynamic typing in a static language

**Problem**: Python's dynamic typing allows runtime type modifications; Go is statically typed.

**Solution**:
- Use Go's reflection for runtime type inspection
- Implement `interface{}` (or `any` in Go 1.18+) for dynamic values
- Type assertions to safely convert values
- Implement `Validator` interface for custom validation

```go
// Example of handling dynamic types
func convertToTargetType(value interface{}, targetType reflect.Type) (interface{}, error) {
    switch targetType.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        switch v := value.(type) {
        case float64:
            return int64(v), nil
        case string:
            return strconv.ParseInt(v, 10, 64)
        default:
            return nil, fmt.Errorf("cannot convert %T to int", value)
        }
    // Handle other types...
    default:
        return nil, fmt.Errorf("unsupported target type: %v", targetType)
    }
}
```

### Challenge 2: Decorator pattern implementation

**Problem**: Python's decorators are heavily used in pydantic-ai; Go doesn't have direct equivalents.

**Solution**:
- Use higher-order functions to wrap behavior
- Implement the functional options pattern for configuration
- Use method chaining for a fluent API

```go
// Example of function-based tool registration instead of decorators
func (a *Agent[D, O]) AddTool(name, description string, fn interface{}) *Agent[D, O] {
    tool := ToolFunc[D](name, description, fn)
    a.tools[name] = tool
    return a
}

// Usage
agent.AddTool(
    "search",
    "Search the web for information",
    func(ctx *RunContext[UserDeps], query string) (string, error) {
        // Implementation
        return "Search results", nil
    },
)
```

### Challenge 3: Structured output validation

**Problem**: Ensuring LLM outputs conform to predefined schemas.

**Solution**:
- Generate JSON schema from Go structs using reflection
- Send schema to LLM as part of the prompt
- Validate responses against the schema
- Re-prompt when validation fails

```go
// Example of result validation
func validateAndConvertResult[O any](result string) (O, error) {
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(result), &data); err != nil {
        return *new(O), fmt.Errorf("invalid JSON: %w", err)
    }
    
    var output O
    if err := mapstructure.Decode(data, &output); err != nil {
        return *new(O), fmt.Errorf("schema validation failed: %w", err)
    }
    
    return output, nil
}
```

## Implementation roadmap

1. **Core interfaces**: Define the fundamental interfaces for the system
2. **Validation system**: Implement struct tag-based validation
3. **JSON schema generation**: Create schema generator from Go types
4. **Model implementations**: Connect to OpenAI, Anthropic, etc.
5. **Agent implementation**: Core agent functionality
6. **Tool system**: Function registration and execution
7. **Dependency injection**: Context-based dependency system
8. **Streaming support**: Implement streaming responses
9. **Hook system**: Add debugging and monitoring hooks
10. **Examples and documentation**: Create comprehensive examples

## Sample usage

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/goagent"
)

// Dependencies for the agent
type SupportDeps struct {
    CustomerID int
    DB         *DatabaseConn
}

// Output structure
type SupportOutput struct {
    SupportAdvice string `json:"support_advice" validate:"required"`
    BlockCard     bool   `json:"block_card"`
    Risk          int    `json:"risk" validate:"gte=0,lte=10"`
}

// Create a database connection
type DatabaseConn struct{}

func (db *DatabaseConn) CustomerName(id int) (string, error) {
    // Mock implementation
    return "John Doe", nil
}

func (db *DatabaseConn) CustomerBalance(id int, includePending bool) (float64, error) {
    // Mock implementation
    return 123.45, nil
}

func main() {
    // Create an OpenAI model
    model := goagent.NewOpenAIModel("gpt-4", "your-api-key")
    
    // Create an agent
    agent := goagent.NewAgent[SupportDeps, SupportOutput](
        model,
        goagent.WithSystemPrompt("You are a support agent in our bank, give the customer support and judge the risk level of their query."),
        goagent.WithOutputType((*SupportOutput)(nil)),
    )
    
    // Add a dynamic system prompt
    agent.AddSystemPrompt(func(ctx *goagent.RunContext[SupportDeps]) (string, error) {
        name, err := ctx.Deps().DB.CustomerName(ctx.Deps().CustomerID)
        if err != nil {
            return "", err
        }
        return fmt.Sprintf("The customer's name is %q", name), nil
    })
    
    // Register a tool
    agent.AddTool(
        "customer_balance",
        "Returns the customer's current account balance",
        func(ctx *goagent.RunContext[SupportDeps], includePending bool) (float64, error) {
            return ctx.Deps().DB.CustomerBalance(ctx.Deps().CustomerID, includePending)
        },
    )
    
    // Add a logging hook
    agent.WithHook(goagent.NewLoggingHook(goagent.DefaultLogger))
    
    // Run the agent
    ctx := context.Background()
    deps := SupportDeps{
        CustomerID: 123,
        DB:         &DatabaseConn{},
    }
    
    result, err := agent.Run(ctx, "What is my balance?", deps)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Support advice: %s\n", result.SupportAdvice)
    fmt.Printf("Block card: %v\n", result.BlockCard)
    fmt.Printf("Risk level: %d\n", result.Risk)
}
```

## Conclusion

This plan outlines a Go port of pydantic-ai that respects Go's idioms and strengths while maintaining the core functionality of the original library. The implementation leverages Go's strong typing, concurrency model, and interface-based design to create a robust framework for building LLM applications.

By following this blueprint, you can create a powerful, idiomatic Go library that enables developers to build production-grade LLM applications with structured outputs, strong type safety, and minimal external dependencies.