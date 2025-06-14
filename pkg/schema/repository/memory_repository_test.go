package repository

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestInMemorySchemaRepository_CRUD(t *testing.T) {
	repo := NewInMemorySchemaRepository()

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
	err := repo.Save("test-id", schema)
	if err != nil {
		t.Fatalf("Failed to save schema: %v", err)
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

	_, err = repo.Get("test-id")
	if err == nil {
		t.Error("Expected error when getting deleted schema")
	}
}

func TestInMemorySchemaRepository_Versioning(t *testing.T) {
	repo := NewInMemorySchemaRepository()

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

func TestInMemorySchemaRepository_ExportImport(t *testing.T) {
	repo1 := NewInMemorySchemaRepository()

	// Add some schemas
	schemas := map[string]*domain.Schema{
		"schema1": {Type: "object", Title: "Schema 1"},
		"schema2": {Type: "array", Title: "Schema 2"},
		"schema3": {Type: "string", Title: "Schema 3"},
	}

	for id, schema := range schemas {
		err := repo1.Save(id, schema)
		if err != nil {
			t.Fatalf("Failed to save schema %s: %v", id, err)
		}
	}

	// Export
	exported, err := repo1.Export()
	if err != nil {
		t.Fatalf("Failed to export: %v", err)
	}

	// Import into new repository
	repo2 := NewInMemorySchemaRepository()
	err = repo2.Import(exported)
	if err != nil {
		t.Fatalf("Failed to import: %v", err)
	}

	// Verify imported data
	if repo2.Count() != repo1.Count() {
		t.Errorf("Expected %d schemas, got %d", repo1.Count(), repo2.Count())
	}

	for id, originalSchema := range schemas {
		imported, err := repo2.Get(id)
		if err != nil {
			t.Errorf("Failed to get imported schema %s: %v", id, err)
			continue
		}
		if imported.Title != originalSchema.Title {
			t.Errorf("Schema %s: expected title %s, got %s", id, originalSchema.Title, imported.Title)
		}
	}
}

func TestInMemorySchemaRepository_Concurrency(t *testing.T) {
	repo := NewInMemorySchemaRepository()
	var wg sync.WaitGroup
	numGoroutines := 100

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
	if repo.Count() != numGoroutines {
		t.Errorf("Expected %d schemas, got %d", numGoroutines, repo.Count())
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

func TestInMemorySchemaRepository_EdgeCases(t *testing.T) {
	repo := NewInMemorySchemaRepository()

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
}

func TestInMemorySchemaRepository_List(t *testing.T) {
	repo := NewInMemorySchemaRepository()

	// Empty list
	ids := repo.List()
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
	ids = repo.List()
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

func TestInMemorySchemaRepository_Clear(t *testing.T) {
	repo := NewInMemorySchemaRepository()

	// Add schemas
	for i := 0; i < 5; i++ {
		if err := repo.Save(fmt.Sprintf("schema-%d", i), &domain.Schema{Type: "object"}); err != nil {
			t.Errorf("Failed to save schema-%d: %v", i, err)
		}
	}

	if repo.Count() != 5 {
		t.Errorf("Expected 5 schemas, got %d", repo.Count())
	}

	// Clear
	repo.Clear()

	if repo.Count() != 0 {
		t.Errorf("Expected 0 schemas after clear, got %d", repo.Count())
	}

	// Verify schemas are gone
	_, err := repo.Get("schema-0")
	if err == nil {
		t.Error("Expected error when getting schema after clear")
	}
}

func TestSchemaVersion_Timestamps(t *testing.T) {
	repo := NewInMemorySchemaRepository()

	// Save schema
	before := time.Now()
	if err := repo.Save("timed", &domain.Schema{Type: "object"}); err != nil {
		t.Fatalf("Failed to save timed schema: %v", err)
	}
	after := time.Now()

	versions, _ := repo.ListVersions("timed")
	if len(versions) != 1 {
		t.Fatalf("Expected 1 version, got %d", len(versions))
	}

	v := versions[0]
	if v.CreatedAt.Before(before) || v.CreatedAt.After(after) {
		t.Error("CreatedAt timestamp out of expected range")
	}

	if v.UpdatedAt != v.CreatedAt {
		t.Error("UpdatedAt should equal CreatedAt for new version")
	}

	if v.Version != 1 {
		t.Errorf("Expected version 1, got %d", v.Version)
	}
}

func TestInMemorySchemaRepository_ComplexSchema(t *testing.T) {
	repo := NewInMemorySchemaRepository()

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

	// Export and re-import to verify serialization
	exported, _ := repo.Export()
	repo2 := NewInMemorySchemaRepository()
	if err := repo2.Import(exported); err != nil {
		t.Fatalf("Failed to import exported data: %v", err)
	}

	reimported, _ := repo2.Get("complex")
	if reimported.Title != complexSchema.Title {
		t.Error("Complex schema not properly serialized/deserialized")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
