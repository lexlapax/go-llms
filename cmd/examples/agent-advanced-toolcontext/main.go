// ABOUTME: Demonstrates advanced ToolContext features with LLM agents calling tools
// ABOUTME: Shows event emission, progress reporting, retry handling, and state access

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: agent-advanced-toolcontext <example> [prompt]")
		fmt.Println("\nAvailable examples:")
		fmt.Println("  progress   - Tool with progress reporting")
		fmt.Println("  events     - Tool with event emission")
		fmt.Println("  retry      - Tool with retry handling")
		fmt.Println("  state      - Tool accessing agent state")
		fmt.Println("  all        - Demonstrate all features")
		fmt.Println("\nExample:")
		fmt.Println("  agent-advanced-toolcontext progress \"Process data with progress updates\"")
		os.Exit(1)
	}

	example := os.Args[1]

	// Get prompt from args or use default
	var prompt string
	if len(os.Args) > 2 {
		prompt = strings.Join(os.Args[2:], " ")
	}

	// Create LLM provider from environment
	llmProvider, err := createProvider()
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	// Run the selected example
	ctx := context.Background()

	switch example {
	case "progress":
		runProgressExample(ctx, llmProvider, prompt)
	case "events":
		runEventsExample(ctx, llmProvider, prompt)
	case "retry":
		runRetryExample(ctx, llmProvider, prompt)
	case "state":
		runStateAccessExample(ctx, llmProvider, prompt)
	case "all":
		runAllFeaturesExample(ctx, llmProvider, prompt)
	default:
		fmt.Printf("Unknown example: %s\n", example)
		os.Exit(1)
	}
}

func createProvider() (ldomain.Provider, error) {
	// Try to create provider from environment
	if os.Getenv("OPENAI_API_KEY") != "" {
		return provider.NewOpenAIProvider(
			os.Getenv("OPENAI_API_KEY"),
			"gpt-4o",
		), nil
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return provider.NewAnthropicProvider(
			os.Getenv("ANTHROPIC_API_KEY"),
			"claude-3-opus-20240229",
		), nil
	}

	if os.Getenv("GEMINI_API_KEY") != "" {
		return provider.NewGeminiProvider(
			os.Getenv("GEMINI_API_KEY"),
			"gemini-pro",
		), nil
	}

	// Try to create from GO_LLMS environment variables
	p, _, _, err := llmutil.ProviderFromEnv()
	if err == nil {
		return p, nil
	}

	return nil, fmt.Errorf("no LLM provider configured. Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY")
}

// Progress Reporting Example
func runProgressExample(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== Progress Reporting Example ===")

	// Create a tool that reports progress
	progressFunc := func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {

		// Parse parameters
		paramMap, ok := params.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid parameters")
		}

		dataSize := 100
		if size, ok := paramMap["size"].(float64); ok {
			dataSize = int(size)
		}

		results := []string{}

		// Process data with progress reporting
		for i := 0; i < dataSize; i++ {
			// Emit progress events
			if ctx.Events != nil {
				progress := (i + 1) * 100 / dataSize
				ctx.Events.EmitProgress(i+1, dataSize, fmt.Sprintf("Processing item %d of %d", i+1, dataSize))

				// Emit milestone events
				if progress%25 == 0 {
					ctx.Events.EmitMessage(fmt.Sprintf("Reached %d%% completion", progress))
				}
			}

			// Simulate processing
			time.Sleep(50 * time.Millisecond)
			results = append(results, fmt.Sprintf("Processed item %d", i+1))
		}

		// Emit completion event
		if ctx.Events != nil {
			ctx.Events.EmitMessage("Processing complete!")
		}

		return map[string]interface{}{
			"processed": dataSize,
			"results":   results[:5], // Return first 5 results
			"status":    "completed",
		}, nil
	}

	// Create the tool
	progressTool := atools.NewTool(
		"data_processor",
		"Processes data with progress reporting",
		progressFunc,
		nil, // Schema can be nil for simple tools
	)

	// Create agent with the tool
	agent := core.NewAgent("progress-demo", llmProvider)
	agent.AddTool(progressTool)
	agent.SetSystemPrompt(`You are a data processing assistant with a tool that reports progress.

Available tool:
- data_processor: Processes data with progress reporting. Parameters: {"size": number}

When asked to process data, use the tool and explain that progress is being reported.`)

	if customPrompt == "" {
		customPrompt = "Process 20 items of data and show me the progress"
	}

	runAgentWithEventMonitoring(ctx, agent, customPrompt)
}

// Event Emission Example
func runEventsExample(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== Event Emission Example ===")

	// Create a tool that emits various events
	eventFunc := func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {

		if ctx.Events != nil {
			// Emit different types of events
			ctx.Events.EmitMessage("Starting event generation...")

			ctx.Events.EmitCustom("debug", "Debug: Initializing event generator")

			ctx.Events.EmitCustom("warning", "Warning: This is a test warning")

			ctx.Events.EmitCustom("info", "Info: Processing events")

			// Custom event
			ctx.Events.EmitCustom("metrics", map[string]interface{}{
				"cpu_usage": 45.2,
				"memory_mb": 128,
				"requests":  1000,
			})

			// Nested events with progress
			for i := 0; i < 5; i++ {
				ctx.Events.EmitProgress(i+1, 5, "Generating event batch")
				time.Sleep(200 * time.Millisecond)
			}

			ctx.Events.EmitMessage("Event generation complete!")
		}

		return map[string]interface{}{
			"events_generated": 10,
			"status":           "success",
		}, nil
	}

	// Create the tool
	eventTool := atools.NewTool(
		"event_generator",
		"Generates various types of events",
		eventFunc,
		nil,
	)

	agent := core.NewAgent("events-demo", llmProvider)
	agent.AddTool(eventTool)
	agent.SetSystemPrompt(`You are an event monitoring assistant.

Available tool:
- event_generator: Generates various types of events for monitoring

When asked about events or monitoring, use this tool to demonstrate event generation.`)

	if customPrompt == "" {
		customPrompt = "Show me how events work in the system"
	}

	runAgentWithEventMonitoring(ctx, agent, customPrompt)
}

// Retry Handling Example
func runRetryExample(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== Retry Handling Example ===")

	// Create a tool that sometimes fails and uses retry info
	retryFunc := func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {

		// Check retry count
		retryCount := 0
		if ctx.Retry > 0 {
			retryCount = ctx.Retry
			if ctx.Events != nil {
				ctx.Events.EmitCustom("info", fmt.Sprintf("Retry attempt #%d", retryCount))
			}
		}

		// Simulate unreliable behavior using local random source
		// #nosec G404 - This is for demo purposes, not security-critical
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		// Higher chance of success with more retries
		successChance := 0.3 + float32(retryCount)*0.2
		if r.Float32() > successChance {
			if ctx.Events != nil {
				ctx.Events.EmitError(fmt.Errorf("service temporarily unavailable"))
			}
			return nil, fmt.Errorf("service temporarily unavailable (retry %d)", retryCount)
		}

		// Success!
		if ctx.Events != nil {
			ctx.Events.EmitMessage(fmt.Sprintf("Success after %d retries!", retryCount))
		}

		return map[string]interface{}{
			"data":    "Successfully retrieved data",
			"retries": retryCount,
			"status":  "success",
		}, nil
	}

	// Create the tool
	retryTool := atools.NewTool(
		"unreliable_service",
		"Calls an unreliable service that may fail",
		retryFunc,
		nil,
	)

	agent := core.NewAgent("retry-demo", llmProvider)
	agent.AddTool(retryTool)

	// Note: Retry configuration would be set at agent creation time
	// For this demo, we'll simulate retries in the tool itself

	agent.SetSystemPrompt(`You are a resilient service assistant.

Available tool:
- unreliable_service: Calls a service that may fail and require retries

When asked to call the service, use the tool. The system will automatically retry if it fails.`)

	if customPrompt == "" {
		customPrompt = "Call the unreliable service and get me some data"
	}

	runAgentWithEventMonitoring(ctx, agent, customPrompt)
}

// State Access Example
func runStateAccessExample(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== State Access Example ===")

	// Create a tool that accesses agent state
	stateFunc := func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {

		result := map[string]interface{}{
			"agent_info": map[string]interface{}{
				"name": ctx.Agent.Name,
				"id":   ctx.Agent.ID,
				"type": ctx.Agent.Type,
			},
			"run_id": ctx.RunID,
		}

		// Access state through StateReader
		if ctx.State != nil {
			// Read various state values
			if userName, exists := ctx.State.Get("user_name"); exists {
				result["user_name"] = userName
				if ctx.Events != nil {
					ctx.Events.EmitMessage(fmt.Sprintf("Found user: %v", userName))
				}
			}

			if preferences, exists := ctx.State.Get("preferences"); exists {
				result["preferences"] = preferences
			}

			// Count state entries
			stateCount := 0
			// Check known keys
			keys := []string{"user_name", "preferences", "context", "prompt", "analysis_type"}
			for _, key := range keys {
				if _, exists := ctx.State.Get(key); exists {
					stateCount++
					if ctx.Events != nil {
						ctx.Events.EmitCustom("debug", fmt.Sprintf("State key '%s' found", key))
					}
				}
			}
			result["state_entries"] = stateCount
		}

		return result, nil
	}

	// Create the tool
	stateTool := atools.NewTool(
		"state_inspector",
		"Inspects and uses agent state",
		stateFunc,
		nil,
	)

	agent := core.NewAgent("state-demo", llmProvider)
	agent.AddTool(stateTool)
	agent.SetSystemPrompt(`You are a context-aware assistant that can inspect state.

Available tool:
- state_inspector: Inspects agent and state information

When asked about context or state, use this tool to show what information is available.`)

	if customPrompt == "" {
		customPrompt = "What information do you have about me and the current context?"
	}

	// Run with pre-populated state
	state := domain.NewState()
	state.Set("user_name", "Alice")
	state.Set("preferences", map[string]interface{}{
		"language": "Go",
		"theme":    "dark",
	})
	state.Set("context", "demonstration")
	state.Set("prompt", customPrompt)

	runAgentWithState(ctx, agent, state)
}

// processAnalysisStep handles the step-specific logic for the analysis tool
func processAnalysisStep(step string, results map[string]interface{}, ctx *domain.ToolContext) {
	switch step {
	case "Data Collection":
		results["data_points"] = 1000
	case "Preprocessing":
		results["cleaned_records"] = 950
	case "Analysis":
		results["patterns_found"] = 5
		// Emit custom metrics
		if ctx.Events != nil {
			ctx.Events.EmitCustom("analysis_metrics", map[string]interface{}{
				"accuracy":           0.95,
				"confidence":         0.87,
				"processing_time_ms": 1500,
			})
		}
	case "Report Generation":
		results["report_sections"] = 3
	}
}

// All Features Example
func runAllFeaturesExample(ctx context.Context, llmProvider ldomain.Provider, customPrompt string) {
	fmt.Println("=== All Features Example ===")

	// Create a comprehensive tool that uses all features
	advancedFunc := func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {

		// 1. Check retry status
		if ctx.Retry > 0 && ctx.Events != nil {
			ctx.Events.EmitCustom("warning", fmt.Sprintf("Running retry attempt #%d", ctx.Retry))
		}

		// 2. Access state
		analysisType := "default"
		if ctx.State != nil {
			if aType, exists := ctx.State.Get("analysis_type"); exists {
				analysisType = fmt.Sprintf("%v", aType)
				if ctx.Events != nil {
					ctx.Events.EmitCustom("info", fmt.Sprintf("Using analysis type: %s", analysisType))
				}
			}
		}

		// 3. Emit various events
		if ctx.Events != nil {
			ctx.Events.EmitMessage("Starting comprehensive analysis...")
			ctx.Events.EmitCustom("debug", fmt.Sprintf("Agent: %s, RunID: %s", ctx.Agent.Name, ctx.RunID))
		}

		// 4. Progress reporting for multi-step analysis
		steps := []string{"Data Collection", "Preprocessing", "Analysis", "Report Generation"}
		results := make(map[string]interface{})

		for i, step := range steps {
			if ctx.Events != nil {
				ctx.Events.EmitProgress(i+1, len(steps), step)
			}

			// Simulate work
			time.Sleep(500 * time.Millisecond)

			// Add step results
			processAnalysisStep(step, results, ctx)
		}

		// 5. Final results with state info
		results["analysis_type"] = analysisType
		results["agent_id"] = ctx.Agent.ID
		results["completed_at"] = time.Now().Format(time.RFC3339)

		if ctx.Events != nil {
			ctx.Events.EmitMessage("Analysis complete!")
		}

		return results, nil
	}

	// Create the tool
	advancedTool := atools.NewTool(
		"advanced_analyzer",
		"Analyzes data using all ToolContext features",
		advancedFunc,
		nil,
	)

	agent := core.NewAgent("advanced-demo", llmProvider)
	agent.AddTool(advancedTool)
	agent.SetSystemPrompt(`You are an advanced analysis assistant with comprehensive tool capabilities.

Available tool:
- advanced_analyzer: Performs analysis using progress reporting, events, state access, and retry handling

Use this tool to demonstrate all the advanced features of the ToolContext.`)

	if customPrompt == "" {
		customPrompt = "Perform a comprehensive analysis and show me all the advanced features"
	}

	// Run with state
	state := domain.NewState()
	state.Set("analysis_type", "comprehensive")
	state.Set("user_level", "advanced")
	state.Set("prompt", customPrompt)

	runAgentWithState(ctx, agent, state)
}

// Helper function to run agent with event monitoring
func runAgentWithEventMonitoring(ctx context.Context, agent *core.LLMAgent, prompt string) {
	fmt.Printf("\nPrompt: %s\n", prompt)
	fmt.Println("\n--- Starting Execution ---")

	// Create event dispatcher for monitoring
	dispatcher := core.NewEventDispatcher(100)

	// Monitor events
	dispatcher.Subscribe(domain.EventHandlerFunc(func(event domain.Event) error {
		timestamp := time.Now().Format("15:04:05.000")
		switch event.Type {
		case domain.EventProgress:
			if data, ok := event.Data.(domain.ProgressEventData); ok {
				fmt.Printf("[%s] [PROGRESS] %d/%d - %s\n",
					timestamp, data.Current, data.Total, data.Message)
			}
		case domain.EventMessage:
			fmt.Printf("[%s] [MESSAGE] %v\n", timestamp, event.Data)
		case domain.EventToolError:
			fmt.Printf("[%s] [TOOL ERROR] %v\n", timestamp, event.Data)
		case domain.EventAgentError:
			fmt.Printf("[%s] [ERROR] %v\n", timestamp, event.Data)
		default:
			// Handle custom events from tools
			if strings.HasPrefix(string(event.Type), "tool.") {
				// Handle tool-specific custom events
				parts := strings.Split(string(event.Type), ".")
				if len(parts) >= 3 {
					fmt.Printf("[%s] [%s:%s] %v\n", timestamp, parts[1], parts[2], event.Data)
				} else {
					fmt.Printf("[%s] [%s] %v\n", timestamp, event.Type, event.Data)
				}
			} else {
				fmt.Printf("[%s] [%s] %v\n", timestamp, event.Type, event.Data)
			}
		}
		return nil
	}))

	// Attach dispatcher to agent
	agent.SetEventDispatcher(dispatcher)

	// Create initial state
	state := domain.NewState()
	state.Set("prompt", prompt)

	// Add enhanced hook
	agent.WithHook(&enhancedToolLogger{})

	// Run the agent
	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Give events time to flush
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n--- Execution Complete ---")

	// Display response
	if response, ok := result.Get("response"); ok {
		fmt.Printf("\nResponse:\n%v\n", response)
	} else if output, ok := result.Get("output"); ok {
		fmt.Printf("\nOutput:\n%v\n", output)
	} else if resultData, ok := result.Get("result"); ok {
		fmt.Printf("\nResult:\n%v\n", resultData)
	}
}

// Helper to run agent with pre-populated state
func runAgentWithState(ctx context.Context, agent *core.LLMAgent, state *domain.State) {
	prompt, _ := state.Get("prompt")
	fmt.Printf("\nPrompt: %v\n", prompt)
	fmt.Printf("Initial State: %d entries\n", len(state.Values()))
	fmt.Println("\n--- Starting Execution ---")

	// Create event dispatcher
	dispatcher := core.NewEventDispatcher(100)

	// Monitor events (same as above)
	dispatcher.Subscribe(domain.EventHandlerFunc(func(event domain.Event) error {
		timestamp := time.Now().Format("15:04:05.000")
		switch event.Type {
		case domain.EventProgress:
			if data, ok := event.Data.(domain.ProgressEventData); ok {
				fmt.Printf("[%s] [PROGRESS] %d/%d - %s\n",
					timestamp, data.Current, data.Total, data.Message)
			}
		case domain.EventMessage:
			fmt.Printf("[%s] [MESSAGE] %v\n", timestamp, event.Data)
		default:
			// Handle custom events from tools
			if strings.HasPrefix(string(event.Type), "tool.") {
				// Handle tool-specific custom events
				parts := strings.Split(string(event.Type), ".")
				if len(parts) >= 3 {
					fmt.Printf("[%s] [%s:%s] %v\n", timestamp, parts[1], parts[2], event.Data)
				} else {
					fmt.Printf("[%s] [%s] %v\n", timestamp, event.Type, event.Data)
				}
			} else {
				fmt.Printf("[%s] [%s] %v\n", timestamp, event.Type, event.Data)
			}
		}
		return nil
	}))

	agent.SetEventDispatcher(dispatcher)
	agent.WithHook(&enhancedToolLogger{})

	// Run with the provided state
	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println("\n--- Execution Complete ---")

	// Display response
	if response, ok := result.Get("response"); ok {
		fmt.Printf("\nResponse:\n%v\n", response)
	} else if output, ok := result.Get("output"); ok {
		fmt.Printf("\nOutput:\n%v\n", output)
	} else if resultData, ok := result.Get("result"); ok {
		fmt.Printf("\nResult:\n%v\n", resultData)
	}
}

// Enhanced tool logger that shows ToolContext usage
type enhancedToolLogger struct{}

func (h *enhancedToolLogger) BeforeGenerate(ctx context.Context, messages []ldomain.Message) {}

func (h *enhancedToolLogger) AfterGenerate(ctx context.Context, response ldomain.Response, err error) {
}

func (h *enhancedToolLogger) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	fmt.Printf("\n[TOOL CALL] Starting '%s'\n", tool)
	if len(params) > 0 {
		fmt.Printf("[TOOL CALL] Parameters: %v\n", params)
	}
}

func (h *enhancedToolLogger) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	if err != nil {
		fmt.Printf("[TOOL ERROR] '%s' failed: %v\n", tool, err)
	} else {
		fmt.Printf("[TOOL RESULT] '%s' completed\n", tool)
		if result != nil {
			// Pretty print result if it's a map
			if resultMap, ok := result.(map[string]interface{}); ok {
				for k, v := range resultMap {
					fmt.Printf("  - %s: %v\n", k, v)
				}
			} else {
				fmt.Printf("  Result: %v\n", result)
			}
		}
	}
}
