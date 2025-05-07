package main

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestGetAPIKey(t *testing.T) {
	// Save old environment variables
	oldOpenAIKey := os.Getenv("OPENAI_API_KEY")
	oldAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	
	// Cleanup function to restore environment
	defer func() {
		os.Setenv("OPENAI_API_KEY", oldOpenAIKey)
		os.Setenv("ANTHROPIC_API_KEY", oldAnthropicKey)
		viper.Reset()
	}()
	
	// Clear environment variables for the test
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	
	// Test case 1: Get API key from environment variable - OpenAI
	t.Run("GetFromEnvOpenAI", func(t *testing.T) {
		// Set environment variable
		os.Setenv("OPENAI_API_KEY", "test-openai-key")
		
		// Get API key
		key, err := getAPIKey("openai")
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
		
		// Get API key
		key, err := getAPIKey("anthropic")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if key != "test-anthropic-key" {
			t.Errorf("Expected API key 'test-anthropic-key', got: %s", key)
		}
	})
	
	// Test case 3: Get API key from viper config
	t.Run("GetFromConfig", func(t *testing.T) {
		// Set config value
		viper.Set("providers.openai.api_key", "config-openai-key")
		
		// Get API key
		key, err := getAPIKey("openai")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if key != "config-openai-key" {
			t.Errorf("Expected API key 'config-openai-key', got: %s", key)
		}
	})
	
	// Test case 4: No API key configured
	t.Run("NoAPIKey", func(t *testing.T) {
		// Clear environment variables and config
		os.Unsetenv("OPENAI_API_KEY")
		viper.Set("providers.openai.api_key", "")
		
		// Try to get API key
		_, err := getAPIKey("openai")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}