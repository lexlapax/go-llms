// ABOUTME: Enhanced mock tool implementations with call tracking and response mapping
// ABOUTME: Provides configurable tools with success/failure modes and comprehensive testing support

package mocks

import (
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ToolCall represents a recorded tool execution
type ToolCall struct {
	Input     map[string]interface{}
	Output    map[string]interface{}
	Error     error
	Context   *domain.ToolContext
	Timestamp time.Time
	Duration  time.Duration
}

// ExpectedCall represents an expected tool call for verification
type ExpectedCall struct {
	InputMatcher func(input map[string]interface{}) bool
	Description  string
	MinCalls     int
	MaxCalls     int
	ActualCalls  int
}

// MockTool is an enhanced mock implementation of the Tool interface
type MockTool struct {
	// Configuration
	ToolName        string
	ToolDescription string
	ToolCategory    string
	ToolTags        []string
	ToolVersion     string
	ResponseMap     map[string]interface{} // Input pattern to response mapping
	DefaultOutput   interface{}            // Default output when no pattern matches
	ErrorRate       float64                // Probability of returning an error (0.0-1.0)

	// Behavior hooks
	OnExecute  func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error)
	OnValidate func(input map[string]interface{}) error

	// Backward compatibility: simple executor that takes interface{} params
	Executor func(ctx *domain.ToolContext, params interface{}) (interface{}, error)

	// Schema configuration
	Schema      *sdomain.Schema // Backward compatibility alias for ParamSchema
	ParamSchema *sdomain.Schema
	OutSchema   *sdomain.Schema

	// Metadata
	UsageInstr      string
	ToolExamples    []domain.ToolExample
	ToolConstraints []string
	ErrorGuid       map[string]string

	// Assertions
	ExpectedCalls []ExpectedCall

	// State
	mu             sync.RWMutex
	callHistory    []ToolCall
	executionCount int
}

// NewMockTool creates a new mock tool with default configuration
func NewMockTool(name, description string) *MockTool {
	return &MockTool{
		ToolName:        name,
		ToolDescription: description,
		ToolCategory:    "test",
		ToolTags:        []string{"test", "mock"},
		ToolVersion:     "1.0.0",
		ResponseMap:     make(map[string]interface{}),
		DefaultOutput:   map[string]interface{}{"result": "mock result"},
		ErrorGuid:       make(map[string]string),
		callHistory:     make([]ToolCall, 0),
	}
}

// WithResponseMapping adds a response mapping for specific inputs
func (t *MockTool) WithResponseMapping(inputPattern string, output interface{}) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ResponseMap[inputPattern] = output
	return t
}

// WithErrorRate sets the probability of returning an error
func (t *MockTool) WithErrorRate(rate float64) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ErrorRate = rate
	return t
}

// WithParameterSchema sets the parameter schema
func (t *MockTool) WithParameterSchema(schema *sdomain.Schema) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ParamSchema = schema
	return t
}

// WithOutputSchema sets the output schema
func (t *MockTool) WithOutputSchema(schema *sdomain.Schema) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.OutSchema = schema
	return t
}

// WithCategory sets the tool category
func (t *MockTool) WithCategory(category string) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ToolCategory = category
	return t
}

// WithTags sets the tool tags
func (t *MockTool) WithTags(tags ...string) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ToolTags = tags
	return t
}

// WithVersion sets the tool version
func (t *MockTool) WithVersion(version string) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ToolVersion = version
	return t
}

// WithUsageInstructions sets usage instructions
func (t *MockTool) WithUsageInstructions(instructions string) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.UsageInstr = instructions
	return t
}

// WithExamples sets tool examples
func (t *MockTool) WithExamples(examples ...domain.ToolExample) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ToolExamples = examples
	return t
}

// WithConstraints sets tool constraints
func (t *MockTool) WithConstraints(constraints ...string) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ToolConstraints = constraints
	return t
}

// WithErrorGuidance sets error guidance
func (t *MockTool) WithErrorGuidance(guidance map[string]string) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ErrorGuid = guidance
	return t
}

// WithExecutor sets a custom executor function
func (t *MockTool) WithExecutor(executor func(ctx *domain.ToolContext, params interface{}) (interface{}, error)) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Executor = executor
	return t
}

// WithOnExecute sets a custom execute handler
func (t *MockTool) WithOnExecute(handler func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error)) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.OnExecute = handler
	return t
}

// Name returns the tool name
func (t *MockTool) Name() string {
	return t.ToolName
}

// Description returns the tool description
func (t *MockTool) Description() string {
	return t.ToolDescription
}

// Execute runs the tool with the given parameters
func (t *MockTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	start := time.Now()

	// Check backward-compatible executor first
	if t.Executor != nil {
		result, err := t.Executor(ctx, params)
		// Record call with params as-is
		inputMap := map[string]interface{}{"params": params}
		outputMap := map[string]interface{}{"result": result}
		t.recordCall(inputMap, outputMap, err, ctx, start)
		return result, err
	}

	// Convert params to map for easier handling
	inputMap, ok := params.(map[string]interface{})
	if !ok {
		// Try to convert
		inputMap = map[string]interface{}{"input": params}
	}

	// Check behavior hook
	if t.OnExecute != nil {
		output, err := t.OnExecute(ctx, inputMap)
		t.recordCall(inputMap, output, err, ctx, start)
		return output, err
	}

	// Simulate error based on error rate
	if t.shouldSimulateError() {
		err := fmt.Errorf("simulated error for tool %s", t.Name())
		t.recordCall(inputMap, nil, err, ctx, start)
		return nil, err
	}

	// Find matching response
	output := t.findMatchingResponse(inputMap)

	// Convert output to map if needed
	outputMap, ok := output.(map[string]interface{})
	if !ok {
		outputMap = map[string]interface{}{"result": output}
	}

	t.recordCall(inputMap, outputMap, nil, ctx, start)
	return outputMap, nil
}

// ParameterSchema returns the parameter schema
func (t *MockTool) ParameterSchema() *sdomain.Schema {
	if t.ParamSchema != nil {
		return t.ParamSchema
	}
	// Backward compatibility: check Schema field
	return t.Schema
}

// OutputSchema returns the output schema
func (t *MockTool) OutputSchema() *sdomain.Schema {
	return t.OutSchema
}

// UsageInstructions returns usage instructions
func (t *MockTool) UsageInstructions() string {
	return t.UsageInstr
}

// Examples returns tool examples
func (t *MockTool) Examples() []domain.ToolExample {
	return t.ToolExamples
}

// Constraints returns tool constraints
func (t *MockTool) Constraints() []string {
	return t.ToolConstraints
}

// ErrorGuidance returns error guidance
func (t *MockTool) ErrorGuidance() map[string]string {
	return t.ErrorGuid
}

// Category returns the tool category
func (t *MockTool) Category() string {
	return t.ToolCategory
}

// Tags returns the tool tags
func (t *MockTool) Tags() []string {
	return t.ToolTags
}

// Version returns the tool version
func (t *MockTool) Version() string {
	return t.ToolVersion
}

// IsDeterministic returns whether the tool is deterministic
func (t *MockTool) IsDeterministic() bool {
	return true // Mock tools are deterministic by default
}

// IsDestructive returns whether the tool is destructive
func (t *MockTool) IsDestructive() bool {
	return false // Mock tools are not destructive
}

// RequiresConfirmation returns whether the tool requires confirmation
func (t *MockTool) RequiresConfirmation() bool {
	return false // Mock tools don't require confirmation
}

// EstimatedLatency returns the estimated latency
func (t *MockTool) EstimatedLatency() string {
	return "fast" // Mock tools are fast
}

// ToMCPDefinition converts to MCP tool definition
func (t *MockTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:        t.ToolName,
		Description: t.ToolDescription,
		InputSchema: t.ParamSchema,
	}
}

// GetCallHistory returns the call history
func (t *MockTool) GetCallHistory() []ToolCall {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]ToolCall, len(t.callHistory))
	copy(history, t.callHistory)
	return history
}

// GetExecutionCount returns the total number of executions
func (t *MockTool) GetExecutionCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.executionCount
}

// Reset clears the call history and resets counters
func (t *MockTool) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.callHistory = make([]ToolCall, 0)
	t.executionCount = 0

	// Reset expected call counters
	for i := range t.ExpectedCalls {
		t.ExpectedCalls[i].ActualCalls = 0
	}
}

// AssertExecuted verifies that the tool was executed with specific inputs
func (t *MockTool) AssertExecuted(inputMatcher func(input map[string]interface{}) bool) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, call := range t.callHistory {
		if inputMatcher(call.Input) {
			return true
		}
	}

	return false
}

// ExpectCall adds an expected call for verification
func (t *MockTool) ExpectCall(description string, inputMatcher func(map[string]interface{}) bool, minCalls, maxCalls int) *MockTool {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ExpectedCalls = append(t.ExpectedCalls, ExpectedCall{
		InputMatcher: inputMatcher,
		Description:  description,
		MinCalls:     minCalls,
		MaxCalls:     maxCalls,
		ActualCalls:  0,
	})

	return t
}

// VerifyExpectations checks if all expected calls were made
func (t *MockTool) VerifyExpectations() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, expected := range t.ExpectedCalls {
		actualCalls := 0
		for _, call := range t.callHistory {
			if expected.InputMatcher(call.Input) {
				actualCalls++
			}
		}

		if actualCalls < expected.MinCalls {
			return fmt.Errorf("expected at least %d calls for %s, got %d", expected.MinCalls, expected.Description, actualCalls)
		}

		if expected.MaxCalls > 0 && actualCalls > expected.MaxCalls {
			return fmt.Errorf("expected at most %d calls for %s, got %d", expected.MaxCalls, expected.Description, actualCalls)
		}
	}

	return nil
}

// Private methods

func (t *MockTool) findMatchingResponse(input map[string]interface{}) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Try to find a matching pattern based on input
	inputStr := fmt.Sprintf("%v", input)
	for pattern, response := range t.ResponseMap {
		if pattern == inputStr {
			return response
		}
	}

	// Check for specific field matches
	for pattern, response := range t.ResponseMap {
		if val, ok := input[pattern]; ok && fmt.Sprintf("%v", val) != "" {
			return response
		}
	}

	// Return default output
	return t.DefaultOutput
}

func (t *MockTool) shouldSimulateError() bool {
	if t.ErrorRate <= 0 {
		return false
	}
	if t.ErrorRate >= 1 {
		return true
	}

	// Simple error simulation based on execution count
	return (t.executionCount+1)%int(1/t.ErrorRate) == 0
}

func (t *MockTool) recordCall(input, output map[string]interface{}, err error, ctx *domain.ToolContext, start time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	call := ToolCall{
		Input:     input,
		Output:    output,
		Error:     err,
		Context:   ctx,
		Timestamp: start,
		Duration:  time.Since(start),
	}

	t.callHistory = append(t.callHistory, call)
	t.executionCount++

	// Update expected call counters
	for i := range t.ExpectedCalls {
		if t.ExpectedCalls[i].InputMatcher(input) {
			t.ExpectedCalls[i].ActualCalls++
		}
	}
}

// Helper functions for creating common mock tools

// CreateCalculatorTool creates a mock calculator tool
func CreateCalculatorTool() *MockTool {
	tool := NewMockTool("calculator", "Perform mathematical calculations")

	tool.WithParameterSchema(&sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"expression": {
				Type:        "string",
				Description: "The mathematical expression to evaluate",
			},
		},
		Required: []string{"expression"},
	})

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		expr, ok := input["expression"].(string)
		if !ok {
			return nil, fmt.Errorf("expression must be a string")
		}

		// Simple mock calculations
		switch expr {
		case "2+2":
			return map[string]interface{}{"result": 4}, nil
		case "10/2":
			return map[string]interface{}{"result": 5}, nil
		default:
			return map[string]interface{}{"result": 42}, nil // Default answer
		}
	}

	return tool
}

// CreateWebSearchTool creates a mock web search tool
func CreateWebSearchTool() *MockTool {
	tool := NewMockTool("web_search", "Search the web for information")

	tool.WithParameterSchema(&sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"query": {
				Type:        "string",
				Description: "The search query",
			},
			"max_results": {
				Type:        "integer",
				Description: "Maximum number of results to return",
			},
		},
		Required: []string{"query"},
	})

	tool.WithResponseMapping("golang testing", map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"title":   "Testing in Go",
				"url":     "https://golang.org/doc/tutorial/add-a-test",
				"snippet": "Learn how to write tests in Go",
			},
			{
				"title":   "Go Testing Package",
				"url":     "https://pkg.go.dev/testing",
				"snippet": "Package testing provides support for automated testing",
			},
		},
	})

	return tool
}

// CreateFileTool creates a mock file operation tool
func CreateFileTool() *MockTool {
	tool := NewMockTool("file_ops", "Perform file operations")

	tool.WithParameterSchema(&sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"operation": {
				Type:        "string",
				Description: "The operation to perform (read, write, delete)",
				Enum:        []string{"read", "write", "delete"},
			},
			"path": {
				Type:        "string",
				Description: "The file path",
			},
			"content": {
				Type:        "string",
				Description: "Content for write operations",
			},
		},
		Required: []string{"operation", "path"},
	})

	// Virtual file system
	files := map[string]string{
		"/tmp/test.txt":  "Hello, World!",
		"/tmp/data.json": `{"name": "test", "value": 123}`,
	}

	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		op := input["operation"].(string)
		path := input["path"].(string)

		switch op {
		case "read":
			if content, exists := files[path]; exists {
				return map[string]interface{}{"content": content}, nil
			}
			return nil, fmt.Errorf("file not found: %s", path)
		case "write":
			content := input["content"].(string)
			files[path] = content
			return map[string]interface{}{"success": true}, nil
		case "delete":
			delete(files, path)
			return map[string]interface{}{"success": true}, nil
		default:
			return nil, fmt.Errorf("unknown operation: %s", op)
		}
	}

	return tool
}
