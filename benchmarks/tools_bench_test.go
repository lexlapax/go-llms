package benchmarks

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BenchmarkToolCreation benchmarks tool creation performance
func BenchmarkToolCreation(b *testing.B) {
	// Simple function for tools
	simpleFunc := func(params struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}) (map[string]interface{}, error) {
		result := params.A * params.B
		return map[string]interface{}{
			"result":      result,
			"calculation": params.A,
			"a":           params.A,
			"b":           params.B,
		}, nil
	}

	// Parameter schema
	paramSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"a": {
				Type:        "number",
				Description: "The first number",
			},
			"b": {
				Type:        "number",
				Description: "The second number",
			},
		},
		Required: []string{"a", "b"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tools.NewTool(
			"multiply",
			"Multiply two numbers",
			simpleFunc,
			paramSchema,
		)
	}
}

// BenchmarkToolParameterHandling benchmarks tool parameter handling performance
func BenchmarkToolParameterHandling(b *testing.B) {
	// Prepare test data
	ctx := context.Background()
	params := map[string]interface{}{
		"a": 10.5,
		"b": 20.5,
	}

	// Create tool
	tool := tools.NewTool(
		"multiply",
		"Multiply two numbers",
		func(params struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}) (map[string]interface{}, error) {
			result := params.A * params.B
			return map[string]interface{}{
				"result":      result,
				"calculation": params.A,
				"a":           params.A,
				"b":           params.B,
			}, nil
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"a": {
					Type:        "number",
					Description: "The first number",
				},
				"b": {
					Type:        "number",
					Description: "The second number",
				},
			},
			Required: []string{"a", "b"},
		},
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkToolConversionTypes benchmarks tool type conversion performance
func BenchmarkToolConversionTypes(b *testing.B) {
	ctx := context.Background()

	// Benchmark conversions with different parameter types
	testCases := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name: "IntToFloat",
			params: map[string]interface{}{
				"a": 10,
				"b": 20,
			},
		},
		{
			name: "StringToFloat",
			params: map[string]interface{}{
				"a": "10.5",
				"b": "20.5",
			},
		},
		{
			name: "MixedTypes",
			params: map[string]interface{}{
				"a": 10.5,
				"b": "20.5",
			},
		},
	}

	for _, tc := range testCases {
		// Create tool with the same function for each test case
		tool := tools.NewTool(
			"multiply",
			"Multiply two numbers",
			func(params struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}) (map[string]interface{}, error) {
				result := params.A * params.B
				return map[string]interface{}{
					"result": result,
					"a":      params.A,
					"b":      params.B,
				}, nil
			},
			&schemaDomain.Schema{
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"a": {Type: "number"},
					"b": {Type: "number"},
				},
				Required: []string{"a", "b"},
			},
		)

		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := tool.Execute(ctx, tc.params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCommonTools benchmarks common tool implementations
func BenchmarkCommonTools(b *testing.B) {
	ctx := context.Background()

	// Test cases for different tools
	testCases := []struct {
		name   string
		tool   domain.Tool
		params interface{}
	}{
		{
			name:   "ReadFile",
			tool:   tools.ReadFile(),
			params: tools.ReadFileParams{Path: "README.md"},
		},
		{
			name:   "ExecuteCommand",
			tool:   tools.ExecuteCommand(),
			params: tools.ExecuteCommandParams{Command: "echo 'hello world'", Timeout: 1.0},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := tc.tool.Execute(ctx, tc.params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}