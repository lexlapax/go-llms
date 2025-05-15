// Package workflow provides agent workflow implementations.
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// CachedAgent extends MultiAgent with response caching capabilities
// It's designed to avoid redundant LLM API calls for repeated or similar requests
type CachedAgent struct {
	// Embed DefaultAgent directly to avoid copying MultiAgent's sync.Map
	DefaultAgent

	// MultiAgent components we need (we'll initialize these manually)
	multiProviderMetrics *MultiProviderMetrics
	providerContextCache sync.Map

	// Response cache for storing and retrieving LLM responses
	responseCache *ResponseCache

	// Configuration for caching behavior
	cacheConfig CacheConfig

	// Cache statistics for monitoring
	cacheStats CacheStats
}

// executeMultipleToolsParallel delegates to MultiAgent's implementation
func (a *CachedAgent) executeMultipleToolsParallel(ctx context.Context, toolNames []string, paramsArray []interface{}) (string, error) {
	// Create a MultiAgent instance to delegate to
	multiAgent := &MultiAgent{DefaultAgent: a.DefaultAgent}
	return multiAgent.executeMultipleToolsParallel(ctx, toolNames, paramsArray)
}

// CacheConfig holds configuration for the cache behavior
type CacheConfig struct {
	// Whether to enable caching (can be toggled at runtime)
	Enabled bool

	// Maximum time to keep a cached entry
	TTL time.Duration

	// Maximum number of entries in the cache
	Capacity int

	// Whether to use fuzzy matching for cache lookups
	// If true, will match similar queries even if not exactly the same
	FuzzyMatching bool

	// Threshold for fuzzy matching (0.0-1.0)
	FuzzyThreshold float64

	// Whether to share cache across all agent instances
	UseGlobalCache bool
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits                    int
	Misses                  int
	StoredResponses         int
	EvictedResponses        int
	FuzzyMatchSuccesses     int
	FuzzyMatchFailures      int
	AverageResponseSavingMs int64
	LastCacheCleanup        time.Time
}

// NewCachedAgent creates a new agent with caching capabilities
func NewCachedAgent(provider ldomain.Provider) *CachedAgent {
	multiAgent := NewMultiAgent(provider)

	// Default cache configuration
	config := CacheConfig{
		Enabled:        true,
		TTL:            10 * time.Minute,
		Capacity:       100,
		FuzzyMatching:  true,
		FuzzyThreshold: 0.85,
		UseGlobalCache: false,
	}

	var cache *ResponseCache
	if config.UseGlobalCache {
		cache = GetGlobalResponseCache()
	} else {
		cache = NewResponseCache(config.Capacity, config.TTL)
	}

	// Create a new CachedAgent without copying the MultiAgent struct
	// This avoids copying the sync.Map which contains a sync.noCopy
	agent := &CachedAgent{
		responseCache: cache,
		cacheConfig:   config,
		cacheStats:    CacheStats{LastCacheCleanup: time.Now()},
	}

	// Embed the MultiAgent using pointer semantics
	agent.DefaultAgent = multiAgent.DefaultAgent
	agent.multiProviderMetrics = multiAgent.multiProviderMetrics
	agent.providerContextCache = sync.Map{} // Create a new sync.Map instead of copying

	return agent
}

// Run executes the agent with given inputs
// This implementation adds caching for responses
func (a *CachedAgent) Run(ctx context.Context, input string) (interface{}, error) {
	return a.run(ctx, input, nil)
}

// RunWithSchema executes the agent and validates output against a schema
// This implementation adds caching for responses
func (a *CachedAgent) RunWithSchema(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	return a.run(ctx, input, schema)
}

// run is the internal implementation of Run and RunWithSchema
// This version includes response caching for improved performance
func (a *CachedAgent) run(ctx context.Context, input string, schema *sdomain.Schema) (interface{}, error) {
	// Prepare the prompt
	prompt := input
	if schema != nil {
		// Enhance the prompt with schema information
		enhancedPrompt, err := processor.EnhancePromptWithSchema(input, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to enhance prompt with schema: %w", err)
		}
		prompt = enhancedPrompt
	}

	// Create messages for the conversation - optimized version
	messages := a.createInitialMessages(prompt)

	// Check cache before making API calls - only if caching is enabled
	if a.cacheConfig.Enabled && schema == nil {
		// We only cache text generation, not structured generation
		// Try to get from cache
		modelOption := []ldomain.Option{}
		if a.modelName != "" {
			modelOption = append(modelOption, ldomain.WithModel(a.modelName))
		}

		var cachedResponse ldomain.Response
		var cacheHit bool

		if a.cacheConfig.FuzzyMatching {
			cachedResponse, cacheHit = a.getFuzzyMatchFromCache(messages, modelOption)
		} else {
			cachedResponse, cacheHit = a.responseCache.Get(messages, modelOption)
		}

		if cacheHit {
			// We found a cached response, use it
			a.cacheStats.Hits++

			// Process cached response for tool calls
			toolCalls, multiParams, shouldCallMultipleTools := a.ExtractMultipleToolCalls(cachedResponse.Content)

			if shouldCallMultipleTools && len(toolCalls) > 0 {
				// Process tool calls in parallel
				toolResponses, err := a.executeMultipleToolsParallel(ctx, toolCalls, multiParams)
				if err == nil {
					// Add the assistant message
					messages = append(messages, ldomain.Message{
						Role:    ldomain.RoleAssistant,
						Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: cachedResponse.Content}},
					})

					// Add tool results
					messages = append(messages, ldomain.Message{
						Role:    ldomain.RoleUser,
						Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: toolResponses}},
					})

					// Continue with a new generation for the tool results
					// This generation is NOT cached to ensure we get fresh results
					var options []ldomain.Option
					if a.modelName != "" {
						options = append(options, ldomain.WithModel(a.modelName))
					}

					// Call hooks before generate
					a.notifyBeforeGenerate(ctx, messages)

					resp, genErr := a.llmProvider.GenerateMessage(ctx, messages, options...)

					// Call hooks after generate
					a.notifyAfterGenerate(ctx, resp, genErr)

					if genErr != nil {
						return nil, fmt.Errorf("LLM generation failed after tool calls: %w", genErr)
					}

					return resp.Content, nil
				}
			}

			// Check for single tool call
			toolCall, params, shouldCallTool := a.ExtractToolCall(cachedResponse.Content)
			if shouldCallTool {
				// Execute the tool
				tool, found := a.tools[toolCall]
				if found {
					// Call hooks before tool call
					a.notifyBeforeToolCall(ctx, toolCall, params)

					// Execute the tool
					toolResult, toolErr := tool.Execute(ctx, params)

					// Call hooks after tool call
					a.notifyAfterToolCall(ctx, toolCall, toolResult, toolErr)

					// Format tool result
					var toolRespContent string
					if toolErr != nil {
						toolRespContent = fmt.Sprintf("Error: %v", toolErr)
					} else {
						switch v := toolResult.(type) {
						case string:
							toolRespContent = v
						case nil:
							toolRespContent = "Tool executed successfully with no output"
						default:
							jsonBytes, err := json.Marshal(toolResult)
							if err != nil {
								toolRespContent = fmt.Sprintf("%v", toolResult)
							} else {
								toolRespContent = string(jsonBytes)
							}
						}
					}

					// Add messages and generate a response
					messages = append(messages, ldomain.Message{
						Role:    ldomain.RoleAssistant,
						Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: cachedResponse.Content}},
					})

					messages = append(messages, ldomain.Message{
						Role:    ldomain.RoleUser,
						Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: fmt.Sprintf("Tool '%s' result: %s", toolCall, toolRespContent)}},
					})

					// Generate final response - NOT cached
					var options []ldomain.Option
					if a.modelName != "" {
						options = append(options, ldomain.WithModel(a.modelName))
					}

					// Call hooks
					a.notifyBeforeGenerate(ctx, messages)
					finalResp, finalErr := a.llmProvider.GenerateMessage(ctx, messages, options...)
					a.notifyAfterGenerate(ctx, finalResp, finalErr)

					if finalErr != nil {
						return nil, fmt.Errorf("LLM generation failed after tool call: %w", finalErr)
					}

					return finalResp.Content, nil
				}
			}

			// If no tool calls or tool not found, just return the cached response
			return cachedResponse.Content, nil
		} else {
			a.cacheStats.Misses++
		}
	}

	// Cache cleanup (lazy - only do it sometimes)
	if a.shouldCleanupCache() {
		go a.cleanupCache()
	}

	// Not in cache, proceed with normal agent operation
	// Main agent loop - continue until we have a result or error
	var finalResponse interface{}
	maxIterations := 10 // Prevent infinite loops
	for i := 0; i < maxIterations; i++ {
		// Call hooks before generate
		a.notifyBeforeGenerate(ctx, messages)

		// Generate response
		var resp ldomain.Response
		var genErr error
		startTime := time.Now()

		if schema != nil {
			// If we have a schema, use structured generation
			var result interface{}
			result, genErr = a.llmProvider.GenerateWithSchema(ctx, prompt, schema)
			if genErr == nil {
				// Convert result to final format if needed
				finalResponse = result
				break // We have a valid structured result, exit the loop
			}
		} else {
			// Regular text generation
			var options []ldomain.Option
			if a.modelName != "" {
				options = append(options, ldomain.WithModel(a.modelName))
			}
			resp, genErr = a.llmProvider.GenerateMessage(ctx, messages, options...)

			// Cache the response if generation was successful and caching is enabled
			if genErr == nil && a.cacheConfig.Enabled && i == 0 {
				// Only cache the first response (not follow-up responses after tool calls)
				// Also store the model name as source
				modelName := a.modelName
				if modelName == "" {
					modelName = "default"
				}
				a.responseCache.Set(messages, options, resp, modelName)
				a.cacheStats.StoredResponses++

				// Track time saved for future cache hits
				a.cacheStats.AverageResponseSavingMs = (a.cacheStats.AverageResponseSavingMs + time.Since(startTime).Milliseconds()) / 2
			}
		}

		// Call hooks after generate
		a.notifyAfterGenerate(ctx, resp, genErr)

		if genErr != nil {
			return nil, fmt.Errorf("LLM generation failed: %w", genErr)
		}

		// If we're doing structured output, the response is in finalResponse
		if schema != nil && finalResponse != nil {
			return finalResponse, nil
		}

		// First check for multiple tool calls (OpenAI format)
		toolCalls, multiParams, shouldCallMultipleTools := a.ExtractMultipleToolCalls(resp.Content)

		if shouldCallMultipleTools && len(toolCalls) > 0 {
			// Process tool calls in parallel
			toolResponses, err := a.executeMultipleToolsParallel(ctx, toolCalls, multiParams)
			if err != nil {
				return nil, fmt.Errorf("error executing tools: %w", err)
			}

			// Add the assistant message and tool results
			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleAssistant,
				Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
			})

			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleUser,
				Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: toolResponses}},
			})

			continue
		}

		// Fall back to legacy single tool call extraction if multiple tools weren't found
		toolCall, params, shouldCallTool := a.ExtractToolCall(resp.Content)
		if !shouldCallTool {
			// No tool call, just return the response content
			return resp.Content, nil
		}

		// Find the requested tool
		tool, found := a.tools[toolCall]
		if !found {
			// Tool not found, append error message and continue
			errMsg := fmt.Sprintf("Tool '%s' not found. Available tools: %s",
				toolCall, strings.Join(a.getToolNames(), ", "))

			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleAssistant,
				Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
			})

			messages = append(messages, ldomain.Message{
				Role:    ldomain.RoleUser,
				Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: fmt.Sprintf("Tool error: %s", errMsg)}},
			})
			continue
		}

		// Call hooks before tool call
		a.notifyBeforeToolCall(ctx, toolCall, params)

		// Execute the tool
		toolResult, toolErr := tool.Execute(ctx, params)

		// Call hooks after tool call
		a.notifyAfterToolCall(ctx, toolCall, toolResult, toolErr)

		// Add the result to messages
		var toolRespContent string
		if toolErr != nil {
			toolRespContent = fmt.Sprintf("Error: %v", toolErr)
		} else {
			// Convert tool result to string if needed
			switch v := toolResult.(type) {
			case string:
				toolRespContent = v
			case nil:
				toolRespContent = "Tool executed successfully with no output"
			default:
				// Try to marshal to JSON
				jsonBytes, err := json.Marshal(toolResult)
				if err != nil {
					toolRespContent = fmt.Sprintf("%v", toolResult)
				} else {
					toolRespContent = string(jsonBytes)
				}
			}
		}

		// Add the assistant message and tool result to the conversation
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleAssistant,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: resp.Content}},
		})

		// Use user role instead of tool role for better OpenAI compatibility
		messages = append(messages, ldomain.Message{
			Role:    ldomain.RoleUser,
			Content: []ldomain.ContentPart{{Type: ldomain.ContentTypeText, Text: fmt.Sprintf("Tool '%s' result: %s", toolCall, toolRespContent)}},
		})
	}

	// If we have a schema and final response, return it
	if schema != nil && finalResponse != nil {
		return finalResponse, nil
	}

	// If we reached max iterations, return what we have
	return "Agent reached maximum iterations without final result", nil
}

// getFuzzyMatchFromCache tries to find a suitable cached response even if not an exact match
func (a *CachedAgent) getFuzzyMatchFromCache(messages []ldomain.Message, options []ldomain.Option) (ldomain.Response, bool) {
	// First try an exact match
	if resp, found := a.responseCache.Get(messages, options); found {
		return resp, true
	}

	// If we only have system and one user message, try fuzzy matching on content
	if len(messages) != 2 || messages[0].Role != ldomain.RoleSystem || messages[1].Role != ldomain.RoleUser {
		return ldomain.Response{}, false
	}

	// Get target message to match
	targetMsg := ""
	if len(messages[1].Content) > 0 && messages[1].Content[0].Type == ldomain.ContentTypeText {
		targetMsg = messages[1].Content[0].Text
	}
	_ = quickHash(targetMsg) // Just compute hash for future implementation

	// TODO: Implement proper fuzzy matching by comparing with other cached entries
	// For now, we just use a simple message hash comparison

	// Cache hit not found
	a.cacheStats.FuzzyMatchFailures++
	return ldomain.Response{}, false
}

// shouldCleanupCache determines if it's time to clean up the cache
func (a *CachedAgent) shouldCleanupCache() bool {
	// Run cleanup approximately every 5 minutes
	return time.Since(a.cacheStats.LastCacheCleanup) > 5*time.Minute
}

// cleanupCache performs cache maintenance
func (a *CachedAgent) cleanupCache() {
	prevSize := a.responseCache.GetStats()["size"].(int)
	a.responseCache.Cleanup()
	newSize := a.responseCache.GetStats()["size"].(int)

	// Track evictions
	a.cacheStats.EvictedResponses += prevSize - newSize
	a.cacheStats.LastCacheCleanup = time.Now()
}

// GetCacheStats returns statistics about the cache
func (a *CachedAgent) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Add our stats
	stats["hits"] = a.cacheStats.Hits
	stats["misses"] = a.cacheStats.Misses
	stats["hit_ratio"] = float64(0)
	if (a.cacheStats.Hits + a.cacheStats.Misses) > 0 {
		stats["hit_ratio"] = float64(a.cacheStats.Hits) / float64(a.cacheStats.Hits+a.cacheStats.Misses)
	}
	stats["stored_responses"] = a.cacheStats.StoredResponses
	stats["evicted_responses"] = a.cacheStats.EvictedResponses
	stats["fuzzy_match_successes"] = a.cacheStats.FuzzyMatchSuccesses
	stats["fuzzy_match_failures"] = a.cacheStats.FuzzyMatchFailures
	stats["avg_response_saving_ms"] = a.cacheStats.AverageResponseSavingMs
	stats["last_cleanup"] = a.cacheStats.LastCacheCleanup.Format(time.RFC3339)

	// Add cache configuration
	stats["config"] = map[string]interface{}{
		"enabled":         a.cacheConfig.Enabled,
		"ttl_minutes":     a.cacheConfig.TTL.Minutes(),
		"capacity":        a.cacheConfig.Capacity,
		"fuzzy_matching":  a.cacheConfig.FuzzyMatching,
		"fuzzy_threshold": a.cacheConfig.FuzzyThreshold,
		"global_cache":    a.cacheConfig.UseGlobalCache,
	}

	// Merge with cache internals stats
	for k, v := range a.responseCache.GetStats() {
		stats["cache_"+k] = v
	}

	return stats
}

// EnableCaching turns caching on or off
func (a *CachedAgent) EnableCaching(enabled bool) {
	a.cacheConfig.Enabled = enabled
}

// SetCacheTTL sets the cache time-to-live
func (a *CachedAgent) SetCacheTTL(ttl time.Duration) {
	a.cacheConfig.TTL = ttl
}

// SetCacheCapacity sets the maximum number of entries in the cache
func (a *CachedAgent) SetCacheCapacity(capacity int) {
	a.cacheConfig.Capacity = capacity
}

// ClearCache empties the entire cache
func (a *CachedAgent) ClearCache() {
	a.responseCache.Clear()
	a.cacheStats.StoredResponses = 0
	a.cacheStats.EvictedResponses = 0
}

// WithModel specifies which LLM model to use
// Override to reset context cache when model changes
func (a *CachedAgent) WithModel(modelName string) domain.Agent {
	// Call the parent implementation (DefaultAgent)
	a.DefaultAgent.WithModel(modelName)

	// Clear response cache since model has changed
	a.ClearCache()

	return a
}

// quickHash generates a fast non-crypto hash of a string
func quickHash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return strconv.FormatUint(h.Sum64(), 16)
}
