// ABOUTME: Example demonstrating OpenAPI spec discovery and validation with the API Client Tool
// ABOUTME: Shows how to discover endpoints, validate requests, and interact with APIs using OpenAPI specs

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
	ctx := context.Background()

	// Create LLM provider
	// using one of three provider/model combinations:
	// 1. "openai/gpt-4o"
	// 2. "anthropic/claude-3-7-sonnet-latest"
	// 3. "gemini/gemini-2.0-flash"
	providerString := "gemini/gemini-2.0-flash"
	provider, err := llmutil.NewProviderFromString(providerString)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Parse provider string to get provider and model info
	providerName, modelName, _ := llmutil.ParseProviderModelString(providerString)

	// Print provider information
	fmt.Printf("Provider: %s\n", providerName)
	if modelName != "" {
		fmt.Printf("Model: %s\n\n", modelName)
	} else {
		fmt.Printf("Model: (default for provider)\n\n")
	}

	// Create API Client Tool
	apiTool := web.NewAPIClientTool()

	// Create an LLM agent
	deps := core.LLMDeps{
		Provider: provider,
	}
	agent := core.NewLLMAgent("openapi-demo",
		"I help you discover and interact with APIs using OpenAPI specifications",
		deps,
	)

	// Add logging hook if DEBUG=1
	if os.Getenv("DEBUG") == "1" {
		// Create slog logger that outputs to stderr
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stderr, opts))

		// Create logging hook with debug level
		loggingHook := core.NewLoggingHook(logger, core.LogLevelDebug)
		agent.WithHook(loggingHook)
		log.Println("Debug logging enabled")
	}

	// Set system prompt
	agent.SetSystemPrompt(`You are an OpenAPI-aware API interaction assistant. You MUST immediately use the api_client tool when requested.

IMPORTANT: When the user says "IMMEDIATELY use the api_client tool", you must respond with a tool call, not a description.

Use the api_client tool to:
1. Discover available operations from OpenAPI/Swagger specifications
2. Validate API requests against the spec before sending
3. Make API calls with proper parameters and authentication
4. Provide helpful guidance based on the API documentation

When working with OpenAPI specs:
- First use discover_operations=true to explore available endpoints
- Pay attention to required parameters and authentication requirements
- Use the spec URL for validation when making actual API calls
- Help users understand the API structure and capabilities

DO NOT just describe what you would do - actually make the API calls using the tool.`)

	// Add the API client tool
	agent.AddTool(apiTool)

	// Example 1: GitHub API Discovery (GitHub provides OpenAPI spec)
	fmt.Println("=== Example 1: GitHub API OpenAPI Discovery ===")
	state1 := domain.NewState()
	state1.Set("user_input", `Use the api_client tool to discover what endpoints are available in the GitHub API. 
Use their OpenAPI spec at: https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json
Show me the first 10 operations you find.`)

	result1, err := agent.Run(ctx, state1)
	if err != nil {
		log.Printf("Error in example 1: %v", err)
	} else {
		printLastMessage(result1)
	}

	// Example 2: PetStore API Discovery (Classic OpenAPI example)
	fmt.Println("\n=== Example 2: PetStore API Discovery ===")
	state2 := domain.NewState()
	state2.Set("user_input", `Use the api_client tool to explore the PetStore API using its OpenAPI spec at:
https://petstore3.swagger.io/api/v3/openapi.json
What operations are available for managing pets?`)

	result2, err := agent.Run(ctx, state2)
	if err != nil {
		log.Printf("Error in example 2: %v", err)
	} else {
		printLastMessage(result2)
	}

	// Example 3: Making a validated call to PetStore
	fmt.Println("\n=== Example 3: Validated API Call to PetStore ===")
	state3 := domain.NewState()
	state3.Set("user_input", `IMMEDIATELY use the api_client tool to make a GET request to the PetStore API endpoint /pet/findByStatus with query parameter status=available.
Base URL: https://petstore3.swagger.io/api/v3
OpenAPI spec for validation: https://petstore3.swagger.io/api/v3/openapi.json`)

	result3, err := agent.Run(ctx, state3)
	if err != nil {
		log.Printf("Error in example 3: %v", err)
	} else {
		printLastMessage(result3)
	}

	// Example 4: JSONPlaceholder API (No OpenAPI spec, but we can still use the tool)
	fmt.Println("\n=== Example 4: JSONPlaceholder API ===")
	state4 := domain.NewState()
	state4.Set("user_input", `IMMEDIATELY use the api_client tool to:
1. First, make a GET request to https://jsonplaceholder.typicode.com/posts?_limit=5 to fetch the first 5 posts
2. Then, make a POST request to https://jsonplaceholder.typicode.com/posts with body: {"title": "Test Post", "body": "Created with OpenAPI discovery example", "userId": 1}`)

	result4, err := agent.Run(ctx, state4)
	if err != nil {
		log.Printf("Error in example 4: %v", err)
	} else {
		printLastMessage(result4)
	}

	// Example 5: API with authentication (if GitHub token available)
	if apiKey := os.Getenv("GITHUB_API_KEY"); apiKey != "" {
		fmt.Println("\n=== Example 5: Authenticated GitHub API with OpenAPI ===")
		state5 := domain.NewState()
		// Store the API key in state for automatic authentication
		state5.Set("github_api_key", apiKey)
		state5.Set("user_input", `Use the api_client tool with the GitHub OpenAPI spec to understand the authentication requirements.
Then list my repositories at https://api.github.com/user/repos.
Use the spec at: https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json`)

		result5, err := agent.Run(ctx, state5)
		if err != nil {
			log.Printf("Error in example 5: %v", err)
		} else {
			printLastMessage(result5)
		}
	}

	// Example 6: Error handling with OpenAPI validation
	fmt.Println("\n=== Example 6: OpenAPI Validation Error Handling ===")
	state6 := domain.NewState()
	state6.Set("user_input", `IMMEDIATELY use the api_client tool to make a POST request to https://petstore3.swagger.io/api/v3/pet with this INVALID body (missing required name field):
{"id": 999, "status": "available"}
Include the OpenAPI spec URL for validation: https://petstore3.swagger.io/api/v3/openapi.json`)

	result6, err := agent.Run(ctx, state6)
	if err != nil {
		log.Printf("Error in example 6: %v", err)
	} else {
		printLastMessage(result6)
	}
}

func printLastMessage(state *domain.State) {
	// Try to get the response from various possible keys
	responseKeys := []string{"response", "output", "result", "answer", "reply"}

	for _, key := range responseKeys {
		if value, exists := state.Get(key); exists {
			fmt.Printf("Response: %v\n", value)
			return
		}
	}

	// If no response found, print the whole state for debugging
	fmt.Printf("State: %+v\n", state.Values())
}
