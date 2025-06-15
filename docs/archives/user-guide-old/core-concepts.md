# Core Concepts

Understanding these core concepts will help you get the most out of go-llms.

## Overview

go-llms is built around five key concepts:

1. **Providers** - Connect to LLMs (OpenAI, Anthropic, etc.)
2. **Messages** - Structure conversations with LLMs
3. **Schemas** - Define and validate structured data
4. **Agents** - Create autonomous entities that use tools
5. **State** - Manage data flow through your application

## Providers

Providers are your connection to LLMs. Each provider implements the same interface, making it easy to switch between them.

### The Provider Interface

```go
type Provider interface {
    Generate(ctx context.Context, prompt string, opts ...Option) (string, error)
    GenerateMessage(ctx context.Context, messages []Message, opts ...Option) (*Message, error)
    GenerateWithSchema(ctx context.Context, prompt string, schema interface{}, opts ...Option) (interface{}, error)
    Stream(ctx context.Context, prompt string, opts ...Option) (<-chan string, error)
    StreamMessage(ctx context.Context, messages []Message, opts ...Option) (<-chan string, error)
}
```

### Key Points

- **Unified Interface**: All providers implement the same methods
- **Context Support**: Use context for timeouts and cancellation
- **Options**: Configure generation with temperature, max tokens, etc.
- **Streaming**: Real-time token streaming for better UX

### Example

```go
// Any provider works the same way
provider := provider.NewOpenAIProvider(key, "gpt-4o")
// or
provider := provider.NewAnthropicProvider(key, "claude-3-5-sonnet-latest")
// or
provider := provider.NewGeminiProvider(key, "gemini-2.0-flash-lite")

// Same interface for all
response, err := provider.Generate(ctx, "Hello!")
```

## Messages

Messages represent the conversation with an LLM. They have roles and can contain multiple types of content.

### Message Structure

```go
type Message struct {
    Role       Role          // system, user, assistant, or tool
    Content    string        // text content
    ToolCalls  []ToolCall    // tool invocations
    MultiContent []ContentPart // images, videos, etc.
}
```

### Roles

- **System**: Instructions for the LLM's behavior
- **User**: Input from the user
- **Assistant**: Responses from the LLM
- **Tool**: Results from tool execution

### Example Conversation

```go
messages := []Message{
    {Role: RoleSystem, Content: "You are a helpful assistant"},
    {Role: RoleUser, Content: "What's the weather like?"},
    {Role: RoleAssistant, Content: "I'll need to check that for you. What's your location?"},
    {Role: RoleUser, Content: "New York City"},
}

response, err := provider.GenerateMessage(ctx, messages)
```

### Multimodal Content

Messages can include images, videos, and other content:

```go
// Image message
imageData, _ := os.ReadFile("chart.png")
msg := domain.NewImageMessage(
    domain.RoleUser,
    imageData,
    "image/png",
    "What does this chart show?",
)
```

## Schemas

Schemas define the structure of data you want from LLMs. They're based on JSON Schema and enable reliable structured output.

### Basic Schema

```go
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]*domain.Schema{
        "name": {
            Type:        "string",
            Description: "Product name",
        },
        "price": {
            Type:        "number",
            Minimum:     float64Ptr(0),
            Description: "Price in USD",
        },
        "inStock": {
            Type:        "boolean",
            Description: "Availability",
        },
    },
    Required: []string{"name", "price"},
}
```

### Using Schemas

```go
// Method 1: Direct generation
type Product struct {
    Name    string  `json:"name"`
    Price   float64 `json:"price"`
    InStock bool    `json:"inStock"`
}

var product Product
err := provider.GenerateWithSchema(ctx, "Tell me about iPhone 15", &product)

// Method 2: Manual validation
response, _ := provider.Generate(ctx, enhancedPrompt)
processor := processor.NewJsonProcessor()
result, err := processor.Process(schema, response)
```

### Validation Features

- **Type checking**: Ensure correct data types
- **Constraints**: Min/max values, string patterns, etc.
- **Required fields**: Enforce mandatory data
- **Custom validators**: Add your own validation logic

## Agents

Agents are autonomous entities that can use tools to accomplish tasks. They combine LLMs with executable functions.

### Agent Components

1. **LLM Provider**: The "brain" that makes decisions
2. **Tools**: Functions the agent can call
3. **State**: Working memory for the agent
4. **Hooks**: Monitor and control agent behavior

### Simple Agent

```go
// Create an agent
agent := core.NewLLMAgent("assistant", provider)

// Give it tools
agent.AddTool(searchTool)
agent.AddTool(calculatorTool)

// Set its personality
agent.SetSystemPrompt("You are a helpful research assistant")

// Let it work
state := domain.NewState().Set("task", "Find the population of Tokyo and calculate its density")
result, err := agent.Run(ctx, state)
```

### How Agents Work

1. Agent receives a task via state
2. LLM decides what tools to use
3. Tools are executed with parameters
4. Results go back to the LLM
5. Process repeats until task is complete

## State

State is how data flows through your application. It's a flexible key-value store that agents and workflows use to share information.

### Working with State

```go
// Create state
state := domain.NewState()

// Set values
state.Set("user_id", 123)
state.Set("preferences", map[string]interface{}{
    "language": "en",
    "theme": "dark",
})

// Get values
userID, exists := state.Get("user_id")
if !exists {
    // Handle missing value
}

// Clone for isolation
newState := state.Clone()

// Merge states
state.Merge(otherState)
```

### State in Agents

```go
// Input state
inputState := domain.NewState()
inputState.Set("task", "Analyze this data")
inputState.Set("data", myData)

// Run agent
resultState, err := agent.Run(ctx, inputState)

// Get results
analysis, _ := resultState.Get("analysis")
summary, _ := resultState.Get("summary")
```

## Putting It All Together

Here's how these concepts work together:

```go
// 1. Create a provider (connection to LLM)
provider := provider.NewOpenAIProvider(apiKey, "gpt-4o")

// 2. Define what you want (schema)
type Analysis struct {
    Summary     string   `json:"summary"`
    KeyPoints   []string `json:"keyPoints"`
    Sentiment   string   `json:"sentiment"`
}

// 3. Create an agent with tools
agent := core.NewLLMAgent("analyzer", provider)
agent.AddTool(webSearchTool)
agent.AddTool(sentimentTool)

// 4. Prepare the task (state)
state := domain.NewState()
state.Set("url", "https://example.com/article")
state.Set("output_schema", Analysis{})

// 5. Run the agent
result, _ := agent.Run(ctx, state)

// 6. Get structured output
var analysis Analysis
output, _ := result.Get("output")
json.Unmarshal(output.([]byte), &analysis)
```

## Best Practices

### Providers
- Use environment variables for API keys
- Set appropriate timeouts via context
- Handle rate limits gracefully
- Consider using multi-provider for reliability

### Messages
- Keep system prompts focused and clear
- Include examples in system messages
- Maintain conversation context appropriately
- Use multimodal content when it adds value

### Schemas
- Start simple, add complexity as needed
- Use descriptions for better LLM understanding
- Test schemas with various inputs
- Enable type coercion for flexibility

### Agents
- Give agents focused, specific tasks
- Provide clear system prompts
- Monitor with hooks for debugging
- Limit iterations to prevent runaway costs

### State
- Keep state minimal and relevant
- Use meaningful keys
- Clone state when isolation is needed
- Clean up large data after use

## Next Steps

Now that you understand the core concepts:

- Learn about [Providers](providers.md) in detail
- Explore [Structured Output](structured-output.md) patterns
- Build [Agents](agents.md) for your use cases
- Master [Tools](tools.md) and [Workflows](workflows.md)

Ready to dive deeper? Choose your path and let's build something great!