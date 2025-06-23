// Package domain defines core types and interfaces for structured LLM outputs.
//
// This package provides the foundational elements for parsing, validating,
// and working with structured outputs from language models. It includes
// parsers for various output formats and integration with schema validation.
//
// Core Concepts:
//
// Parser Interface:
// The Parser interface defines the contract for converting LLM text outputs
// into structured Go values. Implementations handle different output formats
// such as JSON, YAML, XML, and custom formats.
//
// Parser Types:
//   - JSONParser: Extracts and parses JSON from LLM responses
//   - YAMLParser: Handles YAML-formatted outputs
//   - XMLParser: Processes XML responses
//   - RegexParser: Uses patterns to extract structured data
//   - CustomParser: Allows domain-specific parsing logic
//
// Schema Integration:
// Parsers work with the schema package to validate extracted data:
//
//	schema := &schema.Schema{
//	    Type: "object",
//	    Properties: map[string]*schema.Schema{
//	        "name": {Type: "string"},
//	        "age":  {Type: "number"},
//	    },
//	}
//
//	parser := NewJSONParser(schema)
//	result, err := parser.Parse(llmOutput)
//
// Error Handling:
// The package provides detailed error types for different parsing failures:
//   - ParseError: General parsing failures
//   - ValidationError: Schema validation failures
//   - FormatError: Unexpected format issues
//
// Best Practices:
//   - Always provide schemas for predictable parsing
//   - Use appropriate parser types for expected formats
//   - Handle partial extraction gracefully
//   - Implement custom parsers for domain-specific needs
//
// Example Usage:
//
//	// Create a parser with schema
//	parser := domain.NewJSONParser(mySchema)
//
//	// Parse LLM output
//	output := "The result is: {\"status\": \"success\", \"data\": [1,2,3]}"
//	result, err := parser.Parse(output)
//	if err != nil {
//	    log.Printf("Parse error: %v", err)
//	}
//
//	// Use the structured result
//	data := result.(map[string]interface{})
package domain
