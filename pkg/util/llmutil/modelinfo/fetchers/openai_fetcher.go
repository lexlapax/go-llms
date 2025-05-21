package fetchers

import (
	"encoding/json"
	"fmt"
	"io" // Added io import
	"net/http"
	"os"
	"strings"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

// OpenAIFetcher fetches model information from the OpenAI API.
type OpenAIFetcher struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewOpenAIFetcher creates a new OpenAIFetcher.
// If baseURL is empty, it defaults to "https://api.openai.com/v1".
// If client is nil, it defaults to http.DefaultClient.
func NewOpenAIFetcher(baseURL string, client *http.Client) *OpenAIFetcher {
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	httpClient := client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &OpenAIFetcher{BaseURL: baseURL, HTTPClient: httpClient}
}

// OpenAIAPIModel represents a single model object from the OpenAI API response.
// We only include fields we are interested in for now.
type OpenAIAPIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
	// Other fields like 'permission', 'root', 'parent' are ignored for now.
}

// OpenAIAPIResponse is the top-level structure of the OpenAI /v1/models API response.
type OpenAIAPIResponse struct {
	Object string           `json:"object"`
	Data   []OpenAIAPIModel `json:"data"`
}

// FetchModels retrieves model information from the OpenAI API.
func (f *OpenAIFetcher) FetchModels() ([]domain.Model, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	requestURL := f.BaseURL + "/models"
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to %s: %w", requestURL, err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to OpenAI API using custom client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("OpenAI API request failed with status code: %d (and error reading response body: %w)", resp.StatusCode, readErr)
		}
		errorBodyStr := string(errorBodyBytes)
		return nil, fmt.Errorf("OpenAI API request failed with status code: %d, body: %s", resp.StatusCode, errorBodyStr)
	}

	var apiResponse OpenAIAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode OpenAI API response: %w", err)
	}

	var models []domain.Model
	for _, apiModel := range apiResponse.Data {
		// Basic placeholder capabilities
		placeholderCapabilities := domain.Capabilities{
			Text:            domain.MediaTypeCapability{Read: true, Write: true}, // Assume text primarily
			Image:           domain.MediaTypeCapability{Read: false, Write: false},
			Audio:           domain.MediaTypeCapability{Read: false, Write: false},
			Video:           domain.MediaTypeCapability{Read: false, Write: false},
			File:            domain.MediaTypeCapability{Read: false, Write: false},
			FunctionCalling: false, // Placeholder - actual capability varies by model
			Streaming:       true,  // Most OpenAI models support streaming
			JSONMode:        false, // Placeholder - actual capability varies by model
		}

		// Basic placeholder pricing
		placeholderPricing := domain.Pricing{
			InputPer1kTokens:  0.0,
			OutputPer1kTokens: 0.0,
		}

		// Try to infer some very basic capabilities for known model prefixes
		if strings.HasPrefix(apiModel.ID, "gpt-4") {
			// More advanced models might have function calling, JSON mode, etc.
			// This is still a placeholder and would need proper lookup
			placeholderCapabilities.FunctionCalling = true
			if strings.Contains(apiModel.ID, "vision") || strings.Contains(apiModel.ID, "gpt-4o") {
				placeholderCapabilities.Image.Read = true
			}
		}
		if strings.HasPrefix(apiModel.ID, "dall-e") {
			placeholderCapabilities.Text.Read = true   // Takes prompt
			placeholderCapabilities.Text.Write = false // Does not output text
			placeholderCapabilities.Image.Read = false // Does not take image input in base API
			placeholderCapabilities.Image.Write = true // Outputs image
		}
		if strings.HasPrefix(apiModel.ID, "tts") {
			placeholderCapabilities.Text.Read = true // Takes text
			placeholderCapabilities.Text.Write = false
			placeholderCapabilities.Audio.Write = true // Outputs audio
		}
		if strings.HasPrefix(apiModel.ID, "whisper") {
			placeholderCapabilities.Audio.Read = true // Takes audio
			placeholderCapabilities.Text.Write = true // Outputs text
		}

		model := domain.Model{
			Provider:         "openai",
			Name:             apiModel.ID,
			DisplayName:      apiModel.ID,                                              // Placeholder, could be improved with a mapping
			Description:      "Model data fetched from OpenAI API.",                    // Placeholder
			DocumentationURL: "https://platform.openai.com/docs/models/" + apiModel.ID, // Basic guess
			Capabilities:     placeholderCapabilities,
			ContextWindow:    0,                               // Placeholder, e.g. 8192 for gpt-3.5-turbo
			MaxOutputTokens:  0,                               // Placeholder, e.g. 4096 for many models
			TrainingCutoff:   "",                              // Placeholder, e.g. "2023-09"
			ModelFamily:      extractModelFamily(apiModel.ID), // Basic inference
			Pricing:          placeholderPricing,
			LastUpdated:      "", // Placeholder, could use apiModel.Created if formatted
		}
		models = append(models, model)
	}

	return models, nil
}

// extractModelFamily tries to infer a model family from the model ID.
// This is a very basic heuristic.
func extractModelFamily(modelID string) string {
	if modelID == "" {
		return "unknown"
	}
	if strings.HasPrefix(modelID, "text-embedding") {
		return "text-embedding"
	}
	if strings.HasPrefix(modelID, "gpt-") {
		return "gpt" // Covers gpt-3.5, gpt-4, gpt-4o etc.
	}
	if strings.HasPrefix(modelID, "dall-e") { // Check for "dall-e" specifically
		// Find the last part of "dall-e-..." which is usually the version
		parts := strings.Split(modelID, "-")
		if len(parts) >= 2 && parts[0] == "dall" && parts[1] == "e" {
			return "dall-e"
		}
		return parts[0] // Fallback if not "dall-e-X"
	}

	parts := strings.Split(modelID, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown" // Should be unreachable if modelID is not empty
}

// Helper function to create float pointers - not needed for domain.Model
// func float64Ptr(v float64) *float64 { return &v }
// Helper function to create int pointers - not needed for domain.Model
// func intPtr(v int) *int { return &v }
