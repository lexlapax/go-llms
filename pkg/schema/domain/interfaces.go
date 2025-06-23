// Package domain defines the core domain models and interfaces for schema validation.
package domain

// ABOUTME: Core schema domain models and interfaces for JSON validation
// ABOUTME: Defines Schema, Property, and validation contract types

// Schema represents a validation schema for structured data.
// It follows JSON Schema specifications and supports complex validation rules
// including conditional validation, composition, and nested schemas.
type Schema struct {
	Type                 string              `json:"type"`
	Properties           map[string]Property `json:"properties,omitempty"`
	Required             []string            `json:"required,omitempty"`
	AdditionalProperties *bool               `json:"additionalProperties,omitempty"`
	Description          string              `json:"description,omitempty"`
	Title                string              `json:"title,omitempty"`

	// Conditional validation
	If    *Schema   `json:"if,omitempty"`
	Then  *Schema   `json:"then,omitempty"`
	Else  *Schema   `json:"else,omitempty"`
	AllOf []*Schema `json:"allOf,omitempty"`
	AnyOf []*Schema `json:"anyOf,omitempty"`
	OneOf []*Schema `json:"oneOf,omitempty"`
	Not   *Schema   `json:"not,omitempty"`
}

// Property represents a property in a schema.
// It defines validation constraints for individual fields including type,
// format, ranges, patterns, and nested object structures.
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

	// Conditional validation (added to support AnyOf, OneOf, Not in properties)
	AnyOf []*Schema `json:"anyOf,omitempty"`
	OneOf []*Schema `json:"oneOf,omitempty"`
	Not   *Schema   `json:"not,omitempty"`
}

// ValidationResult represents the outcome of a validation.
// It includes a boolean status and detailed error messages for any
// validation failures encountered.
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// Validator defines the contract for schema validation.
// Implementations validate data against JSON schemas and provide
// detailed error reporting for validation failures.
type Validator interface {
	// Validate checks if data conforms to the schema.
	// The data parameter should be a JSON string. Returns validation results
	// with detailed error messages if validation fails.
	Validate(schema *Schema, data string) (*ValidationResult, error)

	// ValidateStruct validates a Go struct against a schema.
	// This method marshals the struct to JSON internally before validation.
	// Useful for validating Go objects before serialization.
	ValidateStruct(schema *Schema, obj interface{}) (*ValidationResult, error)
}

// SchemaRepository defines storage operations for schemas.
// Implementations provide persistent storage and retrieval of
// schema definitions for reuse across the application.
type SchemaRepository interface {
	// Get retrieves a schema by ID.
	// Returns an error if the schema doesn't exist.
	Get(id string) (*Schema, error)

	// Save stores a schema with the given ID.
	// Overwrites any existing schema with the same ID.
	Save(id string, schema *Schema) error

	// Delete removes a schema by ID.
	// Returns an error if the schema doesn't exist.
	Delete(id string) error
}

// SchemaGenerator generates JSON schemas from Go types.
// Implementations use reflection to analyze Go types and produce
// corresponding JSON schema definitions.
type SchemaGenerator interface {
	// GenerateSchema generates a JSON schema from a Go type.
	// The obj parameter should be an instance or pointer to the type.
	// Struct tags can be used to customize the generated schema.
	GenerateSchema(obj interface{}) (*Schema, error)
}
