// ABOUTME: Example demonstrating advanced authentication features of api_client tool
// ABOUTME: Shows OAuth2, custom headers, session management, and auto-auth detection

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
	log.Println("=== API Client Advanced Authentication Example ===")

	// Check if API key is set
	apiKey := llmutil.GetAPIKeyFromEnv("openai")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}
	modelName := os.Getenv("OPENAI_MODEL")
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	// Create LLM agent
	agent, err := core.NewAgentFromString("auth-demo", fmt.Sprintf("openai:%s:%s", apiKey, modelName))
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Enable debug logging if DEBUG=1
	if os.Getenv("DEBUG") == "1" {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		agent.WithHook(core.NewLoggingHook(logger, core.LogLevelDebug))
	}

	ctx := context.Background()

	// Example 1: OAuth2 Authentication
	log.Println("\n--- Example 1: OAuth2 Authentication ---")
	state1 := domain.NewState()
	state1.Set("oauth2_token", "ghp_example_github_token")
	state1.Set("user_input", `Use the api_client tool to fetch my GitHub user profile.
The base_url is https://api.github.com and the endpoint is /user.
Use OAuth2 authentication with the token from my state.`)

	result1, err := agent.Run(ctx, state1)
	if err != nil {
		log.Printf("Error with OAuth2 example: %v", err)
	} else {
		if output, exists := result1.Get("output"); exists {
			log.Printf("OAuth2 Response: %v", output)
		}
	}

	// Example 2: Custom Header Authentication
	log.Println("\n--- Example 2: Custom Header Authentication ---")
	state2 := domain.NewState()
	state2.Set("custom_api_key", "sk_custom_12345")
	state2.Set("user_input", `Use the api_client tool to make a request to https://api.custom.com/v1/data.
Use custom authentication with header name "X-Custom-Key" and the value from custom_api_key in my state.
Add the prefix "Bearer" to the header value.`)

	result2, err := agent.Run(ctx, state2)
	if err != nil {
		log.Printf("Error with custom header example: %v", err)
	} else {
		if output, exists := result2.Get("output"); exists {
			log.Printf("Custom Header Response: %v", output)
		}
	}

	// Example 3: Auto-detect Authentication from State
	log.Println("\n--- Example 3: Auto-detect Authentication ---")
	state3 := domain.NewState()
	state3.Set("github_token", "ghp_auto_detected_token")
	state3.Set("user_input", `Use the api_client tool to fetch the golang/go repository info from GitHub.
The base_url is https://api.github.com and the endpoint is /repos/golang/go.
Don't specify authentication - let it auto-detect from my state.`)

	result3, err := agent.Run(ctx, state3)
	if err != nil {
		log.Printf("Error with auto-detect example: %v", err)
	} else {
		if output, exists := result3.Get("output"); exists {
			log.Printf("Auto-detect Response: %v", output)
		}
	}

	// Example 4: API Key in Different Locations
	log.Println("\n--- Example 4: API Key in Query Parameter ---")
	state4 := domain.NewState()
	state4.Set("weather_api_key", "demo_weather_key")
	state4.Set("user_input", `Use the api_client tool to get weather data from https://api.weather.com/v1/current.
The endpoint needs a query parameter for the city (city=London).
Use API key authentication with the key in a query parameter named "apikey".
Use the weather_api_key from my state.`)

	result4, err := agent.Run(ctx, state4)
	if err != nil {
		log.Printf("Error with query param auth example: %v", err)
	} else {
		if output, exists := result4.Get("output"); exists {
			log.Printf("Query Param Auth Response: %v", output)
		}
	}

	// Example 5: Session/Cookie Management
	log.Println("\n--- Example 5: Session Management ---")
	state5 := domain.NewState()
	state5.Set("user_input", `First, use the api_client tool to login to https://api.sessiondemo.com/auth/login
with POST method and body {"username": "demo", "password": "pass123"}.
Enable session management with enable_session=true.

Then, make another request to https://api.sessiondemo.com/user/profile
with session management enabled to use the cookies from the login.`)

	result5, err := agent.Run(ctx, state5)
	if err != nil {
		log.Printf("Error with session example: %v", err)
	} else {
		if output, exists := result5.Get("output"); exists {
			log.Printf("Session Management Response: %v", output)
		}
	}

	// Example 6: OAuth2 Configuration
	log.Println("\n--- Example 6: OAuth2 Token Exchange ---")
	state6 := domain.NewState()
	state6.Set("oauth2_config", map[string]interface{}{
		"token_url":     "https://auth.example.com/oauth/token",
		"client_id":     "demo_client_id",
		"client_secret": "demo_client_secret",
		"flow":          "client_credentials",
		"scope":         "read write",
	})
	state6.Set("user_input", `I have OAuth2 configuration in my state. Can you explain how to use it
with the api_client tool to obtain an access token using the client credentials flow?
Don't actually make the request, just explain the parameters needed.`)

	result6, err := agent.Run(ctx, state6)
	if err != nil {
		log.Printf("Error with OAuth2 config example: %v", err)
	} else {
		if output, exists := result6.Get("output"); exists {
			log.Printf("OAuth2 Config Explanation: %v", output)
		}
	}

	// Example 7: Multiple Authentication Methods
	log.Println("\n--- Example 7: Trying Different Auth Methods ---")
	state7 := domain.NewState()
	state7.Set("api_key", "fallback_key_123")
	state7.Set("bearer_token", "fallback_bearer_456")
	state7.Set("user_input", `I need to access an API at https://api.multiauth.com/data but I'm not sure
which authentication method it uses. Can you try making a request and let the
auto-detection figure out the right auth method from my state?`)

	result7, err := agent.Run(ctx, state7)
	if err != nil {
		log.Printf("Error with multi-auth example: %v", err)
	} else {
		if output, exists := result7.Get("output"); exists {
			log.Printf("Multi-auth Response: %v", output)
		}
	}

	log.Println("\n=== Authentication Examples Complete ===")
	log.Println("\nKey features demonstrated:")
	log.Println("1. OAuth2 bearer token authentication")
	log.Println("2. Custom header authentication with prefix")
	log.Println("3. Automatic auth detection from state")
	log.Println("4. API key in query parameters")
	log.Println("5. Session/cookie management")
	log.Println("6. OAuth2 configuration for token exchange")
	log.Println("7. Fallback authentication methods")
}
