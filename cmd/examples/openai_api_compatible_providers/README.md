# OpenAI API Compatible Providers Example

This example demonstrates how to use go-llms with LLM providers that implement the OpenAI API specification. It shows two common use cases:

1. **OpenRouter** - A service that provides access to various LLM providers through an OpenAI-compatible API
2. **Ollama** - A tool for running LLMs locally, also with an OpenAI-compatible API

## Overview

The go-llms library makes it easy to use providers that implement the same API as supported providers (like OpenAI) by allowing you to override the base URL. This is particularly useful for:

- Using API gateways or proxies in front of supported providers
- Working with self-hosted or local models
- Integrating with providers that offer OpenAI-compatible APIs

## Three Ways to Configure OpenAI API Compatible Providers

### Method 1: Direct Provider Instantiation

Use the interface-based provider option system to create a provider with a custom base URL:

```go
baseURLOption := domain.NewBaseURLOption("https://custom-endpoint.com")
provider := provider.NewOpenAIProvider(
    apiKey,
    modelName,
    baseURLOption,
)
```

You can add additional options as needed:

```go
// Custom HTTP client with timeout
httpClientOption := domain.NewHTTPClientOption(&http.Client{
    Timeout: 30 * time.Second,
})

// Custom headers for the API requests
headersOption := domain.NewHeadersOption(map[string]string{
    "X-Custom-Header": "value",
})

provider := provider.NewOpenAIProvider(
    apiKey,
    modelName,
    baseURLOption,
    httpClientOption,
    headersOption,
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

### API Endpoint Details

OpenRouter's API requires some specific configuration:

- Base URL should be set to `https://openrouter.ai/api` (without `/v1` as our OpenAI provider adds it)
- OpenRouter works best with the `GenerateMessage` method rather than `Generate`
- Required HTTP headers:
  - `HTTP-Referer`: Your site URL for attribution
  - `X-Title`: Your site name for attribution

### Configuration

Set the following environment variables:

```bash
# Required
export OPENROUTER_API_KEY="your-openrouter-api-key"

# Optional - defaults to mistralai/mistral-small-3.1-24b-instruct:free
export OPENROUTER_MODEL="mistralai/mistral-small-3.1-24b-instruct:free"

# For Method 3
export OPENAI_BASE_URL="https://openrouter.ai/api"
export OPENAI_API_KEY="your-openrouter-api-key"
export OPENAI_MODEL="mistralai/mistral-small-3.1-24b-instruct:free"
```

## Ollama Example

[Ollama](https://ollama.ai/) lets you run LLMs locally. It also provides an OpenAI-compatible API that integrates well with go-llms.

### API Endpoint Details

Ollama's API has some specific requirements:

- Base URL should be set to your Ollama instance (default: `http://localhost:11434`)
- Although Ollama doesn't need an API key, the OpenAI provider requires a non-empty key
- You must provide a dummy API key (`dummy-key`) which will be ignored by Ollama
- Ollama works with both `Generate` and `GenerateMessage` methods

### Configuration

Set the following environment variables:

```bash
# Optional - defaults to http://localhost:11434
export OLLAMA_HOST="http://localhost:11434"

# Optional - defaults to llama3.2:3b
export OLLAMA_MODEL="llama3.2:3b"

# For Method 3
export OPENAI_BASE_URL="http://localhost:11434"
export OPENAI_API_KEY="dummy-key"  # Dummy key for Ollama (required but ignored)
export OPENAI_MODEL="llama3.2:3b"
```

## Running the Example

Make sure Ollama is running locally if you want to test that example.

```bash
# Build the example
make example EXAMPLE=openai_api_compatible_providers

# Run with OpenRouter
export OPENROUTER_API_KEY="your-api-key"
./bin/openai_api_compatible_providers

# Run with Ollama
export OLLAMA_HOST="http://localhost:11434"
./bin/openai_api_compatible_providers

# Run with both
export OPENROUTER_API_KEY="your-api-key"
export OLLAMA_HOST="http://localhost:11434"
./bin/openai_api_compatible_providers
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

1. For OpenAI-compatible providers, pay close attention to the API endpoint paths. The go-llms OpenAI provider automatically adds `/v1` and the appropriate endpoint path:
   - `/v1/completions` for the `Generate` method
   - `/v1/chat/completions` for the `GenerateMessage` method

2. Each provider may require different endpoint paths and authentication methods:
   - **OpenRouter**: Use `https://openrouter.ai/api` (without `/v1`) and attribution headers
   - **Ollama**: Use base URL `http://localhost:11434` and a dummy API key
   - **Others**: Check their documentation for specific requirements

3. The compatibility applies primarily to text generation endpoints. Not all features (like function calling) may be available or work the same way across providers.

4. Each provider may have different model parameters and limitations. Check the provider's documentation for details.

5. When using local models like with Ollama, be aware of the memory and processing power requirements.