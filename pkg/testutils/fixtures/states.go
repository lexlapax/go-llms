// ABOUTME: Pre-configured state fixtures for common testing scenarios
// ABOUTME: Provides ready-to-use state objects with typical data patterns and artifacts

package fixtures

import (
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// EmptyTestState creates a completely empty state for testing
func EmptyTestState() *domain.State {
	return domain.NewState()
}

// BasicTestState creates a state with basic test data
func BasicTestState() *domain.State {
	state := domain.NewState()

	// Basic values
	state.Set("id", "test-123")
	state.Set("name", "Test Entity")
	state.Set("status", "active")
	state.Set("created_at", time.Now().Format(time.RFC3339))

	// Nested data
	data := map[string]interface{}{
		"type":        "example",
		"category":    "test",
		"priority":    1,
		"enabled":     true,
		"description": "A basic test entity for testing purposes",
	}
	state.Set("data", data)

	// Array data
	tags := []string{"test", "example", "basic"}
	state.Set("tags", tags)

	return state
}

// WorkflowTestState creates a state representing a workflow execution
func WorkflowTestState() *domain.State {
	state := domain.NewState()

	// Workflow identification
	state.Set("workflow_id", "wf-test-001")
	state.Set("workflow_name", "Test Workflow")
	state.Set("workflow_version", "1.0.0")

	// Workflow steps
	steps := []map[string]interface{}{
		{
			"name":        "initialize",
			"description": "Initialize workflow",
			"status":      "pending",
			"order":       0,
		},
		{
			"name":        "process",
			"description": "Process data",
			"status":      "pending",
			"order":       1,
		},
		{
			"name":        "finalize",
			"description": "Finalize workflow",
			"status":      "pending",
			"order":       2,
		},
	}
	state.Set("steps", steps)

	// Current state
	state.Set("current_step", 0)
	state.Set("status", "running")
	state.Set("progress", 0.0)

	// Execution context
	context := map[string]interface{}{
		"started_at":  time.Now().Format(time.RFC3339),
		"timeout":     300, // 5 minutes
		"retry_count": 0,
		"max_retries": 3,
	}
	state.Set("execution_context", context)

	return state
}

// ConversationTestState creates a state with conversation history
func ConversationTestState() *domain.State {
	state := domain.NewState()

	// Conversation metadata
	state.Set("conversation_id", "conv-test-001")
	state.Set("participant_count", 2)
	state.Set("started_at", time.Now().Add(-10*time.Minute).Format(time.RFC3339))

	// Add conversation messages
	systemMsg := domain.NewMessage(domain.RoleSystem, "You are a helpful assistant for testing purposes.")
	userMsg := domain.NewMessage(domain.RoleUser, "Hello, I need help with testing.")
	assistantMsg := domain.NewMessage(domain.RoleAssistant, "Hello! I'm happy to help you with testing. What would you like to test?")

	state.AddMessage(systemMsg)
	state.AddMessage(userMsg)
	state.AddMessage(assistantMsg)

	// Conversation context
	context := map[string]interface{}{
		"topic":      "testing",
		"domain":     "general",
		"language":   "en",
		"model":      "test-model",
		"max_tokens": 1000,
	}
	state.Set("context", context)

	// Current conversation state
	state.Set("awaiting_response", false)
	state.Set("last_activity", time.Now().Format(time.RFC3339))

	return state
}

// ErrorTestState creates a state representing an error condition
func ErrorTestState() *domain.State {
	state := domain.NewState()

	// Error flags
	state.Set("has_error", true)
	state.Set("error_code", "TEST_ERROR_001")
	state.Set("error_message", "This is a test error for testing error handling")
	state.Set("error_timestamp", time.Now().Format(time.RFC3339))

	// Error details
	errorDetails := map[string]interface{}{
		"type":     "validation",
		"source":   "input",
		"field":    "test_field",
		"actual":   "invalid_value",
		"expected": "valid_value",
		"stack_trace": []string{
			"at TestFunction (test.go:123)",
			"at ValidateInput (validator.go:456)",
			"at ProcessRequest (handler.go:789)",
		},
	}
	state.Set("error_details", errorDetails)

	// Retry information
	state.Set("retry_count", 2)
	state.Set("max_retries", 3)
	state.Set("next_retry_at", time.Now().Add(5*time.Second).Format(time.RFC3339))

	// Original request context (for debugging)
	originalContext := map[string]interface{}{
		"request_id": "req-test-001",
		"user_id":    "user-test-123",
		"session_id": "session-test-456",
		"timestamp":  time.Now().Add(-30 * time.Second).Format(time.RFC3339),
		"parameters": map[string]interface{}{
			"test_param":  "invalid_value",
			"other_param": 42,
		},
	}
	state.Set("original_context", originalContext)

	return state
}

// StateWithArtifacts creates a state containing various artifacts
func StateWithArtifacts() *domain.State {
	state := domain.NewState()

	// Create test artifacts
	reportContent := []byte("Test Report Content\nThis is a sample PDF report for testing purposes.")
	reportArtifact := domain.NewArtifact("Test Report", domain.ArtifactTypeDocument, reportContent)
	reportArtifact.WithMimeType("application/pdf")
	reportArtifact.WithMetadata("format", "pdf")
	reportArtifact.WithMetadata("pages", 3)
	reportArtifact.WithMetadata("author", "test-agent")
	reportArtifact.WithMetadata("version", "1.0")

	dataContent := []byte(`{"test": true, "data": [1, 2, 3], "message": "test artifact"}`)
	dataArtifact := domain.NewArtifact("Test Data", domain.ArtifactTypeData, dataContent)
	dataArtifact.WithMimeType("application/json")
	dataArtifact.WithMetadata("schema", "test-schema-v1")
	dataArtifact.WithMetadata("validated", true)
	dataArtifact.WithMetadata("source", "test-generator")

	// Add artifacts to state
	state.AddArtifact(reportArtifact)
	state.AddArtifact(dataArtifact)

	// Add artifact references in state data
	state.Set("artifacts", []string{reportArtifact.ID, dataArtifact.ID})
	state.Set("primary_artifact", reportArtifact.ID)
	state.Set("artifact_count", 2)

	// Additional state data
	state.Set("processing_complete", true)
	state.Set("output_ready", true)

	return state
}

// StateWithMetadata creates a state with comprehensive metadata
func StateWithMetadata() *domain.State {
	state := domain.NewState()

	// Set basic state data
	state.Set("task", "metadata_test")
	state.Set("status", "completed")
	state.Set("result", "success")

	// Set comprehensive metadata
	state.SetMetadata("created_by", "test-agent")
	state.SetMetadata("session_id", "session-test-001")
	state.SetMetadata("request_id", "req-test-001")
	state.SetMetadata("environment", "test")
	state.SetMetadata("version", "1.0.0")
	state.SetMetadata("created_at", time.Now().Format(time.RFC3339))

	// Set array metadata
	tags := []string{"testing", "fixtures", "automation"}
	state.SetMetadata("tags", tags)

	// Set nested metadata
	config := map[string]interface{}{
		"debug_mode": true,
		"log_level":  "info",
		"timeout":    30.0,
		"retries":    3,
		"features": map[string]bool{
			"async":     true,
			"streaming": false,
			"caching":   true,
		},
	}
	state.SetMetadata("config", config)

	// Performance metadata
	performance := map[string]interface{}{
		"duration_ms":    150,
		"memory_used_mb": 2.5,
		"cpu_time_ms":    45,
		"network_calls":  3,
		"cache_hits":     2,
		"cache_misses":   1,
	}
	state.SetMetadata("performance", performance)

	// Tracking metadata
	tracking := map[string]interface{}{
		"trace_id":       "trace-test-001",
		"span_id":        "span-test-001",
		"parent_span_id": "span-parent-001",
		"correlation_id": "corr-test-001",
		"operation_name": "test_operation",
	}
	state.SetMetadata("tracking", tracking)

	return state
}

// StateChain creates a series of related states for testing state transitions
func StateChain() []*domain.State {
	states := make([]*domain.State, 3)

	// Initial state
	states[0] = domain.NewState()
	states[0].Set("phase", "initial")
	states[0].Set("step", 1)
	states[0].Set("status", "started")
	states[0].Set("data", map[string]interface{}{"value": 10})

	// Processing state
	states[1] = states[0].Clone()
	states[1].Set("phase", "processing")
	states[1].Set("step", 2)
	states[1].Set("status", "processing")
	states[1].Set("data", map[string]interface{}{"value": 20})

	// Final state
	states[2] = states[1].Clone()
	states[2].Set("phase", "final")
	states[2].Set("step", 3)
	states[2].Set("status", "completed")
	states[2].Set("data", map[string]interface{}{"value": 30})

	return states
}

// LargeTestState creates a state with large amounts of data for performance testing
func LargeTestState() *domain.State {
	state := domain.NewState()

	// Large array of data
	largeArray := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeArray[i] = map[string]interface{}{
			"id":        i,
			"name":      "Item " + string(rune(i)),
			"value":     i * 2,
			"timestamp": time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			"active":    i%2 == 0,
			"metadata": map[string]interface{}{
				"category": "test",
				"priority": i % 5,
				"tags":     []string{"tag1", "tag2", "tag3"},
			},
		}
	}
	state.Set("large_array", largeArray)

	// Large nested object
	largeObject := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		key := "key_" + string(rune(i))
		largeObject[key] = map[string]interface{}{
			"nested_data": make(map[string]interface{}),
			"array_data":  make([]string, 50),
		}
	}
	state.Set("large_object", largeObject)

	// Metadata for size tracking
	state.SetMetadata("size_category", "large")
	state.SetMetadata("estimated_size_mb", 5.0)
	state.SetMetadata("performance_test", true)

	return state
}
