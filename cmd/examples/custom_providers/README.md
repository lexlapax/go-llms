# Custom Providers Example

This example demonstrates how to use go-llms with custom LLM providers by leveraging the flexibility of the library's provider system. It shows two common use cases:

1. **OpenRouter** - A service that provides access to various LLM providers through an OpenAI-compatible API
2. **Ollama** - A tool for running LLMs locally, also with an OpenAI-compatible API

## Overview

The go-llms library makes it easy to use providers that implement the same API as supported providers (like OpenAI) by allowing you to override the base URL. This is particularly useful for:

- Using API gateways or proxies in front of supported providers
- Working with self-hosted or local models
- Integrating with providers that offer OpenAI-compatible APIs

## Three Ways to Configure Custom Providers

### Method 1: Direct Provider Instantiation

Use provider-specific option functions when creating a provider:

```go
provider := provider.NewOpenAIProvider(
    apiKey,
    modelName,
    provider.WithBaseURL("https://custom-endpoint.com"),
)
```

### Method 2: Using ModelConfig and CreateProvider

Use the convenience function with ModelConfig:

```go
config := llmutil.ModelConfig{
    Provider: "openai",
    Model:    modelName,
    APIKey:   apiKey,
    BaseURL:  "https://custom-endpoint.com",
}

customProvider, err := llmutil.CreateProvider(config)
```

### Method 3: Using Environment Variables

Set the appropriate environment variables and use ProviderFromEnv():

```bash
export OPENAI_BASE_URL="https://custom-endpoint.com"
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL="your-model-name"
```

```go
provider, providerName, modelName, err := llmutil.ProviderFromEnv()
```

## OpenRouter Example

[OpenRouter](https://openrouter.ai/) provides a unified API to access different LLM providers. It uses an OpenAI-compatible API, making it easy to integrate with go-llms.

### Configuration

Set the following environment variables:

```bash
# Required
export OPENROUTER_API_KEY="your-openrouter-api-key"

# Optional - defaults to anthropic/claude-3-5-sonnet
export OPENROUTER_MODEL="anthropic/claude-3-5-sonnet"

# For Method 3
export OPENAI_BASE_URL="https://openrouter.ai/api"
export OPENAI_API_KEY="your-openrouter-api-key"
export OPENAI_MODEL="anthropic/claude-3-5-sonnet"
```

## Ollama Example

[Ollama](https://ollama.ai/) lets you run LLMs locally. It also provides an OpenAI-compatible API that integrates well with go-llms.

### Configuration

Set the following environment variables:

```bash
# Optional - defaults to http://localhost:11434
export OLLAMA_HOST="http://localhost:11434"

# Optional - defaults to llama3
export OLLAMA_MODEL="llama3"

# For Method 3
export OPENAI_BASE_URL="http://localhost:11434"
export OPENAI_API_KEY=""  # Empty string for Ollama
export OPENAI_MODEL="llama3"
```

## Running the Example

Make sure Ollama is running locally if you want to test that example.

```bash
# Build the example
make example EXAMPLE=custom_providers

# Run with OpenRouter
export OPENROUTER_API_KEY="your-api-key"
./bin/custom_providers

# Run with Ollama
export OLLAMA_HOST="http://localhost:11434"
./bin/custom_providers

# Run with both
export OPENROUTER_API_KEY="your-api-key"
export OLLAMA_HOST="http://localhost:11434"
./bin/custom_providers
```

## Other Compatible Providers

This approach can be used with many other providers that offer OpenAI-compatible APIs, including:

- [Together.ai](https://www.together.ai/)
- [Groq](https://groq.com/)
- [Anyscale](https://www.anyscale.com/)
- [Fireworks.ai](https://fireworks.ai/)
- [DeepInfra](https://deepinfra.com/)
- Self-hosted models with [vLLM](https://github.com/vllm-project/vllm)

## Important Notes

1. The compatibility applies primarily to text generation endpoints. Not all features (like function calling) may be available or work the same way across providers.

2. Each provider may have different model parameters and limitations. Check the provider's documentation for details.

3. Some providers may require additional headers or authentication methods. For these cases, you might need to extend the library with a custom provider implementation.

4. When using local models like with Ollama, be aware of the memory and processing power requirements.