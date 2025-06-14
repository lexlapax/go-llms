package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestEmptyTestState(t *testing.T) {
	state := EmptyTestState()
	assert.NotNil(t, state)
	assert.Empty(t, state.Keys())
	assert.Empty(t, state.Values())
	assert.Empty(t, state.Artifacts())
	assert.Empty(t, state.Messages())
}

func TestBasicTestState(t *testing.T) {
	state := BasicTestState()
	assert.NotNil(t, state)

	// Check basic values
	id, exists := state.Get("id")
	require.True(t, exists)
	assert.Equal(t, "test-123", id)

	name, exists := state.Get("name")
	require.True(t, exists)
	assert.Equal(t, "Test Entity", name)

	status, exists := state.Get("status")
	require.True(t, exists)
	assert.Equal(t, "active", status)

	// Check nested data
	data, exists := state.Get("data")
	require.True(t, exists)
	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "example", dataMap["type"])
	assert.Equal(t, "test", dataMap["category"])
}

func TestWorkflowTestState(t *testing.T) {
	state := WorkflowTestState()
	assert.NotNil(t, state)

	// Check workflow data
	workflowID, exists := state.Get("workflow_id")
	require.True(t, exists)
	assert.Equal(t, "wf-test-001", workflowID)

	steps, exists := state.Get("steps")
	require.True(t, exists)
	stepsList, ok := steps.([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, stepsList, 3)

	// Check first step
	assert.Equal(t, "initialize", stepsList[0]["name"])
	assert.Equal(t, "pending", stepsList[0]["status"])

	// Check current step
	currentStep, exists := state.Get("current_step")
	require.True(t, exists)
	assert.Equal(t, 0, currentStep)
}

func TestConversationTestState(t *testing.T) {
	state := ConversationTestState()
	assert.NotNil(t, state)

	// Check conversation ID
	conversationID, exists := state.Get("conversation_id")
	require.True(t, exists)
	assert.Equal(t, "conv-test-001", conversationID)

	// Check messages
	messages := state.Messages()
	assert.Len(t, messages, 3)

	// Check first message (system)
	assert.Equal(t, domain.RoleSystem, messages[0].Role)
	assert.Contains(t, messages[0].Content, "assistant")

	// Check second message (user)
	assert.Equal(t, domain.RoleUser, messages[1].Role)
	assert.Equal(t, "Hello, I need help with testing.", messages[1].Content)

	// Check third message (assistant)
	assert.Equal(t, domain.RoleAssistant, messages[2].Role)
	assert.Contains(t, messages[2].Content, "happy to help")

	// Check context
	context, exists := state.Get("context")
	require.True(t, exists)
	contextMap, ok := context.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "testing", contextMap["topic"])
	assert.Equal(t, "general", contextMap["domain"])
}

func TestErrorTestState(t *testing.T) {
	state := ErrorTestState()
	assert.NotNil(t, state)

	// Check error info
	hasError, exists := state.Get("has_error")
	require.True(t, exists)
	assert.Equal(t, true, hasError)

	errorCode, exists := state.Get("error_code")
	require.True(t, exists)
	assert.Equal(t, "TEST_ERROR_001", errorCode)

	errorMessage, exists := state.Get("error_message")
	require.True(t, exists)
	assert.Equal(t, "This is a test error for testing error handling", errorMessage)

	// Check retry info
	retryCount, exists := state.Get("retry_count")
	require.True(t, exists)
	assert.Equal(t, 2, retryCount)

	maxRetries, exists := state.Get("max_retries")
	require.True(t, exists)
	assert.Equal(t, 3, maxRetries)

	// Check error details
	errorDetails, exists := state.Get("error_details")
	require.True(t, exists)
	detailsMap, ok := errorDetails.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "validation", detailsMap["type"])
	assert.Equal(t, "input", detailsMap["source"])
}

func TestStateWithArtifacts(t *testing.T) {
	state := StateWithArtifacts()
	assert.NotNil(t, state)

	// Check that we have artifacts
	artifacts := state.Artifacts()
	assert.Len(t, artifacts, 2)

	// Find the report artifact (ID is generated, so we find by name)
	var reportArtifact, dataArtifact *domain.Artifact
	for _, artifact := range artifacts {
		switch artifact.Name {
		case "Test Report":
			reportArtifact = artifact
		case "Test Data":
			dataArtifact = artifact
		}
	}

	// Check report artifact
	require.NotNil(t, reportArtifact)
	assert.Equal(t, "Test Report", reportArtifact.Name)
	assert.Equal(t, "application/pdf", reportArtifact.MimeType)
	assert.Equal(t, domain.ArtifactTypeDocument, reportArtifact.Type)

	reportData, err := reportArtifact.Data()
	require.NoError(t, err)
	assert.Contains(t, string(reportData), "Test Report Content")

	// Check data artifact
	require.NotNil(t, dataArtifact)
	assert.Equal(t, "Test Data", dataArtifact.Name)
	assert.Equal(t, "application/json", dataArtifact.MimeType)
	assert.Equal(t, domain.ArtifactTypeData, dataArtifact.Type)

	dataContent, err := dataArtifact.Data()
	require.NoError(t, err)
	assert.Contains(t, string(dataContent), "test")

	// Check artifact references in state
	artifactRefs, exists := state.Get("artifacts")
	require.True(t, exists)
	refsList, ok := artifactRefs.([]string)
	require.True(t, ok)
	assert.Contains(t, refsList, reportArtifact.ID)
	assert.Contains(t, refsList, dataArtifact.ID)
}

func TestStateWithMetadata(t *testing.T) {
	state := StateWithMetadata()
	assert.NotNil(t, state)

	// Check metadata exists
	metadata := state.GetAllMetadata()
	assert.NotEmpty(t, metadata)

	// Check specific metadata
	createdBy, exists := state.GetMetadata("created_by")
	require.True(t, exists)
	assert.Equal(t, "test-agent", createdBy)

	sessionID, exists := state.GetMetadata("session_id")
	require.True(t, exists)
	assert.Equal(t, "session-test-001", sessionID)

	environment, exists := state.GetMetadata("environment")
	require.True(t, exists)
	assert.Equal(t, "test", environment)

	tags, exists := state.GetMetadata("tags")
	require.True(t, exists)
	tagsList, ok := tags.([]string)
	require.True(t, ok)
	assert.Contains(t, tagsList, "testing")
	assert.Contains(t, tagsList, "fixtures")
	assert.Contains(t, tagsList, "automation")

	// Check nested metadata
	config, exists := state.GetMetadata("config")
	require.True(t, exists)
	configMap, ok := config.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, configMap["debug_mode"])
	assert.Equal(t, "info", configMap["log_level"])
	assert.Equal(t, 30.0, configMap["timeout"])
}
