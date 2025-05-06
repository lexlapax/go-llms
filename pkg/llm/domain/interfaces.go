package domain

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Provider defines the contract for LLM providers
type Provider interface {
	// Generate produces text from a prompt
	Generate(ctx context.Context, prompt string, options ...Option) (string, error)
	
	// GenerateMessage produces text from a list of messages
	GenerateMessage(ctx context.Context, messages []Message, options ...Option) (Response, error)
	
	// GenerateWithSchema produces structured output conforming to a schema
	GenerateWithSchema(ctx context.Context, prompt string, schema *domain.Schema, options ...Option) (interface{}, error)
	
	// Stream streams responses token by token
	Stream(ctx context.Context, prompt string, options ...Option) (ResponseStream, error)
	
	// StreamMessage streams responses from a list of messages
	StreamMessage(ctx context.Context, messages []Message, options ...Option) (ResponseStream, error)
}

// ModelRegistry manages available LLM models
type ModelRegistry interface {
	// RegisterModel adds a model to the registry
	RegisterModel(name string, provider Provider) error
	
	// GetModel retrieves a model by name
	GetModel(name string) (Provider, error)
	
	// ListModels returns all available models
	ListModels() []string
}