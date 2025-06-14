// Package repository provides implementations of the SchemaRepository interface
package repository

// ABOUTME: In-memory implementation of SchemaRepository with thread-safe operations
// ABOUTME: Provides schema storage with versioning and export/import capabilities

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// SchemaVersion represents a versioned schema entry
type SchemaVersion struct {
	Schema    *domain.Schema `json:"schema"`
	Version   int            `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// InMemorySchemaRepository provides thread-safe in-memory storage for schemas
type InMemorySchemaRepository struct {
	schemas map[string][]SchemaVersion // ID -> versions (latest is last)
	current map[string]int             // ID -> current version index
	mu      sync.RWMutex
}

// NewInMemorySchemaRepository creates a new in-memory schema repository
func NewInMemorySchemaRepository() *InMemorySchemaRepository {
	return &InMemorySchemaRepository{
		schemas: make(map[string][]SchemaVersion),
		current: make(map[string]int),
	}
}

// Get retrieves the current version of a schema by ID
func (r *InMemorySchemaRepository) Get(id string) (*domain.Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.schemas[id]
	if !exists || len(versions) == 0 {
		return nil, fmt.Errorf("schema not found: %s", id)
	}

	currentIdx := r.current[id]
	if currentIdx < 0 || currentIdx >= len(versions) {
		currentIdx = len(versions) - 1 // Use latest version
	}

	return versions[currentIdx].Schema, nil
}

// Save stores a schema, creating a new version
func (r *InMemorySchemaRepository) Save(id string, schema *domain.Schema) error {
	if id == "" {
		return fmt.Errorf("schema ID cannot be empty")
	}
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	versions := r.schemas[id]

	newVersion := SchemaVersion{
		Schema:    schema,
		Version:   len(versions) + 1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.schemas[id] = append(versions, newVersion)
	r.current[id] = len(r.schemas[id]) - 1

	return nil
}

// Delete removes all versions of a schema
func (r *InMemorySchemaRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.schemas[id]; !exists {
		return fmt.Errorf("schema not found: %s", id)
	}

	delete(r.schemas, id)
	delete(r.current, id)

	return nil
}

// GetVersion retrieves a specific version of a schema
func (r *InMemorySchemaRepository) GetVersion(id string, version int) (*domain.Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.schemas[id]
	if !exists || len(versions) == 0 {
		return nil, fmt.Errorf("schema not found: %s", id)
	}

	if version < 1 || version > len(versions) {
		return nil, fmt.Errorf("invalid version %d for schema %s (available: 1-%d)", version, id, len(versions))
	}

	return versions[version-1].Schema, nil
}

// ListVersions returns all versions of a schema
func (r *InMemorySchemaRepository) ListVersions(id string) ([]SchemaVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.schemas[id]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", id)
	}

	// Return a copy to prevent external modification
	result := make([]SchemaVersion, len(versions))
	copy(result, versions)
	return result, nil
}

// SetCurrentVersion sets the active version for a schema
func (r *InMemorySchemaRepository) SetCurrentVersion(id string, version int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions, exists := r.schemas[id]
	if !exists || len(versions) == 0 {
		return fmt.Errorf("schema not found: %s", id)
	}

	if version < 1 || version > len(versions) {
		return fmt.Errorf("invalid version %d for schema %s (available: 1-%d)", version, id, len(versions))
	}

	r.current[id] = version - 1
	return nil
}

// List returns all schema IDs
func (r *InMemorySchemaRepository) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.schemas))
	for id := range r.schemas {
		ids = append(ids, id)
	}
	return ids
}

// Export serializes all schemas to JSON
func (r *InMemorySchemaRepository) Export() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := struct {
		Schemas map[string][]SchemaVersion `json:"schemas"`
		Current map[string]int             `json:"current"`
	}{
		Schemas: r.schemas,
		Current: r.current,
	}

	return json.MarshalIndent(data, "", "  ")
}

// Import deserializes schemas from JSON
func (r *InMemorySchemaRepository) Import(data []byte) error {
	var imported struct {
		Schemas map[string][]SchemaVersion `json:"schemas"`
		Current map[string]int             `json:"current"`
	}

	if err := json.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.schemas = imported.Schemas
	r.current = imported.Current

	return nil
}

// Clear removes all schemas
func (r *InMemorySchemaRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schemas = make(map[string][]SchemaVersion)
	r.current = make(map[string]int)
}

// Count returns the number of schemas
func (r *InMemorySchemaRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.schemas)
}
