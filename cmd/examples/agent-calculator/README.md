# Calculator Tool with LLM Agent Example

This example demonstrates the use of the built-in calculator tool in two ways:
1. **Direct tool usage** - Calling the calculator tool directly from code
2. **LLM agent integration** - Using the calculator tool through an LLM agent

## Built-in Calculator Tool

The calculator tool provides a comprehensive set of mathematical operations:

### Basic Arithmetic
- Addition (`add`, `+`)
- Subtraction (`subtract`, `-`)
- Multiplication (`multiply`, `*`)
- Division (`divide`, `/`)
- Power (`power`, `^`, `**`)
- Modulo (`mod`, `%`)
- Absolute value (`abs`)

### Scientific Functions
- Square root (`sqrt`)
- Cube root (`cbrt`)
- Natural logarithm (`log`)
- Base 10 logarithm (`log10`)
- Base 2 logarithm (`log2`)
- Exponential (`exp`)
- Custom base logarithm (use `log` with `operand2` as base)

### Trigonometry
- Sine (`sin`)
- Cosine (`cos`)
- Tangent (`tan`)
- Arcsine (`asin`)
- Arccosine (`acos`)
- Arctangent (`atan`)
- Hyperbolic sine (`sinh`)
- Hyperbolic cosine (`cosh`)
- Hyperbolic tangent (`tanh`)

### Rounding
- Floor (`floor`)
- Ceiling (`ceil`)
- Round (`round`)

### Advanced Operations
- Factorial (`factorial`)
- Greatest Common Divisor (`gcd`)
- Least Common Multiple (`lcm`)

### Mathematical Constants
- Pi (`pi`)
- Euler's number (`e`)

## Running the Example

### Direct Tool Usage (Default)
```bash
go run main.go
```

This mode demonstrates direct usage of the calculator tool without an LLM. It shows:
- Basic arithmetic operations
- Scientific functions
- Trigonometry
- Advanced operations
- Error handling

### LLM Agent Mode
```bash
go run main.go llm
```

This mode demonstrates using the calculator tool through an LLM agent. The agent:
- Accepts natural language prompts
- Determines when to use the calculator tool
- Calls the tool with appropriate parameters
- Returns results in natural language

## Environment Variables

For real LLM providers (instead of mock), set one of:
- `OPENAI_API_KEY` - For OpenAI GPT models
- `ANTHROPIC_API_KEY` - For Anthropic Claude models
- `GEMINI_API_KEY` - For Google Gemini models
- `GO_LLMS_*` - For GO_LLMS environment configuration

Without API keys, the example uses a mock provider that simulates tool usage.

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
...
```

### LLM Mode
```
=== Built-in Calculator Tool with LLM Agent ===

--- Prompt: What is 25 * 17? ---
Response: The result of multiplying 25 by 17 is 425.

--- Prompt: Calculate the square root of 144 ---
Response: The square root of 144 is 12.
...
```

## Key Features

- **Type-safe parameters**: Uses structured types for tool inputs
- **Comprehensive error handling**: Validates inputs and handles edge cases
- **Event emission**: Emits tool call and result events for observability
- **Natural language interface**: LLM converts between natural language and tool calls
- **Mock provider support**: Can run without API keys for testing

## Code Structure

- `createToolContext()`: Creates minimal context for direct tool usage
- `runDirectExample()`: Demonstrates all calculator operations
- `runLLMExample()`: Shows LLM agent integration
- `createLLMProvider()`: Creates provider from environment or mock
- `printResult()`: Formats calculator results for display

This example showcases the power of built-in tools and how they integrate seamlessly with LLM agents for natural language interfaces.