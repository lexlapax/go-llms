package processor

import (
	"fmt"
	"strings"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	optimizedJson "github.com/lexlapax/go-llms/pkg/util/json"
)

// EnhanceWithOptions adds schema information to a prompt with additional options - optimized version
func (p *PromptEnhancer) EnhanceWithOptions(prompt string, schema *schemaDomain.Schema, options map[string]interface{}) (string, error) {
	// Get the enhanced prompt first
	enhancedPrompt, err := p.Enhance(prompt, schema)
	if err != nil {
		return "", err
	}

	// Pre-calculate JSON for examples if present
	exampleJSONs := make([][]byte, 0)
	if examples, ok := options["examples"].([]map[string]interface{}); ok && len(examples) > 0 {
		exampleJSONs = make([][]byte, len(examples))
		for i, example := range examples {
			exampleJSON, err := optimizedJson.MarshalIndent(example, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal example %d: %w", i, err)
			}
			exampleJSONs[i] = exampleJSON
		}
	}

	// Calculate total capacity needed
	totalCapacity := len(enhancedPrompt) + 100 // Base size + buffer

	// Add space for instructions
	if instructions, ok := options["instructions"].(string); ok {
		totalCapacity += len(instructions) + 30
	}

	// Add space for format
	if format, ok := options["format"].(string); ok {
		totalCapacity += len(format) + 30
	}

	// Add space for examples
	for _, exampleJSON := range exampleJSONs {
		totalCapacity += len(exampleJSON) + 50 // JSON + formatting
	}

	// Create a new string builder with the calculated capacity
	var builder strings.Builder
	builder.Grow(totalCapacity)

	// Add the enhanced prompt
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
		builder.WriteString("Format your response as ")
		builder.WriteString(format)
		builder.WriteString("\n\n")
	}

	// Add examples if present
	if len(exampleJSONs) > 0 {
		builder.WriteString("Here are some examples of valid responses:\n\n")

		for i, exampleJSON := range exampleJSONs {
			builder.WriteString("Example ")
			builder.WriteString(fmt.Sprintf("%d", i+1))
			builder.WriteString(":\n```json\n")
			builder.Write(exampleJSON)
			builder.WriteString("\n```\n\n")
		}
	}

	return builder.String(), nil
}