// Package docs provides documentation generation capabilities for go-llms components.
// It supports generating documentation in multiple formats including OpenAPI, Markdown,
// and JSON. The package provides interfaces and types that are bridge-friendly for
// integration with other systems.
package docs

// ABOUTME: Defines core interfaces for documentation generation including Generator and Documentable
// ABOUTME: Supports multiple output formats (OpenAPI, Markdown, JSON) with bridge-friendly types

import (
	"context"
	"encoding/json"
)

// Generator defines the interface for documentation generators.
// Implementations can generate documentation in various formats from
// a collection of Documentable items.
type Generator interface {
	// GenerateOpenAPI generates OpenAPI 3.0 specification
	GenerateOpenAPI(ctx context.Context, items []Documentable) (*OpenAPISpec, error)

	// GenerateMarkdown generates Markdown documentation
	GenerateMarkdown(ctx context.Context, items []Documentable) (string, error)

	// GenerateJSON generates JSON documentation
	GenerateJSON(ctx context.Context, items []Documentable) ([]byte, error)
}

// Documentable represents an item that can be documented.
// Any component that implements this interface can have its
// documentation automatically generated.
type Documentable interface {
	// GetDocumentation returns the documentation for this item
	GetDocumentation() Documentation
}

// Documentation contains all documentation details for an item.
// It provides comprehensive information including descriptions,
// examples, schemas, and metadata for documentation generation.
type Documentation struct {
	// Basic information
	Name        string `json:"name"`        // Name of the component
	Description string `json:"description"` // Brief description

	// Extended information
	LongDescription string   `json:"longDescription,omitempty"` // Detailed description
	Category        string   `json:"category,omitempty"`        // Category for grouping
	Tags            []string `json:"tags,omitempty"`            // Tags for discovery
	Version         string   `json:"version,omitempty"`         // Version information
	Deprecated      bool     `json:"deprecated,omitempty"`      // Deprecation status
	DeprecationNote string   `json:"deprecationNote,omitempty"` // Deprecation details

	// Usage information
	Examples []Example `json:"examples,omitempty"` // Usage examples

	// Schema information
	Schema  *Schema            `json:"schema,omitempty"`  // Input/output schema
	Schemas map[string]*Schema `json:"schemas,omitempty"` // Multiple schemas (e.g., input/output)

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Additional metadata
}

// Example represents a usage example.
// Examples demonstrate how to use a component with concrete inputs,
// expected outputs, and code snippets. They are essential for
// helping users understand practical usage patterns.
type Example struct {
	Name        string      `json:"name"`                  // Example name
	Description string      `json:"description,omitempty"` // What this example shows
	Input       interface{} `json:"input,omitempty"`       // Example input
	Output      interface{} `json:"output,omitempty"`      // Expected output
	Code        string      `json:"code,omitempty"`        // Code snippet
	Language    string      `json:"language,omitempty"`    // Code language
}

// Schema represents a JSON schema (bridge-friendly).
// It provides a subset of JSON Schema Draft 7 for describing
// data structures, validation rules, and type constraints.
// This type is designed to be easily serializable and compatible
// with various documentation formats.
type Schema struct {
	Type        string                 `json:"type,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]*Schema     `json:"properties,omitempty"`
	Items       *Schema                `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Enum        []interface{}          `json:"enum,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	MinLength   *int                   `json:"minLength,omitempty"`
	MaxLength   *int                   `json:"maxLength,omitempty"`
	Minimum     *float64               `json:"minimum,omitempty"`
	Maximum     *float64               `json:"maximum,omitempty"`
	Additional  map[string]interface{} `json:"additionalProperties,omitempty"`
}

// MarshalJSON ensures Schema is JSON serializable.
// This custom marshaler handles the Schema type's serialization
// to JSON format, preserving all schema properties correctly.
//
// Returns the JSON representation or an error if marshaling fails.
func (s *Schema) MarshalJSON() ([]byte, error) {
	type Alias Schema
	return json.Marshal((*Alias)(s))
}

// UnmarshalJSON ensures Schema can be deserialized from JSON.
// This custom unmarshaler handles the Schema type's deserialization
// from JSON format, properly reconstructing all schema properties.
//
// Parameters:
//   - data: The JSON data to unmarshal
//
// Returns an error if unmarshaling fails.
func (s *Schema) UnmarshalJSON(data []byte) error {
	type Alias Schema
	return json.Unmarshal(data, (*Alias)(s))
}
