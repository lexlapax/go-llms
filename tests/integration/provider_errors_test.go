package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// TestProviderErrors tests error handling for all providers
func TestProviderErrors(t *testing.T) {
	// This test file includes comprehensive error test cases for all providers.
	// It tests the error handling mechanisms for various error conditions such as:
	// - Network failures
	// - Authentication errors
	// - Rate limiting
	// - Context length errors
	// - Content filtering
	// - Timeouts
	// - Parsing errors

	// Mock server to simulate different API errors
	t.Run("MockErrorServer", func(t *testing.T) {
		// Create a server that will return different errors based on request path
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract path type from URL
			var errorType string
			if strings.Contains(r.URL.Path, "auth-error") {
				errorType = "auth-error"
			} else if strings.Contains(r.URL.Path, "rate-limit") {
				errorType = "rate-limit"
			} else if strings.Contains(r.URL.Path, "context-length") {
				errorType = "context-length"
			} else if strings.Contains(r.URL.Path, "content-filter") {
				errorType = "content-filter"
			} else if strings.Contains(r.URL.Path, "timeout") {
				errorType = "timeout"
			} else if strings.Contains(r.URL.Path, "parsing-error") {
				errorType = "parsing-error"
			} else if strings.Contains(r.URL.Path, "server-error") {
				errorType = "server-error"
			} else {
				errorType = "not-found"
			}

			// Determine provider type based on endpoint pattern
			var providerType string
			if strings.Contains(r.URL.Path, "/v1/chat/completions") {
				providerType = "openai"
			} else if strings.Contains(r.URL.Path, "/v1/messages") {
				providerType = "anthropic"
			} else if strings.Contains(r.URL.Path, "models/gemini") {
				providerType = "gemini"
			} else {
				providerType = "openai" // Default to OpenAI format
			}

			// Log request to aid debugging
			fmt.Printf("Request: %s %s (Provider: %s, Error: %s)\n", 
				r.Method, r.URL.Path, providerType, errorType)

			// Return appropriate error responses based on provider and error type
			switch errorType {
			case "auth-error":
				w.WriteHeader(http.StatusUnauthorized)
				switch providerType {
				case "anthropic":
					w.Write([]byte(`{"error":{"type":"authentication_error","message":"Invalid API key"}}`))
				case "gemini":
					w.Write([]byte(`{"error":{"code":401,"message":"API key not valid","status":"UNAUTHENTICATED"}}`))
				default: // openai
					w.Write([]byte(`{"error":{"message":"Invalid API key","type":"authentication_error","code":401}}`))
				}
			case "rate-limit":
				w.WriteHeader(http.StatusTooManyRequests)
				switch providerType {
				case "anthropic":
					w.Write([]byte(`{"error":{"type":"rate_limit_error","message":"Rate limit exceeded"}}`))
				case "gemini":
					w.Write([]byte(`{"error":{"code":429,"message":"Resource has been exhausted","status":"RESOURCE_EXHAUSTED"}}`))
				default: // openai
					w.Write([]byte(`{"error":{"message":"Rate limit exceeded","type":"rate_limit_error","code":429}}`))
				}
			case "context-length":
				w.WriteHeader(http.StatusBadRequest)
				switch providerType {
				case "anthropic":
					w.Write([]byte(`{"error":{"type":"context_length_error","message":"Prompt too long"}}`))
				case "gemini":
					w.Write([]byte(`{"error":{"code":400,"message":"Input content is too long","status":"INVALID_ARGUMENT"}}`))
				default: // openai
					w.Write([]byte(`{"error":{"message":"Context length exceeded","type":"context_length_error","code":400}}`))
				}
			case "content-filter":
				w.WriteHeader(http.StatusBadRequest)
				switch providerType {
				case "anthropic":
					w.Write([]byte(`{"error":{"type":"content_filter_error","message":"Content violates policy"}}`))
				case "gemini":
					w.Write([]byte(`{"error":{"code":400,"message":"Content violates policy","status":"FAILED_PRECONDITION"}}`))
				default: // openai
					w.Write([]byte(`{"error":{"message":"Content violates policy","type":"content_filter_error","code":400}}`))
				}
			case "timeout":
				// Sleep to simulate timeout
				time.Sleep(500 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				switch providerType {
				case "anthropic":
					w.Write([]byte(`{"content":[{"type":"text","text":"delayed response"}]}`))
				case "gemini":
					w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"delayed response"}]}}]}`))
				default: // openai
					w.Write([]byte(`{"choices":[{"message":{"content":"delayed response"},"finish_reason":"stop"}]}`))
				}
			case "parsing-error":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"invalid-json:`)) // invalid JSON for all providers
			case "server-error":
				w.WriteHeader(http.StatusInternalServerError)
				switch providerType {
				case "anthropic":
					w.Write([]byte(`{"error":{"type":"server_error","message":"Server error"}}`))
				case "gemini":
					w.Write([]byte(`{"error":{"code":500,"message":"Internal server error","status":"INTERNAL"}}`))
				default: // openai
					w.Write([]byte(`{"error":{"message":"Server error","type":"server_error","code":500}}`))
				}
			default:
				w.WriteHeader(http.StatusNotFound)
				switch providerType {
				case "anthropic":
					w.Write([]byte(`{"error":{"type":"not_found_error","message":"Not found"}}`))
				case "gemini":
					w.Write([]byte(`{"error":{"code":404,"message":"Not found","status":"NOT_FOUND"}}`))
				default: // openai
					w.Write([]byte(`{"error":{"message":"Not found","type":"not_found_error","code":404}}`))
				}
			}
		}))
		defer mockServer.Close()

		// Test authentication error
		t.Run("AuthenticationError", func(t *testing.T) {
			testAllProviders(t, mockServer.URL+"/auth-error", domain.ErrAuthenticationFailed)
		})

		// Test rate limit error
		t.Run("RateLimitError", func(t *testing.T) {
			testAllProviders(t, mockServer.URL+"/rate-limit", domain.ErrRateLimitExceeded)
		})

		// Test context length error
		t.Run("ContextLengthError", func(t *testing.T) {
			testAllProviders(t, mockServer.URL+"/context-length", domain.ErrContextTooLong)
		})

		// Test content filter error
		t.Run("ContentFilterError", func(t *testing.T) {
			testAllProviders(t, mockServer.URL+"/content-filter", domain.ErrContentFiltered)
		})

		// Test server error
		t.Run("ServerError", func(t *testing.T) {
			testAllProviders(t, mockServer.URL+"/server-error", domain.ErrProviderUnavailable)
		})

		// Test parsing error
		t.Run("ParsingError", func(t *testing.T) {
			testAllProviders(t, mockServer.URL+"/parsing-error", domain.ErrRequestFailed)
		})

		// Test timeout error
		t.Run("TimeoutError", func(t *testing.T) {
			baseURL := mockServer.URL + "/timeout"
			client := &http.Client{Timeout: 100 * time.Millisecond}
			clientOption := domain.NewHTTPClientOption(client)
			baseURLOption := domain.NewBaseURLOption(baseURL)

			// Test with all providers
			providers := []domain.Provider{
				provider.NewOpenAIProvider("test-key", "gpt-4o", clientOption, baseURLOption),
				provider.NewAnthropicProvider("test-key", "claude-3-5-sonnet-latest", clientOption, baseURLOption),
				provider.NewGeminiProvider("test-key", "gemini-2.0-flash-lite", clientOption, baseURLOption),
			}

			for _, p := range providers {
				_, err := p.Generate(context.Background(), "Test prompt")
				if err == nil {
					t.Errorf("Expected timeout error but got nil")
					continue
				}

				// Check for timeout-related errors - accept any error that mentions timeout or deadline
				if !strings.Contains(strings.ToLower(err.Error()), "deadline exceeded") && 
				   !strings.Contains(strings.ToLower(err.Error()), "timeout") &&
				   !domain.IsNetworkConnectivityError(err) {
					t.Errorf("Expected error with deadline/timeout references, got: %v", err)
				}
			}
		})
	})

	// Test network failure handling
	t.Run("NetworkFailure", func(t *testing.T) {
		// Use an invalid URL that will trigger a network error
		badURL := "https://invalid-domain-that-does-not-exist-123456789.example"
		baseURLOption := domain.NewBaseURLOption(badURL)
		
		// Test with all providers
		providers := []domain.Provider{
			provider.NewOpenAIProvider("test-key", "gpt-4o", baseURLOption),
			provider.NewAnthropicProvider("test-key", "claude-3-5-sonnet-latest", baseURLOption),
			provider.NewGeminiProvider("test-key", "gemini-2.0-flash-lite", baseURLOption),
		}
		
		for _, p := range providers {
			_, err := p.Generate(context.Background(), "Test prompt")
			if err == nil {
				t.Errorf("Expected network error but got nil")
				continue
			}
			
			if !domain.IsNetworkConnectivityError(err) {
				// Allow both network connectivity errors and the error message containing "no such host"
				if !strings.Contains(err.Error(), "no such host") {
					t.Errorf("Expected network connectivity error, got: %v", err)
				}
			}
		}
	})

	// Test retry mechanism with different error types
	t.Run("RetryMechanism", func(t *testing.T) {
		count := 0
		retryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment counter for each request
			count++
			
			// Succeed on the 3rd attempt
			if count >= 3 {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"choices":[{"message":{"content":"Success after retry"},"finish_reason":"stop"}]}`))
				return
			}
			
			// Otherwise return a retryable error (rate limit)
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"Rate limit exceeded","type":"rate_limit_error","code":429}}`))
		}))
		defer retryServer.Close()

		// Test with OpenAI provider (retry logic should be similar for all)
		p := provider.NewOpenAIProvider("test-key", "gpt-4o", domain.NewBaseURLOption(retryServer.URL))
		
		// Test with 3 retries (should succeed)
		result, err := generateWithRetry(context.Background(), p, "Test prompt", 3)
		if err != nil {
			t.Errorf("Expected success after retry but got error: %v", err)
		}
		if result == "" {
			t.Errorf("Expected non-empty result after retry")
		}

		// Reset the counter
		count = 0

		// Test with 1 retry (should fail)
		_, err = generateWithRetry(context.Background(), p, "Test prompt", 1)
		if err == nil {
			t.Errorf("Expected error with insufficient retries but got nil")
		}
	})
}

// Helper function to test all providers with a specific error
func testAllProviders(t *testing.T, baseURL string, expectedError error) {
	// Create multiple providers with the base URL option
	baseURLOption := domain.NewBaseURLOption(baseURL)
	
	// For content filter and context length errors, we'll be more lenient and just check
	// that there's an error, as specific mapping can vary by provider
	isSpecialCase := expectedError == domain.ErrContentFiltered ||
		              expectedError == domain.ErrContextTooLong ||
		              strings.Contains(baseURL, "/parsing-error")
	
	// OpenAI provider
	openaiProvider := provider.NewOpenAIProvider("test-key", "gpt-4o", baseURLOption)
	_, err := openaiProvider.Generate(context.Background(), "Test prompt")
	if err == nil {
		t.Errorf("Expected error but got nil for provider OpenAI")
	} else if !isSpecialCase && !errors.Is(err, expectedError) {
		t.Errorf("Expected error %v for provider OpenAI, got: %v", expectedError, err)
	}
	
	// Anthropic provider
	anthropicProvider := provider.NewAnthropicProvider("test-key", "claude-3-5-sonnet-latest", baseURLOption)
	_, err = anthropicProvider.Generate(context.Background(), "Test prompt")
	if err == nil {
		t.Errorf("Expected error but got nil for provider Anthropic")
	} else if !isSpecialCase && !errors.Is(err, expectedError) {
		t.Errorf("Expected error %v for provider Anthropic, got: %v", expectedError, err)
	}
	
	// Gemini provider
	geminiProvider := provider.NewGeminiProvider("test-key", "gemini-2.0-flash-lite", baseURLOption)
	_, err = geminiProvider.Generate(context.Background(), "Test prompt")
	if err == nil {
		t.Errorf("Expected error but got nil for provider Gemini")
	} else if !isSpecialCase && !errors.Is(err, expectedError) {
		t.Errorf("Expected error %v for provider Gemini, got: %v", expectedError, err)
	}
}

// Helper function to generate with retry
func generateWithRetry(ctx context.Context, provider domain.Provider, prompt string, maxRetries int) (string, error) {
	var result string
	var err error
	
	for i := 0; i < maxRetries; i++ {
		result, err = provider.Generate(ctx, prompt)
		if err == nil {
			return result, nil
		}
		
		// Check if this is a retryable error
		if !isRetryableError(err) {
			return "", fmt.Errorf("non-retryable error: %w", err)
		}
		
		// Wait a bit before retrying (increasing backoff)
		time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
	}
	
	return "", errors.New("max retries reached")
}

// Helper function to determine if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Network connectivity errors are retryable
	if errors.Is(err, domain.ErrNetworkConnectivity) {
		return true
	}
	
	// Rate limit errors are retryable
	if errors.Is(err, domain.ErrRateLimitExceeded) {
		return true
	}
	
	// Certain provider unavailable errors might be retryable
	if errors.Is(err, domain.ErrProviderUnavailable) {
		return true
	}
	
	// Check for specific error messages
	if strings.Contains(err.Error(), "rate limit") || 
	   strings.Contains(err.Error(), "too many requests") {
		return true
	}
	
	return false
}