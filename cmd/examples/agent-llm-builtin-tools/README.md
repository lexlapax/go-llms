# LLM Agent with Built-in Tools Example

This example demonstrates how to use LLM agents with all categories of built-in tools. It shows proper system prompts and tool integration patterns for different use cases.

## Overview

LLM agents can use tools to extend their capabilities beyond text generation. This example shows how to:
- Attach built-in tools to LLM agents
- Write effective system prompts that guide tool usage
- Handle different tool categories for specific tasks
- Monitor tool calls and results

## Prerequisites

You need at least one LLM provider configured:

```bash
# Option 1: OpenAI
export OPENAI_API_KEY="your-api-key"

# Option 2: Anthropic  
export ANTHROPIC_API_KEY="your-api-key"

# Option 3: Google Gemini
export GEMINI_API_KEY="your-api-key"

# Option 4: Using GO_LLMS environment variables
export GO_LLMS_PROVIDER="openai"
export GO_LLMS_MODEL="gpt-4-turbo-preview"
export GO_LLMS_OPENAI_API_KEY="your-api-key"
```

## Running the Examples

```bash
# Build the example
go build -o agent-llm-builtin-tools .

# Run different scenarios
./agent-llm-builtin-tools research "Find information about Go generics"
./agent-llm-builtin-tools files "List all Go files and show their sizes"
./agent-llm-builtin-tools system "Show system memory usage"
./agent-llm-builtin-tools data "Parse this JSON: {\"name\":\"test\", \"value\":42}"
./agent-llm-builtin-tools datetime "When is the next Friday the 13th?"
./agent-llm-builtin-tools feeds "Get latest news from Hacker News"

# Run with all tools available
./agent-llm-builtin-tools all-tools "Show me what you can do"
```

## Available Scenarios

### 1. Research Assistant (`research`)
Uses web tools to research topics and compile reports:
- `web_search` - Search for information
- `web_fetch` - Retrieve full web pages
- `web_scrape` - Extract structured data
- `file_write` - Save findings

Example prompts:
- "Research Go generics and create a summary"
- "Find the latest AI developments and save to report.txt"
- "Compare Python and Go for web development"

### 2. File Manager (`files`)
Manages files and directories:
- `file_list` - List directory contents
- `file_read` - Read file contents
- `file_write` - Create or modify files
- `file_move` - Move or rename files
- `file_search` - Search within files

Example prompts:
- "List all .go files in the current directory"
- "Find files containing 'TODO' comments"
- "Create a backup of config.json"

### 3. System Administrator (`system`)
Monitors and manages system resources:
- `get_system_info` - System details
- `process_list` - Running processes
- `get_environment_variable` - Environment config
- `execute_command` - Run commands (carefully)

Example prompts:
- "Show system memory and CPU info"
- "List top 10 processes by CPU usage"
- "Check if GOPATH is set"

### 4. Data Analyst (`data`)
Processes and analyzes structured data:
- `json_process` - JSON operations
- `csv_process` - CSV analysis
- `xml_process` - XML parsing
- `data_transform` - Data transformations

Example prompts:
- "Calculate average from this data: [1,2,3,4,5]"
- "Parse CSV: name,age\\nAlice,30\\nBob,25"
- "Extract all email addresses from this JSON"

### 5. Scheduler (`datetime`)
Handles date and time operations:
- `datetime_now` - Current time
- `datetime_parse` - Parse dates
- `datetime_calculate` - Date math
- `datetime_format` - Format dates
- `datetime_compare` - Compare dates
- `datetime_info` - Date details
- `datetime_convert` - Timezone conversion

Example prompts:
- "What day is 30 days from now?"
- "Convert 3pm EST to PST"
- "How many business days until December 25?"

### 6. News Curator (`feeds`)
Processes RSS/Atom feeds:
- `feed_discover` - Find feeds
- `feed_fetch` - Get feed content
- `feed_filter` - Filter items
- `feed_aggregate` - Combine feeds
- `feed_extract` - Extract data

Example prompts:
- "Get latest tech news from TechCrunch"
- "Find feeds on golang.org"
- "Filter news about AI from the last week"

### 7. Universal Assistant (`all-tools`)
Has access to ALL tools - demonstrates the full capability of the system.

## Key Concepts

### System Prompts
The system prompt is crucial for guiding the LLM's tool usage. A good system prompt:
1. Lists available tools with clear descriptions
2. Explains when to use each tool
3. Provides safety guidelines for dangerous operations
4. Gives step-by-step workflows for complex tasks

### Tool Integration
```go
// Add tools to an agent
agent.AddTool(tools.MustGetTool("web_search"))

// The LLM will automatically:
// 1. Decide when to use the tool
// 2. Format parameters correctly
// 3. Process the results
// 4. Continue with the task
```

### Tool Call Monitoring
The example includes a hook that logs all tool calls:
```
[Tool Call] web_search with params: map[query:Go generics]
[Tool Result] web_search: [{"title":"Go Generics Tutorial","url":"..."}]
```

## Safety Considerations

Some tools can modify system state:
- `file_delete` - Can delete files
- `file_write` - Can overwrite files
- `execute_command` - Can run any command
- `file_move` - Can move/rename files

The example includes safety prompts that encourage:
- Confirmation before destructive operations
- Read-only commands for system tools
- Backup creation before modifications
- Clear explanations of what will happen

## Extending the Example

To add your own scenarios:

1. Create a new function like `runMyScenario()`
2. Select appropriate tools
3. Write a system prompt that guides their usage
4. Add it to the switch statement in main()

## Common Patterns

### Research Workflow
1. Search for information
2. Fetch detailed content
3. Extract key data
4. Summarize findings
5. Save results

### Data Processing Pipeline
1. Read/receive data
2. Parse into structured format
3. Apply transformations
4. Generate insights
5. Output results

### System Monitoring
1. Check system status
2. List processes
3. Examine specific metrics
4. Generate report

## Troubleshooting

### "No LLM provider configured"
Set one of the environment variables listed in Prerequisites.

### "Tool not found"
Make sure all tool packages are imported (the `_` imports at the top of main.go).

### Tool returns unexpected results
Check that your prompt clearly specifies what you want. The LLM interprets your request and decides tool parameters.

### Rate limiting
Some tools (especially web tools) may have rate limits. Add delays if needed.

## Next Steps

- Try combining multiple scenarios
- Create custom workflows using multiple tools
- Experiment with different system prompts
- Add your own tools to the registry