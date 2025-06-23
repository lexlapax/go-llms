// ABOUTME: GraphQL client implementation for api_client tool with schema parsing and validation
// ABOUTME: Provides LLM-friendly GraphQL query execution, introspection, and error handling

package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

// GraphQLClient handles GraphQL operations for the api_client tool
type GraphQLClient struct {
	endpoint   string
	httpClient *http.Client
	headers    map[string]string
	schema     *ast.Schema
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data       interface{}            `json:"data"`
	Errors     []GraphQLError         `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// NewGraphQLClient creates a new GraphQL client for executing queries, mutations, and introspection.
// It handles GraphQL operations with automatic query parsing, schema validation when available,
// comprehensive error handling, and support for variables and operation names while integrating
// seamlessly with the api_client tool for LLM-friendly GraphQL interactions.
func NewGraphQLClient(endpoint string, httpClient *http.Client, headers map[string]string) *GraphQLClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &GraphQLClient{
		endpoint:   endpoint,
		httpClient: httpClient,
		headers:    headers,
	}
}

// Execute performs a GraphQL query or mutation with full support for variables and operation names.
// It parses and validates the query against the schema if available, constructs the proper request format,
// handles authentication through headers, and returns structured responses with separate data and error fields
// for comprehensive error handling in GraphQL's unique error model.
func (c *GraphQLClient) Execute(ctx context.Context, query string, variables map[string]interface{}, operationName string) (*GraphQLResponse, error) {
	// Parse the query
	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL query: %w", err)
	}

	// Validate against schema if available
	if c.schema != nil {
		errs := validator.Validate(c.schema, doc)
		if len(errs) > 0 {
			return nil, fmt.Errorf("GraphQL validation errors: %s", formatValidationErrors(errs))
		}
	}

	// Create request
	reqBody := GraphQLRequest{
		Query:         query,
		Variables:     variables,
		OperationName: operationName,
	}

	// Encode request
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode GraphQL request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return &gqlResp, nil
}

// Introspect performs a GraphQL introspection query to discover the API schema.
// It executes a comprehensive introspection query that retrieves all types, fields, arguments,
// and their relationships, enabling dynamic schema discovery for GraphQL endpoints and
// supporting the api_client tool's GraphQL discovery mode for LLM-friendly exploration.
func (c *GraphQLClient) Introspect(ctx context.Context) (*ast.Schema, error) {
	introspectionQuery := `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					...FullType
				}
			}
		}

		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				args {
					...InputValue
				}
				type {
					...TypeRef
				}
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces {
				...TypeRef
			}
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
			possibleTypes {
				...TypeRef
			}
		}

		fragment InputValue on __InputValue {
			name
			description
			type { ...TypeRef }
			defaultValue
		}

		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`

	resp, err := c.Execute(ctx, introspectionQuery, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to execute introspection query: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("introspection query returned errors: %s", formatGraphQLErrors(resp.Errors))
	}

	// Extract __schema from response data
	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("introspection result is not a map")
	}

	_, ok = dataMap["__schema"]
	if !ok {
		return nil, fmt.Errorf("__schema not found in introspection result")
	}

	// For now, we'll store the raw introspection result
	// TODO: Implement proper schema parsing from introspection result
	schema := &ast.Schema{
		Query: &ast.Definition{
			Name: "Query",
			Kind: ast.Object,
		},
	}

	c.schema = schema
	return schema, nil
}

// GetSchema returns the cached schema
func (c *GraphQLClient) GetSchema() *ast.Schema {
	return c.schema
}

// SetSchema sets the schema for validation
func (c *GraphQLClient) SetSchema(schema *ast.Schema) {
	c.schema = schema
}

// formatValidationErrors formats GraphQL validation errors
func formatValidationErrors(errs gqlerror.List) string {
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// formatGraphQLErrors formats GraphQL response errors
func formatGraphQLErrors(errors []GraphQLError) string {
	var messages []string
	for _, err := range errors {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// GenerateLLMGuidance generates helpful guidance for GraphQL errors to assist LLMs in error resolution.
// It analyzes error messages to provide context-specific suggestions for common GraphQL issues like
// field existence, variable type mismatches, syntax errors, and authentication problems,
// making it easier for LLMs to understand and correct GraphQL query issues.
func GenerateLLMGuidance(err error, schema *ast.Schema) string {
	errMsg := err.Error()

	// Field not found errors
	if strings.Contains(errMsg, "field") && strings.Contains(errMsg, "doesn't exist") {
		return "This field doesn't exist on the type. Check the schema discovery or introspection results to see available fields."
	}

	// Variable type errors
	if strings.Contains(errMsg, "variable") && strings.Contains(errMsg, "type") {
		return "Variable type mismatch. Ensure variables match the expected types defined in the query."
	}

	// Syntax errors
	if strings.Contains(errMsg, "syntax error") || strings.Contains(errMsg, "parse") {
		return "GraphQL syntax error. Check for missing braces, parentheses, or incorrect query structure."
	}

	// Authentication errors
	if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "forbidden") {
		return "Authentication required. Ensure you've provided the correct authentication credentials."
	}

	return "GraphQL operation failed. Use discover_graphql to explore available operations and their structure."
}
