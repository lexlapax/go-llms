// Package domain defines core domain models and interfaces for structured LLM outputs.
package domain

import (
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Processor defines the contract for structured output processing
type Processor interface {
	// Process processes a raw output string against a schema
	Process(schema *schemaDomain.Schema, output string) (interface{}, error)

	// ProcessTyped processes a raw output string against a schema and maps it to a specific type
	ProcessTyped(schema *schemaDomain.Schema, output string, target interface{}) error
}

// PromptEnhancer defines the contract for enhancing prompts with schema information
type PromptEnhancer interface {
	// Enhance adds schema information to a prompt
	Enhance(prompt string, schema *schemaDomain.Schema) (string, error)

	// EnhanceWithOptions adds schema information to a prompt with additional options
	EnhanceWithOptions(prompt string, schema *schemaDomain.Schema, options map[string]interface{}) (string, error)
}
