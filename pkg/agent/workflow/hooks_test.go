package workflow

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestLoggingHook(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Create hooks with different log levels
	basicHook := NewLoggingHook(logger, LogLevelBasic)
	detailedHook := NewLoggingHook(logger, LogLevelDetailed)
	debugHook := NewLoggingHook(logger, LogLevelDebug)

	// Create test messages and response
	messages := []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleSystem, "You are a helpful assistant."),
		ldomain.NewTextMessage(ldomain.RoleUser, "What is the capital of France?"),
	}
	response := ldomain.Response{Content: "The capital of France is Paris."}

	// Test BeforeGenerate with different log levels
	t.Run("BeforeGenerate", func(t *testing.T) {
		ctx := context.Background()

		buf.Reset()
		basicHook.BeforeGenerate(ctx, messages)
		if !contains(buf.String(), "Generating response") {
			t.Errorf("Basic log should include generating message, got: %s", buf.String())
		}

		buf.Reset()
		detailedHook.BeforeGenerate(ctx, messages)
		if !contains(buf.String(), "Message count") || !contains(buf.String(), "2") {
			t.Errorf("Detailed log should include message count, got: %s", buf.String())
		}

		buf.Reset()
		debugHook.BeforeGenerate(ctx, messages)
		if !contains(buf.String(), "Message details") || !contains(buf.String(), "system") {
			t.Errorf("Debug log should include message details, got: %s", buf.String())
		}
	})

	// Test AfterGenerate with different log levels
	t.Run("AfterGenerate", func(t *testing.T) {
		ctx := context.Background()

		buf.Reset()
		basicHook.AfterGenerate(ctx, response, nil)
		if !contains(buf.String(), "Response generated") {
			t.Errorf("Basic log should include response generated, got: %s", buf.String())
		}

		// Test error case
		buf.Reset()
		err := fmt.Errorf("test error")
		basicHook.AfterGenerate(ctx, response, err)
		if !contains(buf.String(), "Generation failed") || !contains(buf.String(), "test error") {
			t.Errorf("Error log should include error message, got: %s", buf.String())
		}
	})

	// Test tool calls
	t.Run("ToolCalls", func(t *testing.T) {
		ctx := context.Background()
		params := map[string]interface{}{
			"query": "weather in Paris",
		}
		result := map[string]interface{}{
			"temperature": 22,
			"condition":   "sunny",
		}

		buf.Reset()
		detailedHook.BeforeToolCall(ctx, "weather", params)
		if !contains(buf.String(), "Calling tool") || !contains(buf.String(), "weather") {
			t.Errorf("Tool call log should include tool name, got: %s", buf.String())
		}

		buf.Reset()
		detailedHook.AfterToolCall(ctx, "weather", result, nil)
		if !contains(buf.String(), "executed successfully") {
			t.Errorf("Tool result log should indicate success, got: %s", buf.String())
		}

		// Test error case
		buf.Reset()
		toolErr := fmt.Errorf("tool failed")
		detailedHook.AfterToolCall(ctx, "weather", nil, toolErr)
		if !contains(buf.String(), "Tool call failed") || !contains(buf.String(), "tool failed") {
			t.Errorf("Tool error log should include error message, got: %s", buf.String())
		}
	})
}

func TestMetricsHook(t *testing.T) {
	hook := NewMetricsHook()
	ctx := WithMetrics(context.Background())

	// Test BeforeGenerate and AfterGenerate
	t.Run("GenerateMetrics", func(t *testing.T) {
		messages := []ldomain.Message{
			ldomain.NewTextMessage(ldomain.RoleSystem, "You are a helpful assistant."),
			ldomain.NewTextMessage(ldomain.RoleUser, "What is the capital of France?"),
		}
		response := ldomain.Response{Content: "The capital of France is Paris."}

		hook.BeforeGenerate(ctx, messages)
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		hook.AfterGenerate(ctx, response, nil)

		metrics := hook.GetMetrics()

		if metrics.Requests != 1 {
			t.Errorf("Expected 1 request, got %d", metrics.Requests)
		}

		if metrics.TotalTokens <= 0 {
			t.Errorf("Expected positive token count, got %d", metrics.TotalTokens)
		}

		if metrics.AverageGenTimeMs < 10 {
			t.Errorf("Expected generation time >= 10ms, got %.2f", metrics.AverageGenTimeMs)
		}
	})

	// Test BeforeToolCall and AfterToolCall
	t.Run("ToolMetrics", func(t *testing.T) {
		params := map[string]interface{}{
			"query": "weather in Paris",
		}
		result := map[string]interface{}{
			"temperature": 22,
			"condition":   "sunny",
		}

		hook.BeforeToolCall(ctx, "weather", params)
		// Simulate tool execution time
		time.Sleep(15 * time.Millisecond)
		hook.AfterToolCall(ctx, "weather", result, nil)

		metrics := hook.GetMetrics()

		if metrics.ToolCalls != 1 {
			t.Errorf("Expected 1 tool call, got %d", metrics.ToolCalls)
		}

		if toolStats, exists := metrics.ToolStats["weather"]; !exists {
			t.Error("Expected weather tool stats to exist")
		} else {
			if toolStats.Calls != 1 {
				t.Errorf("Expected 1 weather tool call, got %d", toolStats.Calls)
			}

			if toolStats.AverageTimeMs < 15 {
				t.Errorf("Expected tool execution time >= 15ms, got %.2f", toolStats.AverageTimeMs)
			}
		}
	})

	// Test error metrics
	t.Run("ErrorMetrics", func(t *testing.T) {
		err := fmt.Errorf("test error")

		hook.BeforeGenerate(ctx, []ldomain.Message{})
		hook.AfterGenerate(ctx, ldomain.Response{}, err)

		hook.BeforeToolCall(ctx, "search", nil)
		hook.AfterToolCall(ctx, "search", nil, err)

		metrics := hook.GetMetrics()

		if metrics.ErrorCount != 2 {
			t.Errorf("Expected 2 errors, got %d", metrics.ErrorCount)
		}
	})

	// Test reset
	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		metrics := hook.GetMetrics()

		if metrics.Requests != 0 || metrics.ToolCalls != 0 || metrics.ErrorCount != 0 {
			t.Errorf("Reset should zero all counts, got requests=%d, toolCalls=%d, errors=%d",
				metrics.Requests, metrics.ToolCalls, metrics.ErrorCount)
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
