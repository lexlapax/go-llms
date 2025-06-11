# Tool Development Guide

This guide explains how to create custom tools for Go-LLMs agents using the enhanced ToolBuilder pattern introduced in v0.3.0.

## Table of Contents

- [Overview](#overview)
- [Tool Architecture](#tool-architecture)
- [Creating a Basic Tool](#creating-a-basic-tool)
- [Using ToolBuilder Pattern](#using-toolbuilder-pattern)
- [Tool Registration](#tool-registration)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Migration Guide](#migration-guide)

## Overview

Tools in Go-LLMs are reusable components that agents can use to perform specific tasks. They provide:

- **Type-safe interfaces** for parameters and outputs
- **Rich metadata** for LLM guidance
- **Event emission** for monitoring
- **State integration** for context awareness
- **Error handling** with helpful guidance

## Tool Architecture

### Core Components

1. **Tool Interface** (`domain.Tool`): The main interface all tools implement
2. **ToolBuilder**: Pattern for creating tools with comprehensive metadata
3. **ToolContext**: Runtime context providing state, events, and agent info
4. **Parameter/Output Schemas**: JSON schemas for validation

### Tool Lifecycle

```
1. Tool Creation → 2. Registration → 3. Discovery → 4. Execution
```

## Creating a Basic Tool

### Simple Function-Based Tool

For simple tools, you can use `NewToolBuilder` directly:

```go
package mytools

import (
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/tools"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Define parameter and result structs
type GreetingParams struct {
    Name     string `json:"name" description:"Name of the person to greet"`
    Language string `json:"language,omitempty" description:"Language for greeting (en, es, fr)"`
}

type GreetingResult struct {
    Greeting string `json:"greeting" description:"The greeting message"`
    Language string `json:"language" description:"Language used"`
}

// Create the execution function
func greetingExecute(ctx *domain.ToolContext, params GreetingParams) (*GreetingResult, error) {
    // Emit start event
    ctx.EmitEvent(domain.EventTypeToolStart, map[string]interface{}{
        "tool": "greeting",
        "params": params,
    })
    
    // Default to English if not specified
    if params.Language == "" {
        params.Language = "en"
    }
    
    // Generate greeting based on language
    var greeting string
    switch params.Language {
    case "es":
        greeting = "¡Hola, " + params.Name + "!"
    case "fr":
        greeting = "Bonjour, " + params.Name + "!"
    default:
        greeting = "Hello, " + params.Name + "!"
    }
    
    result := &GreetingResult{
        Greeting: greeting,
        Language: params.Language,
    }
    
    // Emit completion event
    ctx.EmitEvent(domain.EventTypeToolComplete, result)
    
    return result, nil
}

// Create the tool using ToolBuilder
func CreateGreetingTool() domain.Tool {
    paramSchema := &sdomain.Schema{
        Type: "object",
        Properties: map[string]*sdomain.Schema{
            "name": {
                Type:        "string",
                Description: "Name of the person to greet",
            },
            "language": {
                Type:        "string",
                Description: "Language for greeting (en, es, fr)",
                Default:     "en",
                Enum:        []interface{}{"en", "es", "fr"},
            },
        },
        Required: []string{"name"},
    }
    
    outputSchema := &sdomain.Schema{
        Type: "object",
        Properties: map[string]*sdomain.Schema{
            "greeting": {
                Type:        "string",
                Description: "The greeting message",
            },
            "language": {
                Type:        "string",
                Description: "Language used",
            },
        },
        Required: []string{"greeting", "language"},
    }
    
    return tools.NewToolBuilder("greeting", "Generate personalized greetings in multiple languages").
        WithCategory("text").
        WithTags("greeting", "language", "personalization").
        WithFunction(greetingExecute).
        WithParameterSchema(paramSchema).
        WithOutputSchema(outputSchema).
        WithUsageInstructions(`Use this tool to generate personalized greetings in different languages.
The tool supports English (en), Spanish (es), and French (fr).
If no language is specified, English will be used by default.`).
        WithConstraints(
            "Only supports English, Spanish, and French",
            "Name parameter is required",
            "Maximum name length is 100 characters",
        ).
        WithExamples(
            tools.Example{
                Name:        "English greeting",
                Description: "Generate an English greeting",
                Input:       map[string]interface{}{"name": "Alice"},
                Output:      map[string]interface{}{"greeting": "Hello, Alice!", "language": "en"},
            },
            tools.Example{
                Name:        "Spanish greeting",
                Description: "Generate a Spanish greeting",
                Input:       map[string]interface{}{"name": "Carlos", "language": "es"},
                Output:      map[string]interface{}{"greeting": "¡Hola, Carlos!", "language": "es"},
            },
        ).
        WithErrorGuidance(map[string]string{
            "empty name":           "The name parameter cannot be empty",
            "unsupported language": "Only en, es, and fr languages are supported",
        }).
        Build()
}
```

## Using ToolBuilder Pattern

The ToolBuilder pattern provides a fluent interface for creating tools with rich metadata:

### Required Components

```go
builder := tools.NewToolBuilder(name, description).
    WithFunction(executeFn).           // Execution function
    WithParameterSchema(paramSchema).  // Input validation
    WithOutputSchema(outputSchema)     // Output structure
```

### Optional Enhancements

```go
builder.
    // Categorization
    WithCategory("text").
    WithTags("greeting", "language").
    
    // Behavior hints
    WithDeterministic(true).
    WithDestructive(false).
    WithEstimatedLatency("fast").
    
    // LLM guidance
    WithUsageInstructions("detailed usage guide...").
    WithConstraints("constraint1", "constraint2").
    WithExamples(example1, example2).
    WithErrorGuidance(errorMap).
    
    // Dependencies
    WithDependencies("other_tool").
    WithAPIConfig(apiConfig)
```

## Tool Registration

### Manual Registration

```go
import "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"

func init() {
    tool := CreateGreetingTool()
    tools.RegisterTool(tool)
}
```

### Auto-Registration Pattern

Built-in tools use auto-registration via package init:

```go
package mytools

func init() {
    // Register all tools in this package
    tools.RegisterTool(CreateGreetingTool())
    tools.RegisterTool(CreateTranslationTool())
    // ... more tools
}
```

Then import the package to trigger registration:

```go
import _ "myproject/tools/mytools"
```

## Advanced Features

### State Integration

Access agent state from within tools:

```go
func stateAwareExecute(ctx *domain.ToolContext, params MyParams) (*MyResult, error) {
    // Read from state
    if apiKey, ok := ctx.State.Get("api_key"); ok {
        // Use API key from state
    }
    
    // Read agent metadata
    agentType := ctx.Agent.Type
    agentID := ctx.Agent.ID
    
    // Check previous values
    if lastResult, ok := ctx.State.Get("last_greeting"); ok {
        // Reference previous execution
    }
    
    return result, nil
}
```

### Event Emission

Emit events for monitoring and debugging:

```go
func eventEmittingExecute(ctx *domain.ToolContext, params MyParams) (*MyResult, error) {
    // Progress events
    ctx.EmitProgress(0, 100, "Starting processing")
    
    // Custom events
    ctx.EmitCustom("validation_complete", map[string]interface{}{
        "params": params,
        "valid": true,
    })
    
    // Error events
    if err != nil {
        ctx.EmitError(err)
        return nil, err
    }
    
    // Message events
    ctx.EmitMessage("Processing completed successfully")
    
    return result, nil
}
```

### Authentication Support

For tools that interact with external services:

```go
func authenticatedExecute(ctx *domain.ToolContext, params APIParams) (*APIResult, error) {
    // Check for authentication in state
    auth := extractAuthFromState(ctx.State)
    
    if auth == nil && params.APIKey == "" {
        return nil, fmt.Errorf("authentication required: provide api_key parameter or set in agent state")
    }
    
    // Use authentication
    client := createClient(auth)
    return client.Execute(params)
}
```

### Error Handling

Provide context-aware error messages:

```go
func robustExecute(ctx *domain.ToolContext, params MyParams) (*MyResult, error) {
    // Validation
    if params.Value < 0 {
        return nil, fmt.Errorf("value must be non-negative, got %d", params.Value)
    }
    
    // Operation with error context
    result, err := performOperation(params)
    if err != nil {
        // Wrap with context
        return nil, fmt.Errorf("operation failed for input %v: %w", params, err)
    }
    
    return result, nil
}
```

## Best Practices

### 1. Comprehensive Metadata

Always provide rich metadata for better LLM understanding:

```go
builder.
    WithUsageInstructions("detailed multi-line instructions...").
    WithConstraints(
        "Input must be valid JSON",
        "Maximum payload size is 1MB",
        "Requires authentication",
    ).
    WithExamples(
        // Provide 3-7 examples covering common cases
        basicExample,
        advancedExample,
        errorExample,
    )
```

### 2. Type Safety

Use strongly-typed parameters and results:

```go
// Good: Strongly typed
type SearchParams struct {
    Query    string   `json:"query"`
    Filters  []string `json:"filters,omitempty"`
    MaxItems int      `json:"max_items,omitempty"`
}

// Avoid: Generic map[string]interface{}
```

### 3. Event Emission

Emit meaningful events for observability:

```go
ctx.EmitEvent(domain.EventTypeToolStart, map[string]interface{}{
    "tool": "my_tool",
    "params": params,
    "timestamp": time.Now(),
})

// During execution
ctx.EmitProgress(completed, total, fmt.Sprintf("Processed %d/%d items", completed, total))

// On completion
ctx.EmitEvent(domain.EventTypeToolComplete, map[string]interface{}{
    "tool": "my_tool",
    "duration": time.Since(startTime),
    "result_size": len(result.Items),
})
```

### 4. Error Messages

Provide actionable error messages:

```go
// Good: Specific and actionable
return nil, fmt.Errorf("API rate limit exceeded (429): wait %d seconds before retry", retryAfter)

// Avoid: Generic
return nil, fmt.Errorf("request failed")
```

### 5. Resource Management

Clean up resources properly:

```go
func resourceAwareExecute(ctx *domain.ToolContext, params MyParams) (*MyResult, error) {
    // Use context for cancellation
    client := createClient()
    defer client.Close()
    
    // Respect context cancellation
    select {
    case <-ctx.Context.Done():
        return nil, ctx.Context.Err()
    default:
        return client.Execute(ctx.Context, params)
    }
}
```

## Migration Guide

### From Old Pattern to ToolBuilder

Old pattern (pre-v0.3.0):
```go
func CreateOldTool() domain.Tool {
    return tools.NewTool(
        "my_tool",
        "My tool description",
        paramSchema,
        outputSchema,
        executeFn,
    )
}
```

New pattern (v0.3.0+):
```go
func CreateNewTool() domain.Tool {
    return tools.NewToolBuilder("my_tool", "My tool description").
        WithFunction(executeFn).
        WithParameterSchema(paramSchema).
        WithOutputSchema(outputSchema).
        WithCategory("category").
        WithTags("tag1", "tag2").
        WithUsageInstructions("How to use this tool...").
        WithConstraints("constraint1", "constraint2").
        WithExamples(example1, example2).
        WithErrorGuidance(errorMap).
        Build()
}
```

### Key Differences

1. **Fluent Interface**: Chain configuration methods
2. **Rich Metadata**: Add instructions, examples, constraints
3. **Better Organization**: Categories and tags
4. **LLM Optimization**: Error guidance and usage instructions
5. **MCP Compatibility**: Export to Model Context Protocol format

### Migration Checklist

- [ ] Replace `tools.NewTool` with `tools.NewToolBuilder`
- [ ] Add `.WithFunction()` for execution function
- [ ] Add `.Build()` at the end
- [ ] Add category and tags
- [ ] Add usage instructions
- [ ] Add 3-7 examples
- [ ] Add constraints documentation
- [ ] Add error guidance map
- [ ] Test with LLM agents

## Example: Complete Tool

Here's a complete example of a web search tool:

```go
package webtools

import (
    "context"
    "fmt"
    "net/url"
    "time"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/tools"
    sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

type WebSearchParams struct {
    Query       string   `json:"query" description:"Search query"`
    MaxResults  int      `json:"max_results,omitempty" description:"Maximum results (1-50)"`
    Language    string   `json:"language,omitempty" description:"Language code (en, es, fr, etc)"`
    TimeRange   string   `json:"time_range,omitempty" description:"Time range (day, week, month, year)"`
}

type WebSearchResult struct {
    Results []SearchItem `json:"results" description:"Search results"`
    Total   int         `json:"total" description:"Total results found"`
    Query   string      `json:"query" description:"Original query"`
}

type SearchItem struct {
    Title       string    `json:"title"`
    URL         string    `json:"url"`
    Snippet     string    `json:"snippet"`
    PublishedAt time.Time `json:"published_at,omitempty"`
}

func webSearchExecute(ctx *domain.ToolContext, params WebSearchParams) (*WebSearchResult, error) {
    // Start event
    ctx.EmitEvent(domain.EventTypeToolStart, map[string]interface{}{
        "tool": "web_search",
        "query": params.Query,
    })
    
    // Validate parameters
    if params.Query == "" {
        return nil, fmt.Errorf("search query is required")
    }
    
    if params.MaxResults <= 0 {
        params.MaxResults = 10
    } else if params.MaxResults > 50 {
        params.MaxResults = 50
    }
    
    // Check for API key in state
    apiKey, _ := ctx.State.Get("search_api_key").(string)
    if apiKey == "" {
        return nil, fmt.Errorf("search API key not found in state")
    }
    
    // Progress event
    ctx.EmitProgress(1, 3, "Executing search query")
    
    // Perform search (simplified)
    results, total, err := performSearch(ctx.Context, apiKey, params)
    if err != nil {
        ctx.EmitError(err)
        return nil, fmt.Errorf("search failed: %w", err)
    }
    
    ctx.EmitProgress(2, 3, fmt.Sprintf("Found %d results", total))
    
    // Build result
    result := &WebSearchResult{
        Results: results,
        Total:   total,
        Query:   params.Query,
    }
    
    // Complete event
    ctx.EmitProgress(3, 3, "Search completed")
    ctx.EmitEvent(domain.EventTypeToolComplete, map[string]interface{}{
        "tool": "web_search",
        "results_count": len(results),
        "total": total,
    })
    
    return result, nil
}

func CreateWebSearchTool() domain.Tool {
    paramSchema := &sdomain.Schema{
        Type: "object",
        Properties: map[string]*sdomain.Schema{
            "query": {
                Type:        "string",
                Description: "Search query",
                MinLength:   1,
                MaxLength:   500,
            },
            "max_results": {
                Type:        "integer",
                Description: "Maximum number of results to return (1-50)",
                Default:     10,
                Minimum:     1,
                Maximum:     50,
            },
            "language": {
                Type:        "string",
                Description: "Language code for results (ISO 639-1)",
                Pattern:     "^[a-z]{2}$",
                Default:     "en",
            },
            "time_range": {
                Type:        "string",
                Description: "Filter results by time",
                Enum:        []interface{}{"day", "week", "month", "year"},
            },
        },
        Required: []string{"query"},
    }
    
    outputSchema := &sdomain.Schema{
        Type: "object",
        Properties: map[string]*sdomain.Schema{
            "results": {
                Type: "array",
                Items: &sdomain.Schema{
                    Type: "object",
                    Properties: map[string]*sdomain.Schema{
                        "title":        {Type: "string"},
                        "url":          {Type: "string", Format: "uri"},
                        "snippet":      {Type: "string"},
                        "published_at": {Type: "string", Format: "date-time"},
                    },
                    Required: []string{"title", "url", "snippet"},
                },
            },
            "total": {
                Type:        "integer",
                Description: "Total number of results found",
            },
            "query": {
                Type:        "string",
                Description: "The original search query",
            },
        },
        Required: []string{"results", "total", "query"},
    }
    
    return tools.NewToolBuilder("web_search", "Search the web for information").
        WithCategory("web").
        WithTags("search", "web", "research", "information").
        WithVersion("2.0.0").
        WithDeterministic(false).
        WithDestructive(false).
        WithRequiresConfirmation(false).
        WithEstimatedLatency("medium").
        WithFunction(webSearchExecute).
        WithParameterSchema(paramSchema).
        WithOutputSchema(outputSchema).
        WithUsageInstructions(`Use this tool to search the web for information on any topic.
        
The tool returns relevant web pages with titles, URLs, and text snippets.
Results are ranked by relevance and recency.

Tips for effective searches:
- Use specific keywords for better results
- Use quotes for exact phrases: "machine learning"
- Use minus to exclude terms: python -snake
- Specify time_range for recent information
- Set appropriate max_results based on needs`).
        WithConstraints(
            "Requires search_api_key in agent state",
            "Query length must be between 1-500 characters",
            "Maximum 50 results per search",
            "Some results may be filtered for safety",
            "Results depend on search engine availability",
        ).
        WithExamples(
            tools.Example{
                Name:        "Basic search",
                Description: "Simple keyword search",
                Input:       map[string]interface{}{"query": "golang generics"},
                Output: map[string]interface{}{
                    "results": []interface{}{
                        map[string]interface{}{
                            "title":   "Go 1.18 Release Notes",
                            "url":     "https://go.dev/doc/go1.18",
                            "snippet": "Go 1.18 includes support for generics...",
                        },
                    },
                    "total": 1250,
                    "query": "golang generics",
                },
            },
            tools.Example{
                Name:        "Search with options",
                Description: "Search with language and time filters",
                Input: map[string]interface{}{
                    "query":       "machine learning news",
                    "max_results": 5,
                    "language":    "en",
                    "time_range":  "week",
                },
                Output: map[string]interface{}{
                    "results": []interface{}{
                        map[string]interface{}{
                            "title":        "Latest ML Breakthrough",
                            "url":          "https://example.com/ml-news",
                            "snippet":      "Researchers announce new model...",
                            "published_at": "2024-01-15T10:00:00Z",
                        },
                    },
                    "total": 89,
                    "query": "machine learning news",
                },
            },
        ).
        WithErrorGuidance(map[string]string{
            "empty query":         "Provide a non-empty search query",
            "api key missing":     "Set search_api_key in agent state before using this tool",
            "rate limit":          "Too many requests, wait before searching again",
            "invalid time_range":  "Use one of: day, week, month, year",
            "network error":       "Check internet connection and try again",
        }).
        WithDependencies("search_api_key").
        Build()
}

// Mock search implementation
func performSearch(ctx context.Context, apiKey string, params WebSearchParams) ([]SearchItem, int, error) {
    // This would normally call a real search API
    // For this example, return mock data
    return []SearchItem{
        {
            Title:       "Example Result",
            URL:         "https://example.com",
            Snippet:     "This is an example search result",
            PublishedAt: time.Now(),
        },
    }, 1, nil
}
```

## Summary

The ToolBuilder pattern in Go-LLMs v0.3.0+ provides a powerful and flexible way to create tools with rich metadata that helps LLMs use them effectively. By following the patterns and best practices in this guide, you can create tools that are:

- **Easy to discover** through the registry system
- **Simple to use** with clear interfaces
- **Well-documented** with examples and constraints
- **Observable** through event emission
- **Reliable** with proper error handling
- **LLM-friendly** with comprehensive metadata

For more examples, explore the built-in tools in `pkg/agent/builtins/tools/` or see the [Built-in Tools Guide](builtin-tools.md).