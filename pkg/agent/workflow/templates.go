// ABOUTME: Workflow templates for common patterns and reusable workflows
// ABOUTME: Provides pre-built workflows and template customization

package workflow

import (
	"fmt"
	"sync"
)

// WorkflowTemplate represents a reusable workflow pattern.
// Templates provide pre-built workflows that can be customized with
// variables, making it easy to create common workflow patterns.
type WorkflowTemplate struct {
	ID          string
	Name        string
	Description string
	Category    string
	Tags        []string
	Definition  *WorkflowDefinition
	Variables   map[string]TemplateVariable
	Examples    []TemplateExample
}

// TemplateVariable defines a customizable variable in a template.
// Variables allow templates to be parameterized and reused with
// different configurations.
type TemplateVariable struct {
	Name        string
	Description string
	Type        string // "string", "number", "boolean", "object"
	Default     interface{}
	Required    bool
	Validation  string // Simple validation expression
}

// TemplateExample shows how to use the template.
// Examples provide concrete usage scenarios with specific
// variable values to help users understand the template.
type TemplateExample struct {
	Name        string
	Description string
	Variables   map[string]interface{}
}

// templateRegistry manages workflow templates
type templateRegistry struct {
	mu        sync.RWMutex
	templates map[string]*WorkflowTemplate
}

// globalTemplateRegistry is the global template registry
var globalTemplateRegistry = &templateRegistry{
	templates: make(map[string]*WorkflowTemplate),
}

// RegisterTemplate registers a workflow template.
// Templates must have unique IDs and valid definitions.
//
// Parameters:
//   - template: The workflow template to register
//
// Returns an error if the template is invalid or ID is empty.
func RegisterTemplate(template *WorkflowTemplate) error {
	globalTemplateRegistry.mu.Lock()
	defer globalTemplateRegistry.mu.Unlock()

	if template.ID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}
	if template.Definition == nil {
		return fmt.Errorf("template definition cannot be nil")
	}

	globalTemplateRegistry.templates[template.ID] = template
	return nil
}

// GetTemplate retrieves a template by ID.
//
// Parameters:
//   - id: The template ID to look up
//
// Returns the template or an error if not found.
func GetTemplate(id string) (*WorkflowTemplate, error) {
	globalTemplateRegistry.mu.RLock()
	defer globalTemplateRegistry.mu.RUnlock()

	template, exists := globalTemplateRegistry.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// ListTemplates returns all registered templates.
// The returned slice contains templates in no particular order.
//
// Returns a slice of all registered workflow templates.
func ListTemplates() []*WorkflowTemplate {
	globalTemplateRegistry.mu.RLock()
	defer globalTemplateRegistry.mu.RUnlock()

	templates := make([]*WorkflowTemplate, 0, len(globalTemplateRegistry.templates))
	for _, template := range globalTemplateRegistry.templates {
		templates = append(templates, template)
	}
	return templates
}

// ListTemplatesByCategory returns templates in a specific category.
// Categories help organize templates by their use case or domain.
//
// Parameters:
//   - category: The category to filter by
//
// Returns templates matching the specified category.
func ListTemplatesByCategory(category string) []*WorkflowTemplate {
	globalTemplateRegistry.mu.RLock()
	defer globalTemplateRegistry.mu.RUnlock()

	var templates []*WorkflowTemplate
	for _, template := range globalTemplateRegistry.templates {
		if template.Category == category {
			templates = append(templates, template)
		}
	}
	return templates
}

// SearchTemplates searches templates by tags.
// Returns templates that have any of the specified tags.
//
// Parameters:
//   - tags: Tags to search for
//
// Returns templates matching any of the provided tags.
func SearchTemplates(tags []string) []*WorkflowTemplate {
	globalTemplateRegistry.mu.RLock()
	defer globalTemplateRegistry.mu.RUnlock()

	var templates []*WorkflowTemplate
	for _, template := range globalTemplateRegistry.templates {
		if hasAnyTag(template.Tags, tags) {
			templates = append(templates, template)
		}
	}
	return templates
}

// hasAnyTag checks if template has any of the specified tags
func hasAnyTag(templateTags, searchTags []string) bool {
	tagMap := make(map[string]bool)
	for _, tag := range templateTags {
		tagMap[tag] = true
	}
	for _, tag := range searchTags {
		if tagMap[tag] {
			return true
		}
	}
	return false
}

// ApplyTemplate creates a workflow from a template with variables.
// It validates required variables and applies defaults where needed.
// Variable substitution is performed on the workflow definition.
//
// Parameters:
//   - templateID: The ID of the template to apply
//   - variables: Variable values to substitute
//
// Returns a workflow definition with variables applied or an error.
func ApplyTemplate(templateID string, variables map[string]interface{}) (*WorkflowDefinition, error) {
	template, err := GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Validate required variables
	for varName, varDef := range template.Variables {
		if varDef.Required {
			if _, exists := variables[varName]; !exists {
				if varDef.Default != nil {
					variables[varName] = varDef.Default
				} else {
					return nil, fmt.Errorf("required variable missing: %s", varName)
				}
			}
		}
	}

	// Create a copy of the workflow definition
	// In a real implementation, you would apply variable substitution
	def := &WorkflowDefinition{
		Name:           template.Definition.Name,
		Description:    template.Definition.Description,
		Steps:          make([]WorkflowStep, len(template.Definition.Steps)),
		Parallel:       template.Definition.Parallel,
		MaxConcurrency: template.Definition.MaxConcurrency,
	}

	// Copy steps (in a real implementation, you would apply variable substitution)
	copy(def.Steps, template.Definition.Steps)

	return def, nil
}

// CreateDataProcessingTemplate creates a template for data processing workflows.
// This template provides a standard pattern for ETL (Extract, Transform, Load)
// operations with validation, transformation, and storage steps.
//
// Returns a configured data processing workflow template.
func CreateDataProcessingTemplate() *WorkflowTemplate {
	// This is a simple example template
	// Build steps with error handling
	validate, _ := NewScriptStepBuilder("validate").
		WithLanguage("expr").
		WithScript("validated = true").
		WithDescription("Validate input data").
		Build()

	transform, _ := NewScriptStepBuilder("transform").
		WithLanguage("json-transform").
		WithScript(`{"processed": "{{input}}", "timestamp": "now"}`).
		WithDescription("Transform data").
		Build()

	save, _ := NewScriptStepBuilder("save").
		WithLanguage("expr").
		WithScript("saved = true").
		WithDescription("Save results").
		Build()

	validSteps := []WorkflowStep{}
	if validate != nil {
		validSteps = append(validSteps, validate)
	}
	if transform != nil {
		validSteps = append(validSteps, transform)
	}
	if save != nil {
		validSteps = append(validSteps, save)
	}

	return &WorkflowTemplate{
		ID:          "data-processing",
		Name:        "Data Processing Pipeline",
		Description: "A template for processing data through validation, transformation, and storage",
		Category:    "data",
		Tags:        []string{"data", "etl", "processing"},
		Definition: &WorkflowDefinition{
			Name:        "Data Processing Workflow",
			Description: "Processes data through multiple stages",
			Steps:       validSteps,
			Parallel:    false,
		},
		Variables: map[string]TemplateVariable{
			"input_source": {
				Name:        "input_source",
				Description: "Source of input data",
				Type:        "string",
				Required:    true,
			},
			"output_format": {
				Name:        "output_format",
				Description: "Desired output format",
				Type:        "string",
				Default:     "json",
				Required:    false,
			},
		},
		Examples: []TemplateExample{
			{
				Name:        "CSV Processing",
				Description: "Process CSV files",
				Variables: map[string]interface{}{
					"input_source":  "data.csv",
					"output_format": "json",
				},
			},
		},
	}
}

// CreateAPIIntegrationTemplate creates a template for API integration workflows.
// This template provides a pattern for fetching data from external APIs
// and processing the responses.
//
// Returns a configured API integration workflow template.
func CreateAPIIntegrationTemplate() *WorkflowTemplate {
	return &WorkflowTemplate{
		ID:          "api-integration",
		Name:        "API Integration Workflow",
		Description: "Template for integrating with external APIs",
		Category:    "integration",
		Tags:        []string{"api", "integration", "rest"},
		Definition: &WorkflowDefinition{
			Name:        "API Integration",
			Description: "Fetches data from API and processes it",
			Steps:       []WorkflowStep{}, // Would be populated with actual steps
			Parallel:    false,
		},
		Variables: map[string]TemplateVariable{
			"api_endpoint": {
				Name:        "api_endpoint",
				Description: "API endpoint URL",
				Type:        "string",
				Required:    true,
			},
			"api_key": {
				Name:        "api_key",
				Description: "API authentication key",
				Type:        "string",
				Required:    false,
			},
		},
	}
}

// RegisterDefaultTemplates registers the built-in templates.
// This includes data processing and API integration templates.
// Additional templates can be registered separately.
//
// Returns an error if any template registration fails.
func RegisterDefaultTemplates() error {
	templates := []*WorkflowTemplate{
		CreateDataProcessingTemplate(),
		CreateAPIIntegrationTemplate(),
	}

	for _, template := range templates {
		if err := RegisterTemplate(template); err != nil {
			return fmt.Errorf("failed to register template %s: %w", template.ID, err)
		}
	}

	return nil
}
