# Built-in Web Tools Example

This example demonstrates the web tools available in the go-llms library.

## Overview

The built-in web tools provide functionality for:
- Fetching content from URLs with custom headers and timeouts
- Searching the web using DuckDuckGo
- Scraping web pages to extract structured data
- Making advanced HTTP requests with various methods and authentication

## Running the Example

```bash
go run main.go
```

## Available Web Tools

1. **web_fetch** - Fetch content from URLs
   - Custom timeouts
   - Request headers
   - Response metadata capture

2. **web_search** - Search the web
   - DuckDuckGo integration
   - Result limiting
   - Safe search filtering

3. **web_scrape** - Extract data from HTML
   - CSS-like selectors
   - Link extraction
   - Metadata parsing
   - Text content extraction

4. **http_request** - Advanced HTTP operations
   - All HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
   - Authentication support (Basic, Bearer, API Key)
   - Custom headers and query parameters
   - Request body support (JSON, form data, XML, plain text)
   - Redirect control

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
    "max_results": 5,
    "safe_search": "moderate",
})
```

### Web Scraping
```go
scrapeTool := tools.MustGetTool("web_scrape")
result, _ := scrapeTool.Execute(ctx, map[string]interface{}{
    "url": "https://example.com",
    "extract_options": map[string]interface{}{
        "include_text": true,
        "include_links": true,
        "selectors": []string{"h1", ".content", "#main"},
    },
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

### Authenticated Request
```go
result, _ := httpTool.Execute(ctx, map[string]interface{}{
    "url": "https://api.example.com/protected",
    "method": "GET",
    "auth_type": "bearer",
    "auth_bearer_token": "your-token-here",
})
```

## Key Features

- **Timeout Control**: All tools support configurable timeouts
- **Header Management**: Custom headers for all HTTP operations
- **Authentication**: Multiple auth methods (Basic, Bearer, API Key)
- **Safe Search**: Content filtering for web searches
- **Structured Extraction**: CSS selectors for targeted scraping
- **Response Metadata**: Capture status codes, headers, and timing
- **Error Handling**: Comprehensive error reporting

## Integration with Agents

These tools can be used with agents for web-based workflows:

```go
agent := workflow.NewAgent(
    "web-researcher",
    provider,
    workflow.WithTools(
        tools.MustGetTool("web_search"),
        tools.MustGetTool("web_fetch"),
        tools.MustGetTool("web_scrape"),
    ),
)

// Agent can now search, fetch, and extract information from the web
response, _ := agent.Run(ctx, workflow.UserMessage(
    "Search for Go generics tutorials and summarize the top results",
))
```

## Security Considerations

- The tools validate URLs and prevent access to private IP ranges by default
- Authentication credentials should be handled securely
- Be mindful of rate limits when making multiple requests
- Respect robots.txt and website terms of service