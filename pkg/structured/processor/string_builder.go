package processor

import (
	"strings"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// OptimizedStringBuilder provides an enhanced string builder with better capacity estimation
// for complex schemas and prompt enhancement operations
type OptimizedStringBuilder struct {
	sb strings.Builder
}

// EstimateSchemaCapacity calculates a more accurate initial capacity for a schema
// based on its complexity and structure, returning the estimated capacity in bytes
func EstimateSchemaCapacity(schema *schemaDomain.Schema, prompt string, includeSchemaJSON bool, schemaJSONLength int) int {
	// Base capacity starts with the prompt length and standard boilerplate text
	capacity := len(prompt) + 500 // Base text for prompt enhancement

	// If we're including the schema JSON, add its length plus formatting
	if includeSchemaJSON {
		capacity += schemaJSONLength + 50 // JSON + markdown formatting
	}

	// Add space for schema type and basic schema info
	capacity += 50 // "Type: object", etc.

	// If it's an object schema, calculate property space more accurately
	if schema.Type == "object" {
		// Space for required fields list
		if len(schema.Required) > 0 {
			// Each required field name plus formatting
			capacity += len(strings.Join(schema.Required, ", ")) + 30
		}

		// Space for properties section
		if len(schema.Properties) > 0 {
			// Header for properties section
			capacity += 30 // "Properties:" + newlines

			// For each property, estimate the space needed
			for name, prop := range schema.Properties {
				// Property name, type and basic formatting
				propSize := len(name) + len(prop.Type) + 20

				// Add space for description if present
				if prop.Description != "" {
					propSize += len(prop.Description) + 10
				}

				// Add space for validations like min/max values
				if prop.Minimum != nil || prop.Maximum != nil {
					propSize += 30
				}

				// Add space for enum values
				if len(prop.Enum) > 0 {
					propSize += len(strings.Join(prop.Enum, ", ")) + 30
				}

				// If this property is an object or array, add space for nested properties
				if prop.Properties != nil || prop.Items != nil {
					propSize += 100 // Additional space for nested structure

					// If it has nested properties, add more space
					if prop.Properties != nil {
						propSize += len(prop.Properties) * 50
					}

					// If it has array items, add space for item description
					if prop.Items != nil {
						propSize += 50
						if prop.Items.Properties != nil {
							propSize += len(prop.Items.Properties) * 50
						}
					}
				}

				capacity += propSize
			}
		}
	} else if schema.Type == "array" {
		// For array schemas, add space for item description
		capacity += 100

		// If it has items, add space for item details
		if schema.Properties != nil && schema.Properties[""].Items != nil {
			capacity += 100
		}
	}

	// Add space for conditional validation (if/then/else)
	if schema.If != nil || schema.Then != nil || schema.Else != nil {
		capacity += 300
	}

	// Add space for anyOf, oneOf, allOf
	if len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 || len(schema.AllOf) > 0 {
		capacity += 300
	}

	// Ensure we have enough capacity for very complex schemas
	if len(schema.Properties) > 20 {
		capacity += 1000 // Add extra buffer for very complex schemas
	}

	return capacity
}

// NewOptimizedBuilder creates a new OptimizedStringBuilder with the given initial capacity
func NewOptimizedBuilder(initialCapacity int) *OptimizedStringBuilder {
	b := &OptimizedStringBuilder{}
	b.sb.Grow(initialCapacity)
	return b
}

// NewSchemaPromptBuilder creates an OptimizedStringBuilder with capacity optimized for a schema prompt
func NewSchemaPromptBuilder(prompt string, schema *schemaDomain.Schema, schemaJSONLength int) *OptimizedStringBuilder {
	capacity := EstimateSchemaCapacity(schema, prompt, true, schemaJSONLength)
	return NewOptimizedBuilder(capacity)
}

// NewSchemaWithOptionsBuilder creates an OptimizedStringBuilder with capacity optimized for a schema prompt with options
func NewSchemaWithOptionsBuilder(basePrompt string, options map[string]interface{}) *OptimizedStringBuilder {
	// Start with base prompt length
	capacity := len(basePrompt) + 100

	// Add space for instructions
	if instructions, ok := options["instructions"].(string); ok {
		capacity += len(instructions) + 50
	}

	// Add space for format
	if format, ok := options["format"].(string); ok {
		capacity += len(format) + 50
	}

	// Add space for examples
	if examples, ok := options["examples"].([]map[string]interface{}); ok {
		// More accurate estimation based on number and size of examples
		examplesEstimate := 200 // Base examples header

		// For each example, estimate the space more accurately
		for _, example := range examples {
			// Start with base example formatting
			exampleSize := 100 // Example number, code block formatting, etc.

			// Roughly estimate the serialized JSON size
			// Each key-value pair takes ~20-30 bytes plus the actual data
			exampleSize += len(example) * 50

			examplesEstimate += exampleSize
		}

		capacity += examplesEstimate
	}

	return NewOptimizedBuilder(capacity)
}

// WriteString adds a string to the builder
func (b *OptimizedStringBuilder) WriteString(s string) (int, error) {
	return b.sb.WriteString(s)
}

// Write adds a byte slice to the builder
func (b *OptimizedStringBuilder) Write(p []byte) (int, error) {
	return b.sb.Write(p)
}

// String returns the built string
func (b *OptimizedStringBuilder) String() string {
	return b.sb.String()
}

// Len returns the current length of the content
func (b *OptimizedStringBuilder) Len() int {
	return b.sb.Len()
}

// Cap returns the current capacity of the builder
func (b *OptimizedStringBuilder) Cap() int {
	return b.sb.Cap()
}

// Reset resets the builder to empty
func (b *OptimizedStringBuilder) Reset() {
	b.sb.Reset()
}

// Grow increases the builder's capacity
func (b *OptimizedStringBuilder) Grow(n int) {
	b.sb.Grow(n)
}
