// ABOUTME: Tests for GraphQL client functionality including query execution and discovery
// ABOUTME: Covers introspection, error handling, and caching mechanisms

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// TestGraphQLClientExecute tests basic GraphQL query execution
func TestGraphQLClientExecute(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Check content type
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}

		// Parse request
		var req GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Check query
		if req.Query == "" {
			t.Error("query is empty")
		}

		// Send response
		resp := GraphQLResponse{
			Data: map[string]interface{}{
				"viewer": map[string]interface{}{
					"login": "testuser",
					"name":  "Test User",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewGraphQLClient(server.URL, nil, nil)

	// Execute query
	query := `query { viewer { login name } }`
	resp, err := client.Execute(context.Background(), query, nil, "")
	if err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}

	// Check response
	if resp.Data == nil {
		t.Error("response data is nil")
	}

	// Check specific fields
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data is not a map")
	}

	viewer, ok := data["viewer"].(map[string]interface{})
	if !ok {
		t.Fatal("viewer is not a map")
	}

	if login := viewer["login"]; login != "testuser" {
		t.Errorf("expected login testuser, got %v", login)
	}
}

// TestGraphQLClientWithVariables tests query execution with variables
func TestGraphQLClientWithVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GraphQLRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Check variables
		if req.Variables == nil {
			t.Error("variables are nil")
		}

		if id, ok := req.Variables["id"].(string); !ok || id != "123" {
			t.Errorf("expected id 123, got %v", req.Variables["id"])
		}

		resp := GraphQLResponse{
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "Test User",
				},
			},
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGraphQLClient(server.URL, nil, nil)

	query := `query GetUser($id: ID!) { user(id: $id) { id name } }`
	variables := map[string]interface{}{"id": "123"}

	resp, err := client.Execute(context.Background(), query, variables, "")
	if err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}

	if resp.Data == nil {
		t.Error("response data is nil")
	}
}

// TestGraphQLClientWithErrors tests error handling
func TestGraphQLClientWithErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GraphQLResponse{
			Data: nil,
			Errors: []GraphQLError{
				{
					Message: "Field 'invalidField' doesn't exist on type 'User'",
					Path:    []interface{}{"user", "invalidField"},
					Extensions: map[string]interface{}{
						"code": "FIELD_NOT_FOUND",
					},
				},
			},
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGraphQLClient(server.URL, nil, nil)

	query := `query { user { invalidField } }`
	resp, err := client.Execute(context.Background(), query, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(resp.Errors))
	}

	if resp.Errors[0].Message != "Field 'invalidField' doesn't exist on type 'User'" {
		t.Errorf("unexpected error message: %s", resp.Errors[0].Message)
	}
}

// TestGraphQLCache tests caching functionality
func TestGraphQLCache(t *testing.T) {
	cache := NewGraphQLCache(1 * time.Minute)
	defer cache.Close()

	// Test schema caching
	schema := &ast.Schema{
		Query: &ast.Definition{
			Name: "Query",
			Kind: ast.Object,
		},
	}

	endpoint := "https://api.example.com/graphql"
	cache.SetSchema(endpoint, schema, 5*time.Second)

	// Get from cache
	cachedSchema, found := cache.GetSchema(endpoint)
	if !found {
		t.Error("schema not found in cache")
	}

	if cachedSchema.Query.Name != "Query" {
		t.Error("cached schema doesn't match")
	}

	// Test discovery caching
	discovery := &GraphQLDiscoveryResult{
		Endpoint: endpoint,
		Operations: GraphQLOperations{
			Queries: []GraphQLOperationInfo{
				{
					Name:        "viewer",
					Description: "Current user",
				},
			},
		},
	}

	cache.SetDiscovery(endpoint, discovery, 5*time.Second)

	cachedDiscovery, found := cache.GetDiscovery(endpoint)
	if !found {
		t.Error("discovery not found in cache")
	}

	if len(cachedDiscovery.Operations.Queries) != 1 {
		t.Error("cached discovery doesn't match")
	}

	// Test expiration
	time.Sleep(6 * time.Second)

	_, found = cache.GetSchema(endpoint)
	if found {
		t.Error("schema should have expired")
	}

	_, found = cache.GetDiscovery(endpoint)
	if found {
		t.Error("discovery should have expired")
	}
}

// TestDiscoverOperations tests GraphQL operation discovery
func TestDiscoverOperations(t *testing.T) {
	// Create a simple schema
	schemaStr := `
		type Query {
			"Get the current user"
			viewer: User
			"Get a user by ID"
			user(id: ID!): User
		}
		
		type Mutation {
			"Create a new user"
			createUser(input: CreateUserInput!): User
		}
		
		type User {
			id: ID!
			name: String!
			email: String
		}
		
		input CreateUserInput {
			name: String!
			email: String
		}
	`

	schema, err := gqlparser.LoadSchema(&ast.Source{Input: schemaStr})
	if err != nil {
		t.Fatalf("failed to load schema: %v", err)
	}

	discovery, err := DiscoverOperations(schema, "https://api.example.com/graphql")
	if err != nil {
		t.Fatalf("failed to discover operations: %v", err)
	}

	// Check queries
	if len(discovery.Operations.Queries) != 2 {
		t.Errorf("expected 2 queries, got %d", len(discovery.Operations.Queries))
	}

	// Check viewer query
	viewerOp := discovery.Operations.Queries[1] // sorted alphabetically
	if viewerOp.Name != "viewer" {
		t.Errorf("expected viewer query, got %s", viewerOp.Name)
	}
	if viewerOp.Description != "Get the current user" {
		t.Errorf("unexpected description: %s", viewerOp.Description)
	}
	if viewerOp.Returns != "User" {
		t.Errorf("expected return type User, got %s", viewerOp.Returns)
	}

	// Check user query
	userOp := discovery.Operations.Queries[0] // sorted alphabetically
	if userOp.Name != "user" {
		t.Errorf("expected user query, got %s", userOp.Name)
	}
	if len(userOp.Arguments) != 1 {
		t.Errorf("expected 1 argument, got %d", len(userOp.Arguments))
	}
	if userOp.Arguments[0].Name != "id" || !userOp.Arguments[0].Required {
		t.Error("unexpected argument configuration")
	}

	// Check mutations
	if len(discovery.Operations.Mutations) != 1 {
		t.Errorf("expected 1 mutation, got %d", len(discovery.Operations.Mutations))
	}

	// Check types
	if _, ok := discovery.Types["User"]; !ok {
		t.Error("User type not found in discovery")
	}
}

// TestGenerateLLMGuidance tests error guidance generation
func TestGenerateLLMGuidance(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:     "field not found",
			errMsg:   "field 'foo' doesn't exist on type 'User'",
			expected: "This field doesn't exist on the type. Check the schema discovery or introspection results to see available fields.",
		},
		{
			name:     "variable type error",
			errMsg:   "variable 'id' type mismatch",
			expected: "Variable type mismatch. Ensure variables match the expected types defined in the query.",
		},
		{
			name:     "syntax error",
			errMsg:   "syntax error at position 10",
			expected: "GraphQL syntax error. Check for missing braces, parentheses, or incorrect query structure.",
		},
		{
			name:     "unauthorized",
			errMsg:   "unauthorized access",
			expected: "Authentication required. Ensure you've provided the correct authentication credentials.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errMsg)
			guidance := GenerateLLMGuidance(err, nil)
			if guidance != tt.expected {
				t.Errorf("expected guidance %q, got %q", tt.expected, guidance)
			}
		})
	}
}

// TestDetectGraphQLAuth has been replaced by unified auth tests in auth package
// The old detectGraphQLAuth function has been removed in favor of auth.DetectAuthFromState
/*
func TestDetectGraphQLAuth(t *testing.T) {
	// This test is obsolete - authentication detection is now handled by
	// the unified auth package. See TestAPIClientTool_GraphQLAuthFromState
	// for the new implementation tests.
}
*/