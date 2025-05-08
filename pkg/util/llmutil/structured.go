package llmutil

import (
	"context"
	"fmt"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/adapter/reflection"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// StructuredResponse is a convenient wrapper for structured output operations
type StructuredResponse[T any] struct {
	Data    T         // The typed result
	Schema  *schemaDomain.Schema // The schema used for validation
	Raw     string    // The raw response from the LLM
	Metrics Metrics   // Performance metrics
}

// Metrics captures performance metrics for structured generation
type Metrics struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	LatencyMs        int
}

// GenerateTyped generates a typed response from an LLM
func GenerateTyped[T any](
	ctx context.Context,
	provider domain.Provider,
	prompt string,
	options ...domain.Option,
) (*StructuredResponse[T], error) {
	// Generate schema from type T
	var target T
	schema, err := reflection.GenerateSchema(target)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema: %w", err)
	}

	// Generate with schema
	result, err := provider.GenerateWithSchema(ctx, prompt, schema, options...)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	// Create validator and processor
	validator := validation.NewValidator()
	proc := processor.NewStructuredProcessor(validator)

	// Process the result
	var data T
	if err := proc.ProcessTyped(schema, fmt.Sprintf("%v", result), &data); err != nil {
		return nil, fmt.Errorf("failed to process response: %w", err)
	}

	// Create response object
	response := &StructuredResponse[T]{
		Data:   data,
		Schema: schema,
		Raw:    fmt.Sprintf("%v", result),
		// Metrics would be populated from the provider if available
	}

	return response, nil
}

// EnhancePromptWithExamples adds examples to a schema-based prompt
func EnhancePromptWithExamples[T any](
	prompt string,
	examples []T,
) (string, error) {
	if len(examples) == 0 {
		return prompt, nil
	}

	// Generate schema from type T
	schema, err := reflection.GenerateSchema(examples[0])
	if err != nil {
		return "", fmt.Errorf("failed to generate schema: %w", err)
	}

	// Create enhancer
	enhancer := processor.NewPromptEnhancer()

	// Convert examples to map
	exampleMaps := make([]map[string]interface{}, 0, len(examples))
	for _, example := range examples {
		// This is a simplification - in a real implementation we'd need to convert to map
		exampleMaps = append(exampleMaps, map[string]interface{}{"example": example})
	}

	// Enhance prompt with examples
	options := map[string]interface{}{
		"examples": exampleMaps,
	}

	return enhancer.EnhanceWithOptions(prompt, schema, options)
}

// SanitizeStructuredOutput performs basic validations on structured output
func SanitizeStructuredOutput[T any](data T) (T, error) {
	// Generate schema from type T
	schema, err := reflection.GenerateSchema(data)
	if err != nil {
		var empty T
		return empty, fmt.Errorf("failed to generate schema: %w", err)
	}

	// Create validator
	validator := validation.NewValidator()

	// Convert data to JSON
	proc := processor.NewStructuredProcessor(validator)
	jsonStr, err := proc.ToJSON(data)
	if err != nil {
		var empty T
		return empty, fmt.Errorf("failed to convert to JSON: %w", err)
	}

	// Validate the JSON against the schema
	validationResult, err := validator.Validate(schema, jsonStr)
	if err != nil {
		var empty T
		return empty, fmt.Errorf("validation error: %w", err)
	}

	if !validationResult.Valid {
		var empty T
		return empty, fmt.Errorf("output does not conform to schema: %v", validationResult.Errors)
	}

	return data, nil
}

// ExtractField extracts a specific field from a structured response
func ExtractField[T any, F any](data T, fieldName string) (F, error) {
	var empty F
	
	// This is a placeholder - in a real implementation, 
	// we would use reflection to extract the field from the struct
	// For now, we'll return an error
	return empty, fmt.Errorf("field extraction not implemented")
}