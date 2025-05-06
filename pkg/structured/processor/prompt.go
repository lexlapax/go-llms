package processor

import (
	"encoding/json"
	"fmt"
	"strings"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/domain"
)

// PromptEnhancer adds schema information to prompts
type PromptEnhancer struct{}

// NewPromptEnhancer creates a new prompt enhancer
func NewPromptEnhancer() domain.PromptEnhancer {
	return &PromptEnhancer{}
}

// Enhance adds schema information to a prompt
func (p *PromptEnhancer) Enhance(prompt string, schema *schemaDomain.Schema) (string, error) {
	// Convert schema to JSON
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Build enhanced prompt
	var enhancedPrompt strings.Builder
	enhancedPrompt.WriteString(prompt)
	enhancedPrompt.WriteString("\n\n")
	enhancedPrompt.WriteString("Please provide your response as a valid JSON object that conforms to the following JSON schema:\n\n")
	enhancedPrompt.WriteString("```json\n")
	enhancedPrompt.Write(schemaJSON)
	enhancedPrompt.WriteString("\n```\n\n")

	// Add requirements for the output
	enhancedPrompt.WriteString("Your response must be valid JSON only, following these guidelines:\n")
	enhancedPrompt.WriteString("1. Do not include any explanations, markdown code blocks, or additional text before or after the JSON.\n")
	enhancedPrompt.WriteString("2. Ensure all required fields are included.\n")

	// Add type-specific instructions
	switch schema.Type {
	case "object":
		if len(schema.Required) > 0 {
			enhancedPrompt.WriteString(fmt.Sprintf("3. The required fields are: %s.\n", strings.Join(schema.Required, ", ")))
		}

		// Add descriptions for properties if available
		enhancedPrompt.WriteString("4. Field descriptions:\n")
		for name, prop := range schema.Properties {
			if prop.Description != "" {
				enhancedPrompt.WriteString(fmt.Sprintf("   - %s: %s\n", name, prop.Description))
			}
		}

		// Add enum values if available
		for name, prop := range schema.Properties {
			if len(prop.Enum) > 0 {
				enhancedPrompt.WriteString(fmt.Sprintf("   - %s must be one of: %s\n", name, strings.Join(prop.Enum, ", ")))
			}
		}

	case "array":
		enhancedPrompt.WriteString("3. Format your response as a JSON array of items.\n")
		if schema.Properties != nil && schema.Properties[""].Items != nil {
			itemType := schema.Properties[""].Items.Type
			enhancedPrompt.WriteString(fmt.Sprintf("4. Each item should be a %s.\n", itemType))
		}
	}

	return enhancedPrompt.String(), nil
}

// EnhanceWithOptions adds schema information to a prompt with additional options
func (p *PromptEnhancer) EnhanceWithOptions(prompt string, schema *schemaDomain.Schema, options map[string]interface{}) (string, error) {
	enhancedPrompt, err := p.Enhance(prompt, schema)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(enhancedPrompt)
	builder.WriteString("\n")

	// Add custom instructions if provided
	if instructions, ok := options["instructions"].(string); ok {
		builder.WriteString("Additional instructions: ")
		builder.WriteString(instructions)
		builder.WriteString("\n\n")
	}

	// Add format requirements if provided
	if format, ok := options["format"].(string); ok {
		builder.WriteString(fmt.Sprintf("Format your response as %s\n\n", format))
	}

	// Add examples if provided
	if examples, ok := options["examples"].([]map[string]interface{}); ok && len(examples) > 0 {
		builder.WriteString("Here are some examples of valid responses:\n\n")

		for i, example := range examples {
			exampleJSON, err := json.MarshalIndent(example, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal example %d: %w", i, err)
			}

			builder.WriteString("Example ")
			builder.WriteString(fmt.Sprintf("%d", i+1))
			builder.WriteString(":\n```json\n")
			builder.Write(exampleJSON)
			builder.WriteString("\n```\n\n")
		}
	}

	return builder.String(), nil
}
