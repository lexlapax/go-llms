package fixtures

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestChatGPTMockProvider(t *testing.T) {
	provider := ChatGPTMockProvider()
	assert.NotNil(t, provider)
	assert.Equal(t, "chatgpt-mock", provider.ProviderName)

	ctx := context.Background()

	// Test hello pattern
	resp, err := provider.GenerateMessage(ctx, []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "hello world"),
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "Hello! How can I assist you today?")

	// Test explain pattern
	resp, err = provider.GenerateMessage(ctx, []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "explain quantum physics"),
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "I'd be happy to explain")

	// Test code pattern
	resp, err = provider.GenerateMessage(ctx, []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "write some code"),
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "Here's a code example")
	assert.Contains(t, resp.Content, "```python")

	// Test default response
	resp, err = provider.GenerateMessage(ctx, []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "random message"),
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "I understand your request")
}

func TestClaudeMockProvider(t *testing.T) {
	provider := ClaudeMockProvider()
	assert.NotNil(t, provider)
	assert.Equal(t, "claude-mock", provider.ProviderName)

	ctx := context.Background()

	// Test hello pattern
	resp, err := provider.GenerateMessage(ctx, []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "hello"),
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "Hello! I'm Claude")

	// Test analyze pattern
	resp, err = provider.GenerateMessage(ctx, []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "analyze this data"),
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "I'll analyze this carefully")

	// Test explain pattern
	resp, err = provider.GenerateMessage(ctx, []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "please explain"),
	})
	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "I'll explain this concept thoroughly")
}

func TestErrorMockProvider(t *testing.T) {
	testCases := []struct {
		errorType   string
		expectedMsg string
	}{
		{"rate_limit", "Rate limit exceeded"},
		{"auth", "Authentication failed"},
		{"network", "Network error"},
		{"invalid_response", "Invalid response format"},
		{"unknown", "An unknown error occurred"},
	}

	ctx := context.Background()
	messages := []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "test"),
	}

	for _, tc := range testCases {
		t.Run(tc.errorType, func(t *testing.T) {
			provider := ErrorMockProvider(tc.errorType)
			_, err := provider.GenerateMessage(ctx, messages)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
		})
	}
}

func TestSlowMockProvider(t *testing.T) {
	delay := 100 * time.Millisecond
	provider := SlowMockProvider(delay)

	ctx := context.Background()
	messages := []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "test"),
	}

	start := time.Now()
	resp, err := provider.GenerateMessage(ctx, messages)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "This response was intentionally delayed")
	assert.GreaterOrEqual(t, duration, delay)
}

func TestStreamingMockProvider(t *testing.T) {
	provider := StreamingMockProvider()

	ctx := context.Background()
	messages := []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "test"),
	}

	// Test streaming mode
	stream, err := provider.StreamMessage(ctx, messages)
	require.NoError(t, err)

	var chunks []string
	for token := range stream {
		chunks = append(chunks, token.Text)
		if token.Finished {
			break
		}
	}

	assert.Greater(t, len(chunks), 1)
	fullText := ""
	for _, chunk := range chunks {
		fullText += chunk
	}
	assert.Equal(t, "This is a streaming response that arrives in chunks.", fullText)

	// Test non-streaming mode
	resp, err := provider.GenerateMessage(ctx, messages)
	assert.NoError(t, err)
	assert.Equal(t, "This is a streaming response that arrives in chunks.", resp.Content)
}
