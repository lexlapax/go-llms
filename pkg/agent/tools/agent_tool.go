// ABOUTME: AgentTool wraps a BaseAgent to expose it as a Tool, enabling agents to be used as tools
// ABOUTME: This bridges the State-based agent system with the parameter-based tool system

package tools

import (
	"context"
	"fmt"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// AgentTool wraps a BaseAgent to expose it as a Tool.
// This allows agents to be used within the tool ecosystem, enabling
// composition of agents as tools within other agents or workflows.
type AgentTool struct {
	agent        domain.BaseAgent
	paramSchema  *sdomain.Schema
	stateMapper  StateMapper
	resultMapper ResultMapper
}

// StateMapper converts tool parameters to agent State.
// It bridges the gap between parameter-based tool inputs and state-based agent inputs.
type StateMapper func(ctx context.Context, params interface{}) (*domain.State, error)

// ResultMapper converts agent State result to tool result.
// It extracts the relevant output from the agent's result state for tool consumers.
type ResultMapper func(ctx context.Context, state *domain.State) (interface{}, error)

// NewAgentTool creates a new AgentTool wrapper with default mappers.
// The default state mapper handles common parameter formats, and the default
// result mapper extracts standard result fields.
//
// Parameters:
//   - agent: The agent to wrap as a tool
//
// Returns a new AgentTool instance.
func NewAgentTool(agent domain.BaseAgent) *AgentTool {
	return &AgentTool{
		agent:        agent,
		stateMapper:  DefaultStateMapper,
		resultMapper: DefaultResultMapper,
	}
}

// WithParameterSchema sets the parameter schema for the tool.
// This overrides any schema derived from the agent.
//
// Parameters:
//   - schema: The parameter schema to use
//
// Returns the AgentTool for method chaining.
func (at *AgentTool) WithParameterSchema(schema *sdomain.Schema) *AgentTool {
	at.paramSchema = schema
	return at
}

// WithStateMapper sets a custom state mapper.
// This allows customization of how tool parameters are converted to agent state.
//
// Parameters:
//   - mapper: The custom state mapper function
//
// Returns the AgentTool for method chaining.
func (at *AgentTool) WithStateMapper(mapper StateMapper) *AgentTool {
	at.stateMapper = mapper
	return at
}

// WithResultMapper sets a custom result mapper.
// This allows customization of how agent results are extracted and formatted.
//
// Parameters:
//   - mapper: The custom result mapper function
//
// Returns the AgentTool for method chaining.
func (at *AgentTool) WithResultMapper(mapper ResultMapper) *AgentTool {
	at.resultMapper = mapper
	return at
}

// Name returns the tool name, which is derived from the wrapped agent's name.
func (at *AgentTool) Name() string {
	return at.agent.Name()
}

// Description returns the tool description, which is derived from the wrapped agent's description.
func (at *AgentTool) Description() string {
	return at.agent.Description()
}

// Execute runs the agent with mapped parameters.
// It performs the following steps:
// 1. Maps tool parameters to agent state
// 2. Merges any additional state from the tool context
// 3. Runs the agent
// 4. Maps the agent's result state to tool output
//
// Parameters:
//   - ctx: The tool execution context
//   - params: The tool parameters
//
// Returns the mapped result or an error.
func (at *AgentTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	// Map parameters to State
	state, err := at.stateMapper(ctx.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to map parameters to state: %w", err)
	}

	// Merge any state from the tool context
	// This allows the calling agent to pass configuration
	if ctx.State != nil {
		for _, key := range ctx.State.Keys() {
			if _, alreadyExists := state.Get(key); !alreadyExists { // Don't override mapped params
				if val, exists := ctx.State.Get(key); exists {
					state.Set(key, val)
				}
			}
		}
	}

	// Run the agent with the standard context
	// The agent will create its own internal state management
	resultState, err := at.agent.Run(ctx.Context, state)
	if err != nil {
		// Emit error event if possible
		if ctx.Events != nil {
			ctx.Events.EmitError(err)
		}
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// Map result State to tool result
	result, err := at.resultMapper(ctx.Context, resultState)
	if err != nil {
		return nil, fmt.Errorf("failed to map state to result: %w", err)
	}

	// Emit success event if we have an emitter
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Agent '%s' executed successfully", at.agent.Name()))
	}

	return result, nil
}

// ParameterSchema returns the tool's parameter schema.
// It uses the explicitly set schema if available, otherwise derives from
// the agent's input schema, or returns a generic object schema.
func (at *AgentTool) ParameterSchema() *sdomain.Schema {
	if at.paramSchema != nil {
		return at.paramSchema
	}

	// Try to derive from agent's input schema
	if at.agent.InputSchema() != nil {
		return at.agent.InputSchema()
	}

	// Return a generic object schema
	return &sdomain.Schema{
		Type:        "object",
		Description: fmt.Sprintf("Parameters for %s agent", at.agent.Name()),
	}
}

// DefaultStateMapper converts parameters to State using a simple mapping.
// It handles:
// - map[string]interface{}: Direct mapping of keys to state
// - string: Stored as "input" key
// - *domain.State: Cloned and returned
// - Other types: Stored as "params" key
//
// Parameters:
//   - ctx: Context for the operation
//   - params: The parameters to map
//
// Returns the mapped state or an error.
func DefaultStateMapper(ctx context.Context, params interface{}) (*domain.State, error) {
	state := domain.NewState()

	switch p := params.(type) {
	case map[string]interface{}:
		// Direct mapping of map parameters
		for k, v := range p {
			state.Set(k, v)
		}
	case string:
		// Single string parameter goes to "input" key
		state.Set("input", p)
	case *domain.State:
		// Already a State, return a clone
		return p.Clone(), nil
	default:
		// Store as "params" key
		state.Set("params", params)
	}

	return state, nil
}

// DefaultResultMapper extracts result from State.
// It checks for common result keys in order: "result", "output", "response".
// If none are found, it returns all state values as a map.
//
// Parameters:
//   - ctx: Context for the operation
//   - state: The agent's result state
//
// Returns the extracted result or an error.
func DefaultResultMapper(ctx context.Context, state *domain.State) (interface{}, error) {
	// Check for common result keys
	if result, exists := state.Get("result"); exists {
		return result, nil
	}
	if output, exists := state.Get("output"); exists {
		return output, nil
	}
	if response, exists := state.Get("response"); exists {
		return response, nil
	}

	// Return the entire state values as a map
	return state.Values(), nil
}

// CreateStateMapper creates a state mapper with field mappings.
// The returned mapper transforms parameter fields according to the provided mappings,
// and includes any unmapped fields as-is.
//
// Parameters:
//   - fieldMappings: Map of parameter field names to state field names
//
// Returns a StateMapper function.
func CreateStateMapper(fieldMappings map[string]string) StateMapper {
	return func(ctx context.Context, params interface{}) (*domain.State, error) {
		state := domain.NewState()

		if paramsMap, ok := params.(map[string]interface{}); ok {
			for paramKey, stateKey := range fieldMappings {
				if value, exists := paramsMap[paramKey]; exists {
					state.Set(stateKey, value)
				}
			}
			// Also include unmapped fields
			for k, v := range paramsMap {
				if _, mapped := fieldMappings[k]; !mapped {
					state.Set(k, v)
				}
			}
		} else {
			return DefaultStateMapper(ctx, params)
		}

		return state, nil
	}
}

// CreateResultMapper creates a result mapper that extracts specific fields.
// If a single field is specified, its value is returned directly.
// If multiple fields are specified, a map of field values is returned.
//
// Parameters:
//   - fields: The state fields to extract
//
// Returns a ResultMapper function.
func CreateResultMapper(fields ...string) ResultMapper {
	return func(ctx context.Context, state *domain.State) (interface{}, error) {
		if len(fields) == 1 {
			// Single field extraction
			if value, exists := state.Get(fields[0]); exists {
				return value, nil
			}
			return nil, fmt.Errorf("field %s not found in state", fields[0])
		}

		// Multiple fields extraction
		result := make(map[string]interface{})
		for _, field := range fields {
			if value, exists := state.Get(field); exists {
				result[field] = value
			}
		}

		if len(result) == 0 {
			return nil, fmt.Errorf("none of the requested fields found in state")
		}

		return result, nil
	}
}

// OutputSchema returns the schema for tool output.
// It attempts to derive this from the agent's output schema.
func (at *AgentTool) OutputSchema() *sdomain.Schema {
	// Try to derive from agent's output schema
	if at.agent.OutputSchema() != nil {
		return at.agent.OutputSchema()
	}
	return nil
}

// UsageInstructions returns usage instructions for the tool.
// It includes information about the wrapped agent.
func (at *AgentTool) UsageInstructions() string {
	return fmt.Sprintf("This tool wraps the '%s' agent. %s", at.agent.Name(), at.agent.Description())
}

// Examples returns usage examples for the tool.
// Currently returns nil but could be enhanced to derive from agent examples.
func (at *AgentTool) Examples() []domain.ToolExample {
	// Could potentially derive from agent examples in the future
	return nil
}

// Constraints returns operational constraints for the tool.
// These describe limitations and requirements for successful execution.
func (at *AgentTool) Constraints() []string {
	return []string{
		"Agent must complete its task successfully",
		"Result format depends on the agent implementation",
	}
}

// ErrorGuidance returns guidance for common error scenarios.
// It helps users understand and resolve tool execution errors.
func (at *AgentTool) ErrorGuidance() map[string]string {
	return map[string]string{
		"agent_error":         "The underlying agent failed to execute. Check the agent logs for details.",
		"state_mapping_error": "Failed to convert parameters to agent state format.",
	}
}

// Category returns the tool category, which is always "agent" for wrapped agents.
func (at *AgentTool) Category() string {
	return "agent"
}

// Tags returns descriptive tags for the tool.
// These help with tool discovery and classification.
func (at *AgentTool) Tags() []string {
	return []string{"agent", "wrapper"}
}

// Version returns the tool version.
// Currently returns a fixed version but could be enhanced to include agent version.
func (at *AgentTool) Version() string {
	return "1.0.0"
}

// IsDeterministic returns whether the tool produces deterministic results.
// Returns false because agents typically use LLMs which are non-deterministic.
func (at *AgentTool) IsDeterministic() bool {
	// Agents are generally non-deterministic due to LLM usage
	return false
}

// IsDestructive returns whether the tool performs destructive operations.
// This depends on the wrapped agent's behavior.
func (at *AgentTool) IsDestructive() bool {
	// Depends on the wrapped agent's behavior
	return false
}

// RequiresConfirmation returns whether the tool requires user confirmation before execution.
// Currently returns false but could be enhanced based on agent properties.
func (at *AgentTool) RequiresConfirmation() bool {
	return false
}

// EstimatedLatency returns the estimated execution latency.
// Returns "slow" because agents typically involve LLM calls.
func (at *AgentTool) EstimatedLatency() string {
	return "slow" // Agents typically involve LLM calls
}

// ToMCPDefinition exports the tool definition in MCP (Model Context Protocol) format.
// This enables the tool to be used in MCP-compatible systems.
func (at *AgentTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:         at.Name(),
		Description:  at.Description(),
		InputSchema:  at.ParameterSchema(),
		OutputSchema: at.OutputSchema(),
		Annotations: map[string]interface{}{
			"type":          "agent_wrapper",
			"category":      "agent",
			"version":       "1.0.0",
			"wrapped_agent": at.agent.Name(),
		},
	}
}
