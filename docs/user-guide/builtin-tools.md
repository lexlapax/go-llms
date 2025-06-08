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

The `api_client` tool provides a high-level interface for making REST API calls with intelligent error handling, authentication support, and OpenAPI specification discovery.

### Overview

**Tool Name**: `api_client`  
**Category**: `web`  
**Version**: `2.0.0` (Updated January 7, 2025 with OpenAPI support)

### Current Features

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

4. **OpenAPI/Swagger Support** (Added in Phase 2)
   - OpenAPI 3.0/3.1 specification parsing
   - Automatic operation discovery from specs
   - Request validation against OpenAPI schemas
   - Parameter and request body validation
   - LLM-friendly operation guidance
   - Support for JSON and YAML spec formats
   - **New in v2.0.0** (January 8, 2025):
     - Automatic server URL resolution from OpenAPI specs
     - Security scheme detection and automatic authentication
     - Enhanced error guidance with OpenAPI context
     - Operation metadata with parameter counts and requirements

### Planned Features (Future Phases)

- **Phase 3**: GraphQL query support
- **Phase 4**: Advanced authentication (OAuth2, JWT refresh)
- **Phase 5**: Rate limiting, response caching, pagination handling, streaming responses

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
- `openapi_spec` (string): URL to OpenAPI/Swagger spec for automatic discovery and validation. When provided, enables operation discovery mode
- `discover_operations` (boolean): If true, returns available operations from the OpenAPI spec instead of making an API call

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

#### 7. OpenAPI Discovery

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/",
  "openapi_spec": "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json",
  "discover_operations": true
}
```

#### 8. OpenAPI-Validated Request

```json
{
  "base_url": "https://petstore3.swagger.io/api/v3",
  "endpoint": "/pet/findByStatus",
  "method": "GET",
  "query_params": {
    "status": "available"
  },
  "openapi_spec": "https://petstore3.swagger.io/api/v3/openapi.json"
}
```

#### 9. Automatic Server URL Resolution

When `base_url` is omitted, the tool automatically uses the first server URL from the OpenAPI spec:

```json
{
  "endpoint": "/pet/findByStatus",
  "method": "GET",
  "query_params": {
    "status": "available"
  },
  "openapi_spec": "https://petstore3.swagger.io/api/v3/openapi.json"
}
```

#### 10. Automatic Authentication Detection

The tool can detect and apply authentication from agent state based on OpenAPI security schemes:

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/protected/resource",
  "method": "GET",
  "openapi_spec": "https://api.example.com/openapi.json"
}
```

If the OpenAPI spec defines security requirements and credentials are found in agent state (e.g., `api_key`, `bearer_token`), they will be automatically applied.

### Response Format

The API Client returns a structured response with the following fields:

**Standard API Response:**
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

**OpenAPI Discovery Response:**
```json
{
  "success": true,
  "operations": [
    {
      "operationId": "repos/get",
      "method": "GET",
      "path": "/repos/{owner}/{repo}",
      "summary": "Get a repository",
      "description": "Get a repository by owner and repo name",
      "tags": ["repos"],
      "pathParameterCount": 2,
      "deprecated": false
    },
    {
      "operationId": "repos/listForUser",
      "method": "GET",
      "path": "/users/{username}/repos",
      "summary": "List repositories for a user",
      "tags": ["repos"],
      "pathParameterCount": 1,
      "queryParameterCount": 4,
      "deprecated": false
    }
  ],
  "spec_info": {
    "title": "GitHub v3 REST API",
    "version": "1.1.4",
    "description": "GitHub's v3 REST API"
  },
  "servers": [
    "https://api.github.com"
  ],
  "security_schemes": {
    "oauth2": {
      "type": "oauth2",
      "description": "GitHub OAuth2 authentication"
    },
    "bearer": {
      "type": "http",
      "scheme": "bearer",
      "description": "Bearer token authentication"
    }
  },
  "total_operations": 912,
  "llm_guidance": "API: GitHub v3 REST API v1.1.4\nDescription: GitHub's v3 REST API\n\nAvailable servers:\n- https://api.github.com\n\nAuthentication methods:\n- oauth2: OAuth2 authentication\n- bearer: Bearer token authentication\n\nTotal operations: 912\n\nOperations by category:\n\nrepos:\n- GET /repos/{owner}/{repo} - Get a repository (ID: repos/get)\n- GET /users/{username}/repos - List repositories for a user (ID: repos/listForUser)\n  Requires: 1 path params, 4 query params\n\nTo use an operation, provide the endpoint path and method. The tool will guide you on required parameters."
}
```

### Error Handling

For error responses, the tool provides detailed context:

**Standard Error Response:**
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

**OpenAPI Validation Error:**
```json
{
  "success": false,
  "error_message": "Request validation failed",
  "validation_errors": [
    "name: Required field is missing",
    "species: Value 'elephant' is not in allowed enum: [dog, cat, bird]"
  ],
  "error_guidance": "Fix validation errors:\n1. Add 'name' field (required, string, minLength: 1)\n2. Change 'species' to one of: dog, cat, bird",
  "suggestions": [
    "Add the missing 'name' field",
    "Use a valid species value from the enum"
  ]
}
```

**OpenAPI-Enhanced Error Response (401 Unauthorized):**
```json
{
  "success": false,
  "status_code": 401,
  "error_message": "API returned status 401",
  "error_guidance": "Authentication required. Provide valid credentials using the 'auth' parameter.\n\nThis endpoint requires one of these authentication methods:\n- bearer: Bearer token authentication\n- apiKey: API key in header 'X-API-Key'\n\nProvide credentials using the 'auth' parameter or store them in agent state.",
  "error_details": {
    "message": "Unauthorized"
  }
}
```

**OpenAPI-Enhanced Error Response (400 Bad Request):**
```json
{
  "success": false,
  "status_code": 400,
  "error_message": "API returned status 400",
  "error_guidance": "Bad request. Check that all required parameters are provided and properly formatted.\n\nRequired parameters for this endpoint:\n- owner (path): Repository owner\n- repo (path): Repository name\n- Request body is required\n  Description: The pet object that needs to be added to the store",
  "error_details": {
    "message": "Invalid request format"
  }
}
```

**Error Categories:**
- **4xx Client Errors**: URL, authentication, or request format issues
- **5xx Server Errors**: API server problems (retry recommended)
- **Network Errors**: Connection timeouts or DNS resolution failures
- **Validation Errors**: Invalid parameter values or missing required fields (enhanced with OpenAPI)
- **OpenAPI Errors**: Spec parsing failures or invalid operation references

### LLM Integration

The API Client is designed to work seamlessly with LLM agents:

1. **Descriptive Errors**: Error messages include guidance for LLMs on how to fix issues
2. **Flexible Parameters**: Accepts both structured and string parameters
3. **Context Awareness**: Can infer common patterns and suggest fixes
4. **State Integration**: Stores credentials and base URLs in agent state
5. **OpenAPI Discovery**: LLMs can explore APIs dynamically using OpenAPI specs
6. **Smart Validation**: Provides actionable guidance when requests don't match API schemas
7. **Operation Metadata**: Rich documentation from OpenAPI specs helps LLMs understand APIs
8. **Automatic Configuration**: The tool can automatically:
   - Resolve server URLs from OpenAPI specs (no need to specify base_url)
   - Detect and apply authentication from agent state based on OpenAPI security schemes
   - Provide parameter-specific guidance on validation errors
   - List allowed HTTP methods when a 405 error occurs
   - Show required authentication methods for 401 errors

### OpenAPI Features Deep Dive

#### Automatic Server URL Resolution
When an OpenAPI spec is provided but `base_url` is omitted, the tool will:
1. Parse the OpenAPI spec to find server definitions
2. Use the first server URL as the base URL
3. Handle relative server URLs by extracting the host from the spec URL
4. Emit an `auto_base_url` event for debugging

#### Security Scheme Detection
The tool automatically detects authentication requirements from OpenAPI specs:
1. Checks operation-level security requirements
2. Falls back to global security requirements if not specified
3. Searches agent state for matching credentials:
   - API keys: looks for `{scheme}_api_key`, `api_key`, `apiKey`, or the actual header name
   - Bearer tokens: looks for `{scheme}_token`, `bearer_token`, `access_token`, `token`
   - Basic auth: looks for `{scheme}_username/password`, `api_username/password`, `username/password`
4. Automatically applies found credentials to requests

#### Enhanced Error Guidance
When OpenAPI specs are available, error messages include:
- **400 Bad Request**: Lists all required parameters with descriptions
- **401 Unauthorized**: Shows available authentication methods with details
- **403 Forbidden**: Includes endpoint-specific permission requirements
- **404 Not Found**: Lists required path parameters
- **405 Method Not Allowed**: Shows all allowed HTTP methods for the endpoint

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

### Practical OpenAPI Example

Here's a complete example showing how to use OpenAPI features with an LLM agent:

```go
// Step 1: Store credentials in agent state
agent.State.Set("github_token", os.Getenv("GITHUB_TOKEN"))

// Step 2: Discover available operations
result, _ := agent.ExecuteTool("api_client", map[string]interface{}{
    "openapi_spec": "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json",
    "discover_operations": true,
})

// Step 3: Make an API call (automatic base URL and auth)
result, _ = agent.ExecuteTool("api_client", map[string]interface{}{
    "endpoint": "/repos/lexlapax/go-llms",
    "method": "GET",
    "openapi_spec": "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json",
})

// The tool will:
// 1. Automatically resolve base_url from the OpenAPI spec (https://api.github.com)
// 2. Detect that this endpoint requires authentication
// 3. Find and apply the github_token from agent state as Bearer auth
// 4. Validate the request against the OpenAPI schema
// 5. Provide enhanced error guidance if something goes wrong
```

### Tips for LLM Agents

1. **Start with Discovery**: Always begin by discovering operations to understand what's available
2. **Let OpenAPI Guide You**: The spec provides parameter requirements, so you don't need to guess
3. **Trust Automatic Features**: Let the tool handle server URLs and authentication when possible
4. **Use Validation**: Request validation catches errors before sending, saving API calls
5. **Follow Error Guidance**: The tool provides specific instructions for fixing issues

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