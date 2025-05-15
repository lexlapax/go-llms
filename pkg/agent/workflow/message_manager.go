// Package workflow provides agent workflow implementations.
package workflow

import (
	"sync"
	"time"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// MessageManager handles efficient management of conversation message history
// It provides windowing, summarization, and memory optimization for large conversational contexts
type MessageManager struct {
	// Maximum number of messages to keep in full form
	maxMessages int

	// Maximum number of tokens to keep in the context
	maxTokens int

	// Current messages
	messages []ldomain.Message

	// Cache for token counts
	tokenCounts map[string]int

	// Thread safety
	mu sync.RWMutex

	// Time-based tracking for messages
	messageTimestamps []time.Time

	// Configuration
	config MessageManagerConfig

	// Pool for message objects to reduce allocations
	messagePool *sync.Pool
}

// MessageManagerConfig holds configuration options for message management
type MessageManagerConfig struct {
	// Whether to use token-based truncation (vs message count)
	UseTokenTruncation bool

	// Whether to keep all system messages
	KeepAllSystemMessages bool

	// Whether to use sliding window (recent messages only)
	UseSlidingWindow bool

	// Whether to compress older messages
	CompressOlderMessages bool

	// Time threshold for compressing older messages
	CompressionTimeThreshold time.Duration
}

// NewMessageManager creates a new message manager with the given configuration
func NewMessageManager(maxMessages, maxTokens int, config MessageManagerConfig) *MessageManager {
	// Use sensible defaults
	if maxMessages <= 0 {
		maxMessages = 100 // Default to 100 messages
	}

	if maxTokens <= 0 {
		maxTokens = 16000 // Default to 16K tokens
	}

	return &MessageManager{
		maxMessages:       maxMessages,
		maxTokens:         maxTokens,
		messages:          make([]ldomain.Message, 0, maxMessages),
		tokenCounts:       make(map[string]int),
		messageTimestamps: make([]time.Time, 0, maxMessages),
		config:            config,
		messagePool: &sync.Pool{
			New: func() interface{} {
				return &ldomain.Message{}
			},
		},
	}
}

// AddMessage adds a message to the history
func (m *MessageManager) AddMessage(message ldomain.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clone the message to prevent external modification
	newMsg := m.cloneMessage(message)

	// Add the message
	m.messages = append(m.messages, newMsg)
	m.messageTimestamps = append(m.messageTimestamps, time.Now())

	// Calculate and cache token count
	contentKey := getContentKey(newMsg.Content)
	tokens := m.estimateTokensFromContentParts(newMsg.Content)
	m.tokenCounts[contentKey] = tokens

	// Apply truncation if needed
	m.applyTruncation()
}

// AddMessages adds multiple messages to the history
func (m *MessageManager) AddMessages(messages []ldomain.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Pre-allocate to avoid resizing
	if cap(m.messages)-len(m.messages) < len(messages) {
		// Need to grow the slice
		newCap := len(m.messages) + len(messages) + 10 // Add some buffer
		newMessages := make([]ldomain.Message, len(m.messages), newCap)
		copy(newMessages, m.messages)
		m.messages = newMessages

		// Also grow timestamps slice
		newTimestamps := make([]time.Time, len(m.messageTimestamps), newCap)
		copy(newTimestamps, m.messageTimestamps)
		m.messageTimestamps = newTimestamps
	}

	// Add all messages
	now := time.Now()
	for _, msg := range messages {
		// Clone the message
		newMsg := m.cloneMessage(msg)

		// Add it
		m.messages = append(m.messages, newMsg)
		m.messageTimestamps = append(m.messageTimestamps, now)

		// Cache token count
		contentKey := getContentKey(newMsg.Content)
		tokens := m.estimateTokensFromContentParts(newMsg.Content)
		m.tokenCounts[contentKey] = tokens
	}

	// Apply truncation if needed
	m.applyTruncation()
}

// GetMessages returns the current message history
func (m *MessageManager) GetMessages() []ldomain.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a deep copy to prevent external modification
	result := make([]ldomain.Message, len(m.messages))
	for i, msg := range m.messages {
		result[i] = m.cloneMessage(msg)
	}

	return result
}

// GetMessagesForModel returns a message history optimized for the model's context window
func (m *MessageManager) GetMessagesForModel(modelTokenLimit int) []ldomain.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// If no specific model limit, return all messages
	if modelTokenLimit <= 0 {
		return m.GetMessages()
	}

	// Always include system messages
	var systemMessages []ldomain.Message
	var nonSystemMessages []ldomain.Message

	// Separate system and non-system messages
	for _, msg := range m.messages {
		if msg.Role == ldomain.RoleSystem {
			systemMessages = append(systemMessages, m.cloneMessage(msg))
		} else {
			// Store message with its timestamp index for later sorting
			nonSystemMsg := m.cloneMessage(msg)
			nonSystemMessages = append(nonSystemMessages, nonSystemMsg)
		}
	}

	// Calculate total tokens in system messages
	systemTokens := 0
	for _, msg := range systemMessages {
		systemTokens += m.estimateTokensFromContentParts(msg.Content)
	}

	// Start with all system messages
	result := systemMessages
	remainingTokens := modelTokenLimit - systemTokens

	// Now add as many non-system messages as fit, prioritizing recent ones
	// Start from most recent
	for i := len(nonSystemMessages) - 1; i >= 0 && remainingTokens > 0; i-- {
		msg := nonSystemMessages[i]
		tokens := m.estimateTokensFromContentParts(msg.Content)

		if tokens <= remainingTokens {
			result = append(result, msg)
			remainingTokens -= tokens
		} else if tokens > 32 && remainingTokens > 32 {
			// If the message is too big but we have some space, truncate it
			truncated := m.truncateMessage(msg, remainingTokens)
			result = append(result, truncated)
			break
		}
	}

	return m.sortMessagesInOrder(result)
}

// Reset clears all messages
func (m *MessageManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = m.messages[:0]
	m.messageTimestamps = m.messageTimestamps[:0]
	m.tokenCounts = make(map[string]int)
}

// SetSystemPrompt replaces all system messages with a single one
func (m *MessageManager) SetSystemPrompt(prompt string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove all existing system messages
	newMessages := make([]ldomain.Message, 0, len(m.messages))
	newTimestamps := make([]time.Time, 0, len(m.messages))

	for i, msg := range m.messages {
		if msg.Role != ldomain.RoleSystem {
			newMessages = append(newMessages, msg)
			newTimestamps = append(newTimestamps, m.messageTimestamps[i])
		}
	}

	// Add the new system message at the beginning
	if prompt != "" {
		sysMsg := ldomain.NewTextMessage(ldomain.RoleSystem, prompt)

		// Insert at the beginning
		newMessages = append([]ldomain.Message{sysMsg}, newMessages...)
		newTimestamps = append([]time.Time{time.Now()}, newTimestamps...)

		// Update token count for the new message
		contentKey := getContentKey(sysMsg.Content)
		m.tokenCounts[contentKey] = m.estimateTokensFromContentParts(sysMsg.Content)
	}

	m.messages = newMessages
	m.messageTimestamps = newTimestamps
}

// GetTokenCount returns the current total token count
func (m *MessageManager) GetTokenCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0
	for _, msg := range m.messages {
		total += m.estimateTokensFromContentParts(msg.Content)
	}

	return total
}

// GetMessageCount returns the number of messages
func (m *MessageManager) GetMessageCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.messages)
}

// Internal helper methods

// applyTruncation applies truncation based on the configuration
func (m *MessageManager) applyTruncation() {
	// Check if we need to truncate
	if len(m.messages) <= m.maxMessages {
		return
	}

	// If token-based truncation is enabled
	if m.config.UseTokenTruncation {
		// Calculate current token usage
		totalTokens := 0
		for _, msg := range m.messages {
			totalTokens += m.estimateTokensFromContentParts(msg.Content)
		}

		// If under token limit, no need to truncate
		if totalTokens <= m.maxTokens {
			return
		}

		// Need to remove oldest messages until under token limit
		// (keeping system messages if configured)
		newMessages := make([]ldomain.Message, 0, m.maxMessages)
		newTimestamps := make([]time.Time, 0, m.maxMessages)

		// Always keep system messages if configured
		if m.config.KeepAllSystemMessages {
			for i, msg := range m.messages {
				if msg.Role == ldomain.RoleSystem {
					newMessages = append(newMessages, msg)
					newTimestamps = append(newTimestamps, m.messageTimestamps[i])
					totalTokens -= m.estimateTokensFromContentParts(msg.Content)
				}
			}
		}

		// Add most recent messages until we hit the token limit
		for i := len(m.messages) - 1; i >= 0 && totalTokens > m.maxTokens; i-- {
			// Skip system messages which we've already processed
			if m.config.KeepAllSystemMessages && m.messages[i].Role == ldomain.RoleSystem {
				continue
			}

			// Add message and update token count
			newMessages = append([]ldomain.Message{m.messages[i]}, newMessages...)
			newTimestamps = append([]time.Time{m.messageTimestamps[i]}, newTimestamps...)
			totalTokens -= m.estimateTokensFromContentParts(m.messages[i].Content)
		}

		m.messages = newMessages
		m.messageTimestamps = newTimestamps
	} else {
		// Simple message count truncation
		// Keep the last maxMessages messages
		excess := len(m.messages) - m.maxMessages
		if excess <= 0 {
			return
		}

		// If keeping system messages, we need to collect them first
		if m.config.KeepAllSystemMessages {
			newMessages := make([]ldomain.Message, 0, m.maxMessages)
			newTimestamps := make([]time.Time, 0, m.maxMessages)

			// Keep all system messages
			for i, msg := range m.messages {
				if msg.Role == ldomain.RoleSystem {
					newMessages = append(newMessages, msg)
					newTimestamps = append(newTimestamps, m.messageTimestamps[i])
				}
			}

			// Calculate how many non-system messages we can keep
			nonSystemToKeep := m.maxMessages - len(newMessages)
			if nonSystemToKeep <= 0 {
				// If we have more system messages than max, just keep the first maxMessages
				m.messages = m.messages[:m.maxMessages]
				m.messageTimestamps = m.messageTimestamps[:m.maxMessages]
				return
			}

			// Add the most recent non-system messages
			nonSystemCount := 0
			for i := len(m.messages) - 1; i >= 0 && nonSystemCount < nonSystemToKeep; i-- {
				if m.messages[i].Role != ldomain.RoleSystem {
					newMessages = append([]ldomain.Message{m.messages[i]}, newMessages...)
					newTimestamps = append([]time.Time{m.messageTimestamps[i]}, newTimestamps...)
					nonSystemCount++
				}
			}

			m.messages = newMessages
			m.messageTimestamps = newTimestamps
		} else {
			// Simple truncation - keep most recent messages
			m.messages = m.messages[excess:]
			m.messageTimestamps = m.messageTimestamps[excess:]
		}
	}
}

// estimateTokens estimates the number of tokens in a string
// This is a simple approximation, models might tokenize differently
func (m *MessageManager) estimateTokens(text string) int {
	// Check if we already calculated this
	if count, ok := m.tokenCounts[text]; ok {
		return count
	}

	// Simple heuristic: approximately 4 characters per token for English text
	// Plus overhead for message formatting, role, etc.
	tokenCount := len(text)/4 + 5

	// Cache for future reference
	m.tokenCounts[text] = tokenCount
	return tokenCount
}

// estimateTokensFromContentParts estimates tokens from ContentPart array
func (m *MessageManager) estimateTokensFromContentParts(contentParts []ldomain.ContentPart) int {
	// Generate a key for the content parts
	contentKey := getContentKey(contentParts)
	
	// Check if we already calculated this
	if count, ok := m.tokenCounts[contentKey]; ok {
		return count
	}
	
	totalTokens := 0
	
	// Add up tokens for each part
	for _, part := range contentParts {
		switch part.Type {
		case ldomain.ContentTypeText:
			totalTokens += m.estimateTokens(part.Text)
		case ldomain.ContentTypeImage:
			// Images typically use more tokens than text
			// A rough estimate: 1000 tokens per image
			totalTokens += 1000
		case ldomain.ContentTypeFile, ldomain.ContentTypeVideo, ldomain.ContentTypeAudio:
			// Other media types: estimate 500 tokens
			totalTokens += 500
		}
	}
	
	// Add overhead for multipart message
	if len(contentParts) > 1 {
		totalTokens += 10 * len(contentParts)
	}
	
	// Cache for future reference
	m.tokenCounts[contentKey] = totalTokens
	return totalTokens
}

// getContentKey generates a unique key for content parts for caching
func getContentKey(contentParts []ldomain.ContentPart) string {
	if len(contentParts) == 0 {
		return "empty_content"
	}
	
	// For simple text-only content, use the text directly as the key
	if len(contentParts) == 1 && contentParts[0].Type == ldomain.ContentTypeText {
		return contentParts[0].Text
	}
	
	// For multimodal content, build a key that includes type info
	key := ""
	for i, part := range contentParts {
		key += string(part.Type) + ":"
		
		switch part.Type {
		case ldomain.ContentTypeText:
			key += part.Text
		case ldomain.ContentTypeImage:
			if part.Image != nil {
				if part.Image.Source.Type == ldomain.SourceTypeURL {
					key += part.Image.Source.URL
				} else {
					// Use first 20 chars of base64 data as fingerprint
					data := part.Image.Source.Data
					if len(data) > 20 {
						data = data[:20]
					}
					key += data
				}
			}
		case ldomain.ContentTypeFile:
			if part.File != nil {
				key += part.File.FileName
			}
		case ldomain.ContentTypeVideo, ldomain.ContentTypeAudio:
			key += string(part.Type) + "_content"
		}
		
		if i < len(contentParts)-1 {
			key += "|"
		}
	}
	
	return key
}

// cloneMessage creates a deep copy of a message
func (m *MessageManager) cloneMessage(msg ldomain.Message) ldomain.Message {
	// Clone the ContentParts array
	clonedContent := make([]ldomain.ContentPart, len(msg.Content))
	
	for i, part := range msg.Content {
		clonedPart := ldomain.ContentPart{
			Type: part.Type,
			Text: part.Text,
		}
		
		// Deep copy any media content
		if part.Image != nil {
			imgCopy := *part.Image // Copy the struct
			clonedPart.Image = &imgCopy
		}
		
		if part.File != nil {
			fileCopy := *part.File // Copy the struct
			clonedPart.File = &fileCopy
		}
		
		if part.Video != nil {
			videoCopy := *part.Video // Copy the struct
			clonedPart.Video = &videoCopy
		}
		
		if part.Audio != nil {
			audioCopy := *part.Audio // Copy the struct
			clonedPart.Audio = &audioCopy
		}
		
		clonedContent[i] = clonedPart
	}
	
	return ldomain.Message{
		Role:    msg.Role,
		Content: clonedContent,
	}
}

// truncateMessage truncates a message to fit within token limit
func (m *MessageManager) truncateMessage(msg ldomain.Message, tokenLimit int) ldomain.Message {
	// Calculate how many tokens we currently have
	currentTokens := m.estimateTokensFromContentParts(msg.Content)

	// If already under limit, return as is
	if currentTokens <= tokenLimit {
		return msg
	}

	// Estimate how many characters to keep (4 chars ~= 1 token)
	charsToKeep := (tokenLimit - 5) * 4 // Subtract overhead
	if charsToKeep < 0 {
		charsToKeep = 0
	}

	// Create a cloned message with truncated content
	result := m.cloneMessage(msg)
	
	// Only truncate text content parts
	if len(result.Content) > 0 {
		for i, part := range result.Content {
			if part.Type == ldomain.ContentTypeText {
				// Truncate the text part
				text := part.Text
				if len(text) > charsToKeep {
					// Try to truncate at a space to avoid cutting words
					if charsToKeep > 20 {
						// Look for a space near the truncation point
						for j := charsToKeep - 1; j > charsToKeep-20 && j >= 0; j-- {
							if text[j] == ' ' {
								charsToKeep = j
								break
							}
						}
					}
					
					// Apply truncation
					text = text[:charsToKeep] + "..."
					result.Content[i].Text = text
				}
			}
			// For non-text parts, we leave them as is for now
			// In a more sophisticated implementation, we might remove them to save tokens
		}
	}
	
	return result
}

// sortMessagesInOrder sorts messages to maintain the correct conversation order
func (m *MessageManager) sortMessagesInOrder(msgs []ldomain.Message) []ldomain.Message {
	// System messages go first, then user/assistant messages in chronological order
	var systemMsgs []ldomain.Message
	var nonSystemMsgs []ldomain.Message

	for _, msg := range msgs {
		if msg.Role == ldomain.RoleSystem {
			systemMsgs = append(systemMsgs, msg)
		} else {
			nonSystemMsgs = append(nonSystemMsgs, msg)
		}
	}

	// Combine the messages in the right order
	result := make([]ldomain.Message, 0, len(msgs))
	result = append(result, systemMsgs...)
	result = append(result, nonSystemMsgs...)

	return result
}
