package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/llmutil"
	"github.com/lexlapax/go-llms/pkg/modelinfo/domain"
)

// ABOUTME: Example application demonstrating the ModelInfo service
// ABOUTME: Fetches and displays model capabilities from various LLM providers

type config struct {
	Provider     string
	NamePattern  string
	Capability   string
	FreshData    bool
	CachePath    string
	PrettyPrint  bool
	ShowMetadata bool
}

func main() {
	// Parse command line flags
	cfg := parseFlags()

	// Setup the model inventory options
	opts := &llmutil.GetAvailableModelsOptions{
		UseCache:    !cfg.FreshData,
		CachePath:   cfg.CachePath,
		MaxCacheAge: 24 * time.Hour, // Cache for 24 hours by default
	}

	// Check for API keys and warn if missing
	checkAPIKeys()

	// Fetch all available models
	fmt.Fprintln(os.Stderr, "Fetching model information...")
	inventory, err := llmutil.GetAvailableModels(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching model information: %v\n", err)
		// We'll still try to process what we have, even if there was an error
		// There might be partial data from caches or hardcoded sources
		if inventory == nil {
			os.Exit(1)
		}
	}

	// Filter models based on command line flags
	filteredModels := filterModels(inventory.Models, cfg)

	// Prepare output based on selected options
	var output interface{}
	if cfg.ShowMetadata {
		// Include metadata in output
		output = domain.ModelInventory{
			Metadata: inventory.Metadata,
			Models:   filteredModels,
		}
	} else {
		// Only show models
		output = filteredModels
	}

	// Print the result as JSON
	writeJSON(output, cfg.PrettyPrint)
}

func parseFlags() config {
	cfg := config{}

	// Define command line flags
	flag.StringVar(&cfg.Provider, "provider", "", "Filter by provider (openai, anthropic, gemini)")
	flag.StringVar(&cfg.NamePattern, "name", "", "Filter by model name pattern")
	flag.StringVar(&cfg.Capability, "capability", "",
		"Filter by capability (text-input, text-output, image-input, image-output, "+
			"audio-input, audio-output, video-input, video-output, function-calling, streaming, json-mode)")
	flag.BoolVar(&cfg.FreshData, "fresh", false, "Force fresh data (ignore cache)")
	flag.StringVar(&cfg.CachePath, "cache-path", "", "Custom cache file location")
	flag.BoolVar(&cfg.PrettyPrint, "pretty", true, "Pretty print JSON output")
	flag.BoolVar(&cfg.ShowMetadata, "metadata", false, "Include metadata in output")

	// Parse flags
	flag.Parse()

	return cfg
}

func filterModels(models []domain.Model, cfg config) []domain.Model {
	var result []domain.Model

	for _, model := range models {
		// Filter by provider
		if cfg.Provider != "" && !strings.EqualFold(model.Provider, cfg.Provider) {
			continue
		}

		// Filter by name pattern
		if cfg.NamePattern != "" && !strings.Contains(
			strings.ToLower(model.Name),
			strings.ToLower(cfg.NamePattern)) {
			continue
		}

		// Filter by capability
		if cfg.Capability != "" && !hasCapability(model.Capabilities, cfg.Capability) {
			continue
		}

		// All filters passed, include the model
		result = append(result, model)
	}

	return result
}

func hasCapability(capabilities domain.Capabilities, capabilityName string) bool {
	switch capabilityName {
	case "text-input":
		return capabilities.Text.Read
	case "text-output":
		return capabilities.Text.Write
	case "image-input":
		return capabilities.Image.Read
	case "image-output":
		return capabilities.Image.Write
	case "audio-input":
		return capabilities.Audio.Read
	case "audio-output":
		return capabilities.Audio.Write
	case "video-input":
		return capabilities.Video.Read
	case "video-output":
		return capabilities.Video.Write
	case "function-calling":
		return capabilities.FunctionCalling
	case "streaming":
		return capabilities.Streaming
	case "json-mode":
		return capabilities.JSONMode
	default:
		return false
	}
}

func writeJSON(data interface{}, pretty bool) {
	var bytes []byte
	var err error

	if pretty {
		bytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(bytes))
}

func checkAPIKeys() {
	// Check for provider API keys and warn if missing
	openaiKey := os.Getenv("OPENAI_API_KEY")
	// Anthropic has hardcoded data, so not strictly required
	geminiKey := os.Getenv("GEMINI_API_KEY")

	var missingKeys []string

	if openaiKey == "" {
		missingKeys = append(missingKeys, "OPENAI_API_KEY")
	}
	if geminiKey == "" {
		missingKeys = append(missingKeys, "GEMINI_API_KEY")
	}

	// Only show a warning if keys are missing
	if len(missingKeys) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: The following API keys are not set: %s\n", strings.Join(missingKeys, ", "))
		fmt.Fprintln(os.Stderr, "Only hardcoded model data or previously cached data may be available for these providers.")
	}
}
