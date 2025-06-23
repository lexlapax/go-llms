// ABOUTME: Provides utility functions for common state transformations and operations
// ABOUTME: Includes message filtering, metadata manipulation, and state updates

package core

import (
	"context"
	"sort"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// FilterMessages creates a new state with only messages that match the predicate.
// The predicate function is called for each message; only messages returning true are kept.
// Values and artifacts are copied unchanged to the new state.
func FilterMessages(state *domain.State, predicate func(domain.Message) bool) *domain.State {
	result := domain.NewState()

	// Copy values
	for key, value := range state.Values() {
		result.Set(key, value)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	// Filter messages
	for _, msg := range state.Messages() {
		if predicate(msg) {
			result.AddMessage(msg)
		}
	}

	return result
}

// MapMessages creates a new state with transformed messages.
// The mapper function is applied to each message to produce a new message.
// Values and artifacts are copied unchanged to the new state.
func MapMessages(state *domain.State, mapper func(domain.Message) domain.Message) *domain.State {
	result := domain.NewState()

	// Copy values
	for key, value := range state.Values() {
		result.Set(key, value)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	// Map messages
	for _, msg := range state.Messages() {
		result.AddMessage(mapper(msg))
	}

	return result
}

// FilterMetadata creates a new state with only metadata that matches the predicate.
// The predicate is called for each key-value pair; only pairs returning true are kept.
// Messages and artifacts are copied unchanged to the new state.
func FilterMetadata(state *domain.State, predicate func(key string, value interface{}) bool) *domain.State {
	result := domain.NewState()

	// Copy only values that match the predicate
	for key, value := range state.Values() {
		if predicate(key, value) {
			result.Set(key, value)
		}
	}

	// Copy messages
	for _, msg := range state.Messages() {
		result.AddMessage(msg)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	return result
}

// MapMetadata creates a new state with transformed metadata values.
// The mapper function is applied to each key-value pair to produce new values.
// Messages and artifacts are preserved in the cloned state.
func MapMetadata(state *domain.State, mapper func(key string, value interface{}) interface{}) *domain.State {
	result := state.Clone()

	// Map all metadata values
	for key, value := range state.Values() {
		result.Set(key, mapper(key, value))
	}

	return result
}

// UpdateMetadata creates a new state with updated metadata values.
// Values in the updates map override existing values with the same keys.
// Messages and artifacts are preserved in the cloned state.
func UpdateMetadata(state *domain.State, updates map[string]interface{}) *domain.State {
	result := state.Clone()

	// Apply updates
	for key, value := range updates {
		result.Set(key, value)
	}

	return result
}

// RemoveMetadataKeys creates a new state with specified keys removed.
// The listed keys are deleted from the state's metadata.
// Messages and artifacts are preserved in the cloned state.
func RemoveMetadataKeys(state *domain.State, keys ...string) *domain.State {
	result := state.Clone()

	// Remove specified keys
	for _, key := range keys {
		result.Delete(key)
	}

	return result
}

// TruncateMessages creates a new state keeping only the last N messages.
// If the state has fewer than maxMessages, all messages are kept.
// Values and artifacts are copied unchanged to the new state.
func TruncateMessages(state *domain.State, maxMessages int) *domain.State {
	result := domain.NewState()

	// Copy values
	for key, value := range state.Values() {
		result.Set(key, value)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	// Keep only last N messages
	messages := state.Messages()
	start := 0
	if len(messages) > maxMessages {
		start = len(messages) - maxMessages
	}

	for i := start; i < len(messages); i++ {
		result.AddMessage(messages[i])
	}

	return result
}

// SortMessages creates a new state with messages sorted by the given comparison function.
// The less function should return true if message i should come before message j.
// Values and artifacts are copied unchanged to the new state.
func SortMessages(state *domain.State, less func(i, j domain.Message) bool) *domain.State {
	result := domain.NewState()

	// Copy values
	for key, value := range state.Values() {
		result.Set(key, value)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	// Sort messages
	messages := state.Messages()
	sorted := make([]domain.Message, len(messages))
	copy(sorted, messages)

	sort.Slice(sorted, func(i, j int) bool {
		return less(sorted[i], sorted[j])
	})

	for _, msg := range sorted {
		result.AddMessage(msg)
	}

	return result
}

// MergeStates creates a new state by merging multiple states.
// States are merged in order, with later states overriding values from earlier ones.
// Messages and artifacts from all states are combined.
func MergeStates(states ...*domain.State) *domain.State {
	result := domain.NewState()

	// Merge each state in order
	for _, state := range states {
		if state != nil {
			result.Merge(state)
		}
	}

	return result
}

// CloneWithMessages creates a new state with the same metadata but different messages.
// All values and artifacts are copied from the original state, but messages are replaced
// with the provided list.
func CloneWithMessages(state *domain.State, messages []domain.Message) *domain.State {
	result := domain.NewState()

	// Copy values
	for key, value := range state.Values() {
		result.Set(key, value)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	// Add new messages
	for _, msg := range messages {
		result.AddMessage(msg)
	}

	return result
}

// CloneWithMetadata creates a new state with different metadata but same messages.
// All messages and artifacts are copied from the original state, but values are replaced
// with the provided metadata map.
func CloneWithMetadata(state *domain.State, metadata map[string]interface{}) *domain.State {
	result := domain.NewState()

	// Set new metadata
	for key, value := range metadata {
		result.Set(key, value)
	}

	// Copy messages
	for _, msg := range state.Messages() {
		result.AddMessage(msg)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	return result
}

// GroupMessagesByRole groups messages by their role.
// Returns a map where keys are roles and values are slices of messages with that role.
// Messages appear in the same order as in the original state.
func GroupMessagesByRole(state *domain.State) map[domain.Role][]domain.Message {
	groups := make(map[domain.Role][]domain.Message)

	for _, msg := range state.Messages() {
		groups[msg.Role] = append(groups[msg.Role], msg)
	}

	return groups
}

// CountMessagesByRole counts messages by their role.
// Returns a map where keys are roles and values are the count of messages with that role.
// Useful for analyzing conversation balance and patterns.
func CountMessagesByRole(state *domain.State) map[domain.Role]int {
	counts := make(map[domain.Role]int)

	for _, msg := range state.Messages() {
		counts[msg.Role]++
	}

	return counts
}

// GetLatestMessageByRole gets the most recent message for a given role.
// Returns the last message with the specified role and true if found,
// or an empty message and false if no message with that role exists.
func GetLatestMessageByRole(state *domain.State, role domain.Role) (domain.Message, bool) {
	messages := state.Messages()

	// Iterate backwards to find the latest
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == role {
			return messages[i], true
		}
	}

	return domain.Message{}, false
}

// GetMessagesSince returns all messages after the given timestamp.
// Messages are returned in their original order. Useful for retrieving
// recent conversation history or changes since a checkpoint.
func GetMessagesSince(state *domain.State, since time.Time) []domain.Message {
	var result []domain.Message

	for _, msg := range state.Messages() {
		if msg.Timestamp.After(since) {
			result = append(result, msg)
		}
	}

	return result
}

// HasMetadataKey checks if a metadata key exists.
// Returns true if the key is present in the state's values,
// regardless of the associated value (even if nil).
func HasMetadataKey(state *domain.State, key string) bool {
	_, exists := state.Get(key)
	return exists
}

// GetMetadataKeys returns all metadata keys.
// The returned slice contains all keys present in the state's values.
// The order of keys is not guaranteed to be consistent.
func GetMetadataKeys(state *domain.State) []string {
	values := state.Values()
	keys := make([]string, 0, len(values))

	for key := range values {
		keys = append(keys, key)
	}

	return keys
}

// ClearMessages creates a new state with no messages but same metadata.
// All values and artifacts are preserved, only messages are removed.
// Useful for resetting conversation history while keeping context.
func ClearMessages(state *domain.State) *domain.State {
	result := domain.NewState()

	// Copy values
	for key, value := range state.Values() {
		result.Set(key, value)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	// No messages added

	return result
}

// ClearMetadata creates a new state with no metadata but same messages.
// All messages and artifacts are preserved, only values are cleared.
// Useful for removing context while keeping conversation history.
func ClearMetadata(state *domain.State) *domain.State {
	result := domain.NewState()

	// Copy messages
	for _, msg := range state.Messages() {
		result.AddMessage(msg)
	}

	// Copy artifacts
	for _, artifact := range state.Artifacts() {
		result.AddArtifact(artifact)
	}

	// No metadata added

	return result
}

// ConvertToStateTransform converts a utility function to a StateTransform.
// This adapter allows simple state transformation functions to be used
// where a StateTransform is required, always returning nil error.
func ConvertToStateTransform(fn func(*domain.State) *domain.State) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		return fn(state), nil
	}
}

// AddMessagePrefix adds a prefix to all message contents.
// The prefix is prepended to the content of each message.
// Values and artifacts are preserved unchanged.
func AddMessagePrefix(state *domain.State, prefix string) *domain.State {
	return MapMessages(state, func(msg domain.Message) domain.Message {
		newMsg := msg
		newMsg.Content = prefix + msg.Content
		return newMsg
	})
}

// AddMessageSuffix adds a suffix to all message contents.
// The suffix is appended to the content of each message.
// Values and artifacts are preserved unchanged.
func AddMessageSuffix(state *domain.State, suffix string) *domain.State {
	return MapMessages(state, func(msg domain.Message) domain.Message {
		newMsg := msg
		newMsg.Content = msg.Content + suffix
		return newMsg
	})
}

// CloneWithModifications creates a clone and applies modifications to it.
// The modifier function receives the cloned state and can modify it in-place.
// This ensures the original state remains unchanged.
func CloneWithModifications(state *domain.State, modifier func(*domain.State)) *domain.State {
	clone := state.Clone()
	modifier(clone)
	return clone
}

// ChainUtilityTransforms chains multiple state transformation utility functions.
// Transforms are applied in order, with each receiving the output of the previous.
// Returns a single function that applies all transforms sequentially.
func ChainUtilityTransforms(transforms ...func(*domain.State) *domain.State) func(*domain.State) *domain.State {
	return func(state *domain.State) *domain.State {
		current := state
		for _, transform := range transforms {
			current = transform(current)
		}
		return current
	}
}

// WithTimestamp adds a timestamp metadata to the state.
// Adds a "timestamp" key with the current time as the value.
// Useful for tracking when state snapshots were created.
func WithTimestamp(state *domain.State) *domain.State {
	result := state.Clone()
	result.Set("timestamp", time.Now())
	return result
}

// WithMessageCount adds a message count metadata to the state.
// Adds a "message_count" key with the number of messages as the value.
// Useful for tracking conversation length without counting messages.
func WithMessageCount(state *domain.State) *domain.State {
	result := state.Clone()
	result.Set("message_count", len(state.Messages()))
	return result
}

// WithID adds an ID metadata to the state.
// Adds an "id" key with the provided ID as the value.
// Useful for tracking and identifying specific state instances.
func WithID(state *domain.State, id string) *domain.State {
	result := state.Clone()
	result.Set("id", id)
	return result
}

// SetMetadataValue creates a function that sets a specific metadata value.
// Returns a transformation function that adds or updates the specified key-value pair.
// This is useful for creating reusable state modifiers.
func SetMetadataValue(key string, value interface{}) func(*domain.State) *domain.State {
	return func(state *domain.State) *domain.State {
		result := state.Clone()
		result.Set(key, value)
		return result
	}
}

// AppendMessage creates a function that appends a message to the state.
// Returns a transformation function that adds the message to the state's message list.
// This is useful for creating reusable message appenders.
func AppendMessage(msg domain.Message) func(*domain.State) *domain.State {
	return func(state *domain.State) *domain.State {
		result := state.Clone()
		result.AddMessage(msg)
		return result
	}
}

// ConditionalUtilityTransform creates a conditional transformation function.
// If the condition returns true, applies thenTransform; otherwise returns state unchanged.
// This is useful for creating transforms that only apply under certain conditions.
func ConditionalUtilityTransform(condition func(*domain.State) bool, thenTransform func(*domain.State) *domain.State) func(*domain.State) *domain.State {
	return func(state *domain.State) *domain.State {
		if condition(state) {
			return thenTransform(state)
		}
		return state
	}
}
