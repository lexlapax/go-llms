// ABOUTME: Example of using the api_client tool to interact with GraphQL APIs
// ABOUTME: Demonstrates discovery, queries with variables, and mutations

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

	// Check for required API keys
	githubToken := os.Getenv("GITHUB_API_KEY")
	if githubToken == "" {
		log.Println("Warning: GITHUB_API_KEY not set, GitHub GraphQL examples will fail")
	} else {
		log.Printf("GitHub API key found: %s...%s", githubToken[:10], githubToken[len(githubToken)-4:])
	}

	// Create LLM provider
	// using one of three provider/model combinations:
	// 1. "openai/gpt-4o"
	// 2. "anthropic/claude-3-7-sonnet-latest"
	// 3. "gemini/gemini-2.0-flash"
	providerString := "openai/gpt-4o"
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
	agent := core.NewLLMAgent("graphql-demo",
		"GraphQL API exploration assistant",
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
	}

	// Set system prompt
	agent.SetSystemPrompt(`You are a GraphQL API exploration assistant. You MUST use the api_client tool to interact with GraphQL endpoints.

When asked to explore GraphQL APIs:
1. First use discover_graphql to understand the schema
2. Then execute specific queries based on what you find
3. Explain the results clearly

IMPORTANT: Always make the actual API call using the api_client tool. Do not just describe what you would do - actually do it.`)

	// Add the API tool
	agent.AddTool(apiTool)

	// Example 1: Discover GitHub GraphQL Schema
	fmt.Println("=== Example 1: Discovering GitHub GraphQL Schema ===")
	state := domain.NewState()
	// Set GitHub API key in state for authentication
	if githubToken != "" {
		state.Set("github_token", githubToken)
	}
	state.Set("user_input", "Use the api_client tool to discover what GraphQL operations are available at GitHub's GraphQL API endpoint https://api.github.com/graphql.")

	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error discovering schema: %v", err)
	} else {
		if content, exists := result.Get("content"); exists {
			fmt.Println(content)
		}
	}

	// Example 2: Execute a simple query
	fmt.Println("\n=== Example 2: Query Current User ===")
	state = domain.NewState()
	// Set GitHub API key in state for authentication
	if githubToken != "" {
		state.Set("github_token", githubToken)
	}
	state.Set("user_input", `Use the api_client tool to query the current authenticated user from GitHub GraphQL API. 
Get their login, name, email, and bio using this query:
query {
  viewer {
    login
    name
    email
    bio
  }
}`)

	result, err = agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error querying user: %v", err)
	} else {
		if content, exists := result.Get("content"); exists {
			fmt.Println(content)
		}
	}

	// Example 3: Query with variables
	fmt.Println("\n=== Example 3: Query Repository with Variables ===")
	state = domain.NewState()
	// Set GitHub API key in state for authentication
	if githubToken != "" {
		state.Set("github_token", githubToken)
	}
	state.Set("user_input", `Use the api_client tool to query information about the golang/go repository using variables. 
Use this query:
query GetRepo($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    name
    description
    stargazerCount
    forkCount
    primaryLanguage {
      name
    }
  }
}

With variables:
{
  "owner": "golang",
  "name": "go"
}`)

	result, err = agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error querying repository: %v", err)
	} else {
		if content, exists := result.Get("content"); exists {
			fmt.Println(content)
		}
	}

	// Example 4: Public GraphQL API (no auth required)
	fmt.Println("\n=== Example 4: Countries GraphQL API ===")
	state = domain.NewState()
	state.Set("user_input", `Use the api_client tool to discover and query the Countries GraphQL API at https://countries.trevorblades.com/graphql. 
First discover what's available, then query for information about the United States (code: "US").`)

	result, err = agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error with Countries API: %v", err)
	} else {
		if content, exists := result.Get("content"); exists {
			fmt.Println(content)
		}
	}

	// Example 5: Error handling
	fmt.Println("\n=== Example 5: GraphQL Error Handling ===")
	state = domain.NewState()
	// Set GitHub API key in state for authentication
	if githubToken != "" {
		state.Set("github_token", githubToken)
	}
	state.Set("user_input", `Use the api_client tool to execute an invalid GraphQL query to see error handling:
query {
  viewer {
    invalidField
    anotherBadField
  }
}`)

	result, err = agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error demonstration: %v", err)
	} else {
		if content, exists := result.Get("content"); exists {
			fmt.Println(content)
		}
	}

	// Example 6: Complex nested query
	fmt.Println("\n=== Example 6: Complex Nested Query ===")
	state = domain.NewState()
	// Set GitHub API key in state for authentication
	if githubToken != "" {
		state.Set("github_token", githubToken)
	}
	state.Set("user_input", `Use the api_client tool to get the 5 most recent repositories for the current user with their languages:
query {
  viewer {
    login
    repositories(first: 5, orderBy: {field: CREATED_AT, direction: DESC}) {
      nodes {
        name
        description
        createdAt
        languages(first: 3) {
          nodes {
            name
            color
          }
        }
      }
    }
  }
}`)

	result, err = agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error with nested query: %v", err)
	} else {
		if content, exists := result.Get("content"); exists {
			fmt.Println(content)
		}
	}
}
