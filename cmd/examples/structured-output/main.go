// ABOUTME: Example demonstrating structured output parsing and validation
// ABOUTME: Shows parsing LLM responses with recovery and format conversion

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/llm/outputs"
)

func main() {
	ctx := context.Background()

	// Example 1: Parse JSON with recovery
	fmt.Println("=== Example 1: JSON Parsing with Recovery ===")
	parseJSONWithRecovery(ctx)

	// Example 2: Validate against schema
	fmt.Println("\n=== Example 2: Schema Validation ===")
	validateAgainstSchema(ctx)

	// Example 3: Convert between formats
	fmt.Println("\n=== Example 3: Format Conversion ===")
	convertFormats(ctx)

	// Example 4: Bridge integration
	fmt.Println("\n=== Example 4: Bridge Integration ===")
	bridgeIntegration(ctx)
}

func parseJSONWithRecovery(ctx context.Context) {
	// Simulated LLM output with common issues
	llmOutput := "Here's the JSON response for the user profile:\n\n```json\n{\n  \"name\": \"John Doe\",\n  \"age\": 30,\n  \"email\": \"john@example.com\",\n  \"interests\": [\"programming\", \"AI\", \"music\"],\n  \"verified\": true,\n}\n```\n\nNote: The user has been active since 2020."

	parser := outputs.NewJSONParser()

	// Try parsing with recovery
	result, err := parser.ParseWithRecovery(ctx, llmOutput, &outputs.RecoveryOptions{
		ExtractFromMarkdown: true,
		FixCommonIssues:     true,
		MaxAttempts:         3,
	})

	if err != nil {
		log.Printf("Failed to parse: %v", err)
		return
	}

	fmt.Printf("Successfully parsed: %+v\n", result)
}

func validateAgainstSchema(ctx context.Context) {
	// Define a schema for user profile
	schema := &outputs.OutputSchema{
		Type: outputs.TypeObject,
		Properties: map[string]*outputs.OutputSchema{
			"name": {
				Type:     outputs.TypeString,
				Required: boolPtr(true),
			},
			"age": {
				Type:    outputs.TypeInteger,
				Minimum: float64Ptr(0),
				Maximum: float64Ptr(150),
			},
			"email": {
				Type:   outputs.TypeString,
				Format: "email",
			},
			"interests": {
				Type: outputs.TypeArray,
				Items: &outputs.OutputSchema{
					Type: outputs.TypeString,
				},
				MinItems: intPtr(1),
				MaxItems: intPtr(10),
			},
			"verified": {
				Type: outputs.TypeBoolean,
			},
		},
		RequiredProperties: []string{"name", "email"},
	}

	// Sample data to validate
	data := map[string]interface{}{
		"name":      "Jane Smith",
		"age":       25,
		"email":     "jane@example.com",
		"interests": []interface{}{"reading", "hiking"},
		"verified":  false,
	}

	validator := outputs.NewValidator()
	result, err := validator.Validate(ctx, data, schema)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if result.Valid {
		fmt.Println("✓ Data is valid!")
	} else {
		fmt.Println("✗ Validation failed:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Path, err.Message)
		}

		if len(result.Suggestions) > 0 {
			fmt.Println("\nSuggestions:")
			for _, suggestion := range result.Suggestions {
				fmt.Printf("  - %s: %s\n", suggestion.Path, suggestion.Description)
			}
		}
	}
}

func convertFormats(ctx context.Context) {
	// Sample data
	jsonData := `{
		"product": "Go-LLMs",
		"version": "0.3.5",
		"features": ["parsing", "validation", "conversion"]
	}`

	converter := outputs.NewConverter()

	// Convert JSON to YAML
	yamlResult, err := converter.ConvertString(ctx, jsonData, outputs.FormatJSON, outputs.FormatYAML, nil)
	if err != nil {
		log.Printf("Failed to convert to YAML: %v", err)
		return
	}

	fmt.Println("YAML output:")
	fmt.Println(yamlResult)

	// Convert JSON to XML
	xmlResult, err := converter.ConvertString(ctx, jsonData, outputs.FormatJSON, outputs.FormatXML, &outputs.ConversionOptions{
		Pretty:      true,
		RootElement: "package",
	})
	if err != nil {
		log.Printf("Failed to convert to XML: %v", err)
		return
	}

	fmt.Println("\nXML output:")
	fmt.Println(xmlResult)
}

func bridgeIntegration(ctx context.Context) {
	// Create bridge adapter
	bridge := outputs.NewBridgeAdapter()

	// Define a bridge schema (from go-llmspell)
	bridgeSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task": map[string]interface{}{
				"type":     "string",
				"required": true,
				"enum":     []interface{}{"create", "update", "delete"},
			},
			"target": map[string]interface{}{
				"type":     "string",
				"required": true,
			},
			"parameters": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": true,
			},
		},
		"required": []interface{}{"task", "target"},
	}

	// Convert bridge schema to OutputSchema
	schema, err := bridge.ConvertSchemaFromBridge(bridgeSchema)
	if err != nil {
		log.Printf("Failed to convert schema: %v", err)
		return
	}

	// Simulated LLM output
	llmOutput := `The command structure is: {"task": "create", "target": "function", "parameters": {"name": "hello"}}`

	// Parse and validate
	result, err := bridge.ParseAndValidate(ctx, llmOutput, schema)
	if err != nil {
		log.Printf("Failed to parse and validate: %v", err)
		return
	}

	if result.Success {
		fmt.Println("✓ Bridge parsing successful!")
		fmt.Printf("Format: %s\n", result.Format)
		fmt.Printf("Data: %+v\n", result.Data)
	} else {
		fmt.Printf("✗ Bridge parsing failed: %s\n", result.Error)
	}

	// Get parser info
	info := bridge.GetParserInfo()
	fmt.Printf("\nAvailable parsers: %v\n", info["parsers"])
	fmt.Printf("Features: %v\n", info["features"])
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
