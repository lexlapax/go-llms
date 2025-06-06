package main

// ABOUTME: Example demonstrating complex workflow composition patterns
// ABOUTME: Shows how to build sophisticated workflows by composing simpler workflow agents

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Workflow Composition Examples ===")

	// Example 1: Nested workflows - workflows within workflows
	nestedWorkflowExample(ctx)

	// Example 2: Pipeline composition - chaining workflows
	pipelineExample(ctx)

	// Example 3: Complex orchestration - mixing workflow types
	complexOrchestrationExample(ctx)

	// Example 4: Dynamic workflow composition
	dynamicCompositionExample(ctx)
}

func nestedWorkflowExample(ctx context.Context) {
	fmt.Println("=== Nested Workflows Example ===")
	fmt.Println("Demonstrating workflows that contain other workflows...")

	// Create sub-workflow 1: Data validation workflow
	validationWorkflow := workflow.NewSequentialAgent("validation-workflow").
		AddAgent(createMockAgent("format-checker", "Checking data format...", 50*time.Millisecond)).
		AddAgent(createMockAgent("schema-validator", "Validating against schema...", 75*time.Millisecond))

	// Create sub-workflow 2: Data processing workflow
	processingWorkflow := workflow.NewParallelAgent("processing-workflow").
		WithMaxConcurrency(2).
		AddAgent(createMockAgent("data-cleaner", "Cleaning data...", 100*time.Millisecond)).
		AddAgent(createMockAgent("data-enricher", "Enriching data...", 150*time.Millisecond))

	// Create sub-workflow 3: Report generation workflow
	reportWorkflow := workflow.NewSequentialAgent("report-workflow").
		AddAgent(createMockAgent("summary-generator", "Generating summary...", 100*time.Millisecond)).
		AddAgent(createMockAgent("visualization-creator", "Creating visualizations...", 200*time.Millisecond))

	// Create the main workflow that orchestrates sub-workflows
	mainWorkflow := workflow.NewSequentialAgent("data-pipeline").
		AddAgent(validationWorkflow).
		AddAgent(processingWorkflow).
		AddAgent(reportWorkflow)

	// Run the nested workflow
	state := domain.NewState()
	state.Set("data", "Sample dataset for processing")

	fmt.Println("Starting nested workflow execution...")
	start := time.Now()

	_, err := mainWorkflow.Run(ctx, state)
	if err != nil {
		log.Printf("Workflow error: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("\nCompleted in %v\n", duration)

	// Show execution path
	status := mainWorkflow.Status()
	fmt.Println("\nExecution phases:")
	for step, stepStatus := range status.Steps {
		if stepStatus.State == workflow.StepStateCompleted {
			fmt.Printf("✓ %s (completed at %s)\n", step, stepStatus.EndTime.Format("15:04:05"))
		}
	}

	fmt.Println()
}

func pipelineExample(ctx context.Context) {
	fmt.Println("=== Pipeline Composition Example ===")
	fmt.Println("Building a complex pipeline from simple workflow components...")

	// Create a research pipeline with multiple stages

	// Stage 1: Information gathering (parallel)
	gatheringWorkflow := workflow.NewParallelAgent("info-gathering").
		AddAgent(createResearchAgent("web-searcher", "Searching web for information...")).
		AddAgent(createResearchAgent("doc-analyzer", "Analyzing documents...")).
		AddAgent(createResearchAgent("data-collector", "Collecting relevant data..."))

	// Stage 2: Analysis (conditional based on data type)
	analysisWorkflow := workflow.NewConditionalAgent("analysis").
		AddAgent("quantitative", func(s *domain.State) bool {
			if dataType, exists := s.Get("data_type"); exists {
				return strings.Contains(fmt.Sprintf("%v", dataType), "numeric")
			}
			return false
		}, createMockAgent("stats-analyzer", "Performing statistical analysis...", 200*time.Millisecond)).
		AddAgent("qualitative", func(s *domain.State) bool {
			if dataType, exists := s.Get("data_type"); exists {
				return strings.Contains(fmt.Sprintf("%v", dataType), "text")
			}
			return false
		}, createMockAgent("text-analyzer", "Performing text analysis...", 150*time.Millisecond)).
		SetDefaultAgent(createMockAgent("general-analyzer", "Performing general analysis...", 100*time.Millisecond))

	// Stage 3: Synthesis (sequential)
	synthesisWorkflow := workflow.NewSequentialAgent("synthesis").
		AddAgent(createMockAgent("insight-extractor", "Extracting key insights...", 100*time.Millisecond)).
		AddAgent(createMockAgent("conclusion-generator", "Generating conclusions...", 150*time.Millisecond))

	// Create the pipeline by composing workflows
	researchPipeline := workflow.NewSequentialAgent("research-pipeline").
		AddAgent(gatheringWorkflow).
		AddAgent(analysisWorkflow).
		AddAgent(synthesisWorkflow)

	// Test with different data types
	testCases := []struct {
		name     string
		dataType string
	}{
		{"Numeric Research", "numeric data"},
		{"Text Research", "text documents"},
		{"Mixed Research", "general information"},
	}

	for _, tc := range testCases {
		fmt.Printf("\nRunning pipeline for: %s\n", tc.name)

		state := domain.NewState()
		state.Set("research_topic", tc.name)
		state.Set("data_type", tc.dataType)

		start := time.Now()
		result, err := researchPipeline.Run(ctx, state)
		if err != nil {
			log.Printf("Pipeline error: %v", err)
			continue
		}

		duration := time.Since(start)

		if response, exists := result.Get("response"); exists {
			fmt.Printf("Result: %v (completed in %v)\n", response, duration)
		}
	}

	fmt.Println()
}

func complexOrchestrationExample(ctx context.Context) {
	fmt.Println("=== Complex Orchestration Example ===")
	fmt.Println("Demonstrating advanced workflow composition with mixed patterns...")

	// Create a complex document processing system

	// Initial validation loop (retry up to 3 times)
	validationLoop := workflow.WhileLoop("validation-loop",
		func(state *domain.State, iteration int) bool {
			// Continue while not validated and under 3 attempts
			if valid, exists := state.Get("validated"); exists && valid.(bool) {
				return false // Stop if validated
			}
			return iteration < 3 // Max 3 attempts
		},
		workflow.NewAgentStep("validator", createValidationAgent("document-validator")))

	// Parallel extraction workflows
	extractionWorkflow := workflow.NewParallelAgent("extraction").
		AddAgent(createMockAgent("metadata-extractor", "Extracting metadata...", 100*time.Millisecond)).
		AddAgent(createMockAgent("content-extractor", "Extracting content...", 150*time.Millisecond)).
		AddAgent(createMockAgent("reference-extractor", "Extracting references...", 120*time.Millisecond))

	// Conditional processing based on document type
	processingWorkflow := workflow.NewConditionalAgent("processing").
		AddAgent("pdf", func(s *domain.State) bool {
			return checkDocType(s, "pdf")
		}, createMockAgent("pdf-processor", "Processing PDF document...", 200*time.Millisecond)).
		AddAgent("docx", func(s *domain.State) bool {
			return checkDocType(s, "docx")
		}, createMockAgent("docx-processor", "Processing Word document...", 150*time.Millisecond)).
		SetDefaultAgent(createMockAgent("generic-processor", "Processing document...", 100*time.Millisecond))

	// Create retry workflow separately
	retryFlow := workflow.NewSequentialAgent("retry-flow").
		AddAgent(createMockAgent("quality-improver", "Improving quality...", 200*time.Millisecond)).
		AddAgent(createMockAgent("re-processor", "Reprocessing...", 150*time.Millisecond))

	// Quality check with conditional retry
	qualityCheckWorkflow := workflow.NewConditionalAgent("quality-check").
		AddAgent("needs-retry", func(s *domain.State) bool {
			if quality, exists := s.Get("quality_score"); exists {
				if score, ok := quality.(float64); ok {
					return score < 0.8
				}
			}
			return false
		}, retryFlow).
		SetDefaultAgent(createMockAgent("quality-approved", "Quality check passed!", 50*time.Millisecond))

	// Compose everything into a sophisticated document processing system
	documentSystem := workflow.NewSequentialAgent("document-processing-system").
		AddAgent(validationLoop).
		AddAgent(extractionWorkflow).
		AddAgent(processingWorkflow).
		AddAgent(qualityCheckWorkflow)

	// Test the system
	fmt.Println("Processing a complex document through the system...")

	state := domain.NewState()
	state.Set("document", "sample_document.pdf")
	state.Set("doc_type", "pdf")
	state.Set("validated", false) // Will be set to true by validator

	start := time.Now()
	_, err := documentSystem.Run(ctx, state)
	if err != nil {
		log.Printf("System error: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("\nDocument processing completed in %v\n", duration)

	// Show the execution flow
	status := documentSystem.Status()
	fmt.Println("\nExecution flow:")
	for step, stepStatus := range status.Steps {
		if stepStatus.State == workflow.StepStateCompleted {
			fmt.Printf("✓ %s\n", step)
		}
	}

	fmt.Println()
}

func dynamicCompositionExample(ctx context.Context) {
	fmt.Println("=== Dynamic Workflow Composition Example ===")
	fmt.Println("Building workflows dynamically based on configuration...")

	// Simulate loading workflow configuration
	workflowConfig := []struct {
		Type    string
		Name    string
		Agents  []string
		Options map[string]interface{}
	}{
		{
			Type:   "parallel",
			Name:   "data-processing",
			Agents: []string{"cleaner", "validator", "enricher"},
			Options: map[string]interface{}{
				"maxConcurrency": 2,
			},
		},
		{
			Type:   "sequential",
			Name:   "reporting",
			Agents: []string{"analyzer", "reporter"},
			Options: map[string]interface{}{
				"stopOnError": false,
			},
		},
	}

	// Build workflows dynamically from configuration
	var workflows []domain.BaseAgent

	for _, config := range workflowConfig {
		var wf domain.BaseAgent

		switch config.Type {
		case "parallel":
			parallelWf := workflow.NewParallelAgent(config.Name)
			if maxConc, ok := config.Options["maxConcurrency"].(int); ok {
				parallelWf.WithMaxConcurrency(maxConc)
			}
			for _, agentName := range config.Agents {
				parallelWf.AddAgent(createMockAgent(agentName,
					fmt.Sprintf("%s processing...", agentName), 100*time.Millisecond))
			}
			wf = parallelWf

		case "sequential":
			seqWf := workflow.NewSequentialAgent(config.Name)
			if stopOnErr, ok := config.Options["stopOnError"].(bool); ok {
				seqWf.WithStopOnError(stopOnErr)
			}
			for _, agentName := range config.Agents {
				seqWf.AddAgent(createMockAgent(agentName,
					fmt.Sprintf("%s processing...", agentName), 100*time.Millisecond))
			}
			wf = seqWf
		}

		if wf != nil {
			workflows = append(workflows, wf)
		}
	}

	// Create main workflow from dynamic components
	mainWorkflow := workflow.NewSequentialAgent("dynamic-workflow")
	for _, wf := range workflows {
		mainWorkflow.AddAgent(wf)
	}

	// Run the dynamically composed workflow
	fmt.Println("Running dynamically composed workflow...")

	state := domain.NewState()
	state.Set("input", "Dynamic workflow input data")

	start := time.Now()
	_, err := mainWorkflow.Run(ctx, state)
	if err != nil {
		log.Printf("Dynamic workflow error: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("\nDynamic workflow completed in %v\n", duration)

	// Show what was executed
	fmt.Println("\nDynamic workflow structure:")
	for i, config := range workflowConfig {
		fmt.Printf("%d. %s workflow '%s' with agents: %v\n",
			i+1, config.Type, config.Name, config.Agents)
	}
}

// Helper functions

func createMockAgent(name, response string, delay time.Duration) domain.BaseAgent {
	return &mockAgent{
		BaseAgent: core.NewBaseAgent(name, "Mock agent", domain.AgentTypeCustom),
		response:  response,
		delay:     delay,
	}
}

func createResearchAgent(name, response string) domain.BaseAgent {
	return &mockAgent{
		BaseAgent: core.NewBaseAgent(name, "Research agent", domain.AgentTypeCustom),
		response:  response,
		delay:     time.Duration(100+len(response)) * time.Millisecond, // Vary by response length
	}
}

func createValidationAgent(name string) domain.BaseAgent {
	return &validationAgent{
		BaseAgent: core.NewBaseAgent(name, "Validation agent", domain.AgentTypeCustom),
		attempts:  0,
	}
}

func checkDocType(state *domain.State, expectedType string) bool {
	if docType, exists := state.Get("doc_type"); exists {
		return strings.EqualFold(fmt.Sprintf("%v", docType), expectedType)
	}
	return false
}

// Mock agent implementations

type mockAgent struct {
	domain.BaseAgent
	response string
	delay    time.Duration
}

func (m *mockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Simulate processing
	select {
	case <-time.After(m.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	newState := state.Clone()
	newState.Set("response", m.response)

	// Add quality score for quality check workflow
	newState.Set("quality_score", 0.9) // High quality by default

	return newState, nil
}

type validationAgent struct {
	domain.BaseAgent
	attempts int
}

func (v *validationAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	v.attempts++

	newState := state.Clone()

	// Simulate validation - succeeds on 2nd attempt
	if v.attempts >= 2 {
		newState.Set("validated", true)
		newState.Set("response", "Document validated successfully!")
	} else {
		newState.Set("validated", false)
		newState.Set("response", fmt.Sprintf("Validation attempt %d failed, retrying...", v.attempts))
	}

	return newState, nil
}
