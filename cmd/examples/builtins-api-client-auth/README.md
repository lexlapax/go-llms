# API Client Advanced Authentication Example

This example demonstrates the advanced authentication features of the `api_client` tool in the go-llms library.

## Features Demonstrated

### 1. OAuth2 Bearer Token Authentication
- Using OAuth2 access tokens for API authentication
- Automatic "Bearer" prefix handling

### 2. Custom Header Authentication
- Setting custom authentication headers
- Adding prefixes to header values (e.g., "Token", "Bearer")
- Supporting non-standard authentication schemes

### 3. Automatic Authentication Detection
- Auto-detecting authentication from agent state
- Provider-specific detection (GitHub, GitLab, etc.)
- Fallback to generic auth patterns

### 4. API Key in Different Locations
- Header-based API keys
- Query parameter API keys
- Cookie-based API keys

### 5. Session/Cookie Management
- Maintaining sessions across multiple requests
- Automatic cookie handling
- Session persistence in agent state

### 6. OAuth2 Configuration
- Client credentials flow
- Authorization code flow
- Token refresh mechanisms

### 7. Multiple Authentication Methods
- Trying different auth methods
- Fallback authentication strategies

## Running the Example

```bash
# Basic usage
go run main.go

# With debug logging
DEBUG=1 go run main.go

# With custom model
OPENAI_MODEL=gpt-4o go run main.go
```

## Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key (required)
- `OPENAI_MODEL`: Model to use (defaults to gpt-4o-mini)
- `DEBUG`: Set to "1" to enable debug logging

## Authentication State Management

The example shows how to store authentication credentials in the agent's state:

```go
// OAuth2 token
agent.UpdateState("oauth2_token", "your-access-token")

// Custom API key
agent.UpdateState("custom_api_key", "your-api-key")

// OAuth2 configuration
agent.UpdateState("oauth2_config", map[string]interface{}{
    "token_url": "https://oauth.provider.com/token",
    "client_id": "your-client-id",
    "client_secret": "your-client-secret",
})
```

## Auto-Detection Patterns

The api_client tool can automatically detect authentication based on state keys:

### Provider-Specific
- GitHub: `github_token`, `github_api_key`, `GITHUB_TOKEN`
- GitLab: `gitlab_token`, `gitlab_api_key`, `GITLAB_TOKEN`

### Generic Patterns
- Bearer tokens: `api_token`, `access_token`, `bearer_token`, `token`
- API keys: `api_key`, `apikey`, `x_api_key`

## Security Best Practices

1. **Never hardcode credentials** - Always use environment variables or secure storage
2. **Use state for credentials** - Keep credentials in agent state, not in prompts
3. **Enable session management** - Use `enable_session=true` for stateful APIs
4. **Validate SSL certificates** - The tool validates certificates by default
5. **Use appropriate auth methods** - OAuth2 for user authentication, API keys for service auth

## Example API Responses

The example includes mock responses to demonstrate different authentication scenarios:

1. **Successful OAuth2**: Returns user profile with 200 status
2. **Custom Header**: Returns data with custom authentication
3. **Auto-detected**: Automatically applies the right auth method
4. **Query Parameter**: Includes API key in URL query string
5. **Session Management**: Maintains cookies across requests

## Troubleshooting

Common authentication errors and solutions:

- **401 Unauthorized**: Check if credentials are correct and properly formatted
- **403 Forbidden**: Verify that the credentials have the required permissions
- **Missing auth**: Ensure credentials are in agent state with correct key names
- **Wrong auth type**: Check API documentation for the correct authentication method

## Advanced Usage

### OAuth2 Token Refresh

```go
// Set refresh token in state
agent.UpdateState("refresh_token", "your-refresh-token")

// Tool will automatically refresh expired tokens
```

### Custom Authentication Schemes

```go
// Digest authentication
agent.UpdateState("auth", map[string]interface{}{
    "type": "custom",
    "header_name": "Authorization",
    "header_value": "your-digest-hash",
    "prefix": "Digest",
})
```

### Multi-Tenant Authentication

```go
// Different credentials for different APIs
agent.UpdateState("github_token", "github-token")
agent.UpdateState("gitlab_token", "gitlab-token")
agent.UpdateState("custom_api_key", "custom-key")

// Tool auto-detects based on base URL
```

## Related Documentation

- [API Client Tool Documentation](../../../docs/user-guide/builtin-tools.md#api-client)
- [Authentication Guide](../../../docs/technical/authentication.md)
- [Security Best Practices](../../../docs/security.md)