// Package domain defines types and interfaces for JSON schema validation and generation.
//
// This package provides the core abstractions for working with JSON schemas,
// including validation, storage, and generation from Go types.
//
// # Core Types
//
// Schema: Represents a JSON Schema with properties, constraints, and validation rules.
// It follows the JSON Schema specification and supports common validation keywords.
//
// Property: Represents a single property within a schema, including its type,
// constraints, and nested properties for objects and arrays.
//
// ValidationResult: Contains the results of validating data against a schema,
// including any errors and their locations.
//
// # Interfaces
//
// Validator: Interface for validating JSON data against schemas.
// Implementations handle the actual validation logic and error reporting.
//
// SchemaRepository: Interface for storing and retrieving schemas.
// Supports loading schemas from files, URLs, or other sources.
//
// SchemaGenerator: Interface for generating schemas from Go types.
// Uses reflection or other methods to create schemas automatically.
//
// # Schema Types
//
// The package supports all JSON Schema primitive types:
//   - string: Text values with format and pattern validation
//   - number/integer: Numeric values with range constraints
//   - boolean: True/false values
//   - array: Lists with item schemas and length constraints
//   - object: Structured data with property schemas
//   - null: Null values
//
// # Validation Features
//
// Schemas support various validation keywords:
//   - Type checking and coercion
//   - Required fields
//   - Pattern matching for strings
//   - Minimum/maximum for numbers
//   - Array length and uniqueness
//   - Object property constraints
//   - Conditional validation (if/then/else)
//   - Schema composition (allOf, anyOf, oneOf)
//
// # Usage Example
//
//	// Define a schema
//	schema := &domain.Schema{
//	    Type: "object",
//	    Properties: map[string]domain.Property{
//	        "name": {Type: "string", MinLength: ptr(1)},
//	        "age":  {Type: "integer", Minimum: ptr(0.0)},
//	    },
//	    Required: []string{"name"},
//	}
//
//	// Validate data
//	validator := validation.NewValidator()
//	result := validator.Validate(schema, data)
//	if !result.Valid {
//	    // Handle validation errors
//	}
package domain
