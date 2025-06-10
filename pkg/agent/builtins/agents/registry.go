// ABOUTME: Agent-specific registry with template system for pre-configured agents
// ABOUTME: Provides discovery and composition helpers for agent templates

package agents

import (
	"fmt"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	lldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// AgentRegistry extends the base registry with agent-specific functionality
type AgentRegistry interface {
	builtins.Registry[AgentTemplate]

	// RegisterAgent is a convenience method that accepts AgentMetadata
	RegisterAgent(name string, template AgentTemplate, metadata AgentMetadata) error

	// ListByCapability returns agents with specific capability
	ListByCapability(capability string) []builtins.RegistryEntry[AgentTemplate]

	// ListRequiringTool returns agents that require a specific tool
	ListRequiringTool(toolName string) []builtins.RegistryEntry[AgentTemplate]
}

// AgentTemplate provides a pre-configured agent
type AgentTemplate interface {
	// Build creates a new agent instance with the template's configuration
	Build(provider lldomain.Provider, opts ...AgentOption) domain.Agent

	// Metadata returns information about this template
	Metadata() AgentMetadata
}

// AgentMetadata extends base metadata for agents
type AgentMetadata struct {
	builtins.Metadata
	RequiredTools []string `json:"required_tools,omitempty"` // Names of required tools
	OptionalTools []string `json:"optional_tools,omitempty"` // Names of optional tools
	Capabilities  []string `json:"capabilities,omitempty"`   // What the agent can do
	SystemPrompt  string   `json:"system_prompt,omitempty"`  // Default system prompt
}

// AgentOption allows customization of template agents
type AgentOption func(*agentConfig)

// agentConfig holds configuration options for building agents
type agentConfig struct {
	additionalTools []string
	customPrompt    string
	modelOverride   string
	includeAllTools bool
	hooks           []domain.Hook
}

// Common agent options

// WithAdditionalTools adds extra tools beyond the template's defaults
func WithAdditionalTools(toolNames ...string) AgentOption {
	return func(cfg *agentConfig) {
		cfg.additionalTools = append(cfg.additionalTools, toolNames...)
	}
}

// WithCustomPrompt overrides the template's system prompt
func WithCustomPrompt(prompt string) AgentOption {
	return func(cfg *agentConfig) {
		cfg.customPrompt = prompt
	}
}

// WithModelOverride specifies a different model than the default
func WithModelOverride(model string) AgentOption {
	return func(cfg *agentConfig) {
		cfg.modelOverride = model
	}
}

// WithAllOptionalTools includes all optional tools defined in the template
func WithAllOptionalTools() AgentOption {
	return func(cfg *agentConfig) {
		cfg.includeAllTools = true
	}
}

// WithHooks adds monitoring hooks to the agent
func WithHooks(hooks ...domain.Hook) AgentOption {
	return func(cfg *agentConfig) {
		cfg.hooks = append(cfg.hooks, hooks...)
	}
}

// agentRegistry implements AgentRegistry
type agentRegistry struct {
	builtins.Registry[AgentTemplate]
}

// Agents is the global registry for built-in agent templates
var Agents AgentRegistry = &agentRegistry{
	Registry: builtins.NewRegistry[AgentTemplate](),
}

// RegisterAgent registers an agent template with agent-specific metadata
func (r *agentRegistry) RegisterAgent(name string, template AgentTemplate, metadata AgentMetadata) error {
	// Validate agent metadata
	if err := validateAgentMetadata(name, template, metadata); err != nil {
		return fmt.Errorf("invalid agent metadata: %w", err)
	}

	// Convert AgentMetadata to base Metadata for registration
	baseMetadata := metadata.Metadata

	// Store metadata in template for later retrieval
	// This would typically be done by having the template store the metadata

	return r.Register(name, template, baseMetadata)
}

// ListByCapability returns agents with specific capability
func (r *agentRegistry) ListByCapability(capability string) []builtins.RegistryEntry[AgentTemplate] {
	allAgents := r.List()
	var filtered []builtins.RegistryEntry[AgentTemplate]

	for _, entry := range allAgents {
		// Get metadata from template
		metadata := entry.Component.Metadata()
		for _, cap := range metadata.Capabilities {
			if cap == capability {
				filtered = append(filtered, entry)
				break
			}
		}
	}

	return filtered
}

// ListRequiringTool returns agents that require a specific tool
func (r *agentRegistry) ListRequiringTool(toolName string) []builtins.RegistryEntry[AgentTemplate] {
	allAgents := r.List()
	var filtered []builtins.RegistryEntry[AgentTemplate]

	for _, entry := range allAgents {
		// Get metadata from template
		metadata := entry.Component.Metadata()
		for _, tool := range metadata.RequiredTools {
			if tool == toolName {
				filtered = append(filtered, entry)
				break
			}
		}
	}

	return filtered
}

// validateAgentMetadata ensures the agent metadata is valid
func validateAgentMetadata(name string, template AgentTemplate, metadata AgentMetadata) error {
	// Ensure name matches
	if metadata.Name != "" && metadata.Name != name {
		return fmt.Errorf("metadata name '%s' does not match registration name '%s'", metadata.Name, name)
	}

	// Validate required tools are not duplicated in optional tools
	requiredSet := make(map[string]bool)
	for _, tool := range metadata.RequiredTools {
		if tool == "" {
			return fmt.Errorf("empty required tool name")
		}
		requiredSet[tool] = true
	}

	for _, tool := range metadata.OptionalTools {
		if tool == "" {
			return fmt.Errorf("empty optional tool name")
		}
		if requiredSet[tool] {
			return fmt.Errorf("tool '%s' cannot be both required and optional", tool)
		}
	}

	// Validate capabilities
	for _, cap := range metadata.Capabilities {
		if cap == "" {
			return fmt.Errorf("empty capability string")
		}
	}

	return nil
}

// Helper functions

// MustRegisterAgent registers an agent template or panics on error
func MustRegisterAgent(name string, template AgentTemplate, metadata AgentMetadata) {
	if err := Agents.RegisterAgent(name, template, metadata); err != nil {
		panic(fmt.Sprintf("failed to register agent '%s': %v", name, err))
	}
}

// GetAgent retrieves an agent template by name
func GetAgent(name string) (AgentTemplate, bool) {
	return Agents.Get(name)
}

// MustGetAgent retrieves an agent template by name or panics
func MustGetAgent(name string) AgentTemplate {
	return Agents.MustGet(name)
}

// BuildAgent creates an agent instance from a template name
func BuildAgent(name string, provider lldomain.Provider, opts ...AgentOption) (domain.Agent, error) {
	template, found := GetAgent(name)
	if !found {
		return nil, fmt.Errorf("agent template '%s' not found", name)
	}
	return template.Build(provider, opts...), nil
}

// MustBuildAgent creates an agent instance or panics
func MustBuildAgent(name string, provider lldomain.Provider, opts ...AgentOption) domain.Agent {
	agent, err := BuildAgent(name, provider, opts...)
	if err != nil {
		panic(err)
	}
	return agent
}
