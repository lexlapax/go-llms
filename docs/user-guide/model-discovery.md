# Model Discovery Guide

This guide covers the model discovery and inventory system in Go-LLMs, which automatically fetches and caches information about available models from various LLM providers.

## Overview

The model discovery system provides:

- **Automatic model discovery** from OpenAI, Anthropic, and Google Gemini
- **Capability detection** for features like multimodal support, function calling, and streaming
- **Intelligent caching** to reduce API calls and improve performance
- **Provider aggregation** to get a unified view across all providers
- **Flexible filtering** to find models matching specific criteria

## Architecture

The model discovery system consists of several components:

```
pkg/util/llmutil/
├── model_inventory.go          # High-level inventory functions
├── model_inventory_test.go     # Comprehensive tests
└── modelinfo/                  # Core model information services
    ├── service.go              # Main orchestration service
    ├── domain/                 # Data structures and interfaces
    │   └── inventory.go        # Model and capability definitions
    ├── fetchers/               # Provider-specific fetchers
    │   ├── openai_fetcher.go   # OpenAI API integration
    │   ├── anthropic_fetcher.go # Anthropic model data
    │   └── google_fetcher.go   # Google Gemini API integration
    └── cache/                  # Caching functionality
        └── file_cache.go       # File-based caching system
```

## Quick Start

### Basic Model Discovery

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
    // Get all available models with default settings
    inventory, err := llmutil.GetAvailableModels(nil)
    if err != nil {
        log.Fatalf("Failed to get models: %v", err)
    }

    fmt.Printf("Found %d models from %d providers\n", 
        len(inventory.Models), len(inventory.Metadata.Providers))

    // List all models
    for _, model := range inventory.Models {
        fmt.Printf("- %s (%s): %s\n", 
            model.Name, model.Provider, model.DisplayName)
    }
}
```

### With Custom Options

```go
// Configure discovery options
opts := &llmutil.GetAvailableModelsOptions{
    UseCache:    true,                    // Enable caching (default: true)
    MaxCacheAge: 6 * time.Hour,          // Cache for 6 hours (default: 24h)
    CachePath:   "/tmp/my-models.json",  // Custom cache location
}

inventory, err := llmutil.GetAvailableModels(opts)
if err != nil {
    log.Fatalf("Failed to get models: %v", err)
}
```

## Model Information Structure

Each model in the inventory contains comprehensive information:

```go
type Model struct {
    Name             string       `json:"name"`              // Model identifier
    Provider         string       `json:"provider"`          // Provider name (openai, anthropic, gemini)
    DisplayName      string       `json:"display_name"`      // Human-readable name
    Description      string       `json:"description"`       // Model description
    ModelFamily      string       `json:"model_family"`      // Model family (gpt, claude, gemini)
    Version          string       `json:"version"`           // Model version
    ContextWindow    int          `json:"context_window"`    // Maximum context window size
    MaxOutputTokens  int          `json:"max_output_tokens"` // Maximum output tokens
    Capabilities     Capabilities `json:"capabilities"`      // Model capabilities
    Pricing          *Pricing     `json:"pricing,omitempty"` // Pricing information (if available)
    CreatedAt        string       `json:"created_at"`        // Creation timestamp
    UpdatedAt        string       `json:"updated_at"`        // Last update timestamp
}

type Capabilities struct {
    Text            MediaTypeCapability `json:"text"`
    Image           MediaTypeCapability `json:"image"`
    Audio           MediaTypeCapability `json:"audio"`
    Video           MediaTypeCapability `json:"video"`
    FunctionCalling bool                `json:"function_calling"`
    Streaming       bool                `json:"streaming"`
    JSONMode        bool                `json:"json_mode"`
}

type MediaTypeCapability struct {
    Read  bool `json:"read"`  // Can process this media type as input
    Write bool `json:"write"` // Can generate this media type as output
}
```

## API Keys and Authentication

The model discovery system requires API keys for accessing provider APIs:

```bash
# Set API keys as environment variables
export OPENAI_API_KEY="your-openai-api-key"
export ANTHROPIC_API_KEY="your-anthropic-api-key"  # Optional - has hardcoded fallback data
export GEMINI_API_KEY="your-gemini-api-key"
```

**Note:** If an API key is missing, the system will:
- For OpenAI and Gemini: Skip that provider and log a warning
- For Anthropic: Fall back to hardcoded model data

## Filtering and Discovery

### Filter by Provider

```go
inventory, _ := llmutil.GetAvailableModels(nil)

var openaiModels []domain.Model
var anthropicModels []domain.Model

for _, model := range inventory.Models {
    switch model.Provider {
    case "openai":
        openaiModels = append(openaiModels, model)
    case "anthropic":
        anthropicModels = append(anthropicModels, model)
    }
}

fmt.Printf("OpenAI has %d models\n", len(openaiModels))
fmt.Printf("Anthropic has %d models\n", len(anthropicModels))
```

### Filter by Capabilities

```go
inventory, _ := llmutil.GetAvailableModels(nil)

// Find models that can read images
var multimodalModels []domain.Model
for _, model := range inventory.Models {
    if model.Capabilities.Image.Read {
        multimodalModels = append(multimodalModels, model)
    }
}

// Find models with function calling
var functionCallingModels []domain.Model
for _, model := range inventory.Models {
    if model.Capabilities.FunctionCalling {
        functionCallingModels = append(functionCallingModels, model)
    }
}

// Find models with streaming support
var streamingModels []domain.Model
for _, model := range inventory.Models {
    if model.Capabilities.Streaming {
        streamingModels = append(streamingModels, model)
    }
}
```

### Filter by Context Window Size

```go
inventory, _ := llmutil.GetAvailableModels(nil)

// Find models with large context windows (>100k tokens)
var largeContextModels []domain.Model
for _, model := range inventory.Models {
    if model.ContextWindow > 100000 {
        largeContextModels = append(largeContextModels, model)
    }
}

fmt.Printf("Models with >100k context: %d\n", len(largeContextModels))
for _, model := range largeContextModels {
    fmt.Printf("- %s: %d tokens\n", model.Name, model.ContextWindow)
}
```

### Filter by Model Family

```go
inventory, _ := llmutil.GetAvailableModels(nil)

// Group models by family
families := make(map[string][]domain.Model)
for _, model := range inventory.Models {
    families[model.ModelFamily] = append(families[model.ModelFamily], model)
}

for family, models := range families {
    fmt.Printf("%s family has %d models\n", family, len(models))
}
```

## Caching System

The model discovery system includes an intelligent caching mechanism to improve performance and reduce API calls.

### Cache Configuration

```go
opts := &llmutil.GetAvailableModelsOptions{
    UseCache:    true,                     // Enable/disable caching
    MaxCacheAge: 12 * time.Hour,          // How long to keep cached data
    CachePath:   "/custom/cache/path.json", // Where to store cache file
}
```

### Cache Behavior

- **Cache Hit**: If valid cached data exists, it's returned immediately
- **Cache Miss**: Data is fetched from providers, then cached for future use
- **Cache Expiry**: Expired cache is automatically refreshed
- **Cache Location**: Defaults to system cache directory (`~/.cache/go-llms/model_inventory.json`)

### Disabling Cache

```go
// Force fresh data from providers
opts := &llmutil.GetAvailableModelsOptions{
    UseCache: false,
}

inventory, err := llmutil.GetAvailableModels(opts)
```

### Cache File Format

The cache file is stored as JSON and includes:

```json
{
  "metadata": {
    "version": "1.0",
    "last_updated": "2025-01-17T10:30:00Z",
    "providers": ["openai", "anthropic", "gemini"],
    "total_models": 42
  },
  "models": [...],
  "fetched_at": "2025-01-17T10:30:00Z"
}
```

## Advanced Usage

### Custom Model Info Service

For advanced use cases, you can work directly with the ModelInfo service:

```go
import (
    "net/http"
    "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"
    "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/fetchers"
)

// Create custom HTTP client
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

// Create service with custom fetchers
service := modelinfo.NewServiceWithCustomFetchers(
    fetchers.NewOpenAIFetcher("https://api.openai.com/v1", httpClient),
    fetchers.NewGoogleFetcher("https://generativelanguage.googleapis.com/v1beta", httpClient),
    &fetchers.AnthropicFetcher{}, // Uses hardcoded data
)

// Fetch models directly
inventory, err := service.FetchAllModels()
if err != nil {
    log.Fatalf("Failed to fetch models: %v", err)
}
```

### Custom Cache Implementation

You can implement custom caching by working with the cache package directly:

```go
import "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/cache"

// Save inventory to custom location
err := cache.SaveInventory(inventory, "/my/custom/cache.json")
if err != nil {
    log.Fatalf("Failed to save cache: %v", err)
}

// Load inventory from custom location
cachedInventory, err := cache.LoadInventory("/my/custom/cache.json")
if err != nil {
    if os.IsNotExist(err) {
        log.Println("Cache file not found, will fetch fresh data")
    } else {
        log.Fatalf("Failed to load cache: %v", err)
    }
}

// Check if cache is still valid
if cache.IsCacheValid(cachedInventory, 24*time.Hour) {
    fmt.Println("Cache is still valid")
} else {
    fmt.Println("Cache has expired")
}
```

## Error Handling

The model discovery system includes comprehensive error handling:

```go
inventory, err := llmutil.GetAvailableModels(nil)
if err != nil {
    // Check error type
    switch {
    case strings.Contains(err.Error(), "API key"):
        log.Println("API key issue - check your environment variables")
    case strings.Contains(err.Error(), "network"):
        log.Println("Network connectivity issue")
    case strings.Contains(err.Error(), "cache"):
        log.Println("Cache-related issue - will fetch fresh data")
    default:
        log.Printf("Unexpected error: %v", err)
    }
    
    // Even if there's an error, partial data might be available
    if inventory != nil && len(inventory.Models) > 0 {
        log.Printf("Got partial data: %d models from cache/hardcoded sources", 
            len(inventory.Models))
    }
}
```

## Performance Considerations

### Concurrent Fetching

The model discovery system fetches data from multiple providers concurrently to minimize latency:

```go
// All providers are queried simultaneously, not sequentially
inventory, err := llmutil.GetAvailableModels(nil) // ~2-3 seconds instead of ~6-9 seconds
```

### Cache Optimization

- Cache files are stored in JSON format for human readability and debugging
- Cache validation is performed at the timestamp level
- Memory usage is optimized with struct field ordering

### API Rate Limiting

Be aware of provider rate limits:

- **OpenAI**: Generally generous limits for the models endpoint
- **Gemini**: Has rate limits that may affect frequent calls
- **Anthropic**: Uses hardcoded data, so no API calls

## CLI Example

The `cmd/examples/modelinfo` provides a complete CLI application:

```bash
# Get all models
go run cmd/examples/modelinfo/main.go

# Filter by provider
go run cmd/examples/modelinfo/main.go -provider=openai

# Filter by capability
go run cmd/examples/modelinfo/main.go -capability=image-input

# Force fresh data
go run cmd/examples/modelinfo/main.go -fresh

# Pretty print with metadata
go run cmd/examples/modelinfo/main.go -pretty -metadata
```

## Best Practices

1. **Use caching in production**: Always enable caching to reduce API calls and improve performance

2. **Handle partial failures gracefully**: Some providers may be unavailable, but others might still work

3. **Set appropriate cache ages**: Balance between fresh data and performance (6-24 hours is usually good)

4. **Check capabilities before using models**: Not all models support all features

5. **Monitor API key usage**: Model discovery does make API calls that count against your quotas

6. **Consider offline fallbacks**: The system includes hardcoded Anthropic data as a fallback

## Troubleshooting

### Common Issues

**"No models found"**
- Check that API keys are set correctly
- Verify network connectivity
- Try disabling cache with `UseCache: false`

**"Context deadline exceeded"**
- Increase HTTP client timeout
- Check network stability
- Some providers may have slower API response times

**"Cache permission errors"**
- Ensure write permissions for cache directory
- Try setting a custom cache path in a writable location

**"API key not valid"**
- Verify API keys are correct and have necessary permissions
- Some providers require specific scopes or permissions for the models endpoint

### Debug Mode

Enable detailed logging by setting environment variables:

```bash
export LOG_LEVEL=debug
go run your-app.go
```

Or use the CLI with verbose output:

```bash
go run cmd/examples/modelinfo/main.go -fresh 2>/dev/null | head -10
```

## Related Documentation

- [Getting Started Guide](getting-started.md) - Basic usage patterns
- [Provider Options Guide](provider-options.md) - Configuring providers
- [Error Handling Guide](error-handling.md) - Error handling patterns
- [API Documentation](../api/README.md) - Complete API reference