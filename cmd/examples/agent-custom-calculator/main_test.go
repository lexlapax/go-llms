package main

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestCalculatorAgent_BasicOperations(t *testing.T) {
	calculator := NewCalculatorAgent("test-calculator")
	ctx := context.Background()

	tests := []struct {
		name      string
		operation string
		operand1  interface{}
		operand2  interface{}
		expected  float64
		wantErr   bool
	}{
		{
			name:      "Addition",
			operation: "add",
			operand1:  10.5,
			operand2:  5.2,
			expected:  15.7,
			wantErr:   false,
		},
		{
			name:      "Addition with alias",
			operation: "+",
			operand1:  3,
			operand2:  7,
			expected:  10,
			wantErr:   false,
		},
		{
			name:      "Subtraction",
			operation: "subtract",
			operand1:  20,
			operand2:  8,
			expected:  12,
			wantErr:   false,
		},
		{
			name:      "Subtraction with alias",
			operation: "-",
			operand1:  15.5,
			operand2:  3.2,
			expected:  12.3,
			wantErr:   false,
		},
		{
			name:      "Multiplication",
			operation: "multiply",
			operand1:  4,
			operand2:  6,
			expected:  24,
			wantErr:   false,
		},
		{
			name:      "Multiplication with alias",
			operation: "*",
			operand1:  2.5,
			operand2:  4,
			expected:  10,
			wantErr:   false,
		},
		{
			name:      "Division",
			operation: "divide",
			operand1:  20,
			operand2:  4,
			expected:  5,
			wantErr:   false,
		},
		{
			name:      "Division with alias",
			operation: "/",
			operand1:  15,
			operand2:  3,
			expected:  5,
			wantErr:   false,
		},
		{
			name:      "Power",
			operation: "power",
			operand1:  2,
			operand2:  8,
			expected:  256,
			wantErr:   false,
		},
		{
			name:      "Power with alias ^",
			operation: "^",
			operand1:  3,
			operand2:  3,
			expected:  27,
			wantErr:   false,
		},
		{
			name:      "Power with alias **",
			operation: "**",
			operand1:  5,
			operand2:  2,
			expected:  25,
			wantErr:   false,
		},
		{
			name:      "Square root",
			operation: "sqrt",
			operand1:  16,
			operand2:  0,
			expected:  4,
			wantErr:   false,
		},
		{
			name:      "Modulo",
			operation: "mod",
			operand1:  17,
			operand2:  5,
			expected:  2,
			wantErr:   false,
		},
		{
			name:      "Modulo with alias",
			operation: "%",
			operand1:  20,
			operand2:  6,
			expected:  2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := domain.NewState()
			state.Set("operation", tt.operation)
			state.Set("operand1", tt.operand1)
			state.Set("operand2", tt.operand2)

			result, err := calculator.Run(ctx, state)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			resultValue, exists := result.Get("result")
			if !exists {
				t.Errorf("Result value not found in state")
				return
			}

			resultFloat, ok := resultValue.(float64)
			if !ok {
				t.Errorf("Result is not float64, got %T", resultValue)
				return
			}

			// Use approximate comparison for floating point
			if abs(resultFloat-tt.expected) > 0.0001 {
				t.Errorf("Expected %.4f, got %.4f", tt.expected, resultFloat)
			}

			// Check that computed flag is set
			computed, exists := result.Get("computed")
			if !exists || computed != true {
				t.Errorf("Expected computed flag to be true")
			}
		})
	}
}

func TestCalculatorAgent_ErrorCases(t *testing.T) {
	calculator := NewCalculatorAgent("test-calculator")
	ctx := context.Background()

	tests := []struct {
		name      string
		setupFunc func() *domain.State
		wantErr   string
	}{
		{
			name: "Missing operation",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operand1", 5)
				state.Set("operand2", 3)
				return state
			},
			wantErr: "missing required field 'operation'",
		},
		{
			name: "Missing operand1",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operation", "add")
				state.Set("operand2", 3)
				return state
			},
			wantErr: "missing required field 'operand1'",
		},
		{
			name: "Missing operand2",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operation", "add")
				state.Set("operand1", 5)
				return state
			},
			wantErr: "missing required field 'operand2'",
		},
		{
			name: "Invalid operand1 type",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operation", "add")
				state.Set("operand1", "not_a_number")
				state.Set("operand2", 3)
				return state
			},
			wantErr: "operand1 must be a number, got string",
		},
		{
			name: "Invalid operand2 type",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operation", "add")
				state.Set("operand1", 5)
				state.Set("operand2", "not_a_number")
				return state
			},
			wantErr: "operand2 must be a number, got string",
		},
		{
			name: "Division by zero",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operation", "divide")
				state.Set("operand1", 10)
				state.Set("operand2", 0)
				return state
			},
			wantErr: "division by zero",
		},
		{
			name: "Modulo by zero",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operation", "mod")
				state.Set("operand1", 10)
				state.Set("operand2", 0)
				return state
			},
			wantErr: "modulo by zero",
		},
		{
			name: "Square root of negative number",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operation", "sqrt")
				state.Set("operand1", -4)
				state.Set("operand2", 0)
				return state
			},
			wantErr: "cannot take square root of negative number",
		},
		{
			name: "Unsupported operation",
			setupFunc: func() *domain.State {
				state := domain.NewState()
				state.Set("operation", "factorial")
				state.Set("operand1", 5)
				state.Set("operand2", 0)
				return state
			},
			wantErr: "unsupported operation: factorial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setupFunc()

			result, err := calculator.Run(ctx, state)

			if err == nil {
				t.Errorf("Expected error containing '%s' but got none. Result: %v", tt.wantErr, result)
				return
			}

			if err.Error() != tt.wantErr {
				t.Errorf("Expected error '%s', got '%s'", tt.wantErr, err.Error())
			}
		})
	}
}

func TestCalculatorAgent_TypeConversion(t *testing.T) {
	calculator := NewCalculatorAgent("test-calculator")
	ctx := context.Background()

	tests := []struct {
		name     string
		operand1 interface{}
		operand2 interface{}
		expected float64
	}{
		{
			name:     "int operands",
			operand1: 10,
			operand2: 5,
			expected: 15,
		},
		{
			name:     "int32 operands",
			operand1: int32(12),
			operand2: int32(3),
			expected: 15,
		},
		{
			name:     "int64 operands",
			operand1: int64(20),
			operand2: int64(8),
			expected: 28,
		},
		{
			name:     "float32 operands",
			operand1: float32(7.5),
			operand2: float32(2.5),
			expected: 10,
		},
		{
			name:     "uint operands",
			operand1: uint(15),
			operand2: uint(5),
			expected: 20,
		},
		{
			name:     "mixed types",
			operand1: int(10),
			operand2: float64(2.5),
			expected: 12.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := domain.NewState()
			state.Set("operation", "add")
			state.Set("operand1", tt.operand1)
			state.Set("operand2", tt.operand2)

			result, err := calculator.Run(ctx, state)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			resultValue, exists := result.Get("result")
			if !exists {
				t.Errorf("Result value not found in state")
				return
			}

			resultFloat, ok := resultValue.(float64)
			if !ok {
				t.Errorf("Result is not float64, got %T", resultValue)
				return
			}

			if abs(resultFloat-tt.expected) > 0.0001 {
				t.Errorf("Expected %.4f, got %.4f", tt.expected, resultFloat)
			}
		})
	}
}

func TestCalculatorAgent_Validate(t *testing.T) {
	calculator := NewCalculatorAgent("test-calculator")

	err := calculator.Validate()
	if err != nil {
		t.Errorf("Validation should pass for properly configured calculator agent, got error: %v", err)
	}
}

func TestCalculatorAgent_StatePreservation(t *testing.T) {
	calculator := NewCalculatorAgent("test-calculator")
	ctx := context.Background()

	state := domain.NewState()
	state.Set("operation", "add")
	state.Set("operand1", 10)
	state.Set("operand2", 5)
	state.Set("user_data", "preserve_me")
	state.Set("extra_info", "also_preserve")

	result, err := calculator.Run(ctx, state)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Check that original data is preserved
	userData, exists := result.Get("user_data")
	if !exists || userData != "preserve_me" {
		t.Errorf("Expected user_data to be preserved")
	}

	extraInfo, exists := result.Get("extra_info")
	if !exists || extraInfo != "also_preserve" {
		t.Errorf("Expected extra_info to be preserved")
	}

	// Check that new data is added
	resultValue, exists := result.Get("result")
	if !exists {
		t.Errorf("Expected result to be added")
	} else if resultValue != float64(15) {
		t.Errorf("Expected result to be 15, got %v", resultValue)
	}

	computed, exists := result.Get("computed")
	if !exists || computed != true {
		t.Errorf("Expected computed flag to be set")
	}
}

func TestConvertToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		wantOk   bool
	}{
		{"float64", float64(3.14), 3.14, true},
		{"float32", float32(2.5), 2.5, true},
		{"int", int(42), 42, true},
		{"int32", int32(100), 100, true},
		{"int64", int64(123), 123, true},
		{"uint", uint(50), 50, true},
		{"uint32", uint32(75), 75, true},
		{"uint64", uint64(200), 200, true},
		{"string", "not_a_number", 0, false},
		{"bool", true, 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := convertToFloat64(tt.input)

			if ok != tt.wantOk {
				t.Errorf("Expected ok=%v, got ok=%v", tt.wantOk, ok)
				return
			}

			if tt.wantOk && abs(result-tt.expected) > 0.0001 {
				t.Errorf("Expected %.4f, got %.4f", tt.expected, result)
			}
		})
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
