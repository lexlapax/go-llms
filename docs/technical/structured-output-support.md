# Structured Output Support

The `outputs` package provides comprehensive support for parsing, validating, and converting structured outputs from LLMs.

## Features

- **Format Detection & Parsing**: Automatically detect and parse JSON, YAML, and XML outputs
- **Recovery Mechanisms**: Handle malformed LLM outputs with intelligent recovery
- **Schema Validation**: Validate outputs against defined schemas with detailed error reporting
- **Format Conversion**: Convert between JSON, YAML, and XML formats
- **Bridge Integration**: Support for go-llmspell scripting engine integration

## Components

### Parser Interface
```go
type Parser interface {
    Name() string
    Parse(ctx context.Context, output string) (interface{}, error)
    ParseWithRecovery(ctx context.Context, output string, opts *RecoveryOptions) (interface{}, error)
    ParseWithSchema(ctx context.Context, output string, schema *OutputSchema) (interface{}, error)
    CanParse(output string) bool
}
```

### Available Parsers
- **JSONParser**: Handles JSON with recovery for common LLM issues
- **YAMLParser**: Parses YAML with indentation fixes
- **XMLParser**: Processes XML with tag recovery

### Recovery Options
```go
type RecoveryOptions struct {
    ExtractFromMarkdown bool  // Extract from code blocks
    FixCommonIssues     bool  // Fix trailing commas, quotes, etc.
    StrictMode          bool  // Disable recovery
    MaxAttempts         int   // Max recovery attempts
    Schema              *OutputSchema // Schema guidance
}
```

## Usage Examples

### Basic Parsing
```go
parser := outputs.NewJSONParser()
result, err := parser.Parse(ctx, llmOutput)
```

### Parsing with Recovery
```go
parser := outputs.NewJSONParser()
result, err := parser.ParseWithRecovery(ctx, llmOutput, &outputs.RecoveryOptions{
    ExtractFromMarkdown: true,
    FixCommonIssues:     true,
    MaxAttempts:         3,
})
```

### Schema Validation
```go
schema := &outputs.OutputSchema{
    Type: outputs.TypeObject,
    Properties: map[string]*outputs.OutputSchema{
        "name": {Type: outputs.TypeString, Required: boolPtr(true)},
        "age":  {Type: outputs.TypeInteger, Minimum: float64Ptr(0)},
    },
    RequiredProperties: []string{"name"},
}

validator := outputs.NewValidator()
result, err := validator.Validate(ctx, data, schema)
```

### Format Conversion
```go
converter := outputs.NewConverter()
yamlOutput, err := converter.ConvertString(ctx, jsonInput, 
    outputs.FormatJSON, outputs.FormatYAML, nil)
```

### Bridge Integration
```go
bridge := outputs.NewBridgeAdapter()
result, err := bridge.ParseAndValidate(ctx, llmOutput, schema)
```

## Common Recovery Scenarios

The parsers handle common LLM output issues:

1. **Markdown Code Blocks**: Extracts JSON/YAML/XML from code blocks
2. **Trailing Commas**: Removes trailing commas in objects/arrays
3. **Quote Issues**: Fixes single quotes and unquoted keys
4. **Missing Decimals**: Fixes `.5` to `0.5`
5. **Unclosed Tags**: Attempts to close unclosed XML tags
6. **Text Wrapping**: Extracts structured data from explanatory text

## Schema Types

- `TypeString`: String values with format and pattern support
- `TypeNumber`: Numeric values with min/max constraints
- `TypeInteger`: Integer values
- `TypeBoolean`: Boolean values
- `TypeArray`: Arrays with item schemas and length constraints
- `TypeObject`: Objects with property schemas

## Integration with go-llmspell

The bridge adapter provides seamless integration with go-llmspell:

```go
// Convert go-llmspell schema format
schema, err := bridge.ConvertSchemaFromBridge(bridgeSchema)

// Parse and validate in one step
result, err := bridge.ParseAndValidate(ctx, output, schema)

// Fix malformed outputs
fixed, err := bridge.FixOutput(ctx, output, hints)
```

## Error Handling

Validation provides detailed error information:

```go
type ValidationError struct {
    Path     string  // JSON path to error
    Field    string  // Field name
    Message  string  // Error description
    Code     string  // Error code
    Expected string  // What was expected
    Actual   string  // What was found
}
```

## Performance Considerations

- Parsers attempt direct parsing before recovery
- Recovery attempts are limited by MaxAttempts
- Schema validation is performed recursively
- Format conversion preserves type information when possible

## Future Enhancements

- Streaming parser support
- Additional format support (TOML, MessagePack)
- Schema inference from examples
- LLM-specific output patterns
- Advanced recovery strategies