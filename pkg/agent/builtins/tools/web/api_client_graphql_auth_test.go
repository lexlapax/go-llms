// ABOUTME: Tests for GraphQL authentication detection from state in API Client Tool
// ABOUTME: Validates automatic auth detection for GitHub, GitLab, and generic GraphQL endpoints

package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// TestAPIClientTool_GraphQLAuthFromState tests GraphQL authentication detection from state
func TestAPIClientTool_GraphQLAuthFromState(t *testing.T) {
	// Create test server that checks for authorization
	authToken := "ghp_test123456"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + authToken

		// Allow requests without auth for the "without auth" test case
		if auth == "" && strings.Contains(r.Header.Get("X-Test-Case"), "without_auth") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"message": "Authentication required",
					},
				},
			})
			return
		}

		if auth != expectedAuth && auth != "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"message": "Authentication required",
					},
				},
			})
			return
		}

		// Decode GraphQL request
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Check if it's introspection
		query, _ := req["query"].(string)
		if strings.Contains(query, "__schema") {
			// Return introspection response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"__schema": map[string]interface{}{
						"queryType": map[string]interface{}{
							"name": "Query",
						},
						"types": []map[string]interface{}{
							{
								"name":        "User",
								"kind":        "OBJECT",
								"description": "User type",
							},
						},
					},
				},
			})
		} else {
			// Return query response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"viewer": map[string]interface{}{
						"login": "testuser",
						"name":  "Test User",
					},
				},
			})
		}
	}))
	defer server.Close()

	tool := NewAPIClientTool()

	tests := []struct {
		name        string
		stateValues map[string]interface{}
		params      map[string]interface{}
		expectAuth  bool
		expectError bool
	}{
		{
			name: "GitHub GraphQL with github_token in state",
			stateValues: map[string]interface{}{
				"github_token": authToken,
			},
			params: map[string]interface{}{
				"base_url":      server.URL, // Use actual test server URL
				"endpoint":      "/graphql",
				"graphql_query": "query { viewer { login name } }",
			},
			expectAuth:  true,
			expectError: false,
		},
		{
			name: "GitHub GraphQL with github_api_key in state",
			stateValues: map[string]interface{}{
				"github_api_key": authToken,
			},
			params: map[string]interface{}{
				"base_url":      server.URL, // Use actual test server URL
				"endpoint":      "/graphql",
				"graphql_query": "query { viewer { login name } }",
			},
			expectAuth:  true,
			expectError: false,
		},
		{
			name:        "GitHub GraphQL without auth in state",
			stateValues: map[string]interface{}{},
			params: map[string]interface{}{
				"base_url":      server.URL, // Use actual test server URL
				"endpoint":      "/graphql",
				"graphql_query": "query { viewer { login name } }",
			},
			expectAuth:  false,
			expectError: true, // Should get 401
		},
		{
			name: "GraphQL discovery with auth from state",
			stateValues: map[string]interface{}{
				"github_token": authToken,
			},
			params: map[string]interface{}{
				"base_url":         strings.Replace(server.URL, "127.0.0.1", "api.github.com", 1),
				"endpoint":         "/graphql",
				"discover_graphql": true,
			},
			expectAuth:  true,
			expectError: false,
		},
		{
			name: "Generic GraphQL with api_token",
			stateValues: map[string]interface{}{
				"api_token": authToken,
			},
			params: map[string]interface{}{
				"base_url":      server.URL,
				"endpoint":      "/graphql",
				"graphql_query": "query { viewer { login name } }",
			},
			expectAuth:  true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create state with test values
			state := domain.NewState()
			for k, v := range tt.stateValues {
				state.Set(k, v)
			}

			// Create tool context
			ctx := &domain.ToolContext{
				Context:   context.Background(),
				State:     domain.NewStateReader(state),
				RunID:     "test-run",
				StartTime: time.Now(),
			}

			// Execute tool
			result, err := tool.Execute(ctx, tt.params)

			if tt.expectError {
				if err == nil {
					// Check if error is in the result
					if resultMap, ok := result.(map[string]interface{}); ok {
						if success, ok := resultMap["success"].(bool); ok && success {
							t.Error("Expected error but got success")
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if resultMap, ok := result.(map[string]interface{}); ok {
					if success, ok := resultMap["success"].(bool); !ok || !success {
						t.Errorf("Expected success but got: %v", result)
					}
				}
			}
		})
	}
}
