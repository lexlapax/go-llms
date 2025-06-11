// ABOUTME: Example demonstrating the use of the built-in calculator tool
// ABOUTME: Shows direct tool usage and LLM agent integration with the calculator

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	toolmath "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// Helper to create minimal tool context for testing
func createToolContext() *agentDomain.ToolContext {
	ctx := context.Background()
	agentInfo := agentDomain.AgentInfo{
		ID:          "test-agent",
		Name:        "Calculator Test",
		Description: "Testing calculator tool",
		Type:        agentDomain.AgentTypeCustom,
	}

	state := agentDomain.NewState()
	stateReader := agentDomain.NewStateReader(state)

	return &agentDomain.ToolContext{
		Context: ctx,
		State:   stateReader,
		Agent:   agentInfo,
		RunID:   "test-run-001",
	}
}

func main() {
	// Check for command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "direct":
			// Run direct tool usage example
			runDirectExample()
			return
		case "info":
			// Show tool information
			showToolInfo()
			return
		}
	}

	// Default: Run LLM integration example
	runLLMExample()
}

func showToolInfo() {
	fmt.Println("=== Calculator Tool Information ===")
	fmt.Println()

	// Get the calculator tool from registry
	calculator, ok := tools.GetTool("calculator")
	if !ok {
		log.Fatalf("Failed to get calculator tool")
	}

	fmt.Printf("Tool Name: %s\n", calculator.Name())
	fmt.Printf("Description: %s\n", calculator.Description())
	fmt.Printf("Category: %s\n", calculator.Category())
	fmt.Printf("Version: %s\n", calculator.Version())
	fmt.Printf("Deterministic: %v\n", calculator.IsDeterministic())
	fmt.Printf("Destructive: %v\n", calculator.IsDestructive())
	fmt.Printf("Requires Confirmation: %v\n", calculator.RequiresConfirmation())
	fmt.Printf("Estimated Latency: %s\n", calculator.EstimatedLatency())
	fmt.Printf("Tags: %v\n", calculator.Tags())

	fmt.Println("\nUsage Instructions:")
	fmt.Println(calculator.UsageInstructions())

	fmt.Println("\nConstraints:")
	for _, constraint := range calculator.Constraints() {
		fmt.Printf("- %s\n", constraint)
	}

	fmt.Println("\nExamples:")
	for i, example := range calculator.Examples() {
		fmt.Printf("%d. %s - %s\n", i+1, example.Name, example.Description)
		fmt.Printf("   Input: %v\n", example.Input)
		fmt.Printf("   Output: %v\n", example.Output)
		if example.Explanation != "" {
			fmt.Printf("   Explanation: %s\n", example.Explanation)
		}
		fmt.Println()
	}
}

func runDirectExample() {
	fmt.Println("=== Built-in Calculator Tool Example (Direct Usage) ===")
	fmt.Println()

	// Get the calculator tool from registry
	calculator, ok := tools.GetTool("calculator")
	if !ok {
		log.Fatalf("Failed to get calculator tool")
	}

	fmt.Printf("Tool: %s\n", calculator.Name())
	fmt.Printf("Description: %s\n\n", calculator.Description())

	// Create tool context
	toolCtx := createToolContext()

	// Example 1: Basic Arithmetic
	fmt.Println("--- Basic Arithmetic ---")

	// Addition
	addResult, addErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "add",
		"operand1":  10.5,
		"operand2":  5.2,
	})
	if addErr != nil {
		log.Printf("Addition failed: %v", addErr)
	} else {
		printResult("10.5 + 5.2", addResult)
	}

	// Division
	divResult, divErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "divide",
		"operand1":  20.0,
		"operand2":  4.0,
	})
	if divErr != nil {
		log.Printf("Division failed: %v", divErr)
	} else {
		printResult("20 / 4", divResult)
	}

	// Power
	powerResult, powerErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "power",
		"operand1":  2.0,
		"operand2":  8.0,
	})
	if powerErr != nil {
		log.Printf("Power failed: %v", powerErr)
	} else {
		printResult("2^8", powerResult)
	}

	// Example 2: Square Root and Logarithms
	fmt.Println("\n--- Square Root and Logarithms ---")

	// Square root
	sqrtResult, sqrtErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "sqrt",
		"operand1":  144.0,
	})
	if sqrtErr != nil {
		log.Printf("Square root failed: %v", sqrtErr)
	} else {
		printResult("√144", sqrtResult)
	}

	// Natural logarithm
	logResult, logErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "log",
		"operand1":  2.718281828,
	})
	if logErr != nil {
		log.Printf("Natural log failed: %v", logErr)
	} else {
		printResult("ln(e)", logResult)
	}

	// Example 3: Trigonometry (radians)
	fmt.Println("\n--- Trigonometry ---")

	// Pi constant
	piResult, piErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "pi",
	})
	if piErr != nil {
		log.Printf("Pi failed: %v", piErr)
	} else {
		printResult("π", piResult)
	}

	// Sine of pi/2
	sinResult, sinErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "sin",
		"operand1":  1.5707963267948966, // pi/2
	})
	if sinErr != nil {
		log.Printf("Sine failed: %v", sinErr)
	} else {
		printResult("sin(π/2)", sinResult)
	}

	// Example 4: Advanced Operations
	fmt.Println("\n--- Advanced Operations ---")

	// Factorial
	factorialResult, factErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "factorial",
		"operand1":  5.0,
	})
	if factErr != nil {
		log.Printf("Factorial failed: %v", factErr)
	} else {
		printResult("5!", factorialResult)
	}

	// GCD
	gcdResult, gcdErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "gcd",
		"operand1":  48.0,
		"operand2":  18.0,
	})
	if gcdErr != nil {
		log.Printf("GCD failed: %v", gcdErr)
	} else {
		printResult("GCD(48, 18)", gcdResult)
	}

	// Example 5: Error Handling
	fmt.Println("\n--- Error Handling ---")

	// Division by zero
	divZeroResult, divZeroErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "divide",
		"operand1":  10.0,
		"operand2":  0.0,
	})
	if divZeroErr != nil {
		fmt.Printf("Division by zero: %v (expected)\n", divZeroErr)
	} else {
		printResult("10 / 0", divZeroResult)
	}

	// Square root of negative number
	sqrtNegResult, sqrtNegErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "sqrt",
		"operand1":  -16.0,
	})
	if sqrtNegErr != nil {
		fmt.Printf("√(-16): %v (expected)\n", sqrtNegErr)
	} else {
		printResult("√(-16)", sqrtNegResult)
	}

	// Display all available operations
	fmt.Println("\n--- Available Operations ---")
	fmt.Println("Basic: add(+), subtract(-), multiply(*), divide(/), power(^,**), mod(%), abs")
	fmt.Println("Roots/Logs: sqrt, cbrt, log, log10, log2, exp")
	fmt.Println("Trigonometry: sin, cos, tan, asin, acos, atan, sinh, cosh, tanh")
	fmt.Println("Rounding: floor, ceil, round")
	fmt.Println("Advanced: factorial, gcd, lcm")
	fmt.Println("Constants: pi, e, phi, tau, sqrt2, sqrte, sqrtpi, sqrtphi, ln2, ln10, log2e, log10e")
}

func printResult(operation string, result interface{}) {
	// Import the correct type from the math package
	if calcResult, ok := result.(*toolmath.CalculatorResult); ok {
		if calcResult.Success {
			fmt.Printf("%s = %.6f\n", operation, calcResult.Result)
		} else {
			fmt.Printf("%s = ERROR: %s\n", operation, calcResult.Error)
		}
	} else {
		fmt.Printf("%s = %v\n", operation, result)
	}
}

// runLLMExample demonstrates using the calculator tool with an LLM agent
func runLLMExample() {
	fmt.Println("=== Built-in Calculator Tool with LLM Agent ===")
	fmt.Println()

	ctx := context.Background()

	// Create LLM provider using provider/model string
	providerString := "anthropic/claude-3-7-sonnet-latest" // Default to Claude
	if os.Getenv("OPENAI_API_KEY") != "" && os.Getenv("ANTHROPIC_API_KEY") == "" {
		providerString = "openai/gpt-4o"
	} else if os.Getenv("GEMINI_API_KEY") != "" && os.Getenv("ANTHROPIC_API_KEY") == "" {
		providerString = "gemini/gemini-2.0-flash"
	}

	provider, err := llmutil.NewProviderFromString(providerString)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Parse provider string to get provider and model info
	providerName, modelName, _ := llmutil.ParseProviderModelString(providerString)

	// Print provider information
	fmt.Printf("Provider: %s\n", providerName)
	if modelName != "" {
		fmt.Printf("Model: %s\n\n", modelName)
	} else {
		fmt.Printf("Model: (default for provider)\n\n")
	}

	// Get calculator tool from registry
	calculator, ok := tools.GetTool("calculator")
	if !ok {
		log.Fatalf("Failed to get calculator tool")
	}

	// Create LLM agent with calculator tool
	deps := core.LLMDeps{
		Provider: provider,
	}
	agent := core.NewLLMAgent("math-assistant", "Math Assistant with Calculator", deps)

	// Add logging hooks if DEBUG is enabled
	if os.Getenv("DEBUG") == "1" {
		// Create slog logger that outputs to stderr
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
		loggingHook := core.NewLoggingHook(logger, core.LogLevelDebug)
		agent.WithHook(loggingHook)
		log.Println("Debug logging enabled")
	}

	// Add the calculator tool
	agent.AddTool(calculator)

	// Set system prompt that leverages the tool's built-in documentation
	agent.SetSystemPrompt(`You are a helpful math assistant. You MUST use the calculator tool to perform ALL calculations when requested.

The calculator tool provides comprehensive mathematical capabilities. When asked to perform any calculation:
1. Always use the calculator tool
2. Do not attempt to calculate results yourself
3. Use the tool's output to provide accurate results

IMPORTANT: You must actually use the calculator tool for every calculation. Do not just describe what you would do - actually do it.

The tool will guide you on proper usage, including which operations need one or two operands.`)

	// Example prompts showcasing different calculator features
	examples := []struct {
		title  string
		prompt string
	}{
		{
			title:  "Basic Arithmetic",
			prompt: "What is 25 * 17?",
		},
		{
			title:  "Square Root",
			prompt: "Calculate the square root of 144",
		},
		{
			title:  "Power Operation",
			prompt: "What is 2 to the power of 10?",
		},
		{
			title:  "Trigonometry",
			prompt: "Find the sine of pi/2 radians",
		},
		{
			title:  "Advanced Math - Factorial",
			prompt: "Calculate 15! (factorial)",
		},
		{
			title:  "Complex Calculation",
			prompt: "Calculate (phi squared) minus (sqrt(5))",
		},
	}

	// Run examples
	for i, example := range examples {
		fmt.Printf("\n=== Example %d: %s ===\n", i+1, example.title)

		// Create state with the prompt
		state := agentDomain.NewState()
		state.Set("user_input", example.prompt)

		// Run the agent
		result, runErr := agent.Run(ctx, state)
		if runErr != nil {
			log.Printf("Error: %v", runErr)
			continue
		}

		// Extract and display the response
		printLastMessage(result)
	}

	fmt.Println("\n=== Instructions ===")
	fmt.Println("To run this example:")
	fmt.Println("1. Default (LLM integration): go run main.go")
	fmt.Println("2. Direct tool usage: go run main.go direct")
	fmt.Println("3. Tool information: go run main.go info")
	fmt.Println("\nEnvironment variables:")
	fmt.Println("- Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY for real LLM")
	fmt.Println("- Set DEBUG=1 to enable detailed logging")
}

func printLastMessage(state *agentDomain.State) {
	// Try to get the response from various possible keys
	responseKeys := []string{"response", "output", "result", "answer", "reply"}

	for _, key := range responseKeys {
		if value, exists := state.Get(key); exists {
			fmt.Printf("Response: %v\n", value)
			return
		}
	}

	// If no response found in common keys, check messages
	if messages := state.Messages(); len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		fmt.Printf("Response: %v\n", lastMsg.Content)
		return
	}

	// Last resort: print available keys for debugging
	fmt.Printf("No response found. Available keys: %v\n", state.Keys())
}
