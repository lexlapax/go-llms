// Package provider implements various LLM providers.
package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Error definitions for multi-provider operations
var (
	ErrNoProviders       = errors.New("no providers configured")
	ErrNoSuccessfulCalls = errors.New("no successful responses from any providers")
	ErrContextCanceled   = errors.New("operation canceled due to context cancellation")
	ErrProviderTimeout   = errors.New("provider operation timed out")
)

// SelectionStrategy defines how to select results when multiple providers return valid responses
type SelectionStrategy int

const (
	// StrategyFastest returns the first successful result
	StrategyFastest SelectionStrategy = iota
	// StrategyPrimary uses the primary provider, falling back to others on failure
	StrategyPrimary
	// StrategyConsensus attempts to find consensus among multiple results (future)
	StrategyConsensus
)

// ProviderWeight defines the weight of a provider in a multi-provider setup
type ProviderWeight struct {
	Provider domain.Provider
	Weight   float64 // 0.0 to 1.0, default 1.0
	Name     string  // Optional name for the provider
}

// fallbackResult stores a result from a provider with metadata
type fallbackResult struct {
	provider    string
	content     string
	response    domain.Response
	structured  interface{}
	err         error
	elapsedTime time.Duration
	weight      float64 // Provider weight, used for weighted consensus
}

// MultiProvider implements domain.Provider interface and distributes operations
// across multiple underlying providers with fallback and selection strategies
type MultiProvider struct {
	providers       []ProviderWeight
	selectionStrat  SelectionStrategy
	defaultTimeout  time.Duration
	primaryProvider int             // Index of primary provider for StrategyPrimary
	consensusConfig consensusConfig // Configuration for consensus algorithms
}

// NewMultiProvider creates a new provider that distributes operations across multiple providers
func NewMultiProvider(providers []ProviderWeight, strategy SelectionStrategy) *MultiProvider {
	// Default timeout of 30 seconds
	return &MultiProvider{
		providers:       providers,
		selectionStrat:  strategy,
		defaultTimeout:  30 * time.Second,
		consensusConfig: defaultConsensusConfig(),
	}
}

// WithTimeout configures the default timeout for provider operations
func (mp *MultiProvider) WithTimeout(timeout time.Duration) *MultiProvider {
	mp.defaultTimeout = timeout
	return mp
}

// WithPrimaryProvider sets the index of the primary provider for StrategyPrimary
func (mp *MultiProvider) WithPrimaryProvider(index int) *MultiProvider {
	if index >= 0 && index < len(mp.providers) {
		mp.primaryProvider = index
	}
	return mp
}

// WithConsensusStrategy sets the consensus strategy to use
func (mp *MultiProvider) WithConsensusStrategy(strategy ConsensusStrategy) *MultiProvider {
	mp.consensusConfig.Strategy = strategy
	return mp
}

// WithSimilarityThreshold sets the similarity threshold for ConsensusSimilarity strategy
// The threshold should be between 0.0 and 1.0, with higher values requiring more similarity
func (mp *MultiProvider) WithSimilarityThreshold(threshold float64) *MultiProvider {
	// Ensure threshold is within valid range
	if threshold < 0.0 {
		threshold = 0.0
	}
	if threshold > 1.0 {
		threshold = 1.0
	}
	mp.consensusConfig.SimilarityThreshold = threshold
	return mp
}

// Generate produces text from a prompt concurrently using multiple providers
func (mp *MultiProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	if len(mp.providers) == 0 {
		return "", ErrNoProviders
	}

	// Apply the configured timeout if not overridden in the context
	ctx, cancel := applyTimeoutFromContext(ctx, mp.defaultTimeout)
	defer cancel()

	// Use sequential execution for primary strategy to ensure deterministic results
	if mp.selectionStrat == StrategyPrimary {
		return mp.sequentialGenerateForPrimary(ctx, prompt, options)
	}

	// Use concurrent execution for other strategies
	results := mp.concurrentGenerate(ctx, prompt, options)

	// Apply selection strategy
	return mp.selectTextResult(results)
}

// GenerateMessage produces text from a list of messages using multiple providers
func (mp *MultiProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	if len(mp.providers) == 0 {
		return domain.Response{}, ErrNoProviders
	}

	// Apply the configured timeout if not overridden in the context
	ctx, cancel := applyTimeoutFromContext(ctx, mp.defaultTimeout)
	defer cancel()

	// Use sequential execution for primary strategy to ensure deterministic results
	if mp.selectionStrat == StrategyPrimary {
		return mp.sequentialGenerateMessageForPrimary(ctx, messages, options)
	}

	// Use concurrent execution for other strategies
	results := mp.concurrentGenerateMessage(ctx, messages, options)

	// Apply selection strategy
	return mp.selectMessageResult(results)
}

// GenerateWithSchema produces structured output conforming to a schema using multiple providers
func (mp *MultiProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	if len(mp.providers) == 0 {
		return nil, ErrNoProviders
	}

	// Apply the configured timeout if not overridden in the context
	ctx, cancel := applyTimeoutFromContext(ctx, mp.defaultTimeout)
	defer cancel()

	// Use sequential execution for primary strategy to ensure deterministic results
	if mp.selectionStrat == StrategyPrimary {
		return mp.sequentialGenerateWithSchemaForPrimary(ctx, prompt, schema, options)
	}

	// Use concurrent execution for other strategies
	results := mp.concurrentGenerateWithSchema(ctx, prompt, schema, options)

	// Apply selection strategy
	return mp.selectStructuredResult(results)
}

// Stream streams responses token by token from the fastest or primary provider
// Note: Unlike the other methods, Stream doesn't try all providers concurrently as this would require
// complex token aggregation logic. Instead, it follows the selected strategy more directly.
func (mp *MultiProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	if len(mp.providers) == 0 {
		return nil, ErrNoProviders
	}

	// Apply the configured timeout if not overridden in the context
	ctx, cancel := applyTimeoutFromContext(ctx, mp.defaultTimeout)

	// For streaming, we select the provider upfront rather than aggregating results
	// This approach is more straightforward for streaming responses
	selectedProviderIdx := mp.selectProviderForStreaming()

	// Get a channel from the pool
	responseStream, responseCh := domain.GetChannelPool().GetResponseStream()

	// Start the stream in a goroutine
	go func() {
		defer cancel() // Ensure the context is canceled when we're done
		// We're not returning the channel to the pool here because:
		// 1. close(responseCh) will be called, making the channel unusable for reuse
		// 2. The channel pool's Put method checks if the channel is closed and won't reuse it

		// Try the selected provider first
		stream, err := mp.providers[selectedProviderIdx].Provider.Stream(ctx, prompt, options...)
		if err == nil {
			// Forward tokens from the provider to our response channel
			mp.forwardStream(ctx, stream, responseCh)
			return
		}

		// If the primary fails, try each provider in order
		for i, pw := range mp.providers {
			if i == selectedProviderIdx {
				continue // Skip the one we already tried
			}

			// Check if the context was canceled
			select {
			case <-ctx.Done():
				close(responseCh)
				return
			default:
			}

			stream, err := pw.Provider.Stream(ctx, prompt, options...)
			if err == nil {
				// Forward tokens from the provider to our response channel
				mp.forwardStream(ctx, stream, responseCh)
				return
			}
		}

		// If all providers failed, send an error token
		select {
		case <-ctx.Done():
		case responseCh <- domain.Token{
			Text:     "[ERROR: All providers failed]",
			Finished: true,
		}:
		}
		close(responseCh)
	}()

	return responseStream, nil
}

// StreamMessage streams responses from a list of messages
func (mp *MultiProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	if len(mp.providers) == 0 {
		return nil, ErrNoProviders
	}

	// Apply the configured timeout if not overridden in the context
	ctx, cancel := applyTimeoutFromContext(ctx, mp.defaultTimeout)

	// For streaming, we select the provider upfront rather than aggregating results
	selectedProviderIdx := mp.selectProviderForStreaming()

	// Get a channel from the pool
	responseStream, responseCh := domain.GetChannelPool().GetResponseStream()

	// Start the stream in a goroutine
	go func() {
		defer cancel() // Ensure the context is canceled when we're done
		// We're not returning the channel to the pool here because:
		// 1. close(responseCh) will be called, making the channel unusable for reuse
		// 2. The channel pool's Put method checks if the channel is closed and won't reuse it

		// Try the selected provider first
		stream, err := mp.providers[selectedProviderIdx].Provider.StreamMessage(ctx, messages, options...)
		if err == nil {
			// Forward tokens from the provider to our response channel
			mp.forwardStream(ctx, stream, responseCh)
			return
		}

		// If the primary fails, try each provider in order
		for i, pw := range mp.providers {
			if i == selectedProviderIdx {
				continue // Skip the one we already tried
			}

			// Check if the context was canceled
			select {
			case <-ctx.Done():
				close(responseCh)
				return
			default:
			}

			stream, err := pw.Provider.StreamMessage(ctx, messages, options...)
			if err == nil {
				// Forward tokens from the provider to our response channel
				mp.forwardStream(ctx, stream, responseCh)
				return
			}
		}

		// If all providers failed, send an error token
		select {
		case <-ctx.Done():
		case responseCh <- domain.Token{
			Text:     "[ERROR: All providers failed]",
			Finished: true,
		}:
		}
		close(responseCh)
	}()

	return responseStream, nil
}

// Helper methods for concurrent operations

// concurrentGenerate runs Generate on all providers concurrently
func (mp *MultiProvider) concurrentGenerate(ctx context.Context, prompt string, options []domain.Option) []fallbackResult {
	resultCh := make(chan fallbackResult, len(mp.providers))
	var wg sync.WaitGroup

	// Launch a goroutine for each provider
	for i, pw := range mp.providers {
		wg.Add(1)
		go func(idx int, providerWeight ProviderWeight) {
			defer wg.Done()

			providerName := providerWeight.Name
			if providerName == "" {
				providerName = fmt.Sprintf("provider_%d", idx)
			}

			startTime := time.Now()
			content, err := providerWeight.Provider.Generate(ctx, prompt, options...)
			elapsed := time.Since(startTime)

			// Send result regardless of error status
			resultCh <- fallbackResult{
				provider:    providerName,
				content:     content,
				err:         err,
				elapsedTime: elapsed,
				weight:      providerWeight.Weight,
			}
		}(i, pw)
	}

	// Close the channel when all providers have completed
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect all results
	var results []fallbackResult
	for result := range resultCh {
		results = append(results, result)
	}

	return results
}

// concurrentGenerateMessage runs GenerateMessage on all providers concurrently
func (mp *MultiProvider) concurrentGenerateMessage(ctx context.Context, messages []domain.Message, options []domain.Option) []fallbackResult {
	resultCh := make(chan fallbackResult, len(mp.providers))
	var wg sync.WaitGroup

	// Launch a goroutine for each provider
	for i, pw := range mp.providers {
		wg.Add(1)
		go func(idx int, providerWeight ProviderWeight) {
			defer wg.Done()

			providerName := providerWeight.Name
			if providerName == "" {
				providerName = fmt.Sprintf("provider_%d", idx)
			}

			startTime := time.Now()
			response, err := providerWeight.Provider.GenerateMessage(ctx, messages, options...)
			elapsed := time.Since(startTime)

			// Send result regardless of error status
			resultCh <- fallbackResult{
				provider:    providerName,
				response:    response,
				err:         err,
				elapsedTime: elapsed,
				weight:      providerWeight.Weight,
			}
		}(i, pw)
	}

	// Close the channel when all providers have completed
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect all results
	var results []fallbackResult
	for result := range resultCh {
		results = append(results, result)
	}

	return results
}

// concurrentGenerateWithSchema runs GenerateWithSchema on all providers concurrently
func (mp *MultiProvider) concurrentGenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options []domain.Option) []fallbackResult {
	resultCh := make(chan fallbackResult, len(mp.providers))
	var wg sync.WaitGroup

	// Launch a goroutine for each provider
	for i, pw := range mp.providers {
		wg.Add(1)
		go func(idx int, providerWeight ProviderWeight) {
			defer wg.Done()

			providerName := providerWeight.Name
			if providerName == "" {
				providerName = fmt.Sprintf("provider_%d", idx)
			}

			startTime := time.Now()
			result, err := providerWeight.Provider.GenerateWithSchema(ctx, prompt, schema, options...)
			elapsed := time.Since(startTime)

			// Send result regardless of error status
			resultCh <- fallbackResult{
				provider:    providerName,
				structured:  result,
				err:         err,
				elapsedTime: elapsed,
				weight:      providerWeight.Weight,
			}
		}(i, pw)
	}

	// Close the channel when all providers have completed
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect all results
	var results []fallbackResult
	for result := range resultCh {
		results = append(results, result)
	}

	return results
}

// forwardStream forwards tokens from a source stream to a destination channel
func (mp *MultiProvider) forwardStream(ctx context.Context, sourceStream domain.ResponseStream, destCh chan domain.Token) {
	for {
		select {
		case <-ctx.Done():
			// Context was canceled, close the destination channel
			close(destCh)
			return
		case token, ok := <-sourceStream:
			if !ok {
				// Source stream closed, close destination channel
				close(destCh)
				return
			}

			// Forward the token to the destination
			select {
			case <-ctx.Done():
				close(destCh)
				return
			case destCh <- token:
				// Token forwarded successfully
				if token.Finished {
					close(destCh)
					return
				}
			}
		}
	}
}

// Result selection helpers

// selectTextResult selects a text result based on the configured strategy
func (mp *MultiProvider) selectTextResult(results []fallbackResult) (string, error) {
	if len(results) == 0 {
		return "", ErrNoSuccessfulCalls
	}

	switch mp.selectionStrat {
	case StrategyFastest:
		// Sort by time and return the first successful one
		sortResultsByTime(results)
		for _, result := range results {
			if result.err == nil {
				return result.content, nil
			}
		}

	case StrategyPrimary:
		// Try primary provider first, then fall back
		primaryIdx := mp.primaryProvider
		if primaryIdx >= 0 && primaryIdx < len(results) && results[primaryIdx].err == nil {
			return results[primaryIdx].content, nil
		}

		// Primary failed, try others
		for _, result := range results {
			if result.err == nil {
				return result.content, nil
			}
		}

	case StrategyConsensus:
		// Set the global consensus configuration to use our settings
		// This allows the selectConsensusTextResult to use our configuration
		globalConsensusConfig = &mp.consensusConfig

		// Use enhanced consensus algorithms
		return selectConsensusTextResult(results)
	}

	// If we get here, all providers failed
	var errMsg string
	for _, result := range results {
		if result.err != nil {
			errMsg += fmt.Sprintf("[%s: %v] ", result.provider, result.err)
		}
	}
	return "", fmt.Errorf("%w: %s", ErrNoSuccessfulCalls, errMsg)
}

// selectMessageResult selects a response result based on the configured strategy
func (mp *MultiProvider) selectMessageResult(results []fallbackResult) (domain.Response, error) {
	if len(results) == 0 {
		return domain.Response{}, ErrNoSuccessfulCalls
	}

	switch mp.selectionStrat {
	case StrategyFastest:
		// Sort by time and return the first successful one
		sortResultsByTime(results)
		for _, result := range results {
			if result.err == nil {
				return result.response, nil
			}
		}

	case StrategyPrimary:
		// Try primary provider first, then fall back
		primaryIdx := mp.primaryProvider
		if primaryIdx >= 0 && primaryIdx < len(results) && results[primaryIdx].err == nil {
			return results[primaryIdx].response, nil
		}

		// Primary failed, try others
		for _, result := range results {
			if result.err == nil {
				return result.response, nil
			}
		}

	case StrategyConsensus:
		// Set the global consensus configuration to use our settings
		// This allows the selectConsensusTextResult to use our configuration
		globalConsensusConfig = &mp.consensusConfig

		// Use enhanced consensus algorithms for message responses
		responsePool := domain.GetResponsePool()
		consensusText, err := selectConsensusTextResult(results)
		if err != nil {
			return domain.Response{}, err
		}
		return responsePool.NewResponse(consensusText), nil
	}

	// If we get here, all providers failed
	var errMsg string
	for _, result := range results {
		if result.err != nil {
			errMsg += fmt.Sprintf("[%s: %v] ", result.provider, result.err)
		}
	}
	return domain.Response{}, fmt.Errorf("%w: %s", ErrNoSuccessfulCalls, errMsg)
}

// selectStructuredResult selects a structured result based on the configured strategy
func (mp *MultiProvider) selectStructuredResult(results []fallbackResult) (interface{}, error) {
	if len(results) == 0 {
		return nil, ErrNoSuccessfulCalls
	}

	switch mp.selectionStrat {
	case StrategyFastest:
		// Sort by time and return the first successful one
		sortResultsByTime(results)
		for _, result := range results {
			if result.err == nil {
				return result.structured, nil
			}
		}

	case StrategyPrimary:
		// Try primary provider first, then fall back
		primaryIdx := mp.primaryProvider
		if primaryIdx >= 0 && primaryIdx < len(results) && results[primaryIdx].err == nil {
			return results[primaryIdx].structured, nil
		}

		// Primary failed, try others
		for _, result := range results {
			if result.err == nil {
				return result.structured, nil
			}
		}

	case StrategyConsensus:
		// Set the global consensus configuration to use our settings
		// This allows the selectConsensusTextResult to use our configuration
		globalConsensusConfig = &mp.consensusConfig

		// For structured results, we need specialized handling based on schema types
		// Optimized implementation: use similarity-based grouping on JSON representations
		// and return the most common structure

		// Pre-allocate with estimated capacity to avoid resizing
		structuredResults := make([]string, 0, len(results))
		structuredMap := make(map[string]interface{}, len(results))

		// Convert structured results to JSON strings for comparison
		for _, result := range results {
			if result.err == nil && result.structured != nil {
				// Convert to JSON for comparison
				jsonBytes, err := json.Marshal(result.structured)
				if err == nil {
					jsonStr := string(jsonBytes)
					structuredResults = append(structuredResults, jsonStr)
					structuredMap[jsonStr] = result.structured
				}
			}
		}

		// Fast path: If we only have one structured result, return it immediately
		if len(structuredResults) == 1 {
			return structuredMap[structuredResults[0]], nil
		}

		// If we have multiple structured results, find the most common one
		if len(structuredResults) > 0 {
			// Create fallback results for text-based consensus algorithms
			// Pre-allocate with exact capacity
			textResults := make([]fallbackResult, len(structuredResults))

			// Create a map from JSON string to original result index for faster lookups
			jsonToResultIndex := make(map[string]int, len(results))
			for i, r := range results {
				if r.err == nil && r.structured != nil {
					jsonBytes, jsonErr := json.Marshal(r.structured)
					if jsonErr == nil {
						jsonToResultIndex[string(jsonBytes)] = i
					}
				}
			}

			// Populate text results with efficient lookups
			for i, jsonStr := range structuredResults {
				// Initialize with default values
				textResults[i] = fallbackResult{
					content: jsonStr,
					err:     nil,
					weight:  1.0, // Default weight
				}

				// If we can find the original result, use its metadata
				if origIndex, ok := jsonToResultIndex[jsonStr]; ok {
					origResult := results[origIndex]
					textResults[i].weight = origResult.weight
					textResults[i].elapsedTime = origResult.elapsedTime
					textResults[i].provider = origResult.provider
				}
			}

			// Use consensus algorithms to find the most common structure
			consensusJSON, err := selectConsensusTextResult(textResults)
			if err == nil && consensusJSON != "" {
				// Return the structured object matching the consensus JSON
				if structured, ok := structuredMap[consensusJSON]; ok {
					return structured, nil
				}
			}
		}

		// Fallback to the first successful result if consensus fails
		for _, result := range results {
			if result.err == nil && result.structured != nil {
				return result.structured, nil
			}
		}
	}

	// If we get here, all providers failed or returned nil
	var errMsg string
	for _, result := range results {
		if result.err != nil {
			errMsg += fmt.Sprintf("[%s: %v] ", result.provider, result.err)
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrNoSuccessfulCalls, errMsg)
}

// selectProviderForStreaming selects which provider to use for streaming based on the strategy
func (mp *MultiProvider) selectProviderForStreaming() int {
	switch mp.selectionStrat {
	case StrategyPrimary:
		if mp.primaryProvider >= 0 && mp.primaryProvider < len(mp.providers) {
			return mp.primaryProvider
		}
		return 0 // Default to first provider
	case StrategyFastest, StrategyConsensus:
		// For fastest and consensus, just start with the first provider
		// (consensus doesn't really apply to streaming)
		return 0
	default:
		return 0
	}
}

// Utility functions

// applyTimeoutFromContext applies a timeout to a context if it doesn't already have one
func applyTimeoutFromContext(ctx context.Context, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	// Check if the context already has a deadline
	deadline, ok := ctx.Deadline()
	if !ok {
		// No deadline, apply our default timeout
		return context.WithTimeout(ctx, defaultTimeout)
	}

	// Context already has a deadline, use that
	timeRemaining := time.Until(deadline)
	if timeRemaining <= 0 {
		// Deadline already passed, use a minimal timeout
		return context.WithTimeout(ctx, 1*time.Millisecond)
	}

	// Return the original context with a cancel function
	childCtx, cancel := context.WithCancel(ctx)
	return childCtx, cancel
}

// sortResultsByTime sorts results by elapsed time (fastest first)
func sortResultsByTime(results []fallbackResult) {
	// Simple bubble sort for a small number of providers
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].elapsedTime < results[i].elapsedTime {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// The legacy_selectConsensusTextResult function has been removed
// as the implementation has moved to consensus.go
