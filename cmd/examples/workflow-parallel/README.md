# Parallel Workflow Example

This example demonstrates how to use parallel workflow agents to execute multiple agents concurrently.

## Overview

The parallel workflow agent executes multiple agents at the same time, with various options for controlling concurrency and merging results. This is useful for:

- Getting multiple perspectives on the same topic
- Racing agents to get the fastest response
- Distributing work across multiple specialized agents
- Gathering data from multiple sources simultaneously

## Features Demonstrated

1. **Concurrent Execution**: Multiple agents run simultaneously
2. **Concurrency Control**: Limit the number of agents running at once
3. **Merge Strategies**:
   - `MergeAll`: Combine all results into one state
   - `MergeFirst`: Use the first completed result
   - `MergeCustom`: Apply custom logic to merge results
4. **Timeout Support**: Set maximum time for workflow completion
5. **Error Handling**: Configure how to handle partial failures

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

### 1. Parallel Analysis
Analyzes a topic from multiple perspectives simultaneously:
- Technical analyst (Claude)
- Business analyst (GPT-4)
- Ethical analyst (Claude)

All three analyses run concurrently and results are merged.

### 2. Racing Agents
Multiple agents race to provide the fastest response:
- Fast agent (100ms)
- Medium agent (300ms)
- Slow agent (500ms)

With `MergeFirst` strategy, the workflow completes as soon as the first agent finishes.

### 3. Custom Merge
Demonstrates custom merge logic by calculating average scores from multiple scoring agents.

## Code Structure

```go
// Create parallel workflow
workflow := workflow.NewParallelAgent("analysis").
    WithMaxConcurrency(3).
    WithMergeStrategy(workflow.MergeAll).
    AddAgent(agent1).
    AddAgent(agent2).
    AddAgent(agent3)

// Run workflow
result, err := workflow.Run(ctx, initialState)
```

## Configuration Options

- **WithMaxConcurrency(int)**: Limit concurrent executions
- **WithMergeStrategy(MergeStrategy)**: How to combine results
- **WithMergeFunc(func)**: Custom merge logic
- **WithTimeout(duration)**: Maximum execution time

## Merge Strategies

### MergeAll (Default)
All agent results are stored in `parallel_results` map:
```go
results := result.Get("parallel_results")
// results["agent1"] contains agent1's output
// results["agent2"] contains agent2's output
```

### MergeFirst
Returns the state from the first agent to complete successfully.

### MergeCustom
Apply custom logic to combine results:
```go
workflow.WithMergeFunc(func(results map[string]*domain.State) *domain.State {
    // Custom merge logic
    merged := domain.NewState()
    // Process results...
    return merged
})
```

## Next Steps

See the conditional workflow example for branching logic, or the loop workflow example for iterative processing.