# Schema Generation Example

This example demonstrates how to automatically generate JSON schemas from Go structs using Go-LLMs reflection capabilities. It showcases a key feature of Go-LLMs that enables type-safe interactions with LLMs by generating JSON schemas that LLMs can use to structure their outputs.

## Overview

Schema generation is a core feature that enables Go-LLMs to provide type safety when working with LLM outputs. This example:

1. Defines several Go structs representing a common e-commerce domain model
2. Demonstrates automatic generation of JSON schemas from these structs
3. Shows how the generated schemas can be used with LLMs
4. Highlights advanced features like nested objects, arrays, and validation constraints

## Domain Model

The example includes several linked domain entities:

- `Address`: A physical mailing address
- `Customer`: A user of the e-commerce system
- `Product`: An item available for purchase
- `ProductReview`: A customer review of a product
- `OrderItem`: A line item within an order
- `Order`: A purchase made by a customer

Each entity is defined as a Go struct with appropriate JSON tags, validation constraints, and field descriptions.

## Schema Generation Features

The example demonstrates these schema generation features:

### Basic Types

- String, integer, number (float), and boolean Go types
- Time.Time handling with appropriate formats

### Validation Constraints

- Required fields using `validate:"required"` tag
- String pattern validation with regex patterns
- Minimum/maximum value constraints for numbers
- Minimum/maximum length constraints for strings

### Nested Objects

- Complex object properties like Order.ShippingAddress
- Handling of pointers to objects

### Arrays and Slices

- Handling arrays of primitive types (string, int, etc.)
- Arrays of complex objects (e.g., Product.Reviews)

### Enums

- Enumeration values using `validate:"oneof=..."` tag
- Constants defined as Go types for type safety

## Using the Generated Schemas

The example shows how to integrate generated schemas with structured output processing:

1. Define your domain model as Go structs with proper tags
2. Generate schemas using `reflection.GenerateSchema()`
3. Use the schemas with the structured output processor
4. Process LLM responses directly into typed Go structs

This enables end-to-end type safety from your domain model to the LLM output.

## Running the Example

Build and run the example:

```bash
make example EXAMPLE=schema
./bin/schema
```

This will:
1. Generate schemas for all the domain entities
2. Display the schemas in the console
3. Save the schemas as JSON files in the `schemas` directory
4. Demonstrate how to use the schemas with an LLM

## Performance Considerations

Schema generation uses reflection, which can be computationally expensive. For optimal performance:

1. Generate schemas once during application initialization
2. Cache the generated schemas for reuse
3. Consider using `processor.SetSchemaCache()` for efficient schema caching

The example includes performance tests that measure schema generation speed.

## Best Practices

When defining structs for schema generation:

1. Use descriptive field names and consistent JSON naming conventions
2. Add `description` tags to provide context for the LLM
3. Mark required fields with `validate:"required"` 
4. Use appropriate validation constraints
5. Define enums using Go constants and `validate:"oneof=..."` tags
6. Use pointers for optional complex objects

## Integration with LLMs

The generated schemas integrate seamlessly with Go-LLMs' structured output processing:

```go
// Define your domain model
type Product struct {
    ID    string  `json:"id" validate:"required" description:"Product identifier"`
    Name  string  `json:"name" validate:"required" description:"Product name"`
    Price float64 `json:"price" validate:"min=0" description:"Product price"`
}

// Generate schema (do this once and cache)
schema, _ := reflection.GenerateSchema(Product{})

// Create processor with your LLM provider
processor := processor.NewProcessor(llmProvider)

// Process LLM response directly into a typed struct
var product Product
err := processor.ProcessTyped(ctx, prompt, &product)
```

This enables you to work with LLM outputs as native Go structs with full type safety.