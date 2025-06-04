# Calculator Agent Example

This example demonstrates how to create a custom agent that performs pure computational logic without LLM dependencies. It shows stateless computation patterns, input validation, and error handling.

## Overview

The Calculator Agent is a custom agent that implements mathematical operations using the Go-LLMs agent framework. It demonstrates:

- **Stateless computation patterns**: Pure mathematical functions without external dependencies
- **Input validation and error handling**: Robust parameter checking and meaningful error messages
- **Simple custom agent implementation**: Clean, minimal code structure
- **Integration with workflow agents**: Can be used as a building block in larger workflows

## Features

### Supported Operations

- **Addition**: `add`, `+`
- **Subtraction**: `subtract`, `-`
- **Multiplication**: `multiply`, `*`
- **Division**: `divide`, `/`
- **Power/Exponentiation**: `power`, `^`, `**`
- **Square Root**: `sqrt` (uses operand1 only)
- **Modulo**: `mod`, `%`

### Input Validation

- Type conversion for various numeric types (int, float32, float64, etc.)
- Division by zero detection
- Square root of negative numbers prevention
- Missing parameter detection
- Unsupported operation handling

### Error Handling

- Comprehensive error messages with context
- Graceful handling of edge cases
- Input validation before computation

## Usage

### Basic Usage

```go
// Create calculator agent
calculator := NewCalculatorAgent("calculator")

// Create input state
state := domain.NewState()
state.Set("operation", "add")
state.Set("operand1", 10.5)
state.Set("operand2", 5.2)

// Run calculation
result, err := calculator.Run(ctx, state)
if err != nil {
    log.Fatal(err)
}

// Get result
if value, exists := result.Get("result"); exists {
    fmt.Printf("Result: %.2f\n", value) // Output: 15.70
}
```

### Running the Example

```bash
# Navigate to the calculator directory
cd cmd/examples/agent-custom-calculator

# Run the example
go run main.go

# Or build and run
go build -o calculator .
./calculator
```

### Running Tests

```bash
# Run tests with verbose output
go test -v

# Run tests with coverage
go test -v -cover

# Run specific test
go test -v -run TestCalculatorAgent_BasicOperations
```

## Input State Format

The calculator agent expects the following fields in the input state:

```go
state := domain.NewState()
state.Set("operation", "add")    // Required: operation to perform
state.Set("operand1", 10.5)      // Required: first operand (numeric)
state.Set("operand2", 5.2)       // Required: second operand (numeric)
```

### Supported Numeric Types

The agent automatically converts between numeric types:
- `int`, `int32`, `int64`
- `uint`, `uint32`, `uint64`
- `float32`, `float64`

## Output State Format

The calculator agent adds the following fields to the output state:

```go
// All original state data is preserved
result.Get("result")     // float64: calculation result
result.Get("computed")   // bool: true (indicates computation was performed)
```

## Examples

### Addition
```go
state.Set("operation", "add")
state.Set("operand1", 10.5)
state.Set("operand2", 5.2)
// Result: 15.7
```

### Division
```go
state.Set("operation", "divide")
state.Set("operand1", 20)
state.Set("operand2", 4)
// Result: 5.0
```

### Power
```go
state.Set("operation", "power")
state.Set("operand1", 2)
state.Set("operand2", 8)
// Result: 256.0
```

### Square Root
```go
state.Set("operation", "sqrt")
state.Set("operand1", 16)
state.Set("operand2", 0)  // Not used but required
// Result: 4.0
```

## Error Handling Examples

### Division by Zero
```go
state.Set("operation", "divide")
state.Set("operand1", 10)
state.Set("operand2", 0)
// Error: "division by zero"
```

### Invalid Operation
```go
state.Set("operation", "factorial")
state.Set("operand1", 5)
state.Set("operand2", 0)
// Error: "unsupported operation: factorial"
```

### Invalid Operand Type
```go
state.Set("operation", "add")
state.Set("operand1", "not_a_number")
state.Set("operand2", 5)
// Error: "operand1 must be a number, got string"
```

## Integration with Workflow Agents

The calculator agent can be used as part of larger workflows:

### Sequential Workflow
```go
import "github.com/lexlapax/go-llms/pkg/agent/workflow"

// Create a calculation pipeline
pipeline := workflow.NewSequentialAgent("math-pipeline").
    AddAgent(calculator1).     // First calculation
    AddAgent(calculator2).     // Second calculation
    AddAgent(calculator3)      // Final calculation

result, err := pipeline.Run(ctx, inputState)
```

### Parallel Workflow
```go
// Perform multiple calculations in parallel
parallel := workflow.NewParallelAgent("parallel-math").
    AddAgent(calculator1).     // Calculate sum
    AddAgent(calculator2).     // Calculate product
    AddAgent(calculator3)      // Calculate difference

result, err := parallel.Run(ctx, inputState)
```

### Conditional Workflow
```go
// Choose calculation based on input
conditional := workflow.NewConditionalAgent("smart-calc").
    AddBranch("simple", func(state *domain.State) bool {
        return state.GetString("complexity") == "simple"
    }, simpleCalculator).
    AddBranch("advanced", func(state *domain.State) bool {
        return state.GetString("complexity") == "advanced"
    }, advancedCalculator)
```

## Architecture

The Calculator Agent demonstrates key custom agent patterns:

### State Management
```go
func (c *CalculatorAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Clone input to avoid mutations
    resultState := input.Clone()
    
    // Add computed results
    resultState.Set("result", result)
    resultState.Set("computed", true)
    
    return resultState, nil
}
```

### Event Emission
```go
// Emit start event
c.EmitEvent(domain.EventAgentStart, map[string]interface{}{
    "agent": c.Name(),
    "type":  "calculator",
})

// Emit completion event
c.EmitEvent(domain.EventAgentComplete, map[string]interface{}{
    "agent":     c.Name(),
    "operation": operation,
    "result":    result,
})
```

### Validation
```go
func (c *CalculatorAgent) Validate() error {
    // Call base validation
    if err := c.BaseAgentImpl.Validate(); err != nil {
        return err
    }
    
    // Calculator-specific validation
    // (none needed for this stateless agent)
    return nil
}
```

## Testing

The calculator agent includes comprehensive tests covering:

- **Basic Operations**: All supported mathematical operations
- **Error Cases**: Division by zero, invalid inputs, missing parameters
- **Type Conversion**: Various numeric type combinations
- **State Preservation**: Original data preservation in output state
- **Edge Cases**: Negative square roots, unsupported operations

### Test Coverage

- `TestCalculatorAgent_BasicOperations`: Tests all supported operations
- `TestCalculatorAgent_ErrorCases`: Tests error handling scenarios
- `TestCalculatorAgent_TypeConversion`: Tests numeric type conversion
- `TestCalculatorAgent_Validate`: Tests agent validation
- `TestCalculatorAgent_StatePreservation`: Tests state management
- `TestConvertToFloat64`: Tests type conversion helper function

## Best Practices Demonstrated

1. **State Cloning**: Always clone input state to avoid mutations
2. **Comprehensive Validation**: Check all required parameters before processing
3. **Type Safety**: Safe type conversion with proper error handling
4. **Event Emission**: Emit events for monitoring and debugging
5. **Error Context**: Provide detailed error messages with context
6. **Stateless Design**: No internal state makes the agent thread-safe and predictable

## Use Cases

The Calculator Agent pattern is useful for:

- **Data Processing Pipelines**: Mathematical transformations in data workflows
- **Business Logic**: Complex calculations in business processes
- **Scientific Computing**: Mathematical operations in research workflows
- **Financial Applications**: Currency conversions, interest calculations
- **Engineering Simulations**: Mathematical models and computations

This example demonstrates how custom agents can provide specialized functionality while integrating seamlessly with the broader Go-LLMs agent ecosystem.