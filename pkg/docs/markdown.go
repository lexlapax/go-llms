package docs

// ABOUTME: Markdown documentation generator for tools and components
// ABOUTME: Generates human-readable Markdown documentation from Documentable items

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MarkdownGenerator generates Markdown documentation.
// It converts Documentable items into human-readable markdown format
// suitable for documentation sites, README files, and other text-based
// documentation needs.
type MarkdownGenerator struct {
	config GeneratorConfig
}

// NewMarkdownGenerator creates a new Markdown generator.
//
// Parameters:
//   - config: Configuration for the generator including title, version, and formatting options
//
// Returns a configured MarkdownGenerator instance.
func NewMarkdownGenerator(config GeneratorConfig) *MarkdownGenerator {
	return &MarkdownGenerator{
		config: config,
	}
}

// GenerateOpenAPI is not implemented by MarkdownGenerator.
// This method exists to satisfy the Generator interface but returns an error
// as markdown generation doesn't produce OpenAPI specifications.
//
// Parameters:
//   - ctx: The context (unused)
//   - items: The items to document (unused)
//
// Returns an error indicating OpenAPI generation is not supported.
func (g *MarkdownGenerator) GenerateOpenAPI(ctx context.Context, items []Documentable) (*OpenAPISpec, error) {
	return nil, fmt.Errorf("OpenAPI generation not supported by MarkdownGenerator")
}

// GenerateMarkdown generates Markdown documentation from documentable items.
// It creates a comprehensive markdown document with table of contents,
// grouped sections, metadata tables, schemas, and examples.
//
// Parameters:
//   - ctx: The context for the operation
//   - items: The documentable items to include
//
// Returns formatted markdown string or an error.
func (g *MarkdownGenerator) GenerateMarkdown(ctx context.Context, items []Documentable) (string, error) {
	var builder strings.Builder

	// Add title and description
	fmt.Fprintf(&builder, "# %s\n\n", g.config.Title)
	if g.config.Description != "" {
		fmt.Fprintf(&builder, "%s\n\n", g.config.Description)
	}

	// Add version info
	if g.config.Version != "" {
		fmt.Fprintf(&builder, "**Version:** %s\n\n", g.config.Version)
	}

	// Group items by category if specified
	groups := g.groupItems(items)

	// Sort group names for consistent output
	var groupNames []string
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	// Generate table of contents
	if len(groupNames) > 1 {
		builder.WriteString("## Table of Contents\n\n")
		for _, groupName := range groupNames {
			anchor := strings.ToLower(strings.ReplaceAll(groupName, " ", "-"))
			fmt.Fprintf(&builder, "- [%s](#%s)\n", groupName, anchor)
		}
		builder.WriteString("\n")
	}

	// Process each group
	for _, groupName := range groupNames {
		groupItems := groups[groupName]
		if err := g.addGroupToMarkdown(&builder, groupName, groupItems); err != nil {
			return "", fmt.Errorf("failed to add group %s: %w", groupName, err)
		}
	}

	return builder.String(), nil
}

// GenerateJSON generates JSON representation of the documentation.
// It creates a structured JSON document containing all documentation items
// along with generator metadata.
//
// Parameters:
//   - ctx: The context for the operation
//   - items: The documentable items to include
//
// Returns JSON bytes or an error.
func (g *MarkdownGenerator) GenerateJSON(ctx context.Context, items []Documentable) ([]byte, error) {
	docs := make([]Documentation, 0, len(items))
	for _, item := range items {
		docs = append(docs, item.GetDocumentation())
	}

	result := map[string]interface{}{
		"title":       g.config.Title,
		"description": g.config.Description,
		"version":     g.config.Version,
		"items":       docs,
		"metadata":    g.config.CustomMetadata,
	}

	return json.MarshalIndent(result, "", "  ")
}

// groupItems groups documentable items by category.
// It supports grouping by category, type, or a default grouping.
// Items within each group are sorted alphabetically by name.
//
// Parameters:
//   - items: The documentable items to group
//
// Returns a map of group names to sorted items.
func (g *MarkdownGenerator) groupItems(items []Documentable) map[string][]Documentable {
	groups := make(map[string][]Documentable)

	for _, item := range items {
		doc := item.GetDocumentation()
		category := ""

		switch g.config.GroupBy {
		case "category":
			category = doc.Category
		case "type":
			// Could be extended to group by type
			category = "Tools"
		default:
			category = "Components"
		}

		if category == "" {
			category = "Uncategorized"
		}

		groups[category] = append(groups[category], item)
	}

	// Sort items within each group
	for _, groupItems := range groups {
		sort.Slice(groupItems, func(i, j int) bool {
			return groupItems[i].GetDocumentation().Name < groupItems[j].GetDocumentation().Name
		})
	}

	return groups
}

// addGroupToMarkdown adds a group of items to the markdown.
// It creates a section header for the group and adds all items
// within that group to the documentation.
//
// Parameters:
//   - builder: The string builder for output
//   - groupName: The name of the group
//   - items: The items in this group
//
// Returns an error if any item fails to render.
func (g *MarkdownGenerator) addGroupToMarkdown(builder *strings.Builder, groupName string, items []Documentable) error {
	if groupName != "default" && groupName != "" {
		fmt.Fprintf(builder, "## %s\n\n", groupName)
	}

	for _, item := range items {
		if err := g.addItemToMarkdown(builder, item); err != nil {
			return fmt.Errorf("failed to add item %s: %w", item.GetDocumentation().Name, err)
		}
	}

	return nil
}

// addItemToMarkdown adds a single documentable item to the markdown.
// It renders the item with title, descriptions, metadata, schemas,
// and examples based on the generator configuration.
//
// Parameters:
//   - builder: The string builder for output
//   - item: The documentable item to render
//
// Returns an error if rendering fails.
func (g *MarkdownGenerator) addItemToMarkdown(builder *strings.Builder, item Documentable) error {
	doc := item.GetDocumentation()

	// Add item title
	fmt.Fprintf(builder, "### %s\n\n", doc.Name)

	// Add deprecated notice
	if doc.Deprecated {
		builder.WriteString("**⚠️ DEPRECATED**")
		if doc.DeprecationNote != "" {
			fmt.Fprintf(builder, ": %s", doc.DeprecationNote)
		}
		builder.WriteString("\n\n")
	}

	// Add description
	if doc.Description != "" {
		fmt.Fprintf(builder, "%s\n\n", doc.Description)
	}

	// Add long description
	if doc.LongDescription != "" {
		fmt.Fprintf(builder, "%s\n\n", doc.LongDescription)
	}

	// Add metadata table
	g.addMetadataTable(builder, doc)

	// Add schema information
	if g.config.IncludeSchemas {
		g.addSchemaSection(builder, doc)
	}

	// Add examples
	if g.config.IncludeExamples && len(doc.Examples) > 0 {
		g.addExamplesSection(builder, doc.Examples)
	}

	builder.WriteString("---\n\n")

	return nil
}

// addMetadataTable adds a metadata table for the item.
// It creates a markdown table containing category, tags, version,
// and any custom metadata fields.
//
// Parameters:
//   - builder: The string builder for output
//   - doc: The documentation containing metadata
func (g *MarkdownGenerator) addMetadataTable(builder *strings.Builder, doc Documentation) {
	var hasMetadata bool
	var rows []string

	if doc.Category != "" {
		rows = append(rows, fmt.Sprintf("| **Category** | %s |", doc.Category))
		hasMetadata = true
	}

	if len(doc.Tags) > 0 {
		tags := strings.Join(doc.Tags, ", ")
		rows = append(rows, fmt.Sprintf("| **Tags** | %s |", tags))
		hasMetadata = true
	}

	if doc.Version != "" {
		rows = append(rows, fmt.Sprintf("| **Version** | %s |", doc.Version))
		hasMetadata = true
	}

	// Add custom metadata
	if doc.Metadata != nil {
		for key, value := range doc.Metadata {
			if valueStr, ok := value.(string); ok {
				rows = append(rows, fmt.Sprintf("| **%s** | %s |", key, valueStr))
				hasMetadata = true
			}
		}
	}

	if hasMetadata {
		builder.WriteString("| Property | Value |\n")
		builder.WriteString("|----------|-------|\n")
		for _, row := range rows {
			fmt.Fprintf(builder, "%s\n", row)
		}
		builder.WriteString("\n")
	}
}

// addSchemaSection adds schema information.
// It renders both single schemas and named schema collections
// in a hierarchical markdown format.
//
// Parameters:
//   - builder: The string builder for output
//   - doc: The documentation containing schemas
func (g *MarkdownGenerator) addSchemaSection(builder *strings.Builder, doc Documentation) {
	if doc.Schema != nil {
		builder.WriteString("#### Schema\n\n")
		g.addSchemaMarkdown(builder, doc.Schema, 0)
		builder.WriteString("\n")
	}

	if doc.Schemas != nil {
		for schemaName, schema := range doc.Schemas {
			caser := cases.Title(language.English)
			fmt.Fprintf(builder, "#### %s Schema\n\n", caser.String(schemaName))
			g.addSchemaMarkdown(builder, schema, 0)
			builder.WriteString("\n")
		}
	}
}

// addSchemaMarkdown adds schema information in markdown format.
// It recursively renders schema properties with proper indentation
// to show the hierarchical structure.
//
// Parameters:
//   - builder: The string builder for output
//   - schema: The schema to render
//   - indent: The current indentation level
func (g *MarkdownGenerator) addSchemaMarkdown(builder *strings.Builder, schema *Schema, indent int) {
	indentStr := strings.Repeat("  ", indent)

	if schema.Type != "" {
		fmt.Fprintf(builder, "%s- **Type**: %s\n", indentStr, schema.Type)
	}

	if schema.Description != "" {
		fmt.Fprintf(builder, "%s- **Description**: %s\n", indentStr, schema.Description)
	}

	if len(schema.Required) > 0 {
		fmt.Fprintf(builder, "%s- **Required**: %s\n", indentStr, strings.Join(schema.Required, ", "))
	}

	if len(schema.Enum) > 0 {
		enumStrs := make([]string, len(schema.Enum))
		for i, e := range schema.Enum {
			enumStrs[i] = fmt.Sprintf("`%v`", e)
		}
		fmt.Fprintf(builder, "%s- **Enum**: %s\n", indentStr, strings.Join(enumStrs, ", "))
	}

	if schema.Default != nil {
		fmt.Fprintf(builder, "%s- **Default**: `%v`\n", indentStr, schema.Default)
	}

	if len(schema.Properties) > 0 {
		fmt.Fprintf(builder, "%s- **Properties**:\n", indentStr)
		for propName, propSchema := range schema.Properties {
			fmt.Fprintf(builder, "%s  - **%s**:\n", indentStr, propName)
			g.addSchemaMarkdown(builder, propSchema, indent+2)
		}
	}

	if schema.Items != nil {
		fmt.Fprintf(builder, "%s- **Items**:\n", indentStr)
		g.addSchemaMarkdown(builder, schema.Items, indent+1)
	}
}

// addExamplesSection adds examples section.
// It renders each example with name, description, and code/input/output
// in a clear, readable format with syntax highlighting hints.
//
// Parameters:
//   - builder: The string builder for output
//   - examples: The examples to render
func (g *MarkdownGenerator) addExamplesSection(builder *strings.Builder, examples []Example) {
	builder.WriteString("#### Examples\n\n")

	for i, example := range examples {
		if example.Name != "" {
			fmt.Fprintf(builder, "##### Example %d: %s\n\n", i+1, example.Name)
		} else {
			fmt.Fprintf(builder, "##### Example %d\n\n", i+1)
		}

		if example.Description != "" {
			fmt.Fprintf(builder, "%s\n\n", example.Description)
		}

		if example.Code != "" {
			lang := example.Language
			if lang == "" {
				lang = "json"
			}
			fmt.Fprintf(builder, "```%s\n%s\n```\n\n", lang, example.Code)
		} else {
			// Show input/output as JSON
			if example.Input != nil {
				builder.WriteString("**Input:**\n```json\n")
				if inputJSON, err := json.MarshalIndent(example.Input, "", "  "); err == nil {
					builder.WriteString(string(inputJSON))
				} else {
					fmt.Fprintf(builder, "%v", example.Input)
				}
				builder.WriteString("\n```\n\n")
			}

			if example.Output != nil {
				builder.WriteString("**Output:**\n```json\n")
				if outputJSON, err := json.MarshalIndent(example.Output, "", "  "); err == nil {
					builder.WriteString(string(outputJSON))
				} else {
					fmt.Fprintf(builder, "%v", example.Output)
				}
				builder.WriteString("\n```\n\n")
			}
		}
	}
}
