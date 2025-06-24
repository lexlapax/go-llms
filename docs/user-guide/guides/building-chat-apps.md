# Building Chat Applications

> **[User Guide](../README.md) / Guides / Building Chat Applications**

Learn how to build conversational AI applications from simple chatbots to sophisticated multi-turn conversation systems. This guide covers everything from basic setup to advanced features like memory, tools, and conversation management.

## What You'll Build

By the end of this guide, you'll have built:
1. **Simple Chatbot** - Basic question-and-answer system
2. **Memory-Enabled Chat** - Remembers conversation history
3. **Tool-Enabled Assistant** - Can search web, calculate, access files
4. **Multi-User Chat System** - Handles multiple conversations

## Prerequisites

- Completed [Quick Start](../getting-started/quickstart.md)
- Understanding of [Key Concepts](../getting-started/key-concepts.md)
- API key for your chosen provider

## 1. Simple Chatbot

Let's start with a basic chatbot that responds to user messages:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)

func main() {
    // Create provider
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    openaiProvider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create chat agent
    chatAgent := core.NewLLMAgent("chatbot", "gpt-4", core.LLMDeps{
        Provider: openaiProvider,
}
    
    // Configure personality
    chatAgent.SetSystemPrompt(`You are a friendly and helpful assistant. 
    Keep responses concise but informative. Show personality and warmth.`)
    
    fmt.Println("🤖 Chatbot started! Type 'quit' to exit.")
    fmt.Println("=====================================")
    
    scanner := bufio.NewScanner(os.Stdin)
    
    for {
        fmt.Print("You: ")
        if !scanner.Scan() {
            break
        }
        
        userInput := strings.TrimSpace(scanner.Text())
        if userInput == "quit" {
            fmt.Println("👋 Goodbye!")
            break
        }
        
        if userInput == "" {
            continue
        }
        
        // Create state with user input
        state := domain.NewState()
        state.Set("user_input", userInput)
        
        // Get AI response
        result, err := chatAgent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        if response, exists := result.Get("response"); exists {
            fmt.Printf("🤖 %s\n\n", response)
        }
    }
}
```

### Run Your First Chatbot

```bash
go run simple-chatbot.go
```

Example conversation:
```
🤖 Chatbot started! Type 'quit' to exit.
=====================================
You: Hello! What's your name?
🤖 Hi there! I'm your friendly AI assistant. I don't have a specific name, but you can call me whatever you'd like! How can I help you today?

You: What can you do?
🤖 I can help with a wide variety of tasks! I can answer questions, help with writing, explain concepts, assist with problem-solving, provide information on topics, help with coding, and much more. What would you like to work on?
```

## 2. Memory-Enabled Chat

Now let's add conversation memory so the chatbot remembers previous messages:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)

type ChatHistory struct {
    messages []domain.Message
}

func (ch *ChatHistory) AddUserMessage(text string) {
    ch.messages = append(ch.messages, domain.NewMessage(domain.RoleUser, text))
}

func (ch *ChatHistory) AddAssistantMessage(text string) {
    ch.messages = append(ch.messages, domain.NewMessage(domain.RoleAssistant, text))
}

func (ch *ChatHistory) GetMessages() []domain.Message {
    return ch.messages
}

func main() {
    // Create provider
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    openaiProvider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create chat agent
    chatAgent := core.NewLLMAgent("memory-chatbot", "gpt-4", core.LLMDeps{
        Provider: openaiProvider,
}
    
    // Configure with memory-aware personality
    chatAgent.SetSystemPrompt(`You are a friendly assistant with a good memory. 
    You remember what users tell you in the conversation and can reference previous topics.
    Be conversational and build on the context of the ongoing discussion.`)
    
    // Initialize conversation history
    history := &ChatHistory{}
    
    fmt.Println("🧠 Memory-Enabled Chatbot started! Type 'quit' to exit.")
    fmt.Println("I'll remember our conversation as we chat!")
    fmt.Println("================================================")
    
    scanner := bufio.NewScanner(os.Stdin)
    
    for {
        fmt.Print("You: ")
        if !scanner.Scan() {
            break
        }
        
        userInput := strings.TrimSpace(scanner.Text())
        if userInput == "quit" {
            fmt.Println("👋 It was great chatting with you! I'll remember our conversation.")
            break
        }
        
        if userInput == "" {
            continue
        }
        
        // Add user message to history
        history.AddUserMessage(userInput)
        
        // Create state with conversation history
        state := domain.NewState()
        state.Set("messages", history.GetMessages())
        
        // Get AI response
        result, err := chatAgent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        if response, exists := result.Get("response"); exists {
            responseText := response.(string)
            fmt.Printf("🧠 %s\n\n", responseText)
            
            // Add assistant response to history
            history.AddAssistantMessage(responseText)
        }
    }
}
```

### Enhanced Conversation Example

```
🧠 Memory-Enabled Chatbot started! Type 'quit' to exit.
I'll remember our conversation as we chat!
================================================
You: Hi! My name is Sarah and I'm a software engineer.
🧠 Nice to meet you, Sarah! It's great to chat with a fellow software engineer. What kind of development do you work on?

You: I mainly work with Python and machine learning.
🧠 That's fantastic, Sarah! Python and ML are such exciting fields. Are you working on any interesting machine learning projects at the moment?

You: What was my name again?
🧠 Your name is Sarah! You mentioned you're a software engineer who works with Python and machine learning.
```

## 3. Tool-Enabled Assistant

Let's add tools so your chatbot can perform actions like web search and calculations:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "math"
    "os"
    "strconv"
    "strings"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/agent/tools"
    "github.com/lexlapax/go-llms/pkg/schema"
)

func createCalculatorTool() domain.Tool {
    calculatorSchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "expression": {
                Type:        "string",
                Description: "Mathematical expression to evaluate (e.g., '2 + 3 * 4')",
            },
        },
        Required: []string{"expression"},
    }
    
    return tools.NewTool(
        "calculator",
        "Perform mathematical calculations",
        calculatorSchema,
        func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
            expression := params["expression"].(string)
            
            // Simple calculator - in production, use a proper math parser
            switch {
            case strings.Contains(expression, "+"):
                parts := strings.Split(expression, "+")
                if len(parts) == 2 {
                    a, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
                    b, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
                    return a + b, nil
                }
            case strings.Contains(expression, "*"):
                parts := strings.Split(expression, "*")
                if len(parts) == 2 {
                    a, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
                    b, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
                    return a * b, nil
                }
            case strings.Contains(expression, "sqrt"):
                numStr := strings.TrimSpace(strings.Replace(expression, "sqrt", "", 1))
                num, _ := strconv.ParseFloat(numStr, 64)
                return math.Sqrt(num), nil
            }
            
            return "Unable to calculate: " + expression, nil
        },
    )
}

func createTimeTool() domain.Tool {
    timeSchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "format": {
                Type:        "string",
                Description: "Time format requested (e.g., 'current', 'date', 'timestamp')",
                Default:     "current",
            },
        },
    }
    
    return tools.NewTool(
        "get_time",
        "Get current time and date information",
        timeSchema,
        func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
            format := "current"
            if f, ok := params["format"]; ok {
                format = f.(string)
            }
            
            switch format {
            case "date":
                return "Today is " + time.Now().Format("Monday, January 2, 2006"), nil
            case "timestamp":
                return time.Now().Unix(), nil
            default:
                return "Current time: " + time.Now().Format("3:04 PM MST"), nil
            }
        },
    )
}

func main() {
    // Create provider
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    openaiProvider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create assistant agent
    assistant := core.NewLLMAgent("smart-assistant", "gpt-4", core.LLMDeps{
        Provider: openaiProvider,
}
    
    // Configure with tool-aware personality
    assistant.SetSystemPrompt(`You are a helpful AI assistant with access to tools.
    When users ask for calculations, use the calculator tool.
    When users ask about time or date, use the get_time tool.
    Always explain what you're doing when you use tools.`)
    
    // Add tools
    assistant.AddTool(createCalculatorTool())
    assistant.AddTool(createTimeTool())
    
    fmt.Println("🛠️  Smart Assistant with Tools!")
    fmt.Println("I can help with calculations, time, and general questions.")
    fmt.Println("Try: 'What's 15 * 24?' or 'What time is it?'")
    fmt.Println("=======================================================")
    
    scanner := bufio.NewScanner(os.Stdin)
    
    for {
        fmt.Print("You: ")
        if !scanner.Scan() {
            break
        }
        
        userInput := strings.TrimSpace(scanner.Text())
        if userInput == "quit" {
            fmt.Println("👋 Goodbye! Feel free to come back anytime you need assistance.")
            break
        }
        
        if userInput == "" {
            continue
        }
        
        // Create state with user input
        state := domain.NewState()
        state.Set("user_input", userInput)
        
        // Get AI response (may use tools)
        result, err := assistant.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        if response, exists := result.Get("response"); exists {
            fmt.Printf("🛠️  %s\n\n", response)
        }
    }
}
```

### Tool Usage Examples

```
🛠️  Smart Assistant with Tools!
I can help with calculations, time, and general questions.
Try: 'What's 15 * 24?' or 'What time is it?'
=======================================================
You: What's 15 * 24?
🛠️  I'll calculate that for you! 15 * 24 = 360

You: What time is it?
🛠️  Let me check the current time for you. Current time: 2:45 PM MST

You: What's the square root of 144?
🛠️  I'll calculate the square root of 144 for you. sqrt(144) = 12
```

## 4. Production-Ready Chat System

![Chat Architecture](../../images/chat-architecture.svg)
*Production chat application architecture with configuration, session management, and error handling*

Here's a more robust chat system with proper error handling, configuration, and structure:

```go
package main

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/errors"
)

type ChatConfig struct {
    Provider    string `json:"provider"`
    Model       string `json:"model"`
    MaxHistory  int    `json:"max_history"`
    Timeout     int    `json:"timeout_seconds"`
    SystemPrompt string `json:"system_prompt"`
}

type ChatSession struct {
    ID       string           `json:"id"`
    Messages []domain.Message `json:"messages"`
    Created  time.Time       `json:"created"`
    Updated  time.Time       `json:"updated"`
}

type ChatApp struct {
    config   ChatConfig
    agent    domain.BaseAgent
    session  *ChatSession
}

func NewChatApp(configPath string) (*ChatApp, error) {
    // Load configuration
    config, err := loadConfig(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    
    // Create provider
    var prov domain.Provider
    switch config.Provider {
    case "openai":
        apiKey := os.Getenv("OPENAI_API_KEY")
        if apiKey == "" {
            return nil, fmt.Errorf("OPENAI_API_KEY not set")
        }
        prov = provider.NewOpenAIProvider(apiKey, config.Model)
    case "anthropic":
        apiKey := os.Getenv("ANTHROPIC_API_KEY")
        if apiKey == "" {
            return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
        }
        prov = provider.NewAnthropicProvider(apiKey, config.Model)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
    }
    
    // Create agent
    agent := core.NewLLMAgent("chat-app", config.Model, core.LLMDeps{
        Provider: prov,
}
    
    agent.SetSystemPrompt(config.SystemPrompt)
    agent.SetTimeout(time.Duration(config.Timeout) * time.Second)
    
    // Create session
    session := &ChatSession{
        ID:      fmt.Sprintf("session_%d", time.Now().Unix()),
        Created: time.Now(),
        Updated: time.Now(),
    }
    
    return &ChatApp{
        config:  config,
        agent:   agent,
        session: session,
    }, nil
}

func (app *ChatApp) Chat(userInput string) (string, error) {
    // Add user message to session
    userMsg := domain.NewMessage(domain.RoleUser, userInput)
    app.session.Messages = append(app.session.Messages, userMsg)
    
    // Trim history if too long
    if len(app.session.Messages) > app.config.MaxHistory*2 {
        // Keep system message + recent messages
        recentCount := app.config.MaxHistory * 2
        app.session.Messages = app.session.Messages[len(app.session.Messages)-recentCount:]
    }
    
    // Create state with conversation
    state := domain.NewState()
    state.Set("messages", app.session.Messages)
    state.SetMetadata("session_id", app.session.ID)
    
    // Get response with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 
        time.Duration(app.config.Timeout)*time.Second)
    defer cancel()
    
    result, err := app.agent.Run(ctx, state)
    if err != nil {
        return "", app.handleError(err)
    }
    
    response, exists := result.Get("response")
    if !exists {
        return "", fmt.Errorf("no response received")
    }
    
    responseText := response.(string)
    
    // Add assistant response to session
    assistantMsg := domain.NewMessage(domain.RoleAssistant, responseText)
    app.session.Messages = append(app.session.Messages, assistantMsg)
    app.session.Updated = time.Now()
    
    return responseText, nil
}

func (app *ChatApp) handleError(err error) error {
    // Provide user-friendly error messages
    var providerErr *errors.ProviderError
    if errors.As(err, &providerErr) {
        switch providerErr.Type {
        case errors.ErrTypeRateLimit:
            return fmt.Errorf("rate limit exceeded, please wait a moment and try again")
        case errors.ErrTypeAuthentication:
            return fmt.Errorf("authentication failed, please check your API key")
        case errors.ErrTypeContextLength:
            return fmt.Errorf("conversation too long, please start a new session")
        default:
            return fmt.Errorf("provider error: %s", providerErr.Message)
        }
    }
    
    return err
}

func loadConfig(configPath string) (ChatConfig, error) {
    // Default configuration
    config := ChatConfig{
        Provider:    "openai",
        Model:       "gpt-4",
        MaxHistory:  10,
        Timeout:     30,
        SystemPrompt: "You are a helpful, friendly, and knowledgeable AI assistant.",
    }
    
    // Try to load from file
    if configPath != "" {
        data, err := os.ReadFile(configPath)
        if err == nil {
            json.Unmarshal(data, &config)
        }
    }
    
    return config, nil
}

func main() {
    // Create chat app
    app, err := NewChatApp("chat-config.json")
    if err != nil {
        log.Fatal("Failed to initialize chat app:", err)
    }
    
    fmt.Println("💬 Production Chat System v1.0")
    fmt.Printf("Using %s/%s\n", app.config.Provider, app.config.Model)
    fmt.Printf("Session: %s\n", app.session.ID)
    fmt.Println("Type 'quit' to exit, 'help' for commands")
    fmt.Println("=====================================")
    
    scanner := bufio.NewScanner(os.Stdin)
    
    for {
        fmt.Print("You: ")
        if !scanner.Scan() {
            break
        }
        
        userInput := strings.TrimSpace(scanner.Text())
        
        switch userInput {
        case "quit", "exit":
            fmt.Printf("👋 Session %s ended. Thanks for chatting!\n", app.session.ID)
            return
        case "help":
            fmt.Println("Commands:")
            fmt.Println("  quit/exit - End the conversation")
            fmt.Println("  help      - Show this help")
            fmt.Println("  clear     - Clear conversation history")
            continue
        case "clear":
            app.session.Messages = []domain.Message{}
            fmt.Println("🧹 Conversation history cleared!")
            continue
        case "":
            continue
        }
        
        // Get AI response
        response, err := app.Chat(userInput)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        fmt.Printf("🤖 %s\n\n", response)
    }
}
```

Create a `chat-config.json` file:

```json
{
    "provider": "openai",
    "model": "gpt-4",
    "max_history": 10,
    "timeout_seconds": 30,
    "system_prompt": "You are a helpful AI assistant. Be conversational, informative, and concise. Remember the context of our conversation."
}
```

## Key Features Covered

✅ **Basic Chat Interface** - Terminal-based conversation  
✅ **Conversation Memory** - Remembers previous messages  
✅ **Tool Integration** - Calculator, time, and custom tools  
✅ **Error Handling** - Graceful error recovery  
✅ **Configuration** - Configurable providers and settings  
✅ **Session Management** - Conversation persistence  
✅ **Production Ready** - Timeouts, rate limiting, proper structure  

## Next Steps

Now that you've built chat applications, you can:

1. **Add More Tools** - [Custom Tools Guide](custom-tools.md)
2. **Web Interface** - [Web Applications Guide](web-applications.md)  
3. **Multi-User Support** - [APIs and Services Guide](apis-and-services.md)
4. **Advanced Features** - [Agent Communication](agent-communication.md)

## Common Patterns

### Message Formatting
```go
func formatMessage(role domain.Role, content string) domain.Message {
    return domain.NewMessage(role, content)
}
```

### Error Recovery
```go
func (app *ChatApp) safeChat(input string) string {
    response, err := app.Chat(input)
    if err != nil {
        return "I'm sorry, I encountered an issue. Please try again."
    }
    return response
}
```

### Tool Result Integration
```go
// Tools automatically integrated into conversation flow
if toolResult, exists := result.Get("tool_results"); exists {
    fmt.Printf("🔧 Used tools: %v\n", toolResult)
}
```

---

**Ready for more?** → [Build data extractors](building-data-extractors.md) | **Need tools?** → [Custom Tools Guide](../tools/custom-tools.md)