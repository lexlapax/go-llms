// ABOUTME: Type definitions for structured output parsing and validation
// ABOUTME: Provides schema types for LLM output validation

package outputs

// Type represents schema types for LLM outputs.
// It defines the basic data types that can be used in output schemas
// for structured generation and validation.
type Type string

const (
	// TypeString represents string type
	TypeString Type = "string"
	// TypeNumber represents number type
	TypeNumber Type = "number"
	// TypeInteger represents integer type
	TypeInteger Type = "integer"
	// TypeBoolean represents boolean type
	TypeBoolean Type = "boolean"
	// TypeArray represents array type
	TypeArray Type = "array"
	// TypeObject represents object type
	TypeObject Type = "object"
	// TypeNull represents null type
	TypeNull Type = "null"
)

// OutputSchema is a standalone schema for LLM outputs.
// It provides a comprehensive structure for defining expected output formats,
// including type constraints, validation rules, and nested object definitions.
// This schema is used to validate and parse LLM responses into structured data.
type OutputSchema struct {
	// Type is the schema type
	Type Type `json:"type"`

	// Description of the schema
	Description string `json:"description,omitempty"`

	// Format for string types (email, date-time, etc.)
	Format string `json:"format,omitempty"`

	// Pattern for string validation
	Pattern string `json:"pattern,omitempty"`

	// Enum for allowed values
	Enum []string `json:"enum,omitempty"`

	// Number constraints
	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`

	// Array constraints
	MinItems *int          `json:"minItems,omitempty"`
	MaxItems *int          `json:"maxItems,omitempty"`
	Items    *OutputSchema `json:"items,omitempty"`

	// Object properties
	Properties           map[string]*OutputSchema `json:"properties,omitempty"`
	RequiredProperties   []string                 `json:"required,omitempty"`
	AdditionalProperties *bool                    `json:"additionalProperties,omitempty"`

	// Required flag for individual properties
	Required *bool `json:"isRequired,omitempty"`
}
