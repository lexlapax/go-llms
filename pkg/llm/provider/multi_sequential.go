// Package provider implements various LLM providers.
package provider

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// sequentialGenerateForPrimary runs Generate sequentially for the primary provider strategy
// This is a fixed implementation that avoids race conditions in the concurrent version
func (mp *MultiProvider) sequentialGenerateForPrimary(ctx context.Context, prompt string, options []domain.Option) (string, error) {
	// If we're not using StrategyPrimary, fall back to concurrent implementation
	if mp.selectionStrat != StrategyPrimary {
		results := mp.concurrentGenerate(ctx, prompt, options)
		return mp.selectTextResult(results)
	}

	// Get the primary provider index
	primaryIdx := mp.primaryProvider
	if primaryIdx < 0 || primaryIdx >= len(mp.providers) {
		primaryIdx = 0 // Default to first provider if invalid index
	}

	// Try the primary provider first
	primaryProvider := mp.providers[primaryIdx]
	content, err := primaryProvider.Provider.Generate(ctx, prompt, options...)
	if err == nil {
		return content, nil
	}

	// If primary fails, try the other providers sequentially
	for i, pw := range mp.providers {
		// Skip the primary we already tried
		if i == primaryIdx {
			continue
		}

		// Check if context is canceled
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Continue with next provider
		}

		content, err := pw.Provider.Generate(ctx, prompt, options...)
		if err == nil {
			return content, nil
		}
	}

	// All providers failed
	return "", ErrNoSuccessfulCalls
}

// sequentialGenerateMessageForPrimary runs GenerateMessage sequentially for the primary provider strategy
func (mp *MultiProvider) sequentialGenerateMessageForPrimary(ctx context.Context, messages []domain.Message, options []domain.Option) (domain.Response, error) {
	// If we're not using StrategyPrimary, fall back to concurrent implementation
	if mp.selectionStrat != StrategyPrimary {
		results := mp.concurrentGenerateMessage(ctx, messages, options)
		return mp.selectMessageResult(results)
	}

	// Get the primary provider index
	primaryIdx := mp.primaryProvider
	if primaryIdx < 0 || primaryIdx >= len(mp.providers) {
		primaryIdx = 0 // Default to first provider if invalid index
	}

	// Try the primary provider first
	primaryProvider := mp.providers[primaryIdx]
	response, err := primaryProvider.Provider.GenerateMessage(ctx, messages, options...)
	if err == nil {
		return response, nil
	}

	// If primary fails, try the other providers sequentially
	for i, pw := range mp.providers {
		// Skip the primary we already tried
		if i == primaryIdx {
			continue
		}

		// Check if context is canceled
		select {
		case <-ctx.Done():
			return domain.Response{}, ctx.Err()
		default:
			// Continue with next provider
		}

		response, err := pw.Provider.GenerateMessage(ctx, messages, options...)
		if err == nil {
			return response, nil
		}
	}

	// All providers failed
	return domain.Response{}, ErrNoSuccessfulCalls
}

// sequentialGenerateWithSchemaForPrimary runs GenerateWithSchema sequentially for the primary provider strategy
func (mp *MultiProvider) sequentialGenerateWithSchemaForPrimary(ctx context.Context, prompt string, schema *schemaDomain.Schema, options []domain.Option) (interface{}, error) {
	// If we're not using StrategyPrimary, fall back to concurrent implementation
	if mp.selectionStrat != StrategyPrimary {
		results := mp.concurrentGenerateWithSchema(ctx, prompt, schema, options)
		return mp.selectStructuredResult(results)
	}

	// Get the primary provider index
	primaryIdx := mp.primaryProvider
	if primaryIdx < 0 || primaryIdx >= len(mp.providers) {
		primaryIdx = 0 // Default to first provider if invalid index
	}

	// Try the primary provider first
	primaryProvider := mp.providers[primaryIdx]
	result, err := primaryProvider.Provider.GenerateWithSchema(ctx, prompt, schema, options...)
	if err == nil {
		return result, nil
	}

	// If primary fails, try the other providers sequentially
	for i, pw := range mp.providers {
		// Skip the primary we already tried
		if i == primaryIdx {
			continue
		}

		// Check if context is canceled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue with next provider
		}

		result, err := pw.Provider.GenerateWithSchema(ctx, prompt, schema, options...)
		if err == nil {
			return result, nil
		}
	}

	// All providers failed
	return nil, ErrNoSuccessfulCalls
}
