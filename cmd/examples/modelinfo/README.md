# Model Info Example

This example demonstrates how to use the ModelInfo service to fetch and display information about models available from various LLM providers.

## Features

- Fetches model information from OpenAI, Anthropic, and Google
- Caches results to reduce redundant API calls
- Displays model details in structured JSON format 
- Supports filtering by provider, capability, and model name

## Usage

```bash
# Fetch and display all models
go run main.go

# Fetch and display models from a specific provider
go run main.go --provider openai

# Fetch and display models with specific capabilities
go run main.go --capability image-input

# Available capability filters:
# - text-input
# - text-output  
# - image-input
# - image-output
# - audio-input
# - audio-output
# - video-input
# - video-output
# - function-calling
# - streaming
# - json-mode

# Filter by model name pattern
go run main.go --name "gpt-4"

# Force fresh data (ignore cache)
go run main.go --fresh

# Specify custom cache file location
go run main.go --cache-path ./my-cache.json
```

## Environment Variables

The example requires API keys for providers you want to fetch data from:

- `OPENAI_API_KEY` - OpenAI API key
- `ANTHROPIC_API_KEY` - Anthropic API key (optional - will use hardcoded data)
- `GEMINI_API_KEY` - Google/Gemini API key

Without these keys, the example will fall back to hardcoded model information where available.