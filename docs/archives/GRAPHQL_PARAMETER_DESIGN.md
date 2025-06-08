# GraphQL Parameter Design for API Client Tool

## Overview

This document details how GraphQL support will be integrated into the existing `api_client` tool parameter structure, maintaining consistency with the current REST and OpenAPI functionality.

## Parameter Integration Strategy

### Option 1: Extend Existing Parameters (Recommended)

Add GraphQL-specific parameters to the existing `api_client` tool:

```go
// Additional parameters for GraphQL support
"graphql_query": {
    Type:        "string",
    Description: "GraphQL query or mutation string. When provided, the tool operates in GraphQL mode",
},
"graphql_variables": {
    Type:        "object",
    Description: "Variables for the GraphQL query (e.g., {'userId': '123', 'limit': 10})",
},
"graphql_operation_name": {
    Type:        "string",
    Description: "Name of the operation to execute when query contains multiple operations",
},
"discover_graphql": {
    Type:        "boolean",
    Description: "If true, performs introspection to discover available queries, mutations, and types",
},
"graphql_operation": {
    Type:        "string",
    Description: "Specific operation to get details about (e.g., 'query.user' or 'mutation.createUser')",
}
```

### Option 2: Separate GraphQL Mode

Create a mode parameter that switches between REST and GraphQL:

```go
"mode": {
    Type:        "string",
    Description: "API mode: 'rest' (default) or 'graphql'",
    Enum:        []string{"rest", "graphql"},
}
```

## Recommended Implementation

### 1. Detection Logic

The tool will automatically detect GraphQL mode when:
- `graphql_query` parameter is provided
- `discover_graphql` is set to true
- The endpoint ends with `/graphql` or `/gql`

### 2. Parameter Compatibility

When in GraphQL mode:
- `method` is always POST (GraphQL standard)
- `endpoint` is the GraphQL endpoint (e.g., `/graphql`)
- `body` parameter is ignored (GraphQL query takes precedence)
- `path_params` are not used (GraphQL uses variables)
- `query_params` can still be used for endpoint-specific needs
- All authentication methods remain compatible

### 3. Updated Parameter Schema

```go
// GraphQL-specific parameters to add
"graphql_query": {
    Type:        "string",
    Description: "GraphQL query or mutation. Example: 'query { user(id: $userId) { name email } }'",
},
"graphql_variables": {
    Type:        "object",
    Description: "Variables for the GraphQL query. Example: {'userId': '123'}",
},
"graphql_operation_name": {
    Type:        "string",
    Description: "Operation name when multiple operations exist in the query",
},
"discover_graphql": {
    Type:        "boolean",
    Description: "Discover available GraphQL operations via introspection",
},
"graphql_fragments": {
    Type:        "object",
    Description: "Named fragments to include. Example: {'userFields': 'fragment userFields on User { id name email }'}",
},
"max_graphql_depth": {
    Type:        "integer",
    Description: "Maximum depth for GraphQL queries (default: 5, max: 10)",
}
```

## Usage Examples

### 1. Simple GraphQL Query

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/graphql",
  "graphql_query": "query { viewer { login name email } }",
  "auth": {
    "type": "bearer",
    "token": "github_token_here"
  }
}
```

### 2. Query with Variables

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/graphql",
  "graphql_query": "query GetRepo($owner: String!, $name: String!) { repository(owner: $owner, name: $name) { name description stargazerCount } }",
  "graphql_variables": {
    "owner": "golang",
    "name": "go"
  },
  "auth": {
    "type": "bearer",
    "token": "github_token_here"
  }
}
```

### 3. GraphQL Discovery

```json
{
  "base_url": "https://api.github.com",
  "endpoint": "/graphql",
  "discover_graphql": true,
  "auth": {
    "type": "bearer",
    "token": "github_token_here"
  }
}
```

### 4. Mutation Example

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

## Response Format

### GraphQL Query Response

```json
{
  "success": true,
  "status_code": 200,
  "data": {
    "viewer": {
      "login": "octocat",
      "name": "The Octocat",
      "email": "octocat@github.com"
    }
  },
  "graphql_extensions": {
    "requestId": "abc123",
    "cost": 3
  }
}
```

### GraphQL Error Response

```json
{
  "success": false,
  "status_code": 200,
  "data": null,
  "error_message": "Field 'invalidField' doesn't exist on type 'User'",
  "error_details": {
    "errors": [
      {
        "message": "Field 'invalidField' doesn't exist on type 'User'",
        "path": ["viewer", "invalidField"],
        "extensions": {
          "code": "FIELD_NOT_FOUND"
        }
      }
    ]
  },
  "error_guidance": "The field 'invalidField' is not available on the User type. Available fields include: login, name, email, bio, company, location. Try: query { viewer { login name email } }"
}
```

### GraphQL Discovery Response

```json
{
  "success": true,
  "status_code": 200,
  "graphql_schema": {
    "queries": [
      {
        "name": "viewer",
        "description": "The currently authenticated user",
        "example": "query { viewer { login name email } }",
        "returns": "User",
        "arguments": []
      },
      {
        "name": "repository",
        "description": "Lookup a repository",
        "example": "query { repository(owner: \"owner\", name: \"repo\") { name } }",
        "returns": "Repository",
        "arguments": [
          {
            "name": "owner",
            "type": "String!",
            "description": "The repository owner"
          },
          {
            "name": "name",
            "type": "String!",
            "description": "The repository name"
          }
        ]
      }
    ],
    "mutations": [
      {
        "name": "createIssue",
        "description": "Create a new issue",
        "example": "mutation { createIssue(input: {repositoryId: \"...\", title: \"...\"}) { issue { number } } }",
        "returns": "CreateIssuePayload",
        "input_type": "CreateIssueInput"
      }
    ],
    "types": {
      "User": {
        "description": "A user account",
        "fields": ["login", "name", "email", "bio", "company"]
      }
    }
  }
}
```

## Validation and Error Handling

1. **Query Validation**: Validate GraphQL syntax before sending
2. **Variable Type Checking**: Ensure variables match expected types
3. **Schema Validation**: If schema is cached, validate query against it
4. **Depth Limiting**: Prevent excessively deep queries
5. **Error Translation**: Convert GraphQL errors to LLM-friendly guidance

## Implementation Priority

1. **Phase 1**: Basic GraphQL query execution
2. **Phase 2**: Variable support and error handling
3. **Phase 3**: Discovery and introspection
4. **Phase 4**: Advanced features (fragments, subscriptions)
5. **Phase 5**: Schema caching and validation