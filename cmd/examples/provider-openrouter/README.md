# OpenRouter Provider Example

This example demonstrates how to use the OpenRouter provider to access 400+ models from various providers through a single API.

## Features Demonstrated

1. **Basic Generation** - Using free models for text generation
2. **Streaming** - Real-time streaming responses
3. **Model Discovery** - Fetching available models and their capabilities
4. **Multi-Provider Access** - Using models from different providers (OpenAI, Anthropic, Google, Meta, etc.)

## Prerequisites

1. OpenRouter API key from https://openrouter.ai/
2. Set the environment variable: `export OPENROUTER_API_KEY="your-api-key"`

## Important Notes

- OpenRouter recommends including `HTTP-Referer` and `X-Title` headers for better rate limits and rankings
- Some models may return 405 errors if they're deprecated or the model ID is incorrect
- Check https://openrouter.ai/models for the latest available models

## Running the Example

```bash
# Set your API key
export OPENROUTER_API_KEY="your-api-key"

# Run the example
go run cmd/examples/provider-openrouter/main.go
```

## Key Concepts

### Free Models
OpenRouter provides several free models (marked with `:free` suffix) that don't require credits:
- `huggingface/zephyr-7b-beta:free`
- `nousresearch/nous-capybara-7b:free`
- `mistralai/mistral-7b-instruct:free`

### Model Naming
Models follow the format: `provider/model-name[:variant]`
- `openai/gpt-4` - OpenAI's GPT-4
- `anthropic/claude-3-opus` - Anthropic's Claude 3 Opus
- `google/gemini-pro` - Google's Gemini Pro
- `meta-llama/llama-3-70b-instruct` - Meta's Llama 3 70B

### Special Features
- **Automatic Fallbacks**: OpenRouter can automatically fallback to similar models if one is unavailable
- **Cost Optimization**: Choose models based on performance/cost trade-offs
- **No Regional Restrictions**: Access models regardless of your location
- **BYOK Support**: Bring Your Own Key for underlying providers (5% fee)

## Example Output

```
=== Example 1: Basic Generation with Free Model ===
Response from huggingface/zephyr-7b-beta:free:
1. Performance: Go is designed for high performance with efficient memory management...
2. Concurrency: Built-in goroutines and channels make concurrent programming simple...
3. Simplicity: Clean syntax and standard library make it easy to build robust services...

=== Example 2: Streaming Generation ===
Streaming response: Simple syntax flows,
Goroutines dance in parallel,
Clean code, happy dev.

=== Example 3: Model Discovery ===
Found 400+ models available through OpenRouter

Models by provider:
  openai: 15 models
  anthropic: 8 models
  google: 5 models
  meta-llama: 20 models
  ...

Free models available: 12

=== Example 4: Using Specific Provider Models ===
Testing anthropic/claude-3-haiku:
  Response: The key to happiness is finding meaning and purpose in life...
```

## Cost Considerations

- OpenRouter charges based on the underlying model costs
- Free models have no usage costs
- Prices are typically shown per 1M tokens
- Check https://openrouter.ai/models for current pricing

## Environment Variables

- `OPENROUTER_API_KEY`: Your OpenRouter API key (required)
- `OPENROUTER_API_BASE`: Custom API base URL (optional, defaults to https://openrouter.ai/api/v1)