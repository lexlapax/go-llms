package processor

import (
	"fmt"
	"strings"
	"sync"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/domain"
	optimizedJson "github.com/lexlapax/go-llms/pkg/util/json"
)

// Global schema cache for reuse across all enhancers
var (
	globalSchemaCache *SchemaCache
	schemaCacheMutex  sync.Once
)

// getSchemaCache returns the singleton schema cache instance
func getSchemaCache() *SchemaCache {
	schemaCacheMutex.Do(func() {
		globalSchemaCache = NewSchemaCache()
	})
	return globalSchemaCache
}

// PromptEnhancer adds schema information to prompts
type PromptEnhancer struct {
	// No fields needed as we use global caches
}

// NewPromptEnhancer creates a new prompt enhancer
func NewPromptEnhancer() domain.PromptEnhancer {
	return &PromptEnhancer{}
}

// getSchemaJSON retrieves the JSON representation of a schema, using cache when possible
func getSchemaJSON(schema *schemaDomain.Schema) ([]byte, error) {
	// Get the schema cache
	cache := getSchemaCache()

	// Generate a cache key for the schema
	cacheKey := GenerateSchemaKey(schema)

	// Check cache first
	if cachedJSON, found := cache.Get(cacheKey); found {
		return cachedJSON, nil
	}

	// Cache miss - marshal the schema to JSON
	schemaJSON, err := optimizedJson.MarshalSchemaIndent(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Store in cache for future use
	cache.Set(cacheKey, schemaJSON)

	return schemaJSON, nil
}

// Enhance adds schema information to a prompt - optimized version
func (p *PromptEnhancer) Enhance(prompt string, schema *schemaDomain.Schema) (string, error) {
	// Get schema JSON using cache
	schemaJSON, err := getSchemaJSON(schema)
	if err != nil {
		return "", err
	}

	// Calculate initial capacity based on input sizes to reduce allocations
	// Start with prompt length + schema JSON length + standard text (~500 bytes)
	initialCapacity := len(prompt) + len(schemaJSON) + 500

	// Account for property descriptions (est. ~50 bytes per property)
	if schema.Type == "object" {
		initialCapacity += len(schema.Properties) * 50
	}

	// Build enhanced prompt with capacity pre-allocation
	var enhancedPrompt strings.Builder
	enhancedPrompt.Grow(initialCapacity)

	// Add the base prompt
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
			// Pre-join required fields to reduce allocations
			requiredFields := strings.Join(schema.Required, ", ")
			enhancedPrompt.WriteString("3. The required fields are: ")
			enhancedPrompt.WriteString(requiredFields)
			enhancedPrompt.WriteString(".\n")
		}

		// Add descriptions for properties if available
		if len(schema.Properties) > 0 {
			enhancedPrompt.WriteString("4. Field descriptions:\n")

			// Fast path: only process properties with descriptions
			hasDescriptions := false
			for _, prop := range schema.Properties {
				if prop.Description != "" {
					hasDescriptions = true
					break
				}
			}

			if hasDescriptions {
				for name, prop := range schema.Properties {
					if prop.Description != "" {
						enhancedPrompt.WriteString("   - ")
						enhancedPrompt.WriteString(name)
						enhancedPrompt.WriteString(": ")
						enhancedPrompt.WriteString(prop.Description)
						enhancedPrompt.WriteString("\n")
					}
				}
			}
		}

		// Add enum values if available
		for name, prop := range schema.Properties {
			if len(prop.Enum) > 0 {
				// Pre-join enum values to reduce allocations
				enumValues := strings.Join(prop.Enum, ", ")
				enhancedPrompt.WriteString("   - ")
				enhancedPrompt.WriteString(name)
				enhancedPrompt.WriteString(" must be one of: ")
				enhancedPrompt.WriteString(enumValues)
				enhancedPrompt.WriteString("\n")
			}
		}

	case "array":
		enhancedPrompt.WriteString("3. Format your response as a JSON array of items.\n")
		if schema.Properties != nil && schema.Properties[""].Items != nil {
			itemType := schema.Properties[""].Items.Type
			enhancedPrompt.WriteString("4. Each item should be a ")
			enhancedPrompt.WriteString(itemType)
			enhancedPrompt.WriteString(".\n")
		}
	}

	return enhancedPrompt.String(), nil
}

// EnhanceWithOptions adds schema information to a prompt with additional options - optimized version
func (p *PromptEnhancer) EnhanceWithOptions(prompt string, schema *schemaDomain.Schema, options map[string]interface{}) (string, error) {
	enhancedPrompt, err := p.Enhance(prompt, schema)
	if err != nil {
		return "", err
	}

	// Calculate additional capacity needed for options
	additionalCapacity := len(enhancedPrompt) + 100 // Base size + buffer

	// Estimate space for instructions
	if instructions, ok := options["instructions"].(string); ok {
		additionalCapacity += len(instructions) + 30
	}

	// Estimate space for format
	if format, ok := options["format"].(string); ok {
		additionalCapacity += len(format) + 30
	}

	// Estimate space for examples
	if examples, ok := options["examples"].([]map[string]interface{}); ok {
		// Rough estimate: 200 bytes per example
		additionalCapacity += len(examples) * 200
	}

	var builder strings.Builder
	builder.Grow(additionalCapacity)
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

	// Add examples if provided
	if examples, ok := options["examples"].([]map[string]interface{}); ok && len(examples) > 0 {
		builder.WriteString("Here are some examples of valid responses:\n\n")

		for i, example := range examples {
			// Try to reuse the JSON marshaling if possible (future optimization)
			exampleJSON, err := optimizedJson.MarshalIndent(example, "", "  ")
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

// EnhancePromptWithSchema is a convenience function that enhances a prompt with schema information
// Optimized to use the singleton enhancer instance for better performance
func EnhancePromptWithSchema(prompt string, schema *schemaDomain.Schema) (string, error) {
	// Use the default enhancer singleton instead of creating a new one each time
	enhancer := GetDefaultPromptEnhancer()
	return enhancer.Enhance(prompt, schema)
}
