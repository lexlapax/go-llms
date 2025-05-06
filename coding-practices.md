# Blueprint for go-llms: A Pydantic-like Validation Framework for Go

Go-llms offers a robust framework for validating and structuring LLM outputs in Go applications. This blueprint details how to implement a system that captures pydantic-ai's core functionality while embracing Go's idioms and best practices.

## Core architecture follows clean design with vertical slices

Go-llms implements a clean architecture organized by feature rather than technical layer, making the codebase intuitive to navigate while maintaining proper separation of concerns. Each feature slice contains all necessary components from domain logic to external integrations.

The system's core remains isolated from external dependencies through interfaces, allowing for a highly testable system that can evolve independently of specific LLM providers or validation technologies.

**The primary goal of the library is simple**: validate structured outputs from LLMs against predefined schemas while following Go idioms and best practices.

## Directory structure

The project utilizes a feature-based vertical slicing approach where code is organized by capability rather than technical layers:

```
go-llms/
├── cmd/                       # Application entry points
│   └── examples/              # Example applications
├── internal/                  # Internal packages
├── pkg/                       # Public packages
│   ├── schema/                # Schema definition and validation feature
│   │   ├── domain/            # Core domain models
│   │   ├── validation/        # Validation logic
│   │   └── adapter/           # External adapters (JSON, OpenAPI)
│   ├── llm/                   # LLM integration feature
│   │   ├── domain/            # Core LLM domain models
│   │   ├── provider/          # LLM provider implementations
│   │   └── prompt/            # Prompt templating
│   ├── structured/            # Structured output feature
│   │   ├── domain/            # Structured output domain
│   │   ├── processor/         # Output processors
│   │   └── adapter/           # Integration adapters
│   └── agent/                 # Agent feature (tools, workflows)
│       ├── domain/            # Agent domain models
│       ├── tools/             # Tool implementations
│       └── workflow/          # Agent workflows
└── examples/                  # Usage examples
```

Within each feature slice, the code is organized into domain, application, and infrastructure layers, maintaining clean architecture principles while keeping related functionality together.

## Test-driven development approach

Go-llms follows strict TDD principles, with tests driving the design and implementation of all components. This ensures high quality, well-tested code from the start.

### Red-Green-Refactor cycle example

Here's a concrete example showing the implementation of the schema validation feature using TDD:

#### 1. Red: Write a failing test first

```go
// schema/validation/validator_test.go
func TestObjectValidation(t *testing.T) {
    schema := &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "name": {Type: "string", Required: true},
            "age": {Type: "integer", Minimum: float64Ptr(0)},
        },
    }
    
    validator := NewValidator()
    
    // Valid case
    t.Run("valid object", func(t *testing.T) {
        input := `{"name": "John", "age": 30}`
        result, err := validator.Validate(schema, input)
        
        if err != nil {
            t.Fatalf("Unexpected error: %v", err)
        }
        if !result.Valid {
            t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
        }
    })
    
    // Invalid case
    t.Run("invalid object - missing required", func(t *testing.T) {
        input := `{"age": 30}`
        result, err := validator.Validate(schema, input)
        
        if err != nil {
            t.Fatalf("Unexpected error: %v", err)
        }
        if result.Valid {
            t.Errorf("Expected invalid result")
        }
        if !containsError(result.Errors, "name", "required") {
            t.Errorf("Expected 'required' error for 'name', got: %v", result.Errors)
        }
    })
}
```

#### 2. Green: Implement just enough to pass the test

```go
// schema/validation/validator.go
func (v *Validator) Validate(schema *domain.Schema, jsonStr string) (*domain.ValidationResult, error) {
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }
    
    result := &domain.ValidationResult{Valid: true}
    var errors []string
    
    // Check required properties
    for name, prop := range schema.Properties {
        if prop.Required {
            if _, exists := data[name]; !exists {
                errors = append(errors, fmt.Sprintf("property %q is required", name))
            }
        }
    }
    
    // Basic type validation
    for name, value := range data {
        if prop, exists := schema.Properties[name]; exists {
            switch prop.Type {
            case "integer", "number":
                if _, ok := value.(float64); !ok {
                    errors = append(errors, fmt.Sprintf("property %q must be a number", name))
                }
            case "string":
                if _, ok := value.(string); !ok {
                    errors = append(errors, fmt.Sprintf("property %q must be a string", name))
                }
            }
        }
    }
    
    if len(errors) > 0 {
        result.Valid = false
        result.Errors = errors
    }
    
    return result, nil
}
```

#### 3. Refactor: Improve the implementation

```go
// schema/validation/validator.go
func (v *Validator) Validate(schema *domain.Schema, jsonStr string) (*domain.ValidationResult, error) {
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }
    
    return v.ValidateMap(schema, data)
}

func (v *Validator) ValidateMap(schema *domain.Schema, data map[string]interface{}) (*domain.ValidationResult, error) {
    result := &domain.ValidationResult{Valid: true}
    var errors []string
    
    // Validate required properties
    errors = append(errors, v.validateRequired(schema, data)...)
    
    // Validate property types and constraints
    errors = append(errors, v.validateProperties(schema, data)...)
    
    if len(errors) > 0 {
        result.Valid = false
        result.Errors = errors
    }
    
    return result, nil
}

func (v *Validator) validateRequired(schema *domain.Schema, data map[string]interface{}) []string {
    var errors []string
    
    for name, prop := range schema.Properties {
        if prop.Required {
            if _, exists := data[name]; !exists {
                errors = append(errors, fmt.Sprintf("property %q is required", name))
            }
        }
    }
    
    return errors
}

func (v *Validator) validateProperties(schema *domain.Schema, data map[string]interface{}) []string {
    var errors []string
    
    for name, value := range data {
        if prop, exists := schema.Properties[name]; exists {
            propErrors := v.validateProperty(name, prop, value)
            errors = append(errors, propErrors...)
        }
    }
    
    return errors
}

// Additional validation methods for different types...
```

This example demonstrates how the TDD cycle works in practice, starting with a failing test, implementing just enough code to make it pass, and then refactoring for better design.

## Interface definitions for clean architecture

Go-llms uses interfaces at domain boundaries to ensure proper separation of concerns. Here are some key interfaces:

### Schema validation domain

```go
// schema/domain/interfaces.go
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

// SchemaRepository defines storage operations for schemas
type SchemaRepository interface {
    // Get retrieves a schema by ID
    Get(id string) (*Schema, error)
    
    // Save stores a schema
    Save(id string, schema *Schema) error
    
    // Delete removes a schema
    Delete(id string) error
}
```

### LLM provider domain

```go
// llm/domain/interfaces.go
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

// ModelRegistry manages available LLM models
type ModelRegistry interface {
    // RegisterModel adds a model to the registry
    RegisterModel(name string, provider Provider) error
    
    // GetModel retrieves a model by name
    GetModel(name string) (Provider, error)
    
    // ListModels returns all available models
    ListModels() []string
}
```

### Agent domain

```go
// agent/domain/interfaces.go
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
    
    // WithModel specifies which LLM model to use
    WithModel(modelName string) Agent
}
```

These interfaces provide clear contracts between different parts of the system while allowing implementation details to vary independently.

## Core validation components

The validation system forms the backbone of go-llms, providing the ability to validate structured outputs from LLMs against predefined schemas.

### Schema definition

Schemas are defined using Go structs that can be serialized to JSON Schema:

```go
// schema/domain/schema.go
package domain

// Schema represents a JSON Schema compatible definition
type Schema struct {
    Type                 string              `json:"type"`
    Properties           map[string]Property `json:"properties,omitempty"`
    Required             []string            `json:"required,omitempty"`
    AdditionalProperties *bool               `json:"additionalProperties,omitempty"`
    Description          string              `json:"description,omitempty"`
    Title                string              `json:"title,omitempty"`
}

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
    Required    []string    `json:"required,omitempty"`
    Properties  map[string]Property `json:"properties,omitempty"`
}
```

### Schema validation

The validation component provides comprehensive validation against schemas:

```go
// schema/validation/validator.go
package validation

import (
    "encoding/json"
    "fmt"
    "regexp"
    "strings"
    
    "github.com/yourusername/go-llms/pkg/schema/domain"
)

// DefaultValidator implements schema validation
type DefaultValidator struct{}

// NewValidator creates a new validator
func NewValidator() *DefaultValidator {
    return &DefaultValidator{}
}

// Validate validates a JSON string against a schema
func (v *DefaultValidator) Validate(schema *domain.Schema, jsonStr string) (*domain.ValidationResult, error) {
    var data interface{}
    
    // Parse JSON
    if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }
    
    // Validate against schema
    result := &domain.ValidationResult{Valid: true}
    errors := v.validateValue("", schema, data)
    
    if len(errors) > 0 {
        result.Valid = false
        result.Errors = errors
    }
    
    return result, nil
}

// Validation implementation...
```

## LLM integration approaches

Go-llms integrates with multiple LLM providers through a common interface, allowing applications to easily switch between providers.

### Provider interface

```go
// llm/domain/provider.go
package domain

import "context"

// Provider defines the interface for LLM providers
type Provider interface {
    // Generate produces text from a prompt
    Generate(ctx context.Context, prompt string, options ...Option) (string, error)
    
    // GenerateWithSchema produces structured output conforming to a schema
    GenerateWithSchema(ctx context.Context, prompt string, schema interface{}, options ...Option) (interface{}, error)
    
    // Stream streams responses token by token
    Stream(ctx context.Context, prompt string, options ...Option) (<-chan Token, error)
}
```

### OpenAI implementation

```go
// llm/provider/openai.go
package provider

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/yourusername/go-llms/pkg/llm/domain"
    "github.com/yourusername/go-llms/pkg/schema/domain"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
    apiKey     string
    model      string
    validator  domain.Validator
}

// Generate sends a prompt to OpenAI and returns the response
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
    // Implementation...
}

// GenerateWithSchema generates LLM output that conforms to a schema
func (p *OpenAIProvider) GenerateWithSchema(ctx context.Context, prompt string, schema interface{}, options ...domain.Option) (interface{}, error) {
    // Convert schema to JSON Schema
    schemaJSON, err := json.Marshal(schema)
    if err != nil {
        return nil, err
    }
    
    // Add schema to prompt
    enhancedPrompt := fmt.Sprintf(
        "You must respond with JSON that exactly matches this schema: %s\n\n%s",
        schemaJSON, prompt,
    )
    
    // Try to generate valid output, with retries
    var output string
    for attempt := 0; attempt < 3; attempt++ {
        output, err = p.Generate(ctx, enhancedPrompt, options...)
        if err != nil {
            return nil, err
        }
        
        // Validate output against schema
        result, err := p.validator.Validate(schema.(*domain.Schema), output)
        if err == nil && result.Valid {
            // Parse JSON and return
            var parsed interface{}
            if err := json.Unmarshal([]byte(output), &parsed); err == nil {
                return parsed, nil
            }
        }
        
        // If invalid, enhance prompt with error feedback
        if result != nil && len(result.Errors) > 0 {
            errorMsg := strings.Join(result.Errors, ", ")
            enhancedPrompt = fmt.Sprintf(
                "Your previous response had validation errors: %s\n\nYou must respond with JSON that exactly matches this schema: %s\n\n%s",
                errorMsg, schemaJSON, prompt,
            )
        }
    }
    
    return nil, fmt.Errorf("failed to generate valid output after 3 attempts")
}
```

## Implementation challenges and solutions

### Type coercion

**Challenge**: Go lacks the dynamic type conversion capabilities of Python.

**Solution**: Implement a custom coercion system that handles common LLM output patterns:

```go
// schema/validation/coercion.go
package validation

import (
    "strconv"
    "strings"
)

// Coerce attempts to convert a value to the expected type
func (v *DefaultValidator) Coerce(targetType string, value interface{}) (interface{}, bool) {
    switch targetType {
    case "string":
        return coerceToString(value)
    case "integer":
        return coerceToInteger(value)
    case "number":
        return coerceToNumber(value)
    case "boolean":
        return coerceToBoolean(value)
    default:
        return value, false
    }
}

func coerceToString(value interface{}) (interface{}, bool) {
    switch v := value.(type) {
    case string:
        return v, true
    case float64:
        return strconv.FormatFloat(v, 'f', -1, 64), true
    case int:
        return strconv.Itoa(v), true
    case bool:
        return strconv.FormatBool(v), true
    default:
        return nil, false
    }
}

// Additional coercion functions...
```

### Schema generation from Go structs

**Challenge**: Creating schemas manually is tedious and error-prone.

**Solution**: Use reflection to automatically generate schemas from Go structs:

```go
// schema/adapter/reflection/schema_generator.go
package reflection

import (
    "reflect"
    
    "github.com/yourusername/go-llms/pkg/schema/domain"
)

// GenerateSchema creates a schema from a Go struct
func GenerateSchema(obj interface{}) (*domain.Schema, error) {
    t := reflect.TypeOf(obj)
    
    // Handle pointers
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    
    // Only handle structs
    if t.Kind() != reflect.Struct {
        return nil, fmt.Errorf("schema generation only supports structs, got %s", t.Kind())
    }
    
    schema := &domain.Schema{
        Type:       "object",
        Properties: make(map[string]domain.Property),
    }
    
    var required []string
    
    // Process each field
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        
        // Skip unexported fields
        if !field.IsExported() {
            continue
        }
        
        // Get JSON field name or use struct field name
        jsonTag := field.Tag.Get("json")
        name := field.Name
        if jsonTag != "" {
            parts := strings.Split(jsonTag, ",")
            if parts[0] != "" && parts[0] != "-" {
                name = parts[0]
            }
        }
        
        // Create property
        prop := generateProperty(field)
        schema.Properties[name] = prop
        
        // Check if required
        if tag := field.Tag.Get("validate"); strings.Contains(tag, "required") {
            required = append(required, name)
        }
    }
    
    if len(required) > 0 {
        schema.Required = required
    }
    
    return schema, nil
}

// Implementation details...
```

### Testing with external dependencies

**Challenge**: Testing components that rely on external LLM APIs is difficult.

**Solution**: Create mock implementations that can be used in tests:

```go
// llm/provider/mock.go
package provider

import (
    "context"
    
    "github.com/yourusername/go-llms/pkg/llm/domain"
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
                ch <- domain.Token{Text: "This", Finished: false}
                ch <- domain.Token{Text: " is", Finished: false}
                ch <- domain.Token{Text: " a", Finished: false}
                ch <- domain.Token{Text: " mock", Finished: false}
                ch <- domain.Token{Text: " response", Finished: true}
            }()
            return ch, nil
        },
    }
}

// Generate returns the mock generate function result
func (p *MockProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
    return p.generateFunc(ctx, prompt, options...)
}

// WithGenerateFunc sets a custom generate function
func (p *MockProvider) WithGenerateFunc(f func(ctx context.Context, prompt string, options ...domain.Option) (string, error)) *MockProvider {
    p.generateFunc = f
    return p
}

// Additional methods...
```

## Example usage patterns

### Basic validation of LLM output

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/yourusername/go-llms/pkg/llm/provider"
    "github.com/yourusername/go-llms/pkg/schema/domain"
    "github.com/yourusername/go-llms/pkg/schema/validation"
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
    llmProvider := provider.NewOpenAIProvider("your-api-key", "gpt-4")
    
    // Generate structured output
    prompt := "Generate information about a person including their name, age, and email."
    
    result, err := llmProvider.GenerateWithSchema(context.Background(), prompt, schema)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // Use the validated result
    person := result.(map[string]interface{})
    fmt.Printf("Name: %s\n", person["name"])
    fmt.Printf("Age: %v\n", person["age"])
    fmt.Printf("Email: %s\n", person["email"])
}

func float64Ptr(v float64) *float64 {
    return &v
}
```

### Creating an agent with tools

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/yourusername/go-llms/pkg/agent"
    "github.com/yourusername/go-llms/pkg/llm/provider"
    "github.com/yourusername/go-llms/pkg/schema/domain"
)

// WeatherResponse defines the schema for weather data
type WeatherResponse struct {
    Temperature float64 `json:"temperature"`
    Condition   string  `json:"condition"`
    Location    string  `json:"location"`
}

func main() {
    // Create an LLM provider
    llmProvider := provider.NewOpenAIProvider("your-api-key", "gpt-4")
    
    // Create an agent
    weatherAgent := agent.NewAgent(llmProvider).
        SetSystemPrompt("You are a helpful weather assistant.").
        AddTool(agent.NewTool("get_weather", "Get current weather", getWeather))
    
    // Define output schema
    outputSchema := &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "response": {Type: "string", Description: "A human-friendly weather report"},
        },
        Required: []string{"response"},
    }
    
    // Run the agent
    result, err := weatherAgent.RunWithSchema(
        context.Background(),
        "What's the weather like in New York?",
        outputSchema,
    )
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // Use the result
    fmt.Println(result.(map[string]interface{})["response"])
}

// getWeather is a tool function that retrieves weather data
func getWeather(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    location, ok := params["location"].(string)
    if !ok {
        return nil, fmt.Errorf("location parameter is required")
    }
    
    // In a real implementation, this would call a weather API
    return WeatherResponse{
        Temperature: 72.5,
        Condition:   "Sunny",
        Location:    location,
    }, nil
}
```

### Generating schemas from Go structs

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    
    "github.com/yourusername/go-llms/pkg/schema/adapter/reflection"
)

// Person defines a struct for a person
type Person struct {
    Name     string   `json:"name" validate:"required"`
    Age      int      `json:"age" validate:"min=0,max=120"`
    Email    string   `json:"email" validate:"required,email"`
    Address  Address  `json:"address"`
    Hobbies  []string `json:"hobbies"`
}

// Address defines a struct for an address
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city" validate:"required"`
    Country string `json:"country" validate:"required"`
    ZipCode string `json:"zipCode"`
}

func main() {
    // Generate schema from Go struct
    schema, err := reflection.GenerateSchema(Person{})
    if err != nil {
        log.Fatalf("Error generating schema: %v", err)
    }
    
    // Convert schema to JSON
    schemaJSON, err := json.MarshalIndent(schema, "", "  ")
    if err != nil {
        log.Fatalf("Error marshaling schema: %v", err)
    }
    
    fmt.Println(string(schemaJSON))
}
```

## Conclusion

This blueprint provides a comprehensive guide for implementing "go-llms", a Go port of pydantic-ai that follows test-driven development principles, feature-based vertical slicing, and clean architecture patterns. 

By following this approach, you'll create a robust library that validates LLM outputs against predefined schemas, supports multiple LLM providers, and follows Go best practices. The modular design and clear separation of concerns make the codebase maintainable and extensible, while comprehensive testing ensures reliability.

With go-llms, Go developers can build LLM-powered applications with strong typing and data validation, bringing the powerful validation capabilities of pydantic-ai to the Go ecosystem.