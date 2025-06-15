# Documentation Generation

> **[Documentation Home](/docs/README.md) / [Technical Documentation](/docs/technical/README.md) / Documentation Generation**

The documentation generation system in go-llms provides automatic generation of API documentation in multiple formats (OpenAPI, Markdown, JSON) from tool metadata and other components.

## Overview

The `pkg/docs` package implements a comprehensive documentation generation system that integrates with the tool discovery system to provide:

- **OpenAPI 3.0 Specifications** - Complete API documentation for all tools
- **Markdown Documentation** - Human-readable documentation with examples
- **JSON Documentation** - Structured documentation for programmatic access
- **Bridge Integration** - All formats are JSON-serializable for go-llmspell

## Architecture

### Core Components

1. **Generator Interface** (`pkg/docs/generator.go`)
   - `GenerateOpenAPI()` - Generate OpenAPI 3.0 specifications
   - `GenerateMarkdown()` - Generate human-readable Markdown
   - `GenerateJSON()` - Generate structured JSON documentation

2. **Documentable Interface**
   - Components implement this interface to provide documentation
   - Returns `Documentation` struct with metadata, schemas, examples

3. **Tool Integration** (`pkg/docs/tools.go`)
   - `GenerateToolDocumentation()` - Convert ToolInfo to Documentation
   - `GenerateToolOpenAPI()` - Generate OpenAPI for tools
   - Schema conversion utilities

4. **Discovery Integration** (`pkg/docs/integration.go`)
   - `ToolDocumentationIntegrator` - Central integration point
   - Batch operations for all tools
   - Category and tag-based filtering
   - Enhanced tool help generation

## Usage

### Basic Documentation Generation

```go
// Create integrator with discovery system
discovery := tools.NewDiscovery()
config := docs.GeneratorConfig{
    Title:       "My API Documentation",
    Description: "API for my tools",
    Version:     "1.0.0",
}
integrator := docs.NewToolDocumentationIntegrator(discovery, config)

// Generate OpenAPI for all tools
openAPISpec, err := integrator.GenerateOpenAPIForAllTools(ctx)

// Generate Markdown documentation
markdown, err := integrator.GenerateMarkdownForAllTools(ctx)

// Generate for specific category
categoryDocs, err := integrator.GenerateDocsForCategory(ctx, "file")
```

### Individual Tool Documentation

```go
// Convert tool info to documentation
toolInfo := discovery.ListTools()[0]
doc, err := docs.GenerateToolDocumentation(toolInfo)

// Generate OpenAPI operation
operation, err := docs.ConvertToolInfoToOpenAPIOperation(toolInfo)
```

### Batch Operations with Filtering

```go
options := docs.BatchGenerationOptions{
    Categories:      []string{"file", "web"},
    Tags:            []string{"api", "http"},
    IncludeExamples: true,
    IncludeSchemas:  true,
    GroupByCategory: true,
    OutputFormat:    "json",
}

result, err := integrator.BatchGenerate(ctx, options)
```

## Documentation Formats

### OpenAPI 3.0

The generated OpenAPI specification includes:
- Complete paths for all tool endpoints
- Request/response schemas
- Parameter documentation
- Examples and constraints
- Security schemes (if applicable)

Example structure:
```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "Go-LLMs Tool Documentation",
    "version": "1.0.0"
  },
  "paths": {
    "/tools/{toolName}/execute": {
      "post": {
        "summary": "Execute tool",
        "requestBody": {...},
        "responses": {...}
      }
    }
  }
}
```

### Markdown Documentation

Generated Markdown includes:
- Table of contents
- Category grouping
- Detailed descriptions
- Schema documentation
- Usage examples
- Metadata tables

### JSON Documentation

Structured format for programmatic access:
```json
[
  {
    "name": "file_read",
    "description": "Read file contents",
    "category": "file",
    "schemas": {
      "input": {...},
      "output": {...}
    },
    "examples": [...]
  }
]
```

## Bridge Integration

All documentation types are designed for go-llmspell bridge compatibility:

1. **JSON Serializable** - All types can be marshaled to JSON
2. **Schema Conversion** - Tool schemas convert to bridge format
3. **Example Preservation** - Tool examples included in documentation
4. **Metadata Enhancement** - Additional metadata for bridge consumption

## Implementation Details

### Schema Conversion

The system converts between different schema formats:
- `ToolInfo.ParameterSchema` (json.RawMessage) → `Schema` type
- Preserves validation constraints
- Handles nested schemas
- Supports all JSON Schema features

### Pattern-Based Generation

OpenAPI generation uses patterns:
- Tools exposed as `/tools/{name}/execute` endpoints
- Consistent request/response structure
- Error handling with standard codes

### Performance Considerations

- Lazy loading of tool metadata
- Efficient batch operations
- Concurrent generation support
- Minimal memory allocation

## Example

See `cmd/examples/docs-generation/` for a complete working example that:
- Discovers all 33 built-in tools
- Generates OpenAPI specification (142KB)
- Creates Markdown documentation (207KB)
- Demonstrates filtering and batch operations

## Future Enhancements

- Interactive API explorer
- GraphQL schema generation
- Postman collection export
- AsyncAPI for event-driven tools
- Multi-language documentation

## Related Documentation

- [Tool Discovery API](tool-discovery-api.md) - Tool discovery system
- [Schema System](schema-system.md) - Schema generation and validation
- [Testing Framework](testing.md) - Testing the documentation system