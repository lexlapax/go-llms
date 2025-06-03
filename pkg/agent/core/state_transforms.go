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

// FilterTransform removes keys matching pattern
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

// MapTransform applies function to all values
func MapTransform(fn func(interface{}) interface{}) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := state.Clone()
		for key, value := range state.Values() {
			result.Set(key, fn(value))
		}
		return result, nil
	}
}

// ValidateTransform ensures state matches schema
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

// PrefixKeysTransform adds prefix to all keys
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

// SelectKeysTransform keeps only specified keys
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

// RenameKeysTransform renames keys based on mapping
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

// MergeTransform merges another state into the current state
func MergeTransform(other *domain.State) StateTransform {
	return func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := state.Clone()
		result.Merge(other)
		return result, nil
	}
}

// ClearMessagesTransform removes all messages
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

// LimitMessagesTransform keeps only the last N messages
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

// TransformValues applies a transform to specific keys
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

// ConditionalTransform applies a transform based on a condition
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

// ChainTransforms applies multiple transforms in sequence
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

// FilterMessagesByRole keeps only messages with specific roles
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

// NormalizeKeysTransform normalizes key names (e.g., lowercase, replace spaces)
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

// FlattenTransform flattens nested structures with dot notation
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

// flattenStateValue recursively flattens a value
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
