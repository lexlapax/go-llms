package fetchers

// ABOUTME: OpenRouter model fetcher implementation for discovering available models
// ABOUTME: Fetches model information from OpenRouter's /api/v1/models endpoint

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

// OpenRouterFetcher fetches model information from OpenRouter
type OpenRouterFetcher struct {
	BaseURL string
	APIKey  string
}

// NewOpenRouterFetcher creates a new OpenRouter model fetcher
func NewOpenRouterFetcher(apiKey string) *OpenRouterFetcher {
	return &OpenRouterFetcher{
		BaseURL: "https://openrouter.ai/api/v1",
		APIKey:  apiKey,
	}
}

// openRouterModelsResponse represents the response from OpenRouter's models endpoint
type openRouterModelsResponse struct {
	Data []openRouterModel `json:"data"`
}

// openRouterModel represents a single model from OpenRouter
type openRouterModel struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Created          int64                  `json:"created,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Pricing          openRouterPricing      `json:"pricing"`
	ContextLength    int                    `json:"context_length"`
	Architecture     openRouterArchitecture `json:"architecture,omitempty"`
	TopProvider      openRouterTopProvider  `json:"top_provider,omitempty"`
	PerRequestLimits map[string]interface{} `json:"per_request_limits,omitempty"`
}

type openRouterPricing struct {
	Prompt     string `json:"prompt"`     // USD per token
	Completion string `json:"completion"` // USD per token
	Request    string `json:"request"`    // USD per request
	Image      string `json:"image"`      // USD per image
}

type openRouterArchitecture struct {
	Modality     string `json:"modality"`
	Tokenizer    string `json:"tokenizer"`
	InstructType string `json:"instruct_type"`
}

type openRouterTopProvider struct {
	MaxCompletionTokens int  `json:"max_completion_tokens,omitempty"`
	IsModerated         bool `json:"is_moderated,omitempty"`
}

// FetchModels retrieves available models from OpenRouter
func (f *OpenRouterFetcher) FetchModels() ([]domain.Model, error) {
	requestURL := f.BaseURL + "/models"

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add authorization header if API key is provided
	if f.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+f.APIKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching models: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var response openRouterModelsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	models := make([]domain.Model, 0, len(response.Data))
	for _, orModel := range response.Data {
		model := domain.Model{
			Name:          orModel.ID,
			Provider:      "openrouter",
			ContextWindow: orModel.ContextLength,
			Description:   orModel.Description,
			LastUpdated:   time.Unix(orModel.Created, 0).Format("2006-01-02"),
		}

		// Extract capabilities from architecture
		if orModel.Architecture.Modality != "" {
			modality := strings.ToLower(orModel.Architecture.Modality)
			if strings.Contains(modality, "text") {
				model.Capabilities.Text.Read = true
				model.Capabilities.Text.Write = true
			}
			if strings.Contains(modality, "image") {
				model.Capabilities.Image.Read = true
			}
			if strings.Contains(modality, "multimodal") {
				model.Capabilities.Text.Read = true
				model.Capabilities.Text.Write = true
				model.Capabilities.Image.Read = true
			}
		}

		// Add streaming capability (most models support it)
		model.Capabilities.Streaming = true

		// Extract actual provider from model ID (e.g., "openai/gpt-4" -> "openai")
		if parts := strings.SplitN(orModel.ID, "/", 2); len(parts) > 0 {
			model.Provider = "openrouter:" + parts[0]
		}

		// Set max output tokens if available
		if orModel.TopProvider.MaxCompletionTokens > 0 {
			model.MaxOutputTokens = orModel.TopProvider.MaxCompletionTokens
		}

		// Set pricing
		if orModel.Pricing.Prompt != "" {
			promptPrice, _ := parsePrice(orModel.Pricing.Prompt)
			model.Pricing.InputPer1kTokens = promptPrice * 1000
		}
		if orModel.Pricing.Completion != "" {
			completionPrice, _ := parsePrice(orModel.Pricing.Completion)
			model.Pricing.OutputPer1kTokens = completionPrice * 1000
		}

		models = append(models, model)
	}

	return models, nil
}

// parsePrice converts a string price to float64
func parsePrice(price string) (float64, error) {
	if price == "" || price == "0" {
		return 0, nil
	}

	var p float64
	_, err := fmt.Sscanf(price, "%f", &p)
	return p, err
}
