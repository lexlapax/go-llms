// Example demonstrating schema repository usage
package main

// ABOUTME: Example showing how to use schema repositories for storage
// ABOUTME: Demonstrates both in-memory and file-based implementations

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
)

func main() {
	fmt.Println("Schema Repository Example")
	fmt.Println("========================")

	// Example 1: In-Memory Repository
	fmt.Println("\n1. In-Memory Repository Example")
	demoInMemoryRepository()

	// Example 2: File-Based Repository
	fmt.Println("\n2. File-Based Repository Example")
	demoFileRepository()

	// Example 3: Schema Versioning
	fmt.Println("\n3. Schema Versioning Example")
	demoSchemaVersioning()

	// Example 4: Export/Import
	fmt.Println("\n4. Export/Import Example")
	demoExportImport()
}

func demoInMemoryRepository() {
	repo := repository.NewInMemorySchemaRepository()

	// Create a user schema
	userSchema := &domain.Schema{
		Type:        "object",
		Title:       "User",
		Description: "User profile schema",
		Properties: map[string]domain.Property{
			"id": {
				Type:        "string",
				Format:      "uuid",
				Description: "Unique identifier",
			},
			"name": {
				Type:        "string",
				MinLength:   intPtr(1),
				MaxLength:   intPtr(100),
				Description: "User's full name",
			},
			"email": {
				Type:        "string",
				Format:      "email",
				Description: "Email address",
			},
			"age": {
				Type:        "integer",
				Minimum:     float64Ptr(0),
				Maximum:     float64Ptr(150),
				Description: "User's age",
			},
		},
		Required: []string{"id", "name", "email"},
	}

	// Save schema
	if err := repo.Save("user-v1", userSchema); err != nil {
		log.Fatalf("Failed to save schema: %v", err)
	}
	fmt.Println("✓ Saved user schema")

	// Retrieve schema
	retrieved, err := repo.Get("user-v1")
	if err != nil {
		log.Fatalf("Failed to get schema: %v", err)
	}
	fmt.Printf("✓ Retrieved schema: %s\n", retrieved.Title)

	// List all schemas
	ids := repo.List()
	fmt.Printf("✓ Total schemas: %d\n", len(ids))
}

func demoFileRepository() {
	// Create temporary directory for demo
	tmpDir := filepath.Join(os.TempDir(), "schema-repo-demo")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	repo, err := repository.NewFileSchemaRepository(tmpDir)
	if err != nil {
		log.Fatalf("Failed to create file repository: %v", err)
	}

	// Create a product schema
	productSchema := &domain.Schema{
		Type:  "object",
		Title: "Product",
		Properties: map[string]domain.Property{
			"sku": {
				Type:    "string",
				Pattern: "^[A-Z]{3}-[0-9]{4}$",
			},
			"name": {
				Type:      "string",
				MinLength: intPtr(1),
			},
			"price": {
				Type:    "number",
				Minimum: float64Ptr(0),
			},
			"tags": {
				Type: "array",
				Items: &domain.Property{
					Type: "string",
				},
				UniqueItems: boolPtr(true),
			},
		},
		Required: []string{"sku", "name", "price"},
	}

	// Save schema
	if err := repo.Save("product", productSchema); err != nil {
		log.Fatalf("Failed to save schema: %v", err)
	}
	fmt.Println("✓ Saved product schema to file")

	// Verify files were created
	schemaDir := filepath.Join(tmpDir, "product")
	files, _ := os.ReadDir(schemaDir)
	fmt.Printf("✓ Created files: ")
	for i, file := range files {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(file.Name())
	}
	fmt.Println()
}

func demoSchemaVersioning() {
	repo := repository.NewInMemorySchemaRepository()

	// Version 1: Basic schema
	v1 := &domain.Schema{
		Type:        "object",
		Description: "API Request v1",
		Properties: map[string]domain.Property{
			"method": {Type: "string", Enum: []string{"GET", "POST"}},
			"path":   {Type: "string"},
		},
		Required: []string{"method", "path"},
	}

	if err := repo.Save("api-request", v1); err != nil {
		log.Fatalf("Failed to save version 1: %v", err)
	}
	fmt.Println("✓ Saved version 1")

	// Version 2: Add headers
	v2 := &domain.Schema{
		Type:        "object",
		Description: "API Request v2 - Added headers",
		Properties: map[string]domain.Property{
			"method": {Type: "string", Enum: []string{"GET", "POST", "PUT", "DELETE"}},
			"path":   {Type: "string"},
			"headers": {
				Type:                 "object",
				AdditionalProperties: boolPtr(true),
			},
		},
		Required: []string{"method", "path"},
	}

	if err := repo.Save("api-request", v2); err != nil {
		log.Fatalf("Failed to save version 2: %v", err)
	}
	fmt.Println("✓ Saved version 2")

	// Version 3: Add body
	v3 := &domain.Schema{
		Type:        "object",
		Description: "API Request v3 - Added body",
		Properties: map[string]domain.Property{
			"method": {Type: "string", Enum: []string{"GET", "POST", "PUT", "DELETE", "PATCH"}},
			"path":   {Type: "string"},
			"headers": {
				Type:                 "object",
				AdditionalProperties: boolPtr(true),
			},
			"body": {
				Type:                 "object",
				AdditionalProperties: boolPtr(true),
			},
		},
		Required: []string{"method", "path"},
	}

	if err := repo.Save("api-request", v3); err != nil {
		log.Fatalf("Failed to save version 3: %v", err)
	}
	fmt.Println("✓ Saved version 3")

	// List versions
	versions, _ := repo.ListVersions("api-request")
	fmt.Printf("✓ Total versions: %d\n", len(versions))

	// Get specific version
	v2Retrieved, _ := repo.GetVersion("api-request", 2)
	fmt.Printf("✓ Version 2 description: %s\n", v2Retrieved.Description)

	// Set current version back to v2
	if err := repo.SetCurrentVersion("api-request", 2); err != nil {
		log.Printf("Failed to set current version: %v", err)
	}
	current, _ := repo.Get("api-request")
	fmt.Printf("✓ Current version description: %s\n", current.Description)
}

func demoExportImport() {
	// Create source repository with data
	source := repository.NewInMemorySchemaRepository()

	schemas := map[string]*domain.Schema{
		"config": {
			Type:  "object",
			Title: "Configuration",
			Properties: map[string]domain.Property{
				"debug":   {Type: "boolean"},
				"timeout": {Type: "integer", Minimum: float64Ptr(0)},
			},
		},
		"response": {
			Type:  "object",
			Title: "API Response",
			Properties: map[string]domain.Property{
				"status": {Type: "integer"},
				"data":   {Type: "object"},
			},
		},
	}

	for id, schema := range schemas {
		if err := source.Save(id, schema); err != nil {
			log.Printf("Failed to save schema %s: %v", id, err)
		}
	}
	fmt.Printf("✓ Created %d schemas in source repository\n", source.Count())

	// Export schemas
	exported, err := source.Export()
	if err != nil {
		log.Fatalf("Failed to export: %v", err)
	}
	fmt.Printf("✓ Exported %d bytes of data\n", len(exported))

	// Import into new repository
	target := repository.NewInMemorySchemaRepository()
	if err := target.Import(exported); err != nil {
		log.Fatalf("Failed to import: %v", err)
	}
	fmt.Printf("✓ Imported into target repository\n")

	// Verify import
	fmt.Printf("✓ Target repository count: %d\n", target.Count())
	for id := range schemas {
		if _, err := target.Get(id); err != nil {
			fmt.Printf("✗ Failed to find schema: %s\n", id)
		}
	}

	// Pretty print one schema
	config, _ := target.Get("config")
	data, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println("\nExported/Imported Config Schema:")
	fmt.Println(string(data))
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}
