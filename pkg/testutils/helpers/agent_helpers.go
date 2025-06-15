// ABOUTME: Helper functions for creating mock agents with common behaviors
// ABOUTME: Provides reusable patterns for test agent creation across the codebase

package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// CreateMockAgentWithDefault creates a mock agent with default behavior that marks execution in state
func CreateMockAgentWithDefault(name string) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Default behavior: add a marker to the state
		newState := state.Clone()
		newState.Set(fmt.Sprintf("%s_executed", name), true)
		newState.Set("last_agent", name)
		return newState, nil
	}
	return agent
}

// CreateMockAgentWithError creates a mock agent that returns an error
func CreateMockAgentWithError(name string) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	agent.AddError(fmt.Errorf("mock error from %s", name))
	return agent
}

// CreateMockAgentWithDelay creates a mock agent that delays execution
func CreateMockAgentWithDelay(name string, delay time.Duration) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Use a timer to respect context cancellation
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
			// Delay completed
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		}

		// Default behavior: add a marker to the state
		newState := state.Clone()
		newState.Set(fmt.Sprintf("%s_executed", name), true)
		newState.Set("last_agent", name)
		return newState, nil
	}
	return agent
}

// CreateMockAgentWithResponse creates a mock agent that sets a specific response
func CreateMockAgentWithResponse(name string, response interface{}) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		newState := state.Clone()
		newState.Set("response", response)
		newState.Set(fmt.Sprintf("%s_executed", name), true)
		return newState, nil
	}
	return agent
}

// CreateMockAgentWithRunFunc creates a mock agent with a custom run function
func CreateMockAgentWithRunFunc(name string, runFunc func(ctx context.Context, state *domain.State) (*domain.State, error)) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	agent.OnRun = runFunc
	return agent
}
