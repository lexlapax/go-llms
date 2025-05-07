package integration

import (
	"context"
	"os"
	"strings"
	"testing"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestProviderIntegration tests the provider integration
// These tests require actual API keys and are skipped by default
func TestProviderIntegration(t *testing.T) {
	// Test OpenAI provider
	t.Run("OpenAI", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable not set, skipping test")
		}

		// Create an OpenAI provider
		openai := provider.NewOpenAIProvider(apiKey, "gpt-3.5-turbo")

		// Test simple generation
		t.Run("Generate", func(t *testing.T) {
			resp, err := openai.Generate(context.Background(), "Hello, how are you?")
			if err != nil {
				t.Fatalf("OpenAI Generate failed: %v", err)
			}

			if resp == "" {
				t.Errorf("Expected non-empty response, got empty string")
			}
		})

		// Test generation with messages
		t.Run("GenerateMessage", func(t *testing.T) {
			messages := []ldomain.Message{
				{Role: ldomain.RoleSystem, Content: "You are a helpful assistant."},
				{Role: ldomain.RoleUser, Content: "What is the capital of France?"},
			}

			resp, err := openai.GenerateMessage(context.Background(), messages)
			if err != nil {
				t.Fatalf("OpenAI GenerateMessage failed: %v", err)
			}

			if resp.Content == "" {
				t.Errorf("Expected non-empty response content, got empty string")
			}

			// Check that the response mentions Paris
			if !strings.Contains(resp.Content, "Paris") {
				t.Errorf("Expected response to mention 'Paris', got: %s", resp.Content)
			}
		})

		// Test generation with schema
		t.Run("GenerateWithSchema", func(t *testing.T) {
			schema := &sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"capital": {
						Type:        "string",
						Description: "The capital city",
					},
					"country": {
						Type:        "string",
						Description: "The country",
					},
				},
				Required: []string{"capital", "country"},
			}

			resp, err := openai.GenerateWithSchema(
				context.Background(),
				"What is the capital of France?",
				schema,
			)
			if err != nil {
				t.Fatalf("OpenAI GenerateWithSchema failed: %v", err)
			}

			// Check the response structure
			data, ok := resp.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected map[string]interface{}, got: %T", resp)
			}

			capital, ok := data["capital"].(string)
			if !ok || capital != "Paris" {
				t.Errorf("Expected capital 'Paris', got: %v", data["capital"])
			}

			country, ok := data["country"].(string)
			if !ok || country != "France" {
				t.Errorf("Expected country 'France', got: %v", data["country"])
			}
		})
	})

	// Test Anthropic provider
	t.Run("Anthropic", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			t.Skip("ANTHROPIC_API_KEY environment variable not set, skipping test")
		}

		// Create an Anthropic provider
		anthropic := provider.NewAnthropicProvider(apiKey, "claude-3-sonnet-20240229")

		// Test simple generation
		t.Run("Generate", func(t *testing.T) {
			resp, err := anthropic.Generate(context.Background(), "Hello, how are you?")
			if err != nil {
				t.Fatalf("Anthropic Generate failed: %v", err)
			}

			if resp == "" {
				t.Errorf("Expected non-empty response, got empty string")
			}
		})

		// Test generation with messages
		t.Run("GenerateMessage", func(t *testing.T) {
			messages := []ldomain.Message{
				{Role: ldomain.RoleSystem, Content: "You are a helpful assistant."},
				{Role: ldomain.RoleUser, Content: "What is the capital of France?"},
			}

			resp, err := anthropic.GenerateMessage(context.Background(), messages)
			if err != nil {
				t.Fatalf("Anthropic GenerateMessage failed: %v", err)
			}

			if resp.Content == "" {
				t.Errorf("Expected non-empty response content, got empty string")
			}

			// Check that the response mentions Paris
			if !strings.Contains(resp.Content, "Paris") {
				t.Errorf("Expected response to mention 'Paris', got: %s", resp.Content)
			}
		})

		// Test generation with schema
		t.Run("GenerateWithSchema", func(t *testing.T) {
			schema := &sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"capital": {
						Type:        "string",
						Description: "The capital city",
					},
					"country": {
						Type:        "string",
						Description: "The country",
					},
				},
				Required: []string{"capital", "country"},
			}

			resp, err := anthropic.GenerateWithSchema(
				context.Background(),
				"What is the capital of France?",
				schema,
			)
			if err != nil {
				t.Fatalf("Anthropic GenerateWithSchema failed: %v", err)
			}

			// Check the response structure
			data, ok := resp.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected map[string]interface{}, got: %T", resp)
			}

			capital, ok := data["capital"].(string)
			if !ok || capital != "Paris" {
				t.Errorf("Expected capital 'Paris', got: %v", data["capital"])
			}

			country, ok := data["country"].(string)
			if !ok || country != "France" {
				t.Errorf("Expected country 'France', got: %v", data["country"])
			}
		})
	})
}
