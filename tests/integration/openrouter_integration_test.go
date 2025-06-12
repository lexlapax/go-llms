package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIfNoOpenRouterKey(t *testing.T) {
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("Skipping OpenRouter integration test: OPENROUTER_API_KEY not set")
	}
}

func TestOpenRouterIntegration_BasicGeneration(t *testing.T) {
	skipIfNoOpenRouterKey(t)

	apiKey := os.Getenv("OPENROUTER_API_KEY")

	// Use a free model for testing
	llm := provider.NewOpenRouterProvider(apiKey, "huggingface/zephyr-7b-beta:free")

	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Say 'Hello from OpenRouter' and nothing else."),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(50))
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Contains(t, response.Content, "Hello from OpenRouter")
}

func TestOpenRouterIntegration_Streaming(t *testing.T) {
	skipIfNoOpenRouterKey(t)

	apiKey := os.Getenv("OPENROUTER_API_KEY")

	// Use a free model
	llm := provider.NewOpenRouterProvider(apiKey, "mistralai/mistral-7b-instruct:free")

	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Count from 1 to 5"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := llm.StreamMessage(ctx, messages, domain.WithMaxTokens(50))
	require.NoError(t, err)
	require.NotNil(t, stream)
	var fullContent string
	chunkCount := 0

	for token := range stream {
		if token.Text != "" {
			fullContent += token.Text
			chunkCount++
		}
	}

	assert.Greater(t, chunkCount, 1, "Expected multiple chunks in streaming response")
	assert.NotEmpty(t, fullContent)
	t.Logf("Received %d chunks, full content: %s", chunkCount, fullContent)
}

func TestOpenRouterIntegration_DifferentModels(t *testing.T) {
	skipIfNoOpenRouterKey(t)

	apiKey := os.Getenv("OPENROUTER_API_KEY")

	// Test different types of models
	testCases := []struct {
		name    string
		model   string
		canSkip bool // Some models might not be available
	}{
		{
			name:    "Free Hugging Face Model",
			model:   "huggingface/zephyr-7b-beta:free",
			canSkip: false,
		},
		{
			name:    "OpenAI Model",
			model:   "openai/gpt-3.5-turbo",
			canSkip: true, // Requires credits
		},
		{
			name:    "Mistral Free Model",
			model:   "mistralai/mistral-7b-instruct:free",
			canSkip: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llm := provider.NewOpenRouterProvider(apiKey, tc.model)

			messages := []domain.Message{
				domain.NewTextMessage(domain.RoleUser, "What is 2+2?"),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(20))

			if err != nil && tc.canSkip {
				t.Skipf("Model %s not available or requires credits: %v", tc.model, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Contains(t, response.Content, "4")
		})
	}
}

func TestOpenRouterIntegration_ErrorHandling(t *testing.T) {
	skipIfNoOpenRouterKey(t)

	testCases := []struct {
		name          string
		apiKey        string
		model         string
		expectedError string
	}{
		{
			name:          "Invalid API Key",
			apiKey:        "invalid-key",
			model:         "huggingface/zephyr-7b-beta:free",
			expectedError: "401",
		},
		{
			name:          "Invalid Model",
			apiKey:        os.Getenv("OPENROUTER_API_KEY"),
			model:         "nonexistent/model",
			expectedError: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llm := provider.NewOpenRouterProvider(tc.apiKey, tc.model)

			messages := []domain.Message{
				domain.NewTextMessage(domain.RoleUser, "Test"),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := llm.GenerateMessage(ctx, messages)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestOpenRouterIntegration_CustomBaseURL(t *testing.T) {
	skipIfNoOpenRouterKey(t)

	// Test with environment variable override
	customURL := "https://openrouter.ai/api/v1" // Same as default, but tests the mechanism
	_ = os.Setenv("OPENROUTER_API_BASE", customURL)
	defer func() { _ = os.Unsetenv("OPENROUTER_API_BASE") }()

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	llm := provider.NewOpenRouterProvider(apiKey, "huggingface/zephyr-7b-beta:free")

	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Say 'Custom URL works'"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(50))
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Contains(t, response.Content, "Custom URL works")
}
