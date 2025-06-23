// Package domain provides core types and interfaces for the agent framework.
// This file defines the fundamental contracts for tools, agents, and registries
// that form the foundation of the agent system architecture.
package domain

// ABOUTME: Core interfaces for agent system including Tool and Agent definitions
// ABOUTME: Establishes contracts for tool execution and agent workflows

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Tool represents an executable capability that can be invoked by LLMs.
// Tools provide structured interfaces for LLMs to interact with external systems,
// APIs, and perform specific tasks with clear schemas and usage guidance.
// Each tool defines its parameters, outputs, constraints, and provides examples.
type Tool interface {
	// Core functionality
	Name() string                                                      // Unique identifier for the tool
	Description() string                                               // Brief description of what the tool does
	Execute(ctx *ToolContext, params interface{}) (interface{}, error) // Execute the tool with given parameters

	// Schema definitions
	ParameterSchema() *domain.Schema // JSON Schema for input parameters
	OutputSchema() *domain.Schema    // JSON Schema for output structure

	// LLM guidance
	UsageInstructions() string        // Detailed instructions on when and how to use the tool
	Examples() []ToolExample          // Concrete examples showing tool usage
	Constraints() []string            // Limitations and constraints of the tool
	ErrorGuidance() map[string]string // Map of error types to helpful guidance

	// Metadata
	Category() string // Category for grouping (e.g., "math", "web", "file")
	Tags() []string   // Tags for discovery and filtering
	Version() string  // Tool version for compatibility tracking

	// Behavioral hints
	IsDeterministic() bool      // Same input always produces same output
	IsDestructive() bool        // Tool modifies state or has side effects
	RequiresConfirmation() bool // User confirmation needed before execution
	EstimatedLatency() string   // Expected execution time: "fast", "medium", "slow"

	// MCP compatibility
	ToMCPDefinition() MCPToolDefinition // Export tool definition in MCP format
}

// Agent coordinates interactions between LLMs, tools, and workflows.
// Agents manage the execution flow, handle tool invocations, maintain state,
// and orchestrate complex multi-step processes with support for schema validation.
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

// AgentRegistry provides agent discovery and management capabilities.
// It maintains a registry of available agents with lookup by ID or name
// and supports listing all registered agents for discovery.
type AgentRegistry interface {
	// Register an agent
	Register(agent BaseAgent) error
	// Get agent by ID
	Get(agentID string) (BaseAgent, error)
	// Get agent by name
	GetByName(name string) (BaseAgent, error)
	// List all agents
	List() []BaseAgent
}

// ToolExample provides concrete usage examples for LLMs.
// Examples help language models understand when and how to use tools
// by showing realistic scenarios with expected inputs and outputs.
type ToolExample struct {
	Name        string      `json:"name"`        // Short name for the example
	Description string      `json:"description"` // What this example demonstrates
	Scenario    string      `json:"scenario"`    // When to use this approach
	Input       interface{} `json:"input"`       // Example input parameters
	Output      interface{} `json:"output"`      // Expected output
	Explanation string      `json:"explanation"` // Why this works and what to learn
}

// MCPToolDefinition represents a tool in Model Context Protocol format.
// This provides compatibility with MCP-compatible clients and systems
// by exposing tools in the standard MCP tool definition structure.
type MCPToolDefinition struct {
	Name         string                 `json:"name"`                   // Tool identifier
	Description  string                 `json:"description"`            // Tool description
	InputSchema  interface{}            `json:"inputSchema,omitempty"`  // Parameter schema
	OutputSchema interface{}            `json:"outputSchema,omitempty"` // Output schema
	Annotations  map[string]interface{} `json:"annotations,omitempty"`  // Additional metadata
}
