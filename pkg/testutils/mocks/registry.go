// ABOUTME: Mock registry for centralized management of all mock instances
// ABOUTME: Provides global registration, lookup, and reset capabilities for testing

package mocks

import (
	"fmt"
	"sync"
)

// MockType represents the type of mock
type MockType string

const (
	TypeProvider MockType = "provider"
	TypeTool     MockType = "tool"
	TypeAgent    MockType = "agent"
	TypeState    MockType = "state"
	TypeEvent    MockType = "event"
)

// Mock represents a generic mock interface
type Mock interface {
	Reset()
	Name() string
}

// Registry manages all mock instances
type Registry struct {
	mu    sync.RWMutex
	mocks map[MockType]map[string]Mock
}

// globalRegistry is the default registry instance
var globalRegistry = NewRegistry()

// NewRegistry creates a new mock registry
func NewRegistry() *Registry {
	return &Registry{
		mocks: make(map[MockType]map[string]Mock),
	}
}

// Register adds a mock to the registry
func (r *Registry) Register(mockType MockType, name string, mock Mock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mocks[mockType] == nil {
		r.mocks[mockType] = make(map[string]Mock)
	}

	if _, exists := r.mocks[mockType][name]; exists {
		return fmt.Errorf("mock %s of type %s already registered", name, mockType)
	}

	r.mocks[mockType][name] = mock
	return nil
}

// Unregister removes a mock from the registry
func (r *Registry) Unregister(mockType MockType, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mocks[mockType] == nil {
		return fmt.Errorf("no mocks of type %s registered", mockType)
	}

	if _, exists := r.mocks[mockType][name]; !exists {
		return fmt.Errorf("mock %s of type %s not found", name, mockType)
	}

	delete(r.mocks[mockType], name)
	return nil
}

// Get retrieves a mock from the registry
func (r *Registry) Get(mockType MockType, name string) (Mock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.mocks[mockType] == nil {
		return nil, fmt.Errorf("no mocks of type %s registered", mockType)
	}

	mock, exists := r.mocks[mockType][name]
	if !exists {
		return nil, fmt.Errorf("mock %s of type %s not found", name, mockType)
	}

	return mock, nil
}

// GetProvider retrieves a mock provider from the registry
func (r *Registry) GetProvider(name string) (*MockProvider, error) {
	mock, err := r.Get(TypeProvider, name)
	if err != nil {
		return nil, err
	}

	provider, ok := mock.(*MockProvider)
	if !ok {
		return nil, fmt.Errorf("mock %s is not a MockProvider", name)
	}

	return provider, nil
}

// GetTool retrieves a mock tool from the registry
func (r *Registry) GetTool(name string) (*MockTool, error) {
	mock, err := r.Get(TypeTool, name)
	if err != nil {
		return nil, err
	}

	tool, ok := mock.(*MockTool)
	if !ok {
		return nil, fmt.Errorf("mock %s is not a MockTool", name)
	}

	return tool, nil
}

// List returns all mocks of a specific type
func (r *Registry) List(mockType MockType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.mocks[mockType] == nil {
		return []string{}
	}

	names := make([]string, 0, len(r.mocks[mockType]))
	for name := range r.mocks[mockType] {
		names = append(names, name)
	}

	return names
}

// ListAll returns all registered mocks grouped by type
func (r *Registry) ListAll() map[MockType][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[MockType][]string)
	for mockType, mocks := range r.mocks {
		names := make([]string, 0, len(mocks))
		for name := range mocks {
			names = append(names, name)
		}
		result[mockType] = names
	}

	return result
}

// Reset resets all mocks in the registry
func (r *Registry) Reset() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, mocks := range r.mocks {
		for _, mock := range mocks {
			mock.Reset()
		}
	}
}

// ResetType resets all mocks of a specific type
func (r *Registry) ResetType(mockType MockType) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.mocks[mockType] == nil {
		return
	}

	for _, mock := range r.mocks[mockType] {
		mock.Reset()
	}
}

// Clear removes all mocks from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mocks = make(map[MockType]map[string]Mock)
}

// ClearType removes all mocks of a specific type
func (r *Registry) ClearType(mockType MockType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.mocks, mockType)
}

// Global registry functions

// Register adds a mock to the global registry
func Register(mockType MockType, name string, mock Mock) error {
	return globalRegistry.Register(mockType, name, mock)
}

// Unregister removes a mock from the global registry
func Unregister(mockType MockType, name string) error {
	return globalRegistry.Unregister(mockType, name)
}

// Get retrieves a mock from the global registry
func Get(mockType MockType, name string) (Mock, error) {
	return globalRegistry.Get(mockType, name)
}

// GetProvider retrieves a mock provider from the global registry
func GetProvider(name string) (*MockProvider, error) {
	return globalRegistry.GetProvider(name)
}

// GetTool retrieves a mock tool from the global registry
func GetTool(name string) (*MockTool, error) {
	return globalRegistry.GetTool(name)
}

// List returns all mocks of a specific type from the global registry
func List(mockType MockType) []string {
	return globalRegistry.List(mockType)
}

// ListAll returns all registered mocks from the global registry
func ListAll() map[MockType][]string {
	return globalRegistry.ListAll()
}

// Reset resets all mocks in the global registry
func Reset() {
	globalRegistry.Reset()
}

// ResetType resets all mocks of a specific type in the global registry
func ResetType(mockType MockType) {
	globalRegistry.ResetType(mockType)
}

// Clear removes all mocks from the global registry
func Clear() {
	globalRegistry.Clear()
}

// ClearType removes all mocks of a specific type from the global registry
func ClearType(mockType MockType) {
	globalRegistry.ClearType(mockType)
}

// MockTool Name method is already implemented

// Helper functions for common mock registration patterns

// RegisterProvider is a convenience function to register a mock provider
func RegisterProvider(name string, provider *MockProvider) error {
	return Register(TypeProvider, name, provider)
}

// RegisterTool is a convenience function to register a mock tool
func RegisterTool(name string, tool *MockTool) error {
	return Register(TypeTool, name, tool)
}

// WithRegistry creates a new isolated registry for testing
func WithRegistry(fn func(r *Registry)) {
	registry := NewRegistry()
	fn(registry)
}
