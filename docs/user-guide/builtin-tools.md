# Built-in Tools Guide

This guide provides detailed documentation for all built-in tools in Go-LLMs, with a focus on advanced tools like the API Client.

## Table of Contents

1. [API Client Tool](#api-client-tool)
2. [Web Tools](#web-tools)
3. [File Tools](#file-tools)
4. [System Tools](#system-tools)
5. [Data Tools](#data-tools)
6. [Date Time Tools](#date-time-tools)
7. [Feed Tools](#feed-tools)

## API Client Tool

The `api_client` tool is a powerful, high-level API interaction tool that provides advanced features beyond basic HTTP requests. It supports REST APIs, GraphQL, and OpenAPI specifications with intelligent handling of authentication, pagination, rate limiting, and response transformation.

### Overview

**Tool Name**: `api_client`  
**Category**: `web`  
**Version**: `1.0.0`

### Key Features

1. **Multi-Protocol Support**
   - REST API (primary focus)
   - GraphQL queries and mutations
   - OpenAPI/Swagger specification discovery
   - Automatic endpoint discovery from specs

2. **Authentication Methods**
   - API Key (header, query, cookie)
   - Bearer Token (JWT, OAuth2)
   - Basic Authentication
   - OAuth2 flows (client credentials, authorization code)
   - Custom authentication headers
   - Session/cookie management

3. **Advanced Capabilities**
   - **Auto-Pagination**: Automatically follows pagination links
   - **Rate Limiting**: Respects rate limit headers with intelligent backoff
   - **Response Caching**: Cache responses with configurable TTL
   - **Request Templates**: Store and reuse common request patterns
   - **Response Transformation**: Extract data using JSONPath or JQ-like queries
   - **Error Recovery**: Smart retries with exponential backoff
   - **Mock Mode**: Optional mock responses for testing

### Parameters

```go
type APIClientParams struct {
    // Core request parameters
    Endpoint      string            `json:"endpoint"`        // Full URL or relative path
    Method        string            `json:"method"`          // HTTP method (GET, POST, etc.)
    APIType       string            `json:"api_type"`        // "rest", "graphql", or "openapi"
    
    // OpenAPI discovery
    SpecURL       string            `json:"spec_url"`        // OpenAPI spec URL
    OperationID   string            `json:"operation_id"`    // OpenAPI operation ID
    
    // Authentication
    Auth          AuthConfig        `json:"auth"`            // Authentication configuration
    
    // Request data
    PathParams    map[string]string `json:"path_params"`     // URL path parameters
    QueryParams   map[string]string `json:"query_params"`    // Query parameters
    Headers       map[string]string `json:"headers"`         // Custom headers
    Body          interface{}       `json:"body"`            // Request body
    
    // GraphQL specific
    GraphQLQuery  string            `json:"graphql_query"`   // GraphQL query
    Variables     map[string]interface{} `json:"variables"`  // GraphQL variables
    
    // Response handling
    ResponsePath  string            `json:"response_path"`   // JSONPath to extract
    Transform     string            `json:"transform"`       // JQ-like transformation
    
    // Advanced options
    Timeout       int               `json:"timeout"`         // Timeout in seconds
    CacheTTL      int               `json:"cache_ttl"`       // Cache TTL in seconds
    RetryCount    int               `json:"retry_count"`     // Max retry attempts
    FollowPages   bool              `json:"follow_pages"`    // Auto-follow pagination
    MockResponse  interface{}       `json:"mock_response"`   // Optional mock response
}
```

### Authentication Configuration

```go
type AuthConfig struct {
    Type          string            `json:"type"`            // "api_key", "bearer", "oauth2", "basic"
    APIKey        string            `json:"api_key"`         // API key value
    KeyLocation   string            `json:"key_location"`    // "header", "query", "cookie"
    KeyName       string            `json:"key_name"`        // Name of the key parameter
    Token         string            `json:"token"`           // Bearer token
    Username      string            `json:"username"`        // Basic auth username
    Password      string            `json:"password"`        // Basic auth password
    OAuth2Config  OAuth2Config      `json:"oauth2"`          // OAuth2 configuration
}
```

### Usage Examples

#### 1. Simple REST API Call

```go
// Call GitHub API to get repository information
result, err := apiClient.Execute(ctx, APIClientParams{
    Endpoint: "https://api.github.com/repos/lexlapax/go-llms",
    Method:   "GET",
    APIType:  "rest",
})
```

#### 2. Authenticated API Request

```go
// Call an API with Bearer token authentication
result, err := apiClient.Execute(ctx, APIClientParams{
    Endpoint: "https://api.example.com/users",
    Method:   "GET",
    APIType:  "rest",
    Auth: AuthConfig{
        Type:  "bearer",
        Token: "your-api-token",
    },
})
```

#### 3. OpenAPI Discovery

```go
// Discover and call an endpoint from OpenAPI spec
result, err := apiClient.Execute(ctx, APIClientParams{
    APIType:     "openapi",
    SpecURL:     "https://api.example.com/swagger.json",
    OperationID: "listUsers",
    QueryParams: map[string]string{
        "limit": "10",
    },
})
```

#### 4. GraphQL Query

```go
// Execute a GraphQL query
result, err := apiClient.Execute(ctx, APIClientParams{
    Endpoint:     "https://api.github.com/graphql",
    APIType:      "graphql",
    GraphQLQuery: `
        query($owner: String!, $repo: String!) {
            repository(owner: $owner, name: $repo) {
                stargazerCount
                forkCount
            }
        }
    `,
    Variables: map[string]interface{}{
        "owner": "lexlapax",
        "repo":  "go-llms",
    },
    Auth: AuthConfig{
        Type:  "bearer",
        Token: "your-github-token",
    },
})
```

#### 5. Auto-Pagination

```go
// Fetch all pages of results automatically
result, err := apiClient.Execute(ctx, APIClientParams{
    Endpoint:    "https://api.example.com/items",
    Method:      "GET",
    APIType:     "rest",
    FollowPages: true,  // Automatically fetch all pages
    QueryParams: map[string]string{
        "per_page": "100",
    },
})
```

#### 6. Response Transformation

```go
// Extract specific data from response using JSONPath
result, err := apiClient.Execute(ctx, APIClientParams{
    Endpoint:     "https://api.example.com/data",
    Method:       "GET",
    APIType:      "rest",
    ResponsePath: "$.items[*].name",  // Extract all item names
})
```

### Error Handling

The API Client provides detailed error messages with context:

- **Authentication Errors**: Clear indication of missing or invalid credentials
- **Rate Limit Errors**: Information about retry-after headers and backoff
- **Network Errors**: Connection issues with retry suggestions
- **Validation Errors**: OpenAPI schema validation failures
- **API Errors**: Structured error responses from the API

### State Integration

The API Client integrates with the agent state for:

- **Credential Storage**: Store API keys and tokens in state
- **Base URL Management**: Configure base URLs for relative endpoints
- **Default Headers**: Set common headers for all requests
- **Session Management**: Maintain cookies and session tokens

Example:
```go
// Store credentials in state
ctx.State.Set("github_token", "your-token")
ctx.State.Set("api_base_url", "https://api.github.com")

// The tool will automatically use these values
result, err := apiClient.Execute(ctx, APIClientParams{
    Endpoint: "/user/repos",  // Relative to base URL
    Method:   "GET",
    APIType:  "rest",
})
```

### Performance Considerations

1. **Caching**: Responses are cached based on URL and parameters
2. **Connection Pooling**: Reuses HTTP connections for efficiency
3. **Concurrent Requests**: Supports parallel requests with rate limiting
4. **Memory Management**: Streams large responses to avoid memory issues

### Security Best Practices

1. Never log sensitive credentials
2. Use environment variables for API keys
3. Validate SSL certificates
4. Sanitize user input in queries
5. Respect rate limits to avoid bans

## Web Tools

[Documentation for other web tools like web_fetch, web_search, etc.]

## File Tools

[Documentation for file tools]

## System Tools

[Documentation for system tools]

## Data Tools

[Documentation for data processing tools]

## Date Time Tools

[Documentation for date/time tools]

## Feed Tools

[Documentation for feed processing tools]