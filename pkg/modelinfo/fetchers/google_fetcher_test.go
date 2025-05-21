package fetchers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/modelinfo/domain"
)

// Helper to modify the package-level googleAIAPIURLBase for testing and restore it.
// Note: Since googleAIAPIURLBase is a const, we can't change it directly.
// Instead, we'll pass the full mock server URL to a (hypothetical) constructor or modify how the URL is built in the fetcher.
// For this test, we'll assume the fetcher uses the global const and thus we need a way to control the full URL.
// A better way would be to make the URL configurable in the fetcher.
// For now, the test will construct the full URL for the mock server and the fetcher will need to be able to use it.
// For simplicity in test, we can't directly override const. The test will have to work around it or the main code would need a slight refactor.
// Let's assume for the test that the URL construction is: fmt.Sprintf("%s?key=%s", GoogleAIAPIURLBase, apiKey)
// We will make GoogleAIAPIURLBase a var for testing purposes.

func TestGoogleFetcher_FetchModels_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path for Google is part of the BaseURL + "/models"
		// The key is a query parameter, which is fine.
		// No specific path check here as long as the server gets the request.
		if r.URL.Query().Get("key") != "testapikey" {
			t.Fatalf("Expected API key 'testapikey', got '%s'", r.URL.Query().Get("key"))
		}
		w.WriteHeader(http.StatusOK)
		response := GoogleAIResponse{
			Models: []GoogleAIModel{
				{
					Name:                       "models/gemini-1.5-pro-latest",
					Version:                    "v1",
					DisplayName:                "Gemini 1.5 Pro",
					Description:                "Most capable model.",
					InputTokenLimit:            1048576,
					OutputTokenLimit:           8192,
					SupportedGenerationMethods: []string{"generateContent", "streamGenerateContent"},
				},
				{
					Name:                       "models/gemini-1.0-pro",
					Version:                    "v1", // Example version
					DisplayName:                "Gemini 1.0 Pro",
					Description:                "Previous generation pro model.",
					InputTokenLimit:            30720,
					OutputTokenLimit:           2048,
					SupportedGenerationMethods: []string{"generateContent", "streamGenerateContent"},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalApiKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "testapikey")
	defer os.Setenv("GEMINI_API_KEY", originalApiKey)

	fetcher := NewGoogleFetcher(server.URL, http.DefaultClient) // Use constructor with mock server URL
	models, err := fetcher.FetchModels()

	if err != nil {
		t.Fatalf("FetchModels() returned an unexpected error: %v", err)
	}

	if len(models) != 2 {
		t.Errorf("FetchModels() returned %d models, expected 2", len(models))
	}

	var gemini15Pro *domain.Model
	for i := range models {
		if models[i].Name == "gemini-1.5-pro-latest" {
			gemini15Pro = &models[i]
			break
		}
	}

	if gemini15Pro == nil {
		t.Fatal("Expected model 'gemini-1.5-pro-latest' not found.")
	}

	if gemini15Pro.Provider != "google" {
		t.Errorf("Expected gemini-1.5-pro-latest provider to be 'google', got '%s'", gemini15Pro.Provider)
	}
	if gemini15Pro.DisplayName != "Gemini 1.5 Pro" {
		t.Errorf("Expected DisplayName 'Gemini 1.5 Pro', got '%s'", gemini15Pro.DisplayName)
	}
	if gemini15Pro.ContextWindow != 1048576 {
		t.Errorf("Expected ContextWindow 1048576, got %d", gemini15Pro.ContextWindow)
	}
	if gemini15Pro.MaxOutputTokens != 8192 {
		t.Errorf("Expected MaxOutputTokens 8192, got %d", gemini15Pro.MaxOutputTokens)
	}
	if !gemini15Pro.Capabilities.Streaming {
		t.Error("Expected gemini-1.5-pro-latest to have streaming capability inferred.")
	}
	if !gemini15Pro.Capabilities.FunctionCalling { // Based on current inference logic
		t.Error("Expected gemini-1.5-pro-latest to have function calling capability inferred.")
	}
	if !gemini15Pro.Capabilities.Image.Read { // Based on current inference logic for 1.5 pro
		t.Error("Expected gemini-1.5-pro-latest to have image read capability inferred.")
	}
}

func TestGoogleFetcher_FetchModels_APIKeyMissing(t *testing.T) {
	originalApiKey := os.Getenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", originalApiKey)

	fetcher := NewGoogleFetcher("", http.DefaultClient) // API key check happens first
	_, err := fetcher.FetchModels()

	if err == nil {
		t.Fatal("FetchModels() expected an error when API key is missing, but got nil")
	}
	if !strings.Contains(err.Error(), "GEMINI_API_KEY environment variable not set") {
		t.Errorf("FetchModels() expected error message about missing API key, got: %v", err)
	}
}

func TestGoogleFetcher_FetchModels_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden) // Example: API key invalid or access denied
		fmt.Fprintln(w, `{"error": {"message": "API key not valid. Please pass a valid API key.", "status": "PERMISSION_DENIED"}}`)
	}))
	defer server.Close()

	originalApiKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "invalidkey")
	defer os.Setenv("GEMINI_API_KEY", originalApiKey)

	fetcher := NewGoogleFetcher(server.URL, http.DefaultClient) // Use constructor
	_, err := fetcher.FetchModels()

	if err == nil {
		t.Fatal("FetchModels() expected an error on API error status, but got nil")
	}
	if !strings.Contains(err.Error(), "Google AI API request failed with status code: 403") {
		t.Errorf("FetchModels() expected error message about API failure (403), got: %v", err)
	}
	if !strings.Contains(err.Error(), "API key not valid") {
		t.Errorf("FetchModels() expected error message to contain API error body, got: %v", err)
	}
}

func TestGoogleFetcher_FetchModels_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"models": [`) // Malformed JSON
	}))
	defer server.Close()

	originalApiKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "testapikey")
	defer os.Setenv("GEMINI_API_KEY", originalApiKey)

	fetcher := NewGoogleFetcher(server.URL, http.DefaultClient) // Use constructor
	_, err := fetcher.FetchModels()

	if err == nil {
		t.Fatal("FetchModels() expected an error on invalid JSON response, but got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode Google AI API response") {
		t.Errorf("FetchModels() expected error message about JSON decoding failure, got: %v", err)
	}
}

func TestGoogleFetcher_GuessModelFamily(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected string
	}{
		{"Gemini Pro", "gemini-pro", "gemini"},
		{"Gemini 1.5 Flash", "gemini-1.5-flash-latest", "gemini"},
		{"Text Bison", "text-bison-001", "text"},
		{"Embedding Gecko", "embedding-gecko-001", "embedding"},
		{"AQA", "aqa", "aqa"}, // Example of a model that doesn't fit other prefixes
		{"Unknown", "some-custom-model", "some"},
		{"Empty", "", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guessModelFamily(tt.modelID)
			if got != tt.expected {
				t.Errorf("guessModelFamily(%q) = %q, want %q", tt.modelID, got, tt.expected)
			}
		})
	}
}
