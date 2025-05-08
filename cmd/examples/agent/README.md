# Agent Example

This example demonstrates how to use the agent workflow capabilities of the Go-LLMs library to create a system that can use tools to interact with its environment.

## Overview

The Agent example showcases:

1. Creating an agent with multiple LLM provider options (OpenAI, Anthropic, or Mock)
2. Adding various monitoring hooks for logging and metrics
3. Configuring the agent with different tools
4. Running the agent with different types of requests
5. Response caching to improve performance on repeated queries
6. Parallel tool execution
7. Structured outputs with schemas

## Tools Demonstrated

The agent is configured with the following tools:

- **Get Current Date** - Retrieves current date and time information
- **Calculator** - Performs mathematical calculations
- **Web Fetch** - Retrieves content from URLs
- **Read File** - Reads content from files
- **Write File** - Writes content to files
- **Execute Command** - Executes simple system commands (restricted to safe operations)

## Examples

1. **Basic Tool Usage** - Simple query utilizing calculator and date tools
2. **Multiple Tool Usage** - Query triggering multiple tool calls in one response
3. **Parallel Tool Execution** - Operations performed in parallel for efficiency
4. **Caching with Repeated Queries** - Demonstrates performance improvements with cached responses
5. **Structured Output** - Complex analysis with schema-driven structured response

## Running the Example

To run the example:

```bash
# With OpenAI API key
export OPENAI_API_KEY=your_api_key_here
make example EXAMPLE=agent
./bin/agent

# With Anthropic API key
export ANTHROPIC_API_KEY=your_api_key_here
make example EXAMPLE=agent
./bin/agent

# Without API keys (uses mock provider)
make example EXAMPLE=agent
./bin/agent
```

## Customizing

The agent system is highly customizable:

- Different LLM providers can be swapped in
- Custom tools can be added using the `AddTool` method
- Hooks can be added to monitor and log agent activity
- The system prompt can be modified to change agent behavior
- Schema definitions can be created for structured outputs

## Key Features

- **Performance Optimization** - Uses cached responses and parallel tool execution
- **Monitoring** - Comprehensive metrics for tool usage and performance
- **Logging** - Detailed logging of agent operations (to console and file)
- **Custom Hooks** - Example of creating custom monitoring hooks
- **Structured Output** - Schema-based validation of complex outputs