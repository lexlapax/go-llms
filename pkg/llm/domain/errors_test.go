// ABOUTME: This file contains tests for error handling in the domain package.
// ABOUTME: It verifies error creation, messaging, and type checking functionality.

package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsupportedContentTypeError(t *testing.T) {
	// Create a new unsupported content type error
	provider := "test-provider"
	contentType := ContentTypeVideo
	err := NewUnsupportedContentTypeError(provider, contentType)

	// Test the error message
	expectedMsg := "provider test-provider does not support content type video"
	assert.Equal(t, expectedMsg, err.Error())

	// Test unwrapping the error
	assert.True(t, errors.Is(err, ErrUnsupportedContentType))

	// Test IsUnsupportedContentTypeError
	assert.True(t, IsUnsupportedContentTypeError(err))
	
	// Test that the provider name is in the error message
	assert.True(t, strings.Contains(err.Error(), provider))
	
	// Test that the content type is in the error message
	assert.True(t, strings.Contains(err.Error(), string(contentType)))
}

func TestProviderError(t *testing.T) {
	// Create a new provider error with an underlying unsupported content type error
	baseErr := NewUnsupportedContentTypeError("test-provider", ContentTypeVideo)
	providerErr := NewProviderError(
		"test-provider",
		"GenerateMessage",
		400,
		"Unable to process video content",
		baseErr,
	)

	// Test that the provider error contains information about the operation
	assert.Contains(t, providerErr.Error(), "GenerateMessage")
	
	// Test that the underlying error can be unwrapped and identified
	unwrappedErr := errors.Unwrap(providerErr)
	assert.True(t, errors.Is(unwrappedErr, ErrUnsupportedContentType))
}

func TestContentTypeSupport(t *testing.T) {
	// Test that we can create and identify different content type errors
	videoErr := NewUnsupportedContentTypeError("test-provider", ContentTypeVideo)
	audioErr := NewUnsupportedContentTypeError("test-provider", ContentTypeAudio)
	
	// Test that each error mentions the specific content type
	assert.Contains(t, videoErr.Error(), "video")
	assert.Contains(t, audioErr.Error(), "audio")
	
	// Test that both are recognized as unsupported content type errors
	assert.True(t, IsUnsupportedContentTypeError(videoErr))
	assert.True(t, IsUnsupportedContentTypeError(audioErr))
}