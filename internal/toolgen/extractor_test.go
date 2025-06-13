package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractToolMetadata(t *testing.T) {
	testCode := `
package datetime

import (
	"context"
	"time"
	
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func init() {
	tools.MustRegisterTool("datetime_now", createDateTimeNowTool(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "datetime_now",
			Category:    "datetime",
			Tags:        []string{"time", "date", "timezone"},
			Description: "Get current date and time in various formats and timezones",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic usage",
					Description: "Get current time in UTC",
					Code:        ` + "`" + `{"timezone": "UTC"}` + "`" + `,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

func createDateTimeNowTool() domain.Tool {
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"timezone": {
				Type:        "string",
				Description: "Timezone name (e.g., 'UTC', 'America/New_York')",
				Default:     "UTC",
			},
			"format": {
				Type:        "string",
				Description: "Output format",
				Enum:        []interface{}{"ISO8601", "RFC3339", "Unix", "UnixMilli"},
				Default:     "ISO8601",
			},
		},
		Required: []string{},
	}

	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"datetime": {
				Type:        "string",
				Description: "Current date and time in requested format",
			},
			"timezone": {
				Type:        "string",
				Description: "Timezone used",
			},
			"unix_timestamp": {
				Type:        "integer",
				Description: "Unix timestamp (seconds since epoch)",
			},
		},
		Required: []string{"datetime", "timezone", "unix_timestamp"},
	}

	return tools.NewToolBuilder("datetime_now", "Get current date and time").
		WithCategory("datetime").
		WithTags("time", "date", "timezone", "now", "current").
		WithVersion("1.0.0").
		WithFunction(dateTimeNowExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions("Use this tool to get the current date and time.").
		WithExamples(
			tools.Example{
				Name:        "Get current UTC time",
				Description: "Retrieve current time in UTC timezone",
				Input:       map[string]interface{}{"timezone": "UTC"},
				Output:      map[string]interface{}{
					"datetime": "2024-01-15T10:30:00Z",
					"timezone": "UTC", 
					"unix_timestamp": 1705318200,
				},
			},
		).
		WithConstraints(
			"Timezone must be a valid IANA timezone name",
			"Unix timestamps are in seconds since January 1, 1970 UTC",
		).
		WithDeterministic(false).
		WithDestructive(false).
		WithRequiresConfirmation(false).
		WithEstimatedLatency("fast").
		Build()
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", testCode, parser.ParseComments)
	require.NoError(t, err)

	extractor := NewExtractor()
	metadata, err := extractor.ExtractFromFile(file)
	require.NoError(t, err)
	require.Len(t, metadata, 1)

	// Check extracted metadata
	tool := metadata[0]
	assert.Equal(t, "datetime_now", tool.Name)
	assert.Equal(t, "Get current date and time in various formats and timezones", tool.Description)
	assert.Equal(t, "datetime", tool.Category)
	assert.Equal(t, []string{"time", "date", "timezone"}, tool.Tags)
	assert.Equal(t, "1.0.0", tool.Version)

	// Check resource usage
	assert.Equal(t, "low", tool.ResourceUsage.Memory)
	assert.False(t, tool.ResourceUsage.Network)
	assert.False(t, tool.ResourceUsage.FileSystem)
	assert.True(t, tool.ResourceUsage.Concurrency)

	// Check builder metadata
	assert.Equal(t, "Use this tool to get the current date and time.", tool.UsageInstructions)
	assert.Len(t, tool.Constraints, 2)
	assert.False(t, tool.IsDeterministic)
	assert.False(t, tool.IsDestructive)
	assert.Equal(t, "fast", tool.EstimatedLatency)
}

func TestExtractParameterSchema(t *testing.T) {
	testCode := `&sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"query": {
			Type:        "string",
			Description: "Search query",
			MinLength:   1,
		},
		"limit": {
			Type:        "integer",
			Description: "Maximum results",
			Minimum:     1,
			Maximum:     100,
			Default:     10,
		},
	},
	Required: []string{"query"},
}
`

	expr, err := parser.ParseExpr(testCode)
	require.NoError(t, err)

	extractor := NewExtractor()
	schemaInterface := extractor.extractSchema(expr)

	assert.NotNil(t, schemaInterface)
	schema, ok := schemaInterface.(map[string]interface{})
	require.True(t, ok, "schema should be a map")
	assert.Equal(t, "object", schema["type"])

	propsInterface, ok := schema["properties"]
	require.True(t, ok, "schema should have properties")
	props, ok := propsInterface.(map[string]interface{})
	require.True(t, ok, "properties should be a map")
	assert.Len(t, props, 2)

	queryPropInterface, ok := props["query"]
	require.True(t, ok, "should have query property")
	queryProp, ok := queryPropInterface.(map[string]interface{})
	require.True(t, ok, "query property should be a map")
	assert.Equal(t, "string", queryProp["type"])
	assert.Equal(t, "Search query", queryProp["description"])
	assert.Equal(t, float64(1), queryProp["minLength"])

	limitPropInterface, ok := props["limit"]
	require.True(t, ok, "should have limit property")
	limitProp, ok := limitPropInterface.(map[string]interface{})
	require.True(t, ok, "limit property should be a map")
	assert.Equal(t, "integer", limitProp["type"])
	assert.Equal(t, float64(10), limitProp["default"])

	requiredInterface, ok := schema["required"]
	require.True(t, ok, "schema should have required")
	required, ok := requiredInterface.([]string)
	require.True(t, ok, "required should be a string slice")
	assert.Equal(t, []string{"query"}, required)
}

/*
func TestExtractFromBuilderCalls(t *testing.T) {
	testCode := `tools.NewToolBuilder("web_search", "Search the web").
	WithCategory("web").
	WithTags("search", "web", "internet").
	WithVersion("2.0.0").
	WithFunction(webSearchExecute).
	WithParameterSchema(paramSchema).
	WithOutputSchema(outputSchema).
	WithUsageInstructions("Use this tool to search the web for information.").
	WithExamples(
		tools.Example{
			Name:        "Basic search",
			Description: "Simple web search",
			Input:       map[string]interface{}{"query": "golang tutorials"},
		},
	).
	WithConstraints(
		"Results may vary based on search engine",
		"Rate limited to 100 requests per minute",
	).
	WithErrorGuidance(map[string]string{
		"rate_limit": "Wait 60 seconds before retrying",
		"no_results": "Try different search terms",
	}).
	WithDeterministic(false).
	WithDestructive(false).
	WithRequiresConfirmation(false).
	WithEstimatedLatency("medium").
	Build()
`

	expr, err := parser.ParseExpr(testCode)
	require.NoError(t, err)

	extractor := NewExtractor()
	builder := extractor.extractBuilderMetadata(expr.(*ast.CallExpr))

	assert.Equal(t, "web_search", builder.Name)
	assert.Equal(t, "Search the web", builder.Description)
	assert.Equal(t, "web", builder.Category)
	assert.Equal(t, []string{"search", "web", "internet"}, builder.Tags)
	assert.Equal(t, "2.0.0", builder.Version)
	assert.Equal(t, "Use this tool to search the web for information.", builder.UsageInstructions)
	assert.Len(t, builder.Constraints, 2)
	assert.Len(t, builder.ErrorGuidance, 2)
	assert.False(t, builder.IsDeterministic)
	assert.Equal(t, "medium", builder.EstimatedLatency)
}
*/

func TestGenerateMetadataFile(t *testing.T) {
	metadata := []ToolMetadata{
		{
			Name:        "calculator",
			Description: "Performs calculations",
			Category:    "math",
			Tags:        []string{"math", "arithmetic"},
			Version:     "1.0.0",
			Package:     "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math",
		},
		{
			Name:        "web_search",
			Description: "Searches the web",
			Category:    "web",
			Tags:        []string{"search", "web"},
			Version:     "2.0.0",
			Package:     "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web",
		},
	}

	generator := NewGenerator()
	code, err := generator.GenerateMetadataFile(metadata)
	require.NoError(t, err)

	// Verify generated code contains expected elements
	assert.Contains(t, code, "package tools")
	assert.Contains(t, code, "// Code generated by toolgen. DO NOT EDIT.")
	assert.Contains(t, code, "var ToolManifest = map[string]ToolInfo{")
	assert.Contains(t, code, `"calculator": {`)
	assert.Contains(t, code, `"web_search": {`)
	assert.Contains(t, code, `Name:        "calculator"`)
	assert.Contains(t, code, `Description: "Performs calculations"`)
	assert.Contains(t, code, `Category:    "math"`)
	assert.Contains(t, code, `Tags:        []string{"math", "arithmetic"}`)

	// Verify factory map and functions
	assert.Contains(t, code, "var toolFactories = map[string]ToolFactory{")
	assert.Contains(t, code, "func createCalculatorFactory() ToolFactory")
	assert.Contains(t, code, "func createWebSearchFactory() ToolFactory")
	assert.Contains(t, code, `"calculator": createCalculatorFactory()`)
	assert.Contains(t, code, `"web_search": createWebSearchFactory()`)
}

func TestParseDirectory(t *testing.T) {
	// This test would require a test directory structure
	// For now, we'll skip it as it requires filesystem setup
	t.Skip("Requires test directory structure")
}

// Helper function tests
func TestExtractStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic string literal",
			input:    `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "String with escapes",
			input:    `"hello \"world\""`,
			expected: `hello "world"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpr(tt.input)
			require.NoError(t, err)

			extractor := NewExtractor()
			result := extractor.extractString(expr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSliceValues(t *testing.T) {
	input := `[]string{"tag1", "tag2", "tag3"}`

	expr, err := parser.ParseExpr(input)
	require.NoError(t, err)

	extractor := NewExtractor()
	result := extractor.extractStringSlice(expr.(*ast.CompositeLit))

	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, result)
}
