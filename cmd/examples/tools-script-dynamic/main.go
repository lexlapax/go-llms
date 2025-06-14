// ABOUTME: Example of dynamic script-based tool registration
// ABOUTME: Shows how go-llmspell and other scripting bridges can register tools at runtime

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// SimpleScriptHandler is a basic script handler for demonstration
type SimpleScriptHandler struct{}

func (h *SimpleScriptHandler) Execute(ctx context.Context, script string, toolCtx *domain.ToolContext, params interface{}) (interface{}, error) {
	// For demonstration, just evaluate simple JavaScript-like expressions
	switch script {
	case "add":
		if paramMap, ok := params.(map[string]interface{}); ok {
			var a, b float64
			var aOk, bOk bool

			// Handle both int and float64 types
			if aVal, exists := paramMap["a"]; exists {
				switch v := aVal.(type) {
				case float64:
					a, aOk = v, true
				case int:
					a, aOk = float64(v), true
				}
			}

			if bVal, exists := paramMap["b"]; exists {
				switch v := bVal.(type) {
				case float64:
					b, bOk = v, true
				case int:
					b, bOk = float64(v), true
				}
			}

			if aOk && bOk {
				return a + b, nil
			}
		}
		return nil, fmt.Errorf("invalid parameters for add operation")

	case "greet":
		if paramMap, ok := params.(map[string]interface{}); ok {
			if name, ok := paramMap["name"].(string); ok {
				return fmt.Sprintf("Hello, %s!", name), nil
			}
		}
		return "Hello, World!", nil

	case "factorial":
		if paramMap, ok := params.(map[string]interface{}); ok {
			var n float64
			var nOk bool

			// Handle both int and float64 types
			if nVal, exists := paramMap["n"]; exists {
				switch v := nVal.(type) {
				case float64:
					n, nOk = v, true
				case int:
					n, nOk = float64(v), true
				}
			}

			if nOk && n >= 0 {
				result := 1.0
				for i := 2.0; i <= n; i++ {
					result *= i
				}
				return result, nil
			}
		}
		return nil, fmt.Errorf("invalid parameter for factorial operation")

	default:
		return nil, fmt.Errorf("unknown script: %s", script)
	}
}

func (h *SimpleScriptHandler) Validate(script string) error {
	validScripts := []string{"add", "greet", "factorial"}
	for _, valid := range validScripts {
		if script == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid script: %s", script)
}

func (h *SimpleScriptHandler) Engine() tools.ScriptEngine {
	return tools.ScriptEngineJavaScript
}

func (h *SimpleScriptHandler) SupportsFeature(feature string) bool {
	return feature == "basic_math" || feature == "greetings"
}

func main() {
	fmt.Println("=== Dynamic Script-Based Tool Registration Example ===")

	// Register the script handler
	handler := &SimpleScriptHandler{}
	err := tools.RegisterScriptHandler(handler)
	if err != nil {
		log.Fatalf("Failed to register script handler: %v", err)
	}
	fmt.Println("✓ Registered JavaScript script handler")

	// Create tool discovery instance
	discovery := tools.NewDiscovery()

	// Create a dedicated namespace for this example to avoid conflicts
	err = discovery.CreateNamespace("script-example")
	if err != nil && err.Error() != "namespace script-example already exists" {
		log.Fatalf("Failed to create namespace: %v", err)
	}

	err = discovery.SwitchNamespace("script-example")
	if err != nil {
		log.Fatalf("Failed to switch to script-example namespace: %v", err)
	}
	fmt.Println("✓ Created and switched to 'script-example' namespace")

	// Example 1: Calculator Tool
	calculatorDef := tools.ScriptToolDefinition{
		Name:        "calculator",
		Description: "A simple calculator that can add two numbers",
		Category:    "math",
		Tags:        []string{"calculator", "math", "addition"},
		Version:     "1.0.0",
		Engine:      tools.ScriptEngineJavaScript,
		Script:      "add",
		ParameterSchema: &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"a": {
					Type:        "number",
					Description: "First number",
				},
				"b": {
					Type:        "number",
					Description: "Second number",
				},
			},
			Required: []string{"a", "b"},
		},
		OutputSchema: &schemaDomain.Schema{
			Type:        "number",
			Description: "Sum of the two numbers",
		},
		Examples: []domain.ToolExample{
			{
				Name:        "add_positive_numbers",
				Description: "Add two positive numbers",
				Input:       map[string]interface{}{"a": 5, "b": 3},
				Output:      8,
			},
			{
				Name:        "add_negative_numbers",
				Description: "Add numbers including negatives",
				Input:       map[string]interface{}{"a": -2, "b": 7},
				Output:      5,
			},
		},
		Constraints: []string{
			"Only supports addition of two numbers",
			"Input must be valid numbers",
		},
		ErrorGuidance: map[string]string{
			"invalid_params": "Ensure both 'a' and 'b' parameters are provided as numbers",
			"script_error":   "Check that the calculation parameters are valid numbers",
		},
	}

	// Register calculator tool
	err = tools.RegisterScriptToolWithDiscovery(calculatorDef)
	if err != nil {
		log.Fatalf("Failed to register calculator tool: %v", err)
	}
	fmt.Println("✓ Registered calculator tool")

	// Example 2: Greeting Tool
	greetingDef := tools.ScriptToolDefinition{
		Name:        "greeter",
		Description: "A friendly greeting tool",
		Category:    "utility",
		Tags:        []string{"greeting", "social"},
		Version:     "1.0.0",
		Engine:      tools.ScriptEngineJavaScript,
		Script:      "greet",
		ParameterSchema: &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {
					Type:        "string",
					Description: "Name of the person to greet",
				},
			},
		},
		OutputSchema: &schemaDomain.Schema{
			Type:        "string",
			Description: "Personalized greeting message",
		},
		Examples: []domain.ToolExample{
			{
				Name:        "greet_person",
				Description: "Greet a specific person",
				Input:       map[string]interface{}{"name": "Alice"},
				Output:      "Hello, Alice!",
			},
			{
				Name:        "default_greeting",
				Description: "Default greeting when no name provided",
				Input:       map[string]interface{}{},
				Output:      "Hello, World!",
			},
		},
	}

	// Register greeting tool
	err = tools.RegisterScriptToolWithDiscovery(greetingDef)
	if err != nil {
		log.Fatalf("Failed to register greeting tool: %v", err)
	}
	fmt.Println("✓ Registered greeting tool")

	// Example 3: Factorial Tool
	factorialDef := tools.ScriptToolDefinition{
		Name:        "factorial",
		Description: "Calculate factorial of a number",
		Category:    "math",
		Tags:        []string{"factorial", "math", "recursive"},
		Version:     "1.0.0",
		Engine:      tools.ScriptEngineJavaScript,
		Script:      "factorial",
		ParameterSchema: &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"n": {
					Type:        "number",
					Description: "Number to calculate factorial for (must be non-negative)",
					Minimum:     &[]float64{0}[0],
				},
			},
			Required: []string{"n"},
		},
		OutputSchema: &schemaDomain.Schema{
			Type:        "number",
			Description: "Factorial of the input number",
		},
		Examples: []domain.ToolExample{
			{
				Name:        "small_factorial",
				Description: "Factorial of a small number",
				Input:       map[string]interface{}{"n": 5},
				Output:      120,
			},
			{
				Name:        "zero_factorial",
				Description: "Factorial of zero",
				Input:       map[string]interface{}{"n": 0},
				Output:      1,
			},
		},
		Constraints: []string{
			"Input must be a non-negative integer",
			"Large numbers may cause overflow",
		},
	}

	// Register factorial tool
	err = tools.RegisterScriptToolWithDiscovery(factorialDef)
	if err != nil {
		log.Fatalf("Failed to register factorial tool: %v", err)
	}
	fmt.Println("✓ Registered factorial tool")

	// Demonstrate tool discovery
	fmt.Println("\n=== Tool Discovery ===")
	registeredTools := discovery.GetRegisteredTools()
	fmt.Printf("Found %d registered tools:\n", len(registeredTools))

	for _, tool := range registeredTools {
		if len(tool.Tags) > 0 {
			hasScriptTag := false
			for _, tag := range tool.Tags {
				if tag == "javascript" {
					hasScriptTag = true
					break
				}
			}
			if hasScriptTag {
				fmt.Printf("- %s: %s (Category: %s, Version: %s)\n",
					tool.Name, tool.Description, tool.Category, tool.Version)
			}
		}
	}

	// Demonstrate tool execution
	fmt.Println("\n=== Tool Execution ===")

	// Test calculator
	calcTool, err := discovery.CreateTool("calculator")
	if err != nil {
		log.Fatalf("Failed to create calculator tool: %v", err)
	}

	ctx := &domain.ToolContext{
		Context: context.Background(),
	}

	result, err := calcTool.Execute(ctx, map[string]interface{}{"a": 15.0, "b": 27.0})
	if err != nil {
		log.Fatalf("Failed to execute calculator: %v", err)
	}
	fmt.Printf("Calculator: 15 + 27 = %v\n", result)

	// Test greeter
	greeterTool, err := discovery.CreateTool("greeter")
	if err != nil {
		log.Fatalf("Failed to create greeter tool: %v", err)
	}

	result, err = greeterTool.Execute(ctx, map[string]interface{}{"name": "Go Developer"})
	if err != nil {
		log.Fatalf("Failed to execute greeter: %v", err)
	}
	fmt.Printf("Greeter: %v\n", result)

	// Test factorial
	factorialTool, err := discovery.CreateTool("factorial")
	if err != nil {
		log.Fatalf("Failed to create factorial tool: %v", err)
	}

	result, err = factorialTool.Execute(ctx, map[string]interface{}{"n": 6.0})
	if err != nil {
		log.Fatalf("Failed to execute factorial: %v", err)
	}
	fmt.Printf("Factorial: 6! = %v\n", result)

	// Demonstrate tool metadata
	fmt.Println("\n=== Tool Metadata ===")
	schema, err := discovery.GetToolSchema("calculator")
	if err != nil {
		log.Fatalf("Failed to get calculator schema: %v", err)
	}

	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Printf("Calculator Schema:\n%s\n", schemaJSON)

	// Demonstrate namespace isolation
	fmt.Println("\n=== Namespace Isolation ===")

	// Create a separate namespace for testing
	err = discovery.CreateNamespace("experimental")
	if err != nil {
		log.Fatalf("Failed to create experimental namespace: %v", err)
	}

	// Switch to experimental namespace
	err = discovery.SwitchNamespace("experimental")
	if err != nil {
		log.Fatalf("Failed to switch to experimental namespace: %v", err)
	}

	fmt.Printf("Current namespace: %s\n", discovery.GetCurrentNamespace())

	experimentalTools := discovery.GetRegisteredTools()
	fmt.Printf("Tools in experimental namespace: %d\n", len(experimentalTools))

	// Switch back to default
	err = discovery.SwitchNamespace("default")
	if err != nil {
		log.Fatalf("Failed to switch back to default namespace: %v", err)
	}

	defaultTools := discovery.GetRegisteredTools()
	fmt.Printf("Tools in default namespace: %d\n", len(defaultTools))

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("1. Registering custom script handlers")
	fmt.Println("2. Creating script-based tool definitions")
	fmt.Println("3. Dynamic tool registration with discovery")
	fmt.Println("4. Tool execution with parameters")
	fmt.Println("5. Tool metadata and schema access")
	fmt.Println("6. Namespace isolation for multi-tenant scenarios")
	fmt.Println("\nThis pattern enables go-llmspell and other scripting bridges")
	fmt.Println("to register tools written in various scripting languages!")
}
