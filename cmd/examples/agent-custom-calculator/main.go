// ABOUTME: Example demonstrating a custom agent that performs pure computational logic without LLM dependencies
// ABOUTME: Shows how to create stateless computation patterns with input validation and error handling

package main

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// CalculatorAgent implements pure computational logic without LLM dependencies
type CalculatorAgent struct {
	*core.BaseAgentImpl
}

// NewCalculatorAgent creates a new calculator agent
func NewCalculatorAgent(name string) *CalculatorAgent {
	return &CalculatorAgent{
		BaseAgentImpl: core.NewBaseAgent(name, "Calculator agent for mathematical operations", domain.AgentTypeCustom),
	}
}

// Run performs mathematical calculations based on the input state
func (c *CalculatorAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Emit start event
	c.EmitEvent(domain.EventAgentStart, map[string]interface{}{
		"agent": c.Name(),
		"type":  "calculator",
	})

	// Extract operation and operands from state
	operation, exists := input.Get("operation")
	if !exists {
		return nil, fmt.Errorf("missing required field 'operation'")
	}

	operand1, exists := input.Get("operand1")
	if !exists {
		return nil, fmt.Errorf("missing required field 'operand1'")
	}

	operand2, exists := input.Get("operand2")
	if !exists {
		return nil, fmt.Errorf("missing required field 'operand2'")
	}

	// Convert operands to float64
	num1, ok := convertToFloat64(operand1)
	if !ok {
		return nil, fmt.Errorf("operand1 must be a number, got %T", operand1)
	}

	num2, ok := convertToFloat64(operand2)
	if !ok {
		return nil, fmt.Errorf("operand2 must be a number, got %T", operand2)
	}

	// Perform calculation
	var result float64
	var err error

	switch operation.(string) {
	case "add", "+":
		result = num1 + num2
	case "subtract", "-":
		result = num1 - num2
	case "multiply", "*":
		result = num1 * num2
	case "divide", "/":
		if num2 == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = num1 / num2
	case "power", "^", "**":
		result = math.Pow(num1, num2)
	case "sqrt":
		if num1 < 0 {
			return nil, fmt.Errorf("cannot take square root of negative number")
		}
		result = math.Sqrt(num1)
		// For sqrt, we only use operand1
	case "mod", "%":
		if num2 == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		result = math.Mod(num1, num2)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	// Create result state (clone input to preserve original data)
	resultState := input.Clone()
	resultState.Set("result", result)
	resultState.Set("computed", true)

	// Emit completion event
	c.EmitEvent(domain.EventAgentComplete, map[string]interface{}{
		"agent":     c.Name(),
		"operation": operation,
		"result":    result,
	})

	return resultState, err
}

// Validate ensures the agent is properly configured
func (c *CalculatorAgent) Validate() error {
	// Call base validation
	if err := c.BaseAgentImpl.Validate(); err != nil {
		return err
	}

	// Calculator agent doesn't need additional validation
	// as it's stateless and has no dependencies
	return nil
}

// convertToFloat64 safely converts various numeric types to float64
func convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func main() {
	// Create calculator agent
	calculator := NewCalculatorAgent("calculator")

	// Validate the agent
	if err := calculator.Validate(); err != nil {
		log.Fatalf("Agent validation failed: %v", err)
	}

	ctx := context.Background()

	// Example 1: Addition
	fmt.Println("=== Calculator Agent Example ===")

	state1 := domain.NewState()
	state1.Set("operation", "add")
	state1.Set("operand1", 10.5)
	state1.Set("operand2", 5.2)

	result1, err := calculator.Run(ctx, state1)
	if err != nil {
		log.Printf("Addition failed: %v", err)
	} else {
		if result, exists := result1.Get("result"); exists {
			fmt.Printf("10.5 + 5.2 = %.2f\n", result)
		}
	}

	// Example 2: Division
	state2 := domain.NewState()
	state2.Set("operation", "divide")
	state2.Set("operand1", 20)
	state2.Set("operand2", 4)

	result2, err := calculator.Run(ctx, state2)
	if err != nil {
		log.Printf("Division failed: %v", err)
	} else {
		if result, exists := result2.Get("result"); exists {
			fmt.Printf("20 / 4 = %.2f\n", result)
		}
	}

	// Example 3: Power operation
	state3 := domain.NewState()
	state3.Set("operation", "power")
	state3.Set("operand1", 2)
	state3.Set("operand2", 8)

	result3, err := calculator.Run(ctx, state3)
	if err != nil {
		log.Printf("Power operation failed: %v", err)
	} else {
		if result, exists := result3.Get("result"); exists {
			fmt.Printf("2^8 = %.0f\n", result)
		}
	}

	// Example 4: Square root
	state4 := domain.NewState()
	state4.Set("operation", "sqrt")
	state4.Set("operand1", 16)
	state4.Set("operand2", 0) // Not used for sqrt but required

	result4, err := calculator.Run(ctx, state4)
	if err != nil {
		log.Printf("Square root failed: %v", err)
	} else {
		if result, exists := result4.Get("result"); exists {
			fmt.Printf("√16 = %.2f\n", result)
		}
	}

	// Example 5: Error handling - division by zero
	state5 := domain.NewState()
	state5.Set("operation", "divide")
	state5.Set("operand1", 10)
	state5.Set("operand2", 0)

	result5, err := calculator.Run(ctx, state5)
	if err != nil {
		fmt.Printf("Expected error for division by zero: %v\n", err)
	} else {
		if result, exists := result5.Get("result"); exists {
			fmt.Printf("Unexpected success: %v\n", result)
		}
	}

	// Example 6: Error handling - unsupported operation
	state6 := domain.NewState()
	state6.Set("operation", "factorial")
	state6.Set("operand1", 5)
	state6.Set("operand2", 0)

	result6, err := calculator.Run(ctx, state6)
	if err != nil {
		fmt.Printf("Expected error for unsupported operation: %v\n", err)
	} else {
		if result, exists := result6.Get("result"); exists {
			fmt.Printf("Unexpected success: %v\n", result)
		}
	}

	fmt.Println("\n=== Supported Operations ===")
	fmt.Println("add, +        - Addition")
	fmt.Println("subtract, -   - Subtraction")
	fmt.Println("multiply, *   - Multiplication")
	fmt.Println("divide, /     - Division")
	fmt.Println("power, ^, **  - Exponentiation")
	fmt.Println("sqrt          - Square root (uses operand1 only)")
	fmt.Println("mod, %        - Modulo")
}
