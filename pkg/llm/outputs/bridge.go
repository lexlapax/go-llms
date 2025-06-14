// ABOUTME: Bridge integration support for go-llmspell scripting engine
// ABOUTME: Provides schema conversion and result validation for bridge layer

package outputs

import (
	"context"
	"fmt"
)

// BridgeAdapter provides integration with go-llmspell bridge layer
type BridgeAdapter struct {
	parser    Parser
	validator *Validator
	converter *Converter
}

// NewBridgeAdapter creates a new bridge adapter
func NewBridgeAdapter() *BridgeAdapter {
	return &BridgeAdapter{
		parser:    NewJSONParser(), // Default to JSON
		validator: NewValidator(),
		converter: NewConverter(),
	}
}

// ConvertSchemaFromBridge converts a bridge schema format to OutputSchema
func (a *BridgeAdapter) ConvertSchemaFromBridge(bridgeSchema map[string]interface{}) (*OutputSchema, error) {
	// Extract type
	schemaType, ok := bridgeSchema["type"].(string)
	if !ok {
		return nil, fmt.Errorf("schema must have a type field")
	}

	schema := &OutputSchema{
		Type: Type(schemaType),
	}

	// Extract common fields
	if desc, ok := bridgeSchema["description"].(string); ok {
		schema.Description = desc
	}

	if req, ok := bridgeSchema["required"].(bool); ok {
		schema.Required = &req
	}

	// Handle type-specific fields
	switch schema.Type {
	case TypeString:
		a.convertStringSchema(bridgeSchema, schema)
	case TypeNumber, TypeInteger:
		a.convertNumberSchema(bridgeSchema, schema)
	case TypeBoolean:
		// Boolean doesn't need special handling
	case TypeArray:
		a.convertArraySchema(bridgeSchema, schema)
	case TypeObject:
		a.convertObjectSchema(bridgeSchema, schema)
	}

	return schema, nil
}

// convertStringSchema converts string-specific schema fields
func (a *BridgeAdapter) convertStringSchema(bridge map[string]interface{}, schema *OutputSchema) {
	if format, ok := bridge["format"].(string); ok {
		schema.Format = format
	}

	if pattern, ok := bridge["pattern"].(string); ok {
		schema.Pattern = pattern
	}

	if enum, ok := bridge["enum"].([]interface{}); ok {
		schema.Enum = make([]string, 0, len(enum))
		for _, e := range enum {
			if s, ok := e.(string); ok {
				schema.Enum = append(schema.Enum, s)
			}
		}
	}
}

// convertNumberSchema converts number-specific schema fields
func (a *BridgeAdapter) convertNumberSchema(bridge map[string]interface{}, schema *OutputSchema) {
	if min, ok := bridge["minimum"].(float64); ok {
		schema.Minimum = &min
	} else if min, ok := bridge["minimum"].(int); ok {
		minFloat := float64(min)
		schema.Minimum = &minFloat
	}

	if max, ok := bridge["maximum"].(float64); ok {
		schema.Maximum = &max
	} else if max, ok := bridge["maximum"].(int); ok {
		maxFloat := float64(max)
		schema.Maximum = &maxFloat
	}
}

// convertArraySchema converts array-specific schema fields
func (a *BridgeAdapter) convertArraySchema(bridge map[string]interface{}, schema *OutputSchema) {
	if minItems, ok := bridge["minItems"].(int); ok {
		schema.MinItems = &minItems
	} else if minItems, ok := bridge["minItems"].(float64); ok {
		minInt := int(minItems)
		schema.MinItems = &minInt
	}

	if maxItems, ok := bridge["maxItems"].(int); ok {
		schema.MaxItems = &maxItems
	} else if maxItems, ok := bridge["maxItems"].(float64); ok {
		maxInt := int(maxItems)
		schema.MaxItems = &maxInt
	}

	// Convert items schema
	if items, ok := bridge["items"].(map[string]interface{}); ok {
		itemSchema, err := a.ConvertSchemaFromBridge(items)
		if err == nil {
			schema.Items = itemSchema
		}
	}
}

// convertObjectSchema converts object-specific schema fields
func (a *BridgeAdapter) convertObjectSchema(bridge map[string]interface{}, schema *OutputSchema) {
	// Convert properties
	if props, ok := bridge["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]*OutputSchema)
		for name, propData := range props {
			if propMap, ok := propData.(map[string]interface{}); ok {
				propSchema, err := a.ConvertSchemaFromBridge(propMap)
				if err == nil {
					schema.Properties[name] = propSchema
				}
			}
		}
	}

	// Extract required properties
	if required, ok := bridge["required"].([]interface{}); ok {
		schema.RequiredProperties = make([]string, 0, len(required))
		for _, r := range required {
			if s, ok := r.(string); ok {
				schema.RequiredProperties = append(schema.RequiredProperties, s)
			}
		}
	}

	// Additional properties
	if addProps, ok := bridge["additionalProperties"].(bool); ok {
		schema.AdditionalProperties = &addProps
	}
}

// ParseAndValidate parses output and validates against a schema
func (a *BridgeAdapter) ParseAndValidate(ctx context.Context, output string, schema *OutputSchema) (*BridgeResult, error) {
	// Auto-detect format and parse
	parseResult, err := ParseWithAutoDetection(ctx, output, nil)
	if err != nil {
		// Try with recovery
		parseResult, err = ParseWithAutoDetection(ctx, output, &RecoveryOptions{
			ExtractFromMarkdown: true,
			FixCommonIssues:     true,
			MaxAttempts:         5,
			Schema:              schema,
		})
		if err != nil {
			return &BridgeResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to parse output: %v", err),
			}, nil
		}
	}

	// Validate if schema provided
	var validationResult *ValidationResult
	if schema != nil {
		validationResult, err = a.validator.Validate(ctx, parseResult.Data, schema)
		if err != nil {
			return &BridgeResult{
				Success: false,
				Error:   fmt.Sprintf("Validation error: %v", err),
			}, nil
		}
	}

	return &BridgeResult{
		Success:    validationResult == nil || validationResult.Valid,
		Data:       parseResult.Data,
		Format:     parseResult.Format,
		Validation: validationResult,
	}, nil
}

// BridgeResult represents the result of bridge operations
type BridgeResult struct {
	// Success indicates if the operation was successful
	Success bool `json:"success"`

	// Data is the parsed data (if successful)
	Data interface{} `json:"data,omitempty"`

	// Format is the detected format
	Format string `json:"format,omitempty"`

	// Error contains error message (if failed)
	Error string `json:"error,omitempty"`

	// Validation contains validation results
	Validation *ValidationResult `json:"validation,omitempty"`
}

// ConvertFormat converts data between formats for bridge layer
func (a *BridgeAdapter) ConvertFormat(ctx context.Context, data interface{}, fromFormat, toFormat string) (*BridgeResult, error) {
	from := Format(fromFormat)
	to := Format(toFormat)

	converted, err := a.converter.Convert(ctx, data, from, to, nil)
	if err != nil {
		return &BridgeResult{
			Success: false,
			Error:   fmt.Sprintf("Conversion failed: %v", err),
		}, nil
	}

	return &BridgeResult{
		Success: true,
		Data:    converted,
		Format:  toFormat,
	}, nil
}

// GetSupportedFormats returns the list of supported formats
func (a *BridgeAdapter) GetSupportedFormats() []string {
	return []string{
		string(FormatJSON),
		string(FormatXML),
		string(FormatYAML),
	}
}

// GetParserInfo returns information about available parsers
func (a *BridgeAdapter) GetParserInfo() map[string]interface{} {
	parsers := ListParsers()
	info := make(map[string]interface{})

	info["parsers"] = parsers
	info["defaultParser"] = "json"
	info["features"] = []string{
		"markdown_extraction",
		"error_recovery",
		"schema_guided_parsing",
		"format_auto_detection",
		"common_issue_fixing",
	}

	return info
}

// FixOutput attempts to fix common output issues
func (a *BridgeAdapter) FixOutput(ctx context.Context, output string, hints map[string]interface{}) (*BridgeResult, error) {
	// Determine the format
	format := ""
	if f, ok := hints["format"].(string); ok {
		format = f
	} else {
		// Auto-detect
		detectedFormat, err := DetectFormat(output)
		if err == nil {
			format = string(detectedFormat)
		}
	}

	// Get the appropriate parser
	parser, err := GetParser(format)
	if err != nil {
		// Default to JSON parser
		parser = NewJSONParser()
		format = "json"
	}

	// Try to parse with maximum recovery
	opts := &RecoveryOptions{
		ExtractFromMarkdown: true,
		FixCommonIssues:     true,
		MaxAttempts:         10,
	}

	// Add schema if provided
	if schemaData, ok := hints["schema"].(map[string]interface{}); ok {
		schema, _ := a.ConvertSchemaFromBridge(schemaData)
		opts.Schema = schema
	}

	data, err := parser.ParseWithRecovery(ctx, output, opts)
	if err != nil {
		return &BridgeResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to fix output: %v", err),
		}, nil
	}

	return &BridgeResult{
		Success: true,
		Data:    data,
		Format:  format,
	}, nil
}
