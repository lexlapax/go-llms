// Package domain defines core domain models and interfaces for LLM providers.
package domain

import (
	"errors"
	"fmt"
)

// Common error types
var (
	// ErrRequestFailed is returned when a request to an LLM provider fails.
	ErrRequestFailed = errors.New("request to LLM provider failed")
	
	// ErrResponseParsing is returned when the response from an LLM provider cannot be parsed.
	ErrResponseParsing = errors.New("failed to parse LLM provider response")
	
	// ErrInvalidJSON is returned when the LLM response does not contain valid JSON.
	ErrInvalidJSON = errors.New("response does not contain valid JSON")
	
	// ErrAuthenticationFailed is returned when authentication with the LLM provider fails.
	ErrAuthenticationFailed = errors.New("authentication with LLM provider failed")
	
	// ErrRateLimitExceeded is returned when the LLM provider rate limit is exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	
	// ErrContextTooLong is returned when the input context is too long for the model.
	ErrContextTooLong = errors.New("input context too long")
	
	// ErrProviderUnavailable is returned when the LLM provider is unavailable.
	ErrProviderUnavailable = errors.New("LLM provider unavailable")
	
	// ErrInvalidConfiguration is returned when the provider configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid provider configuration")
	
	// ErrNoResponse is returned when the LLM provider returns no response.
	ErrNoResponse = errors.New("no response from LLM provider")
	
	// ErrTimeout is returned when a request to an LLM provider times out.
	ErrTimeout = errors.New("LLM provider request timed out")
	
	// ErrContentFiltered is returned when content is filtered by the LLM provider.
	ErrContentFiltered = errors.New("content filtered by LLM provider")
	
	// ErrModelNotFound is returned when the requested model is not found.
	ErrModelNotFound = errors.New("model not found")
)

// ProviderError represents an error from an LLM provider with additional context.
type ProviderError struct {
	// Provider is the name of the LLM provider (e.g., "openai", "anthropic").
	Provider string
	
	// Operation is the operation that failed (e.g., "Generate", "Stream").
	Operation string
	
	// StatusCode is the HTTP status code if applicable.
	StatusCode int
	
	// Message is the error message from the provider.
	Message string
	
	// Err is the underlying error.
	Err error
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s %s error (status %d): %s", e.Provider, e.Operation, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("%s %s error: %s", e.Provider, e.Operation, e.Message)
}

// Unwrap returns the underlying error.
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider, operation string, statusCode int, message string, err error) *ProviderError {
	// If no underlying error is provided, use a generic error based on the status code
	if err == nil {
		switch {
		case statusCode == 401:
			err = ErrAuthenticationFailed
		case statusCode == 429:
			err = ErrRateLimitExceeded
		case statusCode >= 500:
			err = ErrProviderUnavailable
		default:
			err = ErrRequestFailed
		}
	}
	
	return &ProviderError{
		Provider:   provider,
		Operation:  operation,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// IsAuthenticationError checks if the error is an authentication error.
func IsAuthenticationError(err error) bool {
	return errors.Is(err, ErrAuthenticationFailed)
}

// IsRateLimitError checks if the error is a rate limit error.
func IsRateLimitError(err error) bool {
	return errors.Is(err, ErrRateLimitExceeded)
}

// IsTimeoutError checks if the error is a timeout error.
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsProviderUnavailableError checks if the error is a provider unavailable error.
func IsProviderUnavailableError(err error) bool {
	return errors.Is(err, ErrProviderUnavailable)
}

// IsContentFilteredError checks if the error is a content filtered error.
func IsContentFilteredError(err error) bool {
	return errors.Is(err, ErrContentFiltered)
}