// ABOUTME: Tests for tool-specific registry functionality
// ABOUTME: Verifies tool metadata validation and resource filtering

package tools

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// Helper function to create mock tools with specific configurations
func createMockTool(name, description string) *mocks.MockTool {
	return mocks.NewMockTool(name, description).
		WithExecutor(func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return "mock result", nil
		})
}

// Create a mock tool with custom properties
func createCustomMockTool(name, description, category string, tags []string, version string) *mocks.MockTool {
	tool := mocks.NewMockTool(name, description).
		WithCategory(category).
		WithTags(tags...).
		WithVersion(version).
		WithExecutor(func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return "mock result", nil
		})

	// Note: MockTool has fixed behavioral properties, but we can still use it for testing
	return tool
}

func TestToolRegistry_RegisterTool(t *testing.T) {
	// Create a new registry for each test to avoid interference
	registry := &toolRegistry{
		Registry:     builtins.NewRegistry[domain.Tool](),
		toolMetadata: make(map[string]ToolMetadata),
	}

	tests := []struct {
		name     string
		toolName string
		tool     domain.Tool
		metadata ToolMetadata
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "successful registration",
			toolName: "test_tool",
			tool:     createMockTool("test_tool", "Test tool"),
			metadata: ToolMetadata{
				Metadata: builtins.Metadata{
					Category:    "test",
					Description: "A test tool",
				},
				RequiredPermissions: []string{"test:read"},
				ResourceUsage: ResourceInfo{
					Memory:      "low",
					Network:     false,
					FileSystem:  true,
					Concurrency: true,
				},
			},
			wantErr: false,
		},
		{
			name:     "tool name mismatch",
			toolName: "test_tool",
			tool:     createMockTool("different_name", "Test tool"),
			metadata: ToolMetadata{
				Metadata: builtins.Metadata{
					Category: "test",
				},
			},
			wantErr: true,
			errMsg:  "invalid tool metadata: tool name 'different_name' does not match registration name 'test_tool'",
		},
		{
			name:     "invalid memory level",
			toolName: "test_tool",
			tool:     createMockTool("test_tool", "Test tool"),
			metadata: ToolMetadata{
				Metadata: builtins.Metadata{
					Category: "test",
				},
				ResourceUsage: ResourceInfo{
					Memory: "extreme",
				},
			},
			wantErr: true,
			errMsg:  "invalid tool metadata: invalid memory usage level: extreme",
		},
		{
			name:     "empty permission",
			toolName: "test_tool",
			tool:     createMockTool("test_tool", "Test tool"),
			metadata: ToolMetadata{
				Metadata: builtins.Metadata{
					Category: "test",
				},
				RequiredPermissions: []string{"test:read", ""},
			},
			wantErr: true,
			errMsg:  "invalid tool metadata: empty permission string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterTool(tt.toolName, tt.tool, tt.metadata)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestToolRegistry_ListByPermission(t *testing.T) {
	registry := &toolRegistry{
		Registry:     builtins.NewRegistry[domain.Tool](),
		toolMetadata: make(map[string]ToolMetadata),
	}

	// Register tools with different permissions
	tool1 := createMockTool("file_reader", "Reads files")
	if err := registry.RegisterTool("file_reader", tool1, ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_reader",
			Description: "Reads files",
			Tags:        []string{"file"},
		},
		RequiredPermissions: []string{"file:read"},
	}); err != nil {
		t.Fatalf("failed to register file_reader: %v", err)
	}

	tool2 := createMockTool("file_writer", "Writes files")
	if err := registry.RegisterTool("file_writer", tool2, ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_writer",
			Description: "Writes files",
			Tags:        []string{"file"},
		},
		RequiredPermissions: []string{"file:write"},
	}); err != nil {
		t.Fatalf("failed to register file_writer: %v", err)
	}

	tool3 := createMockTool("network_fetch", "Fetches from network")
	if err := registry.RegisterTool("network_fetch", tool3, ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "network_fetch",
			Description: "Fetches from network",
			Tags:        []string{"network"},
		},
		RequiredPermissions: []string{"network:access"},
	}); err != nil {
		t.Fatalf("failed to register network_fetch: %v", err)
	}

	// Test permission filtering
	fileReadTools := registry.ListByPermission("file:read")
	if len(fileReadTools) != 1 {
		t.Errorf("expected 1 tool with file:read permission, got %d", len(fileReadTools))
	}

	networkTools := registry.ListByPermission("network:access")
	if len(networkTools) != 1 {
		t.Errorf("expected 1 tool with network:access permission, got %d", len(networkTools))
	}

	noTools := registry.ListByPermission("nonexistent:permission")
	if len(noTools) != 0 {
		t.Errorf("expected 0 tools with nonexistent permission, got %d", len(noTools))
	}
}

func TestToolRegistry_ListByResourceUsage(t *testing.T) {
	registry := &toolRegistry{
		Registry:     builtins.NewRegistry[domain.Tool](),
		toolMetadata: make(map[string]ToolMetadata),
	}

	// Register tools with different resource characteristics
	tool1 := createMockTool("low_memory_tool", "Uses low memory")
	if err := registry.RegisterTool("low_memory_tool", tool1, ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "low_memory_tool",
			Description: "Uses low memory",
		},
		ResourceUsage: ResourceInfo{
			Memory:      "low",
			Concurrency: true,
		},
	}); err != nil {
		t.Fatalf("failed to register low_memory_tool: %v", err)
	}

	tool2 := createMockTool("high_memory_network", "High memory network tool")
	if err := registry.RegisterTool("high_memory_network", tool2, ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "high_memory_network",
			Description: "High memory network tool",
		},
		ResourceUsage: ResourceInfo{
			Memory:      "high",
			Network:     true,
			Concurrency: true,
		},
	}); err != nil {
		t.Fatalf("failed to register high_memory_network: %v", err)
	}

	tool3 := createMockTool("file_tool", "File system tool")
	if err := registry.RegisterTool("file_tool", tool3, ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "file_tool",
			Description: "File system tool",
		},
		ResourceUsage: ResourceInfo{
			Memory:     "medium",
			FileSystem: true,
		},
	}); err != nil {
		t.Fatalf("failed to register file_tool: %v", err)
	}

	tests := []struct {
		name     string
		criteria ResourceCriteria
		expected int
	}{
		{
			name: "filter by max memory",
			criteria: ResourceCriteria{
				MaxMemory: "medium",
			},
			expected: 2, // low_memory_tool and file_tool
		},
		{
			name: "filter by network requirement",
			criteria: ResourceCriteria{
				RequiresNetwork: boolPtr(true),
			},
			expected: 1, // high_memory_network
		},
		{
			name: "filter by no network requirement",
			criteria: ResourceCriteria{
				RequiresNetwork: boolPtr(false),
			},
			expected: 2, // low_memory_tool and file_tool
		},
		{
			name: "filter by filesystem requirement",
			criteria: ResourceCriteria{
				RequiresFileSystem: boolPtr(true),
			},
			expected: 1, // file_tool
		},
		{
			name: "filter by concurrency support",
			criteria: ResourceCriteria{
				RequiresConcurrent: boolPtr(true),
			},
			expected: 2, // low_memory_tool and high_memory_network
		},
		{
			name: "combined filters",
			criteria: ResourceCriteria{
				MaxMemory:          "low",
				RequiresConcurrent: boolPtr(true),
			},
			expected: 1, // only low_memory_tool
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.ListByResourceUsage(tt.criteria)
			if len(results) != tt.expected {
				t.Errorf("expected %d tools matching criteria, got %d", tt.expected, len(results))
			}
		})
	}
}

func TestToolRegistry_GlobalRegistry(t *testing.T) {
	// Save the current registry and restore it after the test
	oldRegistry := Tools
	defer func() { Tools = oldRegistry }()

	// Create a fresh registry for testing
	Tools = &toolRegistry{
		Registry:     builtins.NewRegistry[domain.Tool](),
		toolMetadata: make(map[string]ToolMetadata),
	}

	// Test MustRegisterTool
	tool := createMockTool("global_tool", "Global tool")
	MustRegisterTool("global_tool", tool, ToolMetadata{
		Metadata: builtins.Metadata{
			Category:    "global",
			Description: "A globally registered tool",
		},
		RequiredPermissions: []string{"global:access"},
		ResourceUsage: ResourceInfo{
			Memory: "low",
		},
	})

	// Test GetTool
	retrievedTool, found := GetTool("global_tool")
	if !found {
		t.Error("expected to find globally registered tool")
	}
	if retrievedTool.Name() != "global_tool" {
		t.Errorf("expected tool name 'global_tool', got '%s'", retrievedTool.Name())
	}

	// Test MustGetTool
	mustGetTool := MustGetTool("global_tool")
	if mustGetTool.Name() != "global_tool" {
		t.Errorf("expected tool name 'global_tool', got '%s'", mustGetTool.Name())
	}

	// Test panic on duplicate registration
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	MustRegisterTool("global_tool", tool, ToolMetadata{
		Metadata: builtins.Metadata{Category: "duplicate"},
	})
}

func TestToolRegistry_Validation(t *testing.T) {
	registry := &toolRegistry{
		Registry:     builtins.NewRegistry[domain.Tool](),
		toolMetadata: make(map[string]ToolMetadata),
	}

	// Test various validation scenarios
	tests := []struct {
		name        string
		toolName    string
		tool        domain.Tool
		permissions []string
		memory      string
		wantErr     bool
	}{
		{
			name:        "valid permissions format",
			toolName:    "test1",
			tool:        createMockTool("test1", "Test tool 1"),
			permissions: []string{"resource:action", "another:permission"},
			memory:      "medium",
			wantErr:     false,
		},
		{
			name:        "valid memory levels",
			toolName:    "test2",
			tool:        createMockTool("test2", "Test tool 2"),
			permissions: []string{},
			memory:      "high",
			wantErr:     false,
		},
		{
			name:        "empty memory is valid",
			toolName:    "test3",
			tool:        createMockTool("test3", "Test tool 3"),
			permissions: []string{},
			memory:      "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterTool(tt.toolName, tt.tool, ToolMetadata{
				Metadata: builtins.Metadata{
					Category: "test",
				},
				RequiredPermissions: tt.permissions,
				ResourceUsage: ResourceInfo{
					Memory: tt.memory,
				},
			})

			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}

// Enhanced tests for new registry features

func TestEnhancedToolMetadata(t *testing.T) {
	// Create a fresh registry for testing
	registry := NewTestRegistry()

	t.Run("register tool with enhanced metadata", func(t *testing.T) {
		tool := createCustomMockTool("test_tool", "A test tool", "test",
			[]string{"test", "mock"}, "1.0.0").
			WithUsageInstructions("Use this for testing")

		// Create enhanced metadata that pulls from tool
		metadata := ToolMetadata{
			Metadata: builtins.Metadata{
				Name:        tool.Name(),
				Category:    tool.Category(),
				Tags:        tool.Tags(),
				Description: tool.Description(),
				Version:     tool.Version(),
			},
			RequiredPermissions: []string{"test:read"},
			ResourceUsage: ResourceInfo{
				Memory:      "low",
				Network:     false,
				FileSystem:  false,
				Concurrency: true,
			},
		}

		err := registry.RegisterTool(tool.Name(), tool, metadata)
		if err != nil {
			t.Errorf("Failed to register tool: %v", err)
		}

		// Verify tool is registered
		retrieved, found := registry.Get(tool.Name())
		if !found {
			t.Error("Tool not found after registration")
		}
		if retrieved.Name() != tool.Name() {
			t.Errorf("Retrieved tool name mismatch: got %s, want %s", retrieved.Name(), tool.Name())
		}
	})

	t.Run("enhanced metadata populated from tool interface", func(t *testing.T) {
		registry.Clear()

		tool := createCustomMockTool("auto_tool", "Auto metadata tool", "auto",
			[]string{"auto", "test"}, "2.0.0").
			WithUsageInstructions("Auto-populated instructions").
			WithConstraints("Must be deterministic")

		// Register with minimal metadata - should auto-populate from tool
		metadata := ToolMetadata{
			Metadata: builtins.Metadata{
				Name: tool.Name(),
			},
		}

		err := registry.RegisterTool(tool.Name(), tool, metadata)
		if err != nil {
			t.Errorf("Failed to register tool: %v", err)
		}

		// Get documentation to verify metadata was populated
		doc, err := registry.GetToolDocumentation(tool.Name())
		if err != nil {
			t.Errorf("Failed to get tool documentation: %v", err)
		}

		if doc.Category != "auto" {
			t.Errorf("Category not populated: got %s, want auto", doc.Category)
		}
		if doc.Version != "2.0.0" {
			t.Errorf("Version not populated: got %s, want 2.0.0", doc.Version)
		}
		if doc.UsageInstructions != "Auto-populated instructions" {
			t.Errorf("Usage instructions not populated")
		}
	})
}

func TestMCPExportFunctionality(t *testing.T) {
	registry := NewTestRegistry()
	registry.Clear()

	t.Run("export single tool to MCP", func(t *testing.T) {
		tool := createCustomMockTool("calculator", "Performs calculations", "math",
			[]string{"calculation", "math"}, "2.0.0").
			WithExamples(domain.ToolExample{
				Name:        "Addition",
				Description: "Add two numbers",
				Input:       map[string]interface{}{"operation": "add", "a": 5, "b": 3},
				Output:      8,
			}).
			WithConstraints("Only basic operations").
			WithErrorGuidance(map[string]string{
				"division_by_zero": "Cannot divide by zero",
			})

		metadata := ToolMetadata{
			Metadata: builtins.Metadata{
				Name:        tool.Name(),
				Category:    tool.Category(),
				Description: tool.Description(),
				Version:     tool.Version(),
			},
		}

		// Register the tool
		if err := registry.RegisterTool(tool.Name(), tool, metadata); err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		// Export to MCP
		mcp, err := registry.ExportToMCP(tool.Name())
		if err != nil {
			t.Errorf("Failed to export tool to MCP: %v", err)
		}

		// Verify MCP structure
		if mcp.Name != tool.Name() {
			t.Errorf("MCP name mismatch: got %s, want %s", mcp.Name, tool.Name())
		}
		if mcp.Description != tool.Description() {
			t.Errorf("MCP description mismatch")
		}
	})

	t.Run("export all tools to MCP catalog", func(t *testing.T) {
		// Clear and register multiple tools
		registry.Clear()

		toolsToRegister := []domain.Tool{
			createCustomMockTool("web_search", "Search the web", "web",
				[]string{"web", "search"}, "1.0.0"),
			createCustomMockTool("file_read", "Read files", "file",
				[]string{"file", "read"}, "1.0.0"),
		}

		for _, tool := range toolsToRegister {
			metadata := ToolMetadata{
				Metadata: builtins.Metadata{
					Name:        tool.Name(),
					Category:    tool.Category(),
					Description: tool.Description(),
					Version:     tool.Version(),
				},
			}
			if err := registry.RegisterTool(tool.Name(), tool, metadata); err != nil {
				t.Errorf("Failed to register tool %s: %v", tool.Name(), err)
			}
		}

		// Export all to MCP catalog
		catalog, err := registry.ExportAllToMCP()
		if err != nil {
			t.Errorf("Failed to export catalog: %v", err)
		}

		if len(catalog.Tools) != 2 {
			t.Errorf("Expected 2 tools in catalog, got %d", len(catalog.Tools))
		}

		// Verify catalog metadata
		if catalog.Version == "" {
			t.Error("Catalog should have version")
		}
		if catalog.Description == "" {
			t.Error("Catalog should have description")
		}
	})
}

func TestEnhancedResourceFiltering(t *testing.T) {
	registry := NewTestRegistry()
	registry.Clear()

	// Register tools with different resource requirements
	tools := []struct {
		tool     domain.Tool
		metadata ToolMetadata
	}{
		{
			tool: createCustomMockTool("web_tool", "Web tool", "web",
				[]string{"web"}, "1.0.0"),
			metadata: ToolMetadata{
				Metadata: builtins.Metadata{
					Name:     "web_tool",
					Category: "web",
				},
				RequiredPermissions: []string{"network:read"},
				ResourceUsage: ResourceInfo{
					Memory:  "medium",
					Network: true,
				},
			},
		},
		{
			tool: createCustomMockTool("file_tool", "File tool", "file",
				[]string{"file"}, "1.0.0"),
			metadata: ToolMetadata{
				Metadata: builtins.Metadata{
					Name:     "file_tool",
					Category: "file",
				},
				RequiredPermissions: []string{"file:read", "file:write"},
				ResourceUsage: ResourceInfo{
					Memory:     "low",
					FileSystem: true,
				},
			},
		},
		{
			tool: createCustomMockTool("compute_tool", "Compute tool", "compute",
				[]string{"compute"}, "1.0.0"),
			metadata: ToolMetadata{
				Metadata: builtins.Metadata{
					Name:     "compute_tool",
					Category: "compute",
				},
				ResourceUsage: ResourceInfo{
					Memory:      "high",
					Concurrency: true,
				},
			},
		},
	}

	// Register all tools
	for _, tt := range tools {
		if err := registry.RegisterTool(tt.tool.Name(), tt.tool, tt.metadata); err != nil {
			t.Errorf("Failed to register tool %s: %v", tt.tool.Name(), err)
		}
	}

	t.Run("filter by permission", func(t *testing.T) {
		networkTools := registry.ListByPermission("network:read")
		if len(networkTools) != 1 {
			t.Errorf("Expected 1 network tool, got %d", len(networkTools))
		}

		fileWriteTools := registry.ListByPermission("file:write")
		if len(fileWriteTools) != 1 {
			t.Errorf("Expected 1 file write tool, got %d", len(fileWriteTools))
		}
	})

	t.Run("filter by resource usage", func(t *testing.T) {
		// Tools requiring network
		networkRequired := true
		criteria := ResourceCriteria{
			RequiresNetwork: &networkRequired,
		}
		networkTools := registry.ListByResourceUsage(criteria)
		if len(networkTools) != 1 {
			t.Errorf("Expected 1 network tool, got %d", len(networkTools))
		}

		// Tools with memory <= medium
		criteria = ResourceCriteria{
			MaxMemory: "medium",
		}
		lowMemTools := registry.ListByResourceUsage(criteria)
		if len(lowMemTools) != 2 { // web_tool (medium) and file_tool (low)
			t.Errorf("Expected 2 low/medium memory tools, got %d", len(lowMemTools))
		}

		// Tools supporting concurrency
		concurrent := true
		criteria = ResourceCriteria{
			RequiresConcurrent: &concurrent,
		}
		concurrentTools := registry.ListByResourceUsage(criteria)
		if len(concurrentTools) != 1 {
			t.Errorf("Expected 1 concurrent tool, got %d", len(concurrentTools))
		}
	})
}

func TestToolDocumentation(t *testing.T) {
	registry := NewTestRegistry()
	registry.Clear()

	tool := createCustomMockTool("doc_tool", "Documentation test tool", "test",
		[]string{"test", "documentation"}, "3.0.0").
		WithUsageInstructions("Detailed usage instructions here").
		WithExamples(domain.ToolExample{
			Name:        "Example 1",
			Description: "Basic example",
			Input:       "test input",
			Output:      "test output",
		}).
		WithConstraints("Constraint 1", "Constraint 2").
		WithErrorGuidance(map[string]string{
			"error1": "How to fix error 1",
		})

	metadata := ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        tool.Name(),
			Category:    tool.Category(),
			Tags:        tool.Tags(),
			Description: tool.Description(),
			Version:     tool.Version(),
		},
		RequiredPermissions: []string{"test:read"},
		ResourceUsage: ResourceInfo{
			Memory: "low",
		},
	}

	// Register the tool
	if err := registry.RegisterTool(tool.Name(), tool, metadata); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Get documentation
	doc, err := registry.GetToolDocumentation(tool.Name())
	if err != nil {
		t.Errorf("Failed to get tool documentation: %v", err)
	}

	// Verify all fields
	if doc.Name != tool.Name() {
		t.Errorf("Doc name mismatch")
	}
	if doc.UsageInstructions != tool.UsageInstructions() {
		t.Errorf("Doc usage instructions mismatch")
	}
	if len(doc.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(doc.Examples))
	}
	if len(doc.Constraints) != 2 {
		t.Errorf("Expected 2 constraints, got %d", len(doc.Constraints))
	}
	if len(doc.RequiredPermissions) != 1 {
		t.Errorf("Expected 1 permission, got %d", len(doc.RequiredPermissions))
	}
	if !doc.IsDeterministic {
		t.Error("Expected deterministic tool")
	}
}
