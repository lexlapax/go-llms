package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents our application configuration
type Config struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	Verbose  bool   `yaml:"verbose"`
	Output   string `yaml:"output"`

	Providers struct {
		OpenAI struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		} `yaml:"openai"`

		Anthropic struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		} `yaml:"anthropic"`

		Gemini struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		} `yaml:"gemini"`
	} `yaml:"providers"`
}

// Global config instance
var config Config

// InitOptimizedConfig loads configuration from file and environment
func InitOptimizedConfig(configFile string) error {
	// Set defaults
	config = Config{
		Provider: "openai",
		Output:   "text",
	}
	config.Providers.OpenAI.DefaultModel = "gpt-4o"
	config.Providers.Anthropic.DefaultModel = "claude-3-5-sonnet-latest"
	config.Providers.Gemini.DefaultModel = "gemini-2.0-flash-lite"

	// Load from config file
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
	loadProviderEnvVars("openai", "OPENAI")
	loadProviderEnvVars("anthropic", "ANTHROPIC")
	loadProviderEnvVars("gemini", "GEMINI")

	// Also check for standard API key environment variables (backward compatibility)
	if val := os.Getenv("OPENAI_API_KEY"); val != "" && config.Providers.OpenAI.APIKey == "" {
		config.Providers.OpenAI.APIKey = val
	}
	if val := os.Getenv("ANTHROPIC_API_KEY"); val != "" && config.Providers.Anthropic.APIKey == "" {
		config.Providers.Anthropic.APIKey = val
	}
	if val := os.Getenv("GEMINI_API_KEY"); val != "" && config.Providers.Gemini.APIKey == "" {
		config.Providers.Gemini.APIKey = val
	}
}

// loadProviderEnvVars loads provider-specific environment variables
func loadProviderEnvVars(provider, envPrefix string) {
	// API Key
	envVar := fmt.Sprintf("GO_LLMS_PROVIDERS_%s_API_KEY", envPrefix)
	if val := os.Getenv(envVar); val != "" {
		switch provider {
		case "openai":
			config.Providers.OpenAI.APIKey = val
		case "anthropic":
			config.Providers.Anthropic.APIKey = val
		case "gemini":
			config.Providers.Gemini.APIKey = val
		}
	}

	// Default Model
	envVar = fmt.Sprintf("GO_LLMS_PROVIDERS_%s_DEFAULT_MODEL", envPrefix)
	if val := os.Getenv(envVar); val != "" {
		switch provider {
		case "openai":
			config.Providers.OpenAI.DefaultModel = val
		case "anthropic":
			config.Providers.Anthropic.DefaultModel = val
		case "gemini":
			config.Providers.Gemini.DefaultModel = val
		}
	}
}

// GetOptimizedAPIKey retrieves the API key for a provider
func GetOptimizedAPIKey(provider string) (string, error) {
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

// GetOptimizedProvider returns the configured provider and model
func GetOptimizedProvider() (string, string, error) {
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
