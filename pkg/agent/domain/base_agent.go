// ABOUTME: Core interface defining the base contract for all agent types in the system
// ABOUTME: Provides hierarchy management, lifecycle hooks, and execution methods for agents

package domain

import (
	"context"
	"time"

	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BaseAgent defines the core interface that all agent types must implement.
// It provides a common contract for agent identification, hierarchy management,
// execution, lifecycle hooks, and configuration across all agent implementations.
type BaseAgent interface {
	// Identification
	ID() string          // Unique identifier
	Name() string        // Human-readable name
	Description() string // Agent description
	Type() AgentType     // Agent type (LLM, Sequential, Parallel, etc.)

	// Hierarchy Management
	Parent() BaseAgent
	SetParent(parent BaseAgent) error
	SubAgents() []BaseAgent
	AddSubAgent(agent BaseAgent) error
	RemoveSubAgent(name string) error
	FindAgent(name string) BaseAgent
	FindSubAgent(name string) BaseAgent

	// Execution
	Run(ctx context.Context, input *State) (*State, error)
	RunAsync(ctx context.Context, input *State) (<-chan Event, error)

	// Lifecycle Hooks
	Initialize(ctx context.Context) error
	BeforeRun(ctx context.Context, state *State) error
	AfterRun(ctx context.Context, state *State, result *State, err error) error
	Cleanup(ctx context.Context) error

	// Schema Definition
	InputSchema() *sdomain.Schema
	OutputSchema() *sdomain.Schema

	// Configuration
	Config() AgentConfig
	WithConfig(config AgentConfig) BaseAgent
	Validate() error

	// Metadata
	Metadata() map[string]interface{}
	SetMetadata(key string, value interface{})
}

// AgentType represents the category of agent implementation.
// Different types support different execution patterns and behaviors.
type AgentType string

const (
	AgentTypeLLM         AgentType = "llm"
	AgentTypeSequential  AgentType = "sequential"
	AgentTypeParallel    AgentType = "parallel"
	AgentTypeConditional AgentType = "conditional"
	AgentTypeLoop        AgentType = "loop"
	AgentTypeCustom      AgentType = "custom"
)

// AgentConfig holds common configuration options for all agent types.
// It includes timeouts, retry policies, execution modes, and state management settings.
// Agent-specific configurations can be stored in the Custom map.
type AgentConfig struct {
	// Common configuration
	Timeout    time.Duration `json:"timeout,omitempty"`
	MaxRetries int           `json:"max_retries,omitempty"`
	RetryDelay time.Duration `json:"retry_delay,omitempty"`

	// Execution configuration
	Async        bool `json:"async,omitempty"`
	StreamEvents bool `json:"stream_events,omitempty"`

	// State configuration
	ShareState   bool `json:"share_state,omitempty"`
	IsolateState bool `json:"isolate_state,omitempty"`

	// Custom configuration
	Custom map[string]interface{} `json:"custom,omitempty"`
}
