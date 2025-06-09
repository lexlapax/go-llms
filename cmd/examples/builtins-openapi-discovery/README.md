# OpenAPI Discovery Example

This example demonstrates how to use the API Client Tool with OpenAPI/Swagger specifications for:
- Discovering available API endpoints
- Validating requests against the spec
- Making type-safe API calls

## Features Demonstrated

1. **OpenAPI Discovery Mode**: Fetch and parse OpenAPI specs to discover available endpoints
2. **Automatic Server URL Resolution**: Automatically use server URLs from OpenAPI specs
3. **Security Scheme Detection**: Identify and apply authentication methods from specs
4. **Request Validation**: Validate parameters and request bodies before sending
5. **Enhanced Error Guidance**: Get OpenAPI-aware error messages with specific parameter requirements
6. **Multiple API Examples**: GitHub, PetStore, and JSONPlaceholder APIs
7. **Authentication**: Automatic authentication detection and application from agent state

## Running the Example

```bash
# From the root of the project
go run cmd/examples/builtins-openapi-discovery/main.go

# Or build and run
make build-example EXAMPLE=builtins-openapi-discovery
./bin/builtins-openapi-discovery

# Enable debug logging to see agent operations
DEBUG=1 go run cmd/examples/builtins-openapi-discovery/main.go
```

## Environment Variables

- `OPENAI_API_KEY`: Required for the LLM agent
- `GITHUB_API_KEY`: Optional, enables authenticated GitHub API examples
- `DEBUG`: Set to "1" to enable detailed logging of agent operations, tool calls, and API interactions

## APIs Used

1. **GitHub API**
   - OpenAPI Spec: https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json
   - Shows real-world API discovery with a large spec

2. **PetStore API**
   - OpenAPI Spec: https://petstore3.swagger.io/api/v3/openapi.json
   - Classic OpenAPI example, great for testing

3. **JSONPlaceholder**
   - No OpenAPI spec
   - Shows the tool works with APIs that don't have specs

## Key Concepts

### Discovery Mode

When `discover_operations` is set to true, the tool fetches and parses the OpenAPI spec:

```json
{
  "base_url": "https://api.example.com",
  "endpoint": "/not-used",
  "openapi_spec": "https://api.example.com/openapi.json",
  "discover_operations": true
}
```

The discovery response includes:
- All available operations with metadata
- Server URLs from the spec
- Security schemes (API key, Bearer, OAuth2, etc.)
- LLM-friendly guidance organized by tags
- Total operation count

### Validation Mode

When making actual API calls with an `openapi_spec` URL, the tool validates:
- Path parameters
- Query parameters
- Request body schema
- Authentication requirements

### Automatic Features

1. **Server URL Resolution**: If `base_url` is not provided, the tool uses the first server URL from the OpenAPI spec
2. **Authentication Detection**: The tool looks for credentials in agent state using common key names
3. **Enhanced Error Messages**: Errors include OpenAPI-specific guidance about required parameters and allowed methods

## Example Output

The agent will:
1. Discover available operations from each API
2. Show operation summaries organized by tags
3. Display authentication requirements and server URLs
4. Make validated API calls with automatic auth detection
5. Handle errors with OpenAPI-aware guidance
6. Provide specific parameter requirements on 400 errors
7. List allowed methods on 405 errors
8. Show required authentication methods on 401 errors