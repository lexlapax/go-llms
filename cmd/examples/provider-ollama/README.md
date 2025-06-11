# Ollama Provider Example

This example demonstrates how to use the dedicated Ollama provider in go-llms.

## Features Demonstrated

1. **Model Discovery** - List all locally available Ollama models
2. **Text Generation** - Generate responses using the convenience wrapper
3. **Streaming** - Stream responses token by token
4. **Conversations** - Multi-turn conversations with context
5. **Custom Configuration** - Configure host, timeout, and other options
6. **Provider Comparison** - Shows both convenience wrapper and direct OpenAI provider usage

## Prerequisites

1. Install Ollama from https://ollama.com
2. Pull at least one model:
   ```bash
   ollama pull llama3.2:3b
   ```

## Running the Example

```bash
# Run with default settings (localhost:11434, llama3.2:3b)
go run main.go

# Run with custom host
OLLAMA_HOST=http://192.168.1.100:11434 go run main.go

# Run with different model
OLLAMA_MODEL=mistral:7b go run main.go

# Run with both custom host and model
OLLAMA_HOST=http://custom:11434 OLLAMA_MODEL=codellama:13b go run main.go
```

## Code Overview

### Using the Convenience Wrapper

The Ollama provider offers a dedicated convenience wrapper that configures sensible defaults:

```go
provider := provider.NewOllamaProvider("llama3.2:3b",
    provider.WithOllamaHost("http://localhost:11434"),
    provider.WithOllamaTimeout(2 * time.Minute),
)
```

### Model Discovery

List all available models on your Ollama instance:

```go
fetcher := fetchers.NewOllamaFetcher("http://localhost:11434", nil)
models, err := fetcher.FetchModels()
```

### Streaming Responses

```go
stream, err := provider.Stream(ctx, "Write a story",
    domain.WithTemperature(0.8),
)

for token := range stream {
    fmt.Print(token.Text)
}
```

## Available Ollama Models

Popular models you can use with Ollama:

| Model | Size | Description |
|-------|------|-------------|
| llama3.2:3b | 2.0GB | Llama 3.2 3B - Fast, efficient general purpose |
| llama3.2:1b | 1.3GB | Llama 3.2 1B - Very fast, good for simple tasks |
| mistral:7b | 4.1GB | Mistral 7B - Excellent general purpose |
| codellama:13b | 7.4GB | Code Llama - Specialized for coding |
| gemma2:2b | 1.6GB | Google's Gemma 2 - Efficient and capable |
| phi3:mini | 2.3GB | Microsoft Phi-3 - Strong reasoning |
| qwen2.5:7b | 4.7GB | Qwen 2.5 - Multilingual support |

### Vision Models

| Model | Size | Description |
|-------|------|-------------|
| llava:7b | 4.5GB | LLaVA - Multimodal (text + images) |
| bakllava:7b | 4.5GB | BakLLaVA - Alternative vision model |
| llama3.2-vision:11b | 7.9GB | Llama 3.2 Vision - Latest multimodal |

## Ollama vs OpenAI Provider

The Ollama provider is a convenience wrapper around the OpenAI provider. These are equivalent:

```go
// Using Ollama convenience wrapper
provider1 := provider.NewOllamaProvider("llama3.2:3b")

// Using OpenAI provider directly
provider2 := provider.NewOpenAIProvider(
    "dummy-key",
    "llama3.2:3b",
    domain.NewBaseURLOption("http://localhost:11434"),
    domain.NewHTTPClientOption(&http.Client{Timeout: 120 * time.Second}),
)
```

## Troubleshooting

### Connection Errors

If you get connection errors:
1. Ensure Ollama is running: `ollama serve`
2. Check the host/port: `curl http://localhost:11434/api/tags`
3. Verify firewall settings if using a remote host

### Model Not Found

If you get "model not found" errors:
1. List available models: `ollama list`
2. Pull the required model: `ollama pull <model-name>`

### Timeout Errors

For large models or slow systems:
- Increase timeout: `provider.WithOllamaTimeout(5 * time.Minute)`
- Use smaller models (e.g., `:1b` or `:3b` variants)

## Performance Tips

1. **Model Size**: Smaller models (1B-3B) are much faster
2. **Quantization**: Models with higher quantization (Q8) are more accurate but slower
3. **Context Window**: Longer contexts require more memory and time
4. **Streaming**: Use streaming for better user experience with long responses

## Advanced Usage

### Custom HTTP Client

```go
client := &http.Client{
    Timeout: 5 * time.Minute,
    Transport: &http.Transport{
        MaxIdleConns:        10,
        IdleConnTimeout:     30 * time.Second,
        DisableCompression:  true,
    },
}

provider := provider.NewOllamaProvider("llama3.2:3b",
    domain.NewHTTPClientOption(client),
)
```

### Using with Structured Output

```go
type CodeExample struct {
    Language    string `json:"language"`
    Code        string `json:"code"`
    Explanation string `json:"explanation"`
}

schema := schema.GenerateSchema[CodeExample]()
result, err := provider.GenerateWithSchema(ctx,
    "Show me how to write a hello world in Go",
    schema,
)
```