package benchmarks

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BenchmarkAgentContextInit benchmarks the agent context initialization
func BenchmarkAgentContextInit(b *testing.B) {
	// Create a hook to capture message creation
	var systemMessages []string
	messageHook := &testHook{
		beforeGenerateFunc: func(ctx context.Context, messages []ldomain.Message) {
			if len(messages) > 0 && messages[0].Role == ldomain.RoleSystem {
				systemMessages = append(systemMessages, messages[0].Content)
			}
		},
	}

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

		webFetchTool := tools.WebFetch()

		return mathTool, webFetchTool
	}

	// First benchmark the optimized agent's context initialization
	b.Run("OptimizedInitialMessages", func(b *testing.B) {
		// Create mock provider
		mockProvider := provider.NewMockProvider()

		// Create tools
		mathTool, webFetchTool := createTools()

		// Create agent with tools
		agent := workflow.NewAgent(mockProvider)
		agent.SetSystemPrompt("You are a helpful assistant.")
		agent.AddTool(mathTool)
		agent.AddTool(webFetchTool)
		agent.WithHook(messageHook)

		// Test context
		ctx := context.Background()
		input := "Can you help me with a calculation?"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Clear previous messages
			systemMessages = nil
			// Run will call createInitialMessages internally
			_, _ = agent.Run(ctx, input)
		}
	})

	// Then benchmark the unoptimized agent's context initialization
	b.Run("UnoptimizedInitialMessages", func(b *testing.B) {
		// Create mock provider
		mockProvider := provider.NewMockProvider()

		// Create tools
		mathTool, webFetchTool := createTools()

		// Create unoptimized agent with tools
		agent := workflow.NewUnoptimizedAgent(mockProvider)
		agent.SetSystemPrompt("You are a helpful assistant.")
		agent.AddTool(mathTool)
		agent.AddTool(webFetchTool)
		agent.WithHook(messageHook)

		// Test context
		ctx := context.Background()
		input := "Can you help me with a calculation?"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Clear previous messages
			systemMessages = nil
			// Run will call createInitialMessages internally
			_, _ = agent.Run(ctx, input)
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

// BenchmarkAgentToolExtraction benchmarks the tool call extraction methods
func BenchmarkAgentToolExtraction(b *testing.B) {
	// Create sample providers for both agent types
	mockProvider := provider.NewMockProvider()
	optimizedAgent := workflow.NewAgent(mockProvider)
	unoptimizedAgent := workflow.NewUnoptimizedAgent(mockProvider)

	// Test cases with different formats of tool calls
	testCases := []struct {
		name    string
		content string
	}{
		{
			name: "OpenAIFormat",
			content: `{
				"tool_calls": [
					{
						"id": "call_123",
						"type": "function",
						"function": {
							"name": "get_weather",
							"arguments": "{\"location\":\"San Francisco\",\"unit\":\"celsius\"}"
						}
					}
				]
			}`,
		},
		{
			name:    "SimpleJSONFormat",
			content: `{"tool": "search", "params": {"query": "best restaurants in New York"}}`,
		},
		{
			name:    "MarkdownCodeBlock",
			content: "I'll use the search tool to find information for you.\n\n```json\n{\"tool\": \"search\", \"params\": {\"query\": \"golang performance optimization\"}}\n```",
		},
		{
			name: "MultipleToolCalls",
			content: `{
				"tool_calls": [
					{
						"id": "call_123",
						"type": "function",
						"function": {
							"name": "get_weather",
							"arguments": "{\"location\":\"San Francisco\",\"unit\":\"celsius\"}"
						}
					},
					{
						"id": "call_124",
						"type": "function",
						"function": {
							"name": "get_time",
							"arguments": "{\"timezone\":\"PST\"}"
						}
					}
				]
			}`,
		},
		{
			name:    "TextFormat",
			content: "Tool: search\nParams: {\"query\": \"golang performance tips\"}",
		},
	}

	// Benchmark the optimized extractToolCall method
	b.Run("OptimizedExtractToolCall", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					toolName, params, shouldCall := optimizedAgent.ExtractToolCall(tc.content)
					// Use the results to prevent the compiler from optimizing the call away
					if shouldCall && toolName == "" && params == nil {
						b.Fatal("unexpected benchmark result")
					}
				}
			})
		}
	})

	// Benchmark the unoptimized extractToolCall method
	b.Run("UnoptimizedExtractToolCall", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					toolName, params, shouldCall := unoptimizedAgent.ExtractToolCall(tc.content)
					// Use the results to prevent the compiler from optimizing the call away
					if shouldCall && toolName == "" && params == nil {
						b.Fatal("unexpected benchmark result")
					}
				}
			})
		}
	})

	// Only benchmark the multiple tool calls for the relevant test cases
	multiToolTestCases := []struct {
		name    string
		content string
	}{
		{
			name: "OpenAIMultipleTools",
			content: `{
				"tool_calls": [
					{
						"id": "call_123",
						"type": "function",
						"function": {
							"name": "get_weather",
							"arguments": "{\"location\":\"San Francisco\",\"unit\":\"celsius\"}"
						}
					},
					{
						"id": "call_124",
						"type": "function",
						"function": {
							"name": "get_time",
							"arguments": "{\"timezone\":\"PST\"}"
						}
					}
				]
			}`,
		},
		{
			name:    "MarkdownMultipleTools",
			content: "I'll use multiple tools to get information for you.\n\n```json\n{\"tool_calls\": [{\"function\": {\"name\": \"get_weather\", \"arguments\": \"{\\\"location\\\":\\\"San Francisco\\\",\\\"unit\\\":\\\"celsius\\\"}\"}}]}\n```",
		},
	}

	// Benchmark the optimized extractMultipleToolCalls method
	b.Run("OptimizedExtractMultipleToolCalls", func(b *testing.B) {
		for _, tc := range multiToolTestCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					toolNames, paramsArray, shouldCall := optimizedAgent.ExtractMultipleToolCalls(tc.content)
					// Use the results to prevent the compiler from optimizing the call away
					if shouldCall && len(toolNames) == 0 && len(paramsArray) == 0 {
						b.Fatal("unexpected benchmark result")
					}
				}
			})
		}
	})

	// Benchmark the unoptimized extractMultipleToolCalls method
	b.Run("UnoptimizedExtractMultipleToolCalls", func(b *testing.B) {
		for _, tc := range multiToolTestCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					toolNames, paramsArray, shouldCall := unoptimizedAgent.ExtractMultipleToolCalls(tc.content)
					// Use the results to prevent the compiler from optimizing the call away
					if shouldCall && len(toolNames) == 0 && len(paramsArray) == 0 {
						b.Fatal("unexpected benchmark result")
					}
				}
			})
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

// testHook is a helper struct for testing hooks in agents
type testHook struct {
	beforeGenerateFunc func(ctx context.Context, messages []ldomain.Message)
	afterGenerateFunc  func(ctx context.Context, response ldomain.Response, err error)
	beforeToolCallFunc func(ctx context.Context, tool string, params map[string]interface{})
	afterToolCallFunc  func(ctx context.Context, tool string, result interface{}, err error)
}

func (h *testHook) BeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	if h.beforeGenerateFunc != nil {
		h.beforeGenerateFunc(ctx, messages)
	}
}

func (h *testHook) AfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	if h.afterGenerateFunc != nil {
		h.afterGenerateFunc(ctx, response, err)
	}
}

func (h *testHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	if h.beforeToolCallFunc != nil {
		h.beforeToolCallFunc(ctx, tool, params)
	}
}

func (h *testHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	if h.afterToolCallFunc != nil {
		h.afterToolCallFunc(ctx, tool, result, err)
	}
}
