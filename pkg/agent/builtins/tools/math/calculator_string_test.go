// ABOUTME: Unit tests for the calculator tool's string constant handling
// ABOUTME: Tests conversion of "pi", "e", and numeric strings to float64

package math

import (
	"math"
	"testing"
)

func TestCalculator_StringConstants(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name       string
		operation  string
		operand1   interface{}
		operand2   interface{}
		expected   float64
		shouldFail bool
	}{
		{"Pi times E with strings", "multiply", "pi", "e", math.Pi * math.E, false},
		{"Pi times E with π symbol", "multiply", "π", "e", math.Pi * math.E, false},
		{"Mixed string and number", "multiply", "pi", 2.0, math.Pi * 2, false},
		{"String number", "add", "3.14", "2.86", 6.0, false},
		{"Invalid string", "add", "hello", 5, 0, true}, // Should fail
		{"Case insensitive PI", "multiply", "PI", 2, math.Pi * 2, false},
		{"Case insensitive E", "multiply", "E", 3, math.E * 3, false},
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

			if tt.shouldFail {
				if calcResult.Success {
					t.Fatalf("expected failure for invalid string")
				}
				return
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
