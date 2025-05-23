package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestGeminiExample(t *testing.T) {
	// Skip test if no API key provided
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create generation config option
	generationConfigOption := domain.NewGeminiGenerationConfigOption().
		WithTemperature(0.7).
		WithTopK(40).
		WithMaxOutputTokens(256)

	// Create provider with options
	geminiProvider := provider.NewGeminiProvider(
		apiKey,
		"gemini-2.0-flash-lite",
		generationConfigOption,
	)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test simple text generation
	response, err := geminiProvider.Generate(ctx, "Say hello!")
	if err != nil {
		t.Fatalf("Error generating text: %v", err)
	}

	if response == "" {
		t.Error("Empty response received from Gemini API")
	} else {
		t.Logf("Received response: %s", response)
	}
}
