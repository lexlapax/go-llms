package fetchers

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

func TestAnthropicFetcher_FetchModels(t *testing.T) {
	fetcher := AnthropicFetcher{}
	models, err := fetcher.FetchModels()

	if err != nil {
		t.Fatalf("FetchModels() returned an unexpected error: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("FetchModels() returned an empty slice of models, expected some models.")
	}

	// Check for a specific number of models if it's stable
	// For now, we know it's 4 based on the hardcoded list.
	expectedModelCount := 4
	if len(models) != expectedModelCount {
		t.Errorf("FetchModels() returned %d models, expected %d", len(models), expectedModelCount)
	}

	// Test a specific model - Claude 3 Opus
	var opusModel *domain.Model
	for i := range models {
		if models[i].Name == "claude-3-opus-20240229" {
			opusModel = &models[i]
			break
		}
	}

	if opusModel == nil {
		t.Fatal("Expected model 'claude-3-opus-20240229' not found in the results.")
	}

	if opusModel.Provider != "anthropic" {
		t.Errorf("Expected Opus model provider to be 'anthropic', got '%s'", opusModel.Provider)
	}
	if opusModel.DisplayName != "Claude 3 Opus" {
		t.Errorf("Expected Opus model DisplayName to be 'Claude 3 Opus', got '%s'", opusModel.DisplayName)
	}
	if opusModel.ContextWindow != 200000 {
		t.Errorf("Expected Opus model ContextWindow to be 200000, got %d", opusModel.ContextWindow)
	}
	if opusModel.ModelFamily != "claude-3" {
		t.Errorf("Expected Opus model ModelFamily to be 'claude-3', got '%s'", opusModel.ModelFamily)
	}
	if !opusModel.Capabilities.FunctionCalling {
		t.Error("Expected Opus model Capabilities.FunctionCalling to be true")
	}
	if opusModel.Pricing.InputPer1kTokens != 0.015 {
		t.Errorf("Expected Opus model Pricing.InputPer1kTokens to be 0.015, got %f", opusModel.Pricing.InputPer1kTokens)
	}

	// Test another specific model - Claude 2.1
	var claude21Model *domain.Model
	for i := range models {
		if models[i].Name == "claude-2.1" {
			claude21Model = &models[i]
			break
		}
	}

	if claude21Model == nil {
		t.Fatal("Expected model 'claude-2.1' not found in the results.")
	}

	if claude21Model.Provider != "anthropic" {
		t.Errorf("Expected Claude 2.1 model provider to be 'anthropic', got '%s'", claude21Model.Provider)
	}
	if claude21Model.ContextWindow != 200000 {
		t.Errorf("Expected Claude 2.1 model ContextWindow to be 200000, got %d", claude21Model.ContextWindow)
	}
	if claude21Model.ModelFamily != "claude-2" {
		t.Errorf("Expected Claude 2.1 model ModelFamily to be 'claude-2', got '%s'", claude21Model.ModelFamily)
	}
	if claude21Model.Capabilities.FunctionCalling {
		t.Error("Expected Claude 2.1 model Capabilities.FunctionCalling to be false")
	}
}
