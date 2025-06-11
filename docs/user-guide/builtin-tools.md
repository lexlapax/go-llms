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

The `api_client` tool provides a high-level interface for making REST API and GraphQL calls with intelligent error handling, authentication support, OpenAPI specification discovery, and GraphQL introspection.

### Overview

**Tool Name**: `api_client`  
**Category**: `web`  
**Version**: `3.0.0` (Updated January 8, 2025 with GraphQL support)

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

5. **GraphQL Support** (Added in Phase 3 - v3.0.0)
   - GraphQL query and mutation execution
   - Variable support for parameterized queries
   - Schema introspection and discovery
   - Operation-specific error handling
   - GraphQL-aware response formatting
   - Caching of GraphQL schemas and discovery results
   - LLM-friendly operation discovery
   - **New in v3.0.0** (January 8, 2025):
     - Full GraphQL query/mutation support
     - Schema introspection with caching
     - Variable type validation
     - GraphQL-specific error guidance

### Planned Features (Future Phases)

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

**GraphQL Parameters (v3.0.0):**
- `graphql_query` (string): GraphQL query or mutation string. When provided, the tool operates in GraphQL mode
- `graphql_variables` (object): Variables for the GraphQL query (e.g., {'userId': '123', 'limit': 10})
- `graphql_operation_name` (string): Name of the operation to execute when query contains multiple operations
- `discover_graphql` (boolean): If true, performs introspection to discover available queries, mutations, and types
- `max_graphql_depth` (integer): Maximum depth for GraphQL queries (default: 5, max: 10)

### Authentication Configuration

The tool features unified authentication middleware with two approaches:

#### Automatic Authentication (Recommended)

Store credentials in agent state for automatic detection and application:

```python
# In your agent code:
state.Set("github_token", "ghp_xxx")      # For GitHub APIs
state.Set("api_key", "key_xxx")           # For generic APIs
state.Set("api_username", "user")         # For basic auth
state.Set("api_password", "pass")         # For basic auth
```

The tool automatically detects and applies appropriate authentication based on:
- URL patterns (e.g., github.com, gitlab.com)
- OpenAPI security schemes
- Generic patterns in state

#### Manual Authentication (When Needed)

The `auth` parameter supports multiple authentication types:

**API Key Authentication:**
```json
{
  "type": "api_key",
  "api_key": "your-api-key",
  "key_location": "header",  // or "query" or "cookie"
  "key_name": "X-API-Key"
}
```

**Bearer Token Authentication:**
```json
{
  "type": "bearer",
  "token": "your-bearer-token"
}
```

**Basic Authentication:**
```json
{
  "type": "basic",
  "username": "your-username",
  "password": "your-password"
}
```

**OAuth2 Authentication:**
```json
{
  "type": "oauth2",
  "access_token": "your-access-token"
}
```

**Custom Header Authentication:**
```json
{
  "type": "custom",
  "header_name": "X-Custom-Auth",
  "header_value": "your-secret",
  "prefix": "Token"  // Optional prefix (e.g., "Bearer", "Token")
}
```

**Security Note:** For production use, always prefer storing credentials in agent state to prevent exposing them to LLMs.

#### Session/Cookie Management

Enable session management to maintain state across requests:

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/login",
  "method": "POST",
  "enable_session": true,
  "body": {
    "username": "user",
    "password": "pass"
  }
}
```

Subsequent requests with `enable_session: true` will automatically include cookies from previous responses.

#### OAuth2 Configuration

For OAuth2 token exchange, provide configuration in the `oauth2_config` parameter:

```json
{
  "oauth2_config": {
    "token_url": "https://auth.example.com/oauth/token",
    "auth_url": "https://auth.example.com/oauth/authorize",
    "client_id": "your-client-id",
    "client_secret": "your-client-secret",
    "redirect_uri": "http://localhost:8080/callback",
    "scope": "read write"
  }
}
```

The tool supports:
- Client credentials flow (server-to-server authentication)
- Authorization code flow (user authentication)
- Automatic token refresh using refresh tokens
- JWT token expiry detection

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

The tool features unified authentication middleware that automatically detects and applies credentials from agent state:

**How it works:**
1. Store credentials in agent state (e.g., `state.Set("github_token", "ghp_xxx")`)
2. The tool detects credentials based on:
   - URL patterns (e.g., GitHub, GitLab)
   - OpenAPI security schemes
   - Generic auth patterns in state
3. Authentication is applied transparently - LLMs never see credentials

**Example:**
```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/user/repos",
  "method": "GET"
}
```

If `github_token` or `github_api_key` exists in agent state, it will be automatically applied as a Bearer token.

**Supported auth types:**
- API Key (header or query parameter)
- Bearer Token
- Basic Authentication
- OAuth2 (placeholder for future implementation)

**Security best practice:** Never pass credentials in tool parameters. Always store them in agent state.

#### 11. GraphQL Discovery

Discover available GraphQL operations through introspection:

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/graphql",
  "discover_graphql": true
}
```

#### 12. Simple GraphQL Query

Execute a basic GraphQL query (authentication automatically applied from state):

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/graphql",
  "graphql_query": "query { viewer { login name email bio } }"
}
```

#### 13. GraphQL Query with Variables

Use variables for dynamic GraphQL queries:

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/graphql",
  "graphql_query": "query GetRepo($owner: String!, $name: String!) { repository(owner: $owner, name: $name) { name description stargazerCount } }",
  "graphql_variables": {
    "owner": "golang",
    "name": "go"
  }
}
```

#### 14. GraphQL Mutation

Execute a GraphQL mutation:

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/graphql",
  "graphql_query": "mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id name email } }",
  "graphql_variables": {
    "input": {
      "name": "John Doe",
      "email": "john@example.com"
    }
  },
  "auth": {
    "type": "api_key",
    "api_key": "your-api-key",
    "key_location": "header",
    "key_name": "X-API-Key"
  }
}
```

#### 15. OAuth2 Authentication

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/user/profile",
  "method": "GET",
  "auth": {
    "type": "oauth2",
    "access_token": "your-oauth2-access-token"
  }
}
```

#### 16. Custom Header Authentication

```json
{
  "base_url": "https://api.custom.com",
  "endpoint": "/data",
  "method": "GET",
  "auth": {
    "type": "custom",
    "header_name": "X-Service-Token",
    "header_value": "secret123",
    "prefix": "Bearer"
  }
}
```

#### 17. API Key in Query Parameter

```json
{
  "base_url": "https://api.weather.com",
  "endpoint": "/current",
  "method": "GET",
  "query_params": {
    "city": "London"
  },
  "auth": {
    "type": "api_key",
    "api_key": "your-weather-key",
    "key_location": "query",
    "key_name": "apikey"
  }
}
```

#### 18. Session Management Example

```json
{
  "base_url": "https://api.sessiondemo.com",
  "endpoint": "/auth/login",
  "method": "POST",
  "enable_session": true,
  "body": {
    "username": "demo_user",
    "password": "demo_pass"
  }
}
```

Follow-up request using session:
```json
{
  "base_url": "https://api.sessiondemo.com",
  "endpoint": "/user/profile",
  "method": "GET",
  "enable_session": true
}
```

#### 19. OAuth2 Client Credentials Flow

Store OAuth2 config in state:
```python
state.Set("oauth2_config", {
  "token_url": "https://auth.example.com/oauth/token",
  "client_id": "your-client-id",
  "client_secret": "your-client-secret",
  "flow": "client_credentials",
  "scope": "read write"
})
```

The tool will automatically exchange credentials for an access token when needed.

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

**GraphQL Query Response:**
```json
{
  "success": true,
  "status_code": 200,
  "data": {
    "viewer": {
      "login": "octocat",
      "name": "The Octocat",
      "email": "octocat@github.com",
      "bio": "GitHub's mascot"
    }
  }
}
```

**GraphQL Discovery Response:**
```json
{
  "success": true,
  "status_code": 200,
  "graphql_schema": {
    "endpoint": "https://api.github.com/graphql",
    "operations": {
      "queries": [
        {
          "name": "Query",
          "description": "Root query type",
          "example": "query { ... }"
        }
      ],
      "mutations": []
    },
    "types": {
      "User": {
        "kind": "OBJECT",
        "description": "A user is an individual's account on GitHub",
        "fields": []
      },
      "Repository": {
        "kind": "OBJECT",
        "description": "A repository contains the content for a project",
        "fields": []
      }
    }
  }
}
```

**GraphQL Error Response:**
```json
{
  "success": false,
  "status_code": 200,
  "data": null,
  "error_message": "Field 'invalidField' doesn't exist on type 'User'",
  "error_details": [
    {
      "message": "Field 'invalidField' doesn't exist on type 'User'",
      "path": ["viewer", "invalidField"],
      "extensions": {
        "code": "FIELD_NOT_FOUND"
      }
    }
  ],
  "error_guidance": "GraphQL query returned errors. Check the error_details for specific field errors",
  "graphql_extensions": {
    "requestId": "abc123"
  }
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
8. **GraphQL Support**: LLMs can explore and execute GraphQL queries with:
   - Schema introspection for discovering available operations
   - Variable type validation and guidance
   - GraphQL-specific error messages with field paths
   - Automatic query structure validation
9. **Automatic Configuration**: The tool can automatically:
   - Resolve server URLs from OpenAPI specs (no need to specify base_url)
   - Detect and apply authentication from agent state based on OpenAPI security schemes
   - Provide parameter-specific guidance on validation errors
   - List allowed HTTP methods when a 405 error occurs
   - Show required authentication methods for 401 errors
   - Cache GraphQL schemas for improved performance

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
agent.State.Set("github_token", os.Getenv("GITHUB_API_KEY"))

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

### GraphQL Features Deep Dive

#### Schema Introspection
The tool supports GraphQL schema discovery through introspection:
1. Automatically queries the GraphQL endpoint for schema information
2. Caches schemas for 15 minutes to reduce API calls
3. Returns simplified operation listings for LLM consumption
4. Identifies available queries, mutations, and custom types

#### Query Validation
GraphQL queries are validated before execution:
1. Syntax validation using gqlparser
2. Variable type checking
3. Query depth limiting (default: 5, max: 10)
4. Fragment support (inline fragments currently supported)

#### Error Handling
GraphQL-specific error handling provides:
1. Field-level error messages with paths
2. Type mismatch guidance
3. Authentication error detection
4. Query complexity warnings

### Practical GraphQL Example

Here's a complete example showing GraphQL features with an LLM agent:

```go
// Store GitHub token in agent state
state := agent.GetState()
state.Set("github_token", os.Getenv("GITHUB_API_KEY"))

// Step 1: Discover GraphQL schema (auth applied automatically)
result, _ := agent.ExecuteTool("api_client", map[string]interface{}{
    "base_url": "https://api.github.com",
    "endpoint": "/graphql",
    "discover_graphql": true,
})

// Step 2: Execute a GraphQL query with variables
result, _ = agent.ExecuteTool("api_client", map[string]interface{}{
    "base_url": "https://api.github.com",
    "endpoint": "/graphql",
    "graphql_query": `
        query GetRepo($owner: String!, $name: String!) {
            repository(owner: $owner, name: $name) {
                name
                description
                stargazerCount
                issues(first: 5, states: OPEN) {
                    nodes {
                        title
                        number
                    }
                }
            }
        }
    `,
    "graphql_variables": map[string]interface{}{
        "owner": "golang",
        "name": "go",
    },
})
```

### Tips for LLM Agents

1. **Start with Discovery**: Always begin by discovering operations to understand what's available
2. **Let OpenAPI Guide You**: The spec provides parameter requirements, so you don't need to guess
3. **Trust Automatic Features**: Let the tool handle server URLs and authentication when possible
4. **Use Validation**: Request validation catches errors before sending, saving API calls
5. **Follow Error Guidance**: The tool provides specific instructions for fixing issues
6. **GraphQL Best Practices**:
   - Use introspection to understand the schema before querying
   - Include only necessary fields to reduce response size
   - Use variables for dynamic values instead of string concatenation
   - Handle GraphQL errors separately from HTTP errors
   - Cache discovery results when exploring the same API multiple times

## Web Tools

### Web Search Tool

**Tool Name**: `web_search`  
**Version**: `2.0.0`

Search the web using multiple search engines (DuckDuckGo, Brave, Tavily, Serpapi, Serper.dev).

**Key Features:**
- Multiple search engine support with automatic fallback
- Configurable result count and safe search
- API key support for premium engines
- Automatic result deduplication

**Parameters:**
- `query` (string, required): Search query
- `max_results` (integer): Maximum results (default: 10)
- `engine` (string): Search engine to use
- `safe_search` (string): Safe search level (strict/moderate/off)
- `api_key` (string): API key for premium engines

### Web Fetch Tool

**Tool Name**: `web_fetch`  
**Version**: `2.0.0`

Fetch and extract content from web pages.

**Key Features:**
- Clean text extraction from HTML
- Metadata extraction (title, description, author)
- Custom user agent support
- Timeout control
- Following redirects

**Parameters:**
- `url` (string, required): URL to fetch
- `timeout` (integer): Timeout in seconds (default: 30)
- `user_agent` (string): Custom user agent
- `follow_redirects` (boolean): Follow HTTP redirects (default: true)
- `extract_metadata` (boolean): Extract page metadata (default: true)

### Web Scrape Tool

**Tool Name**: `web_scrape`  
**Version**: `2.0.0`

Extract structured data from web pages using CSS selectors.

**Key Features:**
- CSS selector support for precise data extraction
- Multiple selector support
- Attribute extraction
- Table extraction to JSON
- Custom extraction patterns

**Parameters:**
- `url` (string, required): URL to scrape
- `selectors` (map): CSS selectors for extraction
- `extract_tables` (boolean): Extract tables as JSON
- `timeout` (integer): Timeout in seconds
- `user_agent` (string): Custom user agent

### HTTP Request Tool

**Tool Name**: `http_request`  
**Version**: `2.0.0`

Make raw HTTP requests with full control.

**Key Features:**
- All HTTP methods support
- Custom headers and body
- Response header access
- Binary and text response handling
- Detailed timing information

**Parameters:**
- `url` (string, required): Target URL
- `method` (string): HTTP method (default: GET)
- `headers` (map): Custom headers
- `body` (string): Request body
- `timeout` (integer): Timeout in seconds
- `follow_redirects` (boolean): Follow redirects

## File Tools

### File Read Tool

**Tool Name**: `file_read`  
**Version**: `2.0.0`

Read files with advanced options for large files and specific ranges.

**Key Features:**
- Large file support with streaming
- Line range selection
- Binary file detection
- Encoding detection
- Metadata extraction

**Parameters:**
- `path` (string, required): File path to read
- `start_line` (integer): Starting line number
- `end_line` (integer): Ending line number
- `encoding` (string): File encoding (default: UTF-8)
- `binary_mode` (boolean): Read as binary

### File Write Tool

**Tool Name**: `file_write`  
**Version**: `2.0.0`

Write files with safety features and atomic operations.

**Key Features:**
- Atomic writes with temporary files
- Automatic backup creation
- Append mode support
- Directory creation
- Permission preservation

**Parameters:**
- `path` (string, required): File path to write
- `content` (string, required): Content to write
- `mode` (string): Write mode (overwrite/append)
- `create_dirs` (boolean): Create parent directories
- `backup` (boolean): Create backup before overwriting

### File List Tool

**Tool Name**: `file_list`  
**Version**: `2.0.0`

List files and directories with filtering options.

**Key Features:**
- Pattern matching (glob)
- Size and date filtering
- Recursive listing
- Hidden file inclusion
- Detailed file information

**Parameters:**
- `path` (string, required): Directory path
- `pattern` (string): File pattern (e.g., *.txt)
- `recursive` (boolean): Include subdirectories
- `include_hidden` (boolean): Include hidden files
- `min_size` (integer): Minimum file size
- `max_size` (integer): Maximum file size

### File Delete Tool

**Tool Name**: `file_delete`  
**Version**: `2.0.0`

Safely delete files with confirmation options.

**Key Features:**
- Safe deletion with checks
- Pattern-based deletion
- Recursive directory deletion
- Dry run mode
- Deletion logging

**Parameters:**
- `path` (string, required): File or directory path
- `pattern` (string): Pattern for multiple files
- `recursive` (boolean): Delete directories recursively
- `dry_run` (boolean): Preview without deleting
- `force` (boolean): Skip safety checks

### File Move Tool

**Tool Name**: `file_move`  
**Version**: `2.0.0`

Move or rename files and directories.

**Key Features:**
- Atomic move operations
- Cross-filesystem support
- Overwrite protection
- Directory moving
- Batch operations

**Parameters:**
- `source` (string, required): Source path
- `destination` (string, required): Destination path
- `overwrite` (boolean): Overwrite if exists
- `create_dirs` (boolean): Create parent directories

### File Search Tool

**Tool Name**: `file_search`  
**Version**: `2.0.0`

Search for content within files.

**Key Features:**
- Regular expression support
- Multi-file search
- Context lines
- Binary file skipping
- Performance optimization

**Parameters:**
- `path` (string, required): Search path
- `pattern` (string, required): Search pattern (regex)
- `file_pattern` (string): File name pattern
- `recursive` (boolean): Search subdirectories
- `case_sensitive` (boolean): Case sensitive search
- `context_lines` (integer): Lines of context

## System Tools

### Execute Command Tool

**Tool Name**: `execute_command`  
**Version**: `2.0.0`

Execute system commands safely with timeouts and constraints.

**Key Features:**
- Command timeout support
- Working directory control
- Environment variable setting
- Output capture (stdout/stderr)
- Safe command validation

**Parameters:**
- `command` (string, required): Command to execute
- `args` (array): Command arguments
- `working_dir` (string): Working directory
- `env` (map): Environment variables
- `timeout` (integer): Timeout in seconds
- `shell` (boolean): Use shell execution

### Get Environment Variable Tool

**Tool Name**: `get_environment_variable`  
**Version**: `2.0.0`

Read environment variables with pattern matching.

**Key Features:**
- Single variable reading
- Pattern matching for multiple variables
- Default value support
- Variable expansion
- Security filtering

**Parameters:**
- `name` (string): Variable name
- `pattern` (string): Pattern for multiple variables
- `default` (string): Default value if not found
- `expand` (boolean): Expand variable references

### Get System Info Tool

**Tool Name**: `get_system_info`  
**Version**: `2.0.0`

Gather comprehensive system information.

**Key Features:**
- OS and architecture details
- Memory and CPU information
- Disk usage statistics
- Network interface data
- Runtime environment info

**Parameters:**
- `include_env` (boolean): Include environment variables
- `include_network` (boolean): Include network info
- `include_disk` (boolean): Include disk usage
- `include_memory` (boolean): Include memory stats

### Process List Tool

**Tool Name**: `process_list`  
**Version**: `2.0.0`

List and filter system processes.

**Key Features:**
- Process filtering by name/PID
- CPU and memory usage
- Process tree display
- Sorting options
- Command line arguments

**Parameters:**
- `filter` (string): Filter by process name
- `sort_by` (string): Sort field (cpu/memory/pid)
- `limit` (integer): Maximum processes to return
- `include_threads` (boolean): Include thread count
- `include_children` (boolean): Include child processes

## Data Tools

### JSON Process Tool

**Tool Name**: `json_process`  
**Version**: `2.0.0`

Process JSON data with JSONPath queries and transformations.

**Key Features:**
- JSONPath query support
- JSON validation and formatting
- Flattening and unflattening
- Key/value extraction
- Schema validation

**Parameters:**
- `data` (string, required): JSON data to process
- `operation` (string, required): Operation type
- `query` (string): JSONPath query
- `indent` (integer): Pretty print indentation
- `sort_keys` (boolean): Sort object keys

**Operations:**
- `parse`: Validate and parse JSON
- `query`: Execute JSONPath query
- `flatten`: Flatten nested structure
- `prettify`: Format with indentation
- `minify`: Remove whitespace
- `extract_keys`: Get all keys
- `extract_values`: Get all values

### CSV Process Tool

**Tool Name**: `csv_process`  
**Version**: `2.0.0`

Process CSV data with filtering and statistics.

**Key Features:**
- CSV parsing with custom delimiters
- Column filtering and transformation
- Statistical calculations
- Data type inference
- Header handling

**Parameters:**
- `data` (string, required): CSV data
- `operation` (string, required): Operation type
- `delimiter` (string): Field delimiter
- `headers` (boolean): First row is headers
- `columns` (array): Columns to process

**Operations:**
- `parse`: Parse CSV to JSON
- `filter`: Filter rows by conditions
- `transform`: Transform column values
- `stats`: Calculate statistics

### XML Process Tool

**Tool Name**: `xml_process`  
**Version**: `2.0.0`

Process XML data with XPath queries.

**Key Features:**
- XPath 1.0 query support
- XML to JSON conversion
- Namespace handling
- Pretty printing
- Validation

**Parameters:**
- `data` (string, required): XML data
- `operation` (string, required): Operation type
- `xpath` (string): XPath query
- `namespaces` (map): Namespace mappings

**Operations:**
- `parse`: Validate XML
- `query`: Execute XPath query
- `to_json`: Convert to JSON

### Data Transform Tool

**Tool Name**: `data_transform`  
**Version**: `2.0.0`

Apply functional transformations to data collections.

**Key Features:**
- Functional operations (map, filter, reduce)
- Sorting and grouping
- Aggregations
- Type conversions
- Unique value extraction

**Parameters:**
- `data` (array, required): Input data array
- `operation` (string, required): Transform operation
- `expression` (string): Transformation expression
- `field` (string): Field for operations
- `ascending` (boolean): Sort order

**Operations:**
- `filter`: Filter by condition
- `map`: Transform each element
- `reduce`: Aggregate to single value
- `sort`: Sort by field or value
- `group_by`: Group by field value
- `unique`: Extract unique values
- `reverse`: Reverse order

## DateTime Tools

### DateTime Now Tool

**Tool Name**: `datetime_now`  
**Version**: `2.0.0`

Get current date and time in any timezone.

**Key Features:**
- Multiple timezone support
- Various output formats
- Unix timestamp
- ISO 8601 formatting
- Relative time descriptions

**Parameters:**
- `timezone` (string): Target timezone
- `format` (string): Output format
- `unix` (boolean): Return Unix timestamp
- `iso` (boolean): Return ISO 8601 format

### DateTime Parse Tool

**Tool Name**: `datetime_parse`  
**Version**: `2.0.0`

Parse dates from various formats including natural language.

**Key Features:**
- Auto-format detection
- Natural language parsing ("next Monday")
- Multiple format support
- Timezone handling
- Relative date parsing

**Parameters:**
- `date_string` (string, required): Date to parse
- `format` (string): Expected format
- `timezone` (string): Source timezone
- `strict` (boolean): Strict parsing mode

### DateTime Calculate Tool

**Tool Name**: `datetime_calculate`  
**Version**: `2.0.0`

Perform date arithmetic and calculations.

**Key Features:**
- Add/subtract time units
- Business day calculations
- Duration calculations
- Date differences
- Holiday awareness

**Parameters:**
- `date` (string): Base date (default: now)
- `operation` (string, required): Calculation type
- `value` (integer): Amount to add/subtract
- `unit` (string): Time unit
- `business_days` (boolean): Use business days only

**Operations:**
- `add`: Add time to date
- `subtract`: Subtract time from date
- `difference`: Calculate time between dates
- `business_days`: Count business days

### DateTime Format Tool

**Tool Name**: `datetime_format`  
**Version**: `2.0.0`

Format dates with localization support.

**Key Features:**
- Standard format patterns
- Custom formats
- Localization (50+ locales)
- Relative formatting
- Multiple output formats

**Parameters:**
- `date` (string, required): Date to format
- `format` (string): Output format
- `locale` (string): Locale for formatting
- `timezone` (string): Display timezone

### DateTime Compare Tool

**Tool Name**: `datetime_compare`  
**Version**: `2.0.0`

Compare dates and check date ranges.

**Key Features:**
- Date comparisons
- Range checking
- Multiple date sorting
- Relative comparisons
- Business day aware

**Parameters:**
- `date1` (string, required): First date
- `date2` (string, required): Second date
- `operation` (string): Comparison type
- `dates` (array): Multiple dates for sorting

**Operations:**
- `before`: Check if date1 is before date2
- `after`: Check if date1 is after date2
- `equal`: Check if dates are equal
- `between`: Check if date is in range
- `earliest`: Find earliest date
- `latest`: Find latest date

### DateTime Info Tool

**Tool Name**: `datetime_info`  
**Version**: `2.0.0`

Get detailed information about dates.

**Key Features:**
- Day of week/year
- Week number
- Quarter information
- Leap year detection
- Days in month
- Zodiac signs

**Parameters:**
- `date` (string, required): Date to analyze
- `timezone` (string): Timezone for calculations
- `include_all` (boolean): Return all information

### DateTime Convert Tool

**Tool Name**: `datetime_convert`  
**Version**: `2.0.0`

Convert dates between timezones.

**Key Features:**
- Timezone conversion
- DST handling
- Offset calculations
- Multiple timezone support
- UTC conversions

**Parameters:**
- `date` (string, required): Date to convert
- `from_timezone` (string): Source timezone
- `to_timezone` (string, required): Target timezone
- `format` (string): Output format

## Feed Tools

### Feed Fetch Tool

**Tool Name**: `feed_fetch`  
**Version**: `2.0.0`

Fetch and parse RSS, Atom, and JSON feeds.

**Key Features:**
- Multi-format support (RSS 2.0, Atom, JSON Feed)
- Conditional fetching with ETags
- Custom user agents
- Item limiting
- Metadata extraction

**Parameters:**
- `url` (string, required): Feed URL
- `max_items` (integer): Maximum items to return
- `timeout` (integer): Timeout in seconds
- `user_agent` (string): Custom user agent
- `if_modified_since` (string): Conditional fetch
- `etag` (string): ETag for caching

### Feed Discover Tool

**Tool Name**: `feed_discover`  
**Version**: `2.0.0`

Auto-discover feed URLs from websites.

**Key Features:**
- Auto-discovery from HTML
- Common feed path checking
- Link following
- Multiple format detection
- Podcast feed discovery

**Parameters:**
- `url` (string, required): Website URL
- `follow_links` (boolean): Follow page links
- `max_depth` (integer): Link following depth
- `include_podcasts` (boolean): Include podcast feeds

### Feed Filter Tool

**Tool Name**: `feed_filter`  
**Version**: `2.0.0`

Filter feed items by various criteria.

**Key Features:**
- Keyword filtering
- Date range filtering
- Author filtering
- Category filtering
- Custom matching modes

**Parameters:**
- `feed` (object, required): Feed to filter
- `keywords` (array): Keywords to match
- `after` (string): Date after which to include
- `before` (string): Date before which to include
- `authors` (array): Authors to include
- `categories` (array): Categories to match
- `match_all` (boolean): Require all criteria

### Feed Aggregate Tool

**Tool Name**: `feed_aggregate`  
**Version**: `2.0.0`

Combine multiple feeds into one.

**Key Features:**
- Multi-feed merging
- Duplicate removal
- Sorting options
- Item limiting
- Metadata preservation

**Parameters:**
- `feeds` (array, required): Feeds to aggregate
- `title` (string): Aggregated feed title
- `description` (string): Feed description
- `deduplicate` (boolean): Remove duplicates
- `sort_by` (string): Sort field (date/title)
- `max_items` (integer): Maximum total items

### Feed Convert Tool

**Tool Name**: `feed_convert`  
**Version**: `2.0.0`

Convert between feed formats.

**Key Features:**
- Format conversion (RSS ↔ Atom ↔ JSON)
- Pretty printing
- Metadata inclusion
- Validation
- Custom templates

**Parameters:**
- `feed` (object, required): Feed to convert
- `target_type` (string, required): Target format
- `pretty_print` (boolean): Format output
- `include_meta` (boolean): Include metadata
- `validate` (boolean): Validate output

### Feed Extract Tool

**Tool Name**: `feed_extract`  
**Version**: `2.0.0`

Extract specific data from feed items.

**Key Features:**
- Field extraction
- Data flattening
- Custom field mapping
- Metadata extraction
- Export formatting

**Parameters:**
- `feed` (object, required): Feed to process
- `fields` (array, required): Fields to extract
- `max_items` (integer): Maximum items
- `flatten` (boolean): Flatten nested data
- `include_meta` (boolean): Include metadata

## Tool Discovery and Usage

All built-in tools are registered in the global tool registry and can be discovered programmatically:

```go
import "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"

// List all tools
allTools := tools.Tools.List()

// Get tools by category
webTools := tools.Tools.ListByCategory("web")

// Search for tools
results := tools.Tools.Search("json")

// Get a specific tool
tool, err := tools.GetTool("web_search")
```

## Version History

### v2.0.0 (January 2025)
- Migrated all 32 tools to ToolBuilder pattern
- Added comprehensive metadata for LLM guidance
- Enhanced error messages and examples
- Added MCP (Model Context Protocol) compatibility
- Improved authentication support across web tools

### v1.0.0 (December 2024)
- Initial release with 26 core tools
- Basic tool functionality
- Simple parameter validation