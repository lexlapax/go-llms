// ABOUTME: Package testutils provides comprehensive testing infrastructure for go-llms
// ABOUTME: Includes fixtures, mocks, scenario builders, matchers, and test helpers

/*
Package testutils provides a comprehensive testing infrastructure for the go-llms project.

# Overview

The testutils package is designed to make testing LLM-related code easier, more consistent,
and more maintainable. It provides:

- Pre-configured mock objects (fixtures) for common testing scenarios
- A fluent scenario builder API for complex test setups
- Flexible assertion capabilities through matchers
- Comprehensive helper utilities for context creation and state management

# Core Components

## Fixtures

Fixtures provide pre-configured mock objects that simulate realistic behavior:

	// Provider fixtures
	provider := fixtures.ChatGPTMockProvider()        // ChatGPT-like responses
	provider := fixtures.ClaudeMockProvider()         // Claude-like responses
	provider := fixtures.ErrorMockProvider("auth")   // Error simulation
	provider := fixtures.SlowMockProvider(2*time.Second) // Slow response simulation

	// Tool fixtures
	calc := fixtures.CalculatorMockTool()             // Arithmetic operations
	web := fixtures.WebSearchMockTool()               // Web search simulation
	file := fixtures.FileMockTool()                   // File operations

	// Agent fixtures
	agent := fixtures.SimpleMockAgent()               // Basic agent behavior
	agent := fixtures.ResearchMockAgent()             // Research workflows
	agent := fixtures.WorkflowMockAgent()             // Complex workflows

	// State fixtures
	state := fixtures.BasicTestState()                // Basic test data
	state := fixtures.StateWithArtifacts()            // States with artifacts
	state := fixtures.ConversationTestState()         // Conversation history

## Scenario Builder

The scenario builder provides a fluent API for complex test setups:

	scenario.NewScenario(t).
		WithMockProvider("chatgpt", map[string]mocks.Response{
			"(?i).*hello.*": {Content: "Hello! How can I help?"},
		}).
		WithTool(fixtures.CalculatorMockTool()).
		WithAgent(fixtures.ResearchMockAgent()).
		WithInput("query", "research quantum computing").
		ExpectOutput("task_type", matchers.Equals("research")).
		ExpectOutput("query", matchers.Contains("quantum")).
		ExpectNoError().
		Run()

## Matchers

Matchers provide flexible assertion capabilities:

	matchers.Equals("expected")                       // Exact equality
	matchers.Contains("substring")                    // String contains
	matchers.HasPrefix("prefix")                      // String prefix
	matchers.MatchesRegex("pattern")                  // Regex matching
	matchers.IsType[string]()                         // Type checking
	matchers.IsNil()                                  // Nil checking
	matchers.HasLength(5)                             // Length checking

## Helpers

Helpers provide utilities for common testing tasks:

	// Context creation
	ctx := helpers.CreateTestToolContext()
	ctx := helpers.CreateToolContextWithState(data)

	// Event testing
	capture := helpers.NewEventCapture()
	events := capture.GetEvents()
	helpers.AssertEvents(t, events).
		HasType("agent.start").
		HasType("tool.execute").
		InOrder()

# Usage Patterns

## Basic Provider Testing

	func TestProviderGeneration(t *testing.T) {
		provider := fixtures.ChatGPTMockProvider()

		response, err := provider.Generate(context.Background(), "Hello!")
		assert.NoError(t, err)
		assert.Contains(t, response, "Hello")
	}

## Tool Testing with Context

	func TestToolExecution(t *testing.T) {
		tool := fixtures.CalculatorMockTool()
		ctx := helpers.CreateTestToolContext()

		result, err := tool.Execute(ctx, map[string]interface{}{
			"operation": "add",
			"a": 5.0,
			"b": 3.0,
		})

		assert.NoError(t, err)
		assert.Equal(t, 8.0, result["result"])
	}

## Agent Workflow Testing

	func TestAgentWorkflow(t *testing.T) {
		agent := fixtures.ResearchMockAgent()
		input := fixtures.BasicTestState()
		input.Set("query", "AI research trends")

		result, err := agent.Run(context.Background(), input)
		assert.NoError(t, err)

		taskType, _ := result.Get("task_type")
		assert.Equal(t, "research", taskType)
	}

## Complex Scenario Testing

	func TestComplexWorkflow(t *testing.T) {
		scenario.NewScenario(t).
			WithMockProvider("claude", map[string]mocks.Response{
				"(?i).*analyze.*": {
					Content: "Analysis complete",
					Metadata: map[string]interface{}{
						"confidence": 0.95,
					},
				},
			}).
			WithTool(fixtures.WebSearchMockTool()).
			WithAgent(fixtures.WorkflowMockAgent()).
			WithInput("task", "analyze market trends").
			ExpectOutput("status", matchers.Equals("completed")).
			ExpectOutput("confidence", matchers.GreaterThan(0.9)).
			ExpectMetadata("execution_time", matchers.IsType[time.Duration]()).
			ExpectNoError().
			Run()
	}

# Migration from Legacy Testing

For existing code using the old mock implementations, see the migration guide:

	// OLD - Manual mock setup
	mock := provider.NewMockProvider()
	mock.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		return "custom response", nil
	})

	// NEW - Using fixtures
	provider := fixtures.ChatGPTMockProvider()
	provider.WithPatternResponse("(?i).*custom.*", mocks.Response{
		Content: "custom response",
	})

# Performance Considerations

The testing infrastructure is designed for performance:

- Fixtures are lazily initialized and cached
- Mock responses use efficient pattern matching
- State operations are optimized for common test cases
- Thread-safe implementations support concurrent testing

# Thread Safety

All components in the testutils package are thread-safe:

- Fixtures can be used concurrently across goroutines
- Mock providers handle concurrent requests safely
- Event capture works correctly with parallel test execution
- State management is protected with appropriate synchronization

# Best Practices

1. Use specific fixtures rather than generic mocks when possible
2. Leverage pattern-based responses for realistic behavior simulation
3. Use the scenario builder for complex multi-component tests
4. Apply matchers for flexible and readable assertions
5. Take advantage of helper utilities to reduce boilerplate
6. Consider performance implications when creating large test suites

For detailed examples and migration guidance, see the documentation in the
respective subpackages and the TESTING_MIGRATION_GUIDE.md file.

# Package Structure

This package is organized into subpackages:

- fixtures: Pre-configured mock objects for common testing scenarios
- helpers: Utility functions for context creation, event testing, and state management
- matchers: Flexible assertion capabilities for test validation
- mocks: Core mock implementations and registry
- scenario: Fluent API for complex test scenario setup
*/
package testutils
