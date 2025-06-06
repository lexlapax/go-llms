package integration

// ABOUTME: Integration tests for SharedState functionality in multi-agent systems
// ABOUTME: Tests state sharing between parent and child agents with inheritance

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// TestSharedStateBasic tests basic shared state functionality
func TestSharedStateBasic(t *testing.T) {
	// Create parent state
	parentState := domain.NewState()
	parentState.Set("global_config", "production")
	parentState.Set("api_key", "secret123")
	parentState.Set("shared_data", "parent value")

	// Create shared state context
	sharedState := domain.NewSharedStateContext(parentState)

	// Test reading from parent
	configVal, exists := sharedState.Get("global_config")
	if !exists {
		t.Error("Expected to find global_config from parent")
	}
	if config, ok := configVal.(string); !ok || config != "production" {
		t.Errorf("Expected global_config to be 'production', got %v", configVal)
	}

	// Test local override
	sharedState.Set("shared_data", "child value")

	// Verify local value overrides parent
	dataVal, _ := sharedState.Get("shared_data")
	if data, ok := dataVal.(string); !ok || data != "child value" {
		t.Errorf("Expected shared_data to be 'child value', got %v", dataVal)
	}

	// Test new local value
	sharedState.Set("child_only", "local data")
	childVal, exists := sharedState.Get("child_only")
	if !exists {
		t.Error("Expected to find child_only in local state")
	}
	if val, ok := childVal.(string); !ok || val != "local data" {
		t.Errorf("Expected child_only to be 'local data', got %v", childVal)
	}

	// Verify parent doesn't have child's local value
	if _, exists := parentState.Get("child_only"); exists {
		t.Error("Parent should not have child's local value")
	}
}

// TestSharedStateInheritanceConfig tests different inheritance configurations
func TestSharedStateInheritanceConfig(t *testing.T) {
	// Create parent state with various data types
	parentState := domain.NewState()
	parentState.Set("value1", "parent value")
	parentState.Set("value2", 42)

	// Add artifact
	artifact := domain.NewArtifact("doc1", domain.ArtifactTypeDocument, []byte("Parent document"))
	artifact.Metadata = map[string]interface{}{
		"source": "parent",
	}
	parentState.AddArtifact(artifact)

	// Add message
	parentState.AddMessage(domain.Message{
		Role:    "user",
		Content: "Parent message",
	})

	// Add metadata
	parentState.SetMetadata("parent_meta", "meta value")

	// Test 1: Inherit only values (not messages, artifacts, metadata)
	sharedState1 := domain.NewSharedStateContext(parentState).
		WithInheritanceConfig(false, false, false)

	// Should see values
	if val, exists := sharedState1.Get("value1"); !exists {
		t.Error("Should inherit values")
	} else if v, ok := val.(string); !ok || v != "parent value" {
		t.Errorf("Expected 'parent value', got %v", val)
	}

	// Should not see artifacts
	if arts := sharedState1.Artifacts(); len(arts) != 0 {
		t.Error("Should not inherit artifacts when disabled")
	}

	// Should not see messages
	if msgs := sharedState1.Messages(); len(msgs) != 0 {
		t.Error("Should not inherit messages when disabled")
	}

	// Should not see metadata
	if _, exists := sharedState1.GetMetadata("parent_meta"); exists {
		t.Error("Should not inherit metadata when disabled")
	}

	// Test 2: Inherit everything (default)
	sharedState2 := domain.NewSharedStateContext(parentState)

	// Should see everything
	if _, exists := sharedState2.Get("value1"); !exists {
		t.Error("Should inherit values")
	}
	if arts := sharedState2.Artifacts(); len(arts) != 1 {
		t.Error("Should inherit artifacts by default")
	}
	if msgs := sharedState2.Messages(); len(msgs) != 1 {
		t.Error("Should inherit messages by default")
	}
	if _, exists := sharedState2.GetMetadata("parent_meta"); !exists {
		t.Error("Should inherit metadata by default")
	}
}

// TestSharedStateWithSubAgents tests shared state with agent tools
func TestSharedStateWithSubAgents(t *testing.T) {
	// Create a parent state with global configuration
	parentState := domain.NewState()
	parentState.Set("global_config", map[string]interface{}{
		"environment":  "production",
		"debug":        false,
		"api_endpoint": "https://api.example.com",
	})
	parentState.Set("session_id", "SESSION-12345")

	// Create shared state context
	sharedState := domain.NewSharedStateContext(parentState)

	// Add local override
	sharedState.Set("debug", true) // Override parent's debug setting
	sharedState.Set("agent_name", "analyzer")

	// Test that shared state has both parent and local values

	// Should get parent's global_config
	configVal, exists := sharedState.Get("global_config")
	if !exists {
		t.Fatal("Expected to find global_config from parent")
	}
	config := configVal.(map[string]interface{})
	if config["environment"] != "production" {
		t.Errorf("Expected environment to be production, got %v", config["environment"])
	}

	// Should get parent's session_id
	sessionVal, exists := sharedState.Get("session_id")
	if !exists {
		t.Fatal("Expected to find session_id from parent")
	}
	if session := sessionVal.(string); session != "SESSION-12345" {
		t.Errorf("Expected session_id to be SESSION-12345, got %s", session)
	}

	// Should get local override for debug
	debugVal, exists := sharedState.Get("debug")
	if !exists {
		t.Fatal("Expected to find debug in state")
	}
	if debug := debugVal.(bool); !debug {
		t.Error("Expected debug to be true (local override)")
	}

	// Should get local agent_name
	agentVal, exists := sharedState.Get("agent_name")
	if !exists {
		t.Fatal("Expected to find agent_name in local state")
	}
	if agent := agentVal.(string); agent != "analyzer" {
		t.Errorf("Expected agent_name to be analyzer, got %s", agent)
	}

	// Test converting shared state to regular state
	mergedState := sharedState.AsState()

	// Verify all values are present in merged state
	if val, _ := mergedState.Get("global_config"); val == nil {
		t.Error("Expected global_config in merged state")
	}
	if val, _ := mergedState.Get("session_id"); val == nil {
		t.Error("Expected session_id in merged state")
	}
	if val, _ := mergedState.Get("debug"); val == nil {
		t.Error("Expected debug in merged state")
	}
	if val, _ := mergedState.Get("agent_name"); val == nil {
		t.Error("Expected agent_name in merged state")
	}

	// Test cloning shared state
	cloned := sharedState.Clone()

	// Cloned should have same parent but fresh local state
	if val, exists := cloned.Get("global_config"); !exists {
		t.Error("Cloned state should still have access to parent's global_config")
	} else if config := val.(map[string]interface{}); config["environment"] != "production" {
		t.Error("Cloned state should have same parent values")
	}

	// Cloned should not have local overrides
	if _, exists := cloned.Get("agent_name"); exists {
		t.Error("Cloned state should not have local values from original")
	}

	t.Log("SharedStateContext successfully demonstrated parent-child state sharing")
}
