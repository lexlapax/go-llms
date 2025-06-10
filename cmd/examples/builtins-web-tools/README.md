# Built-in Web Tools Example

This example demonstrates the enhanced web tools available in the go-llms library, showcasing the new ToolBuilder interface and comprehensive authentication support.

## Overview

The built-in web tools provide functionality for:
- Fetching content from URLs with authentication, custom headers, and timeouts
- Searching the web using multiple engines (DuckDuckGo, Brave, Tavily, Serpapi, Serper.dev)
- Scraping web pages to extract structured data with authentication support
- Making advanced HTTP requests with comprehensive authentication methods

## Running the Example

```bash
go run main.go
```

## Available Web Tools

1. **web_fetch** (v3.0.0) - Fetch content from URLs with authentication
   - Multiple authentication methods (bearer, basic, API key, OAuth2, custom)
   - Custom timeouts and headers
   - Response metadata capture
   - Automatic auth detection from state

2. **web_search** (v3.0.0) - Multi-engine web search with auth support
   - Multiple search engines (DuckDuckGo, Brave, Tavily, Serpapi, Serper.dev)
   - Engine-specific API key management
   - Result limiting and safe search filtering
   - Authentication parameters for custom endpoints

3. **web_scrape** (v3.0.0) - Extract data from HTML with authentication
   - CSS-like selectors and content extraction
   - Link extraction and metadata parsing
   - Authentication support for protected pages
   - Text content cleaning and processing

4. **http_request** (v2.0.0) - Advanced HTTP operations
   - All HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
   - Comprehensive authentication (Basic, Bearer, API Key, OAuth2, custom)
   - Custom headers and query parameters
   - Request body support (JSON, form data, XML, plain text)
   - Redirect control and timeout management

## Example Usage

### Simple Web Fetch
```go
fetchTool := tools.MustGetTool("web_fetch")
result, _ := fetchTool.Execute(ctx, map[string]interface{}{
    "url": "https://api.github.com/users/github",
    "timeout": 10,
    "headers": map[string]string{
        "Accept": "application/json",
    },
})
```

### Web Search
```go
searchTool := tools.MustGetTool("web_search")
result, _ := searchTool.Execute(ctx, map[string]interface{}{
    "query": "golang generics tutorial",
    "engine": "tavily",
    "max_results": 5,
    "safe_search": true,
})
```

### Web Scraping
```go
scrapeTool := tools.MustGetTool("web_scrape")
result, _ := scrapeTool.Execute(ctx, map[string]interface{}{
    "url": "https://example.com",
    "extract_text": true,
    "extract_links": true,
    "extract_meta": true,
    "selectors": []string{"h1", "p", ".content"},
})
```

### HTTP POST Request
```go
httpTool := tools.MustGetTool("http_request")
result, _ := httpTool.Execute(ctx, map[string]interface{}{
    "url": "https://api.example.com/data",
    "method": "POST",
    "headers": map[string]string{
        "Content-Type": "application/json",
    },
    "body": map[string]interface{}{
        "key": "value",
    },
})
```

### Authenticated Requests

#### Web Fetch with Authentication
```go
fetchTool := tools.MustGetTool("web_fetch")
result, _ := fetchTool.Execute(ctx, map[string]interface{}{
    "url": "https://api.example.com/protected",
    "auth_type": "bearer",
    "auth_token": "your-token-here",
    "timeout": 10,
})
```

#### Web Scrape with Basic Auth
```go
scrapeTool := tools.MustGetTool("web_scrape")
result, _ := scrapeTool.Execute(ctx, map[string]interface{}{
    "url": "https://secure.example.com/content",
    "auth_type": "basic",
    "auth_username": "user",
    "auth_password": "pass",
    "extract_text": true,
})
```

#### HTTP Request with API Key
```go
httpTool := tools.MustGetTool("http_request")
result, _ := httpTool.Execute(ctx, map[string]interface{}{
    "url": "https://api.example.com/data",
    "method": "GET",
    "auth_type": "api_key",
    "auth_api_key": "your-api-key",
    "auth_key_name": "X-API-Key",
})
```

## Key Features

- **Enhanced ToolBuilder Interface**: All tools use the new ToolBuilder pattern with comprehensive metadata
- **Authentication Support**: Five authentication methods (bearer, basic, API key, OAuth2, custom)
- **Automatic Auth Detection**: Tools can detect authentication from agent state
- **Multi-Engine Search**: Support for 5+ search engines with API key management
- **Timeout Control**: Configurable timeouts for all network operations
- **Header Management**: Custom headers and advanced HTTP options
- **Structured Extraction**: CSS selectors and intelligent content parsing
- **Response Metadata**: Comprehensive status codes, headers, and timing information
- **Error Handling**: Detailed error reporting with LLM-friendly guidance

## Integration with Agents

These tools can be used with agents for web-based workflows:

```go
import "github.com/lexlapax/go-llms/pkg/agent/core"

// Create an LLM agent with web tools
agent := core.NewAgent("web-researcher", provider)
agent.SetSystemPrompt("You are a web research assistant with access to search, fetch, and scraping tools.")

// Add web tools to the agent
agent.AddTool(tools.MustGetTool("web_search"))
agent.AddTool(tools.MustGetTool("web_fetch"))
agent.AddTool(tools.MustGetTool("web_scrape"))
agent.AddTool(tools.MustGetTool("http_request"))

// Agent can now search, fetch, and extract information from the web
state := domain.NewState()
state.Set("user_input", "Search for Go generics tutorials and summarize the top results")
result, _ := agent.Run(ctx, state)
```

## Security Considerations

- The tools validate URLs and prevent access to private IP ranges by default
- Authentication credentials should be handled securely
- Be mindful of rate limits when making multiple requests
- Respect robots.txt and website terms of service