// ABOUTME: LLM-powered agent implementation with full Phase 1.5 component integration
// ABOUTME: Provides state-based execution, type-safe dependencies, and comprehensive observability

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// LLMDeps defines the dependencies for LLM agents.
// It encapsulates the required external services including the LLM provider,
// logger for observability, and optional tracer for distributed tracing.
type LLMDeps struct {
	Provider ldomain.Provider
	Logger   *slog.Logger
	Tracer   any // trace.Tracer - keeping as interface to avoid OpenTelemetry dependency
}

// LLMAgent implements BaseAgent for LLM-powered agents.
// It provides full integration with language models, tool execution, state management,
// guardrails, transforms, and comprehensive observability through hooks and events.
// The agent supports hierarchical organization with sub-agents and handoffs.
type LLMAgent struct {
	*BaseAgentImpl
	mu sync.RWMutex

	// LLM Configuration
	systemPrompt string
	modelName    string

	// Tools
	tools map[string]domain.Tool

	// Hooks
	hooks []domain.Hook

	// Enhanced Components
	inputGuardrails  domain.Guardrail
	outputGuardrails domain.Guardrail
	inputTransforms  []StateTransform
	outputTransforms []StateTransform
	handoffs         map[string]domain.Handoff
	tracingHook      *TracingHook

	// Dependencies
	deps LLMDeps

	// Optimization: cached tool descriptions and names
	cachedToolsDescription string
	cachedToolNames        []string

	// Event handling
	eventStream domain.FunctionalEventStream
}

// NewLLMAgent creates a new LLM agent with the specified dependencies.
// The agent is initialized with empty tool and hook collections. Use the fluent
// API methods (WithTool, WithHook, etc.) to configure the agent after creation.
// The agent automatically manages sub-agent handoffs when sub-agents are added.
func NewLLMAgent(name, description string, deps LLMDeps) *LLMAgent {
	baseAgent := NewBaseAgent(name, description, domain.AgentTypeLLM)

	agent := &LLMAgent{
		BaseAgentImpl:    baseAgent,
		deps:             deps,
		tools:            make(map[string]domain.Tool),
		hooks:            make([]domain.Hook, 0),
		handoffs:         make(map[string]domain.Handoff),
		inputTransforms:  make([]StateTransform, 0),
		outputTransforms: make([]StateTransform, 0),
	}

	// Don't add transfer_to_agent tool by default - it will be added when first sub-agent is added

	return agent
}

// NewAgent creates a simple LLM agent with minimal configuration.
// This is a convenience function that provides excellent developer experience
// by requiring only a name and provider. Logger defaults to slog.Default().
// Use NewLLMAgent for more control over dependencies.
func NewAgent(name string, provider ldomain.Provider) *LLMAgent {
	deps := LLMDeps{
		Provider: provider,
		Logger:   slog.Default(),
		Tracer:   nil, // Optional tracing
	}

	return NewLLMAgent(name, "LLM Agent", deps)
}

// NewAgentWithLogger creates an LLM agent with custom logger.
// This convenience function allows specifying a custom logger while keeping
// other dependencies at their defaults. Useful for integrating with existing
// logging infrastructure.
func NewAgentWithLogger(name string, provider ldomain.Provider, logger *slog.Logger) *LLMAgent {
	deps := LLMDeps{
		Provider: provider,
		Logger:   logger,
		Tracer:   nil,
	}

	return NewLLMAgent(name, "LLM Agent", deps)
}

// NewAgentFromString creates an LLM agent from a provider/model string specification.
// It parses the providerModel string to determine the LLM provider and model to use.
//
// Examples:
//   - "openai/gpt-4" - specific provider and model
//   - "gpt-4" - model with inferred provider
//   - "openai" - provider with default model
//   - "claude" - alias for anthropic/claude-3-opus
func NewAgentFromString(name, providerModel string) (*LLMAgent, error) {
	provider, err := llmutil.NewProviderFromString(providerModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider from '%s': %w", providerModel, err)
	}

	return NewAgent(name, provider), nil
}

// NewAgentFromStringWithLogger creates an LLM agent from a string with custom logger
func NewAgentFromStringWithLogger(name, providerModel string, logger *slog.Logger) (*LLMAgent, error) {
	provider, err := llmutil.NewProviderFromString(providerModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider from '%s': %w", providerModel, err)
	}

	return NewAgentWithLogger(name, provider, logger), nil
}

// NewLLMAgentWithSubAgents creates an LLM agent with sub-agents in one call
// This provides a Google ADK-like API for creating multi-agent systems
func NewLLMAgentWithSubAgents(name string, provider ldomain.Provider, subAgents ...domain.BaseAgent) (*LLMAgent, error) {
	agent := NewAgent(name, provider)

	// Add each sub-agent
	for _, subAgent := range subAgents {
		if err := agent.AddSubAgent(subAgent); err != nil {
			return nil, fmt.Errorf("failed to add sub-agent %s: %w", subAgent.Name(), err)
		}
	}

	return agent, nil
}

// NewLLMAgentWithSubAgentsFromString creates an LLM agent with sub-agents from a provider string
func NewLLMAgentWithSubAgentsFromString(name, providerModel string, subAgents ...domain.BaseAgent) (*LLMAgent, error) {
	agent, err := NewAgentFromString(name, providerModel)
	if err != nil {
		return nil, err
	}

	// Add each sub-agent
	for _, subAgent := range subAgents {
		if err := agent.AddSubAgent(subAgent); err != nil {
			return nil, fmt.Errorf("failed to add sub-agent %s: %w", subAgent.Name(), err)
		}
	}

	return agent, nil
}

// Configuration Methods

// SetSystemPrompt sets the system prompt for the agent
func (a *LLMAgent) SetSystemPrompt(prompt string) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.systemPrompt = prompt
	// Invalidate cached tool description
	a.cachedToolsDescription = ""
	return a
}

// WithModel sets the model name to use
func (a *LLMAgent) WithModel(modelName string) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.modelName = modelName
	return a
}

// Tool Management

// AddTool adds a tool to the agent
func (a *LLMAgent) AddTool(tool domain.Tool) *LLMAgent {
	if tool == nil {
		return a
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.tools[tool.Name()] = tool
	// Invalidate cached tool data
	a.cachedToolsDescription = ""
	a.cachedToolNames = nil

	return a
}

// RemoveTool removes a tool from the agent
func (a *LLMAgent) RemoveTool(name string) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.tools, name)
	// Invalidate cached tool data
	a.cachedToolsDescription = ""
	a.cachedToolNames = nil

	return a
}

// GetTool retrieves a tool by name
func (a *LLMAgent) GetTool(name string) (domain.Tool, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	tool, exists := a.tools[name]
	return tool, exists
}

// ListTools returns all tool names
func (a *LLMAgent) ListTools() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.cachedToolNames != nil {
		return a.cachedToolNames
	}

	names := make([]string, 0, len(a.tools))
	for name := range a.tools {
		names = append(names, name)
	}

	a.cachedToolNames = names
	return names
}

// WithSubAgents adds multiple sub-agents at once (builder pattern)
func (a *LLMAgent) WithSubAgents(subAgents ...domain.BaseAgent) *LLMAgent {
	for _, subAgent := range subAgents {
		if err := a.AddSubAgent(subAgent); err != nil {
			// Log error but continue - builder pattern should not fail
			if a.deps.Logger != nil {
				a.deps.Logger.Error("Failed to add sub-agent",
					"agent", a.name,
					"sub_agent", subAgent.Name(),
					"error", err)
			}
		}
	}
	return a
}

// AddSubAgent adds a sub-agent and automatically registers it as a tool
func (a *LLMAgent) AddSubAgent(agent domain.BaseAgent) error {
	// First, call the parent implementation
	if err := a.BaseAgentImpl.AddSubAgent(agent); err != nil {
		return err
	}

	// Add transfer_to_agent tool if this is the first sub-agent
	subAgents := a.SubAgents()
	if len(subAgents) == 1 {
		_, hasTransferTool := a.GetTool("transfer_to_agent")
		if !hasTransferTool {
			a.AddTool(&transferToAgentTool{agent: a})
		}
	}

	// Create a tool wrapper for the sub-agent
	// We can't use tools.AgentTool due to circular dependency, so we create a simple wrapper
	toolWrapper := &subAgentTool{
		subAgent: agent,
	}

	// Add the sub-agent as a tool
	a.AddTool(toolWrapper)

	return nil
}

// RemoveSubAgent removes a sub-agent and its corresponding tool
func (a *LLMAgent) RemoveSubAgent(name string) error {
	// First, remove the tool
	a.RemoveTool(name)

	// Then call parent implementation
	err := a.BaseAgentImpl.RemoveSubAgent(name)
	if err != nil {
		return err
	}

	// Remove transfer_to_agent tool if no more sub-agents
	if len(a.SubAgents()) == 0 {
		a.RemoveTool("transfer_to_agent")
	}

	return nil
}

// TransferTo transfers control to a sub-agent by name with an optional reason
// This is a convenience method that internally uses the transfer_to_agent tool
func (a *LLMAgent) TransferTo(ctx context.Context, agentName string, reason string, input interface{}) (*domain.State, error) {
	// Check if the sub-agent exists
	subAgent := a.FindSubAgent(agentName)
	if subAgent == nil {
		return nil, fmt.Errorf("sub-agent '%s' not found", agentName)
	}

	// Create a handoff
	handoff := domain.NewSimpleHandoff("transfer", agentName)

	// Create state for the transfer
	state := domain.NewState()

	// Add the reason if provided
	if reason != "" {
		state.SetMetadata("transfer_reason", reason)
	}

	// Handle input based on type
	switch v := input.(type) {
	case *domain.State:
		// If input is already a state, merge it
		for key, value := range v.Values() {
			state.Set(key, value)
		}
	case string:
		state.Set("input", v)
	case map[string]interface{}:
		for key, value := range v {
			state.Set(key, value)
		}
	default:
		state.Set("input", input)
	}

	// Execute the handoff
	return handoff.Execute(ctx, state)
}

// GetSubAgentByName retrieves a sub-agent by name
// This is an alias for FindSubAgent for Google ADK compatibility
func (a *LLMAgent) GetSubAgentByName(name string) domain.BaseAgent {
	return a.FindSubAgent(name)
}

// Enhanced Components Integration

// WithInputGuardrails sets input validation guardrails
func (a *LLMAgent) WithInputGuardrails(guardrails domain.Guardrail) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.inputGuardrails = guardrails
	return a
}

// WithOutputGuardrails sets output validation guardrails
func (a *LLMAgent) WithOutputGuardrails(guardrails domain.Guardrail) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.outputGuardrails = guardrails
	return a
}

// WithInputTransforms adds input state transformations
func (a *LLMAgent) WithInputTransforms(transforms ...StateTransform) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.inputTransforms = append(a.inputTransforms, transforms...)
	return a
}

// WithOutputTransforms adds output state transformations
func (a *LLMAgent) WithOutputTransforms(transforms ...StateTransform) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.outputTransforms = append(a.outputTransforms, transforms...)
	return a
}

// WithHandoff adds an agent handoff capability
func (a *LLMAgent) WithHandoff(name string, handoff domain.Handoff) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.handoffs[name] = handoff
	return a
}

// WithTracing sets the tracing hook
func (a *LLMAgent) WithTracing(hook *TracingHook) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tracingHook = hook
	return a
}

// WithEventStream sets the event stream for functional event processing
func (a *LLMAgent) WithEventStream(stream domain.FunctionalEventStream) *LLMAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.eventStream = stream
	return a
}

// WithHook adds a monitoring hook to the agent
func (a *LLMAgent) WithHook(hook domain.Hook) *LLMAgent {
	if hook == nil {
		return a
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.hooks = append(a.hooks, hook)
	return a
}

// Core Execution Methods

// Run executes the agent with the given state
func (a *LLMAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Create run context with dependencies
	runCtx := domain.NewRunContextWithState(ctx, a.deps, input)
	runCtx = runCtx.WithEventEmitter(func(event domain.Event) {
		a.EmitEvent(event.Type, event.Data)
	})

	// 1. Tracing: Start span
	if a.tracingHook != nil {
		var err error
		ctx, err = a.tracingHook.BeforeRun(ctx, a, input)
		if err != nil {
			return nil, fmt.Errorf("tracing hook failed: %w", err)
		}
		defer func() {
			_ = a.tracingHook.AfterRun(ctx, a, input, nil, nil)
		}()
	}

	// 2. Input guardrails validation
	if a.inputGuardrails != nil {
		if err := a.inputGuardrails.Validate(ctx, input); err != nil {
			return nil, domain.NewAgentError(a.ID(), a.Name(), "input_validation", err)
		}
	}

	// 3. Apply input transformations
	processedState := input
	for _, transform := range a.inputTransforms {
		var err error
		processedState, err = transform(ctx, processedState)
		if err != nil {
			return nil, fmt.Errorf("input transform failed: %w", err)
		}
	}

	// 4. Execute with retry logic
	var result *domain.State
	err := a.ExecuteWithRetry(ctx, func() error {
		var execErr error
		result, execErr = a.executeCore(runCtx.WithState(processedState))
		return execErr
	})

	if err != nil {
		return nil, err
	}

	// 5. Apply output transformations
	for _, transform := range a.outputTransforms {
		result, err = transform(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("output transform failed: %w", err)
		}
	}

	// 6. Output guardrails validation
	if a.outputGuardrails != nil {
		if err := a.outputGuardrails.Validate(ctx, result); err != nil {
			return nil, domain.NewAgentError(a.ID(), a.Name(), "output_validation", err)
		}
	}

	return result, nil
}

// RunAsync executes the agent asynchronously, returning an event stream
func (a *LLMAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	eventChan := make(chan domain.Event, 100)

	go func() {
		defer close(eventChan)

		result, err := a.Run(ctx, input)
		if err != nil {
			eventChan <- domain.NewEvent(domain.EventAgentError, a.ID(), a.Name(), err)
		} else {
			eventChan <- domain.NewEvent(domain.EventAgentComplete, a.ID(), a.Name(), result)
		}
	}()

	return eventChan, nil
}

// Core execution logic
func (a *LLMAgent) executeCore(runCtx *domain.RunContext[LLMDeps]) (*domain.State, error) {
	// Extract prompt from state
	prompt, err := a.extractPromptFromState(runCtx.State)
	if err != nil {
		return nil, fmt.Errorf("failed to extract prompt: %w", err)
	}

	// Create messages for conversation
	messages := a.createMessagesFromState(runCtx.State, prompt)

	// Execute agent loop with tool calling
	result, err := a.executeAgentLoop(runCtx, messages)
	if err != nil {
		return nil, err
	}

	// Create result state
	resultState := domain.NewState()
	resultState.Set("result", result)
	resultState.Set("output", result) // Also set output for compatibility
	resultState.Set("prompt", prompt)

	// Copy over any artifacts or metadata from input state
	if runCtx.State != nil {
		for _, artifact := range runCtx.State.Artifacts() {
			resultState.AddArtifact(artifact)
		}
		// Copy metadata by iterating through all artifacts which contain metadata
		for _, artifact := range runCtx.State.Artifacts() {
			if artifact.Metadata != nil {
				for key, value := range artifact.Metadata {
					resultState.SetMetadata(key, value)
				}
			}
		}
	}

	return resultState, nil
}

// Extract prompt from state using various strategies
func (a *LLMAgent) extractPromptFromState(state *domain.State) (string, error) {
	if state == nil {
		return "", fmt.Errorf("state is nil")
	}

	// Try different keys for prompt
	promptKeys := []string{"user_input", "prompt", "input", "message", "query", "text"}

	for _, key := range promptKeys {
		if value, exists := state.Get(key); exists {
			if prompt, ok := value.(string); ok && prompt != "" {
				return prompt, nil
			}
		}
	}

	return "", fmt.Errorf("no valid prompt found in state")
}

// Create messages from state and prompt
func (a *LLMAgent) createMessagesFromState(state *domain.State, prompt string) []ldomain.Message {
	messages := make([]ldomain.Message, 0, 10)

	// Add system message with tool descriptions
	systemContent := a.getSystemContent()
	if systemContent != "" {
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleSystem,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: systemContent}},
		})
	}

	// Check for existing messages in state
	if state != nil {
		if existingMessages, exists := state.Get("messages"); exists {
			if msgs, ok := existingMessages.([]ldomain.Message); ok {
				messages = append(messages, msgs...)
			}
		}
	}

	// Add user input
	messages = append(messages, ldomain.Message{
		Role:    ldomain.RoleUser,
		Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: prompt}},
	})

	return messages
}

// Get system content including prompt and tool descriptions
func (a *LLMAgent) getSystemContent() string {
	// First check with read lock
	a.mu.RLock()
	if a.cachedToolsDescription != "" {
		cached := a.cachedToolsDescription
		a.mu.RUnlock()
		return cached
	}
	a.mu.RUnlock()

	// Need to generate content - acquire write lock
	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring write lock
	if a.cachedToolsDescription != "" {
		return a.cachedToolsDescription
	}

	var content strings.Builder

	// Add system prompt
	if a.systemPrompt != "" {
		content.WriteString(a.systemPrompt)
	}

	// Add enhanced tool documentation if tools exist
	if len(a.tools) > 0 {
		if content.Len() > 0 {
			content.WriteString("\n\n")
		}
		content.WriteString("## Available Tools\n\n")

		// Format each tool with full documentation
		for _, tool := range a.tools {
			toolDoc := a.formatToolDocumentation(tool)
			content.WriteString(toolDoc)
			content.WriteString("\n---\n\n")
		}

		// Add general tool usage instructions
		content.WriteString("### Tool Usage Format\n\n")
		content.WriteString("To use a tool, respond with JSON in one of these formats:\n\n")
		content.WriteString("**Simple format:**\n")
		content.WriteString("```json\n")
		content.WriteString(`{"tool": "tool_name", "params": {...}}`)
		content.WriteString("\n```\n\n")
		content.WriteString("**OpenAI format:**\n")
		content.WriteString("```json\n")
		content.WriteString(`{"tool_calls": [{"id": "call_1", "type": "function", "function": {"name": "tool_name", "arguments": "{...}"}}]}`)
		content.WriteString("\n```\n")
	}

	// Cache the result
	a.cachedToolsDescription = content.String()
	return a.cachedToolsDescription
}

// formatToolDocumentation creates comprehensive documentation for a single tool
func (a *LLMAgent) formatToolDocumentation(tool domain.Tool) string {
	// formatToolDocumentation creates comprehensive documentation for a tool that helps
	// LLMs understand how to use it effectively. The function generates:
	// 1. Tool name and description
	// 2. Detailed input/output schemas with type information
	// 3. Usage examples with input/output pairs
	// 4. Error handling guidance
	// 5. Tags for categorization
	//
	// The documentation is formatted in Markdown for better readability by LLMs.
	// Schema formatting includes:
	// - Type information for all fields
	// - Required field indicators
	// - Enum constraints for restricted values
	// - Nested object properties
	//
	// Examples are particularly important as they show:
	// - Realistic usage scenarios
	// - Expected input/output format
	// - Edge cases and error conditions
	var doc strings.Builder

	// Tool name and description
	doc.WriteString(fmt.Sprintf("### %s\n\n", tool.Name()))
	doc.WriteString(fmt.Sprintf("**Description:** %s\n\n", tool.Description()))

	// Version and category
	if tool.Version() != "" {
		doc.WriteString(fmt.Sprintf("**Version:** %s", tool.Version()))
		if tool.Category() != "" {
			doc.WriteString(fmt.Sprintf(" | **Category:** %s", tool.Category()))
		}
		doc.WriteString("\n\n")
	}

	// Behavioral characteristics
	doc.WriteString("**Characteristics:**\n")
	doc.WriteString(fmt.Sprintf("- Deterministic: %v\n", tool.IsDeterministic()))
	doc.WriteString(fmt.Sprintf("- Destructive: %v\n", tool.IsDestructive()))
	doc.WriteString(fmt.Sprintf("- Requires Confirmation: %v\n", tool.RequiresConfirmation()))
	doc.WriteString(fmt.Sprintf("- Estimated Latency: %s\n\n", tool.EstimatedLatency()))

	// Usage instructions
	if instructions := tool.UsageInstructions(); instructions != "" {
		doc.WriteString("**Usage Instructions:**\n")
		doc.WriteString(instructions)
		doc.WriteString("\n\n")
	}

	// Parameter schema
	if schema := tool.ParameterSchema(); schema != nil {
		doc.WriteString("**Parameters:**\n")
		doc.WriteString(a.formatSchema(schema, "  "))
		doc.WriteString("\n")
	}

	// Output schema
	if schema := tool.OutputSchema(); schema != nil {
		doc.WriteString("**Returns:**\n")
		doc.WriteString(a.formatSchema(schema, "  "))
		doc.WriteString("\n")
	}

	// Constraints
	if constraints := tool.Constraints(); len(constraints) > 0 {
		doc.WriteString("**Constraints:**\n")
		for _, constraint := range constraints {
			doc.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
		doc.WriteString("\n")
	}

	// Examples
	if examples := tool.Examples(); len(examples) > 0 {
		doc.WriteString("**Examples:**\n\n")
		for i, example := range examples {
			doc.WriteString(fmt.Sprintf("*Example %d: %s*\n", i+1, example.Name))
			if example.Description != "" {
				doc.WriteString(fmt.Sprintf("Description: %s\n", example.Description))
			}
			if example.Scenario != "" {
				doc.WriteString(fmt.Sprintf("When to use: %s\n", example.Scenario))
			}
			doc.WriteString("\nInput:\n```json\n")
			inputJSON, _ := json.MarshalIndent(example.Input, "", "  ")
			doc.WriteString(string(inputJSON))
			doc.WriteString("\n```\n\nOutput:\n```json\n")
			outputJSON, _ := json.MarshalIndent(example.Output, "", "  ")
			doc.WriteString(string(outputJSON))
			doc.WriteString("\n```\n")
			if example.Explanation != "" {
				doc.WriteString(fmt.Sprintf("\nExplanation: %s\n", example.Explanation))
			}
			doc.WriteString("\n")
		}
	}

	// Error guidance
	if errorGuidance := tool.ErrorGuidance(); len(errorGuidance) > 0 {
		doc.WriteString("**Error Handling:**\n")
		for errorType, guidance := range errorGuidance {
			doc.WriteString(fmt.Sprintf("- `%s`: %s\n", errorType, guidance))
		}
		doc.WriteString("\n")
	}

	// Tags
	if tags := tool.Tags(); len(tags) > 0 {
		doc.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(tags, ", ")))
	}

	return doc.String()
}

// formatSchema converts a schema to a readable string format
func (a *LLMAgent) formatSchema(schema *sdomain.Schema, indent string) string {
	if schema == nil {
		return indent + "No schema defined\n"
	}

	var sb strings.Builder

	// Type and description
	if schema.Type != "" {
		sb.WriteString(fmt.Sprintf("%sType: %s\n", indent, schema.Type))
	}
	if schema.Description != "" {
		sb.WriteString(fmt.Sprintf("%sDescription: %s\n", indent, schema.Description))
	}

	// Properties for object types
	if schema.Type == "object" && len(schema.Properties) > 0 {
		sb.WriteString(fmt.Sprintf("%sProperties:\n", indent))
		for name, prop := range schema.Properties {
			required := false
			for _, req := range schema.Required {
				if req == name {
					required = true
					break
				}
			}
			sb.WriteString(fmt.Sprintf("%s  - %s (%s%s)", indent, name, prop.Type,
				func() string {
					if required {
						return ", required"
					}
					return ""
				}()))
			if prop.Description != "" {
				sb.WriteString(fmt.Sprintf(": %s", prop.Description))
			}
			sb.WriteString("\n")

			// Add enum values if present
			if len(prop.Enum) > 0 {
				sb.WriteString(fmt.Sprintf("%s    Allowed values: %s\n", indent, strings.Join(prop.Enum, ", ")))
			}
		}
	}

	// Required fields
	if len(schema.Required) > 0 && schema.Type == "object" {
		sb.WriteString(fmt.Sprintf("%sRequired fields: %s\n", indent, strings.Join(schema.Required, ", ")))
	}

	return sb.String()
}

// Execute the main agent loop with tool calling
func (a *LLMAgent) executeAgentLoop(runCtx *domain.RunContext[LLMDeps], messages []ldomain.Message) (any, error) {
	maxIterations := 10

	for i := 0; i < maxIterations; i++ {
		// Emit progress event
		runCtx.EmitProgress(i+1, maxIterations, fmt.Sprintf("Agent iteration %d", i+1))

		// Generate LLM response
		var options []ldomain.Option
		if a.modelName != "" {
			options = append(options, ldomain.WithModel(a.modelName))
		}

		// Notify hooks before generation
		a.notifyBeforeGenerate(runCtx.Context(), messages)

		resp, err := runCtx.Deps().Provider.GenerateMessage(runCtx.Context(), messages, options...)

		// Notify hooks after generation
		a.notifyAfterGenerate(runCtx.Context(), ldomain.Response{Content: resp.Content}, err)

		if err != nil {
			return nil, fmt.Errorf("LLM generation failed: %w", err)
		}

		// Check for tool calls
		toolCalls, params, shouldCallTools := a.extractToolCalls(resp.Content)
		if !shouldCallTools {
			// No tool calls, return the response
			return resp.Content, nil
		}

		// Execute tool calls
		toolResults, err := a.executeToolCalls(runCtx, toolCalls, params)
		if err != nil {
			return nil, fmt.Errorf("tool execution failed: %w", err)
		}

		// Add assistant message and tool results to conversation
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleAssistant,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
		})

		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleUser,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: toolResults}},
		})
	}

	return "Agent reached maximum iterations without final result", nil
}

// Extract tool calls from LLM response
func (a *LLMAgent) extractToolCalls(content string) ([]string, []any, bool) {
	// Try multiple tool call extraction strategies

	// 1. Try OpenAI format
	if calls, params, found := a.extractOpenAIToolCalls(content); found {
		return calls, params, true
	}

	// 2. Try simple JSON format
	if call, param, found := a.extractSimpleToolCall(content); found {
		return []string{call}, []any{param}, true
	}

	// 3. Try XML-like format (for compatibility with test mocks)
	if calls, params, found := a.extractXMLToolCalls(content); found {
		return calls, params, true
	}

	return nil, nil, false
}

// Extract OpenAI format tool calls
func (a *LLMAgent) extractOpenAIToolCalls(content string) ([]string, []any, bool) {
	var openaiResp struct {
		ToolCalls []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	}

	// Try direct JSON parsing
	if err := json.Unmarshal([]byte(content), &openaiResp); err == nil && len(openaiResp.ToolCalls) > 0 {
		return a.processOpenAIToolCalls(openaiResp.ToolCalls)
	}

	// Try extracting from markdown blocks
	if strings.Contains(content, "```") {
		jsonBlocks := a.extractJSONBlocks(content)
		for _, block := range jsonBlocks {
			if err := json.Unmarshal([]byte(block), &openaiResp); err == nil && len(openaiResp.ToolCalls) > 0 {
				return a.processOpenAIToolCalls(openaiResp.ToolCalls)
			}
		}
	}

	return nil, nil, false
}

// Process OpenAI tool calls
func (a *LLMAgent) processOpenAIToolCalls(toolCalls []struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}) ([]string, []any, bool) {
	names := make([]string, 0, len(toolCalls))
	params := make([]any, 0, len(toolCalls))

	for _, call := range toolCalls {
		if call.Function.Name == "" {
			continue
		}

		names = append(names, call.Function.Name)

		// Parse arguments
		var parsedParams any
		if call.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(call.Function.Arguments), &parsedParams); err == nil {
				params = append(params, parsedParams)
			} else {
				params = append(params, call.Function.Arguments)
			}
		} else {
			params = append(params, map[string]any{})
		}
	}

	return names, params, len(names) > 0
}

// Extract simple tool call format
func (a *LLMAgent) extractSimpleToolCall(content string) (string, any, bool) {
	var toolCall struct {
		Tool   string `json:"tool"`
		Params any    `json:"params"`
	}

	// Try direct parsing
	if err := json.Unmarshal([]byte(content), &toolCall); err == nil && toolCall.Tool != "" {
		return toolCall.Tool, toolCall.Params, true
	}

	// Try extracting from markdown blocks
	if strings.Contains(content, "```") {
		jsonBlocks := a.extractJSONBlocks(content)
		for _, block := range jsonBlocks {
			if err := json.Unmarshal([]byte(block), &toolCall); err == nil && toolCall.Tool != "" {
				return toolCall.Tool, toolCall.Params, true
			}
		}
	}

	// Try finding JSON object in plain text
	if startIdx := strings.Index(content, `{"tool"`); startIdx != -1 {
		// Find the matching closing brace
		braceCount := 0
		inString := false
		escapeNext := false

		for i := startIdx; i < len(content); i++ {
			char := content[i]

			if escapeNext {
				escapeNext = false
				continue
			}

			if char == '\\' {
				escapeNext = true
				continue
			}

			if char == '"' && !escapeNext {
				inString = !inString
				continue
			}

			if !inString {
				if char == '{' {
					braceCount++
				} else if char == '}' {
					braceCount--
					if braceCount == 0 {
						// Found the complete JSON object
						jsonStr := content[startIdx : i+1]
						if err := json.Unmarshal([]byte(jsonStr), &toolCall); err == nil && toolCall.Tool != "" {
							return toolCall.Tool, toolCall.Params, true
						}
						break
					}
				}
			}
		}
	}

	return "", nil, false
}

// Extract JSON blocks from markdown
func (a *LLMAgent) extractJSONBlocks(content string) []string {
	if !strings.Contains(content, "```") {
		return nil
	}

	blocks := make([]string, 0)
	lines := strings.Split(content, "\n")
	inBlock := false
	var blockBuilder strings.Builder

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if !inBlock && strings.HasPrefix(trimmedLine, "```") {
			// Check if it might be JSON
			if strings.HasPrefix(trimmedLine, "```json") || trimmedLine == "```" {
				inBlock = true
				blockBuilder.Reset()
			}
			continue
		}

		if inBlock && trimmedLine == "```" {
			inBlock = false
			if blockBuilder.Len() > 0 {
				blocks = append(blocks, blockBuilder.String())
			}
			continue
		}

		if inBlock {
			if blockBuilder.Len() > 0 {
				blockBuilder.WriteByte('\n')
			}
			blockBuilder.WriteString(line)
		}
	}

	return blocks
}

// Extract XML-like tool calls (for test compatibility)
func (a *LLMAgent) extractXMLToolCalls(content string) ([]string, []any, bool) {
	// Look for <tool_calls> tags
	startTag := "<tool_calls>"
	endTag := "</tool_calls>"

	startIdx := strings.Index(content, startTag)
	if startIdx == -1 {
		return nil, nil, false
	}

	endIdx := strings.Index(content[startIdx:], endTag)
	if endIdx == -1 {
		return nil, nil, false
	}

	// Extract the content between tags
	toolCallsContent := content[startIdx+len(startTag) : startIdx+endIdx]
	toolCallsContent = strings.TrimSpace(toolCallsContent)

	// Parse as JSON array
	var toolCalls []struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal([]byte(toolCallsContent), &toolCalls); err != nil {
		return nil, nil, false
	}

	if len(toolCalls) == 0 {
		return nil, nil, false
	}

	names := make([]string, len(toolCalls))
	params := make([]any, len(toolCalls))

	for i, call := range toolCalls {
		names[i] = call.Name
		params[i] = call.Arguments
	}

	return names, params, true
}

// Execute tool calls
func (a *LLMAgent) executeToolCalls(runCtx *domain.RunContext[LLMDeps], toolNames []string, params []any) (string, error) {
	// executeToolCalls executes multiple tool calls and aggregates their results.
	// The function handles:
	// 1. Tool lookup and validation
	// 2. Context creation with state access
	// 3. Event emission and hook notifications
	// 4. Error aggregation without stopping execution
	// 5. Result formatting for LLM consumption

	if len(toolNames) == 0 {
		return "", nil
	}

	// Use string builder for efficient result concatenation
	var results strings.Builder
	results.WriteString("Tool results:\n")

	// Execute each tool call sequentially
	// Parallel execution could be added but requires careful state management
	for i, toolName := range toolNames {
		// Get corresponding parameters (may be nil if not provided)
		var toolParams any
		if i < len(params) {
			toolParams = params[i]
		}

		// Tool lookup with helpful error message
		tool, exists := a.GetTool(toolName)
		if !exists {
			// Provide available tools to help LLM correct itself
			results.WriteString(fmt.Sprintf("Error: Tool '%s' not found. Available tools: %s\n",
				toolName, strings.Join(a.ListTools(), ", ")))
			continue // Don't stop - try remaining tools
		}

		// Create rich execution context for the tool
		// This provides tools with access to:
		// - Agent state (read-only via StateReader)
		// - Parent context (for cancellation)
		// - Agent reference (for handoffs)
		// - Run metadata
		toolContext := domain.NewToolContext(
			runCtx.Context(),
			domain.NewStateReader(runCtx.State),
			a, // LLMAgent implements BaseAgent
			runCtx.RunID,
		)

		// Wire up event emission for observability
		if a.dispatcher != nil {
			eventEmitter := domain.NewToolEventEmitter(a.dispatcher, toolName, a.ID(), a.Name())
			toolContext = toolContext.WithEventEmitter(eventEmitter)
		}

		// Pass retry information for tools that implement retry logic
		if runCtx.Retry > 0 {
			toolContext = toolContext.WithRetry(runCtx.Retry)
		}

		// Notify hooks before tool call
		a.notifyBeforeToolCall(runCtx.Context(), toolName, toolParams)

		// Execute the tool with enhanced context
		result, err := tool.Execute(toolContext, toolParams)

		// Notify hooks after tool call
		a.notifyAfterToolCall(runCtx.Context(), toolName, result, err)

		if err != nil {
			results.WriteString(fmt.Sprintf("Tool '%s' error: %v\n", toolName, err))
			continue
		}

		// Format result for LLM consumption
		// Handle different result types appropriately
		var resultStr string
		switch v := result.(type) {
		case string:
			// String results pass through as-is
			resultStr = v
		case nil:
			// Nil indicates success with no output data
			resultStr = "Tool executed successfully with no output"
		default:
			// Complex types: try JSON first for structure preservation
			if jsonBytes, err := json.Marshal(result); err == nil {
				resultStr = string(jsonBytes)
			} else {
				// Fallback to string representation
				resultStr = fmt.Sprintf("%v", result)
			}
		}

		results.WriteString(fmt.Sprintf("Tool '%s' result: %s\n", toolName, resultStr))
	}

	return results.String(), nil
}

// Hook notification methods

// notifyBeforeGenerate calls all hooks' BeforeGenerate method
func (a *LLMAgent) notifyBeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	a.mu.RLock()
	hooks := a.hooks
	a.mu.RUnlock()

	for _, hook := range hooks {
		hook.BeforeGenerate(ctx, messages)
	}
}

// notifyAfterGenerate calls all hooks' AfterGenerate method
func (a *LLMAgent) notifyAfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	a.mu.RLock()
	hooks := a.hooks
	a.mu.RUnlock()

	for _, hook := range hooks {
		hook.AfterGenerate(ctx, response, err)
	}
}

// notifyBeforeToolCall calls all hooks' BeforeToolCall method
func (a *LLMAgent) notifyBeforeToolCall(ctx context.Context, tool string, params interface{}) {
	a.mu.RLock()
	hooks := a.hooks
	a.mu.RUnlock()

	// Convert params to map if possible
	var paramsMap map[string]interface{}

	switch p := params.(type) {
	case map[string]interface{}:
		paramsMap = p
	case nil:
		paramsMap = make(map[string]interface{})
	default:
		// Try to marshal/unmarshal to convert to map
		if jsonBytes, err := json.Marshal(params); err == nil {
			_ = json.Unmarshal(jsonBytes, &paramsMap)
		}
		// If that fails, wrap in a map
		if paramsMap == nil {
			paramsMap = map[string]interface{}{
				"value": params,
			}
		}
	}

	for _, hook := range hooks {
		hook.BeforeToolCall(ctx, tool, paramsMap)
	}
}

// notifyAfterToolCall calls all hooks' AfterToolCall method
func (a *LLMAgent) notifyAfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	a.mu.RLock()
	hooks := a.hooks
	a.mu.RUnlock()

	for _, hook := range hooks {
		hook.AfterToolCall(ctx, tool, result, err)
	}
}

// subAgentTool wraps a sub-agent as a tool
type subAgentTool struct {
	subAgent domain.BaseAgent
}

// Name returns the tool name (agent name)
func (sat *subAgentTool) Name() string {
	return sat.subAgent.Name()
}

// Description returns the tool description (agent description)
func (sat *subAgentTool) Description() string {
	return sat.subAgent.Description()
}

// Execute runs the sub-agent with the provided parameters
func (sat *subAgentTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	var state *domain.State

	// Check if we should use shared state
	if ctx.State != nil {
		// Create shared state context with parent state
		sharedCtx := domain.NewSharedStateContext(ctx.State)
		state = sharedCtx.LocalState()

		// TODO: When the sub-agent supports RunContext with shared state,
		// we should pass the shared context through
	} else {
		// No parent state, create fresh state
		state = domain.NewState()
	}

	// Handle different parameter types
	switch p := params.(type) {
	case map[string]interface{}:
		for k, v := range p {
			state.Set(k, v)
		}
	case string:
		state.Set("input", p)
	default:
		state.Set("input", params)
	}

	// Execute the sub-agent
	result, err := sat.subAgent.Run(ctx.Context, state)
	if err != nil {
		return nil, err
	}

	// Extract the result
	if output, ok := result.Get("output"); ok {
		return output, nil
	}
	if output, ok := result.Get("result"); ok {
		return output, nil
	}

	// Return the entire state if no specific output
	return result.Values(), nil
}

// ParameterSchema returns the schema for tool parameters
func (sat *subAgentTool) ParameterSchema() *sdomain.Schema {
	// For now, return a simple schema that accepts any object
	// In the future, we could introspect the agent to determine its expected inputs
	return &sdomain.Schema{
		Type:        "object",
		Description: fmt.Sprintf("Parameters for %s agent", sat.subAgent.Name()),
		Properties: map[string]sdomain.Property{
			"input": {
				Type:        "string",
				Description: "Input for the agent",
			},
		},
	}
}

// OutputSchema returns the schema for tool output
func (sat *subAgentTool) OutputSchema() *sdomain.Schema {
	return nil // Sub-agents can return various types
}

// UsageInstructions returns usage instructions
func (sat *subAgentTool) UsageInstructions() string {
	return fmt.Sprintf("Use this to delegate tasks to the %s agent. %s", sat.subAgent.Name(), sat.subAgent.Description())
}

// Examples returns usage examples
func (sat *subAgentTool) Examples() []domain.ToolExample {
	return nil // Sub-agents may have varied examples
}

// Constraints returns tool constraints
func (sat *subAgentTool) Constraints() []string {
	return nil
}

// ErrorGuidance returns error guidance
func (sat *subAgentTool) ErrorGuidance() map[string]string {
	return nil
}

// Category returns the tool category
func (sat *subAgentTool) Category() string {
	return "agent"
}

// Tags returns tool tags
func (sat *subAgentTool) Tags() []string {
	return []string{"agent", "delegation"}
}

// Version returns the tool version
func (sat *subAgentTool) Version() string {
	return "1.0.0"
}

// IsDeterministic returns whether the tool is deterministic
func (sat *subAgentTool) IsDeterministic() bool {
	return false // Sub-agents may not be deterministic
}

// IsDestructive returns whether the tool is destructive
func (sat *subAgentTool) IsDestructive() bool {
	return false // Delegating to agents is not destructive
}

// RequiresConfirmation returns whether the tool requires confirmation
func (sat *subAgentTool) RequiresConfirmation() bool {
	return false
}

// EstimatedLatency returns the estimated latency
func (sat *subAgentTool) EstimatedLatency() string {
	return "medium" // Sub-agents can vary in execution time
}

// ToMCPDefinition exports the tool definition in MCP format
func (sat *subAgentTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:        sat.Name(),
		Description: sat.Description(),
		InputSchema: sat.ParameterSchema(),
		Annotations: map[string]interface{}{
			"type":     "agent",
			"category": "agent",
			"version":  "1.0.0",
		},
	}
}

// transferToAgentTool is a built-in tool that allows dynamic delegation to sub-agents
type transferToAgentTool struct {
	agent *LLMAgent
}

// Name returns the tool name
func (t *transferToAgentTool) Name() string {
	return "transfer_to_agent"
}

// Description returns the tool description
func (t *transferToAgentTool) Description() string {
	return "Transfer control to a sub-agent by name. The sub-agent will handle the task and return the result."
}

// Execute transfers control to the specified sub-agent
func (t *transferToAgentTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	// Extract agent name and optional reason from params
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("transfer_to_agent requires a map with 'agent_name' and optional 'reason'")
	}

	agentName, ok := paramMap["agent_name"].(string)
	if !ok || agentName == "" {
		return nil, fmt.Errorf("transfer_to_agent requires 'agent_name' parameter")
	}

	// Find the sub-agent
	subAgent := t.agent.FindSubAgent(agentName)
	if subAgent == nil {
		// List available sub-agents for better error message
		var available []string
		for _, sub := range t.agent.SubAgents() {
			available = append(available, sub.Name())
		}
		return nil, fmt.Errorf("sub-agent '%s' not found. Available agents: %v", agentName, available)
	}

	// Create a simple handoff
	handoff := domain.NewSimpleHandoff("transfer", agentName)

	// Create a state for the handoff, using shared state if available
	var state *domain.State
	if ctx.State != nil {
		// Create shared state context with parent state
		sharedCtx := domain.NewSharedStateContext(ctx.State)
		state = sharedCtx.LocalState()

		// Copy relevant values from the context state
		for key, value := range ctx.State.Values() {
			state.Set(key, value)
		}
	} else {
		state = domain.NewState()
	}

	// Add any input parameters
	if input, ok := paramMap["input"]; ok {
		state.Set("input", input)
	}

	result, err := handoff.Execute(ctx.Context, state)
	if err != nil {
		return nil, fmt.Errorf("transfer to agent '%s' failed: %w", agentName, err)
	}

	// Extract the result
	if output, ok := result.Get("output"); ok {
		return output, nil
	}
	if output, ok := result.Get("result"); ok {
		return output, nil
	}

	// Return the state values if no specific output
	return result.Values(), nil
}

// ParameterSchema returns the schema for the transfer_to_agent tool
func (t *transferToAgentTool) ParameterSchema() *sdomain.Schema {
	return &sdomain.Schema{
		Type:        "object",
		Description: "Parameters for transferring control to a sub-agent",
		Properties: map[string]sdomain.Property{
			"agent_name": {
				Type:        "string",
				Description: "Name of the sub-agent to transfer control to",
			},
			"reason": {
				Type:        "string",
				Description: "Reason for transferring to this agent (optional)",
			},
			"input": {
				Type:        "string",
				Description: "Optional input to pass to the sub-agent",
			},
		},
		Required: []string{"agent_name"},
	}
}

// OutputSchema returns the schema for tool output
func (t *transferToAgentTool) OutputSchema() *sdomain.Schema {
	return nil // Transfer results can vary by agent
}

// UsageInstructions returns usage instructions
func (t *transferToAgentTool) UsageInstructions() string {
	return "Use this tool to transfer control to one of the available sub-agents. Each sub-agent specializes in different tasks."
}

// Examples returns usage examples
func (t *transferToAgentTool) Examples() []domain.ToolExample {
	var examples []domain.ToolExample
	if t.agent != nil && len(t.agent.SubAgents()) > 0 {
		firstAgent := t.agent.SubAgents()[0]
		examples = append(examples, domain.ToolExample{
			Name:        "Transfer to sub-agent",
			Description: fmt.Sprintf("Transfer control to the %s agent", firstAgent.Name()),
			Scenario:    fmt.Sprintf("When you need to %s", firstAgent.Description()),
			Input: map[string]interface{}{
				"agent_name": firstAgent.Name(),
				"input":      "Process this task",
			},
			Output:      "Result from sub-agent",
			Explanation: "The sub-agent will process the input and return its result",
		})
	}
	return examples
}

// Constraints returns tool constraints
func (t *transferToAgentTool) Constraints() []string {
	return []string{
		"Can only transfer to registered sub-agents",
		"Sub-agent must exist in the agent hierarchy",
	}
}

// ErrorGuidance returns error guidance
func (t *transferToAgentTool) ErrorGuidance() map[string]string {
	return map[string]string{
		"agent_not_found": "The specified agent name does not exist. Check available sub-agents.",
		"invalid_params":  "Parameters must include 'agent_name' as a string.",
	}
}

// Category returns the tool category
func (t *transferToAgentTool) Category() string {
	return "control"
}

// Tags returns tool tags
func (t *transferToAgentTool) Tags() []string {
	return []string{"control", "delegation", "transfer"}
}

// Version returns the tool version
func (t *transferToAgentTool) Version() string {
	return "1.0.0"
}

// IsDeterministic returns whether the tool is deterministic
func (t *transferToAgentTool) IsDeterministic() bool {
	return false // Sub-agents may not be deterministic
}

// IsDestructive returns whether the tool is destructive
func (t *transferToAgentTool) IsDestructive() bool {
	return false
}

// RequiresConfirmation returns whether the tool requires confirmation
func (t *transferToAgentTool) RequiresConfirmation() bool {
	return false
}

// EstimatedLatency returns the estimated latency
func (t *transferToAgentTool) EstimatedLatency() string {
	return "medium"
}

// ToMCPDefinition exports the tool definition in MCP format
func (t *transferToAgentTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.ParameterSchema(),
		Annotations: map[string]interface{}{
			"type":     "control",
			"category": "control",
			"version":  "1.0.0",
		},
	}
}
