// ABOUTME: Tests for MockAgent implementation verifying all agent functionality
// ABOUTME: Covers response queues, sub-agent management, event tracking, and state history

package mocks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockAgent_Basic(t *testing.T) {
	agent := NewMockAgent("test-agent")

	assert.Equal(t, "test-agent", agent.Name())
	assert.Equal(t, "Mock agent test-agent", agent.Description())
	assert.Equal(t, domain.AgentTypeCustom, agent.Type())
	assert.NotEmpty(t, agent.ID())
}

func TestMockAgent_ResponseQueue(t *testing.T) {
	agent := NewMockAgent("test-agent")
	ctx := context.Background()

	// Add responses to queue
	response1 := domain.NewState()
	response1.Set("result", "first")
	response2 := domain.NewState()
	response2.Set("result", "second")

	agent.AddResponse(response1).AddResponse(response2)

	// Run should return queued responses in order
	input := domain.NewState()

	result1, err := agent.Run(ctx, input)
	require.NoError(t, err)
	val1, _ := result1.Get("result")
	assert.Equal(t, "first", val1)

	result2, err := agent.Run(ctx, input)
	require.NoError(t, err)
	val2, _ := result2.Get("result")
	assert.Equal(t, "second", val2)

	// After queue is exhausted, should return default response
	result3, err := agent.Run(ctx, input)
	require.NoError(t, err)
	val3, _ := result3.Get("result")
	assert.Equal(t, "Mock response from test-agent", val3)
}

func TestMockAgent_ErrorQueue(t *testing.T) {
	agent := NewMockAgent("test-agent")
	ctx := context.Background()

	// Add errors to queue
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	agent.AddError(err1).AddError(err2)

	// Add a response after errors
	response := domain.NewState()
	response.Set("result", "success")
	agent.AddResponse(response)

	input := domain.NewState()

	// Should return errors first
	_, gotErr1 := agent.Run(ctx, input)
	assert.Equal(t, err1, gotErr1)

	_, gotErr2 := agent.Run(ctx, input)
	assert.Equal(t, err2, gotErr2)

	// Then return response
	result, err := agent.Run(ctx, input)
	require.NoError(t, err)
	val, _ := result.Get("result")
	assert.Equal(t, "success", val)
}

func TestMockAgent_BehaviorHooks(t *testing.T) {
	agent := NewMockAgent("test-agent")
	ctx := context.Background()

	// Test OnRun hook
	hookCalled := false
	agent.OnRun = func(ctx context.Context, input *domain.State) (*domain.State, error) {
		hookCalled = true
		output := domain.NewState()
		output.Set("hooked", true)
		return output, nil
	}

	input := domain.NewState()
	result, err := agent.Run(ctx, input)
	require.NoError(t, err)
	assert.True(t, hookCalled)
	val, _ := result.Get("hooked")
	assert.Equal(t, true, val)

	// Test lifecycle hooks
	initCalled := false
	agent.OnInitialize = func(ctx context.Context) error {
		initCalled = true
		return nil
	}

	err = agent.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, initCalled)

	beforeCalled := false
	agent.OnBeforeRun = func(ctx context.Context, state *domain.State) error {
		beforeCalled = true
		return nil
	}

	err = agent.BeforeRun(ctx, input)
	require.NoError(t, err)
	assert.True(t, beforeCalled)

	afterCalled := false
	agent.OnAfterRun = func(ctx context.Context, state *domain.State, result *domain.State, err error) error {
		afterCalled = true
		return nil
	}

	err = agent.AfterRun(ctx, input, result, nil)
	require.NoError(t, err)
	assert.True(t, afterCalled)

	cleanupCalled := false
	agent.OnCleanup = func(ctx context.Context) error {
		cleanupCalled = true
		return nil
	}

	err = agent.Cleanup(ctx)
	require.NoError(t, err)
	assert.True(t, cleanupCalled)
}

func TestMockAgent_SubAgentManagement(t *testing.T) {
	parent := NewMockAgent("parent")
	child1 := NewMockAgent("child1")
	child2 := NewMockAgent("child2")

	// Add sub-agents
	err := parent.AddSubAgent(child1)
	require.NoError(t, err)
	assert.Equal(t, parent, child1.Parent())

	err = parent.AddSubAgent(child2)
	require.NoError(t, err)
	assert.Equal(t, parent, child2.Parent())

	// Check sub-agents list
	subAgents := parent.SubAgents()
	assert.Len(t, subAgents, 2)

	// Test duplicate prevention
	err = parent.AddSubAgent(child1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Find sub-agent
	found := parent.FindSubAgent("child1")
	assert.Equal(t, child1, found)

	notFound := parent.FindSubAgent("nonexistent")
	assert.Nil(t, notFound)

	// Find agent in hierarchy
	found = parent.FindAgent("child2")
	assert.Equal(t, child2, found)

	// Remove sub-agent
	err = parent.RemoveSubAgent("child1")
	require.NoError(t, err)
	assert.Nil(t, child1.Parent())

	subAgents = parent.SubAgents()
	assert.Len(t, subAgents, 1)

	// Try to remove non-existent
	err = parent.RemoveSubAgent("nonexistent")
	assert.Error(t, err)
}

func TestMockAgent_CallHistory(t *testing.T) {
	agent := NewMockAgent("test-agent")
	ctx := context.Background()

	// Make several calls
	input1 := domain.NewState()
	input1.Set("id", 1)

	input2 := domain.NewState()
	input2.Set("id", 2)

	response := domain.NewState()
	response.Set("result", "success")
	agent.AddResponse(response)

	_, err := agent.Run(ctx, input1)
	assert.NoError(t, err)
	_, err = agent.Run(ctx, input2)
	assert.NoError(t, err)

	// Check call history
	history := agent.GetCallHistory()
	assert.Len(t, history, 2)

	// Verify first call
	assert.NotNil(t, history[0].Input)
	val, _ := history[0].Input.Get("id")
	assert.Equal(t, 1, val)
	assert.NotNil(t, history[0].Output)
	assert.NoError(t, history[0].Error)
	assert.NotZero(t, history[0].Timestamp)
	assert.NotZero(t, history[0].Duration)

	// Verify second call
	val, _ = history[1].Input.Get("id")
	assert.Equal(t, 2, val)
}

func TestMockAgent_RunAsync(t *testing.T) {
	agent := NewMockAgent("test-agent")
	ctx := context.Background()

	response := domain.NewState()
	response.Set("result", "async success")
	agent.AddResponse(response)

	input := domain.NewState()
	eventChan, err := agent.RunAsync(ctx, input)
	require.NoError(t, err)

	// Collect events
	events := make([]domain.Event, 0)
	for event := range eventChan {
		events = append(events, event)
	}

	// Should have start and complete events
	assert.Len(t, events, 2)
	assert.Equal(t, "agent.start", string(events[0].Type))
	assert.Equal(t, "agent.complete", string(events[1].Type))

	// Check emitted events were recorded
	emitted := agent.GetEmittedEvents()
	assert.Len(t, emitted, 2)
}

func TestMockAgent_Metadata(t *testing.T) {
	agent := NewMockAgent("test-agent")

	// Set metadata
	agent.SetMetadata("key1", "value1")
	agent.SetMetadata("key2", 42)

	// Get metadata
	meta := agent.Metadata()
	assert.Equal(t, "value1", meta["key1"])
	assert.Equal(t, 42, meta["key2"])

	// Verify returned map is a copy
	meta["key3"] = "external"
	actualMeta := agent.Metadata()
	_, exists := actualMeta["key3"]
	assert.False(t, exists)
}

func TestMockAgent_Reset(t *testing.T) {
	agent := NewMockAgent("test-agent")
	ctx := context.Background()

	// Add data
	agent.AddResponse(domain.NewState())
	agent.AddError(errors.New("test error"))

	// Make a call to generate history
	_, _ = agent.Run(ctx, domain.NewState()) // Ignore error as we expect it to fail

	// Emit some events
	event := domain.NewEvent(domain.EventAgentStart, agent.ID(), agent.Name(), nil)
	agent.recordEvent(event)

	// Reset
	agent.Reset()

	// Verify everything is cleared
	assert.Empty(t, agent.ResponseQueue)
	assert.Empty(t, agent.ErrorQueue)
	assert.Empty(t, agent.GetCallHistory())
	assert.Empty(t, agent.GetEmittedEvents())
	assert.Equal(t, 0, agent.queueIndex)
	assert.Equal(t, 0, agent.errorQueueIndex)
}

func TestMockAgent_Configuration(t *testing.T) {
	agent := NewMockAgent("test-agent")

	// Set config
	config := domain.AgentConfig{
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}

	returned := agent.WithConfig(config)
	assert.Equal(t, agent, returned) // Should return self

	// Get config
	gotConfig := agent.Config()
	assert.Equal(t, config.MaxRetries, gotConfig.MaxRetries)
	assert.Equal(t, config.Timeout, gotConfig.Timeout)

	// Validation
	err := agent.Validate()
	require.NoError(t, err)

	// Test validation failure
	agent.AgentName = ""
	err = agent.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}
