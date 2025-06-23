// ABOUTME: Calculator tool for performing mathematical operations
// ABOUTME: Supports basic arithmetic, trigonometry, logarithms, and other mathematical functions

package math

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// CalculatorParams defines parameters for the Calculator tool
type CalculatorParams struct {
	Operation string  `json:"operation" description:"The mathematical operation to perform"`
	Operand1  float64 `json:"operand1" description:"First operand (or single operand for unary operations)"`
	Operand2  float64 `json:"operand2,omitempty" description:"Second operand (optional for unary operations)"`
}

// CalculatorResult defines the result of the Calculator tool
type CalculatorResult struct {
	Result    float64 `json:"result"`
	Operation string  `json:"operation"`
	Operand1  float64 `json:"operand1"`
	Operand2  float64 `json:"operand2,omitempty"`
	Success   bool    `json:"success"`
	Error     string  `json:"error,omitempty"`
}

// calculatorParamSchema defines parameters for the Calculator tool
var calculatorParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"operation": {
			Type:        "string",
			Description: "The mathematical operation to perform",
		},
		"operand1": {
			Type:        "number",
			Description: "First operand (or single operand for unary operations)",
		},
		"operand2": {
			Type:        "number",
			Description: "Second operand (optional for unary operations)",
		},
	},
	Required: []string{"operation"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("calculator", Calculator(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "calculator",
			Category:    "math",
			Tags:        []string{"math", "calculation", "arithmetic", "trigonometry"},
			Description: "Performs mathematical calculations including arithmetic, trigonometry, and logarithms",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic addition",
					Description: "Add two numbers",
					Code:        `Calculator().Execute(ctx, CalculatorParams{Operation: "add", Operand1: 10.5, Operand2: 5.2})`,
				},
				{
					Name:        "Square root",
					Description: "Calculate square root",
					Code:        `Calculator().Execute(ctx, CalculatorParams{Operation: "sqrt", Operand1: 16})`,
				},
				{
					Name:        "Trigonometry",
					Description: "Calculate sine of an angle (in radians)",
					Code:        `Calculator().Execute(ctx, CalculatorParams{Operation: "sin", Operand1: math.Pi/2})`,
				},
				{
					Name:        "Get mathematical constant",
					Description: "Get value of pi",
					Code:        `Calculator().Execute(ctx, CalculatorParams{Operation: "pi"})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: false,
		},
	})
}

// parseOperand converts various input formats to float64
// It handles numeric values, strings that can be parsed as numbers,
// and special constant names like "pi" and "e"
func parseOperand(value interface{}) (float64, error) {
	// Handle nil/null values
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		// Check for special constants
		switch strings.ToLower(v) {
		case "pi", "π":
			return math.Pi, nil
		case "e":
			return math.E, nil
		case "phi", "φ": // Golden ratio
			return math.Phi, nil
		case "tau", "τ": // 2π
			return 2 * math.Pi, nil
		case "sqrt2", "√2": // Square root of 2
			return math.Sqrt2, nil
		case "sqrte", "√e": // Square root of e
			return math.SqrtE, nil
		case "sqrtpi", "√π": // Square root of pi
			return math.SqrtPi, nil
		case "sqrtphi", "√φ": // Square root of phi
			return math.SqrtPhi, nil
		case "ln2": // Natural log of 2
			return math.Ln2, nil
		case "ln10": // Natural log of 10
			return math.Ln10, nil
		case "log2e": // Log base 2 of e
			return math.Log2E, nil
		case "log10e": // Log base 10 of e
			return math.Log10E, nil
		default:
			// Try to parse as number
			return strconv.ParseFloat(v, 64)
		}
	default:
		// Try to convert to string and parse
		str := fmt.Sprintf("%v", value)
		return strconv.ParseFloat(str, 64)
	}
}

// Calculator creates a tool for performing mathematical calculations including basic arithmetic,
// trigonometry, logarithms, and advanced operations like factorial and GCD/LCM.
// The tool supports mathematical constants (pi, e, phi) and handles edge cases like division by zero
// and domain errors gracefully, returning structured error messages in the result.
func Calculator() domain.Tool {
	// Create a wrapper function that handles parameter preprocessing
	wrappedFn := func(ctx *domain.ToolContext, params map[string]interface{}) (*CalculatorResult, error) {
		// Extract and process parameters
		operation, _ := params["operation"].(string)

		// Parse operands with special handling for constants
		var operand1, operand2 float64
		var err error

		if op1, exists := params["operand1"]; exists {
			operand1, err = parseOperand(op1)
			if err != nil {
				return &CalculatorResult{
					Operation: operation,
					Success:   false,
					Error:     fmt.Sprintf("invalid operand1: %v", err),
				}, nil
			}
		}

		if op2, exists := params["operand2"]; exists {
			operand2, err = parseOperand(op2)
			if err != nil {
				return &CalculatorResult{
					Operation: operation,
					Operand1:  operand1,
					Success:   false,
					Error:     fmt.Sprintf("invalid operand2: %v", err),
				}, nil
			}
		}

		// Create typed params and call the main function
		typedParams := CalculatorParams{
			Operation: operation,
			Operand1:  operand1,
			Operand2:  operand2,
		}

		return calculatorExecute(ctx, typedParams)
	}

	// Create output schema for CalculatorResult
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"result": {
				Type:        "number",
				Description: "The calculation result",
			},
			"operation": {
				Type:        "string",
				Description: "The operation performed",
			},
			"operand1": {
				Type:        "number",
				Description: "First operand used",
			},
			"operand2": {
				Type:        "number",
				Description: "Second operand (if applicable)",
			},
			"success": {
				Type:        "boolean",
				Description: "Whether the calculation succeeded",
			},
			"error": {
				Type:        "string",
				Description: "Error message if calculation failed",
			},
		},
		Required: []string{"operation", "success"},
	}

	builder := atools.NewToolBuilder("calculator", "Performs mathematical calculations including arithmetic, trigonometry, and logarithms").
		WithFunction(wrappedFn).
		WithParameterSchema(calculatorParamSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`Use this tool to perform mathematical calculations. It supports:

Basic Arithmetic:
- add (+): Addition of two numbers
- subtract (-): Subtraction (operand1 - operand2)
- multiply (*): Multiplication
- divide (/): Division (checks for division by zero)
- mod (%): Modulo operation
- power (^, **): Exponentiation
- abs: Absolute value

Roots and Logarithms:
- sqrt: Square root (requires non-negative operand)
- cbrt: Cube root
- log: Natural logarithm (base e) or logarithm with custom base
- log10: Base-10 logarithm
- log2: Base-2 logarithm
- exp: e raised to the power of operand1

Trigonometry (angles in radians):
- sin, cos, tan: Standard trigonometric functions
- asin, acos, atan: Inverse trigonometric functions
- sinh, cosh, tanh: Hyperbolic functions

Rounding:
- floor: Round down to nearest integer
- ceil: Round up to nearest integer
- round: Round to nearest integer

Advanced:
- factorial: Calculate n! (requires non-negative integer ≤ 170)
- gcd: Greatest common divisor (requires positive integers)
- lcm: Least common multiple (requires positive integers)

Mathematical Constants (no operands needed):
- pi (π): 3.14159...
- e: Euler's number (2.71828...)
- phi (φ): Golden ratio
- tau (τ): 2π
- sqrt2, sqrte, sqrtpi, sqrtphi: Square roots of constants
- ln2, ln10, log2e, log10e: Logarithmic constants

Special operand values:
- You can use constant names as operands, e.g., operand1: "pi"
- Numbers can be provided as strings and will be parsed`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic addition",
				Description: "Add two decimal numbers",
				Scenario:    "When you need to sum two values",
				Input: map[string]interface{}{
					"operation": "add",
					"operand1":  10.5,
					"operand2":  5.2,
				},
				Output: map[string]interface{}{
					"result":    15.7,
					"operation": "add",
					"operand1":  10.5,
					"operand2":  5.2,
					"success":   true,
				},
				Explanation: "Addition simply sums the two operands",
			},
			{
				Name:        "Square root",
				Description: "Calculate the square root of a number",
				Scenario:    "When you need to find what number squared equals your input",
				Input: map[string]interface{}{
					"operation": "sqrt",
					"operand1":  16,
				},
				Output: map[string]interface{}{
					"result":    4.0,
					"operation": "sqrt",
					"operand1":  16,
					"success":   true,
				},
				Explanation: "Square root is a unary operation - only operand1 is used",
			},
			{
				Name:        "Trigonometry with constants",
				Description: "Calculate sine of π/2",
				Scenario:    "When working with angles and trigonometric functions",
				Input: map[string]interface{}{
					"operation": "sin",
					"operand1":  "pi",
					"operand2":  2,
				},
				Output: map[string]interface{}{
					"result":    1.0,
					"operation": "sin",
					"operand1":  1.5707963267948966, // π/2
					"success":   true,
				},
				Explanation: "First divide pi by 2, then calculate sin. Note: operand1 can be a constant name",
			},
			{
				Name:        "Get mathematical constant",
				Description: "Retrieve the value of π",
				Scenario:    "When you need a precise mathematical constant",
				Input: map[string]interface{}{
					"operation": "pi",
				},
				Output: map[string]interface{}{
					"result":    3.141592653589793,
					"operation": "pi",
					"success":   true,
				},
				Explanation: "Constants don't require operands",
			},
			{
				Name:        "Division by zero error",
				Description: "Handle division by zero gracefully",
				Scenario:    "Error handling example",
				Input: map[string]interface{}{
					"operation": "divide",
					"operand1":  10,
					"operand2":  0,
				},
				Output: map[string]interface{}{
					"operation": "divide",
					"operand1":  10,
					"operand2":  0,
					"success":   false,
					"error":     "division by zero",
				},
				Explanation: "The tool returns errors in the result rather than throwing exceptions",
			},
			{
				Name:        "Factorial calculation",
				Description: "Calculate 5!",
				Scenario:    "When you need to calculate permutations or combinations",
				Input: map[string]interface{}{
					"operation": "factorial",
					"operand1":  5,
				},
				Output: map[string]interface{}{
					"result":    120,
					"operation": "factorial",
					"operand1":  5,
					"success":   true,
				},
				Explanation: "5! = 5 × 4 × 3 × 2 × 1 = 120",
			},
			{
				Name:        "Logarithm with custom base",
				Description: "Calculate log base 2 of 8",
				Scenario:    "When you need logarithms in bases other than e or 10",
				Input: map[string]interface{}{
					"operation": "log",
					"operand1":  8,
					"operand2":  2,
				},
				Output: map[string]interface{}{
					"result":    3,
					"operation": "log",
					"operand1":  8,
					"operand2":  2,
					"success":   true,
				},
				Explanation: "log₂(8) = 3 because 2³ = 8",
			},
		}).
		WithConstraints([]string{
			"Angles for trigonometric functions must be in radians, not degrees",
			"Square root requires non-negative numbers",
			"Logarithms require positive numbers",
			"Division by zero is not allowed",
			"Factorial maximum input is 170 (171! overflows float64)",
			"Factorial requires non-negative integers",
			"GCD and LCM require positive integers",
			"Inverse trig functions (asin, acos) require input between -1 and 1",
			"Results may have floating-point precision limitations",
		}).
		WithErrorGuidance(map[string]string{
			"division by zero": "Cannot divide by zero. Check that operand2 is not zero for division",
			"modulo by zero":   "Cannot calculate modulo with zero divisor. Ensure operand2 is not zero",
			"cannot take square root of negative number": "Square root of negative numbers results in complex numbers, which this tool doesn't support. Use only non-negative values",
			"logarithm of non-positive number":           "Logarithm is only defined for positive numbers. Ensure operand1 is greater than 0",
			"invalid logarithm base":                     "Logarithm base must be positive and not equal to 1",
			"asin domain error":                          "asin requires input between -1 and 1 inclusive",
			"acos domain error":                          "acos requires input between -1 and 1 inclusive",
			"factorial requires non-negative integer":    "Factorial is only defined for non-negative integers (0, 1, 2, ...)",
			"gcd requires positive integers":             "GCD (Greatest Common Divisor) only works with positive whole numbers",
			"lcm requires positive integers":             "LCM (Least Common Multiple) only works with positive whole numbers",
			"unsupported operation":                      "The operation you specified is not recognized. Check the usage instructions for valid operations",
			"result is NaN":                              "The calculation resulted in 'Not a Number' - this typically happens with invalid operations like 0/0",
			"result is infinite":                         "The calculation resulted in infinity - the result is too large to represent",
		}).
		WithCategory("math").
		WithTags([]string{"math", "calculation", "arithmetic", "trigonometry", "logarithm", "statistics"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "fast")

	return builder.Build()
}

// calculatorExecute is the main calculation logic
func calculatorExecute(ctx *domain.ToolContext, params CalculatorParams) (*CalculatorResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
			ToolName:   "calculator",
			Parameters: params,
			RequestID:  ctx.RunID,
		})
	}

	operation := strings.ToLower(params.Operation)

	// Handle mathematical constants
	switch operation {
	case "pi":
		return &CalculatorResult{
			Result:    math.Pi,
			Operation: operation,
			Success:   true,
		}, nil
	case "e":
		return &CalculatorResult{
			Result:    math.E,
			Operation: operation,
			Success:   true,
		}, nil
	case "phi": // Golden ratio
		return &CalculatorResult{
			Result:    math.Phi,
			Operation: operation,
			Success:   true,
		}, nil
	case "tau": // 2π
		return &CalculatorResult{
			Result:    2 * math.Pi,
			Operation: operation,
			Success:   true,
		}, nil
	case "sqrt2": // Square root of 2
		return &CalculatorResult{
			Result:    math.Sqrt2,
			Operation: operation,
			Success:   true,
		}, nil
	case "sqrte": // Square root of e
		return &CalculatorResult{
			Result:    math.SqrtE,
			Operation: operation,
			Success:   true,
		}, nil
	case "sqrtpi": // Square root of pi
		return &CalculatorResult{
			Result:    math.SqrtPi,
			Operation: operation,
			Success:   true,
		}, nil
	case "sqrtphi": // Square root of phi
		return &CalculatorResult{
			Result:    math.SqrtPhi,
			Operation: operation,
			Success:   true,
		}, nil
	case "ln2": // Natural log of 2
		return &CalculatorResult{
			Result:    math.Ln2,
			Operation: operation,
			Success:   true,
		}, nil
	case "ln10": // Natural log of 10
		return &CalculatorResult{
			Result:    math.Ln10,
			Operation: operation,
			Success:   true,
		}, nil
	case "log2e": // Log base 2 of e
		return &CalculatorResult{
			Result:    math.Log2E,
			Operation: operation,
			Success:   true,
		}, nil
	case "log10e": // Log base 10 of e
		return &CalculatorResult{
			Result:    math.Log10E,
			Operation: operation,
			Success:   true,
		}, nil
	}

	// For other operations, we need at least operand1
	result := &CalculatorResult{
		Operation: operation,
		Operand1:  params.Operand1,
		Operand2:  params.Operand2,
	}

	var calcResult float64
	var err error

	// Perform calculation based on operation
	switch operation {
	// Basic arithmetic
	case "add", "+":
		calcResult = params.Operand1 + params.Operand2
	case "subtract", "-":
		calcResult = params.Operand1 - params.Operand2
	case "multiply", "*":
		calcResult = params.Operand1 * params.Operand2
	case "divide", "/":
		if params.Operand2 == 0 {
			err = fmt.Errorf("division by zero")
			break
		}
		calcResult = params.Operand1 / params.Operand2
	case "power", "^", "**":
		calcResult = math.Pow(params.Operand1, params.Operand2)
	case "mod", "%":
		if params.Operand2 == 0 {
			err = fmt.Errorf("modulo by zero")
			break
		}
		calcResult = math.Mod(params.Operand1, params.Operand2)
	case "abs":
		calcResult = math.Abs(params.Operand1)

	// Roots and logarithms
	case "sqrt":
		if params.Operand1 < 0 {
			err = fmt.Errorf("cannot take square root of negative number")
			break
		}
		calcResult = math.Sqrt(params.Operand1)
	case "cbrt":
		calcResult = math.Cbrt(params.Operand1)
	case "log":
		if params.Operand1 <= 0 {
			err = fmt.Errorf("logarithm of non-positive number")
			break
		}
		// Natural logarithm if no base specified
		if params.Operand2 == 0 {
			calcResult = math.Log(params.Operand1)
		} else {
			// Log with custom base
			if params.Operand2 <= 0 || params.Operand2 == 1 {
				err = fmt.Errorf("invalid logarithm base")
				break
			}
			calcResult = math.Log(params.Operand1) / math.Log(params.Operand2)
		}
	case "log10":
		if params.Operand1 <= 0 {
			err = fmt.Errorf("logarithm of non-positive number")
			break
		}
		calcResult = math.Log10(params.Operand1)
	case "log2":
		if params.Operand1 <= 0 {
			err = fmt.Errorf("logarithm of non-positive number")
			break
		}
		calcResult = math.Log2(params.Operand1)
	case "exp":
		calcResult = math.Exp(params.Operand1)

	// Trigonometry (angles in radians)
	case "sin":
		calcResult = math.Sin(params.Operand1)
	case "cos":
		calcResult = math.Cos(params.Operand1)
	case "tan":
		calcResult = math.Tan(params.Operand1)
	case "asin":
		if params.Operand1 < -1 || params.Operand1 > 1 {
			err = fmt.Errorf("asin domain error: value must be between -1 and 1")
			break
		}
		calcResult = math.Asin(params.Operand1)
	case "acos":
		if params.Operand1 < -1 || params.Operand1 > 1 {
			err = fmt.Errorf("acos domain error: value must be between -1 and 1")
			break
		}
		calcResult = math.Acos(params.Operand1)
	case "atan":
		calcResult = math.Atan(params.Operand1)
	case "sinh":
		calcResult = math.Sinh(params.Operand1)
	case "cosh":
		calcResult = math.Cosh(params.Operand1)
	case "tanh":
		calcResult = math.Tanh(params.Operand1)

	// Rounding
	case "floor":
		calcResult = math.Floor(params.Operand1)
	case "ceil":
		calcResult = math.Ceil(params.Operand1)
	case "round":
		calcResult = math.Round(params.Operand1)

	// Advanced operations
	case "factorial":
		if params.Operand1 < 0 || params.Operand1 != math.Floor(params.Operand1) {
			err = fmt.Errorf("factorial requires non-negative integer")
			break
		}
		calcResult = factorial(int(params.Operand1))
	case "gcd":
		if params.Operand1 <= 0 || params.Operand2 <= 0 ||
			params.Operand1 != math.Floor(params.Operand1) ||
			params.Operand2 != math.Floor(params.Operand2) {
			err = fmt.Errorf("gcd requires positive integers")
			break
		}
		calcResult = float64(gcd(int(params.Operand1), int(params.Operand2)))
	case "lcm":
		if params.Operand1 <= 0 || params.Operand2 <= 0 ||
			params.Operand1 != math.Floor(params.Operand1) ||
			params.Operand2 != math.Floor(params.Operand2) {
			err = fmt.Errorf("lcm requires positive integers")
			break
		}
		calcResult = float64(lcm(int(params.Operand1), int(params.Operand2)))

	default:
		err = fmt.Errorf("unsupported operation: %s", operation)
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if ctx.Events != nil {
			ctx.Events.EmitError(err)
		}
		return result, nil // Return error in result, not as error
	}

	// Check for special float values
	if math.IsNaN(calcResult) {
		result.Success = false
		result.Error = "result is NaN"
	} else if math.IsInf(calcResult, 0) {
		result.Success = false
		result.Error = "result is infinite"
	} else {
		result.Success = true
		result.Result = calcResult
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
			ToolName:  "calculator",
			Result:    result,
			RequestID: ctx.RunID,
		})
	}

	return result, nil
}

// Helper functions

// factorial calculates n!
func factorial(n int) float64 {
	if n < 0 || n > 170 { // 171! overflows float64
		return math.NaN()
	}
	if n == 0 || n == 1 {
		return 1
	}
	result := 1.0
	for i := 2; i <= n; i++ {
		result *= float64(i)
	}
	return result
}

// gcd calculates greatest common divisor
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// lcm calculates least common multiple
func lcm(a, b int) int {
	return (a * b) / gcd(a, b)
}

// MustGetCalculator retrieves the registered Calculator tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetCalculator() domain.Tool {
	return tools.MustGetTool("calculator")
}
