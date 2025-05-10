# Go-LLMs Command Line Interface

This directory contains the command line interface (CLI) for the Go-LLMs library, providing direct access to the library's key features through simple commands.

## Overview

The Go-LLMs CLI offers five main commands:

1. **chat** - Interactive chat session with an LLM
2. **complete** - One-shot text completion
3. **agent** - Tool-equipped agent for complex tasks
4. **structured** - Schema-based structured output generation
5. **completion** - Generate shell autocompletion scripts

All commands support multiple LLM providers (OpenAI, Anthropic, Gemini, or mock providers for testing) and various configuration options.

## Installation

Build the CLI using the Makefile:

```bash
# From the project root
make build

# The binary will be available at bin/go-llms
```

## Configuration

The CLI can be configured through:

1. **Command line flags**
2. **Environment variables**
3. **Configuration file**

### Configuration File

By default, the CLI looks for configuration in:

- `$HOME/.go-llms.yaml`
- `$HOME/.config/go-llms/config.yaml`
- `./go-llms.yaml` (current directory)

You can specify a custom config file with the `--config` flag.

Example configuration file format:

```yaml
provider: anthropic  # Default provider to use
output: text  # Default output format

providers:
  openai:
    api_key: "your-openai-api-key"
    default_model: "gpt-4o"

  anthropic:
    api_key: "your-anthropic-api-key"
    default_model: "claude-3-5-sonnet-latest"

  gemini:
    api_key: "your-gemini-api-key"
    default_model: "gemini-2.0-flash-lite"
```

## Global Flags

These flags apply to all commands:

```
--config string    Configuration file path
--provider string  LLM provider to use (openai, anthropic, gemini, mock) (default: "openai")
--model string     Model to use with the provider
--verbose, -v      Enable verbose output
--output, -o       Output format (text, json) (default: "text")
```

## Commands

### Chat

Start an interactive chat session with an LLM.

```bash
go-llms chat [flags]
```

Flags:
```
-s, --system string        System prompt for the chat
-t, --temperature float32  Temperature for generation (default 0.7)
--max-tokens int           Maximum tokens to generate (default 1000)
--stream                   Stream responses in real-time (auto-enabled in interactive terminals)
--no-stream                Disable streaming (useful for scripts or logging)
```

By default, streaming is automatically enabled when running in an interactive terminal and disabled when the output is redirected or piped. This provides a natural conversational experience while allowing flexibility for scripting.

Example:
```bash
# Start a chat session with default provider (auto-streaming in terminal)
go-llms chat

# Start a chat with streaming explicitly enabled
go-llms chat --stream

# Start a chat with streaming disabled
go-llms chat --no-stream

# Start a chat with a custom system prompt and Anthropic provider
go-llms chat --provider anthropic --system "You are a helpful coding assistant."
```

### Complete

One-shot text completion with an LLM.

```bash
go-llms complete [prompt] [flags]
```

Flags:
```
-t, --temperature float32  Temperature for generation (default 0.7)
--max-tokens int           Maximum tokens to generate (default 1000)
--stream                   Stream the response token by token
```

Example:
```bash
# Get a completion
go-llms complete "Explain quantum computing in simple terms"

# Stream the response
go-llms complete --stream "Write a poem about Go programming"

# Use a specific model
go-llms complete --provider openai --model gpt-4o "Summarize the key features of Go"
```

### Agent

Execute agent workflows with tool integration.

```bash
go-llms agent [input] [flags]
```

Flags:
```
-t, --tools strings         Tools to enable (comma-separated) (default: calculator,date)
                            Available: calculator, date, web, read_file, write_file, execute_command
-s, --system string         System prompt for the agent
-S, --schema string         Path to output schema file
-T, --temperature float32   Temperature for generation (default 0.7)
```

Example:
```bash
# Basic agent with default tools
go-llms agent "What is 25 * 42?"

# Agent with web tool enabled
go-llms agent --tools calculator,date,web_fetch "What is today's date and what is the title of example.com?"

# Agent with file tools and custom system prompt
go-llms agent --tools calculator,date,read_file,write_file \
  --system "You are a helpful assistant that can work with files." \
  "Read the content of ./data.txt and summarize it"

# Agent with structured output
go-llms agent --schema ./schemas/summary.json \
  "Analyze this text and extract the key points"
```

### Structured

Generate structured outputs using schemas.

```bash
go-llms structured [prompt] [flags]
```

Flags:
```
-s, --schema string         Path to schema file (required)
-O, --output-file string    File to write result to
--validate                  Validate output against schema (default true)
-t, --temperature float32   Temperature for generation (default 0.7)
--max-tokens int            Maximum tokens to generate (default 1000)
```

Example:
```bash
# Generate a structured output based on a schema
go-llms structured --schema ./schemas/person.json \
  "Generate information about a fictional software developer"

# Save the output to a file
go-llms structured --schema ./schemas/recipe.json --output-file recipe.json \
  "Generate a recipe for vegetarian lasagna"
```

## Environment Variables

API keys can be provided through environment variables:

```bash
# For OpenAI
export OPENAI_API_KEY=your-openai-api-key

# For Anthropic
export ANTHROPIC_API_KEY=your-anthropic-api-key

# For Google Gemini
export GEMINI_API_KEY=your-gemini-api-key
```

## Tools Reference

The agent command supports these tools:

| Tool Name        | Description                                 | Parameters                                     |
|------------------|---------------------------------------------|------------------------------------------------|
| calculator       | Performs mathematical calculations          | expression: mathematical expression (e.g., "2 * 3") |
| date             | Gets current date and time information      | None                                           |
| web_fetch        | Fetches content from a URL                  | url: URL to fetch                              |
| read_file        | Reads content from a file                   | path: Path to file                             |
| write_file       | Writes content to a file                    | path: Path to file, content: Content to write  |
| execute_command  | Executes a shell command                    | command: Command to execute, timeout: Timeout in seconds |

## Schema Format

The `structured` command and `agent` command (with `--schema`) require a schema file in JSON format following this structure:

```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Person's name"
    },
    "age": {
      "type": "integer",
      "minimum": 0,
      "maximum": 120
    },
    "email": {
      "type": "string",
      "format": "email",
      "description": "Email address"
    }
  },
  "required": ["name", "email"]
}
```

## Examples

### Chat Example
```bash
$ go-llms chat --provider anthropic
Chat session started with anthropic using claude-3-5-sonnet-latest
Streaming mode enabled: responses will appear in real-time
Type 'exit' or 'quit' to end the session
-----------------------------------------

User: What are the benefits of using Go for LLM applications?
Assistant: Go offers several benefits for building LLM applications:

1. Excellent concurrency support through goroutines and channels, making it ideal for handling multiple LLM requests simultaneously
2. Strong performance with near-C speeds but with memory safety
3. Static typing helps catch errors at compile time
4. Built-in testing framework
5. Simple dependency management
6. Fast compilation and single binary deployment
7. Excellent standard library including HTTP server/client implementations
8. Low memory footprint compared to languages like Python or Java
9. Clean syntax that's easy to read and maintain
10. Strong cross-platform support

These qualities make Go particularly well-suited for building reliable, high-performance LLM services.

User: exit
Ending chat session
```

With streaming enabled, the response appears progressively in real-time rather than all at once after a delay.

### Complete Example
```bash
$ go-llms complete --stream "Write a haiku about Go programming"
Streaming response:
Concurrent routines
Simple syntax flows like streams
Fast code, peace of mind
```

### Agent Example
```bash
$ go-llms agent --tools calculator,date "What is today's date and what is 25 * 42?"
Agent Result:
--------------
Today's date is May 8, 2025. The calculation of 25 * 42 equals 1050.
```

### Structured Example
```bash
$ go-llms structured --schema ./schemas/person.json "Generate information about a fictional programmer"
Generating structured output...

Structured Output:
{
  "name": "Alex Morgan",
  "age": 34,
  "email": "alex.morgan@example.com",
  "profession": "Software Engineer",
  "skills": [
    "Go",
    "Kubernetes",
    "Machine Learning",
    "Distributed Systems"
  ],
  "yearsOfExperience": 12,
  "currentProject": "Building an LLM-powered code assistant"
}
```

## Advanced Usage

### Using Different Providers

```bash
# Use OpenAI
go-llms --provider openai chat

# Use Anthropic
go-llms --provider anthropic complete "Explain concurrency in Go"

# Use Google Gemini
go-llms --provider gemini complete "Explain Go interfaces"

# Use mock provider for testing
go-llms --provider mock agent "What is 2+2?"
```

### Verbose Output

Use the `-v` flag to see detailed logs:

```bash
go-llms -v agent --tools calculator,date,web "Check the title of example.com and calculate 15*7"
```

### Custom Models

Override the default model:

```bash
go-llms --provider openai --model gpt-4-turbo-preview chat
go-llms --provider anthropic --model claude-3-opus-20240229 complete "Tell me a story"
go-llms --provider gemini --model gemini-2.0-pro-latest complete "Explain quantum computing"
```

### Shell Autocompletion

Generate shell autocompletion scripts to enable tab-completion for commands, flags, and arguments:

```bash
# For Bash
go-llms completion bash > ~/.bash_completion
source ~/.bash_completion

# For Zsh
go-llms completion zsh > "${fpath[1]}/_go-llms"
source ~/.zshrc

# For Fish
go-llms completion fish > ~/.config/fish/completions/go-llms.fish

# For PowerShell
go-llms completion powershell > go-llms.ps1
```

This enables tab-completion for all commands, subcommands, and flags, making the CLI easier to use interactively.