// ABOUTME: Centralized message fixtures for testing
// ABOUTME: Provides reusable message creation functions for various test scenarios

package fixtures

import (
	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// CreateSampleMessages creates test message arrays of different sizes
// This is useful for testing message handling and performance
func CreateSampleMessages(size int) []domain.Message {
	// Pre-allocate the slice
	messages := make([]domain.Message, 0, size)

	// Always start with a system message
	messages = append(messages, domain.NewTextMessage(domain.RoleSystem, "You are a helpful assistant that provides accurate and concise information."))

	// Add user-assistant message pairs
	for i := 1; i < size; i += 2 {
		// Add user message
		if i < size {
			messages = append(messages, domain.NewTextMessage(domain.RoleUser, "This is a user message for testing performance."))
		}

		// Add assistant message
		if i+1 < size {
			messages = append(messages, domain.NewTextMessage(domain.RoleAssistant, "This is an assistant response for testing performance."))
		}
	}

	return messages
}

// CreateMessagesWithTools creates test message arrays that include tool calls
// This is useful for testing tool-related functionality
func CreateMessagesWithTools(size int) []domain.Message {
	// Pre-allocate the slice
	messages := make([]domain.Message, 0, size)

	// Add system message
	messages = append(messages, domain.NewTextMessage(domain.RoleSystem, "You are a helpful assistant that can use tools."))

	// Add user message first
	messages = append(messages, domain.NewTextMessage(domain.RoleUser, "I need help with a calculation."))

	// Fill remaining messages with assistant-tool pairs
	remaining := size - 2
	for i := 0; i < remaining; i += 2 {
		// Add assistant message with tool call
		if i < remaining {
			messages = append(messages, domain.NewTextMessage(domain.RoleAssistant, "I'll use the calculator tool to help you."))
		}

		// Add tool response
		if i+1 < remaining {
			messages = append(messages, domain.NewTextMessage(domain.RoleTool, "Result: 42"))
		}
	}

	return messages
}

// CreateSimpleConversation creates a basic user-assistant conversation
func CreateSimpleConversation() []domain.Message {
	return []domain.Message{
		domain.NewTextMessage(domain.RoleSystem, "You are a helpful assistant."),
		domain.NewTextMessage(domain.RoleUser, "Hello, how are you?"),
		domain.NewTextMessage(domain.RoleAssistant, "I'm doing well, thank you for asking! How can I help you today?"),
	}
}

// CreateMultiTurnConversation creates a multi-turn conversation with context
func CreateMultiTurnConversation() []domain.Message {
	return []domain.Message{
		domain.NewTextMessage(domain.RoleSystem, "You are a knowledgeable assistant."),
		domain.NewTextMessage(domain.RoleUser, "What is the capital of France?"),
		domain.NewTextMessage(domain.RoleAssistant, "The capital of France is Paris."),
		domain.NewTextMessage(domain.RoleUser, "What is its population?"),
		domain.NewTextMessage(domain.RoleAssistant, "Paris has a population of approximately 2.2 million people in the city proper, and over 12 million in the greater metropolitan area."),
		domain.NewTextMessage(domain.RoleUser, "What are some famous landmarks there?"),
		domain.NewTextMessage(domain.RoleAssistant, "Some famous landmarks in Paris include the Eiffel Tower, the Louvre Museum, Notre-Dame Cathedral, the Arc de Triomphe, and the Champs-Élysées."),
	}
}

// CreateToolCallingConversation creates a conversation with tool calls
func CreateToolCallingConversation() []domain.Message {
	return []domain.Message{
		domain.NewTextMessage(domain.RoleSystem, "You are an assistant with access to tools."),
		domain.NewTextMessage(domain.RoleUser, "What is 25 * 4?"),
		domain.NewTextMessage(domain.RoleAssistant, "I'll calculate that for you."),
		domain.NewTextMessage(domain.RoleTool, "100"),
		domain.NewTextMessage(domain.RoleAssistant, "25 multiplied by 4 equals 100."),
	}
}

// MessageSizes provides common test sizes
var MessageSizes = struct {
	Small  int
	Medium int
	Large  int
}{
	Small:  3,  // System + User + Assistant
	Medium: 7,  // System + 3 exchanges
	Large:  21, // System + 10 exchanges
}
