package main

import (
	"os"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/providers/structs"
)

func TestGetAPIKey(t *testing.T) {
	// Save old environment variables
	oldOpenAIKey := os.Getenv("OPENAI_API_KEY")
	oldAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	// Cleanup function to restore environment
	defer func() {
		os.Setenv("OPENAI_API_KEY", oldOpenAIKey)
		os.Setenv("ANTHROPIC_API_KEY", oldAnthropicKey)
		// Reset koanf instance
		k = koanf.New(".")
	}()

	// Clear environment variables for the test
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Test case 1: Get API key from environment variable - OpenAI
	t.Run("GetFromEnvOpenAI", func(t *testing.T) {
		// Set environment variable
		os.Setenv("OPENAI_API_KEY", "test-openai-key")

		// Reset koanf instance
		k = koanf.New(".")
		if err := InitConfig(""); err != nil {
			t.Fatalf("Failed to init config: %v", err)
		}

		// Get API key
		key, err := GetAPIKey("openai")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if key != "test-openai-key" {
			t.Errorf("Expected API key 'test-openai-key', got: %s", key)
		}
	})

	// Test case 2: Get API key from environment variable - Anthropic
	t.Run("GetFromEnvAnthropic", func(t *testing.T) {
		// Set environment variable
		os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")

		// Reset koanf instance
		k = koanf.New(".")
		if err := InitConfig(""); err != nil {
			t.Fatalf("Failed to init config: %v", err)
		}

		// Get API key
		key, err := GetAPIKey("anthropic")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if key != "test-anthropic-key" {
			t.Errorf("Expected API key 'test-anthropic-key', got: %s", key)
		}
	})

	// Test case 3: Get API key from config - OpenAI
	t.Run("GetFromConfigOpenAI", func(t *testing.T) {
		// Clear environment variable
		os.Unsetenv("OPENAI_API_KEY")

		// Reset koanf and set config
		k = koanf.New(".")
		config := DefaultConfig()
		config.Providers.OpenAI.APIKey = "config-openai-key"
		
		// Load the config
		if err := k.Load(structs.Provider(config, "koanf"), nil); err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Get API key
		key, err := GetAPIKey("openai")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if key != "config-openai-key" {
			t.Errorf("Expected API key 'config-openai-key', got: %s", key)
		}
	})

	// Test case 4: Error when no API key available
	t.Run("ErrorWhenNoAPIKey", func(t *testing.T) {
		// Clear all environment variables
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")

		// Reset koanf instance with empty config
		k = koanf.New(".")
		if err := InitConfig(""); err != nil {
			t.Fatalf("Failed to init config: %v", err)
		}

		// Try to get API key - should fail
		_, err := GetAPIKey("openai")
		if err == nil {
			t.Error("Expected error when no API key available, got nil")
		}
	})
}

// Test the GetProvider function
func TestGetProvider(t *testing.T) {
	// Save old config
	originalK := k
	defer func() {
		k = originalK
	}()

	t.Run("DefaultProvider", func(t *testing.T) {
		k = koanf.New(".")
		config := DefaultConfig()
		if err := k.Load(structs.Provider(config, "koanf"), nil); err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		ctx := &Context{
			Config: &config,
			CLI:    &CLI{},
		}

		provider, model, err := ctx.GetProviderInfo()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if provider != "openai" {
			t.Errorf("Expected provider 'openai', got: %s", provider)
		}

		if model != "gpt-4o" {
			t.Errorf("Expected model 'gpt-4o', got: %s", model)
		}
	})

	t.Run("OverrideProvider", func(t *testing.T) {
		k = koanf.New(".")
		config := DefaultConfig()
		if err := k.Load(structs.Provider(config, "koanf"), nil); err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		ctx := &Context{
			Config: &config,
			CLI: &CLI{
				Provider: "anthropic",
				Model:    "claude-3",
			},
		}

		provider, model, err := ctx.GetProviderInfo()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if provider != "anthropic" {
			t.Errorf("Expected provider 'anthropic', got: %s", provider)
		}

		if model != "claude-3" {
			t.Errorf("Expected model 'claude-3', got: %s", model)
		}
	})
}