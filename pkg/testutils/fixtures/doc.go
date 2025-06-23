// Package fixtures provides test data and utilities for testing LLM integrations.
//
// This package contains reusable test fixtures, mock responses, and helper
// functions to facilitate testing of LLM providers, agents, and tools without
// making actual API calls or incurring costs.
//
// Fixture Categories:
//
// Provider Fixtures:
//   - Mock API responses for each provider (OpenAI, Anthropic, etc.)
//   - Error response scenarios
//   - Rate limit responses
//   - Streaming response chunks
//
// Message Fixtures:
//   - Sample conversations in various formats
//   - Multi-turn dialogues
//   - System prompts and templates
//   - Tool calling examples
//
// Model Fixtures:
//   - Model metadata and capabilities
//   - Pricing information
//   - Token counting examples
//   - Context window limits
//
// Response Fixtures:
//   - Successful completion responses
//   - Partial responses
//   - Error responses
//   - Timeout scenarios
//
// Usage Patterns:
//
//	// Load a standard fixture
//	response := fixtures.OpenAICompletionResponse()
//
//	// Create custom test scenarios
//	messages := fixtures.NewConversation().
//	    WithSystemPrompt("You are a helpful assistant").
//	    WithUserMessage("Hello").
//	    WithAssistantMessage("Hi! How can I help?").
//	    Build()
//
//	// Mock provider responses
//	mockProvider := fixtures.NewMockProvider().
//	    WithResponse("prompt1", "response1").
//	    WithError("error_prompt", errors.New("API error")).
//	    Build()
//
// Test Helpers:
//
//	// Compare responses ignoring timestamps
//	fixtures.AssertResponsesEqual(t, expected, actual)
//
//	// Validate message format
//	fixtures.ValidateMessages(t, messages)
//
//	// Generate random test data
//	testData := fixtures.RandomConversation(10) // 10 turns
//
// File Organization:
//   - providers/: Provider-specific response fixtures
//   - messages/: Conversation and message examples
//   - responses/: Common response patterns
//   - errors/: Error scenarios and edge cases
//   - helpers/: Test utility functions
//
// Best Practices:
//   - Use fixtures to avoid API calls in tests
//   - Create minimal fixtures for specific test cases
//   - Reuse common patterns across tests
//   - Document any provider-specific quirks
//   - Keep fixtures up-to-date with API changes
package fixtures
