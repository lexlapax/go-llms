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

// MockJavaScriptHandler is a mock JavaScript handler for testing
type MockJavaScriptHandler struct{}

// NewMockJavaScriptHandler creates a new mock JavaScript handler
func NewMockJavaScriptHandler() *MockJavaScriptHandler {
	return &MockJavaScriptHandler{}
}

// Execute implements ScriptHandler
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

// Language implements ScriptHandler
func (h *MockJavaScriptHandler) Language() string {
	return "javascript"
}

// Validate implements ScriptHandler
func (h *MockJavaScriptHandler) Validate(script string) error {
	// Basic validation
	if script == "" {
		return fmt.Errorf("script cannot be empty")
	}
	// In a real implementation, you would parse the JavaScript
	return nil
}

// ExpressionHandler handles simple expression evaluation
type ExpressionHandler struct{}

// NewExpressionHandler creates a new expression handler
func NewExpressionHandler() *ExpressionHandler {
	return &ExpressionHandler{}
}

// Execute implements ScriptHandler
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

// Language implements ScriptHandler
func (h *ExpressionHandler) Language() string {
	return "expr"
}

// Validate implements ScriptHandler
func (h *ExpressionHandler) Validate(script string) error {
	if script == "" {
		return fmt.Errorf("expression cannot be empty")
	}
	return nil
}

// JSONTransformHandler handles JSON transformations
type JSONTransformHandler struct{}

// NewJSONTransformHandler creates a new JSON transform handler
func NewJSONTransformHandler() *JSONTransformHandler {
	return &JSONTransformHandler{}
}

// Execute implements ScriptHandler
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

// Language implements ScriptHandler
func (h *JSONTransformHandler) Language() string {
	return "json-transform"
}

// Validate implements ScriptHandler
func (h *JSONTransformHandler) Validate(script string) error {
	var test map[string]interface{}
	return json.Unmarshal([]byte(script), &test)
}

// RegisterDefaultHandlers registers the default/mock handlers
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
