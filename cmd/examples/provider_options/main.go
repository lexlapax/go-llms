package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
	// Check if any API keys are available
	openAIKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if openAIKey == "" && anthropicKey == "" && geminiKey == "" {
		fmt.Println("No API keys found in environment variables.")
		fmt.Println("Please set at least one of: OPENAI_API_KEY, ANTHROPIC_API_KEY, GEMINI_API_KEY")
		os.Exit(1)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Demonstrate common options
	fmt.Println("=== Common Provider Options ===")
	demoCommonOptions(ctx, openAIKey, anthropicKey, geminiKey)

	// Demonstrate provider-specific options
	fmt.Println("\n=== Provider-Specific Options ===")
	demoProviderSpecificOptions(ctx, openAIKey, anthropicKey, geminiKey)

	// Demonstrate options with environment variables
	fmt.Println("\n=== Options with Environment Variables ===")
	demoOptionsWithEnv()

	// Demonstrate the ModelConfig approach
	fmt.Println("\n=== Using ModelConfig ===")
	demoModelConfig(ctx, openAIKey, anthropicKey, geminiKey)
}

// demoCommonOptions demonstrates options that work across all providers
func demoCommonOptions(ctx context.Context, openAIKey, anthropicKey, geminiKey string) {
	// Create common options
	fmt.Println("Creating common options:")
	fmt.Println("- Custom HTTP client with 15s timeout")
	fmt.Println("- Custom headers")

	// Create a custom HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	httpClientOption := domain.NewHTTPClientOption(httpClient)

	// Create custom headers
	headers := map[string]string{
		"X-Custom-Header": "custom-value",
	}
	headersOption := domain.NewHeadersOption(headers)

	// Try to create providers with the common options
	if openAIKey != "" {
		fmt.Println("\nCreating OpenAI provider with common options...")
		openaiProvider := provider.NewOpenAIProvider(openAIKey, "gpt-4o", httpClientOption, headersOption)
		
		// Use the provider
		response, err := openaiProvider.Generate(ctx, "Say hello!")
		if err != nil {
			fmt.Printf("OpenAI error: %v\n", err)
		} else {
			fmt.Printf("OpenAI response: %s\n", response)
		}
	}

	if anthropicKey != "" {
		fmt.Println("\nCreating Anthropic provider with common options...")
		anthropicProvider := provider.NewAnthropicProvider(anthropicKey, "claude-3-5-sonnet-latest", httpClientOption, headersOption)
		
		// Use the provider
		response, err := anthropicProvider.Generate(ctx, "Say hello!")
		if err != nil {
			fmt.Printf("Anthropic error: %v\n", err)
		} else {
			fmt.Printf("Anthropic response: %s\n", response)
		}
	}

	if geminiKey != "" {
		fmt.Println("\nCreating Gemini provider with common options...")
		geminiProvider := provider.NewGeminiProvider(geminiKey, "gemini-2.0-flash-lite", httpClientOption, headersOption)
		
		// Use the provider
		response, err := geminiProvider.Generate(ctx, "Say hello!")
		if err != nil {
			fmt.Printf("Gemini error: %v\n", err)
		} else {
			fmt.Printf("Gemini response: %s\n", response)
		}
	}
}

// demoProviderSpecificOptions demonstrates options specific to each provider
func demoProviderSpecificOptions(ctx context.Context, openAIKey, anthropicKey, geminiKey string) {
	// OpenAI-specific options
	if openAIKey != "" {
		fmt.Println("\nCreating OpenAI provider with specific options:")
		fmt.Println("- Organization option")
		fmt.Println("- Logit bias option")

		// Usually you would use your actual org ID
		orgOption := domain.NewOpenAIOrganizationOption("org-demo")
		
		// Discourage the token for newline
		logitBiasOption := domain.NewOpenAILogitBiasOption(map[string]float64{
			"50256": -100,
		})

		// Create the provider with specific options
		openaiProvider := provider.NewOpenAIProvider(openAIKey, "gpt-4o", orgOption, logitBiasOption)
		
		// Use the provider
		response, err := openaiProvider.Generate(ctx, "Say hello in one word!")
		if err != nil {
			fmt.Printf("OpenAI error: %v\n", err)
		} else {
			fmt.Printf("OpenAI response: %s\n", response)
		}
	}

	// Anthropic-specific options
	if anthropicKey != "" {
		fmt.Println("\nCreating Anthropic provider with specific options:")
		fmt.Println("- System prompt option")
		fmt.Println("- Metadata option")

		systemPromptOption := domain.NewAnthropicSystemPromptOption(
			"You are a helpful assistant who speaks in a very concise way.")
		
		metadataOption := domain.NewAnthropicMetadataOption(map[string]string{
			"user_id":    "user123",
			"session_id": "session456",
		})

		// Create the provider with specific options
		anthropicProvider := provider.NewAnthropicProvider(anthropicKey, "claude-3-5-sonnet-latest", 
			systemPromptOption, metadataOption)
		
		// Use the provider
		response, err := anthropicProvider.Generate(ctx, "Say hello!")
		if err != nil {
			fmt.Printf("Anthropic error: %v\n", err)
		} else {
			fmt.Printf("Anthropic response: %s\n", response)
		}
	}

	// Gemini-specific options
	if geminiKey != "" {
		fmt.Println("\nCreating Gemini provider with specific options:")
		fmt.Println("- Generation config option (topK)")
		fmt.Println("- Safety settings option")

		// Generation config with topK
		generationConfigOption := domain.NewGeminiGenerationConfigOption().WithTopK(20)
		
		// Safety settings option
		safetySettings := []map[string]interface{}{
			{
				"category":  "HARM_CATEGORY_HARASSMENT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
		}
		safetySettingsOption := domain.NewGeminiSafetySettingsOption(safetySettings)

		// Create the provider with specific options
		geminiProvider := provider.NewGeminiProvider(geminiKey, "gemini-2.0-flash-lite", 
			generationConfigOption, safetySettingsOption)
		
		// Use the provider
		response, err := geminiProvider.Generate(ctx, "Say hello!")
		if err != nil {
			fmt.Printf("Gemini error: %v\n", err)
		} else {
			fmt.Printf("Gemini response: %s\n", response)
		}
	}
}

// demoOptionsWithEnv demonstrates creating options based on environment variables
func demoOptionsWithEnv() {
	// Get a provider using the environment variables
	_, providerName, modelName, err := llmutil.ProviderFromEnv()
	if err != nil {
		fmt.Printf("Error creating provider from environment: %v\n", err)
		return
	}

	fmt.Printf("Created %s provider using model %s from environment variables\n",
		providerName, modelName)
	
	// The ProviderFromEnv function automatically applies options based on environment variables:
	// - OPENAI_BASE_URL, ANTHROPIC_BASE_URL, GEMINI_BASE_URL for custom base URLs
	
	// You could add more environment-based options by extending ProviderFromEnv
}

// demoModelConfig demonstrates using the ModelConfig approach
func demoModelConfig(ctx context.Context, openAIKey, anthropicKey, geminiKey string) {
	// Example using OpenAI with ModelConfig
	if openAIKey != "" {
		fmt.Println("\nCreating OpenAI provider with ModelConfig:")
		
		// Create a ModelConfig
		config := llmutil.ModelConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			APIKey:   openAIKey,
			BaseURL:  "", // Use default
		}

		// Create provider from config
		llmProvider, err := llmutil.CreateProvider(config)
		if err != nil {
			fmt.Printf("Error creating provider: %v\n", err)
			return
		}

		// Use the provider
		response, err := llmProvider.Generate(ctx, "Say hello!")
		if err != nil {
			fmt.Printf("Provider error: %v\n", err)
		} else {
			fmt.Printf("Provider response: %s\n", response)
		}
	}
}