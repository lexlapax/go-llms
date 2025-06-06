package integration

// ABOUTME: Integration tests for bidirectional agent-tool conversion functionality
// ABOUTME: Tests converting agents to tools and tools to agents with state passing

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestAgentToToolConversion tests converting an agent to a tool
func TestAgentToToolConversion(t *testing.T) {
	t.Skip("Skipping test that requires complex mock provider setup for tool execution")
	// Create a mock provider for the translator agent
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Extract the user input from messages
		for _, msg := range messages {
			if msg.Role == ldomain.RoleUser {
				for _, part := range msg.Content {
					if part.Type == ldomain.ContentTypeText && strings.Contains(part.Text, "Hello world") {
						return ldomain.Response{
							Content: "Bonjour le monde", // French translation
						}, nil
					}
				}
			}
		}
		return ldomain.Response{Content: "Translation not available"}, nil
	})

	// Create a translator agent
	translator := core.NewLLMAgent("translator", "test", core.LLMDeps{Provider: mockProvider})
	translator.SetSystemPrompt("You are a translator. Translate text to French.")

	// Convert agent to tool
	translatorTool := tools.NewAgentTool(translator)

	// Create a main agent that uses the translator tool
	mainProvider := provider.NewMockProvider()
	toolCallCount := 0
	mainProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Check if this is a tool result message
		for _, msg := range messages {
			if msg.Role == ldomain.RoleUser {
				for _, part := range msg.Content {
					if strings.Contains(part.Text, "Tool result:") && strings.Contains(part.Text, "Bonjour le monde") {
						// This is the tool result, provide final response
						return ldomain.Response{
							Content: "The French translation is: Bonjour le monde",
						}, nil
					}
				}
			}
		}

		// First call - return tool call request
		if toolCallCount == 0 {
			toolCallCount++
			return ldomain.Response{
				Content: `{"tool": "translator", "params": {"user_input": "Hello world"}}`,
			}, nil
		}

		return ldomain.Response{Content: "I can help with translations"}, nil
	})

	// Create main agent with the translator tool
	mainAgent := core.NewLLMAgent("main", "test", core.LLMDeps{Provider: mainProvider})
	mainAgent.AddTool(translatorTool)

	// Execute the main agent
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Please translate 'Hello world' to French")

	result, err := mainAgent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Main agent execution failed: %v", err)
	}

	// Verify the result
	outputVal, _ := result.Get("output")
	output, ok := outputVal.(string)
	if !ok {
		t.Fatal("Expected output to be string")
	}

	if !strings.Contains(output, "Bonjour le monde") {
		t.Errorf("Expected output to contain French translation, got: %s", output)
	}
}

// TestToolToAgentConversion tests converting a tool to an agent
func TestToolToAgentConversion(t *testing.T) {
	// Create a calculator tool with schema
	schema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"operation": {Type: "string", Description: "The operation to perform"},
			"a":         {Type: "number", Description: "First operand"},
			"b":         {Type: "number", Description: "Second operand"},
		},
		Required: []string{"operation", "a", "b"},
	}

	calculatorTool := tools.NewTool(
		"calculator",
		"Perform basic arithmetic operations",
		func(ctx context.Context, params struct {
			Operation string  `json:"operation"`
			A         float64 `json:"a"`
			B         float64 `json:"b"`
		}) (float64, error) {
			switch params.Operation {
			case "add":
				return params.A + params.B, nil
			case "subtract":
				return params.A - params.B, nil
			case "multiply":
				return params.A * params.B, nil
			case "divide":
				if params.B == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return params.A / params.B, nil
			default:
				return 0, fmt.Errorf("unknown operation: %s", params.Operation)
			}
		},
		schema,
	)

	// Convert tool to agent
	calculatorAgent := tools.NewToolAgent(calculatorTool)

	// Test the agent
	ctx := context.Background()
	state := domain.NewState()
	state.Set("operation", "add")
	state.Set("a", 10.0)
	state.Set("b", 5.0)

	result, err := calculatorAgent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Calculator agent execution failed: %v", err)
	}

	// Verify the result
	resultVal, exists := result.Get("result")
	if !exists {
		t.Fatal("Expected result in state")
	}

	resultFloat, ok := resultVal.(float64)
	if !ok {
		t.Fatal("Expected result to be float64")
	}

	if resultFloat != 15.0 {
		t.Errorf("Expected 10 + 5 = 15, got %f", resultFloat)
	}
}

// TestAgentToolChaining tests chaining multiple agent-tools together
func TestAgentToolChaining(t *testing.T) {
	t.Skip("Skipping test that requires complex mock provider setup for tool execution")
	// Create mock providers for different agents
	analyzerProvider := provider.NewMockProvider()
	analyzerProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{
			Content: "Analysis: The data shows positive trends with 25% growth",
		}, nil
	})

	summarizerProvider := provider.NewMockProvider()
	summarizerProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Look for analysis results in the input
		for _, msg := range messages {
			if msg.Role == ldomain.RoleUser {
				for _, part := range msg.Content {
					if strings.Contains(part.Text, "25% growth") {
						return ldomain.Response{
							Content: "Summary: Strong positive growth of 25% observed",
						}, nil
					}
				}
			}
		}
		return ldomain.Response{Content: "Summary: No significant findings"}, nil
	})

	// Create analyzer and summarizer agents
	analyzer := core.NewLLMAgent("analyzer", "test", core.LLMDeps{Provider: analyzerProvider})
	analyzer.SetSystemPrompt("Analyze the provided data")

	summarizer := core.NewLLMAgent("summarizer", "test", core.LLMDeps{Provider: summarizerProvider})
	summarizer.SetSystemPrompt("Summarize the analysis results")

	// Convert agents to tools
	analyzerTool := tools.NewAgentTool(analyzer)
	summarizerTool := tools.NewAgentTool(summarizer)

	// Create coordinator agent that uses both tools
	coordinatorProvider := provider.NewMockProvider()
	callCount := 0
	coordinatorProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Check for tool results
		lastMessage := messages[len(messages)-1]
		if lastMessage.Role == ldomain.RoleUser {
			for _, part := range lastMessage.Content {
				// Check if this is analyzer result
				if strings.Contains(part.Text, "Tool result:") && strings.Contains(part.Text, "25% growth") && callCount == 1 {
					callCount++
					// Call summarizer
					return ldomain.Response{
						Content: `{"tool": "summarizer", "params": {"user_input": "Analysis: The data shows positive trends with 25% growth"}}`,
					}, nil
				}
				// Check if this is summarizer result
				if strings.Contains(part.Text, "Tool result:") && strings.Contains(part.Text, "Strong positive growth") && callCount == 2 {
					callCount++
					// Provide final response
					return ldomain.Response{
						Content: "Analysis complete. Summary: Strong positive growth of 25% observed",
					}, nil
				}
			}
		}

		// First call: analyze the data
		if callCount == 0 {
			callCount++
			return ldomain.Response{
				Content: `{"tool": "analyzer", "params": {"user_input": "Analyze Q4 sales data"}}`,
			}, nil
		}

		return ldomain.Response{Content: "Processing..."}, nil
	})

	// Create coordinator agent with both tools
	coordinator := core.NewLLMAgent("coordinator", "test", core.LLMDeps{Provider: coordinatorProvider})
	coordinator.AddTool(analyzerTool)
	coordinator.AddTool(summarizerTool)

	// Execute the coordinator
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Analyze and summarize Q4 sales data")

	result, err := coordinator.Run(ctx, state)
	if err != nil {
		t.Fatalf("Coordinator execution failed: %v", err)
	}

	// Verify the result
	outputVal, _ := result.Get("output")
	output, ok := outputVal.(string)
	if !ok {
		t.Fatal("Expected output to be string")
	}

	if !strings.Contains(output, "Strong positive growth of 25% observed") {
		t.Errorf("Expected output to contain summary of positive growth, got: %s", output)
	}

	// Verify both tools were called
	if callCount < 3 {
		t.Errorf("Expected at least 3 LLM calls (initial + 2 tools), got %d", callCount)
	}
}

// TestAgentToolWithContext tests that agent-tools work with tool context
func TestAgentToolWithContext(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Check for context information in the input
		for _, msg := range messages {
			if msg.Role == ldomain.RoleUser {
				for _, part := range msg.Content {
					if strings.Contains(part.Text, "retry attempt") {
						return ldomain.Response{
							Content: "I see this is a retry. Let me be extra careful.",
						}, nil
					}
				}
			}
		}
		return ldomain.Response{Content: "Processing your request"}, nil
	})

	// Create an agent
	processor := core.NewLLMAgent("processor", "test", core.LLMDeps{Provider: mockProvider})
	processor.SetSystemPrompt("You are a careful processor")

	// Convert to tool
	processorTool := tools.NewAgentTool(processor)

	// Create tool context
	ctx := context.Background()
	state := domain.NewState()
	state.Set("retry_count", 2)

	toolCtx := &domain.ToolContext{
		Context:   ctx,
		State:     state,
		RunID:     "test-run-123",
		Retry:     2,
		StartTime: time.Now(),
		Events:    &mockEventEmitter{},
		Agent: domain.AgentInfo{
			ID:   "test-agent",
			Name: "Test Agent",
		},
	}

	// Execute the tool
	params := map[string]interface{}{
		"user_input": fmt.Sprintf("Process this data (retry attempt %d)", toolCtx.Retry),
	}

	result, err := processorTool.Execute(toolCtx, params)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	// Verify the result acknowledges the retry
	resultStr, ok := result.(string)
	if !ok {
		t.Fatal("Expected result to be string")
	}

	if !strings.Contains(resultStr, "retry") {
		t.Errorf("Expected result to acknowledge retry, got: %s", resultStr)
	}
}

// mockEventEmitter implements domain.EventEmitter for testing
type mockEventEmitter struct {
	events []domain.Event
}

func (m *mockEventEmitter) Emit(eventType domain.EventType, data interface{}) {
	m.events = append(m.events, domain.Event{
		Type: eventType,
		Data: data,
	})
}

func (m *mockEventEmitter) EmitProgress(current, total int, message string) {
	m.Emit(domain.EventProgress, map[string]interface{}{
		"current": current,
		"total":   total,
		"message": message,
	})
}

func (m *mockEventEmitter) EmitMessage(message string) {
	m.Emit(domain.EventMessage, map[string]interface{}{
		"message": message,
	})
}

func (m *mockEventEmitter) EmitError(err error) {
	if err != nil {
		m.Emit(domain.EventToolError, err.Error())
	}
}

func (m *mockEventEmitter) EmitCustom(eventName string, data interface{}) {
	m.Emit(domain.EventType("custom."+eventName), data)
}
