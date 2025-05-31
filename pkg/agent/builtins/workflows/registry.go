// ABOUTME: Workflow-specific registry with pattern system for multi-agent workflows
// ABOUTME: Provides workflow builder utilities and routing patterns

package workflows

import (
	"context"
	"fmt"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// WorkflowRegistry extends the base registry with workflow-specific functionality
type WorkflowRegistry interface {
	builtins.Registry[WorkflowPattern]

	// RegisterWorkflow is a convenience method that accepts WorkflowMetadata
	RegisterWorkflow(name string, pattern WorkflowPattern, metadata WorkflowMetadata) error

	// ListByStageCount returns workflows with specific number of stages
	ListByStageCount(minStages, maxStages int) []builtins.RegistryEntry[WorkflowPattern]

	// ListRequiringAgent returns workflows that require a specific agent
	ListRequiringAgent(agentName string) []builtins.RegistryEntry[WorkflowPattern]
}

// WorkflowPattern defines a reusable workflow structure
type WorkflowPattern interface {
	// Build creates a workflow instance
	Build(opts ...WorkflowOption) Workflow

	// Metadata returns information about this pattern
	Metadata() WorkflowMetadata
}

// WorkflowMetadata extends base metadata for workflows
type WorkflowMetadata struct {
	builtins.Metadata
	RequiredAgents []string `json:"required_agents,omitempty"` // Names of required agents
	OptionalAgents []string `json:"optional_agents,omitempty"` // Names of optional agents
	Stages         []string `json:"stages,omitempty"`          // Workflow stages in order
	RoutingType    string   `json:"routing_type,omitempty"`    // Type of routing (sequential, parallel, conditional)
}

// Workflow represents an executable workflow
type Workflow interface {
	// Execute runs the workflow
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// AddAgent adds an agent to the workflow
	AddAgent(name string, agent domain.Agent) error

	// SetRouter configures routing between agents
	SetRouter(router Router) error

	// GetStages returns the workflow stages
	GetStages() []string
}

// Router determines how data flows between agents
type Router interface {
	// Route determines next agent and transforms data
	Route(ctx context.Context, from string, result interface{}) (nextAgent string, transformedInput interface{}, err error)

	// IsComplete checks if workflow should terminate
	IsComplete(ctx context.Context, lastAgent string, result interface{}) bool
}

// WorkflowOption allows customization of workflow patterns
type WorkflowOption func(*workflowConfig)

// workflowConfig holds configuration options for building workflows
type workflowConfig struct {
	agentOverrides  map[string]domain.Agent
	routerOverride  Router
	maxIterations   int
	timeoutOverride *int
	metadata        map[string]interface{}
}

// Common workflow options

// WithAgentOverride replaces a specific agent in the workflow
func WithAgentOverride(stageName string, agent domain.Agent) WorkflowOption {
	return func(cfg *workflowConfig) {
		if cfg.agentOverrides == nil {
			cfg.agentOverrides = make(map[string]domain.Agent)
		}
		cfg.agentOverrides[stageName] = agent
	}
}

// WithRouterOverride uses a custom router
func WithRouterOverride(router Router) WorkflowOption {
	return func(cfg *workflowConfig) {
		cfg.routerOverride = router
	}
}

// WithMaxIterations sets maximum workflow iterations
func WithMaxIterations(max int) WorkflowOption {
	return func(cfg *workflowConfig) {
		cfg.maxIterations = max
	}
}

// WithTimeout sets workflow timeout in seconds
func WithTimeout(seconds int) WorkflowOption {
	return func(cfg *workflowConfig) {
		cfg.timeoutOverride = &seconds
	}
}

// WithMetadata adds custom metadata to the workflow
func WithMetadata(key string, value interface{}) WorkflowOption {
	return func(cfg *workflowConfig) {
		if cfg.metadata == nil {
			cfg.metadata = make(map[string]interface{})
		}
		cfg.metadata[key] = value
	}
}

// workflowRegistry implements WorkflowRegistry
type workflowRegistry struct {
	builtins.Registry[WorkflowPattern]
}

// Workflows is the global registry for built-in workflow patterns
var Workflows WorkflowRegistry = &workflowRegistry{
	Registry: builtins.NewRegistry[WorkflowPattern](),
}

// RegisterWorkflow registers a workflow pattern with workflow-specific metadata
func (r *workflowRegistry) RegisterWorkflow(name string, pattern WorkflowPattern, metadata WorkflowMetadata) error {
	// Validate workflow metadata
	if err := validateWorkflowMetadata(name, pattern, metadata); err != nil {
		return fmt.Errorf("invalid workflow metadata: %w", err)
	}

	// Convert WorkflowMetadata to base Metadata for registration
	baseMetadata := metadata.Metadata

	return r.Registry.Register(name, pattern, baseMetadata)
}

// ListByStageCount returns workflows with specific number of stages
func (r *workflowRegistry) ListByStageCount(minStages, maxStages int) []builtins.RegistryEntry[WorkflowPattern] {
	allWorkflows := r.List()
	var filtered []builtins.RegistryEntry[WorkflowPattern]

	for _, entry := range allWorkflows {
		metadata := entry.Component.Metadata()
		stageCount := len(metadata.Stages)
		if stageCount >= minStages && stageCount <= maxStages {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// ListRequiringAgent returns workflows that require a specific agent
func (r *workflowRegistry) ListRequiringAgent(agentName string) []builtins.RegistryEntry[WorkflowPattern] {
	allWorkflows := r.List()
	var filtered []builtins.RegistryEntry[WorkflowPattern]

	for _, entry := range allWorkflows {
		metadata := entry.Component.Metadata()
		for _, agent := range metadata.RequiredAgents {
			if agent == agentName {
				filtered = append(filtered, entry)
				break
			}
		}
	}

	return filtered
}

// validateWorkflowMetadata ensures the workflow metadata is valid
func validateWorkflowMetadata(name string, pattern WorkflowPattern, metadata WorkflowMetadata) error {
	// Ensure name matches
	if metadata.Name != "" && metadata.Name != name {
		return fmt.Errorf("metadata name '%s' does not match registration name '%s'", metadata.Name, name)
	}

	// Validate stages
	if len(metadata.Stages) == 0 {
		return fmt.Errorf("workflow must have at least one stage")
	}

	for i, stage := range metadata.Stages {
		if stage == "" {
			return fmt.Errorf("empty stage name at position %d", i)
		}
	}

	// Validate required agents are not duplicated in optional agents
	requiredSet := make(map[string]bool)
	for _, agent := range metadata.RequiredAgents {
		if agent == "" {
			return fmt.Errorf("empty required agent name")
		}
		requiredSet[agent] = true
	}

	for _, agent := range metadata.OptionalAgents {
		if agent == "" {
			return fmt.Errorf("empty optional agent name")
		}
		if requiredSet[agent] {
			return fmt.Errorf("agent '%s' cannot be both required and optional", agent)
		}
	}

	// Validate routing type
	if metadata.RoutingType != "" {
		validTypes := map[string]bool{
			"sequential":  true,
			"parallel":    true,
			"conditional": true,
			"custom":      true,
		}
		if !validTypes[metadata.RoutingType] {
			return fmt.Errorf("invalid routing type: %s", metadata.RoutingType)
		}
	}

	return nil
}

// Helper functions

// MustRegisterWorkflow registers a workflow pattern or panics on error
func MustRegisterWorkflow(name string, pattern WorkflowPattern, metadata WorkflowMetadata) {
	if err := Workflows.RegisterWorkflow(name, pattern, metadata); err != nil {
		panic(fmt.Sprintf("failed to register workflow '%s': %v", name, err))
	}
}

// GetWorkflow retrieves a workflow pattern by name
func GetWorkflow(name string) (WorkflowPattern, bool) {
	return Workflows.Get(name)
}

// MustGetWorkflow retrieves a workflow pattern by name or panics
func MustGetWorkflow(name string) WorkflowPattern {
	return Workflows.MustGet(name)
}

// BuildWorkflow creates a workflow instance from a pattern name
func BuildWorkflow(name string, opts ...WorkflowOption) (Workflow, error) {
	pattern, found := GetWorkflow(name)
	if !found {
		return nil, fmt.Errorf("workflow pattern '%s' not found", name)
	}
	return pattern.Build(opts...), nil
}

// MustBuildWorkflow creates a workflow instance or panics
func MustBuildWorkflow(name string, opts ...WorkflowOption) Workflow {
	workflow, err := BuildWorkflow(name, opts...)
	if err != nil {
		panic(err)
	}
	return workflow
}

// Common Router implementations

// SequentialRouter routes through stages in order
func SequentialRouter(stages []string) Router {
	return &sequentialRouter{stages: stages}
}

type sequentialRouter struct {
	stages       []string
	currentIndex int
}

func (r *sequentialRouter) Route(ctx context.Context, from string, result interface{}) (string, interface{}, error) {
	// Find current stage
	for i, stage := range r.stages {
		if stage == from {
			if i+1 < len(r.stages) {
				return r.stages[i+1], result, nil
			}
			break
		}
	}
	return "", nil, nil // No next stage
}

func (r *sequentialRouter) IsComplete(ctx context.Context, lastAgent string, result interface{}) bool {
	return lastAgent == r.stages[len(r.stages)-1]
}

// ConditionalRouter routes based on conditions
type ConditionalRouter struct {
	conditions map[string]func(interface{}) string
	fallback   string
}

func (r *ConditionalRouter) Route(ctx context.Context, from string, result interface{}) (string, interface{}, error) {
	if condition, exists := r.conditions[from]; exists {
		next := condition(result)
		if next != "" {
			return next, result, nil
		}
	}
	if r.fallback != "" {
		return r.fallback, result, nil
	}
	return "", nil, nil
}

func (r *ConditionalRouter) IsComplete(ctx context.Context, lastAgent string, result interface{}) bool {
	// Conditional routing is complete when no next agent is determined
	next, _, _ := r.Route(ctx, lastAgent, result)
	return next == ""
}