# Working with Providers

This guide covers everything you need to know about using LLM providers in go-llms.

## Overview

Providers are your gateway to Large Language Models. Go-llms supports multiple providers through a unified interface, making it easy to switch between different LLMs or use multiple providers in the same application.

## Supported Providers

### OpenAI
- **Models**: GPT-4o, GPT-4o-mini, GPT-4 Turbo, GPT-3.5 Turbo
- **Features**: Function calling, JSON mode, streaming, vision
- **Best for**: General-purpose AI, code generation, complex reasoning

### Anthropic
- **Models**: Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Haiku
- **Features**: Large context windows, streaming, vision
- **Best for**: Long documents, analysis, creative writing

### Google Gemini
- **Models**: Gemini 2.0 Flash Lite, Gemini Pro, Gemini Pro Vision
- **Features**: Multimodal by default, streaming
- **Best for**: Multimodal tasks, fast inference

### Ollama
- **Models**: Llama 3.2, Mistral, Phi-3, CodeLlama, and many more
- **Features**: Local hosting, GPU acceleration, model management, streaming
- **Best for**: Privacy, offline use, custom models, cost-effective inference
- **Special**: No API key required, runs locally

### OpenRouter
- **Models**: 400+ models from OpenAI, Anthropic, Google, Meta, and more
- **Features**: Automatic fallbacks, cost optimization, unified API
- **Best for**: Accessing multiple providers through one API, cost-effective model selection
- **Special**: Pay-as-you-go pricing, 68 free models available

### Google Vertex AI
- **Models**: Gemini 1.5 Pro, Gemini 1.5 Flash, Claude (partner models), PaLM 2
- **Features**: Enterprise-grade, regional deployment, IAM integration, streaming
- **Best for**: Enterprise applications, GCP integration, data residency requirements
- **Authentication**: Service account or Application Default Credentials (ADC)
- **Special**: Access to partner models (Claude) in enterprise environment

### OpenAI-Compatible
- **Providers**: LM Studio, vLLM, OpenRouter (also has dedicated wrapper), any OpenAI-compatible API
- **Features**: Local models, custom endpoints
- **Best for**: Privacy, custom models, offline use
- **Important**: Only provide the base URL without `/v1` or `/v1/chat/completions` - the provider adds these automatically

## Basic Usage

### Creating a Provider

```go
import "github.com/lexlapax/go-llms/pkg/llm/provider"

// OpenAI
openai := provider.NewOpenAIProvider(
    "your-api-key",  // or os.Getenv("OPENAI_API_KEY")
    "gpt-4o",
)

// Anthropic
anthropic := provider.NewAnthropicProvider(
    "your-api-key",  // or os.Getenv("ANTHROPIC_API_KEY")
    "claude-3-5-sonnet-latest",
)

// Google Gemini
gemini := provider.NewGeminiProvider(
    "your-api-key",  // or os.Getenv("GOOGLE_API_KEY")
    "gemini-2.0-flash-lite",
)

// Ollama (convenience wrapper)
ollama := provider.NewOllamaProvider("llama3.2:3b")

// Or with custom host
ollama := provider.NewOllamaProvider(
    "llama3.2:3b",
    domain.WithBaseURL("http://my-ollama-server:11434"),
)

// Google Vertex AI
vertexAI, err := provider.NewVertexAIProvider(
    "my-project-id",
    "us-central1",
    "gemini-1.5-flash",
)
if err != nil {
    log.Fatal(err)
}

// OpenAI-Compatible (for other providers)
// IMPORTANT: Only provide base URL without /v1 or /v1/chat/completions
// The provider automatically appends /v1/chat/completions
compatible := provider.NewOpenAIProvider(
    "",  // No API key needed for local
    "model-name",
    domain.WithBaseURL("http://localhost:8080"),  // NOT http://localhost:8080/v1
)

// Example with OpenRouter (OpenAI-compatible)
openrouter := provider.NewOpenRouterProvider(
    "your-api-key",
    "openai/gpt-4o",
    // OpenRouter uses https://openrouter.ai/api as base URL
    // The provider will automatically append /v1/chat/completions
)
```

### Making Requests

All providers implement the same interface:

```go
// Simple text generation
response, err := provider.Generate(
    context.Background(),
    "Explain quantum computing in simple terms",
)

// Conversation with messages
messages := []domain.Message{
    {Role: domain.RoleSystem, Content: "You are a helpful assistant"},
    {Role: domain.RoleUser, Content: "What's the weather like?"},
}
response, err := provider.GenerateMessage(context.Background(), messages)

// Structured output
type Answer struct {
    Explanation string `json:"explanation"`
    Confidence  float64 `json:"confidence"`
}
var answer Answer
err := provider.GenerateWithSchema(
    context.Background(),
    "Is water wet?",
    &answer,
)
```

## Provider Options

### Temperature and Creativity

```go
import "github.com/lexlapax/go-llms/pkg/llm/domain"

// More creative responses
response, err := provider.Generate(
    ctx,
    "Write a poem about coding",
    domain.WithTemperature(0.8),
)

// More deterministic responses
response, err := provider.Generate(
    ctx,
    "List the steps to install Go",
    domain.WithTemperature(0.1),
)
```

### Token Limits

```go
// Limit response length
response, err := provider.Generate(
    ctx,
    "Summarize War and Peace",
    domain.WithMaxTokens(200),
)

// Get more detailed responses
response, err := provider.Generate(
    ctx,
    "Explain machine learning in detail",
    domain.WithMaxTokens(2000),
)
```

### System Prompts

```go
messages := []domain.Message{
    {
        Role: domain.RoleSystem,
        Content: "You are an expert Go programmer. Provide concise, idiomatic Go code.",
    },
    {
        Role: domain.RoleUser,
        Content: "How do I read a JSON file?",
    },
}

response, err := provider.GenerateMessage(ctx, messages)
```

### Streaming Responses

```go
// Stream tokens as they arrive
stream, err := provider.StreamMessage(ctx, messages)
if err != nil {
    return err
}

for chunk := range stream {
    fmt.Print(chunk)
}
```

## Multi-Provider Setup

Use multiple providers for reliability and cost optimization:

### Primary with Fallback

```go
// Use OpenAI primarily, fall back to Anthropic
multi := provider.NewMultiProvider(
    []provider.ProviderConfig{
        {
            Provider: openai,
            Name:     "openai",
            Weight:   1.0,
        },
        {
            Provider: anthropic,
            Name:     "anthropic",
            Weight:   1.0,
        },
    },
    provider.StrategyPrimary,  // Use first, fallback on error
)
```

### Load Balancing

```go
// Distribute requests across providers
multi := provider.NewMultiProvider(
    []provider.ProviderConfig{
        {
            Provider: openai,
            Name:     "openai",
            Weight:   0.7,  // 70% of requests
        },
        {
            Provider: anthropic,
            Name:     "anthropic", 
            Weight:   0.3,  // 30% of requests
        },
    },
    provider.StrategyRandom,  // Random selection by weight
)
```

### Consensus

```go
// Get responses from multiple providers and compare
multi := provider.NewMultiProvider(
    []provider.ProviderConfig{
        {Provider: openai, Name: "openai"},
        {Provider: anthropic, Name: "anthropic"},
        {Provider: gemini, Name: "gemini"},
    },
    provider.StrategyConsensus,  // Get responses from all
)

// Returns most common response
response, err := multi.Generate(ctx, "What is 2+2?")
```

## Advanced Features

### Multimodal Content

```go
// Send images to vision-capable models
imageData, _ := os.ReadFile("chart.png")
msg := domain.NewImageMessage(
    domain.RoleUser,
    imageData,
    "image/png",
    "What does this chart show?",
)

// Works with OpenAI GPT-4V, Anthropic Claude 3, Gemini Pro Vision
response, err := provider.GenerateMessage(ctx, []domain.Message{msg})
```

### Function Calling (OpenAI)

```go
// Define available functions
tools := []domain.Tool{
    {
        Type: "function",
        Function: domain.Function{
            Name:        "get_weather",
            Description: "Get weather for a location",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "location": map[string]interface{}{
                        "type": "string",
                        "description": "City name",
                    },
                },
                "required": []string{"location"},
            },
        },
    },
}

response, err := openai.GenerateMessage(
    ctx,
    messages,
    domain.WithTools(tools),
)
```

### JSON Mode (OpenAI)

```go
// Force JSON output
response, err := openai.Generate(
    ctx,
    "List 3 programming languages with their strengths",
    domain.WithResponseFormat("json_object"),
)
```

## Provider-Specific Features

### OpenAI
- **JSON Mode**: Force valid JSON output
- **Function Calling**: Let the model call functions
- **Vision**: Analyze images with GPT-4V
- **Fine-tuning**: Use custom fine-tuned models

### Anthropic
- **Large Context**: Up to 200k tokens context window
- **System Prompts**: Strong instruction following
- **Vision**: Native multimodal support in Claude 3

### Google Gemini
- **Multimodal Native**: Images, text, code in same request
- **Fast Inference**: Optimized for speed
- **Safety Settings**: Built-in content filtering

### Ollama
- **Local Models**: Run models on your own hardware
- **Model Management**: Easy download and management via CLI
- **GPU Acceleration**: Automatic GPU detection and usage
- **No API Key**: No authentication required for local use
- **Model Discovery**: List available models programmatically

### Google Vertex AI
- **Enterprise Authentication**: Service account or Application Default Credentials (ADC)
- **Regional Deployment**: Deploy models in specific regions for data residency
- **Partner Models**: Access to Claude models in enterprise environment
- **IAM Integration**: Fine-grained access control with Google Cloud IAM
- **Model Versions**: Access to specific model versions for consistency

## Working with OpenAI-Compatible Providers

### Base URL Configuration

When using OpenAI-compatible providers, it's crucial to understand how base URLs are handled:

1. **The OpenAI provider automatically appends `/v1/chat/completions` to your base URL**
2. **You should only provide the base URL without any API endpoints**

#### Correct Examples:
```go
// LM Studio
provider := provider.NewOpenAIProvider(
    "",
    "local-model",
    domain.WithBaseURL("http://localhost:1234"),  // ✓ Correct
)

// vLLM
provider := provider.NewOpenAIProvider(
    "",
    "model-name", 
    domain.WithBaseURL("http://localhost:8000"),  // ✓ Correct
)

// OpenRouter (has convenience wrapper)
provider := provider.NewOpenRouterProvider(
    "api-key",
    "openai/gpt-4o",
    // Automatically uses https://openrouter.ai/api
)

// Custom OpenAI-compatible API
provider := provider.NewOpenAIProvider(
    "api-key",
    "model-name",
    domain.WithBaseURL("https://api.example.com"),  // ✓ Correct
)
```

#### Incorrect Examples:
```go
// DON'T include /v1 or /v1/chat/completions
provider := provider.NewOpenAIProvider(
    "",
    "model", 
    domain.WithBaseURL("http://localhost:1234/v1"),  // ✗ Wrong
)

provider := provider.NewOpenAIProvider(
    "",
    "model",
    domain.WithBaseURL("http://localhost:8000/v1/chat/completions"),  // ✗ Wrong  
)
```

### Common OpenAI-Compatible Providers

1. **LM Studio**
   - Default URL: `http://localhost:1234`
   - No API key required
   - Supports various open-source models

2. **vLLM**
   - Default URL: `http://localhost:8000`
   - High-performance inference server
   - Supports distributed serving

3. **OpenRouter**
   - URL: `https://openrouter.ai/api` (handled by wrapper)
   - Requires API key
   - Access to 400+ models from various providers

4. **Ollama** (Special Case)
   - Has its own wrapper: `NewOllamaProvider`
   - Default URL: `http://localhost:11434`
   - Different API structure, but wrapper handles it

## Environment Configuration

### API Keys

```bash
# Set API keys as environment variables
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GOOGLE_API_KEY="..."

# Ollama configuration (optional)
export OLLAMA_HOST="http://localhost:11434"  # Default
export OLLAMA_MODEL="llama3.2:3b"           # Default model

# OpenRouter configuration
export OPENROUTER_API_KEY="sk-or-..."

# Vertex AI configuration
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
export VERTEX_AI_PROJECT_ID="my-project-id"
export VERTEX_AI_LOCATION="us-central1"
export VERTEX_AI_MODEL="gemini-1.5-flash"

# Use in code
provider := provider.NewOpenAIProvider(
    os.Getenv("OPENAI_API_KEY"),
    "gpt-4o",
)

# Ollama doesn't need an API key
ollama := provider.NewOllamaProvider(
    os.Getenv("OLLAMA_MODEL"),  // or "llama3.2:3b"
)
```

### Model Selection

```go
// Use environment variables for models too
model := os.Getenv("OPENAI_MODEL")
if model == "" {
    model = "gpt-4o-mini"  // Default
}

provider := provider.NewOpenAIProvider(apiKey, model)
```

## Error Handling

### Rate Limits

```go
for retries := 0; retries < 3; retries++ {
    response, err := provider.Generate(ctx, prompt)
    if err == nil {
        return response, nil
    }
    
    // Check for rate limit errors
    if strings.Contains(err.Error(), "rate limit") {
        time.Sleep(time.Second * time.Duration(retries+1))
        continue
    }
    
    return "", err
}
```

### Context Length

```go
// Handle context too long errors
response, err := provider.Generate(ctx, veryLongPrompt)
if err != nil && strings.Contains(err.Error(), "context length") {
    // Truncate and retry
    truncated := truncatePrompt(veryLongPrompt)
    response, err = provider.Generate(ctx, truncated)
}
```

### Network Issues

```go
// Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := provider.Generate(ctx, prompt)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    }
}
```

## Performance Tips

### 1. Choose the Right Model
- Use smaller models (gpt-4o-mini, claude-3-haiku) for simple tasks
- Reserve large models for complex reasoning
- Consider local models for high-volume, low-complexity tasks

### 2. Optimize Prompts
- Be concise but clear
- Use system prompts to set behavior once
- Provide examples for consistent output

### 3. Use Caching
```go
// Cache provider instances
var (
    providerCache map[string]domain.Provider
    cacheMu       sync.RWMutex
)

func GetProvider(name string) domain.Provider {
    cacheMu.RLock()
    defer cacheMu.RUnlock()
    return providerCache[name]
}
```

### 4. Batch Requests
```go
// Process multiple prompts efficiently
prompts := []string{"prompt1", "prompt2", "prompt3"}
responses := make([]string, len(prompts))

var wg sync.WaitGroup
for i, prompt := range prompts {
    wg.Add(1)
    go func(idx int, p string) {
        defer wg.Done()
        responses[idx], _ = provider.Generate(ctx, p)
    }(i, prompt)
}
wg.Wait()
```

## Cost Optimization

### Model Selection Strategy
```go
// Use cheaper models when possible
func SelectModel(complexity string) string {
    switch complexity {
    case "simple":
        return "gpt-4o-mini"
    case "moderate":
        return "gpt-4o"
    case "complex":
        return "gpt-4-turbo"
    default:
        return "gpt-4o-mini"
    }
}
```

### Token Monitoring
```go
// Track token usage
type TokenTracker struct {
    mu    sync.Mutex
    usage map[string]int
}

func (t *TokenTracker) Track(provider string, tokens int) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.usage[provider] += tokens
}
```

## Working with Ollama

### Model Discovery

```go
import "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"

// List available Ollama models
service := modelinfo.NewService()
models, err := service.GetOllamaModels()
if err != nil {
    log.Fatal(err)
}

for _, model := range models {
    fmt.Printf("Model: %s, Context: %d tokens\n", model.Name, model.ContextWindow)
}
```

### Running Multiple Models

```go
// Use different models for different tasks
codingModel := provider.NewOllamaProvider("codellama:7b")
chatModel := provider.NewOllamaProvider("llama3.2:3b")
analysisModel := provider.NewOllamaProvider("mistral:7b")

// Use appropriate model for each task
code, _ := codingModel.Generate(ctx, "Write a Go function to sort a slice")
chat, _ := chatModel.Generate(ctx, "Explain the code to a beginner")
analysis, _ := analysisModel.Generate(ctx, "Review the code for best practices")
```

## Working with Google Vertex AI

Vertex AI is Google Cloud's enterprise AI platform, offering access to Google's models and partner models like Claude in a secure, enterprise environment.

### Authentication

Vertex AI supports multiple authentication methods:

#### 1. Application Default Credentials (ADC)

```go
// ADC will be used automatically if configured
provider, err := provider.NewVertexAIProvider(
    "my-project-id",
    "us-central1",
    "gemini-1.5-flash",
)
```

Set up ADC:
```bash
# Using gcloud CLI
gcloud auth application-default login

# Or set project
gcloud config set project my-project-id
```

#### 2. Service Account

```go
// Using service account JSON file
provider, err := provider.NewVertexAIProvider(
    "my-project-id",
    "us-central1", 
    "gemini-1.5-pro",
    domain.NewVertexAIServiceAccountOption("/path/to/service-account.json"),
)

// Or via environment variable
os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/service-account.json")
provider, err := provider.NewVertexAIProvider(
    "my-project-id",
    "us-central1",
    "gemini-1.5-pro",
)
```

### Available Models

#### Google Models
- `gemini-2.0-flash-preview`: Latest Gemini 2.0 Flash
- `gemini-1.5-pro`: Most capable Gemini model
- `gemini-1.5-flash`: Fast and cost-effective
- `gemini-1.0-pro`: Legacy stable version

#### Partner Models (Claude via Vertex)
- `claude-3-opus@20240229`: Most capable Claude model
- `claude-3-7-sonnet@20241022`: Balanced performance
- `claude-3-5-sonnet@20241022`: Latest Sonnet version
- `claude-3-5-haiku@20241022`: Fast and efficient

### Regional Deployment

```go
// Deploy in specific regions for data residency
usProvider, _ := provider.NewVertexAIProvider(
    "my-project", "us-central1", "gemini-1.5-flash",
)

euProvider, _ := provider.NewVertexAIProvider(
    "my-project", "europe-west4", "gemini-1.5-flash",  
)

asiaProvider, _ := provider.NewVertexAIProvider(
    "my-project", "asia-northeast1", "gemini-1.5-flash",
)
```

### Using Partner Models

```go
// Access Claude through Vertex AI
claude, err := provider.NewVertexAIProvider(
    "my-project-id",
    "us-central1",
    "claude-3-5-sonnet@20241022",
)

response, err := claude.Generate(ctx, "Explain quantum computing")
```

### Multimodal with Gemini

```go
// Analyze images with Gemini models
imageData, _ := os.ReadFile("diagram.png")
imageMsg := domain.NewImageMessage(domain.RoleUser, "Explain this diagram", imageData)

response, err := provider.Generate(ctx, imageMsg)
```

### Environment Configuration

```bash
# Required
export VERTEX_AI_PROJECT_ID="my-project-id"
export VERTEX_AI_LOCATION="us-central1"  # Optional, defaults to us-central1

# For service account auth
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# Model selection
export VERTEX_AI_MODEL="gemini-1.5-flash"
```

### Cost Considerations

- Vertex AI pricing varies by region
- Partner models (Claude) may have different pricing
- Consider using Gemini Flash for cost-effective inference
- Monitor usage through Google Cloud Console

### IAM Permissions

Required permissions for service accounts:
- `aiplatform.endpoints.predict`
- `aiplatform.models.predict`

Example IAM policy:
```bash
gcloud projects add-iam-policy-binding PROJECT_ID \
    --member="serviceAccount:YOUR_SERVICE_ACCOUNT@PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/aiplatform.user"
```

## Testing with Mock Providers

```go
import "github.com/lexlapax/go-llms/pkg/llm/provider"

// Create deterministic mock for testing
mock := provider.NewMockProvider()
mock.SetResponse("What is 2+2?", "4")

// Use in tests
response, err := mock.Generate(ctx, "What is 2+2?")
assert.Equal(t, "4", response)
```

## Best Practices

1. **Always handle errors** - LLM calls can fail for many reasons
2. **Use appropriate timeouts** - Set context timeouts for all calls
3. **Monitor costs** - Track token usage across providers
4. **Cache when possible** - Reuse responses for identical prompts
5. **Choose models wisely** - Match model size to task complexity
6. **Secure API keys** - Never hardcode keys in source code
7. **Plan for failures** - Implement retry logic and fallbacks

## Next Steps

Now that you understand providers:
- Learn about [Structured Output](structured-output.md) for reliable data extraction
- Explore [Agents](agents.md) to build autonomous systems
- Discover [Tools](tools.md) to extend agent capabilities
- Master [Workflows](workflows.md) for complex multi-step processes

Ready to build something amazing? Choose your provider and let's go! 🚀