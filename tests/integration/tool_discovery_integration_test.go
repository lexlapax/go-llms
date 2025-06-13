package integration

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryIntegration_ListTools(t *testing.T) {
	discovery := tools.NewDiscovery()

	// List all tools
	allTools := discovery.ListTools()

	// We should have at least 30+ tools from the metadata
	assert.Greater(t, len(allTools), 30)

	// Check some known tools exist
	toolNames := make(map[string]bool)
	for _, tool := range allTools {
		toolNames[tool.Name] = true
	}

	assert.True(t, toolNames["calculator"])
	assert.True(t, toolNames["web_search"])
	assert.True(t, toolNames["file_read"])
	assert.True(t, toolNames["datetime_now"])
}

func TestDiscoveryIntegration_SearchTools(t *testing.T) {
	discovery := tools.NewDiscovery()

	// Search for datetime tools
	dateTools := discovery.SearchTools("datetime")
	assert.Greater(t, len(dateTools), 5)

	// Search for file tools
	fileTools := discovery.SearchTools("file")
	assert.Greater(t, len(fileTools), 4)

	// Search for JSON tools
	jsonTools := discovery.SearchTools("json")
	assert.Greater(t, len(jsonTools), 2)
}

func TestDiscoveryIntegration_ListByCategory(t *testing.T) {
	discovery := tools.NewDiscovery()

	tests := []struct {
		category      string
		expectedCount int
	}{
		{"datetime", 7},
		{"web", 5},
		{"file", 6},
		{"data", 4},
		{"math", 1},
		{"system", 4},
		{"feed", 6},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			tools := discovery.ListByCategory(tt.category)
			assert.Equal(t, tt.expectedCount, len(tools), "Category %s should have %d tools", tt.category, tt.expectedCount)
		})
	}
}

func TestDiscoveryIntegration_GetToolSchema(t *testing.T) {
	discovery := tools.NewDiscovery()

	// Get schema for calculator
	schema, err := discovery.GetToolSchema("calculator")
	require.NoError(t, err)
	assert.Equal(t, "calculator", schema.Name)
	assert.Contains(t, schema.Description, "mathematical calculations")

	// Get schema for non-existent tool
	_, err = discovery.GetToolSchema("nonexistent")
	assert.Error(t, err)
}

func TestDiscoveryIntegration_GetToolHelp(t *testing.T) {
	discovery := tools.NewDiscovery()

	// Get help for datetime_now
	help, err := discovery.GetToolHelp("datetime_now")
	require.NoError(t, err)

	// Verify help contains expected sections
	assert.Contains(t, help, "Tool: datetime_now")
	assert.Contains(t, help, "Description:")
	assert.Contains(t, help, "Parameters:")
}

func TestDiscoveryIntegration_GetToolExamples(t *testing.T) {
	discovery := tools.NewDiscovery()

	// Test with a tool that has examples (calculator usually has examples)
	examples, err := discovery.GetToolExamples("calculator")
	require.NoError(t, err)

	// Calculator should have at least one example
	assert.NotEmpty(t, examples, "calculator should have examples")
	t.Logf("Found %d examples for calculator", len(examples))

	// Verify example structure
	for i, ex := range examples {
		t.Logf("Example %d: %s - %s", i, ex.Name, ex.Description)
		assert.NotEmpty(t, ex.Name, "example should have a name")
		assert.NotEmpty(t, ex.Description, "example should have a description")
		// Note: Input might be nil for some examples in the metadata
		// The actual tool examples have Input, but the metadata might not preserve it
	}

	// Test with non-existent tool
	_, err = discovery.GetToolExamples("nonexistent_tool")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDiscoveryIntegration_CreateTool(t *testing.T) {
	t.Skip("Tool creation requires actual tool packages to be imported")

	// This test would work when the actual tool packages are imported
	// For now, we skip it as it would require build tags
}
