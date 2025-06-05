// ABOUTME: Example demonstrating conditional workflow execution with branching logic
// ABOUTME: Shows if/else patterns, priority handling, and multiple match scenarios

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

func main() {
	// Example 1: Basic conditional workflow (if/else logic)
	basicConditionalExample()

	// Example 2: Priority-based condition evaluation
	priorityConditionalExample()

	// Example 3: Multiple matching branches
	multipleMatchExample()
}

func basicConditionalExample() {
	fmt.Println("=== Basic Conditional Workflow Example ===")
	fmt.Println("Demonstrating if/else logic with different processing paths...")
	fmt.Println()

	ctx := context.Background()

	// Create specialized agents for different data types
	var textProcessor domain.BaseAgent
	if tp, err := core.NewAgentFromString("text-processor", "claude"); err != nil {
		log.Printf("Using mock text processor: %v", err)
		textProcessor = createMockAgent("text-processor", "Text processing completed", 100*time.Millisecond)
	} else {
		tp.SetSystemPrompt("You are a text processing specialist. Analyze and process text data.")
		textProcessor = tp
	}

	var imageProcessor domain.BaseAgent
	if ip, err := core.NewAgentFromString("image-processor", "gpt-4-vision"); err != nil {
		log.Printf("Using mock image processor: %v", err)
		imageProcessor = createMockAgent("image-processor", "Image processing completed", 200*time.Millisecond)
	} else {
		ip.SetSystemPrompt("You are an image processing specialist. Analyze and process image data.")
		imageProcessor = ip
	}

	var dataProcessor domain.BaseAgent
	if dp, err := core.NewAgentFromString("data-processor", "gpt-4"); err != nil {
		log.Printf("Using mock data processor: %v", err)
		dataProcessor = createMockAgent("data-processor", "Data processing completed", 150*time.Millisecond)
	} else {
		dp.SetSystemPrompt("You are a data processing specialist. Analyze and process structured data.")
		dataProcessor = dp
	}

	// Create conditional workflow
	conditionalWorkflow := workflow.NewConditionalAgent("data-type-processor").
		AddAgent("text-branch", func(state *domain.State) bool {
			if dataType, exists := state.Get("data_type"); exists {
				return dataType == "text"
			}
			return false
		}, textProcessor).
		AddAgent("image-branch", func(state *domain.State) bool {
			if dataType, exists := state.Get("data_type"); exists {
				return dataType == "image"
			}
			return false
		}, imageProcessor).
		AddAgent("data-branch", func(state *domain.State) bool {
			if dataType, exists := state.Get("data_type"); exists {
				return dataType == "structured"
			}
			return false
		}, dataProcessor).
		SetDefaultAgent(createMockAgent("generic-processor", "Generic processing completed", 75*time.Millisecond))

	// Test different data types
	testCases := []struct {
		dataType string
		data     string
		prompt   string
	}{
		{
			"text",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			"Please analyze this text and provide a summary of its content and structure.",
		},
		{
			"image",
			"base64_encoded_image_data_here",
			"Please analyze this image data and describe what you can observe about its format and potential content.",
		},
		{
			"structured",
			`{"name": "John", "age": 30, "city": "New York"}`,
			"Please analyze this structured data and extract key information and insights.",
		},
		{
			"unknown",
			"some unknown data format",
			"Please analyze this data and try to determine its format and content.",
		},
	}

	for _, testCase := range testCases {
		fmt.Printf("Processing %s data...\n", testCase.dataType)

		// Create initial state with proper prompt for LLMAgent
		initialState := domain.NewState()
		initialState.Set("data_type", testCase.dataType)
		initialState.Set("data", testCase.data)
		initialState.Set("prompt", testCase.prompt)

		// Run workflow
		start := time.Now()
		result, err := conditionalWorkflow.Run(ctx, initialState)
		if err != nil {
			log.Printf("Workflow failed for %s: %v", testCase.dataType, err)
			continue
		}

		duration := time.Since(start)

		// Display results
		if response, exists := result.Get("response"); exists {
			fmt.Printf("Result: %v (took %v)\n", response, duration)
		} else {
			fmt.Printf("No specific processing result (took %v)\n", duration)
		}
		fmt.Println()
	}
}

func priorityConditionalExample() {
	fmt.Println("=== Priority-Based Conditional Example ===")
	fmt.Println("Demonstrating priority evaluation with multiple conditions...")
	fmt.Println()

	ctx := context.Background()

	// Create agents for different priority levels
	var criticalAgent domain.BaseAgent
	if ca, err := core.NewAgentFromString("critical-handler", "claude"); err != nil {
		log.Printf("Using mock critical agent: %v", err)
		criticalAgent = createMockAgent("critical-handler", "CRITICAL: Issue escalated to emergency team", 50*time.Millisecond)
	} else {
		ca.SetSystemPrompt("You are a critical issue handler. Provide emergency response for critical severity issues.")
		criticalAgent = ca
	}

	var highAgent domain.BaseAgent
	if ha, err := core.NewAgentFromString("high-handler", "gpt-4"); err != nil {
		log.Printf("Using mock high priority agent: %v", err)
		highAgent = createMockAgent("high-handler", "HIGH: Issue assigned to senior engineer", 100*time.Millisecond)
	} else {
		ha.SetSystemPrompt("You are a senior engineer. Handle high priority issues with detailed analysis.")
		highAgent = ha
	}

	var mediumAgent domain.BaseAgent
	if ma, err := core.NewAgentFromString("medium-handler", "claude"); err != nil {
		log.Printf("Using mock medium priority agent: %v", err)
		mediumAgent = createMockAgent("medium-handler", "MEDIUM: Issue added to team backlog", 75*time.Millisecond)
	} else {
		ma.SetSystemPrompt("You are a team lead. Triage medium priority issues and plan resolution.")
		mediumAgent = ma
	}

	var lowAgent domain.BaseAgent
	if la, err := core.NewAgentFromString("low-handler", "gpt-4"); err != nil {
		log.Printf("Using mock low priority agent: %v", err)
		lowAgent = createMockAgent("low-handler", "LOW: Issue logged for future review", 25*time.Millisecond)
	} else {
		la.SetSystemPrompt("You are a support engineer. Log and categorize low priority issues.")
		lowAgent = la
	}

	// Create conditional workflow with priorities
	priorityWorkflow := workflow.NewConditionalAgent("issue-triage").
		AddBranchWithPriority("critical", func(state *domain.State) bool {
			if severity, exists := state.Get("severity"); exists {
				return severity.(int) >= 9
			}
			return false
		}, createAgentStep("critical", criticalAgent), 100).
		AddBranchWithPriority("high", func(state *domain.State) bool {
			if severity, exists := state.Get("severity"); exists {
				return severity.(int) >= 7
			}
			return false
		}, createAgentStep("high", highAgent), 75).
		AddBranchWithPriority("medium", func(state *domain.State) bool {
			if severity, exists := state.Get("severity"); exists {
				return severity.(int) >= 4
			}
			return false
		}, createAgentStep("medium", mediumAgent), 50).
		AddBranchWithPriority("low", func(state *domain.State) bool {
			if severity, exists := state.Get("severity"); exists {
				return severity.(int) >= 1
			}
			return false
		}, createAgentStep("low", lowAgent), 25)

	// Test different severity levels
	severityLevels := []int{10, 8, 5, 2, 0}

	for _, severity := range severityLevels {
		fmt.Printf("Processing issue with severity %d...\n", severity)

		// Create initial state with proper prompt for LLMAgent
		issueDescription := fmt.Sprintf("System issue with severity level %d", severity)
		prompt := fmt.Sprintf("Please handle this issue: %s. Provide appropriate response based on the severity level.", issueDescription)

		initialState := domain.NewState()
		initialState.Set("severity", severity)
		initialState.Set("issue", issueDescription)
		initialState.Set("prompt", prompt)

		// Run workflow
		start := time.Now()
		result, err := priorityWorkflow.Run(ctx, initialState)
		if err != nil {
			log.Printf("Workflow failed for severity %d: %v", severity, err)
			continue
		}

		duration := time.Since(start)

		// Display results
		if response, exists := result.Get("response"); exists {
			fmt.Printf("Result: %v (took %v)\n", response, duration)
		} else {
			fmt.Printf("No handler matched severity %d (took %v)\n", severity, duration)
		}
		fmt.Println()
	}
}

func multipleMatchExample() {
	fmt.Println("=== Multiple Match Conditional Example ===")
	fmt.Println("Demonstrating workflows that allow multiple branches to execute...")
	fmt.Println()

	ctx := context.Background()

	// Create validation agents
	var syntaxChecker domain.BaseAgent
	if sc, err := core.NewAgentFromString("syntax-checker", "claude"); err != nil {
		log.Printf("Using mock syntax checker: %v", err)
		syntaxChecker = createMockAgent("syntax-checker", "Syntax validation passed", 50*time.Millisecond)
	} else {
		sc.SetSystemPrompt("You are a syntax validator. Check code for syntax errors and provide detailed feedback.")
		syntaxChecker = sc
	}

	var securityChecker domain.BaseAgent
	if sec, err := core.NewAgentFromString("security-checker", "gpt-4"); err != nil {
		log.Printf("Using mock security checker: %v", err)
		securityChecker = createMockAgent("security-checker", "Security scan completed", 100*time.Millisecond)
	} else {
		sec.SetSystemPrompt("You are a security expert. Scan code for security vulnerabilities and provide recommendations.")
		securityChecker = sec
	}

	var performanceChecker domain.BaseAgent
	if pc, err := core.NewAgentFromString("performance-checker", "claude"); err != nil {
		log.Printf("Using mock performance checker: %v", err)
		performanceChecker = createMockAgent("performance-checker", "Performance analysis completed", 150*time.Millisecond)
	} else {
		pc.SetSystemPrompt("You are a performance analyst. Analyze code for performance issues and optimization opportunities.")
		performanceChecker = pc
	}

	var compatibilityChecker domain.BaseAgent
	if cc, err := core.NewAgentFromString("compatibility-checker", "gpt-4"); err != nil {
		log.Printf("Using mock compatibility checker: %v", err)
		compatibilityChecker = createMockAgent("compatibility-checker", "Compatibility check passed", 75*time.Millisecond)
	} else {
		cc.SetSystemPrompt("You are a compatibility expert. Check code for compatibility issues across different platforms and versions.")
		compatibilityChecker = cc
	}

	// Create conditional workflow that allows multiple matches
	validationWorkflow := workflow.NewConditionalAgent("code-validation").
		WithAllowMultipleMatches(true).
		WithEvaluateAllConditions(true).
		AddBranchWithPriority("syntax", func(state *domain.State) bool {
			if checks, exists := state.Get("required_checks"); exists {
				checkList := checks.([]string)
				for _, check := range checkList {
					if check == "syntax" {
						return true
					}
				}
			}
			return false
		}, createAgentStep("syntax", syntaxChecker), 100).
		AddBranchWithPriority("security", func(state *domain.State) bool {
			if checks, exists := state.Get("required_checks"); exists {
				checkList := checks.([]string)
				for _, check := range checkList {
					if check == "security" {
						return true
					}
				}
			}
			return false
		}, createAgentStep("security", securityChecker), 90).
		AddBranchWithPriority("performance", func(state *domain.State) bool {
			if checks, exists := state.Get("required_checks"); exists {
				checkList := checks.([]string)
				for _, check := range checkList {
					if check == "performance" {
						return true
					}
				}
			}
			return false
		}, createAgentStep("performance", performanceChecker), 80).
		AddBranchWithPriority("compatibility", func(state *domain.State) bool {
			if checks, exists := state.Get("required_checks"); exists {
				checkList := checks.([]string)
				for _, check := range checkList {
					if check == "compatibility" {
						return true
					}
				}
			}
			return false
		}, createAgentStep("compatibility", compatibilityChecker), 70)

	// Test different validation scenarios
	testScenarios := []struct {
		name   string
		checks []string
	}{
		{"Basic validation", []string{"syntax"}},
		{"Security review", []string{"syntax", "security"}},
		{"Full validation", []string{"syntax", "security", "performance", "compatibility"}},
		{"Performance focus", []string{"performance", "compatibility"}},
	}

	for _, scenario := range testScenarios {
		fmt.Printf("Running %s...\n", scenario.name)

		// Create initial state with proper prompt for LLMAgent
		codeExample := `
func calculateTotal(items []Item) float64 {
    total := 0.0
    for _, item := range items {
        total += item.Price * item.Quantity
    }
    return total
}`
		prompt := fmt.Sprintf("Please perform %s validation on this Go code: %s", scenario.name, codeExample)

		initialState := domain.NewState()
		initialState.Set("required_checks", scenario.checks)
		initialState.Set("code", codeExample)
		initialState.Set("prompt", prompt)

		// Run workflow
		start := time.Now()
		result, err := validationWorkflow.Run(ctx, initialState)
		if err != nil {
			log.Printf("Workflow failed for %s: %v", scenario.name, err)
			continue
		}

		duration := time.Since(start)

		// Display results
		fmt.Printf("Validation completed in %v\n", duration)
		if response, exists := result.Get("response"); exists {
			fmt.Printf("Final result: %v\n", response)
		}

		// Show workflow status
		status := validationWorkflow.Status()
		completedSteps := 0
		for _, stepStatus := range status.Steps {
			if stepStatus.State == workflow.StepStateCompleted {
				completedSteps++
			}
		}
		fmt.Printf("Completed %d validation steps\n", completedSteps)
		fmt.Println()
	}
}

// Helper functions
func createMockAgent(name, response string, delay time.Duration) domain.BaseAgent {
	agent := &mockAgent{
		BaseAgent: core.NewBaseAgent(name, "Mock agent", domain.AgentTypeCustom),
		delay:     delay,
		response:  response,
	}
	return agent
}

// Helper function to create agent steps
func createAgentStep(name string, agent domain.BaseAgent) workflow.WorkflowStep {
	return workflow.NewAgentStep(name, agent)
}

type mockAgent struct {
	domain.BaseAgent
	delay    time.Duration
	response string
}

func (m *mockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	newState := state.Clone()
	newState.Set("response", m.response)

	return newState, nil
}
