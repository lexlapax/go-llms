// ABOUTME: Example demonstrating the API Client Tool for making REST API calls with LLM agents
// ABOUTME: Shows authentication, error handling, and various HTTP methods with GitHub API

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
	ctx := context.Background()

	// Create LLM provider
	provider, err := llmutil.NewProviderFromString("openai")
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
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
	
	// Set system prompt
	agent.SetSystemPrompt(`You are an API interaction assistant. Use the api_client tool to:
1. Make REST API calls to various services
2. Handle authentication when needed
3. Parse and explain API responses
4. Provide helpful error guidance

When making API calls, be sure to:
- Use the correct HTTP method (GET, POST, PUT, DELETE, etc.)
- Include necessary authentication if required
- Format request bodies as JSON when needed
- Handle errors gracefully and suggest fixes`)
	
	// Add the API client tool
	agent.AddTool(apiTool)

	// Example 1: Simple GET request to GitHub API
	fmt.Println("=== Example 1: Fetching GitHub User Info ===")
	state1 := domain.NewState()
	state1.Set("prompt", "Fetch information about the GitHub user 'octocat' using the GitHub API")

	result1, err := agent.Run(ctx, state1)
	if err != nil {
		log.Printf("Error in example 1: %v", err)
	} else {
		printLastMessage(result1)
	}

	// Example 2: Search GitHub repositories
	fmt.Println("\n=== Example 2: Searching GitHub Repositories ===")
	state2 := domain.NewState()
	state2.Set("prompt", "Search for Go repositories related to 'llm' on GitHub. Show me the top 3 most starred ones.")

	result2, err := agent.Run(ctx, state2)
	if err != nil {
		log.Printf("Error in example 2: %v", err)
	} else {
		printLastMessage(result2)
	}

	// Example 3: Making authenticated API calls (if API key is available)
	if apiKey := os.Getenv("GITHUB_TOKEN"); apiKey != "" {
		fmt.Println("\n=== Example 3: Authenticated API Call ===")
		state3 := domain.NewState()
		state3.Set("prompt", fmt.Sprintf("Check my GitHub rate limit status. Use this token for authentication: %s", apiKey))

		result3, err := agent.Run(ctx, state3)
		if err != nil {
			log.Printf("Error in example 3: %v", err)
		} else {
			printLastMessage(result3)
		}
	}

	// Example 4: POST request (creating a gist if authenticated)
	if apiKey := os.Getenv("GITHUB_TOKEN"); apiKey != "" {
		fmt.Println("\n=== Example 4: Creating a GitHub Gist ===")
		state4 := domain.NewState()
		state4.Set("prompt", fmt.Sprintf(`Create a new GitHub gist with the following:
- Description: "API Client Tool Demo"
- Filename: "hello.go"
- Content: "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello from API Client Tool!\")\n}"
- Make it public
- Use this token: %s`, apiKey))

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
	state5.Set("prompt", "Try to access a non-existent GitHub user 'this-user-definitely-does-not-exist-12345'")

	result5, err := agent.Run(ctx, state5)
	if err != nil {
		log.Printf("Error in example 5: %v", err)
	} else {
		printLastMessage(result5)
	}

	// Example 6: Using path parameters
	fmt.Println("\n=== Example 6: Path Parameters ===")
	state6 := domain.NewState()
	state6.Set("prompt", "Get information about the 'go-llms' repository owned by 'lexlapax' on GitHub")

	result6, err := agent.Run(ctx, state6)
	if err != nil {
		log.Printf("Error in example 6: %v", err)
	} else {
		printLastMessage(result6)
	}

	// Example 7: Working with a different API (JSONPlaceholder)
	fmt.Println("\n=== Example 7: JSONPlaceholder API ===")
	state7 := domain.NewState()
	state7.Set("prompt", "Fetch the first 5 posts from JSONPlaceholder API (https://jsonplaceholder.typicode.com) and summarize them")

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