package fetchers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOllamaFetcher(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		fetcher := NewOllamaFetcher("", nil)
		assert.Equal(t, defaultOllamaBaseURL, fetcher.BaseURL)
		assert.Equal(t, http.DefaultClient, fetcher.HTTPClient)
	})

	t.Run("CustomValues", func(t *testing.T) {
		customURL := "http://custom:11434"
		customClient := &http.Client{Timeout: 30 * time.Second}
		fetcher := NewOllamaFetcher(customURL, customClient)
		assert.Equal(t, customURL, fetcher.BaseURL)
		assert.Equal(t, customClient, fetcher.HTTPClient)
	})

	t.Run("EnvironmentVariable", func(t *testing.T) {
		// Save original value
		originalHost := os.Getenv("OLLAMA_HOST")
		defer func() {
			if originalHost != "" {
				_ = os.Setenv("OLLAMA_HOST", originalHost)
			} else {
				_ = os.Unsetenv("OLLAMA_HOST")
			}
		}()

		// Set custom environment variable
		customHost := "http://env-host:11434"
		_ = os.Setenv("OLLAMA_HOST", customHost)

		fetcher := NewOllamaFetcher("", nil)
		assert.Equal(t, customHost, fetcher.BaseURL)
	})
}

func TestOllamaFetcherFetchModels(t *testing.T) {
	t.Run("SuccessfulFetch", func(t *testing.T) {
		// Create mock response
		mockResponse := OllamaAPIResponse{
			Models: []OllamaModel{
				{
					Name:       "llama3.2:3b",
					Model:      "llama3.2:3b",
					ModifiedAt: time.Now(),
					Size:       2019393189,
					Digest:     "a80c4f17acd55265feec403c7aef86be0c25983ab279d83f3bcd3abbcb5b8b72",
					Details: OllamaModelDetails{
						ParentModel:       "llama3.2",
						Format:            "gguf",
						Family:            "llama",
						Families:          []string{"llama"},
						ParameterSize:     "3B",
						QuantizationLevel: "Q4_K_M",
					},
				},
				{
					Name:       "mistral:7b",
					Model:      "mistral:7b",
					ModifiedAt: time.Now(),
					Size:       4113858977,
					Digest:     "61e88e884507ba5e06c49b40e6226884b2a16e872382c2b44a42f2d119d804a5",
					Details: OllamaModelDetails{
						ParentModel:       "mistral",
						Format:            "gguf",
						Family:            "mistral",
						Families:          []string{"mistral"},
						ParameterSize:     "7B",
						QuantizationLevel: "Q4_0",
					},
				},
			},
		}

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/tags", r.URL.Path)
			assert.Equal(t, "GET", r.Method)

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		// Create fetcher with test server URL
		fetcher := NewOllamaFetcher(server.URL, nil)
		models, err := fetcher.FetchModels()

		require.NoError(t, err)
		assert.Len(t, models, 2)

		// Check first model
		assert.Equal(t, "ollama", models[0].Provider)
		assert.Equal(t, "llama3.2:3b", models[0].Name)
		assert.Equal(t, "llama", models[0].ModelFamily)
		assert.Equal(t, 8192, models[0].ContextWindow)
		assert.True(t, models[0].Capabilities.Text.Read)
		assert.True(t, models[0].Capabilities.Text.Write)
		assert.True(t, models[0].Capabilities.Streaming)
		assert.True(t, models[0].Capabilities.FunctionCalling)
		assert.Equal(t, 0.0, models[0].Pricing.InputPer1kTokens)
		assert.Equal(t, 0.0, models[0].Pricing.OutputPer1kTokens)

		// Check second model
		assert.Equal(t, "mistral:7b", models[1].Name)
		assert.Equal(t, "mistral", models[1].ModelFamily)
		assert.Equal(t, 8192, models[1].ContextWindow)
	})

	t.Run("ServerError", func(t *testing.T) {
		// Create test server that returns an error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		fetcher := NewOllamaFetcher(server.URL, nil)
		models, err := fetcher.FetchModels()

		assert.Error(t, err)
		assert.Nil(t, models)
		assert.Contains(t, err.Error(), "status code: 500")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Create test server that returns invalid JSON
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		fetcher := NewOllamaFetcher(server.URL, nil)
		models, err := fetcher.FetchModels()

		assert.Error(t, err)
		assert.Nil(t, models)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("ConnectionError", func(t *testing.T) {
		// Use an invalid URL to simulate connection error
		fetcher := NewOllamaFetcher("http://invalid-host-that-does-not-exist:99999", nil)
		models, err := fetcher.FetchModels()

		assert.Error(t, err)
		assert.Nil(t, models)
	})
}

func TestEstimateContextWindow(t *testing.T) {
	fetcher := &OllamaFetcher{}

	testCases := []struct {
		name           string
		model          OllamaModel
		expectedWindow int
	}{
		{
			name: "Llama3",
			model: OllamaModel{
				Name: "llama3.2:3b",
			},
			expectedWindow: 8192,
		},
		{
			name: "Llama2",
			model: OllamaModel{
				Name: "llama2:7b",
			},
			expectedWindow: 4096,
		},
		{
			name: "Mistral7B",
			model: OllamaModel{
				Name: "mistral:7b",
			},
			expectedWindow: 8192,
		},
		{
			name: "Gemma",
			model: OllamaModel{
				Name: "gemma:2b",
			},
			expectedWindow: 8192,
		},
		{
			name: "Qwen2",
			model: OllamaModel{
				Name: "qwen2:7b",
			},
			expectedWindow: 32768,
		},
		{
			name: "Qwen",
			model: OllamaModel{
				Name: "qwen:4b",
			},
			expectedWindow: 8192,
		},
		{
			name: "Phi",
			model: OllamaModel{
				Name: "phi:2.7b",
			},
			expectedWindow: 4096,
		},
		{
			name: "CodeLlama",
			model: OllamaModel{
				Name: "codellama:13b",
			},
			expectedWindow: 16384,
		},
		{
			name: "Unknown",
			model: OllamaModel{
				Name: "unknown:model",
			},
			expectedWindow: 4096, // Default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			window := fetcher.estimateContextWindow(tc.model)
			assert.Equal(t, tc.expectedWindow, window)
		})
	}
}

func TestSupportsVision(t *testing.T) {
	fetcher := &OllamaFetcher{}

	testCases := []struct {
		name           string
		model          OllamaModel
		expectedVision bool
	}{
		{
			name: "LLaVA",
			model: OllamaModel{
				Name: "llava:7b",
			},
			expectedVision: true,
		},
		{
			name: "BakLLaVA",
			model: OllamaModel{
				Name: "bakllava:latest",
			},
			expectedVision: true,
		},
		{
			name: "Moondream",
			model: OllamaModel{
				Name: "moondream:1.8b",
			},
			expectedVision: true,
		},
		{
			name: "Llama3.2Vision",
			model: OllamaModel{
				Name: "llama3.2-vision:11b",
			},
			expectedVision: true,
		},
		{
			name: "MiniCPM-V",
			model: OllamaModel{
				Name: "minicpm-v:8b",
			},
			expectedVision: true,
		},
		{
			name: "RegularLlama",
			model: OllamaModel{
				Name: "llama3.2:3b",
			},
			expectedVision: false,
		},
		{
			name: "Mistral",
			model: OllamaModel{
				Name: "mistral:7b",
			},
			expectedVision: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasVision := fetcher.supportsVision(tc.model)
			assert.Equal(t, tc.expectedVision, hasVision)
		})
	}
}
