package benchmarks

import (
	"net/http"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// testHTTPClient is a mock HTTP client that returns a predefined response
type testHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (c *testHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.DoFunc(req)
}

// createSampleMessages creates test message arrays of different sizes
func createSampleMessages(size int) []domain.Message {
	// Pre-allocate the slice
	messages := make([]domain.Message, 0, size)

	// Always start with a system message
	messages = append(messages, domain.Message{
		Role:    domain.RoleSystem,
		Content: "You are a helpful assistant that provides accurate and concise information.",
	})

	// Add user-assistant message pairs
	for i := 1; i < size; i += 2 {
		// Add user message
		if i < size {
			messages = append(messages, domain.Message{
				Role:    domain.RoleUser,
				Content: "This is a user message for testing performance.",
			})
		}

		// Add assistant message
		if i+1 < size {
			messages = append(messages, domain.Message{
				Role:    domain.RoleAssistant,
				Content: "This is an assistant response for testing performance.",
			})
		}
	}

	return messages
}

// createMessagesWithTools creates test message arrays that include tool calls
func createMessagesWithTools(size int) []domain.Message {
	// Pre-allocate the slice
	messages := make([]domain.Message, 0, size)

	// Add system message
	messages = append(messages, domain.Message{
		Role:    domain.RoleSystem,
		Content: "You are a helpful assistant that can use tools.",
	})

	// Add user message first
	messages = append(messages, domain.Message{
		Role:    domain.RoleUser,
		Content: "I need help with a calculation.",
	})

	// Fill remaining messages with assistant-tool pairs
	remaining := size - 2
	for i := 0; i < remaining; i += 2 {
		// Add assistant message with tool call
		if i < remaining {
			messages = append(messages, domain.Message{
				Role:    domain.RoleAssistant,
				Content: "I'll use the calculator tool to help you.",
			})
		}

		// Add tool response
		if i+1 < remaining {
			messages = append(messages, domain.Message{
				Role:    domain.RoleTool,
				Content: "Result: 42",
			})
		}
	}

	return messages
}

// BenchmarkProviderMessageConversion benchmarks the message conversion process
func BenchmarkProviderMessageConversion(b *testing.B) {
	// Create sample message arrays of different sizes
	smallMessages := createSampleMessages(3)  // System + User + Assistant
	mediumMessages := createSampleMessages(7) // System + 3 exchanges
	largeMessages := createSampleMessages(21) // System + 10 exchanges

	// Messages with tool calls
	toolMessages := createMessagesWithTools(7) // System + User + 2 tool exchanges

	// Create unoptimized providers - we're just testing conversion, not actual API calls
	// so we use dummy API keys and models
	openaiProvider := provider.NewOpenAIProvider("dummy-key", "gpt-4")
	anthropicProvider := provider.NewAnthropicProvider("dummy-key", "claude-3-5-sonnet-latest")

	// Benchmark OpenAI message conversion with different message sizes
	b.Run("OpenAI_SmallMessages", func(b *testing.B) {
		runOpenAIMessageConversionBenchmark(b, openaiProvider, smallMessages)
	})

	b.Run("OpenAI_MediumMessages", func(b *testing.B) {
		runOpenAIMessageConversionBenchmark(b, openaiProvider, mediumMessages)
	})

	b.Run("OpenAI_LargeMessages", func(b *testing.B) {
		runOpenAIMessageConversionBenchmark(b, openaiProvider, largeMessages)
	})

	b.Run("OpenAI_ToolMessages", func(b *testing.B) {
		runOpenAIMessageConversionBenchmark(b, openaiProvider, toolMessages)
	})

	// Benchmark Anthropic message conversion with different message sizes
	b.Run("Anthropic_SmallMessages", func(b *testing.B) {
		runAnthropicMessageConversionBenchmark(b, anthropicProvider, smallMessages)
	})

	b.Run("Anthropic_MediumMessages", func(b *testing.B) {
		runAnthropicMessageConversionBenchmark(b, anthropicProvider, mediumMessages)
	})

	b.Run("Anthropic_LargeMessages", func(b *testing.B) {
		runAnthropicMessageConversionBenchmark(b, anthropicProvider, largeMessages)
	})

	b.Run("Anthropic_ToolMessages", func(b *testing.B) {
		runAnthropicMessageConversionBenchmark(b, anthropicProvider, toolMessages)
	})
}

// runOpenAIMessageConversionBenchmark benchmarks the message conversion process for OpenAI
func runOpenAIMessageConversionBenchmark(b *testing.B, p *provider.OpenAIProvider, messages []domain.Message) {
	// Run the benchmark with the optimized conversion method
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Call the optimized conversion method directly
		oaiMessages := p.ConvertMessagesToOpenAIFormat(messages)
		
		// We need to use the result to prevent the compiler from optimizing away the call
		if len(oaiMessages) == 0 {
			b.Fatalf("Expected non-empty oaiMessages, got empty slice")
		}
	}
}

// runAnthropicMessageConversionBenchmark benchmarks the message conversion process for Anthropic
func runAnthropicMessageConversionBenchmark(b *testing.B, p *provider.AnthropicProvider, messages []domain.Message) {
	// Run the benchmark with the optimized conversion method
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Call the optimized conversion method directly
		anthMessages, systemMessage := p.ConvertMessagesToAnthropicFormat(messages)
		
		// We need to use the results to prevent the compiler from optimizing away the call
		if len(systemMessage) == 0 && len(anthMessages) == 0 {
			b.Fatalf("Expected non-empty results, got empty data")
		}
	}
}