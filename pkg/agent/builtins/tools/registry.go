// ABOUTME: Tool-specific registry with resource usage and permission tracking
// ABOUTME: Extends base registry with tool-specific metadata and validation

package tools

import (
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ToolRegistry extends the base registry with tool-specific functionality
type ToolRegistry interface {
	builtins.Registry[domain.Tool]

	// RegisterTool is a convenience method that accepts ToolMetadata
	RegisterTool(name string, tool domain.Tool, metadata ToolMetadata) error

	// ListByPermission returns tools requiring specific permission
	ListByPermission(permission string) []builtins.RegistryEntry[domain.Tool]

	// ListByResourceUsage returns tools matching resource criteria
	ListByResourceUsage(criteria ResourceCriteria) []builtins.RegistryEntry[domain.Tool]

	// ExportToMCP exports a single tool to MCP format
	ExportToMCP(name string) (domain.MCPToolDefinition, error)

	// ExportAllToMCP exports all tools to an MCP catalog
	ExportAllToMCP() (MCPCatalog, error)

	// GetToolDocumentation returns comprehensive documentation for a tool
	GetToolDocumentation(name string) (ToolDocumentation, error)
}

// ToolMetadata extends base metadata for tools
type ToolMetadata struct {
	builtins.Metadata
	RequiredPermissions []string     `json:"required_permissions,omitempty"`
	ResourceUsage       ResourceInfo `json:"resource_usage,omitempty"`

	// Enhanced metadata from tool interface
	UsageInstructions    string               `json:"usage_instructions,omitempty"`
	Examples             []domain.ToolExample `json:"examples,omitempty"`
	Constraints          []string             `json:"constraints,omitempty"`
	ErrorGuidance        map[string]string    `json:"error_guidance,omitempty"`
	IsDeterministic      bool                 `json:"is_deterministic"`
	IsDestructive        bool                 `json:"is_destructive"`
	RequiresConfirmation bool                 `json:"requires_confirmation"`
	EstimatedLatency     string               `json:"estimated_latency,omitempty"`
}

// ResourceInfo describes resource requirements
type ResourceInfo struct {
	Memory      string `json:"memory,omitempty"`      // "low", "medium", "high"
	Network     bool   `json:"network,omitempty"`     // requires network access
	FileSystem  bool   `json:"file_system,omitempty"` // requires file system access
	Concurrency bool   `json:"concurrency,omitempty"` // thread-safe for concurrent use
}

// ResourceCriteria for filtering tools by resource usage
type ResourceCriteria struct {
	MaxMemory          string // Filter tools with memory usage <= this level
	RequiresNetwork    *bool  // Filter by network requirement (nil = don't care)
	RequiresFileSystem *bool  // Filter by file system requirement (nil = don't care)
	RequiresConcurrent *bool  // Filter by concurrency support (nil = don't care)
}

// MCPCatalog represents a catalog of tools in MCP format
type MCPCatalog struct {
	Version     string                     `json:"version"`
	Description string                     `json:"description"`
	Tools       []domain.MCPToolDefinition `json:"tools"`
	Metadata    map[string]interface{}     `json:"metadata,omitempty"`
}

// ToolDocumentation provides comprehensive documentation for a tool
type ToolDocumentation struct {
	Name                 string               `json:"name"`
	Description          string               `json:"description"`
	Category             string               `json:"category"`
	Tags                 []string             `json:"tags"`
	Version              string               `json:"version"`
	UsageInstructions    string               `json:"usage_instructions"`
	Examples             []domain.ToolExample `json:"examples"`
	Constraints          []string             `json:"constraints"`
	ErrorGuidance        map[string]string    `json:"error_guidance"`
	RequiredPermissions  []string             `json:"required_permissions"`
	ResourceUsage        ResourceInfo         `json:"resource_usage"`
	IsDeterministic      bool                 `json:"is_deterministic"`
	IsDestructive        bool                 `json:"is_destructive"`
	RequiresConfirmation bool                 `json:"requires_confirmation"`
	EstimatedLatency     string               `json:"estimated_latency"`
	ParameterSchema      interface{}          `json:"parameter_schema,omitempty"`
	OutputSchema         interface{}          `json:"output_schema,omitempty"`
}

// toolRegistry implements ToolRegistry
type toolRegistry struct {
	builtins.Registry[domain.Tool]
	// Store enhanced metadata separately for retrieval
	toolMetadata map[string]ToolMetadata
	mu           sync.RWMutex
}

// Tools is the global registry for built-in tools
var Tools ToolRegistry = &toolRegistry{
	Registry:     builtins.NewRegistry[domain.Tool](),
	toolMetadata: make(map[string]ToolMetadata),
}

// RegisterTool registers a tool with tool-specific metadata
func (r *toolRegistry) RegisterTool(name string, tool domain.Tool, metadata ToolMetadata) error {
	// Validate tool metadata
	if err := validateToolMetadata(name, tool, metadata); err != nil {
		return fmt.Errorf("invalid tool metadata: %w", err)
	}

	// If metadata doesn't have values from tool interface, populate them
	if metadata.UsageInstructions == "" && tool.UsageInstructions() != "" {
		metadata.UsageInstructions = tool.UsageInstructions()
	}
	if len(metadata.Examples) == 0 && tool.Examples() != nil {
		metadata.Examples = tool.Examples()
	}
	if len(metadata.Constraints) == 0 && tool.Constraints() != nil {
		metadata.Constraints = tool.Constraints()
	}
	if metadata.ErrorGuidance == nil && tool.ErrorGuidance() != nil {
		metadata.ErrorGuidance = tool.ErrorGuidance()
	}
	if metadata.Category == "" && tool.Category() != "" {
		metadata.Category = tool.Category()
	}
	if len(metadata.Tags) == 0 && tool.Tags() != nil {
		metadata.Tags = tool.Tags()
	}
	if metadata.Version == "" && tool.Version() != "" {
		metadata.Version = tool.Version()
	}
	metadata.IsDeterministic = tool.IsDeterministic()
	metadata.IsDestructive = tool.IsDestructive()
	metadata.RequiresConfirmation = tool.RequiresConfirmation()
	if metadata.EstimatedLatency == "" && tool.EstimatedLatency() != "" {
		metadata.EstimatedLatency = tool.EstimatedLatency()
	}

	// Store the enhanced metadata
	r.mu.Lock()
	r.toolMetadata[name] = metadata
	r.mu.Unlock()

	// Convert ToolMetadata to base Metadata for registration
	baseMetadata := metadata.Metadata

	// Register with base registry
	return r.Register(name, tool, baseMetadata)
}

// ListByPermission returns tools requiring specific permission
func (r *toolRegistry) ListByPermission(permission string) []builtins.RegistryEntry[domain.Tool] {
	allTools := r.List()
	var filtered []builtins.RegistryEntry[domain.Tool]

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, entry := range allTools {
		// Check enhanced metadata for permissions
		if metadata, exists := r.toolMetadata[entry.Metadata.Name]; exists {
			for _, perm := range metadata.RequiredPermissions {
				if perm == permission {
					filtered = append(filtered, entry)
					break
				}
			}
		}
	}

	return filtered
}

// ListByResourceUsage returns tools matching resource criteria
func (r *toolRegistry) ListByResourceUsage(criteria ResourceCriteria) []builtins.RegistryEntry[domain.Tool] {
	allTools := r.List()
	var filtered []builtins.RegistryEntry[domain.Tool]

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, entry := range allTools {
		// Check enhanced metadata for resource usage
		if metadata, exists := r.toolMetadata[entry.Metadata.Name]; exists {
			if matchesResourceCriteriaEnhanced(metadata.ResourceUsage, criteria) {
				filtered = append(filtered, entry)
			}
		}
	}

	return filtered
}

// validateToolMetadata ensures the tool metadata is valid
func validateToolMetadata(name string, tool domain.Tool, metadata ToolMetadata) error {
	// Ensure tool name matches
	if tool.Name() != name {
		return fmt.Errorf("tool name '%s' does not match registration name '%s'", tool.Name(), name)
	}

	// Validate memory usage level
	if metadata.ResourceUsage.Memory != "" {
		validMemoryLevels := map[string]bool{"low": true, "medium": true, "high": true}
		if !validMemoryLevels[metadata.ResourceUsage.Memory] {
			return fmt.Errorf("invalid memory usage level: %s", metadata.ResourceUsage.Memory)
		}
	}

	// Validate permissions format
	for _, perm := range metadata.RequiredPermissions {
		if perm == "" {
			return fmt.Errorf("empty permission string")
		}
		// Could add more validation here (e.g., format like "resource:action")
	}

	return nil
}

// matchesResourceCriteriaEnhanced checks if resource info matches criteria
func matchesResourceCriteriaEnhanced(info ResourceInfo, criteria ResourceCriteria) bool {
	// Check memory constraint
	if criteria.MaxMemory != "" && info.Memory != "" {
		memoryLevels := map[string]int{"low": 1, "medium": 2, "high": 3}
		maxLevel, maxExists := memoryLevels[criteria.MaxMemory]
		infoLevel, infoExists := memoryLevels[info.Memory]

		if maxExists && infoExists && infoLevel > maxLevel {
			return false
		}
	}

	// Check network requirement
	if criteria.RequiresNetwork != nil && *criteria.RequiresNetwork != info.Network {
		return false
	}

	// Check file system requirement
	if criteria.RequiresFileSystem != nil && *criteria.RequiresFileSystem != info.FileSystem {
		return false
	}

	// Check concurrency support
	if criteria.RequiresConcurrent != nil && *criteria.RequiresConcurrent != info.Concurrency {
		return false
	}

	return true
}

/*
// matchesResourceCriteria checks if tags indicate matching resource usage
func matchesResourceCriteria(tags []string, criteria ResourceCriteria) bool {
	// This is a simplified implementation using tags
	// In a real implementation, we'd check against stored ResourceInfo

	// Check memory constraint
	if criteria.MaxMemory != "" {
		memoryLevels := map[string]int{"low": 1, "medium": 2, "high": 3}
		maxLevel := memoryLevels[criteria.MaxMemory]

		for _, tag := range tags {
			if tag == "memory:medium" && maxLevel < 2 {
				return false
			}
			if tag == "memory:high" && maxLevel < 3 {
				return false
			}
		}
	}

	// Check network requirement
	if criteria.RequiresNetwork != nil {
		hasNetwork := false
		for _, tag := range tags {
			if tag == "network" {
				hasNetwork = true
				break
			}
		}
		if *criteria.RequiresNetwork != hasNetwork {
			return false
		}
	}

	// Check file system requirement
	if criteria.RequiresFileSystem != nil {
		hasFileSystem := false
		for _, tag := range tags {
			if tag == "filesystem" {
				hasFileSystem = true
				break
			}
		}
		if *criteria.RequiresFileSystem != hasFileSystem {
			return false
		}
	}

	// Check concurrency support
	if criteria.RequiresConcurrent != nil {
		hasConcurrency := false
		for _, tag := range tags {
			if tag == "concurrent" {
				hasConcurrency = true
				break
			}
		}
		if *criteria.RequiresConcurrent != hasConcurrency {
			return false
		}
	}

	return true
}
*/

// ExportToMCP exports a single tool to MCP format
func (r *toolRegistry) ExportToMCP(name string) (domain.MCPToolDefinition, error) {
	tool, found := r.Get(name)
	if !found {
		return domain.MCPToolDefinition{}, fmt.Errorf("tool '%s' not found", name)
	}

	// Use the tool's own MCP export method
	return tool.ToMCPDefinition(), nil
}

// ExportAllToMCP exports all tools to an MCP catalog
func (r *toolRegistry) ExportAllToMCP() (MCPCatalog, error) {
	allTools := r.List()

	catalog := MCPCatalog{
		Version:     "1.0.0",
		Description: "Go-LLMs Tool Catalog",
		Tools:       make([]domain.MCPToolDefinition, 0, len(allTools)),
		Metadata: map[string]interface{}{
			"generated_at": time.Now().UTC().Format(time.RFC3339),
			"tool_count":   len(allTools),
		},
	}

	for _, entry := range allTools {
		mcp := entry.Component.ToMCPDefinition()
		catalog.Tools = append(catalog.Tools, mcp)
	}

	return catalog, nil
}

// GetToolDocumentation returns comprehensive documentation for a tool
func (r *toolRegistry) GetToolDocumentation(name string) (ToolDocumentation, error) {
	tool, found := r.Get(name)
	if !found {
		return ToolDocumentation{}, fmt.Errorf("tool '%s' not found", name)
	}

	// Get enhanced metadata
	r.mu.RLock()
	metadata, hasMetadata := r.toolMetadata[name]
	r.mu.RUnlock()

	doc := ToolDocumentation{
		Name:                 tool.Name(),
		Description:          tool.Description(),
		Category:             tool.Category(),
		Tags:                 tool.Tags(),
		Version:              tool.Version(),
		UsageInstructions:    tool.UsageInstructions(),
		Examples:             tool.Examples(),
		Constraints:          tool.Constraints(),
		ErrorGuidance:        tool.ErrorGuidance(),
		IsDeterministic:      tool.IsDeterministic(),
		IsDestructive:        tool.IsDestructive(),
		RequiresConfirmation: tool.RequiresConfirmation(),
		EstimatedLatency:     tool.EstimatedLatency(),
	}

	// Add schema information
	if tool.ParameterSchema() != nil {
		doc.ParameterSchema = tool.ParameterSchema()
	}
	if tool.OutputSchema() != nil {
		doc.OutputSchema = tool.OutputSchema()
	}

	// Add metadata if available
	if hasMetadata {
		doc.RequiredPermissions = metadata.RequiredPermissions
		doc.ResourceUsage = metadata.ResourceUsage
	}

	return doc, nil
}

// Clear removes all entries (useful for testing)
func (r *toolRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear enhanced metadata
	r.toolMetadata = make(map[string]ToolMetadata)

	// Clear base registry
	r.Registry.Clear()
}

// NewTestRegistry creates a new registry instance for testing
func NewTestRegistry() ToolRegistry {
	return &toolRegistry{
		Registry:     builtins.NewRegistry[domain.Tool](),
		toolMetadata: make(map[string]ToolMetadata),
	}
}

// Helper functions for common tool registrations

// MustRegisterTool registers a tool or panics on error
func MustRegisterTool(name string, tool domain.Tool, metadata ToolMetadata) {
	if err := Tools.RegisterTool(name, tool, metadata); err != nil {
		panic(fmt.Sprintf("failed to register tool '%s': %v", name, err))
	}
}

// GetTool retrieves a tool by name with type assertion
func GetTool(name string) (domain.Tool, bool) {
	return Tools.Get(name)
}

// MustGetTool retrieves a tool by name or panics
func MustGetTool(name string) domain.Tool {
	return Tools.MustGet(name)
}
