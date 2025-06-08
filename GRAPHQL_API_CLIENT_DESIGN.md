# GraphQL Support Design for API Client Tool

## Overview

This document outlines the design for adding GraphQL support to the `api_client` tool, focusing on making GraphQL accessible and intuitive for LLM consumption while maintaining consistency with our existing REST and OpenAPI patterns.

## Design Goals

1. **LLM-Friendly Interface**: Make GraphQL queries and mutations easy for LLMs to construct and execute
2. **Schema Discovery**: Automatic introspection and schema understanding
3. **Validation & Guidance**: Proactive error prevention with helpful suggestions
4. **Performance**: Efficient caching of schemas and query optimization
5. **Consistency**: Follow patterns established in REST and OpenAPI support

## Key Components

### 1. GraphQL Parameters for api_client

```go
// New parameters for api_client tool
type GraphQLParams struct {
    // Main GraphQL parameters
    GraphQLEndpoint   string              `json:"graphql_endpoint,omitempty"`   // GraphQL endpoint URL
    GraphQLQuery      string              `json:"graphql_query,omitempty"`      // GraphQL query or mutation
    GraphQLVariables  map[string]any      `json:"graphql_variables,omitempty"`  // Query variables
    
    // Discovery parameters
    DiscoverGraphQL   bool                `json:"discover_graphql,omitempty"`   // Discover schema operations
    GraphQLOperation  string              `json:"graphql_operation,omitempty"`  // Specific operation to explore
    
    // Advanced parameters
    GraphQLFragments  map[string]string   `json:"graphql_fragments,omitempty"`  // Named fragments
    OperationName     string              `json:"operation_name,omitempty"`     // Named operation
    MaxDepth          int                 `json:"max_depth,omitempty"`          // Max query depth (default: 5)
}
```

### 2. GraphQL Schema Representation

```go
// LLM-friendly schema representation
type GraphQLSchema struct {
    // Core schema information
    QueryType        *TypeInfo
    MutationType     *TypeInfo
    SubscriptionType *TypeInfo
    Types            map[string]*TypeInfo
    
    // LLM guidance
    Description      string
    Examples         []QueryExample
    CommonPatterns   []QueryPattern
}

type TypeInfo struct {
    Name        string
    Kind        string              // OBJECT, SCALAR, ENUM, etc.
    Description string
    Fields      map[string]*FieldInfo
    EnumValues  []string            // For ENUM types
}

type FieldInfo struct {
    Name         string
    Type         string              // Human-readable type representation
    Description  string
    Args         map[string]*ArgInfo
    IsRequired   bool
    IsList       bool
    Examples     []string            // Example values or usage
}
```

### 3. Discovery Response Format

```json
{
  "graphql_schema": {
    "endpoint": "https://api.github.com/graphql",
    "operations": {
      "queries": [
        {
          "name": "viewer",
          "description": "The currently authenticated user",
          "example": "query { viewer { login name email } }",
          "available_fields": ["login", "name", "email", "bio", "company", "repositories"]
        },
        {
          "name": "repository",
          "description": "Lookup a repository by owner and name",
          "example": "query { repository(owner: \"octocat\", name: \"hello-world\") { name description stargazerCount } }",
          "required_args": ["owner", "name"],
          "available_fields": ["name", "description", "stargazerCount", "forkCount", "issues"]
        }
      ],
      "mutations": [
        {
          "name": "createIssue",
          "description": "Create a new issue",
          "example": "mutation { createIssue(input: {repositoryId: \"...\", title: \"Bug Report\"}) { issue { number title } } }",
          "input_type": "CreateIssueInput",
          "required_fields": ["repositoryId", "title"]
        }
      ]
    },
    "types": {
      "User": {
        "fields": ["login", "name", "email", "bio", "company"],
        "description": "A user is an individual's account on GitHub"
      }
    }
  }
}
```

### 4. Query Building Assistance

The tool will provide intelligent query building assistance:

```go
// Example of LLM-friendly query building
type QueryBuilder struct {
    schema *GraphQLSchema
    
    // Build query with field selection
    BuildQuery(typeName string, fields []string) string
    
    // Suggest fields based on partial input
    SuggestFields(typeName string, prefix string) []string
    
    // Validate query against schema
    ValidateQuery(query string) ValidationResult
    
    // Generate example queries
    GenerateExamples(operationType string) []QueryExample
}
```

### 5. Error Handling with Guidance

GraphQL-specific error handling with LLM guidance:

```go
type GraphQLError struct {
    Message    string
    Path       []string              // Path to the error in the query
    Extensions map[string]any        // Additional error details
    
    // LLM guidance
    Suggestion     string            // How to fix the error
    ValidExample   string            // Working example
    Documentation  string            // Link to relevant docs
}
```

## Usage Examples

### 1. Schema Discovery

```yaml
# Discover available operations
- tool: api_client
  parameters:
    graphql_endpoint: "https://api.github.com/graphql"
    discover_graphql: true
    headers:
      Authorization: "Bearer ${GITHUB_TOKEN}"
```

### 2. Simple Query

```yaml
# Get current user information
- tool: api_client
  parameters:
    graphql_endpoint: "https://api.github.com/graphql"
    graphql_query: |
      query {
        viewer {
          login
          name
          email
          bio
        }
      }
    headers:
      Authorization: "Bearer ${GITHUB_TOKEN}"
```

### 3. Query with Variables

```yaml
# Get repository information
- tool: api_client
  parameters:
    graphql_endpoint: "https://api.github.com/graphql"
    graphql_query: |
      query GetRepo($owner: String!, $name: String!) {
        repository(owner: $owner, name: $name) {
          name
          description
          stargazerCount
          forkCount
        }
      }
    graphql_variables:
      owner: "golang"
      name: "go"
    headers:
      Authorization: "Bearer ${GITHUB_TOKEN}"
```

### 4. Mutation Example

```yaml
# Create an issue
- tool: api_client
  parameters:
    graphql_endpoint: "https://api.github.com/graphql"
    graphql_query: |
      mutation CreateIssue($input: CreateIssueInput!) {
        createIssue(input: $input) {
          issue {
            number
            title
            url
          }
        }
      }
    graphql_variables:
      input:
        repositoryId: "MDEwOlJlcG9zaXRvcnkxMjk2MjY5"
        title: "New Feature Request"
        body: "Description of the feature"
    headers:
      Authorization: "Bearer ${GITHUB_TOKEN}"
```

## Implementation Strategy

### Phase 1: Infrastructure (Day 1)
1. Select GraphQL client library (likely `github.com/hasura/go-graphql-client` or `github.com/shurcooL/graphql`)
2. Integrate with existing api_client tool structure
3. Implement introspection query support
4. Create schema caching mechanism

### Phase 2: Query Execution (Day 2)
1. Implement query parser and validator
2. Add variable support
3. Create response handling
4. Implement error translation

### Phase 3: Advanced Features (Day 3)
1. Add mutation support
2. Implement fragments
3. Add query complexity analysis
4. Support subscriptions (if feasible)

### Phase 4: LLM Integration (Day 4)
1. Create discovery responses
2. Generate examples from schema
3. Implement query suggestions
4. Add comprehensive error guidance

### Phase 5: Testing & Documentation (Day 5)
1. Create comprehensive test suite
2. Write documentation
3. Create examples
4. Performance benchmarks

## Caching Strategy

Similar to OpenAPI caching:
- In-memory cache for schemas with TTL
- Operation index for fast lookup
- Lazy loading of schema details
- Memory pooling for efficiency

## Security Considerations

1. **Query Depth Limiting**: Prevent deeply nested queries
2. **Query Complexity**: Calculate and limit query complexity
3. **Rate Limiting**: Respect GraphQL rate limits
4. **Authentication**: Support various auth methods (Bearer, API Key, etc.)

## Testing Strategy

1. **Unit Tests**: Test each component in isolation
2. **Integration Tests**: Test with real GraphQL APIs (GitHub, Shopify)
3. **Mock Tests**: Test with mock GraphQL servers
4. **Performance Tests**: Benchmark query execution and caching

## Success Metrics

1. **Ease of Use**: LLMs can successfully construct and execute GraphQL queries
2. **Performance**: Schema caching provides <1ms lookups
3. **Error Reduction**: Validation prevents 90%+ of common GraphQL errors
4. **Coverage**: Support for 95%+ of GraphQL specification features