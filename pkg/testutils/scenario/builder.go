// ABOUTME: ScenarioBuilder provides a fluent API for building test scenarios
// ABOUTME: Simplifies complex test setup and assertions for bridge testing

package scenario

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// ScenarioBuilder provides a fluent API for building test scenarios
type ScenarioBuilder struct {
	t            testing.TB
	providers    map[string]*mocks.MockProvider
	tools        map[string]*mocks.MockTool
	agents       map[string]*mocks.MockAgent
	mainAgent    *mocks.MockAgent
	state        *domain.State
	timeout      time.Duration
	context      context.Context
	expectations []expectation
	errors       []error
	eventEmitter *mocks.MockEventEmitter
}

// expectation represents an expected outcome
type expectation struct {
	name        string
	checkFunc   func() (bool, string)
	description string
}

// NewScenario creates a new scenario builder
func NewScenario(t testing.TB) *ScenarioBuilder {
	return &ScenarioBuilder{
		t:            t,
		providers:    make(map[string]*mocks.MockProvider),
		tools:        make(map[string]*mocks.MockTool),
		agents:       make(map[string]*mocks.MockAgent),
		state:        domain.NewState(),
		timeout:      30 * time.Second,
		context:      context.Background(),
		expectations: []expectation{},
		errors:       []error{},
		eventEmitter: mocks.NewMockEventEmitter("scenario-emitter", "scenario"),
	}
}

// Configuration Methods

// WithMockProvider adds a mock provider with predefined responses
func (s *ScenarioBuilder) WithMockProvider(name string, responses map[string]llmdomain.Response) *ScenarioBuilder {
	provider := mocks.NewMockProvider(name)

	// Add pattern-based responses
	for pattern, response := range responses {
		provider.WithPatternResponse(pattern, mocks.Response{
			Content: response.Content,
		})
	}

	s.providers[name] = provider
	return s
}

// WithTool adds a tool to the scenario
func (s *ScenarioBuilder) WithTool(tool *mocks.MockTool) *ScenarioBuilder {
	s.tools[tool.ToolName] = tool
	return s
}

// WithAgent adds an agent to the scenario
func (s *ScenarioBuilder) WithAgent(agent *mocks.MockAgent) *ScenarioBuilder {
	s.agents[agent.Name()] = agent

	// Set as main agent if it's the first one
	if s.mainAgent == nil {
		s.mainAgent = agent
	}

	return s
}

// WithMainAgent sets the main agent for the scenario
func (s *ScenarioBuilder) WithMainAgent(agent *mocks.MockAgent) *ScenarioBuilder {
	s.mainAgent = agent
	s.agents[agent.Name()] = agent
	return s
}

// WithInput adds input to the state
func (s *ScenarioBuilder) WithInput(key string, value interface{}) *ScenarioBuilder {
	s.state.Set(key, value)
	return s
}

// WithState sets the entire state
func (s *ScenarioBuilder) WithState(state *domain.State) *ScenarioBuilder {
	s.state = state
	return s
}

// WithTimeout sets the execution timeout
func (s *ScenarioBuilder) WithTimeout(duration time.Duration) *ScenarioBuilder {
	s.timeout = duration
	return s
}

// WithContext sets a custom context
func (s *ScenarioBuilder) WithContext(ctx context.Context) *ScenarioBuilder {
	s.context = ctx
	return s
}

// WithEventEmitter sets a custom event emitter
func (s *ScenarioBuilder) WithEventEmitter(emitter *mocks.MockEventEmitter) *ScenarioBuilder {
	s.eventEmitter = emitter
	return s
}

// Expectation Methods

// ExpectOutput expects a specific output in the state
func (s *ScenarioBuilder) ExpectOutput(key string, matcher Matcher) *ScenarioBuilder {
	s.expectations = append(s.expectations, expectation{
		name:        fmt.Sprintf("output[%s]", key),
		description: fmt.Sprintf("output %s %s", key, matcher.Description()),
		checkFunc: func() (bool, string) {
			value, exists := s.state.Get(key)
			if !exists {
				return false, fmt.Sprintf("output key %q not found in state", key)
			}
			return matcher.Match(value)
		},
	})
	return s
}

// ExpectToolCall expects a specific tool to be called
func (s *ScenarioBuilder) ExpectToolCall(toolName string, inputMatcher Matcher) *ScenarioBuilder {
	s.expectations = append(s.expectations, expectation{
		name:        fmt.Sprintf("tool_call[%s]", toolName),
		description: fmt.Sprintf("tool %s called with %s", toolName, inputMatcher.Description()),
		checkFunc: func() (bool, string) {
			tool, exists := s.tools[toolName]
			if !exists {
				return false, fmt.Sprintf("tool %q not registered", toolName)
			}

			calls := tool.GetCallHistory()
			if len(calls) == 0 {
				return false, fmt.Sprintf("tool %q was not called", toolName)
			}

			// Check if any call matches
			for _, call := range calls {
				if ok, _ := inputMatcher.Match(call.Input); ok {
					return true, ""
				}
			}

			return false, fmt.Sprintf("no call to tool %q matched the input criteria", toolName)
		},
	})
	return s
}

// ExpectEvent expects a specific event to be emitted
func (s *ScenarioBuilder) ExpectEvent(eventType string, dataMatcher Matcher) *ScenarioBuilder {
	s.expectations = append(s.expectations, expectation{
		name:        fmt.Sprintf("event[%s]", eventType),
		description: fmt.Sprintf("event %s with data %s", eventType, dataMatcher.Description()),
		checkFunc: func() (bool, string) {
			events := s.eventEmitter.GetEvents()

			for _, event := range events {
				if string(event.Type) == eventType {
					if ok, _ := dataMatcher.Match(event.Data); ok {
						return true, ""
					}
				}
			}

			return false, fmt.Sprintf("no event of type %q matched the data criteria", eventType)
		},
	})
	return s
}

// ExpectError expects an error matching the criteria
func (s *ScenarioBuilder) ExpectError(errorMatcher Matcher) *ScenarioBuilder {
	s.expectations = append(s.expectations, expectation{
		name:        "error",
		description: fmt.Sprintf("error %s", errorMatcher.Description()),
		checkFunc: func() (bool, string) {
			if len(s.errors) == 0 {
				return false, "expected error but none occurred"
			}

			// Check if any error matches
			for _, err := range s.errors {
				if ok, _ := errorMatcher.Match(err); ok {
					return true, ""
				}
			}

			// Try matching error messages
			for _, err := range s.errors {
				if ok, _ := errorMatcher.Match(err.Error()); ok {
					return true, ""
				}
			}

			return false, fmt.Sprintf("no error matched the criteria, got: %v", s.errors)
		},
	})
	return s
}

// ExpectNoError expects no errors
func (s *ScenarioBuilder) ExpectNoError() *ScenarioBuilder {
	s.expectations = append(s.expectations, expectation{
		name:        "no_error",
		description: "no error",
		checkFunc: func() (bool, string) {
			if len(s.errors) > 0 {
				return false, fmt.Sprintf("expected no error but got: %v", s.errors)
			}
			return true, ""
		},
	})
	return s
}

// ExpectAgentCall expects the main agent to be called
func (s *ScenarioBuilder) ExpectAgentCall(stateMatcher Matcher) *ScenarioBuilder {
	s.expectations = append(s.expectations, expectation{
		name:        "agent_call",
		description: fmt.Sprintf("agent called with state %s", stateMatcher.Description()),
		checkFunc: func() (bool, string) {
			if s.mainAgent == nil {
				return false, "no main agent set"
			}

			history := s.mainAgent.GetCallHistory()
			if len(history) == 0 {
				return false, "main agent was not called"
			}

			// Check if any call matches
			for _, call := range history {
				if ok, _ := stateMatcher.Match(call.Input); ok {
					return true, ""
				}
			}

			return false, "no agent call matched the state criteria"
		},
	})
	return s
}

// ExpectProviderCall expects a provider to be called
func (s *ScenarioBuilder) ExpectProviderCall(providerName string, messageMatcher Matcher) *ScenarioBuilder {
	s.expectations = append(s.expectations, expectation{
		name:        fmt.Sprintf("provider_call[%s]", providerName),
		description: fmt.Sprintf("provider %s called with %s", providerName, messageMatcher.Description()),
		checkFunc: func() (bool, string) {
			provider, exists := s.providers[providerName]
			if !exists {
				return false, fmt.Sprintf("provider %q not registered", providerName)
			}

			history := provider.GetCallHistory()
			if len(history) == 0 {
				return false, fmt.Sprintf("provider %q was not called", providerName)
			}

			// Check if any call matches
			for _, call := range history {
				if ok, _ := messageMatcher.Match(call.Messages); ok {
					return true, ""
				}
			}

			return false, fmt.Sprintf("no call to provider %q matched the message criteria", providerName)
		},
	})
	return s
}

// Execution Methods

// Run executes the scenario and returns the final state
func (s *ScenarioBuilder) Run() *domain.State {
	ctx, cancel := context.WithTimeout(s.context, s.timeout)
	defer cancel()

	return s.RunWithContext(ctx)
}

// RunWithContext executes the scenario with the given context
func (s *ScenarioBuilder) RunWithContext(ctx context.Context) *domain.State {
	// Clear any previous errors
	s.errors = []error{}

	// Execute the main agent if set
	if s.mainAgent != nil {
		result, err := s.mainAgent.Run(ctx, s.state)
		if err != nil {
			s.errors = append(s.errors, err)
		} else if result != nil {
			s.state = result
		}
	}

	// Run all expectations
	s.verify()

	return s.state
}

// RunTool executes a specific tool in the scenario
func (s *ScenarioBuilder) RunTool(toolName string, input interface{}) (interface{}, error) {
	tool, exists := s.tools[toolName]
	if !exists {
		err := fmt.Errorf("tool %q not found", toolName)
		s.errors = append(s.errors, err)
		return nil, err
	}

	// Create tool context
	toolCtx := &domain.ToolContext{
		Context: s.context,
		State:   s.state,
	}

	result, err := tool.Execute(toolCtx, input)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	return result, err
}

// verify runs all expectations and reports failures
func (s *ScenarioBuilder) verify() {
	var failures []string

	for _, exp := range s.expectations {
		ok, msg := exp.checkFunc()
		if !ok {
			failures = append(failures, fmt.Sprintf("%s: %s", exp.name, msg))
		}
	}

	if len(failures) > 0 {
		s.t.Errorf("Scenario expectations failed:\n%s", strings.Join(failures, "\n"))
	}
}

// Helper Methods

// GetState returns the current state
func (s *ScenarioBuilder) GetState() *domain.State {
	return s.state
}

// GetErrors returns all errors that occurred
func (s *ScenarioBuilder) GetErrors() []error {
	return s.errors
}

// GetProvider returns a provider by name
func (s *ScenarioBuilder) GetProvider(name string) *mocks.MockProvider {
	return s.providers[name]
}

// GetTool returns a tool by name
func (s *ScenarioBuilder) GetTool(name string) *mocks.MockTool {
	return s.tools[name]
}

// GetAgent returns an agent by name
func (s *ScenarioBuilder) GetAgent(name string) *mocks.MockAgent {
	return s.agents[name]
}

// GetEventEmitter returns the event emitter
func (s *ScenarioBuilder) GetEventEmitter() *mocks.MockEventEmitter {
	return s.eventEmitter
}

// Reset clears the scenario for reuse
func (s *ScenarioBuilder) Reset() *ScenarioBuilder {
	// Reset all mocks
	for _, provider := range s.providers {
		provider.Reset()
	}

	for _, tool := range s.tools {
		tool.Reset()
	}

	for _, agent := range s.agents {
		agent.Reset()
	}

	// Clear state
	s.state = domain.NewState()
	s.errors = []error{}
	s.expectations = []expectation{}
	s.eventEmitter.Reset()

	return s
}
