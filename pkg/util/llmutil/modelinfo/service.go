package modelinfo

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/fetchers"
)

// ModelInfoService aggregates model information from various providers.
type ModelInfoService struct {
	openAIFetcher    *fetchers.OpenAIFetcher
	googleFetcher    *fetchers.GoogleFetcher
	anthropicFetcher *fetchers.AnthropicFetcher
	// Add other fetchers here if more providers are supported
}

// NewServiceWithCustomFetchers creates a ModelInfoService with specific fetcher instances.
// Useful for testing or custom provider configurations.
func NewServiceWithCustomFetchers(
	openAIFetcher *fetchers.OpenAIFetcher,
	googleFetcher *fetchers.GoogleFetcher,
	anthropicFetcher *fetchers.AnthropicFetcher,
) *ModelInfoService {
	return &ModelInfoService{
		openAIFetcher:    openAIFetcher,
		googleFetcher:    googleFetcher,
		anthropicFetcher: anthropicFetcher,
	}
}

// defaultNewModelInfoService is the default implementation for creating a ModelInfoService.
func defaultNewModelInfoService() *ModelInfoService {
	return NewServiceWithCustomFetchers(
		fetchers.NewOpenAIFetcher("", http.DefaultClient), // Uses default internal URL
		fetchers.NewGoogleFetcher("", http.DefaultClient), // Uses default internal URL
		&fetchers.AnthropicFetcher{},                      // Remains as is
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
		fmt.Printf("%s\n", errMsg) // Placeholder for actual logging
		fetcherErrors = append(fetcherErrors, errMsg)
	} else {
		allModels = append(allModels, openAIModels...)
	}

	// Fetch from Google
	googleModels, err := s.googleFetcher.FetchModels()
	if err != nil {
		errMsg := fmt.Sprintf("Error fetching Google models: %v", err)
		fmt.Printf("%s\n", errMsg) // Placeholder for actual logging
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
		fmt.Printf("%s\n", errMsg) // Placeholder for actual logging
		fetcherErrors = append(fetcherErrors, errMsg)
	} else {
		allModels = append(allModels, anthropicModels...)
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
