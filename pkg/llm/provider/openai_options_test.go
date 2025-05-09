package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestOpenAIProviderOptions(t *testing.T) {
	// Create a mock HTTP server to simulate OpenAI API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check organization header if it's set
		orgHeader := r.Header.Get("OpenAI-Organization")

		// Just a simple response for the test
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4o",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "This is a test response. Organization: %s"
				},
				"finish_reason": "stop"
			}]
		}`, orgHeader)

		_, err := w.Write([]byte(response))
		if err != nil {
			t.Errorf("Error writing response: %v", err)
			return
		}
	}))
	defer mockServer.Close()

	t.Run("BaseURLOption", func(t *testing.T) {
		baseURLOption := domain.NewBaseURLOption(mockServer.URL)
		provider := NewOpenAIProvider("test-key", "gpt-4o", baseURLOption)

		if provider.baseURL != mockServer.URL {
			t.Errorf("Expected baseURL to be %s, got %s", mockServer.URL, provider.baseURL)
		}
	})

	t.Run("HTTPClientOption", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30 * time.Second}
		clientOption := domain.NewHTTPClientOption(customClient)
		provider := NewOpenAIProvider("test-key", "gpt-4o", clientOption)

		if provider.httpClient != customClient {
			t.Error("HTTPClientOption failed to set the HTTP client")
		}
	})

	t.Run("OrganizationOption", func(t *testing.T) {
		orgID := "org-123456"
		orgOption := domain.NewOpenAIOrganizationOption(orgID)
		baseURLOption := domain.NewBaseURLOption(mockServer.URL)

		provider := NewOpenAIProvider("test-key", "gpt-4o", orgOption, baseURLOption)

		// Check that the option field was updated
		if provider.organization != orgID {
			t.Errorf("Expected organization to be %s, got %s", orgID, provider.organization)
		}

		// Make a request and check if the organization header was sent
		ctx := context.Background()
		response, err := provider.Generate(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Response should contain the organization ID
		if !strings.Contains(response, orgID) {
			t.Errorf("Expected response to contain organization ID %s, got: %s", orgID, response)
		}
	})

	t.Run("LogitBiasOption", func(t *testing.T) {
		logitBias := map[string]float64{"1234": 0.5, "5678": -0.5}
		logitBiasOption := domain.NewOpenAILogitBiasOption(logitBias)
		provider := NewOpenAIProvider("test-key", "gpt-4o", logitBiasOption)

		if len(provider.logitBias) != len(logitBias) {
			t.Errorf("Expected logitBias to have %d items, got %d", len(logitBias), len(provider.logitBias))
		}

		for k, v := range logitBias {
			if provider.logitBias[k] != v {
				t.Errorf("Expected logitBias[%s] to be %f, got %f", k, v, provider.logitBias[k])
			}
		}
	})

	t.Run("Multiple Options", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30 * time.Second}
		orgID := "org-123456"
		logitBias := map[string]float64{"1234": 0.5, "5678": -0.5}

		options := []domain.ProviderOption{
			domain.NewBaseURLOption(mockServer.URL),
			domain.NewHTTPClientOption(customClient),
			domain.NewOpenAIOrganizationOption(orgID),
			domain.NewOpenAILogitBiasOption(logitBias),
		}

		provider := NewOpenAIProvider("test-key", "gpt-4o", options...)

		if provider.baseURL != mockServer.URL {
			t.Errorf("Expected baseURL to be %s, got %s", mockServer.URL, provider.baseURL)
		}

		if provider.httpClient != customClient {
			t.Error("HTTPClientOption failed to set the HTTP client")
		}

		if provider.organization != orgID {
			t.Errorf("Expected organization to be %s, got %s", orgID, provider.organization)
		}

		if len(provider.logitBias) != len(logitBias) {
			t.Errorf("Expected logitBias to have %d items, got %d", len(logitBias), len(provider.logitBias))
		}

		// Also test that the options actually affect the API request
		ctx := context.Background()
		response, err := provider.Generate(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Response should contain the organization ID
		if !strings.Contains(response, orgID) {
			t.Errorf("Expected response to contain organization ID %s, got: %s", orgID, response)
		}
	})
}
