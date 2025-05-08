# Advanced Schema Validation Features

> **[Documentation Home](/REFERENCE.md) / [User Guide](/docs/user-guide/) / Advanced Validation**

This document describes the advanced schema validation features available in the Go-LLMs schema validation package.

*Related: [Getting Started](getting-started.md) | [API Reference](/docs/api/schema.md) | [Performance Optimization](/docs/technical/performance.md#validation-optimization)*

## Table of Contents

1. [Array Validation](#array-validation)
2. [Number Validation](#number-validation)
3. [Multi-Format String Validation](#multi-format-string-validation)
4. [Conditional Validation](#conditional-validation)
5. [Custom Validation Functions](#custom-validation-functions)

## Array Validation

The schema validation package now supports advanced array validation features:

### Min/Max Items

You can specify the minimum and maximum number of items an array should contain:

```go
schema := &domain.Schema{
    Type: "array",
    Properties: map[string]domain.Property{
        "": {
            MinItems: intPtr(2),  // Array must have at least 2 items
            MaxItems: intPtr(10), // Array must have at most 10 items
            Items: &domain.Property{
                Type: "string",
            },
        },
    },
}
```

### Unique Items

You can require that all array items be unique:

```go
schema := &domain.Schema{
    Type: "array",
    Properties: map[string]domain.Property{
        "": {
            UniqueItems: boolPtr(true), // All items must be unique
            Items: &domain.Property{
                Type: "string",
            },
        },
    },
}
```

The uniqueness check works with primitive types as well as complex objects, comparing their JSON representation for uniqueness.

## Number Validation

In addition to the existing minimum and maximum value constraints, the validation package now supports exclusive ranges:

### Exclusive Minimum/Maximum

```go
schema := &domain.Schema{
    Type: "number",
    Properties: map[string]domain.Property{
        "": {
            ExclusiveMinimum: float64Ptr(5.0), // Value must be > 5.0
            ExclusiveMaximum: float64Ptr(10.0), // Value must be < 10.0
        },
    },
}
```

With these constraints:
- `4.9` is invalid (less than exclusive minimum)
- `5.0` is invalid (equal to exclusive minimum)
- `7.5` is valid (between exclusive minimum and maximum)
- `10.0` is invalid (equal to exclusive maximum)
- `10.1` is invalid (greater than exclusive maximum)

## Multi-Format String Validation

The validation package now supports multiple alternative formats for string validation using pipe (`|`) or comma (`,`) separators:

```go
schema := &domain.Schema{
    Type: "string",
    Properties: map[string]domain.Property{
        "": {
            Format: "email|hostname",  // String must be a valid email OR hostname
        },
    },
}
```

The validator will check each format in sequence and consider the validation successful if the value matches any of the specified formats.

### Supported Formats

- `email` - Valid email address
- `date-time` - ISO8601 date-time
- `date` - ISO8601 date
- `uri`, `url` - Valid URI/URL
- `uuid` - Valid UUID
- `duration` - Valid duration string
- `ip` - IPv4 or IPv6 address
- `ipv4` - IPv4 address
- `ipv6` - IPv6 address
- `hostname` - Valid hostname
- `base64` - Valid base64-encoded string
- `json` - Valid JSON string

## Conditional Validation

The validation package now supports conditional validation using JSON Schema constructs like if-then-else, allOf, anyOf, oneOf, and not:

### If-Then-Else

Apply different validation rules based on a condition:

```go
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "type": {Type: "string"},
        "value": {Type: "string"},
    },
    If: &domain.Schema{
        Properties: map[string]domain.Property{
            "type": {
                Enum: []string{"email"},
            },
        },
    },
    Then: &domain.Schema{
        Properties: map[string]domain.Property{
            "value": {
                Format: "email",
            },
        },
    },
    Else: &domain.Schema{
        Properties: map[string]domain.Property{
            "value": {
                MinLength: intPtr(3),
            },
        },
    },
}
```

In this example:
- If `type` is "email", then `value` must be a valid email
- Otherwise, `value` must be at least 3 characters long

### AllOf, AnyOf, OneOf

Combine schemas with different logical operators:

```go
// AllOf - Data must be valid against ALL schemas
schema := &domain.Schema{
    AllOf: []*domain.Schema{
        {
            Type: "object",
            Properties: map[string]domain.Property{
                "name": {Type: "string", MinLength: intPtr(1)},
            },
            Required: []string{"name"},
        },
        {
            Type: "object",
            Properties: map[string]domain.Property{
                "age": {Type: "integer", Minimum: float64Ptr(0)},
            },
            Required: []string{"age"},
        },
    },
}

// AnyOf - Data must be valid against AT LEAST ONE schema
schema := &domain.Schema{
    AnyOf: []*domain.Schema{
        {
            Type: "object",
            Properties: map[string]domain.Property{
                "name": {Type: "string"},
            },
            Required: []string{"name"},
        },
        {
            Type: "object",
            Properties: map[string]domain.Property{
                "id": {Type: "integer"},
            },
            Required: []string{"id"},
        },
    },
}

// OneOf - Data must be valid against EXACTLY ONE schema
schema := &domain.Schema{
    OneOf: []*domain.Schema{
        {
            Type: "object",
            Properties: map[string]domain.Property{
                "name": {Type: "string"},
            },
            Required: []string{"name"},
        },
        {
            Type: "object",
            Properties: map[string]domain.Property{
                "id": {Type: "integer"},
            },
            Required: []string{"id"},
        },
    },
}
```

### Not

Invert validation logic with the Not operator:

```go
schema := &domain.Schema{
    Not: &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "type": {
                Enum: []string{"admin"},
            },
        },
    },
}
```

In this example, the data is valid if it does NOT match the inner schema - in other words, if `type` is not "admin".

## Custom Validation Functions

The validation package now supports custom validation functions that can be registered and applied to properties:

### Registering Custom Validators

```go
// Define a custom validator function
func ValidateEven(value interface{}, displayPath string) []string {
    var errors []string
    
    if num, ok := value.(float64); ok {
        if int(num)%2 != 0 {
            errors = append(errors, fmt.Sprintf("%s must be an even number", displayPath))
        }
    } else {
        errors = append(errors, fmt.Sprintf("%s must be a number", displayPath))
    }
    
    return errors
}

// Register the validator
RegisterCustomValidator("even", ValidateEven)
```

### Using Custom Validators in Schemas

```go
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "count": {
            Type:            "integer",
            CustomValidator: "even", // Must be an even number
        },
    },
}
```

### Built-in Custom Validators

The package includes several pre-registered custom validators:

- `nonEmpty` - String must not be empty or contain only whitespace
- `alphanumeric` - String must contain only alphanumeric characters
- `noWhitespace` - String must not contain whitespace
- `positive` - Number must be greater than zero
- `nonNegative` - Number must be greater than or equal to zero

### Enabling Custom Validation

Custom validation is disabled by default. Enable it by passing the WithCustomValidation option:

```go
validator := NewValidator(WithCustomValidation(true))
```

## Usage Example

Here's a complete example demonstrating several advanced features:

```go
package main

import (
    "fmt"
    "encoding/json"
    
    "github.com/lexlapax/go-llms/pkg/schema/domain"
    "github.com/lexlapax/go-llms/pkg/schema/validation"
)

func main() {
    // Define a schema with advanced validation features
    schema := &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "name": {
                Type:            "string",
                MinLength:       intPtr(2),
                CustomValidator: "nonEmpty",
            },
            "email": {
                Type:   "string",
                Format: "email",
            },
            "age": {
                Type:            "integer",
                Minimum:         float64Ptr(18),
                CustomValidator: "positive",
            },
            "tags": {
                Type: "array",
                MinItems:    intPtr(1),
                MaxItems:    intPtr(5),
                UniqueItems: boolPtr(true),
                Items: &domain.Property{
                    Type: "string",
                },
            },
            "metadata": {
                Type: "object",
                Properties: map[string]domain.Property{
                    "type": {Type: "string"},
                    "value": {Type: "string"},
                },
                If: &domain.Schema{
                    Properties: map[string]domain.Property{
                        "type": {
                            Enum: []string{"url"},
                        },
                    },
                },
                Then: &domain.Schema{
                    Properties: map[string]domain.Property{
                        "value": {
                            Format: "uri",
                        },
                    },
                },
            },
        },
        Required: []string{"name", "email", "age"},
    }
    
    // Sample valid data
    validData := `{
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30,
        "tags": ["user", "customer"],
        "metadata": {
            "type": "url",
            "value": "https://example.com"
        }
    }`
    
    // Create validator with all features enabled
    validator := validation.NewValidator(
        validation.WithCoercion(true),
        validation.WithCustomValidation(true),
    )
    
    // Validate the data
    result, err := validator.Validate(schema, validData)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    if result.Valid {
        fmt.Println("Data is valid!")
    } else {
        fmt.Println("Validation failed:")
        for _, err := range result.Errors {
            fmt.Printf("- %s\n", err)
        }
    }
}

// Helper functions
func intPtr(i int) *int {
    return &i
}

func float64Ptr(f float64) *float64 {
    return &f
}

func boolPtr(b bool) *bool {
    return &b
}
```

## Performance Considerations

When using advanced validation features, keep these performance considerations in mind:

1. **Custom Validators**: These add overhead, only enable when needed
2. **Conditional Validation**: Complex conditions with deeply nested schemas can impact performance
3. **OneOf/AnyOf with Many Schemas**: These require validating against multiple schemas, which multiplies validation time
4. **UniqueItems on Large Arrays**: The uniqueness check can be expensive for large arrays, especially with complex objects

The validator includes various performance optimizations:

- Object pooling for reuse of validation results and error buffers
- Fast paths for common validation scenarios
- Regex caching to avoid recompilation
- Optimized uniqueness checks with different algorithms for small vs. large arrays