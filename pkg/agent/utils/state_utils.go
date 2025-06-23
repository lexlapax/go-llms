// ABOUTME: Provides utility functions for state manipulation and validation
// ABOUTME: Includes helpers for state comparison, extraction, and transformation

package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ExtractValue extracts a value from state using a path.
// The path supports dot notation for nested access and array indexing.
//
// Path format examples:
//   - "key" - simple key access
//   - "key.nested.value" - nested object access
//   - "key[0].nested" - array index with nested access
//   - "items[2].name" - access name field of third array item
//
// Parameters:
//   - state: The state to extract from
//   - path: The extraction path
//
// Returns the extracted value or an error if the path is invalid.
func ExtractValue(state *domain.State, path string) (interface{}, error) {
	if state == nil {
		return nil, fmt.Errorf("state is nil")
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	var current interface{} = state.Values()

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part.key]
			if !ok {
				return nil, fmt.Errorf("key '%s' not found at path '%s'", part.key, strings.Join(getPathKeys(parts[:i+1]), "."))
			}
			current = val

		case []interface{}:
			if part.isIndex {
				if part.index < 0 || part.index >= len(v) {
					return nil, fmt.Errorf("index %d out of bounds at path '%s'", part.index, strings.Join(getPathKeys(parts[:i+1]), "."))
				}
				current = v[part.index]
			} else {
				return nil, fmt.Errorf("cannot access key '%s' on array at path '%s'", part.key, strings.Join(getPathKeys(parts[:i]), "."))
			}

		default:
			if i < len(parts)-1 {
				return nil, fmt.Errorf("cannot traverse further at path '%s'", strings.Join(getPathKeys(parts[:i+1]), "."))
			}
			return current, nil
		}
	}

	return current, nil
}

// SetValue sets a value in state using a path.
// It creates intermediate structures as needed to support the path.
//
// Parameters:
//   - state: The state to modify
//   - path: The path where to set the value
//   - value: The value to set
//
// Returns an error if the path is invalid or cannot be created.
func SetValue(state *domain.State, path string, value interface{}) error {
	if state == nil {
		return fmt.Errorf("state is nil")
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	// If it's a simple path, set directly
	if len(parts) == 1 && !parts[0].isIndex {
		state.Set(parts[0].key, value)
		return nil
	}

	// For complex paths, we need to build/traverse the structure
	values := state.Values()
	current := ensureStructure(values, parts[:len(parts)-1])

	lastPart := parts[len(parts)-1]
	switch v := current.(type) {
	case map[string]interface{}:
		v[lastPart.key] = value
	case []interface{}:
		if lastPart.isIndex {
			if lastPart.index >= 0 && lastPart.index < len(v) {
				v[lastPart.index] = value
			} else {
				return fmt.Errorf("index %d out of bounds", lastPart.index)
			}
		} else {
			return fmt.Errorf("cannot set key on array")
		}
	default:
		return fmt.Errorf("cannot set value on %T", current)
	}

	// Update the state with modified values
	for k, v := range values {
		state.Set(k, v)
	}

	return nil
}

// CompareStates compares two states and returns differences.
// It identifies added, removed, and modified values between states.
//
// Parameters:
//   - state1: The first state (baseline)
//   - state2: The second state (comparison)
//
// Returns a StateDiff containing all differences.
func CompareStates(state1, state2 *domain.State) StateDiff {
	diff := StateDiff{
		Added:    make(map[string]interface{}),
		Removed:  make(map[string]interface{}),
		Modified: make(map[string]ValueChange),
	}

	if state1 == nil && state2 == nil {
		return diff
	}

	if state1 == nil {
		diff.Added = state2.Values()
		return diff
	}

	if state2 == nil {
		diff.Removed = state1.Values()
		return diff
	}

	values1 := state1.Values()
	values2 := state2.Values()

	// Check for removed and modified values
	for k, v1 := range values1 {
		v2, exists := values2[k]
		if !exists {
			diff.Removed[k] = v1
		} else if !deepEqual(v1, v2) {
			diff.Modified[k] = ValueChange{
				Old: v1,
				New: v2,
			}
		}
	}

	// Check for added values
	for k, v2 := range values2 {
		if _, exists := values1[k]; !exists {
			diff.Added[k] = v2
		}
	}

	return diff
}

// StateDiff represents differences between two states.
// It categorizes changes into added, removed, and modified values.
type StateDiff struct {
	Added    map[string]interface{}
	Removed  map[string]interface{}
	Modified map[string]ValueChange
}

// ValueChange represents a changed value between states.
// It captures both the old and new values for comparison.
type ValueChange struct {
	Old interface{}
	New interface{}
}

// IsEmpty returns true if there are no differences.
// This is useful for checking if two states are identical.
func (d StateDiff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Modified) == 0
}

// ValidateState validates state against common rules.
// It checks for circular references and size limits to ensure
// the state is safe to process and store.
//
// Parameters:
//   - state: The state to validate
//
// Returns an error if validation fails.
func ValidateState(state *domain.State) error {
	if state == nil {
		return fmt.Errorf("state is nil")
	}

	values := state.Values()

	// Check for circular references
	if err := checkCircularReferences(values, make(map[interface{}]bool)); err != nil {
		return fmt.Errorf("circular reference detected: %w", err)
	}

	// Check state size
	size := estimateSize(values)
	maxSize := int64(10 * 1024 * 1024) // 10MB default
	if size > maxSize {
		return fmt.Errorf("state size %d exceeds maximum %d", size, maxSize)
	}

	return nil
}

// CopyValues creates a deep copy of values.
// It uses JSON marshaling for reliable deep copying,
// with a fallback to shallow copy if marshaling fails.
//
// Parameters:
//   - values: The map to copy
//
// Returns a deep copy of the values.
func CopyValues(values map[string]interface{}) map[string]interface{} {
	if values == nil {
		return nil
	}

	// Use JSON marshal/unmarshal for deep copy
	data, err := json.Marshal(values)
	if err != nil {
		// Fallback to shallow copy
		result := make(map[string]interface{})
		for k, v := range values {
			result[k] = v
		}
		return result
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		// Fallback to shallow copy
		result = make(map[string]interface{})
		for k, v := range values {
			result[k] = v
		}
	}

	return result
}

// FilterState creates a new state with only specified keys.
// Artifacts and messages are preserved in the filtered state.
//
// Parameters:
//   - state: The state to filter
//   - keys: The keys to include in the filtered state
//
// Returns a new state containing only the specified keys.
func FilterState(state *domain.State, keys []string) *domain.State {
	if state == nil || len(keys) == 0 {
		return domain.NewState()
	}

	filtered := domain.NewState()
	values := state.Values()

	for _, key := range keys {
		if value, ok := values[key]; ok {
			filtered.Set(key, value)
		}
	}

	// Copy artifacts and messages
	for _, artifact := range state.Artifacts() {
		filtered.AddArtifact(artifact)
	}

	for _, msg := range state.Messages() {
		filtered.AddMessage(msg)
	}

	return filtered
}

// TransformState applies a transformation function to all values.
// This is useful for data conversion, sanitization, or enrichment.
// Artifacts and messages are preserved in the transformed state.
//
// Parameters:
//   - state: The state to transform
//   - transform: Function to apply to each key-value pair
//
// Returns a new transformed state or an error if transformation fails.
func TransformState(state *domain.State, transform func(key string, value interface{}) (interface{}, error)) (*domain.State, error) {
	if state == nil {
		return nil, fmt.Errorf("state is nil")
	}

	transformed := domain.NewState()

	for k, v := range state.Values() {
		newValue, err := transform(k, v)
		if err != nil {
			return nil, fmt.Errorf("transform failed for key '%s': %w", k, err)
		}
		transformed.Set(k, newValue)
	}

	// Copy artifacts and messages
	for _, artifact := range state.Artifacts() {
		transformed.AddArtifact(artifact)
	}

	for _, msg := range state.Messages() {
		transformed.AddMessage(msg)
	}

	return transformed, nil
}

// pathPart represents a single component of a path.
// It can be either a key name or an array index.
type pathPart struct {
	key     string
	isIndex bool
	index   int
}

// parsePath parses a path string into path components.
// It handles dot notation and array indexing.
func parsePath(path string) []pathPart {
	var parts []pathPart
	current := ""

	for i := 0; i < len(path); i++ {
		ch := path[i]

		switch ch {
		case '.':
			if current != "" {
				parts = append(parts, pathPart{key: current})
				current = ""
			}
		case '[':
			if current != "" {
				parts = append(parts, pathPart{key: current})
				current = ""
			}
			// Parse index
			j := i + 1
			for j < len(path) && path[j] != ']' {
				j++
			}
			if j < len(path) {
				indexStr := path[i+1 : j]
				index := 0
				_, _ = fmt.Sscanf(indexStr, "%d", &index)
				parts = append(parts, pathPart{isIndex: true, index: index})
				i = j
			}
		default:
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, pathPart{key: current})
	}

	return parts
}

// getPathKeys converts path parts back to string representations.
// Array indices are formatted as "[n]".
func getPathKeys(parts []pathPart) []string {
	keys := make([]string, len(parts))
	for i, p := range parts {
		if p.isIndex {
			keys[i] = fmt.Sprintf("[%d]", p.index)
		} else {
			keys[i] = p.key
		}
	}
	return keys
}

// ensureStructure creates nested structures as needed to support a path.
// It creates maps for object paths and arrays for index paths.
func ensureStructure(data map[string]interface{}, parts []pathPart) interface{} {
	if len(parts) == 0 {
		return data
	}

	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if part.isIndex {
				// Need to create array
				arr := make([]interface{}, part.index+1)
				v[part.key] = arr
				current = arr
			} else {
				if _, ok := v[part.key]; !ok {
					v[part.key] = make(map[string]interface{})
				}
				current = v[part.key]
			}
		case []interface{}:
			if part.isIndex {
				// Expand array if needed
				for len(v) <= part.index {
					v = append(v, nil)
				}
				if v[part.index] == nil {
					v[part.index] = make(map[string]interface{})
				}
				current = v[part.index]
			}
		}
	}

	return current
}

// deepEqual performs deep equality comparison of two values.
// It uses JSON comparison for complex types with reflection as fallback.
func deepEqual(a, b interface{}) bool {
	// Use JSON comparison for simplicity
	aJSON, err1 := json.Marshal(a)
	bJSON, err2 := json.Marshal(b)

	if err1 != nil || err2 != nil {
		return reflect.DeepEqual(a, b)
	}

	return string(aJSON) == string(bJSON)
}

// checkCircularReferences detects circular references in data structures.
// It tracks visited objects to identify cycles.
func checkCircularReferences(value interface{}, visited map[interface{}]bool) error {
	switch v := value.(type) {
	case map[string]interface{}:
		if visited[v] {
			return fmt.Errorf("circular reference in map")
		}
		visited[v] = true
		for _, val := range v {
			if err := checkCircularReferences(val, visited); err != nil {
				return err
			}
		}
		delete(visited, v)
	case []interface{}:
		if visited[v] {
			return fmt.Errorf("circular reference in array")
		}
		visited[v] = true
		for _, val := range v {
			if err := checkCircularReferences(val, visited); err != nil {
				return err
			}
		}
		delete(visited, v)
	}
	return nil
}

// estimateSize estimates the size of a value in bytes.
// It uses JSON encoding as a size approximation.
func estimateSize(value interface{}) int64 {
	// Simple size estimation using JSON
	data, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return int64(len(data))
}
