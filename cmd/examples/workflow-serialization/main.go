// ABOUTME: Example demonstrating workflow serialization and script-based steps
// ABOUTME: Shows JSON/YAML serialization, script steps, and template usage

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

func main() {
	fmt.Println("=== Workflow Serialization Example ===")

	// Register script handlers
	if err := workflow.RegisterDefaultHandlers(); err != nil {
		log.Fatal("Failed to register handlers:", err)
	}

	// Example 1: Create and serialize a workflow with script steps
	fmt.Println("\n1. Creating Workflow with Script Steps:")
	demonstrateScriptWorkflow()

	// Example 2: Serialize to different formats
	fmt.Println("\n2. Serialization Formats:")
	demonstrateSerializationFormats()

	// Example 3: Deserialize from bridge format
	fmt.Println("\n3. Bridge Layer Deserialization:")
	demonstrateBridgeDeserialization()

	// Example 4: Use workflow templates
	fmt.Println("\n4. Workflow Templates:")
	demonstrateTemplates()

	// Example 5: Custom script handlers
	fmt.Println("\n5. Custom Script Handlers:")
	demonstrateCustomHandlers()
}

func demonstrateScriptWorkflow() {
	// Create script steps
	validateStep, err := workflow.NewScriptStepBuilder("validate").
		WithLanguage("expr").
		WithScript("validated = true").
		WithDescription("Validate input data").
		WithTimeout(5 * time.Second).
		Build()
	if err != nil {
		log.Printf("Failed to create validate step: %v", err)
		return
	}

	transformStep, err := workflow.NewScriptStepBuilder("transform").
		WithLanguage("json-transform").
		WithScript(`{"result": "{{input}}", "timestamp": "2024-01-01"}`).
		WithDescription("Transform data to result format").
		Build()
	if err != nil {
		log.Printf("Failed to create transform step: %v", err)
		return
	}

	// Create workflow definition
	wf := &workflow.WorkflowDefinition{
		Name:        "Data Processing Workflow",
		Description: "Process data through validation and transformation",
		Steps: []workflow.WorkflowStep{
			validateStep,
			transformStep,
		},
	}

	// Execute the workflow
	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("input", "test-data")

	workflowState := &workflow.WorkflowState{
		State:    initialState,
		Metadata: make(map[string]interface{}),
	}

	fmt.Println("  Executing workflow steps:")
	for _, step := range wf.Steps {
		fmt.Printf("  - Running step: %s\n", step.Name())
		workflowState, err = step.Execute(ctx, workflowState)
		if err != nil {
			log.Printf("    Error: %v", err)
			continue
		}

		// Show state after step
		values := workflowState.Values()
		fmt.Printf("    State after step: %v\n", values)
	}
}

func demonstrateSerializationFormats() {
	// Create a workflow
	step, _ := workflow.NewScriptStepBuilder("example").
		WithLanguage("javascript").
		WithScript("return state").
		WithDescription("Example step").
		Build()

	wf := &workflow.WorkflowDefinition{
		Name:        "Example Workflow",
		Description: "Demonstrates serialization",
		Steps:       []workflow.WorkflowStep{step},
		Parallel:    false,
	}

	// Serialize to JSON
	jsonData, err := workflow.SerializeWorkflow(wf, "json")
	if err != nil {
		log.Printf("JSON serialization failed: %v", err)
		return
	}
	fmt.Println("  JSON format:")
	fmt.Printf("  %s\n", jsonData)

	// Serialize to pretty JSON
	prettyJSON, err := workflow.SerializeWorkflow(wf, "json-pretty")
	if err != nil {
		log.Printf("Pretty JSON serialization failed: %v", err)
		return
	}
	fmt.Println("\n  Pretty JSON format:")
	fmt.Printf("  %s\n", prettyJSON)

	// Serialize to YAML
	yamlData, err := workflow.SerializeWorkflow(wf, "yaml")
	if err != nil {
		log.Printf("YAML serialization failed: %v", err)
		return
	}
	fmt.Println("\n  YAML format:")
	fmt.Printf("  %s\n", yamlData)
}

func demonstrateBridgeDeserialization() {
	// Simulate data from bridge layer
	bridgeData := map[string]interface{}{
		"name":        "Bridge Workflow",
		"description": "Created from scripting engine",
		"version":     "1.0",
		"parallel":    false,
		"steps": []interface{}{
			map[string]interface{}{
				"name":        "step1",
				"type":        "script",
				"description": "First step",
				"script": map[string]interface{}{
					"language": "expr",
					"source":   "x = 10",
					"timeout":  "10s",
				},
			},
			map[string]interface{}{
				"name":        "step2",
				"type":        "script",
				"description": "Second step",
				"script": map[string]interface{}{
					"language": "json-transform",
					"source":   `{"x": "{{x}}", "y": 20}`,
				},
			},
		},
	}

	// Deserialize
	wf, err := workflow.DeserializeDefinition(bridgeData)
	if err != nil {
		log.Printf("Failed to deserialize: %v", err)
		return
	}

	fmt.Printf("  Deserialized workflow: %s\n", wf.Name)
	fmt.Printf("  Steps: %d\n", len(wf.Steps))
	for _, step := range wf.Steps {
		if scriptStep, ok := step.(*workflow.ScriptStep); ok {
			fmt.Printf("    - %s (%s): %s\n",
				scriptStep.Name(),
				scriptStep.Language(),
				scriptStep.Description())
		}
	}
}

func demonstrateTemplates() {
	// Register default templates
	if err := workflow.RegisterDefaultTemplates(); err != nil {
		log.Printf("Failed to register templates: %v", err)
		return
	}

	// List available templates
	templates := workflow.ListTemplates()
	fmt.Printf("  Available templates: %d\n", len(templates))
	for _, tmpl := range templates {
		fmt.Printf("    - %s: %s\n", tmpl.ID, tmpl.Name)
		fmt.Printf("      Category: %s\n", tmpl.Category)
		fmt.Printf("      Tags: %v\n", tmpl.Tags)
	}

	// Apply a template
	variables := map[string]interface{}{
		"input_source": "data.csv",
	}

	wf, err := workflow.ApplyTemplate("data-processing", variables)
	if err != nil {
		log.Printf("Failed to apply template: %v", err)
		return
	}

	fmt.Printf("\n  Applied template workflow: %s\n", wf.Name)
	fmt.Printf("  Steps: %d\n", len(wf.Steps))
}

func demonstrateCustomHandlers() {
	// Create a custom handler
	customHandler := &workflow.ScriptHandlerFunc{
		LanguageFn: func() string { return "custom" },
		ValidateFn: func(script string) error {
			return nil // Accept all scripts
		},
		ExecuteFn: func(ctx context.Context, state *workflow.WorkflowState, script string, env map[string]interface{}) (*workflow.WorkflowState, error) {
			// Custom logic
			newState := domain.NewState()
			newState.Set("custom_result", fmt.Sprintf("Processed: %s", script))

			return &workflow.WorkflowState{
				State:    newState,
				Metadata: state.Metadata,
			}, nil
		},
	}

	// Register the custom handler
	if err := workflow.RegisterScriptHandler("custom", customHandler); err != nil {
		log.Printf("Failed to register custom handler: %v", err)
		return
	}

	// Create a step using the custom handler
	customStep, err := workflow.NewScriptStep("custom-step", "custom", "do something special")
	if err != nil {
		log.Printf("Failed to create custom step: %v", err)
		return
	}

	// Execute the custom step
	ctx := context.Background()
	state := &workflow.WorkflowState{
		State:    domain.NewState(),
		Metadata: make(map[string]interface{}),
	}

	result, err := customStep.Execute(ctx, state)
	if err != nil {
		log.Printf("Custom step failed: %v", err)
		return
	}

	values := result.Values()
	fmt.Printf("  Custom handler result: %v\n", values["custom_result"])

	// Show registered languages
	languages := workflow.ListScriptLanguages()
	fmt.Printf("\n  Registered script languages: %v\n", languages)
}
