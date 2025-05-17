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

func TestChatCmdStreamingFlag(t *testing.T) {
	// Create a new chat command
	chatCmd := newChatCmd()

	// Verify the stream flag exists
	streamFlag := chatCmd.Flags().Lookup("stream")
	if streamFlag == nil {
		t.Fatal("Expected 'stream' flag to exist on chat command")
	}

	// Verify the no-stream flag exists
	noStreamFlag := chatCmd.Flags().Lookup("no-stream")
	if noStreamFlag == nil {
		t.Fatal("Expected 'no-stream' flag to exist on chat command")
	}

	// Check default values (should be false)
	defaultStreamValue, err := chatCmd.Flags().GetBool("stream")
	if err != nil {
		t.Fatalf("Error getting stream flag value: %v", err)
	}
	if defaultStreamValue {
		t.Errorf("Expected default stream flag value to be false, got true")
	}

	defaultNoStreamValue, err := chatCmd.Flags().GetBool("no-stream")
	if err != nil {
		t.Fatalf("Error getting no-stream flag value: %v", err)
	}
	if defaultNoStreamValue {
		t.Errorf("Expected default no-stream flag value to be false, got true")
	}

	// Test with stream flag explicitly set
	t.Run("WithStreamFlag", func(t *testing.T) {
		cmd := newChatCmd()
		args := []string{"--stream"}
		cmd.SetArgs(args)
		flags := cmd.Flags()

		// We need to manually parse the flags since we're not executing the command
		err = flags.Parse(args)
		if err != nil {
			t.Fatalf("Error parsing flags: %v", err)
		}

		// Check that the flag value was correctly set
		streamValue, err := flags.GetBool("stream")
		if err != nil {
			t.Fatalf("Error getting stream flag value: %v", err)
		}
		if !streamValue {
			t.Errorf("Expected stream flag value to be true after setting, got false")
		}
	})

	// Test with no-stream flag explicitly set
	t.Run("WithNoStreamFlag", func(t *testing.T) {
		cmd := newChatCmd()
		args := []string{"--no-stream"}
		cmd.SetArgs(args)
		flags := cmd.Flags()

		// We need to manually parse the flags since we're not executing the command
		err = flags.Parse(args)
		if err != nil {
			t.Fatalf("Error parsing flags: %v", err)
		}

		// Check that the flag value was correctly set
		noStreamValue, err := flags.GetBool("no-stream")
		if err != nil {
			t.Fatalf("Error getting no-stream flag value: %v", err)
		}
		if !noStreamValue {
			t.Errorf("Expected no-stream flag value to be true after setting, got false")
		}
	})
}
