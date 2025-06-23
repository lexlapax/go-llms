// ABOUTME: Tool discovery system providing metadata-first access without imports
// ABOUTME: Enables dynamic tool exploration for scripting engines and CLI tools

package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ToolInfo represents lightweight tool metadata for discovery.
// It provides essential information about a tool without requiring
// the tool to be loaded or instantiated.
type ToolInfo struct {
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Category        string          `json:"category"`
	Tags            []string        `json:"tags"`
	Version         string          `json:"version"`
	ParameterSchema json.RawMessage `json:"parameter_schema,omitempty"`
	OutputSchema    json.RawMessage `json:"output_schema,omitempty"`
	Examples        []Example       `json:"examples,omitempty"`
	UsageHint       string          `json:"usage_hint,omitempty"`
	Package         string          `json:"package,omitempty"` // For lazy loading
}

// Example represents a simplified example structure for tool usage.
// It includes both input and expected output for demonstration purposes.
type Example struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Input       json.RawMessage `json:"input"`
	Output      json.RawMessage `json:"output,omitempty"`
}

// ToolFactory is a function that creates a tool on demand.
// This enables lazy loading and reduces memory usage for unused tools.
type ToolFactory func() (domain.Tool, error)

// ToolDiscovery provides metadata-first tool discovery with dynamic registration.
// It enables tools to be discovered, explored, and instantiated without
// requiring all tools to be loaded into memory. This interface supports
// dynamic registration, versioning, and multi-tenant isolation.
type ToolDiscovery interface {
	// Static discovery methods
	ListTools() []ToolInfo
	SearchTools(query string) []ToolInfo
	ListByCategory(category string) []ToolInfo
	GetToolSchema(name string) (*ToolSchema, error)
	GetToolExamples(name string) ([]domain.ToolExample, error)
	CreateTool(name string) (domain.Tool, error)
	CreateTools(names ...string) (map[string]domain.Tool, error)
	GetToolHelp(name string) (string, error)

	// Dynamic registration methods (REQUIRED FOR DOWNSTREAM)
	RegisterTool(info ToolInfo, factory ToolFactory) error
	UnregisterTool(name string) error
	GetRegisteredTools() []ToolInfo

	// Registry persistence (DOWNSTREAM REQUIREMENT)
	SaveRegistry(writer interface{}) error // io.Writer for serialization
	LoadRegistry(reader interface{}) error // io.Reader for deserialization

	// Tool versioning support
	RegisterToolVersion(info ToolInfo, factory ToolFactory, version string) error
	GetToolVersions(name string) []string
	CreateToolVersion(name, version string) (domain.Tool, error)

	// Multi-tenant support (for downstream isolation)
	CreateNamespace(namespace string) error
	ListNamespaces() []string
	SwitchNamespace(namespace string) error
	GetCurrentNamespace() string
}

// ToolSchema contains detailed schema information for a tool.
// It provides comprehensive metadata including parameters, outputs,
// examples, and usage guidance.
type ToolSchema struct {
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Parameters    interface{}          `json:"parameters,omitempty"`
	Output        interface{}          `json:"output,omitempty"`
	Examples      []domain.ToolExample `json:"examples,omitempty"`
	Constraints   []string             `json:"constraints,omitempty"`
	ErrorGuidance map[string]string    `json:"error_guidance,omitempty"`
}

// ToolVersionInfo holds versioned tool information.
// It tracks multiple versions of a tool and their corresponding factories.
type ToolVersionInfo struct {
	ToolInfo
	Versions map[string]ToolFactory `json:"-"` // version -> factory mapping
}

// NamespaceRegistry holds tools for a specific namespace.
// This enables multi-tenant isolation and separate tool environments.
type NamespaceRegistry struct {
	Name      string                     `json:"name"`
	Metadata  map[string]ToolVersionInfo `json:"metadata"`
	Factories map[string]ToolFactory     `json:"-"` // Current version factories
}

// toolDiscovery implements ToolDiscovery with enhanced features.
// It provides thread-safe operations, multi-namespace support,
// and dynamic tool registration capabilities.
type toolDiscovery struct {
	// Multi-namespace support
	namespaces       map[string]*NamespaceRegistry
	currentNamespace string

	// Thread safety
	mu sync.RWMutex
}

// globalDiscovery is the singleton discovery instance
var (
	globalDiscovery *toolDiscovery
	discoveryOnce   sync.Once
)

// NewDiscovery returns the global discovery instance.
// It uses a singleton pattern to ensure consistent tool metadata
// across the application.
//
// Returns the global ToolDiscovery instance.
func NewDiscovery() ToolDiscovery {
	discoveryOnce.Do(func() {
		globalDiscovery = &toolDiscovery{
			namespaces:       make(map[string]*NamespaceRegistry),
			currentNamespace: "default",
		}
		// Create default namespace
		_ = globalDiscovery.CreateNamespace("default")
		// Initialize with metadata from registry
		globalDiscovery.initializeFromRegistry()
	})
	return globalDiscovery
}

// initializeFromRegistry populates metadata from existing registry.
// This is called during initialization to load built-in tools.
func (d *toolDiscovery) initializeFromRegistry() {
	// Metadata will be populated by the generated registry_metadata.go init() function
	// which calls RegisterToolMetadata for each tool
	// No need to do anything here as the metadata is registered at package init time
}

// RegisterToolMetadata registers tool metadata without the tool instance.
// This is typically called from init() functions to register built-in tools.
//
// Parameters:
//   - info: Tool metadata information
//   - factory: Factory function to create the tool
//
// Returns an error if registration fails.
func RegisterToolMetadata(info ToolInfo, factory ToolFactory) error {
	discovery := NewDiscovery().(*toolDiscovery)
	return discovery.RegisterTool(info, factory)
}

// getCurrentRegistry returns the current namespace registry, creating it if needed.
// This ensures a registry always exists for the current namespace.
func (d *toolDiscovery) getCurrentRegistry() *NamespaceRegistry {
	registry, exists := d.namespaces[d.currentNamespace]
	if !exists {
		// This should not happen as we create default namespace in NewDiscovery
		_ = d.CreateNamespace(d.currentNamespace)
		registry = d.namespaces[d.currentNamespace]
	}
	return registry
}

// ListTools returns all available tools without loading them.
// This provides efficient tool discovery without memory overhead.
//
// Returns a slice of ToolInfo for all tools in the current namespace.
func (d *toolDiscovery) ListTools() []ToolInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	registry := d.getCurrentRegistry()
	result := make([]ToolInfo, 0, len(registry.Metadata))
	for _, versionInfo := range registry.Metadata {
		result = append(result, versionInfo.ToolInfo)
	}
	return result
}

// SearchTools searches tools by keyword.
// It searches in tool names, descriptions, and tags.
//
// Parameters:
//   - query: The search query (case-insensitive)
//
// Returns matching tools from the current namespace.
func (d *toolDiscovery) SearchTools(query string) []ToolInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	registry := d.getCurrentRegistry()
	query = strings.ToLower(query)
	var results []ToolInfo

	for _, versionInfo := range registry.Metadata {
		info := versionInfo.ToolInfo
		// Search in name
		if strings.Contains(strings.ToLower(info.Name), query) {
			results = append(results, info)
			continue
		}

		// Search in description
		if strings.Contains(strings.ToLower(info.Description), query) {
			results = append(results, info)
			continue
		}

		// Search in tags
		for _, tag := range info.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, info)
				break
			}
		}
	}

	return results
}

// ListByCategory returns tools in a specific category.
// Categories help organize tools by functionality.
//
// Parameters:
//   - category: The category to filter by
//
// Returns tools matching the category in the current namespace.
func (d *toolDiscovery) ListByCategory(category string) []ToolInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	registry := d.getCurrentRegistry()
	var results []ToolInfo
	for _, versionInfo := range registry.Metadata {
		if versionInfo.Category == category {
			results = append(results, versionInfo.ToolInfo)
		}
	}
	return results
}

// GetToolSchema returns detailed schema for a specific tool.
// This includes parameters, outputs, examples, and constraints.
//
// Parameters:
//   - name: The tool name
//
// Returns the tool schema or an error if not found.
func (d *toolDiscovery) GetToolSchema(name string) (*ToolSchema, error) {
	d.mu.RLock()
	registry := d.getCurrentRegistry()
	versionInfo, exists := registry.Metadata[name]
	d.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	info := versionInfo.ToolInfo

	schema := &ToolSchema{
		Name:        info.Name,
		Description: info.Description,
	}

	// Parse parameter schema
	if len(info.ParameterSchema) > 0 {
		var params interface{}
		if err := json.Unmarshal(info.ParameterSchema, &params); err == nil {
			schema.Parameters = params
		}
	}

	// Parse output schema
	if len(info.OutputSchema) > 0 {
		var output interface{}
		if err := json.Unmarshal(info.OutputSchema, &output); err == nil {
			schema.Output = output
		}
	}

	// Convert examples back to domain.ToolExample
	for _, ex := range info.Examples {
		var input, output interface{}
		_ = json.Unmarshal(ex.Input, &input)
		_ = json.Unmarshal(ex.Output, &output)

		schema.Examples = append(schema.Examples, domain.ToolExample{
			Name:        ex.Name,
			Description: ex.Description,
			Input:       input,
			Output:      output,
		})
	}

	// Try to get more details from the actual tool if it's loaded
	if tool, found := tools.GetTool(name); found {
		schema.Constraints = tool.Constraints()
		schema.ErrorGuidance = tool.ErrorGuidance()
	}

	return schema, nil
}

// GetToolExamples returns examples for a specific tool.
// Examples help users understand proper tool usage.
//
// Parameters:
//   - name: The tool name
//
// Returns tool examples or an error if not found.
func (d *toolDiscovery) GetToolExamples(name string) ([]domain.ToolExample, error) {
	schema, err := d.GetToolSchema(name)
	if err != nil {
		return nil, err
	}
	return schema.Examples, nil
}

// CreateTool instantiates a tool by name.
// This uses the registered factory to create the tool on demand.
//
// Parameters:
//   - name: The tool name
//
// Returns the created tool or an error if not found.
func (d *toolDiscovery) CreateTool(name string) (domain.Tool, error) {
	d.mu.RLock()
	registry := d.getCurrentRegistry()
	factory, exists := registry.Factories[name]
	d.mu.RUnlock()

	if !exists {
		// Try to get from registry if already loaded
		if tool, found := tools.GetTool(name); found {
			return tool, nil
		}
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return factory()
}

// CreateTools instantiates multiple tools.
// This is a convenience method for batch tool creation.
//
// Parameters:
//   - names: Tool names to create
//
// Returns a map of created tools or an error if any creation fails.
func (d *toolDiscovery) CreateTools(names ...string) (map[string]domain.Tool, error) {
	result := make(map[string]domain.Tool)

	for _, name := range names {
		tool, err := d.CreateTool(name)
		if err != nil {
			return nil, fmt.Errorf("failed to create tool %s: %w", name, err)
		}
		result[name] = tool
	}

	return result, nil
}

// GetToolHelp generates help text for a tool.
// This provides comprehensive usage information in a human-readable format.
//
// Parameters:
//   - name: The tool name
//
// Returns formatted help text or an error if not found.
func (d *toolDiscovery) GetToolHelp(name string) (string, error) {
	schema, err := d.GetToolSchema(name)
	if err != nil {
		return "", err
	}

	var help strings.Builder
	help.WriteString(fmt.Sprintf("Tool: %s\n", schema.Name))
	help.WriteString(fmt.Sprintf("Description: %s\n", schema.Description))

	if schema.Parameters != nil {
		help.WriteString("\nParameters:\n")
		paramJSON, _ := json.MarshalIndent(schema.Parameters, "  ", "  ")
		help.Write(paramJSON)
		help.WriteString("\n")
	}

	if len(schema.Examples) > 0 {
		help.WriteString("\nExamples:\n")
		for _, ex := range schema.Examples {
			help.WriteString(fmt.Sprintf("  - %s: %s\n", ex.Name, ex.Description))
			if ex.Input != nil {
				inputJSON, _ := json.MarshalIndent(ex.Input, "    ", "  ")
				help.WriteString("    Input:\n    ")
				help.Write(inputJSON)
				help.WriteString("\n")
			}
		}
	}

	if len(schema.Constraints) > 0 {
		help.WriteString("\nConstraints:\n")
		for _, c := range schema.Constraints {
			help.WriteString(fmt.Sprintf("  - %s\n", c))
		}
	}

	return help.String(), nil
}

// GetToolMetadata returns the metadata for all tools without requiring imports.
// This is a convenience function for scripting bridges and dynamic tool access.
//
// Returns a map of tool names to their metadata.
func GetToolMetadata() map[string]ToolInfo {
	discovery := NewDiscovery()
	tools := discovery.ListTools()

	result := make(map[string]ToolInfo)
	for _, tool := range tools {
		result[tool.Name] = tool
	}
	return result
}

// ========== DYNAMIC REGISTRATION METHODS (REQUIRED FOR DOWNSTREAM) ==========

// RegisterTool registers a new tool at runtime.
// This is critical for scripting engines and dynamic tool creation.
//
// Parameters:
//   - info: Tool metadata
//   - factory: Factory function to create the tool
//
// Returns an error if the tool already exists.
func (d *toolDiscovery) RegisterTool(info ToolInfo, factory ToolFactory) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	registry := d.getCurrentRegistry()

	if _, exists := registry.Metadata[info.Name]; exists {
		return fmt.Errorf("tool %s already registered in namespace %s", info.Name, d.currentNamespace)
	}

	// Set version if not provided
	if info.Version == "" {
		info.Version = "1.0.0"
	}

	// Create version info
	versionInfo := ToolVersionInfo{
		ToolInfo: info,
		Versions: make(map[string]ToolFactory),
	}
	versionInfo.Versions[info.Version] = factory

	registry.Metadata[info.Name] = versionInfo
	registry.Factories[info.Name] = factory
	return nil
}

// UnregisterTool removes a tool from the registry.
// This allows dynamic tool management and cleanup.
//
// Parameters:
//   - name: The tool name to remove
//
// Returns an error if the tool is not found.
func (d *toolDiscovery) UnregisterTool(name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	registry := d.getCurrentRegistry()

	if _, exists := registry.Metadata[name]; !exists {
		return fmt.Errorf("tool %s not found in namespace %s", name, d.currentNamespace)
	}

	delete(registry.Metadata, name)
	delete(registry.Factories, name)
	return nil
}

// GetRegisteredTools returns all registered tools in current namespace.
// This is equivalent to ListTools() for the current namespace.
//
// Returns a slice of all registered tools.
func (d *toolDiscovery) GetRegisteredTools() []ToolInfo {
	return d.ListTools() // ListTools already works with current namespace
}

// ========== REGISTRY PERSISTENCE (DOWNSTREAM REQUIREMENT) ==========

// RegistrySnapshot represents a serializable snapshot of the registry.
// It captures the complete state for persistence and restoration.
type RegistrySnapshot struct {
	Namespaces map[string]*NamespaceRegistry `json:"namespaces"`
	Current    string                        `json:"current_namespace"`
}

// SaveRegistry serializes the registry to a writer.
// This enables persistence of dynamically registered tools.
//
// Parameters:
//   - writer: An io.Writer to write the serialized data
//
// Returns an error if serialization fails.
func (d *toolDiscovery) SaveRegistry(writer interface{}) error {
	w, ok := writer.(io.Writer)
	if !ok {
		return fmt.Errorf("writer must implement io.Writer")
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	snapshot := RegistrySnapshot{
		Namespaces: d.namespaces,
		Current:    d.currentNamespace,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(snapshot)
}

// LoadRegistry deserializes the registry from a reader.
// Note that factories must be re-registered after loading.
//
// Parameters:
//   - reader: An io.Reader to read the serialized data
//
// Returns an error if deserialization fails.
func (d *toolDiscovery) LoadRegistry(reader interface{}) error {
	r, ok := reader.(io.Reader)
	if !ok {
		return fmt.Errorf("reader must implement io.Reader")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	var snapshot RegistrySnapshot
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&snapshot); err != nil {
		return fmt.Errorf("failed to decode registry: %w", err)
	}

	// Restore namespaces (factories will be nil, need to be re-registered)
	d.namespaces = snapshot.Namespaces
	d.currentNamespace = snapshot.Current

	// Initialize factory maps for each namespace
	for _, ns := range d.namespaces {
		if ns.Factories == nil {
			ns.Factories = make(map[string]ToolFactory)
		}
	}

	return nil
}

// ========== TOOL VERSIONING SUPPORT ==========

// RegisterToolVersion registers a specific version of a tool.
// This enables version management and backward compatibility.
//
// Parameters:
//   - info: Tool metadata
//   - factory: Factory function for this version
//   - version: Version string (e.g., "2.0.0")
//
// Returns an error if registration fails.
func (d *toolDiscovery) RegisterToolVersion(info ToolInfo, factory ToolFactory, version string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	registry := d.getCurrentRegistry()

	versionInfo, exists := registry.Metadata[info.Name]
	if !exists {
		// Create new tool with this version
		versionInfo = ToolVersionInfo{
			ToolInfo: info,
			Versions: make(map[string]ToolFactory),
		}
	}

	// Set the version
	info.Version = version
	versionInfo.Versions[version] = factory

	// If this is the latest version, update the main metadata and factory
	if !exists || isNewerVersion(version, versionInfo.Version) {
		versionInfo.ToolInfo = info
		registry.Factories[info.Name] = factory
	}

	registry.Metadata[info.Name] = versionInfo
	return nil
}

// GetToolVersions returns all available versions of a tool.
// This helps with version discovery and selection.
//
// Parameters:
//   - name: The tool name
//
// Returns a slice of version strings.
func (d *toolDiscovery) GetToolVersions(name string) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	registry := d.getCurrentRegistry()
	versionInfo, exists := registry.Metadata[name]
	if !exists {
		return []string{}
	}

	versions := make([]string, 0, len(versionInfo.Versions))
	for version := range versionInfo.Versions {
		versions = append(versions, version)
	}
	return versions
}

// CreateToolVersion creates a tool instance of a specific version.
// This enables using older versions for compatibility.
//
// Parameters:
//   - name: The tool name
//   - version: The specific version to create
//
// Returns the versioned tool or an error if not found.
func (d *toolDiscovery) CreateToolVersion(name, version string) (domain.Tool, error) {
	d.mu.RLock()
	registry := d.getCurrentRegistry()
	versionInfo, exists := registry.Metadata[name]
	d.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	factory, versionExists := versionInfo.Versions[version]
	if !versionExists {
		return nil, fmt.Errorf("tool %s version %s not found", name, version)
	}

	return factory()
}

// ========== MULTI-TENANT SUPPORT (FOR DOWNSTREAM ISOLATION) ==========

// CreateNamespace creates a new tool namespace.
// Namespaces provide isolation for multi-tenant scenarios.
//
// Parameters:
//   - namespace: The namespace name
//
// Returns an error if the namespace already exists.
func (d *toolDiscovery) CreateNamespace(namespace string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.namespaces[namespace]; exists {
		return fmt.Errorf("namespace %s already exists", namespace)
	}

	d.namespaces[namespace] = &NamespaceRegistry{
		Name:      namespace,
		Metadata:  make(map[string]ToolVersionInfo),
		Factories: make(map[string]ToolFactory),
	}
	return nil
}

// ListNamespaces returns all available namespaces.
// This helps with namespace discovery and management.
//
// Returns a slice of namespace names.
func (d *toolDiscovery) ListNamespaces() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	namespaces := make([]string, 0, len(d.namespaces))
	for name := range d.namespaces {
		namespaces = append(namespaces, name)
	}
	return namespaces
}

// SwitchNamespace changes the current namespace.
// All subsequent operations will use the new namespace.
//
// Parameters:
//   - namespace: The namespace to switch to
//
// Returns an error if the namespace doesn't exist.
func (d *toolDiscovery) SwitchNamespace(namespace string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.namespaces[namespace]; !exists {
		return fmt.Errorf("namespace %s does not exist", namespace)
	}

	d.currentNamespace = namespace
	return nil
}

// GetCurrentNamespace returns the currently active namespace.
// The default namespace is "default".
//
// Returns the current namespace name.
func (d *toolDiscovery) GetCurrentNamespace() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentNamespace
}

// ========== HELPER FUNCTIONS ==========

// isNewerVersion compares two semantic version strings.
// This is a simple implementation using string comparison.
// In production, use proper semantic versioning libraries.
//
// Parameters:
//   - v1: First version
//   - v2: Second version
//
// Returns true if v1 is newer than v2.
func isNewerVersion(v1, v2 string) bool {
	// Simple string comparison for now - in production would use proper semver
	return v1 > v2
}
