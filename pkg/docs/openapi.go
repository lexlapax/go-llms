package docs

// ABOUTME: OpenAPI 3.0 documentation generator for tools and components
// ABOUTME: Generates OpenAPI specifications from Documentable items with tool-specific support

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// OpenAPIGenerator generates OpenAPI 3.0 specifications.
// It converts Documentable items into OpenAPI specifications with
// paths, schemas, and examples suitable for API documentation.
type OpenAPIGenerator struct {
	config GeneratorConfig
}

// NewOpenAPIGenerator creates a new OpenAPI generator.
//
// Parameters:
//   - config: Configuration for the generator including title, version, and base URL
//
// Returns a configured OpenAPIGenerator instance.
func NewOpenAPIGenerator(config GeneratorConfig) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		config: config,
	}
}

// GenerateOpenAPI generates an OpenAPI specification from documentable items.
// It creates a complete OpenAPI 3.0.3 specification with paths, schemas,
// tags, and components based on the provided items.
//
// Parameters:
//   - ctx: The context for the operation
//   - items: The documentable items to include
//
// Returns an OpenAPI specification or an error.
func (g *OpenAPIGenerator) GenerateOpenAPI(ctx context.Context, items []Documentable) (*OpenAPISpec, error) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: &Info{
			Title:       g.config.Title,
			Description: g.config.Description,
			Version:     g.config.Version,
		},
		Paths: make(map[string]*PathItem),
		Components: &Components{
			Schemas:       make(map[string]*Schema),
			RequestBodies: make(map[string]*RequestBody),
		},
		Tags: []Tag{},
	}

	// Add server if base URL is provided
	if g.config.BaseURL != "" {
		spec.Servers = []Server{
			{
				URL:         g.config.BaseURL,
				Description: "Default server",
			},
		}
	}

	// Group items by category if specified
	groups := g.groupItems(items)

	// Process each group
	for category, groupItems := range groups {
		// Add tag for the category
		if category != "" {
			spec.Tags = append(spec.Tags, Tag{
				Name:        category,
				Description: fmt.Sprintf("Operations for %s", category),
			})
		}

		// Process each item in the group
		for _, item := range groupItems {
			if err := g.addItemToSpec(spec, item, category); err != nil {
				return nil, fmt.Errorf("failed to add item %s: %w", item.GetDocumentation().Name, err)
			}
		}
	}

	return spec, nil
}

// GenerateMarkdown is not implemented by OpenAPIGenerator.
// This method exists to satisfy the Generator interface but returns an error
// as OpenAPI generation doesn't produce markdown.
//
// Parameters:
//   - ctx: The context (unused)
//   - items: The items to document (unused)
//
// Returns an error indicating markdown generation is not supported.
func (g *OpenAPIGenerator) GenerateMarkdown(ctx context.Context, items []Documentable) (string, error) {
	return "", fmt.Errorf("markdown generation not supported by OpenAPIGenerator")
}

// GenerateJSON generates JSON representation of the OpenAPI spec.
// It first generates the OpenAPI specification, then serializes it to JSON.
//
// Parameters:
//   - ctx: The context for the operation
//   - items: The documentable items to include
//
// Returns JSON bytes of the OpenAPI spec or an error.
func (g *OpenAPIGenerator) GenerateJSON(ctx context.Context, items []Documentable) ([]byte, error) {
	spec, err := g.GenerateOpenAPI(ctx, items)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(spec, "", "  ")
}

// GenerateOpenAPIForTool generates OpenAPI documentation for a single tool.
// This is a convenience function for generating documentation for individual tools.
//
// Parameters:
//   - tool: The tool to document
//   - config: Generator configuration
//
// Returns an OpenAPI specification for the tool or an error.
func GenerateOpenAPIForTool(tool Documentable, config GeneratorConfig) (*OpenAPISpec, error) {
	generator := NewOpenAPIGenerator(config)
	return generator.GenerateOpenAPI(context.Background(), []Documentable{tool})
}

// groupItems groups documentable items by category.
// It supports grouping by category, type, or using a default group.
// This enables organized presentation in the OpenAPI specification.
//
// Parameters:
//   - items: The documentable items to group
//
// Returns a map of group names to items.
func (g *OpenAPIGenerator) groupItems(items []Documentable) map[string][]Documentable {
	groups := make(map[string][]Documentable)

	for _, item := range items {
		doc := item.GetDocumentation()
		category := ""

		switch g.config.GroupBy {
		case "category":
			category = doc.Category
		case "type":
			// Could be extended to group by type
			category = "default"
		default:
			category = "default"
		}

		if category == "" {
			category = "default"
		}

		groups[category] = append(groups[category], item)
	}

	return groups
}

// addItemToSpec adds a documentable item to the OpenAPI spec.
// It creates paths, operations, schemas, and examples for the item,
// organizing them within the appropriate sections of the specification.
//
// Parameters:
//   - spec: The OpenAPI specification to add to
//   - item: The documentable item to add
//   - category: The category for tagging
//
// Returns an error if the item cannot be added.
func (g *OpenAPIGenerator) addItemToSpec(spec *OpenAPISpec, item Documentable, category string) error {
	doc := item.GetDocumentation()

	// Create path for the item
	path := fmt.Sprintf("/tools/%s", strings.ToLower(doc.Name))

	// Create operation
	operation := &Operation{
		Summary:     doc.Description,
		Description: doc.LongDescription,
		OperationID: doc.Name,
		Tags:        []string{category},
		Responses: map[string]*Response{
			"200": {
				Description: "Successful response",
			},
			"400": {
				Description: "Invalid request",
			},
			"500": {
				Description: "Internal error",
			},
		},
	}

	// Add request body if there's an input schema
	if doc.Schema != nil || (doc.Schemas != nil && doc.Schemas["input"] != nil) {
		inputSchema := doc.Schema
		if doc.Schemas != nil && doc.Schemas["input"] != nil {
			inputSchema = doc.Schemas["input"]
		}

		// Add schema to components
		schemaName := fmt.Sprintf("%sInput", doc.Name)
		spec.Components.Schemas[schemaName] = inputSchema

		// Create request body
		operation.RequestBody = &RequestBody{
			Description: fmt.Sprintf("Input parameters for %s", doc.Name),
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Type:        "object",
						Title:       "$ref",
						Description: fmt.Sprintf("#/components/schemas/%s", schemaName),
					},
				},
			},
		}
	}

	// Add response schema if available
	if doc.Schemas != nil && doc.Schemas["output"] != nil {
		outputSchema := doc.Schemas["output"]
		schemaName := fmt.Sprintf("%sOutput", doc.Name)
		spec.Components.Schemas[schemaName] = outputSchema

		// Update successful response
		operation.Responses["200"] = &Response{
			Description: "Successful response",
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Type:        "object",
						Title:       "$ref",
						Description: fmt.Sprintf("#/components/schemas/%s", schemaName),
					},
				},
			},
		}
	}

	// Add examples if configured
	if g.config.IncludeExamples && len(doc.Examples) > 0 {
		examples := make(map[string]Example)
		for i, ex := range doc.Examples {
			examples[fmt.Sprintf("example%d", i+1)] = ex
		}

		if operation.RequestBody != nil && operation.RequestBody.Content["application/json"].Examples == nil {
			mt := operation.RequestBody.Content["application/json"]
			mt.Examples = examples
			operation.RequestBody.Content["application/json"] = mt
		}
	}

	// Add deprecation info
	if doc.Deprecated {
		operation.Deprecated = true
		if doc.DeprecationNote != "" {
			operation.Description = fmt.Sprintf("%s\n\n**Deprecated**: %s", operation.Description, doc.DeprecationNote)
		}
	}

	// Add the operation to the spec
	spec.Paths[path] = &PathItem{
		Post: operation,
	}

	return nil
}

// ConvertSchemaToOpenAPI converts our Schema type to OpenAPI schema format.
// It recursively transforms the Schema structure into a map suitable for
// JSON serialization in OpenAPI specifications.
//
// Parameters:
//   - schema: The schema to convert
//
// Returns a map representing the OpenAPI schema.
func ConvertSchemaToOpenAPI(schema *Schema) map[string]interface{} {
	result := make(map[string]interface{})

	if schema.Type != "" {
		result["type"] = schema.Type
	}
	if schema.Title != "" {
		result["title"] = schema.Title
	}
	if schema.Description != "" {
		result["description"] = schema.Description
	}
	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for k, v := range schema.Properties {
			props[k] = ConvertSchemaToOpenAPI(v)
		}
		result["properties"] = props
	}
	if schema.Items != nil {
		result["items"] = ConvertSchemaToOpenAPI(schema.Items)
	}
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}
	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}
	if schema.Default != nil {
		result["default"] = schema.Default
	}
	if schema.Format != "" {
		result["format"] = schema.Format
	}
	if schema.Pattern != "" {
		result["pattern"] = schema.Pattern
	}
	if schema.MinLength != nil {
		result["minLength"] = *schema.MinLength
	}
	if schema.MaxLength != nil {
		result["maxLength"] = *schema.MaxLength
	}
	if schema.Minimum != nil {
		result["minimum"] = *schema.Minimum
	}
	if schema.Maximum != nil {
		result["maximum"] = *schema.Maximum
	}

	// Add any additional properties
	for k, v := range schema.Additional {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	return result
}
