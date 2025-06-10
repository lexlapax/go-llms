// ABOUTME: Unit tests for OpenAPI specification parsing and validation
// ABOUTME: Tests JSON/YAML parsing, spec validation, and operation discovery

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestOpenAPIParser_ParseSpec tests parsing of OpenAPI specifications
func TestOpenAPIParser_ParseSpec(t *testing.T) {
	parser := NewOpenAPIParser()

	testCases := []struct {
		name        string
		spec        string
		format      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid OpenAPI 3.0 JSON",
			spec: `{
				"openapi": "3.0.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				},
				"paths": {
					"/test": {
						"get": {
							"summary": "Test endpoint",
							"responses": {
								"200": {
									"description": "Success"
								}
							}
						}
					}
				}
			}`,
			format:      "JSON",
			expectError: false,
		},
		{
			name: "Valid OpenAPI 3.1 YAML",
			spec: `openapi: "3.1.0"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /test:
    get:
      summary: "Test endpoint"
      responses:
        "200":
          description: "Success"`,
			format:      "YAML",
			expectError: false,
		},
		{
			name: "Missing openapi field",
			spec: `{
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				},
				"paths": {}
			}`,
			format:      "JSON",
			expectError: true,
			errorMsg:    "missing required field: openapi",
		},
		{
			name: "Missing info.title",
			spec: `{
				"openapi": "3.0.0",
				"info": {
					"version": "1.0.0"
				},
				"paths": {}
			}`,
			format:      "JSON",
			expectError: true,
			errorMsg:    "missing required field: info.title",
		},
		{
			name: "Missing info.version",
			spec: `{
				"openapi": "3.0.0",
				"info": {
					"title": "Test API"
				},
				"paths": {}
			}`,
			format:      "JSON",
			expectError: true,
			errorMsg:    "missing required field: info.version",
		},
		{
			name: "Unsupported OpenAPI version",
			spec: `{
				"openapi": "2.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				},
				"paths": {}
			}`,
			format:      "JSON",
			expectError: true,
			errorMsg:    "unsupported OpenAPI version",
		},
		{
			name: "Missing paths, components, and webhooks",
			spec: `{
				"openapi": "3.0.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				}
			}`,
			format:      "JSON",
			expectError: true,
			errorMsg:    "specification must contain at least one of: paths, components, or webhooks",
		},
		{
			name: "Invalid JSON",
			spec: `{
				"openapi": "3.0.0"
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				},
				"paths": {}
			}`,
			format:      "JSON",
			expectError: true,
			errorMsg:    "failed to parse spec as JSON or YAML",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := parser.ParseSpec([]byte(tc.spec), "test")

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMsg != "" && !containsString(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if spec == nil {
					t.Errorf("Expected spec but got nil")
				}
			}
		})
	}
}

// TestOpenAPIParser_FetchSpec tests fetching specs from URLs
func TestOpenAPIParser_FetchSpec(t *testing.T) {
	// Create test server with various responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid-spec":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"openapi": "3.0.0",
				"info": map[string]interface{}{
					"title":   "Test API",
					"version": "1.0.0",
				},
				"paths": map[string]interface{}{
					"/test": map[string]interface{}{
						"get": map[string]interface{}{
							"summary": "Test endpoint",
							"responses": map[string]interface{}{
								"200": map[string]interface{}{
									"description": "Success",
								},
							},
						},
					},
				},
			})
		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not found"))
		case "/invalid-spec":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"invalid": "spec"}`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	parser := NewOpenAPIParser()

	testCases := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid spec",
			path:        "/valid-spec",
			expectError: false,
		},
		{
			name:        "Not found",
			path:        "/not-found",
			expectError: true,
			errorMsg:    "failed to fetch OpenAPI spec: HTTP 404",
		},
		{
			name:        "Invalid spec",
			path:        "/invalid-spec",
			expectError: true,
			errorMsg:    "invalid OpenAPI spec",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := server.URL + tc.path
			spec, err := parser.FetchSpec(url)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMsg != "" && !containsString(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if spec == nil {
					t.Errorf("Expected spec but got nil")
				}
			}
		})
	}
}

// TestOpenAPISpec_GetOperations tests operation discovery
func TestOpenAPISpec_GetOperations(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: InfoObject{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]PathItem{
			"/users": {
				Get: &Operation{
					OperationID: "getUsers",
					Summary:     "Get all users",
					Description: "Retrieve a list of all users",
					Tags:        []string{"users"},
					Parameters: []Parameter{
						{
							Name:        "limit",
							In:          "query",
							Description: "Maximum number of users to return",
							Required:    false,
							Schema: &Schema{
								Type:    "integer",
								Minimum: func() *float64 { v := 1.0; return &v }(),
								Maximum: func() *float64 { v := 100.0; return &v }(),
							},
						},
					},
					Responses: map[string]Response{
						"200": {
							Description: "Successful response",
						},
					},
				},
				Post: &Operation{
					OperationID: "createUser",
					Summary:     "Create a user",
					Description: "Create a new user",
					Tags:        []string{"users"},
					RequestBody: &RequestBody{
						Description: "User data",
						Required:    true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"name": {
											Type: "string",
										},
										"email": {
											Type:   "string",
											Format: "email",
										},
									},
									Required: []string{"name", "email"},
								},
							},
						},
					},
					Responses: map[string]Response{
						"201": {
							Description: "User created successfully",
						},
					},
				},
			},
			"/users/{id}": {
				Get: &Operation{
					OperationID: "getUserById",
					Summary:     "Get user by ID",
					Description: "Retrieve a specific user by their ID",
					Tags:        []string{"users"},
					Parameters: []Parameter{
						{
							Name:        "id",
							In:          "path",
							Description: "User ID",
							Required:    true,
							Schema: &Schema{
								Type: "string",
							},
						},
					},
					Responses: map[string]Response{
						"200": {
							Description: "Successful response",
						},
						"404": {
							Description: "User not found",
						},
					},
				},
				Delete: &Operation{
					OperationID: "deleteUser",
					Summary:     "Delete user",
					Description: "Delete a specific user",
					Tags:        []string{"users"},
					Deprecated:  true,
					Parameters: []Parameter{
						{
							Name:        "id",
							In:          "path",
							Description: "User ID",
							Required:    true,
							Schema: &Schema{
								Type: "string",
							},
						},
					},
					Responses: map[string]Response{
						"204": {
							Description: "User deleted successfully",
						},
						"404": {
							Description: "User not found",
						},
					},
				},
			},
		},
	}

	operations := spec.GetOperations()

	// Verify we got the expected number of operations
	expectedCount := 4 // GET /users, POST /users, GET /users/{id}, DELETE /users/{id}
	if len(operations) != expectedCount {
		t.Errorf("Expected %d operations, got %d", expectedCount, len(operations))
	}

	// Verify operation details
	operationMap := make(map[string]OperationInfo)
	for _, op := range operations {
		key := op.Method + " " + op.Path
		operationMap[key] = op
	}

	// Test GET /users
	if op, exists := operationMap["GET /users"]; exists {
		if op.OperationID != "getUsers" {
			t.Errorf("Expected operationId 'getUsers', got '%s'", op.OperationID)
		}
		if op.Summary != "Get all users" {
			t.Errorf("Expected summary 'Get all users', got '%s'", op.Summary)
		}
		if len(op.Parameters) != 1 {
			t.Errorf("Expected 1 parameter, got %d", len(op.Parameters))
		}
		if len(op.Tags) != 1 || op.Tags[0] != "users" {
			t.Errorf("Expected tags ['users'], got %v", op.Tags)
		}
	} else {
		t.Errorf("GET /users operation not found")
	}

	// Test POST /users
	if op, exists := operationMap["POST /users"]; exists {
		if op.OperationID != "createUser" {
			t.Errorf("Expected operationId 'createUser', got '%s'", op.OperationID)
		}
		if op.RequestBody == nil {
			t.Errorf("Expected request body, got nil")
		} else if !op.RequestBody.Required {
			t.Errorf("Expected required request body")
		}
	} else {
		t.Errorf("POST /users operation not found")
	}

	// Test DELETE /users/{id} (deprecated)
	if op, exists := operationMap["DELETE /users/{id}"]; exists {
		if !op.Deprecated {
			t.Errorf("Expected deprecated operation")
		}
	} else {
		t.Errorf("DELETE /users/{id} operation not found")
	}
}

// TestOpenAPISpec_GetBaseURL tests base URL extraction
func TestOpenAPISpec_GetBaseURL(t *testing.T) {
	testCases := []struct {
		name        string
		spec        *OpenAPISpec
		expectedURL string
	}{
		{
			name: "Single server",
			spec: &OpenAPISpec{
				Servers: []ServerObject{
					{URL: "https://api.example.com"},
				},
			},
			expectedURL: "https://api.example.com",
		},
		{
			name: "Multiple servers",
			spec: &OpenAPISpec{
				Servers: []ServerObject{
					{URL: "https://api.example.com"},
					{URL: "https://staging-api.example.com"},
				},
			},
			expectedURL: "https://api.example.com",
		},
		{
			name:        "No servers",
			spec:        &OpenAPISpec{},
			expectedURL: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := tc.spec.GetBaseURL()
			if url != tc.expectedURL {
				t.Errorf("Expected URL '%s', got '%s'", tc.expectedURL, url)
			}
		})
	}
}

// TestOpenAPISpec_GetSecuritySchemes tests security scheme extraction
func TestOpenAPISpec_GetSecuritySchemes(t *testing.T) {
	spec := &OpenAPISpec{
		Components: &ComponentsObject{
			SecuritySchemes: map[string]SecurityScheme{
				"ApiKeyAuth": {
					Type: "apiKey",
					In:   "header",
					Name: "X-API-Key",
				},
				"BearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
				"OAuth2": {
					Type: "oauth2",
					Flows: &OAuthFlows{
						AuthorizationCode: &OAuthFlow{
							AuthorizationURL: "https://example.com/oauth/authorize",
							TokenURL:         "https://example.com/oauth/token",
							Scopes: map[string]string{
								"read":  "Read access",
								"write": "Write access",
							},
						},
					},
				},
			},
		},
	}

	schemes := spec.GetSecuritySchemes()

	if len(schemes) != 3 {
		t.Errorf("Expected 3 security schemes, got %d", len(schemes))
	}

	// Test API key auth
	if apiKey, exists := schemes["ApiKeyAuth"]; exists {
		if apiKey.Type != "apiKey" {
			t.Errorf("Expected type 'apiKey', got '%s'", apiKey.Type)
		}
		if apiKey.In != "header" {
			t.Errorf("Expected in 'header', got '%s'", apiKey.In)
		}
		if apiKey.Name != "X-API-Key" {
			t.Errorf("Expected name 'X-API-Key', got '%s'", apiKey.Name)
		}
	} else {
		t.Errorf("ApiKeyAuth scheme not found")
	}

	// Test Bearer auth
	if bearer, exists := schemes["BearerAuth"]; exists {
		if bearer.Type != "http" {
			t.Errorf("Expected type 'http', got '%s'", bearer.Type)
		}
		if bearer.Scheme != "bearer" {
			t.Errorf("Expected scheme 'bearer', got '%s'", bearer.Scheme)
		}
		if bearer.BearerFormat != "JWT" {
			t.Errorf("Expected bearerFormat 'JWT', got '%s'", bearer.BearerFormat)
		}
	} else {
		t.Errorf("BearerAuth scheme not found")
	}

	// Test OAuth2
	if oauth, exists := schemes["OAuth2"]; exists {
		if oauth.Type != "oauth2" {
			t.Errorf("Expected type 'oauth2', got '%s'", oauth.Type)
		}
		if oauth.Flows == nil || oauth.Flows.AuthorizationCode == nil {
			t.Errorf("Expected OAuth2 authorization code flow")
		}
	} else {
		t.Errorf("OAuth2 scheme not found")
	}
}

// TestOpenAPISpec_EmptyComponents tests handling of specs with no components
func TestOpenAPISpec_EmptyComponents(t *testing.T) {
	spec := &OpenAPISpec{}

	schemes := spec.GetSecuritySchemes()
	if schemes == nil {
		t.Errorf("Expected empty map, got nil")
	}
	if len(schemes) != 0 {
		t.Errorf("Expected empty map, got %d items", len(schemes))
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr || strings.Contains(s, substr)))
}
