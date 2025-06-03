// ABOUTME: Tests for the enhanced RunContext with type-safe dependency injection
// ABOUTME: including metadata, state access, and event emission capabilities

package domain

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Example dependency types for testing
type TestConfig struct {
	APIKey     string
	MaxRetries int
	Timeout    time.Duration
}

type TestDatabase struct {
	ConnectionString string
	Connected        bool
}

func TestRunContext_Basic(t *testing.T) {
	// Create base context
	ctx := context.Background()

	// Create test dependencies
	deps := &TestConfig{
		APIKey:     "test-key",
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}

	// Create test state
	state := NewState()
	state.SetMetadata("test-key", "test-value")

	// Create run context
	runCtx := NewRunContextWithState(ctx, deps, state)

	// Test context retrieval
	if runCtx.Context() != ctx {
		t.Error("Context() should return the base context")
	}

	// Test dependencies retrieval
	if runCtx.Deps() != deps {
		t.Error("Deps() should return the provided dependencies")
	}

	// Test state retrieval
	if runCtx.State != state {
		t.Error("State should return the provided state")
	}
}

func TestRunContext_Metadata(t *testing.T) {
	ctx := context.Background()
	deps := &TestConfig{}
	state := NewState()

	_ = NewRunContextWithState(ctx, deps, state)

	// Set metadata on state
	state.SetMetadata("key1", "value1")
	state.SetMetadata("key2", 42)
	state.SetMetadata("key3", true)

	// Get metadata from state
	tests := []struct {
		key      string
		expected interface{}
		exists   bool
	}{
		{"key1", "value1", true},
		{"key2", 42, true},
		{"key3", true, true},
		{"missing", nil, false},
	}

	for _, tt := range tests {
		value, exists := state.GetMetadata(tt.key)
		if exists != tt.exists {
			t.Errorf("GetMetadata(%s) exists = %v, want %v", tt.key, exists, tt.exists)
		}
		if value != tt.expected {
			t.Errorf("GetMetadata(%s) = %v, want %v", tt.key, value, tt.expected)
		}
	}
}

func TestRunContext_Timeout(t *testing.T) {
	ctx := context.Background()
	deps := &TestConfig{Timeout: 5 * time.Second}
	state := NewState()

	runCtx := NewRunContextWithState(ctx, deps, state)

	// Test basic RunContext fields
	if runCtx.RunID == "" {
		t.Error("RunID should not be empty")
	}

	if runCtx.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}

	// Test WithRetry
	runCtx2 := runCtx.WithRetry(2)
	if runCtx2.Retry != 2 {
		t.Errorf("WithRetry(2) should set Retry to 2, got %d", runCtx2.Retry)
	}

	// Test Elapsed
	time.Sleep(10 * time.Millisecond)
	elapsed := runCtx.Elapsed()
	if elapsed < 10*time.Millisecond {
		t.Errorf("Elapsed should be at least 10ms, got %v", elapsed)
	}
}

func TestRunContext_Events(t *testing.T) {
	ctx := context.Background()
	deps := &TestConfig{}
	state := NewState()

	runCtx := NewRunContextWithState(ctx, deps, state)

	// Test event emission
	var capturedEvents []Event
	runCtx = runCtx.WithEventEmitter(func(e Event) {
		capturedEvents = append(capturedEvents, e)
	})

	// Emit progress event
	runCtx.EmitProgress(50, 100, "Halfway done")

	// Emit message event
	runCtx.EmitMessage("Processing complete")

	// Check captured events
	if len(capturedEvents) != 2 {
		t.Errorf("Expected 2 events, got %d", len(capturedEvents))
	}

	if capturedEvents[0].Type != EventProgress {
		t.Errorf("First event type = %v, want EventProgress", capturedEvents[0].Type)
	}

	if capturedEvents[1].Type != EventMessage {
		t.Errorf("Second event type = %v, want EventMessage", capturedEvents[1].Type)
	}
}

func TestRunContext_WithState(t *testing.T) {
	ctx := context.Background()
	deps := &TestConfig{}
	state1 := NewState()
	state1.Set("key1", "value1")

	runCtx := NewRunContextWithState(ctx, deps, state1)

	// Verify initial state
	if runCtx.State != state1 {
		t.Error("RunContext should have initial state")
	}

	// Create new context with different state
	state2 := NewState()
	state2.Set("key2", "value2")

	runCtx2 := runCtx.WithState(state2)

	// Original context should be unchanged
	if runCtx.State != state1 {
		t.Error("Original context state should be unchanged")
	}

	// New context should have new state
	if runCtx2.State != state2 {
		t.Error("New context should have new state")
	}
}

func TestRunContext_MultipleDependencies(t *testing.T) {
	// Example with struct containing multiple dependencies
	type AppDependencies struct {
		Config   *TestConfig
		Database *TestDatabase
		Logger   interface{} // Could be a logger interface
	}

	ctx := context.Background()
	deps := &AppDependencies{
		Config: &TestConfig{
			APIKey:     "secret",
			MaxRetries: 5,
			Timeout:    1 * time.Minute,
		},
		Database: &TestDatabase{
			ConnectionString: "postgres://localhost/test",
			Connected:        true,
		},
		Logger: nil, // In real app, this would be a logger
	}

	state := NewState()
	runCtx := NewRunContextWithState(ctx, deps, state)

	// Access nested dependencies
	appDeps := runCtx.Deps()
	if appDeps.Config.APIKey != "secret" {
		t.Errorf("Config.APIKey = %q, want %q", appDeps.Config.APIKey, "secret")
	}
	if !appDeps.Database.Connected {
		t.Error("Database should be connected")
	}
}

func TestRunContext_ThreadSafety(t *testing.T) {
	ctx := context.Background()
	deps := &TestConfig{}
	state := NewState()

	_ = NewRunContextWithState(ctx, deps, state)

	// Concurrent metadata access
	done := make(chan bool)

	// Writers
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := fmt.Sprintf("key%d", n)
			state.SetMetadata(key, n)
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := fmt.Sprintf("key%d", n)
			time.Sleep(1 * time.Millisecond) // Give writers a chance
			state.GetMetadata(key)
			done <- true
		}(i)
	}

	// State setters
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := fmt.Sprintf("value%d", n)
			state.Set(key, n)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 30; i++ {
		<-done
	}

	// Verify we have state values
	values := state.Values()
	if len(values) < 5 {
		t.Errorf("Expected at least 5 values, got %d", len(values))
	}
}

func TestRunContext_Example(t *testing.T) {
	// Example showing how RunContext would be used in an agent

	// Define agent dependencies
	type AgentDeps struct {
		LLMClient    interface{} // Would be actual LLM client
		ToolRegistry interface{} // Would be tool registry
		Config       struct {
			MaxRetries     int
			TimeoutSeconds int
			Temperature    float64
		}
	}

	// Create dependencies
	deps := &AgentDeps{
		LLMClient:    nil, // Would be actual client
		ToolRegistry: nil, // Would be actual registry
		Config: struct {
			MaxRetries     int
			TimeoutSeconds int
			Temperature    float64
		}{
			MaxRetries:     3,
			TimeoutSeconds: 30,
			Temperature:    0.7,
		},
	}

	// Create initial state
	state := NewState()
	state.AddMessage(NewMessage(RoleUser, "Hello, can you help me?"))

	// Create run context
	ctx := context.Background()
	runCtx := NewRunContextWithState(ctx, deps, state)

	// Set execution metadata on state
	state.SetMetadata("request_id", "req-123")
	state.SetMetadata("user_id", "user-456")

	// Test event emission with custom emitter
	var capturedEvents []Event
	runCtx = runCtx.WithEventEmitter(func(e Event) {
		capturedEvents = append(capturedEvents, e)
	})

	// Simulate agent processing
	runCtx.EmitMessage("Starting agent processing")

	// Simulate progress
	runCtx.EmitProgress(50, 100, "Processing request")

	// Update state
	runCtx.State.AddMessage(NewMessage(RoleAssistant, "I found some information for you."))

	// Final progress
	runCtx.EmitProgress(100, 100, "Completed")

	// Verify execution
	if len(capturedEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(capturedEvents))
	}

	messages := runCtx.State.Messages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

func TestGetFromState(t *testing.T) {
	ctx := context.Background()
	deps := &TestConfig{}

	// Test with nil state
	rc := NewRunContext(ctx, deps)
	if rc.State != nil {
		t.Error("NewRunContext should have nil State")
	}

	// Test with state
	state := NewState()
	state.Set("key1", "value1")
	state.Set("key2", 42)

	rc = rc.WithState(state)

	// Test existing key
	val1, ok := rc.State.Get("key1")
	if !ok || val1 != "value1" {
		t.Errorf("State.Get(\"key1\") should return value1, got %v", val1)
	}

	// Test another key
	val2, ok := rc.State.Get("key2")
	if !ok || val2 != 42 {
		t.Errorf("State.Get(\"key2\") should return 42, got %v", val2)
	}

	// Test non-existent key
	_, ok = rc.State.Get("missing")
	if ok {
		t.Error("State.Get(\"missing\") should return false")
	}
}
