# Architecture Analysis: Adding "List Models" Capability to Go-LLMs

## Current Provider API Support Research

We've analyzed the three primary providers to determine if they offer model listing capabilities:

### 1. OpenAI
- **API Support**: Yes, robust support via `GET /models` endpoint
- **Documentation**: [OpenAI Models API](https://platform.openai.com/docs/api-reference/models)
- **Response Format**: Returns detailed list with model IDs, owners, permissions
- **Implementation**: Simple REST endpoint call with standard authentication
- **Example Response**:
  ```json
  {
    "object": "list",
    "data": [
      {
        "id": "gpt-4o",
        "object": "model",
        "created": 1683758102,
        "owned_by": "openai",
        "permission": [
          {
            "id": "...",
            "object": "permission",
            "created": 1683758102,
            "allow_create_engine": false,
            "allow_sampling": true,
            "allow_logprobs": true,
            "allow_search_indices": false,
            "allow_view": true,
            "allow_fine_tuning": false,
            "organization": "*",
            "group": null,
            "is_blocking": false
          }
        ],
        "root": "gpt-4o",
        "parent": null
      },
      /* More models... */
    ]
  }
  ```

### 2. Anthropic
- **API Support**: Limited - no public endpoint specifically for listing models
- **Documentation**: Only documents available models in their API documentation
- **Available Models**: Fixed set (claude-3-opus, claude-3-sonnet, claude-3-haiku, etc.)
- **Implementation Challenge**: Would require hard-coding available models or scraping docs
- **Reference**: [Anthropic API Reference](https://docs.anthropic.com/claude/reference/selecting-a-model)
- **Limitation**: The model list would need to be manually maintained as Anthropic releases new models

### 3. Google Gemini
- **API Support**: Yes, via the Google AI API
- **Documentation**: [Gemini API Models](https://ai.google.dev/api/rest)
- **Available Models**: Limited set of Gemini models
- **Implementation**: Similar to OpenAI with standard REST endpoint
- **Example Models**: gemini-1.5-pro, gemini-1.5-flash, etc.

## Architecture Decision Analysis

There are several approaches we could take to implement model listing capability:

### Option 1: Add to Core Provider Interface
```go
type Provider interface {
    // Existing methods
    Generate(...)
    GenerateMessage(...)
    
    // New method
    ListModels() ([]ModelInfo, error)
}
```

**Pros:**
- Standardized interface across providers
- Easy access from existing provider instances
- Consistent developer experience

**Cons:**
- Forces implementation on all providers including those without API support
- Would complicate custom provider implementations
- Inconsistent with separation of concerns (generation vs. discovery)
- Hard to represent provider-specific model attributes
- Increases interface surface area for minimal gain
- Makes mock implementations more complex for testing

### Option 2: Dedicated Model Registry Package

```go
// pkg/llm/modelregistry/registry.go
package modelregistry

type ModelInfo struct {
    ID           string
    Provider     string
    Capabilities map[string]bool // e.g., "chat", "image", "embedding" 
    MaxTokens    int
    // Additional metadata
}

// Provider-specific implementations
func ListOpenAIModels(options ...ProviderOption) ([]ModelInfo, error)
func ListAnthropicModels(options ...ProviderOption) ([]ModelInfo, error)
func ListGeminiModels(options ...ProviderOption) ([]ModelInfo, error)

// Generic registry
type Registry interface {
    ListModels(provider string) ([]ModelInfo, error)
    GetModelDetails(provider, modelID string) (*ModelInfo, error)
    RefreshCache() error
}
```

**Pros:**
- Clear separation of concerns
- Only implemented for providers with native support
- Can include caching for better performance
- Maintains clean core Provider interface
- Can be extended with more functionality (recommendations, filtering)
- Allows for provider-specific model attributes
- Better testability with clear boundaries

**Cons:**
- Additional package to maintain
- Some code duplication in authentication/client handling
- Requires separate configuration from provider instances

### Option 3: Command-Line Tool Only

**Pros:**
- No changes to library code
- Perfect for developer exploration use case
- Simpler to implement quickly
- Doesn't increase complexity of the core library

**Cons:**
- Not available programmatically
- Limited usefulness for applications
- Still requires underlying code to make API calls, so doesn't actually save much work

## Recommendation

We recommend implementing **Option 2: Dedicated Model Registry Package** with a command-line interface built on top of it.

This approach:
1. Respects separation of concerns
2. Works with inconsistent provider APIs
3. Still makes functionality available programmatically
4. Provides a developer-friendly CLI
5. Allows for future extension to more providers and capabilities

## Implementation Plan

1. **Create Model Registry Package**
   - Define common `ModelInfo` struct
   - Implement provider-specific listing functions
   - Create caching layer for performance
   - Add registry interface for extensibility

2. **Implement Provider-Specific Listings**
   - OpenAI: Direct API call to `/models`
   - Anthropic: Static list with version checking capability
   - Gemini: API call to Google AI endpoints
   - OpenAI-compatible: Parameterized handling of compatible providers

3. **Add Command-Line Interface**
   - Command: `llm models list [--provider=...]`
   - Pretty-print formatted output 
   - Filtering options (capabilities, context size, etc.)
   - JSON output option for programmatic consumption

4. **Documentation**
   - API documentation for programmatic usage
   - CLI documentation with examples
   - Integration examples in README

5. **Testing**
   - Unit tests with mocked HTTP responses
   - Integration tests with actual API calls (optional)
   - CLI testing with test fixtures

## Sample Code Structure

```
pkg/
  llm/
    modelregistry/
      registry.go         # Core interfaces and types
      openai.go           # OpenAI-specific implementation
      anthropic.go        # Anthropic-specific implementation
      gemini.go           # Gemini-specific implementation
      cache.go            # Caching implementation
      filters.go          # Filtering utilities
  
cmd/
  llm/
    commands/
      models.go           # CLI command implementation
```

## Future Extensions

This architecture could be extended to support:

1. Model capability discovery (what features each model supports)
2. Model selection recommendations based on requirements
3. Cost estimation based on token usage
4. Automated fallback selection
5. Regular updates of model information for providers without API endpoints

By implementing this as a separate package, we maintain the clean core library design while providing powerful model discovery capabilities for developers.