package processor

// ABOUTME: Schema-aware prompt enhancement with caching and templates
// ABOUTME: Generates prompts that guide LLMs to produce valid JSON output

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
	// Enhance transforms a user prompt by appending JSON schema instructions
	// to guide LLMs in producing structured outputs. The function:
	// 1. Converts the schema to JSON (with caching for performance)
	// 2. Builds an enhanced prompt with clear formatting instructions
	// 3. Adds schema-specific guidance based on type (object, array, etc.)
	// 4. Includes examples for complex schemas
	// 5. Uses efficient string building to minimize allocations

	// Get schema JSON using cache to avoid repeated marshaling
	schemaJSON, err := getSchemaJSON(schema)
	if err != nil {
		return "", err
	}

	// Create optimized string builder with better capacity estimation
	enhancedPrompt := NewSchemaPromptBuilder(prompt, schema, len(schemaJSON))

	// Add the base prompt
	if _, err := enhancedPrompt.WriteString(prompt); err != nil {
		return "", fmt.Errorf("failed to write prompt: %w", err)
	}
	if _, err := enhancedPrompt.WriteString("\n\n"); err != nil {
		return "", fmt.Errorf("failed to write newline: %w", err)
	}
	if _, err := enhancedPrompt.WriteString("Please provide your response as a valid JSON object that conforms to the following JSON schema:\n\n"); err != nil {
		return "", fmt.Errorf("failed to write instructions: %w", err)
	}
	if _, err := enhancedPrompt.WriteString("```json\n"); err != nil {
		return "", fmt.Errorf("failed to write code block start: %w", err)
	}
	if _, err := enhancedPrompt.Write(schemaJSON); err != nil {
		return "", fmt.Errorf("failed to write schema JSON: %w", err)
	}
	if _, err := enhancedPrompt.WriteString("\n```\n\n"); err != nil {
		return "", fmt.Errorf("failed to write code block end: %w", err)
	}

	// Add requirements for the output
	if _, err := enhancedPrompt.WriteString("Your response must be valid JSON only, following these guidelines:\n"); err != nil {
		return "", fmt.Errorf("failed to write guidelines header: %w", err)
	}
	if _, err := enhancedPrompt.WriteString("1. Do not include any explanations, markdown code blocks, or additional text before or after the JSON.\n"); err != nil {
		return "", fmt.Errorf("failed to write guideline 1: %w", err)
	}
	if _, err := enhancedPrompt.WriteString("2. Ensure all required fields are included.\n"); err != nil {
		return "", fmt.Errorf("failed to write guideline 2: %w", err)
	}

	// Add type-specific instructions based on schema type
	// Different types require different guidance for LLMs
	switch schema.Type {
	case "object":
		// For objects, emphasize required fields and provide field descriptions
		if len(schema.Required) > 0 {
			// Pre-join required fields to reduce allocations
			requiredFields := strings.Join(schema.Required, ", ")
			if _, err := enhancedPrompt.WriteString("3. The required fields are: "); err != nil {
				return "", fmt.Errorf("failed to write required fields prefix: %w", err)
			}
			if _, err := enhancedPrompt.WriteString(requiredFields); err != nil {
				return "", fmt.Errorf("failed to write required fields list: %w", err)
			}
			if _, err := enhancedPrompt.WriteString(".\n"); err != nil {
				return "", fmt.Errorf("failed to write required fields suffix: %w", err)
			}
		}

		// Add descriptions for properties if available
		// Property descriptions help LLMs understand the expected content
		if len(schema.Properties) > 0 {
			if _, err := enhancedPrompt.WriteString("4. Field descriptions:\n"); err != nil {
				return "", fmt.Errorf("failed to write field descriptions header: %w", err)
			}

			// Fast path optimization: only process properties with descriptions
			// This avoids iterating through properties twice in most cases
			hasDescriptions := false
			for _, prop := range schema.Properties {
				if prop.Description != "" {
					hasDescriptions = true
					break
				}
			}

			if hasDescriptions {
				// Write field descriptions in a clear format
				for name, prop := range schema.Properties {
					if prop.Description != "" {
						if _, err := enhancedPrompt.WriteString("   - "); err != nil {
							return "", fmt.Errorf("failed to write description indent: %w", err)
						}
						if _, err := enhancedPrompt.WriteString(name); err != nil {
							return "", fmt.Errorf("failed to write property name: %w", err)
						}
						if _, err := enhancedPrompt.WriteString(": "); err != nil {
							return "", fmt.Errorf("failed to write separator: %w", err)
						}
						if _, err := enhancedPrompt.WriteString(prop.Description); err != nil {
							return "", fmt.Errorf("failed to write property description: %w", err)
						}
						if _, err := enhancedPrompt.WriteString("\n"); err != nil {
							return "", fmt.Errorf("failed to write newline: %w", err)
						}
					}
				}
			}
		}

		// Add enum constraints if available
		// Enums restrict fields to specific values, which is crucial for LLMs
		for name, prop := range schema.Properties {
			if len(prop.Enum) > 0 {
				// Pre-join enum values to reduce allocations
				enumValues := strings.Join(prop.Enum, ", ")
				if _, err := enhancedPrompt.WriteString("   - "); err != nil {
					return "", fmt.Errorf("failed to write enum indent: %w", err)
				}
				if _, err := enhancedPrompt.WriteString(name); err != nil {
					return "", fmt.Errorf("failed to write enum property name: %w", err)
				}
				if _, err := enhancedPrompt.WriteString(" must be one of: "); err != nil {
					return "", fmt.Errorf("failed to write enum prefix: %w", err)
				}
				if _, err := enhancedPrompt.WriteString(enumValues); err != nil {
					return "", fmt.Errorf("failed to write enum values: %w", err)
				}
				if _, err := enhancedPrompt.WriteString("\n"); err != nil {
					return "", fmt.Errorf("failed to write enum newline: %w", err)
				}
			}
		}

	case "array":
		// Array type requires different instructions
		if _, err := enhancedPrompt.WriteString("3. Format your response as a JSON array of items.\n"); err != nil {
			return "", fmt.Errorf("failed to write array format instruction: %w", err)
		}
		if schema.Properties != nil && schema.Properties[""].Items != nil {
			itemType := schema.Properties[""].Items.Type
			if _, err := enhancedPrompt.WriteString("4. Each item should be a "); err != nil {
				return "", fmt.Errorf("failed to write item type prefix: %w", err)
			}
			if _, err := enhancedPrompt.WriteString(itemType); err != nil {
				return "", fmt.Errorf("failed to write item type: %w", err)
			}
			if _, err := enhancedPrompt.WriteString(".\n"); err != nil {
				return "", fmt.Errorf("failed to write item type suffix: %w", err)
			}
		}
	}

	return enhancedPrompt.String(), nil
}

// EnhancePromptWithSchema is a convenience function that enhances a prompt with schema information
// Optimized to use the singleton enhancer instance for better performance
func EnhancePromptWithSchema(prompt string, schema *schemaDomain.Schema) (string, error) {
	// Use the default enhancer singleton instead of creating a new one each time
	enhancer := GetDefaultPromptEnhancer()
	return enhancer.Enhance(prompt, schema)
}
