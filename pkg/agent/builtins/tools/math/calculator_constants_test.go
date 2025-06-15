// ABOUTME: Unit tests for mathematical constants in the calculator tool
// ABOUTME: Tests all supported constants both as operations and as operand strings

package math

import (
	"math"
	"testing"

	"github.com/lexlapax/go-llms/pkg/testutils/helpers"
)

func TestCalculator_AllConstants(t *testing.T) {
	tool := Calculator()
	ctx := helpers.CreateTestToolContext()

	// Test constants as operations
	constantTests := []struct {
		name      string
		operation string
		expected  float64
	}{
		{"Pi constant", "pi", math.Pi},
		{"E constant", "e", math.E},
		{"Phi constant", "phi", math.Phi},
		{"Tau constant", "tau", 2 * math.Pi},
		{"Sqrt2 constant", "sqrt2", math.Sqrt2},
		{"SqrtE constant", "sqrte", math.SqrtE},
		{"SqrtPi constant", "sqrtpi", math.SqrtPi},
		{"SqrtPhi constant", "sqrtphi", math.SqrtPhi},
		{"Ln2 constant", "ln2", math.Ln2},
		{"Ln10 constant", "ln10", math.Ln10},
		{"Log2E constant", "log2e", math.Log2E},
		{"Log10E constant", "log10e", math.Log10E},
	}

	for _, tt := range constantTests {
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

func TestCalculator_ConstantsAsOperands(t *testing.T) {
	tool := Calculator()
	ctx := helpers.CreateTestToolContext()

	// Test using constants as string operands
	tests := []struct {
		name      string
		operation string
		operand1  interface{}
		operand2  interface{}
		expected  float64
	}{
		{"Phi times 2", "multiply", "phi", 2, math.Phi * 2},
		{"Tau divided by 2", "divide", "tau", 2, math.Pi}, // tau/2 = 2π/2 = π
		{"Sqrt2 squared", "power", "sqrt2", 2, 2.0},
		{"E plus ln2", "add", "e", "ln2", math.E + math.Ln2},
		{"Log base 2 of 8", "log", 8, 2, 3}, // Verify regular log still works
		{"Pi plus phi", "add", "pi", "phi", math.Pi + math.Phi},
		{"SqrtPhi times SqrtPhi", "multiply", "sqrtphi", "sqrtphi", math.Phi},
		{"Case insensitive PHI", "multiply", "PHI", 3, math.Phi * 3},
		{"Unicode φ", "multiply", "φ", 2, math.Phi * 2},
		{"Unicode τ", "divide", "τ", 2, math.Pi},
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

			tolerance := 0.0000001
			if math.Abs(calcResult.Result-tt.expected) > tolerance {
				t.Errorf("expected %v, got %v (diff: %v)", tt.expected, calcResult.Result,
					math.Abs(calcResult.Result-tt.expected))
			}
		})
	}
}
