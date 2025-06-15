# Authentication System

> **[Documentation Home](/docs/README.md) / [Technical Documentation](/docs/technical/README.md) / Authentication**

This document describes the authentication system in go-llms, particularly how it integrates with the API Client tool and web tools.

## Overview

The authentication system (`pkg/util/auth/`) provides a unified way to handle various authentication methods across HTTP requests. It's designed to keep credentials secure by storing them in agent state rather than passing them through LLM prompts.

## Key Components

### AuthConfig

```go
type AuthConfig struct {
    Type string                 `json:"type"` // "api_key", "bearer", "basic", "oauth2", "custom"
    Data map[string]interface{} `json:"data"` // Auth-specific data
}
```

### Authentication Detection

The system uses a flexible token detection approach that works with any URL (including test servers):

```go
func DetectAuthFromState(state StateReader, baseURL string, schemes map[string]AuthScheme) *AuthConfig {
    // Try generic auth detection first, which includes provider-specific tokens
    if auth := detectGenericAuthWithProviderTokens(state); auth != nil {
        return auth
    }
    
    // If schemes are provided, try to match against them
    if len(schemes) > 0 {
        return detectFromSchemes(state, schemes)
    }
    
    // Fall back to basic generic auth detection
    return detectGenericAuth(state)
}
```

## Authentication Methods

### 1. API Key Authentication

Supports API keys in headers, query parameters, or cookies:

```go
auth := map[string]interface{}{
    "type": "api_key",
    "api_key": "your-key",
    "key_location": "header",  // or "query", "cookie"
    "key_name": "X-API-Key",
}
```

### 2. Bearer Token Authentication

Standard bearer token in Authorization header:

```go
auth := map[string]interface{}{
    "type": "bearer",
    "token": "your-token",
}
```

### 3. Basic Authentication

HTTP Basic authentication:

```go
auth := map[string]interface{}{
    "type": "basic",
    "username": "user",
    "password": "pass",
}
```

### 4. OAuth2 Authentication

Supports access tokens with future support for full OAuth2 flows:

```go
auth := map[string]interface{}{
    "type": "oauth2",
    "access_token": "your-access-token",
}
```

### 5. Custom Header Authentication

For APIs with non-standard authentication headers:

```go
auth := map[string]interface{}{
    "type": "custom",
    "header_name": "X-Custom-Auth",
    "header_value": "your-value",
    "prefix": "Token",  // Optional prefix
}
```

## Token Detection from State

The system automatically detects authentication tokens from agent state using a comprehensive list of common patterns:

### Provider-Specific Tokens

```go
// GitHub tokens
"github_token", "github_api_key", "GITHUB_TOKEN", "GITHUB_API_KEY",
"github_personal_access_token", "github_pat", "gh_token"

// GitLab tokens
"gitlab_token", "gitlab_api_key", "GITLAB_TOKEN", "GITLAB_API_KEY",
"gitlab_personal_access_token", "gitlab_pat"

// Generic tokens
"api_token", "access_token", "bearer_token", "auth_token", "token",
"API_TOKEN", "ACCESS_TOKEN", "BEARER_TOKEN", "AUTH_TOKEN", "TOKEN"
```

### Generic API Keys

```go
"api_key", "apikey", "x_api_key", "X_API_KEY"
```

## OpenAPI Integration

When OpenAPI schemas are available, the system can:

1. Detect security requirements from the schema
2. Match schema-defined authentication methods with state values
3. Provide detailed error messages about missing authentication

### Schema-Based Detection

```go
func detectFromSchemes(state StateReader, schemes map[string]AuthScheme) *AuthConfig {
    for schemeName, scheme := range schemes {
        switch scheme.Type {
        case "apiKey":
            if auth := detectAPIKeyFromState(state, schemeName, scheme); auth != nil {
                return auth
            }
        case "http":
            if auth := detectHTTPAuthFromState(state, schemeName, scheme); auth != nil {
                return auth
            }
        }
    }
    return nil
}
```

## Best Practices

### 1. Store Credentials in State

Always store credentials in agent state rather than passing them in tool parameters:

```go
// Good
state.Set("github_token", "ghp_xxx")

// Avoid
params["auth"] = map[string]interface{}{
    "type": "bearer",
    "token": "ghp_xxx",  // Exposed to LLM
}
```

### 2. Use Environment Variables

Load credentials from environment variables:

```go
state.Set("github_token", os.Getenv("GITHUB_TOKEN"))
state.Set("api_key", os.Getenv("API_KEY"))
```

### 3. Provider-Agnostic Design

The authentication system works with any URL, not just hardcoded provider patterns. This ensures:
- Test servers work correctly
- Self-hosted instances are supported
- Custom APIs can use standard authentication

## Recent Changes (January 2025)

### Removal of URL-Specific Detection

Previously, the system used hardcoded URL patterns to detect authentication requirements:

```go
// OLD APPROACH - REMOVED
func detectURLSpecificAuth(normalizedURL string, state StateReader) *AuthConfig {
    if strings.Contains(normalizedURL, "github.com") {
        // GitHub-specific detection
    }
    if strings.Contains(normalizedURL, "gitlab.com") {
        // GitLab-specific detection
    }
}
```

This approach had limitations:
- Didn't work with test servers
- Didn't support self-hosted instances
- Required code changes to add new providers

### Current Approach

The new approach uses generic token detection that works with any URL:

```go
func detectGenericAuthWithProviderTokens(state StateReader) *AuthConfig {
    // Try all known token patterns regardless of URL
    tokenKeys := []string{
        // GitHub tokens
        "github_token", "github_api_key", ...
        // GitLab tokens
        "gitlab_token", "gitlab_api_key", ...
        // Generic tokens
        "api_token", "access_token", ...
    }
    
    for _, key := range tokenKeys {
        if value, exists := state.Get(key); exists {
            if token, ok := value.(string); ok && token != "" {
                return &AuthConfig{
                    Type: "bearer",
                    Data: map[string]interface{}{
                        "token": token,
                    },
                }
            }
        }
    }
    return nil
}
```

Benefits:
- Works with any URL including test servers
- No hardcoded provider patterns
- Automatically supports new providers if they follow common naming conventions

## Future Enhancements

A provider registry pattern is planned (see TODO.md) that will allow:
- Configuration-driven provider definitions
- Custom authentication patterns via YAML/JSON
- Response-based authentication detection (401 + WWW-Authenticate)
- OAuth2 discovery via .well-known endpoints
- Extensibility without code changes

## Integration with Tools

### API Client Tool

The API Client tool (`pkg/agent/builtins/tools/web/api_client.go`) integrates seamlessly with the authentication system:

1. Checks for explicit auth parameter
2. Falls back to state-based detection
3. Applies authentication to requests
4. Provides helpful error messages for missing auth

### Example Usage

```go
// Store credentials in state
state.Set("github_token", os.Getenv("GITHUB_TOKEN"))

// Make authenticated request (no auth parameter needed)
result, err := tool.Execute(ctx, map[string]interface{}{
    "base_url": "https://api.github.com",
    "endpoint": "/user/repos",
    "method": "GET",
})
```

## Testing

The authentication system has comprehensive tests covering:
- All authentication types
- Token detection from state
- OpenAPI schema integration
- Error handling
- Security (no credential leakage)

See `pkg/util/auth/auth_test.go` for test coverage.