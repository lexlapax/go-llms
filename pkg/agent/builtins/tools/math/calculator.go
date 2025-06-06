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

// Calculator creates a tool for performing mathematical calculations
// This is a built-in tool optimized for:
// - Wide range of mathematical operations
// - Error handling for edge cases (division by zero, domain errors)
// - Support for both unary and binary operations
// - Mathematical constants
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

	return atools.NewTool(
		"calculator",
		"Performs mathematical calculations including arithmetic, trigonometry, and logarithms",
		wrappedFn,
		calculatorParamSchema,
	)
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
