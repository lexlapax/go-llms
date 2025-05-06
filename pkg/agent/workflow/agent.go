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

// DefaultAgent implements the Agent interface
type DefaultAgent struct {
	llmProvider  ldomain.Provider
	tools        map[string]domain.Tool
	hooks        []domain.Hook
	systemPrompt string
	modelName    string
}

// NewAgent creates a new agent with an LLM provider
func NewAgent(provider ldomain.Provider) *DefaultAgent {
	return &DefaultAgent{
		llmProvider: provider,
		tools:       make(map[string]domain.Tool),
		hooks:       make([]domain.Hook, 0),
	}
}

// AddTool registers a tool with the agent
func (a *DefaultAgent) AddTool(tool domain.Tool) domain.Agent {
	a.tools[tool.Name()] = tool
	return a
}

// SetSystemPrompt configures the agent's system prompt
func (a *DefaultAgent) SetSystemPrompt(prompt string) domain.Agent {
	a.systemPrompt = prompt
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

// Run executes the agent with given inputs
func (a *DefaultAgent) Run(ctx context.Context, input string) (interface{}, error) {
	return a.run(ctx, input, nil)
}

// RunWithSchema executes the agent and validates output against a schema
func (a *DefaultAgent) RunWithSchema(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	return a.run(ctx, input, schema)
}

// run is the internal implementation of Run and RunWithSchema
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

	// Create messages for the conversation
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

		// Check for tool calls
		toolCall, params, shouldCallTool := a.extractToolCall(resp.Content)
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
				Content: resp.Content,
			})
			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleTool,
				Content: errMsg,
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
			Content: resp.Content,
		})
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleTool,
			Content: toolRespContent,
		})
	}

	// If we have a schema and final response, return it
	if schema != nil && finalResponse != nil {
		return finalResponse, nil
	}

	// If we reached max iterations, return what we have
	return "Agent reached maximum iterations without final result", nil
}

// extractToolCall attempts to identify a tool call in the response content
func (a *DefaultAgent) extractToolCall(content string) (string, interface{}, bool) {
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

// extractJSONBlocks extracts JSON blocks from markdown-formatted text
func extractJSONBlocks(content string) []string {
	var blocks []string
	lines := strings.Split(content, "\n")
	var currentBlock []string
	inBlock := false
	jsonBlockMarker := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for block start
		if !inBlock && (strings.HasPrefix(trimmedLine, "```json") ||
			(strings.HasPrefix(trimmedLine, "```") && !strings.HasPrefix(trimmedLine, "```yaml") &&
				!strings.HasPrefix(trimmedLine, "```python") && !strings.HasPrefix(trimmedLine, "```go") &&
				!strings.HasPrefix(trimmedLine, "```js") && !strings.HasPrefix(trimmedLine, "```java"))) {

			inBlock = true
			jsonBlockMarker = strings.HasPrefix(trimmedLine, "```json")
			currentBlock = []string{}
			continue
		}

		// Check for block end
		if inBlock && trimmedLine == "```" {
			inBlock = false
			if len(currentBlock) > 0 {
				// Try to validate if the block contains JSON
				joined := strings.Join(currentBlock, "\n")
				if jsonBlockMarker || isValidJSON(joined) {
					blocks = append(blocks, joined)
				}
			}
			continue
		}

		// Add line to current block
		if inBlock {
			currentBlock = append(currentBlock, line)
		}
	}

	return blocks
}

// isValidJSON checks if a string is valid JSON
func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// getToolNames returns a list of available tool names
func (a *DefaultAgent) getToolNames() []string {
	var names []string
	for name := range a.tools {
		names = append(names, name)
	}
	return names
}

// createInitialMessages creates the initial messages for the conversation
func (a *DefaultAgent) createInitialMessages(input string) []ldomain.Message {
	var messages []ldomain.Message

	// Add system message if provided
	if a.systemPrompt != "" {
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleSystem,
			Content: a.systemPrompt,
		})
	}

	// Add available tools to system prompt if any
	if len(a.tools) > 0 {
		toolsDesc := a.getToolsDescription()
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleSystem,
			Content: toolsDesc,
		})
	}

	// Add user input
	messages = append(messages, ldomain.Message{
		Role:    ldomain.RoleUser,
		Content: input,
	})

	return messages
}

// getToolsDescription creates a description of available tools
func (a *DefaultAgent) getToolsDescription() string {
	if len(a.tools) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("You have access to the following tools:\n\n")

	for name, tool := range a.tools {
		builder.WriteString(fmt.Sprintf("Tool: %s\n", name))
		builder.WriteString(fmt.Sprintf("Description: %s\n", tool.Description()))

		// Add parameter schema if available
		schema := tool.ParameterSchema()
		if schema != nil {
			paramSchemaJSON, err := json.MarshalIndent(schema, "", "  ")
			if err == nil {
				builder.WriteString(fmt.Sprintf("Parameters: %s\n", string(paramSchemaJSON)))
			}
		}

		builder.WriteString("\n")
	}

	builder.WriteString("\nTo use a tool, respond with:\n")
	builder.WriteString("```json\n{\"tool\": \"tool_name\", \"params\": {...}}\n```\n")

	return builder.String()
}

// Notification functions for hooks

func (a *DefaultAgent) notifyBeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	for _, hook := range a.hooks {
		hook.BeforeGenerate(ctx, messages)
	}
}

func (a *DefaultAgent) notifyAfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	for _, hook := range a.hooks {
		hook.AfterGenerate(ctx, response, err)
	}
}

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

func (a *DefaultAgent) notifyAfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	for _, hook := range a.hooks {
		hook.AfterToolCall(ctx, tool, result, err)
	}
}
