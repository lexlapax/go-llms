package fetchers

// ABOUTME: Ollama API client for fetching locally available models
// ABOUTME: Retrieves model information from local Ollama instance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

const defaultOllamaBaseURL = "http://localhost:11434"

// OllamaFetcher fetches model information from a local Ollama instance.
type OllamaFetcher struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewOllamaFetcher creates a new OllamaFetcher.
// If baseURL is empty, it defaults to "http://localhost:11434".
// If client is nil, it defaults to http.DefaultClient.
func NewOllamaFetcher(baseURL string, client *http.Client) *OllamaFetcher {
	if baseURL == "" {
		// Check environment variable first
		baseURL = os.Getenv("OLLAMA_HOST")
		if baseURL == "" {
			baseURL = defaultOllamaBaseURL
		}
	}
	httpClient := client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &OllamaFetcher{BaseURL: baseURL, HTTPClient: httpClient}
}

// OllamaModelDetails represents detailed information about a model from Ollama
type OllamaModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// OllamaModel represents a single model object from the Ollama API response.
type OllamaModel struct {
	Name       string             `json:"name"`
	Model      string             `json:"model"`
	ModifiedAt time.Time          `json:"modified_at"`
	Size       int64              `json:"size"`
	Digest     string             `json:"digest"`
	Details    OllamaModelDetails `json:"details"`
}

// OllamaAPIResponse is the response structure from Ollama's /api/tags endpoint
type OllamaAPIResponse struct {
	Models []OllamaModel `json:"models"`
}

// FetchModels retrieves model information from the Ollama API.
func (f *OllamaFetcher) FetchModels() ([]domain.Model, error) {
	requestURL := f.BaseURL + "/api/tags"
	resp, err := f.HTTPClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Ollama API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		errorBodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("ollama API request failed with status code: %d (and error reading response body: %w)", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("ollama API request failed with status code: %d, body: %s", resp.StatusCode, string(errorBodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var ollamaResponse OllamaAPIResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Ollama API response: %w", err)
	}

	models := make([]domain.Model, 0, len(ollamaResponse.Models))
	for _, ollamaModel := range ollamaResponse.Models {
		model := f.convertToModel(ollamaModel)
		models = append(models, model)
	}

	return models, nil
}

// convertToModel converts an Ollama model to the domain Model format
func (f *OllamaFetcher) convertToModel(ollamaModel OllamaModel) domain.Model {
	// Extract model family and size info from the name
	// e.g., "llama3.2:3b" -> family: "llama", size: "3b"
	modelFamily := ollamaModel.Details.Family
	if modelFamily == "" {
		// Try to extract from name
		parts := strings.Split(ollamaModel.Name, ":")
		if len(parts) > 0 {
			// Remove version numbers from the base name
			baseName := parts[0]
			// Remove trailing numbers and dots
			for i := len(baseName) - 1; i >= 0; i-- {
				if baseName[i] != '.' && (baseName[i] < '0' || baseName[i] > '9') {
					modelFamily = baseName[:i+1]
					break
				}
			}
		}
	}

	// Determine context window based on model family and parameter size
	contextWindow := f.estimateContextWindow(ollamaModel)

	// All Ollama models support streaming and function calling (through OpenAI API)
	capabilities := domain.Capabilities{
		Text: domain.MediaTypeCapability{
			Read:  true,
			Write: true,
		},
		Image: domain.MediaTypeCapability{
			Read:  f.supportsVision(ollamaModel),
			Write: false, // Ollama doesn't generate images
		},
		Audio: domain.MediaTypeCapability{
			Read:  false,
			Write: false,
		},
		Video: domain.MediaTypeCapability{
			Read:  false,
			Write: false,
		},
		File: domain.MediaTypeCapability{
			Read:  false,
			Write: false,
		},
		FunctionCalling: true, // All models support function calling via OpenAI API
		Streaming:       true, // All Ollama models support streaming
		JSONMode:        true, // Available through OpenAI-compatible API
	}

	// Format the size in GB
	sizeGB := float64(ollamaModel.Size) / (1024 * 1024 * 1024)
	description := fmt.Sprintf("%s model (%s parameters, %.1fGB) - %s quantization",
		modelFamily,
		ollamaModel.Details.ParameterSize,
		sizeGB,
		ollamaModel.Details.QuantizationLevel,
	)

	return domain.Model{
		Provider:         "ollama",
		Name:             ollamaModel.Name,
		DisplayName:      ollamaModel.Name,
		Description:      description,
		DocumentationURL: "https://ollama.com/library/" + strings.Split(ollamaModel.Name, ":")[0],
		Capabilities:     capabilities,
		ContextWindow:    contextWindow,
		MaxOutputTokens:  contextWindow / 2, // Conservative estimate
		TrainingCutoff:   "",                // Not available from Ollama API
		ModelFamily:      modelFamily,
		Pricing: domain.Pricing{
			InputPer1kTokens:  0.0, // Free for local inference
			OutputPer1kTokens: 0.0, // Free for local inference
		},
		LastUpdated: ollamaModel.ModifiedAt.Format("2006-01-02"),
	}
}

// estimateContextWindow estimates the context window based on model information
func (f *OllamaFetcher) estimateContextWindow(model OllamaModel) int {
	// Check for known model families and their typical context windows
	name := strings.ToLower(model.Name)

	// Llama models
	if strings.Contains(name, "llama") {
		if strings.Contains(name, "llama3") || strings.Contains(name, "llama-3") {
			return 8192 // Llama 3 models typically have 8k context
		}
		if strings.Contains(name, "llama2") || strings.Contains(name, "llama-2") {
			return 4096 // Llama 2 models typically have 4k context
		}
	}

	// Mistral models
	if strings.Contains(name, "mistral") {
		if strings.Contains(name, "7b") {
			return 8192 // Mistral 7B has 8k context
		}
	}

	// Gemma models
	if strings.Contains(name, "gemma") {
		return 8192 // Gemma models typically have 8k context
	}

	// Qwen models
	if strings.Contains(name, "qwen") {
		if strings.Contains(name, "qwen2") {
			return 32768 // Qwen2 models often have 32k context
		}
		return 8192
	}

	// Phi models
	if strings.Contains(name, "phi") {
		return 4096 // Phi models typically have 4k context
	}

	// CodeLlama
	if strings.Contains(name, "codellama") {
		return 16384 // CodeLlama often has 16k context
	}

	// Default conservative estimate
	return 4096
}

// supportsVision checks if the model supports vision/image input
func (f *OllamaFetcher) supportsVision(model OllamaModel) bool {
	name := strings.ToLower(model.Name)

	// Known vision models
	visionModels := []string{
		"llava",
		"bakllava",
		"moondream",
		"llama3.2-vision",
		"minicpm-v",
	}

	for _, vm := range visionModels {
		if strings.Contains(name, vm) {
			return true
		}
	}

	return false
}
