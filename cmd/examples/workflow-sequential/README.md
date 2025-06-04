# Sequential Workflow Example

This example demonstrates how to use the new workflow agents to create sequential multi-step processes.

## Overview

The sequential workflow agent executes a series of agents one after another, passing the state from each agent to the next. This is useful for:

- Multi-step research workflows
- Data processing pipelines
- Document generation workflows
- Any process that requires sequential steps

## Features Demonstrated

1. **Sequential Execution**: Agents run one after another
2. **State Passing**: Each agent receives the state from the previous agent
3. **Error Handling**: Configure whether to stop on error or continue
4. **Retry Logic**: Built-in retry support for failed steps
5. **Event Monitoring**: Track workflow progress through events

## Usage

```bash
# Run with default mock provider
go run main.go

# Run with actual LLM providers (requires API keys)
export ANTHROPIC_API_KEY=your-key
export OPENAI_API_KEY=your-key
go run main.go
```

## Example Workflow

The main example creates a research workflow:

1. **Question Generator** (Claude): Generates research questions about a topic
2. **Researcher** (GPT-4): Answers the generated questions
3. **Summarizer** (Claude): Summarizes the findings into key insights

## Code Structure

```go
// Create agents
questionGenerator, _ := core.NewAgentFromString("generator", "claude")
researcher, _ := core.NewAgentFromString("researcher", "gpt-4")
summarizer, _ := core.NewAgentFromString("summarizer", "claude")

// Create workflow
workflow := workflow.NewSequentialAgent("research").
    WithStopOnError(true).
    AddAgent(questionGenerator).
    AddAgent(researcher).
    AddAgent(summarizer)

// Run workflow
result, err := workflow.Run(ctx, initialState)
```

## Configuration Options

- **WithStopOnError(bool)**: Stop workflow on first error (default: true)
- **WithMaxRetries(int)**: Maximum retries per step (default: 0)
- **SetErrorHandler(ErrorHandler)**: Custom error handling logic

## Next Steps

See the parallel workflow example for running agents concurrently, or the conditional workflow example for branching logic.