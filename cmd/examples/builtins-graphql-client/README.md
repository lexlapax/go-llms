# GraphQL Client Example

This example demonstrates how to use the `api_client` tool to interact with GraphQL APIs, including:
- Schema discovery through introspection
- Executing queries with and without variables
- Handling GraphQL-specific errors
- Working with nested queries

## Prerequisites

For GitHub GraphQL API examples:
```bash
export GITHUB_API_KEY="your-github-personal-access-token"
```

For the LLM provider:
```bash
export ANTHROPIC_API_KEY="your-api-key"
# or
export OPENAI_API_KEY="your-api-key"
# or
export GEMINI_API_KEY="your-api-key"
```

## Running the Example

```bash
# Run with default settings
go run cmd/examples/builtins-graphql-client/main.go

# Run with debug logging
DEBUG=1 go run cmd/examples/builtins-graphql-client/main.go

# Use a specific model
ANTHROPIC_MODEL=claude-3-5-sonnet-latest go run cmd/examples/builtins-graphql-client/main.go
```

## Features Demonstrated

### 1. Schema Discovery
The example shows how to discover available GraphQL operations using introspection:
```go
state.Set("user_input", "Use the api_client tool to discover what GraphQL operations are available...")
```

### 2. Simple Queries
Execute basic GraphQL queries without variables:
```graphql
query {
  viewer {
    login
    name
    email
    bio
  }
}
```

### 3. Parameterized Queries
Use GraphQL variables for dynamic queries:
```graphql
query GetRepo($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    name
    description
    stargazerCount
  }
}
```

### 4. Public APIs
The example includes the Countries GraphQL API which doesn't require authentication, demonstrating that the tool works with various GraphQL endpoints.

### 5. Error Handling
Shows how GraphQL errors are handled and reported:
- Field validation errors
- Type mismatches
- Authentication failures

### 6. Complex Queries
Demonstrates nested queries with multiple levels:
```graphql
query {
  viewer {
    repositories(first: 5) {
      nodes {
        name
        languages(first: 3) {
          nodes {
            name
          }
        }
      }
    }
  }
}
```

## GraphQL-Specific Parameters

The `api_client` tool supports these GraphQL parameters:

- `graphql_query`: The GraphQL query or mutation string
- `graphql_variables`: Variables for parameterized queries
- `graphql_operation_name`: Name when multiple operations exist
- `discover_graphql`: Set to true for schema introspection
- `max_graphql_depth`: Maximum query depth (default: 5)

## Example Output

```
=== Example 1: Discovering GitHub GraphQL Schema ===
I'll discover the available GraphQL operations at GitHub's API.

The GitHub GraphQL API offers a rich set of operations:

**Queries Available:**
1. `viewer` - Get information about the currently authenticated user
2. `repository` - Look up a repository by owner and name
3. `user` - Look up a user by login
4. `organization` - Look up an organization by login
5. `search` - Search for various items (repositories, users, issues, etc.)

**Key Types:**
- User: Contains fields like login, name, email, bio, repositories, followers
- Repository: Contains name, description, stargazerCount, issues, pullRequests
- Organization: Contains login, name, description, members, repositories

=== Example 2: Query Current User ===
Here's the information about the current authenticated user:

- Login: octocat
- Name: The Octocat
- Email: octocat@github.com
- Bio: GitHub's mascot

...
```

## Troubleshooting

1. **Authentication Errors**: Ensure your GitHub token has the necessary scopes
2. **Rate Limiting**: GitHub GraphQL API has rate limits based on query complexity
3. **Schema Changes**: GraphQL schemas can evolve; use discovery to check current fields

## Advanced Usage

### Custom Headers
```go
params := map[string]interface{}{
    "headers": map[string]interface{}{
        "X-Custom-Header": "value",
    },
}
```

### Timeout Configuration
```go
params := map[string]interface{}{
    "timeout": "60s", // For complex queries
}
```

### Using Fragments (coming in Phase 3)
```go
params := map[string]interface{}{
    "graphql_fragments": map[string]interface{}{
        "userFields": "fragment userFields on User { login name email }",
    },
}
```