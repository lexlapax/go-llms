// ABOUTME: Tests for StateManager including state lifecycle, transformations, and merging
// ABOUTME: Validates state persistence, transforms, and merge strategies

package core_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestStateManagerBasicOperations(t *testing.T) {
	sm := core.NewStateManager()

	// Create and save state
	state := domain.NewState()
	state.Set("key1", "value1")
	state.Set("key2", 42)

	err := sm.SaveState(state)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Load state
	loaded, err := sm.LoadState(state.ID())
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify loaded state
	val1, ok := loaded.Get("key1")
	if !ok || val1 != "value1" {
		t.Error("Loaded state has incorrect value for key1")
	}

	val2, ok := loaded.Get("key2")
	if !ok || val2 != 42 {
		t.Error("Loaded state has incorrect value for key2")
	}

	// Verify it's a clone (different ID)
	if loaded.ID() == state.ID() {
		t.Error("Loaded state should be a clone with different ID")
	}

	// Test loading non-existent state
	_, err = sm.LoadState("non-existent")
	if err == nil {
		t.Error("Should error when loading non-existent state")
	}

	// Test saving nil state
	err = sm.SaveState(nil)
	if err == nil {
		t.Error("Should error when saving nil state")
	}
}

func TestStateManagerDelete(t *testing.T) {
	sm := core.NewStateManager()

	// Save state
	state := domain.NewState()
	state.Set("key", "value")
	_ = sm.SaveState(state)

	// Delete state
	err := sm.DeleteState(state.ID())
	if err != nil {
		t.Fatalf("Failed to delete state: %v", err)
	}

	// Try to load deleted state
	_, err = sm.LoadState(state.ID())
	if err == nil {
		t.Error("Should error when loading deleted state")
	}

	// Try to delete non-existent state
	err = sm.DeleteState("non-existent")
	if err == nil {
		t.Error("Should error when deleting non-existent state")
	}
}

func TestStateManagerList(t *testing.T) {
	sm := core.NewStateManager()

	// Save multiple states
	states := make([]*domain.State, 3)
	for i := 0; i < 3; i++ {
		states[i] = domain.NewState()
		states[i].Set("index", i)
		_ = sm.SaveState(states[i])
	}

	// List states
	ids := sm.ListStates()
	if len(ids) != 3 {
		t.Errorf("Expected 3 states, got %d", len(ids))
	}

	// Verify all state IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, state := range states {
		if !idMap[state.ID()] {
			t.Errorf("State %s not found in list", state.ID())
		}
	}
}

func TestStateManagerTransforms(t *testing.T) {
	sm := core.NewStateManager()
	ctx := context.Background()

	// Register a custom transform
	sm.RegisterTransform("double_values", func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()
		for k, v := range state.Values() {
			if num, ok := v.(int); ok {
				result.Set(k, num*2)
			} else {
				result.Set(k, v)
			}
		}
		return result, nil
	})

	// Create state
	state := domain.NewState()
	state.Set("num1", 10)
	state.Set("num2", 20)
	state.Set("str", "hello")

	// Apply transform
	transformed, err := sm.ApplyTransform(ctx, "double_values", state)
	if err != nil {
		t.Fatalf("Failed to apply transform: %v", err)
	}

	// Verify transformation
	val1, _ := transformed.Get("num1")
	if val1 != 20 {
		t.Errorf("Expected 20, got %v", val1)
	}

	val2, _ := transformed.Get("num2")
	if val2 != 40 {
		t.Errorf("Expected 40, got %v", val2)
	}

	val3, _ := transformed.Get("str")
	if val3 != "hello" {
		t.Errorf("String should remain unchanged, got %v", val3)
	}

	// Test non-existent transform
	_, err = sm.ApplyTransform(ctx, "non-existent", state)
	if err == nil {
		t.Error("Should error with non-existent transform")
	}
}

func TestStateManagerBuiltinTransforms(t *testing.T) {
	sm := core.NewStateManager()
	ctx := context.Background()

	t.Run("Filter Transform", func(t *testing.T) {
		state := domain.NewState()
		state.Set("keep1", "value1")
		state.Set("keep2", "value2")
		state.Set("remove1", "value3")
		state.Set("remove2", "value4")

		// Set filter keys in metadata
		state.SetMetadata("filter_keys", []string{"remove1", "remove2"})

		transformed, err := sm.ApplyTransform(ctx, "filter", state)
		if err != nil {
			t.Fatalf("Failed to apply filter transform: %v", err)
		}

		// Verify filtered keys are removed
		if _, ok := transformed.Get("remove1"); ok {
			t.Error("remove1 should be filtered out")
		}
		if _, ok := transformed.Get("remove2"); ok {
			t.Error("remove2 should be filtered out")
		}

		// Verify kept keys remain
		if val, ok := transformed.Get("keep1"); !ok || val != "value1" {
			t.Error("keep1 should remain")
		}
		if val, ok := transformed.Get("keep2"); !ok || val != "value2" {
			t.Error("keep2 should remain")
		}
	})

	t.Run("Flatten Transform", func(t *testing.T) {
		state := domain.NewState()
		state.Set("simple", "value")
		state.Set("nested", map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "deep_value",
			},
			"another": "value2",
		})
		state.Set("array", []interface{}{"a", "b", "c"})

		transformed, err := sm.ApplyTransform(ctx, "flatten", state)
		if err != nil {
			t.Fatalf("Failed to apply flatten transform: %v", err)
		}

		// Check flattened values
		if val, ok := transformed.Get("simple"); !ok || val != "value" {
			t.Error("Simple value should remain")
		}

		if val, ok := transformed.Get("nested.level1.level2"); !ok || val != "deep_value" {
			t.Error("Nested value should be flattened")
		}

		if val, ok := transformed.Get("nested.another"); !ok || val != "value2" {
			t.Error("Another nested value should be flattened")
		}

		if val, ok := transformed.Get("array[0]"); !ok || val != "a" {
			t.Error("Array element 0 should be flattened")
		}
	})

	t.Run("Sanitize Transform", func(t *testing.T) {
		state := domain.NewState()
		state.Set("username", "john")
		state.Set("password", "secret123")
		state.Set("api_key", "abc123")
		state.Set("user_token", "xyz789")
		state.Set("normal_data", "visible")

		transformed, err := sm.ApplyTransform(ctx, "sanitize", state)
		if err != nil {
			t.Fatalf("Failed to apply sanitize transform: %v", err)
		}

		// Check sensitive data is redacted
		if val, _ := transformed.Get("password"); val != "[REDACTED]" {
			t.Errorf("Password should be redacted, got %v", val)
		}

		if val, _ := transformed.Get("api_key"); val != "[REDACTED]" {
			t.Errorf("API key should be redacted, got %v", val)
		}

		if val, _ := transformed.Get("user_token"); val != "[REDACTED]" {
			t.Errorf("Token should be redacted, got %v", val)
		}

		// Check normal data remains
		if val, _ := transformed.Get("username"); val != "john" {
			t.Error("Username should remain unchanged")
		}

		if val, _ := transformed.Get("normal_data"); val != "visible" {
			t.Error("Normal data should remain unchanged")
		}
	})
}

func TestStateManagerValidators(t *testing.T) {
	sm := core.NewStateManager()

	// Register a validator
	sm.RegisterValidator("has_required", func(state *domain.State) error {
		if _, ok := state.Get("required_field"); !ok {
			return errorf("missing required_field")
		}
		return nil
	})

	// Test valid state
	validState := domain.NewState()
	validState.Set("required_field", "value")
	validState.Set("other", "data")

	err := sm.ValidateState("has_required", validState)
	if err != nil {
		t.Errorf("Valid state should pass validation: %v", err)
	}

	// Test invalid state
	invalidState := domain.NewState()
	invalidState.Set("other", "data")

	err = sm.ValidateState("has_required", invalidState)
	if err == nil {
		t.Error("Invalid state should fail validation")
	}

	// Test non-existent validator
	err = sm.ValidateState("non-existent", validState)
	if err == nil {
		t.Error("Should error with non-existent validator")
	}
}

func TestMergeStrategies(t *testing.T) {
	sm := core.NewStateManager()

	// Create test states
	state1 := domain.NewState()
	state1.Set("key1", "value1")
	state1.Set("common", "original")

	state2 := domain.NewState()
	state2.Set("key2", "value2")
	state2.Set("common", "modified")

	state3 := domain.NewState()
	state3.Set("key3", "value3")

	states := []*domain.State{state1, state2, state3}

	t.Run("MergeStrategyLast", func(t *testing.T) {
		result, err := sm.MergeStates(states, domain.MergeStrategyLast)
		if err != nil {
			t.Fatalf("Failed to merge: %v", err)
		}

		// Should only have values from last state
		if _, ok := result.Get("key1"); ok {
			t.Error("Should not have key1 from first state")
		}

		if _, ok := result.Get("key2"); ok {
			t.Error("Should not have key2 from second state")
		}

		if val, ok := result.Get("key3"); !ok || val != "value3" {
			t.Error("Should have key3 from last state")
		}
	})

	t.Run("MergeStrategyMergeAll", func(t *testing.T) {
		result, err := sm.MergeStates(states, domain.MergeStrategyMergeAll)
		if err != nil {
			t.Fatalf("Failed to merge: %v", err)
		}

		// Should have all keys, with later values overriding
		if val, ok := result.Get("key1"); !ok || val != "value1" {
			t.Error("Should have key1")
		}

		if val, ok := result.Get("key2"); !ok || val != "value2" {
			t.Error("Should have key2")
		}

		if val, ok := result.Get("key3"); !ok || val != "value3" {
			t.Error("Should have key3")
		}

		if val, ok := result.Get("common"); !ok || val != "modified" {
			t.Error("Common key should be overridden by later state")
		}
	})

	t.Run("MergeStrategyUnion", func(t *testing.T) {
		// Create states with duplicate values
		s1 := domain.NewState()
		s1.Set("nums", 1)
		s1.Set("unique1", "a")

		s2 := domain.NewState()
		s2.Set("nums", 2)
		s2.Set("unique2", "b")

		s3 := domain.NewState()
		s3.Set("nums", 1) // Duplicate value
		s3.Set("unique3", "c")

		result, err := sm.MergeStates([]*domain.State{s1, s2, s3}, domain.MergeStrategyUnion)
		if err != nil {
			t.Fatalf("Failed to merge: %v", err)
		}

		// Check unique values are preserved
		if val, ok := result.Get("unique1"); !ok || val != "a" {
			t.Error("Should have unique1")
		}

		// Check values with multiple occurrences
		if val, ok := result.Get("nums"); !ok {
			t.Error("Should have nums")
		} else {
			// Should be an array with unique values
			if arr, ok := val.([]interface{}); !ok {
				t.Error("nums should be an array")
			} else if len(arr) != 2 {
				t.Errorf("nums should have 2 unique values, got %d", len(arr))
			}
		}
	})

	t.Run("EmptyStates", func(t *testing.T) {
		_, err := sm.MergeStates([]*domain.State{}, domain.MergeStrategyLast)
		if err == nil {
			t.Error("Should error when merging empty states")
		}
	})

	t.Run("WithNilStates", func(t *testing.T) {
		// Include nil states in the mix
		statesWithNil := []*domain.State{state1, nil, state2, nil, state3}

		result, err := sm.MergeStates(statesWithNil, domain.MergeStrategyMergeAll)
		if err != nil {
			t.Fatalf("Should handle nil states gracefully: %v", err)
		}

		// Should still have all non-nil values
		if val, ok := result.Get("key1"); !ok || val != "value1" {
			t.Error("Should have key1")
		}
		if val, ok := result.Get("key2"); !ok || val != "value2" {
			t.Error("Should have key2")
		}
		if val, ok := result.Get("key3"); !ok || val != "value3" {
			t.Error("Should have key3")
		}
	})
}

func TestStateSnapshot(t *testing.T) {
	sm := core.NewStateManager()

	// Create multiple states
	states := make(map[string]*domain.State)

	state1 := domain.NewState()
	state1.Set("data", "state1")
	states["agent1"] = state1

	state2 := domain.NewState()
	state2.Set("data", "state2")
	states["agent2"] = state2

	// Create snapshot
	snapshot := sm.CreateSnapshot(states)

	if snapshot.Timestamp == "" {
		t.Error("Snapshot should have timestamp")
	}

	if len(snapshot.States) != 2 {
		t.Errorf("Snapshot should have 2 states, got %d", len(snapshot.States))
	}

	// Verify states are cloned
	for name, originalState := range states {
		snapshotState, ok := snapshot.States[name]
		if !ok {
			t.Errorf("Snapshot missing state for %s", name)
			continue
		}

		if snapshotState.ID() == originalState.ID() {
			t.Error("Snapshot states should be clones")
		}

		// Verify data is preserved
		origData, _ := originalState.Get("data")
		snapData, _ := snapshotState.Get("data")
		if origData != snapData {
			t.Error("Snapshot should preserve state data")
		}
	}

	// Test with nil states
	statesWithNil := map[string]*domain.State{
		"agent1": state1,
		"agent2": nil,
		"agent3": state2,
	}

	snapshot2 := sm.CreateSnapshot(statesWithNil)
	if len(snapshot2.States) != 2 {
		t.Error("Snapshot should skip nil states")
	}
}

// Benchmark tests
func BenchmarkStateManagerSaveLoad(b *testing.B) {
	sm := core.NewStateManager()
	state := domain.NewState()

	// Pre-populate state
	for i := 0; i < 100; i++ {
		state.Set(sprintf("key%d", i), i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = sm.SaveState(state)
		_, _ = sm.LoadState(state.ID())
	}
}

func BenchmarkStateManagerMerge(b *testing.B) {
	sm := core.NewStateManager()

	// Create states
	states := make([]*domain.State, 10)
	for i := 0; i < 10; i++ {
		states[i] = domain.NewState()
		for j := 0; j < 10; j++ {
			states[i].Set(sprintf("key%d_%d", i, j), j)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = sm.MergeStates(states, domain.MergeStrategyMergeAll)
	}
}

// Helper functions for testing
func errorf(format string, a ...interface{}) error {
	return &testError{msg: fmt.Sprintf(format, a...)}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

