# Go-LLMs v0.3.1 Release Notes

**Release Date**: June 10, 2025

## Overview

Go-LLMs v0.3.1 is a major milestone release that completes the Tool System Enhancement initiative, bringing comprehensive improvements to all 32 built-in tools with the new ToolBuilder pattern. This release provides enhanced LLM integration, better error handling, and full Model Context Protocol (MCP) compatibility.

## Key Highlights

### ✨ ToolBuilder Pattern Migration Complete

All 32 built-in tools have been migrated to the enhanced ToolBuilder pattern, providing:

- **Rich Metadata**: Comprehensive usage instructions, constraints, and examples for each tool
- **LLM-Optimized**: Enhanced error messages and guidance specifically designed for LLM consumption
- **MCP Compatible**: Full support for Model Context Protocol export
- **Improved Discovery**: Better categorization and search capabilities

### 🛠️ Tool Categories Enhanced

1. **Web Tools** (6 tools) - Enhanced with authentication support:
   - `api_client` v3.0.0: GraphQL support, OpenAPI discovery, automatic auth detection
   - `web_search` v2.0.0: Multi-engine support with API key management
   - `web_fetch`, `web_scrape`, `http_request`: Improved error handling

2. **File Tools** (6 tools) - Safety and performance improvements:
   - Large file support with streaming
   - Atomic write operations
   - Enhanced search with regex patterns
   - Binary file detection

3. **System Tools** (4 tools) - Better security and control:
   - Safe command execution with timeouts
   - Environment variable pattern matching
   - Comprehensive system information
   - Process filtering and monitoring

4. **Data Tools** (4 tools) - Advanced processing capabilities:
   - JSONPath query support
   - CSV statistics and filtering
   - XML to JSON conversion
   - Functional transformations (map, filter, reduce)

5. **DateTime Tools** (7 tools) - Comprehensive time operations:
   - Natural language date parsing
   - Business day calculations
   - Multi-timezone support
   - Localized formatting

6. **Feed Tools** (6 tools) - Complete feed processing:
   - Multi-format support (RSS, Atom, JSON Feed)
   - Feed discovery and aggregation
   - Content filtering and extraction
   - Format conversion

### 📚 Documentation Improvements

- **Tool Development Guide**: Comprehensive guide for creating custom tools
- **Built-in Tools Guide**: Detailed documentation for all 32 tools
- **Examples Gallery**: Organized showcase of all example use cases
- **Enhanced READMEs**: Updated documentation across all examples

### 🚀 Performance

- **Test Coverage**: 44.3% overall coverage with 280+ tests passing
- **Benchmark Results**: 
  - API Client: ~115μs for simple GET requests
  - Tool Execution: ~6.3μs per tool call
  - State Operations: ~67ns for get/set operations
  - Workflow Execution: ~22μs for sequential workflows

## Breaking Changes

None. This release maintains full backward compatibility with v0.3.0.

## New Features

### API Client Tool (v3.0.0)

- **GraphQL Support**: Full query/mutation execution with variables
- **Schema Introspection**: Automatic GraphQL schema discovery
- **OpenAPI Integration**: Automatic server URL resolution
- **Enhanced Authentication**: Automatic credential detection from state
- **Improved Error Guidance**: Context-aware error messages

### Enhanced Tool Metadata

All tools now include:
- 3-7 usage examples with input/output
- Comprehensive constraints documentation
- Error guidance mapping
- Resource usage indicators
- MCP export capability

### Calculator Tool (v2.0.0)

- Extended mathematical constants (phi, tau, sqrt variants)
- Enhanced LLM integration mode as default
- Provider/model information display
- DEBUG environment variable support

## Improvements

### Error Handling
- Context-aware error messages across all tools
- LLM-friendly guidance for common errors
- Validation with actionable suggestions

### Authentication
- Unified authentication middleware for web tools
- Support for Bearer, API Key, Basic, OAuth2, Custom headers
- Automatic detection based on URL patterns
- Secure credential storage in agent state

### Documentation
- Created comprehensive tool development guide
- Updated built-in tools documentation
- Added examples gallery with 40+ examples
- Improved cross-linking and navigation

## Bug Fixes

- Fixed hardcoded URL detection in authentication
- Resolved token detection patterns
- Fixed linting issues across feed tools
- Corrected example patterns in agent-calculator

## Migration Guide

No migration required for v0.3.0 users. The ToolBuilder pattern is backward compatible.

For users upgrading from earlier versions, see the [Migration Guide](docs/MIGRATION_GUIDE_PHASE5.md).

## Examples

### Using the Enhanced API Client

```go
// Automatic authentication from state
state.Set("github_token", os.Getenv("GITHUB_TOKEN"))

// GraphQL query with automatic auth
result, _ := agent.ExecuteTool("api_client", map[string]interface{}{
    "base_url": "https://api.github.com",
    "endpoint": "/graphql",
    "graphql_query": `query { viewer { login name } }`,
})
```

### Using Enhanced Tools with LLM

```go
// All tools now have rich metadata for LLMs
agent := core.NewLLMAgent("assistant", "Assistant", deps)
agent.AddTool(tools.MustGetTool("web_search"))
agent.AddTool(tools.MustGetTool("file_read"))
agent.AddTool(tools.MustGetTool("data_transform"))

// LLM can now better understand tool usage
result, _ := agent.Run(ctx, state)
```

## What's Next

- **Phase 5**: Advanced API Client capabilities (pagination, rate limiting, caching)
- **Phase 6**: Model Context Protocol support for agents
- **Built-in Agents**: Text, Research, Coding, Data agents
- **Multi-Agent Workflows**: Pipeline, MapReduce, Consensus patterns

## Contributors

This release represents a significant effort in enhancing the tool system. Special thanks to all contributors and users who provided feedback during the development process.

## Installation

```bash
go get github.com/lexlapax/go-llms@v0.3.1
```

## Requirements

- Go 1.23.0 or higher
- No additional dependencies required

## License

MIT License - see LICENSE file for details