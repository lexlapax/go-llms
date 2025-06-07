# Tool Enhancement Implementation Plan

## Overview

This document outlines the comprehensive plan to enhance the tool system with better LLM guidance, removing backward compatibility requirements and implementing a completely new tool interface that provides rich metadata and usage instructions for LLMs.

## Goals

1. Create a new Tool interface with comprehensive LLM guidance
2. Migrate all 32 built-in tools to the new interface
3. Update all 31 examples to use the enhanced tools
4. Completely overhaul tool documentation
5. Prepare for MCP (Model Context Protocol) compatibility

## Phase 1: Core Infrastructure Changes

### 1.1 Create New Tool Interface (TDD First)

**Location**: `pkg/agent/domain/tool.go` (new file)

**Test file first**: `pkg/agent/domain/tool_test.go`

Test cases:
- Tool with minimal implementation
- Tool with full LLM guidance
- Schema validation for parameters and output
- Example validation
- MCP export functionality

**New interface structure**:
```go
type Tool interface {
    // Core functionality
    Name() string
    Description() string
    Execute(ctx *ToolContext, params interface{}) (interface{}, error)
    
    // Schema definitions
    ParameterSchema() *schema.Schema
    OutputSchema() *schema.Schema
    
    // LLM guidance
    UsageInstructions() string
    Examples() []ToolExample
    Constraints() []string
    ErrorGuidance() map[string]string // error type -> guidance
    
    // Metadata
    Category() string
    Tags() []string
    Version() string
    
    // Behavioral hints
    IsDeterministic() bool
    IsDestructive() bool
    RequiresConfirmation() bool
    EstimatedLatency() string // "fast", "medium", "slow"
    
    // MCP compatibility
    ToMCPDefinition() MCPToolDefinition
}

type ToolExample struct {
    Name        string
    Description string
    Scenario    string // When to use this example
    Input       interface{}
    Output      interface{}
    Explanation string // Why this works
}

type MCPToolDefinition struct {
    Name         string                 `json:"name"`
    Description  string                 `json:"description"`
    InputSchema  interface{}            `json:"inputSchema"`
    OutputSchema interface{}            `json:"outputSchema,omitempty"`
    Annotations  map[string]interface{} `json:"annotations,omitempty"`
}
```

**Implementation steps**:
1. Write comprehensive tests for the new interface
2. Create interface definition
3. Run `go fmt ./pkg/agent/domain/`
4. Run `golangci-lint run ./pkg/agent/domain/`
5. Run `go test ./pkg/agent/domain/`

### 1.2 Create Base Tool Implementation

**Location**: `pkg/agent/tools/base_tool_v2.go`

Provides default implementations and helpers with builder pattern for easier construction.

**Implementation steps**:
1. Write tests for BaseToolV2 and builder
2. Implement BaseToolV2 with all interface methods
3. Create builder pattern for easy tool construction
4. Add validation methods
5. Run fmt, lint, test

### 1.3 Update Tool Registry

**Location**: `pkg/agent/builtins/tools/registry.go`

Enhanced metadata structure and registry updates to support new tool interface, including MCP export functionality.

**Implementation steps**:
1. Write tests for new registry methods
2. Implement registry enhancements
3. Add MCP export functionality
4. Run fmt, lint, test

### 1.4 Update LLM Agent Tool Description

**Location**: `pkg/agent/core/llm_agent.go`

Enhanced system content generation that automatically creates comprehensive tool documentation from metadata.

**Implementation steps**:
1. Write tests for new formatting functions
2. Implement enhanced system content generation
3. Add helper methods for schema formatting
4. Run fmt, lint, test

## Phase 2: Tool Migration Plan

### 2.1 Migration Order (by complexity)

Total: 32 tools across 7 categories

1. **Math Tools** (1 tool)
   - `calculator` - Most complex, good test case

2. **System Tools** (4 tools)
   - `execute_command` - High risk, needs careful guidance
   - `get_environment_variable` - Simple, good for testing
   - `get_system_info` - Read-only, safe
   - `process_list` - Medium complexity

3. **File Tools** (6 tools)
   - `file_read` - Common use, needs good examples
   - `file_write` - Destructive, needs constraints
   - `file_list` - Complex parameters
   - `file_delete` - High risk, needs confirmation
   - `file_move` - Destructive, needs guidance
   - `file_search` - Complex, needs examples

4. **Web Tools** (4 tools)
   - `web_search` - Multiple engines, complex
   - `web_fetch` - External dependency
   - `web_scrape` - Complex selectors
   - `http_request` - Very flexible, needs guidance

5. **Data Tools** (4 tools)
   - `json_process` - Complex operations
   - `csv_process` - Multiple transforms
   - `xml_process` - XPath complexity
   - `data_transform` - Generic operations

6. **DateTime Tools** (7 tools)
   - `datetime_now` - Simple baseline
   - `datetime_info` - Information extraction
   - `datetime_calculate` - Complex operations
   - `datetime_parse` - Format complexity
   - `datetime_format` - Output formatting
   - `datetime_convert` - Timezone handling
   - `datetime_compare` - Comparison logic

7. **Feed Tools** (6 tools)
   - `feed_fetch` - External dependency
   - `feed_discover` - Auto-discovery
   - `feed_filter` - Complex filtering
   - `feed_aggregate` - Multiple sources
   - `feed_convert` - Format conversion
   - `feed_extract` - Content extraction

### 2.2 Calculator Tool Migration Example

The calculator tool will serve as the template for migrating all other tools. It will include:

- Comprehensive usage instructions
- Multiple examples covering different scenarios
- Clear constraints and limitations
- Error guidance for common mistakes
- Full parameter and output schemas
- Behavioral metadata (deterministic, non-destructive, fast)

## Phase 3: Example Refactoring Plan

### 3.1 Examples to Update (31 examples total)

**Order of refactoring**:
1. `agent-calculator` - Remove custom system prompt
2. `agent-simple-llm` - Update tool usage
3. `agent-llm-builtin-tools` - Showcase new guidance
4. `agent-tools-conversion` - Update conversion logic
5. `builtins-data-tools` - Update for new interface
6. `builtins-datetime-tools` - Update for new interface
7. `builtins-feed-tools` - Update for new interface
8. `builtins-file-tools` - Update for new interface
9. `builtins-system-tools` - Update for new interface
10. `builtins-web-tools` - Update for new interface
11. `builtins-web-search-parallel` - Update for new interface
12. All remaining agent examples using tools

### 3.2 Example Refactoring Pattern

**Before**:
```go
agent.SetSystemPrompt(`You are a helpful math assistant with access to a calculator tool.

When asked to perform calculations, use the calculator tool by responding with:
{"tool": "calculator", "params": {"operation": "...", "operand1": ..., "operand2": ...}}

The calculator supports:
[... 200 lines of manual tool documentation ...]
`)
```

**After**:
```go
agent.SetSystemPrompt(`You are a helpful math assistant.`)
// Tool documentation now auto-generated from tool metadata
```

## Phase 4: Documentation Overhaul

### 4.1 Technical Documentation Updates

**New file**: `docs/technical/tools.md`

Structure:
1. Tool System Architecture
2. Tool Interface Reference
3. Creating Custom Tools
4. Tool Metadata Best Practices
5. LLM Guidance Patterns
6. MCP Compatibility
7. Testing Tools
8. Performance Considerations

### 4.2 User Guide Updates

**New file**: `docs/user-guide/tool-development.md`

Structure:
1. Quick Start Guide
2. Tool Examples Gallery
3. Common Patterns
4. Troubleshooting
5. Migration from v1 Tools

**Update**: `docs/user-guide/builtin-tools.md`
- Update all tool documentation to reflect new interface
- Add guidance on how tools self-document
- Include MCP export examples

## Phase 5: Implementation Schedule

### Week 1: Core Infrastructure
- Day 1: New Tool interface + tests
- Day 2: BaseToolV2 implementation
- Day 3: Registry updates
- Day 4: LLM Agent enhancements
- Day 5: Integration testing

### Week 2: Tool Migration (Part 1)
- Day 1: Calculator tool (template for others)
- Day 2: System tools (4)
- Day 3: File tools (6)
- Day 4: Web tools (4)
- Day 5: Testing & fixes

### Week 3: Tool Migration (Part 2)
- Day 1: Data tools (4)
- Day 2: DateTime tools (7)
- Day 3: Feed tools (6)
- Day 4: Example updates (first half)
- Day 5: Example updates (second half)

### Week 4: Documentation & Polish
- Day 1-2: Technical documentation
- Day 3-4: User guide documentation
- Day 5: Final testing & release prep

## Testing Strategy

1. **Unit Tests**: Each tool has comprehensive tests
2. **Integration Tests**: Tool + LLM Agent interaction
3. **Example Tests**: All examples must pass
4. **MCP Export Tests**: Verify MCP compatibility
5. **Performance Tests**: No regression in tool execution

## Success Criteria

1. All 32 tools provide comprehensive LLM guidance
2. No custom system prompts needed for tool usage
3. MCP export generates valid definitions
4. All 31 examples work without modification
5. Documentation is complete and accurate
6. No performance regression
7. All tests pass, code is formatted and linted

## Migration Checklist

### For each tool:
- [ ] Write tests for new interface implementation
- [ ] Implement new Tool interface
- [ ] Add comprehensive usage instructions
- [ ] Create 3-5 examples with scenarios
- [ ] Define constraints and limitations
- [ ] Add error guidance
- [ ] Define output schema
- [ ] Set behavioral metadata
- [ ] Test MCP export
- [ ] Run fmt, lint, vet
- [ ] Update any examples using the tool
- [ ] Update documentation

### Global tasks:
- [ ] Create new Tool interface
- [ ] Implement BaseToolV2
- [ ] Update registry
- [ ] Enhance LLM Agent
- [ ] Update all examples
- [ ] Write technical documentation
- [ ] Write user guide
- [ ] Performance testing
- [ ] Release preparation