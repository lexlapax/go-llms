// ABOUTME: Workflow serialization for bridge layer integration and persistence
// ABOUTME: Converts workflows to/from JSON and YAML formats for downstream consumption

package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// SerializableWorkflowDefinition is a bridge-friendly representation of WorkflowDefinition
type SerializableWorkflowDefinition struct {
	Name           string                 `json:"name" yaml:"name"`
	Description    string                 `json:"description" yaml:"description"`
	Version        string                 `json:"version" yaml:"version"`
	Steps          []SerializableStep     `json:"steps" yaml:"steps"`
	Parallel       bool                   `json:"parallel" yaml:"parallel"`
	MaxConcurrency int                    `json:"max_concurrency,omitempty" yaml:"max_concurrency,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" yaml:"updated_at"`
}

// SerializableStep represents a workflow step in serializable format
type SerializableStep struct {
	Name        string                 `json:"name" yaml:"name"`
	Type        string                 `json:"type" yaml:"type"` // "agent", "script", "conditional", "loop", "parallel"
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Config      map[string]interface{} `json:"config" yaml:"config"`
	Script      *ScriptConfig          `json:"script,omitempty" yaml:"script,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ScriptConfig holds script-specific configuration
type ScriptConfig struct {
	Language    string                 `json:"language" yaml:"language"`
	Source      string                 `json:"source" yaml:"source"`
	Environment map[string]interface{} `json:"environment,omitempty" yaml:"environment,omitempty"`
	Timeout     string                 `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// WorkflowSerializer handles workflow serialization
type WorkflowSerializer interface {
	// Serialize converts a workflow definition to bytes
	Serialize(def *WorkflowDefinition) ([]byte, error)
	// Deserialize creates a workflow definition from bytes
	Deserialize(data []byte) (*WorkflowDefinition, error)
	// Format returns the serialization format
	Format() string
}

// JSONWorkflowSerializer serializes workflows to JSON
type JSONWorkflowSerializer struct {
	pretty bool
}

// NewJSONWorkflowSerializer creates a new JSON serializer
func NewJSONWorkflowSerializer(pretty bool) *JSONWorkflowSerializer {
	return &JSONWorkflowSerializer{pretty: pretty}
}

// Serialize implements WorkflowSerializer
func (s *JSONWorkflowSerializer) Serialize(def *WorkflowDefinition) ([]byte, error) {
	serializable, err := ToSerializable(def)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to serializable: %w", err)
	}

	if s.pretty {
		return json.MarshalIndent(serializable, "", "  ")
	}
	return json.Marshal(serializable)
}

// Deserialize implements WorkflowSerializer
func (s *JSONWorkflowSerializer) Deserialize(data []byte) (*WorkflowDefinition, error) {
	var serializable SerializableWorkflowDefinition
	if err := json.Unmarshal(data, &serializable); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return FromSerializable(&serializable)
}

// Format implements WorkflowSerializer
func (s *JSONWorkflowSerializer) Format() string {
	return "json"
}

// YAMLWorkflowSerializer serializes workflows to YAML
type YAMLWorkflowSerializer struct{}

// NewYAMLWorkflowSerializer creates a new YAML serializer
func NewYAMLWorkflowSerializer() *YAMLWorkflowSerializer {
	return &YAMLWorkflowSerializer{}
}

// Serialize implements WorkflowSerializer
func (s *YAMLWorkflowSerializer) Serialize(def *WorkflowDefinition) ([]byte, error) {
	serializable, err := ToSerializable(def)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to serializable: %w", err)
	}

	return yaml.Marshal(serializable)
}

// Deserialize implements WorkflowSerializer
func (s *YAMLWorkflowSerializer) Deserialize(data []byte) (*WorkflowDefinition, error) {
	var serializable SerializableWorkflowDefinition
	if err := yaml.Unmarshal(data, &serializable); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return FromSerializable(&serializable)
}

// Format implements WorkflowSerializer
func (s *YAMLWorkflowSerializer) Format() string {
	return "yaml"
}

// ToSerializable converts a WorkflowDefinition to its serializable form
func ToSerializable(def *WorkflowDefinition) (*SerializableWorkflowDefinition, error) {
	if def == nil {
		return nil, fmt.Errorf("workflow definition cannot be nil")
	}

	serializable := &SerializableWorkflowDefinition{
		Name:           def.Name,
		Description:    def.Description,
		Version:        "1.0",
		Parallel:       def.Parallel,
		MaxConcurrency: def.MaxConcurrency,
		Steps:          make([]SerializableStep, 0, len(def.Steps)),
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Convert steps
	for _, step := range def.Steps {
		serStep, err := serializeStep(step)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize step %s: %w", step.Name(), err)
		}
		serializable.Steps = append(serializable.Steps, serStep)
	}

	return serializable, nil
}

// FromSerializable creates a WorkflowDefinition from its serializable form
func FromSerializable(serializable *SerializableWorkflowDefinition) (*WorkflowDefinition, error) {
	if serializable == nil {
		return nil, fmt.Errorf("serializable definition cannot be nil")
	}

	def := &WorkflowDefinition{
		Name:           serializable.Name,
		Description:    serializable.Description,
		Parallel:       serializable.Parallel,
		MaxConcurrency: serializable.MaxConcurrency,
		Steps:          make([]WorkflowStep, 0, len(serializable.Steps)),
	}

	// Convert steps
	for _, serStep := range serializable.Steps {
		step, err := deserializeStep(serStep)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize step %s: %w", serStep.Name, err)
		}
		def.Steps = append(def.Steps, step)
	}

	return def, nil
}

// serializeStep converts a WorkflowStep to its serializable form
func serializeStep(step WorkflowStep) (SerializableStep, error) {
	// Determine step type
	stepType := "custom"
	config := make(map[string]interface{})

	switch s := step.(type) {
	case *AgentStep:
		stepType = "agent"
		config["agent_type"] = fmt.Sprintf("%T", s.agent)
	case *ScriptStep:
		stepType = "script"
		return SerializableStep{
			Name:        s.Name(),
			Type:        stepType,
			Description: s.description,
			Script: &ScriptConfig{
				Language:    s.language,
				Source:      s.script,
				Environment: s.environment,
				Timeout:     s.timeout.String(),
			},
			Metadata: s.metadata,
		}, nil
	default:
		// For custom step types, store type information
		config["step_type"] = fmt.Sprintf("%T", step)
	}

	return SerializableStep{
		Name:   step.Name(),
		Type:   stepType,
		Config: config,
	}, nil
}

// deserializeStep creates a WorkflowStep from its serializable form
func deserializeStep(serStep SerializableStep) (WorkflowStep, error) {
	switch serStep.Type {
	case "script":
		if serStep.Script == nil {
			return nil, fmt.Errorf("script configuration missing for script step")
		}

		timeout, err := time.ParseDuration(serStep.Script.Timeout)
		if err != nil && serStep.Script.Timeout != "" {
			return nil, fmt.Errorf("invalid timeout duration: %w", err)
		}

		return &ScriptStep{
			name:        serStep.Name,
			description: serStep.Description,
			language:    serStep.Script.Language,
			script:      serStep.Script.Source,
			environment: serStep.Script.Environment,
			timeout:     timeout,
			metadata:    serStep.Metadata,
		}, nil

	case "agent":
		// For agent steps, we need a factory or registry to recreate the agent
		// This is a placeholder that would need integration with agent registry
		return nil, fmt.Errorf("agent deserialization requires agent registry (not implemented)")

	default:
		// For custom steps, we need a step factory
		return nil, fmt.Errorf("unknown step type: %s", serStep.Type)
	}
}

// SerializeWorkflow is a convenience function to serialize a workflow
func SerializeWorkflow(def *WorkflowDefinition, format string) ([]byte, error) {
	serializer := GetWorkflowSerializer(format)
	return serializer.Serialize(def)
}

// DeserializeWorkflow is a convenience function to deserialize a workflow
func DeserializeWorkflow(data []byte, format string) (*WorkflowDefinition, error) {
	serializer := GetWorkflowSerializer(format)
	return serializer.Deserialize(data)
}

// GetWorkflowSerializer returns a serializer for the specified format
func GetWorkflowSerializer(format string) WorkflowSerializer {
	switch format {
	case "yaml", "yml":
		return NewYAMLWorkflowSerializer()
	case "json-pretty":
		return NewJSONWorkflowSerializer(true)
	default:
		return NewJSONWorkflowSerializer(false)
	}
}

// DeserializeDefinition creates a workflow from a map (for bridge layer)
func DeserializeDefinition(data map[string]interface{}) (*WorkflowDefinition, error) {
	// Convert map to JSON first
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	// Then deserialize normally
	return DeserializeWorkflow(jsonData, "json")
}
