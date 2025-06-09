// ABOUTME: Example demonstrating the API Client Tool for making REST API calls with LLM agents
// ABOUTME: Shows authentication, error handling, and various HTTP methods with GitHub API

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
	providerString := "anthropic/claude-3-7-sonnet-latest"
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
	agent := core.NewLLMAgent("api-demo",
		"I help you interact with REST APIs",
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
	agent.SetSystemPrompt(`You are an API interaction assistant. You MUST use the api_client tool to make actual API calls when requested.

Use the api_client tool to:
1. Make REST API calls to various services
2. Handle authentication when needed
3. Parse and explain API responses
4. Provide helpful error guidance

When making API calls, be sure to:
- Use the correct HTTP method (GET, POST, PUT, DELETE, etc.)
- Include necessary authentication if required
- Format request bodies as JSON when needed
- Handle errors gracefully and suggest fixes

IMPORTANT: Always make the actual API call using the api_client tool. Do not just describe what you would do - actually do it.`)

	// Add the API client tool
	agent.AddTool(apiTool)

	// Example 1: Simple GET request to GitHub API
	fmt.Println("=== Example 1: Fetching GitHub User Info ===")
	state1 := domain.NewState()
	state1.Set("user_input", "Use the api_client tool to fetch information about the GitHub user 'octocat' from the GitHub API")

	result1, err := agent.Run(ctx, state1)
	if err != nil {
		log.Printf("Error in example 1: %v", err)
	} else {
		printLastMessage(result1)
	}

	// Example 2: Search GitHub repositories
	fmt.Println("\n=== Example 2: Searching GitHub Repositories ===")
	state2 := domain.NewState()
	state2.Set("user_input", "Use the api_client tool to search for Go repositories related to 'llm' on GitHub. Make the API call to GitHub's search endpoint with parameters to find repositories with language:go and query 'llm', sorted by stars. Show me the top 3 results.")

	result2, err := agent.Run(ctx, state2)
	if err != nil {
		log.Printf("Error in example 2: %v", err)
	} else {
		printLastMessage(result2)
	}

	// Example 3: Making authenticated API calls (if API key is available)
	if apiKey := os.Getenv("GITHUB_API_KEY"); apiKey != "" {
		fmt.Println("\n=== Example 3: Authenticated API Call ===")
		state3 := domain.NewState()
		// Store the API key in state for automatic authentication
		state3.Set("github_api_key", apiKey)
		state3.Set("user_input", "Use the api_client tool to check my GitHub rate limit status at https://api.github.com/rate_limit")

		result3, err := agent.Run(ctx, state3)
		if err != nil {
			log.Printf("Error in example 3: %v", err)
		} else {
			printLastMessage(result3)
		}
	}

	// Example 4: POST request (creating a gist if authenticated)
	if apiKey := os.Getenv("GITHUB_API_KEY"); apiKey != "" {
		fmt.Println("\n=== Example 4: Creating a GitHub Gist ===")
		state4 := domain.NewState()
		// Store the API key in state for automatic authentication
		state4.Set("github_api_key", apiKey)
		state4.Set("user_input", `Use the api_client tool to create a new GitHub gist at https://api.github.com/gists with the following:
- Description: "API Client Tool Demo"
- Filename: "hello.go"
- Content: "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello from API Client Tool!\")\n}"
- Make it public`)

		result4, err := agent.Run(ctx, state4)
		if err != nil {
			log.Printf("Error in example 4: %v", err)
		} else {
			printLastMessage(result4)
		}
	}

	// Example 5: Error handling demonstration
	fmt.Println("\n=== Example 5: Error Handling ===")
	state5 := domain.NewState()
	state5.Set("user_input", "Use the api_client tool to try to access a non-existent GitHub user 'this-user-definitely-does-not-exist-12345'")

	result5, err := agent.Run(ctx, state5)
	if err != nil {
		log.Printf("Error in example 5: %v", err)
	} else {
		printLastMessage(result5)
	}

	// Example 6: Using path parameters
	fmt.Println("\n=== Example 6: Path Parameters ===")
	state6 := domain.NewState()
	state6.Set("user_input", "Use the api_client tool to get information about the 'go-llms' repository owned by 'lexlapax' on GitHub")

	result6, err := agent.Run(ctx, state6)
	if err != nil {
		log.Printf("Error in example 6: %v", err)
	} else {
		printLastMessage(result6)
	}

	// Example 7: Working with a different API (JSONPlaceholder)
	fmt.Println("\n=== Example 7: JSONPlaceholder API ===")
	state7 := domain.NewState()
	state7.Set("user_input", "Use the api_client tool to fetch the first 5 posts from JSONPlaceholder API (https://jsonplaceholder.typicode.com) and summarize them")

	result7, err := agent.Run(ctx, state7)
	if err != nil {
		log.Printf("Error in example 7: %v", err)
	} else {
		printLastMessage(result7)
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
