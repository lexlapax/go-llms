# Dynamic Script-Based Tool Registration Example

This example demonstrates the enhanced tool discovery system with dynamic script-based tool registration, which is **critical for go-llmspell integration** and other scripting engines.

## Overview

The v0.3.5.3 Enhanced Tool Discovery System enables:

- **Dynamic tool registration** at runtime
- **Script-based tools** for multiple programming languages
- **Namespace isolation** for multi-tenant scenarios
- **Tool versioning** support
- **Registry persistence** for plugin architectures

## Features Demonstrated

### 1. Script Handler Registration
```go
handler := &SimpleScriptHandler{}
err := tools.RegisterScriptHandler(handler)
```

### 2. Dynamic Tool Creation
```go
toolDef := tools.ScriptToolDefinition{
    Name:        "calculator",
    Description: "A simple calculator",
    Engine:      tools.ScriptEngineJavaScript,
    Script:      "add",
    // ... schemas and examples
}

err = tools.RegisterScriptToolWithDiscovery(toolDef)
```

### 3. Tool Discovery and Execution
```go
discovery := tools.NewDiscovery()
tool, err := discovery.CreateTool("calculator")
result, err := tool.Execute(ctx, params)
```

### 4. Namespace Isolation
```go
err = discovery.CreateNamespace("experimental")
err = discovery.SwitchNamespace("experimental")
```

## go-llmspell Integration

This system enables go-llmspell to:

1. **Register script handlers** for JavaScript, Lua, Tengo, etc.
2. **Create tools dynamically** from script definitions
3. **Isolate tools by tenant** using namespaces
4. **Persist tool registrations** for plugin architectures
5. **Version tools** for compatibility management

## Running the Example

```bash
cd cmd/examples/tools-script-dynamic
go run main.go
```

## Expected Output

```
=== Dynamic Script-Based Tool Registration Example ===
✓ Registered JavaScript script handler
✓ Registered calculator tool
✓ Registered greeting tool
✓ Registered factorial tool

=== Tool Discovery ===
Found 3+ registered tools:
- calculator: A simple calculator that can add two numbers (Category: math, Version: 1.0.0)
- greeter: A friendly greeting tool (Category: utility, Version: 1.0.0)
- factorial: Calculate factorial of a number (Category: math, Version: 1.0.0)

=== Tool Execution ===
Calculator: 15 + 27 = 42
Greeter: Hello, Go Developer!
Factorial: 6! = 720

=== Tool Metadata ===
Calculator Schema:
{
  "name": "calculator",
  "description": "A simple calculator that can add two numbers",
  "parameters": { ... },
  "examples": [ ... ]
}

=== Namespace Isolation ===
Current namespace: experimental
Tools in experimental namespace: 0
Tools in default namespace: 3+

=== Example Complete ===
```

## Key Components

### ScriptHandler Interface
- `Execute()` - Run scripts with context and parameters
- `Validate()` - Check script syntax and validity  
- `Engine()` - Return supported script engine type
- `SupportsFeature()` - Check for specific capabilities

### ScriptToolDefinition
- Complete tool metadata including schemas
- Script engine specification
- Examples and constraints
- Error guidance for LLMs

### Enhanced ToolDiscovery
- `RegisterTool()` - Dynamic tool registration
- `UnregisterTool()` - Runtime tool removal
- `CreateNamespace()` - Multi-tenant isolation
- `SaveRegistry()`/`LoadRegistry()` - Persistence

## Bridge Integration Pattern

For go-llmspell and other scripting bridges:

```go
// 1. Register script engine handler
jsHandler := NewJavaScriptHandler()
tools.RegisterScriptHandler(jsHandler)

// 2. Create tool from script definition
toolDef := tools.ScriptToolDefinition{
    Engine: tools.ScriptEngineJavaScript,
    Script: "function calculate(params) { return params.a + params.b; }",
    // ... metadata
}

// 3. Register with discovery system
tools.RegisterScriptToolWithDiscovery(toolDef)

// 4. Tools are now available for LLM agent use
```

This enables scripting engines to provide dynamic, runtime-extensible tool capabilities to LLM agents!