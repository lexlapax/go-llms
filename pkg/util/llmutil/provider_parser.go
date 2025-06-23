// ABOUTME: Provider string parser for creating LLM providers from simple string specifications
// ABOUTME: Supports formats like "openai/gpt-4", aliases, and automatic API key resolution

package llmutil

import (
	"fmt"
	"strings"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// ParseProviderModelString parses a provider/model specification string into separate provider and model components.
// It supports multiple formats and can infer missing information from known patterns.
//
// Formats supported:
//   - "provider/model" (e.g., "openai/gpt-4")
//   - "model" (e.g., "gpt-4" - provider inferred)
//   - "provider" (e.g., "openai" - model from env or default)
func ParseProviderModelString(spec string) (provider, model string, err error) {
	if spec == "" {
		return "", "", fmt.Errorf("empty provider/model specification")
	}

	// Handle special cases
	if spec == "mock" {
		return "mock", "", nil
	}

	// Split by "/" to check for provider/model format
	parts := strings.SplitN(spec, "/", 2)

	if len(parts) == 2 {
		// Format: provider/model
		provider = normalizeProvider(parts[0])
		model = parts[1]
		return provider, model, nil
	}

	// Single part - could be provider or model
	part := parts[0]

	// Check if it's a known provider first
	if isKnownProvider(part) {
		provider = normalizeProvider(part)
		// Model will be determined from environment or defaults
		return provider, "", nil
	}

	// Check if it's a known alias
	if fullSpec, ok := modelAliases[part]; ok {
		return ParseProviderModelString(fullSpec)
	}

	// Try to infer provider from model name
	inferredProvider := inferProviderFromModel(part)
	if inferredProvider != "" {
		return inferredProvider, part, nil
	}

	// Unable to parse
	return "", "", fmt.Errorf("unable to parse '%s' - use format 'provider/model' (e.g., 'openai/gpt-4')", spec)
}

// ParseProviderModelWithOptions parses extended format with options
// Format: "provider/model:option" (e.g., "openai/gpt-4:streaming")
func ParseProviderModelWithOptions(spec string) (provider, model, useCase string, err error) {
	// Split by ":" to separate options
	parts := strings.SplitN(spec, ":", 2)

	// Parse the base spec
	provider, model, err = ParseProviderModelString(parts[0])
	if err != nil {
		return "", "", "", err
	}

	// Extract use case if provided
	if len(parts) == 2 {
		useCase = parts[1]
	}

	return provider, model, useCase, nil
}

// NewProviderFromString creates a provider from a string specification.
// It parses provider/model strings, handles aliases, infers missing information,
// and automatically retrieves API keys from environment variables.
// Supports formats like "openai/gpt-4", "gpt-4", "claude", "openai/gpt-4:streaming".
func NewProviderFromString(spec string) (domain.Provider, error) {
	// Parse the specification with options
	provider, model, useCase, err := ParseProviderModelWithOptions(spec)
	if err != nil {
		return nil, err
	}

	// Create model config
	config := ModelConfig{
		Provider: provider,
		Model:    model,
		UseCase:  useCase,
	}

	// CreateProvider will handle API key lookup from environment
	llmProvider, err := CreateProvider(config)
	if err != nil {
		// Enhance error message with helpful hints
		if strings.Contains(err.Error(), "API key") {
			return nil, fmt.Errorf("%w. Set %s or %s environment variable",
				err,
				getAPIKeyEnvVar(provider),
				getGoLLMsAPIKeyEnvVar(provider))
		}
		return nil, err
	}

	return llmProvider, nil
}

// normalizeProvider converts provider aliases to canonical names
func normalizeProvider(provider string) string {
	// Don't normalize gemini to google here - let CreateProvider handle it
	return strings.ToLower(provider)
}

// isKnownProvider checks if a string is a known provider name
func isKnownProvider(s string) bool {
	normalized := normalizeProvider(s)
	switch normalized {
	case "openai", "anthropic", "google", "gemini", "mock", "ollama", "openrouter", "vertexai":
		return true
	default:
		return false
	}
}

// inferProviderFromModel attempts to determine provider from model name patterns
func inferProviderFromModel(model string) string {
	model = strings.ToLower(model)

	// Check prefixes
	for prefix, provider := range modelPrefixToProvider {
		if strings.HasPrefix(model, prefix) {
			return provider
		}
	}

	// Check if model contains certain keywords
	for keyword, provider := range modelKeywordToProvider {
		if strings.Contains(model, keyword) {
			return provider
		}
	}

	return ""
}

// getAPIKeyEnvVar returns the standard API key environment variable name
func getAPIKeyEnvVar(provider string) string {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "google", "gemini":
		return "GEMINI_API_KEY"
	case "ollama":
		return "OLLAMA_API_KEY"
	case "openrouter":
		return "OPENROUTER_API_KEY"
	default:
		return strings.ToUpper(provider) + "_API_KEY"
	}
}

// getGoLLMsAPIKeyEnvVar returns the go-llms specific API key environment variable name
func getGoLLMsAPIKeyEnvVar(provider string) string {
	switch provider {
	case "openai":
		return "GO_LLMS_OPENAI_API_KEY"
	case "anthropic":
		return "GO_LLMS_ANTHROPIC_API_KEY"
	case "google", "gemini":
		return "GO_LLMS_GEMINI_API_KEY"
	case "ollama":
		return "GO_LLMS_OLLAMA_API_KEY"
	case "openrouter":
		return "GO_LLMS_OPENROUTER_API_KEY"
	default:
		return "GO_LLMS_" + strings.ToUpper(provider) + "_API_KEY"
	}
}

// Model name patterns for provider detection
var modelPrefixToProvider = map[string]string{
	"gpt-":            "openai",
	"o1-":             "openai",
	"dall-e":          "openai",
	"text-embedding":  "openai",
	"chatgpt":         "openai",
	"claude-":         "anthropic",
	"gemini-":         "gemini",
	"embedding-gecko": "gemini",
	"llama":           "ollama",
	"mistral":         "ollama",
	"codellama":       "ollama",
	"gemma":           "ollama",
	"qwen":            "ollama",
	"phi":             "ollama",
}

var modelKeywordToProvider = map[string]string{
	"turbo": "openai", // for gpt-3.5-turbo, gpt-4-turbo
}

// Model aliases for convenience
// You can easily modify these aliases as needed
var modelAliases = map[string]string{
	// OpenAI aliases
	"gpt-3.5":      "openai/gpt-3.5-turbo",
	"gpt-4":        "openai/gpt-4",
	"gpt-4o":       "openai/gpt-4o",
	"gpt-4.1":      "openai/gpt-4.1",           // convenience alias
	"gpt-4.1-mini": "openai/gpt-4.1-mini",      // convenience alias
	"gpt-4.1-nano": "openai/gpt-4.1-nano",      // convenience alias
	"o1":           "openai/o1",                // convenience alias
	"o1-mini":      "openai/o1-mini",           // convenience alias
	"o1-pro":       "openai/o1-pro",            // convenience alias
	"o4-mini":      "openai/o4-mini",           // convenience alias
	"o3":           "openai/o3",                // convenience alias
	"o3-mini":      "openai/o3-mini",           // convenience alias
	"codex":        "openai/codex-mini-latest", // convenience alias
	"codex-mini":   "openai/codex-mini-latest", // convenience alias

	// Anthropic aliases
	"claude":          "anthropic/claude-3-7-sonnet-latest",
	"claude-3":        "anthropic/claude-3-7-sonnet-latest",
	"claude-3-opus":   "anthropic/claude-3-opus-latest",
	"claude-3-sonnet": "anthropic/claude-3-7-sonnet-latest",
	"claude-3-haiku":  "anthropic/claude-3-5-haiku-latest",
	"claude-4-opus":   "anthropic/claude-opus-4-20250514",
	"claude-4-sonnet": "anthropic/claude-sonnet-4-20250514",
	"opus":            "anthropic/claude-3-opus-latest",
	"opus-4":          "anthropic/claude-opus-4-20250514",
	"sonnet":          "anthropic/claude-3-7-sonnet-latest",
	"sonnet-4":        "anthropic/claude-sonnet-4-20250514",
	"haiku":           "anthropic/claude-3-5-haiku-latest",

	// Google/Gemini aliases
	// Note: "gemini" is now treated as a provider, not an alias
	"gemini-pro":       "gemini/gemini-2.5-pro-preview-05-06",
	"gemini-flash":     "gemini/gemini-2.5-flash-preview-05-20",
	"gemini-2.5":       "gemini/gemini-2.5-pro-preview-05-06",
	"gemini-2.5-flash": "gemini/gemini-2.5-flash-preview-05-20",
	"gemini-2.5-pro":   "gemini/gemini-2.5-pro-preview-05-06",
	"flash":            "gemini/gemini-2.0-flash",
	"flash-lite":       "gemini/gemini-2.0-flash-lite",

	// Ollama aliases
	"llama3":        "ollama/llama3.2:3b",
	"llama3.2":      "ollama/llama3.2:3b",
	"mistral":       "ollama/mistral:7b",
	"mistral-7b":    "ollama/mistral:7b",
	"codellama":     "ollama/codellama:13b",
	"codellama-13b": "ollama/codellama:13b",
	"gemma2":        "ollama/gemma2:2b",
	"qwen2":         "ollama/qwen2.5:7b",
	"phi3":          "ollama/phi3:mini",

	// OpenRouter aliases
	"deepseek-qwen3":       "deepseek/deepseek-r1-0528-qwen3-8b:free",
	"deepseek":             "deepseek/deepseek-r1-0528:free",
	"openrouter/codellama": "openrouter/ollama/codellama:13b",
}
