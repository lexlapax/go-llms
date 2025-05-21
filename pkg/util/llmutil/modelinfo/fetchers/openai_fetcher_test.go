package fetchers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

func TestOpenAIFetcher_FetchModels_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("Expected path to be '/', got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer testapikey" {
			t.Fatalf("Expected Authorization header 'Bearer testapikey', got '%s'", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		response := OpenAIAPIResponse{
			Object: "list",
			Data: []OpenAIAPIModel{
				{ID: "gpt-3.5-turbo", Object: "model", Created: 1677610602, OwnedBy: "openai"},
				{ID: "dall-e-2", Object: "model", Created: 1677610603, OwnedBy: "openai"},
				{ID: "text-embedding-ada-002", Object: "model", Created: 1677610604, OwnedBy: "openai"},
			},
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	originalApiKey := os.Getenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_API_KEY", "testapikey")
	defer os.Setenv("OPENAI_API_KEY", originalApiKey)

	fetcher := NewOpenAIFetcher(server.URL, http.DefaultClient) // Use constructor with mock server URL
	models, err := fetcher.FetchModels()

	if err != nil {
		t.Fatalf("FetchModels() returned an unexpected error: %v", err)
	}

	if len(models) != 3 {
		t.Errorf("FetchModels() returned %d models, expected 3", len(models))
	}

	var gptModel *domain.Model
	for i := range models {
		if models[i].Name == "gpt-3.5-turbo" {
			gptModel = &models[i]
			break
		}
	}

	if gptModel == nil {
		t.Fatal("Expected model 'gpt-3.5-turbo' not found.")
	}

	if gptModel.Provider != "openai" {
		t.Errorf("Expected gpt-3.5-turbo provider to be 'openai', got '%s'", gptModel.Provider)
	}
	if gptModel.ModelFamily != "gpt" { // Based on extractModelFamily
		t.Errorf("Expected gpt-3.5-turbo ModelFamily to be 'gpt', got '%s'", gptModel.ModelFamily)
	}
	// Check one of the inferred capabilities
	if !gptModel.Capabilities.Text.Read || !gptModel.Capabilities.Text.Write {
		t.Error("Expected gpt-3.5-turbo to have text read/write capabilities")
	}
	if gptModel.Capabilities.Image.Read { // gpt-3.5-turbo should not have image read by default inference
		t.Error("Expected gpt-3.5-turbo to NOT have image read capability by default inference")
	}
}

func TestOpenAIFetcher_FetchModels_APIKeyMissing(t *testing.T) {
	originalApiKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalApiKey)

	// Instantiate with default or empty URL, as API key check happens first
	fetcher := NewOpenAIFetcher("", http.DefaultClient)
	_, err := fetcher.FetchModels()

	if err == nil {
		t.Fatal("FetchModels() expected an error when API key is missing, but got nil")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY environment variable not set") {
		t.Errorf("FetchModels() expected error message about missing API key, got: %v", err)
	}
}

func TestOpenAIFetcher_FetchModels_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"error": {"message": "Internal server error", "type": "server_error"}}`)
	}))
	defer server.Close()

	originalApiKey := os.Getenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_API_KEY", "testapikey")
	defer os.Setenv("OPENAI_API_KEY", originalApiKey)

	fetcher := NewOpenAIFetcher(server.URL, http.DefaultClient) // Use constructor
	_, err := fetcher.FetchModels()

	if err == nil {
		t.Fatal("FetchModels() expected an error on API error status, but got nil")
	}
	if !strings.Contains(err.Error(), "OpenAI API request failed with status code: 500") {
		t.Errorf("FetchModels() expected error message about API failure, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Internal server error") {
		t.Errorf("FetchModels() expected error message to contain API error body, got: %v", err)
	}
}

func TestOpenAIFetcher_FetchModels_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"object": "list", "data": [`) // Malformed JSON
	}))
	defer server.Close()

	originalApiKey := os.Getenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_API_KEY", "testapikey")
	defer os.Setenv("OPENAI_API_KEY", originalApiKey)

	fetcher := NewOpenAIFetcher(server.URL, http.DefaultClient) // Use constructor
	_, err := fetcher.FetchModels()

	if err == nil {
		t.Fatal("FetchModels() expected an error on invalid JSON response, but got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode OpenAI API response") {
		t.Errorf("FetchModels() expected error message about JSON decoding failure, got: %v", err)
	}
}

func TestOpenAIFetcher_ExtractModelFamily(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected string
	}{
		{"GPT-4 Turbo", "gpt-4-turbo-preview", "gpt"},
		{"GPT-3.5 Turbo", "gpt-3.5-turbo-0125", "gpt"},
		{"DALL-E 3", "dall-e-3", "dall-e"},
		{"TTS", "tts-1-hd", "tts"},
		{"Whisper", "whisper-1", "whisper"},
		{"Ada Embedding", "text-embedding-ada-002", "text-embedding"},
		{"Babbage", "babbage-002", "babbage"},
		{"Unknown", "custom-model-x", "custom"},
		{"Empty", "", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModelFamily(tt.modelID)
			if got != tt.expected {
				t.Errorf("extractModelFamily(%q) = %q, want %q", tt.modelID, got, tt.expected)
			}
		})
	}
}
