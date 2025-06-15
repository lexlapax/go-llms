package docs

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
)

func TestGenerateToolDocumentation(t *testing.T) {
	// Create a sample ToolInfo
	sampleSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
		},
		"required": []string{"query"},
	}
	schemaBytes, _ := json.Marshal(sampleSchema)

	toolInfo := tools.ToolInfo{
		Name:            "test_tool",
		Description:     "A test tool for documentation",
		Category:        "test",
		Tags:            []string{"testing", "example"},
		Version:         "1.0.0",
		ParameterSchema: json.RawMessage(schemaBytes),
		Examples: []tools.Example{
			{
				Name:        "basic_search",
				Description: "Basic search example",
				Input:       json.RawMessage(`{"query": "test"}`),
				Output:      json.RawMessage(`{"results": ["item1", "item2"]}`),
			},
		},
		UsageHint: "Use this tool to perform test operations",
		Package:   "test/package",
	}

	// Generate documentation
	doc, err := GenerateToolDocumentation(toolInfo)
	if err != nil {
		t.Fatalf("GenerateToolDocumentation failed: %v", err)
	}

	// Verify basic fields
	if doc.Name != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", doc.Name)
	}

	if doc.Description != "A test tool for documentation" {
		t.Errorf("Expected description to match")
	}

	if doc.Category != "test" {
		t.Errorf("Expected category 'test', got '%s'", doc.Category)
	}

	if len(doc.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(doc.Tags))
	}

	if doc.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", doc.Version)
	}

	// Verify schemas
	if doc.Schemas == nil {
		t.Fatal("Expected schemas to be populated")
	}

	if _, exists := doc.Schemas["input"]; !exists {
		t.Error("Expected input schema to exist")
	}

	// Verify examples
	if len(doc.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(doc.Examples))
	}

	example := doc.Examples[0]
	if example.Name != "basic_search" {
		t.Errorf("Expected example name 'basic_search', got '%s'", example.Name)
	}

	// Verify metadata
	if doc.Metadata["package"] != "test/package" {
		t.Errorf("Expected package metadata to be set")
	}
}

func TestConvertToolInfoToOpenAPIOperation(t *testing.T) {
	// Create a sample ToolInfo
	sampleSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"input": map[string]interface{}{
				"type":        "string",
				"description": "Input value",
			},
		},
	}
	schemaBytes, _ := json.Marshal(sampleSchema)

	toolInfo := tools.ToolInfo{
		Name:            "convert_test",
		Description:     "Test conversion tool",
		Category:        "conversion",
		ParameterSchema: json.RawMessage(schemaBytes),
		OutputSchema:    json.RawMessage(schemaBytes),
		UsageHint:       "This tool converts data",
	}

	// Convert to OpenAPI operation
	operation, err := ConvertToolInfoToOpenAPIOperation(toolInfo)
	if err != nil {
		t.Fatalf("ConvertToolInfoToOpenAPIOperation failed: %v", err)
	}

	// Verify operation fields
	if operation.OperationID != "execute_convert_test" {
		t.Errorf("Expected operation ID 'execute_convert_test', got '%s'", operation.OperationID)
	}

	if operation.Summary != "Test conversion tool" {
		t.Errorf("Expected summary to match description")
	}

	if len(operation.Tags) != 1 || operation.Tags[0] != "conversion" {
		t.Errorf("Expected tags to contain category")
	}

	// Verify request body
	if operation.RequestBody == nil {
		t.Fatal("Expected request body to be set")
	}

	if !operation.RequestBody.Required {
		t.Error("Expected request body to be required")
	}

	// Verify responses
	if len(operation.Responses) == 0 {
		t.Fatal("Expected responses to be set")
	}

	if _, exists := operation.Responses["200"]; !exists {
		t.Error("Expected 200 response to exist")
	}

	if _, exists := operation.Responses["400"]; !exists {
		t.Error("Expected 400 response to exist")
	}

	if _, exists := operation.Responses["500"]; !exists {
		t.Error("Expected 500 response to exist")
	}
}

func TestGenerateToolOpenAPI(t *testing.T) {
	ctx := context.Background()

	// Create sample tools
	tools := []tools.ToolInfo{
		{
			Name:        "tool1",
			Description: "First test tool",
			Category:    "test",
			Version:     "1.0.0",
		},
		{
			Name:        "tool2",
			Description: "Second test tool",
			Category:    "test",
			Version:     "1.0.0",
		},
	}

	config := GeneratorConfig{
		Title:       "Test API",
		Description: "Test tool API",
		Version:     "1.0.0",
		BaseURL:     "https://api.test.com",
	}

	// Generate OpenAPI spec
	spec, err := GenerateToolOpenAPI(ctx, tools, config)
	if err != nil {
		t.Fatalf("GenerateToolOpenAPI failed: %v", err)
	}

	// Verify spec structure
	if spec.OpenAPI != "3.0.0" {
		t.Errorf("Expected OpenAPI version '3.0.0', got '%s'", spec.OpenAPI)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", spec.Info.Title)
	}

	if len(spec.Paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(spec.Paths))
	}

	// Verify tool paths exist
	if _, exists := spec.Paths["/tools/tool1/execute"]; !exists {
		t.Error("Expected tool1 path to exist")
	}

	if _, exists := spec.Paths["/tools/tool2/execute"]; !exists {
		t.Error("Expected tool2 path to exist")
	}

	// Verify servers
	if len(spec.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(spec.Servers))
	}

	if spec.Servers[0].URL != "https://api.test.com" {
		t.Errorf("Expected server URL 'https://api.test.com', got '%s'", spec.Servers[0].URL)
	}
}

func TestGenerateToolMarkdown(t *testing.T) {
	ctx := context.Background()

	// Create sample tools
	tools := []tools.ToolInfo{
		{
			Name:        "markdown_tool",
			Description: "Tool for markdown testing",
			Category:    "docs",
			Tags:        []string{"markdown", "test"},
			Version:     "1.0.0",
		},
	}

	config := GeneratorConfig{
		Title:           "Markdown Test",
		Description:     "Test markdown generation",
		Version:         "1.0.0",
		IncludeExamples: true,
		IncludeSchemas:  true,
	}

	// Generate Markdown
	markdown, err := GenerateToolMarkdown(ctx, tools, config)
	if err != nil {
		t.Fatalf("GenerateToolMarkdown failed: %v", err)
	}

	// Verify markdown contains expected content
	if len(markdown) == 0 {
		t.Error("Expected non-empty markdown")
	}

	// Check for title
	if !contains(markdown, "# Markdown Test") {
		t.Error("Expected title in markdown")
	}

	// Check for tool name
	if !contains(markdown, "### markdown_tool") {
		t.Error("Expected tool name in markdown")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			contains(s[1:], substr) ||
			(len(s) > 0 && s[:len(substr)] == substr))
}

func TestConvertJSONRawMessageToSchema(t *testing.T) {
	// Test schema conversion
	jsonSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name field",
			},
			"age": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
				"maximum": 120,
			},
		},
		"required": []string{"name"},
	}

	jsonBytes, _ := json.Marshal(jsonSchema)
	schema, err := convertJSONRawMessageToSchema(jsonBytes)
	if err != nil {
		t.Fatalf("convertJSONRawMessageToSchema failed: %v", err)
	}

	// Verify conversion
	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", schema.Type)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Expected required field 'name'")
	}

	// Check property details
	nameProperty := schema.Properties["name"]
	if nameProperty.Type != "string" {
		t.Errorf("Expected name property type 'string', got '%s'", nameProperty.Type)
	}

	ageProperty := schema.Properties["age"]
	if ageProperty.Type != "integer" {
		t.Errorf("Expected age property type 'integer', got '%s'", ageProperty.Type)
	}

	if ageProperty.Minimum == nil || *ageProperty.Minimum != 0 {
		t.Error("Expected age minimum to be 0")
	}

	if ageProperty.Maximum == nil || *ageProperty.Maximum != 120 {
		t.Error("Expected age maximum to be 120")
	}
}
