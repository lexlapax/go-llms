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
		maxMessages:      maxMessages,
		maxTokens:        maxTokens,
		messages:         make([]ldomain.Message, 0, maxMessages),
		tokenCounts:      make(map[string]int),
		messageTimestamps: make([]time.Time, 0, maxMessages),
		config:           config,
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
	tokens := m.estimateTokens(newMsg.Content)
	m.tokenCounts[newMsg.Content] = tokens
	
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
		tokens := m.estimateTokens(newMsg.Content)
		m.tokenCounts[newMsg.Content] = tokens
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
		systemTokens += m.estimateTokens(msg.Content)
	}
	
	// Start with all system messages
	result := systemMessages
	remainingTokens := modelTokenLimit - systemTokens
	
	// Now add as many non-system messages as fit, prioritizing recent ones
	// Start from most recent
	for i := len(nonSystemMessages) - 1; i >= 0 && remainingTokens > 0; i-- {
		msg := nonSystemMessages[i]
		tokens := m.estimateTokens(msg.Content)
		
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
		sysMsg := ldomain.Message{
			Role:    ldomain.RoleSystem,
			Content: prompt,
		}
		
		// Insert at the beginning
		newMessages = append([]ldomain.Message{sysMsg}, newMessages...)
		newTimestamps = append([]time.Time{time.Now()}, newTimestamps...)
		
		// Update token count for the new message
		m.tokenCounts[prompt] = m.estimateTokens(prompt)
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
		total += m.estimateTokens(msg.Content)
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
			totalTokens += m.estimateTokens(msg.Content)
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
					totalTokens -= m.estimateTokens(msg.Content)
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
			totalTokens -= m.estimateTokens(m.messages[i].Content)
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

// cloneMessage creates a deep copy of a message
func (m *MessageManager) cloneMessage(msg ldomain.Message) ldomain.Message {
	return ldomain.Message{
		Role:    msg.Role,
		Content: msg.Content,
	}
}

// truncateMessage truncates a message to fit within token limit
func (m *MessageManager) truncateMessage(msg ldomain.Message, tokenLimit int) ldomain.Message {
	// Calculate how many tokens we currently have
	currentTokens := m.estimateTokens(msg.Content)
	
	// If already under limit, return as is
	if currentTokens <= tokenLimit {
		return msg
	}
	
	// Estimate how many characters to keep (4 chars ~= 1 token)
	charsToKeep := (tokenLimit - 5) * 4 // Subtract overhead
	if charsToKeep < 0 {
		charsToKeep = 0
	}
	
	// Truncate the content
	truncatedContent := msg.Content
	if len(truncatedContent) > charsToKeep {
		// Try to truncate at a space to avoid cutting words
		if charsToKeep > 20 {
			// Look for a space near the truncation point
			for i := charsToKeep - 1; i > charsToKeep-20 && i >= 0; i-- {
				if truncatedContent[i] == ' ' {
					charsToKeep = i
					break
				}
			}
		}
		
		truncatedContent = truncatedContent[:charsToKeep] + "..."
	}
	
	return ldomain.Message{
		Role:    msg.Role,
		Content: truncatedContent,
	}
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