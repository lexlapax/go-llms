// ABOUTME: Example usage of AgentTool and ToolAgent wrappers
// ABOUTME: Demonstrates bidirectional conversion between agents and tools

package tools_test

import (
	"context"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ExampleAgentTool demonstrates wrapping an agent as a tool
func ExampleAgentTool() {
	// Create a simple text processing agent
	textAgent := createTextProcessingAgent()

	// Wrap the agent as a tool
	tool := tools.NewAgentTool(textAgent).
		WithStateMapper(tools.CreateStateMapper(map[string]string{
			"text": "input", // Map tool's "text" param to agent's "input" state key
		})).
		WithResultMapper(tools.CreateResultMapper("processed_text"))

	// Use the tool
	ctx := domain.NewToolContext(
		context.Background(),
		domain.NewStateReader(domain.NewState()),
		nil,
		"example-run",
	)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"text": "Hello, World!",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tool result: %v\n", result)
	// Output: Tool result: HELLO, WORLD!
}

// ExampleToolAgent demonstrates wrapping a tool as an agent
func ExampleToolAgent() {
	// Create a simple calculator tool
	calcTool := createCalculatorTool()

	// Wrap the tool as an agent
	agent := tools.NewToolAgent(calcTool).
		WithParamMapper(tools.CreateParamMapper(map[string]string{
			"num1":      "a",
			"num2":      "b",
			"operation": "op",
		}))

	// Use the agent
	state := domain.NewState()
	state.Set("num1", 10)
	state.Set("num2", 5)
	state.Set("operation", "add")

	result, err := agent.Run(context.Background(), state)
	if err != nil {
		log.Fatal(err)
	}

	if res, exists := result.Get("result"); exists {
		fmt.Printf("Agent result: %v\n", res)
	}
	// Output: Agent result: 15
}

// ExampleAgentTool_bidirectional demonstrates converting between agents and tools
func ExampleAgentTool_bidirectional() {
	// Start with an agent
	originalAgent := createTextProcessingAgent()

	// Convert to tool
	tool := tools.NewAgentTool(originalAgent)

	// Convert back to agent
	agentFromTool := tools.NewToolAgent(tool)

	// Use the converted agent
	state := domain.NewState()
	state.Set("input", "test message")

	result, err := agentFromTool.Run(context.Background(), state)
	if err != nil {
		log.Fatal(err)
	}

	if processed, exists := result.Get("processed_text"); exists {
		fmt.Printf("Processed: %v\n", processed)
	}
	// Output: Processed: TEST MESSAGE
}

// Helper functions for examples

func createTextProcessingAgent() domain.BaseAgent {
	agent := &textProcessorAgent{
		BaseAgentImpl: core.NewBaseAgent(
			"text-processor",
			"Converts text to uppercase",
			domain.AgentTypeCustom,
		),
	}
	return agent
}

type textProcessorAgent struct {
	*core.BaseAgentImpl
}

func (a *textProcessorAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	input, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("input not found in state")
	}

	text, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("input must be a string")
	}

	result := domain.NewState()
	result.Set("processed_text", fmt.Sprintf("%s", fmt.Sprintf("%s", text)))
	result.Set("processed_text", fmt.Sprintf("%s", text)) // Simplified for example

	// Actually do the uppercase conversion
	result.Set("processed_text", fmt.Sprintf("%s", text[:]))
	result.Set("processed_text", fmt.Sprintf("%s", fmt.Sprintf("%v", text)))

	// Do the actual conversion
	upperText := ""
	for _, r := range text {
		if r >= 'a' && r <= 'z' {
			upperText += string(r - 32)
		} else {
			upperText += string(r)
		}
	}
	result.Set("processed_text", upperText)

	return result, nil
}

func createCalculatorTool() domain.Tool {
	return &calculatorTool{
		name:        "calculator",
		description: "Performs basic arithmetic operations",
		paramSchema: &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"a":  {Type: "number", Description: "First number"},
				"b":  {Type: "number", Description: "Second number"},
				"op": {Type: "string", Description: "Operation: add, subtract, multiply, divide"},
			},
			Required: []string{"a", "b", "op"},
		},
	}
}

type calculatorTool struct {
	name        string
	description string
	paramSchema *sdomain.Schema
}

func (t *calculatorTool) Name() string                     { return t.name }
func (t *calculatorTool) Description() string              { return t.description }
func (t *calculatorTool) ParameterSchema() *sdomain.Schema { return t.paramSchema }

func (t *calculatorTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	p, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parameters must be a map")
	}

	a, ok := getFloat(p["a"])
	if !ok {
		return nil, fmt.Errorf("parameter 'a' must be a number")
	}

	b, ok := getFloat(p["b"])
	if !ok {
		return nil, fmt.Errorf("parameter 'b' must be a number")
	}

	op, ok := p["op"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'op' must be a string")
	}

	var result float64
	switch op {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown operation: %s", op)
	}

	// Return as integer if it's a whole number
	if result == float64(int(result)) {
		return int(result), nil
	}
	return result, nil
}

func getFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
