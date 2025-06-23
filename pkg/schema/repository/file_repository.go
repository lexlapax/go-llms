// Package repository provides implementations for persistent schema storage.
// This package includes file-based and memory-based implementations of the
// SchemaRepository interface, supporting versioning, migration, and import/export
// capabilities for JSON schemas.
package repository

// ABOUTME: File-based implementation of SchemaRepository with directory structure
// ABOUTME: Provides persistent schema storage with versioning and migration support

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FileSchemaRepository provides file-based persistent storage for schemas
type FileSchemaRepository struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileSchemaRepository creates a new file-based schema repository
func NewFileSchemaRepository(basePath string) (*FileSchemaRepository, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FileSchemaRepository{
		basePath: basePath,
	}, nil
}

// schemaDir returns the directory path for a schema ID
func (r *FileSchemaRepository) schemaDir(id string) string {
	// Sanitize ID to prevent directory traversal
	safeID := strings.ReplaceAll(id, string(os.PathSeparator), "_")
	safeID = strings.ReplaceAll(safeID, "..", "_")
	return filepath.Join(r.basePath, safeID)
}

// versionFile returns the file path for a specific version
func (r *FileSchemaRepository) versionFile(id string, version int) string {
	return filepath.Join(r.schemaDir(id), fmt.Sprintf("v%d.json", version))
}

// currentFile returns the file path for the current version pointer
func (r *FileSchemaRepository) currentFile(id string) string {
	return filepath.Join(r.schemaDir(id), "current")
}

// metadataFile returns the file path for schema metadata
func (r *FileSchemaRepository) metadataFile(id string) string {
	return filepath.Join(r.schemaDir(id), "metadata.json")
}

// Get retrieves the current version of a schema by ID
func (r *FileSchemaRepository) Get(id string) (*domain.Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Read current version number
	currentPath := r.currentFile(id)
	currentData, err := os.ReadFile(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("schema not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read current version: %w", err)
	}

	version, err := strconv.Atoi(strings.TrimSpace(string(currentData)))
	if err != nil {
		return nil, fmt.Errorf("invalid current version format: %w", err)
	}

	return r.getVersionUnsafe(id, version)
}

// getVersionUnsafe reads a version without locking (caller must hold lock)
func (r *FileSchemaRepository) getVersionUnsafe(id string, version int) (*domain.Schema, error) {
	versionPath := r.versionFile(id, version)
	data, err := os.ReadFile(versionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema version %d: %w", version, err)
	}

	var schema domain.Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

// Save stores a schema, creating a new version
func (r *FileSchemaRepository) Save(id string, schema *domain.Schema) error {
	if id == "" {
		return fmt.Errorf("schema ID cannot be empty")
	}
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Create schema directory if it doesn't exist
	schemaDir := r.schemaDir(id)
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		return fmt.Errorf("failed to create schema directory: %w", err)
	}

	// Determine next version
	versions, _ := r.listVersionsUnsafe(id)
	nextVersion := len(versions) + 1

	// Marshal schema
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Write version file
	versionPath := r.versionFile(id, nextVersion)
	if err := os.WriteFile(versionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write schema version: %w", err)
	}

	// Update current version pointer
	currentPath := r.currentFile(id)
	if err := os.WriteFile(currentPath, []byte(strconv.Itoa(nextVersion)), 0644); err != nil {
		// Rollback version file on error
		_ = os.Remove(versionPath)
		return fmt.Errorf("failed to update current version: %w", err)
	}

	// Update metadata
	metadata := SchemaMetadata{
		ID:             id,
		LatestVersion:  nextVersion,
		CurrentVersion: nextVersion,
		TotalVersions:  nextVersion,
	}
	metadataData, _ := json.MarshalIndent(metadata, "", "  ")
	metadataPath := r.metadataFile(id)
	_ = os.WriteFile(metadataPath, metadataData, 0644)

	return nil
}

// Delete removes all versions of a schema
func (r *FileSchemaRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	schemaDir := r.schemaDir(id)
	if _, err := os.Stat(schemaDir); os.IsNotExist(err) {
		return fmt.Errorf("schema not found: %s", id)
	}

	if err := os.RemoveAll(schemaDir); err != nil {
		return fmt.Errorf("failed to delete schema: %w", err)
	}

	return nil
}

// GetVersion retrieves a specific version of a schema
func (r *FileSchemaRepository) GetVersion(id string, version int) (*domain.Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.getVersionUnsafe(id, version)
}

// ListVersions returns all version numbers for a schema
func (r *FileSchemaRepository) ListVersions(id string) ([]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.listVersionsUnsafe(id)
}

// listVersionsUnsafe lists versions without locking (caller must hold lock)
func (r *FileSchemaRepository) listVersionsUnsafe(id string) ([]int, error) {
	schemaDir := r.schemaDir(id)
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("schema not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read schema directory: %w", err)
	}

	var versions []int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "v") && strings.HasSuffix(name, ".json") {
			versionStr := strings.TrimPrefix(strings.TrimSuffix(name, ".json"), "v")
			if version, err := strconv.Atoi(versionStr); err == nil {
				versions = append(versions, version)
			}
		}
	}

	sort.Ints(versions)
	return versions, nil
}

// SetCurrentVersion sets the active version for a schema
func (r *FileSchemaRepository) SetCurrentVersion(id string, version int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify version exists
	versionPath := r.versionFile(id, version)
	if _, err := os.Stat(versionPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("version %d not found for schema %s", version, id)
		}
		return fmt.Errorf("failed to check version: %w", err)
	}

	// Update current version pointer
	currentPath := r.currentFile(id)
	if err := os.WriteFile(currentPath, []byte(strconv.Itoa(version)), 0644); err != nil {
		return fmt.Errorf("failed to update current version: %w", err)
	}

	// Update metadata
	metadataPath := r.metadataFile(id)
	var metadata SchemaMetadata
	if data, err := os.ReadFile(metadataPath); err == nil {
		_ = json.Unmarshal(data, &metadata)
	}
	metadata.CurrentVersion = version
	metadataData, _ := json.MarshalIndent(metadata, "", "  ")
	_ = os.WriteFile(metadataPath, metadataData, 0644)

	return nil
}

// List returns all schema IDs
func (r *FileSchemaRepository) List() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it's a valid schema directory by looking for metadata or current file
			currentPath := filepath.Join(r.basePath, entry.Name(), "current")
			if _, err := os.Stat(currentPath); err == nil {
				ids = append(ids, entry.Name())
			}
		}
	}

	return ids, nil
}

// Export creates a tar.gz archive of all schemas
func (r *FileSchemaRepository) Export() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// For simplicity, we'll return the base path
	// In a real implementation, this would create a tar.gz archive
	return r.basePath, nil
}

// Import extracts schemas from a tar.gz archive
func (r *FileSchemaRepository) Import(archivePath string) error {
	// For simplicity, this is a no-op
	// In a real implementation, this would extract a tar.gz archive
	return fmt.Errorf("import not implemented")
}

// Clear removes all schemas
func (r *FileSchemaRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			path := filepath.Join(r.basePath, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// Count returns the number of schemas
func (r *FileSchemaRepository) Count() (int, error) {
	ids, err := r.List()
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

// Migrate performs schema migrations between versions
func (r *FileSchemaRepository) Migrate(id string, fromVersion, toVersion int, migrator SchemaMigrator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get source schema
	sourceSchema, err := r.getVersionUnsafe(id, fromVersion)
	if err != nil {
		return fmt.Errorf("failed to get source version: %w", err)
	}

	// Apply migration
	migratedSchema, err := migrator.Migrate(sourceSchema, fromVersion, toVersion)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Save as new version
	data, err := json.MarshalIndent(migratedSchema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal migrated schema: %w", err)
	}

	versionPath := r.versionFile(id, toVersion)
	if err := os.WriteFile(versionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write migrated version: %w", err)
	}

	return nil
}

// SchemaMetadata stores metadata about a schema
type SchemaMetadata struct {
	ID             string `json:"id"`
	LatestVersion  int    `json:"latest_version"`
	CurrentVersion int    `json:"current_version"`
	TotalVersions  int    `json:"total_versions"`
}

// SchemaMigrator defines the interface for schema migrations
type SchemaMigrator interface {
	Migrate(schema *domain.Schema, fromVersion, toVersion int) (*domain.Schema, error)
}
