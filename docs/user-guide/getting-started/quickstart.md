# Quick Start

> **[User Guide](../README.md) / Getting Started / Quick Start**

Get your first AI conversation working in 5 minutes! This guide assumes you have Go installed and an OpenAI API key.

![Quick Start Steps](../../images/quickstart-steps.svg)
*Visual guide to the 4-step quick start process*

## Prerequisites

- Go 1.21 or later installed
- OpenAI API key ([get one here](https://platform.openai.com/api-keys))
- 5 minutes of your time

## Step 1: Create a New Go Project

```bash
# Create a new directory
mkdir my-ai-app
cd my-ai-app

# Initialize Go module
go mod init my-ai-app

# Add go-llms dependency
go get github.com/lexlapax/go-llms
```

## Step 2: Set Your API Key

```bash
# Set environment variable (replace with your actual key)
export OPENAI_API_KEY="sk-your-openai-key-here"

# Or create a .env file
echo "OPENAI_API_KEY=sk-your-openai-key-here" > .env
```

## Step 3: Create Your First AI Program

Create a file called `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    agentdomain "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Get API key from environment
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    // Create a provider
    openaiProvider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create an agent
    agent := core.NewLLMAgent("assistant", "A helpful AI assistant", core.LLMDeps{
        Provider: openaiProvider,
    })
    
    // Set system prompt
    agent.SetSystemPrompt("You are a helpful assistant.")
    
    // Create state with user input
    state := agentdomain.NewState()
    state.Set("user_input", "Hello! What can you help me with today?")
    
    // Run the agent
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal("Error running agent:", err)
    }
    
    // Get and print the response
    if response, exists := result.Get("response"); exists {
        fmt.Println("User: Hello! What can you help me with today?")
        fmt.Println("AI:", response)
    }
}
```

## Step 4: Run Your Program

```bash
go run main.go
```

You should see output like:
```
User: Hello! What can you help me with today?
AI: Hello! I'm here to help you with a wide variety of tasks. I can assist with questions, writing, analysis, coding, math, creative projects, and much more. What would you like to work on?
```

## 🎉 Congratulations!

You've successfully created your first AI-powered application with go-llms! 

## What Just Happened?

1. **Provider**: You created an OpenAI provider that connects to GPT-4
2. **Agent**: You created an LLM agent that can have conversations
3. **State**: You used state to pass data (your message) to the agent
4. **Execution**: The agent processed your input and generated a response

## Next Steps

### Try Different Prompts

Modify the `user_input` to try different questions:

```go
state.Set("user_input", "Explain quantum computing in simple terms")
// or
state.Set("user_input", "Write a haiku about programming")
// or  
state.Set("user_input", "What are the benefits of Go programming language?")
```

### Add a System Prompt

Make your agent more specialized:

```go
agent.SetSystemPrompt("You are a expert Go programming tutor. Always provide code examples and explain concepts clearly.")
```

### Try a Different Provider

Switch to Claude (Anthropic):

```go
// Get Anthropic API key
anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
if anthropicKey == "" {
    log.Fatal("Please set ANTHROPIC_API_KEY environment variable")
}

// Create Anthropic provider
anthropicProvider := provider.NewAnthropicProvider(anthropicKey, "claude-3-sonnet-20240229")

// Use in agent
agent := core.NewLLMAgent("assistant", "A helpful AI assistant", core.LLMDeps{
    Provider: anthropicProvider,
})
```

### Common Issues and Solutions

#### "Module not found" error
```bash
go mod tidy
go mod download
```

#### "API key not found" error
```bash
# Check your environment variable
echo $OPENAI_API_KEY

# Make sure it starts with "sk-"
```

#### Rate limit errors
```bash
# Wait a moment and try again, or use gpt-3.5-turbo
openaiProvider := provider.NewOpenAIProvider(apiKey, "gpt-3.5-turbo")
```

## What's Next?

Now that you have a working AI application, you can:

1. **[Learn Key Concepts](key-concepts.md)** - Understand providers, agents, and state
2. **[Complete Installation Guide](installation.md)** - Set up a proper development environment
3. **[First Steps Tutorial](first-steps.md)** - Build 3 progressively complex programs
4. **[Choose Your Provider](choosing-providers.md)** - Learn about different AI providers

## Need Help?

- **Quick Questions**: [API Quick Reference](../reference/api-quick-reference.md)
- **Installation Issues**: [Installation Guide](installation.md)
- **Concept Questions**: [Key Concepts](key-concepts.md)
- **More Examples**: [Beginner Projects](../examples/beginner-projects.md)

---

**Working?** → [Learn the key concepts](key-concepts.md) | **Having issues?** → [Complete installation guide](installation.md)