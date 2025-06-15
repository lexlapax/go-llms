// ABOUTME: Test scenario builders for complex testing patterns
// ABOUTME: Provides fluent API for building multi-component test scenarios

package fixtures

import (
	"context"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// ScenarioBuilder provides a fluent API for building complex test scenarios
type ScenarioBuilder struct {
	agent    *mocks.MockAgent
	provider *mocks.MockProvider
	tools    []*mocks.MockTool
	state    *domain.State
	emitter  *mocks.MockEventEmitter
	context  context.Context
	timeout  time.Duration
	metadata map[string]interface{}
}

// NewScenario creates a new scenario builder
func NewScenario() *ScenarioBuilder {
	return &ScenarioBuilder{
		tools:    make([]*mocks.MockTool, 0),
		state:    domain.NewState(),
		context:  context.Background(),
		timeout:  30 * time.Second,
		metadata: make(map[string]interface{}),
	}
}

// WithAgent sets the agent for the scenario
func (s *ScenarioBuilder) WithAgent(agentID string) *ScenarioBuilder {
	s.agent = mocks.NewMockAgent(agentID)
	return s
}

// WithSimpleAgent adds a simple mock agent
func (s *ScenarioBuilder) WithSimpleAgent() *ScenarioBuilder {
	s.agent = SimpleMockAgent()
	return s
}

// WithResearchAgent adds a research mock agent
func (s *ScenarioBuilder) WithResearchAgent() *ScenarioBuilder {
	s.agent = ResearchMockAgent()
	return s
}

// WithWorkflowAgent adds a workflow mock agent
func (s *ScenarioBuilder) WithWorkflowAgent() *ScenarioBuilder {
	s.agent = WorkflowMockAgent()
	return s
}

// WithComplexWorkflowAgent adds a complex workflow mock agent
func (s *ScenarioBuilder) WithComplexWorkflowAgent() *ScenarioBuilder {
	s.agent = ComplexWorkflowMockAgent()
	return s
}

// WithConcurrentAgent adds a concurrent mock agent
func (s *ScenarioBuilder) WithConcurrentAgent() *ScenarioBuilder {
	s.agent = ConcurrentMockAgent()
	return s
}

// WithErrorRecoveryAgent adds an error recovery mock agent
func (s *ScenarioBuilder) WithErrorRecoveryAgent() *ScenarioBuilder {
	s.agent = ErrorRecoveryMockAgent()
	return s
}

// WithStatefulAgent adds a stateful mock agent
func (s *ScenarioBuilder) WithStatefulAgent() *ScenarioBuilder {
	s.agent = StatefulMockAgent()
	return s
}

// WithProvider sets the LLM provider for the scenario
func (s *ScenarioBuilder) WithProvider(providerID string) *ScenarioBuilder {
	s.provider = mocks.NewMockProvider(providerID)
	return s
}

// WithChatGPTProvider adds a ChatGPT mock provider
func (s *ScenarioBuilder) WithChatGPTProvider() *ScenarioBuilder {
	s.provider = ChatGPTMockProvider()
	return s
}

// WithClaudeProvider adds a Claude mock provider
func (s *ScenarioBuilder) WithClaudeProvider() *ScenarioBuilder {
	s.provider = ClaudeMockProvider()
	return s
}

// WithErrorProvider adds an error-prone provider
func (s *ScenarioBuilder) WithErrorProvider(errorType string) *ScenarioBuilder {
	s.provider = ErrorMockProvider(errorType)
	return s
}

// WithSlowProvider adds a slow provider with delay
func (s *ScenarioBuilder) WithSlowProvider(delay time.Duration) *ScenarioBuilder {
	s.provider = SlowMockProvider(delay)
	return s
}

// WithStreamingProvider adds a streaming provider
func (s *ScenarioBuilder) WithStreamingProvider() *ScenarioBuilder {
	s.provider = StreamingMockProvider()
	return s
}

// WithTool adds a tool to the scenario
func (s *ScenarioBuilder) WithTool(tool *mocks.MockTool) *ScenarioBuilder {
	s.tools = append(s.tools, tool)
	return s
}

// WithCalculatorTool adds a calculator tool
func (s *ScenarioBuilder) WithCalculatorTool() *ScenarioBuilder {
	s.tools = append(s.tools, CalculatorMockTool())
	return s
}

// WithWebSearchTool adds a web search tool
func (s *ScenarioBuilder) WithWebSearchTool() *ScenarioBuilder {
	s.tools = append(s.tools, WebSearchMockTool())
	return s
}

// WithFileTool adds a file management tool
func (s *ScenarioBuilder) WithFileTool() *ScenarioBuilder {
	s.tools = append(s.tools, FileMockTool())
	return s
}

// WithErrorTool adds a tool that randomly fails
func (s *ScenarioBuilder) WithErrorTool(errorRate float64) *ScenarioBuilder {
	s.tools = append(s.tools, ErrorMockTool(errorRate))
	return s
}

// WithState sets initial state data
func (s *ScenarioBuilder) WithState(key string, value interface{}) *ScenarioBuilder {
	s.state.Set(key, value)
	return s
}

// WithMessage adds a message to the initial state
func (s *ScenarioBuilder) WithMessage(role domain.Role, content string) *ScenarioBuilder {
	s.state.AddMessage(domain.NewMessage(role, content))
	return s
}

// WithUserMessage adds a user message
func (s *ScenarioBuilder) WithUserMessage(content string) *ScenarioBuilder {
	return s.WithMessage(domain.RoleUser, content)
}

// WithAssistantMessage adds an assistant message
func (s *ScenarioBuilder) WithAssistantMessage(content string) *ScenarioBuilder {
	return s.WithMessage(domain.RoleAssistant, content)
}

// WithSystemMessage adds a system message
func (s *ScenarioBuilder) WithSystemMessage(content string) *ScenarioBuilder {
	return s.WithMessage(domain.RoleSystem, content)
}

// WithEventEmitter adds an event emitter
func (s *ScenarioBuilder) WithEventEmitter(agentID, agentName string) *ScenarioBuilder {
	s.emitter = mocks.NewMockEventEmitter(agentID, agentName)
	return s
}

// WithContext sets the context for the scenario
func (s *ScenarioBuilder) WithContext(ctx context.Context) *ScenarioBuilder {
	s.context = ctx
	return s
}

// WithTimeout sets the timeout for the scenario
func (s *ScenarioBuilder) WithTimeout(timeout time.Duration) *ScenarioBuilder {
	s.timeout = timeout
	return s
}

// WithMetadata adds metadata to the scenario
func (s *ScenarioBuilder) WithMetadata(key string, value interface{}) *ScenarioBuilder {
	s.metadata[key] = value
	return s
}

// Scenario represents a fully built test scenario
type Scenario struct {
	Agent    *mocks.MockAgent
	Provider *mocks.MockProvider
	Tools    []*mocks.MockTool
	State    *domain.State
	Emitter  *mocks.MockEventEmitter
	Context  context.Context
	Timeout  time.Duration
	Metadata map[string]interface{}
}

// Build creates the final scenario
func (s *ScenarioBuilder) Build() *Scenario {
	// Set defaults if not provided
	if s.agent == nil {
		s.agent = SimpleMockAgent()
	}
	if s.provider == nil {
		s.provider = ChatGPTMockProvider()
	}
	if s.emitter == nil {
		s.emitter = mocks.NewMockEventEmitter("test-agent", "Test Agent")
	}

	return &Scenario{
		Agent:    s.agent,
		Provider: s.provider,
		Tools:    s.tools,
		State:    s.state,
		Emitter:  s.emitter,
		Context:  s.context,
		Timeout:  s.timeout,
		Metadata: s.metadata,
	}
}

// Execute runs the scenario and returns the result
func (scenario *Scenario) Execute() (*domain.State, error) {
	if scenario.Agent == nil {
		return nil, nil
	}

	// Add timeout to context if needed
	ctx := scenario.Context
	if scenario.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, scenario.Timeout)
		defer cancel()
	}

	// Execute the agent with the state
	return scenario.Agent.Run(ctx, scenario.State)
}

// CreateToolContext creates a tool context from the scenario
func (scenario *Scenario) CreateToolContext() *domain.ToolContext {
	stateReader := domain.NewStateReader(scenario.State)
	tc := domain.NewToolContext(scenario.Context, stateReader, scenario.Agent, "test-run")

	if scenario.Emitter != nil {
		tc = tc.WithEventEmitter(scenario.Emitter)
	}

	return tc
}

// Pre-built scenario templates

// SimpleAgentScenario creates a basic agent scenario
func SimpleAgentScenario() *Scenario {
	return NewScenario().
		WithSimpleAgent().
		WithChatGPTProvider().
		WithUserMessage("Hello, world!").
		Build()
}

// ResearchScenario creates a research workflow scenario
func ResearchScenario(query string) *Scenario {
	return NewScenario().
		WithResearchAgent().
		WithChatGPTProvider().
		WithWebSearchTool().
		WithState("query", query).
		WithUserMessage("Please research: " + query).
		Build()
}

// CalculationScenario creates a mathematical calculation scenario
func CalculationScenario(operation string, a, b float64) *Scenario {
	return NewScenario().
		WithSimpleAgent().
		WithChatGPTProvider().
		WithCalculatorTool().
		WithState("operation", operation).
		WithState("a", a).
		WithState("b", b).
		WithUserMessage("Please calculate").
		Build()
}

// FileProcessingScenario creates a file processing scenario
func FileProcessingScenario(operation, path string) *Scenario {
	builder := NewScenario().
		WithWorkflowAgent().
		WithChatGPTProvider().
		WithFileTool().
		WithState("operation", operation).
		WithState("path", path)

	if operation == "write" {
		builder = builder.WithState("content", "Test file content")
	}

	return builder.WithUserMessage("Please process file: " + path).Build()
}

// ErrorHandlingScenario creates a scenario that tests error handling
func ErrorHandlingScenario() *Scenario {
	return NewScenario().
		WithSimpleAgent().
		WithErrorProvider("rate_limit").
		WithErrorTool(0.5). // 50% error rate
		WithUserMessage("Test error handling").
		WithTimeout(5 * time.Second).
		Build()
}

// StreamingScenario creates a scenario that tests streaming responses
func StreamingScenario() *Scenario {
	return NewScenario().
		WithSimpleAgent().
		WithStreamingProvider().
		WithUserMessage("Tell me a story").
		WithTimeout(10 * time.Second).
		Build()
}

// MultiToolScenario creates a scenario with multiple tools
func MultiToolScenario() *Scenario {
	return NewScenario().
		WithWorkflowAgent().
		WithChatGPTProvider().
		WithCalculatorTool().
		WithWebSearchTool().
		WithFileTool().
		WithUserMessage("Use multiple tools to solve this complex task").
		WithTimeout(15 * time.Second).
		Build()
}

// ConversationScenario creates a multi-turn conversation scenario
func ConversationScenario() *Scenario {
	return NewScenario().
		WithStatefulAgent().
		WithChatGPTProvider().
		WithUserMessage("Hello, I'm starting a conversation").
		WithAssistantMessage("Hello! I'm ready to help you").
		WithUserMessage("Let's continue our discussion").
		Build()
}

// HookTestingScenario creates a scenario for testing hook functionality
func HookTestingScenario(hookImpl interface{}, testType string) *Scenario {
	builder := NewScenario().
		WithSimpleAgent().
		WithChatGPTProvider()

	// Configure based on test type
	switch testType {
	case "basic":
		builder = builder.
			WithUserMessage("Calculate 2 + 2").
			WithMetadata("expected_hooks", map[string]int{
				"before_generate": 1,
				"after_generate":  1,
				"before_tool":     0,
				"after_tool":      0,
			})

	case "with_tools":
		builder = builder.
			WithCalculatorTool().
			WithUserMessage("Calculate 2 + 2").
			WithMetadata("expected_hooks", map[string]int{
				"before_generate": 2, // Initial + after tool result
				"after_generate":  2,
				"before_tool":     1,
				"after_tool":      1,
			}).
			WithMetadata("expected_order", []string{
				"BeforeGenerate",
				"AfterGenerate",
				"BeforeToolCall:calculator",
				"AfterToolCall:calculator",
				"BeforeGenerate",
				"AfterGenerate",
			})

	case "error_handling":
		// Use error provider
		builder = builder.
			WithErrorProvider("provider_error").
			WithUserMessage("Test input").
			WithMetadata("expected_error", true).
			WithMetadata("expected_hooks", map[string]int{
				"before_generate": 1,
				"after_generate":  1,
				"before_tool":     0,
				"after_tool":      0,
			})

	case "concurrent":
		builder = builder.
			WithUserMessage("Concurrent test input").
			WithMetadata("concurrent_runs", 10).
			WithMetadata("expected_hooks_per_run", map[string]int{
				"before_generate": 1,
				"after_generate":  1,
			})

	default:
		builder = builder.WithUserMessage("Default test input")
	}

	scenario := builder.Build()

	// Store hook reference for later verification
	if hookImpl != nil {
		scenario.Metadata["test_hook"] = hookImpl
	}

	return scenario
}

// WorkflowHookScenario creates a scenario for testing hooks in workflow agents
func WorkflowHookScenario(agent1Hook, agent2Hook interface{}) *Scenario {
	return NewScenario().
		WithWorkflowAgent().
		WithChatGPTProvider().
		WithUserMessage("Process this sequentially").
		WithMetadata("agent1_hook", agent1Hook).
		WithMetadata("agent2_hook", agent2Hook).
		WithMetadata("expected_responses", []string{
			"Result from agent 1",
			"Result from agent 2",
		}).
		Build()
}

// MetricsHookScenario creates a scenario for testing metrics hooks
func MetricsHookScenario(numRuns int) *Scenario {
	return NewScenario().
		WithSimpleAgent().
		WithChatGPTProvider().
		WithUserMessage("Test input for metrics").
		WithMetadata("num_runs", numRuns).
		WithMetadata("expected_metrics", map[string]interface{}{
			"requests":    numRuns,
			"error_count": 0,
		}).
		Build()
}
