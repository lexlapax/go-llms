// ABOUTME: Tests for tool-specific registry functionality
// ABOUTME: Verifies tool metadata validation and resource filtering

package tools

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// mockTool implements domain.Tool for testing
type mockTool struct {
	name        string
	description string
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return m.description }
func (m *mockTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	return "mock result", nil
}
func (m *mockTool) ParameterSchema() *sdomain.Schema { return nil }

func TestToolRegistry_RegisterTool(t *testing.T) {
	// Create a new registry for each test to avoid interference
	registry := &toolRegistry{
		Registry: builtins.NewRegistry[domain.Tool](),
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
			tool:     &mockTool{name: "test_tool", description: "Test tool"},
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
			tool:     &mockTool{name: "different_name", description: "Test tool"},
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
			tool:     &mockTool{name: "test_tool", description: "Test tool"},
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
			tool:     &mockTool{name: "test_tool", description: "Test tool"},
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
		Registry: builtins.NewRegistry[domain.Tool](),
	}

	// Register tools with different permissions (using tags as proxy)
	tool1 := &mockTool{name: "file_reader", description: "Reads files"}
	if err := registry.Register("file_reader", tool1, builtins.Metadata{
		Tags: []string{"file", "perm:file:read"},
	}); err != nil {
		t.Fatalf("failed to register file_reader: %v", err)
	}

	tool2 := &mockTool{name: "file_writer", description: "Writes files"}
	if err := registry.Register("file_writer", tool2, builtins.Metadata{
		Tags: []string{"file", "perm:file:write"},
	}); err != nil {
		t.Fatalf("failed to register file_writer: %v", err)
	}

	tool3 := &mockTool{name: "network_fetch", description: "Fetches from network"}
	if err := registry.Register("network_fetch", tool3, builtins.Metadata{
		Tags: []string{"network", "perm:network:access"},
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
		Registry: builtins.NewRegistry[domain.Tool](),
	}

	// Register tools with different resource characteristics (using tags)
	tool1 := &mockTool{name: "low_memory_tool", description: "Uses low memory"}
	if err := registry.Register("low_memory_tool", tool1, builtins.Metadata{
		Tags: []string{"memory:low", "concurrent"},
	}); err != nil {
		t.Fatalf("failed to register low_memory_tool: %v", err)
	}

	tool2 := &mockTool{name: "high_memory_network", description: "High memory network tool"}
	if err := registry.Register("high_memory_network", tool2, builtins.Metadata{
		Tags: []string{"memory:high", "network", "concurrent"},
	}); err != nil {
		t.Fatalf("failed to register high_memory_network: %v", err)
	}

	tool3 := &mockTool{name: "file_tool", description: "File system tool"}
	if err := registry.Register("file_tool", tool3, builtins.Metadata{
		Tags: []string{"memory:medium", "filesystem"},
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
		Registry: builtins.NewRegistry[domain.Tool](),
	}

	// Test MustRegisterTool
	tool := &mockTool{name: "global_tool", description: "Global tool"}
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
		Registry: builtins.NewRegistry[domain.Tool](),
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
			tool:        &mockTool{name: "test1"},
			permissions: []string{"resource:action", "another:permission"},
			memory:      "medium",
			wantErr:     false,
		},
		{
			name:        "valid memory levels",
			toolName:    "test2",
			tool:        &mockTool{name: "test2"},
			permissions: []string{},
			memory:      "high",
			wantErr:     false,
		},
		{
			name:        "empty memory is valid",
			toolName:    "test3",
			tool:        &mockTool{name: "test3"},
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
