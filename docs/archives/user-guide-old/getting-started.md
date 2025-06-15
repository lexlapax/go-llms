# Getting Started

Welcome to go-llms! This guide will get you up and running with the basics in minutes.

## Installation

```bash
go get github.com/lexlapax/go-llms
```

## Your First Program

Here's the simplest way to use go-llms:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Create a provider (OpenAI example)
    provider := provider.NewOpenAIProvider(
        "your-api-key",  // or use os.Getenv("OPENAI_API_KEY")
        "gpt-4o",
    )
    
    // Generate text
    response, err := provider.Generate(
        context.Background(), 
        "Explain quantum computing in one sentence",
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response)
}
```

## Structured Output

Want to get structured data from LLMs? Here's how:

```go
// Define what you want
type City struct {
    Name       string `json:"name"`
    Country    string `json:"country"`
    Population int    `json:"population"`
}

// Ask for it
var city City
err := provider.GenerateWithSchema(
    context.Background(),
    "Tell me about Tokyo",
    &city,
)

// Use it
fmt.Printf("%s, %s has %d people\n", 
    city.Name, city.Country, city.Population)
```

## Using Multiple Providers

Need reliability? Use multiple providers:

```go
import "github.com/lexlapax/go-llms/pkg/llm/provider"

// Create providers
openai := provider.NewOpenAIProvider(openaiKey, "gpt-4o")
anthropic := provider.NewAnthropicProvider(anthropicKey, "claude-3-5-sonnet-latest")

// Combine them
multi := provider.NewMultiProvider(
    []provider.ProviderConfig{
        {Provider: openai, Name: "openai"},
        {Provider: anthropic, Name: "anthropic"},
    },
    provider.StrategyPrimary, // Use OpenAI, fallback to Anthropic
)

// Use like any provider
response, _ := multi.Generate(ctx, "Hello!")
```

## Agents with Tools

Create an intelligent agent that can use tools:

```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
)

// Create an agent
agent := core.NewLLMAgent("assistant", provider)

// Add a built-in tool
calculator, _ := tools.GetTool("calculator")
agent.AddTool(calculator)

// Let the agent work
state := domain.NewState().Set("input", "What's 15% of 200?")
result, _ := agent.Run(context.Background(), state)

output, _ := result.Get("output")
fmt.Println(output) // "15% of 200 is 30"
```

## What's Next?

You've learned the basics! Here's where to go next:

### Learn Core Concepts
→ **[Core Concepts](core-concepts.md)** - Understand providers, messages, schemas, and agents

### Choose Your Path

**Working with LLMs:**
- [Providers Guide](providers.md) - Configure different LLM providers
- [Structured Output](structured-output.md) - Extract structured data reliably

**Building Agents:**
- [Agents Guide](agents.md) - Create autonomous agents
- [Tools Guide](tools.md) - Use and create tools
- [Workflows Guide](workflows.md) - Build complex workflows

**Going Deeper:**
- [Examples Gallery](examples-gallery.md) - See real-world examples
- [API Reference](../api/) - Detailed API documentation

## Quick Tips

1. **API Keys**: Store them in environment variables, not in code
2. **Error Handling**: Always check errors - LLM calls can fail
3. **Context**: Use `context.Context` for timeouts and cancellation
4. **Structured Output**: Define schemas for reliable data extraction

## Need Help?

- Check the [API Reference](../api/) for detailed documentation
- Browse [Examples](../../cmd/examples/) for working code
- Read about [Error Handling](error-handling.md) for robust applications

---

Ready to build something amazing? Let's go! 🚀