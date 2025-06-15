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

// Tests for enhanced provider fixtures

func TestOpenAIMockProvider(t *testing.T) {
	provider := OpenAIMockProvider()

	ctx := context.Background()

	// Test summarize pattern
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "Please summarize this document for me.",
		}},
	}}

	resp, err := provider.GenerateMessage(ctx, messages)
	require.NoError(t, err)
	assert.Contains(t, resp.Content, "summary")

	// Verify call was recorded with metadata
	history := provider.GetCallHistory()
	assert.Len(t, history, 1)

	// Check that the response pattern was matched (content confirms the pattern worked)
	assert.Equal(t, "Here's a concise summary of the key points:", resp.Content)
}

func TestAnthropicMockProvider(t *testing.T) {
	provider := AnthropicMockProvider()

	ctx := context.Background()

	// Test research pattern
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "I need you to research AI trends.",
		}},
	}}

	resp, err := provider.GenerateMessage(ctx, messages)
	require.NoError(t, err)
	assert.Contains(t, resp.Content, "research")

	// Check that the response pattern was matched
	assert.Equal(t, "I'll conduct thorough research on this topic. Let me analyze the available information and provide you with comprehensive findings.", resp.Content)
}

func TestGeminiMockProvider(t *testing.T) {
	provider := GeminiMockProvider()

	ctx := context.Background()

	// Test creative pattern
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "Create something creative for my project.",
		}},
	}}

	resp, err := provider.GenerateMessage(ctx, messages)
	require.NoError(t, err)
	assert.Contains(t, resp.Content, "creative")

	// Check that the response pattern was matched
	assert.Equal(t, "I'll create something creative for you! Here's an innovative approach to your request.", resp.Content)
}

func TestRealisticStreamingProvider(t *testing.T) {
	provider := RealisticStreamingProvider()

	ctx := context.Background()
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "Stream me a response",
		}},
	}}

	stream, err := provider.StreamMessage(ctx, messages)
	require.NoError(t, err)
	require.NotNil(t, stream)

	var tokens []string
	var finished bool

	// Collect streaming tokens with timeout
	timeout := time.After(5 * time.Second)
	for !finished {
		select {
		case token, ok := <-stream:
			if !ok {
				finished = true
				break
			}
			tokens = append(tokens, token.Text)
			if token.Finished {
				finished = true
			}
		case <-timeout:
			t.Fatal("Streaming timeout")
		}
	}

	// Verify we received the expected content
	fullText := ""
	for _, token := range tokens {
		fullText += token
	}
	assert.Contains(t, fullText, "realistic")
	assert.Contains(t, fullText, "streaming")
	assert.True(t, len(tokens) > 1, "Should receive multiple tokens")
}

func TestFastStreamingProvider(t *testing.T) {
	provider := FastStreamingProvider()

	ctx := context.Background()
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "Fast response please",
		}},
	}}

	start := time.Now()
	stream, err := provider.StreamMessage(ctx, messages)
	require.NoError(t, err)

	var tokenCount int
	for token := range stream {
		tokenCount++
		if token.Finished {
			break
		}
	}

	duration := time.Since(start)
	assert.True(t, duration < 500*time.Millisecond, "Fast streaming should complete quickly")
	assert.True(t, tokenCount > 1, "Should receive multiple tokens")
}

func TestRateLimitErrorProvider(t *testing.T) {
	provider := RateLimitErrorProvider()

	ctx := context.Background()
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "This should fail with rate limit",
		}},
	}}

	_, err := provider.GenerateMessage(ctx, messages)
	require.Error(t, err)

	// Check if it's a rate limit error
	rateLimitErr, ok := err.(*RateLimitError)
	require.True(t, ok, "Error should be a RateLimitError")
	assert.Contains(t, rateLimitErr.Message, "Rate limit exceeded")
	assert.Equal(t, 60*time.Second, rateLimitErr.RetryAfter)
}

func TestAuthenticationErrorProvider(t *testing.T) {
	provider := AuthenticationErrorProvider()

	ctx := context.Background()
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "This should fail with auth error",
		}},
	}}

	_, err := provider.GenerateMessage(ctx, messages)
	require.Error(t, err)

	// Check if it's an authentication error
	authErr, ok := err.(*AuthenticationError)
	require.True(t, ok, "Error should be an AuthenticationError")
	assert.Contains(t, authErr.Message, "Invalid API key")
}

func TestNetworkErrorProvider(t *testing.T) {
	provider := NetworkErrorProvider()

	ctx := context.Background()
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "This should fail with network error",
		}},
	}}

	_, err := provider.GenerateMessage(ctx, messages)
	require.Error(t, err)

	// Check if it's a network error
	netErr, ok := err.(*NetworkError)
	require.True(t, ok, "Error should be a NetworkError")
	assert.Contains(t, netErr.Message, "timeout")
	assert.True(t, netErr.Timeout)
}

func TestIntermittentErrorProvider(t *testing.T) {
	provider := IntermittentErrorProvider(0.7) // 70% success rate

	ctx := context.Background()
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "Test intermittent failures",
		}},
	}}

	successCount := 0
	errorCount := 0

	// Try multiple calls to test intermittent behavior
	for i := 0; i < 20; i++ {
		_, err := provider.GenerateMessage(ctx, messages)
		if err != nil {
			errorCount++
			assert.Contains(t, err.Error(), "intermittent_error")
		} else {
			successCount++
		}
	}

	// Should have both successes and failures
	assert.True(t, successCount > 0, "Should have some successful calls")
	assert.True(t, errorCount > 0, "Should have some failed calls")
}

func TestConfiguredOpenAIProvider(t *testing.T) {
	provider := ConfiguredOpenAIProvider("gpt-4o", 0.7, 500)

	ctx := context.Background()
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "Test configured provider",
		}},
	}}

	resp, err := provider.GenerateMessage(ctx, messages)
	require.NoError(t, err)

	// Verify content is returned
	assert.Equal(t, "I'll help you with that request.", resp.Content)

	// Provider configuration is stored in fixture but not directly accessible in response
}

func TestConfiguredAnthropicProvider(t *testing.T) {
	provider := ConfiguredAnthropicProvider("claude-3-opus", 1000, 0.3)

	ctx := context.Background()
	messages := []ldomain.Message{{
		Role: "user",
		Content: []ldomain.ContentPart{{
			Type: ldomain.ContentTypeText,
			Text: "Test configured Claude provider",
		}},
	}}

	resp, err := provider.GenerateMessage(ctx, messages)
	require.NoError(t, err)

	// Verify content is returned
	assert.Equal(t, "I understand your request. Let me provide you with a thoughtful and comprehensive response.", resp.Content)

	// Provider configuration is stored in fixture but not directly accessible in response
}

func TestBasicMockProvider(t *testing.T) {
	provider := BasicMockProvider()
	assert.NotNil(t, provider)
	assert.Equal(t, "basic-mock", provider.ProviderName)

	ctx := context.Background()
	messages := []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "test message"),
	}

	resp, err := provider.GenerateMessage(ctx, messages)
	assert.NoError(t, err)
	assert.Equal(t, "Test response", resp.Content)
}

func TestBasicMockProviderWithContent(t *testing.T) {
	content := "Custom test content"
	provider := BasicMockProviderWithContent(content)
	assert.NotNil(t, provider)
	assert.Equal(t, "basic-mock", provider.ProviderName)

	ctx := context.Background()
	messages := []ldomain.Message{
		ldomain.NewTextMessage(ldomain.RoleUser, "test message"),
	}

	resp, err := provider.GenerateMessage(ctx, messages)
	assert.NoError(t, err)
	assert.Equal(t, content, resp.Content)
}

func TestErrorTypes(t *testing.T) {
	t.Run("RateLimitError", func(t *testing.T) {
		err := &RateLimitError{
			Message:    "Rate limit exceeded",
			RetryAfter: 30 * time.Second,
		}
		assert.Equal(t, "Rate limit exceeded", err.Error())
		assert.Equal(t, 30*time.Second, err.RetryAfter)
	})

	t.Run("AuthenticationError", func(t *testing.T) {
		err := &AuthenticationError{
			Message: "Invalid credentials",
		}
		assert.Equal(t, "Invalid credentials", err.Error())
	})

	t.Run("NetworkError", func(t *testing.T) {
		err := &NetworkError{
			Message: "Connection timeout",
			Timeout: true,
		}
		assert.Equal(t, "Connection timeout", err.Error())
		assert.True(t, err.Timeout)
	})
}
