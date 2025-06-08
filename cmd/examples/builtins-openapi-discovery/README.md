# OpenAPI Discovery Example

This example demonstrates how to use the API Client Tool with OpenAPI/Swagger specifications for:
- Discovering available API endpoints
- Validating requests against the spec
- Making type-safe API calls

## Features Demonstrated

1. **OpenAPI Discovery Mode**: Fetch and parse OpenAPI specs to discover available endpoints
2. **Request Validation**: Validate parameters and request bodies before sending
3. **Multiple API Examples**: GitHub, PetStore, and JSONPlaceholder APIs
4. **Error Handling**: See how validation catches errors before they reach the API
5. **Authentication**: Use API keys with OpenAPI-documented endpoints

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
- `GITHUB_TOKEN`: Optional, enables authenticated GitHub API examples
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

### Validation Mode

When making actual API calls with an `openapi_spec` URL, the tool validates:
- Path parameters
- Query parameters
- Request body schema
- Authentication requirements

## Example Output

The agent will:
1. Discover available operations from each API
2. Show operation summaries and required parameters
3. Make validated API calls
4. Handle errors with helpful guidance