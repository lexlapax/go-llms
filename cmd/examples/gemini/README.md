# Gemini Example

This example demonstrates how to use the Gemini provider with Go-LLMs to interact with Google's Gemini models. It covers all the key features of the Gemini integration, from basic text generation to advanced configuration options.

## Features

- Text generation with prompt
- Provider-specific options (GeminiGenerationConfigOption and GeminiSafetySettingsOption)
- Message-based conversation
- Structured data generation with schema validation
- Streaming responses
- Generation parameter comparison (temperature, top-k, top-p)
- Error handling

## Prerequisites

- Go 1.21 or later
- A Google API key for Gemini

## Running the Example

1. Set your Gemini API key as an environment variable:

```bash
export GEMINI_API_KEY=your_api_key_here
```

2. Build and run the example:

```bash
make example EXAMPLE=gemini
./bin/gemini
```

## What This Example Demonstrates

### 1. Simple Text Generation

```go
// Create a provider with default settings
provider := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash-lite")

// Generate text from a prompt
response, err := provider.Generate(ctx, "What are the major features of Go?")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Response: %s\n", response)
```

### 2. Provider-Specific Options

#### Generation Configuration

```go
// Create generation config with specific parameters
generationConfig := domain.NewGeminiGenerationConfigOption().
    WithTemperature(0.7).
    WithTopK(40).
    WithMaxOutputTokens(1024).
    WithTopP(0.95)

// Create provider with generation config
provider := provider.NewGeminiProvider(
    apiKey,
    "gemini-2.0-flash-lite",
    generationConfig,
)
```

#### Safety Settings

```go
// Create safety settings to filter harmful content
safetySettings := []map[string]interface{}{
    {
        "category": "HARM_CATEGORY_HARASSMENT",
        "threshold": "BLOCK_MEDIUM_AND_ABOVE",
    },
    {
        "category": "HARM_CATEGORY_HATE_SPEECH",
        "threshold": "BLOCK_MEDIUM_AND_ABOVE",
    },
}
safetySettingsOption := domain.NewGeminiSafetySettingsOption(safetySettings)

// Create provider with safety settings
provider := provider.NewGeminiProvider(
    apiKey,
    "gemini-2.0-flash-lite",
    safetySettingsOption,
)
```

### 3. Message-Based Conversation

```go
// Create a sequence of messages
messages := []domain.Message{
    {Role: domain.RoleUser, Content: "What is machine learning?"},
    {Role: domain.RoleAssistant, Content: "Machine learning is a branch of artificial intelligence..."},
    {Role: domain.RoleUser, Content: "What are some common algorithms?"},
}

// Generate a response to the conversation
response, err := provider.GenerateMessage(ctx, messages)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Response: %s\n", response.Content)
```

### 4. Structured Output with Schema Validation

```go
// Define a struct for the output
type Recipe struct {
    Name        string   `json:"name" description:"The name of the recipe"`
    Ingredients []string `json:"ingredients" description:"List of ingredients"`
    Steps       []string `json:"steps" description:"Step-by-step cooking instructions"`
    PrepTime    int      `json:"prepTime" description:"Preparation time in minutes"`
    CookTime    int      `json:"cookTime" description:"Cooking time in minutes"`
}

// Generate a schema from the struct
schema := schema.GenerateSchema(Recipe{})

// Generate structured output
result, err := provider.GenerateWithSchema(
    ctx,
    "Create a recipe for chocolate chip cookies",
    schema,
)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

// Cast the result to the correct type
recipe := result.(*Recipe)
fmt.Printf("Recipe: %s\n", recipe.Name)
fmt.Println("Ingredients:")
for _, ingredient := range recipe.Ingredients {
    fmt.Printf("- %s\n", ingredient)
}
```

### 5. Streaming Responses

```go
// Stream a response token by token
stream, err := provider.Stream(ctx, "Tell me a story about a dragon.")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Println("Streaming response:")
for token := range stream {
    fmt.Print(token.Text) // Print each token as it arrives
    if token.Finished {
        fmt.Println() // Print newline when complete
    }
}
```

### 6. Generation Parameter Comparison

The example shows how different parameter settings affect generation:

```go
// Low temperature for deterministic output
lowTempConfig := domain.NewGeminiGenerationConfigOption().
    WithTemperature(0.1)

// High temperature for creative output
highTempConfig := domain.NewGeminiGenerationConfigOption().
    WithTemperature(0.9)

// Different top-k values
lowTopKConfig := domain.NewGeminiGenerationConfigOption().
    WithTopK(5)

highTopKConfig := domain.NewGeminiGenerationConfigOption().
    WithTopK(100)

// Compare outputs from different configurations
generateAndCompare(ctx, apiKey, "Explain how batteries work", lowTempConfig, highTempConfig)
generateAndCompare(ctx, apiKey, "Write a short poem about coding", lowTopKConfig, highTopKConfig)
```

### 7. Error Handling

```go
// Handle common error types
response, err := provider.Generate(ctx, "Complex prompt that might be filtered")
if err != nil {
    switch {
    case errors.Is(err, domain.ErrContentFiltered):
        fmt.Println("Content was filtered by safety settings")
    case errors.Is(err, domain.ErrRateLimitExceeded):
        fmt.Println("Rate limit exceeded, try again later")
    case errors.Is(err, domain.ErrContextTooLong):
        fmt.Println("Prompt is too long for model context window")
    default:
        fmt.Printf("Other error: %v\n", err)
    }
    return
}
```

## Gemini Model Capabilities

This example uses the "gemini-2.0-flash-lite" model by default, which is optimized for:
- Fast response times
- Cost-effective generation
- Everyday tasks and queries

The Gemini provider supports multiple models with different capabilities:

| Model | Key Features | Use Cases |
|-------|--------------|-----------|
| gemini-2.0-flash-lite | Fast, efficient, cost-effective | Everyday tasks, simple queries |
| gemini-2.0-flash | Balanced performance and capabilities | General-purpose applications |
| gemini-2.0-pro | Highest capabilities, more detailed | Complex reasoning, creative tasks |
| gemini-1.5-flash | Legacy model, basic capabilities | Basic text generation |
| gemini-1.5-pro | Legacy model, strong capabilities | Detailed content generation |

## Configuration Parameters

### Generation Parameters

The example demonstrates various generation parameters:

- **Temperature**: Controls randomness in token selection
  - Higher values (0.7-1.0): More diverse and creative outputs
  - Lower values (0.1-0.3): More focused and deterministic outputs
  - Default: 0.7

- **Top-K Sampling**: Limits token selection to the K most likely next tokens
  - Higher values (40-100): Broader vocabulary selection
  - Lower values (5-20): More conservative output
  - Default: 40

- **Top-P (Nucleus) Sampling**: Considers only tokens whose cumulative probability exceeds P
  - Higher values (0.9-1.0): More diverse vocabulary
  - Lower values (0.5-0.8): More focused vocabulary
  - Works alongside temperature to balance creativity and coherence
  - Default: 0.95

- **Max Output Tokens**: Maximum length of the generated response
  - Higher values: Longer responses (up to model limits)
  - Lower values: More concise responses
  - Default: 1024

### Safety Settings

The example shows how to configure content filtering using safety settings:

- **Categories**: Different types of potentially harmful content
  - `HARM_CATEGORY_HARASSMENT`: Offensive, bullying, or threatening content
  - `HARM_CATEGORY_HATE_SPEECH`: Content that attacks or discriminates
  - `HARM_CATEGORY_SEXUALLY_EXPLICIT`: Explicit or adult content
  - `HARM_CATEGORY_DANGEROUS_CONTENT`: Harmful instructions or illegal content

- **Thresholds**: Filtering levels for each category (from most to least restrictive)
  - `BLOCK_LOW_AND_ABOVE`: Blocks almost all potentially harmful content
  - `BLOCK_MEDIUM_AND_ABOVE`: Blocks moderate and high severity content
  - `BLOCK_ONLY_HIGH`: Blocks only high severity content
  - `BLOCK_NONE`: No content filtering

## Implementation Notes

- The example handles automatic role mapping between Go-LLMs standard roles and Gemini API roles
- Streaming uses Server-Sent Events (SSE) format with automatic handling
- Structured output enhances prompts with schema information
- Error handling maps Gemini-specific errors to standard domain errors
- API key is stored securely in environment variables

## Advanced Usage

For more advanced usage of the Gemini provider, including combining multiple provider options, implementing retry logic, and integrating with multi-provider setups, see the [Provider Options Guide](/docs/user-guide/provider-options.md) and [Multi-Provider Example](/cmd/examples/multi/README.md).