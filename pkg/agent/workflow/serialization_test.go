package workflow

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWorkflowSerialization(t *testing.T) {
	// Register handlers for testing
	err := RegisterDefaultHandlers()
	require.NoError(t, err)

	// Create a workflow definition
	def := &WorkflowDefinition{
		Name:        "Test Workflow",
		Description: "A test workflow for serialization",
		Parallel:    false,
		Steps: []WorkflowStep{
			&AgentStep{
				name:  "step1",
				agent: nil, // Mock agent
			},
		},
	}

	t.Run("JSON Serialization", func(t *testing.T) {
		serializer := NewJSONWorkflowSerializer(false)

		// Serialize
		data, err := serializer.Serialize(def)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Check it's valid JSON
		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, "Test Workflow", jsonData["name"])
		assert.Equal(t, "A test workflow for serialization", jsonData["description"])
		assert.Equal(t, false, jsonData["parallel"])
	})

	t.Run("JSON Pretty Serialization", func(t *testing.T) {
		serializer := NewJSONWorkflowSerializer(true)

		// Serialize
		data, err := serializer.Serialize(def)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "\n")
		assert.Contains(t, string(data), "  ")
	})

	t.Run("YAML Serialization", func(t *testing.T) {
		serializer := NewYAMLWorkflowSerializer()

		// Serialize
		data, err := serializer.Serialize(def)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Check it's valid YAML
		var yamlData map[string]interface{}
		err = yaml.Unmarshal(data, &yamlData)
		assert.NoError(t, err)
		assert.Equal(t, "Test Workflow", yamlData["name"])
	})
}

func TestScriptStepSerialization(t *testing.T) {
	// Register handlers
	err := RegisterDefaultHandlers()
	require.NoError(t, err)

	// Create a script step
	scriptStep, err := NewScriptStepBuilder("transform").
		WithLanguage("javascript").
		WithScript("return 'success'").
		WithDescription("Transform data").
		WithTimeout(10*time.Second).
		WithEnvironment("debug", true).
		WithMetadata("version", "1.0").
		Build()
	require.NoError(t, err)

	// Create workflow with script step
	def := &WorkflowDefinition{
		Name:        "Script Workflow",
		Description: "Workflow with script steps",
		Steps:       []WorkflowStep{scriptStep},
	}

	t.Run("Script Step JSON Serialization", func(t *testing.T) {
		serializer := NewJSONWorkflowSerializer(false)

		// Serialize
		data, err := serializer.Serialize(def)
		assert.NoError(t, err)

		// Check JSON structure
		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		assert.NoError(t, err)

		steps := jsonData["steps"].([]interface{})
		assert.Len(t, steps, 1)

		step := steps[0].(map[string]interface{})
		assert.Equal(t, "transform", step["name"])
		assert.Equal(t, "script", step["type"])

		script := step["script"].(map[string]interface{})
		assert.Equal(t, "javascript", script["language"])
		assert.Equal(t, "return 'success'", script["source"])
	})

	t.Run("Script Step Deserialization", func(t *testing.T) {
		// Serialize first
		serializer := NewJSONWorkflowSerializer(false)
		data, err := serializer.Serialize(def)
		require.NoError(t, err)

		// Deserialize
		deserializedDef, err := serializer.Deserialize(data)
		assert.NoError(t, err)
		assert.NotNil(t, deserializedDef)
		assert.Equal(t, "Script Workflow", deserializedDef.Name)
		assert.Len(t, deserializedDef.Steps, 1)

		// Check script step
		step := deserializedDef.Steps[0].(*ScriptStep)
		assert.Equal(t, "transform", step.Name())
		assert.Equal(t, "javascript", step.Language())
		assert.Equal(t, "return 'success'", step.Script())
		assert.Equal(t, 10*time.Second, step.Timeout())
	})
}

func TestSerializableWorkflowDefinition(t *testing.T) {
	t.Run("ToSerializable", func(t *testing.T) {
		def := &WorkflowDefinition{
			Name:           "Test",
			Description:    "Test workflow",
			Parallel:       true,
			MaxConcurrency: 5,
			Steps:          []WorkflowStep{},
		}

		serializable, err := ToSerializable(def)
		assert.NoError(t, err)
		assert.NotNil(t, serializable)
		assert.Equal(t, "Test", serializable.Name)
		assert.Equal(t, "Test workflow", serializable.Description)
		assert.Equal(t, true, serializable.Parallel)
		assert.Equal(t, 5, serializable.MaxConcurrency)
		assert.Equal(t, "1.0", serializable.Version)
		assert.NotZero(t, serializable.CreatedAt)
		assert.NotZero(t, serializable.UpdatedAt)
	})

	t.Run("FromSerializable", func(t *testing.T) {
		serializable := &SerializableWorkflowDefinition{
			Name:           "Test",
			Description:    "Test workflow",
			Version:        "1.0",
			Parallel:       true,
			MaxConcurrency: 3,
			Steps:          []SerializableStep{},
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		def, err := FromSerializable(serializable)
		assert.NoError(t, err)
		assert.NotNil(t, def)
		assert.Equal(t, "Test", def.Name)
		assert.Equal(t, "Test workflow", def.Description)
		assert.Equal(t, true, def.Parallel)
		assert.Equal(t, 3, def.MaxConcurrency)
	})
}

func TestDeserializeDefinition(t *testing.T) {
	// Create a map representation
	data := map[string]interface{}{
		"name":        "Bridge Workflow",
		"description": "Created from bridge",
		"version":     "1.0",
		"parallel":    false,
		"steps": []interface{}{
			map[string]interface{}{
				"name": "step1",
				"type": "script",
				"script": map[string]interface{}{
					"language": "expr",
					"source":   "result = 'done'",
				},
			},
		},
	}

	def, err := DeserializeDefinition(data)
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, "Bridge Workflow", def.Name)
	assert.Equal(t, "Created from bridge", def.Description)
	assert.Len(t, def.Steps, 1)
}

func TestGetWorkflowSerializer(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"json", "json"},
		{"yaml", "yaml"},
		{"yml", "yaml"},
		{"json-pretty", "json"},
		{"unknown", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			serializer := GetWorkflowSerializer(tt.format)
			assert.NotNil(t, serializer)
			assert.Equal(t, tt.expected, serializer.Format())
		})
	}
}

func TestSerializationHelpers(t *testing.T) {
	// Register handlers
	err := RegisterDefaultHandlers()
	require.NoError(t, err)

	def := &WorkflowDefinition{
		Name:        "Helper Test",
		Description: "Testing helper functions",
		Steps:       []WorkflowStep{},
	}

	t.Run("SerializeWorkflow", func(t *testing.T) {
		data, err := SerializeWorkflow(def, "json")
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify it's valid JSON
		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		assert.NoError(t, err)
	})

	t.Run("DeserializeWorkflow", func(t *testing.T) {
		// Serialize first
		data, err := SerializeWorkflow(def, "json")
		require.NoError(t, err)

		// Deserialize
		deserializedDef, err := DeserializeWorkflow(data, "json")
		assert.NoError(t, err)
		assert.NotNil(t, deserializedDef)
		assert.Equal(t, def.Name, deserializedDef.Name)
		assert.Equal(t, def.Description, deserializedDef.Description)
	})
}
