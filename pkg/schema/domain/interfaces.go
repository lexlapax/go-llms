// Package domain defines the core domain models and interfaces for schema validation.
package domain

// Schema represents a validation schema for structured data
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

// Property represents a property in a schema
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

// ValidationResult represents the outcome of a validation
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
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

// SchemaGenerator generates JSON schema from Go types
type SchemaGenerator interface {
	// GenerateSchema generates a JSON schema from a Go type
	GenerateSchema(obj interface{}) (*Schema, error)
}
