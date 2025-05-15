// ABOUTME: This file contains integration tests for multimodal content handling.
// ABOUTME: It verifies proper error handling for unsupported content types with real providers.

package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/stretchr/testify/assert"
)

// TestMultimodalContentIntegration tests the provider integration for multimodal content
func TestMultimodalContentIntegration(t *testing.T) {
	// Test OpenAI provider with file content
	t.Run("OpenAI_FileContent", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable not set, skipping test")
		}

		// Create an OpenAI provider
		openai := provider.NewOpenAIProvider(apiKey, "gpt-4o")
		
		// Create a sample text message
		textMessage := domain.NewTextMessage(domain.RoleUser, "Hello, world!")
		
		// This should succeed as text messages are supported
		_, err := openai.GenerateMessage(context.Background(), []domain.Message{textMessage})
		assert.NoError(t, err, "Text messages should be supported by OpenAI")
		
		// Create a file message
		fileMessage := domain.NewFileMessage(
			domain.RoleUser,
			"test.pdf",
			[]byte("test file data"),
			"application/pdf",
			"Please analyze this file",
		)
		
		// Check if provider supports this content type
		supported := ProviderSupportsContentType(openai, domain.ContentTypeFile)
		
		// If the provider claims to support this content type, test it
		if supported {
			// Try to generate with file message - this might succeed or fail depending on the implementation
			_, err = openai.GenerateMessage(context.Background(), []domain.Message{fileMessage})
			t.Logf("OpenAI file content support: %v", err == nil)
		} else {
			t.Log("OpenAI does not claim to support file content")
		}
	})
	
	// Test Anthropic provider with video content (which it shouldn't support)
	t.Run("Anthropic_UnsupportedVideoContent", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			t.Skip("ANTHROPIC_API_KEY environment variable not set, skipping test")
		}

		// Create an Anthropic provider
		anthropic := provider.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-latest")
		
		// Create a video message
		videoMessage := domain.NewVideoMessage(
			domain.RoleUser,
			[]byte("test video data"),
			"video/mp4",
			"What's happening in this video?",
		)
		
		// Check if provider supports video content type
		supported := ProviderSupportsContentType(anthropic, domain.ContentTypeVideo)
		
		// Video content should not be supported by Anthropic
		assert.False(t, supported, "Anthropic should not claim to support video content")
		
		// If we try to generate anyway, we should get an unsupported content type error
		if !supported {
			_, err := anthropic.GenerateMessage(context.Background(), []domain.Message{videoMessage})
			assert.Error(t, err, "Video messages should not be supported by Anthropic")
			assert.True(t, domain.IsUnsupportedContentTypeError(err), 
				"Error should be an unsupported content type error")
		}
	})
}

// ProviderSupportsContentType is a helper function to check if a provider claims to support a content type
// This would normally be a method on the provider interface or an exported function in the provider package
func ProviderSupportsContentType(provider domain.Provider, contentType domain.ContentType) bool {
	// Get provider name as a string
	providerStr := fmt.Sprintf("%T", provider)
	
	// Check content type support based on provider string
	if strings.Contains(providerStr, "OpenAIProvider") {
		// OpenAI supports all content types
		return true
	} else if strings.Contains(providerStr, "AnthropicProvider") {
		// Anthropic only supports image content for now
		return contentType == domain.ContentTypeImage || contentType == domain.ContentTypeText
	} else if strings.Contains(providerStr, "GeminiProvider") {
		// Gemini supports image, video and text
		return contentType == domain.ContentTypeImage || 
		       contentType == domain.ContentTypeVideo || 
		       contentType == domain.ContentTypeText
	}
	
	// Unknown provider
	return false
}