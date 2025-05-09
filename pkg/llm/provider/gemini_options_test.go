package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestGeminiProviderOptions(t *testing.T) {
	// Create a mock HTTP server to simulate Gemini API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// These would be used for more complex validation in a real integration test
		// Just extract and discard for now to avoid unused variable warnings
		_, _ = requestBody["generationConfig"].(map[string]interface{})
		_, _ = requestBody["safetySettings"].([]interface{})

		// Create sample response
		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{
								"text": "This is a test response.",
							},
						},
					},
					"finishReason": "STOP",
				},
			},
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Error encoding response: %v", err)
			return
		}
	}))
	defer mockServer.Close()

	t.Run("BaseURLOption", func(t *testing.T) {
		baseURLOption := domain.NewBaseURLOption(mockServer.URL)
		provider := NewGeminiProvider("test-key", "gemini-2.0-flash-lite", baseURLOption)

		if provider.baseURL != mockServer.URL {
			t.Errorf("Expected baseURL to be %s, got %s", mockServer.URL, provider.baseURL)
		}
	})

	t.Run("HTTPClientOption", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30 * time.Second}
		clientOption := domain.NewHTTPClientOption(customClient)
		provider := NewGeminiProvider("test-key", "gemini-2.0-flash-lite", clientOption)

		if provider.httpClient != customClient {
			t.Error("HTTPClientOption failed to set the HTTP client")
		}
	})

	t.Run("GeminiGenerationConfigOption", func(t *testing.T) {
		customTopK := 25
		generationConfigOption := domain.NewGeminiGenerationConfigOption().WithTopK(customTopK)
		provider := NewGeminiProvider("test-key", "gemini-2.0-flash-lite", generationConfigOption)

		// Check that the topK field was updated
		if provider.topK != customTopK {
			t.Errorf("Expected topK to be %d, got %d", customTopK, provider.topK)
		}

		// For a proper end-to-end test, we would need to capture and verify the actual API request
		// This would be done in an integration test with the mock server
	})

	t.Run("GeminiSafetySettingsOption", func(t *testing.T) {
		safetySettings := []map[string]interface{}{
			{
				"category":  "HARM_CATEGORY_HARASSMENT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
		}
		safetySettingsOption := domain.NewGeminiSafetySettingsOption(safetySettings)
		provider := NewGeminiProvider("test-key", "gemini-2.0-flash-lite", safetySettingsOption)

		// Check that the safetySettings field was updated
		if len(provider.safetySettings) != len(safetySettings) {
			t.Errorf("Expected safetySettings to have %d items, got %d", len(safetySettings), len(provider.safetySettings))
		}

		if len(provider.safetySettings) > 0 {
			if provider.safetySettings[0]["category"] != "HARM_CATEGORY_HARASSMENT" {
				t.Errorf("Expected safetySettings[0]['category'] to be 'HARM_CATEGORY_HARASSMENT', got %v",
					provider.safetySettings[0]["category"])
			}
			if provider.safetySettings[0]["threshold"] != "BLOCK_MEDIUM_AND_ABOVE" {
				t.Errorf("Expected safetySettings[0]['threshold'] to be 'BLOCK_MEDIUM_AND_ABOVE', got %v",
					provider.safetySettings[0]["threshold"])
			}
		}
	})

	t.Run("Multiple Options", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30 * time.Second}
		customTopK := 30
		safetySettings := []map[string]interface{}{
			{
				"category":  "HARM_CATEGORY_HARASSMENT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
		}

		options := []domain.ProviderOption{
			domain.NewBaseURLOption(mockServer.URL),
			domain.NewHTTPClientOption(customClient),
			domain.NewGeminiGenerationConfigOption().WithTopK(customTopK),
			domain.NewGeminiSafetySettingsOption(safetySettings),
		}

		provider := NewGeminiProvider("test-key", "gemini-2.0-flash-lite", options...)

		// Check that all options were applied correctly
		if provider.baseURL != mockServer.URL {
			t.Errorf("Expected baseURL to be %s, got %s", mockServer.URL, provider.baseURL)
		}
		if provider.httpClient != customClient {
			t.Error("HTTPClientOption failed to set the HTTP client")
		}
		if provider.topK != customTopK {
			t.Errorf("Expected topK to be %d, got %d", customTopK, provider.topK)
		}
		if len(provider.safetySettings) != len(safetySettings) {
			t.Errorf("Expected safetySettings to have %d items, got %d", len(safetySettings), len(provider.safetySettings))
		}
	})
}

func TestGeminiProviderWithRequestOptions(t *testing.T) {
	// Create a mock HTTP server that captures and validates the request
	var capturedRequest map[string]interface{}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the request body for inspection
		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Return a simple response
		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{
								"text": "Test response",
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Error encoding response: %v", err)
			return
		}
	}))
	defer mockServer.Close()

	t.Run("GenerationConfig in Request", func(t *testing.T) {
		// Create provider with topK option
		customTopK := 25
		generationConfigOption := domain.NewGeminiGenerationConfigOption().WithTopK(customTopK)
		baseURLOption := domain.NewBaseURLOption(mockServer.URL)

		provider := NewGeminiProvider("test-key", "gemini-2.0-flash-lite", generationConfigOption, baseURLOption)

		// Make a request
		ctx := context.Background()
		_, err := provider.Generate(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify the topK was included in the request
		generationConfig, ok := capturedRequest["generationConfig"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected generationConfig in request")
		}

		if topK, ok := generationConfig["topK"].(float64); !ok || int(topK) != customTopK {
			t.Errorf("Expected topK to be %d, got %v", customTopK, generationConfig["topK"])
		}
	})

	t.Run("SafetySettings in Request", func(t *testing.T) {
		// Reset captured request
		capturedRequest = nil

		// Create provider with safety settings
		safetySettings := []map[string]interface{}{
			{
				"category":  "HARM_CATEGORY_DANGEROUS",
				"threshold": "BLOCK_ONLY_HIGH",
			},
		}
		safetySettingsOption := domain.NewGeminiSafetySettingsOption(safetySettings)
		baseURLOption := domain.NewBaseURLOption(mockServer.URL)

		provider := NewGeminiProvider("test-key", "gemini-2.0-flash-lite", safetySettingsOption, baseURLOption)

		// Make a request
		ctx := context.Background()
		_, err := provider.Generate(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify the safety settings were included in the request
		settings, ok := capturedRequest["safetySettings"].([]interface{})
		if !ok {
			t.Fatal("Expected safetySettings in request")
		}

		if len(settings) != 1 {
			t.Fatalf("Expected 1 safety setting, got %d", len(settings))
		}

		setting, ok := settings[0].(map[string]interface{})
		if !ok {
			t.Fatal("Expected safety setting to be a map")
		}

		if category, ok := setting["category"].(string); !ok || category != "HARM_CATEGORY_DANGEROUS" {
			t.Errorf("Expected category 'HARM_CATEGORY_DANGEROUS', got %v", setting["category"])
		}

		if threshold, ok := setting["threshold"].(string); !ok || threshold != "BLOCK_ONLY_HIGH" {
			t.Errorf("Expected threshold 'BLOCK_ONLY_HIGH', got %v", setting["threshold"])
		}
	})
}
