# Structured Output Example

This example demonstrates the structured output parsing and validation capabilities of go-llms.

## Features Demonstrated

1. **JSON Parsing with Recovery** - Parse LLM outputs that may contain markdown formatting or common JSON errors
2. **Schema Validation** - Validate parsed data against defined schemas with detailed error reporting
3. **Format Conversion** - Convert between JSON, YAML, and XML formats
4. **Bridge Integration** - Integration with go-llmspell scripting engine

## Running the Example

```bash
go run main.go
```

## Example Output

The example shows four different use cases:

### 1. JSON Parsing with Recovery
Demonstrates parsing JSON from LLM output that includes markdown code blocks and trailing commas:
- Extracts JSON from markdown code blocks
- Fixes common JSON issues (trailing commas, etc.)
- Provides multiple recovery strategies

### 2. Schema Validation
Shows how to validate data against a schema:
- Define schemas with type constraints
- Validate required fields
- Check format constraints (email, etc.)
- Get detailed error messages and fix suggestions

### 3. Format Conversion
Converts data between different formats:
- JSON to YAML conversion
- JSON to XML conversion with custom root element
- Pretty printing options

### 4. Bridge Integration
Demonstrates go-llmspell integration:
- Convert bridge schemas to OutputSchema format
- Parse and validate LLM outputs
- Get parser information and capabilities

## Code Structure

The example is organized into separate functions for each feature:
- `parseJSONWithRecovery()` - JSON parsing with error recovery
- `validateAgainstSchema()` - Schema-based validation
- `convertFormats()` - Format conversion examples
- `bridgeIntegration()` - Bridge adapter usage

## Related Documentation

- [Structured Output Package Documentation](../../../pkg/llm/outputs/README.md)
- [Output Parser API](../../../pkg/llm/outputs/parser.go)
- [Validator API](../../../pkg/llm/outputs/validator.go)