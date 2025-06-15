// ABOUTME: Integration helpers for the documentation system with tool discovery
// ABOUTME: Provides batch operations and bridges between discovery and documentation systems

package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
)

// ToolDocumentationIntegrator provides high-level integration between discovery and documentation
type ToolDocumentationIntegrator struct {
	discovery tools.ToolDiscovery
	config    GeneratorConfig
}

// NewToolDocumentationIntegrator creates a new integrator with the discovery system
func NewToolDocumentationIntegrator(discovery tools.ToolDiscovery, config GeneratorConfig) *ToolDocumentationIntegrator {
	return &ToolDocumentationIntegrator{
		discovery: discovery,
		config:    config,
	}
}

// GenerateDocsForAllTools generates documentation for all tools in the discovery system
func (i *ToolDocumentationIntegrator) GenerateDocsForAllTools(ctx context.Context) ([]Documentation, error) {
	toolInfos := i.discovery.ListTools()
	docs := make([]Documentation, len(toolInfos))

	for idx, toolInfo := range toolInfos {
		doc, err := GenerateToolDocumentation(toolInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate documentation for tool %s: %w", toolInfo.Name, err)
		}
		docs[idx] = doc
	}

	return docs, nil
}

// GenerateOpenAPIForAllTools creates an OpenAPI specification for all discovered tools
func (i *ToolDocumentationIntegrator) GenerateOpenAPIForAllTools(ctx context.Context) (*OpenAPISpec, error) {
	toolInfos := i.discovery.ListTools()
	return GenerateToolOpenAPI(ctx, toolInfos, i.config)
}

// GenerateMarkdownForAllTools creates markdown documentation for all discovered tools
func (i *ToolDocumentationIntegrator) GenerateMarkdownForAllTools(ctx context.Context) (string, error) {
	toolInfos := i.discovery.ListTools()
	return GenerateToolMarkdown(ctx, toolInfos, i.config)
}

// GenerateDocsForCategory generates documentation for tools in a specific category
func (i *ToolDocumentationIntegrator) GenerateDocsForCategory(ctx context.Context, category string) ([]Documentation, error) {
	toolInfos := i.discovery.ListByCategory(category)
	docs := make([]Documentation, len(toolInfos))

	for idx, toolInfo := range toolInfos {
		doc, err := GenerateToolDocumentation(toolInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate documentation for tool %s: %w", toolInfo.Name, err)
		}
		docs[idx] = doc
	}

	return docs, nil
}

// GenerateDocsForSearchQuery generates documentation for tools matching a search query
func (i *ToolDocumentationIntegrator) GenerateDocsForSearchQuery(ctx context.Context, query string) ([]Documentation, error) {
	toolInfos := i.discovery.SearchTools(query)
	docs := make([]Documentation, len(toolInfos))

	for idx, toolInfo := range toolInfos {
		doc, err := GenerateToolDocumentation(toolInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate documentation for tool %s: %w", toolInfo.Name, err)
		}
		docs[idx] = doc
	}

	return docs, nil
}

// GenerateOpenAPIForCategory creates OpenAPI spec for tools in a specific category
func (i *ToolDocumentationIntegrator) GenerateOpenAPIForCategory(ctx context.Context, category string) (*OpenAPISpec, error) {
	toolInfos := i.discovery.ListByCategory(category)

	// Update config title to reflect category
	categoryConfig := i.config
	categoryConfig.Title = fmt.Sprintf("%s - %s Category Tools", i.config.Title, category)

	return GenerateToolOpenAPI(ctx, toolInfos, categoryConfig)
}

// IntegrateWithToolHelp enhances the existing GetToolHelp with documentation formatting
func (i *ToolDocumentationIntegrator) IntegrateWithToolHelp(ctx context.Context, toolName string) (string, error) {
	// Get the basic help from discovery
	basicHelp, err := i.discovery.GetToolHelp(toolName)
	if err != nil {
		return "", err
	}

	// Try to get the tool info for enhanced documentation
	toolInfos := i.discovery.SearchTools(toolName)

	// Find exact match
	var targetTool tools.ToolInfo
	found := false
	for _, toolInfo := range toolInfos {
		if toolInfo.Name == toolName {
			targetTool = toolInfo
			found = true
			break
		}
	}

	if !found {
		// Return basic help if we can't find the tool
		return basicHelp, nil
	}

	// Generate enhanced documentation
	doc, err := GenerateToolDocumentation(targetTool)
	if err != nil {
		return basicHelp, nil // Fall back to basic help
	}

	// Format enhanced help
	var help strings.Builder
	help.WriteString(fmt.Sprintf("# %s\n\n", doc.Name))
	help.WriteString(fmt.Sprintf("**Description:** %s\n\n", doc.Description))

	if doc.LongDescription != "" {
		help.WriteString(fmt.Sprintf("**Usage:** %s\n\n", doc.LongDescription))
	}

	if doc.Category != "" {
		help.WriteString(fmt.Sprintf("**Category:** %s\n\n", doc.Category))
	}

	if len(doc.Tags) > 0 {
		help.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(doc.Tags, ", ")))
	}

	if doc.Version != "" {
		help.WriteString(fmt.Sprintf("**Version:** %s\n\n", doc.Version))
	}

	// Add schema information
	if inputSchema, exists := doc.Schemas["input"]; exists {
		help.WriteString("## Input Schema\n\n")
		help.WriteString(formatSchemaAsMarkdown(inputSchema))
		help.WriteString("\n")
	}

	if outputSchema, exists := doc.Schemas["output"]; exists {
		help.WriteString("## Output Schema\n\n")
		help.WriteString(formatSchemaAsMarkdown(outputSchema))
		help.WriteString("\n")
	}

	// Add examples
	if len(doc.Examples) > 0 {
		help.WriteString("## Examples\n\n")
		for _, example := range doc.Examples {
			help.WriteString(fmt.Sprintf("### %s\n\n", example.Name))
			if example.Description != "" {
				help.WriteString(fmt.Sprintf("%s\n\n", example.Description))
			}

			if example.Input != nil {
				help.WriteString("**Input:**\n```json\n")
				if inputJSON, err := prettyJSON(example.Input); err == nil {
					help.WriteString(inputJSON)
				}
				help.WriteString("\n```\n\n")
			}

			if example.Output != nil {
				help.WriteString("**Output:**\n```json\n")
				if outputJSON, err := prettyJSON(example.Output); err == nil {
					help.WriteString(outputJSON)
				}
				help.WriteString("\n```\n\n")
			}
		}
	}

	// Add original basic help as fallback section
	if basicHelp != "" {
		help.WriteString("## Additional Information\n\n")
		help.WriteString("```\n")
		help.WriteString(basicHelp)
		help.WriteString("\n```\n")
	}

	return help.String(), nil
}

// BatchGenerationOptions provides options for batch generation operations
type BatchGenerationOptions struct {
	// Categories to include (empty means all)
	Categories []string

	// Tags to filter by (empty means all)
	Tags []string

	// IncludeExamples whether to include examples in output
	IncludeExamples bool

	// IncludeSchemas whether to include schemas in output
	IncludeSchemas bool

	// GroupByCategory whether to group tools by category in output
	GroupByCategory bool

	// OutputFormat specifies the format: "openapi", "markdown", "json"
	OutputFormat string
}

// BatchGenerate performs batch generation with advanced filtering options
func (i *ToolDocumentationIntegrator) BatchGenerate(ctx context.Context, options BatchGenerationOptions) (interface{}, error) {
	// Get all tools
	allTools := i.discovery.ListTools()

	// Filter by categories if specified
	var filteredTools []tools.ToolInfo
	if len(options.Categories) > 0 {
		categorySet := make(map[string]bool)
		for _, cat := range options.Categories {
			categorySet[cat] = true
		}

		for _, tool := range allTools {
			if categorySet[tool.Category] {
				filteredTools = append(filteredTools, tool)
			}
		}
	} else {
		filteredTools = allTools
	}

	// Filter by tags if specified
	if len(options.Tags) > 0 {
		tagSet := make(map[string]bool)
		for _, tag := range options.Tags {
			tagSet[tag] = true
		}

		var tagFilteredTools []tools.ToolInfo
		for _, tool := range filteredTools {
			hasMatchingTag := false
			for _, toolTag := range tool.Tags {
				if tagSet[toolTag] {
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				tagFilteredTools = append(tagFilteredTools, tool)
			}
		}
		filteredTools = tagFilteredTools
	}

	// Update config based on options
	batchConfig := i.config
	batchConfig.IncludeExamples = options.IncludeExamples
	batchConfig.IncludeSchemas = options.IncludeSchemas
	if options.GroupByCategory {
		batchConfig.GroupBy = "category"
	}

	// Generate based on output format
	switch options.OutputFormat {
	case "openapi", "":
		return GenerateToolOpenAPI(ctx, filteredTools, batchConfig)
	case "markdown":
		return GenerateToolMarkdown(ctx, filteredTools, batchConfig)
	case "json":
		docs := make([]Documentation, len(filteredTools))
		for i, toolInfo := range filteredTools {
			doc, err := GenerateToolDocumentation(toolInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to generate documentation for tool %s: %w", toolInfo.Name, err)
			}
			docs[i] = doc
		}
		return docs, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", options.OutputFormat)
	}
}

// GetToolCategories returns all unique categories from discovered tools
func (i *ToolDocumentationIntegrator) GetToolCategories() []string {
	toolInfos := i.discovery.ListTools()
	categorySet := make(map[string]bool)

	for _, tool := range toolInfos {
		if tool.Category != "" {
			categorySet[tool.Category] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for category := range categorySet {
		categories = append(categories, category)
	}

	return categories
}

// GetToolTags returns all unique tags from discovered tools
func (i *ToolDocumentationIntegrator) GetToolTags() []string {
	toolInfos := i.discovery.ListTools()
	tagSet := make(map[string]bool)

	for _, tool := range toolInfos {
		for _, tag := range tool.Tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags
}

// Convenience functions for common operations

// GenerateToolsOpenAPI is a convenience function to generate OpenAPI spec for all tools
func GenerateToolsOpenAPI(ctx context.Context, config GeneratorConfig) (*OpenAPISpec, error) {
	discovery := tools.NewDiscovery()
	integrator := NewToolDocumentationIntegrator(discovery, config)
	return integrator.GenerateOpenAPIForAllTools(ctx)
}

// GenerateToolsMarkdown is a convenience function to generate markdown for all tools
func GenerateToolsMarkdown(ctx context.Context, config GeneratorConfig) (string, error) {
	discovery := tools.NewDiscovery()
	integrator := NewToolDocumentationIntegrator(discovery, config)
	return integrator.GenerateMarkdownForAllTools(ctx)
}

// GenerateToolsJSON is a convenience function to generate JSON docs for all tools
func GenerateToolsJSON(ctx context.Context, config GeneratorConfig) ([]Documentation, error) {
	discovery := tools.NewDiscovery()
	integrator := NewToolDocumentationIntegrator(discovery, config)
	return integrator.GenerateDocsForAllTools(ctx)
}

// Helper functions

// formatSchemaAsMarkdown formats a schema as markdown text
func formatSchemaAsMarkdown(schema *Schema) string {
	var builder strings.Builder

	if schema.Type != "" {
		builder.WriteString(fmt.Sprintf("- **Type**: %s\n", schema.Type))
	}

	if schema.Description != "" {
		builder.WriteString(fmt.Sprintf("- **Description**: %s\n", schema.Description))
	}

	if len(schema.Required) > 0 {
		builder.WriteString(fmt.Sprintf("- **Required**: %s\n", strings.Join(schema.Required, ", ")))
	}

	if len(schema.Properties) > 0 {
		builder.WriteString("- **Properties**:\n")
		for propName, propSchema := range schema.Properties {
			builder.WriteString(fmt.Sprintf("  - **%s**: %s", propName, propSchema.Type))
			if propSchema.Description != "" {
				builder.WriteString(fmt.Sprintf(" - %s", propSchema.Description))
			}
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// prettyJSON formats an interface as pretty-printed JSON
func prettyJSON(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
