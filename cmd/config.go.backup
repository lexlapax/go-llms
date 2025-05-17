package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

// Global koanf instance
var k = koanf.New(".")

// Config represents our application configuration
type Config struct {
	Provider string `koanf:"provider" json:"provider"`
	Model    string `koanf:"model" json:"model"`
	Verbose  bool   `koanf:"verbose" json:"verbose"`
	Output   string `koanf:"output" json:"output"`

	Providers struct {
		OpenAI struct {
			APIKey       string `koanf:"api_key" json:"api_key"`
			DefaultModel string `koanf:"default_model" json:"default_model"`
		} `koanf:"openai" json:"openai"`

		Anthropic struct {
			APIKey       string `koanf:"api_key" json:"api_key"`
			DefaultModel string `koanf:"default_model" json:"default_model"`
		} `koanf:"anthropic" json:"anthropic"`

		Gemini struct {
			APIKey       string `koanf:"api_key" json:"api_key"`
			DefaultModel string `koanf:"default_model" json:"default_model"`
		} `koanf:"gemini" json:"gemini"`
	} `koanf:"providers" json:"providers"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Provider: "openai",
		Output:   "text",
		Providers: struct {
			OpenAI struct {
				APIKey       string `koanf:"api_key" json:"api_key"`
				DefaultModel string `koanf:"default_model" json:"default_model"`
			} `koanf:"openai" json:"openai"`
			Anthropic struct {
				APIKey       string `koanf:"api_key" json:"api_key"`
				DefaultModel string `koanf:"default_model" json:"default_model"`
			} `koanf:"anthropic" json:"anthropic"`
			Gemini struct {
				APIKey       string `koanf:"api_key" json:"api_key"`
				DefaultModel string `koanf:"default_model" json:"default_model"`
			} `koanf:"gemini" json:"gemini"`
		}{
			OpenAI: struct {
				APIKey       string `koanf:"api_key" json:"api_key"`
				DefaultModel string `koanf:"default_model" json:"default_model"`
			}{
				DefaultModel: "gpt-4o",
			},
			Anthropic: struct {
				APIKey       string `koanf:"api_key" json:"api_key"`
				DefaultModel string `koanf:"default_model" json:"default_model"`
			}{
				DefaultModel: "claude-3-5-sonnet-latest",
			},
			Gemini: struct {
				APIKey       string `koanf:"api_key" json:"api_key"`
				DefaultModel string `koanf:"default_model" json:"default_model"`
			}{
				DefaultModel: "gemini-2.0-flash-lite",
			},
		},
	}
}

// InitConfig loads configuration from various sources
func InitConfig(configFile string) error {
	// Load defaults
	if err := k.Load(structs.Provider(DefaultConfig(), "koanf"), nil); err != nil {
		return fmt.Errorf("error loading default config: %w", err)
	}

	// Load from config file if specified
	if configFile != "" {
		if err := k.Load(file.Provider(configFile), yaml.Parser()); err == nil {
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
				if err := k.Load(file.Provider(path), yaml.Parser()); err == nil {
					fmt.Printf("Using config file: %s\n", path)
					break
				}
			}
		}
	}

	// Load environment variables
	// Map GO_LLMS_PROVIDER to provider, GO_LLMS_PROVIDERS_OPENAI_API_KEY to providers.openai.api_key
	if err := k.Load(env.Provider("GO_LLMS_", ".", func(s string) string {
		return strings.Replace(
			strings.ToLower(strings.TrimPrefix(s, "GO_LLMS_")),
			"_", ".", -1)
	}), nil); err != nil {
		return fmt.Errorf("error loading environment variables: %w", err)
	}

	// Also check for standard API key environment variables
	// This preserves backward compatibility with existing scripts
	envMappings := map[string]string{
		"OPENAI_API_KEY":    "providers.openai.api_key",
		"ANTHROPIC_API_KEY": "providers.anthropic.api_key",
		"GEMINI_API_KEY":    "providers.gemini.api_key",
	}

	for envVar, configKey := range envMappings {
		if val := os.Getenv(envVar); val != "" && k.String(configKey) == "" {
			k.Set(configKey, val)
		}
	}

	return nil
}

// GetAPIKey retrieves the API key for a provider
func GetAPIKey(provider string) (string, error) {
	key := k.String(fmt.Sprintf("providers.%s.api_key", provider))
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

// GetProvider returns the configured provider and model
func GetProvider() (string, string, error) {
	provider := k.String("provider")
	model := k.String("model")

	// If no model specified, get the default for the provider
	if model == "" {
		model = k.String(fmt.Sprintf("providers.%s.default_model", provider))
		if model == "" {
			return "", "", fmt.Errorf("no model specified and no default model configured for provider %s", provider)
		}
	}

	return provider, model, nil
}