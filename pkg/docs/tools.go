// ABOUTME: Tool-specific documentation generation that integrates with the discovery system
// ABOUTME: Converts ToolInfo structures to Documentation format with OpenAPI support for tools

package docs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
)

// GenerateToolDocumentation converts a ToolInfo from the discovery system to Documentation format
func GenerateToolDocumentation(toolInfo tools.ToolInfo) (Documentation, error) {
	doc := Documentation{
		Name:            toolInfo.Name,
		Description:     toolInfo.Description,
		LongDescription: toolInfo.UsageHint,
		Category:        toolInfo.Category,
		Tags:            toolInfo.Tags,
		Version:         toolInfo.Version,
		Metadata:        make(map[string]interface{}),
	}

	// Add package information to metadata
	if toolInfo.Package != "" {
		doc.Metadata["package"] = toolInfo.Package
	}

	// Convert parameter schema
	if len(toolInfo.ParameterSchema) > 0 {
		paramSchema, err := convertJSONRawMessageToSchema(toolInfo.ParameterSchema)
		if err != nil {
			return doc, fmt.Errorf("failed to convert parameter schema: %w", err)
		}
		if doc.Schemas == nil {
			doc.Schemas = make(map[string]*Schema)
		}
		doc.Schemas["input"] = paramSchema
	}

	// Convert output schema
	if len(toolInfo.OutputSchema) > 0 {
		outputSchema, err := convertJSONRawMessageToSchema(toolInfo.OutputSchema)
		if err != nil {
			return doc, fmt.Errorf("failed to convert output schema: %w", err)
		}
		if doc.Schemas == nil {
			doc.Schemas = make(map[string]*Schema)
		}
		doc.Schemas["output"] = outputSchema
	}

	// Convert examples
	doc.Examples = make([]Example, len(toolInfo.Examples))
	for i, toolExample := range toolInfo.Examples {
		var input, output interface{}

		// Parse input
		if len(toolExample.Input) > 0 {
			if err := json.Unmarshal(toolExample.Input, &input); err != nil {
				return doc, fmt.Errorf("failed to parse example input: %w", err)
			}
		}

		// Parse output
		if len(toolExample.Output) > 0 {
			if err := json.Unmarshal(toolExample.Output, &output); err != nil {
				return doc, fmt.Errorf("failed to parse example output: %w", err)
			}
		}

		doc.Examples[i] = Example{
			Name:        toolExample.Name,
			Description: toolExample.Description,
			Input:       input,
			Output:      output,
			Language:    "json", // Tools typically use JSON
		}
	}

	return doc, nil
}

// GenerateToolOpenAPI creates an OpenAPI specification specifically for tools
func GenerateToolOpenAPI(ctx context.Context, toolInfos []tools.ToolInfo, config GeneratorConfig) (*OpenAPISpec, error) {
	// Convert tools to documentable items
	documentables := make([]Documentable, len(toolInfos))
	for i, toolInfo := range toolInfos {
		doc, err := GenerateToolDocumentation(toolInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate documentation for tool %s: %w", toolInfo.Name, err)
		}
		documentables[i] = &documentableWrapper{doc: doc}
	}

	// Use the standard OpenAPI generator
	generator := &standardGenerator{config: config}
	return generator.GenerateOpenAPI(ctx, documentables)
}

// GenerateToolMarkdown creates markdown documentation for tools
func GenerateToolMarkdown(ctx context.Context, toolInfos []tools.ToolInfo, config GeneratorConfig) (string, error) {
	// Convert tools to documentable items
	documentables := make([]Documentable, len(toolInfos))
	for i, toolInfo := range toolInfos {
		doc, err := GenerateToolDocumentation(toolInfo)
		if err != nil {
			return "", fmt.Errorf("failed to generate documentation for tool %s: %w", toolInfo.Name, err)
		}
		documentables[i] = &documentableWrapper{doc: doc}
	}

	// Use the standard markdown generator
	generator := &standardGenerator{config: config}
	return generator.GenerateMarkdown(ctx, documentables)
}

// ConvertToolInfoToOpenAPIOperation converts a single ToolInfo to an OpenAPI operation
func ConvertToolInfoToOpenAPIOperation(toolInfo tools.ToolInfo) (*Operation, error) {
	operation := &Operation{
		Summary:     toolInfo.Description,
		Description: toolInfo.UsageHint,
		OperationID: fmt.Sprintf("execute_%s", toolInfo.Name),
		Tags:        []string{toolInfo.Category},
		Responses:   make(map[string]*Response),
	}

	// Add tool metadata to description
	if toolInfo.UsageHint == "" && len(toolInfo.Tags) > 0 {
		operation.Description = fmt.Sprintf("Tool tags: %v", toolInfo.Tags)
	}

	// Convert parameter schema to request body
	if len(toolInfo.ParameterSchema) > 0 {
		paramSchema, err := convertJSONRawMessageToSchema(toolInfo.ParameterSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert parameter schema: %w", err)
		}

		operation.RequestBody = &RequestBody{
			Description: fmt.Sprintf("Parameters for %s tool", toolInfo.Name),
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: paramSchema,
				},
			},
		}

		// Add examples to request body if available
		if len(toolInfo.Examples) > 0 {
			examples := make(map[string]Example)
			for _, ex := range toolInfo.Examples {
				var input interface{}
				if len(ex.Input) > 0 {
					_ = json.Unmarshal(ex.Input, &input) // Best effort parsing
				}
				examples[ex.Name] = Example{
					Name:        ex.Name,
					Description: ex.Description,
					Input:       input,
				}
			}
			mediaType := operation.RequestBody.Content["application/json"]
			mediaType.Examples = examples
			operation.RequestBody.Content["application/json"] = mediaType
		}
	}

	// Convert output schema to response
	if len(toolInfo.OutputSchema) > 0 {
		outputSchema, err := convertJSONRawMessageToSchema(toolInfo.OutputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert output schema: %w", err)
		}

		operation.Responses["200"] = &Response{
			Description: fmt.Sprintf("Successful execution of %s", toolInfo.Name),
			Content: map[string]MediaType{
				"application/json": {
					Schema: outputSchema,
				},
			},
		}
	} else {
		// Default success response
		operation.Responses["200"] = &Response{
			Description: "Tool executed successfully",
		}
	}

	// Add error responses
	operation.Responses["400"] = &Response{
		Description: "Invalid input parameters",
	}
	operation.Responses["500"] = &Response{
		Description: "Tool execution failed",
	}

	return operation, nil
}

// convertJSONRawMessageToSchema converts json.RawMessage to our Schema type
func convertJSONRawMessageToSchema(raw json.RawMessage) (*Schema, error) {
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(raw, &schemaMap); err != nil {
		return nil, err
	}

	schema := &Schema{}

	// Basic type conversion
	if t, ok := schemaMap["type"].(string); ok {
		schema.Type = t
	}
	if title, ok := schemaMap["title"].(string); ok {
		schema.Title = title
	}
	if desc, ok := schemaMap["description"].(string); ok {
		schema.Description = desc
	}
	if format, ok := schemaMap["format"].(string); ok {
		schema.Format = format
	}
	if pattern, ok := schemaMap["pattern"].(string); ok {
		schema.Pattern = pattern
	}

	// Properties conversion
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]*Schema)
		for name, prop := range props {
			if propBytes, err := json.Marshal(prop); err == nil {
				if propSchema, err := convertJSONRawMessageToSchema(propBytes); err == nil {
					schema.Properties[name] = propSchema
				}
			}
		}
	}

	// Required fields
	if req, ok := schemaMap["required"].([]interface{}); ok {
		schema.Required = make([]string, len(req))
		for i, r := range req {
			if s, ok := r.(string); ok {
				schema.Required[i] = s
			}
		}
	}

	// Enum values
	if enum, ok := schemaMap["enum"].([]interface{}); ok {
		schema.Enum = enum
	}

	// Default value
	if def, ok := schemaMap["default"]; ok {
		schema.Default = def
	}

	// Numeric constraints
	if min, ok := schemaMap["minimum"].(float64); ok {
		schema.Minimum = &min
	}
	if max, ok := schemaMap["maximum"].(float64); ok {
		schema.Maximum = &max
	}

	// String constraints
	if minLen, ok := schemaMap["minLength"].(float64); ok {
		minLenInt := int(minLen)
		schema.MinLength = &minLenInt
	}
	if maxLen, ok := schemaMap["maxLength"].(float64); ok {
		maxLenInt := int(maxLen)
		schema.MaxLength = &maxLenInt
	}

	// Items for arrays
	if items, ok := schemaMap["items"]; ok {
		if itemsBytes, err := json.Marshal(items); err == nil {
			if itemsSchema, err := convertJSONRawMessageToSchema(itemsBytes); err == nil {
				schema.Items = itemsSchema
			}
		}
	}

	// Additional properties
	if additional, ok := schemaMap["additionalProperties"]; ok {
		if additionalMap, ok := additional.(map[string]interface{}); ok {
			schema.Additional = additionalMap
		}
	}

	return schema, nil
}

// documentableWrapper wraps Documentation to implement Documentable interface
type documentableWrapper struct {
	doc Documentation
}

func (w *documentableWrapper) GetDocumentation() Documentation {
	return w.doc
}

// standardGenerator provides the standard implementation for generating documentation
type standardGenerator struct {
	config GeneratorConfig
}

func (g *standardGenerator) GenerateOpenAPI(ctx context.Context, items []Documentable) (*OpenAPISpec, error) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:       g.config.Title,
			Description: g.config.Description,
			Version:     g.config.Version,
		},
		Paths:      make(map[string]*PathItem),
		Components: &Components{},
	}

	// Add base URL if provided
	if g.config.BaseURL != "" {
		spec.Servers = []Server{
			{
				URL:         g.config.BaseURL,
				Description: "Tool execution server",
			},
		}
	}

	// Process each documentable item
	for _, item := range items {
		doc := item.GetDocumentation()

		// Create path for tool execution
		path := fmt.Sprintf("/tools/%s/execute", doc.Name)
		pathItem := &PathItem{
			Summary:     doc.Description,
			Description: doc.LongDescription,
		}

		// Create POST operation for tool execution
		operation := &Operation{
			Summary:     fmt.Sprintf("Execute %s tool", doc.Name),
			Description: doc.Description,
			OperationID: fmt.Sprintf("execute_%s", doc.Name),
			Tags:        doc.Tags,
			Responses:   make(map[string]*Response),
		}

		// Add input schema as request body
		if inputSchema, exists := doc.Schemas["input"]; exists {
			operation.RequestBody = &RequestBody{
				Description: fmt.Sprintf("Input parameters for %s", doc.Name),
				Required:    true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: inputSchema,
					},
				},
			}
		}

		// Add output schema as response
		if outputSchema, exists := doc.Schemas["output"]; exists {
			operation.Responses["200"] = &Response{
				Description: fmt.Sprintf("Successful execution of %s", doc.Name),
				Content: map[string]MediaType{
					"application/json": {
						Schema: outputSchema,
					},
				},
			}
		} else {
			operation.Responses["200"] = &Response{
				Description: "Tool executed successfully",
			}
		}

		// Add standard error responses
		operation.Responses["400"] = &Response{Description: "Invalid input parameters"}
		operation.Responses["500"] = &Response{Description: "Tool execution failed"}

		pathItem.Post = operation
		spec.Paths[path] = pathItem

		// Add tool category as tag
		if doc.Category != "" {
			found := false
			for _, tag := range spec.Tags {
				if tag.Name == doc.Category {
					found = true
					break
				}
			}
			if !found {
				spec.Tags = append(spec.Tags, Tag{
					Name:        doc.Category,
					Description: fmt.Sprintf("Tools in the %s category", doc.Category),
				})
			}
		}
	}

	return spec, nil
}

func (g *standardGenerator) GenerateMarkdown(ctx context.Context, items []Documentable) (string, error) {
	// Use the existing markdown generator
	generator := NewMarkdownGenerator(g.config)
	return generator.GenerateMarkdown(ctx, items)
}

func (g *standardGenerator) GenerateJSON(ctx context.Context, items []Documentable) ([]byte, error) {
	docs := make([]Documentation, len(items))
	for i, item := range items {
		docs[i] = item.GetDocumentation()
	}
	return json.MarshalIndent(docs, "", "  ")
}
