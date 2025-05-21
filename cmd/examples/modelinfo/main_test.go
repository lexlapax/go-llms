package main

import (
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llmutil"
	"github.com/lexlapax/go-llms/pkg/modelinfo/domain"
)

// TestModelInfoExample is a simple test to verify the example code works correctly
func TestModelInfoExample(t *testing.T) {
	// Check if we're running in short mode - some CI environments may not have API keys set
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Test fetching models with default options
	opts := &llmutil.GetAvailableModelsOptions{
		UseCache:    true,
		MaxCacheAge: 1 * time.Hour,
	}

	inventory, err := llmutil.GetAvailableModels(opts)
	if err != nil {
		// Allow partial failures if some API keys are missing
		t.Logf("Warning: %v", err)
	}

	if inventory == nil {
		t.Fatal("Expected to get an inventory object, got nil")
	}

	// At minimum we should have Anthropic models from hardcoded data
	anthropicFound := false
	for _, model := range inventory.Models {
		if model.Provider == "anthropic" {
			anthropicFound = true
			break
		}
	}

	if !anthropicFound {
		t.Error("Expected to find at least one Anthropic model")
	}

	// Test the filtering function for different capabilities
	cfg := config{
		Provider:    "",
		NamePattern: "",
		Capability:  "text-output",
	}

	filteredModels := filterModels(inventory.Models, cfg)
	if len(filteredModels) == 0 {
		t.Error("Expected to find at least one model with text output capability")
	}

	// Test that we can identify text-processing models
	textModelsCount := 0
	for _, model := range inventory.Models {
		if model.Capabilities.Text.Read && model.Capabilities.Text.Write {
			textModelsCount++
		}
	}

	if textModelsCount == 0 {
		t.Error("Expected to find at least one model with text read/write capabilities")
	}

	// Check that metadata is populated
	if inventory.Metadata.Version == "" {
		t.Error("Expected metadata.Version to be set")
	}
	if inventory.Metadata.LastUpdated == "" {
		t.Error("Expected metadata.LastUpdated to be set")
	}
}

// TestHasCapability verifies the capability filter function
func TestHasCapability(t *testing.T) {
	// Create test capabilities
	cap := domain.Capabilities{
		Text: domain.MediaTypeCapability{
			Read:  true,
			Write: true,
		},
		Image: domain.MediaTypeCapability{
			Read:  true,
			Write: false,
		},
		Audio: domain.MediaTypeCapability{
			Read:  false,
			Write: true,
		},
		FunctionCalling: true,
		Streaming:       true,
		JSONMode:        false,
	}

	tests := []struct {
		capability string
		expected   bool
	}{
		{"text-input", true},
		{"text-output", true},
		{"image-input", true},
		{"image-output", false},
		{"audio-input", false},
		{"audio-output", true},
		{"video-input", false},
		{"video-output", false},
		{"function-calling", true},
		{"streaming", true},
		{"json-mode", false},
		{"invalid-capability", false},
	}

	for _, test := range tests {
		t.Run(test.capability, func(t *testing.T) {
			if hasCapability(cap, test.capability) != test.expected {
				t.Errorf("hasCapability(%s) = %v; want %v", 
					test.capability, !test.expected, test.expected)
			}
		})
	}
}