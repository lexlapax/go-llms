// Package domain defines the core types and interfaces for LLM interactions.
//
// This package contains the fundamental building blocks used throughout the LLM library,
// including message formats, provider interfaces, options, and response types.
//
// # Core Types
//
// Message: Represents a single message in a conversation with role and content.
// Messages can contain text, images, and other multimodal content.
//
// Response: Represents the LLM's response including generated text, token usage,
// and additional metadata.
//
// Option: Functional options for configuring generation requests including
// temperature, max tokens, and other parameters.
//
// # Interfaces
//
// Provider: The primary interface that all LLM providers must implement.
// It defines methods for text generation, message-based chat, and streaming.
//
// ModelRegistry: Interface for providers that support model discovery and listing.
//
// # Message Roles
//
// The package defines standard roles for messages:
//   - System: Instructions that guide the model's behavior
//   - User: Input from the user
//   - Assistant: Responses from the AI model
//   - Tool: Results from function/tool calls
//
// # Content Types
//
// Messages support multiple content types:
//   - Text: Standard text content
//   - Image: Images as base64 or URLs
//   - Audio: Audio content for speech models
//   - Video: Video content for multimodal models
//   - File: General file attachments
//
// # Error Handling
//
// The package defines common error types that providers should use:
//   - Network errors for connectivity issues
//   - Rate limit errors for API throttling
//   - Authentication errors for credential issues
//   - Validation errors for invalid inputs
//
// # Usage Example
//
//	// Create a message
//	msg := domain.Message{
//	    Role:    domain.RoleUser,
//	    Content: "What is the weather like?",
//	}
//
//	// Configure options
//	opts := []domain.Option{
//	    domain.WithTemperature(0.7),
//	    domain.WithMaxTokens(100),
//	}
//
//	// Use with a provider
//	response, err := provider.GenerateMessage(ctx, []domain.Message{msg}, opts...)
package domain
