package provider

import (
	"hash/fnv"
	"sync"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// MessageCache provides caching for converted messages to avoid repeated conversions
type MessageCache struct {
	lock  sync.RWMutex
	cache map[uint64]interface{}
}

// NewMessageCache creates a new message cache with a default capacity
func NewMessageCache() *MessageCache {
	return &MessageCache{
		cache: make(map[uint64]interface{}, 10), // Default capacity of 10 conversations
	}
}

// Get retrieves a cached message conversion for the given key
// Returns the cached value and a boolean indicating if it was found
func (c *MessageCache) Get(key uint64) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value, ok := c.cache[key]
	return value, ok
}

// Set stores a message conversion in the cache
func (c *MessageCache) Set(key uint64, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.cache[key] = value
}

// Clear empties the cache
func (c *MessageCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// If the cache is already large, allocate a new one with the same capacity
	if len(c.cache) > 10 {
		c.cache = make(map[uint64]interface{}, len(c.cache))
	} else {
		// Otherwise just clear the map
		for k := range c.cache {
			delete(c.cache, k)
		}
	}
}

// GenerateMessagesKey creates a hash key for a message array
// This is used for cache lookups to avoid repeated message conversions
func GenerateMessagesKey(messages []domain.Message) uint64 {
	hasher := fnv.New64()

	// Write each message to the hasher
	for _, msg := range messages {
		// Add role to hash
		hasher.Write([]byte(msg.Role))

		// Handle multimodal content
		for _, part := range msg.Content {
			// Add content type to hash
			hasher.Write([]byte(part.Type))

			// Add content based on type
			switch part.Type {
			case domain.ContentTypeText:
				hasher.Write([]byte(part.Text))
			case domain.ContentTypeImage:
				if part.Image != nil {
					hasher.Write([]byte(part.Image.Source.Type))
					if part.Image.Source.Type == domain.SourceTypeBase64 {
						hasher.Write([]byte(part.Image.Source.MediaType))
						// Only hash a portion of the data to avoid excessive memory usage
						if len(part.Image.Source.Data) > 100 {
							hasher.Write([]byte(part.Image.Source.Data[:100]))
						} else {
							hasher.Write([]byte(part.Image.Source.Data))
						}
					} else {
						hasher.Write([]byte(part.Image.Source.URL))
					}
				}
			case domain.ContentTypeFile:
				if part.File != nil {
					hasher.Write([]byte(part.File.FileName))
					hasher.Write([]byte(part.File.MimeType))
					// Only hash a portion of the data to avoid excessive memory usage
					if len(part.File.FileData) > 100 {
						hasher.Write([]byte(part.File.FileData[:100]))
					} else {
						hasher.Write([]byte(part.File.FileData))
					}
				}
			case domain.ContentTypeVideo:
				if part.Video != nil {
					hasher.Write([]byte(part.Video.Source.Type))
					if part.Video.Source.Type == domain.SourceTypeBase64 {
						hasher.Write([]byte(part.Video.Source.MediaType))
						// Only hash a portion of the data to avoid excessive memory usage
						if len(part.Video.Source.Data) > 100 {
							hasher.Write([]byte(part.Video.Source.Data[:100]))
						} else {
							hasher.Write([]byte(part.Video.Source.Data))
						}
					} else {
						hasher.Write([]byte(part.Video.Source.URL))
					}
				}
			case domain.ContentTypeAudio:
				if part.Audio != nil {
					hasher.Write([]byte(part.Audio.Source.Type))
					if part.Audio.Source.Type == domain.SourceTypeBase64 {
						hasher.Write([]byte(part.Audio.Source.MediaType))
						// Only hash a portion of the data to avoid excessive memory usage
						if len(part.Audio.Source.Data) > 100 {
							hasher.Write([]byte(part.Audio.Source.Data[:100]))
						} else {
							hasher.Write([]byte(part.Audio.Source.Data))
						}
					} else {
						hasher.Write([]byte(part.Audio.Source.URL))
					}
				}
			}
		}
	}

	return hasher.Sum64()
}

// Removed unused preAllocateMessages function

// ConvertMessageToMap converts a domain.Message to a map with pre-allocated fields
// This reduces allocations during message conversion
func ConvertMessageToMap(msg domain.Message) map[string]interface{} {
	// Pre-allocate the map with enough capacity for common fields
	result := make(map[string]interface{}, 5)

	// Add standard fields
	result["role"] = string(msg.Role)
	result["content"] = msg.Content

	return result
}

// BuildRequestBody creates a request body map with pre-allocated fields
// This reduces allocations during request creation
func BuildRequestBody(model string, capacity int) map[string]interface{} {
	// Pre-allocate the map with enough capacity for common fields
	requestBody := make(map[string]interface{}, capacity)
	requestBody["model"] = model

	return requestBody
}

// AddOptionToRequestBody adds an option to the request body if it has a non-zero value
func AddOptionToRequestBody(requestBody map[string]interface{}, key string, value interface{}) {
	// Skip zero values unless they're booleans (which could be intentionally false)
	switch v := value.(type) {
	case int:
		if v != 0 {
			requestBody[key] = v
		}
	case float64:
		if v != 0 {
			requestBody[key] = v
		}
	case string:
		if v != "" {
			requestBody[key] = v
		}
	case []string:
		if len(v) > 0 {
			requestBody[key] = v
		}
	case bool:
		requestBody[key] = v
	default:
		// For other types, add only if not nil
		if value != nil {
			requestBody[key] = value
		}
	}
}
