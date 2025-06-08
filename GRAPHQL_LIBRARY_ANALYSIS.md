# GraphQL Library Analysis for Go

## Overview

This document analyzes the main GraphQL client libraries available for Go to determine the best fit for our api_client tool integration.

## Libraries Evaluated

### 1. github.com/shurcooL/graphql

**Pros:**
- Type-safe GraphQL client using Go structs
- Simple and intuitive API
- Good for static queries
- Well-maintained by Dmitri Shuralyov (Go team member)
- No external dependencies

**Cons:**
- Requires defining Go structs for each query
- Not suitable for dynamic query construction
- Limited introspection support
- Not ideal for LLM-generated queries

**Example:**
```go
var query struct {
    Viewer struct {
        Login string
        Name  string
    }
}
client.Query(context.Background(), &query, nil)
```

### 2. github.com/hasura/go-graphql-client

**Pros:**
- Fork of shurcooL/graphql with more features
- Supports subscriptions
- Better error handling
- More active development
- Supports custom scalar types

**Cons:**
- Still requires Go structs for queries
- Not suitable for dynamic queries from LLMs
- Limited introspection capabilities

### 3. github.com/machinebox/graphql

**Pros:**
- Supports dynamic query construction
- Simple API for runtime queries
- Good for LLM-generated queries
- Supports variables, headers, and custom HTTP clients
- Clean error handling

**Cons:**
- Less actively maintained (last update 2020)
- No built-in subscription support
- Basic feature set

**Example:**
```go
client := graphql.NewClient(endpoint)
req := graphql.NewRequest(`
    query {
        viewer {
            login
            name
        }
    }
`)
var resp map[string]interface{}
err := client.Run(ctx, req, &resp)
```

### 4. github.com/Khan/genqlient

**Pros:**
- Code generation from GraphQL queries
- Type-safe with generated code
- Good for known queries
- Excellent performance

**Cons:**
- Requires build-time code generation
- Not suitable for dynamic queries
- Cannot handle LLM-generated queries

### 5. github.com/graphql-go/graphql

**Pros:**
- Full GraphQL implementation (server and client)
- Complete introspection support
- Can parse and validate GraphQL queries
- Supports the entire GraphQL spec

**Cons:**
- Primarily a server library
- Heavier than needed for just client functionality
- More complex API

### 6. Custom Implementation

**Pros:**
- Full control over implementation
- Can optimize for LLM use case
- Minimal dependencies
- Can reuse existing HTTP client code

**Cons:**
- More development effort
- Need to implement query parsing
- Need to handle GraphQL spec compliance

## Recommendation

For our use case, I recommend a **hybrid approach**:

1. **Use github.com/graphql-go/graphql for:**
   - Schema parsing and validation
   - Introspection query support
   - Query validation against schema
   - Type system understanding

2. **Custom implementation for:**
   - Query execution (reuse our existing HTTP client)
   - Response handling
   - Error formatting for LLMs
   - Caching layer

3. **Optional: Use github.com/vektah/gqlparser/v2 for:**
   - Lightweight query parsing
   - Schema parsing without full server implementation
   - AST manipulation

## Implementation Plan

```go
// Core dependencies
import (
    "github.com/vektah/gqlparser/v2"        // For parsing
    "github.com/vektah/gqlparser/v2/ast"    // For AST
    "github.com/vektah/gqlparser/v2/parser" // For parsing queries
    "github.com/vektah/gqlparser/v2/validator" // For validation
)

// Our GraphQL client structure
type GraphQLClient struct {
    endpoint   string
    httpClient *http.Client
    schema     *ast.Schema
    cache      *GraphQLCache
}

// Execute dynamic queries
func (c *GraphQLClient) Execute(query string, variables map[string]interface{}) (interface{}, error) {
    // Parse query
    doc, err := parser.ParseQuery(&ast.Source{Input: query})
    if err != nil {
        return nil, err
    }
    
    // Validate against schema if available
    if c.schema != nil {
        errs := validator.Validate(c.schema, doc)
        if len(errs) > 0 {
            return nil, formatErrors(errs)
        }
    }
    
    // Execute query using our HTTP client
    return c.executeHTTP(doc, variables)
}
```

## Decision

**Selected Approach**: Custom implementation with `github.com/vektah/gqlparser/v2`

**Reasons:**
1. **gqlparser** is the parser used by gqlgen (most popular GraphQL server for Go)
2. Lightweight and focused on parsing/validation
3. Actively maintained
4. Supports the full GraphQL specification
5. Allows us to build exactly what we need for LLM integration
6. Can reuse our existing HTTP client infrastructure
7. Enables custom error messages and guidance for LLMs

**Benefits:**
- Full control over LLM-friendly features
- Minimal dependencies
- Consistent with our existing patterns
- Optimal performance through custom caching
- Better integration with our tool system