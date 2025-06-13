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

func TestToolDiscovery_ListTools(t *testing.T) {
	// Create a fresh discovery instance for testing
	discovery := &toolDiscovery{
		metadata:  make(map[string]ToolInfo),
		factories: make(map[string]ToolFactory),
	}

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

	discovery.metadata["calculator"] = info1
	discovery.metadata["web_search"] = info2

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
	discovery := &toolDiscovery{
		metadata: map[string]ToolInfo{
			"calculator": {
				Name:        "calculator",
				Description: "Performs mathematical calculations",
				Category:    "math",
				Tags:        []string{"arithmetic", "math", "compute"},
			},
			"json_process": {
				Name:        "json_process",
				Description: "Process and query JSON data",
				Category:    "data",
				Tags:        []string{"json", "data", "parse"},
			},
			"file_read": {
				Name:        "file_read",
				Description: "Read files from filesystem",
				Category:    "file",
				Tags:        []string{"file", "read", "io"},
			},
		},
		factories: make(map[string]ToolFactory),
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
			resultNames := make([]string, len(results))
			for i, r := range results {
				resultNames[i] = r.Name
			}

			assert.Len(t, resultNames, len(tt.expected))
			for _, expectedName := range tt.expected {
				assert.Contains(t, resultNames, expectedName)
			}
		})
	}
}

func TestToolDiscovery_ListByCategory(t *testing.T) {
	discovery := &toolDiscovery{
		metadata: map[string]ToolInfo{
			"calculator": {
				Name:     "calculator",
				Category: "math",
			},
			"statistics": {
				Name:     "statistics",
				Category: "math",
			},
			"web_search": {
				Name:     "web_search",
				Category: "web",
			},
		},
		factories: make(map[string]ToolFactory),
	}

	// Test math category
	mathTools := discovery.ListByCategory("math")
	assert.Len(t, mathTools, 2)

	// Test web category
	webTools := discovery.ListByCategory("web")
	assert.Len(t, webTools, 1)
	assert.Equal(t, "web_search", webTools[0].Name)

	// Test non-existent category
	emptyTools := discovery.ListByCategory("nonexistent")
	assert.Empty(t, emptyTools)
}

func TestToolDiscovery_GetToolSchema(t *testing.T) {
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"expression": {
				Type:        "string",
				Description: "Math expression",
			},
		},
		Required: []string{"expression"},
	}

	outputSchema := &sdomain.Schema{
		Type: "number",
	}

	// Convert schemas to JSON
	paramJSON, _ := json.Marshal(paramSchema)
	outputJSON, _ := json.Marshal(outputSchema)

	discovery := &toolDiscovery{
		metadata: map[string]ToolInfo{
			"calculator": {
				Name:            "calculator",
				Description:     "Performs calculations",
				ParameterSchema: paramJSON,
				OutputSchema:    outputJSON,
				Examples: []Example{
					{
						Name:        "Addition",
						Description: "Add two numbers",
						Input:       json.RawMessage(`{"expression": "2 + 2"}`),
						Output:      json.RawMessage(`4`),
					},
				},
			},
		},
		factories: make(map[string]ToolFactory),
	}

	// Test getting schema for existing tool
	schema, err := discovery.GetToolSchema("calculator")
	require.NoError(t, err)
	assert.Equal(t, "calculator", schema.Name)
	assert.Equal(t, "Performs calculations", schema.Description)
	assert.NotNil(t, schema.Parameters)
	assert.NotNil(t, schema.Output)
	assert.Len(t, schema.Examples, 1)

	// Test getting schema for non-existent tool
	_, err = discovery.GetToolSchema("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestToolDiscovery_CreateTool(t *testing.T) {
	mockCalc := &mockTool{
		name:        "calculator",
		description: "Test calculator",
	}

	discovery := &toolDiscovery{
		metadata: map[string]ToolInfo{
			"calculator": {
				Name: "calculator",
			},
		},
		factories: map[string]ToolFactory{
			"calculator": func() (domain.Tool, error) {
				return mockCalc, nil
			},
		},
	}

	// Test creating existing tool
	tool, err := discovery.CreateTool("calculator")
	require.NoError(t, err)
	assert.Equal(t, "calculator", tool.Name())
	assert.Equal(t, "Test calculator", tool.Description())

	// Test creating non-existent tool
	_, err = discovery.CreateTool("nonexistent")
	assert.Error(t, err)
}

func TestToolDiscovery_CreateTools(t *testing.T) {
	mockCalc := &mockTool{name: "calculator"}
	mockWeb := &mockTool{name: "web_search"}

	discovery := &toolDiscovery{
		metadata: map[string]ToolInfo{
			"calculator": {Name: "calculator"},
			"web_search": {Name: "web_search"},
		},
		factories: map[string]ToolFactory{
			"calculator": func() (domain.Tool, error) { return mockCalc, nil },
			"web_search": func() (domain.Tool, error) { return mockWeb, nil },
		},
	}

	// Test creating multiple tools
	tools, err := discovery.CreateTools("calculator", "web_search")
	require.NoError(t, err)
	assert.Len(t, tools, 2)
	assert.NotNil(t, tools["calculator"])
	assert.NotNil(t, tools["web_search"])

	// Test with non-existent tool
	_, err = discovery.CreateTools("calculator", "nonexistent")
	assert.Error(t, err)
}

func TestToolDiscovery_GetToolExamples(t *testing.T) {
	discovery := &toolDiscovery{
		metadata: map[string]ToolInfo{
			"calculator": {
				Name:        "calculator",
				Description: "Performs mathematical calculations",
				Examples: []Example{
					{
						Name:        "Basic addition",
						Description: "Add two numbers",
						Input:       json.RawMessage(`{"expression": "2 + 2"}`),
						Output:      json.RawMessage(`{"result": 4}`),
					},
					{
						Name:        "Complex calculation",
						Description: "Evaluate complex expression",
						Input:       json.RawMessage(`{"expression": "(10 + 5) * 2"}`),
						Output:      json.RawMessage(`{"result": 30}`),
					},
				},
			},
		},
		factories: make(map[string]ToolFactory),
	}

	// Test successful retrieval
	examples, err := discovery.GetToolExamples("calculator")
	require.NoError(t, err)
	assert.Len(t, examples, 2)

	// Verify first example
	assert.Equal(t, "Basic addition", examples[0].Name)
	assert.Equal(t, "Add two numbers", examples[0].Description)
	assert.Equal(t, map[string]interface{}{"expression": "2 + 2"}, examples[0].Input)
	assert.Equal(t, map[string]interface{}{"result": float64(4)}, examples[0].Output)

	// Verify second example
	assert.Equal(t, "Complex calculation", examples[1].Name)
	assert.Equal(t, "Evaluate complex expression", examples[1].Description)

	// Test non-existent tool
	_, err = discovery.GetToolExamples("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool nonexistent not found")
}

func TestToolDiscovery_GetToolHelp(t *testing.T) {
	paramSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "Math expression to evaluate",
			},
		},
		"required": []string{"expression"},
	}

	paramJSON, _ := json.Marshal(paramSchema)

	discovery := &toolDiscovery{
		metadata: map[string]ToolInfo{
			"calculator": {
				Name:            "calculator",
				Description:     "Performs mathematical calculations",
				ParameterSchema: paramJSON,
				Examples: []Example{
					{
						Name:        "Basic addition",
						Description: "Add two numbers",
						Input:       json.RawMessage(`{"expression": "2 + 2"}`),
					},
				},
			},
		},
		factories: make(map[string]ToolFactory),
	}

	help, err := discovery.GetToolHelp("calculator")
	require.NoError(t, err)

	// Verify help contains expected sections
	assert.Contains(t, help, "Tool: calculator")
	assert.Contains(t, help, "Description: Performs mathematical calculations")
	assert.Contains(t, help, "Parameters:")
	assert.Contains(t, help, "expression")
	assert.Contains(t, help, "Examples:")
	assert.Contains(t, help, "Basic addition")
}

func TestRegisterToolMetadata(t *testing.T) {
	// This test modifies global state, so we skip it to avoid interfering with other tests
	// The functionality is already tested through the generated registry_metadata.go
	t.Skip("Skipping test that modifies global discovery state")
}

func TestGetToolMetadata(t *testing.T) {
	// Test the global GetToolMetadata function
	metadata := GetToolMetadata()

	// Should have all tools
	assert.Greater(t, len(metadata), 30)

	// Check specific tool
	calc, exists := metadata["calculator"]
	assert.True(t, exists)
	assert.Equal(t, "calculator", calc.Name)
	assert.Equal(t, "math", calc.Category)
	assert.NotEmpty(t, calc.Description)
	assert.Contains(t, calc.Tags, "math")

	// Verify it returns a copy, not the internal map
	delete(metadata, "calculator")
	metadata2 := GetToolMetadata()
	_, exists = metadata2["calculator"]
	assert.True(t, exists, "deleting from returned map should not affect internal state")
}

func TestToolInfo_JSONMarshaling(t *testing.T) {
	info := ToolInfo{
		Name:            "test_tool",
		Description:     "A test tool",
		Category:        "test",
		Tags:            []string{"test", "example"},
		Version:         "1.0.0",
		ParameterSchema: json.RawMessage(`{"type": "object"}`),
		Examples: []Example{
			{
				Name:  "Test",
				Input: json.RawMessage(`{"test": true}`),
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(info)
	require.NoError(t, err)

	// Test unmarshaling
	var decoded ToolInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info.Name, decoded.Name)
	assert.Equal(t, info.Description, decoded.Description)
	assert.Equal(t, info.Category, decoded.Category)
	assert.Equal(t, info.Tags, decoded.Tags)
	assert.Equal(t, info.Version, decoded.Version)
}
