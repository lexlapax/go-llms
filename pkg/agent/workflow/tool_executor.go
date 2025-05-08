// Package workflow provides agent workflow implementations.
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolName    string
	Result      interface{}
	Error       error
	ElapsedTime time.Duration
	Status      ToolExecutionStatus
}

// ToolExecutionStatus represents the status of a tool execution
type ToolExecutionStatus string

const (
	// ToolStatusSuccess indicates the tool executed successfully
	ToolStatusSuccess ToolExecutionStatus = "success"
	// ToolStatusError indicates the tool failed with an error
	ToolStatusError ToolExecutionStatus = "error"
	// ToolStatusNotFound indicates the requested tool was not found
	ToolStatusNotFound ToolExecutionStatus = "not_found"
	// ToolStatusTimeout indicates the tool execution timed out
	ToolStatusTimeout ToolExecutionStatus = "timeout"
)

// ToolExecutor provides parallel execution of tools with proper concurrency control
type ToolExecutor struct {
	tools            map[string]domain.Tool
	maxConcurrent    int
	executionTimeout time.Duration
	hooks            []domain.Hook

	// Semaphore to limit concurrent executions
	sem chan struct{}
	// For tracking active executions
	activeCount int
	mu          sync.RWMutex
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(tools map[string]domain.Tool, maxConcurrent int, timeout time.Duration, hooks []domain.Hook) *ToolExecutor {
	if maxConcurrent <= 0 {
		maxConcurrent = 5 // Default to 5 concurrent executions
	}

	if timeout <= 0 {
		timeout = 30 * time.Second // Default 30 second timeout
	}

	return &ToolExecutor{
		tools:            tools,
		maxConcurrent:    maxConcurrent,
		executionTimeout: timeout,
		hooks:            hooks,
		sem:              make(chan struct{}, maxConcurrent),
	}
}

// ExecuteToolsParallel executes multiple tools in parallel with proper concurrency control
// Returns a map of tool names to results
func (e *ToolExecutor) ExecuteToolsParallel(ctx context.Context, toolNames []string, paramsArray []interface{}) map[string]ToolResult {
	if len(toolNames) == 0 {
		return map[string]ToolResult{}
	}

	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.executionTimeout)
	defer cancel()

	// Create channels for results
	resultCh := make(chan ToolResult, len(toolNames))

	// Create a mutex to protect the ctxDone flag
	var ctxMu sync.Mutex
	ctxDone := false

	// Start a goroutine to watch for context cancellation
	go func() {
		<-execCtx.Done()
		ctxMu.Lock()
		ctxDone = true
		ctxMu.Unlock()
		// Release any blocked goroutines waiting for semaphore
		for i := 0; i < e.maxConcurrent; i++ {
			select {
			case e.sem <- struct{}{}:
			default:
			}
		}
	}()

	// Start goroutines for each tool
	var wg sync.WaitGroup
	for i, toolName := range toolNames {
		wg.Add(1)
		go func(idx int, name string, params interface{}) {
			defer wg.Done()

			// Check if context is already cancelled before acquiring semaphore
			ctxMu.Lock()
			isDone := ctxDone
			ctxMu.Unlock()

			if isDone {
				resultCh <- ToolResult{
					ToolName: name,
					Status:   ToolStatusTimeout,
					Error:    fmt.Errorf("operation cancelled due to timeout"),
				}
				return
			}

			// Acquire semaphore (or fail if context is done)
			select {
			case e.sem <- struct{}{}:
				// Got the semaphore, continue
				e.mu.Lock()
				e.activeCount++
				e.mu.Unlock()

				defer func() {
					<-e.sem // Release semaphore
					e.mu.Lock()
					e.activeCount--
					e.mu.Unlock()
				}()
			case <-execCtx.Done():
				resultCh <- ToolResult{
					ToolName: name,
					Status:   ToolStatusTimeout,
					Error:    fmt.Errorf("operation cancelled due to timeout"),
				}
				return
			}

			// Execute the tool with its own timeout
			result := e.executeTool(execCtx, name, params)
			resultCh <- result
		}(i, toolName, paramsArray[i])
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	results := make(map[string]ToolResult, len(toolNames))
	for result := range resultCh {
		results[result.ToolName] = result
	}

	return results
}

// executeTool executes a single tool and returns its result
func (e *ToolExecutor) executeTool(ctx context.Context, toolName string, params interface{}) ToolResult {
	// Check if tool exists
	tool, found := e.tools[toolName]
	if !found {
		return ToolResult{
			ToolName: toolName,
			Status:   ToolStatusNotFound,
			Error:    fmt.Errorf("tool not found: %s", toolName),
		}
	}

	// Convert params to map for hooks
	var paramsMap map[string]interface{}
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

	// Call hooks before tool call
	for _, hook := range e.hooks {
		hook.BeforeToolCall(ctx, toolName, paramsMap)
	}

	// Execute the tool with timeout
	startTime := time.Now()
	toolResult, toolErr := tool.Execute(ctx, params)
	elapsed := time.Since(startTime)

	// Create result
	result := ToolResult{
		ToolName:    toolName,
		Result:      toolResult,
		Error:       toolErr,
		ElapsedTime: elapsed,
	}

	if toolErr != nil {
		result.Status = ToolStatusError
	} else {
		result.Status = ToolStatusSuccess
	}

	// Call hooks after tool call
	for _, hook := range e.hooks {
		hook.AfterToolCall(ctx, toolName, toolResult, toolErr)
	}

	return result
}

// FormatToolResults formats the results of tool executions into a string
func (e *ToolExecutor) FormatToolResults(results map[string]ToolResult) string {
	var output strings.Builder
	output.WriteString("Tool results:\n")

	for _, result := range results {
		switch result.Status {
		case ToolStatusNotFound:
			output.WriteString(fmt.Sprintf("Error: Tool '%s' not found. Available tools: %s\n",
				result.ToolName, strings.Join(e.getToolNames(), ", ")))
		case ToolStatusTimeout:
			output.WriteString(fmt.Sprintf("Error: Tool '%s' execution timed out after %v\n",
				result.ToolName, e.executionTimeout))
		case ToolStatusError:
			output.WriteString(fmt.Sprintf("Tool '%s' error: %v\n\n", result.ToolName, result.Error))
		case ToolStatusSuccess:
			// Format the successful result
			var toolRespContent string
			switch v := result.Result.(type) {
			case string:
				toolRespContent = v
			case nil:
				toolRespContent = "Tool executed successfully with no output"
			default:
				jsonBytes, err := json.Marshal(result.Result)
				if err != nil {
					toolRespContent = fmt.Sprintf("%v", result.Result)
				} else {
					toolRespContent = string(jsonBytes)
				}
			}
			output.WriteString(fmt.Sprintf("Tool '%s' result: %s\n\n", result.ToolName, toolRespContent))
		}
	}

	return output.String()
}

// getToolNames returns a list of available tool names
func (e *ToolExecutor) getToolNames() []string {
	names := make([]string, 0, len(e.tools))
	for name := range e.tools {
		names = append(names, name)
	}
	return names
}

// GetActiveCount returns the number of currently executing tools
func (e *ToolExecutor) GetActiveCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.activeCount
}

// SetMaxConcurrent updates the maximum number of concurrent tool executions
func (e *ToolExecutor) SetMaxConcurrent(max int) {
	if max <= 0 {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// If increasing concurrency, create a new semaphore
	if max > e.maxConcurrent {
		newSem := make(chan struct{}, max)
		// Copy existing elements
		for i := 0; i < len(e.sem); i++ {
			newSem <- struct{}{}
		}
		e.sem = newSem
	}
	// If decreasing, we just set the value and let it adjust over time
	e.maxConcurrent = max
}

// SetExecutionTimeout updates the tool execution timeout
func (e *ToolExecutor) SetExecutionTimeout(timeout time.Duration) {
	if timeout <= 0 {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.executionTimeout = timeout
}
