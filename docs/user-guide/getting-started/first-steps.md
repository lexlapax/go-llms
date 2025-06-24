# First Steps: Your First 3 Programs

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Getting Started](../../user-guide/getting-started) / First Steps**

Learn Go-LLMs step-by-step by building three progressively more complex applications. Each example builds on the previous one, introducing new concepts and capabilities.

## Prerequisites

- [Installation completed](installation.md) ✅
- At least one API key configured (or Ollama running) ✅
- Basic Go knowledge ✅

## Learning Path Overview

![Learning Progression](../../images/quickstart-steps.svg)

1. **Hello AI** - Basic text generation (5 minutes)
2. **Smart Assistant** - Agent with built-in tools (10 minutes)
3. **Data Extractor** - Structured outputs with validation (10 minutes)

Each program is self-contained and introduces 2-3 new concepts.

---

## Program 1: Hello AI
*Learn: Basic providers, text generation, error handling*

### Goal
Create your first AI conversation with minimal code.

### Code

Create `hello_ai.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)

func main() {
    fmt.Println("🤖 Hello AI - Your First Go-LLMs Program")
    fmt.Println("=========================================")

    // Step 1: Create a provider
    // Try multiple providers with fallback to mock
    var p domain.Provider
    var err error

    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        fmt.Println("✓ Using OpenAI GPT-4o-mini")
        p = provider.NewOpenAIProvider(apiKey, "gpt-4o-mini")
    } else if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        fmt.Println("✓ Using Anthropic Claude")
        p = provider.NewAnthropicProvider(apiKey, "claude-3-5-haiku-latest")
    } else {
        fmt.Println("ℹ No API keys found, using mock provider")
        p = provider.NewMockProvider()
    }

    // Step 2: Create an agent
    agent := core.NewLLMAgent("assistant", "gpt-4o-mini", core.LLMDeps{
        Provider: p,
}

    // Step 3: Set system prompt
    agent.SetSystemPrompt("You are a helpful assistant that gives concise, friendly responses.")

    // Step 4: Have a conversation
    state := domain.NewState()
    
    conversations := []string{
        "Hello! What can you help me with?",
        "Explain quantum computing in one sentence.",
        "What's the weather like on Mars?",
    }

    for i, question := range conversations {
        fmt.Printf("\n--- Conversation %d ---\n", i+1)
        fmt.Printf("You: %s\n", question)

        // Set the user input
        state.Set("user_input", question)

        // Run the agent
        result, err := agent.Run(context.Background(), state)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }

        // Get the response
        if response, exists := result.Get("response"); exists {
            fmt.Printf("AI: %v\n", response)
        }
    }

    fmt.Println("\n🎉 Congratulations! You've completed your first AI conversation.")
}
```

### Run It

```bash
# With OpenAI (recommended for beginners)
export OPENAI_API_KEY="your-key-here"
go run hello_ai.go

# Or with Anthropic
export ANTHROPIC_API_KEY="your-key-here"
go run hello_ai.go

# Or without any API keys (uses mock)
go run hello_ai.go
```

### What You Learned

✅ **Provider Creation** - How to create and configure LLM providers  
✅ **Agent Basics** - Creating agents that wrap providers with additional capabilities  
✅ **State Management** - Using state objects to pass data between interactions  
✅ **Error Handling** - Graceful fallbacks when API keys aren't available  
✅ **Multiple Providers** - Supporting different providers in the same application  

### Try These Experiments

- Change the system prompt to make the AI respond differently
- Try different providers by setting different environment variables
- Modify the questions and see how responses change

---

## Program 2: Smart Assistant  
*Learn: Built-in tools, agent capabilities, tool integration*

### Goal
Create an AI assistant that can use tools to perform web searches, calculations, and file operations.

### Code

Create `smart_assistant.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    
    // Import built-in tools
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
)

func main() {
    fmt.Println("🛠️ Smart Assistant - Agent with Tools")
    fmt.Println("=====================================")

    // Step 1: Create an agent using the simple string-based API
    agent, err := core.NewAgentFromString("smart-assistant", "openai/gpt-4o-mini")
    if err != nil {
        // Fallback to mock if no API key
        fmt.Printf("⚠️ Provider creation failed (%v), using mock agent\n", err)
        mockAgent, _ := core.NewAgentFromString("smart-assistant", "mock")
        agent = mockAgent
    }

    // Step 2: Configure the agent
    agent.SetSystemPrompt(`You are a smart assistant with access to tools. 
Use tools when helpful to provide accurate, up-to-date information.
Be conversational and explain what you're doing when using tools.`)

    // Step 3: Add built-in tools
    fmt.Println("🔧 Adding tools to the assistant...")

    // Calculator for math operations
    calculator := math.NewCalculatorTool()
    agent.AddTool(calculator)
    fmt.Println("✓ Calculator tool added")

    // Web search (requires API key)
    if searchKey := os.Getenv("SEARCH_API_KEY"); searchKey != "" {
        webSearch := web.NewWebSearchTool(searchKey)
        agent.AddTool(webSearch)
        fmt.Println("✓ Web search tool added")
    } else {
        fmt.Println("ℹ SEARCH_API_KEY not set, skipping web search tool")
    }

    // File operations
    fileRead := file.NewFileReadTool()
    agent.AddTool(fileRead)
    fmt.Println("✓ File read tool added")

    // System information
    sysInfo := system.NewSystemInfoTool()
    agent.AddTool(sysInfo)
    fmt.Println("✓ System info tool added")

    // Step 4: Interactive tasks that use different tools
    tasks := []struct {
        description string
        prompt      string
    }{
        {
            description: "Math calculation",
            prompt:      "Calculate the compound interest on $1000 at 5% annual rate for 10 years",
        },
        {
            description: "System information",
            prompt:      "What operating system am I running and what's the current time?",
        },
        {
            description: "File operation", 
            prompt:      "Read the contents of go.mod file if it exists",
        },
        {
            description: "Web search (if available)",
            prompt:      "What's the latest stable version of Go?",
        },
    }

    state := domain.NewState()

    for i, task := range tasks {
        fmt.Printf("\n--- Task %d: %s ---\n", i+1, task.description)
        fmt.Printf("Request: %s\n", task.prompt)

        // Set the user input
        state.Set("user_input", task.prompt)

        // Run the agent
        result, err := agent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }

        // Get the response
        if response, exists := result.Get("response"); exists {
            fmt.Printf("Assistant: %v\n", response)
        }

        // Show tool usage if any tools were called
        if toolCalls, exists := result.Get("tool_calls"); exists {
            fmt.Printf("🔧 Tools used: %v\n", toolCalls)
        }
    }

    fmt.Println("\n🎉 Great! You've created an assistant with multiple capabilities.")
    fmt.Println("💡 Try adding more tools or asking different questions!")
}
```

### Run It

```bash
# Basic functionality (works without additional API keys)
export OPENAI_API_KEY="your-key"
go run smart_assistant.go

# Enhanced with web search
export SEARCH_API_KEY="your-search-key"  # Optional
go run smart_assistant.go
```

### What You Learned

✅ **String-based Agent Creation** - Using the convenient `NewAgentFromString` API  
✅ **Built-in Tools** - Adding pre-built tools for common tasks  
✅ **Tool Categories** - Math, web, file, and system tools  
✅ **Tool Integration** - How agents automatically use tools when appropriate  
✅ **State Persistence** - How state carries over between tool calls  
✅ **Capability Management** - Graceful degradation when optional tools aren't available  

### Try These Experiments

- Add more tools from different categories
- Ask questions that require multiple tool calls
- Create custom prompts that test specific tool combinations
- Monitor what tools get called for different types of questions

---

## Program 3: Data Extractor
*Learn: Structured outputs, schemas, validation, error recovery*

### Goal
Extract structured data from unstructured text with guaranteed JSON output that matches your schema.

### Code

Create `data_extractor.go`:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/schema/domain" as schemaDomain
    "github.com/lexlapax/go-llms/pkg/schema/validation"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// Customer represents a customer record
type Customer struct {
    Name        string   `json:"name"`
    Email       string   `json:"email"`
    Phone       string   `json:"phone"`
    Company     string   `json:"company"`
    Industry    string   `json:"industry"`
    Interests   []string `json:"interests"`
    Priority    string   `json:"priority"`
    NextAction  string   `json:"next_action"`
}

// Review represents a product review
type Review struct {
    ProductName string  `json:"product_name"`
    Rating      int     `json:"rating"`
    Sentiment   string  `json:"sentiment"`
    Summary     string  `json:"summary"`
    Keywords    []string `json:"keywords"`
    Verified    bool    `json:"verified"`
}

func main() {
    fmt.Println("📊 Data Extractor - Structured Outputs")
    fmt.Println("=====================================")

    // Step 1: Create an agent
    agent, err := core.NewAgentFromString("extractor", "openai/gpt-4o-mini")
    if err != nil {
        fmt.Printf("⚠️ Using mock agent: %v\n", err)
        agent, _ = core.NewAgentFromString("extractor", "mock")
    }

    agent.SetSystemPrompt(`You are a data extraction specialist. 
Extract structured information from unstructured text and return valid JSON.
Be accurate and thorough in your extraction.`)

    // Step 2: Define schemas for different data types
    customerSchema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "name":        {Type: "string", Description: "Full name"},
            "email":       {Type: "string", Format: "email", Description: "Email address"},
            "phone":       {Type: "string", Description: "Phone number"},
            "company":     {Type: "string", Description: "Company name"},
            "industry":    {Type: "string", Description: "Industry sector"},
            "interests":   {Type: "array", Items: &schemaDomain.Property{Type: "string"}, Description: "Areas of interest"},
            "priority":    {Type: "string", Enum: []interface{}{"high", "medium", "low"}, Description: "Contact priority"},
            "next_action": {Type: "string", Description: "Recommended next action"},
        },
        Required: []string{"name", "email", "priority"},
    }

    reviewSchema := &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "product_name": {Type: "string", Description: "Product name"},
            "rating":       {Type: "integer", Minimum: float64Ptr(1), Maximum: float64Ptr(5), Description: "Rating 1-5"},
            "sentiment":    {Type: "string", Enum: []interface{}{"positive", "negative", "neutral"}, Description: "Overall sentiment"},
            "summary":      {Type: "string", Description: "Brief summary"},
            "keywords":     {Type: "array", Items: &schemaDomain.Property{Type: "string"}, Description: "Key terms"},
            "verified":     {Type: "boolean", Description: "Is this a verified purchase"},
        },
        Required: []string{"product_name", "rating", "sentiment", "summary"},
    }

    // Step 3: Create validator and processor
    validator := validation.NewValidator()
    structuredProcessor := processor.NewStructuredProcessor(validator)

    // Step 4: Test data extraction scenarios
    scenarios := []struct {
        name   string
        schema *schemaDomain.Schema
        text   string
        target interface{}
    }{
        {
            name:   "Customer Lead Extraction",
            schema: customerSchema,
            target: &Customer{},
            text: `Email from Sarah Johnson (sarah.johnson@techcorp.com) at TechCorp Industries. 
            She's interested in our enterprise software solutions and cloud migration services. 
            Mentioned they're a fintech company looking to modernize their infrastructure. 
            Phone: +1-555-0123. This seems like a high-priority lead - they have budget allocated 
            for Q1 and want to schedule a technical demo next week.`,
        },
        {
            name:   "Product Review Analysis",
            schema: reviewSchema,
            target: &Review{},
            text: `I absolutely love the UltraBook Pro laptop! ⭐⭐⭐⭐⭐ 
            The performance is incredible and the battery life is amazing. 
            Perfect for programming and design work. The build quality feels premium 
            and the screen is gorgeous. Definitely worth the investment. 
            Verified purchase from authorized retailer.`,
        },
    }

    state := domain.NewState()

    for i, scenario := range scenarios {
        fmt.Printf("\n--- Scenario %d: %s ---\n", i+1, scenario.name)
        fmt.Printf("Input text: %s\n", truncateText(scenario.text, 100))

        // Method 1: Using agent with schema
        fmt.Println("\n🤖 Method 1: Agent-based extraction")
        
        // Set the schema on the agent
        agent.SetSchema(scenario.schema)
        
        state.Set("user_input", fmt.Sprintf("Extract structured data from: %s", scenario.text))
        
        result, err := agent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Agent error: %v\n", err)
        } else if structured, exists := result.Get("structured_output"); exists {
            fmt.Println("✓ Agent extracted data:")
            printJSON(structured)
        }

        // Method 2: Direct processing with structured processor
        fmt.Println("\n🔧 Method 2: Direct processor extraction")
        
        prompt := fmt.Sprintf("Extract structured data from this text: %s", scenario.text)
        
        // Simulate LLM response (in real usage, this would come from the LLM)
        rawLLMResponse := generateMockResponse(scenario.name)
        
        processedData, err := structuredProcessor.Process(scenario.schema, rawLLMResponse)
        if err != nil {
            fmt.Printf("❌ Processing error: %v\n", err)
        } else {
            fmt.Println("✓ Processor extracted data:")
            printJSON(processedData)
        }

        // Method 3: Type-safe extraction
        fmt.Println("\n🎯 Method 3: Type-safe extraction")
        
        err = structuredProcessor.ProcessTyped(scenario.schema, rawLLMResponse, scenario.target)
        if err != nil {
            fmt.Printf("❌ Typed processing error: %v\n", err)
        } else {
            fmt.Println("✓ Type-safe extraction:")
            printJSON(scenario.target)
        }
    }

    fmt.Println("\n🎉 Excellent! You've mastered structured data extraction.")
    fmt.Println("💡 Try creating your own schemas for different data types!")
}

// Helper functions
func float64Ptr(v float64) *float64 {
    return &v
}

func truncateText(text string, maxLen int) string {
    if len(text) <= maxLen {
        return text
    }
    return text[:maxLen] + "..."
}

func printJSON(v interface{}) {
    data, _ := json.MarshalIndent(v, "", "  ")
    fmt.Println(string(data))
}

func generateMockResponse(scenarioName string) string {
    switch scenarioName {
    case "Customer Lead Extraction":
        return `{
            "name": "Sarah Johnson",
            "email": "sarah.johnson@techcorp.com",
            "phone": "+1-555-0123",
            "company": "TechCorp Industries",
            "industry": "fintech",
            "interests": ["enterprise software", "cloud migration", "infrastructure modernization"],
            "priority": "high",
            "next_action": "Schedule technical demo for next week"
        }`
    case "Product Review Analysis":
        return `{
            "product_name": "UltraBook Pro laptop",
            "rating": 5,
            "sentiment": "positive", 
            "summary": "Customer loves the performance, battery life, and build quality",
            "keywords": ["performance", "battery life", "programming", "design", "premium", "gorgeous screen"],
            "verified": true
        }`
    default:
        return "{}"
    }
}
```

### Run It

```bash
export OPENAI_API_KEY="your-key"
go run data_extractor.go
```

### What You Learned

✅ **Schema Definition** - Creating JSON schemas for data validation  
✅ **Structured Agents** - Configuring agents to return structured data  
✅ **Multiple Extraction Methods** - Agent-based, processor-based, and type-safe extraction  
✅ **Data Validation** - Automatic validation against schemas  
✅ **Error Recovery** - Graceful handling of invalid or malformed data  
✅ **Type Safety** - Converting JSON to Go structs safely  

### Try These Experiments

- Create schemas for your own data types (events, products, etc.)
- Test with more complex nested data structures
- Try extraction with different providers to see quality differences
- Add custom validation rules to your schemas

---

## What's Next?

🎉 **Congratulations!** You've completed the first steps and learned:

- Basic LLM provider usage
- Agent creation and configuration
- Tool integration and capabilities
- Structured data extraction and validation

### Continue Your Journey

Choose your path based on what interests you most:

#### 🏗️ **Building Applications**
- **[Chat Applications](../guides/building-chat-apps.md)** - Create conversational interfaces
- **[Data Extractors](../guides/building-data-extractors.md)** - Build production data processing
- **[Research Agents](../guides/building-research-agents.md)** - Information gathering systems

#### 🔧 **Advanced Features**  
- **[Agent Communication](../guides/agent-communication.md)** - Multi-agent coordination
- **[Custom Tools](../advanced/custom-tools.md)** - Build your own tools
- **[Workflow Orchestration](../advanced/workflow-orchestration.md)** - Complex automation

#### 📚 **References & Examples**
- **[80+ Examples](../../cmd/examples/README.md)** - Complete working examples
- **[API Quick Reference](../reference/api-quick-reference.md)** - Essential API patterns
- **[Built-in Tools Reference](../reference/built-in-tools-reference.md)** - Complete tool catalog

### Getting Help

- **Quick Questions**: [API Reference](../reference/api-quick-reference.md)
- **Community**: [GitHub Discussions](https://github.com/lexlapax/go-llms/discussions)
- **Issues**: [GitHub Issues](https://github.com/lexlapax/go-llms/issues)

---

**Ready for the next challenge?** Pick your learning path above and keep building! 🚀