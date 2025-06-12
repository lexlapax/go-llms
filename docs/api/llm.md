# LLM API Reference

The llm package (`pkg/llm`) provides interfaces and implementations for interacting with Language Model providers. It offers a unified API across different providers while supporting provider-specific features.

## Overview

The LLM package provides:
- Unified provider interface for text generation
- Support for multiple providers (OpenAI, Anthropic, Google Gemini)
- Multimodal content support (text, images, video, audio)
- Streaming responses
- Structured output generation
- Multi-provider strategies for reliability and performance

## Core Interfaces

### Provider Interface

The foundation for all LLM providers.

```go
type Provider interface {
    // Generate produces text from a prompt
    Generate(ctx context.Context, prompt string, options ...Option) (string, error)
    
    // GenerateMessage produces a response from a sequence of messages
    GenerateMessage(ctx context.Context, messages []Message, options ...Option) (*Message, error)
    
    // GenerateWithSchema produces structured output conforming to a schema
    GenerateWithSchema(ctx context.Context, prompt string, schema interface{}, options ...Option) (interface{}, error)
    
    // Stream streams responses token by token
    Stream(ctx context.Context, prompt string, options ...Option) (<-chan string, error)
    
    // StreamMessage streams a response from a sequence of messages
    StreamMessage(ctx context.Context, messages []Message, options ...Option) (<-chan string, error)
}
```

### Messages

Messages represent conversations with support for multimodal content.

```go
// Role constants
const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleTool      Role = "tool"
)

// Message structure
type Message struct {
    Role       Role          `json:"role"`
    Content    string        `json:"content"`
    ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
    ToolCallID string        `json:"tool_call_id,omitempty"`
    MultiContent []ContentPart `json:"multi_content,omitempty"`
}

// ContentPart for multimodal content
type ContentPart struct {
    Type     ContentType    `json:"type"`
    Text     string         `json:"text,omitempty"`
    ImageURL *ImageURL      `json:"image_url,omitempty"`
    VideoURL *VideoURL      `json:"video_url,omitempty"`
    AudioURL *AudioURL      `json:"audio_url,omitempty"`
}
```

### Creating Messages

```go
// Simple text message
msg := domain.NewTextMessage(domain.RoleUser, "Hello, how are you?")

// Message with image
imageData, _ := os.ReadFile("image.jpg")
msg := domain.NewImageMessage(
    domain.RoleUser,
    imageData,
    "image/jpeg",
    "What's in this image?",
)

// Message with image URL
msg := domain.NewImageURLMessage(
    domain.RoleUser,
    "https://example.com/image.jpg",
    "Describe this image",
)

// Message with video
videoData, _ := os.ReadFile("video.mp4")
msg := domain.NewVideoMessage(
    domain.RoleUser,
    videoData,
    "video/mp4",
    "What happens in this video?",
)
```

## Generation Options

### Request-Level Options

Configure individual requests:

```go
// Temperature controls randomness (0.0-2.0)
response, err := provider.Generate(ctx, prompt,
    domain.WithTemperature(0.7),
)

// Max tokens limits response length
response, err := provider.Generate(ctx, prompt,
    domain.WithMaxTokens(1000),
)

// Stop sequences end generation
response, err := provider.Generate(ctx, prompt,
    domain.WithStopSequences("END", "STOP"),
)

// Top-p for nucleus sampling
response, err := provider.Generate(ctx, prompt,
    domain.WithTopP(0.9),
)

// Frequency penalty reduces repetition
response, err := provider.Generate(ctx, prompt,
    domain.WithFrequencyPenalty(0.5),
)

// Presence penalty encourages new topics
response, err := provider.Generate(ctx, prompt,
    domain.WithPresencePenalty(0.5),
)

// Combine multiple options
response, err := provider.Generate(ctx, prompt,
    domain.WithTemperature(0.7),
    domain.WithMaxTokens(500),
    domain.WithTopP(0.9),
)
```

### Provider-Level Options

Configure providers at creation time:

```go
// Common options across providers
// IMPORTANT: For OpenAI-compatible providers, only provide the base URL
// The provider automatically appends /v1/chat/completions
provider := provider.NewOpenAIProvider(apiKey, model,
    domain.NewBaseURLOption("https://custom-endpoint.com"), // NOT .../v1
    domain.NewHTTPClientOption(customHTTPClient),
    domain.NewTimeoutOption(30000), // 30 seconds
    domain.NewHeadersOption(map[string]string{
        "X-Custom-Header": "value",
    }),
)

// Provider-specific options
openaiProvider := provider.NewOpenAIProvider(apiKey, model,
    domain.NewOpenAIOrganizationOption("org-123"),
)

anthropicProvider := provider.NewAnthropicProvider(apiKey, model,
    domain.NewAnthropicSystemPromptOption("You are helpful."),
    domain.NewAnthropicMetadataOption(map[string]string{
        "user_id": "123",
    }),
)

geminiProvider := provider.NewGeminiProvider(apiKey, model,
    domain.NewGeminiGenerationConfigOption().
        WithTemperature(0.7).
        WithTopK(40),
    domain.NewGeminiSafetySettingsOption(safetySettings),
)
```

For detailed provider options documentation, see the [Provider Options Guide](../user-guide/provider-options.md).

## Provider Implementations

### OpenAI

Supports GPT-3.5, GPT-4, and GPT-4o models.

```go
import "github.com/lexlapax/go-llms/pkg/llm/provider"

// Create provider
openai := provider.NewOpenAIProvider("api-key", "gpt-4o")

// Basic generation
response, _ := openai.Generate(ctx, "Explain quantum computing")

// With messages
messages := []domain.Message{
    domain.NewTextMessage(domain.RoleSystem, "You are a physics teacher"),
    domain.NewTextMessage(domain.RoleUser, "What is quantum entanglement?"),
}
response, _ := openai.GenerateMessage(ctx, messages)

// Streaming
stream, _ := openai.Stream(ctx, "Tell me a story")
for chunk := range stream {
    fmt.Print(chunk)
}
```

### OpenAI-Compatible Providers

Many providers implement the OpenAI API specification. Use the OpenAI provider with a custom base URL:

```go
// IMPORTANT: Only provide the base URL without /v1 or /v1/chat/completions
// The provider automatically appends these paths

// LM Studio (local)
lmstudio := provider.NewOpenAIProvider("", "local-model",
    domain.NewBaseURLOption("http://localhost:1234"), // NOT .../v1
)

// vLLM (self-hosted)
vllm := provider.NewOpenAIProvider("", "model-name",
    domain.NewBaseURLOption("http://localhost:8000"), // NOT .../v1
)

// OpenRouter (requires API key)
openrouter := provider.NewOpenRouterProvider("api-key", "openai/gpt-4o",
    // Uses https://openrouter.ai/api automatically
    domain.NewHeadersOption(map[string]string{
        "HTTP-Referer": "https://your-app.com",
        "X-Title": "Your App",
    }),
)

// Custom OpenAI-compatible API
custom := provider.NewOpenAIProvider("api-key", "model",
    domain.NewBaseURLOption("https://api.example.com"), // NOT .../v1
)
```

### Anthropic

Supports Claude 3 models (Opus, Sonnet, Haiku).

```go
// Create provider
anthropic := provider.NewAnthropicProvider("api-key", "claude-3-5-sonnet-latest")

// With system prompt
anthropic = provider.NewAnthropicProvider("api-key", "claude-3-5-sonnet-latest",
    domain.NewAnthropicSystemPromptOption("You are a helpful coding assistant"),
)

// Generate response
response, _ := anthropic.Generate(ctx, "Write a Go function to reverse a string")
```

### Google Gemini

Supports Gemini 1.5 and 2.0 models.

```go
// Create provider
gemini := provider.NewGeminiProvider("api-key", "gemini-2.0-flash-lite")

// With generation config
gemini = provider.NewGeminiProvider("api-key", "gemini-2.0-flash-lite",
    domain.NewGeminiGenerationConfigOption().
        WithTemperature(0.7).
        WithMaxOutputTokens(2048),
)

// Multimodal generation
imageData, _ := os.ReadFile("diagram.png")
messages := []domain.Message{
    domain.NewImageMessage(domain.RoleUser, imageData, "image/png", 
        "Explain this diagram"),
}
response, _ := gemini.GenerateMessage(ctx, messages)
```

## Multi-Provider

Use multiple providers for reliability and performance.

```go
// Create individual providers
openai := provider.NewOpenAIProvider(openaiKey, "gpt-4o")
anthropic := provider.NewAnthropicProvider(anthropicKey, "claude-3-5-sonnet-latest")
gemini := provider.NewGeminiProvider(geminiKey, "gemini-2.0-flash-lite")

// Create provider configuration
providers := []provider.ProviderConfig{
    {Provider: openai, Weight: 1.0, Name: "openai"},
    {Provider: anthropic, Weight: 1.0, Name: "anthropic"},
    {Provider: gemini, Weight: 1.0, Name: "gemini"},
}
```

### Strategies

#### Fastest Response

```go
// Returns first successful response
multi := provider.NewMultiProvider(providers, provider.StrategyFastest)
response, _ := multi.Generate(ctx, "Quick question: What is 2+2?")
```

#### Primary with Fallback

```go
// Uses primary, falls back on failure
multi := provider.NewMultiProvider(providers, provider.StrategyPrimary).
    WithPrimaryProvider("openai")
response, _ := multi.Generate(ctx, prompt)
```

#### Consensus

```go
// Seeks agreement among providers
multi := provider.NewMultiProvider(providers, provider.StrategyConsensus).
    WithConsensusStrategy(provider.ConsensusSimilarity).
    WithSimilarityThreshold(0.8)
response, _ := multi.Generate(ctx, "What is the capital of France?")
```

## Structured Output

Generate data conforming to schemas.

```go
// Define structure
type Recipe struct {
    Name        string   `json:"name"`
    Ingredients []string `json:"ingredients"`
    Steps       []string `json:"steps"`
    PrepTime    int      `json:"prep_time_minutes"`
}

// Generate with schema
var recipe Recipe
err := provider.GenerateWithSchema(ctx, 
    "Create a recipe for chocolate chip cookies",
    &recipe,
)

// The provider will ensure the output matches the schema
fmt.Printf("Recipe: %s\n", recipe.Name)
fmt.Printf("Prep time: %d minutes\n", recipe.PrepTime)
```

For detailed structured output documentation, see [Structured API Reference](structured.md).

## Error Handling

```go
response, err := provider.Generate(ctx, prompt)
if err != nil {
    // Check for specific error types
    if errors.Is(err, domain.ErrRateLimitExceeded) {
        // Handle rate limiting
        time.Sleep(time.Minute)
    } else if errors.Is(err, domain.ErrContextLengthExceeded) {
        // Reduce prompt size
        prompt = truncatePrompt(prompt)
    } else if errors.Is(err, domain.ErrContentFiltered) {
        // Handle content filtering
        fmt.Println("Content was filtered")
    } else {
        // Generic error handling
        return fmt.Errorf("generation failed: %w", err)
    }
}
```

## Streaming

Handle real-time token streams.

```go
// Basic streaming
stream, err := provider.Stream(ctx, "Write a haiku about programming")
if err != nil {
    return err
}

var fullResponse strings.Builder
for token := range stream {
    fmt.Print(token) // Print as it arrives
    fullResponse.WriteString(token)
}

// Message-based streaming
messages := []domain.Message{
    domain.NewTextMessage(domain.RoleUser, "Tell me a story"),
}
stream, err := provider.StreamMessage(ctx, messages)

// With timeout handling
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

stream, err := provider.Stream(ctx, prompt)
for {
    select {
    case token, ok := <-stream:
        if !ok {
            return // Stream complete
        }
        fmt.Print(token)
    case <-ctx.Done():
        return ctx.Err() // Timeout
    }
}
```

## Examples

### Chat Application

```go
// Initialize conversation
messages := []domain.Message{
    domain.NewTextMessage(domain.RoleSystem, 
        "You are a helpful assistant. Be concise and friendly."),
}

// Chat loop
scanner := bufio.NewScanner(os.Stdin)
for {
    fmt.Print("You: ")
    if !scanner.Scan() {
        break
    }
    
    // Add user message
    userMsg := domain.NewTextMessage(domain.RoleUser, scanner.Text())
    messages = append(messages, userMsg)
    
    // Generate response
    response, err := provider.GenerateMessage(ctx, messages)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        continue
    }
    
    // Add assistant response
    messages = append(messages, *response)
    fmt.Printf("Assistant: %s\n", response.Content)
}
```

### Multi-Modal Analysis

```go
// Analyze image and text together
imageData, _ := os.ReadFile("chart.png")
messages := []domain.Message{
    domain.NewTextMessage(domain.RoleUser, 
        "Please analyze this sales chart and provide insights."),
    domain.NewImageMessage(domain.RoleUser, imageData, "image/png", 
        "What trends do you see?"),
}

response, _ := provider.GenerateMessage(ctx, messages)
fmt.Println("Analysis:", response.Content)
```

### Reliability with Multi-Provider

```go
// Create resilient provider setup
providers := []provider.ProviderConfig{
    {Provider: primary, Weight: 1.0, Name: "primary"},
    {Provider: backup1, Weight: 0.8, Name: "backup1"},
    {Provider: backup2, Weight: 0.6, Name: "backup2"},
}

multi := provider.NewMultiProvider(providers, provider.StrategyPrimary).
    WithPrimaryProvider("primary").
    WithTimeout(10 * time.Second)

// Will automatically failover if primary fails
response, err := multi.Generate(ctx, prompt)
```

## Best Practices

1. **Provider Selection**: Choose providers based on your needs (cost, quality, speed)
2. **Error Handling**: Always handle provider-specific errors appropriately
3. **Rate Limiting**: Implement retry logic with exponential backoff
4. **Context Management**: Use contexts for timeout and cancellation
5. **Streaming**: Use streaming for better user experience with long responses
6. **Multi-Provider**: Use for mission-critical applications requiring high availability

## Integration with Other Components

- **Agents**: Providers power agent LLM capabilities (see [Agent API](agent.md))
- **Structured Output**: Use with schema validation (see [Structured API](structured.md))
- **Utilities**: Leverage provider utilities (see [Utils API](utils.md#llm-utilities))
- **Testing**: Use mock providers for testing (see [Test Utilities](testutils.md))

## See Also

- [Provider Options Guide](../user-guide/provider-options.md) - Detailed provider configuration
- [Multi-Provider Guide](../user-guide/multi-provider.md) - Advanced multi-provider patterns
- [Error Handling Guide](../user-guide/error-handling.md) - Comprehensive error handling
- [Performance Guide](../technical/performance.md) - Optimization techniques