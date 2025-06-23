// ABOUTME: This file defines error types related to LLM provider operations.
// ABOUTME: It includes errors for various failure scenarios including content type support.

package domain

import (
	"fmt"
	"strings"

	"github.com/lexlapax/go-llms/pkg/errors"
)

// Common error types
var (
	// ErrRequestFailed is returned when a request to an LLM provider fails.
	ErrRequestFailed = errors.NewErrorWithCode("llm_request_failed", "request to LLM provider failed").SetRetryable(true)

	// ErrResponseParsing is returned when the response from an LLM provider cannot be parsed.
	ErrResponseParsing = errors.NewErrorWithCode("llm_response_parsing", "failed to parse LLM provider response")

	// ErrInvalidJSON is returned when the LLM response does not contain valid JSON.
	ErrInvalidJSON = errors.NewErrorWithCode("llm_invalid_json", "response does not contain valid JSON")

	// ErrAuthenticationFailed is returned when authentication with the LLM provider fails.
	ErrAuthenticationFailed = errors.NewErrorWithCode("llm_auth_failed", "authentication with LLM provider failed").SetFatal(true)

	// ErrRateLimitExceeded is returned when the LLM provider rate limit is exceeded.
	ErrRateLimitExceeded = errors.NewErrorWithCode("llm_rate_limit", "rate limit exceeded").SetRetryable(true)

	// ErrContextTooLong is returned when the input context is too long for the model.
	ErrContextTooLong = errors.NewErrorWithCode("llm_context_too_long", "input context too long").SetFatal(true)

	// ErrProviderUnavailable is returned when the LLM provider is unavailable.
	ErrProviderUnavailable = errors.NewErrorWithCode("llm_provider_unavailable", "LLM provider unavailable").SetRetryable(true)

	// ErrInvalidConfiguration is returned when the provider configuration is invalid.
	ErrInvalidConfiguration = errors.NewErrorWithCode("llm_invalid_config", "invalid provider configuration").SetFatal(true)

	// ErrNoResponse is returned when the LLM provider returns no response.
	ErrNoResponse = errors.NewErrorWithCode("llm_no_response", "no response from LLM provider").SetRetryable(true)

	// ErrTimeout is returned when a request to an LLM provider times out.
	ErrTimeout = errors.NewErrorWithCode("llm_timeout", "LLM provider request timed out").SetRetryable(true)

	// ErrContentFiltered is returned when content is filtered by the LLM provider.
	ErrContentFiltered = errors.NewErrorWithCode("llm_content_filtered", "content filtered by LLM provider")

	// ErrModelNotFound is returned when the requested model is not found.
	ErrModelNotFound = errors.NewErrorWithCode("llm_model_not_found", "model not found").SetFatal(true)

	// ErrNetworkConnectivity is returned when there are network issues connecting to the provider.
	ErrNetworkConnectivity = errors.NewErrorWithCode("llm_network_error", "network connectivity issues").SetRetryable(true)

	// ErrTokenQuotaExceeded is returned when the user has exceeded their token quota.
	ErrTokenQuotaExceeded = errors.NewErrorWithCode("llm_quota_exceeded", "token quota exceeded").SetFatal(true).SetRetryable(false)

	// ErrInvalidModelParameters is returned when provided model parameters are invalid.
	ErrInvalidModelParameters = errors.NewErrorWithCode("llm_invalid_params", "invalid model parameters").SetFatal(true).SetRetryable(false)

	// ErrUnsupportedContentType is returned when a provider doesn't support a specific content type.
	ErrUnsupportedContentType = errors.NewErrorWithCode("llm_unsupported_content", "content type not supported by provider").SetFatal(true)
)

// ProviderError represents an error from an LLM provider with additional context.
// It extends BaseError with provider-specific information including
// provider name, operation, and HTTP status code.
type ProviderError struct {
	*errors.BaseError

	// Provider is the name of the LLM provider (e.g., "openai", "anthropic").
	Provider string `json:"provider"`

	// Operation is the operation that failed (e.g., "Generate", "Stream").
	Operation string `json:"operation"`

	// StatusCode is the HTTP status code if applicable.
	StatusCode int `json:"status_code,omitempty"`
}

// Error implements the error interface.
// Returns a formatted error message including provider name,
// operation, and optionally the HTTP status code.
//
// Returns the formatted error message.
func (e *ProviderError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s %s error (status %d): %s", e.Provider, e.Operation, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("%s %s error: %s", e.Provider, e.Operation, e.Message)
}

// NewProviderError creates a new ProviderError.
// Automatically selects appropriate base error type based on status code
// and message content. If no underlying error is provided, it intelligently
// maps common scenarios to specific error types.
//
// Parameters:
//   - provider: The name of the LLM provider
//   - operation: The operation that failed
//   - statusCode: HTTP status code (0 if not applicable)
//   - message: Error message
//   - err: Underlying error (nil to auto-detect type)
//
// Returns a new ProviderError with appropriate context.
func NewProviderError(provider, operation string, statusCode int, message string, err error) *ProviderError {
	// If no underlying error is provided, determine the most specific error based on status code or message
	var baseErr *errors.BaseError
	if err == nil {
		lowerMessage := strings.ToLower(message)

		switch {
		// HTTP Status Code based errors
		case statusCode == 401:
			baseErr = ErrAuthenticationFailed
		case statusCode == 429:
			baseErr = ErrRateLimitExceeded
		case statusCode == 400 && (strings.Contains(lowerMessage, "param") || strings.Contains(lowerMessage, "request")):
			baseErr = ErrInvalidModelParameters
		case statusCode == 402 || strings.Contains(lowerMessage, "quota") || strings.Contains(lowerMessage, "billing"):
			baseErr = ErrTokenQuotaExceeded
		case statusCode == 404 && strings.Contains(lowerMessage, "model"):
			baseErr = ErrModelNotFound
		case statusCode >= 500 && statusCode < 505:
			baseErr = ErrProviderUnavailable
		case statusCode == 502 || statusCode == 503 || statusCode == 504:
			baseErr = ErrNetworkConnectivity

		// Message content based errors (when status code doesn't help)
		case strings.Contains(lowerMessage, "context") && strings.Contains(lowerMessage, "length"):
			baseErr = ErrContextTooLong
		case strings.Contains(lowerMessage, "content") && (strings.Contains(lowerMessage, "filter") ||
			strings.Contains(lowerMessage, "policy")):
			baseErr = ErrContentFiltered
		case strings.Contains(lowerMessage, "network") || strings.Contains(lowerMessage, "connection"):
			baseErr = ErrNetworkConnectivity
		default:
			baseErr = ErrRequestFailed
		}
	} else {
		// Wrap the provided error
		baseErr = errors.Wrap(err, message)
	}

	// Create the provider error with enhanced context
	_ = baseErr.WithContext("provider", provider).
		WithContext("operation", operation).
		WithContext("status_code", statusCode).
		WithType("ProviderError")

	return &ProviderError{
		BaseError:  baseErr,
		Provider:   provider,
		Operation:  operation,
		StatusCode: statusCode,
	}
}

// IsAuthenticationError checks if the error is an authentication error.
// Checks for authentication failure error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates authentication failure.
func IsAuthenticationError(err error) bool {
	if baseErr, ok := err.(*errors.BaseError); ok {
		return baseErr.Code == ErrAuthenticationFailed.Code
	}
	return false
}

// IsRateLimitError checks if the error is a rate limit error.
// Checks for rate limit exceeded error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates rate limiting.
func IsRateLimitError(err error) bool {
	if baseErr, ok := err.(*errors.BaseError); ok {
		return baseErr.Code == ErrRateLimitExceeded.Code
	}
	return false
}

// IsTimeoutError checks if the error is a timeout error.
// Checks for request timeout error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates a timeout.
func IsTimeoutError(err error) bool {
	if baseErr, ok := err.(*errors.BaseError); ok {
		return baseErr.Code == ErrTimeout.Code
	}
	return false
}

// IsProviderUnavailableError checks if the error is a provider unavailable error.
// Checks for provider unavailability error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates provider unavailability.
func IsProviderUnavailableError(err error) bool {
	if baseErr, ok := err.(*errors.BaseError); ok {
		return baseErr.Code == ErrProviderUnavailable.Code
	}
	return false
}

// IsContentFilteredError checks if the error is a content filtered error.
// Checks for content filtering error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates content was filtered.
func IsContentFilteredError(err error) bool {
	if baseErr, ok := err.(*errors.BaseError); ok {
		return baseErr.Code == ErrContentFiltered.Code
	}
	return false
}

// IsNetworkConnectivityError checks if the error is a network connectivity error.
// Checks for network connectivity error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates network connectivity issues.
func IsNetworkConnectivityError(err error) bool {
	if baseErr, ok := err.(*errors.BaseError); ok {
		return baseErr.Code == ErrNetworkConnectivity.Code
	}
	return false
}

// IsTokenQuotaExceededError checks if the error is a token quota exceeded error.
// Checks for token quota exceeded error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates quota exhaustion.
func IsTokenQuotaExceededError(err error) bool {
	if baseErr, ok := err.(*errors.BaseError); ok {
		return baseErr.Code == ErrTokenQuotaExceeded.Code
	}
	return false
}

// IsInvalidModelParametersError checks if the error is an invalid model parameters error.
// Checks for invalid model parameter error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates invalid parameters.
func IsInvalidModelParametersError(err error) bool {
	if baseErr, ok := err.(*errors.BaseError); ok {
		return baseErr.Code == ErrInvalidModelParameters.Code
	}
	return false
}

// UnsupportedContentTypeError represents an error when a provider doesn't support a specific content type.
// This error is used when attempting to send content (like images, audio, video)
// to a provider that doesn't support that content type.
type UnsupportedContentTypeError struct {
	*errors.BaseError

	Provider    string      `json:"provider"`
	ContentType ContentType `json:"content_type"`
}

// Error implements the error interface.
// Returns a formatted message indicating which provider and content type.
//
// Returns the formatted error message.
func (e *UnsupportedContentTypeError) Error() string {
	return fmt.Sprintf("provider %s does not support content type %s", e.Provider, e.ContentType)
}

// NewUnsupportedContentTypeError creates a new error for unsupported content types.
// This function creates an error when a provider doesn't support a specific
// content type (e.g., sending images to a text-only model).
//
// Parameters:
//   - provider: The name of the provider
//   - contentType: The unsupported content type
//
// Returns a new UnsupportedContentTypeError.
func NewUnsupportedContentTypeError(provider string, contentType ContentType) *UnsupportedContentTypeError {
	baseErr := errors.Wrap(ErrUnsupportedContentType, fmt.Sprintf("provider %s does not support content type %s", provider, contentType))
	_ = baseErr.WithContext("provider", provider).
		WithContext("content_type", string(contentType)).
		WithType("UnsupportedContentTypeError")

	return &UnsupportedContentTypeError{
		BaseError:   baseErr,
		Provider:    provider,
		ContentType: contentType,
	}
}

// IsUnsupportedContentTypeError checks if the error is an unsupported content type error.
// Checks for unsupported content type error codes.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error indicates unsupported content type.
func IsUnsupportedContentTypeError(err error) bool {
	return errors.As(err, &ErrUnsupportedContentType)
}
