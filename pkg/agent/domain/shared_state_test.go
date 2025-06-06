// ABOUTME: Tests for SharedStateContext implementation that enables parent-child state sharing
// ABOUTME: including inheritance configuration, value fallback, and state isolation

package domain

import (
	"testing"
)

func TestSharedStateContext_BasicOperations(t *testing.T) {
	// Create parent state
	parent := NewState()
	parent.Set("parent_key", "parent_value")
	parent.Set("shared_key", "parent_shared")

	// Create shared state context
	ssc := NewSharedStateContext(parent)

	// Test Get with parent fallback
	t.Run("Get parent value", func(t *testing.T) {
		val, ok := ssc.Get("parent_key")
		if !ok {
			t.Error("Expected to find parent_key")
		}
		if val != "parent_value" {
			t.Errorf("Expected 'parent_value', got %v", val)
		}
	})

	// Test Set only affects local state
	t.Run("Set local value", func(t *testing.T) {
		ssc.Set("local_key", "local_value")

		// Check local state has it
		val, ok := ssc.Get("local_key")
		if !ok {
			t.Error("Expected to find local_key")
		}
		if val != "local_value" {
			t.Errorf("Expected 'local_value', got %v", val)
		}

		// Check parent doesn't have it
		_, ok = parent.Get("local_key")
		if ok {
			t.Error("Parent should not have local_key")
		}
	})

	// Test override parent value
	t.Run("Override parent value", func(t *testing.T) {
		ssc.Set("shared_key", "local_override")

		// Local should have override
		val, ok := ssc.Get("shared_key")
		if !ok {
			t.Error("Expected to find shared_key")
		}
		if val != "local_override" {
			t.Errorf("Expected 'local_override', got %v", val)
		}

		// Parent should still have original
		parentVal, ok := parent.Get("shared_key")
		if !ok {
			t.Error("Parent should still have shared_key")
		}
		if parentVal != "parent_shared" {
			t.Errorf("Parent should still have 'parent_shared', got %v", parentVal)
		}
	})
}

func TestSharedStateContext_Messages(t *testing.T) {
	// Create parent with messages
	parent := NewState()
	parent.AddMessage(NewMessage(RoleSystem, "System message"))
	parent.AddMessage(NewMessage(RoleUser, "User message"))

	// Create shared context
	ssc := NewSharedStateContext(parent)

	t.Run("Inherit messages", func(t *testing.T) {
		messages := ssc.Messages()
		if len(messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(messages))
		}
	})

	t.Run("Add local messages", func(t *testing.T) {
		ssc.LocalState().AddMessage(NewMessage(RoleAssistant, "Assistant message"))

		messages := ssc.Messages()
		if len(messages) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(messages))
		}

		// Parent should still have 2
		if len(parent.Messages()) != 2 {
			t.Errorf("Parent should still have 2 messages, got %d", len(parent.Messages()))
		}
	})
}

func TestSharedStateContext_Artifacts(t *testing.T) {
	// Create parent with artifacts
	parent := NewState()
	artifact1 := NewArtifact("artifact1", ArtifactTypeData, []byte("data1"))
	parent.AddArtifact(artifact1)

	// Create shared context
	ssc := NewSharedStateContext(parent)

	t.Run("Inherit artifacts", func(t *testing.T) {
		artifact, ok := ssc.GetArtifact(artifact1.ID)
		if !ok {
			t.Error("Expected to find artifact1")
		}
		if artifact.Name != "artifact1" {
			t.Errorf("Expected artifact1, got %s", artifact.Name)
		}
	})

	t.Run("Add local artifacts", func(t *testing.T) {
		artifact2 := NewArtifact("artifact2", ArtifactTypeData, []byte("data2"))
		ssc.LocalState().AddArtifact(artifact2)

		// Should have both artifacts
		artifacts := ssc.Artifacts()
		if len(artifacts) != 2 {
			t.Errorf("Expected 2 artifacts, got %d", len(artifacts))
		}

		// Parent should still have 1
		if len(parent.Artifacts()) != 1 {
			t.Errorf("Parent should still have 1 artifact, got %d", len(parent.Artifacts()))
		}
	})
}

func TestSharedStateContext_InheritanceConfig(t *testing.T) {
	// Create parent state with data
	parent := NewState()
	parent.Set("key", "value")
	parent.AddMessage(NewMessage(RoleUser, "Message"))
	parent.AddArtifact(NewArtifact("a1", ArtifactTypeData, []byte("data")))
	parent.SetMetadata("meta", "data")

	// Create shared context and disable inheritance
	ssc := NewSharedStateContext(parent).
		WithInheritanceConfig(false, false, false)

	t.Run("Messages not inherited", func(t *testing.T) {
		messages := ssc.Messages()
		if len(messages) != 0 {
			t.Errorf("Expected 0 messages when inheritance disabled, got %d", len(messages))
		}
	})

	t.Run("Artifacts not inherited", func(t *testing.T) {
		artifacts := ssc.Artifacts()
		if len(artifacts) != 0 {
			t.Errorf("Expected 0 artifacts when inheritance disabled, got %d", len(artifacts))
		}
	})

	t.Run("Metadata not inherited", func(t *testing.T) {
		_, ok := ssc.GetMetadata("meta")
		if ok {
			t.Error("Should not find metadata when inheritance disabled")
		}
	})

	t.Run("Regular values still inherited", func(t *testing.T) {
		// Regular key-value pairs are always inherited
		val, ok := ssc.Get("key")
		if !ok {
			t.Error("Regular values should always be inherited")
		}
		if val != "value" {
			t.Errorf("Expected 'value', got %v", val)
		}
	})
}

func TestSharedStateContext_Clone(t *testing.T) {
	// Create parent and shared context
	parent := NewState()
	parent.Set("parent_key", "parent_value")

	ssc := NewSharedStateContext(parent)
	ssc.Set("local_key", "local_value")

	// Clone the context
	clone := ssc.Clone()

	t.Run("Clone has same parent", func(t *testing.T) {
		val, ok := clone.Get("parent_key")
		if !ok {
			t.Error("Clone should have access to parent values")
		}
		if val != "parent_value" {
			t.Errorf("Expected 'parent_value', got %v", val)
		}
	})

	t.Run("Clone has fresh local state", func(t *testing.T) {
		_, ok := clone.Get("local_key")
		if ok {
			t.Error("Clone should not have original's local values")
		}
	})

	t.Run("Clone modifications don't affect original", func(t *testing.T) {
		clone.Set("clone_key", "clone_value")

		_, ok := ssc.Get("clone_key")
		if ok {
			t.Error("Original should not have clone's values")
		}
	})
}

func TestSharedStateContext_AsState(t *testing.T) {
	// Create parent with various data
	parent := NewState()
	parent.Set("parent_key", "parent_value")
	parent.AddMessage(NewMessage(RoleUser, "Parent message"))
	parent.AddArtifact(NewArtifact("parent_artifact", ArtifactTypeData, []byte("parent")))

	// Create shared context with local data
	ssc := NewSharedStateContext(parent)
	ssc.Set("local_key", "local_value")
	ssc.Set("parent_key", "overridden") // Override parent value
	ssc.LocalState().AddMessage(NewMessage(RoleAssistant, "Local message"))
	ssc.LocalState().AddArtifact(NewArtifact("local_artifact", ArtifactTypeData, []byte("local")))

	// Convert to regular state
	state := ssc.AsState()

	t.Run("Has all values", func(t *testing.T) {
		// Should have overridden value
		val, ok := state.Get("parent_key")
		if !ok || val != "overridden" {
			t.Errorf("Expected overridden value, got %v", val)
		}

		// Should have local value
		val, ok = state.Get("local_key")
		if !ok || val != "local_value" {
			t.Errorf("Expected local value, got %v", val)
		}
	})

	t.Run("Has all messages", func(t *testing.T) {
		messages := state.Messages()
		if len(messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(messages))
		}
	})

	t.Run("Has all artifacts", func(t *testing.T) {
		artifacts := state.Artifacts()
		if len(artifacts) != 2 {
			t.Errorf("Expected 2 artifacts, got %d", len(artifacts))
		}
	})

	t.Run("Is independent of original", func(t *testing.T) {
		state.Set("new_key", "new_value")

		_, ok := ssc.Get("new_key")
		if ok {
			t.Error("Original shared context should not have new state's values")
		}

		_, ok = parent.Get("new_key")
		if ok {
			t.Error("Parent should not have new state's values")
		}
	})
}

func TestSharedStateContext_Keys(t *testing.T) {
	parent := NewState()
	parent.Set("key1", "value1")
	parent.Set("key2", "value2")

	ssc := NewSharedStateContext(parent)
	ssc.Set("key3", "value3")
	ssc.Set("key1", "override") // Override parent key

	keys := ssc.Keys()

	// Should have all 3 unique keys
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	if len(keyMap) != 3 {
		t.Errorf("Expected 3 unique keys, got %d", len(keyMap))
	}

	expectedKeys := []string{"key1", "key2", "key3"}
	for _, expected := range expectedKeys {
		if !keyMap[expected] {
			t.Errorf("Missing expected key: %s", expected)
		}
	}
}

func TestSharedStateContext_MergeToParent(t *testing.T) {
	parent := NewState()
	ssc := NewSharedStateContext(parent)

	// Currently, MergeToParent returns an error because StateReader is read-only
	err := ssc.MergeToParent()
	if err != ErrStateReadOnly {
		t.Errorf("Expected ErrStateReadOnly, got %v", err)
	}
}

func TestSharedStateContext_ThreadSafety(t *testing.T) {
	parent := NewState()
	parent.Set("counter", 0)

	ssc := NewSharedStateContext(parent)

	// Run concurrent operations
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			ssc.Set("key"+string(rune(i)), i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			ssc.Get("counter")
			ssc.Values()
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Verify state is consistent
	values := ssc.Values()
	if len(values) < 1 {
		t.Error("Expected at least one value")
	}
}
