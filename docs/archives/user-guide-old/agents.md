# Building Agents

This guide shows you how to build and use agents - autonomous entities that can use tools to accomplish tasks.

## Overview

Agents in go-llms are powered by LLMs and can:
- Make decisions based on context
- Use tools to interact with external systems
- Maintain state across operations
- Work together in complex workflows

## Types of Agents

### 1. LLM Agents
Basic agents that use an LLM provider to make decisions and optionally use tools.

```go
import "github.com/lexlapax/go-llms/pkg/agent/core"

// Create a simple LLM agent
agent := core.NewLLMAgent("assistant", provider)

// Set its personality
agent.SetSystemPrompt("You are a helpful research assistant")

// Give it tools
agent.AddTool(searchTool)
agent.AddTool(calculatorTool)
```

### 2. Workflow Agents
Pre-built patterns for common orchestration needs:
- **Sequential**: Execute agents one after another
- **Parallel**: Run multiple agents simultaneously
- **Conditional**: Branch based on conditions
- **Loop**: Iterate until a condition is met

### 3. Custom Agents
Build your own agent types for unique requirements.

## Creating Your First Agent

### Basic Agent Setup

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Create provider
    provider := provider.NewOpenAIProvider(
        "your-api-key",
        "gpt-4o",
    )
    
    // Create agent
    agent := core.NewLLMAgent("assistant", provider)
    
    // Set system prompt
    agent.SetSystemPrompt("You are a helpful assistant")
    
    // Run the agent
    state := domain.NewState()
    state.Set("input", "What's the capital of France?")
    
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.GetString("output"))
}
```

### Adding Tools

Tools extend what agents can do:

```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

// Get built-in tools
searchTool, _ := tools.GetTool("web_search")
fetchTool, _ := tools.GetTool("web_fetch")

// Add to agent
agent.AddTool(searchTool)
agent.AddTool(fetchTool)

// Now the agent can search and fetch web content
state := domain.NewState()
state.Set("input", "Find the latest news about AI")

result, _ := agent.Run(context.Background(), state)
```

## Working with State

State is how data flows through agents:

```go
// Create initial state
state := domain.NewState()
state.Set("task", "Analyze sales data")
state.Set("data", salesData)
state.Set("format", "summary")

// Run agent
result, _ := agent.Run(ctx, state)

// Get results
summary, _ := result.Get("summary")
insights, _ := result.Get("insights")
```

## Agent Communication

### Direct Communication

```go
// Agents can work together
researcher := core.NewLLMAgent("researcher", provider)
writer := core.NewLLMAgent("writer", provider)

// Researcher finds information
researchState := domain.NewState()
researchState.Set("topic", "quantum computing")
research, _ := researcher.Run(ctx, researchState)

// Writer creates article
writeState := domain.NewState()
writeState.Set("research", research.Get("findings"))
writeState.Set("style", "blog post")
article, _ := writer.Run(ctx, writeState)
```

### Using Sub-Agents

```go
// Create a coordinator with sub-agents
coordinator := core.NewLLMAgent("coordinator", provider)

// Add sub-agents as tools
coordinator.AddAgent(researcher)
coordinator.AddAgent(writer)
coordinator.AddAgent(editor)

// Coordinator can now delegate tasks
state := domain.NewState()
state.Set("task", "Create a blog post about AI")
result, _ := coordinator.Run(ctx, state)
```

## Workflow Patterns

### Sequential Workflow

```go
import "github.com/lexlapax/go-llms/pkg/agent/workflow"

// Create a data processing pipeline
pipeline := workflow.NewSequentialAgent("data-pipeline")

// Add steps
pipeline.AddAgent(dataExtractor)
pipeline.AddAgent(dataValidator)
pipeline.AddAgent(dataProcessor)
pipeline.AddAgent(reportGenerator)

// Run pipeline
result, _ := pipeline.Run(ctx, inputState)
```

### Parallel Processing

```go
// Process multiple tasks simultaneously
parallel := workflow.NewParallelAgent("analyzer")

// Add parallel tasks
parallel.AddAgent(textAnalyzer)
parallel.AddAgent(sentimentAnalyzer)
parallel.AddAgent(keywordExtractor)

// Run all analyzers at once
results, _ := parallel.Run(ctx, textState)

// Results are merged automatically
sentiment := results.Get("sentiment")
keywords := results.Get("keywords")
```

### Conditional Branching

```go
// Route based on conditions
router := workflow.NewConditionalAgent("router")

// Add branches
router.AddBranch(
    "urgent",
    func(s *domain.State) bool {
        priority := s.GetString("priority")
        return priority == "high"
    },
    urgentHandler,
)

router.AddBranch(
    "normal",
    func(s *domain.State) bool {
        return true // Default branch
    },
    normalHandler,
)

// Automatically routes to correct handler
result, _ := router.Run(ctx, taskState)
```

### Loops

```go
// Iterate until condition is met
refiner := workflow.NewLoopAgent("refiner")
refiner.SetLoopAgent(refinementAgent)
refiner.SetMaxIterations(5)
refiner.SetCondition(func(s *domain.State) bool {
    quality := s.GetFloat64("quality_score")
    return quality >= 0.9
})

// Keeps refining until quality threshold is met
result, _ := refiner.Run(ctx, draftState)
```

## Advanced Features

### System Prompts

```go
// Detailed system prompt for specialized behavior
agent.SetSystemPrompt(`You are an expert data analyst.

Your responsibilities:
1. Analyze data for patterns and insights
2. Provide clear, actionable recommendations
3. Use visualizations when helpful
4. Always validate your findings

Format your responses with:
- Executive Summary
- Key Findings
- Recommendations
- Supporting Data`)
```

### Tool Selection

```go
// Agent automatically selects appropriate tools
agent.AddTool(calculatorTool)
agent.AddTool(webSearchTool)
agent.AddTool(databaseTool)

// Agent will choose the right tool based on the task
state.Set("input", "Calculate the ROI if we invest $10000")
// Agent uses calculator

state.Set("input", "Find the latest market trends")
// Agent uses web search

state.Set("input", "Get last quarter's sales data")
// Agent uses database
```

### Error Handling

```go
// Implement retry logic
func runWithRetry(agent domain.BaseAgent, state *domain.State) (*domain.State, error) {
    var lastErr error
    
    for i := 0; i < 3; i++ {
        result, err := agent.Run(context.Background(), state)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    return nil, fmt.Errorf("failed after 3 attempts: %w", lastErr)
}
```

### Monitoring with Hooks

```go
// Add monitoring
agent.AddHook(&MonitoringHook{
    OnStart: func(ctx context.Context, state *domain.State) {
        log.Printf("Agent %s started", agent.Name())
    },
    OnComplete: func(ctx context.Context, result *domain.State, err error) {
        if err != nil {
            log.Printf("Agent %s failed: %v", agent.Name(), err)
        } else {
            log.Printf("Agent %s completed successfully", agent.Name())
        }
    },
})
```

## Real-World Examples

### Research Assistant

```go
// Create a research assistant
researcher := core.NewLLMAgent("researcher", provider)
researcher.SetSystemPrompt("You are a thorough research assistant")

// Add research tools
researcher.AddTool(webSearchTool)
researcher.AddTool(webFetchTool)
researcher.AddTool(summarizerTool)

// Research a topic
state := domain.NewState()
state.Set("topic", "renewable energy trends 2024")
state.Set("depth", "comprehensive")

result, _ := researcher.Run(ctx, state)

// Get structured research report
report := result.Get("research_report")
sources := result.Get("sources")
```

### Customer Support Agent

```go
// Create support agent
support := core.NewLLMAgent("support", provider)
support.SetSystemPrompt(`You are a helpful customer support agent.
Always be polite and professional.
Try to resolve issues in one interaction.`)

// Add support tools
support.AddTool(knowledgeBaseTool)
support.AddTool(ticketSystemTool)
support.AddTool(escalationTool)

// Handle customer query
state := domain.NewState()
state.Set("customer_id", "12345")
state.Set("issue", "Can't login to my account")

result, _ := support.Run(ctx, state)
```

### Data Analysis Pipeline

```go
// Create analysis pipeline
pipeline := workflow.NewSequentialAgent("analysis-pipeline")

// Data ingestion
ingester := core.NewLLMAgent("ingester", provider)
ingester.AddTool(csvReaderTool)
ingester.AddTool(jsonProcessorTool)

// Data cleaning
cleaner := core.NewLLMAgent("cleaner", provider)
cleaner.AddTool(dataValidatorTool)
cleaner.AddTool(dataTransformTool)

// Analysis
analyzer := core.NewLLMAgent("analyzer", provider)
analyzer.AddTool(statisticsTool)
analyzer.AddTool(visualizationTool)

// Add to pipeline
pipeline.AddAgent(ingester)
pipeline.AddAgent(cleaner)
pipeline.AddAgent(analyzer)

// Run analysis
result, _ := pipeline.Run(ctx, dataState)
```

## Best Practices

### 1. Clear System Prompts
- Be specific about the agent's role
- Define expected behavior clearly
- Include output format requirements

### 2. Tool Selection
- Only add tools the agent needs
- Group related tools together
- Provide clear tool descriptions

### 3. State Management
- Keep state minimal and relevant
- Use meaningful key names
- Clean up large data after use

### 4. Error Recovery
- Implement retry logic for network operations
- Provide fallback behaviors
- Log errors for debugging

### 5. Performance
- Use parallel agents for independent tasks
- Cache results when appropriate
- Set reasonable timeouts

## Common Patterns

### Supervisor Pattern
```go
// Supervisor manages team of specialized agents
supervisor := core.NewLLMAgent("supervisor", provider)
supervisor.AddAgent(researchAgent)
supervisor.AddAgent(analysisAgent)
supervisor.AddAgent(reportAgent)

// Supervisor coordinates the team
state.Set("project", "Market analysis for new product")
result, _ := supervisor.Run(ctx, state)
```

### Pipeline Pattern
```go
// Each agent transforms data for the next
extract := createExtractAgent()
transform := createTransformAgent()
load := createLoadAgent()

pipeline := workflow.NewSequentialAgent("etl")
pipeline.AddAgent(extract)
pipeline.AddAgent(transform)
pipeline.AddAgent(load)
```

### Consensus Pattern
```go
// Multiple agents vote on best answer
agents := []domain.BaseAgent{
    createExpertAgent1(),
    createExpertAgent2(),
    createExpertAgent3(),
}

consensus := workflow.NewParallelAgent("consensus")
for _, agent := range agents {
    consensus.AddAgent(agent)
}

// Merge strategy determines final answer
consensus.SetMergeStrategy(workflow.MergeByVoting)
```

## Debugging Agents

### Enable Detailed Logging
```go
agent.SetDebugMode(true)
agent.AddHook(&LoggingHook{
    LogLevel: "debug",
})
```

### Trace Execution
```go
// Add tracing hook
agent.AddHook(&TracingHook{
    OnToolCall: func(tool string, params map[string]interface{}) {
        fmt.Printf("Calling tool %s with %v\n", tool, params)
    },
})
```

### Inspect State
```go
// Check state at each step
result, _ := agent.Run(ctx, state)
fmt.Printf("Final state keys: %v\n", result.Keys())
for _, key := range result.Keys() {
    fmt.Printf("%s: %v\n", key, result.Get(key))
}
```

## Next Steps

Now that you understand agents:
- Explore [Tools](tools.md) to extend agent capabilities
- Learn about [Workflows](workflows.md) for complex orchestration
- See [Examples Gallery](examples-gallery.md) for more patterns
- Check the [API Reference](../api/agent.md) for detailed documentation

Ready to build intelligent agents? Let's go! 🤖