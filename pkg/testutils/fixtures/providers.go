// ABOUTME: Pre-configured mock providers for common testing scenarios
// ABOUTME: Provides ready-to-use provider fixtures with typical behavior patterns

package fixtures

import (
	"context"
	"errors"
	"time"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// ChatGPTMockProvider creates a mock provider configured to behave like ChatGPT
func ChatGPTMockProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("chatgpt-mock")

	// Add typical ChatGPT responses
	provider.WithPatternResponse("(?i).*hello.*", mocks.Response{
		Content: "Hello! How can I assist you today?",
		Metadata: map[string]interface{}{
			"model": "gpt-4",
			"usage": map[string]interface{}{
				"prompt_tokens":     5,
				"completion_tokens": 10,
				"total_tokens":      15,
			},
		},
	})

	provider.WithPatternResponse("(?i).*explain.*", mocks.Response{
		Content: "I'd be happy to explain that for you. Let me break it down step by step.",
		Metadata: map[string]interface{}{
			"model": "gpt-4",
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		},
	})

	provider.WithPatternResponse("(?i).*code.*", mocks.Response{
		Content: "Here's a code example:\n\n```python\ndef example():\n    return 'Hello, World!'\n```",
		Metadata: map[string]interface{}{
			"model": "gpt-4",
			"usage": map[string]interface{}{
				"prompt_tokens":     8,
				"completion_tokens": 25,
				"total_tokens":      33,
			},
		},
	})

	// Default response for unmatched patterns
	provider.WithDefaultResponse(mocks.Response{
		Content: "I understand your request. Here's my response based on the information provided.",
		Metadata: map[string]interface{}{
			"model": "gpt-4",
			"usage": map[string]interface{}{
				"prompt_tokens":     15,
				"completion_tokens": 15,
				"total_tokens":      30,
			},
		},
	})

	return provider
}

// ClaudeMockProvider creates a mock provider configured to behave like Claude
func ClaudeMockProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("claude-mock")

	// Add typical Claude responses
	provider.WithPatternResponse("(?i).*hello.*", mocks.Response{
		Content: "Hello! I'm Claude, an AI assistant. How may I help you today?",
		Metadata: map[string]interface{}{
			"model": "claude-3-opus",
			"usage": map[string]interface{}{
				"prompt_tokens":     5,
				"completion_tokens": 15,
				"total_tokens":      20,
			},
		},
	})

	provider.WithPatternResponse("(?i).*analyze.*", mocks.Response{
		Content: "I'll analyze this carefully. Let me examine the key aspects and provide you with a comprehensive analysis.",
		Metadata: map[string]interface{}{
			"model": "claude-3-opus",
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 25,
				"total_tokens":      35,
			},
		},
	})

	provider.WithPatternResponse("(?i).*explain.*", mocks.Response{
		Content: "I'll explain this concept thoroughly. Let me start with the fundamentals and build up to the more complex aspects.",
		Metadata: map[string]interface{}{
			"model": "claude-3-opus",
			"usage": map[string]interface{}{
				"prompt_tokens":     8,
				"completion_tokens": 30,
				"total_tokens":      38,
			},
		},
	})

	// Default response
	provider.WithDefaultResponse(mocks.Response{
		Content: "I understand your question. Let me provide a thoughtful and detailed response.",
		Metadata: map[string]interface{}{
			"model": "claude-3-opus",
			"usage": map[string]interface{}{
				"prompt_tokens":     12,
				"completion_tokens": 18,
				"total_tokens":      30,
			},
		},
	})

	return provider
}

// ErrorMockProvider creates a mock provider that always returns errors
func ErrorMockProvider(errorType string) *mocks.MockProvider {
	provider := mocks.NewMockProvider("error-mock")

	// Configure to always return errors
	provider.OnGenerateMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		switch errorType {
		case "rate_limit":
			return ldomain.Response{}, errors.New("rate_limit: Rate limit exceeded - please try again later")
		case "auth":
			return ldomain.Response{}, errors.New("auth: Authentication failed - invalid API key")
		case "network":
			return ldomain.Response{}, errors.New("network: Network error - connection timeout")
		case "invalid_response":
			return ldomain.Response{}, errors.New("invalid_response: Invalid response format from API")
		default:
			return ldomain.Response{}, errors.New("unknown: An unknown error occurred")
		}
	}

	return provider
}

// SlowMockProvider creates a mock provider with configurable response delay
func SlowMockProvider(delay time.Duration) *mocks.MockProvider {
	provider := mocks.NewMockProvider("slow-mock")

	// Add some default responses with delay
	provider.WithDefaultResponse(mocks.Response{
		Content: "This response was intentionally delayed.",
		Delay:   delay,
		Metadata: map[string]interface{}{
			"model":          "slow-model",
			"response_delay": delay.String(),
		},
	})

	return provider
}

// StreamingMockProvider creates a mock provider that simulates streaming responses
func StreamingMockProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("streaming-mock")

	// Configure streaming behavior
	provider.OnStreamMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
		// Create a channel for streaming tokens
		ch := make(chan ldomain.Token)

		go func() {
			defer close(ch)

			// Simulate streaming response
			chunks := []string{
				"This ",
				"is ",
				"a ",
				"streaming ",
				"response ",
				"that ",
				"arrives ",
				"in ",
				"chunks.",
			}

			for i, chunk := range chunks {
				select {
				case <-ctx.Done():
					return
				case ch <- ldomain.Token{
					Text:     chunk,
					Finished: i == len(chunks)-1,
				}:
					// Small delay between chunks
					time.Sleep(50 * time.Millisecond)
				}
			}
		}()

		// Return the channel directly as ResponseStream
		return ch, nil
	}

	// Also support non-streaming mode
	provider.WithDefaultResponse(mocks.Response{
		Content: "This is a streaming response that arrives in chunks.",
		Metadata: map[string]interface{}{
			"model":          "streaming-model",
			"stream_capable": true,
		},
	})

	return provider
}
