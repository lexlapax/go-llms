// ABOUTME: Generic registry interface and implementation for built-in components
// ABOUTME: Provides thread-safe registration, discovery, and metadata management

package builtins

import (
	"fmt"
	"strings"
	"sync"
)

// Registry provides a central registration and discovery mechanism for components
type Registry[T any] interface {
	// Register adds a component to the registry
	Register(name string, component T, metadata Metadata) error

	// Get retrieves a component by name
	Get(name string) (T, bool)

	// MustGet retrieves a component by name or panics if not found
	MustGet(name string) T

	// List returns all registered components
	List() []RegistryEntry[T]

	// ListByCategory returns components in a specific category
	ListByCategory(category string) []RegistryEntry[T]

	// ListByTags returns components matching all provided tags
	ListByTags(tags ...string) []RegistryEntry[T]

	// Search returns components matching the query (searches name, description, tags)
	Search(query string) []RegistryEntry[T]

	// Categories returns all unique categories
	Categories() []string

	// Clear removes all entries (useful for testing)
	Clear()
}

// Metadata describes a registered component
type Metadata struct {
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Tags         []string  `json:"tags"`
	Description  string    `json:"description"`
	Version      string    `json:"version"`
	Examples     []Example `json:"examples,omitempty"`
	Deprecated   bool      `json:"deprecated,omitempty"`
	Experimental bool      `json:"experimental,omitempty"`
}

// Example shows how to use a component
type Example struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
}

// RegistryEntry combines a component with its metadata
type RegistryEntry[T any] struct {
	Component T
	Metadata  Metadata
}

// baseRegistry provides a thread-safe implementation of Registry
type baseRegistry[T any] struct {
	mu      sync.RWMutex
	entries map[string]RegistryEntry[T]
}

// NewRegistry creates a new registry instance
func NewRegistry[T any]() Registry[T] {
	return &baseRegistry[T]{
		entries: make(map[string]RegistryEntry[T]),
	}
}

// Register adds a component to the registry
func (r *baseRegistry[T]) Register(name string, component T, metadata Metadata) error {
	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	// Ensure metadata name matches registration name
	if metadata.Name != "" && metadata.Name != name {
		return fmt.Errorf("metadata name '%s' does not match registration name '%s'", metadata.Name, name)
	}
	metadata.Name = name

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entries[name]; exists {
		return fmt.Errorf("component '%s' is already registered", name)
	}

	r.entries[name] = RegistryEntry[T]{
		Component: component,
		Metadata:  metadata,
	}

	return nil
}

// Get retrieves a component by name
func (r *baseRegistry[T]) Get(name string) (T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.entries[name]
	return entry.Component, exists
}

// MustGet retrieves a component by name or panics if not found
func (r *baseRegistry[T]) MustGet(name string) T {
	component, exists := r.Get(name)
	if !exists {
		panic(fmt.Sprintf("component '%s' not found in registry", name))
	}
	return component
}

// List returns all registered components
func (r *baseRegistry[T]) List() []RegistryEntry[T] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]RegistryEntry[T], 0, len(r.entries))
	for _, entry := range r.entries {
		entries = append(entries, entry)
	}
	return entries
}

// ListByCategory returns components in a specific category
func (r *baseRegistry[T]) ListByCategory(category string) []RegistryEntry[T] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var entries []RegistryEntry[T]
	for _, entry := range r.entries {
		if strings.EqualFold(entry.Metadata.Category, category) {
			entries = append(entries, entry)
		}
	}
	return entries
}

// ListByTags returns components matching all provided tags
func (r *baseRegistry[T]) ListByTags(tags ...string) []RegistryEntry[T] {
	if len(tags) == 0 {
		return r.List()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var entries []RegistryEntry[T]
	for _, entry := range r.entries {
		if containsAllTags(entry.Metadata.Tags, tags) {
			entries = append(entries, entry)
		}
	}
	return entries
}

// Search returns components matching the query
func (r *baseRegistry[T]) Search(query string) []RegistryEntry[T] {
	if query == "" {
		return r.List()
	}

	query = strings.ToLower(query)
	r.mu.RLock()
	defer r.mu.RUnlock()

	var entries []RegistryEntry[T]
	for _, entry := range r.entries {
		if matchesSearch(entry.Metadata, query) {
			entries = append(entries, entry)
		}
	}
	return entries
}

// Categories returns all unique categories
func (r *baseRegistry[T]) Categories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categoryMap := make(map[string]bool)
	for _, entry := range r.entries {
		if entry.Metadata.Category != "" {
			categoryMap[entry.Metadata.Category] = true
		}
	}

	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}
	return categories
}

// Clear removes all entries
func (r *baseRegistry[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries = make(map[string]RegistryEntry[T])
}

// containsAllTags checks if tags contains all searchTags
func containsAllTags(tags, searchTags []string) bool {
	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[strings.ToLower(tag)] = true
	}

	for _, searchTag := range searchTags {
		if !tagMap[strings.ToLower(searchTag)] {
			return false
		}
	}
	return true
}

// matchesSearch checks if metadata matches the search query
func matchesSearch(metadata Metadata, query string) bool {
	// Check name
	if strings.Contains(strings.ToLower(metadata.Name), query) {
		return true
	}

	// Check description
	if strings.Contains(strings.ToLower(metadata.Description), query) {
		return true
	}

	// Check category
	if strings.Contains(strings.ToLower(metadata.Category), query) {
		return true
	}

	// Check tags
	for _, tag := range metadata.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	return false
}
