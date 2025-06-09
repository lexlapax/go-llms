// ABOUTME: Simple tests for GraphQL authentication detection with proper GitHub URL patterns
// ABOUTME: Tests auto-auth detection without complex URL manipulations

package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// TestAPIClientTool_GraphQLAuthSimple tests GraphQL with simple auth detection
func TestAPIClientTool_GraphQLAuthSimple(t *testing.T) {
	// Create test server that doesn't require auth
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the auth headers for debugging
		auth := r.Header.Get("Authorization")
		apiKey := r.Header.Get("X-API-Key")
		t.Logf("Received Authorization header: %s", auth)
		t.Logf("Received X-API-Key header: %s", apiKey)
		
		// Decode GraphQL request
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Return success response regardless of auth
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"login": "testuser",
					"name":  "Test User",
					"auth":  auth, // Echo auth header for verification
					"apikey": apiKey, // Echo API key header for verification
				},
			},
		})
	}))
	defer server.Close()

	tool := NewAPIClientTool()

	tests := []struct {
		name          string
		stateValues   map[string]interface{}
		params        map[string]interface{}
		expectedAuth  string
	}{
		{
			name: "Generic API token becomes bearer",
			stateValues: map[string]interface{}{
				"api_token": "test-token-123",
			},
			params: map[string]interface{}{
				"base_url":      server.URL,
				"endpoint":      "/graphql",
				"graphql_query": "query { viewer { login name } }",
			},
			expectedAuth: "Bearer test-token-123",
		},
		{
			name: "API key with proper type",
			stateValues: map[string]interface{}{
				"api_key": "key-456",
			},
			params: map[string]interface{}{
				"base_url":      server.URL,
				"endpoint":      "/graphql",
				"graphql_query": "query { viewer { login name } }",
			},
			expectedAuth: "", // API key goes in X-API-Key header, not Authorization
		},
		{
			name: "No auth in state",
			stateValues: map[string]interface{}{},
			params: map[string]interface{}{
				"base_url":      server.URL,
				"endpoint":      "/graphql",
				"graphql_query": "query { viewer { login name } }",
			},
			expectedAuth: "",
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
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check result
			if resultMap, ok := result.(map[string]interface{}); ok {
				if !resultMap["success"].(bool) {
					t.Errorf("Expected success but got: %v", result)
				}
				
				// Check auth header was applied correctly
				if data, ok := resultMap["data"].(map[string]interface{}); ok {
					if viewer, ok := data["viewer"].(map[string]interface{}); ok {
						if auth, ok := viewer["auth"].(string); ok {
							if auth != tt.expectedAuth {
								t.Errorf("Expected auth header %q, got %q", tt.expectedAuth, auth)
							}
						}
						// Check API key header for api_key type
						if tt.name == "API key with proper type" {
							if apiKey, ok := viewer["apikey"].(string); ok {
								if apiKey != "key-456" {
									t.Errorf("Expected X-API-Key header 'key-456', got %q", apiKey)
								}
							} else {
								t.Error("Expected X-API-Key header to be set")
							}
						}
					}
				}
			}
		})
	}
}