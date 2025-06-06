package integration

// ABOUTME: End-to-end integration tests for agent with Gemini provider
// ABOUTME: Tests real LLM interactions with tools using Google Gemini models

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

// TestLiveEndToEndAgentGemini tests the agent with Gemini provider
func TestLiveEndToEndAgentGemini(t *testing.T) {
	// Skip if we don't have API keys
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY environment variable not set, skipping live Gemini end-to-end agent test")
	}

	// Create a Gemini provider
	// Using the gemini-2.0-flash model as it has better capabilities for agent workflows
	// than the default flash-lite model
	llm := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash")

	// Create an agent with new architecture
	deps := core.LLMDeps{
		Provider: llm,
	}
	agent := core.NewLLMAgent("gemini-e2e-agent", "gemini-2.0-flash", deps)

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
				if !strings.Contains(lower, "hello") && !strings.Contains(lower, "hi") && !strings.Contains(lower, "good") && !strings.Contains(lower, "well") {
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
				// Check for at least the year, as formatting might vary
				if !strings.Contains(response, today[:4]) {
					t.Errorf("Expected response to contain current year, got: %s", response)
				}
			},
		},
		{
			name:  "Use calculator tool",
			query: "What is 25 times 12?",
			validate: func(t *testing.T, response string) {
				// Should contain the result 300
				if !strings.Contains(response, "300") {
					t.Errorf("Expected response to contain '300', got: %s", response)
				}
			},
		},
		{
			name:  "Division calculation",
			query: "Can you divide 144 by 12 for me?",
			validate: func(t *testing.T, response string) {
				// Should contain the result 12
				if !strings.Contains(response, "12") {
					t.Errorf("Expected response to contain '12', got: %s", response)
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

// TestLiveGeminiComplexWorkflow tests more complex agent workflows with Gemini
func TestLiveGeminiComplexWorkflow(t *testing.T) {
	// Skip if we don't have API keys
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY environment variable not set, skipping live Gemini complex workflow test")
	}

	// Create a Gemini provider
	llm := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash")

	// Create an agent
	deps := core.LLMDeps{
		Provider: llm,
	}
	agent := core.NewLLMAgent("gemini-complex-agent", "gemini-2.0-flash", deps)
	agent.SetSystemPrompt("You are a helpful assistant that can perform complex tasks using multiple tools.")

	// Add a weather tool (simulated)
	agent.AddTool(tools.NewTool(
		"get_weather",
		"Get the current weather for a location",
		func(params struct {
			Location string `json:"location"`
		}) (map[string]interface{}, error) {
			// Simulate weather data
			weatherData := map[string]map[string]interface{}{
				"new york": {
					"temperature": 72,
					"condition":   "partly cloudy",
					"humidity":    65,
					"wind":        "10 mph",
				},
				"london": {
					"temperature": 64,
					"condition":   "rainy",
					"humidity":    85,
					"wind":        "15 mph",
				},
				"tokyo": {
					"temperature": 78,
					"condition":   "sunny",
					"humidity":    55,
					"wind":        "5 mph",
				},
			}

			location := strings.ToLower(params.Location)
			if data, ok := weatherData[location]; ok {
				data["location"] = params.Location
				return data, nil
			}

			// Default weather for unknown locations
			return map[string]interface{}{
				"location":    params.Location,
				"temperature": 70,
				"condition":   "clear",
				"humidity":    60,
				"wind":        "8 mph",
			}, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"location": {
					Type:        "string",
					Description: "The location to get weather for",
				},
			},
			Required: []string{"location"},
		},
	))

	// Add a temperature converter tool
	agent.AddTool(tools.NewTool(
		"convert_temperature",
		"Convert temperature between Celsius and Fahrenheit",
		func(params struct {
			Value float64 `json:"value"`
			From  string  `json:"from"`
			To    string  `json:"to"`
		}) (map[string]interface{}, error) {
			var result float64
			if strings.ToLower(params.From) == "fahrenheit" && strings.ToLower(params.To) == "celsius" {
				result = (params.Value - 32) * 5 / 9
			} else if strings.ToLower(params.From) == "celsius" && strings.ToLower(params.To) == "fahrenheit" {
				result = (params.Value * 9 / 5) + 32
			} else {
				return nil, fmt.Errorf("invalid conversion: %s to %s", params.From, params.To)
			}
			return map[string]interface{}{
				"original_value":  params.Value,
				"original_unit":   params.From,
				"converted_value": result,
				"converted_unit":  params.To,
			}, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"value": {
					Type:        "number",
					Description: "The temperature value to convert",
				},
				"from": {
					Type:        "string",
					Description: "The unit to convert from (Celsius or Fahrenheit)",
				},
				"to": {
					Type:        "string",
					Description: "The unit to convert to (Celsius or Fahrenheit)",
				},
			},
			Required: []string{"value", "from", "to"},
		},
	))

	// Test complex workflow
	t.Run("WeatherAndConversion", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Create initial state
		state := domain.NewState()
		state.Set("user_input", "What's the weather in New York? Also convert the temperature to Celsius.")

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

		t.Logf("Complex workflow response: %s", response)

		// Verify the response contains weather info and conversion
		lower := strings.ToLower(response)
		if !strings.Contains(lower, "new york") {
			t.Error("Expected response to mention New York")
		}
		if !strings.Contains(lower, "celsius") || !strings.Contains(lower, "°c") {
			t.Error("Expected response to contain Celsius conversion")
		}
		// Should contain weather conditions
		if !strings.Contains(lower, "cloudy") && !strings.Contains(lower, "weather") {
			t.Error("Expected response to contain weather information")
		}
	})
}

// TestLiveGeminiErrorRecovery tests error handling and recovery with Gemini
func TestLiveGeminiErrorRecovery(t *testing.T) {
	// Skip if we don't have API keys
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY environment variable not set, skipping live Gemini error recovery test")
	}

	// Create a Gemini provider
	llm := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash")

	// Create an agent
	deps := core.LLMDeps{
		Provider: llm,
	}
	agent := core.NewLLMAgent("gemini-error-agent", "gemini-2.0-flash", deps)
	agent.SetSystemPrompt("You are a helpful assistant. When tools fail, explain the error gracefully.")

	// Add a calculator tool that can fail
	agent.AddTool(tools.NewTool(
		"safe_divide",
		"Safely divide two numbers",
		func(params struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}) (map[string]interface{}, error) {
			if params.B == 0 {
				return nil, fmt.Errorf("cannot divide by zero")
			}
			result := params.A / params.B
			return map[string]interface{}{
				"result": result,
				"a":      params.A,
				"b":      params.B,
			}, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"a": {
					Type:        "number",
					Description: "The dividend",
				},
				"b": {
					Type:        "number",
					Description: "The divisor",
				},
			},
			Required: []string{"a", "b"},
		},
	))

	// Test error handling
	t.Run("DivisionByZero", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Create initial state
		state := domain.NewState()
		state.Set("user_input", "Please divide 100 by 0")

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

		t.Logf("Error handling response: %s", response)

		// Verify the response handles the error gracefully
		lower := strings.ToLower(response)
		if !strings.Contains(lower, "zero") || !strings.Contains(lower, "cannot") || !strings.Contains(lower, "divide") {
			t.Error("Expected response to explain division by zero error")
		}
	})
}
