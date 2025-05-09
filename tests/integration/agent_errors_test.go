package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestAgentErrors tests error handling for agent workflows
func TestAgentErrors(t *testing.T) {
	// This test file focuses on agent error handling
	// It tests error conditions that might occur during agent execution,
	// including provider errors, tool execution errors, and timeout errors.

	// Test agent with failing provider
	t.Run("FailingProvider", func(t *testing.T) {
		// Create a failing provider that always returns errors
		provider := &mockFailingProvider{
			err: errors.New("provider error"),
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Run the agent with the failing provider
		_, err := agent.Run(context.Background(), "Test input")
		if err == nil {
			t.Errorf("Expected error but got nil")
		}
	})

	// Test agent with failing tool
	t.Run("FailingTool", func(t *testing.T) {
		// Create a provider that calls a tool but the tool fails
		provider := &mockToolCallingProvider{
			toolName:     "failing_tool",
			toolParams:   map[string]interface{}{"param": "value"},
			expectError:  true,
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Add a failing tool
		agent.AddTool(tools.NewTool(
			"failing_tool",
			"A tool that always fails",
			func(params struct {
				Param string `json:"param"`
			}) (interface{}, error) {
				return nil, errors.New("tool execution error")
			},
			&schemaDomain.Schema{
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"param": {Type: "string"},
				},
				Required: []string{"param"},
			},
		))

		// Run the agent with the tool call
		_, err := agent.Run(context.Background(), "Test input")
		if err == nil {
			t.Errorf("Expected error but got nil")
		}
	})

	// Test agent with invalid tool name
	t.Run("InvalidToolName", func(t *testing.T) {
		// Create a provider that calls a non-existent tool
		provider := &mockToolCallingProvider{
			toolName:     "nonexistent_tool",
			toolParams:   map[string]interface{}{"param": "value"},
			expectError:  true,
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Run the agent with the invalid tool call
		_, err := agent.Run(context.Background(), "Test input")
		if err == nil {
			t.Errorf("Expected error but got nil")
		}
	})

	// Test agent with invalid tool parameters
	t.Run("InvalidToolParams", func(t *testing.T) {
		// Create a provider that calls a tool with invalid parameters
		provider := &mockToolCallingProvider{
			toolName:     "test_tool",
			toolParams:   map[string]interface{}{"wrong_param": "value"}, // Wrong parameter name
			expectError:  true,
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Add a tool with specific parameters
		agent.AddTool(tools.NewTool(
			"test_tool",
			"A test tool",
			func(params struct {
				RequiredParam string `json:"required_param"`
			}) (interface{}, error) {
				return params.RequiredParam, nil
			},
			&schemaDomain.Schema{
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"required_param": {Type: "string"},
				},
				Required: []string{"required_param"},
			},
		))

		// Run the agent with the invalid tool parameters
		_, err := agent.Run(context.Background(), "Test input")
		if err == nil {
			t.Errorf("Expected error but got nil")
		}
	})

	// Test agent with invalid schema
	t.Run("InvalidSchema", func(t *testing.T) {
		// Create a provider that returns invalid data for the schema
		provider := &mockInvalidSchemaProvider{
			expectError: true,
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Define a schema for validation
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"result": {Type: "number"},
			},
			Required: []string{"result"},
		}

		// Run the agent with the schema validation
		_, err := agent.RunWithSchema(context.Background(), "Test input", schema)
		if err == nil {
			t.Errorf("Expected schema validation error but got nil")
		}
	})

	// Test agent with timeout
	t.Run("Timeout", func(t *testing.T) {
		// Create a provider that takes too long to respond
		provider := &mockDelayProvider{
			delay: 300 * time.Millisecond,
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Create a context with a shorter timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Run the agent with the timeout context
		_, err := agent.Run(ctx, "Test input")
		if err == nil {
			t.Errorf("Expected timeout error but got nil")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected deadline exceeded error, got: %v", err)
		}
	})

	// Test agent with hooks that error
	t.Run("ErrorHooks", func(t *testing.T) {
		// Create a mock provider
		provider := &mockProvider{
			expectError: true,
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Add an error hook
		agent.WithHook(&errorHook{})

		// Run the agent with the error hook
		_, err := agent.Run(context.Background(), "Test input")
		if err == nil {
			t.Errorf("Expected hook error but got nil")
		}
	})

	// Test with custom runWithTimeout function
	t.Run("CustomRunWithTimeout", func(t *testing.T) {
		// Create a provider that takes too long to respond
		provider := &mockDelayProvider{
			delay: 300 * time.Millisecond,
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Run the agent with a short timeout
		_, err := runWithTimeout(agent, "Test input", 100*time.Millisecond)
		if err == nil {
			t.Errorf("Expected timeout error but got nil")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected deadline exceeded error, got: %v", err)
		}
	})

	// Test agent with a tool that returns invalid JSON
	t.Run("InvalidToolResponse", func(t *testing.T) {
		// Create a provider that generates a valid tool call
		provider := &mockToolCallingProvider{
			toolName:     "invalid_response_tool",
			toolParams:   map[string]interface{}{"param": "value"},
			expectError:  true,
		}

		// Create an agent
		agent := workflow.NewAgent(provider)

		// Add a tool that returns a value that can't be serialized to JSON
		agent.AddTool(tools.NewTool(
			"invalid_response_tool",
			"A tool that returns an invalid response",
			func(params struct {
				Param string `json:"param"`
			}) (interface{}, error) {
				// Return a function, which can't be serialized to JSON
				return func() {}, nil
			},
			&schemaDomain.Schema{
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"param": {Type: "string"},
				},
				Required: []string{"param"},
			},
		))

		// Run the agent with the tool call
		_, err := agent.Run(context.Background(), "Test input")
		if err == nil {
			t.Errorf("Expected error for invalid tool response but got nil")
		}
	})
}

// Mock provider that always fails
type mockFailingProvider struct {
	err error
}

func (p *mockFailingProvider) Generate(ctx context.Context, prompt string, options ...llmDomain.Option) (string, error) {
	return "", p.err
}

func (p *mockFailingProvider) GenerateMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.Response, error) {
	return llmDomain.Response{}, p.err
}

func (p *mockFailingProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...llmDomain.Option) (interface{}, error) {
	return nil, p.err
}

func (p *mockFailingProvider) Stream(ctx context.Context, prompt string, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, p.err
}

func (p *mockFailingProvider) StreamMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, p.err
}

// Mock provider that calls a specific tool
type mockToolCallingProvider struct {
	toolName     string
	toolParams   map[string]interface{}
	expectError  bool
	generateResp string
}

func (p *mockToolCallingProvider) Generate(ctx context.Context, prompt string, options ...llmDomain.Option) (string, error) {
	// Return error if expected
	if p.expectError {
		return "", errors.New("mock provider error")
	}
	
	// If custom response is set, return it
	if p.generateResp != "" {
		return p.generateResp, nil
	}
	
	// Create a tool call in JSON format
	return fmt.Sprintf(`{"tool": "%s", "params": %s}`, p.toolName, mapToJSON(p.toolParams)), nil
}

func (p *mockToolCallingProvider) GenerateMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.Response, error) {
	result, err := p.Generate(ctx, "message", options...)
	return llmDomain.Response{Content: result}, err
}

func (p *mockToolCallingProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...llmDomain.Option) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (p *mockToolCallingProvider) Stream(ctx context.Context, prompt string, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

func (p *mockToolCallingProvider) StreamMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

// Convert a map to a JSON string
func mapToJSON(m map[string]interface{}) string {
	jsonStr := "{"
	first := true
	for k, v := range m {
		if !first {
			jsonStr += ","
		}
		first = false
		jsonStr += fmt.Sprintf(`"%s":`, k)
		switch v := v.(type) {
		case string:
			jsonStr += fmt.Sprintf(`"%s"`, v)
		case int, float64:
			jsonStr += fmt.Sprintf(`%v`, v)
		case bool:
			jsonStr += fmt.Sprintf(`%v`, v)
		default:
			jsonStr += `"invalid"`
		}
	}
	jsonStr += "}"
	return jsonStr
}

// Mock provider that returns invalid schema data
type mockInvalidSchemaProvider struct{
	expectError bool
}

func (p *mockInvalidSchemaProvider) Generate(ctx context.Context, prompt string, options ...llmDomain.Option) (string, error) {
	if p.expectError {
		return "", errors.New("mock provider error")
	}
	return "text response", nil
}

func (p *mockInvalidSchemaProvider) GenerateMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.Response, error) {
	if p.expectError {
		return llmDomain.Response{}, errors.New("mock provider error")
	}
	return llmDomain.Response{Content: "text response"}, nil
}

func (p *mockInvalidSchemaProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...llmDomain.Option) (interface{}, error) {
	if p.expectError {
		return nil, errors.New("mock provider error")
	}
	// Return string for a field that should be a number
	return map[string]interface{}{
		"result": "not a number",
	}, nil
}

func (p *mockInvalidSchemaProvider) Stream(ctx context.Context, prompt string, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

func (p *mockInvalidSchemaProvider) StreamMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

// Mock provider that delays responses
type mockDelayProvider struct {
	delay time.Duration
}

func (p *mockDelayProvider) Generate(ctx context.Context, prompt string, options ...llmDomain.Option) (string, error) {
	select {
	case <-time.After(p.delay):
		return "delayed response", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (p *mockDelayProvider) GenerateMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.Response, error) {
	result, err := p.Generate(ctx, "message", options...)
	return llmDomain.Response{Content: result}, err
}

func (p *mockDelayProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...llmDomain.Option) (interface{}, error) {
	select {
	case <-time.After(p.delay):
		return map[string]interface{}{"result": 42}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *mockDelayProvider) Stream(ctx context.Context, prompt string, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

func (p *mockDelayProvider) StreamMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

// Standard mock provider
type mockProvider struct{
	expectError bool
}

func (p *mockProvider) Generate(ctx context.Context, prompt string, options ...llmDomain.Option) (string, error) {
	if p.expectError {
		return "", errors.New("mock provider error")
	}
	return "mock response", nil
}

func (p *mockProvider) GenerateMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.Response, error) {
	if p.expectError {
		return llmDomain.Response{}, errors.New("mock provider error")
	}
	return llmDomain.Response{Content: "mock response"}, nil
}

func (p *mockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...llmDomain.Option) (interface{}, error) {
	if p.expectError {
		return nil, errors.New("mock provider error")
	}
	return map[string]interface{}{"result": 42}, nil
}

func (p *mockProvider) Stream(ctx context.Context, prompt string, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

func (p *mockProvider) StreamMessage(ctx context.Context, messages []llmDomain.Message, options ...llmDomain.Option) (llmDomain.ResponseStream, error) {
	return nil, errors.New("not implemented")
}

// Hook that causes errors
type errorHook struct{}

func (h *errorHook) BeforeGenerate(ctx context.Context, messages []llmDomain.Message) {
	// No error here
}

func (h *errorHook) AfterGenerate(ctx context.Context, response llmDomain.Response, err error) {
	// No error here
}

func (h *errorHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	panic("hook error")
}

func (h *errorHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	// No error here
}

// Helper function to run an agent with a timeout
func runWithTimeout(agent *workflow.DefaultAgent, input string, timeout time.Duration) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return agent.Run(ctx, input)
}