// Package processor implements structured output processing functionality
package processor

import (
	"encoding/json"
	"fmt"
	"strings"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/domain"
)

// StructuredProcessor handles processing of structured LLM outputs
type StructuredProcessor struct {
	validator schemaDomain.Validator
}

// NewStructuredProcessor creates a new structured processor
func NewStructuredProcessor(validator schemaDomain.Validator) domain.Processor {
	return &StructuredProcessor{
		validator: validator,
	}
}

// Process processes a raw output string against a schema
func (p *StructuredProcessor) Process(schema *schemaDomain.Schema, output string) (interface{}, error) {
	// Extract JSON from the output using our optimized extractor
	jsonStr := ExtractJSON(output)

	if jsonStr == "" {
		return nil, fmt.Errorf("no valid JSON found in the output")
	}

	// Parse the JSON
	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate against the schema
	validationResult, err := p.validator.Validate(schema, jsonStr)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if !validationResult.Valid {
		errorDetails := strings.Join(validationResult.Errors, ", ")
		return nil, fmt.Errorf("output does not conform to schema: %s", errorDetails)
	}

	return result, nil
}

// ProcessTyped processes a raw output string against a schema and maps to a target type
func (p *StructuredProcessor) ProcessTyped(schema *schemaDomain.Schema, output string, target interface{}) error {
	// Check if target is a pointer
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	// Extract and validate JSON from the output
	result, err := p.Process(schema, output)
	if err != nil {
		return err
	}

	// Convert the result to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal validated result: %w", err)
	}

	// Unmarshal into the target type
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal into target type: %w", err)
	}

	return nil
}

// ToJSON converts an object to a JSON string
func (p *StructuredProcessor) ToJSON(obj interface{}) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("object cannot be nil")
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// Legacy extraction functions are replaced by the optimized ExtractJSON
