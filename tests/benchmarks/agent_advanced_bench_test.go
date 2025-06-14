package benchmarks

// ABOUTME: Advanced benchmarks for agent features including state management, workflow agents, and hooks
// ABOUTME: Tests performance of new architecture components introduced in Phase 1-5

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// BenchmarkAgentCreation benchmarks different ways of creating agents
func BenchmarkAgentCreation(b *testing.B) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()

	b.Run("DirectCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			deps := core.LLMDeps{
				Provider: mockProvider,
			}
			_ = core.NewLLMAgent("benchmark-agent", "mock", deps)
		}
	})

	b.Run("StringCreation", func(b *testing.B) {
		// Set up environment variable for mock provider
		b.Setenv("GO_LLMS_MOCK_API_KEY", "test-key")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = core.NewAgentFromString("benchmark-agent", "mock")
		}
	})

	b.Run("WithTools", func(b *testing.B) {
		tool1 := mocks.NewMockTool("tool1", "Test tool 1")
		tool2 := mocks.NewMockTool("tool2", "Test tool 2")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			deps := core.LLMDeps{
				Provider: mockProvider,
			}
			agent := core.NewLLMAgent("benchmark-agent", "mock", deps)
			agent.AddTool(tool1)
			agent.AddTool(tool2)
		}
	})

	b.Run("WithHooks", func(b *testing.B) {
		hook := core.NewLLMMetricsHook()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			deps := core.LLMDeps{
				Provider: mockProvider,
			}
			agent := core.NewLLMAgent("benchmark-agent", "mock", deps)
			agent.WithHook(hook)
		}
	})
}

// BenchmarkStateManagement benchmarks state operations
func BenchmarkStateManagement(b *testing.B) {
	b.Run("StateCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = domain.NewState()
		}
	})

	b.Run("StateSetGet", func(b *testing.B) {
		state := domain.NewState()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			state.Set("key", i)
			_, _ = state.Get("key")
		}
	})

	b.Run("StateNestedData", func(b *testing.B) {
		state := domain.NewState()
		nestedData := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"value": 42,
					"array": []int{1, 2, 3, 4, 5},
				},
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			state.Set("nested", nestedData)
			data, _ := state.Get("nested")
			// Access nested value to ensure full traversal
			if m, ok := data.(map[string]interface{}); ok {
				if l1, ok := m["level1"].(map[string]interface{}); ok {
					if l2, ok := l1["level2"].(map[string]interface{}); ok {
						_ = l2["value"]
					}
				}
			}
		}
	})

	b.Run("StateClone", func(b *testing.B) {
		state := domain.NewState()
		state.Set("key1", "value1")
		state.Set("key2", 42)
		state.Set("key3", []int{1, 2, 3})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = state.Clone()
		}
	})

	b.Run("SharedStateContext", func(b *testing.B) {
		parentState := domain.NewState()
		parentState.Set("shared_key", "shared_value")
		sharedContext := domain.NewSharedStateContext(parentState)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			childState := domain.NewState()
			// Simulate shared state access
			if val, exists := sharedContext.Get("shared_key"); exists {
				childState.Set("local_key", val)
			}
		}
	})
}

// BenchmarkToolExecution benchmarks tool execution performance
func BenchmarkToolExecution(b *testing.B) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()

	// Create tools
	simpleTool := mocks.NewMockTool("simple", "Simple tool")
	complexTool := mocks.NewMockTool("complex", "Complex tool")

	b.Run("SingleToolCall", func(b *testing.B) {
		deps := core.LLMDeps{
			Provider: mockProvider,
		}
		agent := core.NewLLMAgent("benchmark-agent", "mock", deps)
		agent.AddTool(simpleTool)

		ctx := context.Background()
		state := domain.NewState()
		state.Set("user_input", "Use the simple tool")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = agent.Run(ctx, state.Clone())
		}
	})

	b.Run("MultipleToolCalls", func(b *testing.B) {
		deps := core.LLMDeps{
			Provider: mockProvider,
		}
		agent := core.NewLLMAgent("benchmark-agent", "mock", deps)
		agent.AddTool(simpleTool)
		agent.AddTool(complexTool)

		ctx := context.Background()
		state := domain.NewState()
		state.Set("user_input", "Use both tools")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = agent.Run(ctx, state.Clone())
		}
	})
}

// BenchmarkWorkflowAgents benchmarks workflow agent performance
func BenchmarkWorkflowAgents(b *testing.B) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()

	// Helper to create agents
	createAgent := func(name string) domain.BaseAgent {
		deps := core.LLMDeps{
			Provider: mockProvider,
		}
		return core.NewLLMAgent(name, "mock", deps)
	}

	b.Run("SequentialWorkflow", func(b *testing.B) {
		seq := workflow.NewSequentialAgent("seq-bench")
		seq.AddAgent(createAgent("step1"))
		seq.AddAgent(createAgent("step2"))
		seq.AddAgent(createAgent("step3"))

		ctx := context.Background()
		state := domain.NewState()
		state.Set("input", "test data")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = seq.Run(ctx, state.Clone())
		}
	})

	b.Run("ParallelWorkflow", func(b *testing.B) {
		par := workflow.NewParallelAgent("par-bench")
		par.AddAgent(createAgent("parallel1"))
		par.AddAgent(createAgent("parallel2"))
		par.AddAgent(createAgent("parallel3"))

		ctx := context.Background()
		state := domain.NewState()
		state.Set("input", "test data")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = par.Run(ctx, state.Clone())
		}
	})

	b.Run("ConditionalWorkflow", func(b *testing.B) {
		cond := workflow.NewConditionalAgent("cond-bench")
		cond.AddAgent(
			"branch1",
			func(state *domain.State) bool {
				val, _ := state.Get("value")
				if num, ok := val.(int); ok {
					return num < 50
				}
				return false
			},
			createAgent("low-branch"),
		)
		cond.AddAgent(
			"branch2",
			func(state *domain.State) bool {
				val, _ := state.Get("value")
				if num, ok := val.(int); ok {
					return num >= 50
				}
				return false
			},
			createAgent("high-branch"),
		)

		ctx := context.Background()
		state := domain.NewState()
		state.Set("value", 75)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cond.Run(ctx, state.Clone())
		}
	})

	b.Run("LoopWorkflow", func(b *testing.B) {
		loop := workflow.NewLoopAgent("loop-bench").
			WithMaxIterations(3)
		loop.SetLoopAgent(createAgent("loop-body"))

		ctx := context.Background()
		state := domain.NewState()
		state.Set("counter", 0)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = loop.Run(ctx, state.Clone())
		}
	})
}

// BenchmarkHookExecution benchmarks hook overhead
func BenchmarkHookExecution(b *testing.B) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()

	b.Run("NoHooks", func(b *testing.B) {
		deps := core.LLMDeps{
			Provider: mockProvider,
		}
		agent := core.NewLLMAgent("benchmark-agent", "mock", deps)

		ctx := context.Background()
		state := domain.NewState()
		state.Set("user_input", "test")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = agent.Run(ctx, state.Clone())
		}
	})

	b.Run("SingleHook", func(b *testing.B) {
		deps := core.LLMDeps{
			Provider: mockProvider,
		}
		agent := core.NewLLMAgent("benchmark-agent", "mock", deps)
		agent.WithHook(core.NewLoggingHook(nil, core.LogLevelBasic))

		ctx := context.Background()
		state := domain.NewState()
		state.Set("user_input", "test")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = agent.Run(ctx, state.Clone())
		}
	})

	b.Run("MultipleHooks", func(b *testing.B) {
		deps := core.LLMDeps{
			Provider: mockProvider,
		}
		agent := core.NewLLMAgent("benchmark-agent", "mock", deps)
		agent.WithHook(core.NewLoggingHook(nil, core.LogLevelBasic))
		agent.WithHook(core.NewLLMMetricsHook())

		ctx := context.Background()
		state := domain.NewState()
		state.Set("user_input", "test")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = agent.Run(ctx, state.Clone())
		}
	})
}

// BenchmarkEventStream benchmarks event streaming performance
func BenchmarkEventStream(b *testing.B) {
	b.Run("EventCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = domain.NewEvent(domain.EventAgentStart, "test-agent-id", "test", map[string]interface{}{
				"agent": "test",
				"time":  time.Now(),
			})
		}
	})

	b.Run("EventStreamOperations", func(b *testing.B) {
		// Create a stream with events
		events := make([]domain.Event, 100)
		for i := range events {
			events[i] = domain.NewEvent(domain.EventToolCall, "tool-agent-id", "test_tool", map[string]interface{}{
				"tool":  "test_tool",
				"index": i,
			})
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			ch := make(chan domain.Event, len(events))

			// Send events to channel
			go func() {
				for _, e := range events {
					ch <- e
				}
				close(ch)
			}()

			stream := domain.NewFunctionalEventStream(ctx, ch)

			// Chain operations
			filtered := stream.Filter(func(e domain.Event) bool {
				return e.Type == domain.EventToolCall
			})

			mapped := filtered.Map(func(e domain.Event) domain.Event {
				// Type assert Data to map before modifying
				if data, ok := e.Data.(map[string]interface{}); ok {
					data["processed"] = true
				}
				return e
			})

			// Force evaluation
			_, _ = mapped.Collect()
		}
	})
}
