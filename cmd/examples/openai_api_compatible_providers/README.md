# OpenAI API Compatible Providers Example

This example demonstrates how to use go-llms with LLM providers that implement the OpenAI API specification. It shows integration with two popular OpenAI API-compatible services:

1. **OpenRouter** - A unified API gateway service that provides access to 100+ models from various providers (Anthropic, Mistral, OpenAI, etc.)
2. **Ollama** - A tool for running large language models locally on your machine with minimal setup

## Overview

Many LLM providers and frameworks have adopted the OpenAI API specification as a standard interface for model interactions. The go-llms library makes it easy to integrate with these providers by configuring the OpenAI provider with a different base URL and any required headers.

This approach offers several benefits:

- **Expanded Model Access**: Use specialized models not directly supported in go-llms
- **Local Deployment**: Run models locally for privacy, lower latency, or offline use
- **Provider Redundancy**: Switch between providers without changing your code
- **Cost Optimization**: Choose the most cost-effective provider for your needs
- **Specialized Features**: Access proprietary improvements from different providers

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

## OpenRouter Integration

[OpenRouter](https://openrouter.ai/) is a unified API gateway that provides access to over 100 models from different providers including Anthropic, Mistral, OpenAI, Meta (Llama models), and others. It offers a standardized interface with transparent pricing and usage limits.

### Key Features

- **Model Variety**: Access to models from multiple providers through a single API
- **Fallback Routing**: Automatic routing to alternative models if primary choices are unavailable
- **Pay-as-you-go**: Simple credit-based pricing with no minimum spend
- **Model Comparison**: Easy comparison of different models' capabilities and costs
- **Free Tier**: Provides free access to selected models for testing and light usage

### API Endpoint Configuration

OpenRouter's API requires specific configuration when used with go-llms:

- **Base URL**: Use `https://openrouter.ai/api` (without `/v1` because the OpenAI provider adds it automatically)
- **API Method**: OpenRouter works best with the `GenerateMessage` method (chat completions endpoint)
- **Required Attribution Headers**:
  - `HTTP-Referer`: Your site URL for attribution (required)
  - `X-Title`: Your site or app name for attribution (recommended)

### Code Example

```go
// Create required headers for OpenRouter
headersOption := domain.NewHeadersOption(map[string]string{
    "HTTP-Referer": "https://your-app-url.com",  // Required attribution
    "X-Title":      "Your App Name",             // Recommended attribution
})

// Set correct base URL (OpenAI provider will add /v1 automatically)
baseURLOption := domain.NewBaseURLOption("https://openrouter.ai/api")

// Create OpenRouter provider
provider := provider.NewOpenAIProvider(
    openRouterAPIKey,
    "mistralai/mistral-small-3.1-24b-instruct:free", // Default free model
    baseURLOption,
    headersOption,
)

// Use with message-based API (recommended)
messages := []domain.Message{
    {Role: domain.RoleSystem, Content: "You are a helpful assistant."},
    {Role: domain.RoleUser, Content: "Tell me about the solar system."},
}

response, err := provider.GenerateMessage(context.Background(), messages)
```

### Environment Variables

```bash
# Required
export OPENROUTER_API_KEY="your-openrouter-api-key"  # Get from openrouter.ai

# Optional - defaults to mistralai/mistral-small-3.1-24b-instruct:free
export OPENROUTER_MODEL="anthropic/claude-3-opus-20240229"

# For environment variables method
export OPENAI_BASE_URL="https://openrouter.ai/api"
export OPENAI_API_KEY="your-openrouter-api-key"
export OPENAI_MODEL="mistralai/mistral-small-3.1-24b-instruct:free"
```

### Popular Models Available on OpenRouter

| Provider | Model | Key Features |
|----------|-------|--------------|
| Anthropic | claude-3-opus-20240229 | High-end reasoning, accuracy, and long context |
| Anthropic | claude-3-sonnet-20240229 | Balanced performance and efficiency |
| Mistral AI | mistral-small-3.1-24b-instruct | High quality, cost-effective model |
| Meta | meta-llama/llama-3-70b-instruct | Open-source, powerful, large context window |
| Cohere | cohere/command-r-plus | Excellent for research and knowledge tasks |
| Aleph Alpha | luminous-supreme | Strong multilingual and reasoning capabilities |

## Ollama Integration

[Ollama](https://ollama.ai/) provides an easy way to run large language models locally on your computer. It handles model downloading, quantization, and serving through an OpenAI-compatible API, making it perfect for local development, privacy-sensitive applications, or offline use.

### Key Features

- **Local Execution**: Run models entirely on your own hardware
- **No API Keys or Costs**: Use models without external API calls or payments
- **Easy Setup**: Simple installation with automatic model management
- **Privacy**: Keep sensitive data local and process offline
- **Customization**: Fine-tune models and create custom model definitions
- **Broad Model Support**: Access to Llama 3, Mistral, Codellama, Gemma, and many others

### API Endpoint Configuration

Ollama's API has specific requirements when integrating with go-llms:

- **Base URL**: Set to your Ollama instance (default: `http://localhost:11434`)
- **API Key**: Ollama doesn't use API keys, but the OpenAI provider requires a non-empty value
  - Use a dummy key like `"dummy-key"` which will be ignored by Ollama
- **API Methods**: Ollama supports both `Generate` (completions) and `GenerateMessage` (chat completions)
- **Timeout**: Consider setting longer timeouts for local models, especially on slower hardware

### Code Example

```go
// Create a longer timeout for local model inference
httpClient := &http.Client{
    Timeout: 60 * time.Second, // Local models may take longer to respond
}
httpClientOption := domain.NewHTTPClientOption(httpClient)

// Set Ollama endpoint URL
baseURLOption := domain.NewBaseURLOption("http://localhost:11434")

// Create Ollama provider (with dummy API key)
provider := provider.NewOpenAIProvider(
    "dummy-key", // Required by provider but ignored by Ollama
    "llama3.2:3b", // Model must be pulled in Ollama before use
    baseURLOption,
    httpClientOption,
)

// Use the provider for both text completion and chat
response, err := provider.Generate(
    context.Background(),
    "Explain quantum computing in simple terms",
    domain.WithTemperature(0.7),
    domain.WithMaxTokens(1000),
)
```

### Environment Variables

```bash
# Optional - defaults to http://localhost:11434
export OLLAMA_HOST="http://localhost:11434"

# Optional - defaults to llama3.2:3b
export OLLAMA_MODEL="llama3.2:3b"

# For environment variables method
export OPENAI_BASE_URL="http://localhost:11434"
export OPENAI_API_KEY="dummy-key"  # Dummy key for Ollama (required but ignored)
export OPENAI_MODEL="llama3.2:3b"
```

### Popular Models Available in Ollama

| Model | Size | Key Features | Command to Pull |
|-------|------|--------------|----------------|
| llama3.2 | 3B, 8B, 70B | Latest Llama model, general purpose | `ollama pull llama3.2:3b` |
| mistral | 7B, 12B | Excellent performance in the 7B-12B range | `ollama pull mistral` |
| codellama | 7B, 13B, 34B | Specialized for code generation | `ollama pull codellama` |
| gemma | 2B, 7B | Google's compact efficient model | `ollama pull gemma:7b` |
| phi | 3B | Microsoft's small but capable model | `ollama pull phi` |
| neural-chat | 7B | Optimized for chat interactions | `ollama pull neural-chat` |

### Optimizing Ollama Performance

- Ensure your machine meets [minimum requirements](https://github.com/ollama/ollama/blob/main/README.md#requirements)
- For better performance, use machines with dedicated GPUs
- Start with smaller models (3B-7B) on machines with limited resources
- Increase `--gpu-layers` parameter for better GPU utilization
- Configure quantization level based on your needs (q4_0 for speed, q8_0 for quality)

## Groq Integration

[Groq](https://groq.com/) is a specialized AI inference service with ultra-fast response times. It offers an OpenAI-compatible API with some of the fastest inference speeds available for LLMs, making it ideal for latency-sensitive applications.

### Key Features

- **Extreme Speed**: Remarkably fast inference times (often 10-100x faster than other services)
- **Low Latency**: Consistently low response times for interactive applications
- **High Throughput**: Process more requests per second than traditional providers
- **Premium Models**: Access to Llama 3, Mistral, and other high-quality models
- **Flexible Pricing**: Usage-based model with competitive rates

### API Endpoint Configuration

Groq's API can be configured with go-llms as follows:

- **Base URL**: Use `https://api.groq.com`
- **API Key**: Requires a Groq API key from your Groq account
- **API Methods**: Supports both `Generate` (completions) and `GenerateMessage` (chat completions)
- **Model Names**: Model identifiers follow the format `llama3-8b-8192`, `mistral-7b-instruct`, etc.

### Code Example

```go
// Set Groq base URL
baseURLOption := domain.NewBaseURLOption("https://api.groq.com")

// Create Groq provider
provider := provider.NewOpenAIProvider(
    groqAPIKey,
    "llama3-8b-8192", // Llama 3 8B model
    baseURLOption,
)

// Use with message-based API
messages := []domain.Message{
    {Role: domain.RoleSystem, Content: "You are a helpful assistant."},
    {Role: domain.RoleUser, Content: "Generate a Python function to calculate prime numbers."},
}

response, err := provider.GenerateMessage(context.Background(), messages)
```

### Environment Variables

```bash
# Required
export GROQ_API_KEY="your-groq-api-key"  # Get from groq.com

# Optional - defaults to llama3-8b-8192
export GROQ_MODEL="mixtral-8x7b-32768"

# For environment variables method
export OPENAI_BASE_URL="https://api.groq.com"
export OPENAI_API_KEY="your-groq-api-key"
export OPENAI_MODEL="llama3-8b-8192"
```

### Popular Models Available on Groq

| Model | Context Size | Key Features |
|-------|-------------|--------------|
| llama3-8b-8192 | 8K | Fast, efficient general-purpose model |
| llama3-70b-8192 | 8K | Highest capability Llama 3 model |
| mixtral-8x7b-32768 | 32K | Mixture of experts model with large context |
| gemma-7b-it | 8K | Google's efficient instruction-tuned model |
| mistral-7b-instruct | 8K | High-quality mid-size model |

## Running the Example

This example demonstrates integration with both OpenRouter (cloud service) and Ollama (local service). You can run it with either or both providers.

### Prerequisites

1. **For OpenRouter**:
   - Sign up at [openrouter.ai](https://openrouter.ai) to get an API key
   - No additional installation required

2. **For Ollama**:
   - Install Ollama from [ollama.ai](https://ollama.ai)
   - Start the Ollama service
   - Pull a model (e.g., `ollama pull llama3.2:3b`)

3. **For Groq**:
   - Sign up at [groq.com](https://groq.com) to get an API key
   - No additional installation required

### Running the Application

```bash
# Build the example
make example EXAMPLE=openai_api_compatible_providers

# Run with OpenRouter
export OPENROUTER_API_KEY="your-api-key"
./bin/openai_api_compatible_providers

# Run with Ollama
export OLLAMA_HOST="http://localhost:11434"
./bin/openai_api_compatible_providers

# Run with Groq
export GROQ_API_KEY="your-api-key"
export GROQ_MODEL="llama3-8b-8192"
./bin/openai_api_compatible_providers

# Run with multiple providers simultaneously
export OPENROUTER_API_KEY="your-api-key"
export OLLAMA_HOST="http://localhost:11434"
export GROQ_API_KEY="your-api-key"
./bin/openai_api_compatible_providers
```

## Additional Compatible Providers

The approach demonstrated in this example can be extended to many other providers that implement the OpenAI API specification:

### Cloud Providers

| Provider | Base URL | Key Features | API Documentation |
|----------|----------|--------------|-------------------|
| [Together.ai](https://www.together.ai/) | `https://api.together.xyz` | Wide model selection, API-compatible | [Docs](https://docs.together.ai/reference/inference) |
| [Anyscale](https://www.anyscale.com/) | `https://api.endpoints.anyscale.com/v1` | Ray-powered inference, optimized serving | [Docs](https://docs.endpoints.anyscale.com/) |
| [Fireworks.ai](https://fireworks.ai/) | `https://api.fireworks.ai/inference` | Fast inference, specialized models | [Docs](https://readme.fireworks.ai/) |
| [DeepInfra](https://deepinfra.com/) | `https://api.deepinfra.com/v1/openai` | Inference infrastructure, model serving | [Docs](https://deepinfra.com/docs) |
| [Perplexity](https://www.perplexity.ai/) | `https://api.perplexity.ai` | Knowledge-focused models, research | [Docs](https://docs.perplexity.ai/) |
| [Lepton](https://www.lepton.ai/) | `https://api.lepton.ai/v1` | Inference optimization, specialized models | [Docs](https://www.lepton.ai/docs) |

### Self-Hosted Options

| Framework | Description | Key Features |
|-----------|-------------|--------------|
| [vLLM](https://github.com/vllm-project/vllm) | High-throughput LLM serving | Continuous batching, PagedAttention |
| [Text Generation Inference](https://github.com/huggingface/text-generation-inference) | HuggingFace's serving framework | Flash Attention, Quantization |
| [LocalAI](https://github.com/go-skynet/LocalAI) | OpenAI API compatible self-hosted inference | Supports many model formats |
| [LM Studio](https://lmstudio.ai/) | Desktop application with API server | User-friendly UI, one-click serving |

## Compatibility Considerations

### API Path Handling

The go-llms OpenAI provider handles API path construction in a specific way:

1. **Base Paths**: The OpenAI provider automatically adds `/v1` to the base URL
   - For OpenRouter: Use `https://openrouter.ai/api` (provider adds `/v1`)
   - For Groq: Use `https://api.groq.com` (provider adds `/v1`)
   - For Ollama: Use `http://localhost:11434` (no `/v1` is needed)

2. **Endpoint Selection**: The provider automatically selects the appropriate endpoint
   - `Generate()` → Uses `/v1/completions` endpoint
   - `GenerateMessage()` → Uses `/v1/chat/completions` endpoint

### Provider-Specific Requirements

Each provider may have unique requirements:

| Provider | API Key | Headers | Special Considerations |
|----------|---------|---------|------------------------|
| OpenRouter | Required | Attribution headers required | Model names include provider prefix |
| Ollama | Dummy value | None | Local service must be running |
| Groq | Required | None | Ultra-fast response times |
| Together.ai | Required | Optional user identification | Model versioning in model names |
| Fireworks.ai | Required | None | Custom error handling |

## Troubleshooting Guide

### Common Issues and Solutions

1. **Connection Errors**
   - **Problem**: "Failed to connect to..."
   - **Solution**: Check that the service is running and the base URL is correct
   - **For Ollama**: Ensure Ollama service is started (`ollama serve`)

2. **Authentication Errors**
   - **Problem**: "Invalid API key" or "Authentication failed"
   - **Solution**: Verify your API key is correct and has the necessary permissions
   - **For OpenRouter**: Ensure your account has credits available

3. **Model Not Found Errors**
   - **Problem**: "Model not found" or "Unknown model"
   - **Solution**: Verify the model name is correct for the specific provider
   - **For Ollama**: Pull the model first with `ollama pull model_name`

4. **Timeout Errors**
   - **Problem**: "Context deadline exceeded" or "Request timed out"
   - **Solution**: Increase the timeout in the HTTP client option
   - **For Local Models**: Local inference can be much slower than cloud APIs

5. **Path Resolution Issues**
   - **Problem**: "Not found" or "Invalid endpoint"
   - **Solution**: Check if the provider requires a specific base URL format
   - **For API Compatibility**: Some providers have specific path requirements

### Performance Optimization

1. **For Ollama (Local Models)**:
   - Use smaller models for faster responses (3B-7B models recommended)
   - Run on hardware with a GPU when possible
   - Use appropriate quantization for your hardware (q4_0, q5_K_M, etc.)
   - Set longer timeouts for complex queries
   - Consider batch processing for non-interactive workloads

2. **For Cloud Providers**:
   - Select appropriate model size for your task (larger isn't always better)
   - Implement client-side caching for common queries
   - Use streaming for interactive applications
   - Consider regional endpoints for lower latency if available

## Additional Resources

- [Go-LLMs Provider Options Guide](/docs/user-guide/provider-options.md) - Detailed documentation on provider options
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference) - The API specification many providers implement
- [Ollama Model Library](https://ollama.ai/library) - Browse available models for Ollama
- [OpenRouter Models](https://openrouter.ai/models) - View all models available through OpenRouter