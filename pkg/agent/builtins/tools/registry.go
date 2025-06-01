// ABOUTME: Tool-specific registry with resource usage and permission tracking
// ABOUTME: Extends base registry with tool-specific metadata and validation

package tools

import (
	"fmt"

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
}

// ToolMetadata extends base metadata for tools
type ToolMetadata struct {
	builtins.Metadata
	RequiredPermissions []string     `json:"required_permissions,omitempty"`
	ResourceUsage       ResourceInfo `json:"resource_usage,omitempty"`
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

// toolRegistry implements ToolRegistry
type toolRegistry struct {
	builtins.Registry[domain.Tool]
	// Additional tool-specific data could go here
}

// Tools is the global registry for built-in tools
var Tools ToolRegistry = &toolRegistry{
	Registry: builtins.NewRegistry[domain.Tool](),
}

// RegisterTool registers a tool with tool-specific metadata
func (r *toolRegistry) RegisterTool(name string, tool domain.Tool, metadata ToolMetadata) error {
	// Validate tool metadata
	if err := validateToolMetadata(name, tool, metadata); err != nil {
		return fmt.Errorf("invalid tool metadata: %w", err)
	}

	// Convert ToolMetadata to base Metadata for registration
	baseMetadata := metadata.Metadata

	// Store the full metadata in a way we can retrieve it later
	// For now, we'll use the base registry and type assert when needed
	return r.Registry.Register(name, tool, baseMetadata)
}

// ListByPermission returns tools requiring specific permission
func (r *toolRegistry) ListByPermission(permission string) []builtins.RegistryEntry[domain.Tool] {
	allTools := r.List()
	var filtered []builtins.RegistryEntry[domain.Tool]

	for _, entry := range allTools {
		// In a real implementation, we'd store and retrieve the full ToolMetadata
		// For now, we'll check tags as a proxy
		for _, tag := range entry.Metadata.Tags {
			if tag == "perm:"+permission {
				filtered = append(filtered, entry)
				break
			}
		}
	}

	return filtered
}

// ListByResourceUsage returns tools matching resource criteria
func (r *toolRegistry) ListByResourceUsage(criteria ResourceCriteria) []builtins.RegistryEntry[domain.Tool] {
	allTools := r.List()
	var filtered []builtins.RegistryEntry[domain.Tool]

	for _, entry := range allTools {
		// In a real implementation, we'd check against stored ResourceInfo
		// For now, we'll use a simple tag-based approach
		if matchesResourceCriteria(entry.Metadata.Tags, criteria) {
			filtered = append(filtered, entry)
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
