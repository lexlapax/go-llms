package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestFileSchemaRepository_CRUD(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	repo, err := NewFileSchemaRepository(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Test Create/Save
	schema := &domain.Schema{
		Type:        "object",
		Title:       "Test Schema",
		Description: "A test schema",
		Properties: map[string]domain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name"},
	}

	// Save schema
	err = repo.Save("test-id", schema)
	if err != nil {
		t.Fatalf("Failed to save schema: %v", err)
	}

	// Verify files were created
	schemaDir := filepath.Join(tmpDir, "test-id")
	if _, err := os.Stat(schemaDir); os.IsNotExist(err) {
		t.Error("Schema directory was not created")
	}

	// Test Get
	retrieved, err := repo.Get("test-id")
	if err != nil {
		t.Fatalf("Failed to get schema: %v", err)
	}

	if retrieved.Title != schema.Title {
		t.Errorf("Expected title %s, got %s", schema.Title, retrieved.Title)
	}

	// Test Update (creates new version)
	schema.Description = "Updated description"
	err = repo.Save("test-id", schema)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}

	retrieved, err = repo.Get("test-id")
	if err != nil {
		t.Fatalf("Failed to get updated schema: %v", err)
	}

	if retrieved.Description != "Updated description" {
		t.Errorf("Expected updated description, got %s", retrieved.Description)
	}

	// Test Delete
	err = repo.Delete("test-id")
	if err != nil {
		t.Fatalf("Failed to delete schema: %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(schemaDir); !os.IsNotExist(err) {
		t.Error("Schema directory was not deleted")
	}

	_, err = repo.Get("test-id")
	if err == nil {
		t.Error("Expected error when getting deleted schema")
	}
}

func TestFileSchemaRepository_Versioning(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileSchemaRepository(tmpDir)

	// Create multiple versions
	for i := 1; i <= 3; i++ {
		schema := &domain.Schema{
			Type:        "object",
			Description: fmt.Sprintf("Version %d", i),
		}
		err := repo.Save("versioned", schema)
		if err != nil {
			t.Fatalf("Failed to save version %d: %v", i, err)
		}
	}

	// Check versions
	versions, err := repo.ListVersions("versioned")
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}

	if len(versions) != 3 {
		t.Fatalf("Expected 3 versions, got %d", len(versions))
	}

	// Verify version numbers
	for i, v := range versions {
		if v != i+1 {
			t.Errorf("Expected version %d, got %d", i+1, v)
		}
	}

	// Get specific version
	v2, err := repo.GetVersion("versioned", 2)
	if err != nil {
		t.Fatalf("Failed to get version 2: %v", err)
	}

	if v2.Description != "Version 2" {
		t.Errorf("Expected 'Version 2', got %s", v2.Description)
	}

	// Set current version
	err = repo.SetCurrentVersion("versioned", 1)
	if err != nil {
		t.Fatalf("Failed to set current version: %v", err)
	}

	current, err := repo.Get("versioned")
	if err != nil {
		t.Fatalf("Failed to get current version: %v", err)
	}

	if current.Description != "Version 1" {
		t.Errorf("Expected 'Version 1' as current, got %s", current.Description)
	}
}

func TestFileSchemaRepository_List(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileSchemaRepository(tmpDir)

	// Empty list
	ids, err := repo.List()
	if err != nil {
		t.Fatalf("Failed to list schemas: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("Expected empty list, got %d items", len(ids))
	}

	// Add schemas
	expected := []string{"id1", "id2", "id3"}
	for _, id := range expected {
		if err := repo.Save(id, &domain.Schema{Type: "object"}); err != nil {
			t.Errorf("Failed to save schema %s: %v", id, err)
		}
	}

	// Check list
	ids, err = repo.List()
	if err != nil {
		t.Fatalf("Failed to list schemas: %v", err)
	}

	if len(ids) != len(expected) {
		t.Errorf("Expected %d items, got %d", len(expected), len(ids))
	}

	// Verify all IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, expectedID := range expected {
		if !idMap[expectedID] {
			t.Errorf("Expected ID %s not found in list", expectedID)
		}
	}
}

func TestFileSchemaRepository_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileSchemaRepository(tmpDir)
	var wg sync.WaitGroup
	numGoroutines := 50 // Reduced for file system operations

	// Concurrent writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			schema := &domain.Schema{
				Type:  "object",
				Title: fmt.Sprintf("Schema %d", idx),
			}
			err := repo.Save(fmt.Sprintf("concurrent-%d", idx), schema)
			if err != nil {
				t.Errorf("Failed to save schema %d: %v", idx, err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all schemas were saved
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("Failed to count schemas: %v", err)
	}
	if count != numGoroutines {
		t.Errorf("Expected %d schemas, got %d", numGoroutines, count)
	}

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_, err := repo.Get(fmt.Sprintf("concurrent-%d", idx))
			if err != nil {
				t.Errorf("Failed to get schema %d: %v", idx, err)
			}
		}(i)
	}
	wg.Wait()
}

func TestFileSchemaRepository_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileSchemaRepository(tmpDir)

	// Empty ID
	err := repo.Save("", &domain.Schema{})
	if err == nil {
		t.Error("Expected error for empty ID")
	}

	// Nil schema
	err = repo.Save("test", nil)
	if err == nil {
		t.Error("Expected error for nil schema")
	}

	// Non-existent schema
	_, err = repo.Get("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent schema")
	}

	// Invalid version
	if err := repo.Save("test", &domain.Schema{Type: "object"}); err != nil {
		t.Fatalf("Failed to save test schema: %v", err)
	}
	_, err = repo.GetVersion("test", 0)
	if err == nil {
		t.Error("Expected error for version 0")
	}

	_, err = repo.GetVersion("test", 999)
	if err == nil {
		t.Error("Expected error for invalid version")
	}

	// Delete non-existent
	err = repo.Delete("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent schema")
	}

	// Directory traversal protection
	err = repo.Save("../escape", &domain.Schema{Type: "object"})
	if err != nil {
		t.Fatalf("Failed to save schema with special chars: %v", err)
	}

	// Check that it was saved in the correct location (../ becomes _)
	escapedDir := filepath.Join(tmpDir, "__escape")
	if _, err := os.Stat(escapedDir); os.IsNotExist(err) {
		t.Error("Schema with special chars not saved correctly")
	}
}

func TestFileSchemaRepository_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileSchemaRepository(tmpDir)

	// Add schemas
	for i := 0; i < 5; i++ {
		if err := repo.Save(fmt.Sprintf("schema-%d", i), &domain.Schema{Type: "object"}); err != nil {
			t.Errorf("Failed to save schema-%d: %v", i, err)
		}
	}

	count, _ := repo.Count()
	if count != 5 {
		t.Errorf("Expected 5 schemas, got %d", count)
	}

	// Clear
	err := repo.Clear()
	if err != nil {
		t.Fatalf("Failed to clear: %v", err)
	}

	count, _ = repo.Count()
	if count != 0 {
		t.Errorf("Expected 0 schemas after clear, got %d", count)
	}

	// Verify schemas are gone
	_, err = repo.Get("schema-0")
	if err == nil {
		t.Error("Expected error when getting schema after clear")
	}
}

func TestFileSchemaRepository_ComplexSchema(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileSchemaRepository(tmpDir)

	// Create a complex schema with nested properties and conditional validation
	complexSchema := &domain.Schema{
		Type:  "object",
		Title: "Complex Schema",
		Properties: map[string]domain.Property{
			"user": {
				Type: "object",
				Properties: map[string]domain.Property{
					"name": {Type: "string", MinLength: intPtr(1)},
					"age":  {Type: "integer", Minimum: float64Ptr(0)},
				},
				Required: []string{"name"},
			},
			"items": {
				Type: "array",
				Items: &domain.Property{
					Type: "object",
					Properties: map[string]domain.Property{
						"id":    {Type: "string"},
						"value": {Type: "number"},
					},
				},
			},
		},
		If: &domain.Schema{
			Properties: map[string]domain.Property{
				"type": {Type: "string", Enum: []string{"premium"}},
			},
		},
		Then: &domain.Schema{
			Required: []string{"items"},
		},
	}

	// Save and retrieve
	err := repo.Save("complex", complexSchema)
	if err != nil {
		t.Fatalf("Failed to save complex schema: %v", err)
	}

	retrieved, err := repo.Get("complex")
	if err != nil {
		t.Fatalf("Failed to get complex schema: %v", err)
	}

	// Verify structure is preserved
	if retrieved.If == nil || retrieved.Then == nil {
		t.Error("Conditional validation not preserved")
	}

	if len(retrieved.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(retrieved.Properties))
	}

	// Verify nested properties
	userProp, ok := retrieved.Properties["user"]
	if !ok {
		t.Error("User property not found")
	} else if len(userProp.Properties) != 2 {
		t.Errorf("Expected 2 user properties, got %d", len(userProp.Properties))
	}
}

func TestFileSchemaRepository_PermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Running as root, skipping permission test")
	}

	// Create a directory with no write permission
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	defer func() { _ = os.Chmod(readOnlyDir, 0755) }() // Cleanup

	_, err := NewFileSchemaRepository(filepath.Join(readOnlyDir, "schemas"))
	if err == nil {
		t.Error("Expected error when creating repository in read-only directory")
	}
}

func TestFileSchemaRepository_Metadata(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileSchemaRepository(tmpDir)

	// Save schema
	schema := &domain.Schema{Type: "object", Title: "Test"}
	if err := repo.Save("meta-test", schema); err != nil {
		t.Fatalf("Failed to save meta-test schema: %v", err)
	}

	// Check metadata file
	metadataPath := filepath.Join(tmpDir, "meta-test", "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("Metadata file was not created")
	}

	// Add more versions
	for i := 2; i <= 3; i++ {
		schema.Description = fmt.Sprintf("Version %d", i)
		if err := repo.Save("meta-test", schema); err != nil {
			t.Errorf("Failed to save version %d: %v", i, err)
		}
	}

	// Set different current version
	if err := repo.SetCurrentVersion("meta-test", 2); err != nil {
		t.Fatalf("Failed to set current version: %v", err)
	}

	// Read metadata
	data, _ := os.ReadFile(metadataPath)
	var metadata SchemaMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if metadata.LatestVersion != 3 {
		t.Errorf("Expected latest version 3, got %d", metadata.LatestVersion)
	}

	if metadata.CurrentVersion != 2 {
		t.Errorf("Expected current version 2, got %d", metadata.CurrentVersion)
	}
}
