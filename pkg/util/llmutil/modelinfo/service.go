package modelinfo

// ABOUTME: Service orchestrator for model information fetching and caching
// ABOUTME: Coordinates between providers to maintain model inventory

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/fetchers"
)

// ModelInfoService aggregates model information from various providers.
type ModelInfoService struct {
	openAIFetcher     *fetchers.OpenAIFetcher
	googleFetcher     *fetchers.GoogleFetcher
	anthropicFetcher  *fetchers.AnthropicFetcher
	ollamaFetcher     *fetchers.OllamaFetcher
	openRouterFetcher *fetchers.OpenRouterFetcher
	// Add other fetchers here if more providers are supported
}

// NewServiceWithCustomFetchers creates a ModelInfoService with specific fetcher instances.
// Useful for testing or custom provider configurations.
func NewServiceWithCustomFetchers(
	openAIFetcher *fetchers.OpenAIFetcher,
	googleFetcher *fetchers.GoogleFetcher,
	anthropicFetcher *fetchers.AnthropicFetcher,
	ollamaFetcher *fetchers.OllamaFetcher,
	openRouterFetcher *fetchers.OpenRouterFetcher,
) *ModelInfoService {
	return &ModelInfoService{
		openAIFetcher:     openAIFetcher,
		googleFetcher:     googleFetcher,
		anthropicFetcher:  anthropicFetcher,
		ollamaFetcher:     ollamaFetcher,
		openRouterFetcher: openRouterFetcher,
	}
}

// defaultNewModelInfoService is the default implementation for creating a ModelInfoService.
func defaultNewModelInfoService() *ModelInfoService {
	return NewServiceWithCustomFetchers(
		fetchers.NewOpenAIFetcher("", http.DefaultClient), // Uses default internal URL
		fetchers.NewGoogleFetcher("", http.DefaultClient), // Uses default internal URL
		&fetchers.AnthropicFetcher{},                      // Remains as is
		fetchers.NewOllamaFetcher("", http.DefaultClient), // Uses default internal URL
		fetchers.NewOpenRouterFetcher(""),                 // Uses no API key by default
	)
}

// NewModelInfoServiceFunc is a package-level variable that can be overridden in tests
// to provide a custom ModelInfoService instance.
var NewModelInfoServiceFunc = defaultNewModelInfoService

// AggregateModels fetches model information from all configured providers and aggregates them.
func (s *ModelInfoService) AggregateModels() (*domain.ModelInventory, error) {
	allModels := []domain.Model{}
	var overallErr error
	var fetcherErrors []string

	// Fetch from OpenAI
	openAIModels, err := s.openAIFetcher.FetchModels()
	if err != nil {
		errMsg := fmt.Sprintf("Error fetching OpenAI models: %v", err)
		fetcherErrors = append(fetcherErrors, errMsg)
	} else {
		allModels = append(allModels, openAIModels...)
	}

	// Fetch from Google
	googleModels, err := s.googleFetcher.FetchModels()
	if err != nil {
		errMsg := fmt.Sprintf("Error fetching Google models: %v", err)
		fetcherErrors = append(fetcherErrors, errMsg)
	} else {
		allModels = append(allModels, googleModels...)
	}

	// Fetch from Anthropic
	anthropicModels, err := s.anthropicFetcher.FetchModels()
	if err != nil {
		// This fetcher currently returns hardcoded data, so an error is unexpected
		// unless the method signature changes or an internal issue occurs.
		errMsg := fmt.Sprintf("Error fetching Anthropic models: %v", err)
		fetcherErrors = append(fetcherErrors, errMsg)
	} else {
		allModels = append(allModels, anthropicModels...)
	}

	// Fetch from Ollama (local models)
	// Note: This is optional as Ollama may not be running
	if s.ollamaFetcher != nil {
		ollamaModels, err := s.ollamaFetcher.FetchModels()
		if err != nil {
			// Don't treat Ollama errors as critical since it's a local service
			// that may not always be running
			errMsg := fmt.Sprintf("Error fetching Ollama models (local service may not be running): %v", err)
			fetcherErrors = append(fetcherErrors, errMsg)
		} else {
			allModels = append(allModels, ollamaModels...)
		}
	}

	// Fetch from OpenRouter
	// Note: This is optional as it requires an API key
	if s.openRouterFetcher != nil {
		openRouterModels, err := s.openRouterFetcher.FetchModels()
		if err != nil {
			// Don't treat OpenRouter errors as critical since it requires an API key
			errMsg := fmt.Sprintf("Error fetching OpenRouter models (API key may not be configured): %v", err)
			fetcherErrors = append(fetcherErrors, errMsg)
		} else {
			allModels = append(allModels, openRouterModels...)
		}
	}

	// Populate metadata
	metadata := domain.Metadata{
		Version:       "1.0.0",
		LastUpdated:   time.Now().Format("2006-01-02"),
		Description:   "Aggregated inventory of LLM models.",
		SchemaVersion: "1", // Assuming schema version 1 for now
	}

	inventory := &domain.ModelInventory{
		Metadata: metadata,
		Models:   allModels,
	}

	if len(fetcherErrors) > 0 {
		// Return a general error if any fetcher failed.
		// For more detailed error handling, a multi-error type could be used.
		overallErr = fmt.Errorf("one or more fetchers failed to retrieve model data; %d errors occurred", len(fetcherErrors))
	}

	return inventory, overallErr
}
