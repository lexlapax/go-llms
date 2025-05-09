package provider

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// mapOpenAIErrorToStandard maps OpenAI API error messages to standard error types
func mapOpenAIErrorToStandard(statusCode int, errorMsg string, operation string) error {
	// Convert error message to lowercase for case-insensitive matching
	lowerErrorMsg := strings.ToLower(errorMsg)

	// Common error patterns for OpenAI
	switch {
	case statusCode == http.StatusUnauthorized || strings.Contains(lowerErrorMsg, "invalid api key"):
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrAuthenticationFailed)

	case statusCode == http.StatusTooManyRequests || strings.Contains(lowerErrorMsg, "rate limit"):
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrRateLimitExceeded)

	case strings.Contains(lowerErrorMsg, "context length"):
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrContextTooLong)

	case strings.Contains(lowerErrorMsg, "content filter"):
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrContentFiltered)

	case strings.Contains(lowerErrorMsg, "model not found"):
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrModelNotFound)

	case strings.Contains(lowerErrorMsg, "quota") || strings.Contains(lowerErrorMsg, "billing"):
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrTokenQuotaExceeded)

	case strings.Contains(lowerErrorMsg, "invalid parameter") || strings.Contains(lowerErrorMsg, "invalid request"):
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrInvalidModelParameters)

	case statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusGatewayTimeout:
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrNetworkConnectivity)

	case statusCode >= 500:
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrProviderUnavailable)

	default:
		return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrRequestFailed)
	}
}

// mapAnthropicErrorToStandard maps Anthropic API error messages to standard error types
func mapAnthropicErrorToStandard(statusCode int, errorType, errorMsg string, operation string) error {
	// Convert error message and type to lowercase for case-insensitive matching
	lowerErrorMsg := strings.ToLower(errorMsg)
	lowerErrorType := strings.ToLower(errorType)

	// Common error patterns for Anthropic
	switch {
	case statusCode == http.StatusUnauthorized ||
		strings.Contains(lowerErrorType, "authentication") ||
		strings.Contains(lowerErrorMsg, "api key"):
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrAuthenticationFailed)

	case statusCode == http.StatusTooManyRequests ||
		strings.Contains(lowerErrorType, "rate_limit") ||
		strings.Contains(lowerErrorMsg, "rate limit"):
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrRateLimitExceeded)

	case strings.Contains(lowerErrorType, "context_length") ||
		strings.Contains(lowerErrorMsg, "context length") ||
		strings.Contains(lowerErrorMsg, "too long"):
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrContextTooLong)

	case strings.Contains(lowerErrorType, "content_filter") ||
		strings.Contains(lowerErrorMsg, "content filtered") ||
		strings.Contains(lowerErrorMsg, "content policy"):
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrContentFiltered)

	case strings.Contains(lowerErrorType, "model_not_found") ||
		strings.Contains(lowerErrorMsg, "model not found"):
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrModelNotFound)

	case strings.Contains(lowerErrorType, "quota") ||
		strings.Contains(lowerErrorMsg, "quota") ||
		strings.Contains(lowerErrorMsg, "billing") ||
		strings.Contains(lowerErrorMsg, "payment"):
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrTokenQuotaExceeded)

	case strings.Contains(lowerErrorType, "invalid_param") ||
		strings.Contains(lowerErrorMsg, "invalid parameter") ||
		strings.Contains(lowerErrorMsg, "invalid request"):
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrInvalidModelParameters)

	case statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusGatewayTimeout ||
		strings.Contains(lowerErrorMsg, "network") ||
		strings.Contains(lowerErrorMsg, "connection") ||
		strings.Contains(lowerErrorType, "connection"):
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrNetworkConnectivity)

	case statusCode >= 500:
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrProviderUnavailable)

	default:
		return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrRequestFailed)
	}
}

// ParseJSONError attempts to extract error information from a JSON error response
// This is a utility function to abstract error parsing logic for different providers
func ParseJSONError(body []byte, statusCode int, provider, operation string) error {
	if len(body) == 0 {
		// If no body, create a generic error based on status code
		return domain.NewProviderError(
			provider,
			operation,
			statusCode,
			fmt.Sprintf("HTTP error: %d", statusCode),
			nil,
		)
	}

	// Look for common error patterns in JSON
	errorRegex := regexp.MustCompile(`"error":\s*\{\s*"message":\s*"([^"]+)"`)
	matches := errorRegex.FindSubmatch(body)

	if len(matches) > 1 {
		errorMsg := string(matches[1])

		switch provider {
		case "openai":
			return mapOpenAIErrorToStandard(statusCode, errorMsg, operation)
		case "anthropic":
			// For Anthropic, we ideally need the error type as well
			// Try to extract it from the JSON
			typeRegex := regexp.MustCompile(`"error":\s*\{\s*"type":\s*"([^"]+)"`)
			typeMatches := typeRegex.FindSubmatch(body)

			var errorType string
			if len(typeMatches) > 1 {
				errorType = string(typeMatches[1])
			}

			return mapAnthropicErrorToStandard(statusCode, errorType, errorMsg, operation)
		default:
			// For other providers, use a generic approach
			return domain.NewProviderError(provider, operation, statusCode, errorMsg, nil)
		}
	}

	// If we couldn't parse the error message, return a generic error
	return domain.NewProviderError(
		provider,
		operation,
		statusCode,
		fmt.Sprintf("Unknown error (status: %d)", statusCode),
		nil,
	)
}

// MultiProviderError represents an error from multiple providers
type MultiProviderError struct {
	// ProviderErrors contains the errors from each provider
	ProviderErrors map[string]error

	// Message is the overall error message
	Message string
}

// Error implements the error interface
func (e *MultiProviderError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	// Build a detailed error message from all provider errors
	var errMsgs []string
	for provider, err := range e.ProviderErrors {
		errMsgs = append(errMsgs, fmt.Sprintf("[%s: %v]", provider, err))
	}

	return fmt.Sprintf("multi-provider errors: %s", strings.Join(errMsgs, " "))
}

// Unwrap returns the first error in the map
// Note: In Go, a multi-error can only unwrap to a single error
func (e *MultiProviderError) Unwrap() error {
	// Return the first error as the unwrapped error
	for _, err := range e.ProviderErrors {
		return err
	}
	return nil
}

// Is checks if the target error is in the provider errors
func (e *MultiProviderError) Is(target error) bool {
	for _, err := range e.ProviderErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// NewMultiProviderError creates a new MultiProviderError
func NewMultiProviderError(providerErrors map[string]error, message string) *MultiProviderError {
	return &MultiProviderError{
		ProviderErrors: providerErrors,
		Message:        message,
	}
}
