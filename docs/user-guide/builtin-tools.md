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

The `api_client` tool provides a high-level interface for making REST API calls with intelligent error handling and authentication support. This is Phase 1 implementation focusing on core REST functionality with plans for OpenAPI and GraphQL support in future phases.

### Overview

**Tool Name**: `api_client`  
**Category**: `web`  
**Version**: `1.0.0`

### Current Features (Phase 1)

1. **REST API Support**
   - Full HTTP method support (GET, POST, PUT, DELETE, PATCH, etc.)
   - JSON request and response handling
   - Path parameter substitution (e.g., `/users/{user_id}/posts`)
   - Query parameter support
   - Custom headers

2. **Authentication Methods**
   - API Key authentication (header, query parameter)
   - Bearer Token authentication
   - Basic Authentication (username/password)
   - Custom authentication headers

3. **Advanced Error Handling**
   - Intelligent error categorization (4xx client errors, 5xx server errors)
   - Detailed error context with API response details
   - LLM-friendly error guidance for troubleshooting
   - Status code interpretation and recommendations

### Planned Features (Future Phases)

- **Phase 2**: OpenAPI/Swagger specification discovery and validation
- **Phase 3**: GraphQL query support, rate limiting, response caching, pagination handling

### Parameters

**Required Parameters:**
- `base_url` (string): The base URL of the API (e.g., "https://api.github.com")
- `endpoint` (string): The API endpoint path (e.g., "/repos/owner/repo")

**Optional Parameters:**
- `method` (string): HTTP method (default: "GET")
- `headers` (map[string]string): Custom HTTP headers
- `query_params` (map[string]interface{}): URL query parameters
- `path_params` (map[string]string): Path parameter substitutions for URLs like `/users/{user_id}`
- `body` (interface{}): Request body (automatically JSON-encoded)
- `auth` (object): Authentication configuration
- `timeout` (integer): Request timeout in seconds (default: 30)

### Authentication Configuration

The `auth` parameter supports the following authentication types:

```json
{
  "type": "api_key",
  "api_key": "your-api-key",
  "key_location": "header",
  "key_name": "X-API-Key"
}
```

```json
{
  "type": "bearer",
  "token": "your-bearer-token"
}
```

```json
{
  "type": "basic",
  "username": "your-username",
  "password": "your-password"
}
```

### Usage Examples

#### 1. Simple GET Request

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/repos/lexlapax/go-llms",
  "method": "GET"
}
```

#### 2. POST Request with JSON Body

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/users",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

#### 3. API Key Authentication

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/protected/data",
  "method": "GET",
  "auth": {
    "type": "api_key",
    "api_key": "your-api-key",
    "key_location": "header",
    "key_name": "X-API-Key"
  }
}
```

#### 4. Bearer Token Authentication

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/user/repos",
  "method": "GET",
  "auth": {
    "type": "bearer",
    "token": "ghp_your_github_token"
  }
}
```

#### 5. Path Parameters

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/users/{user_id}/posts",
  "method": "GET",
  "path_params": {
    "user_id": "12345"
  }
}
```

#### 6. Query Parameters

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/search",
  "method": "GET",
  "query_params": {
    "q": "go llm framework",
    "limit": "10",
    "page": "1"
  }
}
```

### Response Format

The API Client returns a structured response with the following fields:

```json
{
  "success": true,
  "status_code": 200,
  "headers": {
    "content-type": "application/json",
    "server": "GitHub.com"
  },
  "data": {
    "name": "go-llms",
    "full_name": "lexlapax/go-llms",
    "description": "Go library for LLM providers",
    "stargazers_count": 42
  }
}
```

### Error Handling

For error responses, the tool provides detailed context:

```json
{
  "success": false,
  "status_code": 404,
  "error_details": {
    "error": "Not Found",
    "message": "Repository not found"
  },
  "error_guidance": "This is a client-side error (4xx). Check the endpoint URL and ensure the resource exists."
}
```

**Error Categories:**
- **4xx Client Errors**: URL, authentication, or request format issues
- **5xx Server Errors**: API server problems (retry recommended)
- **Network Errors**: Connection timeouts or DNS resolution failures
- **Validation Errors**: Invalid parameter values or missing required fields

### LLM Integration

The API Client is designed to work seamlessly with LLM agents:

1. **Descriptive Errors**: Error messages include guidance for LLMs on how to fix issues
2. **Flexible Parameters**: Accepts both structured and string parameters
3. **Context Awareness**: Can infer common patterns and suggest fixes
4. **State Integration**: Stores credentials and base URLs in agent state

### Performance

**Benchmark Results** (Apple M1 Ultra):
- **Throughput**: 250,000+ requests/second for simple operations
- **Latency**: ~50 microseconds for basic GET requests
- **Memory Usage**: 10-14 KB per request
- **Concurrency**: Excellent parallel performance

### Security Best Practices

1. Store API keys in environment variables or agent state, not in prompts
2. Use HTTPS endpoints only
3. Validate SSL certificates (default behavior)
4. Avoid logging sensitive request/response data
5. Use appropriate authentication methods for each API

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