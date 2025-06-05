# Multi-Provider Workflow Example

This example demonstrates how to implement multi-provider strategies using workflow agents. It shows three common patterns for working with multiple LLM providers:

## Patterns Demonstrated

### 1. Fastest Response Pattern
Uses a parallel workflow with `MergeFirst` strategy to return the first response from any provider. This is useful when:
- You want the lowest latency possible
- All providers give similar quality responses
- You have redundancy requirements

### 2. Consensus Pattern
Uses a parallel workflow with a custom merge function to compare responses from multiple providers. This is useful when:
- You need high confidence in the answer
- Working with factual questions
- Want to detect and handle disagreements between models

### 3. Primary with Fallback Pattern
Uses a sequential workflow with error handling to try providers in order. This is useful when:
- You have a preferred provider but need reliability
- Want to control costs (try cheaper providers first)
- Need graceful degradation

## Running the Example

The example works with real providers if API keys are available, or falls back to mock agents for demonstration:

```bash
# With real providers (set any combination of these)
export OPENAI_API_KEY=your-key
export ANTHROPIC_API_KEY=your-key  
export GEMINI_API_KEY=your-key

# Run the example
go run main.go
```

## Key Concepts

1. **Workflow Agents**: The example uses `ParallelAgent` and `SequentialAgent` to orchestrate multiple LLM agents.

2. **Merge Strategies**: 
   - `MergeFirst`: Returns the first result
   - `MergeAll`: Combines all results
   - Custom merge functions for consensus logic

3. **Error Handling**: The sequential workflow demonstrates fallback behavior with `ErrorActionContinue`.

4. **State Management**: All agents use the state-based execution model with proper state passing between steps.

## Comparison with Provider-Level Multi

This example shows multi-provider patterns at the agent level, which provides:
- Better integration with the agent framework
- Access to hooks, events, and state management
- Ability to mix different types of agents (not just LLM providers)
- More flexibility in orchestration patterns

For provider-level multi-provider usage, see the `multi` and `consensus` examples.