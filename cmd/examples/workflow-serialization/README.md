# Workflow Serialization Example

This example demonstrates workflow serialization capabilities for bridge layer integration, including JSON/YAML serialization, script-based steps, and template usage.

## Overview

The workflow serialization system provides:
- **WorkflowSerializer** for converting workflows to/from JSON and YAML
- **ScriptStep** for embedding scripts in workflows
- **ScriptHandler** registry for multiple scripting languages
- **Templates** for reusable workflow patterns
- **Bridge integration** via map-based deserialization

## Running the Example

```bash
go run main.go
```

## Features Demonstrated

### 1. Script-Based Workflows
Create workflows with embedded scripts:
```go
validateStep, err := workflow.NewScriptStepBuilder("validate").
    WithLanguage("expr").
    WithScript("validated = true").
    WithDescription("Validate input data").
    WithTimeout(5 * time.Second).
    Build()
```

### 2. Serialization Formats
Serialize workflows to different formats:
```go
// JSON
jsonData, err := workflow.SerializeWorkflow(wf, "json")

// Pretty JSON
prettyJSON, err := workflow.SerializeWorkflow(wf, "json-pretty")

// YAML
yamlData, err := workflow.SerializeWorkflow(wf, "yaml")
```

### 3. Bridge Layer Integration
Deserialize workflows from map format (for scripting engines):
```go
bridgeData := map[string]interface{}{
    "name": "Bridge Workflow",
    "steps": []interface{}{
        map[string]interface{}{
            "type": "script",
            "script": map[string]interface{}{
                "language": "javascript",
                "source": "return state",
            },
        },
    },
}

wf, err := workflow.DeserializeDefinition(bridgeData)
```

### 4. Workflow Templates
Use pre-built workflow templates:
```go
// Apply a template with variables
variables := map[string]interface{}{
    "input_source": "data.csv",
}
wf, err := workflow.ApplyTemplate("data-processing", variables)
```

### 5. Custom Script Handlers
Register custom scripting languages:
```go
customHandler := &workflow.ScriptHandlerFunc{
    LanguageFn: func() string { return "custom" },
    ExecuteFn: func(ctx, state, script, env) (*WorkflowState, error) {
        // Custom execution logic
    },
}

workflow.RegisterScriptHandler("custom", customHandler)
```

## Script Languages

The example includes mock handlers for:
- **javascript** - Mock JavaScript execution
- **expr** - Simple expression evaluation
- **json-transform** - JSON-based transformations

Real implementations would use:
- **javascript** - goja or otto engine
- **lua** - gopher-lua
- **tengo** - tengo scripting language
- **expr** - expr or cel-go

## Serialization Format

### JSON Format
```json
{
  "name": "Example Workflow",
  "description": "Workflow description",
  "version": "1.0",
  "steps": [
    {
      "name": "step1",
      "type": "script",
      "script": {
        "language": "javascript",
        "source": "return state",
        "timeout": "30s"
      }
    }
  ]
}
```

### YAML Format
```yaml
name: Example Workflow
description: Workflow description
version: "1.0"
steps:
  - name: step1
    type: script
    script:
      language: javascript
      source: return state
      timeout: 30s
```

## Use Cases

1. **Scripting Engine Integration**: Serialize workflows for go-llmspell
2. **Workflow Storage**: Save workflows to files or databases
3. **Visual Builders**: Create workflows from UI builders
4. **API Integration**: Send/receive workflows via REST APIs
5. **Configuration Management**: Define workflows in config files

## Key Components

- **WorkflowSerializer**: Interface for serialization formats
- **SerializableWorkflowDefinition**: Bridge-friendly workflow format
- **ScriptStep**: Embeds scripts in workflow steps
- **ScriptHandler**: Executes scripts in specific languages
- **WorkflowTemplate**: Reusable workflow patterns

## Related Examples

- `workflow-sequential`: Basic sequential workflows
- `workflow-parallel`: Parallel workflow execution
- `agent-events`: Event system for workflow monitoring
- `types-bridge`: Type conversion for bridge integration