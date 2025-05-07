package benchmarks

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BenchmarkToolCreation benchmarks the creation of tools
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

	// Benchmark original tool creation
	b.Run("OriginalToolCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tools.NewTool(
				"multiply",
				"Multiply two numbers",
				simpleFunc,
				paramSchema,
			)
		}
	})

	// Benchmark optimized tool creation
	b.Run("OptimizedToolCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tools.NewOptimizedTool(
				"multiply",
				"Multiply two numbers",
				simpleFunc,
				paramSchema,
			)
		}
	})
}

// BenchmarkToolParameterHandling benchmarks the parameter handling of tools
func BenchmarkToolParameterHandling(b *testing.B) {
	// Prepare test data
	ctx := context.Background()
	params := map[string]interface{}{
		"a": 10.5,
		"b": 20.5,
	}

	// Create tools
	originalTool := tools.NewTool(
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

	optimizedTool := tools.NewOptimizedTool(
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

	// Benchmark original tool execution
	b.Run("OriginalToolExecution", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := originalTool.Execute(ctx, params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized tool execution
	b.Run("OptimizedToolExecution", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedTool.Execute(ctx, params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkToolConversionTypes benchmarks different types of parameter conversions
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
		// Create tools with the same function for each test case
		originalTool := tools.NewTool(
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

		optimizedTool := tools.NewOptimizedTool(
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

		// Benchmark original implementation
		b.Run("Original_"+tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := originalTool.Execute(ctx, tc.params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		// Benchmark optimized implementation
		b.Run("Optimized_"+tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := optimizedTool.Execute(ctx, tc.params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCommonTools benchmarks the common tool implementations
func BenchmarkCommonTools(b *testing.B) {
	ctx := context.Background()

	// Test cases for different tools
	testCases := []struct {
		name        string
		originalFn  func() (domain.Tool, interface{})
		optimizedFn func() (domain.Tool, interface{})
	}{
		{
			name: "ReadFile",
			originalFn: func() (domain.Tool, interface{}) {
				return tools.ReadFile(), tools.ReadFileParams{Path: "README.md"}
			},
			optimizedFn: func() (domain.Tool, interface{}) {
				return tools.OptimizedReadFile(), tools.ReadFileParams{Path: "README.md"}
			},
		},
		{
			name: "ExecuteCommand",
			originalFn: func() (domain.Tool, interface{}) {
				return tools.ExecuteCommand(), tools.ExecuteCommandParams{
					Command: "echo 'hello world'",
					Timeout: 1.0,
				}
			},
			optimizedFn: func() (domain.Tool, interface{}) {
				return tools.OptimizedExecuteCommand(), tools.ExecuteCommandParams{
					Command: "echo 'hello world'",
					Timeout: 1.0,
				}
			},
		},
	}

	for _, tc := range testCases {
		originalTool, originalParams := tc.originalFn()
		optimizedTool, optimizedParams := tc.optimizedFn()

		// Benchmark original implementation
		b.Run("Original_"+tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := originalTool.Execute(ctx, originalParams)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		// Benchmark optimized implementation
		b.Run("Optimized_"+tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := optimizedTool.Execute(ctx, optimizedParams)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}