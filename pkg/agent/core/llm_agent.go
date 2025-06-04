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
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// LLMDeps defines the dependencies for LLM agents
type LLMDeps struct {
	Provider ldomain.Provider
	Logger   *slog.Logger
	Tracer   any // trace.Tracer - keeping as interface to avoid OpenTelemetry dependency
}

// LLMAgent implements BaseAgent for LLM-powered agents
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

// NewLLMAgent creates a new LLM agent
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

	return agent
}

// NewAgent creates a simple LLM agent with minimal configuration (excellent DX)
func NewAgent(name string, provider ldomain.Provider) *LLMAgent {
	deps := LLMDeps{
		Provider: provider,
		Logger:   slog.Default(),
		Tracer:   nil, // Optional tracing
	}

	return NewLLMAgent(name, "LLM Agent", deps)
}

// NewAgentWithLogger creates an LLM agent with custom logger
func NewAgentWithLogger(name string, provider ldomain.Provider, logger *slog.Logger) *LLMAgent {
	deps := LLMDeps{
		Provider: provider,
		Logger:   logger,
		Tracer:   nil,
	}

	return NewLLMAgent(name, "LLM Agent", deps)
}

// NewAgentFromString creates an LLM agent from a provider/model string specification
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
	promptKeys := []string{"prompt", "input", "message", "query", "text"}

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
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return cached version if available
	if a.cachedToolsDescription != "" {
		return a.cachedToolsDescription
	}

	var content strings.Builder

	// Add system prompt
	if a.systemPrompt != "" {
		content.WriteString(a.systemPrompt)
	}

	// Add tool descriptions if tools exist
	if len(a.tools) > 0 {
		if content.Len() > 0 {
			content.WriteString("\n\n")
		}
		content.WriteString("Available tools:\n")

		for _, tool := range a.tools {
			content.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))
		}

		content.WriteString("\nTo use a tool, respond with JSON in this format:\n")
		content.WriteString(`{"tool": "tool_name", "params": {...}}`)
		content.WriteString("\nOr for OpenAI format:\n")
		content.WriteString(`{"tool_calls": [{"id": "call_1", "type": "function", "function": {"name": "tool_name", "arguments": "{...}"}}]}`)
	}

	// Cache the result
	a.cachedToolsDescription = content.String()
	return a.cachedToolsDescription
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

// Execute tool calls
func (a *LLMAgent) executeToolCalls(runCtx *domain.RunContext[LLMDeps], toolNames []string, params []any) (string, error) {
	if len(toolNames) == 0 {
		return "", nil
	}

	var results strings.Builder
	results.WriteString("Tool results:\n")

	for i, toolName := range toolNames {
		var toolParams any
		if i < len(params) {
			toolParams = params[i]
		}

		// Find the tool
		tool, exists := a.GetTool(toolName)
		if !exists {
			results.WriteString(fmt.Sprintf("Error: Tool '%s' not found. Available tools: %s\n",
				toolName, strings.Join(a.ListTools(), ", ")))
			continue
		}

		// Notify hooks before tool call
		a.notifyBeforeToolCall(runCtx.Context(), toolName, toolParams)

		// Execute the tool
		result, err := tool.Execute(runCtx.Context(), toolParams)

		// Notify hooks after tool call
		a.notifyAfterToolCall(runCtx.Context(), toolName, result, err)

		if err != nil {
			results.WriteString(fmt.Sprintf("Tool '%s' error: %v\n", toolName, err))
			continue
		}

		// Format result
		var resultStr string
		switch v := result.(type) {
		case string:
			resultStr = v
		case nil:
			resultStr = "Tool executed successfully with no output"
		default:
			if jsonBytes, err := json.Marshal(result); err == nil {
				resultStr = string(jsonBytes)
			} else {
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
