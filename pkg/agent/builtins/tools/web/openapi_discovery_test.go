// ABOUTME: Tests for OpenAPI operation discovery and metadata extraction
// ABOUTME: Validates parameter extraction, LLM guidance generation, and search functionality

package web

import (
	"testing"
)

// TestOperationDiscovery_EnumerateOperations tests comprehensive operation enumeration
func TestOperationDiscovery_EnumerateOperations(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	operations := discovery.EnumerateOperations()

	// Verify we got the expected number of operations
	expectedCount := 4 // GET /users, POST /users, GET /users/{id}, DELETE /users/{id}
	if len(operations) != expectedCount {
		t.Errorf("Expected %d operations, got %d", expectedCount, len(operations))
	}

	// Find and test GET /users operation
	var getUsersOp *EnhancedOperationInfo
	for _, op := range operations {
		if op.Method == "GET" && op.Path == "/users" {
			getUsersOp = &op
			break
		}
	}

	if getUsersOp == nil {
		t.Fatal("GET /users operation not found")
	}

	// Test basic metadata
	if getUsersOp.OperationID != "getUsers" {
		t.Errorf("Expected operationId 'getUsers', got '%s'", getUsersOp.OperationID)
	}

	if getUsersOp.Summary != "Get all users" {
		t.Errorf("Expected summary 'Get all users', got '%s'", getUsersOp.Summary)
	}

	if len(getUsersOp.Tags) != 1 || getUsersOp.Tags[0] != "users" {
		t.Errorf("Expected tags ['users'], got %v", getUsersOp.Tags)
	}

	// Test query parameters
	if len(getUsersOp.QueryParameters) != 1 {
		t.Errorf("Expected 1 query parameter, got %d", len(getUsersOp.QueryParameters))
	} else {
		limitParam := getUsersOp.QueryParameters[0]
		if limitParam.Name != "limit" {
			t.Errorf("Expected parameter name 'limit', got '%s'", limitParam.Name)
		}
		if limitParam.Schema.Type != "integer" {
			t.Errorf("Expected parameter type 'integer', got '%s'", limitParam.Schema.Type)
		}
		if limitParam.Schema.Minimum == nil || *limitParam.Schema.Minimum != 1.0 {
			t.Errorf("Expected minimum value 1, got %v", limitParam.Schema.Minimum)
		}
		if limitParam.Schema.Maximum == nil || *limitParam.Schema.Maximum != 100.0 {
			t.Errorf("Expected maximum value 100, got %v", limitParam.Schema.Maximum)
		}
	}

	// Test responses
	if len(getUsersOp.ResponseInfo) == 0 {
		t.Error("Expected response information")
	}

	if response, exists := getUsersOp.ResponseInfo["200"]; exists {
		if response.Description != "Successful response" {
			t.Errorf("Expected response description 'Successful response', got '%s'", response.Description)
		}
	} else {
		t.Error("Expected 200 response not found")
	}

	// Test LLM guidance generation
	guidance := getUsersOp.LLMGuidance
	if guidance.UsageInstructions == "" {
		t.Error("Expected usage instructions to be generated")
	}

	if len(guidance.ParameterGuidance) == 0 {
		t.Error("Expected parameter guidance to be generated")
	}

	if guidance.ParameterGuidance["limit"] == "" {
		t.Error("Expected guidance for 'limit' parameter")
	}

	if len(guidance.Examples) == 0 {
		t.Error("Expected operation examples to be generated")
	}
}

// TestOperationDiscovery_PostOperation tests POST operation with request body
func TestOperationDiscovery_PostOperation(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	operations := discovery.EnumerateOperations()

	// Find POST /users operation
	var postUsersOp *EnhancedOperationInfo
	for _, op := range operations {
		if op.Method == "POST" && op.Path == "/users" {
			postUsersOp = &op
			break
		}
	}

	if postUsersOp == nil {
		t.Fatal("POST /users operation not found")
	}

	// Test request body information
	if postUsersOp.RequestBodyInfo == nil {
		t.Fatal("Expected request body info")
	}

	if !postUsersOp.RequestBodyInfo.Required {
		t.Error("Expected required request body")
	}

	if postUsersOp.RequestBodyInfo.Description != "User data" {
		t.Errorf("Expected description 'User data', got '%s'", postUsersOp.RequestBodyInfo.Description)
	}

	if len(postUsersOp.RequestBodyInfo.ContentTypes) == 0 {
		t.Error("Expected content types")
	}

	// Test request body schema
	schema := postUsersOp.RequestBodyInfo.Schema
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}

	if nameSchema, exists := schema.Properties["name"]; exists {
		if nameSchema.Type != "string" {
			t.Errorf("Expected name type 'string', got '%s'", nameSchema.Type)
		}
	} else {
		t.Error("Expected 'name' property not found")
	}

	if emailSchema, exists := schema.Properties["email"]; exists {
		if emailSchema.Type != "string" {
			t.Errorf("Expected email type 'string', got '%s'", emailSchema.Type)
		}
		if emailSchema.Format != "email" {
			t.Errorf("Expected email format 'email', got '%s'", emailSchema.Format)
		}
	} else {
		t.Error("Expected 'email' property not found")
	}

	if len(schema.Required) != 2 {
		t.Errorf("Expected 2 required properties, got %d", len(schema.Required))
	}

	// Test LLM guidance for POST
	guidance := postUsersOp.LLMGuidance
	if len(guidance.Constraints) == 0 {
		t.Error("Expected constraints for POST operation")
	}

	hasRequestBodyConstraint := false
	for _, constraint := range guidance.Constraints {
		if constraint == "Request body is required" {
			hasRequestBodyConstraint = true
			break
		}
	}
	if !hasRequestBodyConstraint {
		t.Error("Expected request body constraint")
	}
}

// TestOperationDiscovery_PathParameters tests path parameter handling
func TestOperationDiscovery_PathParameters(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	operations := discovery.EnumerateOperations()

	// Find GET /users/{id} operation
	var getUserOp *EnhancedOperationInfo
	for _, op := range operations {
		if op.Method == "GET" && op.Path == "/users/{id}" {
			getUserOp = &op
			break
		}
	}

	if getUserOp == nil {
		t.Fatal("GET /users/{id} operation not found")
	}

	// Test path parameters
	if len(getUserOp.PathParameters) != 1 {
		t.Errorf("Expected 1 path parameter, got %d", len(getUserOp.PathParameters))
	} else {
		idParam := getUserOp.PathParameters[0]
		if idParam.Name != "id" {
			t.Errorf("Expected parameter name 'id', got '%s'", idParam.Name)
		}
		if idParam.In != "path" {
			t.Errorf("Expected parameter in 'path', got '%s'", idParam.In)
		}
		if !idParam.Required {
			t.Error("Expected path parameter to be required")
		}
		if idParam.Schema.Type != "string" {
			t.Errorf("Expected parameter type 'string', got '%s'", idParam.Schema.Type)
		}
	}
}

// TestOperationDiscovery_DeprecatedOperation tests deprecated operation handling
func TestOperationDiscovery_DeprecatedOperation(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	operations := discovery.EnumerateOperations()

	// Find DELETE /users/{id} operation (marked as deprecated)
	var deleteUserOp *EnhancedOperationInfo
	for _, op := range operations {
		if op.Method == "DELETE" && op.Path == "/users/{id}" {
			deleteUserOp = &op
			break
		}
	}

	if deleteUserOp == nil {
		t.Fatal("DELETE /users/{id} operation not found")
	}

	// Test deprecated flag
	if !deleteUserOp.Deprecated {
		t.Error("Expected operation to be marked as deprecated")
	}

	// Test deprecation constraint
	guidance := deleteUserOp.LLMGuidance
	hasDeprecationConstraint := false
	for _, constraint := range guidance.Constraints {
		if constraint == "This operation is deprecated and may be removed in future versions" {
			hasDeprecationConstraint = true
			break
		}
	}
	if !hasDeprecationConstraint {
		t.Error("Expected deprecation constraint")
	}
}

// TestOperationDiscovery_FindOperationByID tests operation lookup by ID
func TestOperationDiscovery_FindOperationByID(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	// Test finding existing operation
	op := discovery.FindOperationByID("getUsers")
	if op == nil {
		t.Fatal("Expected to find operation with ID 'getUsers'")
	}

	if op.Method != "GET" || op.Path != "/users" {
		t.Errorf("Expected GET /users, got %s %s", op.Method, op.Path)
	}

	// Test finding non-existent operation
	op = discovery.FindOperationByID("nonexistent")
	if op != nil {
		t.Error("Expected nil for non-existent operation ID")
	}
}

// TestOperationDiscovery_FindOperationsByTag tests operation lookup by tag
func TestOperationDiscovery_FindOperationsByTag(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	// Test finding operations by tag
	ops := discovery.FindOperationsByTag("users")
	if len(ops) != 4 {
		t.Errorf("Expected 4 operations with tag 'users', got %d", len(ops))
	}

	// Test finding operations by non-existent tag
	ops = discovery.FindOperationsByTag("nonexistent")
	if len(ops) != 0 {
		t.Errorf("Expected 0 operations with non-existent tag, got %d", len(ops))
	}
}

// TestOperationDiscovery_FindOperationsByPath tests operation lookup by path pattern
func TestOperationDiscovery_FindOperationsByPath(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	// Test finding operations by path pattern
	ops := discovery.FindOperationsByPath("/users")
	if len(ops) != 4 {
		t.Errorf("Expected 4 operations matching '/users', got %d", len(ops))
	}

	// Test finding operations by specific path pattern
	ops = discovery.FindOperationsByPath("{id}")
	if len(ops) != 2 {
		t.Errorf("Expected 2 operations matching '{id}', got %d", len(ops))
	}

	// Test finding operations by non-matching pattern
	ops = discovery.FindOperationsByPath("/nonexistent")
	if len(ops) != 0 {
		t.Errorf("Expected 0 operations matching non-existent pattern, got %d", len(ops))
	}
}

// TestOperationDiscovery_GetPathToOperationMap tests path-to-operation mapping
func TestOperationDiscovery_GetPathToOperationMap(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	mapping := discovery.GetPathToOperationMap()

	expectedKeys := []string{
		"GET /users",
		"POST /users",
		"GET /users/{id}",
		"DELETE /users/{id}",
	}

	if len(mapping) != len(expectedKeys) {
		t.Errorf("Expected %d mappings, got %d", len(expectedKeys), len(mapping))
	}

	for _, key := range expectedKeys {
		if _, exists := mapping[key]; !exists {
			t.Errorf("Expected mapping for '%s' not found", key)
		}
	}

	// Test specific mapping
	if op, exists := mapping["GET /users"]; exists {
		if op.OperationID != "getUsers" {
			t.Errorf("Expected operationId 'getUsers', got '%s'", op.OperationID)
		}
	} else {
		t.Error("Expected mapping for 'GET /users' not found")
	}
}

// TestOperationDiscovery_ExampleGeneration tests automatic example generation
func TestOperationDiscovery_ExampleGeneration(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	operations := discovery.EnumerateOperations()

	// Find POST /users operation
	var postUsersOp *EnhancedOperationInfo
	for _, op := range operations {
		if op.Method == "POST" && op.Path == "/users" {
			postUsersOp = &op
			break
		}
	}

	if postUsersOp == nil {
		t.Fatal("POST /users operation not found")
	}

	// Test example generation
	guidance := postUsersOp.LLMGuidance
	if len(guidance.Examples) == 0 {
		t.Fatal("Expected operation examples")
	}

	example := guidance.Examples[0]
	if example.Name != "Basic Usage" {
		t.Errorf("Expected example name 'Basic Usage', got '%s'", example.Name)
	}

	if example.Description == "" {
		t.Error("Expected example description")
	}

	if example.RequestBody == nil {
		t.Error("Expected example request body")
	}

	// Verify generated request body structure
	if reqBody, ok := example.RequestBody.(map[string]interface{}); ok {
		if _, hasName := reqBody["name"]; !hasName {
			t.Error("Expected generated request body to have 'name' field")
		}
		if _, hasEmail := reqBody["email"]; !hasEmail {
			t.Error("Expected generated request body to have 'email' field")
		}
	} else {
		t.Error("Expected request body to be a map")
	}
}

// TestOperationDiscovery_ErrorGuidance tests error guidance generation
func TestOperationDiscovery_ErrorGuidance(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	operations := discovery.EnumerateOperations()

	// Find GET /users/{id} operation
	var getUserOp *EnhancedOperationInfo
	for _, op := range operations {
		if op.Method == "GET" && op.Path == "/users/{id}" {
			getUserOp = &op
			break
		}
	}

	if getUserOp == nil {
		t.Fatal("GET /users/{id} operation not found")
	}

	// Test error guidance
	guidance := getUserOp.LLMGuidance
	if len(guidance.ErrorGuidance) == 0 {
		t.Error("Expected error guidance")
	}

	if guidance404, exists := guidance.ErrorGuidance["404"]; exists {
		if !containsString(guidance404, "Not Found") {
			t.Errorf("Expected 404 guidance to contain 'Not Found', got '%s'", guidance404)
		}
	} else {
		t.Error("Expected 404 error guidance")
	}
}

// TestOperationDiscovery_SchemaValidation tests schema validation integration
func TestOperationDiscovery_SchemaValidation(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	// Test request body validation for POST /users
	requestBody := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	result, err := discovery.ValidateRequestBody("createUser", requestBody)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid request body, got errors: %v", result.Errors)
	}

	// Test invalid request body
	invalidRequestBody := map[string]interface{}{
		"name": "John Doe",
		// Missing required email field
	}

	result, err = discovery.ValidateRequestBody("createUser", invalidRequestBody)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid request body but validation passed")
	}
}

// TestOperationDiscovery_ParameterValidation tests parameter validation
func TestOperationDiscovery_ParameterValidation(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)


	// Test valid parameters for GET /users
	parameters := map[string]interface{}{
		"limit": 50,
	}

	results, err := discovery.ValidateParameters("getUsers", parameters)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected parameter validation results")
	}

	if limitResult, exists := results["limit"]; exists {
		if !limitResult.Valid {
			t.Errorf("Expected valid limit parameter, got errors: %v", limitResult.Errors)
		}
	} else {
		t.Error("Expected limit parameter result")
	}

	// Test invalid parameters (out of range)
	invalidParameters := map[string]interface{}{
		"limit": 200, // Exceeds maximum of 100
	}

	results, err = discovery.ValidateParameters("getUsers", invalidParameters)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if limitResult, exists := results["limit"]; exists {
		if limitResult.Valid {
			t.Errorf("Expected invalid limit parameter but validation passed. Errors: %v", limitResult.Errors)
		}
	} else {
		t.Error("Expected limit parameter result for invalid validation")
	}
}

// TestOperationDiscovery_SchemaOptimization tests schema optimization
func TestOperationDiscovery_SchemaOptimization(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	// Test schema optimization
	err := discovery.OptimizeSchema("getUsers")
	if err != nil {
		t.Errorf("Unexpected error during schema optimization: %v", err)
	}

	// Test optimization for non-existent operation
	err = discovery.OptimizeSchema("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent operation")
	}
}

// TestOperationDiscovery_ParameterCoercion tests parameter value coercion
func TestOperationDiscovery_ParameterCoercion(t *testing.T) {
	spec := createTestOpenAPISpec()
	discovery := NewOperationDiscovery(spec)

	operations := discovery.EnumerateOperations()
	
	// Find GET /users operation
	var getUsersOp *EnhancedOperationInfo
	for _, op := range operations {
		if op.Method == "GET" && op.Path == "/users" {
			getUsersOp = &op
			break
		}
	}

	if getUsersOp == nil {
		t.Fatal("GET /users operation not found")
	}

	// Test coercing string to integer for limit parameter
	if len(getUsersOp.QueryParameters) > 0 {
		limitParam := getUsersOp.QueryParameters[0]
		
		// Coerce string "50" to integer
		coercedValue, err := discovery.CoerceParameterValue(limitParam, "50")
		if err != nil {
			t.Errorf("Unexpected error during coercion: %v", err)
		}

		// Should be converted to int64
		if _, ok := coercedValue.(int64); !ok {
			t.Errorf("Expected coerced value to be int64, got %T", coercedValue)
		}
	}
}

// createTestOpenAPISpec creates a test OpenAPI specification for testing
func createTestOpenAPISpec() *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: InfoObject{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Servers: []ServerObject{
			{URL: "https://api.example.com"},
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
}