package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"gopkg.in/yaml.v3"
)

// SimpleConfig holds our configuration
var config Config

// Init config with defaults
func init() {
	config = DefaultConfig()
}

// LoadConfig loads configuration from file and environment
func LoadConfig(configFile string) error {
	// Try to load from file
	if configFile != "" {
		if err := loadYAMLFile(configFile); err == nil {
			fmt.Printf("Using config file: %s\n", configFile)
		}
	} else {
		// Try standard locations
		home, _ := os.UserHomeDir()
		configPaths := []string{
			filepath.Join(home, ".go-llms.yaml"),
			".go-llms.yaml",
			filepath.Join(home, ".config", "go-llms", "config.yaml"),
		}

		for _, path := range configPaths {
			if _, err := os.Stat(path); err == nil {
				if err := loadYAMLFile(path); err == nil {
					fmt.Printf("Using config file: %s\n", path)
					break
				}
			}
		}
	}

	// Override with environment variables
	loadEnvVars()

	return nil
}

// loadYAMLFile loads configuration from a YAML file
func loadYAMLFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	return yaml.Unmarshal(data, &config)
}

// loadEnvVars loads configuration from environment variables
func loadEnvVars() {
	// Standard format: GO_LLMS_PROVIDER, GO_LLMS_MODEL, etc.
	if val := os.Getenv("GO_LLMS_PROVIDER"); val != "" {
		config.Provider = val
	}
	if val := os.Getenv("GO_LLMS_MODEL"); val != "" {
		config.Model = val
	}
	if val := os.Getenv("GO_LLMS_VERBOSE"); val == "true" {
		config.Verbose = true
	}
	if val := os.Getenv("GO_LLMS_OUTPUT"); val != "" {
		config.Output = val
	}

	// Provider-specific settings
	providers := []string{"openai", "anthropic", "gemini"}
	for _, provider := range providers {
		upperProvider := strings.ToUpper(provider)
		
		// API Keys - both formats
		if val := os.Getenv(fmt.Sprintf("GO_LLMS_PROVIDERS_%s_API_KEY", upperProvider)); val != "" {
			setProviderAPIKey(provider, val)
		}
		if val := os.Getenv(fmt.Sprintf("%s_API_KEY", upperProvider)); val != "" {
			setProviderAPIKey(provider, val)
		}
		
		// Default models
		if val := os.Getenv(fmt.Sprintf("GO_LLMS_PROVIDERS_%s_DEFAULT_MODEL", upperProvider)); val != "" {
			setProviderDefaultModel(provider, val)
		}
	}
}

// Helper functions to set provider config
func setProviderAPIKey(provider, apiKey string) {
	switch provider {
	case "openai":
		config.Providers.OpenAI.APIKey = apiKey
	case "anthropic":
		config.Providers.Anthropic.APIKey = apiKey
	case "gemini":
		config.Providers.Gemini.APIKey = apiKey
	}
}

func setProviderDefaultModel(provider, model string) {
	switch provider {
	case "openai":
		config.Providers.OpenAI.DefaultModel = model
	case "anthropic":
		config.Providers.Anthropic.DefaultModel = model
	case "gemini":
		config.Providers.Gemini.DefaultModel = model
	}
}

// GetSimpleAPIKey retrieves the API key for a provider
func GetSimpleAPIKey(provider string) (string, error) {
	var key string
	
	switch provider {
	case "openai":
		key = config.Providers.OpenAI.APIKey
	case "anthropic":
		key = config.Providers.Anthropic.APIKey
	case "gemini":
		key = config.Providers.Gemini.APIKey
	}
	
	if key == "" {
		// Try environment variable as fallback
		envVar := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
		key = os.Getenv(envVar)
		if key == "" {
			return "", fmt.Errorf("no API key configured for provider %s. Set it in config file or %s environment variable", provider, envVar)
		}
	}
	return key, nil
}

// GetSimpleProvider returns the configured provider and model
func GetSimpleProvider() (string, string, error) {
	provider := config.Provider
	model := config.Model

	// If no model specified, get the default for the provider
	if model == "" {
		switch provider {
		case "openai":
			model = config.Providers.OpenAI.DefaultModel
		case "anthropic":
			model = config.Providers.Anthropic.DefaultModel
		case "gemini":
			model = config.Providers.Gemini.DefaultModel
		}
		
		if model == "" {
			return "", "", fmt.Errorf("no model specified and no default model configured for provider %s", provider)
		}
	}

	return provider, model, nil
}