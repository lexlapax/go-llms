package main

import (
	"os"
	"testing"
)

func TestGetAPIKey(t *testing.T) {
	// Save old environment variables
	oldOpenAIKey := os.Getenv("OPENAI_API_KEY")
	oldAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	// Cleanup function to restore environment
	defer func() {
		os.Setenv("OPENAI_API_KEY", oldOpenAIKey)
		os.Setenv("ANTHROPIC_API_KEY", oldAnthropicKey)
		// Reset config
		config = Config{}
		InitOptimizedConfig("")
	}()

	// Clear environment variables for the test
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Test case 1: Get API key from environment variable - OpenAI
	t.Run("GetFromEnvOpenAI", func(t *testing.T) {
		// Set environment variable
		os.Setenv("OPENAI_API_KEY", "test-openai-key")
		// Reset and reload config
		config = Config{}
		InitOptimizedConfig("")
		
		key, err := GetOptimizedAPIKey("openai")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if key != "test-openai-key" {
			t.Errorf("Expected 'test-openai-key', got '%s'", key)
		}
	})

	// Test case 2: Get API key from environment variable - Anthropic
	t.Run("GetFromEnvAnthropic", func(t *testing.T) {
		// Set environment variable
		os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
		// Reset and reload config
		config = Config{}
		InitOptimizedConfig("")
		
		key, err := GetOptimizedAPIKey("anthropic")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if key != "test-anthropic-key" {
			t.Errorf("Expected 'test-anthropic-key', got '%s'", key)
		}
	})

	// Test case 3: Get API key from config - OpenAI
	t.Run("GetFromConfigOpenAI", func(t *testing.T) {
		// Clear env var
		os.Unsetenv("OPENAI_API_KEY")
		// Set config directly
		config = Config{}
		config.Providers.OpenAI.APIKey = "config-openai-key"
		
		key, err := GetOptimizedAPIKey("openai")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if key != "config-openai-key" {
			t.Errorf("Expected 'config-openai-key', got '%s'", key)
		}
	})

	// Test case 4: Error when no API key is configured
	t.Run("ErrorWhenNoAPIKey", func(t *testing.T) {
		// Clear everything
		os.Unsetenv("OPENAI_API_KEY")
		config = Config{}
		InitOptimizedConfig("")
		
		_, err := GetOptimizedAPIKey("openai")
		if err == nil {
			t.Error("Expected error when no API key is configured")
		}
	})
}

func TestGetProvider(t *testing.T) {
	// Test case 1: Default provider and model
	t.Run("DefaultProvider", func(t *testing.T) {
		config = Config{}
		InitOptimizedConfig("")
		
		provider, model, err := GetOptimizedProvider()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if provider != "openai" {
			t.Errorf("Expected 'openai', got '%s'", provider)
		}
		if model != "gpt-4o" {
			t.Errorf("Expected 'gpt-4o', got '%s'", model)
		}
	})

	// Test case 2: Override provider and model
	t.Run("OverrideProvider", func(t *testing.T) {
		config = Config{
			Provider: "anthropic",
			Model:    "custom-model",
		}
		
		provider, model, err := GetOptimizedProvider()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if provider != "anthropic" {
			t.Errorf("Expected 'anthropic', got '%s'", provider)
		}
		if model != "custom-model" {
			t.Errorf("Expected 'custom-model', got '%s'", model)
		}
	})
}