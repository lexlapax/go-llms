package fetchers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/lexlapax/go-llms/pkg/modelinfo/domain"
)

const (
	googleAIAPIURLBase = "https://generativelanguage.googleapis.com/v1beta/models"
)

// GoogleFetcher fetches model information from the Google AI (Gemini) API.
type GoogleFetcher struct {
	// No fields needed for now, but can hold client or config later.
}

// GoogleAIModel represents a single model object from the Google AI API response.
type GoogleAIModel struct {
	Name                       string   `json:"name"`
	Version                    string   `json:"version"`
	DisplayName                string   `json:"displayName"`
	Description                string   `json:"description"`
	InputTokenLimit            int      `json:"inputTokenLimit"`
	OutputTokenLimit           int      `json:"outputTokenLimit"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	// Other fields like temperatureControls, topPControls, etc., are ignored for now.
}

// GoogleAIResponse is the top-level structure of the Google AI /v1beta/models API response.
type GoogleAIResponse struct {
	Models []GoogleAIModel `json:"models"`
}

// FetchModels retrieves model information from the Google AI (Gemini) API.
func (f *GoogleFetcher) FetchModels() ([]domain.Model, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY environment variable not set")
	}

	url := fmt.Sprintf("%s?key=%s", googleAIAPIURLBase, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Google AI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody strings.Builder
		if _, err := errorBody.ReadFrom(resp.Body); err != nil {
			return nil, fmt.Errorf("Google AI API request failed with status code: %d (error reading body)", resp.StatusCode)
		}
		return nil, fmt.Errorf("Google AI API request failed with status code: %d, body: %s", resp.StatusCode, errorBody.String())
	}

	var apiResponse GoogleAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Google AI API response: %w", err)
	}

	var models []domain.Model
	for _, apiModel := range apiResponse.Models {
		modelName := strings.TrimPrefix(apiModel.Name, "models/")

		// Basic placeholder capabilities
		capabilities := domain.Capabilities{
			Text:            domain.MediaTypeCapability{Read: false, Write: false},
			Image:           domain.MediaTypeCapability{Read: false, Write: false},
			Audio:           domain.MediaTypeCapability{Read: false, Write: false},
			Video:           domain.MediaTypeCapability{Read: false, Write: false},
			File:            domain.MediaTypeCapability{Read: false, Write: false},
			FunctionCalling: false, // Placeholder, Gemini has function calling
			Streaming:       false, // Will infer from supportedGenerationMethods
			JSONMode:        false, // Placeholder
		}

		for _, method := range apiModel.SupportedGenerationMethods {
			if method == "generateContent" {
				capabilities.Text.Read = true // Assumes text input for generateContent
				capabilities.Text.Write = true
			}
			if method == "streamGenerateContent" {
				capabilities.Streaming = true
			}
			// TODO: Infer other capabilities if possible from methods or model name patterns
		}
		
		// Infer some capabilities based on model name (very basic)
		if strings.Contains(modelName, "gemini") {
			capabilities.FunctionCalling = true // Gemini models generally support this
			if strings.Contains(modelName, "pro-vision") || strings.Contains(modelName, "1.5-pro") || strings.Contains(modelName, "1.5-flash"){ // Gemini 1.5 Pro/Flash are multimodal
				capabilities.Image.Read = true
				capabilities.Audio.Read = true 
				// Video and File might also be true for 1.5 Pro depending on specific features enabled
			}
		}


		// Basic placeholder pricing
		placeholderPricing := domain.Pricing{
			InputPer1kTokens:  0.0,
			OutputPer1kTokens: 0.0,
		}

		model := domain.Model{
			Provider:         "google",
			Name:             modelName,
			DisplayName:      apiModel.DisplayName,
			Description:      apiModel.Description,
			DocumentationURL: "https://ai.google.dev/models/" + guessModelFamily(modelName), // Basic guess
			Capabilities:     capabilities,
			ContextWindow:    apiModel.InputTokenLimit,
			MaxOutputTokens:  apiModel.OutputTokenLimit,
			TrainingCutoff:   "", // Placeholder, e.g., "2023-11"
			ModelFamily:      guessModelFamily(modelName) + " (v" + apiModel.Version + ")",
			Pricing:          placeholderPricing,
			LastUpdated:      "", // Placeholder
		}
		models = append(models, model)
	}

	return models, nil
}

// guessModelFamily tries to infer a model family from the model name.
// This is a very basic heuristic.
func guessModelFamily(modelName string) string {
	if strings.HasPrefix(modelName, "gemini-") {
		return "gemini"
	}
	if strings.HasPrefix(modelName, "text-") {
		return "text" // e.g. text-bison
	}
	if strings.HasPrefix(modelName, "embedding-") {
		return "embedding" // e.g. embedding-gecko-001
	}
	// Fallback to the first part of the name if available
	parts := strings.Split(modelName, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}
