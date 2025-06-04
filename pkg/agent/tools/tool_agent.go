// ABOUTME: ToolAgent wraps a Tool to expose it as a BaseAgent, enabling tools to be used as agents
// ABOUTME: This complements AgentTool to provide bidirectional conversion between agents and tools

package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/core"
)

// ToolAgent wraps a Tool to expose it as a BaseAgent
type ToolAgent struct {
	*core.BaseAgentImpl
	tool          domain.Tool
	paramMapper   ParamMapper
	stateUpdater  StateUpdater
	executeFunc   func(context.Context, *domain.State) (*domain.State, error)
}

// ParamMapper extracts tool parameters from agent State
type ParamMapper func(ctx context.Context, state *domain.State) (interface{}, error)

// StateUpdater updates the State with tool execution results
type StateUpdater func(ctx context.Context, state *domain.State, result interface{}, err error) (*domain.State, error)

// NewToolAgent creates a new ToolAgent wrapper
func NewToolAgent(tool domain.Tool) *ToolAgent {
	ta := &ToolAgent{
		BaseAgentImpl: core.NewBaseAgent(
			tool.Name(),
			tool.Description(),
			domain.AgentTypeCustom, // Using custom type for tool agents
		),
		tool:         tool,
		paramMapper:  DefaultParamMapper,
		stateUpdater: DefaultStateUpdater,
	}

	// Set the execution function
	ta.executeFunc = ta.execute

	return ta
}

// Run executes the tool agent
func (ta *ToolAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Call lifecycle hooks from BaseAgentImpl
	if err := ta.BaseAgentImpl.BeforeRun(ctx, input); err != nil {
		return nil, err
	}

	// Execute the tool
	result, err := ta.executeFunc(ctx, input)

	// Call after hook
	if afterErr := ta.BaseAgentImpl.AfterRun(ctx, input, result, err); afterErr != nil {
		if err != nil {
			return nil, fmt.Errorf("execution failed: %w, after hook also failed: %v", err, afterErr)
		}
		return nil, afterErr
	}

	return result, err
}

// WithParamMapper sets a custom parameter mapper
func (ta *ToolAgent) WithParamMapper(mapper ParamMapper) *ToolAgent {
	ta.paramMapper = mapper
	return ta
}

// WithStateUpdater sets a custom state updater
func (ta *ToolAgent) WithStateUpdater(updater StateUpdater) *ToolAgent {
	ta.stateUpdater = updater
	return ta
}

// execute is the internal execution function
func (ta *ToolAgent) execute(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Extract parameters from State
	params, err := ta.paramMapper(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("failed to extract parameters from state: %w", err)
	}

	// Create ToolContext for the tool
	toolContext := domain.NewToolContext(
		ctx,
		domain.NewStateReader(state),
		ta, // ToolAgent implements BaseAgent
		fmt.Sprintf("toolagent-%s-%d", ta.ID(), time.Now().UnixNano()),
	)

	// If we have an event dispatcher, set up event emitter
	// Note: ToolAgent would need to have access to a dispatcher for full functionality
	// For now, tools wrapped as agents won't emit events unless we enhance this

	// Execute the tool with ToolContext
	result, err := ta.tool.Execute(toolContext, params)

	// Update state with results
	newState, updateErr := ta.stateUpdater(ctx, state.Clone(), result, err)
	if updateErr != nil {
		if err != nil {
			return nil, fmt.Errorf("tool execution failed: %w, state update also failed: %v", err, updateErr)
		}
		return nil, fmt.Errorf("failed to update state: %w", updateErr)
	}

	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return newState, nil
}

// DefaultParamMapper extracts parameters from State
func DefaultParamMapper(ctx context.Context, state *domain.State) (interface{}, error) {
	// Check for explicit params key
	if params, exists := state.Get("params"); exists {
		return params, nil
	}

	// Check for input key
	if input, exists := state.Get("input"); exists {
		return input, nil
	}

	// Return entire state values as map
	return state.Values(), nil
}

// DefaultStateUpdater updates state with tool results
func DefaultStateUpdater(ctx context.Context, state *domain.State, result interface{}, err error) (*domain.State, error) {
	if err != nil {
		state.Set("error", err.Error())
		state.Set("success", false)
		return state, nil
	}

	state.Set("result", result)
	state.Set("success", true)

	// If result is a map, also merge it into state
	if resultMap, ok := result.(map[string]interface{}); ok {
		for k, v := range resultMap {
			// Prefix with "output_" to avoid conflicts
			state.Set(fmt.Sprintf("output_%s", k), v)
		}
	}

	return state, nil
}

// CreateParamMapper creates a parameter mapper with field extraction
func CreateParamMapper(fieldMappings map[string]string) ParamMapper {
	return func(ctx context.Context, state *domain.State) (interface{}, error) {
		params := make(map[string]interface{})

		for stateKey, paramKey := range fieldMappings {
			if value, exists := state.Get(stateKey); exists {
				params[paramKey] = value
			}
		}

		if len(params) == 0 {
			return nil, fmt.Errorf("no parameters found in state")
		}

		return params, nil
	}
}

// CreateSingleParamMapper creates a mapper for tools expecting a single parameter
func CreateSingleParamMapper(stateKey string) ParamMapper {
	return func(ctx context.Context, state *domain.State) (interface{}, error) {
		if value, exists := state.Get(stateKey); exists {
			return value, nil
		}
		return nil, fmt.Errorf("required state key %s not found", stateKey)
	}
}

// CreateStateUpdaterWithPrefix creates an updater that prefixes result keys
func CreateStateUpdaterWithPrefix(prefix string) StateUpdater {
	return func(ctx context.Context, state *domain.State, result interface{}, err error) (*domain.State, error) {
		if err != nil {
			state.Set(prefix+"_error", err.Error())
			state.Set(prefix+"_success", false)
			return state, nil
		}

		state.Set(prefix+"_result", result)
		state.Set(prefix+"_success", true)

		// If result is a map, prefix all keys
		if resultMap, ok := result.(map[string]interface{}); ok {
			for k, v := range resultMap {
				state.Set(fmt.Sprintf("%s_%s", prefix, k), v)
			}
		}

		return state, nil
	}
}