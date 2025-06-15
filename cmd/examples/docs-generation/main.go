package main

// ABOUTME: Example demonstrating documentation generation integration with tool discovery
// ABOUTME: Shows OpenAPI, Markdown, and JSON generation for all discovered tools

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/docs"
)

func main() {
	ctx := context.Background()

	log.Println("Go-LLMs Documentation Generation Example")
	log.Println("=======================================")

	// Create documentation config
	config := docs.GeneratorConfig{
		Title:           "Go-LLMs Tool Documentation",
		Description:     "Comprehensive documentation for all available tools in the go-llms library",
		Version:         "1.0.0",
		BaseURL:         "https://api.go-llms.example.com",
		GroupBy:         "category",
		IncludeExamples: true,
		IncludeSchemas:  true,
		CustomMetadata: map[string]interface{}{
			"generator": "go-llms docs integration",
			"generated": "2024",
		},
	}

	// Initialize tool discovery
	discovery := tools.NewDiscovery()
	integrator := docs.NewToolDocumentationIntegrator(discovery, config)

	// Example 1: List all available tools
	log.Println("\nExample 1: Discovering available tools")
	log.Println("=====================================")

	toolList := discovery.ListTools()
	log.Printf("Found %d tools in the discovery system\n", len(toolList))

	// Show first few tools
	for i, tool := range toolList {
		if i >= 5 { // Show only first 5 tools
			log.Printf("... and %d more tools\n", len(toolList)-5)
			break
		}
		log.Printf("- %s (%s): %s\n", tool.Name, tool.Category, tool.Description)
	}

	// Example 2: Generate OpenAPI specification for all tools
	log.Println("\nExample 2: Generating OpenAPI specification")
	log.Println("==========================================")

	openAPISpec, err := integrator.GenerateOpenAPIForAllTools(ctx)
	if err != nil {
		log.Fatalf("Failed to generate OpenAPI spec: %v", err)
	}

	// Save to file
	openAPIJSON, err := json.MarshalIndent(openAPISpec, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal OpenAPI spec: %v", err)
	}

	err = os.WriteFile("tools-openapi.json", openAPIJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write OpenAPI file: %v", err)
	}

	log.Printf("✅ Generated OpenAPI specification with %d paths\n", len(openAPISpec.Paths))
	log.Println("📄 Saved to: tools-openapi.json")

	// Example 3: Generate Markdown documentation
	log.Println("\nExample 3: Generating Markdown documentation")
	log.Println("============================================")

	markdown, err := integrator.GenerateMarkdownForAllTools(ctx)
	if err != nil {
		log.Fatalf("Failed to generate Markdown: %v", err)
	}

	err = os.WriteFile("tools-documentation.md", []byte(markdown), 0644)
	if err != nil {
		log.Fatalf("Failed to write Markdown file: %v", err)
	}

	log.Printf("✅ Generated Markdown documentation (%d characters)\n", len(markdown))
	log.Println("📄 Saved to: tools-documentation.md")

	// Example 4: Generate documentation for specific category
	log.Println("\nExample 4: Category-specific documentation")
	log.Println("=========================================")

	categories := integrator.GetToolCategories()
	log.Printf("Available categories: %v\n", categories)

	if len(categories) > 0 {
		category := categories[0]
		log.Printf("Generating docs for category: %s\n", category)

		categoryDocs, err := integrator.GenerateDocsForCategory(ctx, category)
		if err != nil {
			log.Fatalf("Failed to generate category docs: %v", err)
		}

		log.Printf("✅ Generated %d documentation items for category '%s'\n", len(categoryDocs), category)
	}

	// Example 5: Search-based documentation generation
	log.Println("\nExample 5: Search-based documentation")
	log.Println("====================================")

	searchQuery := "file"
	log.Printf("Searching for tools matching: '%s'\n", searchQuery)

	searchDocs, err := integrator.GenerateDocsForSearchQuery(ctx, searchQuery)
	if err != nil {
		log.Fatalf("Failed to generate search docs: %v", err)
	}

	log.Printf("✅ Found %d tools matching search query '%s'\n", len(searchDocs), searchQuery)
	for _, doc := range searchDocs {
		log.Printf("- %s: %s\n", doc.Name, doc.Description)
	}

	// Example 6: Enhanced tool help integration
	log.Println("\nExample 6: Enhanced tool help")
	log.Println("=============================")

	if len(toolList) > 0 {
		toolName := toolList[0].Name
		log.Printf("Getting enhanced help for tool: %s\n", toolName)

		enhancedHelp, err := integrator.IntegrateWithToolHelp(ctx, toolName)
		if err != nil {
			log.Printf("Failed to get enhanced help: %v", err)
		} else {
			// Show first 200 characters of help
			helpPreview := enhancedHelp
			if len(helpPreview) > 200 {
				helpPreview = helpPreview[:200] + "..."
			}
			log.Printf("Enhanced help preview:\n%s\n", helpPreview)
		}
	}

	// Example 7: Batch generation with options
	log.Println("\nExample 7: Batch generation with custom options")
	log.Println("===============================================")

	batchOptions := docs.BatchGenerationOptions{
		Categories:      []string{}, // All categories
		Tags:            []string{}, // All tags
		IncludeExamples: true,
		IncludeSchemas:  true,
		GroupByCategory: true,
		OutputFormat:    "json",
	}

	batchResult, err := integrator.BatchGenerate(ctx, batchOptions)
	if err != nil {
		log.Fatalf("Failed to perform batch generation: %v", err)
	}

	if docList, ok := batchResult.([]docs.Documentation); ok {
		log.Printf("✅ Batch generated %d documentation items\n", len(docList))

		// Save batch result
		batchJSON, err := json.MarshalIndent(docList, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal batch result: %v", err)
		}

		err = os.WriteFile("tools-batch-docs.json", batchJSON, 0644)
		if err != nil {
			log.Fatalf("Failed to write batch file: %v", err)
		}
		log.Println("📄 Saved to: tools-batch-docs.json")
	}

	// Example 8: Using convenience functions
	log.Println("\nExample 8: Using convenience functions")
	log.Println("=====================================")

	// Generate OpenAPI using convenience function
	convenienctConfig := docs.GeneratorConfig{
		Title:       "Go-LLMs API",
		Description: "Tool execution API",
		Version:     "1.0.0",
	}

	convenientOpenAPI, err := docs.GenerateToolsOpenAPI(ctx, convenienctConfig)
	if err != nil {
		log.Fatalf("Failed to generate via convenience function: %v", err)
	}

	log.Printf("✅ Generated OpenAPI via convenience function with %d paths\n", len(convenientOpenAPI.Paths))

	// Example 9: Demonstrate tool schema conversion
	log.Println("\nExample 9: Individual tool documentation")
	log.Println("=======================================")

	if len(toolList) > 0 {
		tool := toolList[0]
		log.Printf("Converting tool info to documentation: %s\n", tool.Name)

		doc, err := docs.GenerateToolDocumentation(tool)
		if err != nil {
			log.Printf("Failed to convert tool: %v", err)
		} else {
			log.Printf("✅ Generated documentation for tool '%s'\n", doc.Name)
			log.Printf("   Category: %s\n", doc.Category)
			log.Printf("   Tags: %v\n", doc.Tags)
			log.Printf("   Has input schema: %t\n", doc.Schemas != nil && doc.Schemas["input"] != nil)
			log.Printf("   Has output schema: %t\n", doc.Schemas != nil && doc.Schemas["output"] != nil)
			log.Printf("   Examples: %d\n", len(doc.Examples))
		}

		// Convert to OpenAPI operation
		operation, err := docs.ConvertToolInfoToOpenAPIOperation(tool)
		if err != nil {
			log.Printf("Failed to convert to OpenAPI operation: %v", err)
		} else {
			log.Printf("✅ Generated OpenAPI operation for tool '%s'\n", tool.Name)
			log.Printf("   Operation ID: %s\n", operation.OperationID)
			log.Printf("   Has request body: %t\n", operation.RequestBody != nil)
			log.Printf("   Response codes: %v\n", getKeys(operation.Responses))
		}
	}

	log.Println("\n🎉 Documentation generation examples completed!")
	log.Println("Generated files:")
	log.Println("- tools-openapi.json: Complete OpenAPI specification")
	log.Println("- tools-documentation.md: Markdown documentation")
	log.Println("- tools-batch-docs.json: Batch-generated JSON documentation")
	log.Println("\nThese files demonstrate the integration between tool discovery and documentation generation.")
}

// Helper function to get map keys
func getKeys(m map[string]*docs.Response) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
