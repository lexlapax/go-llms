// ABOUTME: Unit tests for the calculator tool
// ABOUTME: Tests all mathematical operations and error handling

package math

import (
	"context"
	"math"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// Helper function to create test context
func createTestContext() *domain.ToolContext {
	ctx := context.Background()
	agentInfo := domain.AgentInfo{
		ID:          "test-agent",
		Name:        "Test Agent",
		Description: "A test agent",
		Type:        domain.AgentTypeLLM,
	}

	state := domain.NewState()
	stateReader := domain.NewStateReader(state)

	tc := &domain.ToolContext{
		Context: ctx,
		State:   stateReader,
		Agent:   agentInfo,
		RunID:   "test-run-123",
	}

	return tc
}

func TestCalculator_BasicArithmetic(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		operation string
		operand1  float64
		operand2  float64
		expected  float64
	}{
		{"Addition", "add", 10.5, 5.2, 15.7},
		{"Addition with +", "+", 10, 5, 15},
		{"Subtraction", "subtract", 20, 8, 12},
		{"Subtraction with -", "-", 10.5, 5.5, 5},
		{"Multiplication", "multiply", 4, 5, 20},
		{"Multiplication with *", "*", 3.5, 2, 7},
		{"Division", "divide", 20, 4, 5},
		{"Division with /", "/", 15, 3, 5},
		{"Power", "power", 2, 3, 8},
		{"Power with ^", "^", 3, 2, 9},
		{"Power with **", "**", 10, 2, 100},
		{"Modulo", "mod", 10, 3, 1},
		{"Modulo with %", "%", 17, 5, 2},
		{"Absolute value", "abs", -15, 0, 15},
		{"Absolute value positive", "abs", 15, 0, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": tt.operation,
				"operand1":  tt.operand1,
				"operand2":  tt.operand2,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult, ok := result.(*CalculatorResult)
			if !ok {
				t.Fatalf("expected *CalculatorResult, got %T", result)
			}

			if !calcResult.Success {
				t.Fatalf("calculation failed: %s", calcResult.Error)
			}

			if math.Abs(calcResult.Result-tt.expected) > 0.0001 {
				t.Errorf("expected %v, got %v", tt.expected, calcResult.Result)
			}
		})
	}
}

func TestCalculator_Roots(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		operation string
		operand1  float64
		expected  float64
	}{
		{"Square root of 16", "sqrt", 16, 4},
		{"Square root of 25", "sqrt", 25, 5},
		{"Square root of 2", "sqrt", 2, math.Sqrt(2)},
		{"Cube root of 8", "cbrt", 8, 2},
		{"Cube root of 27", "cbrt", 27, 3},
		{"Cube root of -8", "cbrt", -8, -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": tt.operation,
				"operand1":  tt.operand1,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult := result.(*CalculatorResult)
			if !calcResult.Success {
				t.Fatalf("calculation failed: %s", calcResult.Error)
			}

			if math.Abs(calcResult.Result-tt.expected) > 0.0001 {
				t.Errorf("expected %v, got %v", tt.expected, calcResult.Result)
			}
		})
	}
}

func TestCalculator_Logarithms(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		operation string
		operand1  float64
		operand2  float64
		expected  float64
	}{
		{"Natural log of e", "log", math.E, 0, 1},
		{"Log base 2 of 8", "log", 8, 2, 3},
		{"Log10 of 100", "log10", 100, 0, 2},
		{"Log2 of 16", "log2", 16, 0, 4},
		{"Exp of 0", "exp", 0, 0, 1},
		{"Exp of 1", "exp", 1, 0, math.E},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": tt.operation,
				"operand1":  tt.operand1,
				"operand2":  tt.operand2,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult := result.(*CalculatorResult)
			if !calcResult.Success {
				t.Fatalf("calculation failed: %s", calcResult.Error)
			}

			if math.Abs(calcResult.Result-tt.expected) > 0.0001 {
				t.Errorf("expected %v, got %v", tt.expected, calcResult.Result)
			}
		})
	}
}

func TestCalculator_Trigonometry(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		operation string
		operand1  float64
		expected  float64
	}{
		{"Sin of 0", "sin", 0, 0},
		{"Sin of π/2", "sin", math.Pi / 2, 1},
		{"Cos of 0", "cos", 0, 1},
		{"Cos of π", "cos", math.Pi, -1},
		{"Tan of 0", "tan", 0, 0},
		{"Asin of 0", "asin", 0, 0},
		{"Asin of 1", "asin", 1, math.Pi / 2},
		{"Acos of 1", "acos", 1, 0},
		{"Acos of 0", "acos", 0, math.Pi / 2},
		{"Atan of 0", "atan", 0, 0},
		{"Atan of 1", "atan", 1, math.Pi / 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": tt.operation,
				"operand1":  tt.operand1,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult := result.(*CalculatorResult)
			if !calcResult.Success {
				t.Fatalf("calculation failed: %s", calcResult.Error)
			}

			if math.Abs(calcResult.Result-tt.expected) > 0.0001 {
				t.Errorf("expected %v, got %v", tt.expected, calcResult.Result)
			}
		})
	}
}

func TestCalculator_Rounding(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		operation string
		operand1  float64
		expected  float64
	}{
		{"Floor of 3.7", "floor", 3.7, 3},
		{"Floor of -3.7", "floor", -3.7, -4},
		{"Ceil of 3.2", "ceil", 3.2, 4},
		{"Ceil of -3.2", "ceil", -3.2, -3},
		{"Round of 3.5", "round", 3.5, 4},
		{"Round of 3.4", "round", 3.4, 3},
		{"Round of -3.5", "round", -3.5, -4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": tt.operation,
				"operand1":  tt.operand1,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult := result.(*CalculatorResult)
			if !calcResult.Success {
				t.Fatalf("calculation failed: %s", calcResult.Error)
			}

			if calcResult.Result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, calcResult.Result)
			}
		})
	}
}

func TestCalculator_Constants(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		operation string
		expected  float64
	}{
		{"Pi constant", "pi", math.Pi},
		{"E constant", "e", math.E},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": tt.operation,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult := result.(*CalculatorResult)
			if !calcResult.Success {
				t.Fatalf("calculation failed: %s", calcResult.Error)
			}

			if calcResult.Result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, calcResult.Result)
			}
		})
	}
}

func TestCalculator_Advanced(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		operation string
		operand1  float64
		operand2  float64
		expected  float64
	}{
		{"Factorial of 0", "factorial", 0, 0, 1},
		{"Factorial of 5", "factorial", 5, 0, 120},
		{"Factorial of 10", "factorial", 10, 0, 3628800},
		{"GCD of 12 and 8", "gcd", 12, 8, 4},
		{"GCD of 15 and 25", "gcd", 15, 25, 5},
		{"LCM of 4 and 6", "lcm", 4, 6, 12},
		{"LCM of 12 and 18", "lcm", 12, 18, 36},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": tt.operation,
				"operand1":  tt.operand1,
				"operand2":  tt.operand2,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult := result.(*CalculatorResult)
			if !calcResult.Success {
				t.Fatalf("calculation failed: %s", calcResult.Error)
			}

			if calcResult.Result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, calcResult.Result)
			}
		})
	}
}

func TestCalculator_ErrorCases(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		operation string
		operand1  float64
		operand2  float64
		expectErr string
	}{
		{"Division by zero", "divide", 10, 0, "division by zero"},
		{"Modulo by zero", "mod", 10, 0, "modulo by zero"},
		{"Square root of negative", "sqrt", -4, 0, "cannot take square root of negative number"},
		{"Log of negative", "log", -1, 0, "logarithm of non-positive number"},
		{"Log of zero", "log10", 0, 0, "logarithm of non-positive number"},
		{"Invalid log base", "log", 10, 1, "invalid logarithm base"},
		{"Asin out of range", "asin", 2, 0, "asin domain error: value must be between -1 and 1"},
		{"Acos out of range", "acos", -2, 0, "acos domain error: value must be between -1 and 1"},
		{"Factorial of negative", "factorial", -5, 0, "factorial requires non-negative integer"},
		{"Factorial of float", "factorial", 5.5, 0, "factorial requires non-negative integer"},
		{"GCD with negative", "gcd", -12, 8, "gcd requires positive integers"},
		{"LCM with zero", "lcm", 0, 5, "lcm requires positive integers"},
		{"Unknown operation", "unknown", 1, 2, "unsupported operation: unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": tt.operation,
				"operand1":  tt.operand1,
				"operand2":  tt.operand2,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult := result.(*CalculatorResult)
			if calcResult.Success {
				t.Fatalf("expected error but calculation succeeded")
			}

			if calcResult.Error != tt.expectErr {
				t.Errorf("expected error '%s', got '%s'", tt.expectErr, calcResult.Error)
			}
		})
	}
}

func TestCalculator_EdgeCases(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	// Test very large factorial
	params := map[string]interface{}{
		"operation": "factorial",
		"operand1":  171, // 171! overflows float64
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calcResult := result.(*CalculatorResult)
	if calcResult.Success {
		t.Error("expected NaN for 171!")
	}

	// Test tan of π/2 (should be very large but not infinite)
	params = map[string]interface{}{
		"operation": "tan",
		"operand1":  math.Pi / 2,
	}

	result, err = tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calcResult = result.(*CalculatorResult)
	if !calcResult.Success {
		t.Errorf("tan(π/2) should succeed, got error: %s", calcResult.Error)
	}
}
