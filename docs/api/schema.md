# Schema Package

The `schema` package provides functionality for defining, validating, and working with schemas that describe structured data formats. It is primarily used to validate the output of language models against expected structures.

## Core Components

### Domain

#### Schema

```go
type Schema struct {
    Type                 string              `json:"type"`
    Properties           map[string]Property `json:"properties,omitempty"`
    Required             []string            `json:"required,omitempty"`
    AdditionalProperties *bool               `json:"additionalProperties,omitempty"`
    Description          string              `json:"description,omitempty"`
    Title                string              `json:"title,omitempty"`
}
```

The `Schema` struct represents a JSON Schema compatible definition that describes the structure of expected data. It supports various types such as objects, arrays, strings, numbers, etc.

#### Property

```go
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

The `Property` struct defines the characteristics of a property in a schema, with support for validation constraints like minimum/maximum values, length limits, patterns, etc.

#### ValidationResult

```go
type ValidationResult struct {
    Valid  bool     `json:"valid"`
    Errors []string `json:"errors,omitempty"`
}
```

The `ValidationResult` struct represents the outcome of a validation operation, indicating whether the data is valid and any validation errors encountered.

### Validator Interface

```go
type Validator interface {
    // Validate checks if data conforms to the schema
    Validate(schema *Schema, data string) (*ValidationResult, error)
    
    // ValidateStruct validates a Go struct against a schema
    ValidateStruct(schema *Schema, obj interface{}) (*ValidationResult, error)
}
```

The `Validator` interface provides methods to validate data against schemas, supporting both raw JSON strings and Go structs.

## Validation Package

The validation package implements the `Validator` interface and provides functionality for validating data against schemas.

### Default Validator

```go
// Create a new validator
validator := validation.NewValidator()

// Validate JSON data against a schema
result, err := validator.Validate(schema, jsonData)
if err != nil {
    // Handle error
}

if !result.Valid {
    // Handle validation errors
    for _, err := range result.Errors {
        fmt.Println(err)
    }
}
```

The default validator supports a wide range of validation operations, including:

- Type validation (string, number, boolean, object, array)
- Numeric constraints (minimum, maximum, exclusiveMinimum, exclusiveMaximum)
- String constraints (minLength, maxLength, pattern, enum)
- Array validation (items, minItems, maxItems, uniqueItems)
- Object validation (required properties, property types, additionalProperties)

## Schema Generation

The `adapter/reflection` package provides utilities for generating schemas from Go structs:

```go
// Generate a schema from a Go struct
schema, err := reflection.GenerateSchema(MyStruct{})
if err != nil {
    // Handle error
}

// Use the generated schema
// ...
```

Schema generation supports:

- Basic types (string, number, boolean, etc.)
- Nested objects and arrays
- Custom tags for defining validation rules
- Optional fields and default values

## Type Coercion

The validation package also provides type coercion functionality to handle flexible inputs:

```go
// Coerce a value to the expected type
coercedValue, success := validator.Coerce("integer", "42")
if success {
    // coercedValue will be of type int64
}
```

Coercion supports converting between compatible types such as:

- String to number/boolean/date
- Number to string
- Boolean to string
- And more

## Example Usage

```go
// Define a schema
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "name": {Type: "string", MinLength: intPtr(3), MaxLength: intPtr(50)},
        "age": {Type: "integer", Minimum: float64Ptr(0), Maximum: float64Ptr(120)},
        "email": {Type: "string", Format: "email"},
    },
    Required: []string{"name", "email"},
}

// Create a validator
validator := validation.NewValidator()

// Validate JSON data
jsonData := `{"name":"John Doe","age":30,"email":"john.doe@example.com"}`
result, err := validator.Validate(schema, jsonData)

if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

if !result.Valid {
    fmt.Println("Validation failed:")
    for _, err := range result.Errors {
        fmt.Printf("- %s\n", err)
    }
    return
}

fmt.Println("Validation successful!")
```

## Helper Functions

```go
// Create integer pointer
func intPtr(i int) *int {
    return &i
}

// Create float64 pointer
func float64Ptr(f float64) *float64 {
    return &f
}
```
