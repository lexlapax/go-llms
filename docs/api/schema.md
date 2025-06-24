# Schema API

JSON Schema validation - Schema definition, validation, and type coercion

## Package Information

- **Import Path**: `github.com/lexlapax/go-llms/pkg/schema`
- **Category**: Core
- **Stability**: Stable (v0.3.x)

## Overview

The Schema package implements JSON Schema validation (draft 7) with additional features for type coercion and custom validators. It's designed to work seamlessly with LLM outputs and structured data extraction.

Key features:
- Full JSON Schema draft 7 support
- Type coercion for common conversions
- Custom validator registration
- Schema composition and references
- Integration with structured output parsing
- Performance-optimized validation

## Core Types

### Schema Definition

Define schemas using the Schema type:

```go
type Schema struct {
    Type        string                 `json:"type,omitempty"`
    Properties  map[string]*Schema     `json:"properties,omitempty"`
    Required    []string               `json:"required,omitempty"`
    Title       string                 `json:"title,omitempty"`
    Description string                 `json:"description,omitempty"`
}
```

### Validation

Validate data against schemas:

```go
validator := schema.NewValidator()
err := validator.Validate(data, schemaDefinition)
if err != nil {
    // Handle validation errors
}
```

### Type Coercion

Automatic type conversion during validation:

```go
// Register custom coercion rules
schema.RegisterCoercion(reflect.TypeOf(""), reflect.TypeOf(0), func(v interface{}) (interface{}, error) {
    return strconv.Atoi(v.(string))
})
```
## Examples

### Define and Validate

```go
personSchema := &schema.Schema{
    Type: "object",
    Properties: map[string]*schema.Schema{
        "name": {Type: "string"},
        "age": {Type: "integer", Minimum: &zero},
    },
    Required: []string{"name"},
}

data := map[string]interface{}{
    "name": "John",
    "age": 30,
}

validator := schema.NewValidator()
err := validator.Validate(data, personSchema)
```
## Best Practices

1. **Define schemas upfront**: Create reusable schema definitions
2. **Use references**: Leverage $ref for schema composition
3. **Validate early**: Validate data at system boundaries
4. **Handle coercion carefully**: Be explicit about type conversions
5. **Cache validators**: Reuse compiled validators for performance
## Error Handling

Validation errors provide detailed information:

```go
err := validator.Validate(data, schema)
if err != nil {
    var validationErr *schema.ValidationError
    if errors.As(err, &validationErr) {
        for _, detail := range validationErr.Details {
            log.Printf("Error at %s: %s", detail.Path, detail.Message)
        }
    }
}
```