// Package workflow provides agent workflow implementations.
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// UnoptimizedDefaultAgent is the original implementation of DefaultAgent
// Keep this for benchmarking and compatibility testing
type UnoptimizedDefaultAgent struct {
	llmProvider  ldomain.Provider
	tools        map[string]domain.Tool
	hooks        []domain.Hook
	systemPrompt string
	modelName    string
}

// DefaultAgent implements the Agent interface with optimizations
type DefaultAgent struct {
	llmProvider  ldomain.Provider
	tools        map[string]domain.Tool
	hooks        []domain.Hook
	systemPrompt string
	modelName    string

	// Optimization: cache tool descriptions to avoid regeneration
	cachedToolsDescription string
	// Optimization: cache tool names to avoid regeneration
	cachedToolNames []string
	// Optimization: pre-allocate message buffer
	messageBuffer []ldomain.Message
}

// NewUnoptimizedAgent creates a new agent with an LLM provider using the unoptimized implementation
// This is kept for benchmarking and compatibility testing
func NewUnoptimizedAgent(provider ldomain.Provider) *UnoptimizedDefaultAgent {
	return &UnoptimizedDefaultAgent{
		llmProvider: provider,
		tools:       make(map[string]domain.Tool),
		hooks:       make([]domain.Hook, 0),
	}
}

// NewAgent creates a new agent with an LLM provider
func NewAgent(provider ldomain.Provider) *DefaultAgent {
	return &DefaultAgent{
		llmProvider: provider,
		tools:       make(map[string]domain.Tool),
		hooks:       make([]domain.Hook, 0),
		// Pre-allocate message buffer with capacity for efficiency
		messageBuffer: make([]ldomain.Message, 0, 10),
	}
}

// AddTool registers a tool with the agent
func (a *DefaultAgent) AddTool(tool domain.Tool) domain.Agent {
	a.tools[tool.Name()] = tool
	// Optimization: invalidate cached tool description and names when tools change
	a.cachedToolsDescription = ""
	a.cachedToolNames = nil
	return a
}

// SetSystemPrompt configures the agent's system prompt
func (a *DefaultAgent) SetSystemPrompt(prompt string) domain.Agent {
	a.systemPrompt = prompt
	// Optimization: invalidate cached tool description as it contains the system prompt
	a.cachedToolsDescription = ""
	return a
}

// WithModel specifies which LLM model to use
func (a *DefaultAgent) WithModel(modelName string) domain.Agent {
	a.modelName = modelName
	return a
}

// WithHook adds a monitoring hook to the agent
func (a *DefaultAgent) WithHook(hook domain.Hook) domain.Agent {
	a.hooks = append(a.hooks, hook)
	return a
}

// Unoptimized agent methods
// AddTool registers a tool with the unoptimized agent
func (a *UnoptimizedDefaultAgent) AddTool(tool domain.Tool) domain.Agent {
	a.tools[tool.Name()] = tool
	return a
}

// SetSystemPrompt configures the unoptimized agent's system prompt
func (a *UnoptimizedDefaultAgent) SetSystemPrompt(prompt string) domain.Agent {
	a.systemPrompt = prompt
	return a
}

// WithModel specifies which LLM model to use for the unoptimized agent
func (a *UnoptimizedDefaultAgent) WithModel(modelName string) domain.Agent {
	a.modelName = modelName
	return a
}

// WithHook adds a monitoring hook to the unoptimized agent
func (a *UnoptimizedDefaultAgent) WithHook(hook domain.Hook) domain.Agent {
	a.hooks = append(a.hooks, hook)
	return a
}

// Run executes the agent with given inputs
func (a *DefaultAgent) Run(ctx context.Context, input string) (interface{}, error) {
	return a.run(ctx, input, nil)
}

// RunWithSchema executes the agent and validates output against a schema
func (a *DefaultAgent) RunWithSchema(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	return a.run(ctx, input, schema)
}

// Run executes the unoptimized agent with given inputs
func (a *UnoptimizedDefaultAgent) Run(ctx context.Context, input string) (interface{}, error) {
	return a.run(ctx, input, nil)
}

// RunWithSchema executes the unoptimized agent and validates output against a schema
func (a *UnoptimizedDefaultAgent) RunWithSchema(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	return a.run(ctx, input, schema)
}

// run is the internal implementation of Run and RunWithSchema for the optimized agent
func (a *DefaultAgent) run(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	// Prepare the prompt
	prompt := input
	if schema != nil {
		// Enhance the prompt with schema information
		enhancedPrompt, err := processor.EnhancePromptWithSchema(input, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to enhance prompt with schema: %w", err)
		}
		prompt = enhancedPrompt
	}

	// Create messages for the conversation - optimized version
	messages := a.createInitialMessages(prompt)

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
			// Process each tool call and collect results
			var allToolsOutput strings.Builder
			allToolsOutput.WriteString("Tool results:\n")

			// Track if any tool calls were successful
			toolCallsMade := 0

			for i, toolName := range toolCalls {
				// Find the requested tool
				tool, found := a.tools[toolName]
				if !found {
					// Tool not found, append error message
					allToolsOutput.WriteString(fmt.Sprintf("Error: Tool '%s' not found. Available tools: %s\n",
						toolName, strings.Join(a.getToolNames(), ", ")))
					continue
				}

				toolCallsMade++
				params := multiParams[i]

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
				allToolsOutput.WriteString(fmt.Sprintf("Tool '%s' result: %s\n\n", toolName, toolRespContent))
			}

			// If we processed at least one tool, continue the conversation
			if toolCallsMade > 0 {
				// Add the assistant message and all tool results
				messages = append(messages, ldomain.Message{
					Role:    ldomain.RoleAssistant,
					Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
				})

				// Add tool results as user message for compatibility
				messages = append(messages, ldomain.Message{
					Role:    ldomain.RoleUser,
					Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: allToolsOutput.String()}},
				})

				continue
			}
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

// ExtractMultipleToolCalls extracts multiple tool calls from response content
// This is used for the OpenAI format where multiple tools can be called at once
// This method is exported for benchmarking purposes
func (a *DefaultAgent) ExtractMultipleToolCalls(content string) ([]string, []interface{}, bool) {
	// Optimization: Pre-allocate arrays with reasonable initial capacity
	toolNames := make([]string, 0, 3) // Most common use case is 1-3 tool calls
	paramsArray := make([]interface{}, 0, 3)

	// Parse as OpenAI format
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

	// Check if content is a direct OpenAI format JSON first
	if len(content) > 0 && (content[0] == '{' || content[0] == '[') {
		// Try to parse the content as OpenAI format
		if err := json.Unmarshal([]byte(content), &openaiResp); err == nil && len(openaiResp.ToolCalls) > 0 {
			// Process all tool calls in one go
			return a.processOpenAIToolCalls(openaiResp.ToolCalls, toolNames, paramsArray)
		}
	}

	// If direct parsing fails, look for JSON blocks in markdown
	// Optimization: only search for JSON blocks if content contains code block markers
	if strings.Contains(content, "```") {
		jsonBlocks := extractJSONBlocks(content)
		for _, block := range jsonBlocks {
			// Clear openaiResp to reuse it
			openaiResp.ToolCalls = openaiResp.ToolCalls[:0]

			if err := json.Unmarshal([]byte(block), &openaiResp); err == nil && len(openaiResp.ToolCalls) > 0 {
				return a.processOpenAIToolCalls(openaiResp.ToolCalls, toolNames, paramsArray)
			}
		}
	}

	return nil, nil, false
}

// processOpenAIToolCalls is a helper method to process tool calls in the OpenAI format
// This avoids code duplication in extractMultipleToolCalls
func (a *DefaultAgent) processOpenAIToolCalls(
	toolCalls []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	},
	toolNames []string,
	paramsArray []interface{},
) ([]string, []interface{}, bool) {
	// Estimate count of valid tools for pre-allocation
	validCount := 0
	for _, toolCall := range toolCalls {
		if toolCall.Function.Name != "" {
			validCount++
		}
	}

	// Pre-allocate result slices to exact size if empty
	if len(toolNames) == 0 && validCount > 0 {
		toolNames = make([]string, 0, validCount)
		paramsArray = make([]interface{}, 0, validCount)
	}

	// Process each tool call
	for _, toolCall := range toolCalls {
		if toolCall.Function.Name == "" {
			continue // Skip tool calls without names
		}

		// Parse the arguments JSON
		var params interface{}

		// Optimization: Only try to parse as JSON if the argument starts with '{' or '['
		args := toolCall.Function.Arguments
		if len(args) > 0 && (args[0] == '{' || args[0] == '[') {
			if err := json.Unmarshal([]byte(args), &params); err == nil {
				toolNames = append(toolNames, toolCall.Function.Name)
				paramsArray = append(paramsArray, params)
				continue
			}
		}

		// If JSON parsing fails or it's not JSON, use the raw arguments string
		toolNames = append(toolNames, toolCall.Function.Name)
		paramsArray = append(paramsArray, args)
	}

	if len(toolNames) > 0 {
		return toolNames, paramsArray, true
	}

	return nil, nil, false
}

// ExtractToolCall attempts to identify a tool call in the response content - optimized version
// This method is exported for benchmarking purposes
func (a *DefaultAgent) ExtractToolCall(content string) (string, interface{}, bool) {
	// Check for OpenAI format first (which includes tool_calls array)
	// Define the OpenAI format struct only once
	type functionStruct struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}

	type toolCallStruct struct {
		ID       string         `json:"id"`
		Type     string         `json:"type"`
		Function functionStruct `json:"function"`
	}

	// Fast path: check if content could be JSON
	if len(content) > 0 && (content[0] == '{' || content[0] == '[') {
		// Try OpenAI format first
		var openaiResp struct {
			ToolCalls []toolCallStruct `json:"tool_calls"`
		}

		// Try to parse full content as OpenAI format first
		if err := json.Unmarshal([]byte(content), &openaiResp); err == nil && len(openaiResp.ToolCalls) > 0 {
			// Use the first tool call
			toolCall := openaiResp.ToolCalls[0]
			if toolCall.Function.Name != "" {
				// Fast path: check if arguments could be JSON
				args := toolCall.Function.Arguments
				if len(args) > 0 && (args[0] == '{' || args[0] == '[') {
					// Parse the arguments JSON
					var params interface{}
					if err := json.Unmarshal([]byte(args), &params); err == nil {
						return toolCall.Function.Name, params, true
					}
				}
				// Even if JSON parsing fails, return the tool name and raw arguments
				return toolCall.Function.Name, toolCall.Function.Arguments, true
			}
		}

		// Try simple tool call format
		var toolCall struct {
			Tool   string      `json:"tool"`
			Params interface{} `json:"params"`
		}

		// Look for JSON embedded in the text - try the full text
		if json.Unmarshal([]byte(content), &toolCall) == nil && toolCall.Tool != "" {
			return toolCall.Tool, toolCall.Params, true
		}
	}

	// Only look for JSON blocks if content contains code block markers (optimization)
	if strings.Contains(content, "```") {
		// Look for JSON blocks in markdown
		jsonBlocks := extractJSONBlocks(content)
		if len(jsonBlocks) > 0 {
			// Try each JSON block
			for _, block := range jsonBlocks {
				// Skip empty blocks (optimization)
				if len(block) < 5 {
					continue
				}

				// Try simple tool call format first (most common)
				var toolCall struct {
					Tool   string      `json:"tool"`
					Params interface{} `json:"params"`
				}

				if json.Unmarshal([]byte(block), &toolCall) == nil && toolCall.Tool != "" {
					return toolCall.Tool, toolCall.Params, true
				}

				// Try OpenAI format within JSON blocks
				var blockOpenAIResp struct {
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				}

				if err := json.Unmarshal([]byte(block), &blockOpenAIResp); err == nil && len(blockOpenAIResp.ToolCalls) > 0 {
					// Use the first tool call
					toolCall := blockOpenAIResp.ToolCalls[0]
					if toolCall.Function.Name != "" {
						// Fast path: check if arguments could be JSON
						args := toolCall.Function.Arguments
						if len(args) > 0 && (args[0] == '{' || args[0] == '[') {
							// Parse the arguments JSON
							var params interface{}
							if err := json.Unmarshal([]byte(args), &params); err == nil {
								return toolCall.Function.Name, params, true
							}
						}
						// Even if JSON parsing fails, return the tool name and raw arguments
						return toolCall.Function.Name, toolCall.Function.Arguments, true
					}
				}
			}
		}
	}

	// Optimization: Only do text-based extraction if necessary (contains "tool:" or "params:")
	if strings.Contains(strings.ToLower(content), "tool:") ||
		strings.Contains(strings.ToLower(content), "params:") ||
		strings.Contains(strings.ToLower(content), "parameters:") {

		// Fallback to simple text-based extraction
		// Pre-allocate with estimated capacity for typical cases
		lines := strings.Split(content, "\n")
		var toolName string
		var paramsJSON strings.Builder
		inParams := false

		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Check for tool name
			lowerLine := strings.ToLower(trimmedLine)
			if strings.HasPrefix(lowerLine, "tool:") {
				toolName = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Tool:"))
				if toolName == "" { // Check for lowercase version if empty
					toolName = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "tool:"))
				}
			} else if strings.HasPrefix(lowerLine, "params:") || strings.HasPrefix(lowerLine, "parameters:") {
				// Extract parameter start
				var paramsPart string
				if strings.HasPrefix(lowerLine, "params:") {
					paramsPart = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Params:"))
					if paramsPart == "" { // Try lowercase
						paramsPart = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "params:"))
					}
				} else {
					paramsPart = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Parameters:"))
					if paramsPart == "" { // Try lowercase
						paramsPart = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "parameters:"))
					}
				}
				paramsJSON.WriteString(paramsPart)
				inParams = true
			} else if inParams {
				if paramsJSON.Len() > 0 {
					paramsJSON.WriteString(" ")
				}
				paramsJSON.WriteString(trimmedLine)
			}
		}

		// If we found a tool name and params JSON, try to parse it
		if toolName != "" && paramsJSON.Len() > 0 {
			paramsJSONStr := paramsJSON.String()
			// Fast path: check if arguments could be JSON
			if len(paramsJSONStr) > 0 && (paramsJSONStr[0] == '{' || paramsJSONStr[0] == '[') {
				var params interface{}
				if json.Unmarshal([]byte(paramsJSONStr), &params) == nil {
					return toolName, params, true
				}
			}
			// Even if JSON parsing fails, return the tool name and the raw params text
			return toolName, paramsJSONStr, true
		}
	}

	return "", nil, false
}

// extractJSONBlocks extracts JSON blocks from markdown-formatted text - optimized version
func extractJSONBlocks(content string) []string {
	// Fast path: if there's no code blocks, return empty
	if !strings.Contains(content, "```") {
		return nil
	}

	// Pre-allocate for expected number of blocks (3 is a reasonable default)
	blocks := make([]string, 0, 3)

	// Calculate the maximum capacity needed based on content size
	estimatedSize := len(content)

	// Pre-allocate string builder to reduce allocations
	var blockBuilder strings.Builder
	lines := strings.Split(content, "\n")
	inBlock := false
	jsonBlockMarker := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for block start - optimize the check
		if !inBlock {
			// Fast check for code block start
			if strings.HasPrefix(trimmedLine, "```") {
				// Check if it's a JSON block or potentially a JSON block
				if strings.HasPrefix(trimmedLine, "```json") {
					// Explicit JSON block
					inBlock = true
					jsonBlockMarker = true
					// Reset the string builder
					blockBuilder.Reset()
					blockBuilder.Grow(estimatedSize / 4) // Estimate block size
				} else if strings.HasPrefix(trimmedLine, "```") {
					// Check if it's not a known non-JSON format
					isNonJsonFormat := strings.HasPrefix(trimmedLine, "```yaml") ||
						strings.HasPrefix(trimmedLine, "```python") ||
						strings.HasPrefix(trimmedLine, "```go") ||
						strings.HasPrefix(trimmedLine, "```js") ||
						strings.HasPrefix(trimmedLine, "```java") ||
						strings.HasPrefix(trimmedLine, "```ruby") ||
						strings.HasPrefix(trimmedLine, "```c") ||
						strings.HasPrefix(trimmedLine, "```cpp") ||
						strings.HasPrefix(trimmedLine, "```csharp") ||
						strings.HasPrefix(trimmedLine, "```php") ||
						strings.HasPrefix(trimmedLine, "```rust") ||
						strings.HasPrefix(trimmedLine, "```shell") ||
						strings.HasPrefix(trimmedLine, "```bash") ||
						strings.HasPrefix(trimmedLine, "```sql") ||
						strings.HasPrefix(trimmedLine, "```typescript")

					if !isNonJsonFormat {
						inBlock = true
						jsonBlockMarker = false
						// Reset the string builder
						blockBuilder.Reset()
						blockBuilder.Grow(estimatedSize / 4) // Estimate block size
					}
				}
			}
			continue
		}

		// Check for block end
		if inBlock && trimmedLine == "```" {
			inBlock = false

			// Only add non-empty blocks
			if blockBuilder.Len() > 0 {
				blockContent := blockBuilder.String()

				// Only add if it's a JSON block or is valid JSON
				if jsonBlockMarker || isValidJSON(blockContent) {
					blocks = append(blocks, blockContent)
				}
			}
			continue
		}

		// Add line to current block
		if inBlock {
			// Add newline between lines if not the first line
			if blockBuilder.Len() > 0 {
				blockBuilder.WriteByte('\n')
			}
			blockBuilder.WriteString(line)
		}
	}

	return blocks
}

// isValidJSON checks if a string is valid JSON - optimized version
func isValidJSON(s string) bool {
	// Fast path: empty string or too short strings are not valid JSON
	if len(s) < 2 {
		return false
	}

	// Must start with { or [
	if s[0] != '{' && s[0] != '[' {
		return false
	}

	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// getToolNames returns a list of available tool names
func (a *DefaultAgent) getToolNames() []string {
	// Optimization: return cached names if available
	if a.cachedToolNames != nil {
		return a.cachedToolNames
	}

	// Pre-allocate slice with the right capacity
	names := make([]string, 0, len(a.tools))
	for name := range a.tools {
		names = append(names, name)
	}

	// Cache for future use
	a.cachedToolNames = names
	return names
}

// createInitialMessages creates the initial messages for the conversation - optimized version
func (a *DefaultAgent) createInitialMessages(input string) []ldomain.Message {
	// Reset message buffer to reuse it
	a.messageBuffer = a.messageBuffer[:0]

	// Create system message, combining prompt and tool descriptions
	var systemContent string
	if a.systemPrompt != "" {
		systemContent = a.systemPrompt
	}

	// Add available tools to system prompt if any
	if len(a.tools) > 0 {
		toolsDesc := a.getToolsDescription()
		if systemContent != "" {
			systemContent += "\n\n" + toolsDesc
		} else {
			systemContent = toolsDesc
		}
	}

	// Add the combined system message if non-empty
	if systemContent != "" {
		a.messageBuffer = append(a.messageBuffer, ldomain.Message{
			Role:    ldomain.RoleSystem,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: systemContent}},
		})
	}

	// Add user input
	a.messageBuffer = append(a.messageBuffer, ldomain.Message{
		Role:    ldomain.RoleUser,
		Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: input}},
	})

	return a.messageBuffer
}

// Unoptimized agent methods that match the original implementation

// ExtractMultipleToolCalls extracts multiple tool calls from response content - unoptimized version
// This method is exported for benchmarking purposes
func (a *UnoptimizedDefaultAgent) ExtractMultipleToolCalls(content string) ([]string, []interface{}, bool) {
	var toolNames []string
	var paramsArray []interface{}

	// Parse as OpenAI format
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

	// Try to parse the content as OpenAI format
	if err := json.Unmarshal([]byte(content), &openaiResp); err == nil && len(openaiResp.ToolCalls) > 0 {
		for _, toolCall := range openaiResp.ToolCalls {
			if toolCall.Function.Name != "" {
				// Parse the arguments JSON
				var params interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err == nil {
					toolNames = append(toolNames, toolCall.Function.Name)
					paramsArray = append(paramsArray, params)
				} else {
					// If JSON parsing fails, use the raw arguments string
					toolNames = append(toolNames, toolCall.Function.Name)
					paramsArray = append(paramsArray, toolCall.Function.Arguments)
				}
			}
		}

		if len(toolNames) > 0 {
			return toolNames, paramsArray, true
		}
	}

	// Look for JSON blocks in markdown that might contain OpenAI format
	jsonBlocks := extractJSONBlocks(content)
	for _, block := range jsonBlocks {
		if err := json.Unmarshal([]byte(block), &openaiResp); err == nil && len(openaiResp.ToolCalls) > 0 {
			for _, toolCall := range openaiResp.ToolCalls {
				if toolCall.Function.Name != "" {
					// Parse the arguments JSON
					var params interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err == nil {
						toolNames = append(toolNames, toolCall.Function.Name)
						paramsArray = append(paramsArray, params)
					} else {
						// If JSON parsing fails, use the raw arguments string
						toolNames = append(toolNames, toolCall.Function.Name)
						paramsArray = append(paramsArray, toolCall.Function.Arguments)
					}
				}
			}

			if len(toolNames) > 0 {
				return toolNames, paramsArray, true
			}
		}
	}

	return nil, nil, false
}

// ExtractToolCall attempts to identify a tool call in the response content - unoptimized version
// This method is exported for benchmarking purposes
func (a *UnoptimizedDefaultAgent) ExtractToolCall(content string) (string, interface{}, bool) {
	// Check for OpenAI format first (which includes tool_calls array)
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

	// Try to parse full content as OpenAI format first
	if err := json.Unmarshal([]byte(content), &openaiResp); err == nil && len(openaiResp.ToolCalls) > 0 {
		// Use the first tool call
		toolCall := openaiResp.ToolCalls[0]
		if toolCall.Function.Name != "" {
			// Parse the arguments JSON
			var params interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err == nil {
				return toolCall.Function.Name, params, true
			}
			// Even if JSON parsing fails, return the tool name and raw arguments
			return toolCall.Function.Name, toolCall.Function.Arguments, true
		}
	}

	// Continue with previous implementation
	// Simple parse function that looks for patterns like:
	// {"tool": "tool_name", "params": {...}}
	// or
	// Tool: tool_name
	// Params: {...}

	// Try JSON parsing first
	var toolCall struct {
		Tool   string      `json:"tool"`
		Params interface{} `json:"params"`
	}

	// Look for JSON embedded in the text - first try the full text
	if json.Unmarshal([]byte(content), &toolCall) == nil && toolCall.Tool != "" {
		return toolCall.Tool, toolCall.Params, true
	}

	// Look for JSON blocks in markdown
	jsonBlocks := extractJSONBlocks(content)
	for _, block := range jsonBlocks {
		if json.Unmarshal([]byte(block), &toolCall) == nil && toolCall.Tool != "" {
			return toolCall.Tool, toolCall.Params, true
		}

		// Try OpenAI format within JSON blocks
		var blockOpenAIResp struct {
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		}

		if err := json.Unmarshal([]byte(block), &blockOpenAIResp); err == nil && len(blockOpenAIResp.ToolCalls) > 0 {
			// Use the first tool call
			toolCall := blockOpenAIResp.ToolCalls[0]
			if toolCall.Function.Name != "" {
				// Parse the arguments JSON
				var params interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err == nil {
					return toolCall.Function.Name, params, true
				}
				// Even if JSON parsing fails, return the tool name and raw arguments
				return toolCall.Function.Name, toolCall.Function.Arguments, true
			}
		}
	}

	// Fallback to simple text-based extraction
	lines := strings.Split(content, "\n")
	var toolName string
	var paramsJSON string
	inParams := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "tool:") {
			toolName = strings.TrimSpace(strings.TrimPrefix(line, "Tool:"))
			toolName = strings.TrimSpace(strings.TrimPrefix(toolName, "tool:"))
		} else if strings.HasPrefix(strings.ToLower(line), "params:") ||
			strings.HasPrefix(strings.ToLower(line), "parameters:") {
			paramsPart := strings.TrimSpace(strings.TrimPrefix(line, "Params:"))
			paramsPart = strings.TrimSpace(strings.TrimPrefix(paramsPart, "params:"))
			paramsPart = strings.TrimSpace(strings.TrimPrefix(paramsPart, "Parameters:"))
			paramsPart = strings.TrimSpace(strings.TrimPrefix(paramsPart, "parameters:"))
			paramsJSON = paramsPart
			inParams = true
		} else if inParams {
			paramsJSON += " " + line
		}
	}

	// If we found a tool name and params JSON, try to parse it
	if toolName != "" && paramsJSON != "" {
		var params interface{}
		if json.Unmarshal([]byte(paramsJSON), &params) == nil {
			return toolName, params, true
		}
		// Even if JSON parsing fails, return the tool name and the raw params text
		return toolName, paramsJSON, true
	}

	return "", nil, false
}

// run is the internal implementation of Run and RunWithSchema for the unoptimized agent
func (a *UnoptimizedDefaultAgent) run(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	// Prepare the prompt
	prompt := input
	if schema != nil {
		// Enhance the prompt with schema information
		enhancedPrompt, err := processor.EnhancePromptWithSchema(input, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to enhance prompt with schema: %w", err)
		}
		prompt = enhancedPrompt
	}

	// Create messages for the conversation - unoptimized version
	messages := a.createInitialMessages(prompt)

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
			// Process each tool call and collect results
			var allToolsOutput strings.Builder
			allToolsOutput.WriteString("Tool results:\n")

			// Track if any tool calls were successful
			toolCallsMade := 0

			for i, toolName := range toolCalls {
				// Find the requested tool
				tool, found := a.tools[toolName]
				if !found {
					// Tool not found, append error message
					allToolsOutput.WriteString(fmt.Sprintf("Error: Tool '%s' not found. Available tools: %s\n",
						toolName, strings.Join(a.getToolNames(), ", ")))
					continue
				}

				toolCallsMade++
				params := multiParams[i]

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
				allToolsOutput.WriteString(fmt.Sprintf("Tool '%s' result: %s\n\n", toolName, toolRespContent))
			}

			// If we processed at least one tool, continue the conversation
			if toolCallsMade > 0 {
				// Add the assistant message and all tool results
				messages = append(messages, ldomain.Message{
					Role:    ldomain.RoleAssistant,
					Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
				})

				// Add tool results as user message for compatibility
				messages = append(messages, ldomain.Message{
					Role:    ldomain.RoleUser,
					Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: allToolsOutput.String()}},
				})

				continue
			}
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

// getToolNames returns a list of available tool names - unoptimized version
func (a *UnoptimizedDefaultAgent) getToolNames() []string {
	var names []string
	for name := range a.tools {
		names = append(names, name)
	}
	return names
}

// createInitialMessages creates the initial messages for the conversation - unoptimized version
func (a *UnoptimizedDefaultAgent) createInitialMessages(input string) []ldomain.Message {
	var messages []ldomain.Message

	// Create system message, combining prompt and tool descriptions
	var systemContent string
	if a.systemPrompt != "" {
		systemContent = a.systemPrompt
	}

	// Add available tools to system prompt if any
	if len(a.tools) > 0 {
		toolsDesc := a.getToolsDescription()
		if systemContent != "" {
			systemContent += "\n\n" + toolsDesc
		} else {
			systemContent = toolsDesc
		}
	}

	// Add the combined system message if non-empty
	if systemContent != "" {
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleSystem,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: systemContent}},
		})
	}

	// Add user input
	messages = append(messages, ldomain.Message{
		Role:    ldomain.RoleUser,
		Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: input}},
	})

	return messages
}



// Notification functions for hooks

// notifyBeforeGenerate calls all hooks' BeforeGenerate method
func (a *DefaultAgent) notifyBeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	for _, hook := range a.hooks {
		hook.BeforeGenerate(ctx, messages)
	}
}

// notifyAfterGenerate calls all hooks' AfterGenerate method
func (a *DefaultAgent) notifyAfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	for _, hook := range a.hooks {
		hook.AfterGenerate(ctx, response, err)
	}
}

// notifyBeforeToolCall calls all hooks' BeforeToolCall method
func (a *DefaultAgent) notifyBeforeToolCall(ctx context.Context, tool string, params interface{}) {
	// Convert params to map if possible
	var paramsMap map[string]interface{}

	// If it's already a map, use it
	switch p := params.(type) {
	case map[string]interface{}:
		paramsMap = p
	default:
		// Try to convert to map using reflection
		v := reflect.ValueOf(params)
		if v.Kind() == reflect.Struct {
			paramsMap = make(map[string]interface{})
			t := v.Type()
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				paramsMap[field.Name] = v.Field(i).Interface()
			}
		} else {
			// Create a map with a single "value" key
			paramsMap = map[string]interface{}{
				"value": params,
			}
		}
	}

	for _, hook := range a.hooks {
		hook.BeforeToolCall(ctx, tool, paramsMap)
	}
}

// notifyAfterToolCall calls all hooks' AfterToolCall method
func (a *DefaultAgent) notifyAfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	for _, hook := range a.hooks {
		hook.AfterToolCall(ctx, tool, result, err)
	}
}

// Notification functions for the unoptimized agent
// These mirror the optimized agent's notification methods

// notifyBeforeGenerate calls all hooks' BeforeGenerate method
func (a *UnoptimizedDefaultAgent) notifyBeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	for _, hook := range a.hooks {
		hook.BeforeGenerate(ctx, messages)
	}
}

// notifyAfterGenerate calls all hooks' AfterGenerate method
func (a *UnoptimizedDefaultAgent) notifyAfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	for _, hook := range a.hooks {
		hook.AfterGenerate(ctx, response, err)
	}
}

// notifyBeforeToolCall calls all hooks' BeforeToolCall method
func (a *UnoptimizedDefaultAgent) notifyBeforeToolCall(ctx context.Context, tool string, params interface{}) {
	// Convert params to map if possible
	var paramsMap map[string]interface{}

	// If it's already a map, use it
	switch p := params.(type) {
	case map[string]interface{}:
		paramsMap = p
	default:
		// Try to convert to map using reflection
		v := reflect.ValueOf(params)
		if v.Kind() == reflect.Struct {
			paramsMap = make(map[string]interface{})
			t := v.Type()
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				paramsMap[field.Name] = v.Field(i).Interface()
			}
		} else {
			// Create a map with a single "value" key
			paramsMap = map[string]interface{}{
				"value": params,
			}
		}
	}

	for _, hook := range a.hooks {
		hook.BeforeToolCall(ctx, tool, paramsMap)
	}
}

// notifyAfterToolCall calls all hooks' AfterToolCall method
func (a *UnoptimizedDefaultAgent) notifyAfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	for _, hook := range a.hooks {
		hook.AfterToolCall(ctx, tool, result, err)
	}
}
