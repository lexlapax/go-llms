// ABOUTME: Enhanced mock provider implementations with pattern-based responses and call tracking
// ABOUTME: Provides deterministic responses, failure simulation, and comprehensive test support
// Package mocks provides mock implementations of LLM providers for testing purposes.
// It includes pattern-based response matching, call tracking, and configurable behaviors
// to facilitate comprehensive testing of LLM-dependent code.
package mocks

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Response represents a mock response configuration
type Response struct {
	// Content is the response text to return
	Content string
	// Error is the error to return (if any)
	Error error
	// Delay simulates response latency
	Delay time.Duration
	// Metadata contains additional response information
	Metadata map[string]interface{}
}

// ProviderCall represents a recorded provider call
type ProviderCall struct {
	// Method is the name of the method called
	Method string
	// Prompt is the text prompt (for non-message calls)
	Prompt string
	// Messages contains the message history (for message-based calls)
	Messages []ldomain.Message
	// Schema is the output schema (for structured calls)
	Schema *sdomain.Schema
	// Options contains the call options
	Options []ldomain.Option
	// Response is the returned response content
	Response string
	// Error is the returned error (if any)
	Error error
	// Timestamp is when the call was made
	Timestamp time.Time
	// Duration is how long the call took
	Duration time.Duration
}

// MockProvider is an enhanced mock provider with pattern-based responses and call tracking
type MockProvider struct {
	// Configuration
	ProviderName    string
	ResponsePattern map[string]Response // Pattern-based responses
	DefaultResponse Response            // Default response when no pattern matches

	// Behavior hooks
	OnGenerate           func(ctx context.Context, prompt string, options ...ldomain.Option) (string, error)
	OnGenerateMessage    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	OnGenerateWithSchema func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
	OnStream             func(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error)
	OnStreamMessage      func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error)

	// State
	mu          sync.RWMutex
	callHistory []ProviderCall
	callCount   int
	lastError   error
}

// NewMockProvider creates a new mock provider with default configuration
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		ProviderName:    name,
		ResponsePattern: make(map[string]Response),
		DefaultResponse: Response{Content: "Default mock response"},
		callHistory:     make([]ProviderCall, 0),
	}
}

// WithPatternResponse adds a pattern-based response
func (p *MockProvider) WithPatternResponse(pattern string, response Response) *MockProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ResponsePattern[pattern] = response
	return p
}

// WithDefaultResponse sets the default response
func (p *MockProvider) WithDefaultResponse(response Response) *MockProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.DefaultResponse = response
	return p
}

// Generate produces text from a prompt
func (p *MockProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	start := time.Now()

	// Check behavior hook first
	if p.OnGenerate != nil {
		response, err := p.OnGenerate(ctx, prompt, options...)
		p.recordCall("Generate", prompt, nil, nil, options, response, err, start)
		return response, err
	}

	// Find matching pattern
	response := p.findMatchingResponse(prompt)

	// Apply delay if configured
	if response.Delay > 0 {
		select {
		case <-time.After(response.Delay):
		case <-ctx.Done():
			err := ctx.Err()
			p.recordCall("Generate", prompt, nil, nil, options, "", err, start)
			return "", err
		}
	}

	p.recordCall("Generate", prompt, nil, nil, options, response.Content, response.Error, start)
	return response.Content, response.Error
}

// GenerateMessage generates a response to a sequence of messages
func (p *MockProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	start := time.Now()

	// Check behavior hook first
	if p.OnGenerateMessage != nil {
		response, err := p.OnGenerateMessage(ctx, messages, options...)
		p.recordCall("GenerateMessage", "", messages, nil, options, response.Content, err, start)
		return response, err
	}

	// Convert messages to prompt for pattern matching
	prompt := ""
	if len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		// Extract text content from the last message
		for _, part := range lastMsg.Content {
			if part.Type == ldomain.ContentTypeText {
				prompt = part.Text
				break
			}
		}
	}

	response := p.findMatchingResponse(prompt)

	// Apply delay if configured
	if response.Delay > 0 {
		select {
		case <-time.After(response.Delay):
		case <-ctx.Done():
			err := ctx.Err()
			p.recordCall("GenerateMessage", prompt, messages, nil, options, "", err, start)
			return ldomain.Response{}, err
		}
	}

	result := ldomain.Response{
		Content: response.Content,
	}

	p.recordCall("GenerateMessage", prompt, messages, nil, options, response.Content, response.Error, start)
	return result, response.Error
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	start := time.Now()

	// Check behavior hook first
	if p.OnGenerateWithSchema != nil {
		response, err := p.OnGenerateWithSchema(ctx, prompt, schema, options...)
		p.recordCall("GenerateWithSchema", prompt, nil, schema, options, fmt.Sprintf("%v", response), err, start)
		return response, err
	}

	// Default structured response based on schema
	response := map[string]interface{}{
		"result": "Mock structured response",
		"schema": schema.Type,
	}

	p.recordCall("GenerateWithSchema", prompt, nil, schema, options, fmt.Sprintf("%v", response), nil, start)
	return response, nil
}

// Stream streams responses token by token
func (p *MockProvider) Stream(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	start := time.Now()

	// Check behavior hook first
	if p.OnStream != nil {
		stream, err := p.OnStream(ctx, prompt, options...)
		p.recordCall("Stream", prompt, nil, nil, options, "stream", err, start)
		return stream, err
	}

	// Find matching response
	response := p.findMatchingResponse(prompt)

	if response.Error != nil {
		p.recordCall("Stream", prompt, nil, nil, options, "", response.Error, start)
		return nil, response.Error
	}

	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)

		// Apply delay if configured
		if response.Delay > 0 {
			select {
			case <-time.After(response.Delay):
			case <-ctx.Done():
				return
			}
		}

		// Stream tokens
		words := splitIntoTokens(response.Content)
		for i, word := range words {
			select {
			case ch <- ldomain.Token{Text: word, Finished: i == len(words)-1}:
			case <-ctx.Done():
				return
			}
		}
	}()

	p.recordCall("Stream", prompt, nil, nil, options, "stream:"+response.Content, nil, start)
	return ch, nil
}

// StreamMessage streams responses token by token with messages
func (p *MockProvider) StreamMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	start := time.Now()

	// Check behavior hook first
	if p.OnStreamMessage != nil {
		stream, err := p.OnStreamMessage(ctx, messages, options...)
		p.recordCall("StreamMessage", "", messages, nil, options, "stream", err, start)
		return stream, err
	}

	// Convert to prompt and use Stream
	prompt := ""
	if len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		// Extract text content from the last message
		for _, part := range lastMsg.Content {
			if part.Type == ldomain.ContentTypeText {
				prompt = part.Text
				break
			}
		}
	}

	return p.Stream(ctx, prompt, options...)
}

// GetCallHistory returns the call history
func (p *MockProvider) GetCallHistory() []ProviderCall {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]ProviderCall, len(p.callHistory))
	copy(history, p.callHistory)
	return history
}

// GetCallCount returns the total number of calls
func (p *MockProvider) GetCallCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.callCount
}

// Reset clears the call history and resets counters
func (p *MockProvider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.callHistory = make([]ProviderCall, 0)
	p.callCount = 0
	p.lastError = nil
}

// AssertCalled verifies that a method was called with specific parameters
func (p *MockProvider) AssertCalled(method string, promptPattern string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	re, err := regexp.Compile(promptPattern)
	if err != nil {
		return false
	}

	for _, call := range p.callHistory {
		if call.Method == method && re.MatchString(call.Prompt) {
			return true
		}
	}

	return false
}

// Private methods

func (p *MockProvider) findMatchingResponse(prompt string) Response {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Try to find a matching pattern
	for pattern, response := range p.ResponsePattern {
		if matched, _ := regexp.MatchString(pattern, prompt); matched {
			return response
		}
	}

	// Return default response
	return p.DefaultResponse
}

func (p *MockProvider) recordCall(method, prompt string, messages []ldomain.Message, schema *sdomain.Schema, options []ldomain.Option, response string, err error, start time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	call := ProviderCall{
		Method:    method,
		Prompt:    prompt,
		Messages:  messages,
		Schema:    schema,
		Options:   options,
		Response:  response,
		Error:     err,
		Timestamp: start,
		Duration:  time.Since(start),
	}

	p.callHistory = append(p.callHistory, call)
	p.callCount++
	p.lastError = err
}

func splitIntoTokens(text string) []string {
	// Simple word-based tokenization
	words := regexp.MustCompile(`\S+|\s+`).FindAllString(text, -1)
	return words
}

// Name returns the provider name to satisfy Mock interface
func (p *MockProvider) Name() string {
	return p.ProviderName
}

// TestMockProvider is a simplified mock provider for backward compatibility
type TestMockProvider struct {
	*MockProvider
	GenerateFunc           func(ctx context.Context, prompt string, options ...ldomain.Option) (string, error)
	GenerateMessageFunc    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	GenerateWithSchemaFunc func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
	StreamFunc             func(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error)
	StreamMessageFunc      func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error)
}

// NewTestMockProvider creates a TestMockProvider with backward compatibility
func NewTestMockProvider() *TestMockProvider {
	mp := NewMockProvider("test")
	tmp := &TestMockProvider{MockProvider: mp}

	// Wire up the function fields to behavior hooks
	mp.OnGenerate = func(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
		if tmp.GenerateFunc != nil {
			return tmp.GenerateFunc(ctx, prompt, options...)
		}
		return mp.DefaultResponse.Content, nil
	}

	mp.OnGenerateMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		if tmp.GenerateMessageFunc != nil {
			return tmp.GenerateMessageFunc(ctx, messages, options...)
		}
		return ldomain.Response{Content: mp.DefaultResponse.Content}, nil
	}

	mp.OnGenerateWithSchema = func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
		if tmp.GenerateWithSchemaFunc != nil {
			return tmp.GenerateWithSchemaFunc(ctx, prompt, schema, options...)
		}
		return map[string]interface{}{"result": "Default structured response"}, nil
	}

	mp.OnStream = func(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
		if tmp.StreamFunc != nil {
			return tmp.StreamFunc(ctx, prompt, options...)
		}
		// Default stream implementation
		ch := make(chan ldomain.Token)
		go func() {
			defer close(ch)
			ch <- ldomain.Token{Text: "Test", Finished: false}
			ch <- ldomain.Token{Text: " response", Finished: true}
		}()
		return ch, nil
	}

	mp.OnStreamMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
		if tmp.StreamMessageFunc != nil {
			return tmp.StreamMessageFunc(ctx, messages, options...)
		}
		// Default stream implementation
		ch := make(chan ldomain.Token)
		go func() {
			defer close(ch)
			ch <- ldomain.Token{Text: "Test", Finished: false}
			ch <- ldomain.Token{Text: " response", Finished: true}
		}()
		return ch, nil
	}

	return tmp
}

// CustomMockProvider maintains backward compatibility with legacy test code
type CustomMockProvider struct {
	*MockProvider
	GenerateMessageFunc    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	GenerateWithSchemaFunc func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
}

// NewCustomMockProvider creates a CustomMockProvider with backward compatibility
func NewCustomMockProvider() *CustomMockProvider {
	mp := NewMockProvider("custom")
	cmp := &CustomMockProvider{MockProvider: mp}

	// Wire up the function fields
	mp.OnGenerateMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		if cmp.GenerateMessageFunc != nil {
			return cmp.GenerateMessageFunc(ctx, messages, options...)
		}
		return ldomain.Response{Content: mp.DefaultResponse.Content}, nil
	}

	mp.OnGenerateWithSchema = func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
		if cmp.GenerateWithSchemaFunc != nil {
			return cmp.GenerateWithSchemaFunc(ctx, prompt, schema, options...)
		}
		return map[string]interface{}{"result": "Default structured response"}, nil
	}

	return cmp
}

// MockStructuredProvider is a specialized mock for structured data testing
type MockStructuredProvider struct {
	*MockProvider
	// Data is the structured data to return
	Data interface{}
}

// NewMockStructuredProvider creates a mock provider that returns specific structured data
func NewMockStructuredProvider(data interface{}) *MockStructuredProvider {
	mp := NewMockProvider("structured")
	msp := &MockStructuredProvider{
		MockProvider: mp,
		Data:         data,
	}

	// Override schema generation to return the configured data
	mp.OnGenerateWithSchema = func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
		return msp.Data, nil
	}

	return msp
}
