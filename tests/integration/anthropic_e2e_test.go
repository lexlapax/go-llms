package integration

// ABOUTME: End-to-end integration tests for agent with Anthropic provider
// ABOUTME: Tests real LLM interactions with tools using Claude models

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestLiveEndToEndAgentAnthropic tests the agent with Anthropic provider
func TestLiveEndToEndAgentAnthropic(t *testing.T) {
	// Skip if we don't have API keys
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY environment variable not set, skipping live Anthropic end-to-end test")
	}

	// Create an Anthropic provider
	llm := provider.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-latest")

	// Create an agent with new architecture
	deps := core.LLMDeps{
		Provider: llm,
	}
	agent := core.NewLLMAgent("anthropic-e2e-agent", "claude-3-5-sonnet-latest", deps)

	// Add a system prompt
	agent.SetSystemPrompt("You are a helpful assistant that can answer questions and use tools.")

	// Add date and calculator tools
	agent.AddTool(tools.NewTool(
		"get_current_date",
		"Get the current date",
		func() map[string]string {
			now := time.Now()
			return map[string]string{
				"date": now.Format("2006-01-02"),
				"time": now.Format("15:04:05"),
				"year": fmt.Sprintf("%d", now.Year()),
			}
		},
		&sdomain.Schema{
			Type:        "object",
			Description: "Returns the current date and time",
		},
	))

	agent.AddTool(tools.NewTool(
		"calculator",
		"Perform basic arithmetic calculations",
		func(params struct {
			Operation string  `json:"operation" description:"The operation to perform: add, subtract, multiply, or divide"`
			A         float64 `json:"a" description:"The first number"`
			B         float64 `json:"b" description:"The second number"`
		}) (map[string]interface{}, error) {
			var result float64
			switch params.Operation {
			case "add":
				result = params.A + params.B
			case "subtract":
				result = params.A - params.B
			case "multiply":
				result = params.A * params.B
			case "divide":
				if params.B == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = params.A / params.B
			default:
				return nil, fmt.Errorf("unknown operation: %s", params.Operation)
			}
			return map[string]interface{}{
				"result":    result,
				"operation": params.Operation,
				"a":         params.A,
				"b":         params.B,
			}, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"operation": {
					Type:        "string",
					Description: "The operation to perform",
					Enum:        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": {
					Type:        "number",
					Description: "The first number",
				},
				"b": {
					Type:        "number",
					Description: "The second number",
				},
			},
			Required: []string{"operation", "a", "b"},
		},
	))

	// Test cases
	testCases := []struct {
		name     string
		query    string
		validate func(t *testing.T, response string)
	}{
		{
			name:  "Simple greeting",
			query: "Hello! How are you today?",
			validate: func(t *testing.T, response string) {
				// Should contain a greeting response
				lower := strings.ToLower(response)
				if !strings.Contains(lower, "hello") && !strings.Contains(lower, "hi") && !strings.Contains(lower, "good") {
					t.Errorf("Expected greeting response, got: %s", response)
				}
			},
		},
		{
			name:  "Use date tool",
			query: "What's today's date?",
			validate: func(t *testing.T, response string) {
				// Should contain today's date
				today := time.Now().Format("2006-01-02")
				if !strings.Contains(response, today[:4]) { // At least the year
					t.Errorf("Expected response to contain current year, got: %s", response)
				}
			},
		},
		{
			name:  "Use calculator tool",
			query: "What is 42 times 17?",
			validate: func(t *testing.T, response string) {
				// Should contain the result 714
				if !strings.Contains(response, "714") {
					t.Errorf("Expected response to contain '714', got: %s", response)
				}
			},
		},
		{
			name:  "Complex calculation",
			query: "Calculate (100 + 50) * 2 for me. First add 100 and 50, then multiply the result by 2.",
			validate: func(t *testing.T, response string) {
				// Should contain 150 (intermediate) and 300 (final)
				if !strings.Contains(response, "300") {
					t.Errorf("Expected response to contain final result '300', got: %s", response)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create initial state
			state := domain.NewState()
			state.Set("user_input", tc.query)

			// Run the agent
			finalState, err := agent.Run(ctx, state)
			if err != nil {
				t.Fatalf("Agent run failed: %v", err)
			}

			// Check final output
			output, ok := finalState.Get("output")
			if !ok {
				t.Fatal("No output in final state")
			}

			response, ok := output.(string)
			if !ok {
				t.Fatal("Output is not a string")
			}

			// Log the response for debugging
			t.Logf("Query: %s", tc.query)
			t.Logf("Response: %s", response)

			// Validate the response
			tc.validate(t, response)
		})
	}
}

// TestLiveAnthropicStreamingAgent tests streaming responses with Anthropic
func TestLiveAnthropicStreamingAgent(t *testing.T) {
	// Skip if we don't have API keys
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY environment variable not set, skipping live Anthropic streaming test")
	}

	// Create an Anthropic provider
	llm := provider.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-latest")

	// Create an agent
	deps := core.LLMDeps{
		Provider: llm,
	}
	agent := core.NewLLMAgent("anthropic-stream-agent", "claude-3-5-sonnet-latest", deps)
	agent.SetSystemPrompt("You are a helpful assistant. Keep your responses brief.")

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Tell me a very short story (2-3 sentences) about a robot.")
	state.Set("stream", true) // Enable streaming if supported

	// Run the agent
	finalState, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Check final output
	output, ok := finalState.Get("output")
	if !ok {
		t.Fatal("No output in final state")
	}

	response, ok := output.(string)
	if !ok {
		t.Fatal("Output is not a string")
	}

	// Log the response
	t.Logf("Streaming response: %s", response)

	// Validate we got a response about a robot
	lower := strings.ToLower(response)
	if !strings.Contains(lower, "robot") {
		t.Errorf("Expected response about a robot, got: %s", response)
	}
}

// TestLiveAnthropicErrorHandling tests error scenarios with real API
func TestLiveAnthropicErrorHandling(t *testing.T) {
	// Create an Anthropic provider with invalid API key
	llm := provider.NewAnthropicProvider("invalid-api-key", "claude-3-5-sonnet-latest")

	// Create an agent
	deps := core.LLMDeps{
		Provider: llm,
	}
	agent := core.NewLLMAgent("anthropic-error-agent", "claude-3-5-sonnet-latest", deps)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Hello!")

	// Run the agent - should fail
	_, err := agent.Run(ctx, state)
	if err == nil {
		t.Fatal("Expected error with invalid API key, got nil")
	}

	// Log the error
	t.Logf("Expected error received: %v", err)

	// Verify it's an authentication error
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "401") && !strings.Contains(errStr, "unauthorized") && !strings.Contains(errStr, "invalid") {
		t.Errorf("Expected authentication error, got: %v", err)
	}
}

