package benchmarks

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BenchmarkAgentSetup benchmarks the agent initialization process
func BenchmarkAgentSetup(b *testing.B) {
	b.Run("AgentWithMultipleTools", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create a mock provider (fast, no network calls)
			mockProvider := provider.NewMockProvider()

			// Create an agent
			agent := workflow.NewAgent(mockProvider)

			// Set system prompt
			agent.SetSystemPrompt("You are a helpful assistant that can answer questions and use tools.")

			// Add hooks for monitoring
			metricsHook := workflow.NewMetricsHook()
			agent.WithHook(metricsHook)

			// Add multiple tools
			agent.AddTool(tools.NewTool(
				"get_current_date",
				"Get the current date",
				func() map[string]string {
					return map[string]string{
						"date": "2025-05-06",
						"time": "12:30:00",
						"year": "2025",
					}
				},
				&schemaDomain.Schema{
					Type:        "object",
					Description: "Returns the current date and time",
				},
			))

			agent.AddTool(tools.NewTool(
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
			))

			agent.AddTool(tools.NewTool(
				"search",
				"Search for information",
				func(params struct {
					Query string `json:"query"`
				}) (string, error) {
					return "Search results for: " + params.Query, nil
				},
				&schemaDomain.Schema{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"query": {
							Type:        "string",
							Description: "The search query",
						},
					},
					Required: []string{"query"},
				},
			))
		}
	})
}

// BenchmarkToolExecution benchmarks the tool execution process
func BenchmarkToolExecution(b *testing.B) {
	// Create mock tools with different parameter types
	noParamTool := tools.NewTool(
		"no_param_tool",
		"A tool with no parameters",
		func() string {
			return "result"
		},
		nil,
	)

	stringParamTool := tools.NewTool(
		"string_param_tool",
		"A tool with string parameter",
		func(p string) string {
			return "processed: " + p
		},
		&schemaDomain.Schema{
			Type: "string",
		},
	)

	structParamTool := tools.NewTool(
		"struct_param_tool",
		"A tool with struct parameter",
		func(params struct {
			Name string  `json:"name"`
			Age  int     `json:"age"`
			Rate float64 `json:"rate"`
		}) map[string]interface{} {
			return map[string]interface{}{
				"processed_name": params.Name,
				"processed_age":  params.Age * 2,
				"processed_rate": params.Rate * 1.5,
			}
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
				"rate": {Type: "number"},
			},
			Required: []string{"name", "age"},
		},
	)

	// Create test context
	ctx := context.Background()

	// Benchmark no-parameter tool
	b.Run("NoParamTool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := noParamTool.Execute(ctx, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark string parameter tool
	b.Run("StringParamTool", func(b *testing.B) {
		param := "test parameter"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := stringParamTool.Execute(ctx, param)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark struct parameter tool
	b.Run("StructParamTool", func(b *testing.B) {
		param := map[string]interface{}{
			"name": "John",
			"age":  30,
			"rate": 15.75,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := structParamTool.Execute(ctx, param)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}