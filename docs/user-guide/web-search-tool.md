# Web Search Tool

> **[Documentation Home](/REFERENCE.md) / [User Guide](/docs/user-guide/) / Web Search Tool**

The Web Search tool provides comprehensive web search capabilities with support for multiple search engines, automatic API key detection, and intelligent result processing.

## Overview

The `web_search` tool allows agents to perform web searches using various search engines:

- **DuckDuckGo** - Free, no API key required (limited to instant answers)
- **Brave Search** - Full web search with comprehensive results (requires API key)
- **Tavily Search** - AI-optimized search designed for LLM applications (requires API key)
- **Serper.dev** - Fast Google search results API (requires API key)
- **Serpapi Search** - Google search results via API (requires API key)

## Features

- **Automatic Engine Selection**: Automatically selects the best available search engine based on API keys
- **Multiple Search Engines**: Support for DuckDuckGo, Brave, Tavily, Serper.dev, and Serpapi
- **Result Filtering**: Configurable result limits and safe search options
- **LLM Optimization**: Tavily provides AI-optimized results with context
- **Timeout Support**: Configurable timeout for search requests
- **Progress Events**: Real-time progress updates during search
- **Explicit API Keys**: Support for programmatic API key injection without environment variables

## Configuration

### Environment Variables

Set these environment variables to enable additional search engines:

```bash
# Brave Search API key
export BRAVE_API_KEY="your-brave-api-key"

# Tavily Search API key (recommended for LLM use)
export TAVILY_API_KEY="your-tavily-api-key"

# Serper.dev API key (fast Google search results)
export SERPERDEV_API_KEY="your-serperdev-api-key"

# Serpapi Search API key (Google search results)
export SERPAPI_API_KEY="your-serpapi-api-key"
```

### API Key Priority

When no engine is explicitly specified, the tool automatically selects based on available API keys:

1. **Tavily** (if TAVILY_API_KEY is set) - Best for LLM applications
2. **Serper.dev** (if SERPERDEV_API_KEY is set) - Fast Google search results
3. **Serpapi** (if SERPAPI_API_KEY is set) - Google search results
4. **Brave** (if BRAVE_API_KEY is set) - Comprehensive web search
5. **DuckDuckGo** (always available) - Limited to instant answers

## Usage

### Basic Search

```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

// Get the web search tool
searchTool, _ := tools.GetTool("web_search")

// Perform a search
result, err := searchTool.Execute(ctx, map[string]interface{}{
    "query": "artificial intelligence trends 2025",
})
```

### Specify Search Engine

```go
// Use Brave Search explicitly
result, err := searchTool.Execute(ctx, map[string]interface{}{
    "query": "golang concurrency patterns",
    "engine": "brave",
})

// Use Tavily Search for LLM-optimized results
result, err := searchTool.Execute(ctx, map[string]interface{}{
    "query": "quantum computing applications",
    "engine": "tavily",
})

// Use Serper.dev Search for fast Google results
result, err := searchTool.Execute(ctx, map[string]interface{}{
    "query": "machine learning algorithms",
    "engine": "serper",
})

// Use Serpapi Search for Google results
result, err := searchTool.Execute(ctx, map[string]interface{}{
    "query": "quantum computing fundamentals",
    "engine": "serpapi",
})
```

### Using Explicit API Keys

For production environments, you can provide API keys directly without relying on environment variables:

```go
// Use explicit API key (overrides environment variable)
result, err := searchTool.Execute(ctx, map[string]interface{}{
    "query":          "machine learning algorithms",
    "engine":         "tavily",
    "engine_api_key": secretManager.GetAPIKey("tavily-prod"),
})

// Parallel searches with different API keys
results := make(chan *WebSearchResults, 3)

// Search with organization's primary account
go func() {
    res, _ := searchTool.Execute(ctx, map[string]interface{}{
        "query":          "market analysis",
        "engine":         "brave",
        "engine_api_key": orgKeys.BravePrimary,
    })
    results <- res
}()

// Search with backup account
go func() {
    res, _ := searchTool.Execute(ctx, map[string]interface{}{
        "query":          "market analysis",
        "engine":         "serpapi",
        "engine_api_key": orgKeys.SerpapiBackup,
    })
    results <- res
}()
```

### Advanced Parameters

```go
result, err := searchTool.Execute(ctx, map[string]interface{}{
    "query":          "machine learning",
    "max_results":    5,              // Limit to 5 results (default: 10, max: 50)
    "safe_search":    true,           // Enable safe search (default: true)
    "timeout":        30,             // Timeout in seconds (default: 30)
    "engine":         "tavily",       // Specify engine explicitly
    "engine_api_key": "optional-key", // Optional: Override environment variable
})
```

## Search Engines

### DuckDuckGo

- **Pros**: No API key required, always available
- **Cons**: Limited to instant answers, not full web search
- **Best for**: Quick facts, definitions, basic information

### Brave Search

- **Pros**: Full web search, comprehensive results, privacy-focused
- **Cons**: Requires API key
- **Best for**: General web search, news, diverse content
- **API**: Get key at https://brave.com/search/api/

### Tavily Search

- **Pros**: AI-optimized, includes answer summaries, designed for LLMs
- **Cons**: Requires API key
- **Best for**: LLM applications, research agents, context-aware search
- **API**: Get key at https://tavily.com/

### Serper.dev

- **Pros**: Fast Google search results, simple API, low latency
- **Cons**: Requires API key
- **Best for**: Quick search results, low-latency applications, real-time search
- **API**: Get key at https://serper.dev/

### Serpapi Search

- **Pros**: Google search results, excellent ranking, comprehensive results
- **Cons**: Requires API key
- **Best for**: General search needs, finding authoritative sources, latest information
- **API**: Get key at https://serpapi.com/

## Result Format

The tool returns a `WebSearchResults` object:

```go
type WebSearchResults struct {
    Query      string         // Original search query
    Engine     string         // Search engine used
    Results    []SearchResult // Array of search results
    TotalFound int           // Number of results found
    TimeMs     int64         // Search time in milliseconds
}

type SearchResult struct {
    Title       string // Result title
    URL         string // Result URL
    Description string // Result description/summary
    Snippet     string // Additional snippet/context
}
```

### Tavily Special Features

When using Tavily, the first result may be an AI-generated summary:

```go
// Check for AI summary
if result.Results[0].URL == "tavily:answer:..." {
    summary := result.Results[0].Description
    // Use the AI-generated summary
}
```

## Integration with Agents

### LLM Agent Example

```go
agent := core.NewAgent("researcher", provider)
agent.AddTool(tools.MustGetTool("web_search"))

// The agent can now use web search in conversations
response, _ := agent.Generate(ctx, []domain.Message{
    {Role: "user", Content: "Search for the latest AI developments"},
})
```

### Custom Agent Example

```go
type ResearchAgent struct {
    *core.BaseAgentImpl
}

func (r *ResearchAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    searchTool := tools.MustGetTool("web_search")
    
    // Perform search
    results, err := searchTool.Execute(&domain.ToolContext{
        Context: ctx,
        State:   domain.NewStateReader(state),
    }, map[string]interface{}{
        "query": state.GetString("topic"),
        "max_results": 10,
    })
    
    // Process results...
    return state, nil
}
```

## Best Practices

1. **API Key Management**: 
   - Development: Use environment variables for convenience
   - Production: Use explicit API keys from secure key management systems
   - Never hardcode API keys in source code
2. **Engine Selection**: Let the tool auto-select unless you need specific features
3. **Result Limits**: Use appropriate limits to avoid overwhelming downstream processing
4. **Error Handling**: Always handle API key missing errors gracefully
5. **Caching**: Consider caching results for repeated queries

### Production API Key Management

```go
// Best practice: Inject keys from secure storage
type SearchConfig struct {
    BraveAPIKey    string
    TavilyAPIKey   string
    SerperdevAPIKey string
    SerpapiAPIKey  string
}

func NewSearchAgent(config SearchConfig) *Agent {
    agent := core.NewAgent("searcher", provider)
    searchTool := tools.MustGetTool("web_search")
    
    // Create wrapper that injects API key
    agent.AddTool(func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        // Auto-select engine based on available keys
        if config.TavilyAPIKey != "" && params["engine"] == nil {
            params["engine"] = "tavily"
            params["engine_api_key"] = config.TavilyAPIKey
        } else if config.SerperdevAPIKey != "" && params["engine"] == nil {
            params["engine"] = "serper"
            params["engine_api_key"] = config.SerperdevAPIKey
        } else if config.SerpapiAPIKey != "" && params["engine"] == nil {
            params["engine"] = "serpapi"
            params["engine_api_key"] = config.SerpapiAPIKey
        }
        
        return searchTool.Execute(ctx, params)
    })
    
    return agent
}
```

## Troubleshooting

### No Results Returned

- **DuckDuckGo**: Only returns instant answers, not general web results
- **Solution**: Use Brave or Tavily for comprehensive web search

### API Key Errors

```bash
# Check if API keys are set
echo $BRAVE_API_KEY
echo $TAVILY_API_KEY
echo $SERPERDEV_API_KEY

# Set API keys
export BRAVE_API_KEY="your-key"
export TAVILY_API_KEY="your-key"
export SERPERDEV_API_KEY="your-key"
```

### Rate Limiting

- Implement exponential backoff for rate-limited APIs
- Consider caching results to reduce API calls
- Use appropriate timeouts

## Example: Research Assistant

See the [agent-custom-research example](/cmd/examples/agent-custom-research/) for a complete implementation of a research assistant that uses the web search tool with automatic engine selection based on available API keys.