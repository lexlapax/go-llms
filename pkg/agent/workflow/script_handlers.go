// ABOUTME: Example script handlers for demonstration and testing
// ABOUTME: Provides mock implementations for JavaScript and expression evaluation

package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// MockJavaScriptHandler is a mock JavaScript handler for testing.
// It provides a simplified JavaScript-like behavior for workflow testing
// without requiring a full JavaScript engine.
type MockJavaScriptHandler struct{}

// NewMockJavaScriptHandler creates a new mock JavaScript handler.
// This handler simulates JavaScript execution for testing purposes.
//
// Returns a new MockJavaScriptHandler instance.
func NewMockJavaScriptHandler() *MockJavaScriptHandler {
	return &MockJavaScriptHandler{}
}

// Execute implements ScriptHandler.
// It simulates JavaScript execution by pattern matching on the script content.
// Supports basic patterns like "return 'success'" and "transform" operations.
//
// Parameters:
//   - ctx: The execution context
//   - state: The current workflow state
//   - script: The JavaScript code to execute
//   - env: Environment variables available to the script
//
// Returns the new workflow state or an error.
func (h *MockJavaScriptHandler) Execute(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error) {
	// This is a mock implementation
	// In a real implementation, you would use a JavaScript engine like goja or otto

	// Create a new state
	newState := domain.NewState()

	// Copy values from previous state
	if state.State != nil {
		for k, v := range state.Values() {
			newState.Set(k, v)
		}
	}

	// Simple mock logic based on script content
	if strings.Contains(script, "return") {
		// Extract mock return value
		if strings.Contains(script, "return 'success'") {
			newState.Set("result", "success")
		} else if strings.Contains(script, "return state") {
			// Return state as-is
		} else if strings.Contains(script, "transform") {
			// Mock transformation
			if val, ok := env["state"].(map[string]interface{}); ok {
				if input, exists := val["input"]; exists {
					newState.Set("transformed", fmt.Sprintf("transformed_%v", input))
				}
			}
		}
	}

	// Create new workflow state
	newWorkflowState := &WorkflowState{
		State:    newState,
		Metadata: make(map[string]interface{}),
	}

	// Copy metadata
	if state.Metadata != nil {
		for k, v := range state.Metadata {
			newWorkflowState.Metadata[k] = v
		}
	}

	return newWorkflowState, nil
}

// Language implements ScriptHandler.
// Returns "javascript" as the supported language identifier.
func (h *MockJavaScriptHandler) Language() string {
	return "javascript"
}

// Validate implements ScriptHandler.
// Performs basic validation to ensure the script is not empty.
// In a real implementation, this would parse and validate JavaScript syntax.
//
// Parameters:
//   - script: The JavaScript code to validate
//
// Returns an error if validation fails.
func (h *MockJavaScriptHandler) Validate(script string) error {
	// Basic validation
	if script == "" {
		return fmt.Errorf("script cannot be empty")
	}
	// In a real implementation, you would parse the JavaScript
	return nil
}

// ExpressionHandler handles simple expression evaluation.
// It provides basic expression evaluation capabilities for workflow scripts,
// supporting simple assignments and string operations.
type ExpressionHandler struct{}

// NewExpressionHandler creates a new expression handler.
// This handler evaluates simple expressions like assignments.
//
// Returns a new ExpressionHandler instance.
func NewExpressionHandler() *ExpressionHandler {
	return &ExpressionHandler{}
}

// Execute implements ScriptHandler.
// It evaluates simple expressions, supporting basic assignments with
// string literals and JSON values.
//
// Parameters:
//   - ctx: The execution context
//   - state: The current workflow state
//   - script: The expression to evaluate
//   - env: Environment variables available to the expression
//
// Returns the new workflow state or an error.
func (h *ExpressionHandler) Execute(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error) {
	// This is a simplified expression evaluator
	// In a real implementation, you might use expr or cel-go

	newState := domain.NewState()

	// Copy values from previous state
	if state.State != nil {
		for k, v := range state.Values() {
			newState.Set(k, v)
		}
	}

	// Simple expression evaluation
	script = strings.TrimSpace(script)

	// Handle simple assignments
	if strings.Contains(script, "=") {
		parts := strings.SplitN(script, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Handle string literals
			if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
				value = strings.Trim(value, "'")
				newState.Set(key, value)
			} else if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
				newState.Set(key, value)
			} else {
				// Try to parse as JSON
				var parsed interface{}
				if err := json.Unmarshal([]byte(value), &parsed); err == nil {
					newState.Set(key, parsed)
				} else {
					newState.Set(key, value)
				}
			}
		}
	}

	// Create new workflow state
	newWorkflowState := &WorkflowState{
		State:    newState,
		Metadata: make(map[string]interface{}),
	}

	// Copy metadata
	if state.Metadata != nil {
		for k, v := range state.Metadata {
			newWorkflowState.Metadata[k] = v
		}
	}

	return newWorkflowState, nil
}

// Language implements ScriptHandler.
// Returns "expr" as the supported language identifier.
func (h *ExpressionHandler) Language() string {
	return "expr"
}

// Validate implements ScriptHandler.
// Ensures the expression is not empty.
//
// Parameters:
//   - script: The expression to validate
//
// Returns an error if validation fails.
func (h *ExpressionHandler) Validate(script string) error {
	if script == "" {
		return fmt.Errorf("expression cannot be empty")
	}
	return nil
}

// JSONTransformHandler handles JSON transformations.
// It applies JSON-based transformations to workflow state,
// supporting template-like variable substitution with {{variable}} syntax.
type JSONTransformHandler struct{}

// NewJSONTransformHandler creates a new JSON transform handler.
// This handler transforms state using JSON-based definitions.
//
// Returns a new JSONTransformHandler instance.
func NewJSONTransformHandler() *JSONTransformHandler {
	return &JSONTransformHandler{}
}

// Execute implements ScriptHandler.
// It applies JSON transformations to create a new state. Supports template
// variable substitution using {{variable}} syntax to reference state values.
//
// Parameters:
//   - ctx: The execution context
//   - state: The current workflow state
//   - script: JSON transformation definition
//   - env: Environment variables available for substitution
//
// Returns the transformed workflow state or an error.
func (h *JSONTransformHandler) Execute(ctx context.Context, state *WorkflowState, script string, env map[string]interface{}) (*WorkflowState, error) {
	// Parse the transformation script as JSON
	var transform map[string]interface{}
	if err := json.Unmarshal([]byte(script), &transform); err != nil {
		return nil, fmt.Errorf("invalid JSON transformation: %w", err)
	}

	newState := domain.NewState()

	// Apply transformations
	for key, value := range transform {
		switch v := value.(type) {
		case string:
			// Handle template-like replacements
			if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
				// Extract variable name
				varName := strings.TrimSpace(strings.Trim(v, "{}"))
				if stateVals, ok := env["state"].(map[string]interface{}); ok {
					if val, exists := stateVals[varName]; exists {
						newState.Set(key, val)
						continue
					}
				}
			}
			newState.Set(key, v)
		default:
			newState.Set(key, v)
		}
	}

	// Create new workflow state
	newWorkflowState := &WorkflowState{
		State:    newState,
		Metadata: make(map[string]interface{}),
	}

	// Copy metadata
	if state.Metadata != nil {
		for k, v := range state.Metadata {
			newWorkflowState.Metadata[k] = v
		}
	}

	return newWorkflowState, nil
}

// Language implements ScriptHandler.
// Returns "json-transform" as the supported language identifier.
func (h *JSONTransformHandler) Language() string {
	return "json-transform"
}

// Validate implements ScriptHandler.
// Validates that the script is valid JSON format.
//
// Parameters:
//   - script: The JSON transformation to validate
//
// Returns an error if the JSON is invalid.
func (h *JSONTransformHandler) Validate(script string) error {
	var test map[string]interface{}
	return json.Unmarshal([]byte(script), &test)
}

// RegisterDefaultHandlers registers the default/mock handlers.
// This includes JavaScript (mock), expression, and JSON transform handlers.
// These handlers are primarily for testing and demonstration purposes.
//
// Returns an error if any handler registration fails.
func RegisterDefaultHandlers() error {
	handlers := map[string]ScriptHandler{
		"javascript":     NewMockJavaScriptHandler(),
		"expr":           NewExpressionHandler(),
		"json-transform": NewJSONTransformHandler(),
	}

	for lang, handler := range handlers {
		if err := RegisterScriptHandler(lang, handler); err != nil {
			return fmt.Errorf("failed to register %s handler: %w", lang, err)
		}
	}

	return nil
}
