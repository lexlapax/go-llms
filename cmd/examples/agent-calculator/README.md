# Calculator Tool with LLM Agent Example

This example demonstrates the use of the **enhanced** built-in calculator tool (v2.0.0) in three ways:
1. **Direct tool usage** - Calling the calculator tool directly from code
2. **LLM agent integration** - Using the calculator tool through an LLM agent with enhanced metadata
3. **Tool information display** - Viewing the comprehensive tool documentation and metadata

## Enhanced Calculator Tool (v2.0.0)

The calculator tool now includes comprehensive LLM guidance with the ToolBuilder pattern:

### Features
- **Rich metadata**: Usage instructions, examples, constraints, and error guidance
- **Smart parameter handling**: Accepts numbers or constant names as strings
- **Comprehensive error messages**: Maps common errors to helpful guidance
- **MCP compatibility**: Exports to Model Context Protocol format

### Supported Operations

#### Basic Arithmetic
- Addition (`add`, `+`)
- Subtraction (`subtract`, `-`)
- Multiplication (`multiply`, `*`)
- Division (`divide`, `/`)
- Power (`power`, `^`, `**`)
- Modulo (`mod`, `%`)
- Absolute value (`abs`)

#### Scientific Functions
- Square root (`sqrt`)
- Cube root (`cbrt`)
- Natural logarithm (`log`)
- Base 10 logarithm (`log10`)
- Base 2 logarithm (`log2`)
- Exponential (`exp`)
- Custom base logarithm (use `log` with `operand2` as base)

#### Trigonometry (angles in radians)
- Sine (`sin`)
- Cosine (`cos`)
- Tangent (`tan`)
- Arcsine (`asin`)
- Arccosine (`acos`)
- Arctangent (`atan`)
- Hyperbolic sine (`sinh`)
- Hyperbolic cosine (`cosh`)
- Hyperbolic tangent (`tanh`)

#### Rounding
- Floor (`floor`)
- Ceiling (`ceil`)
- Round (`round`)

#### Advanced Operations
- Factorial (`factorial`) - max input: 170
- Greatest Common Divisor (`gcd`)
- Least Common Multiple (`lcm`)

#### Mathematical Constants
- Pi (`pi`, `π`) - 3.14159...
- Euler's number (`e`) - 2.71828...
- Golden ratio (`phi`, `φ`) - 1.61803...
- Tau (`tau`, `τ`) - 2π
- Square roots: `sqrt2`, `sqrte`, `sqrtpi`, `sqrtphi`
- Logarithmic constants: `ln2`, `ln10`, `log2e`, `log10e`

## Running the Example

### 1. LLM Agent Mode (Default)
```bash
go run main.go
```

This mode demonstrates using the calculator tool through an LLM agent. The agent:
- Accepts natural language prompts
- Automatically uses the calculator tool based on its enhanced metadata
- Leverages the tool's usage instructions and examples
- Handles errors using the tool's error guidance
- Returns results in natural language

### 2. Direct Tool Usage
```bash
go run main.go direct
```

This mode demonstrates direct usage of the calculator tool without an LLM. It shows:
- Basic arithmetic operations
- Scientific functions with mathematical constants
- Trigonometry with radians
- Advanced operations (factorial, GCD, LCM)
- Using constant names as operands (e.g., "pi", "phi")
- Comprehensive error handling

### 3. Tool Information Mode
```bash
go run main.go info
```

This mode displays the calculator tool's comprehensive metadata:
- Version, category, and tags
- Behavioral characteristics (deterministic, non-destructive, fast)
- Full usage instructions
- All constraints and limitations
- Number of available examples

## Environment Variables

### LLM Provider Configuration
For real LLM providers (instead of mock), set one of:
- `OPENAI_API_KEY` - For OpenAI GPT models
- `ANTHROPIC_API_KEY` - For Anthropic Claude models  
- `GEMINI_API_KEY` - For Google Gemini models

Without API keys, the example uses a mock provider that simulates tool usage.

### Debug Logging
- `DEBUG=1` - Enable detailed logging to see tool calls and agent reasoning

## How It Works

1. **Tool Registration**: The calculator tool is automatically registered when the package is imported
2. **Tool Discovery**: The example retrieves the tool from the registry
3. **Direct Execution**: In direct mode, the tool is called with typed parameters
4. **LLM Integration**: In LLM mode, the agent:
   - Receives natural language input
   - Generates tool calls in JSON format
   - Executes the calculator tool
   - Formats results in natural language

## Example Output

### Direct Mode
```
=== Built-in Calculator Tool Example (Direct Usage) ===

Tool: calculator
Description: Performs mathematical calculations including arithmetic, trigonometry, and logarithms

--- Basic Arithmetic ---
10.5 + 5.2 = 15.700000
20 / 4 = 5.000000
2^8 = 256.000000

--- Mathematical Constants ---
φ (golden ratio) = 1.618034
φ² (phi squared) = 2.618034
τ (tau = 2π) = 6.283185

--- Error Handling ---
Expected error for division by zero: 10 / 0 = ERROR: division by zero
Expected error for sqrt(-4): √(-4) = ERROR: cannot take square root of negative number
```

### LLM Mode
```
=== Built-in Calculator Tool with LLM Agent ===

Provider: anthropic
Model: claude-3-7-sonnet-latest

=== Example 1: Basic Arithmetic ===
Response: 25 multiplied by 17 equals 425.

=== Example 9: Error Handling ===
Response: I cannot calculate the square root of -16 because the square root of negative numbers results in complex numbers, which this calculator tool doesn't support. The tool only works with real numbers.

=== Example 10: Complex Calculation ===
Response: To calculate phi squared minus the square root of 5:
- Phi squared (φ²) = 2.618034
- Square root of 5 = 2.236068
- Result: 2.618034 - 2.236068 = 0.381966
```

### Tool Information Mode
```
=== Calculator Tool Information ===
Name: calculator
Description: Performs mathematical calculations including arithmetic, trigonometry, and logarithms
Version: 2.0.0
Category: math
Tags: [math calculation arithmetic trigonometry logarithm statistics]
Deterministic: true
Destructive: false
Requires Confirmation: false
Estimated Latency: fast

Usage Instructions:
Use this tool to perform mathematical calculations. It supports:

Basic Arithmetic:
- add (+): Addition of two numbers
- subtract (-): Subtraction (operand1 - operand2)
...

Constraints:
- Angles for trigonometric functions must be in radians, not degrees
- Square root requires non-negative numbers
- Logarithms require positive numbers
- Division by zero is not allowed
- Factorial maximum input is 170 (171! overflows float64)
...

Examples available: 7
```

## What's New in v2.0.0

The calculator tool has been enhanced with the ToolBuilder pattern:

1. **Comprehensive Usage Instructions**: Detailed guidance for every operation
2. **Rich Examples**: 7 examples showing various use cases with input/output
3. **Explicit Constraints**: 9 documented limitations (e.g., factorial max 170)
4. **Error Guidance**: 13 error scenarios mapped to helpful messages
5. **Extended Constants**: Added phi, tau, and various mathematical constants
6. **Smart Parameter Handling**: Accept constant names as strings ("pi", "e", etc.)
7. **MCP Export**: Full compatibility with Model Context Protocol
8. **Better LLM Integration**: Enhanced metadata helps LLMs use the tool correctly

## Key Features

- **Type-safe parameters**: Uses structured types for tool inputs
- **Comprehensive error handling**: Validates inputs and handles edge cases with helpful guidance
- **Event emission**: Emits tool call and result events for observability
- **Natural language interface**: LLM converts between natural language and tool calls
- **Enhanced metadata**: Rich documentation for better LLM understanding
- **Mock provider support**: Can run without API keys for testing

## Code Structure

- `createToolContext()`: Creates minimal context for direct tool usage
- `runDirectExample()`: Demonstrates all calculator operations
- `runLLMExample()`: Shows LLM agent integration
- `createLLMProvider()`: Creates provider from environment or mock
- `printResult()`: Formats calculator results for display

This example showcases the power of built-in tools and how they integrate seamlessly with LLM agents for natural language interfaces.