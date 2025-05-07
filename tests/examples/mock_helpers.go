package examples

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Custom mock provider types

// TestMockProvider is a mock provider used specifically for the MultiProvider tests
type TestMockProvider struct {
	generateFunc           func(ctx context.Context, prompt string, options ...ldomain.Option) (string, error)
	generateMessageFunc    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	generateWithSchemaFunc func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
	streamFunc             func(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error)
	streamMessageFunc      func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error)
}

func (p *TestMockProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	if p.generateFunc != nil {
		return p.generateFunc(ctx, prompt, options...)
	}
	// If generateFunc is not set, don't delegate to GenerateMessage to avoid surprises
	// Instead, return a default response
	return "Default mock response (Generate)", nil
}

func (p *TestMockProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	if p.generateMessageFunc != nil {
		return p.generateMessageFunc(ctx, messages, options...)
	}
	return ldomain.Response{Content: "Default mock response"}, nil
}

func (p *TestMockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	if p.generateWithSchemaFunc != nil {
		return p.generateWithSchemaFunc(ctx, prompt, schema, options...)
	}
	return map[string]interface{}{"result": "Default structured response"}, nil
}

func (p *TestMockProvider) Stream(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	if p.streamFunc != nil {
		return p.streamFunc(ctx, prompt, options...)
	}
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Test", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}

func (p *TestMockProvider) StreamMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	if p.streamMessageFunc != nil {
		return p.streamMessageFunc(ctx, messages, options...)
	}
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Test", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}

// CustomMockProvider is a mock LLM provider for other tests
type CustomMockProvider struct {
	generateMessageFunc    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	generateWithSchemaFunc func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
}

// Generate produces text from a prompt
func (p *CustomMockProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	resp, err := p.GenerateMessage(ctx, []ldomain.Message{{Role: ldomain.RoleUser, Content: prompt}}, options...)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateMessage generates a response to a sequence of messages
func (p *CustomMockProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	if p.generateMessageFunc != nil {
		return p.generateMessageFunc(ctx, messages, options...)
	}
	return ldomain.Response{Content: "Default mock response"}, nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *CustomMockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	if p.generateWithSchemaFunc != nil {
		return p.generateWithSchemaFunc(ctx, prompt, schema, options...)
	}
	return map[string]interface{}{"result": "Default structured response"}, nil
}

// Stream streams responses token by token
func (p *CustomMockProvider) Stream(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Test", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}

// StreamMessage streams responses token by token with messages
func (p *CustomMockProvider) StreamMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Test", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}

// MockTool is a mock implementation of the Tool interface for testing
type MockTool struct {
	name        string
	description string
	schema      *sdomain.Schema
	executor    func(ctx context.Context, params interface{}) (interface{}, error)
}

func (t MockTool) Name() string {
	return t.name
}

func (t MockTool) Description() string {
	return t.description
}

func (t MockTool) Execute(ctx context.Context, params interface{}) (interface{}, error) {
	if t.executor != nil {
		return t.executor(ctx, params)
	}
	return nil, nil
}

func (t MockTool) ParameterSchema() *sdomain.Schema {
	return t.schema
}

// Helper function to create a calculator tool for tests
func CreateCalculatorTool() domain.Tool {
	return MockTool{
		name:        "calculator",
		description: "Perform mathematical calculations",
		executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			return map[string]interface{}{
				"result": 4,
			}, nil
		},
		schema: &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	}
}

// mockStructuredProvider is a mock provider that returns structured data
type MockStructuredProvider struct {
	data interface{}
}

func (m *MockStructuredProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	return "Mock response", nil
}

func (m *MockStructuredProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	return ldomain.Response{Content: "Mock response"}, nil
}

func (m *MockStructuredProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	return m.data, nil
}

func (m *MockStructuredProvider) Stream(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Mock", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}

func (m *MockStructuredProvider) StreamMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Mock", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}