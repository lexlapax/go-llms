package tools

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTool implements domain.Tool for testing
type mockTool struct {
	name              string
	description       string
	category          string
	tags              []string
	version           string
	paramSchema       *sdomain.Schema
	outputSchema      *sdomain.Schema
	usageInstructions string
	examples          []domain.ToolExample
	constraints       []string
	errorGuidance     map[string]string
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return m.description }
func (m *mockTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	return fmt.Sprintf("Executed %s with %v", m.name, params), nil
}
func (m *mockTool) ParameterSchema() *sdomain.Schema { return m.paramSchema }
func (m *mockTool) OutputSchema() *sdomain.Schema    { return m.outputSchema }
func (m *mockTool) UsageInstructions() string        { return m.usageInstructions }
func (m *mockTool) Examples() []domain.ToolExample   { return m.examples }
func (m *mockTool) Constraints() []string            { return m.constraints }
func (m *mockTool) ErrorGuidance() map[string]string { return m.errorGuidance }
func (m *mockTool) Category() string                 { return m.category }
func (m *mockTool) Tags() []string                   { return m.tags }
func (m *mockTool) Version() string                  { return m.version }
func (m *mockTool) IsDeterministic() bool            { return true }
func (m *mockTool) IsDestructive() bool              { return false }
func (m *mockTool) RequiresConfirmation() bool       { return false }
func (m *mockTool) EstimatedLatency() string         { return "fast" }
func (m *mockTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:        m.name,
		Description: m.description,
	}
}

func createTestDiscovery() *toolDiscovery {
	discovery := &toolDiscovery{
		namespaces:       make(map[string]*NamespaceRegistry),
		currentNamespace: "test",
	}
	_ = discovery.CreateNamespace("test")
	return discovery
}

func TestToolDiscovery_ListTools(t *testing.T) {
	discovery := createTestDiscovery()

	// Test empty discovery
	tools := discovery.ListTools()
	assert.Empty(t, tools, "Empty discovery should return no tools")

	// Add some tool metadata
	info1 := ToolInfo{
		Name:        "calculator",
		Description: "Performs calculations",
		Category:    "math",
		Tags:        []string{"arithmetic", "math"},
		Version:     "1.0.0",
	}
	info2 := ToolInfo{
		Name:        "web_search",
		Description: "Searches the web",
		Category:    "web",
		Tags:        []string{"search", "web", "internet"},
		Version:     "2.0.0",
	}

	factory := func() (domain.Tool, error) { return &mockTool{name: "test"}, nil }

	_ = discovery.RegisterTool(info1, factory)
	_ = discovery.RegisterTool(info2, factory)

	// Test listing tools
	tools = discovery.ListTools()
	assert.Len(t, tools, 2, "Should return 2 tools")

	// Verify both tools are present
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}
	assert.True(t, toolNames["calculator"])
	assert.True(t, toolNames["web_search"])
}

func TestToolDiscovery_SearchTools(t *testing.T) {
	discovery := createTestDiscovery()

	// Register test tools
	tools := []ToolInfo{
		{
			Name:        "calculator",
			Description: "Performs mathematical calculations",
			Category:    "math",
			Tags:        []string{"arithmetic", "math", "compute"},
		},
		{
			Name:        "json_process",
			Description: "Process and query JSON data",
			Category:    "data",
			Tags:        []string{"json", "data", "parse"},
		},
		{
			Name:        "file_read",
			Description: "Read files from filesystem",
			Category:    "file",
			Tags:        []string{"file", "read", "io"},
		},
	}

	factory := func() (domain.Tool, error) { return &mockTool{name: "test"}, nil }
	for _, tool := range tools {
		_ = discovery.RegisterTool(tool, factory)
	}

	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "Search by name",
			query:    "calc",
			expected: []string{"calculator"},
		},
		{
			name:     "Search by description",
			query:    "JSON",
			expected: []string{"json_process"},
		},
		{
			name:     "Search by tag",
			query:    "data",
			expected: []string{"json_process"},
		},
		{
			name:     "Search multiple matches",
			query:    "file",
			expected: []string{"file_read"},
		},
		{
			name:     "Case insensitive search",
			query:    "MATH",
			expected: []string{"calculator"},
		},
		{
			name:     "No matches",
			query:    "nonexistent",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := discovery.SearchTools(tt.query)
			assert.Len(t, results, len(tt.expected))

			resultNames := make([]string, len(results))
			for i, result := range results {
				resultNames[i] = result.Name
			}

			for _, expected := range tt.expected {
				assert.Contains(t, resultNames, expected)
			}
		})
	}
}

func TestToolDiscovery_ListByCategory(t *testing.T) {
	discovery := createTestDiscovery()

	// Register test tools
	tools := []ToolInfo{
		{Name: "calculator", Category: "math"},
		{Name: "json_process", Category: "data"},
		{Name: "file_read", Category: "file"},
		{Name: "csv_process", Category: "data"},
	}

	factory := func() (domain.Tool, error) { return &mockTool{name: "test"}, nil }
	for _, tool := range tools {
		_ = discovery.RegisterTool(tool, factory)
	}

	// Test listing by category
	mathTools := discovery.ListByCategory("math")
	assert.Len(t, mathTools, 1)
	assert.Equal(t, "calculator", mathTools[0].Name)

	dataTools := discovery.ListByCategory("data")
	assert.Len(t, dataTools, 2)
	toolNames := []string{dataTools[0].Name, dataTools[1].Name}
	assert.Contains(t, toolNames, "json_process")
	assert.Contains(t, toolNames, "csv_process")

	nonexistentTools := discovery.ListByCategory("nonexistent")
	assert.Empty(t, nonexistentTools)
}

func TestToolDiscovery_GetToolSchema(t *testing.T) {
	discovery := createTestDiscovery()

	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"expression": {Type: "string"},
		},
	}

	outputSchema := &sdomain.Schema{
		Type: "number",
	}

	info := ToolInfo{
		Name:        "calculator",
		Description: "A simple calculator",
		Category:    "math",
	}

	// Marshal schemas to JSON for ToolInfo
	paramBytes, _ := json.Marshal(paramSchema)
	outputBytes, _ := json.Marshal(outputSchema)
	info.ParameterSchema = paramBytes
	info.OutputSchema = outputBytes

	factory := func() (domain.Tool, error) { return &mockTool{name: "calculator"}, nil }
	_ = discovery.RegisterTool(info, factory)

	// Test getting schema
	schema, err := discovery.GetToolSchema("calculator")
	require.NoError(t, err)
	assert.Equal(t, "calculator", schema.Name)
	assert.Equal(t, "A simple calculator", schema.Description)
	assert.NotNil(t, schema.Parameters)
	assert.NotNil(t, schema.Output)

	// Test nonexistent tool
	_, err = discovery.GetToolSchema("nonexistent")
	assert.Error(t, err)
}

func TestToolDiscovery_CreateTool(t *testing.T) {
	discovery := createTestDiscovery()

	// Test creating nonexistent tool
	_, err := discovery.CreateTool("nonexistent")
	assert.Error(t, err)

	// Register a tool
	info := ToolInfo{
		Name:        "calculator",
		Description: "A simple calculator",
	}

	factory := func() (domain.Tool, error) {
		return &mockTool{
			name:        "calculator",
			description: "A simple calculator",
		}, nil
	}

	_ = discovery.RegisterTool(info, factory)

	// Test creating existing tool
	tool, err := discovery.CreateTool("calculator")
	require.NoError(t, err)
	assert.Equal(t, "calculator", tool.Name())
	assert.Equal(t, "A simple calculator", tool.Description())
}

func TestToolDiscovery_CreateTools(t *testing.T) {
	discovery := createTestDiscovery()

	// Register multiple tools
	tools := []string{"calculator", "file_reader"}
	factory := func() (domain.Tool, error) { return &mockTool{name: "test"}, nil }

	for _, toolName := range tools {
		info := ToolInfo{Name: toolName}
		_ = discovery.RegisterTool(info, factory)
	}

	// Test creating multiple tools
	createdTools, err := discovery.CreateTools("calculator", "file_reader")
	require.NoError(t, err)
	assert.Len(t, createdTools, 2)
	assert.Contains(t, createdTools, "calculator")
	assert.Contains(t, createdTools, "file_reader")

	// Test creating with nonexistent tool
	_, err = discovery.CreateTools("calculator", "nonexistent")
	assert.Error(t, err)
}

func TestToolDiscovery_GetToolHelp(t *testing.T) {
	discovery := createTestDiscovery()

	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"expression": {
				Type:        "string",
				Description: "Mathematical expression to evaluate",
			},
		},
		Required: []string{"expression"},
	}

	info := ToolInfo{
		Name:        "calculator",
		Description: "Evaluates mathematical expressions",
		Examples: []Example{
			{
				Name:        "simple_addition",
				Description: "Add two numbers",
				Input:       json.RawMessage(`{"expression": "2 + 3"}`),
				Output:      json.RawMessage(`5`),
			},
		},
	}

	// Marshal schema to JSON for ToolInfo
	paramBytes, _ := json.Marshal(paramSchema)
	info.ParameterSchema = paramBytes

	factory := func() (domain.Tool, error) { return &mockTool{name: "calculator"}, nil }
	_ = discovery.RegisterTool(info, factory)

	// Test getting help
	help, err := discovery.GetToolHelp("calculator")
	require.NoError(t, err)
	assert.Contains(t, help, "calculator")
	assert.Contains(t, help, "Evaluates mathematical expressions")
	assert.Contains(t, help, "Parameters")
	assert.Contains(t, help, "Examples")

	// Test nonexistent tool
	_, err = discovery.GetToolHelp("nonexistent")
	assert.Error(t, err)
}

func TestGetToolMetadata(t *testing.T) {
	// This test uses the global discovery instance
	metadata := GetToolMetadata()
	assert.NotNil(t, metadata)
	// The exact content depends on what's registered globally
}
