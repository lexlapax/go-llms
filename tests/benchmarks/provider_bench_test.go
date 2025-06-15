package benchmarks

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils/fixtures"
)

// Removed unused testHTTPClient

// BenchmarkProviderMessageConversion benchmarks the message conversion process
func BenchmarkProviderMessageConversion(b *testing.B) {
	// Create sample message arrays of different sizes
	smallMessages := fixtures.CreateSampleMessages(fixtures.MessageSizes.Small)   // System + User + Assistant
	mediumMessages := fixtures.CreateSampleMessages(fixtures.MessageSizes.Medium) // System + 3 exchanges
	largeMessages := fixtures.CreateSampleMessages(fixtures.MessageSizes.Large)   // System + 10 exchanges

	// Messages with tool calls
	toolMessages := fixtures.CreateMessagesWithTools(7) // System + User + 2 tool exchanges

	// Create unoptimized providers - we're just testing conversion, not actual API calls
	// so we use dummy API keys and models
	openaiProvider := provider.NewOpenAIProvider("dummy-key", "gpt-4")
	anthropicProvider := provider.NewAnthropicProvider("dummy-key", "claude-3-5-sonnet-latest")
	geminiProvider := provider.NewGeminiProvider("dummy-key", "gemini-2.0-flash-lite")

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

	// Benchmark Gemini message conversion with different message sizes
	b.Run("Gemini_SmallMessages", func(b *testing.B) {
		runGeminiMessageConversionBenchmark(b, geminiProvider, smallMessages)
	})

	b.Run("Gemini_MediumMessages", func(b *testing.B) {
		runGeminiMessageConversionBenchmark(b, geminiProvider, mediumMessages)
	})

	b.Run("Gemini_LargeMessages", func(b *testing.B) {
		runGeminiMessageConversionBenchmark(b, geminiProvider, largeMessages)
	})

	b.Run("Gemini_ToolMessages", func(b *testing.B) {
		runGeminiMessageConversionBenchmark(b, geminiProvider, toolMessages)
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

// runGeminiMessageConversionBenchmark benchmarks the message conversion process for Gemini
func runGeminiMessageConversionBenchmark(b *testing.B, p *provider.GeminiProvider, messages []domain.Message) {
	// Run the benchmark with the optimized conversion method
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Call the optimized conversion method directly
		geminiMessages := p.ConvertMessagesToGeminiFormat(messages)

		// We need to use the results to prevent the compiler from optimizing away the call
		if len(geminiMessages) == 0 {
			b.Fatalf("Expected non-empty geminiMessages, got empty slice")
		}
	}
}
