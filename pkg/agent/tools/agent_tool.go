// ABOUTME: AgentTool wraps a BaseAgent to expose it as a Tool, enabling agents to be used as tools
// ABOUTME: This bridges the State-based agent system with the parameter-based tool system

package tools

import (
	"context"
	"fmt"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// AgentTool wraps a BaseAgent to expose it as a Tool
type AgentTool struct {
	agent         domain.BaseAgent
	paramSchema   *sdomain.Schema
	stateMapper   StateMapper
	resultMapper  ResultMapper
}

// StateMapper converts tool parameters to agent State
type StateMapper func(ctx context.Context, params interface{}) (*domain.State, error)

// ResultMapper converts agent State result to tool result
type ResultMapper func(ctx context.Context, state *domain.State) (interface{}, error)

// NewAgentTool creates a new AgentTool wrapper
func NewAgentTool(agent domain.BaseAgent) *AgentTool {
	return &AgentTool{
		agent:        agent,
		stateMapper:  DefaultStateMapper,
		resultMapper: DefaultResultMapper,
	}
}

// WithParameterSchema sets the parameter schema for the tool
func (at *AgentTool) WithParameterSchema(schema *sdomain.Schema) *AgentTool {
	at.paramSchema = schema
	return at
}

// WithStateMapper sets a custom state mapper
func (at *AgentTool) WithStateMapper(mapper StateMapper) *AgentTool {
	at.stateMapper = mapper
	return at
}

// WithResultMapper sets a custom result mapper
func (at *AgentTool) WithResultMapper(mapper ResultMapper) *AgentTool {
	at.resultMapper = mapper
	return at
}

// Name returns the tool name (agent name)
func (at *AgentTool) Name() string {
	return at.agent.Name()
}

// Description returns the tool description (agent description)
func (at *AgentTool) Description() string {
	return at.agent.Description()
}

// Execute runs the agent with mapped parameters
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

// ParameterSchema returns the tool's parameter schema
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

// DefaultStateMapper converts parameters to State using a simple mapping
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

// DefaultResultMapper extracts result from State
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

// CreateStateMapper creates a state mapper with field mappings
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

// CreateResultMapper creates a result mapper that extracts specific fields
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