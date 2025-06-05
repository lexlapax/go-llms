package benchmarks

// ABOUTME: Benchmarks for agent initialization and tool execution
// ABOUTME: Tests performance of agent setup, context creation, and tool calls

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"

	// Import built-in tools
	builtinTools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"

	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BenchmarkAgentContextInit benchmarks the agent context initialization
func BenchmarkAgentContextInit(b *testing.B) {
	// Create tools
	createTools := func() (domain.Tool, domain.Tool) {
		mathTool := tools.NewTool(
			"multiply",
			"Multiply two numbers",
			func(params struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}) (map[string]interface{}, error) {
				return map[string]interface{}{
					"result": params.A * params.B,
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

		webFetchTool, _ := builtinTools.GetTool("web_fetch")

		return mathTool, webFetchTool
	}

	// Benchmark the agent's context initialization
	b.Run("AgentInitialization", func(b *testing.B) {
		// Create mock provider
		mockProvider := provider.NewMockProvider()

		// Create tools
		mathTool, webFetchTool := createTools()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create agent with tools
			deps := core.LLMDeps{
				Provider: mockProvider,
			}
			agent := core.NewLLMAgent("benchmark-agent", "benchmark", deps)
			agent.SetSystemPrompt("You are a helpful assistant.")
			agent.AddTool(mathTool)
			agent.AddTool(webFetchTool)
		}
	})

	// Benchmark agent execution
	b.Run("AgentExecution", func(b *testing.B) {
		// Create mock provider
		mockProvider := provider.NewMockProvider()

		// Create tools
		mathTool, webFetchTool := createTools()

		// Create agent with tools
		deps := core.LLMDeps{
			Provider: mockProvider,
		}
		agent := core.NewLLMAgent("benchmark-agent", "benchmark", deps)
		agent.SetSystemPrompt("You are a helpful assistant.")
		agent.AddTool(mathTool)
		agent.AddTool(webFetchTool)

		// Test context
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Run will call createInitialMessages internally
			state := domain.NewState()
			state.Set("user_input", "Can you help me with a calculation?")
			_, _ = agent.Run(ctx, state)
		}
	})
}

// BenchmarkAgentSetup benchmarks the agent initialization process
func BenchmarkAgentSetup(b *testing.B) {
	b.Run("AgentWithMultipleTools", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create a mock provider (fast, no network calls)
			mockProvider := provider.NewMockProvider()

			// Create an agent
			deps := core.LLMDeps{
				Provider: mockProvider,
			}
			agent := core.NewLLMAgent("benchmark-setup", "benchmark", deps)
			agent.SetSystemPrompt("You are a helpful assistant that can answer questions and use tools.")

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

// BenchmarkTool benchmarks individual tool operations
func BenchmarkTool(b *testing.B) {
	// Create a simple tool with no parameters
	noParamTool := tools.NewTool(
		"no_param",
		"A tool with no parameters",
		func() string {
			return "Result with no parameters"
		},
		&schemaDomain.Schema{
			Type: "object",
		},
	)

	// Create a tool with string parameter
	stringParamTool := tools.NewTool(
		"string_param",
		"A tool with string parameter",
		func(params struct {
			Input string `json:"input"`
		}) string {
			return "Processed: " + params.Input
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"input": {Type: "string"},
			},
			Required: []string{"input"},
		},
	)

	// Create a tool with struct parameters
	structParamTool := tools.NewTool(
		"struct_param",
		"A tool with struct parameters",
		func(params struct {
			A int     `json:"a"`
			B float64 `json:"b"`
			C string  `json:"c"`
		}) map[string]interface{} {
			return map[string]interface{}{
				"sum":    float64(params.A) + params.B,
				"concat": params.C + " processed",
			}
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"a": {Type: "integer"},
				"b": {Type: "number"},
				"c": {Type: "string"},
			},
			Required: []string{"a", "b", "c"},
		},
	)

	// Create a ToolContext for testing
	toolCtx := &domain.ToolContext{
		Context: context.Background(),
		RunID:   "test-run",
	}

	// Benchmark no-parameter tool
	b.Run("NoParameterTool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = noParamTool.Execute(toolCtx, map[string]interface{}{})
		}
	})

	// Benchmark string parameter tool
	b.Run("StringParameterTool", func(b *testing.B) {
		params := map[string]interface{}{
			"input": "test string",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = stringParamTool.Execute(toolCtx, params)
		}
	})

	// Benchmark struct parameter tool
	b.Run("StructParameterTool", func(b *testing.B) {
		params := map[string]interface{}{
			"a": 42,
			"b": 3.14,
			"c": "test",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = structParamTool.Execute(toolCtx, params)
		}
	})
}