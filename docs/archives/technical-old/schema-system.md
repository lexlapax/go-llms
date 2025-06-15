# Schema System Overview

> **[Documentation Home](/docs/README.md) / [Technical Documentation](/docs/technical/README.md) / Schema System**

This document provides a comprehensive overview of the schema system in go-llms, covering schema generation, storage, validation, and structured output handling.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Schema Generation](#schema-generation)
4. [Schema Storage](#schema-storage)
5. [Schema Validation](#schema-validation)
6. [Structured Output Handling](#structured-output-handling)
7. [Integration with go-llmspell](#integration-with-go-llmspell)
8. [Usage Examples](#usage-examples)
9. [Best Practices](#best-practices)
10. [Related Documentation](#related-documentation)

## Overview

The go-llms schema system provides end-to-end support for working with structured data:

1. **Generation**: Create schemas from Go structs
2. **Storage**: Version and manage schemas
3. **Validation**: Validate data against schemas
4. **Output Handling**: Parse and recover structured outputs from LLMs

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Applications                            │
│                   (go-llmspell, etc.)                       │
└─────────────────┬─────────────────────┬────────────────────┘
                  │                     │
┌─────────────────▼─────────────────────▼────────────────────┐
│              Schema Package         Outputs Package          │
├─────────────────────────────────────────────────────────────┤
│  • Schema Generation               • Output Parsing          │
│  • Schema Storage                  • Format Recovery         │
│  • Version Management              • Schema Validation       │
│  • Type Handlers                   • Format Conversion       │
└─────────────────┬─────────────────────┬────────────────────┘
                  │                     │
┌─────────────────▼─────────────────────▼────────────────────┐
│                    Core Domain                              │
│                 (Schema Definition)                         │
└─────────────────────────────────────────────────────────────┘
```

## Schema Generation

The schema package provides two generators for creating schemas from Go structs:

### ReflectionSchemaGenerator

Automatically generates schemas using reflection:

```go
gen := generator.NewReflectionSchemaGenerator()

type User struct {
    ID    string `json:"id" validate:"required,uuid" format:"uuid"`
    Name  string `json:"name" validate:"required,min=1,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age,omitempty" validate:"min=0,max=150"`
}

schema, err := gen.GenerateSchema(User{})
```

### TagSchemaGenerator

Generates schemas primarily from struct tags:

```go
gen := generator.NewTagSchemaGenerator()

type Example struct {
    Field string `schema:"type=string,format=email,required"`
}

schema, err := gen.GenerateSchema(Example{})
```

### Custom Type Handlers

Register handlers for custom types:

```go
gen.RegisterTypeHandler(reflect.TypeOf(time.Duration(0)), 
    func(t reflect.Type, tag reflect.StructTag) (domain.Property, error) {
        return domain.Property{
            Type:        "string",
            Format:      "duration",
            Description: "Duration in string format",
        }, nil
    })
```

## Schema Storage

### InMemorySchemaRepository

Thread-safe in-memory storage with versioning:

```go
repo := repository.NewInMemorySchemaRepository()

// Save a schema
err := repo.Save("user-schema", schema)

// Retrieve versions
current, err := repo.Get("user-schema")
v1, err := repo.GetVersion("user-schema", 1)

// Export/Import
data, err := repo.Export()
```

### FileSchemaRepository

Persistent file-based storage:

```go
repo, err := repository.NewFileSchemaRepository("/path/to/schemas")
// Same API as InMemorySchemaRepository
```

## Schema Validation

The validation system supports comprehensive JSON Schema features:

### Core Features
- Type validation (string, number, integer, boolean, object, array)
- Constraint validation (min/max, pattern, enum, format)
- Required field validation
- Nested object and array validation
- Format validation (email, uri, uuid, etc.)

### Usage

```go
validator := outputs.NewValidator()

schema := &outputs.OutputSchema{
    Type: outputs.TypeObject,
    Properties: map[string]*outputs.OutputSchema{
        "name": {Type: outputs.TypeString, Required: boolPtr(true)},
        "age":  {Type: outputs.TypeInteger, Minimum: float64Ptr(0)},
    },
    RequiredProperties: []string{"name"},
}

result, err := validator.Validate(ctx, data, schema)
```

## Structured Output Handling

The outputs package handles parsing and recovery of LLM outputs:

### Parsers
- **JSONParser**: Handles JSON with recovery
- **YAMLParser**: Parses YAML with fixes
- **XMLParser**: Processes XML with recovery

### Recovery Features
1. Extract from markdown code blocks
2. Fix trailing commas
3. Handle quote issues
4. Fix missing decimals
5. Close unclosed tags
6. Extract from explanatory text

### Usage

```go
// Basic parsing
parser := outputs.NewJSONParser()
result, err := parser.Parse(ctx, llmOutput)

// With recovery
result, err := parser.ParseWithRecovery(ctx, llmOutput, &outputs.RecoveryOptions{
    ExtractFromMarkdown: true,
    FixCommonIssues:     true,
})

// Format conversion
converter := outputs.NewConverter()
yamlOutput, err := converter.ConvertString(ctx, jsonInput, 
    outputs.FormatJSON, outputs.FormatYAML, nil)
```

## Integration with go-llmspell

The system provides seamless integration with the go-llmspell scripting engine:

### Schema Registration

```go
// Register schemas dynamically
repo := repository.NewInMemorySchemaRepository()
err := repo.Save("tool-input-schema", toolSchema)

// Bridge adapter for go-llmspell
bridge := outputs.NewBridgeAdapter()
schema, err := bridge.ConvertSchemaFromBridge(bridgeSchema)
```

### Workflow Integration

```go
// 1. Generate schema from struct
gen := generator.NewReflectionSchemaGenerator()
schema, err := gen.GenerateSchema(MyToolInput{})

// 2. Store schema
repo.Save("my-tool-input", schema)

// 3. Parse LLM output
parser := outputs.NewJSONParser()
data, err := parser.ParseWithRecovery(ctx, llmOutput, recoveryOpts)

// 4. Validate against schema
validator := outputs.NewValidator()
result, err := validator.Validate(ctx, data, schema)
```

## Usage Examples

### Complete Tool Registration Flow

```go
// Define tool input structure
type SearchInput struct {
    Query    string   `json:"query" validate:"required,min=1"`
    Filters  []string `json:"filters,omitempty"`
    MaxItems int      `json:"max_items" validate:"min=1,max=100"`
}

// Generate and register schema
gen := generator.NewReflectionSchemaGenerator()
schema, err := gen.GenerateSchema(SearchInput{})

repo := repository.NewFileSchemaRepository("./schemas")
err = repo.Save("search-tool-input", schema)

// Use in tool execution
func (t *SearchTool) Execute(ctx context.Context, input string) (interface{}, error) {
    // Parse LLM output
    parser := outputs.NewJSONParser()
    data, err := parser.ParseWithRecovery(ctx, input, &outputs.RecoveryOptions{
        ExtractFromMarkdown: true,
        FixCommonIssues:     true,
    })
    
    // Validate against schema
    schema, err := repo.Get("search-tool-input")
    validator := outputs.NewValidator()
    result, err := validator.Validate(ctx, data, schema)
    
    // Execute tool logic
    // ...
}
```

## Best Practices

### Schema Generation
1. Use ReflectionSchemaGenerator for automatic generation
2. Use TagSchemaGenerator for fine control
3. Register custom type handlers for domain types
4. Include validation tags for constraints

### Schema Storage
1. Use versioning for schema evolution
2. Export schemas for backup/migration
3. Use FileSchemaRepository for persistence
4. Implement schema migration strategies

### Output Handling
1. Always use recovery options for LLM outputs
2. Provide schemas for better recovery guidance
3. Log recovery attempts for debugging
4. Test with real LLM outputs

### Performance
1. Cache parsed schemas
2. Reuse parser instances
3. Limit recovery attempts
4. Use streaming for large outputs (future)

## Related Documentation

- [Schema Package Details](schema-package.md) - In-depth schema package documentation
- [Structured Output Support](structured-output-support.md) - Output parsing and recovery details
- [Testing Framework](testing.md) - Testing schemas and outputs
- [go-llmspell Integration](https://github.com/yourusername/go-llmspell) - Scripting engine integration