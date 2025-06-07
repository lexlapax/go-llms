# API Client Tool Example

This example demonstrates how to use the API Client Tool with an LLM agent to interact with REST APIs.

## Features Demonstrated

1. **Basic GET Requests**: Fetching user information from GitHub API
2. **Search Queries**: Using query parameters to search repositories
3. **Authentication**: Making authenticated requests with API tokens
4. **POST Requests**: Creating resources (GitHub gists)
5. **Error Handling**: Graceful handling of 404 and other errors
6. **Path Parameters**: Using dynamic path segments in API endpoints
7. **Multiple APIs**: Working with different REST APIs (GitHub and JSONPlaceholder)

## Running the Example

### Basic Usage (No Authentication)

```bash
go run cmd/examples/builtins-web-api-client/main.go
```

This will run examples that don't require authentication:
- Fetching public GitHub user information
- Searching public repositories
- Handling API errors
- Using path parameters
- Interacting with JSONPlaceholder API

### With GitHub Authentication

To run the authenticated examples (creating gists, checking rate limits), provide a GitHub personal access token:

```bash
GITHUB_TOKEN="your-github-token" go run cmd/examples/builtins-web-api-client/main.go
```

To create a GitHub token:
1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Give it a name and select the "gist" scope
4. Copy the token and use it as shown above

## API Client Tool Capabilities

The API Client Tool supports:

### HTTP Methods
- GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS

### Authentication Types
- **API Key**: Header or query parameter placement
- **Bearer Token**: OAuth2 and JWT tokens
- **Basic Auth**: Username and password

### Request Features
- Path parameter substitution (e.g., `/users/{user_id}`)
- Query parameters
- Custom headers
- JSON request bodies
- Configurable timeouts

### Response Handling
- Automatic JSON parsing
- Status code interpretation
- Error details extraction
- Helpful error guidance for common HTTP errors

## Example API Calls

### Simple GET Request
```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/users/octocat",
  "method": "GET"
}
```

### Authenticated POST Request
```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/gists",
  "method": "POST",
  "auth": {
    "type": "bearer",
    "token": "your-token-here"
  },
  "body": {
    "description": "My Gist",
    "public": true,
    "files": {
      "hello.txt": {
        "content": "Hello, World!"
      }
    }
  }
}
```

### Using Path Parameters
```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/repos/{owner}/{repo}",
  "method": "GET",
  "path_params": {
    "owner": "lexlapax",
    "repo": "go-llms"
  }
}
```

## Error Guidance

The tool provides helpful guidance for common HTTP errors:

- **400 Bad Request**: Check parameter formatting
- **401 Unauthorized**: Add authentication credentials
- **403 Forbidden**: Check permissions for the API key
- **404 Not Found**: Verify endpoint and path parameters
- **429 Too Many Requests**: Implement rate limiting
- **500+ Server Errors**: Retry with exponential backoff

## Best Practices

1. **Security**: Never hardcode API keys in your code
2. **Rate Limiting**: Respect API rate limits
3. **Error Handling**: Always check the `success` field in responses
4. **Timeouts**: Set appropriate timeouts for long-running requests
5. **Retries**: Implement exponential backoff for transient failures