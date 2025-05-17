package multi_provider

import (
	"context"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// MockProvider is a local mock implementation of the Provider interface for testing
type MockProvider struct {
	GenerateFunc           func(ctx context.Context, prompt string, options ...ldomain.Option) (string, error)
	GenerateMessageFunc    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	GenerateWithSchemaFunc func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
	StreamFunc             func(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error)
	StreamMessageFunc      func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error)
}

func (p *MockProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	if p.GenerateFunc != nil {
		return p.GenerateFunc(ctx, prompt, options...)
	}

	// Create a text message from the prompt and use GenerateMessage
	textMsg := ldomain.NewTextMessage(ldomain.RoleUser, prompt)
	resp, err := p.GenerateMessage(ctx, []ldomain.Message{textMsg}, options...)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (p *MockProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	if p.GenerateMessageFunc != nil {
		return p.GenerateMessageFunc(ctx, messages, options...)
	}
	return ldomain.Response{Content: "Default mock response"}, nil
}

func (p *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	if p.GenerateWithSchemaFunc != nil {
		return p.GenerateWithSchemaFunc(ctx, prompt, schema, options...)
	}
	return map[string]interface{}{"result": "Default structured response"}, nil
}

func (p *MockProvider) Stream(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
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

func (p *MockProvider) StreamMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
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
