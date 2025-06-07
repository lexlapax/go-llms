// ABOUTME: Integration tests for API Client Tool OpenAPI discovery and validation features
// ABOUTME: Tests discovery mode, validation, and real API interactions with OpenAPI specs

package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// TestAPIClientTool_OpenAPIDiscovery tests the OpenAPI discovery functionality
func TestAPIClientTool_OpenAPIDiscovery(t *testing.T) {
	// Create a mock OpenAPI spec
	mockSpec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Test API",
			"version":     "1.0.0",
			"description": "A test API for unit tests",
		},
		"paths": map[string]interface{}{
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"operationId": "listUsers",
					"summary":     "List all users",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
						},
					},
				},
				"post": map[string]interface{}{
					"operationId": "createUser",
					"summary":     "Create a new user",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"name": map[string]interface{}{
											"type": "string",
										},
										"email": map[string]interface{}{
											"type": "string",
										},
									},
									"required": []string{"name", "email"},
								},
							},
						},
					},
				},
			},
			"/users/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"operationId": "getUser",
					"summary":     "Get a user by ID",
					"parameters": []map[string]interface{}{
						{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		},
	}

	// Create test server that serves the OpenAPI spec
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockSpec)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Create the tool
	tool := createAPIClientTool()
	ctx := &domain.ToolContext{
		Context: context.Background(),
	}

	// Test discovery mode
	t.Run("Discovery Mode", func(t *testing.T) {
		params := map[string]interface{}{
			"base_url":            server.URL,
			"endpoint":            "/dummy",
			"openapi_spec":        server.URL + "/openapi.json",
			"discover_operations": true,
		}

		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Discovery failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		// Check success
		if !resultMap["success"].(bool) {
			t.Error("Expected success to be true")
		}

		// Check operations
		operations, ok := resultMap["operations"].([]OperationInfo)
		if !ok {
			t.Fatalf("Expected operations to be []OperationInfo, got %T", resultMap["operations"])
		}

		// Should have 3 operations
		if len(operations) != 3 {
			t.Errorf("Expected 3 operations, got %d", len(operations))
		}

		// Check spec info
		specInfo, ok := resultMap["spec_info"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected spec_info to be map, got %T", resultMap["spec_info"])
		}

		if specInfo["title"] != "Test API" {
			t.Errorf("Expected title 'Test API', got %v", specInfo["title"])
		}
	})
}

// TestAPIClientTool_OpenAPIValidation tests request validation against OpenAPI spec
func TestAPIClientTool_OpenAPIValidation(t *testing.T) {
	// Create a more detailed OpenAPI spec with validation rules
	mockSpec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Validation Test API",
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{
			"/pets": map[string]interface{}{
				"post": map[string]interface{}{
					"operationId": "createPet",
					"summary":     "Create a pet",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"name": map[string]interface{}{
											"type":      "string",
											"minLength": 1,
											"maxLength": 50,
										},
										"age": map[string]interface{}{
											"type":    "integer",
											"minimum": 0,
											"maximum": 100,
										},
										"species": map[string]interface{}{
											"type": "string",
											"enum": []string{"dog", "cat", "bird"},
										},
									},
									"required": []string{"name", "species"},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Pet created",
						},
					},
				},
			},
		},
	}

	// Create test servers
	specServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockSpec)
	}))
	defer specServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pets" && r.Method == "POST" {
			// Just return success for valid requests
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      "123",
				"message": "Pet created",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer apiServer.Close()

	// Create the tool
	tool := createAPIClientTool()
	ctx := &domain.ToolContext{
		Context: context.Background(),
	}

	t.Run("Valid Request With Validation", func(t *testing.T) {
		params := map[string]interface{}{
			"base_url":     apiServer.URL,
			"endpoint":     "/pets",
			"method":       "POST",
			"openapi_spec": specServer.URL,
			"body": map[string]interface{}{
				"name":    "Fluffy",
				"species": "cat",
				"age":     5,
			},
		}

		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		// Should succeed
		if !resultMap["success"].(bool) {
			t.Errorf("Expected success, got error: %v", resultMap["error_message"])
		}
	})

	t.Run("Invalid Request - Missing Required Field", func(t *testing.T) {
		params := map[string]interface{}{
			"base_url":     apiServer.URL,
			"endpoint":     "/pets",
			"method":       "POST",
			"openapi_spec": specServer.URL,
			"body": map[string]interface{}{
				"name": "Fluffy",
				// Missing required 'species' field
			},
		}

		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		// Should fail validation
		if resultMap["success"].(bool) {
			t.Error("Expected validation to fail for missing required field")
		}

		// Check for validation errors
		if resultMap["error_message"] != "Request validation failed" {
			t.Errorf("Expected validation error message, got: %v", resultMap["error_message"])
		}
	})

	t.Run("Invalid Request - Enum Violation", func(t *testing.T) {
		params := map[string]interface{}{
			"base_url":     apiServer.URL,
			"endpoint":     "/pets",
			"method":       "POST",
			"openapi_spec": specServer.URL,
			"body": map[string]interface{}{
				"name":    "Fluffy",
				"species": "elephant", // Not in enum
			},
		}

		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		// Should fail validation
		if resultMap["success"].(bool) {
			t.Error("Expected validation to fail for enum violation")
		}
	})
}

// TestAPIClientTool_RealAPIs tests against real public APIs (integration test)
func TestAPIClientTool_RealAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test with real APIs")
	}

	tool := createAPIClientTool()
	ctx := &domain.ToolContext{
		Context: context.Background(),
	}

	t.Run("PetStore API Discovery", func(t *testing.T) {
		params := map[string]interface{}{
			"base_url":            "https://petstore3.swagger.io",
			"endpoint":            "/dummy",
			"openapi_spec":        "https://petstore3.swagger.io/api/v3/openapi.json",
			"discover_operations": true,
		}

		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Discovery failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		// Check success
		if !resultMap["success"].(bool) {
			t.Error("Expected success to be true")
		}

		// Should have operations
		operations, ok := resultMap["operations"].([]OperationInfo)
		if !ok || len(operations) == 0 {
			t.Error("Expected to discover operations from PetStore API")
		}

		// Check if we found some expected operations
		var foundFindPetsByStatus bool
		for _, op := range operations {
			if op.OperationID == "findPetsByStatus" {
				foundFindPetsByStatus = true
				break
			}
		}
		if !foundFindPetsByStatus {
			t.Error("Expected to find 'findPetsByStatus' operation")
		}
	})

	t.Run("GitHub API Discovery", func(t *testing.T) {
		// Note: GitHub's OpenAPI spec is very large, so this might be slow
		params := map[string]interface{}{
			"base_url":            "https://api.github.com",
			"endpoint":            "/dummy",
			"openapi_spec":        "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json",
			"discover_operations": true,
		}

		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Logf("GitHub API discovery failed (might be rate limited): %v", err)
			t.Skip("Skipping GitHub API test")
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		// Check success
		if !resultMap["success"].(bool) {
			t.Error("Expected success to be true")
		}

		// GitHub API should have many operations
		operations, ok := resultMap["operations"].([]OperationInfo)
		if !ok || len(operations) < 100 {
			t.Errorf("Expected many operations from GitHub API, got %d", len(operations))
		}
	})
}