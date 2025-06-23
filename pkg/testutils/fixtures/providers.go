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

// Provider-specific configuration fixtures

// OpenAIMockProvider creates a mock provider configured like OpenAI GPT
func OpenAIMockProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("openai-gpt-4")

	// OpenAI-style responses with usage tracking
	provider.WithPatternResponse("(?i).*summarize.*", mocks.Response{
		Content: "Here's a concise summary of the key points:",
		Metadata: map[string]interface{}{
			"model": "gpt-4-turbo",
			"usage": map[string]interface{}{
				"prompt_tokens":     25,
				"completion_tokens": 15,
				"total_tokens":      40,
			},
			"finish_reason": "stop",
		},
	})

	provider.WithPatternResponse("(?i).*write.*code.*", mocks.Response{
		Content: "```python\n# Here's the code you requested\nprint('Hello, World!')\n```",
		Metadata: map[string]interface{}{
			"model": "gpt-4-turbo",
			"usage": map[string]interface{}{
				"prompt_tokens":     12,
				"completion_tokens": 20,
				"total_tokens":      32,
			},
			"finish_reason": "stop",
		},
	})

	provider.WithDefaultResponse(mocks.Response{
		Content: "I'll help you with that request.",
		Metadata: map[string]interface{}{
			"model": "gpt-4-turbo",
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 8,
				"total_tokens":      18,
			},
			"finish_reason": "stop",
		},
	})

	return provider
}

// AnthropicMockProvider creates a mock provider configured like Anthropic Claude
func AnthropicMockProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("claude-3-5-sonnet")

	// Claude-style responses with detailed analysis
	provider.WithPatternResponse("(?i).*research.*", mocks.Response{
		Content: "I'll conduct thorough research on this topic. Let me analyze the available information and provide you with comprehensive findings.",
		Metadata: map[string]interface{}{
			"model": "claude-3-5-sonnet-20241022",
			"usage": map[string]interface{}{
				"input_tokens":  15,
				"output_tokens": 25,
			},
			"stop_reason": "end_turn",
		},
	})

	provider.WithPatternResponse("(?i).*analyze.*", mocks.Response{
		Content: "I'll analyze this systematically. Here's my detailed breakdown of the key components and their relationships.",
		Metadata: map[string]interface{}{
			"model": "claude-3-5-sonnet-20241022",
			"usage": map[string]interface{}{
				"input_tokens":  12,
				"output_tokens": 22,
			},
			"stop_reason": "end_turn",
		},
	})

	provider.WithDefaultResponse(mocks.Response{
		Content: "I understand your request. Let me provide you with a thoughtful and comprehensive response.",
		Metadata: map[string]interface{}{
			"model": "claude-3-5-sonnet-20241022",
			"usage": map[string]interface{}{
				"input_tokens":  8,
				"output_tokens": 16,
			},
			"stop_reason": "end_turn",
		},
	})

	return provider
}

// GeminiMockProvider creates a mock provider configured like Google Gemini
func GeminiMockProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("gemini-1.5-pro")

	// Gemini-style responses with safety ratings
	provider.WithPatternResponse("(?i).*creative.*", mocks.Response{
		Content: "I'll create something creative for you! Here's an innovative approach to your request.",
		Metadata: map[string]interface{}{
			"model": "gemini-1.5-pro-001",
			"usage": map[string]interface{}{
				"prompt_token_count":     10,
				"candidates_token_count": 18,
				"total_token_count":      28,
			},
			"safety_ratings": []map[string]interface{}{
				{"category": "HARM_CATEGORY_HARASSMENT", "probability": "NEGLIGIBLE"},
				{"category": "HARM_CATEGORY_HATE_SPEECH", "probability": "NEGLIGIBLE"},
			},
			"finish_reason": "STOP",
		},
	})

	provider.WithDefaultResponse(mocks.Response{
		Content: "I can help you with that. Here's my response to your query.",
		Metadata: map[string]interface{}{
			"model": "gemini-1.5-pro-001",
			"usage": map[string]interface{}{
				"prompt_token_count":     8,
				"candidates_token_count": 12,
				"total_token_count":      20,
			},
			"safety_ratings": []map[string]interface{}{
				{"category": "HARM_CATEGORY_HARASSMENT", "probability": "NEGLIGIBLE"},
				{"category": "HARM_CATEGORY_HATE_SPEECH", "probability": "NEGLIGIBLE"},
			},
			"finish_reason": "STOP",
		},
	})

	return provider
}

// Enhanced streaming provider fixtures

// RealisticStreamingProvider creates a provider with realistic streaming patterns
func RealisticStreamingProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("realistic-streaming")

	provider.OnStreamMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
		ch := make(chan ldomain.Token)

		go func() {
			defer close(ch)

			// Simulate realistic streaming with variable delays
			words := []string{"This", "is", "a", "realistic", "streaming", "response", "that", "demonstrates", "how", "LLMs", "actually", "stream", "tokens", "with", "variable", "timing", "and", "occasional", "pauses", "for", "processing."}

			for i, word := range words {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Variable delays to simulate realistic streaming
				var delay time.Duration
				switch {
				case i < 3: // Fast start
					delay = 20 * time.Millisecond
				case i > 10 && i < 13: // Pause for processing
					delay = 150 * time.Millisecond
				case i > 15: // Slower finish
					delay = 80 * time.Millisecond
				default: // Normal speed
					delay = 50 * time.Millisecond
				}

				time.Sleep(delay)

				token := word
				if i < len(words)-1 {
					token += " "
				}

				ch <- ldomain.Token{
					Text:     token,
					Finished: i == len(words)-1,
				}
			}
		}()

		return ch, nil
	}

	provider.WithDefaultResponse(mocks.Response{
		Content: "This is a realistic streaming response that demonstrates how LLMs actually stream tokens with variable timing and occasional pauses for processing.",
		Metadata: map[string]interface{}{
			"model":          "realistic-streaming-model",
			"stream_capable": true,
			"response_time":  "variable",
		},
	})

	return provider
}

// FastStreamingProvider creates a provider with rapid streaming
func FastStreamingProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("fast-streaming")

	provider.OnStreamMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
		ch := make(chan ldomain.Token)

		go func() {
			defer close(ch)

			tokens := []string{"Fast", " streaming", " response", " with", " minimal", " latency", "."}

			for i, token := range tokens {
				select {
				case <-ctx.Done():
					return
				case ch <- ldomain.Token{
					Text:     token,
					Finished: i == len(tokens)-1,
				}:
					time.Sleep(10 * time.Millisecond) // Very fast
				}
			}
		}()

		return ch, nil
	}

	return provider
}

// Comprehensive error scenario fixtures

// RateLimitErrorProvider creates a provider that simulates rate limiting
func RateLimitErrorProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("rate-limit-error")

	provider.OnGenerateMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{}, &RateLimitError{
			Message:    "Rate limit exceeded. Please retry after 60 seconds.",
			RetryAfter: 60 * time.Second,
		}
	}

	return provider
}

// AuthenticationErrorProvider creates a provider that simulates auth failures
func AuthenticationErrorProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("auth-error")

	provider.OnGenerateMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{}, &AuthenticationError{
			Message: "Invalid API key provided. Please check your authentication credentials.",
		}
	}

	return provider
}

// NetworkErrorProvider creates a provider that simulates network issues
func NetworkErrorProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("network-error")

	provider.OnGenerateMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{}, &NetworkError{
			Message: "Request timeout: Unable to connect to the API endpoint.",
			Timeout: true,
		}
	}

	return provider
}

// IntermittentErrorProvider creates a provider that fails occasionally
func IntermittentErrorProvider(successRate float64) *mocks.MockProvider {
	provider := mocks.NewMockProvider("intermittent-error")
	var callCount int

	provider.OnGenerateMessage = func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		callCount++

		// Simulate success/failure based on success rate and call pattern
		if float64(callCount%10)/10.0 < successRate {
			return ldomain.Response{
				Content: "Successful response after intermittent failures.",
			}, nil
		}

		return ldomain.Response{}, errors.New("intermittent_error: Temporary service unavailable")
	}

	return provider
}

// Configuration-specific fixtures

// ConfiguredOpenAIProvider creates an OpenAI provider with specific configuration
func ConfiguredOpenAIProvider(model string, temperature float64, maxTokens int) *mocks.MockProvider {
	provider := OpenAIMockProvider()

	// Override to reflect configuration in metadata
	originalDefault := provider.DefaultResponse
	provider.WithDefaultResponse(mocks.Response{
		Content: originalDefault.Content,
		Metadata: map[string]interface{}{
			"model":       model,
			"temperature": temperature,
			"max_tokens":  maxTokens,
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 15,
				"total_tokens":      25,
			},
			"finish_reason": "stop",
		},
	})

	return provider
}

// ConfiguredAnthropicProvider creates a Claude provider with specific configuration
func ConfiguredAnthropicProvider(model string, maxTokens int, temperature float64) *mocks.MockProvider {
	provider := AnthropicMockProvider()

	// Override to reflect configuration in metadata
	originalDefault := provider.DefaultResponse
	provider.WithDefaultResponse(mocks.Response{
		Content: originalDefault.Content,
		Metadata: map[string]interface{}{
			"model":       model,
			"max_tokens":  maxTokens,
			"temperature": temperature,
			"usage": map[string]interface{}{
				"input_tokens":  8,
				"output_tokens": 16,
			},
			"stop_reason": "end_turn",
		},
	})

	return provider
}

// BasicMockProvider creates a simple provider for basic testing scenarios
func BasicMockProvider() *mocks.MockProvider {
	provider := mocks.NewMockProvider("basic-mock")
	provider.WithDefaultResponse(mocks.Response{
		Content: "Test response",
		Metadata: map[string]interface{}{
			"model": "basic-test-model",
		},
	})
	return provider
}

// BasicMockProviderWithContent creates a provider with specific content
func BasicMockProviderWithContent(content string) *mocks.MockProvider {
	provider := mocks.NewMockProvider("basic-mock")
	provider.WithDefaultResponse(mocks.Response{
		Content: content,
		Metadata: map[string]interface{}{
			"model": "basic-test-model",
		},
	})
	return provider
}

// Custom error types for realistic error simulation

// RateLimitError represents an error when API rate limits are exceeded.
type RateLimitError struct {
	Message    string
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// AuthenticationError represents an error when authentication fails.
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// NetworkError represents a network-related error during API communication.
type NetworkError struct {
	Message string
	Timeout bool
}

func (e *NetworkError) Error() string {
	return e.Message
}
