// Package workflow provides agent workflow implementations.
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// MultiAgent extends DefaultAgent with optimizations for multi-provider scenarios
// It's designed to work efficiently with the MultiProvider
type MultiAgent struct {
	DefaultAgent

	// Optimization: cache for provider contexts to avoid recreating in each call
	providerContextCache sync.Map

	// Enhanced metrics for multi-provider operations
	multiProviderMetrics *MultiProviderMetrics
}

// MultiProviderMetrics tracks metrics specific to multi-provider operations
type MultiProviderMetrics struct {
	mu                sync.Mutex
	providerLatencies map[string][]int64 // Response times per provider
	consensusHits     int                // Times consensus algorithm produced a result
	consensusMisses   int                // Times consensus algorithm failed to produce a result
	fallbackCount     int                // Times fallback to single provider was used
}

// NewMultiAgent creates a new agent optimized for multi-provider scenarios
func NewMultiAgent(provider ldomain.Provider) *MultiAgent {
	baseAgent := NewAgent(provider)
	return &MultiAgent{
		DefaultAgent: *baseAgent,
		multiProviderMetrics: &MultiProviderMetrics{
			providerLatencies: make(map[string][]int64),
		},
	}
}

// Run executes the agent with given inputs
// This implementation adds optimizations for multi-provider scenarios
func (a *MultiAgent) Run(ctx context.Context, input string) (interface{}, error) {
	return a.run(ctx, input, nil)
}

// RunWithSchema executes the agent and validates output against a schema
// This implementation adds optimizations for multi-provider scenarios
func (a *MultiAgent) RunWithSchema(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	return a.run(ctx, input, schema)
}

// run is the internal implementation of Run and RunWithSchema
// This version includes optimizations for multi-provider scenarios
func (a *MultiAgent) run(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	// Prepare the prompt - same as DefaultAgent
	prompt := input
	if schema != nil {
		// Enhance the prompt with schema information
		enhancedPrompt, err := processor.EnhancePromptWithSchema(input, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to enhance prompt with schema: %w", err)
		}
		prompt = enhancedPrompt
	}

	// Create messages for the conversation - reuse the optimized version
	messages := a.createInitialMessages(prompt)

	// Detect if we're using a MultiProvider
	_, isMulti := a.llmProvider.(*provider.MultiProvider)

	// Pre-configure Provider with metadata when using MultiProvider
	if isMulti {
		// Store agent configuration in context to avoid recreating
		ctx = a.enhanceContextForMultiProvider(ctx)
	}

	// Main agent loop - continue until we have a result or error
	var finalResponse interface{}
	maxIterations := 10 // Prevent infinite loops
	for i := 0; i < maxIterations; i++ {
		// Call hooks before generate
		a.notifyBeforeGenerate(ctx, messages)

		// Generate response
		var resp ldomain.Response
		var genErr error
		if schema != nil {
			// If we have a schema, use structured generation
			var result interface{}
			result, genErr = a.llmProvider.GenerateWithSchema(ctx, prompt, schema)
			if genErr == nil {
				// Convert result to final format if needed
				finalResponse = result
				break // We have a valid structured result, exit the loop
			}
		} else {
			// Regular text generation
			var options []ldomain.Option
			if a.modelName != "" {
				options = append(options, ldomain.WithModel(a.modelName))
			}
			resp, genErr = a.llmProvider.GenerateMessage(ctx, messages, options...)
		}

		// Call hooks after generate
		a.notifyAfterGenerate(ctx, resp, genErr)

		if genErr != nil {
			return nil, fmt.Errorf("LLM generation failed: %w", genErr)
		}

		// If we're doing structured output, the response is in finalResponse
		if schema != nil && finalResponse != nil {
			return finalResponse, nil
		}

		// First check for multiple tool calls (OpenAI format)
		toolCalls, multiParams, shouldCallMultipleTools := a.ExtractMultipleToolCalls(resp.Content)

		if shouldCallMultipleTools && len(toolCalls) > 0 {
			// Process tool calls in parallel when there are multiple calls
			toolResponses, err := a.executeMultipleToolsParallel(ctx, toolCalls, multiParams)
			if err != nil {
				return nil, fmt.Errorf("error executing tools: %w", err)
			}

			// Add the assistant message and all tool results
			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleAssistant,
				Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
			})

			// Add tool results as user message
			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleUser,
				Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: toolResponses}},
			})

			continue
		}

		// Fall back to legacy single tool call extraction if multiple tools weren't found
		toolCall, params, shouldCallTool := a.ExtractToolCall(resp.Content)
		if !shouldCallTool {
			// No tool call, just return the response content
			return resp.Content, nil
		}

		// Find the requested tool
		tool, found := a.tools[toolCall]
		if !found {
			// Tool not found, append error message and continue
			errMsg := fmt.Sprintf("Tool '%s' not found. Available tools: %s",
				toolCall, strings.Join(a.getToolNames(), ", "))

			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleAssistant,
				Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
			})

			// Use user role instead of tool role for better OpenAI compatibility
			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleUser,
				Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: fmt.Sprintf("Tool error: %s", errMsg)}},
			})
			continue
		}

		// Call hooks before tool call
		a.notifyBeforeToolCall(ctx, toolCall, params)

		// Execute the tool
		toolResult, toolErr := tool.Execute(ctx, params)

		// Call hooks after tool call
		a.notifyAfterToolCall(ctx, toolCall, toolResult, toolErr)

		// Add the result to messages
		var toolRespContent string
		if toolErr != nil {
			toolRespContent = fmt.Sprintf("Error: %v", toolErr)
		} else {
			// Convert tool result to string if needed
			switch v := toolResult.(type) {
			case string:
				toolRespContent = v
			case nil:
				toolRespContent = "Tool executed successfully with no output"
			default:
				// Try to marshal to JSON
				jsonBytes, err := json.Marshal(toolResult)
				if err != nil {
					toolRespContent = fmt.Sprintf("%v", toolResult)
				} else {
					toolRespContent = string(jsonBytes)
				}
			}
		}

		// Add the assistant message and tool result to the conversation
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleAssistant,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
		})

		// Use user role instead of tool role for better OpenAI compatibility
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleUser,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: fmt.Sprintf("Tool '%s' result: %s", toolCall, toolRespContent)}},
		})
	}

	// If we have a schema and final response, return it
	if schema != nil && finalResponse != nil {
		return finalResponse, nil
	}

	// If we reached max iterations, return what we have
	return "Agent reached maximum iterations without final result", nil
}

// executeMultipleToolsParallel runs multiple tools in parallel and collects their results
func (a *MultiAgent) executeMultipleToolsParallel(ctx context.Context, toolNames []string, paramsArray []interface{}) (string, error) {
	if len(toolNames) == 0 {
		return "", fmt.Errorf("no tools to execute")
	}

	// Create a shared buffer for tool results
	var allToolsOutput strings.Builder
	allToolsOutput.WriteString("Tool results:\n")

	// If there's only one tool, handle it directly without goroutines
	if len(toolNames) == 1 {
		return a.executeSingleTool(ctx, toolNames[0], paramsArray[0], &allToolsOutput)
	}

	// Track results from multiple tools
	var wg sync.WaitGroup
	resultsMutex := sync.Mutex{}
	errorCount := 0
	successCount := 0

	// Process each tool in parallel
	for i, toolName := range toolNames {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()

			// Find the tool
			tool, found := a.tools[name]
			if !found {
				resultsMutex.Lock()
				allToolsOutput.WriteString(fmt.Sprintf("Error: Tool '%s' not found. Available tools: %s\n",
					name, strings.Join(a.getToolNames(), ", ")))
				errorCount++
				resultsMutex.Unlock()
				return
			}

			params := paramsArray[idx]

			// Call hooks before tool call
			a.notifyBeforeToolCall(ctx, name, params)

			// Execute the tool
			toolResult, toolErr := tool.Execute(ctx, params)

			// Call hooks after tool call
			a.notifyAfterToolCall(ctx, name, toolResult, toolErr)

			// Format the result
			var toolRespContent string
			if toolErr != nil {
				toolRespContent = fmt.Sprintf("Error: %v", toolErr)
				resultsMutex.Lock()
				errorCount++
				resultsMutex.Unlock()
			} else {
				resultsMutex.Lock()
				successCount++
				resultsMutex.Unlock()

				// Convert tool result to string
				switch v := toolResult.(type) {
				case string:
					toolRespContent = v
				case nil:
					toolRespContent = "Tool executed successfully with no output"
				default:
					jsonBytes, err := json.Marshal(toolResult)
					if err != nil {
						toolRespContent = fmt.Sprintf("%v", toolResult)
					} else {
						toolRespContent = string(jsonBytes)
					}
				}
			}

			// Add this tool's result to the combined output
			resultsMutex.Lock()
			allToolsOutput.WriteString(fmt.Sprintf("Tool '%s' result: %s\n\n", name, toolRespContent))
			resultsMutex.Unlock()
		}(i, toolName)
	}

	// Wait for all tools to complete
	wg.Wait()

	// If all tools failed, return an error
	if errorCount == len(toolNames) {
		return "", fmt.Errorf("all tools failed to execute")
	}

	return allToolsOutput.String(), nil
}

// executeSingleTool executes a single tool and formats its result
// Used when there's only one tool to avoid goroutine overhead
func (a *MultiAgent) executeSingleTool(ctx context.Context, toolName string, params interface{}, output *strings.Builder) (string, error) {
	// Find the tool
	tool, found := a.tools[toolName]
	if !found {
		output.WriteString(fmt.Sprintf("Error: Tool '%s' not found. Available tools: %s\n",
			toolName, strings.Join(a.getToolNames(), ", ")))
		return output.String(), nil
	}

	// Call hooks before tool call
	a.notifyBeforeToolCall(ctx, toolName, params)

	// Execute the tool
	toolResult, toolErr := tool.Execute(ctx, params)

	// Call hooks after tool call
	a.notifyAfterToolCall(ctx, toolName, toolResult, toolErr)

	// Format the result
	var toolRespContent string
	if toolErr != nil {
		toolRespContent = fmt.Sprintf("Error: %v", toolErr)
	} else {
		// Convert tool result to string
		switch v := toolResult.(type) {
		case string:
			toolRespContent = v
		case nil:
			toolRespContent = "Tool executed successfully with no output"
		default:
			jsonBytes, err := json.Marshal(toolResult)
			if err != nil {
				toolRespContent = fmt.Sprintf("%v", toolResult)
			} else {
				toolRespContent = string(jsonBytes)
			}
		}
	}

	// Add this tool's result to the combined output
	output.WriteString(fmt.Sprintf("Tool '%s' result: %s\n\n", toolName, toolRespContent))

	return output.String(), nil
}

// enhanceContextForMultiProvider adds agent configuration to the context
// This optimizes calls to MultiProvider by avoiding recreation of agent state
func (a *MultiAgent) enhanceContextForMultiProvider(ctx context.Context) context.Context {
	// Create a context key for provider-specific agent state
	type agentContextKey string
	const providerAgentKey agentContextKey = "agent_configuration"

	// Check if we already have an entry in the cache
	if cachedCtx, found := a.providerContextCache.Load(ctx); found {
		return cachedCtx.(context.Context)
	}

	// Create agent configuration to share with providers
	agentConfig := map[string]interface{}{
		"system_prompt":            a.systemPrompt,
		"model_name":               a.modelName,
		"tool_count":               len(a.tools),
		"cached_tool_descriptions": a.cachedToolsDescription,
	}

	// Add a tool schema cache to avoid regenerating JSON schemas
	toolSchemas := make(map[string]interface{}, len(a.tools))
	for name, tool := range a.tools {
		schema := tool.ParameterSchema()
		if schema != nil {
			toolSchemas[name] = schema
		}
	}
	agentConfig["tool_schemas"] = toolSchemas

	// Add the configuration to the context
	enhancedCtx := context.WithValue(ctx, providerAgentKey, agentConfig)

	// Store in cache for future use
	a.providerContextCache.Store(ctx, enhancedCtx)

	return enhancedCtx
}

// Removed unused updateMetrics function

// WithModel specifies which LLM model to use
// Override to reset context cache when model changes
func (a *MultiAgent) WithModel(modelName string) domain.Agent {
	// First call the parent implementation to set the model
	a.DefaultAgent.WithModel(modelName)

	// Reset the context cache since model has changed
	a.providerContextCache = sync.Map{}

	return a
}

// AddTool registers a tool with the agent
// Override to reset context cache when tools change
func (a *MultiAgent) AddTool(tool domain.Tool) domain.Agent {
	// First call the parent implementation to add the tool
	a.DefaultAgent.AddTool(tool)

	// Reset the context cache since tools have changed
	a.providerContextCache = sync.Map{}

	return a
}

// SetSystemPrompt configures the agent's system prompt
// Override to reset context cache when prompt changes
func (a *MultiAgent) SetSystemPrompt(prompt string) domain.Agent {
	// First call the parent implementation to set the prompt
	a.DefaultAgent.SetSystemPrompt(prompt)

	// Reset the context cache since system prompt has changed
	a.providerContextCache = sync.Map{}

	return a
}

// GetMultiProviderMetrics returns the multi-provider specific metrics
func (a *MultiAgent) GetMultiProviderMetrics() map[string]interface{} {
	a.multiProviderMetrics.mu.Lock()
	defer a.multiProviderMetrics.mu.Unlock()

	metrics := make(map[string]interface{})

	// Add consensus metrics
	metrics["consensus_hits"] = a.multiProviderMetrics.consensusHits
	metrics["consensus_misses"] = a.multiProviderMetrics.consensusMisses
	metrics["fallback_count"] = a.multiProviderMetrics.fallbackCount

	// Calculate average latency per provider
	providerAvgLatency := make(map[string]float64)
	for provider, latencies := range a.multiProviderMetrics.providerLatencies {
		if len(latencies) == 0 {
			continue
		}

		total := int64(0)
		for _, latency := range latencies {
			total += latency
		}
		providerAvgLatency[provider] = float64(total) / float64(len(latencies))
	}
	metrics["provider_latencies"] = providerAvgLatency

	return metrics
}
