package fetchers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenRouterFetcher_FetchModels(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "/api/v1/models", r.URL.Path)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": [
				{
					"id": "openai/gpt-4",
					"name": "GPT-4",
					"created": 1680000000,
					"description": "OpenAI's GPT-4 model",
					"context_length": 8192,
					"pricing": {
						"prompt": "0.00003",
						"completion": "0.00006",
						"request": "0",
						"image": "0"
					},
					"architecture": {
						"modality": "text->text",
						"tokenizer": "cl100k_base",
						"instruct_type": "none"
					},
					"top_provider": {
						"max_completion_tokens": 4096,
						"is_moderated": false
					}
				},
				{
					"id": "anthropic/claude-3-opus",
					"name": "Claude 3 Opus",
					"created": 1700000000,
					"description": "Anthropic's most capable model",
					"context_length": 200000,
					"pricing": {
						"prompt": "0.000015",
						"completion": "0.000075",
						"request": "0",
						"image": "0"
					},
					"architecture": {
						"modality": "text+image->text",
						"tokenizer": "claude",
						"instruct_type": "claude"
					},
					"top_provider": {
						"max_completion_tokens": 4096,
						"is_moderated": true
					}
				},
				{
					"id": "huggingface/zephyr-7b-beta:free",
					"name": "Zephyr 7B Beta (Free)",
					"created": 1690000000,
					"description": "A free model from Hugging Face",
					"context_length": 4096,
					"pricing": {
						"prompt": "0",
						"completion": "0",
						"request": "0",
						"image": "0"
					},
					"architecture": {
						"modality": "text->text",
						"tokenizer": "llama",
						"instruct_type": "alpaca"
					},
					"top_provider": {
						"max_completion_tokens": 2048,
						"is_moderated": false
					}
				}
			]
		}`))
	}))
	defer server.Close()

	// Create fetcher with test server URL
	fetcher := &OpenRouterFetcher{
		BaseURL: server.URL + "/api/v1",
		APIKey:  "test-key",
	}

	// Fetch models
	models, err := fetcher.FetchModels()
	require.NoError(t, err)
	require.Len(t, models, 3)

	// Check first model (GPT-4)
	gpt4 := models[0]
	assert.Equal(t, "openai/gpt-4", gpt4.Name)
	assert.Equal(t, "openrouter:openai", gpt4.Provider)
	assert.Equal(t, 8192, gpt4.ContextWindow)
	assert.Equal(t, 4096, gpt4.MaxOutputTokens)
	assert.Equal(t, "OpenAI's GPT-4 model", gpt4.Description)
	assert.Equal(t, "2023-03-28", gpt4.LastUpdated)
	assert.True(t, gpt4.Capabilities.Text.Read)
	assert.True(t, gpt4.Capabilities.Text.Write)
	assert.True(t, gpt4.Capabilities.Streaming)

	// Check second model (Claude 3 Opus)
	claude := models[1]
	assert.Equal(t, "anthropic/claude-3-opus", claude.Name)
	assert.Equal(t, "openrouter:anthropic", claude.Provider)
	assert.Equal(t, 200000, claude.ContextWindow)
	assert.Equal(t, 4096, claude.MaxOutputTokens)
	assert.True(t, claude.Capabilities.Image.Read)
	assert.True(t, claude.Capabilities.Text.Read)
	assert.True(t, claude.Capabilities.Text.Write)
	assert.True(t, claude.Capabilities.Streaming)

	// Check third model (Free model)
	free := models[2]
	assert.Equal(t, "huggingface/zephyr-7b-beta:free", free.Name)
	assert.Equal(t, "openrouter:huggingface", free.Provider)
	assert.Equal(t, 4096, free.ContextWindow)
	assert.Equal(t, 2048, free.MaxOutputTokens)
	assert.True(t, free.Capabilities.Text.Read)
	assert.True(t, free.Capabilities.Text.Write)
	assert.True(t, free.Capabilities.Streaming)
}

func TestOpenRouterFetcher_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		serverFunc    func(w http.ResponseWriter, r *http.Request)
		expectedError string
	}{
		{
			name: "non-200 status",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error": "Invalid API key"}`))
			},
			expectedError: "unexpected status code 401",
		},
		{
			name: "invalid JSON",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`invalid json`))
			},
			expectedError: "parsing JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			fetcher := &OpenRouterFetcher{
				BaseURL: server.URL + "/api/v1",
				APIKey:  "test-key",
			}

			_, err := fetcher.FetchModels()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}
