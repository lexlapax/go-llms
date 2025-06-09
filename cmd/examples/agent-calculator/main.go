// ABOUTME: Example demonstrating the use of the built-in calculator tool
// ABOUTME: Shows direct tool usage and LLM agent integration with the calculator

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
	toolmath "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
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
	if len(os.Args) > 1 && os.Args[1] == "llm" {
		// Run LLM integration example
		runLLMExample()
	} else {
		// Run direct tool usage example
		runDirectExample()
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

	// Example 2: Scientific Functions
	fmt.Println("\n--- Scientific Functions ---")

	// Square root
	sqrtResult, sqrtErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "sqrt",
		"operand1":  16.0,
	})
	if sqrtErr != nil {
		log.Printf("Square root failed: %v", sqrtErr)
	} else {
		printResult("√16", sqrtResult)
	}

	// Natural logarithm
	logResult, logErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "log",
		"operand1":  math.E,
	})
	if logErr != nil {
		log.Printf("Natural log failed: %v", logErr)
	} else {
		printResult("ln(e)", logResult)
	}

	// Logarithm base 2
	log2Result, log2Err := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "log",
		"operand1":  8.0,
		"operand2":  2.0,
	})
	if log2Err != nil {
		log.Printf("Log base 2 failed: %v", log2Err)
	} else {
		printResult("log₂(8)", log2Result)
	}

	// Example 3: Trigonometry
	fmt.Println("\n--- Trigonometry ---")

	// Sin(π/2)
	sinResult, sinErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "sin",
		"operand1":  math.Pi / 2,
	})
	if sinErr != nil {
		log.Printf("Sine failed: %v", sinErr)
	} else {
		printResult("sin(π/2)", sinResult)
	}

	// Cos(π)
	cosResult, cosErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "cos",
		"operand1":  math.Pi,
	})
	if cosErr != nil {
		log.Printf("Cosine failed: %v", cosErr)
	} else {
		printResult("cos(π)", cosResult)
	}

	// Arcsin(1)
	asinResult, asinErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "asin",
		"operand1":  1.0,
	})
	if asinErr != nil {
		log.Printf("Arcsin failed: %v", asinErr)
	} else {
		printResult("arcsin(1) in radians", asinResult)
	}

	// Example 4: Advanced Operations
	fmt.Println("\n--- Advanced Operations ---")

	// Factorial
	factResult, factErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "factorial",
		"operand1":  5.0,
	})
	if factErr != nil {
		log.Printf("Factorial failed: %v", factErr)
	} else {
		printResult("5!", factResult)
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

	// LCM
	lcmResult, lcmErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "lcm",
		"operand1":  12.0,
		"operand2":  18.0,
	})
	if lcmErr != nil {
		log.Printf("LCM failed: %v", lcmErr)
	} else {
		printResult("LCM(12, 18)", lcmResult)
	}

	// Example 5: Mathematical Constants
	fmt.Println("\n--- Mathematical Constants ---")

	// Pi
	piResult, piErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "pi",
	})
	if piErr != nil {
		log.Printf("Pi failed: %v", piErr)
	} else {
		printResult("π", piResult)
	}

	// E
	eResult, eErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "e",
	})
	if eErr != nil {
		log.Printf("E failed: %v", eErr)
	} else {
		printResult("e", eResult)
	}

	// Example 6: Error Handling
	fmt.Println("\n--- Error Handling ---")

	// Division by zero
	divZeroResult, divZeroErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "divide",
		"operand1":  10.0,
		"operand2":  0.0,
	})
	if divZeroErr != nil {
		log.Printf("Expected error for division by zero: %v", divZeroErr)
	} else {
		printResult("10 / 0", divZeroResult)
	}

	// Square root of negative
	negSqrtResult, negSqrtErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "sqrt",
		"operand1":  -4.0,
	})
	if negSqrtErr != nil {
		log.Printf("Expected error for sqrt(-4): %v", negSqrtErr)
	} else {
		printResult("√(-4)", negSqrtResult)
	}

	// Invalid operation
	unkResult, unkErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "unknown",
		"operand1":  1.0,
		"operand2":  2.0,
	})
	if unkErr != nil {
		log.Printf("Expected error for unknown operation: %v", unkErr)
	} else {
		printResult("unknown(1, 2)", unkResult)
	}

	// Example 7: Additional Constants
	fmt.Println("\n--- Additional Mathematical Constants ---")

	// Golden ratio
	phiResult, phiErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "phi",
	})
	if phiErr != nil {
		log.Printf("Phi failed: %v", phiErr)
	} else {
		printResult("φ (golden ratio)", phiResult)
	}

	// Calculate using phi
	phiCalc, phiCalcErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "multiply",
		"operand1":  "phi",
		"operand2":  "phi",
	})
	if phiCalcErr != nil {
		log.Printf("Phi squared failed: %v", phiCalcErr)
	} else {
		printResult("φ² (phi squared)", phiCalc)
	}

	// Tau
	tauResult, tauErr := calculator.Execute(toolCtx, map[string]interface{}{
		"operation": "tau",
	})
	if tauErr != nil {
		log.Printf("Tau failed: %v", tauErr)
	} else {
		printResult("τ (tau = 2π)", tauResult)
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
	if os.Getenv("OPENAI_API_KEY") != "" {
		providerString = "openai/gpt-4o"
	} else if os.Getenv("GEMINI_API_KEY") != "" {
		providerString = "gemini/gemini-2.0-flash"
	}

	provider, err := llmutil.NewProviderFromString(providerString)
	if err != nil {
		// Fall back to mock provider for demonstration
		fmt.Println("Note: No LLM API keys found. Using mock provider for demonstration.")
		fmt.Println("The mock will simulate calculator tool usage.")
		fmt.Println("Tip: Set DEBUG=1 to see detailed logging of agent execution.")
		fmt.Println()
		provider = createMockProvider()
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
			title:  "Advanced Math - GCD",
			prompt: "What is the GCD of 48 and 18?",
		},
		{
			title:  "Logarithms",
			prompt: "Calculate log base 2 of 64",
		},
		{
			title:  "Using Constants",
			prompt: "What is pi times e?",
		},
		{
			title:  "Error Handling",
			prompt: "What is the square root of -16?",
		},
		{
			title:  "Complex Calculation",
			prompt: "Calculate (phi squared) minus (sqrt(5))",
		},
	}

	// Display tool metadata if requested
	if len(os.Args) > 2 && os.Args[2] == "info" {
		fmt.Println("=== Calculator Tool Information ===")
		fmt.Printf("Name: %s\n", calculator.Name())
		fmt.Printf("Description: %s\n", calculator.Description())
		fmt.Printf("Version: %s\n", calculator.Version())
		fmt.Printf("Category: %s\n", calculator.Category())
		fmt.Printf("Tags: %v\n", calculator.Tags())
		fmt.Printf("Deterministic: %v\n", calculator.IsDeterministic())
		fmt.Printf("Destructive: %v\n", calculator.IsDestructive())
		fmt.Printf("Requires Confirmation: %v\n", calculator.RequiresConfirmation())
		fmt.Printf("Estimated Latency: %s\n", calculator.EstimatedLatency())
		fmt.Println("\nUsage Instructions:")
		fmt.Println(calculator.UsageInstructions())
		fmt.Println("\nConstraints:")
		for _, c := range calculator.Constraints() {
			fmt.Printf("- %s\n", c)
		}
		fmt.Println("\nExamples available:", len(calculator.Examples()))
		return
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
	fmt.Println("1. For direct tool usage: ./agent-calculator")
	fmt.Println("2. For LLM integration: ./agent-calculator llm")
	fmt.Println("3. For tool information: ./agent-calculator llm info")
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

// createMockProvider creates a mock provider for demonstration
func createMockProvider() ldomain.Provider {
	mockProvider := provider.NewMockProvider()
	// Track if we've seen a tool result
	hasToolResult := false
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Extract the last user message
		var lastUserMsg string
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				// Get text content from the message
				for _, part := range messages[i].Content {
					if part.Type == ldomain.ContentTypeText {
						lastUserMsg = part.Text
						break
					}
				}
				if lastUserMsg != "" {
					break
				}
			}
		}

		// Check if this is a tool result response
		if strings.Contains(lastUserMsg, "Tool results:") && strings.Contains(lastUserMsg, "Result:") {
			hasToolResult = true
			// Extract the calculation result from the tool result message
			var resultValue string
			lines := strings.Split(lastUserMsg, "\n")
			for _, line := range lines {
				if strings.Contains(line, "Result:") {
					parts := strings.Split(line, "Result:")
					if len(parts) > 1 {
						resultValue = strings.TrimSpace(parts[1])
					}
				}
			}

			// Generate a natural language response based on the tool result
			if resultValue != "" {
				if strings.Contains(messages[len(messages)-3].Content[0].Text, "multiply") ||
					strings.Contains(messages[len(messages)-3].Content[0].Text, "*") {
					return ldomain.Response{
						Content: fmt.Sprintf("The result of multiplying 25 by 17 is %s.", resultValue),
					}, nil
				} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "square root") {
					return ldomain.Response{
						Content: fmt.Sprintf("The square root of 144 is %s.", resultValue),
					}, nil
				} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "power") {
					return ldomain.Response{
						Content: fmt.Sprintf("2 raised to the power of 10 is %s.", resultValue),
					}, nil
				} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "factorial") {
					return ldomain.Response{
						Content: fmt.Sprintf("The factorial of 15 (15!) is %s.", resultValue),
					}, nil
				} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "GCD") {
					return ldomain.Response{
						Content: fmt.Sprintf("The greatest common divisor (GCD) of 48 and 18 is %s.", resultValue),
					}, nil
				} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "log base 2") {
					return ldomain.Response{
						Content: fmt.Sprintf("The logarithm base 2 of 64 is %s.", resultValue),
					}, nil
				} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "sine") ||
					strings.Contains(messages[len(messages)-3].Content[0].Text, "sin") {
					return ldomain.Response{
						Content: fmt.Sprintf("The sine of π/2 radians is %s.", resultValue),
					}, nil
				} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "pi times e") {
					return ldomain.Response{
						Content: fmt.Sprintf("The value of π × e is %s.", resultValue),
					}, nil
				}
			}

			return ldomain.Response{
				Content: fmt.Sprintf("The calculation result is %s.", resultValue),
			}, nil
		}

		// Initial tool call based on the prompt
		if lastUserMsg != "" && !hasToolResult {
			switch {
			case contains(lastUserMsg, "*") || contains(lastUserMsg, "multiply"):
				return ldomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "multiply", "operand1": 25, "operand2": 17}}`,
				}, nil
			case contains(lastUserMsg, "square root"):
				return ldomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "sqrt", "operand1": 144}}`,
				}, nil
			case contains(lastUserMsg, "power"):
				return ldomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "power", "operand1": 2, "operand2": 10}}`,
				}, nil
			case contains(lastUserMsg, "factorial"):
				return ldomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "factorial", "operand1": 15}}`,
				}, nil
			case contains(lastUserMsg, "GCD"):
				return ldomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "gcd", "operand1": 48, "operand2": 18}}`,
				}, nil
			case contains(lastUserMsg, "log base 2"):
				return ldomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "log", "operand1": 64, "operand2": 2}}`,
				}, nil
			case contains(lastUserMsg, "sine") || contains(lastUserMsg, "sin"):
				return ldomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "sin", "operand1": 1.5707963267948966}}`,
				}, nil
			case contains(lastUserMsg, "pi times e"):
				return ldomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "multiply", "operand1": 3.141592653589793, "operand2": 2.718281828459045}}`,
				}, nil
			default:
				return ldomain.Response{
					Content: "I'll help you with that calculation. Let me use the calculator tool.",
				}, nil
			}
		}

		return ldomain.Response{
			Content: "I can help you with mathematical calculations.",
		}, nil
	})

	return mockProvider
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 &&
		(s[0:len(substr)] == substr || contains(s[1:], substr)))
}
