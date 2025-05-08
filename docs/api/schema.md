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
    
    // Conditional validation
    If                  *Schema             `json:"if,omitempty"`
    Then                *Schema             `json:"then,omitempty"`
    Else                *Schema             `json:"else,omitempty"`
    AllOf               []*Schema           `json:"allOf,omitempty"`
    AnyOf               []*Schema           `json:"anyOf,omitempty"`
    OneOf               []*Schema           `json:"oneOf,omitempty"`
    Not                 *Schema             `json:"not,omitempty"`
}
```

The `Schema` struct represents a JSON Schema compatible definition that describes the structure of expected data. It supports various types such as objects, arrays, strings, numbers, etc., as well as advanced conditional validation through logical operators.

#### Property

```go
type Property struct {
    Type                 string              `json:"type"`
    Format               string              `json:"format,omitempty"`
    Description          string              `json:"description,omitempty"`
    Minimum              *float64            `json:"minimum,omitempty"`
    Maximum              *float64            `json:"maximum,omitempty"`
    ExclusiveMinimum     *float64            `json:"exclusiveMinimum,omitempty"`
    ExclusiveMaximum     *float64            `json:"exclusiveMaximum,omitempty"`
    MinLength            *int                `json:"minLength,omitempty"`
    MaxLength            *int                `json:"maxLength,omitempty"`
    MinItems             *int                `json:"minItems,omitempty"`
    MaxItems             *int                `json:"maxItems,omitempty"`
    UniqueItems          *bool               `json:"uniqueItems,omitempty"`
    Pattern              string              `json:"pattern,omitempty"`
    Enum                 []string            `json:"enum,omitempty"`
    Items                *Property           `json:"items,omitempty"`
    Required             []string            `json:"required,omitempty"`
    Properties           map[string]Property `json:"properties,omitempty"`
    AdditionalProperties *bool               `json:"additionalProperties,omitempty"`
    CustomValidator      string              `json:"customValidator,omitempty"`
}
```

The `Property` struct defines the characteristics of a property in a schema, with support for validation constraints like minimum/maximum values, length limits, patterns, etc. It includes advanced array validation features and support for custom validators.

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
// Create a new validator with default settings
validator := validation.NewValidator()

// Create a validator with type coercion enabled
validator := validation.NewValidator(validation.WithCoercion(true))

// Create a validator with custom validation enabled
validator := validation.NewValidator(validation.WithCustomValidation(true))

// Enable both features
validator := validation.NewValidator(
    validation.WithCoercion(true),
    validation.WithCustomValidation(true),
)

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

The validator supports a comprehensive range of validation operations, including:

- Type validation (string, number, boolean, object, array)
- Numeric constraints (minimum, maximum, exclusiveMinimum, exclusiveMaximum)
- String constraints (minLength, maxLength, pattern, enum)
- String format validation (email, date-time, uri, uuid, hostname, ipv4, ipv6, etc.)
- Multi-format string validation with pipe (`|`) or comma (`,`) separators
- Array validation (items, minItems, maxItems, uniqueItems)
- Object validation (required properties, property types, additionalProperties)
- Conditional validation (if-then-else, allOf, anyOf, oneOf, not)
- Custom validation functions

For detailed documentation on advanced validation features, see the [Advanced Validation Guide](../schema/ADVANCED_VALIDATION.md).

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

### Basic Validation

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

### Advanced Validation

```go
// Define a schema with advanced features
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "username": {
            Type: "string", 
            MinLength: intPtr(3), 
            CustomValidator: "alphanumeric",
        },
        "contact": {
            Type: "string",
            Format: "email|phone", // Multiple formats
        },
        "age": {
            Type: "integer",
            ExclusiveMinimum: float64Ptr(18), // Must be > 18
        },
        "tags": {
            Type: "array",
            MinItems: intPtr(1),
            MaxItems: intPtr(5),
            UniqueItems: boolPtr(true),
            Items: &domain.Property{
                Type: "string",
            },
        },
        "account_type": {Type: "string"},
    },
    Required: []string{"username", "contact", "age"},
    
    // Conditional validation: if account_type is "premium", require "payment_info"
    If: &domain.Schema{
        Properties: map[string]domain.Property{
            "account_type": {
                Enum: []string{"premium"},
            },
        },
    },
    Then: &domain.Schema{
        Required: []string{"payment_info"},
        Properties: map[string]domain.Property{
            "payment_info": {Type: "object"},
        },
    },
}

// Create a validator with all features enabled
validator := validation.NewValidator(
    validation.WithCoercion(true),
    validation.WithCustomValidation(true),
)

// Validate JSON data
jsonData := `{
    "username": "john123",
    "contact": "john@example.com",
    "age": 25,
    "tags": ["user", "active", "verified"],
    "account_type": "premium",
    "payment_info": {
        "card_type": "visa",
        "last_four": "1234"
    }
}`

result, err := validator.Validate(schema, jsonData)
// ... handle result
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
