// ABOUTME: Provider testing utilities for response generation and error injection
// ABOUTME: Provides helpers for simulating various provider behaviors and streaming

package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// ResponseGenerator generates test responses based on schemas
type ResponseGenerator struct {
	schema *sdomain.Schema
}

// NewResponseGenerator creates a new response generator
func NewResponseGenerator(schema *sdomain.Schema) *ResponseGenerator {
	return &ResponseGenerator{schema: schema}
}

// Generate generates a response that matches the schema
func (rg *ResponseGenerator) Generate() (string, error) {
	if rg.schema == nil {
		return "Generated response", nil
	}

	// Generate based on schema type
	switch rg.schema.Type {
	case "object":
		return rg.generateObject()
	case "array":
		return rg.generateArray()
	case "string":
		return rg.generateString()
	case "number", "integer":
		return rg.generateNumber()
	case "boolean":
		return "true", nil
	default:
		return "Generated response", nil
	}
}

func (rg *ResponseGenerator) generateObject() (string, error) {
	obj := make(map[string]interface{})

	for name, prop := range rg.schema.Properties {
		value := generatePropertyValue(prop)
		obj[name] = value
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (rg *ResponseGenerator) generateArray() (string, error) {
	arr := []interface{}{
		"item1",
		"item2",
		"item3",
	}

	data, err := json.Marshal(arr)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (rg *ResponseGenerator) generateString() (string, error) {
	// Note: Schema doesn't have Enum, that's on Property
	return "generated string value", nil
}

func (rg *ResponseGenerator) generateNumber() (string, error) {
	return "42", nil
}

func generatePropertyValue(prop sdomain.Property) interface{} {
	switch prop.Type {
	case "string":
		if len(prop.Enum) > 0 {
			return prop.Enum[0]
		}
		return "test value"
	case "number", "integer":
		return 42
	case "boolean":
		return true
	case "array":
		return []interface{}{"item1", "item2"}
	case "object":
		return map[string]interface{}{"nested": "value"}
	default:
		return "default value"
	}
}

// ErrorInjector provides controlled error injection for providers
type ErrorInjector struct {
	errorRate   float64
	errorType   string
	callCount   int
	errorAfter  int
	specificMsg string
}

// NewErrorInjector creates a new error injector
func NewErrorInjector(errorType string) *ErrorInjector {
	return &ErrorInjector{
		errorType:  errorType,
		errorRate:  1.0, // Always error by default
		errorAfter: -1,  // Error immediately by default
	}
}

// WithRate sets the error rate (0.0 to 1.0)
func (ei *ErrorInjector) WithRate(rate float64) *ErrorInjector {
	ei.errorRate = rate
	return ei
}

// WithErrorAfter errors after N successful calls
func (ei *ErrorInjector) WithErrorAfter(n int) *ErrorInjector {
	ei.errorAfter = n
	return ei
}

// WithMessage sets a specific error message
func (ei *ErrorInjector) WithMessage(msg string) *ErrorInjector {
	ei.specificMsg = msg
	return ei
}

// ShouldError determines if an error should be injected
func (ei *ErrorInjector) ShouldError() bool {
	ei.callCount++

	if ei.errorAfter >= 0 {
		return ei.callCount > ei.errorAfter
	}

	// Simple error rate implementation
	return ei.callCount%int(1/ei.errorRate) == 0
}

// GetError returns the appropriate error
func (ei *ErrorInjector) GetError() error {
	if ei.specificMsg != "" {
		return fmt.Errorf("%s: %s", ei.errorType, ei.specificMsg)
	}

	switch ei.errorType {
	case "rate_limit":
		return fmt.Errorf("rate limit exceeded")
	case "auth":
		return fmt.Errorf("authentication failed")
	case "network":
		return fmt.Errorf("network error: connection refused")
	case "timeout":
		return fmt.Errorf("request timeout")
	case "invalid_response":
		return fmt.Errorf("invalid response format")
	default:
		return fmt.Errorf("provider error: %s", ei.errorType)
	}
}

// StreamSimulator simulates streaming responses
type StreamSimulator struct {
	chunks    []string
	delay     time.Duration
	errorAt   int
	chunkSize int
}

// NewStreamSimulator creates a new stream simulator
func NewStreamSimulator(content string) *StreamSimulator {
	// Split content into chunks
	chunkSize := 10
	var chunks []string

	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunks = append(chunks, content[i:end])
	}

	return &StreamSimulator{
		chunks:    chunks,
		delay:     10 * time.Millisecond,
		errorAt:   -1,
		chunkSize: chunkSize,
	}
}

// WithDelay sets the delay between chunks
func (ss *StreamSimulator) WithDelay(delay time.Duration) *StreamSimulator {
	ss.delay = delay
	return ss
}

// WithErrorAt simulates an error at a specific chunk
func (ss *StreamSimulator) WithErrorAt(chunk int) *StreamSimulator {
	ss.errorAt = chunk
	return ss
}

// Stream simulates streaming the response
func (ss *StreamSimulator) Stream(ctx context.Context, callback func(chunk string) error) error {
	for i, chunk := range ss.chunks {
		// Check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Simulate error
		if i == ss.errorAt {
			return fmt.Errorf("stream error at chunk %d", i)
		}

		// Send chunk
		if err := callback(chunk); err != nil {
			return err
		}

		// Delay between chunks
		if i < len(ss.chunks)-1 {
			time.Sleep(ss.delay)
		}
	}

	return nil
}

// ProviderBehavior configures provider behavior for testing
type ProviderBehavior struct {
	ResponseDelay   time.Duration
	StreamingDelay  time.Duration
	ErrorInjector   *ErrorInjector
	ResponsePattern string
	MetadataFunc    func() map[string]interface{}
}

// ApplyToProvider applies the behavior to a mock provider
func (pb *ProviderBehavior) ApplyToProvider(provider *mocks.MockProvider) {
	// Note: MockProvider doesn't have ResponseDelay/StreamingDelay fields
	// These would need to be implemented in the behavior hooks

	if pb.ErrorInjector != nil {
		provider.OnGenerateMessage = func(ctx context.Context, msgs []llmdomain.Message, opts ...llmdomain.Option) (llmdomain.Response, error) {
			if pb.ErrorInjector.ShouldError() {
				return llmdomain.Response{}, pb.ErrorInjector.GetError()
			}
			// Return default behavior
			return llmdomain.Response{}, nil
		}
	}

	if pb.MetadataFunc != nil {
		// Set metadata on responses
		provider.OnGenerateMessage = func(ctx context.Context, msgs []llmdomain.Message, opts ...llmdomain.Option) (llmdomain.Response, error) {
			resp := llmdomain.Response{
				Content: "Response with metadata",
				// Note: Response doesn't have Metadata field in current implementation
			}
			return resp, nil
		}
	}
}

// Common provider behaviors

// SlowProviderBehavior creates a slow provider behavior
func SlowProviderBehavior(delay time.Duration) *ProviderBehavior {
	return &ProviderBehavior{
		ResponseDelay:  delay,
		StreamingDelay: delay / 10,
	}
}

// UnreliableProviderBehavior creates an unreliable provider behavior
func UnreliableProviderBehavior(errorRate float64) *ProviderBehavior {
	return &ProviderBehavior{
		ErrorInjector: NewErrorInjector("network").WithRate(errorRate),
	}
}

// RateLimitedProviderBehavior creates a rate-limited provider behavior
func RateLimitedProviderBehavior(successfulCalls int) *ProviderBehavior {
	return &ProviderBehavior{
		ErrorInjector: NewErrorInjector("rate_limit").WithErrorAfter(successfulCalls),
	}
}

// MetadataProviderBehavior creates a provider that includes metadata
func MetadataProviderBehavior() *ProviderBehavior {
	return &ProviderBehavior{
		MetadataFunc: func() map[string]interface{} {
			return map[string]interface{}{
				"model":       "test-model",
				"temperature": 0.7,
				"max_tokens":  1000,
				"usage": map[string]interface{}{
					"prompt_tokens":     50,
					"completion_tokens": 100,
					"total_tokens":      150,
				},
			}
		},
	}
}
