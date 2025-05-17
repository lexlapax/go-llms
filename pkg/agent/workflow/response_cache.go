// Package workflow provides agent workflow implementations.
package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// ResponseCacheEntry represents a cached response
type ResponseCacheEntry struct {
	Response   ldomain.Response
	Timestamp  time.Time
	UsageCount int
	Source     string // Description of where response came from (model, etc.)
}

// ResponseCache provides a thread-safe cache for LLM responses
// to avoid redundant API calls for the same input
type ResponseCache struct {
	// Main cache storage - key is hash of the messages + options
	cache    map[string]ResponseCacheEntry
	capacity int           // Maximum number of entries to store
	ttl      time.Duration // Time-to-live for cache entries
	mu       sync.RWMutex  // Thread safety
}

// NewResponseCache creates a new response cache with the given capacity and TTL
func NewResponseCache(capacity int, ttl time.Duration) *ResponseCache {
	// Set reasonable defaults if parameters are invalid
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}

	if ttl <= 0 {
		ttl = 5 * time.Minute // Default TTL
	}

	return &ResponseCache{
		cache:    make(map[string]ResponseCacheEntry, capacity),
		capacity: capacity,
		ttl:      ttl,
	}
}

// Get retrieves a cached response for the given messages and options
// Returns the response and a boolean indicating if it was found
func (c *ResponseCache) Get(messages []ldomain.Message, options []ldomain.Option) (ldomain.Response, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Generate cache key from messages and options
	key := c.generateCacheKey(messages, options)

	entry, found := c.cache[key]
	if !found {
		return ldomain.Response{}, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > c.ttl {
		// Will be cleaned up later by Cleanup method
		return ldomain.Response{}, false
	}

	// Update usage count (requires write lock)
	c.mu.RUnlock()
	c.mu.Lock()
	entry.UsageCount++
	c.cache[key] = entry
	c.mu.Unlock()
	c.mu.RLock()

	return entry.Response, true
}

// Set adds or updates a response in the cache
func (c *ResponseCache) Set(messages []ldomain.Message, options []ldomain.Option, response ldomain.Response, source string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean up expired entries if cache is at capacity
	if len(c.cache) >= c.capacity {
		c.cleanupLocked()
	}

	// Generate cache key
	key := c.generateCacheKey(messages, options)

	// Update or insert entry
	entry, found := c.cache[key]
	if found {
		// Update existing entry
		entry.Response = response
		entry.Timestamp = time.Now()
		entry.UsageCount++
		entry.Source = source
	} else {
		// Create new entry
		entry = ResponseCacheEntry{
			Response:   response,
			Timestamp:  time.Now(),
			UsageCount: 1,
			Source:     source,
		}
	}
	c.cache[key] = entry
}

// Invalidate removes a specific entry from the cache
func (c *ResponseCache) Invalidate(messages []ldomain.Message, options []ldomain.Option) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateCacheKey(messages, options)
	delete(c.cache, key)
}

// Cleanup removes expired entries and trims the cache if needed
func (c *ResponseCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cleanupLocked()
}

// cleanupLocked is the internal implementation of Cleanup
// that assumes the mutex is already locked
func (c *ResponseCache) cleanupLocked() {
	now := time.Now()

	// First pass: remove expired entries
	for key, entry := range c.cache {
		if now.Sub(entry.Timestamp) > c.ttl {
			delete(c.cache, key)
		}
	}

	// If still over capacity, remove least recently used entries
	if len(c.cache) > c.capacity {
		// Convert map to slice for sorting
		entries := make([]struct {
			key   string
			entry ResponseCacheEntry
		}, 0, len(c.cache))

		for k, v := range c.cache {
			entries = append(entries, struct {
				key   string
				entry ResponseCacheEntry
			}{k, v})
		}

		// Sort by usage count and timestamp (older and less used first)
		sort.Slice(entries, func(i, j int) bool {
			// Primary sort by usage count
			if entries[i].entry.UsageCount != entries[j].entry.UsageCount {
				return entries[i].entry.UsageCount < entries[j].entry.UsageCount
			}
			// Secondary sort by timestamp
			return entries[i].entry.Timestamp.Before(entries[j].entry.Timestamp)
		})

		// Remove oldest entries until we're under capacity
		for i := 0; i < len(entries) && len(c.cache) > c.capacity; i++ {
			delete(c.cache, entries[i].key)
		}
	}
}

// Clear empties the entire cache
func (c *ResponseCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]ResponseCacheEntry, c.capacity)
}

// GetStats returns statistics about the cache
func (c *ResponseCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["size"] = len(c.cache)
	stats["capacity"] = c.capacity
	stats["ttl_seconds"] = c.ttl.Seconds()

	// Count responses by source
	sourceCount := make(map[string]int)
	for _, entry := range c.cache {
		sourceCount[entry.Source]++
	}
	stats["sources"] = sourceCount

	return stats
}

// generateCacheKey creates a deterministic key for caching
// It hashes the combination of messages and options
func (c *ResponseCache) generateCacheKey(messages []ldomain.Message, options []ldomain.Option) string {
	// Normalize messages to avoid inconsistent caching
	var messageData []map[string]interface{}
	for _, msg := range messages {
		msgData := map[string]interface{}{
			"role": string(msg.Role),
		}

		// Handle content parts
		if len(msg.Content) > 0 {
			// For simple messages with just text, use a simpler representation
			if len(msg.Content) == 1 && msg.Content[0].Type == ldomain.ContentTypeText {
				msgData["content"] = msg.Content[0].Text
			} else {
				// For multimodal content, create a representation of each part
				var contentParts []map[string]interface{}
				for _, part := range msg.Content {
					partData := map[string]interface{}{
						"type": string(part.Type),
					}

					switch part.Type {
					case ldomain.ContentTypeText:
						partData["text"] = part.Text
					case ldomain.ContentTypeImage:
						if part.Image != nil {
							// Just use media type and source type as identifiers
							// to avoid storing large base64 strings in cache keys
							partData["media_type"] = part.Image.Source.MediaType
							partData["source_type"] = string(part.Image.Source.Type)
						}
					case ldomain.ContentTypeFile:
						if part.File != nil {
							partData["file_name"] = part.File.FileName
							partData["mime_type"] = part.File.MimeType
						}
					case ldomain.ContentTypeVideo, ldomain.ContentTypeAudio:
						// Just include the type
						partData["media_content"] = true
					}

					contentParts = append(contentParts, partData)
				}
				msgData["content_parts"] = contentParts
			}
		}

		messageData = append(messageData, msgData)
	}

	// Add options to the key if present
	var optionsData []map[string]string
	for _, opt := range options {
		// Get option name and value through reflection
		optType := fmt.Sprintf("%T", opt)
		optStr := fmt.Sprintf("%v", opt)

		// Add to options data
		optionsData = append(optionsData, map[string]string{
			"type":  optType,
			"value": optStr,
		})
	}

	// Create a combined structure for hashing
	data := map[string]interface{}{
		"messages": messageData,
		"options":  optionsData,
	}

	// Marshal to JSON and hash
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to a simpler key if marshaling fails
		var sb strings.Builder
		for _, msg := range messages {
			sb.WriteString(string(msg.Role))
			sb.WriteString(":")

			// Extract text content for the key
			if len(msg.Content) > 0 {
				for _, part := range msg.Content {
					if part.Type == ldomain.ContentTypeText {
						sb.WriteString(part.Text)
					} else {
						sb.WriteString("[" + string(part.Type) + "]")
					}
				}
			}

			sb.WriteString("|")
		}
		return hashString(sb.String())
	}

	return hashString(string(jsonData))
}

// hashString creates a SHA-256 hash of the input string
func hashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// Global cache for sharing across agents if needed
var globalResponseCache *ResponseCache
var globalCacheOnce sync.Once

// GetGlobalResponseCache returns the shared global response cache
func GetGlobalResponseCache() *ResponseCache {
	globalCacheOnce.Do(func() {
		globalResponseCache = NewResponseCache(200, 10*time.Minute)
	})
	return globalResponseCache
}
