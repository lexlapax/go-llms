# Documentation Generation Integration Example

This example demonstrates the integration between the go-llms documentation generation system and the tool discovery system.

## What This Example Shows

1. **Tool Discovery Integration**: How to use the tool discovery system to find all available tools
2. **OpenAPI Generation**: Generate OpenAPI 3.0 specifications for all discovered tools
3. **Markdown Documentation**: Create human-readable Markdown documentation
4. **JSON Documentation**: Export structured JSON documentation
5. **Category-based Generation**: Generate docs for specific tool categories
6. **Search-based Generation**: Generate docs for tools matching search queries
7. **Enhanced Tool Help**: Integration with existing `GetToolHelp` functionality
8. **Batch Operations**: Advanced batch generation with filtering options
9. **Individual Tool Conversion**: Convert single tools to documentation format

## Features Demonstrated

### Core Integration
- Tool discovery system integration
- Automatic schema conversion from tool metadata
- Example handling and conversion
- Bridge-friendly output formats

### Output Formats
- **OpenAPI 3.0**: Complete API specification with endpoints for tool execution
- **Markdown**: Human-readable documentation with tables and code examples
- **JSON**: Structured data perfect for programmatic consumption

### Advanced Features
- Category-based filtering
- Tag-based filtering
- Search query filtering
- Batch processing with options
- Enhanced help text generation

## Running the Example

```bash
go run main.go
```

## Generated Files

After running the example, you'll find these files:

1. **tools-openapi.json**: Complete OpenAPI 3.0 specification
   - Defines REST endpoints for each tool
   - Includes request/response schemas
   - Tool categorization via tags
   - Example requests and responses

2. **tools-documentation.md**: Markdown documentation
   - Human-readable tool descriptions
   - Usage instructions and examples
   - Schema documentation
   - Categorized organization

3. **tools-batch-docs.json**: Batch-generated JSON documentation
   - Structured tool metadata
   - Complete schema information
   - Examples and usage hints
   - Perfect for programmatic consumption

## Bridge Integration

This example demonstrates the go-llmspell bridge requirements:

- **JSON Serialization**: All types are fully JSON serializable
- **Schema Compatibility**: Tool schemas are converted to documentation schemas
- **Metadata Preservation**: All tool metadata is preserved and enhanced
- **Discovery Integration**: Seamless integration with existing tool discovery

## Key Components Used

- `pkg/docs/tools.go`: Tool-specific documentation generation
- `pkg/docs/integration.go`: Discovery system integration
- `pkg/agent/tools/discovery.go`: Tool discovery system
- Standard documentation generators

## Integration Points

The example shows how the documentation system integrates with:

1. **Tool Discovery**: Automatic discovery of all available tools
2. **Schema System**: Conversion between tool schemas and doc schemas  
3. **Example System**: Handling of tool examples in documentation
4. **Help System**: Enhancement of existing `GetToolHelp` functionality

This integration makes it easy to generate comprehensive documentation for any go-llms application with minimal setup.