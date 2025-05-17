package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	llmDomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Context represents our execution context
type Context struct {
	Config Config
	ctx    context.Context
}

// createProvider creates an LLM provider based on configuration
func (c *Context) createProvider() (llmDomain.Provider, error) {
	providerName := c.Config.Provider
	
	apiKey, err := GetOptimizedAPIKey(providerName)
	if err != nil {
		return nil, err
	}
	
	_, modelName, err := GetOptimizedProvider()
	if err != nil {
		return nil, err
	}
	
	switch providerName {
	case "openai":
		return provider.NewOpenAIProvider(apiKey, modelName), nil
	case "anthropic":
		return provider.NewAnthropicProvider(apiKey, modelName), nil
	case "gemini":
		return provider.NewGeminiProvider(apiKey, modelName), nil
	case "mock":
		return provider.NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

// streamOutput handles streaming output
func streamOutput(ctx context.Context, stream <-chan string, w io.Writer) error {
	for {
		select {
		case chunk, ok := <-stream:
			if !ok {
				return nil
			}
			_, err := w.Write([]byte(chunk))
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// formatOutput formats output based on configuration
func formatOutput(output string, format string) string {
	switch format {
	case "json":
		// For JSON format, ensure it's valid JSON
		output = strings.TrimSpace(output)
		if !strings.HasPrefix(output, "{") && !strings.HasPrefix(output, "[") {
			// Wrap in a simple JSON object if it's not already JSON
			return fmt.Sprintf(`{"output": %q}`, output)
		}
		return output
	default:
		return output
	}
}