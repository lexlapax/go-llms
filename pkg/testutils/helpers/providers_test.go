package helpers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

func TestResponseGenerator(t *testing.T) {
	t.Run("nil schema", func(t *testing.T) {
		rg := NewResponseGenerator(nil)
		resp, err := rg.Generate()
		assert.NoError(t, err)
		assert.Equal(t, "Generated response", resp)
	})

	t.Run("object schema", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
		}

		rg := NewResponseGenerator(schema)
		resp, err := rg.Generate()
		assert.NoError(t, err)
		assert.Contains(t, resp, "name")
		assert.Contains(t, resp, "age")

		// Should be valid JSON
		assert.True(t, strings.HasPrefix(resp, "{"))
		assert.True(t, strings.HasSuffix(resp, "}"))
	})

	t.Run("array schema", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "array",
		}

		rg := NewResponseGenerator(schema)
		resp, err := rg.Generate()
		assert.NoError(t, err)
		assert.Equal(t, `["item1","item2","item3"]`, resp)
	})

	t.Run("string schema", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "string",
		}

		rg := NewResponseGenerator(schema)
		resp, err := rg.Generate()
		assert.NoError(t, err)
		assert.Equal(t, "generated string value", resp)
	})

	t.Run("number schema", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "number",
		}

		rg := NewResponseGenerator(schema)
		resp, err := rg.Generate()
		assert.NoError(t, err)
		assert.Equal(t, "42", resp)
	})

	t.Run("boolean schema", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "boolean",
		}

		rg := NewResponseGenerator(schema)
		resp, err := rg.Generate()
		assert.NoError(t, err)
		assert.Equal(t, "true", resp)
	})
}

func TestErrorInjector(t *testing.T) {
	t.Run("default behavior", func(t *testing.T) {
		ei := NewErrorInjector("network")

		// Should always error by default (rate = 1.0)
		assert.True(t, ei.ShouldError())
		assert.True(t, ei.ShouldError())

		err := ei.GetError()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network")
	})

	t.Run("with rate", func(t *testing.T) {
		ei := NewErrorInjector("auth").WithRate(0.5)

		// With rate 0.5, should error every other call
		errorCount := 0
		for i := 0; i < 10; i++ {
			if ei.ShouldError() {
				errorCount++
			}
		}

		// Should be approximately half
		assert.InDelta(t, 5, errorCount, 2)
	})

	t.Run("with error after", func(t *testing.T) {
		ei := NewErrorInjector("rate_limit").WithErrorAfter(3)

		// First 3 calls should not error
		assert.False(t, ei.ShouldError())
		assert.False(t, ei.ShouldError())
		assert.False(t, ei.ShouldError())

		// Subsequent calls should error
		assert.True(t, ei.ShouldError())
		assert.True(t, ei.ShouldError())
	})

	t.Run("with message", func(t *testing.T) {
		ei := NewErrorInjector("custom").WithMessage("specific error")

		err := ei.GetError()
		assert.Error(t, err)
		assert.Equal(t, "custom: specific error", err.Error())
	})

	t.Run("error types", func(t *testing.T) {
		testCases := []struct {
			errorType string
			expected  string
		}{
			{"rate_limit", "rate limit exceeded"},
			{"auth", "authentication failed"},
			{"network", "network error: connection refused"},
			{"timeout", "request timeout"},
			{"invalid_response", "invalid response format"},
			{"unknown", "provider error: unknown"},
		}

		for _, tc := range testCases {
			t.Run(tc.errorType, func(t *testing.T) {
				ei := NewErrorInjector(tc.errorType)
				err := ei.GetError()
				assert.Contains(t, err.Error(), tc.expected)
			})
		}
	})
}

func TestStreamSimulator(t *testing.T) {
	t.Run("basic streaming", func(t *testing.T) {
		content := "Hello, this is a test stream"
		ss := NewStreamSimulator(content)

		var received []string
		err := ss.Stream(context.Background(), func(chunk string) error {
			received = append(received, chunk)
			return nil
		})

		assert.NoError(t, err)
		assert.Greater(t, len(received), 1) // Should be multiple chunks
		assert.Equal(t, content, strings.Join(received, ""))
	})

	t.Run("with delay", func(t *testing.T) {
		content := "Short"
		ss := NewStreamSimulator(content).WithDelay(50 * time.Millisecond)

		start := time.Now()
		chunks := 0
		err := ss.Stream(context.Background(), func(chunk string) error {
			chunks++
			return nil
		})

		duration := time.Since(start)
		assert.NoError(t, err)
		assert.Equal(t, 1, chunks)                     // "Short" fits in one chunk
		assert.Less(t, duration, 100*time.Millisecond) // No delay after last chunk
	})

	t.Run("with error at chunk", func(t *testing.T) {
		content := "This will error partway through"
		ss := NewStreamSimulator(content).WithErrorAt(1)

		chunks := 0
		err := ss.Stream(context.Background(), func(chunk string) error {
			chunks++
			return nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream error at chunk 1")
		assert.Equal(t, 1, chunks) // Should have received first chunk
	})

	t.Run("context cancellation", func(t *testing.T) {
		content := "This is a longer content that will be streamed"
		ss := NewStreamSimulator(content).WithDelay(100 * time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())

		chunks := 0
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := ss.Stream(ctx, func(chunk string) error {
			chunks++
			return nil
		})

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Greater(t, chunks, 0)
		assert.Less(t, chunks, 5) // Should not have completed
	})

	t.Run("callback error", func(t *testing.T) {
		content := "Test content"
		ss := NewStreamSimulator(content)

		callbackErr := errors.New("callback failed")
		chunks := 0
		err := ss.Stream(context.Background(), func(chunk string) error {
			chunks++
			if chunks == 2 {
				return callbackErr
			}
			return nil
		})

		assert.Equal(t, callbackErr, err)
		assert.Equal(t, 2, chunks)
	})
}

func TestProviderBehavior(t *testing.T) {
	t.Run("apply error injector", func(t *testing.T) {
		provider := mocks.NewMockProvider("test")
		behavior := &ProviderBehavior{
			ErrorInjector: NewErrorInjector("network"),
		}

		behavior.ApplyToProvider(provider)

		// The behavior should have set OnGenerateMessage
		assert.NotNil(t, provider.OnGenerateMessage)

		// Test that it returns an error
		resp, err := provider.GenerateMessage(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network")
		assert.Empty(t, resp.Content)
	})

	t.Run("apply metadata function", func(t *testing.T) {
		provider := mocks.NewMockProvider("test")
		behavior := &ProviderBehavior{
			MetadataFunc: func() map[string]interface{} {
				return map[string]interface{}{
					"test": "metadata",
				}
			},
		}

		behavior.ApplyToProvider(provider)

		// The behavior should have set OnGenerateMessage
		assert.NotNil(t, provider.OnGenerateMessage)

		// Test that it returns response with metadata
		resp, err := provider.GenerateMessage(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, "Response with metadata", resp.Content)
	})
}

func TestCommonBehaviors(t *testing.T) {
	t.Run("slow provider", func(t *testing.T) {
		behavior := SlowProviderBehavior(100 * time.Millisecond)
		assert.Equal(t, 100*time.Millisecond, behavior.ResponseDelay)
		assert.Equal(t, 10*time.Millisecond, behavior.StreamingDelay)
	})

	t.Run("unreliable provider", func(t *testing.T) {
		behavior := UnreliableProviderBehavior(0.3)
		assert.NotNil(t, behavior.ErrorInjector)
		assert.Equal(t, "network", behavior.ErrorInjector.errorType)
		assert.Equal(t, 0.3, behavior.ErrorInjector.errorRate)
	})

	t.Run("rate limited provider", func(t *testing.T) {
		behavior := RateLimitedProviderBehavior(5)
		assert.NotNil(t, behavior.ErrorInjector)
		assert.Equal(t, "rate_limit", behavior.ErrorInjector.errorType)
		assert.Equal(t, 5, behavior.ErrorInjector.errorAfter)
	})

	t.Run("metadata provider", func(t *testing.T) {
		behavior := MetadataProviderBehavior()
		assert.NotNil(t, behavior.MetadataFunc)

		metadata := behavior.MetadataFunc()
		assert.Equal(t, "test-model", metadata["model"])
		assert.Equal(t, 0.7, metadata["temperature"])
		assert.Equal(t, 1000, metadata["max_tokens"])

		usage, ok := metadata["usage"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 50, usage["prompt_tokens"])
		assert.Equal(t, 100, usage["completion_tokens"])
		assert.Equal(t, 150, usage["total_tokens"])
	})
}

func TestGeneratePropertyValue(t *testing.T) {
	t.Run("string property", func(t *testing.T) {
		prop := sdomain.Property{Type: "string"}
		val := generatePropertyValue(prop)
		assert.Equal(t, "test value", val)
	})

	t.Run("string with enum", func(t *testing.T) {
		prop := sdomain.Property{
			Type: "string",
			Enum: []string{"option1", "option2"},
		}
		val := generatePropertyValue(prop)
		assert.Equal(t, "option1", val)
	})

	t.Run("number property", func(t *testing.T) {
		prop := sdomain.Property{Type: "number"}
		val := generatePropertyValue(prop)
		assert.Equal(t, 42, val)
	})

	t.Run("boolean property", func(t *testing.T) {
		prop := sdomain.Property{Type: "boolean"}
		val := generatePropertyValue(prop)
		assert.Equal(t, true, val)
	})

	t.Run("array property", func(t *testing.T) {
		prop := sdomain.Property{Type: "array"}
		val := generatePropertyValue(prop)
		arr, ok := val.([]interface{})
		require.True(t, ok)
		assert.Len(t, arr, 2)
		assert.Equal(t, "item1", arr[0])
		assert.Equal(t, "item2", arr[1])
	})

	t.Run("object property", func(t *testing.T) {
		prop := sdomain.Property{Type: "object"}
		val := generatePropertyValue(prop)
		obj, ok := val.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "value", obj["nested"])
	})

	t.Run("unknown property", func(t *testing.T) {
		prop := sdomain.Property{Type: "unknown"}
		val := generatePropertyValue(prop)
		assert.Equal(t, "default value", val)
	})
}
