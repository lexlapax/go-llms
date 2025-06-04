# Agent Structured Output Example

This example demonstrates how to use LLM agents with structured output validation using schemas. It shows how to get type-safe, validated responses from LLMs by defining Go structs and automatically generating JSON schemas.

## Overview

The Agent Structured Output example showcases:

1. **Schema-driven LLM interactions** - Define Go structs and get validated JSON responses
2. **Type-safe processing** - Direct conversion from LLM responses to Go structs
3. **Complex data structures** - Nested objects, arrays, enums, and validation constraints
4. **Real-world use cases** - Task management, project analysis, and meeting notes
5. **Fallback patterns** - Graceful handling when structured processing isn't available

## Key Features

### Structured Data Types

The example defines several realistic business domain types:

- **Task** - Project task with status, priority, and metadata
- **ProjectAnalysis** - Comprehensive project metrics and insights
- **MeetingNotes** - Structured meeting documentation with action items
- **ActionItem** - Nested object for meeting action items

### Advanced Schema Features

- **Enums** - Status and priority levels with validation
- **Nested Objects** - ActionItems within MeetingNotes
- **Arrays** - Lists of strings and complex objects
- **Optional Fields** - Pointer types for optional data
- **Validation Constraints** - Required fields, min/max values, patterns
- **Time Handling** - Automatic time.Time formatting

## Running the Example

```bash
# With an API key (recommended for full functionality)
export OPENAI_API_KEY=your_key_here
# OR
export ANTHROPIC_API_KEY=your_key_here
# OR  
export GEMINI_API_KEY=your_key_here

# Build and run
go build -o agent-structured-output
./agent-structured-output

# Or use make
make example EXAMPLE=agent-structured-output
./bin/agent-structured-output
```

## Example Outputs

### 1. Structured Task Generation

Input: "Create a realistic software development task for implementing a user authentication system"

Output:
```json
{
  "id": "task-auth-001",
  "title": "Implement User Authentication System",
  "description": "Develop secure user authentication with OAuth2 integration",
  "status": "pending",
  "priority": "high",
  "estimated_hours": 16,
  "tags": ["authentication", "security", "oauth", "backend"],
  "created_at": "2025-02-03T10:30:00Z"
}
```

### 2. Project Analysis

Input: Project data with multiple tasks

Output:
```json
{
  "project_name": "E-commerce Platform Redesign",
  "total_tasks": 8,
  "completed_tasks": 2,
  "pending_tasks": 4,
  "in_progress_tasks": 2,
  "completion_rate": 25,
  "recommendations": [
    "Prioritize security audit due to high risk",
    "Allocate additional resources to mobile optimization",
    "Consider parallel development for payment integration"
  ],
  "risks": [
    "Security vulnerabilities in pending audit",
    "Mobile optimization scope creep",
    "Database performance issues"
  ],
  "next_actions": [
    "Complete security audit immediately",
    "Begin mobile optimization planning",
    "Schedule database performance review"
  ]
}
```

### 3. Meeting Notes Extraction

Input: Meeting transcript

Output:
```json
{
  "meeting_title": "Sprint Planning Session",
  "date": "2025-02-03T14:00:00Z",
  "attendees": ["Sarah", "Mike", "Jenny", "Alex"],
  "duration": 90,
  "key_discussions": [
    "Sprint performance review",
    "Authentication system priority",
    "Technical debt allocation"
  ],
  "decisions": [
    "Authentication system is top priority",
    "20% capacity allocated to technical debt",
    "UI redesign approved"
  ],
  "action_items": [
    {
      "description": "Update user stories for authentication",
      "assigned_to": "Sarah",
      "due_date": "2025-02-07T23:59:59Z",
      "priority": "high",
      "status": "pending"
    }
  ]
}
```

## Schema Validation

The example demonstrates automatic validation:

```go
// Define your domain model
type Task struct {
    ID          string     `json:"id" validate:"required" description:"Unique task identifier"`
    Status      TaskStatus `json:"status" validate:"required,oneof=pending in_progress completed cancelled"`
    Priority    Priority   `json:"priority" validate:"required,oneof=low medium high urgent"`
    EstimatedHours float64 `json:"estimated_hours" validate:"min=0"`
}

// Generate schema automatically
schema, _ := reflection.GenerateSchema(Task{})

// Process LLM response with validation
var task Task
err := processor.ProcessTypedWithSchema(ctx, prompt, schema, &task)
```

## Validation Features

The schemas include comprehensive validation:

- **Required Fields** - `validate:"required"`
- **Enum Values** - `validate:"oneof=value1 value2 value3"`
- **Numeric Ranges** - `validate:"min=0,max=100"`
- **String Patterns** - `pattern:"^[A-Z]{2}$"`
- **Array Constraints** - `validate:"min=1"` for non-empty arrays

## Integration Patterns

### With Agents

```go
// Create agent
agent, _ := core.NewAgentFromString("analyzer", "gpt-4")

// Generate schema from Go struct
schema, _ := reflection.GenerateSchema(ProjectAnalysis{})

// Use structured processor
processor := processor.NewProcessor(llmProvider)
var analysis ProjectAnalysis
err := processor.ProcessTypedWithSchema(ctx, prompt, schema, &analysis)
```

### With Custom Validation

```go
// Add custom validation rules
type CustomTask struct {
    Task
    CustomField string `json:"custom_field" validate:"required,custom_rule"`
}

// Register custom validator
validator.RegisterValidation("custom_rule", customValidationFunc)
```

## Error Handling

The example includes comprehensive error handling:

- **Schema generation errors** - Invalid struct definitions
- **Validation errors** - LLM responses that don't match schema
- **Processing errors** - JSON parsing and type conversion issues
- **Fallback patterns** - Graceful degradation for mock providers

## Performance Considerations

1. **Schema Caching** - Generate schemas once and reuse
2. **Processor Reuse** - Create processor instances once
3. **Validation Optimization** - Use appropriate validation levels
4. **Memory Management** - Consider large response handling

## Best Practices

### Schema Design

1. **Use descriptive field names** and consistent JSON conventions
2. **Add meaningful descriptions** for LLM context
3. **Mark required fields** explicitly
4. **Use appropriate validation constraints**
5. **Define enums** for controlled vocabularies

### Error Recovery

1. **Implement fallback patterns** for when structured processing fails
2. **Validate responses** before using in business logic
3. **Log validation errors** for debugging
4. **Provide meaningful error messages** to users

### Type Safety

1. **Use Go's type system** effectively with struct tags
2. **Leverage pointer types** for optional complex fields
3. **Define custom types** for domain-specific enums
4. **Implement proper JSON marshaling** for complex types

## Use Cases

This pattern is ideal for:

- **Data extraction** from unstructured text
- **Report generation** with consistent formats
- **API responses** that need validation
- **Configuration management** with type safety
- **Workflow automation** with structured inputs/outputs
- **Documentation generation** from conversations
- **Analysis and insights** with standardized metrics

The structured output approach ensures reliability, type safety, and consistency when working with LLM-generated content in production applications.