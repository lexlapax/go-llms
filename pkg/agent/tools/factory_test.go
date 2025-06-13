package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolFactory_WithoutImports(t *testing.T) {
	// Test that factories exist but return errors without imports
	discovery := NewDiscovery()

	// Try to create a tool without importing the package
	tool, err := discovery.CreateTool("calculator")
	assert.Error(t, err)
	assert.Nil(t, tool)
	assert.Contains(t, err.Error(), "not yet loaded")
}

func TestToolFactory_Metadata(t *testing.T) {
	// Test that metadata is available without imports
	discovery := NewDiscovery()

	// Get tool metadata
	schema, err := discovery.GetToolSchema("calculator")
	require.NoError(t, err)
	assert.Equal(t, "calculator", schema.Name)
	assert.NotEmpty(t, schema.Description)

	// List tools
	tools := discovery.ListTools()
	assert.Greater(t, len(tools), 30)

	// Find calculator
	var found bool
	for _, tool := range tools {
		if tool.Name == "calculator" {
			found = true
			assert.Equal(t, "math", tool.Category)
			assert.Contains(t, tool.Tags, "math")
			break
		}
	}
	assert.True(t, found, "calculator tool should be in the list")
}
