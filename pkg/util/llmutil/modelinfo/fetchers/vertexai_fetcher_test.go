package fetchers

import (
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

// mockTokenSource implements oauth2.TokenSource for testing
type mockTokenSource struct {
	token string
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: m.token,
	}, nil
}

func TestVertexAIFetcher_FetchModels(t *testing.T) {
	// Create fetcher with mock auth
	fetcher := &VertexAIFetcher{
		projectID:   "test-project",
		location:    "us-central1",
		tokenSource: &mockTokenSource{token: "test-token"},
	}

	// Fetch models
	models, err := fetcher.FetchModels()
	if err != nil {
		t.Fatalf("FetchModels failed: %v", err)
	}

	// Verify we have models
	if len(models) == 0 {
		t.Fatal("Expected at least one model")
	}

	// Check for expected models
	hasGemini := false
	hasClaude := false

	for _, model := range models {
		if model.Provider == "vertexai:google" {
			hasGemini = true
		}
		if model.Provider == "vertexai:anthropic" {
			hasClaude = true
		}

		// Verify required fields
		if model.Name == "" {
			t.Error("Model name should not be empty")
		}
		if model.ContextWindow <= 0 {
			t.Errorf("Model %s should have positive context window", model.Name)
		}
		if model.MaxOutputTokens <= 0 {
			t.Errorf("Model %s should have positive max output tokens", model.Name)
		}
	}

	if !hasGemini {
		t.Error("Expected at least one Google model")
	}
	if !hasClaude {
		t.Error("Expected at least one partner (Claude) model")
	}
}

func TestVertexAIFetcher_GetModelFamily(t *testing.T) {
	tests := []struct {
		modelID  string
		expected string
	}{
		{"gemini-1.5-pro", "gemini"},
		{"gemini-2.0-flash", "gemini"},
		{"claude-3-opus@20240229", "claude"},
		{"claude-3-5-haiku@20241022", "claude"},
		{"unknown-model", ""},
	}

	for _, test := range tests {
		result := getModelFamily(test.modelID)
		if result != test.expected {
			t.Errorf("getModelFamily(%s) = %s, want %s", test.modelID, result, test.expected)
		}
	}
}

func TestVertexAIFetcher_FetchModelDetails(t *testing.T) {
	fetcher := &VertexAIFetcher{
		projectID:   "test-project",
		location:    "us-central1",
		tokenSource: &mockTokenSource{token: "test-token"},
	}

	// Test finding a known model
	model, err := fetcher.FetchModelDetails("gemini-1.5-pro-001")
	if err != nil {
		t.Fatalf("FetchModelDetails failed: %v", err)
	}

	if model == nil {
		t.Fatal("Expected model to be found")
	}

	if !strings.Contains(model.Name, "gemini-1.5-pro-001") {
		t.Errorf("Expected model name to contain gemini-1.5-pro-001, got %s", model.Name)
	}

	// Test model not found
	_, err = fetcher.FetchModelDetails("non-existent-model")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}
