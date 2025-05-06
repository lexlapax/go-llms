// Package domain defines the core domain models and interfaces for LLM providers.
package domain

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a message in a conversation
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Token represents a token in a streamed response
type Token struct {
	Text     string `json:"text"`
	Finished bool   `json:"finished"`
}

// Response represents a complete response from an LLM
type Response struct {
	Content string `json:"content"`
}

// ResponseStream represents a stream of tokens from an LLM
type ResponseStream <-chan Token