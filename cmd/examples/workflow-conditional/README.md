# Conditional Workflow Example

This example demonstrates how to use conditional workflow agents to implement branching logic in workflows.

## Overview

The conditional workflow agent executes different branches based on state conditions, similar to if/else logic in programming. This is useful for:

- Routing data to specialized processors based on type
- Implementing priority-based handling systems
- Creating decision trees for complex business logic
- Handling multiple conditions with custom priority ordering

## Features Demonstrated

1. **Basic Conditional Logic**: Simple if/else branching based on state values
2. **Priority-Based Evaluation**: Conditions evaluated in priority order (highest first)
3. **Multiple Matching**: Allow multiple branches to execute for comprehensive processing
4. **Default Branches**: Fallback processing when no conditions match
5. **Complex Conditions**: Custom condition functions with access to full state

## Usage

```bash
# Run with mock agents
go run main.go

# Run with actual LLM providers (requires API keys)
export ANTHROPIC_API_KEY=your-key
export OPENAI_API_KEY=your-key
go run main.go
```

## Examples

### 1. Basic Conditional Processing
Routes different data types to specialized processors:
- Text data → Text processor (Claude)
- Image data → Image processor (GPT-4 Vision)
- Structured data → Data processor (GPT-4)
- Unknown data → Generic processor (fallback)

### 2. Priority-Based Issue Triage
Handles issues based on severity with priority ordering:
- Severity 9-10 → Critical handler (Priority 100)
- Severity 7-8 → High priority handler (Priority 75)
- Severity 4-6 → Medium priority handler (Priority 50)
- Severity 1-3 → Low priority handler (Priority 25)
- Severity 0 → No handler (demonstrates no match scenario)

### 3. Multiple Match Validation
Runs multiple validation checks simultaneously:
- Syntax validation
- Security scanning
- Performance analysis
- Compatibility checking

All applicable checks run based on the required_checks configuration.

## Code Structure

### Basic Conditional Workflow
```go
// Create conditional workflow
workflow := workflow.NewConditionalAgent("processor").
    AddAgent("text", func(state *domain.State) bool {
        dataType, _ := state.Get("data_type")
        return dataType == "text"
    }, textProcessor).
    AddAgent("image", func(state *domain.State) bool {
        dataType, _ := state.Get("data_type")
        return dataType == "image"
    }, imageProcessor).
    SetDefaultAgent(genericProcessor)

// Run with state
result, err := workflow.Run(ctx, initialState)
```

### Priority-Based Conditions
```go
workflow := workflow.NewConditionalAgent("triage").
    AddBranchWithPriority("critical", criticalCondition, criticalStep, 100).
    AddBranchWithPriority("high", highCondition, highStep, 75).
    AddBranchWithPriority("medium", mediumCondition, mediumStep, 50)
```

### Multiple Matches
```go
workflow := workflow.NewConditionalAgent("validation").
    WithAllowMultipleMatches(true).
    WithEvaluateAllConditions(true).
    AddAgent("syntax", syntaxCondition, syntaxValidator).
    AddAgent("security", securityCondition, securityValidator)
```

## Configuration Options

- **AddAgent(name, condition, agent)**: Add a branch with agent
- **AddBranch(name, condition, step)**: Add a branch with custom step  
- **AddBranchWithPriority(name, condition, step, priority)**: Add prioritized branch
- **SetDefaultAgent(agent)**: Set fallback agent
- **SetDefaultBranch(step)**: Set fallback step
- **WithAllowMultipleMatches(bool)**: Allow multiple branches to execute
- **WithEvaluateAllConditions(bool)**: Evaluate all conditions even after finding matches
- **WithHook(hook)**: Add monitoring hooks

## Condition Functions

Condition functions receive the current state and return a boolean:

```go
func textCondition(state *domain.State) bool {
    if dataType, exists := state.Get("data_type"); exists {
        return dataType == "text"
    }
    return false
}

func severityCondition(minSeverity int) func(*domain.State) bool {
    return func(state *domain.State) bool {
        if severity, exists := state.Get("severity"); exists {
            return severity.(int) >= minSeverity
        }
        return false
    }
}
```

## Execution Flow

1. **Condition Evaluation**: Conditions evaluated in priority order (highest first)
2. **Branch Execution**: Matching branches execute based on configuration
3. **State Flow**: Each branch receives and can modify the workflow state
4. **Result Merging**: Final state contains results from executed branches
5. **Default Handling**: Default branch executes if no conditions match

## Error Handling

- **Branch Errors**: Individual branch failures can be handled or cause workflow failure
- **Validation**: All branches and conditions validated before execution
- **Graceful Degradation**: Default branches provide fallback behavior

## Metadata

The workflow adds execution metadata including:
- `executed_branches`: List of branches that executed
- `total_branches`: Total number of defined branches  
- `has_default`: Whether a default branch is configured

## Next Steps

See the loop workflow example for iterative processing, or combine conditional and parallel workflows for complex decision trees.