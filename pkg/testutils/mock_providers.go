// Package testutils provides testing utilities for the Go-LLMs library.
package testutils

import (
	"context"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestMockProvider is a mock provider used specifically for MultiProvider tests
type TestMockProvider struct {
	GenerateFunc           func(ctx context.Context, prompt string, options ...ldomain.Option) (string, error)
	GenerateMessageFunc    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	GenerateWithSchemaFunc func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
	StreamFunc             func(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error)
	StreamMessageFunc      func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error)
}

func (p *TestMockProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	if p.GenerateFunc != nil {
		return p.GenerateFunc(ctx, prompt, options...)
	}
	// If generateFunc is not set, don't delegate to GenerateMessage to avoid surprises
	// Instead, return a default response
	return "Default mock response (Generate)", nil
}

func (p *TestMockProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	if p.GenerateMessageFunc != nil {
		return p.GenerateMessageFunc(ctx, messages, options...)
	}
	return ldomain.Response{Content: "Default mock response"}, nil
}

func (p *TestMockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	if p.GenerateWithSchemaFunc != nil {
		return p.GenerateWithSchemaFunc(ctx, prompt, schema, options...)
	}
	return map[string]interface{}{"result": "Default structured response"}, nil
}

func (p *TestMockProvider) Stream(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	if p.StreamFunc != nil {
		return p.StreamFunc(ctx, prompt, options...)
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
	if p.StreamMessageFunc != nil {
		return p.StreamMessageFunc(ctx, messages, options...)
	}
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Test", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}

// CustomMockProvider is a mock LLM provider with customizable behavior
type CustomMockProvider struct {
	GenerateMessageFunc    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	GenerateWithSchemaFunc func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
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
	if p.GenerateMessageFunc != nil {
		return p.GenerateMessageFunc(ctx, messages, options...)
	}
	return ldomain.Response{Content: "Default mock response"}, nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *CustomMockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	if p.GenerateWithSchemaFunc != nil {
		return p.GenerateWithSchemaFunc(ctx, prompt, schema, options...)
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

// MockStructuredProvider is a mock provider that returns structured data
type MockStructuredProvider struct {
	Data interface{}
}

func (m *MockStructuredProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	return "Mock response", nil
}

func (m *MockStructuredProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	return ldomain.Response{Content: "Mock response"}, nil
}

func (m *MockStructuredProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	return m.Data, nil
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
