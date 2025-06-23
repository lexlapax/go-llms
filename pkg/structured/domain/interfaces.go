// Package domain defines core domain models and interfaces for structured LLM outputs.
package domain

// ABOUTME: Core interfaces for structured output processing from LLMs
// ABOUTME: Defines Processor and PromptEnhancer contracts

import (
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Processor defines the contract for structured output processing.
// Implementations extract and validate structured data from LLM text responses
// according to provided JSON schemas.
type Processor interface {
	// Process processes a raw output string against a schema.
	// It extracts structured data from the LLM output and validates it
	// against the provided schema, returning the parsed result.
	Process(schema *schemaDomain.Schema, output string) (interface{}, error)

	// ProcessTyped processes a raw output string against a schema and maps it to a specific type.
	// The target parameter should be a pointer to the desired type. The extracted data
	// will be unmarshaled into this target after schema validation.
	ProcessTyped(schema *schemaDomain.Schema, output string, target interface{}) error

	// ToJSON converts an object to a JSON string.
	// This is useful for serializing structured data back to JSON format.
	ToJSON(obj interface{}) (string, error)
}

// PromptEnhancer defines the contract for enhancing prompts with schema information.
// Implementations add schema constraints and formatting instructions to prompts
// to improve the likelihood of receiving properly structured responses from LLMs.
type PromptEnhancer interface {
	// Enhance adds schema information to a prompt.
	// It augments the original prompt with instructions about the expected
	// output format based on the provided JSON schema.
	Enhance(prompt string, schema *schemaDomain.Schema) (string, error)

	// EnhanceWithOptions adds schema information to a prompt with additional options.
	// Options may include formatting preferences, example outputs, or provider-specific
	// instructions for structured output generation.
	EnhanceWithOptions(prompt string, schema *schemaDomain.Schema, options map[string]interface{}) (string, error)
}
