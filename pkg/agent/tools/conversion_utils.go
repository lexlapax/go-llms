// ABOUTME: Provides utilities for bidirectional conversion between agents and tools
// ABOUTME: Includes registry integration, schema mapping, and common conversion patterns

package tools

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ConversionOptions configures agent-tool conversion behavior.
// It provides various customization points for the conversion process,
// including naming, mapping functions, and event handling.
type ConversionOptions struct {
	// Prefix for converted names
	NamePrefix string

	// Whether to auto-generate mappers from schemas
	AutoGenerateMappers bool

	// Event dispatcher for ToolAgent
	EventDispatcher domain.EventDispatcher

	// Custom state mapper
	StateMapper StateMapper

	// Custom result mapper
	ResultMapper ResultMapper

	// Custom param mapper
	ParamMapper ParamMapper

	// Custom state updater
	StateUpdater StateUpdater
}

// Registry Integration Utilities

// RegisterAgentAsTool converts an agent to a tool and registers it in the registry.
// This enables agents to be used in tool-based contexts while preserving their functionality.
//
// Parameters:
//   - agent: The agent to convert and register
//   - registry: The tool registry to register in
//   - opts: Optional conversion configuration
//
// Returns an error if registration fails.
func RegisterAgentAsTool(agent domain.BaseAgent, registry builtins.Registry[domain.Tool], opts ...ConversionOptions) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}
	if registry == nil {
		return fmt.Errorf("registry cannot be nil")
	}

	var opt ConversionOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Create AgentTool wrapper
	agentTool := NewAgentTool(agent)

	// Apply options
	if opt.StateMapper != nil {
		agentTool.WithStateMapper(opt.StateMapper)
	}
	if opt.ResultMapper != nil {
		agentTool.WithResultMapper(opt.ResultMapper)
	}

	// Auto-generate mappers if requested
	if opt.AutoGenerateMappers && agent.InputSchema() != nil {
		stateMapper, resultMapper := GenerateSmartMappers(agent.InputSchema(), agent.OutputSchema())
		if stateMapper != nil {
			agentTool.WithStateMapper(stateMapper)
		}
		if resultMapper != nil {
			agentTool.WithResultMapper(resultMapper)
		}
	}

	// Register the tool
	toolName := agent.Name()
	if opt.NamePrefix != "" {
		toolName = opt.NamePrefix + toolName
	}

	metadata := builtins.Metadata{
		Description: agent.Description(),
		Category:    "agent-wrapper",
		Tags:        []string{"agent", "converted"},
	}

	return registry.Register(toolName, agentTool, metadata)
}

// ConvertToolCategoryToAgents converts all tools in a category to agents.
// This is useful for batch conversion of related tools.
//
// Parameters:
//   - registry: The tool registry to search
//   - category: The category to convert
//   - opts: Optional conversion configuration
//
// Returns a slice of converted agents or an error.
func ConvertToolCategoryToAgents(registry builtins.Registry[domain.Tool], category string, opts ...ConversionOptions) ([]domain.BaseAgent, error) {
	if registry == nil {
		return nil, fmt.Errorf("registry cannot be nil")
	}

	var opt ConversionOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Search for tools in category
	categoryTools := registry.Search(category)
	if len(categoryTools) == 0 {
		return nil, fmt.Errorf("no tools found in category: %s", category)
	}

	agents := make([]domain.BaseAgent, 0, len(categoryTools))
	for _, toolEntry := range categoryTools {
		tool := toolEntry.Component
		if tool == nil {
			return nil, fmt.Errorf("tool component is nil for %s", toolEntry.Metadata.Name)
		}

		// Create ToolAgent wrapper
		toolAgent := createToolAgentWithOptions(tool, opt)
		agents = append(agents, toolAgent)
	}

	return agents, nil
}

// RegisterAgentsAsTools registers multiple agents as tools in a single operation.
// This is a convenience function for batch registration.
//
// Parameters:
//   - agents: The agents to convert and register
//   - registry: The tool registry to register in
//   - opts: Optional conversion configuration
//
// Returns an error if any registration fails.
func RegisterAgentsAsTools(agents []domain.BaseAgent, registry builtins.Registry[domain.Tool], opts ...ConversionOptions) error {
	if registry == nil {
		return fmt.Errorf("registry cannot be nil")
	}

	for _, agent := range agents {
		if err := RegisterAgentAsTool(agent, registry, opts...); err != nil {
			return fmt.Errorf("failed to register agent %s: %w", agent.Name(), err)
		}
	}

	return nil
}

// Event Dispatcher Integration

// NewToolAgentWithEvents creates a ToolAgent with event dispatcher support.
// This enables the agent to emit events during execution.
//
// Parameters:
//   - tool: The tool to wrap as an agent
//   - dispatcher: The event dispatcher for event emission
//
// Returns a configured ToolAgent.
func NewToolAgentWithEvents(tool domain.Tool, dispatcher domain.EventDispatcher) *ToolAgent {
	ta := NewToolAgent(tool)

	// Set the event dispatcher on the ToolAgent
	ta.eventDispatcher = dispatcher

	return ta
}

// CreateEventForwardingToolContext creates a ToolContext that forwards events to a dispatcher.
// This bridges tool events with the agent event system.
//
// Parameters:
//   - ctx: The base context
//   - dispatcher: The event dispatcher to forward to
//   - agent: The agent associated with the context
//   - runID: The execution run ID
//
// Returns a configured ToolContext.
func CreateEventForwardingToolContext(ctx context.Context, dispatcher domain.EventDispatcher, agent domain.BaseAgent, runID string) *domain.ToolContext {
	// Create a basic tool context
	toolCtx := domain.NewToolContext(ctx, nil, agent, runID)

	// Create an event emitter that forwards to the dispatcher
	if dispatcher != nil {
		emitter := &eventForwardingEmitter{
			dispatcher: dispatcher,
			agentID:    agent.ID(),
			agentName:  agent.Name(),
			runID:      runID,
		}
		toolCtx.Events = emitter
	}

	return toolCtx
}

// eventForwardingEmitter implements domain.ToolEventEmitter by forwarding events to a dispatcher.
// It adds agent metadata to all emitted events.
type eventForwardingEmitter struct {
	dispatcher domain.EventDispatcher
	agentID    string
	agentName  string
	runID      string
}

func (e *eventForwardingEmitter) Emit(eventType domain.EventType, data interface{}) {
	if e.dispatcher != nil {
		e.dispatcher.Dispatch(domain.Event{
			Type:      eventType,
			AgentID:   e.agentID,
			AgentName: e.agentName,
			Data:      data,
		})
	}
}

func (e *eventForwardingEmitter) EmitProgress(current, total int, message string) {
	if e.dispatcher != nil {
		e.dispatcher.Dispatch(domain.Event{
			Type:      domain.EventProgress,
			AgentID:   e.agentID,
			AgentName: e.agentName,
			Data: domain.ProgressEventData{
				Current: current,
				Total:   total,
				Message: message,
			},
		})
	}
}

func (e *eventForwardingEmitter) EmitMessage(message string) {
	if e.dispatcher != nil {
		e.dispatcher.Dispatch(domain.Event{
			Type:      domain.EventMessage,
			AgentID:   e.agentID,
			AgentName: e.agentName,
			Data:      message,
		})
	}
}

func (e *eventForwardingEmitter) EmitError(err error) {
	if e.dispatcher != nil {
		e.dispatcher.Dispatch(domain.Event{
			Type:      domain.EventAgentError,
			AgentID:   e.agentID,
			AgentName: e.agentName,
			Error:     err,
		})
	}
}

func (e *eventForwardingEmitter) EmitCustom(eventName string, data interface{}) {
	if e.dispatcher != nil {
		// Use EventMessage with custom data structure
		e.dispatcher.Dispatch(domain.Event{
			Type:      domain.EventMessage,
			AgentID:   e.agentID,
			AgentName: e.agentName,
			Data: map[string]interface{}{
				"type": eventName,
				"data": data,
			},
		})
	}
}

// Schema Mapping Utilities

// DeriveToolSchemaFromAgent generates a tool parameter schema from agent's input schema.
// This helps maintain schema consistency during conversion.
//
// Parameters:
//   - agent: The agent to derive schema from
//
// Returns a tool-compatible schema or nil if no input schema exists.
func DeriveToolSchemaFromAgent(agent domain.BaseAgent) *sdomain.Schema {
	if agent == nil || agent.InputSchema() == nil {
		return nil
	}

	inputSchema := agent.InputSchema()

	// Create a copy of the schema for tool parameters
	toolSchema := &sdomain.Schema{
		Type:        inputSchema.Type,
		Description: fmt.Sprintf("Parameters for %s agent", agent.Name()),
		Properties:  make(map[string]sdomain.Property),
		Required:    inputSchema.Required,
	}

	// Copy properties
	for name, prop := range inputSchema.Properties {
		toolSchema.Properties[name] = prop
	}

	return toolSchema
}

// ValidateConversionCompatibility checks if agent-tool conversion is valid.
// It performs basic compatibility checks including schema validation.
//
// Parameters:
//   - agent: The agent to validate
//   - tool: The tool to validate against
//
// Returns an error if conversion would be invalid.
func ValidateConversionCompatibility(agent domain.BaseAgent, tool domain.Tool) error {
	// Check basic compatibility
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	// Check schema compatibility if available
	if agent.InputSchema() != nil && tool.ParameterSchema() != nil {
		// TODO: Implement detailed schema compatibility checking
		// For now, just check that they're both object schemas
		if agent.InputSchema().Type != "object" || tool.ParameterSchema().Type != "object" {
			return fmt.Errorf("schema type mismatch: agent=%s, tool=%s",
				agent.InputSchema().Type, tool.ParameterSchema().Type)
		}
	}

	return nil
}

// GenerateSmartMappers creates mappers based on schema analysis.
// It intelligently generates state and result mappers by analyzing schemas.
//
// Parameters:
//   - inputSchema: The input schema to analyze
//   - outputSchema: The output schema to analyze
//
// Returns generated StateMapper and ResultMapper functions.
func GenerateSmartMappers(inputSchema, outputSchema *sdomain.Schema) (StateMapper, ResultMapper) {
	var stateMapper StateMapper
	var resultMapper ResultMapper

	// Generate state mapper based on input schema
	if inputSchema != nil && inputSchema.Type == "object" {
		requiredKeys := inputSchema.Required
		stateMapper = func(ctx context.Context, params interface{}) (*domain.State, error) {
			state := domain.NewState()

			// Handle map parameters
			if paramsMap, ok := params.(map[string]interface{}); ok {
				// Check required fields
				for _, key := range requiredKeys {
					if _, exists := paramsMap[key]; !exists {
						return nil, fmt.Errorf("required parameter %s not found", key)
					}
				}

				// Copy all parameters to state
				for k, v := range paramsMap {
					state.Set(k, v)
				}
			} else {
				// For non-map params, use default mapper
				return DefaultStateMapper(ctx, params)
			}

			return state, nil
		}
	}

	// Generate result mapper based on output schema
	if outputSchema != nil && outputSchema.Type == "object" {
		// If output schema has a single required field, extract just that
		if len(outputSchema.Required) == 1 {
			resultKey := outputSchema.Required[0]
			resultMapper = CreateResultMapper(resultKey)
		} else {
			// Otherwise use default mapper
			resultMapper = DefaultResultMapper
		}
	}

	return stateMapper, resultMapper
}

// Common Conversion Patterns

// WrapLLMAgentAsTool wraps an LLM agent as a tool with sensible defaults.
// It handles common LLM agent patterns like prompt/response mapping.
//
// Parameters:
//   - agent: The LLM agent to wrap
//
// Returns a tool implementation or nil if agent is nil.
func WrapLLMAgentAsTool(agent *core.LLMAgent) domain.Tool {
	if agent == nil {
		return nil
	}

	agentTool := NewAgentTool(agent)

	// LLM agents typically expect "prompt" or "input" in state
	agentTool.WithStateMapper(func(ctx context.Context, params interface{}) (*domain.State, error) {
		state := domain.NewState()

		switch p := params.(type) {
		case string:
			state.Set("prompt", p)
		case map[string]interface{}:
			// Look for common prompt keys
			for _, key := range []string{"prompt", "input", "query", "message"} {
				if val, exists := p[key]; exists {
					state.Set("prompt", val)
					break
				}
			}
			// Also copy all other parameters
			for k, v := range p {
				state.Set(k, v)
			}
		default:
			state.Set("input", params)
		}

		return state, nil
	})

	// Extract response from common keys
	agentTool.WithResultMapper(func(ctx context.Context, state *domain.State) (interface{}, error) {
		// Check common response keys
		for _, key := range []string{"response", "output", "result", "answer"} {
			if val, exists := state.Get(key); exists {
				return val, nil
			}
		}

		// Fallback to default
		return DefaultResultMapper(ctx, state)
	})

	return agentTool
}

// WrapWorkflowAgentAsTool wraps a workflow agent as a tool.
// It uses schema-based mapping when available for better conversion.
//
// Parameters:
//   - agent: The workflow agent to wrap
//
// Returns a tool implementation or nil if agent is nil.
func WrapWorkflowAgentAsTool(agent domain.BaseAgent) domain.Tool {
	if agent == nil {
		return nil
	}

	agentTool := NewAgentTool(agent)

	// Workflow agents often need more complex state setup
	if agent.InputSchema() != nil {
		// Use schema-based mapper
		stateMapper, resultMapper := GenerateSmartMappers(agent.InputSchema(), agent.OutputSchema())
		if stateMapper != nil {
			agentTool.WithStateMapper(stateMapper)
		}
		if resultMapper != nil {
			agentTool.WithResultMapper(resultMapper)
		}
	}

	return agentTool
}

// CreateToolChainFromAgents creates a single tool that chains multiple agents.
// The agents are executed sequentially, with each agent's output becoming
// the next agent's input.
//
// Parameters:
//   - agents: The agents to chain together
//
// Returns a tool that executes the agent chain or nil if no agents provided.
func CreateToolChainFromAgents(agents ...domain.BaseAgent) domain.Tool {
	if len(agents) == 0 {
		return nil
	}

	// Create a custom chain agent implementation
	chainAgent := &chainAgentImpl{
		BaseAgentImpl: core.NewBaseAgent(
			"chain",
			fmt.Sprintf("Chain of %d agents", len(agents)),
			domain.AgentTypeCustom,
		),
		agents: agents,
	}

	// Wrap as tool
	return NewAgentTool(chainAgent)
}

// chainAgentImpl implements a simple agent that chains other agents.
// It executes agents sequentially, passing state between them.
type chainAgentImpl struct {
	*core.BaseAgentImpl
	agents []domain.BaseAgent
}

// Run executes all agents in sequence.
// Each agent receives the output state from the previous agent.
//
// Parameters:
//   - ctx: The execution context
//   - state: The initial state
//
// Returns the final state or an error if any agent fails.
func (c *chainAgentImpl) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	currentState := state.Clone()

	for i, agent := range c.agents {
		result, err := agent.Run(ctx, currentState)
		if err != nil {
			return nil, fmt.Errorf("agent %d (%s) failed: %w", i, agent.Name(), err)
		}
		currentState = result
	}

	return currentState, nil
}

// RoundTripConvert validates that an agent can be converted to tool and back.
// This is useful for testing conversion fidelity.
//
// Parameters:
//   - agent: The agent to test
//
// Returns the round-trip converted agent or an error if conversion fails.
func RoundTripConvert(agent domain.BaseAgent) (domain.BaseAgent, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}

	// Convert agent to tool
	tool := NewAgentTool(agent)

	// Convert tool back to agent
	resultAgent := NewToolAgent(tool)

	// Validate the round trip
	if agent.Name() != resultAgent.Name() {
		return nil, fmt.Errorf("name mismatch after round trip: %s != %s", agent.Name(), resultAgent.Name())
	}

	if agent.Description() != resultAgent.Description() {
		return nil, fmt.Errorf("description mismatch after round trip")
	}

	return resultAgent, nil
}

// Advanced Mapping Utilities

// CreatePathMapper creates a mapper that extracts values using paths.
// Paths use dot notation for nested access (e.g., "user.name").
//
// Parameters:
//   - paths: Map of state keys to parameter paths
//
// Returns a StateMapper that extracts values by path.
func CreatePathMapper(paths map[string]string) StateMapper {
	return func(ctx context.Context, params interface{}) (*domain.State, error) {
		state := domain.NewState()

		// Convert params to map if possible
		var paramsMap map[string]interface{}
		switch p := params.(type) {
		case map[string]interface{}:
			paramsMap = p
		default:
			// Try reflection for struct types
			paramsMap = structToMap(params)
		}

		// Extract values using paths
		for statePath, paramPath := range paths {
			value := extractValueByPath(paramsMap, paramPath)
			if value != nil {
				state.Set(statePath, value)
			}
		}

		return state, nil
	}
}

// CreateTypeConversionMapper creates a mapper with type conversions.
// This allows custom transformation of parameter values.
//
// Parameters:
//   - conversions: Map of parameter keys to conversion functions
//
// Returns a StateMapper that applies type conversions.
func CreateTypeConversionMapper(conversions map[string]func(interface{}) interface{}) StateMapper {
	return func(ctx context.Context, params interface{}) (*domain.State, error) {
		// First use default mapper
		state, err := DefaultStateMapper(ctx, params)
		if err != nil {
			return nil, err
		}

		// Apply type conversions
		for key, converter := range conversions {
			if value, exists := state.Get(key); exists && value != nil {
				converted := converter(value)
				state.Set(key, converted)
			}
		}

		return state, nil
	}
}

// CreateNestedStateMapper handles deeply nested state structures.
// It can either preserve nesting or flatten structures.
//
// Parameters:
//   - flatten: If true, nested structures are flattened to dot notation
//
// Returns a StateMapper for handling nested structures.
func CreateNestedStateMapper(flatten bool) StateMapper {
	return func(ctx context.Context, params interface{}) (*domain.State, error) {
		state := domain.NewState()

		if flatten {
			// Flatten nested structures
			flattened := flattenMap(params, "")
			for k, v := range flattened {
				state.Set(k, v)
			}
		} else {
			// Preserve nesting
			return DefaultStateMapper(ctx, params)
		}

		return state, nil
	}
}

// Helper functions

func createToolAgentWithOptions(tool domain.Tool, opts ConversionOptions) *ToolAgent {
	var ta *ToolAgent

	if opts.EventDispatcher != nil {
		ta = NewToolAgentWithEvents(tool, opts.EventDispatcher)
	} else {
		ta = NewToolAgent(tool)
	}

	// Apply mappers
	if opts.ParamMapper != nil {
		ta.WithParamMapper(opts.ParamMapper)
	}
	if opts.StateUpdater != nil {
		ta.WithStateUpdater(opts.StateUpdater)
	}

	return ta
}

func extractValueByPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if current == nil {
			return nil
		}

		if i == len(parts)-1 {
			return current[part]
		}

		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

func structToMap(v interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Use json tag if available
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name
		if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		result[fieldName] = fieldVal.Interface()
	}

	return result
}

func flattenMap(data interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newKey := key
			if prefix != "" {
				newKey = prefix + "." + key
			}

			// Recursively flatten nested maps
			if nestedMap, ok := value.(map[string]interface{}); ok {
				for k, v := range flattenMap(nestedMap, newKey) {
					result[k] = v
				}
			} else {
				result[newKey] = value
			}
		}
	default:
		if prefix != "" {
			result[prefix] = data
		}
	}

	return result
}
