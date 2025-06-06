// ABOUTME: Edge case tests for the calculator tool
// ABOUTME: Tests handling of null values, pi/2 calculations, and other edge cases

package math

import (
	"math"
	"testing"
)

func TestCalculator_NullHandling(t *testing.T) {
	tool := Calculator()
	ctx := createTestContext()

	tests := []struct {
		name      string
		params    map[string]interface{}
		expectErr bool
		expected  float64
	}{
		{
			name: "sin with null operand2",
			params: map[string]interface{}{
				"operation": "sin",
				"operand1":  math.Pi / 2,
				"operand2":  nil,
			},
			expectErr: false,
			expected:  1.0,
		},
		{
			name: "sin with missing operand2",
			params: map[string]interface{}{
				"operation": "sin",
				"operand1":  math.Pi / 2,
			},
			expectErr: false,
			expected:  1.0,
		},
		{
			name: "sin with pi/2 as string calculation",
			params: map[string]interface{}{
				"operation": "sin",
				"operand1":  1.5707963267948966,
			},
			expectErr: false,
			expected:  1.0,
		},
		{
			name: "divide pi by 2",
			params: map[string]interface{}{
				"operation": "divide",
				"operand1":  "pi",
				"operand2":  2,
			},
			expectErr: false,
			expected:  math.Pi / 2,
		},
		{
			name: "multiply with one null operand",
			params: map[string]interface{}{
				"operation": "multiply",
				"operand1":  5,
				"operand2":  nil,
			},
			expectErr: false,
			expected:  0, // 5 * 0 = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, tt.params)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calcResult := result.(*CalculatorResult)
			if !calcResult.Success {
				t.Fatalf("calculation failed: %s", calcResult.Error)
			}

			tolerance := 0.0000001
			if math.Abs(calcResult.Result-tt.expected) > tolerance {
				t.Errorf("expected %v, got %v", tt.expected, calcResult.Result)
			}
		})
	}
}
