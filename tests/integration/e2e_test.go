package integration

// ABOUTME: End-to-end integration tests with real providers
// ABOUTME: Tests complete workflows from validation to agent execution

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// TestEndToEndWorkflow tests the entire workflow from validation to provider to agent
func TestEndToEndWorkflow(t *testing.T) {
	// Skip if we don't have API keys
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable not set, skipping end-to-end test")
	}

	// Create custom HTTP client with longer timeouts for reliability
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	t.Run("ValidateProcessGenerate", func(t *testing.T) {
		// Create a schema
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"result": {
					Type:        "integer",
					Description: "The result of the calculation",
				},
			},
			Required: []string{"result"},
		}

		// Create a validator
		validator := validation.NewValidator()

		// Create a structured processor
		processor := processor.NewStructuredProcessor(validator)

		// Create an LLM provider with custom client
		clientOption := ldomain.NewHTTPClientOption(httpClient)
		llm := provider.NewOpenAIProvider(apiKey, "gpt-4o", clientOption)

		// Test generation
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := llm.GenerateWithSchema(ctx, "What is 21 * 2? Respond with just the result in the required JSON format.", schema)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		// Convert response to JSON
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal response: %v", err)
		}

		// Process and validate
		result, err := processor.Process(schema, string(jsonBytes))
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		// Check result
		data, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got: %T", result)
		}

		// Check result
		resultValue, ok := data["result"].(float64)
		if !ok {
			t.Errorf("Expected integer result, got: %T", data["result"])
		}

		if resultValue != 42 {
			t.Errorf("Expected result 42, got: %v", resultValue)
		}
	})
}

// TestLiveEndToEndAgent tests the agent with real providers and tools
// This is similar to TestEndToEndAgent but uses real API keys
func TestLiveEndToEndAgent(t *testing.T) {
	// Skip if we don't have API keys
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable not set, skipping live end-to-end agent test")
	}

	// Create custom HTTP client with longer timeouts for reliability
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Create an LLM provider with custom client for better reliability
	clientOption := ldomain.NewHTTPClientOption(httpClient)
	llm := provider.NewOpenAIProvider(apiKey, "gpt-4o", clientOption)

	// Create an agent with new architecture
	deps := core.LLMDeps{
		Provider: llm,
	}
	agent := core.NewLLMAgent("e2e-agent", "gpt-4o", deps)

	// Add a system prompt with explicit instructions to use tools
	agent.SetSystemPrompt(`You are a helpful assistant that can answer questions and use tools.
When asked about date or time information, ALWAYS use the get_current_date tool.
When asked to perform calculations, ALWAYS use the multiply tool.
Do not try to calculate or determine dates yourself - use the provided tools.`)

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

	// Add a calculator tool for multiply
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
				"calculation": fmt.Sprintf("%g * %g = %g", params.A, params.B, result),
				"a":           params.A,
				"b":           params.B,
			}, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
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

	// Test cases that require different tool usage
	testCases := []struct {
		name           string
		query          string
		validateResult func(t *testing.T, response string)
	}{
		{
			name:  "Simple greeting",
			query: "Hello! How are you?",
			validateResult: func(t *testing.T, response string) {
				// Should get a conversational response without tool usage
				lower := strings.ToLower(response)
				if strings.Contains(lower, "tool_calls") || strings.Contains(lower, "multiply") {
					t.Errorf("Expected conversational response without tools, got: %s", response)
				}
			},
		},
		{
			name:  "Use date tool",
			query: "What year is it?",
			validateResult: func(t *testing.T, response string) {
				// Should contain the current year
				currentYear := fmt.Sprintf("%d", time.Now().Year())
				if !strings.Contains(response, currentYear) {
					t.Errorf("Expected response to contain current year %s, got: %s", currentYear, response)
				}
			},
		},
		{
			name:  "Use multiply tool",
			query: "What is 15 times 7?",
			validateResult: func(t *testing.T, response string) {
				// Should contain the result 105
				if !strings.Contains(response, "105") {
					t.Errorf("Expected response to contain '105', got: %s", response)
				}
			},
		},
		{
			name:  "Complex request",
			query: "What is 8 times 9? Also, what's today's date?",
			validateResult: func(t *testing.T, response string) {
				// Should contain both the calculation result and date
				if !strings.Contains(response, "72") {
					t.Errorf("Expected response to contain '72', got: %s", response)
				}
				// Should also have date info (at least the year)
				currentYear := fmt.Sprintf("%d", time.Now().Year())
				if !strings.Contains(response, currentYear) {
					t.Errorf("Expected response to contain year %s, got: %s", currentYear, response)
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

			// Log the full response for debugging
			t.Logf("Query: %s", tc.query)
			t.Logf("Response: %s", response)

			// Validate the response
			tc.validateResult(t, response)
		})
	}
}

// TestStructuredOutputWithAgent tests structured output processing with agent
func TestStructuredOutputWithAgent(t *testing.T) {
	// Skip if we don't have API keys
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable not set, skipping structured output test")
	}

	// Create an LLM provider
	llm := provider.NewOpenAIProvider(apiKey, "gpt-4o")

	// Create an agent
	deps := core.LLMDeps{
		Provider: llm,
	}
	agent := core.NewLLMAgent("structured-output-agent", "gpt-4o", deps)
	agent.SetSystemPrompt("You are a helpful assistant that provides structured data.")

	// Create a tool that returns structured data
	agent.AddTool(tools.NewTool(
		"get_user_info",
		"Get information about a user",
		func(params struct {
			UserID string `json:"user_id"`
		}) (map[string]interface{}, error) {
			// Simulate user data
			users := map[string]map[string]interface{}{
				"123": {
					"name":  "Alice Johnson",
					"age":   28,
					"email": "alice@example.com",
					"role":  "developer",
				},
				"456": {
					"name":  "Bob Smith",
					"age":   35,
					"email": "bob@example.com",
					"role":  "manager",
				},
			}

			if user, ok := users[params.UserID]; ok {
				return user, nil
			}
			return nil, fmt.Errorf("user not found: %s", params.UserID)
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"user_id": {
					Type:        "string",
					Description: "The ID of the user to get information for",
				},
			},
			Required: []string{"user_id"},
		},
	))

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Get me information about user 123")

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

	// Verify response contains user information
	expectedInfo := []string{"Alice Johnson", "28", "alice@example.com", "developer"}
	for _, info := range expectedInfo {
		if !strings.Contains(strings.ToLower(response), strings.ToLower(info)) {
			t.Errorf("Expected response to contain '%s', got: %s", info, response)
		}
	}
}