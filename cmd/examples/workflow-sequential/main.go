// ABOUTME: Example demonstrating sequential workflow execution with multiple LLM agents
// ABOUTME: Shows how to chain agents together for complex multi-step processing

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

func main() {
	// Create a research workflow that:
	// 1. Generates research questions
	// 2. Answers each question
	// 3. Summarizes findings

	ctx := context.Background()

	// Create agents
	questionGenerator, err := core.NewAgentFromString("question-generator", "claude")
	if err != nil {
		log.Fatalf("Failed to create question generator: %v", err)
	}
	questionGenerator.SetSystemPrompt("You are a research assistant. Generate 3 insightful questions about the given topic.")

	researcher, err := core.NewAgentFromString("researcher", "gpt-4")
	if err != nil {
		log.Fatalf("Failed to create researcher: %v", err)
	}
	researcher.SetSystemPrompt("You are an expert researcher. Answer the given questions thoroughly but concisely.")

	summarizer, err := core.NewAgentFromString("summarizer", "claude")
	if err != nil {
		log.Fatalf("Failed to create summarizer: %v", err)
	}
	summarizer.SetSystemPrompt("You are a skilled writer. Summarize the research findings into key insights.")

	// Create sequential workflow
	researchWorkflow := workflow.NewSequentialAgent("research-workflow").
		WithStopOnError(true).
		AddAgent(questionGenerator).
		AddAgent(researcher).
		AddAgent(summarizer)

	// Create initial state with research topic
	initialState := domain.NewState()
	initialState.Set("prompt", "quantum computing and its potential impact on cryptography")
	initialState.Set("instructions", "Research this topic by first generating questions, then answering them, and finally summarizing the findings.")

	// Run workflow
	fmt.Println("Starting research workflow...")
	fmt.Println("Topic: quantum computing and its potential impact on cryptography")
	fmt.Println("---")

	result, err := researchWorkflow.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	// Display results
	if response, exists := result.Get("response"); exists {
		fmt.Printf("Final Summary:\n%v\n", response)
	}

	// Show workflow status
	status := researchWorkflow.Status()
	fmt.Printf("\nWorkflow completed in: %v\n", status.EndTime.Sub(status.StartTime))
	fmt.Printf("Steps executed: %d\n", len(status.Steps))
}

// Alternative example with error handling
/*
func _exampleWithErrorHandling() {
	ctx := context.Background()

	// Create a workflow with custom error handling
	errorHandler := &workflow.DefaultErrorHandler{
		MaxRetries: 2,
	}

	// Create agents using mock provider for demonstration
	agent1 := _createMockAgent("data-fetcher", "Fetch data from source")
	agent2 := _createMockAgent("data-processor", "Process the fetched data")
	agent3 := _createMockAgent("report-generator", "Generate report from processed data")

	// Create workflow
	dataWorkflow := workflow.NewSequentialAgent("data-pipeline").
		WithStopOnError(false).
		WithMaxRetries(1)

	dataWorkflow.SetErrorHandler(errorHandler)
	dataWorkflow.AddAgent(agent1)
	dataWorkflow.AddAgent(agent2)
	dataWorkflow.AddAgent(agent3)

	// Run with error handling
	initialState := domain.NewState()
	initialState.Set("source", "database")

	_, err := dataWorkflow.Run(ctx, initialState)
	if err != nil {
		log.Printf("Workflow completed with errors: %v", err)
	} else {
		log.Println("Workflow completed successfully")
	}

	// Check which steps completed
	status := dataWorkflow.Status()
	for stepName, stepStatus := range status.Steps {
		fmt.Printf("Step %s: %s\n", stepName, stepStatus.State)
	}
}
*/

/*
func _createMockAgent(name, prompt string) domain.BaseAgent {
	// For demonstration - in real use, create actual LLM agents
	agent := core.NewBaseAgent(name, prompt, domain.AgentTypeCustom)
	return agent
}
*/
