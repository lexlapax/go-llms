package domain

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Tool represents a capability the LLM can invoke
type Tool interface {
	// Name returns the tool's name
	Name() string
	
	// Description provides information about the tool
	Description() string
	
	// Execute runs the tool with parameters
	Execute(ctx context.Context, params interface{}) (interface{}, error)
	
	// ParameterSchema returns the schema for the tool parameters
	ParameterSchema() *domain.Schema
}

// Agent coordinates interactions with LLMs
type Agent interface {
	// Run executes the agent with given inputs
	Run(ctx context.Context, input string) (interface{}, error)
	
	// RunWithSchema executes the agent and validates output against a schema
	RunWithSchema(ctx context.Context, input string, schema *domain.Schema) (interface{}, error)
	
	// AddTool registers a tool with the agent
	AddTool(tool Tool) Agent
	
	// SetSystemPrompt configures the agent's system prompt
	SetSystemPrompt(prompt string) Agent
	
	// WithModel specifies which LLM model to use
	WithModel(modelName string) Agent
	
	// WithHook adds a monitoring hook to the agent
	WithHook(hook Hook) Agent
}