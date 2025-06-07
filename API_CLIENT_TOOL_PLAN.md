# API Client Tool Implementation Plan

## Overview

The API Client Tool is designed to help LLMs effectively interact with various types of APIs. Unlike a generic HTTP client, this tool provides LLM-friendly features like automatic API discovery, schema understanding, authentication handling, and intelligent error recovery.

## Core Philosophy

The tool should graduate from simple to complex capabilities:
1. **Basic REST API** - Simple GET/POST requests with JSON
2. **OpenAPI/Swagger** - Automatic discovery and understanding of API specifications
3. **GraphQL** - Query construction and introspection
4. **Advanced Features** - Rate limiting, authentication, caching, retries

## Key Features for LLM Usage

### 1. API Discovery and Understanding
- **Automatic API Documentation Parsing**: Read OpenAPI/Swagger specs to understand endpoints
- **Schema Inference**: Automatically infer request/response schemas from examples
- **Endpoint Discovery**: Find available endpoints through common patterns (/api, /swagger.json, /.well-known/)
- **GraphQL Introspection**: Query GraphQL schemas to understand available operations

### 2. Intelligent Request Construction
- **Parameter Validation**: Validate parameters against discovered schemas
- **Example-Based Learning**: Learn from successful API calls to improve future requests
- **Smart Defaults**: Provide sensible defaults for common parameters
- **Format Detection**: Automatically detect and handle JSON, XML, Form data, etc.

### 3. Authentication Management
- **Multi-Method Support**: 
  - API Key (header, query, cookie)
  - Bearer Token
  - Basic Auth
  - OAuth 2.0 (client credentials, authorization code)
  - Custom auth schemes
- **Credential Storage**: Secure in-memory storage during session
- **Auth Discovery**: Detect authentication requirements from API responses

### 4. Rate Limiting and Throttling
- **Automatic Detection**: Learn rate limits from headers (X-RateLimit-*, Retry-After)
- **Adaptive Throttling**: Slow down requests as limits approach
- **Quota Management**: Track API quotas and warn before exhaustion
- **Burst Control**: Handle burst limits vs sustained rate limits

### 5. Error Handling and Recovery
- **Intelligent Retry**: Exponential backoff with jitter
- **Error Classification**: Distinguish between client errors, server errors, and rate limits
- **Fallback Strategies**: Try alternative endpoints or degraded functionality
- **Error Translation**: Convert technical errors to LLM-friendly explanations

### 6. Response Processing
- **Pagination Handling**: Automatically fetch all pages of results
- **Data Extraction**: Extract specific fields from nested responses
- **Format Conversion**: Convert between JSON, XML, CSV as needed
- **Response Caching**: Cache responses with intelligent TTL

## Implementation Phases

### Phase 1: Basic REST Client (Day 5, Part 1)
- Simple GET, POST, PUT, DELETE operations
- JSON request/response handling
- Basic authentication (API key, Bearer token)
- Error handling with user-friendly messages

### Phase 2: OpenAPI Integration (Day 5, Part 2)
- Parse OpenAPI/Swagger specifications
- Generate available operations from spec
- Validate requests against schemas
- Provide operation descriptions to LLM

### Phase 3: Advanced Features (Future)
- GraphQL support with introspection
- OAuth 2.0 flows
- Webhook handling
- WebSocket support for real-time APIs

## Tool Interface Design

### Parameters Structure
```go
type APIClientParams struct {
    // Basic Configuration
    BaseURL     string            `json:"base_url,omitempty"`      // Required for direct calls
    SpecURL     string            `json:"spec_url,omitempty"`      // URL to OpenAPI spec
    Operation   string            `json:"operation,omitempty"`     // Operation ID from spec
    
    // Direct Request (when no spec)
    Endpoint    string            `json:"endpoint,omitempty"`      
    Method      string            `json:"method,omitempty"`        
    
    // Request Data
    PathParams  map[string]string `json:"path_params,omitempty"`  
    QueryParams map[string]string `json:"query_params,omitempty"` 
    Headers     map[string]string `json:"headers,omitempty"`      
    Body        interface{}       `json:"body,omitempty"`         
    
    // Authentication
    Auth        AuthConfig        `json:"auth,omitempty"`         
    
    // Advanced Options
    Timeout     string            `json:"timeout,omitempty"`       // "30s", "1m"
    RetryConfig *RetryConfig      `json:"retry,omitempty"`        
    RateLimit   *RateLimitConfig  `json:"rate_limit,omitempty"`   
    Cache       *CacheConfig      `json:"cache,omitempty"`        
}

type AuthConfig struct {
    Type        string            `json:"type"`                    // "api_key", "bearer", "basic", "oauth2"
    // API Key Auth
    APIKey      string            `json:"api_key,omitempty"`      
    KeyLocation string            `json:"key_location,omitempty"`  // "header", "query", "cookie"
    KeyName     string            `json:"key_name,omitempty"`     
    
    // Bearer Token
    Token       string            `json:"token,omitempty"`        
    
    // Basic Auth
    Username    string            `json:"username,omitempty"`     
    Password    string            `json:"password,omitempty"`     
    
    // OAuth2
    ClientID     string           `json:"client_id,omitempty"`    
    ClientSecret string           `json:"client_secret,omitempty"`
    TokenURL     string           `json:"token_url,omitempty"`    
    Scopes       []string         `json:"scopes,omitempty"`       
}
```

### Usage Examples for LLMs

#### Example 1: Simple REST API Call
```json
{
    "base_url": "https://api.github.com",
    "endpoint": "/users/{{username}}",
    "method": "GET",
    "path_params": {
        "username": "octocat"
    },
    "headers": {
        "Accept": "application/vnd.github.v3+json"
    }
}
```

#### Example 2: Using OpenAPI Specification
```json
{
    "spec_url": "https://petstore.swagger.io/v2/swagger.json",
    "operation": "getPetById",
    "path_params": {
        "petId": "123"
    }
}
```

#### Example 3: Authenticated Request with Rate Limiting
```json
{
    "base_url": "https://api.example.com",
    "endpoint": "/v1/data",
    "method": "POST",
    "auth": {
        "type": "bearer",
        "token": "{{api_token}}"
    },
    "body": {
        "query": "machine learning"
    },
    "rate_limit": {
        "requests_per_second": 10,
        "burst": 20
    }
}
```

## Error Guidance for LLMs

The tool will provide specific guidance for common errors:

- **401 Unauthorized**: "The API requires authentication. Please provide credentials using the 'auth' parameter."
- **404 Not Found**: "The endpoint '{{endpoint}}' was not found. Check the API documentation or try discovering available endpoints."
- **429 Too Many Requests**: "Rate limit exceeded. The tool will automatically retry after {{retry_after}} seconds."
- **Invalid Schema**: "The request body doesn't match the expected schema. Required fields: {{required_fields}}"

## Success Metrics

1. **Ease of Use**: LLM can successfully call APIs without detailed knowledge of HTTP
2. **Error Recovery**: 90% of recoverable errors handled automatically
3. **Discovery Success**: Can discover and use APIs from just a base URL
4. **Performance**: Efficient caching and rate limiting to minimize API calls

## Future Enhancements

1. **API Monitoring**: Track API health and availability
2. **Cost Tracking**: Monitor API usage costs
3. **Mock Mode**: Test API interactions without making real calls
4. **Batch Operations**: Execute multiple API calls efficiently
5. **Webhook Server**: Receive and process webhook callbacks
6. **API Composition**: Chain multiple API calls together