// ABOUTME: Tool discovery system providing metadata-first access without imports
// ABOUTME: Enables dynamic tool exploration for scripting engines and CLI tools

package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ToolInfo represents lightweight tool metadata for discovery
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

// Example represents a simplified example structure
type Example struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Input       json.RawMessage `json:"input"`
	Output      json.RawMessage `json:"output,omitempty"`
}

// ToolFactory is a function that creates a tool on demand
type ToolFactory func() (domain.Tool, error)

// ToolDiscovery provides metadata-first tool discovery
type ToolDiscovery interface {
	// ListTools returns all available tools without loading them
	ListTools() []ToolInfo

	// SearchTools searches tools by keyword in name, description, or tags
	SearchTools(query string) []ToolInfo

	// ListByCategory returns tools in a specific category
	ListByCategory(category string) []ToolInfo

	// GetToolSchema returns detailed schema for a specific tool
	GetToolSchema(name string) (*ToolSchema, error)

	// GetToolExamples returns examples for a specific tool
	GetToolExamples(name string) ([]domain.ToolExample, error)

	// CreateTool instantiates a tool by name
	CreateTool(name string) (domain.Tool, error)

	// CreateTools instantiates multiple tools
	CreateTools(names ...string) (map[string]domain.Tool, error)

	// GetToolHelp generates help text for a tool
	GetToolHelp(name string) (string, error)
}

// ToolSchema contains detailed schema information
type ToolSchema struct {
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Parameters    interface{}          `json:"parameters,omitempty"`
	Output        interface{}          `json:"output,omitempty"`
	Examples      []domain.ToolExample `json:"examples,omitempty"`
	Constraints   []string             `json:"constraints,omitempty"`
	ErrorGuidance map[string]string    `json:"error_guidance,omitempty"`
}

// toolDiscovery implements ToolDiscovery
type toolDiscovery struct {
	metadata  map[string]ToolInfo
	factories map[string]ToolFactory
	mu        sync.RWMutex
}

// globalDiscovery is the singleton discovery instance
var (
	globalDiscovery *toolDiscovery
	discoveryOnce   sync.Once
)

// NewDiscovery returns the global discovery instance
func NewDiscovery() ToolDiscovery {
	discoveryOnce.Do(func() {
		globalDiscovery = &toolDiscovery{
			metadata:  make(map[string]ToolInfo),
			factories: make(map[string]ToolFactory),
		}
		// Initialize with metadata from registry
		globalDiscovery.initializeFromRegistry()
	})
	return globalDiscovery
}

// initializeFromRegistry populates metadata from existing registry
func (d *toolDiscovery) initializeFromRegistry() {
	// Metadata will be populated by the generated registry_metadata.go init() function
	// which calls RegisterToolMetadata for each tool
	// No need to do anything here as the metadata is registered at package init time
}

// RegisterToolMetadata registers tool metadata without the tool instance
func RegisterToolMetadata(info ToolInfo, factory ToolFactory) error {
	discovery := NewDiscovery().(*toolDiscovery)

	discovery.mu.Lock()
	defer discovery.mu.Unlock()

	if _, exists := discovery.metadata[info.Name]; exists {
		return fmt.Errorf("tool %s already registered", info.Name)
	}

	discovery.metadata[info.Name] = info
	discovery.factories[info.Name] = factory
	return nil
}

// ListTools returns all available tools without loading them
func (d *toolDiscovery) ListTools() []ToolInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]ToolInfo, 0, len(d.metadata))
	for _, info := range d.metadata {
		result = append(result, info)
	}
	return result
}

// SearchTools searches tools by keyword
func (d *toolDiscovery) SearchTools(query string) []ToolInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query = strings.ToLower(query)
	var results []ToolInfo

	for _, info := range d.metadata {
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

// ListByCategory returns tools in a specific category
func (d *toolDiscovery) ListByCategory(category string) []ToolInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var results []ToolInfo
	for _, info := range d.metadata {
		if info.Category == category {
			results = append(results, info)
		}
	}
	return results
}

// GetToolSchema returns detailed schema for a specific tool
func (d *toolDiscovery) GetToolSchema(name string) (*ToolSchema, error) {
	d.mu.RLock()
	info, exists := d.metadata[name]
	d.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

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

// GetToolExamples returns examples for a specific tool
func (d *toolDiscovery) GetToolExamples(name string) ([]domain.ToolExample, error) {
	schema, err := d.GetToolSchema(name)
	if err != nil {
		return nil, err
	}
	return schema.Examples, nil
}

// CreateTool instantiates a tool by name
func (d *toolDiscovery) CreateTool(name string) (domain.Tool, error) {
	d.mu.RLock()
	factory, exists := d.factories[name]
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

// CreateTools instantiates multiple tools
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

// GetToolHelp generates help text for a tool
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

// GetToolMetadata returns the metadata for all tools without requiring imports
// This is a convenience function for scripting bridges
func GetToolMetadata() map[string]ToolInfo {
	discovery := NewDiscovery()
	tools := discovery.ListTools()

	result := make(map[string]ToolInfo)
	for _, tool := range tools {
		result[tool.Name] = tool
	}
	return result
}
