// ABOUTME: Provides built-in state transformation functions for common operations
// ABOUTME: Includes filtering, mapping, validation, and key manipulation transforms

package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Built-in state transformation functions

// FilterTransform creates a transform that removes keys matching a glob pattern.
// The pattern follows filepath.Match syntax (e.g., "temp_*" removes all keys starting with "temp_").
// Messages and artifacts are preserved in the transformed state.
func FilterTransform(pattern string) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := state.Clone()
		for key := range state.Values() {
			if matched, _ := filepath.Match(pattern, key); matched {
				result.Delete(key)
			}
		}
		return result, nil
	}
}

// MapTransform creates a transform that applies a function to all state values.
// The provided function is called for each key-value pair in the state.
// Messages and artifacts are preserved unchanged.
func MapTransform(fn func(interface{}) interface{}) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := state.Clone()
		for key, value := range state.Values() {
			result.Set(key, fn(value))
		}
		return result, nil
	}
}

// ValidateTransform creates a transform that validates the state against a schema.
// If validation fails, the transform returns an error and the state is unchanged.
// This is useful for ensuring state conforms to expected structure before processing.
func ValidateTransform(validator sdomain.Validator, schema *sdomain.Schema) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Use the SchemaValidator from domain
		sv := domain.SchemaValidator(validator, schema)
		if err := sv.Validate(state); err != nil {
			return nil, fmt.Errorf("state validation failed: %w", err)
		}
		return state, nil
	}
}

// PrefixKeysTransform creates a transform that adds a prefix to all state keys.
// For example, with prefix "agent_", key "status" becomes "agent_status".
// Messages and artifacts are preserved unchanged.
func PrefixKeysTransform(prefix string) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()
		for key, value := range state.Values() {
			result.Set(prefix+key, value)
		}
		// Copy messages
		for _, msg := range state.Messages() {
			result.AddMessage(msg)
		}
		// Copy artifacts
		for _, artifact := range state.Artifacts() {
			result.AddArtifact(artifact)
		}
		return result, nil
	}
}

// SelectKeysTransform creates a transform that keeps only specified keys.
// All other keys are removed from the state. This is useful for filtering
// state to only relevant data. Messages and artifacts are preserved.
func SelectKeysTransform(keys ...string) StateTransform {
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()
		for key, value := range state.Values() {
			if keySet[key] {
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
		return result, nil
	}
}

// RenameKeysTransform creates a transform that renames keys based on a mapping.
// Keys not in the mapping are left unchanged. If a target key already exists,
// it will be overwritten. Messages and artifacts are preserved.
func RenameKeysTransform(mapping map[string]string) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := state.Clone()

		// Rename keys
		for oldKey, newKey := range mapping {
			if value, ok := state.Get(oldKey); ok {
				result.Delete(oldKey)
				result.Set(newKey, value)
			}
		}

		return result, nil
	}
}

// MergeTransform creates a transform that merges another state into the current state.
// Values from the other state override values in the current state for matching keys.
// Messages and artifacts from both states are combined.
func MergeTransform(other *domain.State) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := state.Clone()
		result.Merge(other)
		return result, nil
	}
}

// ClearMessagesTransform creates a transform that removes all messages from the state.
// Values and artifacts are preserved. This is useful when message history
// needs to be reset while maintaining other state data.
func ClearMessagesTransform() StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()

		// Copy values
		for key, value := range state.Values() {
			result.Set(key, value)
		}

		// Copy artifacts but not messages
		for _, artifact := range state.Artifacts() {
			result.AddArtifact(artifact)
		}

		return result, nil
	}
}

// LimitMessagesTransform creates a transform that keeps only the last N messages.
// This is useful for maintaining a sliding window of conversation history
// to prevent unbounded growth. Values and artifacts are preserved.
func LimitMessagesTransform(n int) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()

		// Copy values
		for key, value := range state.Values() {
			result.Set(key, value)
		}

		// Copy artifacts
		for _, artifact := range state.Artifacts() {
			result.AddArtifact(artifact)
		}

		// Copy only last N messages
		messages := state.Messages()
		start := 0
		if len(messages) > n {
			start = len(messages) - n
		}
		for i := start; i < len(messages); i++ {
			result.AddMessage(messages[i])
		}

		return result, nil
	}
}

// TransformValues creates a transform that applies specific transformations to selected keys.
// Each key in the map has its own transformation function. Keys not in the map
// are left unchanged. Messages and artifacts are preserved.
func TransformValues(keyTransforms map[string]func(interface{}) interface{}) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := state.Clone()

		for key, transform := range keyTransforms {
			if value, ok := state.Get(key); ok {
				result.Set(key, transform(value))
			}
		}

		return result, nil
	}
}

// ConditionalTransform creates a transform that applies different transforms based on a condition.
// If the condition returns true, thenTransform is applied; otherwise elseTransform is applied.
// If either transform is nil, the state is returned unchanged for that branch.
func ConditionalTransform(condition func(*domain.State) bool, thenTransform, elseTransform StateTransform) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		if condition(state) {
			if thenTransform != nil {
				return thenTransform(ctx, state)
			}
		} else {
			if elseTransform != nil {
				return elseTransform(ctx, state)
			}
		}
		return state, nil
	}
}

// ChainTransforms creates a transform that applies multiple transforms in sequence.
// Each transform receives the output of the previous transform. If any transform
// returns an error, the chain stops and returns that error.
func ChainTransforms(transforms ...StateTransform) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		current := state
		for i, transform := range transforms {
			result, err := transform(ctx, current)
			if err != nil {
				return nil, fmt.Errorf("transform %d failed: %w", i, err)
			}
			current = result
		}
		return current, nil
	}
}

// FilterMessagesByRole creates a transform that keeps only messages with specific roles.
// Messages with roles not in the provided list are removed. This is useful for
// extracting only user messages or only assistant messages. Values and artifacts are preserved.
func FilterMessagesByRole(roles ...string) StateTransform {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()

		// Copy values
		for key, value := range state.Values() {
			result.Set(key, value)
		}

		// Copy artifacts
		for _, artifact := range state.Artifacts() {
			result.AddArtifact(artifact)
		}

		// Filter messages by role
		for _, msg := range state.Messages() {
			if roleSet[string(msg.Role)] {
				result.AddMessage(msg)
			}
		}

		return result, nil
	}
}

// NormalizeKeysTransform creates a transform that normalizes all key names.
// Keys are converted to lowercase and spaces/dashes are replaced with underscores.
// This ensures consistent key naming across different sources.
func NormalizeKeysTransform() StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()

		// Normalize and copy values
		for key, value := range state.Values() {
			// Normalize: lowercase and replace spaces/dashes with underscores
			normalized := strings.ToLower(key)
			normalized = strings.ReplaceAll(normalized, " ", "_")
			normalized = strings.ReplaceAll(normalized, "-", "_")
			result.Set(normalized, value)
		}

		// Copy messages
		for _, msg := range state.Messages() {
			result.AddMessage(msg)
		}

		// Copy artifacts
		for _, artifact := range state.Artifacts() {
			result.AddArtifact(artifact)
		}

		return result, nil
	}
}

// FlattenTransform creates a transform that flattens nested structures using a separator.
// Nested maps become dot-notation keys (e.g., {"user": {"name": "John"}} becomes {"user.name": "John"}).
// Arrays are flattened with index notation (e.g., {"items[0]": "value"}).
func FlattenTransform(separator string) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()

		// Flatten values
		for key, value := range state.Values() {
			flattenStateValue(result, key, value, separator)
		}

		// Copy messages
		for _, msg := range state.Messages() {
			result.AddMessage(msg)
		}

		// Copy artifacts
		for _, artifact := range state.Artifacts() {
			result.AddArtifact(artifact)
		}

		return result, nil
	}
}

// flattenStateValue recursively flattens a value into the state.
// It handles nested maps and arrays, creating flattened keys with the specified separator.
// Base types are stored directly with their prefix as the key.
func flattenStateValue(state *domain.State, prefix string, value interface{}, separator string) {
	switch v := value.(type) {
	case map[string]interface{}:
		for k, val := range v {
			var newKey string
			if prefix != "" {
				newKey = prefix + separator + k
			} else {
				newKey = k
			}
			flattenStateValue(state, newKey, val, separator)
		}
	case []interface{}:
		for i, val := range v {
			newKey := fmt.Sprintf("%s%s%d", prefix, separator, i)
			flattenStateValue(state, newKey, val, separator)
		}
	default:
		state.Set(prefix, value)
	}
}
