// ABOUTME: ToolAgent wraps a Tool to expose it as a BaseAgent, enabling tools to be used as agents
// ABOUTME: This complements AgentTool to provide bidirectional conversion between agents and tools

package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ToolAgent wraps a Tool to expose it as a BaseAgent.
// This enables tools to be used in agent-based contexts, providing
// the inverse functionality of AgentTool.
type ToolAgent struct {
	*core.BaseAgentImpl
	tool            domain.Tool
	paramMapper     ParamMapper
	stateUpdater    StateUpdater
	executeFunc     func(context.Context, *domain.State) (*domain.State, error)
	eventDispatcher domain.EventDispatcher // Added for event support
}

// ParamMapper extracts tool parameters from agent State.
// It bridges the gap between state-based agent inputs and parameter-based tool inputs.
type ParamMapper func(ctx context.Context, state *domain.State) (interface{}, error)

// StateUpdater updates the State with tool execution results.
// It handles both successful results and errors, updating the state accordingly.
type StateUpdater func(ctx context.Context, state *domain.State, result interface{}, err error) (*domain.State, error)

// NewToolAgent creates a new ToolAgent wrapper with default mappers.
// The default param mapper extracts from "params" or "input" keys,
// and the default state updater sets "result" and "success" keys.
//
// Parameters:
//   - tool: The tool to wrap as an agent
//
// Returns a new ToolAgent instance.
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

// Run executes the tool agent.
// It extracts parameters from state, executes the tool, and updates
// the state with results. Lifecycle hooks are called before and after execution.
//
// Parameters:
//   - ctx: The execution context
//   - input: The input state
//
// Returns the updated state or an error.
func (ta *ToolAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Call lifecycle hooks from BaseAgentImpl
	if err := ta.BeforeRun(ctx, input); err != nil {
		return nil, err
	}

	// Execute the tool
	result, err := ta.executeFunc(ctx, input)

	// Call after hook
	if afterErr := ta.AfterRun(ctx, input, result, err); afterErr != nil {
		if err != nil {
			return nil, fmt.Errorf("execution failed: %w, after hook also failed: %v", err, afterErr)
		}
		return nil, afterErr
	}

	return result, err
}

// WithParamMapper sets a custom parameter mapper.
// This allows customization of how state is converted to tool parameters.
//
// Parameters:
//   - mapper: The custom parameter mapper function
//
// Returns the ToolAgent for method chaining.
func (ta *ToolAgent) WithParamMapper(mapper ParamMapper) *ToolAgent {
	ta.paramMapper = mapper
	return ta
}

// WithStateUpdater sets a custom state updater.
// This allows customization of how tool results are stored in state.
//
// Parameters:
//   - updater: The custom state updater function
//
// Returns the ToolAgent for method chaining.
func (ta *ToolAgent) WithStateUpdater(updater StateUpdater) *ToolAgent {
	ta.stateUpdater = updater
	return ta
}

// WithEventDispatcher sets the event dispatcher.
// This enables the tool to emit events during execution.
//
// Parameters:
//   - dispatcher: The event dispatcher to use
//
// Returns the ToolAgent for method chaining.
func (ta *ToolAgent) WithEventDispatcher(dispatcher domain.EventDispatcher) *ToolAgent {
	ta.eventDispatcher = dispatcher
	return ta
}

// execute is the internal execution function.
// It handles parameter extraction, tool execution, and state updates.
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

	// If we have an event dispatcher, enhance the tool context with event emission
	if ta.eventDispatcher != nil {
		emitter := &toolEventEmitter{
			dispatcher: ta.eventDispatcher,
			agentID:    ta.ID(),
			agentName:  ta.Name(),
			runID:      toolContext.RunID,
		}
		toolContext.Events = emitter
	}

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

// DefaultParamMapper extracts parameters from State.
// It checks for "params" and "input" keys in order, falling back
// to the entire state values if neither exists.
//
// Parameters:
//   - ctx: The execution context
//   - state: The state to extract from
//
// Returns the extracted parameters or an error.
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

// DefaultStateUpdater updates state with tool results.
// It sets "result" and "success" keys, and for map results,
// also adds prefixed keys for each map entry.
//
// Parameters:
//   - ctx: The execution context
//   - state: The state to update
//   - result: The tool execution result
//   - err: Any error from tool execution
//
// Returns the updated state.
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

// CreateParamMapper creates a parameter mapper with field extraction.
// The mapper extracts specific fields from state based on the provided mappings.
//
// Parameters:
//   - fieldMappings: Map of state keys to parameter keys
//
// Returns a ParamMapper function.
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

// CreateSingleParamMapper creates a mapper for tools expecting a single parameter.
// It extracts a single value from the specified state key.
//
// Parameters:
//   - stateKey: The state key to extract
//
// Returns a ParamMapper function.
func CreateSingleParamMapper(stateKey string) ParamMapper {
	return func(ctx context.Context, state *domain.State) (interface{}, error) {
		if value, exists := state.Get(stateKey); exists {
			return value, nil
		}
		return nil, fmt.Errorf("required state key %s not found", stateKey)
	}
}

// CreateStateUpdaterWithPrefix creates an updater that prefixes result keys.
// This helps avoid key conflicts when multiple tools update the same state.
//
// Parameters:
//   - prefix: The prefix to add to all result keys
//
// Returns a StateUpdater function.
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

// toolEventEmitter provides event emission for tools wrapped as agents.
// It implements domain.ToolEventEmitter to bridge tool events with the agent event system.
type toolEventEmitter struct {
	dispatcher domain.EventDispatcher
	agentID    string
	agentName  string
	runID      string
}

func (e *toolEventEmitter) Emit(eventType domain.EventType, data interface{}) {
	if e.dispatcher != nil {
		e.dispatcher.Dispatch(domain.Event{
			Type:      eventType,
			AgentID:   e.agentID,
			AgentName: e.agentName,
			Data:      data,
		})
	}
}

func (e *toolEventEmitter) EmitProgress(current, total int, message string) {
	e.Emit(domain.EventProgress, domain.ProgressEventData{
		Current: current,
		Total:   total,
		Message: message,
	})
}

func (e *toolEventEmitter) EmitMessage(message string) {
	e.Emit(domain.EventMessage, message)
}

func (e *toolEventEmitter) EmitError(err error) {
	if e.dispatcher != nil {
		e.dispatcher.Dispatch(domain.Event{
			Type:      domain.EventAgentError,
			AgentID:   e.agentID,
			AgentName: e.agentName,
			Error:     err,
		})
	}
}

func (e *toolEventEmitter) EmitCustom(eventName string, data interface{}) {
	e.Emit(domain.EventMessage, map[string]interface{}{
		"type": eventName,
		"data": data,
	})
}
